"""
Physics Environment gRPC Server
Streams degraded RF signals to the array controller, simulating deep space communication.
"""
import os
import sys
import time
import math
import logging
import json
from concurrent import futures

import grpc
import numpy as np

# Add parent directory to path for proto imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))

from app.ephemeris.skyfield_engine import (
    DSN_STATIONS, SPACECRAFT_TARGETS, AU_TO_KM,
    compute_distance_km, compute_azimuth_elevation,
    compute_relative_velocity, compute_geometric_delay_us
)
from app.propagation.path_loss import fspl_db, antenna_gain_db, received_power_dbm
from app.propagation.doppler_shift import doppler_shift_hz, observed_frequency
from app.propagation.atmosphere import (
    generate_weather, atmospheric_attenuation_db,
    system_noise_temperature_k, WeatherCondition
)
from app.spacecraft.transmitter import (
    generate_spacecraft_signal, DEFAULT_CARRIER_HZ, DEFAULT_TX_POWER_DBM
)

import proto.dsn_pb2 as dsn_pb2
import proto.dsn_pb2_grpc as dsn_pb2_grpc

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

DEFAULT_SAMPLE_RATE = 10000.0
DEFAULT_CHUNK_SIZE = 1000
TX_ANTENNA_GAIN_DB = 36.0  # Typical spacecraft high-gain antenna


class PhysicsEnvironmentServicer(dsn_pb2_grpc.PhysicsEnvironmentServicer):
    """gRPC servicer that streams degraded signals for each antenna."""

    def __init__(self):
        self.start_time = time.time()
        logger.info("Physics Environment initialized")

    def StreamSignals(self, request, context):
        """Stream degraded signal chunks for each requested antenna."""
        spacecraft_id = request.spacecraft_id or "voyager1"
        antenna_ids = list(request.antenna_ids) or list(DSN_STATIONS.keys())[:3]
        carrier_freq = request.carrier_frequency_hz or DEFAULT_CARRIER_HZ
        sample_rate = request.sample_rate_hz or DEFAULT_SAMPLE_RATE
        chunk_size = request.chunk_size or DEFAULT_CHUNK_SIZE

        target = SPACECRAFT_TARGETS.get(spacecraft_id)
        if not target:
            context.abort(grpc.StatusCode.NOT_FOUND, f"Unknown spacecraft: {spacecraft_id}")
            return

        logger.info(f"Starting signal stream for {spacecraft_id} -> antennas: {antenna_ids}")

        # Generate the source signal (same for all antennas since it's one spacecraft)
        num_bits = chunk_size
        i_source, q_source, original_bits = generate_spacecraft_signal(
            num_bits=num_bits,
            modulation="BPSK",
            carrier_freq_hz=carrier_freq,
            sample_rate_hz=sample_rate,
            seed=42
        )

        chunk_idx = 0
        while context.is_active():
            t = time.time() - self.start_time
            
            for antenna_id in antenna_ids:
                station = DSN_STATIONS.get(antenna_id)
                if not station:
                    continue

                # Calculate physics
                distance = compute_distance_km(station, target, t)
                az, el = compute_azimuth_elevation(station, target, t)
                rel_vel = compute_relative_velocity(station, target)
                doppler = doppler_shift_hz(carrier_freq, rel_vel)

                # Path loss
                path_loss = fspl_db(distance, carrier_freq)
                rx_gain = antenna_gain_db(station.dish_diameter_m, carrier_freq)
                rx_power = received_power_dbm(DEFAULT_TX_POWER_DBM, TX_ANTENNA_GAIN_DB, rx_gain, distance, carrier_freq)

                # Atmospheric effects
                weather = generate_weather(antenna_id, seed=int(t * 10) % 1000)
                atm_atten = atmospheric_attenuation_db(weather, max(el, 5.0))
                sys_noise_temp = system_noise_temperature_k(weather)

                # Calculate SNR
                # SNR = Pr - noise_floor
                k_boltzmann = 1.380649e-23
                bandwidth = sample_rate
                noise_power_w = k_boltzmann * sys_noise_temp * bandwidth
                noise_power_dbm = 10.0 * math.log10(noise_power_w) + 30.0
                snr_db = rx_power - atm_atten - noise_power_dbm

                # Degrade the signal: add noise based on SNR
                signal_power = np.mean(i_source**2 + q_source**2)
                if signal_power > 0 and snr_db > -50:
                    noise_power_linear = signal_power / (10.0 ** (snr_db / 10.0))
                else:
                    noise_power_linear = signal_power * 100

                noise_std = math.sqrt(max(noise_power_linear, 1e-20))
                i_noisy = i_source + np.random.normal(0, noise_std, len(i_source))
                q_noisy = q_source + np.random.normal(0, noise_std, len(q_source))

                # Apply random phase offset (geometric delay effect)
                phase_offset = (distance % 1.0) * 2.0 * math.pi
                cos_p = math.cos(phase_offset)
                sin_p = math.sin(phase_offset)
                i_shifted = i_noisy * cos_p - q_noisy * sin_p
                q_shifted = i_noisy * sin_p + q_noisy * cos_p

                # Build and yield the signal chunk
                chunk = dsn_pb2.SignalChunk(
                    antenna_id=antenna_id,
                    timestamp=t,
                    i_samples=i_shifted.tolist(),
                    q_samples=q_shifted.tolist(),
                    snr_db=snr_db,
                    doppler_shift_hz=doppler,
                    path_loss_db=path_loss,
                    atmosphere_attenuation_db=atm_atten,
                    distance_km=distance,
                )
                yield chunk

            chunk_idx += 1
            time.sleep(0.1)  # 10 chunks per second


def serve():
    port = os.environ.get("GRPC_PORT", "50051")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    dsn_pb2_grpc.add_PhysicsEnvironmentServicer_to_server(
        PhysicsEnvironmentServicer(), server
    )
    server.add_insecure_port(f"[::]:{port}")
    server.start()
    logger.info(f"Physics Environment gRPC server running on port {port}")
    server.wait_for_termination()


if __name__ == "__main__":
    serve()

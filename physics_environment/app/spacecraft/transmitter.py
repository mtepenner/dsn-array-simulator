"""
Spacecraft Transmitter
Generates simulated source waveforms using BPSK or QPSK modulation.
"""
import math
import numpy as np
from typing import Tuple, List


# Standard DSN X-band downlink frequency
DEFAULT_CARRIER_HZ = 8.4e9  # 8.4 GHz X-band
DEFAULT_TX_POWER_DBM = 43.0  # ~20W typical spacecraft transmitter
DEFAULT_SYMBOL_RATE = 1000.0  # symbols/second


def generate_random_bits(num_bits: int, seed: int = None) -> np.ndarray:
    """Generate random telemetry data bits."""
    rng = np.random.RandomState(seed)
    return rng.randint(0, 2, num_bits)


def bpsk_modulate(bits: np.ndarray, carrier_freq_hz: float, sample_rate_hz: float,
                   symbol_rate: float = DEFAULT_SYMBOL_RATE) -> Tuple[np.ndarray, np.ndarray]:
    """
    Generate BPSK-modulated IQ samples.
    
    BPSK maps:
        bit 0 -> phase 0 (I=+1, Q=0)
        bit 1 -> phase π (I=-1, Q=0)
    
    Returns (i_samples, q_samples)
    """
    samples_per_symbol = int(sample_rate_hz / symbol_rate)
    total_samples = len(bits) * samples_per_symbol
    
    t = np.arange(total_samples) / sample_rate_hz
    
    # Map bits to symbols: 0 -> +1, 1 -> -1
    symbols = 1.0 - 2.0 * bits.astype(float)
    
    # Repeat each symbol for samples_per_symbol
    symbol_wave = np.repeat(symbols, samples_per_symbol)
    
    # Generate carrier
    carrier_phase = 2.0 * math.pi * carrier_freq_hz * t
    
    # BPSK: I channel carries the data, Q is zero (at baseband)
    # At baseband representation:
    i_samples = symbol_wave
    q_samples = np.zeros_like(symbol_wave)
    
    return i_samples, q_samples


def qpsk_modulate(bits: np.ndarray, carrier_freq_hz: float, sample_rate_hz: float,
                   symbol_rate: float = DEFAULT_SYMBOL_RATE) -> Tuple[np.ndarray, np.ndarray]:
    """
    Generate QPSK-modulated IQ samples.
    
    QPSK maps pairs of bits to phases:
        00 -> π/4    (I=+1/√2, Q=+1/√2)
        01 -> 3π/4   (I=-1/√2, Q=+1/√2)
        11 -> 5π/4   (I=-1/√2, Q=-1/√2)
        10 -> 7π/4   (I=+1/√2, Q=-1/√2)
    
    Returns (i_samples, q_samples)
    """
    # Ensure even number of bits
    if len(bits) % 2 != 0:
        bits = np.append(bits, 0)
    
    samples_per_symbol = int(sample_rate_hz / symbol_rate)
    num_symbols = len(bits) // 2
    
    # Split into even/odd bits
    i_bits = bits[0::2]  # even bits
    q_bits = bits[1::2]  # odd bits
    
    # Map to I/Q: 0 -> +1/√2, 1 -> -1/√2
    inv_sqrt2 = 1.0 / math.sqrt(2.0)
    i_symbols = (1.0 - 2.0 * i_bits.astype(float)) * inv_sqrt2
    q_symbols = (1.0 - 2.0 * q_bits.astype(float)) * inv_sqrt2
    
    # Repeat each symbol
    i_samples = np.repeat(i_symbols, samples_per_symbol)
    q_samples = np.repeat(q_symbols, samples_per_symbol)
    
    return i_samples, q_samples


def add_carrier(i_baseband: np.ndarray, q_baseband: np.ndarray,
                carrier_freq_hz: float, sample_rate_hz: float,
                doppler_hz: float = 0.0) -> Tuple[np.ndarray, np.ndarray]:
    """
    Mix baseband IQ signal up to IF/RF with optional Doppler shift.
    Returns IQ at the shifted frequency.
    """
    t = np.arange(len(i_baseband)) / sample_rate_hz
    effective_freq = carrier_freq_hz + doppler_hz
    
    cos_carrier = np.cos(2.0 * math.pi * effective_freq * t)
    sin_carrier = np.sin(2.0 * math.pi * effective_freq * t)
    
    # Complex mixing
    i_out = i_baseband * cos_carrier - q_baseband * sin_carrier
    q_out = i_baseband * sin_carrier + q_baseband * cos_carrier
    
    return i_out, q_out


def apply_power_scaling(i_samples: np.ndarray, q_samples: np.ndarray,
                        received_power_dbm: float) -> Tuple[np.ndarray, np.ndarray]:
    """Scale signal to match a given received power level in dBm."""
    current_power = np.mean(i_samples**2 + q_samples**2)
    if current_power <= 0:
        return i_samples, q_samples
    
    target_power = 10.0 ** ((received_power_dbm - 30.0) / 10.0)  # dBm to Watts
    scale = math.sqrt(target_power / current_power)
    
    return i_samples * scale, q_samples * scale


def generate_spacecraft_signal(num_bits: int = 100,
                                modulation: str = "BPSK",
                                carrier_freq_hz: float = DEFAULT_CARRIER_HZ,
                                sample_rate_hz: float = 10000.0,
                                symbol_rate: float = DEFAULT_SYMBOL_RATE,
                                seed: int = None) -> Tuple[np.ndarray, np.ndarray, np.ndarray]:
    """
    Generate a complete spacecraft transmission signal.
    
    Returns (i_samples, q_samples, original_bits)
    """
    bits = generate_random_bits(num_bits, seed=seed)
    
    if modulation.upper() == "QPSK":
        i_samples, q_samples = qpsk_modulate(bits, carrier_freq_hz, sample_rate_hz, symbol_rate)
    else:
        i_samples, q_samples = bpsk_modulate(bits, carrier_freq_hz, sample_rate_hz, symbol_rate)
    
    return i_samples, q_samples, bits

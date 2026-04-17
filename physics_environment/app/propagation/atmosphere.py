"""
Atmospheric Effects Model
Adds random noise and attenuation based on simulated Earth weather over each antenna.
"""
import math
import random
from dataclasses import dataclass
from enum import Enum
from typing import Tuple


class WeatherCondition(Enum):
    CLEAR = "clear"
    CLOUDY = "cloudy"
    LIGHT_RAIN = "light_rain"
    HEAVY_RAIN = "heavy_rain"
    THUNDERSTORM = "thunderstorm"


@dataclass
class AtmosphereState:
    """Current atmospheric conditions at a ground station."""
    weather: WeatherCondition
    temperature_c: float
    humidity_percent: float
    wind_speed_m_s: float
    cloud_cover_percent: float
    rain_rate_mm_hr: float

# Typical attenuation values for X-band (~8.4 GHz) in dB
WEATHER_ATTENUATION_DB = {
    WeatherCondition.CLEAR: (0.1, 0.3),
    WeatherCondition.CLOUDY: (0.2, 0.8),
    WeatherCondition.LIGHT_RAIN: (0.5, 2.0),
    WeatherCondition.HEAVY_RAIN: (2.0, 6.0),
    WeatherCondition.THUNDERSTORM: (5.0, 15.0),
}

# Noise temperature contributions in Kelvin
WEATHER_NOISE_TEMP_K = {
    WeatherCondition.CLEAR: (15.0, 25.0),
    WeatherCondition.CLOUDY: (25.0, 50.0),
    WeatherCondition.LIGHT_RAIN: (50.0, 100.0),
    WeatherCondition.HEAVY_RAIN: (100.0, 200.0),
    WeatherCondition.THUNDERSTORM: (200.0, 300.0),
}


def generate_weather(station_id: str, seed: int = None) -> AtmosphereState:
    """Generate random but plausible weather for a ground station."""
    if seed is not None:
        rng = random.Random(seed)
    else:
        rng = random.Random()
    
    weather = rng.choice(list(WeatherCondition))
    
    return AtmosphereState(
        weather=weather,
        temperature_c=rng.uniform(-5.0, 45.0),
        humidity_percent=rng.uniform(10.0, 95.0),
        wind_speed_m_s=rng.uniform(0.0, 30.0),
        cloud_cover_percent=rng.uniform(0.0, 100.0),
        rain_rate_mm_hr={
            WeatherCondition.CLEAR: 0.0,
            WeatherCondition.CLOUDY: 0.0,
            WeatherCondition.LIGHT_RAIN: rng.uniform(0.5, 5.0),
            WeatherCondition.HEAVY_RAIN: rng.uniform(5.0, 25.0),
            WeatherCondition.THUNDERSTORM: rng.uniform(25.0, 100.0),
        }[weather],
    )


def atmospheric_attenuation_db(atmosphere: AtmosphereState, elevation_deg: float) -> float:
    """
    Calculate atmospheric attenuation in dB based on weather and elevation angle.
    Lower elevation = more atmosphere to pass through (cosecant law).
    """
    min_atten, max_atten = WEATHER_ATTENUATION_DB[atmosphere.weather]
    base_atten = random.uniform(min_atten, max_atten)
    
    # Cosecant law: attenuation increases at lower elevation angles
    if elevation_deg < 5.0:
        elevation_deg = 5.0  # Avoid extreme values near horizon
    csc_factor = 1.0 / math.sin(math.radians(elevation_deg))
    
    return base_atten * csc_factor


def atmospheric_noise_temperature_k(atmosphere: AtmosphereState) -> float:
    """
    Calculate the noise temperature contribution from the atmosphere in Kelvin.
    """
    min_temp, max_temp = WEATHER_NOISE_TEMP_K[atmosphere.weather]
    return random.uniform(min_temp, max_temp)


def system_noise_temperature_k(atmosphere: AtmosphereState, 
                                receiver_temp_k: float = 20.0,
                                cosmic_background_k: float = 2.725) -> float:
    """
    Calculate total system noise temperature.
    T_sys = T_receiver + T_atmosphere + T_cosmic
    """
    t_atm = atmospheric_noise_temperature_k(atmosphere)
    return receiver_temp_k + t_atm + cosmic_background_k


def add_atmospheric_noise(samples: list, snr_db: float, atmosphere: AtmosphereState) -> Tuple[list, float]:
    """
    Add atmospheric noise to signal samples.
    Returns (noisy_samples, effective_snr_db).
    """
    import numpy as np
    
    signal = np.array(samples)
    signal_power = np.mean(signal ** 2)
    
    # Additional attenuation from atmosphere
    extra_atten_db = atmospheric_attenuation_db(atmosphere, 45.0)  # assume 45 deg elevation
    effective_snr_db = snr_db - extra_atten_db
    
    # Generate noise
    noise_power = signal_power / (10 ** (effective_snr_db / 10.0)) if effective_snr_db > -50 else signal_power * 100
    noise = np.random.normal(0, math.sqrt(noise_power), len(signal))
    
    noisy_signal = signal + noise
    return noisy_signal.tolist(), effective_snr_db

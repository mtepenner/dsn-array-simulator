"""
Doppler Shift Calculator
Shifts the carrier frequency based on relative velocity between station and spacecraft.
"""
import math

SPEED_OF_LIGHT_KM_S = 299792.458


def doppler_shift_hz(carrier_freq_hz: float, relative_velocity_km_s: float) -> float:
    """
    Calculate the Doppler frequency shift.
    
    Positive velocity = receding (redshift, frequency decreases)
    Negative velocity = approaching (blueshift, frequency increases)
    
    f_observed = f_source * (1 - v/c)  (non-relativistic approximation)
    """
    shift = -carrier_freq_hz * (relative_velocity_km_s / SPEED_OF_LIGHT_KM_S)
    return shift


def observed_frequency(carrier_freq_hz: float, relative_velocity_km_s: float) -> float:
    """
    Calculate the observed frequency after Doppler shift.
    """
    return carrier_freq_hz + doppler_shift_hz(carrier_freq_hz, relative_velocity_km_s)


def relativistic_doppler_shift(carrier_freq_hz: float, relative_velocity_km_s: float) -> float:
    """
    Calculate relativistic Doppler shift for higher accuracy at extreme velocities.
    
    f_observed = f_source * sqrt((1 - beta) / (1 + beta))
    where beta = v/c
    """
    beta = relative_velocity_km_s / SPEED_OF_LIGHT_KM_S
    if abs(beta) >= 1.0:
        return 0.0
    factor = math.sqrt((1.0 - beta) / (1.0 + beta))
    return carrier_freq_hz * factor


def doppler_rate_hz_per_s(carrier_freq_hz: float, acceleration_km_s2: float) -> float:
    """
    Calculate the rate of change of Doppler shift (Hz/s).
    Used for tracking loop bandwidth calculations.
    """
    return -carrier_freq_hz * (acceleration_km_s2 / SPEED_OF_LIGHT_KM_S)

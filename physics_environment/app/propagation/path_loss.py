"""
Free-Space Path Loss (FSPL) Calculator
Calculates signal attenuation over millions of miles of deep space.
"""
import math


def fspl_db(distance_km: float, frequency_hz: float) -> float:
    """
    Calculate Free-Space Path Loss in dB.
    
    FSPL(dB) = 20*log10(d) + 20*log10(f) + 20*log10(4*pi/c)
    
    where:
        d = distance in meters
        f = frequency in Hz
        c = speed of light (m/s)
    """
    if distance_km <= 0 or frequency_hz <= 0:
        return 0.0
    
    c = 299792458.0  # speed of light in m/s
    distance_m = distance_km * 1000.0
    
    fspl = (20.0 * math.log10(distance_m) 
            + 20.0 * math.log10(frequency_hz) 
            + 20.0 * math.log10(4.0 * math.pi / c))
    
    return fspl


def apply_path_loss(signal_power_dbm: float, distance_km: float, frequency_hz: float) -> float:
    """
    Apply free-space path loss to a signal.
    Returns the received signal power in dBm.
    """
    loss = fspl_db(distance_km, frequency_hz)
    return signal_power_dbm - loss


def antenna_gain_db(diameter_m: float, frequency_hz: float, efficiency: float = 0.55) -> float:
    """
    Calculate antenna gain in dB.
    G = eta * (pi * D / lambda)^2
    """
    c = 299792458.0
    wavelength = c / frequency_hz
    gain_linear = efficiency * (math.pi * diameter_m / wavelength) ** 2
    if gain_linear <= 0:
        return 0.0
    return 10.0 * math.log10(gain_linear)


def received_power_dbm(tx_power_dbm: float, tx_gain_db: float, rx_gain_db: float,
                        distance_km: float, frequency_hz: float) -> float:
    """
    Calculate received signal power using the Friis transmission equation.
    Pr = Pt + Gt + Gr - FSPL
    """
    loss = fspl_db(distance_km, frequency_hz)
    return tx_power_dbm + tx_gain_db + rx_gain_db - loss

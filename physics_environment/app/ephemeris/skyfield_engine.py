"""
Skyfield Ephemeris Engine
Calculates exact vectors between Earth stations and spacecraft using simplified models.
"""
import math
import time
from dataclasses import dataclass
from typing import Tuple

@dataclass
class CelestialPosition:
    """Position in 3D space (km)."""
    x: float
    y: float
    z: float

@dataclass
class GroundStation:
    """An Earth-based antenna ground station."""
    station_id: str
    name: str
    latitude: float   # degrees
    longitude: float  # degrees
    elevation_m: float
    dish_diameter_m: float = 34.0

# Pre-defined DSN ground stations (real-world approximate locations)
DSN_STATIONS = {
    "DSS-14": GroundStation("DSS-14", "Goldstone 70m", 35.4267, -116.89, 1001.0, 70.0),
    "DSS-24": GroundStation("DSS-24", "Goldstone 34m", 35.3395, -116.8756, 988.0, 34.0),
    "DSS-25": GroundStation("DSS-25", "Goldstone 34m BWG", 35.3382, -116.8750, 987.0, 34.0),
    "DSS-34": GroundStation("DSS-34", "Canberra 34m", -35.3985, 148.9816, 692.0, 34.0),
    "DSS-35": GroundStation("DSS-35", "Canberra 34m BWG", -35.3955, 148.9786, 694.0, 34.0),
    "DSS-43": GroundStation("DSS-43", "Canberra 70m", -35.4023, 148.9814, 689.0, 70.0),
    "DSS-54": GroundStation("DSS-54", "Madrid 34m", 40.4313, -4.2480, 837.0, 34.0),
    "DSS-55": GroundStation("DSS-55", "Madrid 34m BWG", 40.4325, -4.2510, 835.0, 34.0),
    "DSS-63": GroundStation("DSS-63", "Madrid 70m", 40.4314, -4.2481, 833.0, 70.0),
}

EARTH_RADIUS_KM = 6371.0

@dataclass
class SpacecraftTarget:
    """A deep-space target with orbital parameters."""
    name: str
    distance_au: float  # distance from Earth in AU
    velocity_km_s: float  # relative velocity towards/away from Earth
    ra_deg: float  # right ascension in degrees
    dec_deg: float  # declination in degrees

# Pre-defined spacecraft targets
SPACECRAFT_TARGETS = {
    "voyager1": SpacecraftTarget("Voyager 1", 159.0, 17.0, 257.8, 12.2),
    "voyager2": SpacecraftTarget("Voyager 2", 133.0, 15.4, 296.3, -57.5),
    "newHorizons": SpacecraftTarget("New Horizons", 57.0, 13.8, 282.5, -20.8),
    "mars_orbiter": SpacecraftTarget("Mars Orbiter", 1.5, 24.1, 45.0, 22.0),
}

AU_TO_KM = 1.496e8  # 1 AU in km


def station_ecef(station: GroundStation) -> Tuple[float, float, float]:
    """Convert station lat/lon/elevation to ECEF coordinates (km)."""
    lat_r = math.radians(station.latitude)
    lon_r = math.radians(station.longitude)
    r = EARTH_RADIUS_KM + station.elevation_m / 1000.0
    x = r * math.cos(lat_r) * math.cos(lon_r)
    y = r * math.cos(lat_r) * math.sin(lon_r)
    z = r * math.sin(lat_r)
    return (x, y, z)


def spacecraft_position(target: SpacecraftTarget, t: float) -> Tuple[float, float, float]:
    """
    Calculate spacecraft position in ECEF-like coordinates.
    Uses RA/Dec and distance, with slight time-varying motion.
    """
    dist_km = target.distance_au * AU_TO_KM
    # Add slight radial motion over time
    dist_km += target.velocity_km_s * t
    
    ra_r = math.radians(target.ra_deg)
    dec_r = math.radians(target.dec_deg)
    
    x = dist_km * math.cos(dec_r) * math.cos(ra_r)
    y = dist_km * math.cos(dec_r) * math.sin(ra_r)
    z = dist_km * math.sin(dec_r)
    return (x, y, z)


def compute_distance_km(station: GroundStation, target: SpacecraftTarget, t: float) -> float:
    """Compute distance between a ground station and spacecraft in km."""
    sx, sy, sz = station_ecef(station)
    tx, ty, tz = spacecraft_position(target, t)
    dx = tx - sx
    dy = ty - sy
    dz = tz - sz
    return math.sqrt(dx*dx + dy*dy + dz*dz)


def compute_azimuth_elevation(station: GroundStation, target: SpacecraftTarget, t: float) -> Tuple[float, float]:
    """
    Compute azimuth and elevation from a ground station to a spacecraft.
    Returns (azimuth_deg, elevation_deg).
    """
    sx, sy, sz = station_ecef(station)
    tx, ty, tz = spacecraft_position(target, t)
    
    # Vector from station to spacecraft
    dx = tx - sx
    dy = ty - sy
    dz = tz - sz
    dist = math.sqrt(dx*dx + dy*dy + dz*dz)
    
    lat_r = math.radians(station.latitude)
    lon_r = math.radians(station.longitude)
    
    # Convert to local ENU (East, North, Up)
    e = -math.sin(lon_r) * dx + math.cos(lon_r) * dy
    n = (-math.sin(lat_r) * math.cos(lon_r) * dx 
         - math.sin(lat_r) * math.sin(lon_r) * dy 
         + math.cos(lat_r) * dz)
    u = (math.cos(lat_r) * math.cos(lon_r) * dx 
         + math.cos(lat_r) * math.sin(lon_r) * dy 
         + math.sin(lat_r) * dz)
    
    azimuth = math.degrees(math.atan2(e, n)) % 360.0
    elevation = math.degrees(math.asin(min(1.0, max(-1.0, u / dist))))
    
    return azimuth, elevation


def compute_relative_velocity(station: GroundStation, target: SpacecraftTarget) -> float:
    """
    Compute approximate relative radial velocity in km/s.
    Simplified: uses the spacecraft's bulk velocity plus Earth rotation effect.
    """
    # Earth rotation contribution at station latitude
    earth_rot_speed = 0.4651  # km/s at equator
    lat_r = math.radians(station.latitude)
    v_rot = earth_rot_speed * math.cos(lat_r)
    
    # Component along line of sight (simplified)
    return target.velocity_km_s + v_rot * 0.01  # small contribution


def compute_geometric_delay_us(station1: GroundStation, station2: GroundStation, 
                                target: SpacecraftTarget, t: float) -> float:
    """
    Compute the geometric delay in microseconds between two stations
    observing the same spacecraft signal.
    """
    d1 = compute_distance_km(station1, target, t)
    d2 = compute_distance_km(station2, target, t)
    c_km_s = 299792.458  # speed of light in km/s
    delay_s = (d2 - d1) / c_km_s
    return delay_s * 1e6  # convert to microseconds

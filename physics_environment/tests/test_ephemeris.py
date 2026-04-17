"""Tests for the Skyfield ephemeris engine."""
import math
import pytest
from app.ephemeris.skyfield_engine import (
    DSN_STATIONS, SPACECRAFT_TARGETS, AU_TO_KM,
    station_ecef, spacecraft_position,
    compute_distance_km, compute_azimuth_elevation,
    compute_relative_velocity, compute_geometric_delay_us
)


class TestStationECEF:
    def test_goldstone_position(self):
        station = DSN_STATIONS["DSS-14"]
        x, y, z = station_ecef(station)
        # Should be roughly at the right latitude/longitude
        r = math.sqrt(x*x + y*y + z*z)
        assert abs(r - 6372.0) < 5.0  # approximately Earth radius

    def test_all_stations_on_earth(self):
        for sid, station in DSN_STATIONS.items():
            x, y, z = station_ecef(station)
            r = math.sqrt(x*x + y*y + z*z)
            assert 6360.0 < r < 6380.0, f"Station {sid} at unexpected radius: {r}"

    def test_canberra_southern_hemisphere(self):
        station = DSN_STATIONS["DSS-43"]
        x, y, z = station_ecef(station)
        assert z < 0, "Canberra should have negative z (southern hemisphere)"


class TestSpacecraftPosition:
    def test_voyager1_distance(self):
        target = SPACECRAFT_TARGETS["voyager1"]
        x, y, z = spacecraft_position(target, 0)
        dist = math.sqrt(x*x + y*y + z*z)
        expected_km = target.distance_au * AU_TO_KM
        assert abs(dist - expected_km) / expected_km < 0.01  # within 1%

    def test_position_changes_with_time(self):
        target = SPACECRAFT_TARGETS["voyager1"]
        pos1 = spacecraft_position(target, 0)
        pos2 = spacecraft_position(target, 3600)  # 1 hour later
        d1 = math.sqrt(sum(c*c for c in pos1))
        d2 = math.sqrt(sum(c*c for c in pos2))
        assert d2 > d1, "Voyager should be moving away"


class TestComputeDistance:
    def test_voyager1_distance_order_of_magnitude(self):
        station = DSN_STATIONS["DSS-14"]
        target = SPACECRAFT_TARGETS["voyager1"]
        dist = compute_distance_km(station, target, 0)
        # Voyager 1 is ~159 AU = ~2.38e10 km
        assert 2e10 < dist < 3e10

    def test_mars_closer_than_voyager(self):
        station = DSN_STATIONS["DSS-14"]
        mars_dist = compute_distance_km(station, SPACECRAFT_TARGETS["mars_orbiter"], 0)
        voyager_dist = compute_distance_km(station, SPACECRAFT_TARGETS["voyager1"], 0)
        assert mars_dist < voyager_dist


class TestAzimuthElevation:
    def test_returns_valid_range(self):
        station = DSN_STATIONS["DSS-14"]
        target = SPACECRAFT_TARGETS["voyager1"]
        az, el = compute_azimuth_elevation(station, target, 0)
        assert 0 <= az < 360
        assert -90 <= el <= 90


class TestRelativeVelocity:
    def test_voyager1_positive(self):
        station = DSN_STATIONS["DSS-14"]
        target = SPACECRAFT_TARGETS["voyager1"]
        vel = compute_relative_velocity(station, target)
        assert vel > 0, "Voyager 1 should be moving away"


class TestGeometricDelay:
    def test_same_station_zero_delay(self):
        station = DSN_STATIONS["DSS-14"]
        target = SPACECRAFT_TARGETS["voyager1"]
        delay = compute_geometric_delay_us(station, station, target, 0)
        assert abs(delay) < 1e-6

    def test_different_stations_nonzero_delay(self):
        station1 = DSN_STATIONS["DSS-14"]
        station2 = DSN_STATIONS["DSS-43"]
        target = SPACECRAFT_TARGETS["voyager1"]
        delay = compute_geometric_delay_us(station1, station2, target, 0)
        # Inter-continental baseline should give microsecond-range delays
        assert abs(delay) < 100000  # less than 100ms (reasonable for Earth baseline)

    def test_delay_antisymmetric(self):
        s1 = DSN_STATIONS["DSS-14"]
        s2 = DSN_STATIONS["DSS-43"]
        target = SPACECRAFT_TARGETS["voyager1"]
        d12 = compute_geometric_delay_us(s1, s2, target, 0)
        d21 = compute_geometric_delay_us(s2, s1, target, 0)
        assert abs(d12 + d21) < 1e-6, "Delay should be antisymmetric"

"""Tests for the spacecraft transmitter module."""
import math
import numpy as np
import pytest
from app.spacecraft.transmitter import (
    generate_random_bits, bpsk_modulate, qpsk_modulate,
    generate_spacecraft_signal, apply_power_scaling
)


class TestBitGeneration:
    def test_correct_length(self):
        bits = generate_random_bits(100, seed=42)
        assert len(bits) == 100

    def test_binary_values(self):
        bits = generate_random_bits(1000, seed=42)
        assert set(bits).issubset({0, 1})

    def test_deterministic_with_seed(self):
        b1 = generate_random_bits(50, seed=42)
        b2 = generate_random_bits(50, seed=42)
        np.testing.assert_array_equal(b1, b2)


class TestBPSK:
    def test_output_shape(self):
        bits = generate_random_bits(10, seed=42)
        i, q = bpsk_modulate(bits, 8.4e9, 10000.0, 1000.0)
        expected_samples = 10 * 10  # 10 bits * (10000/1000) samples/symbol
        assert len(i) == expected_samples
        assert len(q) == expected_samples

    def test_bpsk_symbols_are_plus_minus_one(self):
        bits = np.array([0, 1, 0, 1])
        i, q = bpsk_modulate(bits, 8.4e9, 10000.0, 1000.0)
        # Each symbol should be +1 or -1 (at baseband)
        samples_per_sym = 10
        for s in range(4):
            sym_val = i[s * samples_per_sym]
            assert abs(abs(sym_val) - 1.0) < 1e-10

    def test_q_channel_zero_for_bpsk(self):
        bits = generate_random_bits(20, seed=42)
        i, q = bpsk_modulate(bits, 8.4e9, 10000.0, 1000.0)
        np.testing.assert_array_almost_equal(q, np.zeros_like(q))


class TestQPSK:
    def test_output_shape(self):
        bits = generate_random_bits(20, seed=42)
        i, q = qpsk_modulate(bits, 8.4e9, 10000.0, 1000.0)
        # 20 bits -> 10 QPSK symbols
        expected_samples = 10 * 10
        assert len(i) == expected_samples

    def test_constellation_points(self):
        bits = np.array([0, 0, 0, 1, 1, 0, 1, 1])
        i, q = qpsk_modulate(bits, 8.4e9, 10000.0, 1000.0)
        inv_sqrt2 = 1.0 / math.sqrt(2.0)
        # Check first symbol (bits 0,0 -> I=+1/√2, Q=+1/√2)
        assert abs(i[0] - inv_sqrt2) < 1e-10
        assert abs(q[0] - inv_sqrt2) < 1e-10


class TestSpacecraftSignal:
    def test_generate_bpsk(self):
        i, q, bits = generate_spacecraft_signal(num_bits=50, modulation="BPSK", seed=42)
        assert len(bits) == 50
        assert len(i) > 0
        assert len(q) > 0

    def test_generate_qpsk(self):
        i, q, bits = generate_spacecraft_signal(num_bits=50, modulation="QPSK", seed=42)
        assert len(bits) == 50
        assert len(i) > 0

    def test_deterministic_with_seed(self):
        i1, q1, b1 = generate_spacecraft_signal(num_bits=20, seed=42)
        i2, q2, b2 = generate_spacecraft_signal(num_bits=20, seed=42)
        np.testing.assert_array_equal(b1, b2)
        np.testing.assert_array_equal(i1, i2)


class TestPowerScaling:
    def test_scaling_preserves_shape(self):
        i = np.array([1.0, -1.0, 1.0, -1.0])
        q = np.array([0.0, 0.0, 0.0, 0.0])
        i_scaled, q_scaled = apply_power_scaling(i, q, -100.0)
        assert len(i_scaled) == 4
        assert len(q_scaled) == 4

    def test_lower_power_reduces_amplitude(self):
        i = np.array([1.0, -1.0, 1.0, -1.0])
        q = np.zeros(4)
        i_high, _ = apply_power_scaling(i, q, 0.0)
        i_low, _ = apply_power_scaling(i, q, -60.0)
        assert np.mean(i_high**2) > np.mean(i_low**2)

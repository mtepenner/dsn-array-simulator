package dsp

import (
	"math"
	"testing"
)

func TestStationECEF(t *testing.T) {
	// Goldstone approximate coords
	x, y, z := StationECEF(35.4, -116.9, 1000.0)
	r := math.Sqrt(x*x + y*y + z*z)
	// Should be approximately Earth radius + 1km
	if math.Abs(r-6372.0) > 5.0 {
		t.Errorf("Radius = %f, expected ~6372 km", r)
	}
}

func TestStationECEFEquator(t *testing.T) {
	// Equator, prime meridian
	x, _, z := StationECEF(0, 0, 0)
	if math.Abs(z) > 0.01 {
		t.Errorf("Z should be ~0 at equator, got %f", z)
	}
	if x < 6370 {
		t.Errorf("X should be ~Earth radius at equator/prime meridian, got %f", x)
	}
}

func TestGeometricDelayUSSymmetry(t *testing.T) {
	s1 := StationPos{Latitude: 35.4, Longitude: -116.9, Elevation: 1000}
	s2 := StationPos{Latitude: -35.4, Longitude: 148.9, Elevation: 600}

	d12 := GeometricDelayUS(s1, s2, 180.0, 30.0)
	d21 := GeometricDelayUS(s2, s1, 180.0, 30.0)

	if math.Abs(d12+d21) > 1e-6 {
		t.Errorf("Delay should be antisymmetric: d12=%f, d21=%f", d12, d21)
	}
}

func TestGeometricDelaySameStation(t *testing.T) {
	s := StationPos{Latitude: 35.4, Longitude: -116.9, Elevation: 1000}
	delay := GeometricDelayUS(s, s, 0, 0)
	if math.Abs(delay) > 1e-10 {
		t.Errorf("Same station delay should be 0, got %f", delay)
	}
}

func TestSampleDelay(t *testing.T) {
	d := SampleDelay(1.0, 1e6) // 1 µs at 1 MHz sample rate
	if math.Abs(d-1.0) > 1e-10 {
		t.Errorf("SampleDelay = %f, expected 1.0", d)
	}
}

func TestFractionalDelayZero(t *testing.T) {
	signal := []float64{1, 2, 3, 4, 5, 6, 7, 8}
	result := FractionalDelay(signal, 0.0)
	if len(result) != len(signal) {
		t.Fatalf("Length mismatch: %d vs %d", len(result), len(signal))
	}
	// With zero delay, output should approximate input (within windowing effects)
	for i := 2; i < len(signal)-2; i++ { // skip edges
		if math.Abs(result[i]-signal[i]) > 0.5 {
			t.Errorf("At %d: result=%f, expected ~%f", i, result[i], signal[i])
		}
	}
}

func TestFractionalDelayPreservesLength(t *testing.T) {
	signal := make([]float64, 100)
	for i := range signal {
		signal[i] = math.Sin(float64(i) * 0.1)
	}
	result := FractionalDelay(signal, 2.5)
	if len(result) != 100 {
		t.Errorf("Output length = %d, expected 100", len(result))
	}
}

func TestCrossCorrelationDelaySameSignal(t *testing.T) {
	signal := make([]float64, 200)
	for i := range signal {
		signal[i] = math.Sin(float64(i) * 0.3)
	}
	lag := CrossCorrelationDelay(signal, signal, 10)
	if math.Abs(lag) > 0.5 {
		t.Errorf("Same signal lag = %f, expected 0", lag)
	}
}

func TestCrossCorrelationDelayKnownShift(t *testing.T) {
	n := 300
	signal1 := make([]float64, n)
	signal2 := make([]float64, n)
	shift := 5
	for i := 0; i < n; i++ {
		signal1[i] = math.Sin(float64(i) * 0.2)
		j := i + shift
		if j < n {
			signal2[j] = signal1[i]
		}
	}
	lag := CrossCorrelationDelay(signal1, signal2, 20)
	if math.Abs(lag-float64(shift)) > 1.5 {
		t.Errorf("Detected lag = %f, expected ~%d", lag, shift)
	}
}

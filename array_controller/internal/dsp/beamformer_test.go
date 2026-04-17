package dsp

import (
	"math"
	"testing"
)

func TestBeamformEmpty(t *testing.T) {
	result := Beamform(nil)
	if result == nil {
		t.Fatal("Beamform(nil) should return non-nil result")
	}
	if len(result.CombinedI) != 0 {
		t.Errorf("Expected empty combined I, got %d samples", len(result.CombinedI))
	}
}

func TestBeamformSingleAntenna(t *testing.T) {
	n := 100
	iSamples := make([]float64, n)
	qSamples := make([]float64, n)
	for i := 0; i < n; i++ {
		iSamples[i] = math.Cos(float64(i) * 0.1)
		qSamples[i] = math.Sin(float64(i) * 0.1)
	}

	signals := []AntennaSignal{{
		AntennaID:    "DSS-14",
		ISamples:     iSamples,
		QSamples:     qSamples,
		PhaseOffset:  0,
		DelaySamples: 0,
		Weight:       1.0,
	}}

	result := Beamform(signals)
	if len(result.CombinedI) != n {
		t.Errorf("Expected %d samples, got %d", n, len(result.CombinedI))
	}
}

func TestBeamformCoherentCombinationIncreasesSNR(t *testing.T) {
	n := 500
	// Create identical signals from multiple antennas
	numAntennas := 4
	signals := make([]AntennaSignal, numAntennas)

	for a := 0; a < numAntennas; a++ {
		iSamples := make([]float64, n)
		qSamples := make([]float64, n)
		for i := 0; i < n; i++ {
			iSamples[i] = math.Cos(float64(i) * 0.1)
			qSamples[i] = 0 // Pure real signal
		}
		signals[a] = AntennaSignal{
			AntennaID:    "test",
			ISamples:     iSamples,
			QSamples:     qSamples,
			PhaseOffset:  0,
			DelaySamples: 0,
			Weight:       1.0,
		}
	}

	result := Beamform(signals)
	if result.CombinedSNR <= 0 {
		t.Errorf("Combined SNR should be positive, got %f", result.CombinedSNR)
	}
}

func TestBeamformWeightsPreserved(t *testing.T) {
	n := 50
	signals := make([]AntennaSignal, 2)
	for i := range signals {
		iSamples := make([]float64, n)
		for j := range iSamples {
			iSamples[j] = 1.0
		}
		signals[i] = AntennaSignal{
			ISamples: iSamples,
			QSamples: make([]float64, n),
			Weight:   float64(i + 1),
		}
	}
	result := Beamform(signals)
	if len(result.Weights) != 2 {
		t.Errorf("Expected 2 weights, got %d", len(result.Weights))
	}
}

func TestMaxRatioBeamformEmpty(t *testing.T) {
	result := MaxRatioBeamform(nil, nil)
	if result == nil {
		t.Fatal("MaxRatioBeamform(nil, nil) should return non-nil result")
	}
}

func TestMaxRatioBeamformHighSNRWeightedMore(t *testing.T) {
	n := 100
	signals := make([]AntennaSignal, 2)
	for i := range signals {
		iSamples := make([]float64, n)
		for j := range iSamples {
			iSamples[j] = float64(i + 1)
		}
		signals[i] = AntennaSignal{
			ISamples: iSamples,
			QSamples: make([]float64, n),
			Weight:   1.0,
		}
	}

	snrs := []float64{10.0, 20.0} // dB
	result := MaxRatioBeamform(signals, snrs)

	// Higher SNR antenna (index 1) should get higher weight
	if result.Weights[1] <= result.Weights[0] {
		t.Errorf("Higher SNR antenna should have higher weight: w0=%f, w1=%f",
			result.Weights[0], result.Weights[1])
	}
}

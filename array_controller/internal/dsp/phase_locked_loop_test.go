package dsp

import (
	"math"
	"testing"
)

func TestNewPLL(t *testing.T) {
	pll := NewPLL(1000.0, 44100.0, 50.0, 0.707)
	if pll == nil {
		t.Fatal("NewPLL returned nil")
	}
	if pll.CenterFreqHz != 1000.0 {
		t.Errorf("CenterFreqHz = %f, want 1000", pll.CenterFreqHz)
	}
	if pll.SampleRateHz != 44100.0 {
		t.Errorf("SampleRateHz = %f, want 44100", pll.SampleRateHz)
	}
}

func TestPLLLocksOnPureTone(t *testing.T) {
	freq := 100.0
	sampleRate := 10000.0
	pll := NewPLL(freq, sampleRate, 30.0, 0.707)

	// Generate a pure tone at the center frequency
	numSamples := 2000
	iSamples := make([]float64, numSamples)
	qSamples := make([]float64, numSamples)
	for i := 0; i < numSamples; i++ {
		phase := 2.0 * math.Pi * freq * float64(i) / sampleRate
		iSamples[i] = math.Cos(phase)
		qSamples[i] = math.Sin(phase)
	}

	pll.ProcessBlock(iSamples, qSamples)

	if !pll.IsLocked() {
		t.Error("PLL should lock on a pure tone after 2000 samples")
	}
}

func TestPLLPhaseErrorDecreasesOverTime(t *testing.T) {
	freq := 200.0
	sampleRate := 10000.0
	pll := NewPLL(freq, sampleRate, 20.0, 0.707)

	// Process samples and track phase error
	var earlyError, lateError float64
	for i := 0; i < 1000; i++ {
		phase := 2.0 * math.Pi * freq * float64(i) / sampleRate
		iIn := math.Cos(phase)
		qIn := math.Sin(phase)
		pll.ProcessSample(iIn, qIn)

		if i == 50 {
			earlyError = math.Abs(pll.PhaseError())
		}
		if i == 999 {
			lateError = math.Abs(pll.PhaseError())
		}
	}

	if lateError >= earlyError && earlyError > 0.01 {
		t.Errorf("Phase error should decrease: early=%f, late=%f", earlyError, lateError)
	}
}

func TestPLLProcessBlockMatchesSampleBySample(t *testing.T) {
	freq := 150.0
	sampleRate := 10000.0
	n := 100

	iSamples := make([]float64, n)
	qSamples := make([]float64, n)
	for i := 0; i < n; i++ {
		phase := 2.0 * math.Pi * freq * float64(i) / sampleRate
		iSamples[i] = math.Cos(phase)
		qSamples[i] = math.Sin(phase)
	}

	// ProcessBlock
	pll1 := NewPLL(freq, sampleRate, 20.0, 0.707)
	iOut1, qOut1, _ := pll1.ProcessBlock(iSamples, qSamples)

	// Sample-by-sample
	pll2 := NewPLL(freq, sampleRate, 20.0, 0.707)
	iOut2 := make([]float64, n)
	qOut2 := make([]float64, n)
	for i := 0; i < n; i++ {
		iOut2[i], qOut2[i], _ = pll2.ProcessSample(iSamples[i], qSamples[i])
	}

	for i := 0; i < n; i++ {
		if math.Abs(iOut1[i]-iOut2[i]) > 1e-12 {
			t.Errorf("I mismatch at %d: block=%f, sample=%f", i, iOut1[i], iOut2[i])
		}
		if math.Abs(qOut1[i]-qOut2[i]) > 1e-12 {
			t.Errorf("Q mismatch at %d: block=%f, sample=%f", i, qOut1[i], qOut2[i])
		}
	}
}

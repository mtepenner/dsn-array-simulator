package decoder

import (
	"math"
	"testing"
)

func TestBPSKDemodulateKnownBits(t *testing.T) {
	// Create a clean BPSK signal: bit 0 -> +1, bit 1 -> -1
	bits := []byte{0, 1, 0, 1, 1, 0}
	samplesPerSymbol := 10

	iSamples := make([]float64, len(bits)*samplesPerSymbol)
	qSamples := make([]float64, len(bits)*samplesPerSymbol)

	for i, b := range bits {
		val := 1.0
		if b == 1 {
			val = -1.0
		}
		for j := 0; j < samplesPerSymbol; j++ {
			iSamples[i*samplesPerSymbol+j] = val
		}
	}

	result := BPSKDemodulate(iSamples, qSamples, samplesPerSymbol)
	if len(result.Bits) != len(bits) {
		t.Fatalf("Expected %d bits, got %d", len(bits), len(result.Bits))
	}

	for i, expected := range bits {
		if result.Bits[i] != expected {
			t.Errorf("Bit %d: got %d, expected %d", i, result.Bits[i], expected)
		}
	}
}

func TestBPSKDemodulateDefaultSamplesPerSymbol(t *testing.T) {
	n := 100 // 10 symbols with default 10 samples/symbol
	iSamples := make([]float64, n)
	qSamples := make([]float64, n)
	for i := range iSamples {
		iSamples[i] = 1.0
	}

	result := BPSKDemodulate(iSamples, qSamples, 0)
	if len(result.Bits) != 10 {
		t.Errorf("Expected 10 bits with default samples/symbol, got %d", len(result.Bits))
	}
}

func TestQPSKDemodulateKnownSymbols(t *testing.T) {
	// Create QPSK signal for known bits
	// I=+1, Q=+1 -> bits 0,0
	// I=-1, Q=+1 -> bits 1,0
	samplesPerSymbol := 10
	symbols := [][2]float64{
		{1.0, 1.0},   // bits 0,0
		{-1.0, 1.0},  // bits 1,0
		{-1.0, -1.0}, // bits 1,1
		{1.0, -1.0},  // bits 0,1
	}

	n := len(symbols) * samplesPerSymbol
	iSamples := make([]float64, n)
	qSamples := make([]float64, n)

	for s, sym := range symbols {
		for j := 0; j < samplesPerSymbol; j++ {
			iSamples[s*samplesPerSymbol+j] = sym[0]
			qSamples[s*samplesPerSymbol+j] = sym[1]
		}
	}

	result := QPSKDemodulate(iSamples, qSamples, samplesPerSymbol)
	expectedBits := []byte{0, 0, 1, 0, 1, 1, 0, 1}

	if len(result.Bits) != len(expectedBits) {
		t.Fatalf("Expected %d bits, got %d", len(expectedBits), len(result.Bits))
	}

	for i, expected := range expectedBits {
		if result.Bits[i] != expected {
			t.Errorf("Bit %d: got %d, expected %d", i, result.Bits[i], expected)
		}
	}
}

func TestCalculateBERPerfect(t *testing.T) {
	tx := []byte{0, 1, 0, 1, 1}
	rx := []byte{0, 1, 0, 1, 1}
	ber := CalculateBER(tx, rx)
	if ber != 0.0 {
		t.Errorf("BER = %f, expected 0.0", ber)
	}
}

func TestCalculateBERAllErrors(t *testing.T) {
	tx := []byte{0, 0, 0, 0}
	rx := []byte{1, 1, 1, 1}
	ber := CalculateBER(tx, rx)
	if ber != 1.0 {
		t.Errorf("BER = %f, expected 1.0", ber)
	}
}

func TestCalculateBERHalfErrors(t *testing.T) {
	tx := []byte{0, 0, 1, 1}
	rx := []byte{0, 1, 1, 0}
	ber := CalculateBER(tx, rx)
	if math.Abs(ber-0.5) > 1e-10 {
		t.Errorf("BER = %f, expected 0.5", ber)
	}
}

func TestCalculateBEREmpty(t *testing.T) {
	ber := CalculateBER(nil, nil)
	if ber != 1.0 {
		t.Errorf("BER for empty = %f, expected 1.0", ber)
	}
}

func TestEstimateBERFromSNRHighSNR(t *testing.T) {
	// At high SNR (e.g., 20 dB), BER should be very low
	ber := EstimateBERFromSNR(20.0)
	if ber > 1e-6 {
		t.Errorf("BER at 20 dB SNR = %e, expected < 1e-6", ber)
	}
}

func TestEstimateBERFromSNRLowSNR(t *testing.T) {
	// At 0 dB SNR, BER should be significant
	ber := EstimateBERFromSNR(0.0)
	if ber < 0.01 || ber > 0.5 {
		t.Errorf("BER at 0 dB SNR = %f, expected between 0.01 and 0.5", ber)
	}
}

func TestEstimateBERFromSNRMonotonic(t *testing.T) {
	// BER should decrease as SNR increases
	ber1 := EstimateBERFromSNR(0.0)
	ber2 := EstimateBERFromSNR(5.0)
	ber3 := EstimateBERFromSNR(10.0)
	if ber2 >= ber1 {
		t.Errorf("BER not decreasing: %f >= %f", ber2, ber1)
	}
	if ber3 >= ber2 {
		t.Errorf("BER not decreasing: %f >= %f", ber3, ber2)
	}
}

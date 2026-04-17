package decoder

import (
	"math"
)

// DemodulationResult holds the output of demodulating a signal.
type DemodulationResult struct {
	Bits         []byte
	BitErrorRate float64
	SymbolsI     []float64
	SymbolsQ     []float64
	Confidence   float64
}

// BPSKDemodulate demodulates a BPSK signal from IQ samples.
// BPSK: I > 0 -> bit 0, I < 0 -> bit 1
func BPSKDemodulate(iSamples, qSamples []float64, samplesPerSymbol int) *DemodulationResult {
	if samplesPerSymbol <= 0 {
		samplesPerSymbol = 10
	}

	numSymbols := len(iSamples) / samplesPerSymbol
	bits := make([]byte, numSymbols)
	symbolsI := make([]float64, numSymbols)
	symbolsQ := make([]float64, numSymbols)

	var totalConfidence float64

	for s := 0; s < numSymbols; s++ {
		start := s * samplesPerSymbol
		end := start + samplesPerSymbol
		if end > len(iSamples) {
			end = len(iSamples)
		}

		// Integrate over the symbol period
		var sumI, sumQ float64
		for i := start; i < end; i++ {
			sumI += iSamples[i]
			sumQ += qSamples[i]
		}
		avgI := sumI / float64(end-start)
		avgQ := sumQ / float64(end-start)

		symbolsI[s] = avgI
		symbolsQ[s] = avgQ

		// Decision: positive I = bit 0, negative I = bit 1
		if avgI >= 0 {
			bits[s] = 0
		} else {
			bits[s] = 1
		}

		// Confidence based on distance from decision boundary
		totalConfidence += math.Abs(avgI)
	}

	avgConfidence := 0.0
	if numSymbols > 0 {
		avgConfidence = totalConfidence / float64(numSymbols)
	}

	return &DemodulationResult{
		Bits:       bits,
		SymbolsI:   symbolsI,
		SymbolsQ:   symbolsQ,
		Confidence: avgConfidence,
	}
}

// QPSKDemodulate demodulates a QPSK signal from IQ samples.
func QPSKDemodulate(iSamples, qSamples []float64, samplesPerSymbol int) *DemodulationResult {
	if samplesPerSymbol <= 0 {
		samplesPerSymbol = 10
	}

	numSymbols := len(iSamples) / samplesPerSymbol
	bits := make([]byte, numSymbols*2)
	symbolsI := make([]float64, numSymbols)
	symbolsQ := make([]float64, numSymbols)

	var totalConfidence float64

	for s := 0; s < numSymbols; s++ {
		start := s * samplesPerSymbol
		end := start + samplesPerSymbol
		if end > len(iSamples) {
			end = len(iSamples)
		}

		var sumI, sumQ float64
		for i := start; i < end; i++ {
			sumI += iSamples[i]
			sumQ += qSamples[i]
		}
		avgI := sumI / float64(end-start)
		avgQ := sumQ / float64(end-start)

		symbolsI[s] = avgI
		symbolsQ[s] = avgQ

		// QPSK decision
		if avgI >= 0 {
			bits[2*s] = 0
		} else {
			bits[2*s] = 1
		}
		if avgQ >= 0 {
			bits[2*s+1] = 0
		} else {
			bits[2*s+1] = 1
		}

		totalConfidence += math.Sqrt(avgI*avgI + avgQ*avgQ)
	}

	avgConfidence := 0.0
	if numSymbols > 0 {
		avgConfidence = totalConfidence / float64(numSymbols)
	}

	return &DemodulationResult{
		Bits:       bits,
		SymbolsI:   symbolsI,
		SymbolsQ:   symbolsQ,
		Confidence: avgConfidence,
	}
}

// CalculateBER computes the Bit Error Rate between transmitted and received bits.
func CalculateBER(transmitted, received []byte) float64 {
	n := len(transmitted)
	if len(received) < n {
		n = len(received)
	}
	if n == 0 {
		return 1.0
	}

	errors := 0
	for i := 0; i < n; i++ {
		if transmitted[i] != received[i] {
			errors++
		}
	}

	return float64(errors) / float64(n)
}

// EstimateBERFromSNR estimates the theoretical BER for BPSK given SNR in dB.
// BER_BPSK = 0.5 * erfc(sqrt(SNR_linear))
func EstimateBERFromSNR(snrDB float64) float64 {
	snrLinear := math.Pow(10.0, snrDB/10.0)
	return 0.5 * math.Erfc(math.Sqrt(snrLinear))
}

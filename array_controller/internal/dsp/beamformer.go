package dsp

import (
	"math"
)

// BeamformerConfig holds configuration for the beamformer.
type BeamformerConfig struct {
	NumAntennas   int
	SampleRate    float64
	CarrierFreqHz float64
}

// AntennaSignal holds a signal from one antenna with its correction parameters.
type AntennaSignal struct {
	AntennaID    string
	ISamples     []float64
	QSamples     []float64
	PhaseOffset  float64 // radians - phase correction to apply
	DelaySamples float64 // fractional sample delay correction
	Weight       float64 // antenna weighting factor
}

// BeamformResult holds the output of beamforming.
type BeamformResult struct {
	CombinedI   []float64
	CombinedQ   []float64
	CombinedSNR float64
	Weights     []float64
}

// Beamform performs phase-aligned signal combination (beamforming).
// This is the core array processing: phase-shift, delay-correct, and sum all antenna signals.
func Beamform(signals []AntennaSignal) *BeamformResult {
	if len(signals) == 0 {
		return &BeamformResult{}
	}

	// Find the minimum signal length
	minLen := len(signals[0].ISamples)
	for _, sig := range signals[1:] {
		if len(sig.ISamples) < minLen {
			minLen = len(sig.ISamples)
		}
	}

	combinedI := make([]float64, minLen)
	combinedQ := make([]float64, minLen)
	weights := make([]float64, len(signals))

	totalWeight := 0.0
	for i, sig := range signals {
		w := sig.Weight
		if w <= 0 {
			w = 1.0
		}
		weights[i] = w
		totalWeight += w
	}

	for si, sig := range signals {
		// Apply delay correction
		iDelayed := FractionalDelay(sig.ISamples[:minLen], sig.DelaySamples)
		qDelayed := FractionalDelay(sig.QSamples[:minLen], sig.DelaySamples)

		// Apply phase correction
		cosP := math.Cos(-sig.PhaseOffset) // negative to correct the offset
		sinP := math.Sin(-sig.PhaseOffset)

		w := weights[si] / totalWeight

		for i := 0; i < minLen; i++ {
			iCorrected := iDelayed[i]*cosP - qDelayed[i]*sinP
			qCorrected := iDelayed[i]*sinP + qDelayed[i]*cosP

			combinedI[i] += iCorrected * w
			combinedQ[i] += qCorrected * w
		}
	}

	// Estimate combined SNR
	// For N coherently combined antennas, SNR improves by ~N
	combinedSNR := estimateSNR(combinedI, combinedQ)

	return &BeamformResult{
		CombinedI:   combinedI,
		CombinedQ:   combinedQ,
		CombinedSNR: combinedSNR,
		Weights:     weights,
	}
}

// estimateSNR estimates the Signal-to-Noise Ratio of an IQ signal.
func estimateSNR(iSamples, qSamples []float64) float64 {
	n := len(iSamples)
	if n == 0 {
		return 0
	}

	// Calculate signal power (mean of magnitude squared)
	var signalPower float64
	for i := 0; i < n; i++ {
		mag := math.Sqrt(iSamples[i]*iSamples[i] + qSamples[i]*qSamples[i])
		signalPower += mag * mag
	}
	signalPower /= float64(n)

	// Estimate noise as variance of the magnitude
	var meanMag float64
	for i := 0; i < n; i++ {
		meanMag += math.Sqrt(iSamples[i]*iSamples[i] + qSamples[i]*qSamples[i])
	}
	meanMag /= float64(n)

	var variance float64
	for i := 0; i < n; i++ {
		mag := math.Sqrt(iSamples[i]*iSamples[i] + qSamples[i]*qSamples[i])
		diff := mag - meanMag
		variance += diff * diff
	}
	variance /= float64(n)

	if variance < 1e-20 {
		return 60.0 // essentially noise-free
	}

	snr := (meanMag * meanMag) / variance
	return 10.0 * math.Log10(snr)
}

// MaxRatioBeamform performs Maximum Ratio Combining (MRC) beamforming.
// Each antenna is weighted proportionally to its SNR.
func MaxRatioBeamform(signals []AntennaSignal, snrs []float64) *BeamformResult {
	if len(signals) == 0 {
		return &BeamformResult{}
	}

	// Set weights proportional to SNR (linear scale)
	for i := range signals {
		if i < len(snrs) {
			snrLinear := math.Pow(10.0, snrs[i]/10.0)
			signals[i].Weight = snrLinear
		} else {
			signals[i].Weight = 1.0
		}
	}

	return Beamform(signals)
}

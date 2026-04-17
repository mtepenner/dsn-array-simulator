package dsp

import (
	"math"
	"math/cmplx"
)

// PLL implements a Phase-Locked Loop for carrier tracking.
type PLL struct {
	// Loop parameters
	CenterFreqHz  float64
	SampleRateHz  float64
	LoopBandwidth float64
	DampingFactor float64

	// Internal state
	phase      float64
	frequency  float64
	integ      float64
	phaseError float64
	locked     bool
	lockCount  int

	// Proportional and integral gains
	kp float64
	ki float64
}

// NewPLL creates a Phase-Locked Loop with the given parameters.
func NewPLL(centerFreqHz, sampleRateHz, loopBandwidthHz, dampingFactor float64) *PLL {
	// Calculate loop filter gains using Gardner's formula
	theta_n := loopBandwidthHz / sampleRateHz * 2.0 * math.Pi
	denom := 1.0 + 2.0*dampingFactor*theta_n + theta_n*theta_n
	kp := 4.0 * dampingFactor * theta_n / denom
	ki := 4.0 * theta_n * theta_n / denom

	return &PLL{
		CenterFreqHz:  centerFreqHz,
		SampleRateHz:  sampleRateHz,
		LoopBandwidth: loopBandwidthHz,
		DampingFactor: dampingFactor,
		phase:         0,
		frequency:     centerFreqHz / sampleRateHz * 2.0 * math.Pi,
		kp:            kp,
		ki:            ki,
		locked:        false,
	}
}

// ProcessSample runs one sample through the PLL.
// Returns the corrected IQ sample and the current phase estimate.
func (p *PLL) ProcessSample(iIn, qIn float64) (iOut, qOut, phaseEst float64) {
	// Generate local oscillator
	cosPhase := math.Cos(p.phase)
	sinPhase := math.Sin(p.phase)

	// Mix input with local oscillator (downconvert)
	iOut = iIn*cosPhase + qIn*sinPhase
	qOut = -iIn*sinPhase + qIn*cosPhase

	// Phase detector (atan2-based)
	p.phaseError = math.Atan2(qOut, iOut)

	// Loop filter (PI controller)
	p.integ += p.ki * p.phaseError
	freqAdj := p.kp*p.phaseError + p.integ

	// Update phase
	p.phase += p.frequency + freqAdj
	// Wrap phase to [-π, π]
	for p.phase > math.Pi {
		p.phase -= 2.0 * math.Pi
	}
	for p.phase < -math.Pi {
		p.phase += 2.0 * math.Pi
	}

	// Lock detection
	if math.Abs(p.phaseError) < 0.3 {
		p.lockCount++
		if p.lockCount > 100 {
			p.locked = true
		}
	} else {
		p.lockCount = 0
		if p.lockCount < -50 {
			p.locked = false
		}
	}

	phaseEst = p.phase
	return
}

// ProcessBlock runs a block of IQ samples through the PLL.
func (p *PLL) ProcessBlock(iIn, qIn []float64) (iOut, qOut []float64, phaseEstimates []float64) {
	n := len(iIn)
	if len(qIn) < n {
		n = len(qIn)
	}

	iOut = make([]float64, n)
	qOut = make([]float64, n)
	phaseEstimates = make([]float64, n)

	for i := 0; i < n; i++ {
		iOut[i], qOut[i], phaseEstimates[i] = p.ProcessSample(iIn[i], qIn[i])
	}
	return
}

// IsLocked returns whether the PLL has achieved lock.
func (p *PLL) IsLocked() bool {
	return p.locked
}

// PhaseError returns the current phase error.
func (p *PLL) PhaseError() float64 {
	return p.phaseError
}

// CurrentPhase returns the current NCO phase.
func (p *PLL) CurrentPhase() float64 {
	return p.phase
}

// unused import guard
var _ = cmplx.Abs

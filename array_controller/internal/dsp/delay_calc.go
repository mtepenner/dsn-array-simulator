package dsp

import (
	"math"
)

const (
	SpeedOfLightKmS = 299792.458
	EarthRadiusKM   = 6371.0
)

// StationPos represents a ground station position.
type StationPos struct {
	Latitude  float64 // degrees
	Longitude float64 // degrees
	Elevation float64 // meters
}

// StationECEF converts station geodetic coordinates to ECEF (km).
func StationECEF(lat, lon, elevM float64) (x, y, z float64) {
	latR := lat * math.Pi / 180.0
	lonR := lon * math.Pi / 180.0
	r := EarthRadiusKM + elevM/1000.0
	x = r * math.Cos(latR) * math.Cos(lonR)
	y = r * math.Cos(latR) * math.Sin(lonR)
	z = r * math.Sin(latR)
	return
}

// GeometricDelayUS calculates the geometric delay in microseconds between
// two stations receiving the same wavefront from a distant source.
func GeometricDelayUS(station1, station2 StationPos, sourceRA, sourceDec float64) float64 {
	// Station positions in ECEF
	x1, y1, z1 := StationECEF(station1.Latitude, station1.Longitude, station1.Elevation)
	x2, y2, z2 := StationECEF(station2.Latitude, station2.Longitude, station2.Elevation)

	// Baseline vector
	bx := x2 - x1
	by := y2 - y1
	bz := z2 - z1

	// Source direction unit vector
	raR := sourceRA * math.Pi / 180.0
	decR := sourceDec * math.Pi / 180.0
	sx := math.Cos(decR) * math.Cos(raR)
	sy := math.Cos(decR) * math.Sin(raR)
	sz := math.Sin(decR)

	// Geometric delay = dot(baseline, source_direction) / c
	dot := bx*sx + by*sy + bz*sz
	delayS := dot / SpeedOfLightKmS
	return delayS * 1e6 // convert to microseconds
}

// SampleDelay converts a delay in microseconds to a number of samples at the given sample rate.
func SampleDelay(delayUS float64, sampleRateHz float64) float64 {
	return delayUS * 1e-6 * sampleRateHz
}

// FractionalDelay applies a fractional sample delay using sinc interpolation.
func FractionalDelay(samples []float64, delaySamples float64) []float64 {
	n := len(samples)
	result := make([]float64, n)
	intDelay := int(math.Floor(delaySamples))
	fracDelay := delaySamples - float64(intDelay)

	// Sinc interpolation filter (length 8)
	filterLen := 8
	half := filterLen / 2

	for i := 0; i < n; i++ {
		var sum float64
		for j := -half; j < half; j++ {
			idx := i - intDelay + j
			if idx >= 0 && idx < n {
				sincArg := float64(j) - fracDelay
				var sincVal float64
				if math.Abs(sincArg) < 1e-10 {
					sincVal = 1.0
				} else {
					sincVal = math.Sin(math.Pi*sincArg) / (math.Pi * sincArg)
				}
				// Hanning window
				window := 0.5 * (1.0 + math.Cos(2.0*math.Pi*float64(j)/float64(filterLen)))
				sum += samples[idx] * sincVal * window
			}
		}
		result[i] = sum
	}
	return result
}

// CrossCorrelationDelay estimates the delay between two signals using cross-correlation.
func CrossCorrelationDelay(signal1, signal2 []float64, maxLagSamples int) float64 {
	n := len(signal1)
	if len(signal2) < n {
		n = len(signal2)
	}

	bestLag := 0
	bestCorr := -math.MaxFloat64

	for lag := -maxLagSamples; lag <= maxLagSamples; lag++ {
		var corr float64
		var count int
		for i := 0; i < n; i++ {
			j := i + lag
			if j >= 0 && j < n {
				corr += signal1[i] * signal2[j]
				count++
			}
		}
		if count > 0 {
			corr /= float64(count)
		}
		if corr > bestCorr {
			bestCorr = corr
			bestLag = lag
		}
	}

	return float64(bestLag)
}

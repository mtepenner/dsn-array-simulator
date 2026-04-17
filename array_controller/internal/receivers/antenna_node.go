package receivers

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// AntennaNode simulates an individual ground station receiving a noisy signal.
type AntennaNode struct {
	mu sync.RWMutex

	ID            string
	Name          string
	Latitude      float64
	Longitude     float64
	DishDiameterM float64
	ElevationM    float64

	// Current state
	ISamples       []float64
	QSamples       []float64
	SNRDB          float64
	DopplerShiftHz float64
	PathLossDB     float64
	AtmosAttenDB   float64
	DistanceKM     float64
	Timestamp      float64
	PhaseOffset    float64
	DelayUS        float64
	PLLLocked      bool
	AzimuthDeg     float64
	ElevationDeg   float64

	Active bool
}

// DSNStation defines a ground station configuration.
type DSNStation struct {
	ID            string
	Name          string
	Latitude      float64
	Longitude     float64
	ElevationM    float64
	DishDiameterM float64
}

// DefaultStations returns the standard DSN stations.
func DefaultStations() []DSNStation {
	return []DSNStation{
		{ID: "DSS-14", Name: "Goldstone 70m", Latitude: 35.4267, Longitude: -116.89, ElevationM: 1001.0, DishDiameterM: 70.0},
		{ID: "DSS-24", Name: "Goldstone 34m", Latitude: 35.3395, Longitude: -116.8756, ElevationM: 988.0, DishDiameterM: 34.0},
		{ID: "DSS-25", Name: "Goldstone 34m BWG", Latitude: 35.3382, Longitude: -116.8750, ElevationM: 987.0, DishDiameterM: 34.0},
		{ID: "DSS-34", Name: "Canberra 34m", Latitude: -35.3985, Longitude: 148.9816, ElevationM: 692.0, DishDiameterM: 34.0},
		{ID: "DSS-35", Name: "Canberra 34m BWG", Latitude: -35.3955, Longitude: 148.9786, ElevationM: 694.0, DishDiameterM: 34.0},
		{ID: "DSS-43", Name: "Canberra 70m", Latitude: -35.4023, Longitude: 148.9814, ElevationM: 689.0, DishDiameterM: 70.0},
		{ID: "DSS-54", Name: "Madrid 34m", Latitude: 40.4313, Longitude: -4.2480, ElevationM: 837.0, DishDiameterM: 34.0},
		{ID: "DSS-55", Name: "Madrid 34m BWG", Latitude: 40.4325, Longitude: -4.2510, ElevationM: 835.0, DishDiameterM: 34.0},
		{ID: "DSS-63", Name: "Madrid 70m", Latitude: 40.4314, Longitude: -4.2481, ElevationM: 833.0, DishDiameterM: 70.0},
	}
}

// NewAntennaNode creates a new antenna node from a station definition.
func NewAntennaNode(station DSNStation) *AntennaNode {
	return &AntennaNode{
		ID:            station.ID,
		Name:          station.Name,
		Latitude:      station.Latitude,
		Longitude:     station.Longitude,
		DishDiameterM: station.DishDiameterM,
		ElevationM:    station.ElevationM,
		Active:        true,
		PLLLocked:     false,
	}
}

// UpdateSignal updates the antenna node with new signal data.
func (a *AntennaNode) UpdateSignal(iSamples, qSamples []float64, snr, doppler, pathLoss, atmosAtten, distance, timestamp float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ISamples = iSamples
	a.QSamples = qSamples
	a.SNRDB = snr
	a.DopplerShiftHz = doppler
	a.PathLossDB = pathLoss
	a.AtmosAttenDB = atmosAtten
	a.DistanceKM = distance
	a.Timestamp = timestamp
}

// GetSignal returns the current IQ samples (thread-safe).
func (a *AntennaNode) GetSignal() ([]float64, []float64) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	iCopy := make([]float64, len(a.ISamples))
	qCopy := make([]float64, len(a.QSamples))
	copy(iCopy, a.ISamples)
	copy(qCopy, a.QSamples)
	return iCopy, qCopy
}

// SimulateLocalNoise generates simulated noise to add to the received signal.
func SimulateLocalNoise(numSamples int, noisePower float64) ([]float64, []float64) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	iNoise := make([]float64, numSamples)
	qNoise := make([]float64, numSamples)
	std := math.Sqrt(noisePower)
	for i := 0; i < numSamples; i++ {
		iNoise[i] = rng.NormFloat64() * std
		qNoise[i] = rng.NormFloat64() * std
	}
	return iNoise, qNoise
}

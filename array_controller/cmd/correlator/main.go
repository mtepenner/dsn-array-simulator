package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"

	"dsn-array-simulator/array_controller/internal/decoder"
	"dsn-array-simulator/array_controller/internal/dsp"
	"dsn-array-simulator/array_controller/internal/receivers"
)

var ctx = context.Background()

// TelemetryData is the JSON structure sent via WebSocket to the dashboard.
type TelemetryData struct {
	Timestamp    float64            `json:"timestamp"`
	CombinedSNR  float64            `json:"combined_snr_db"`
	BitErrorRate float64            `json:"bit_error_rate"`
	Antennas     []AntennaTelemetry `json:"antennas"`
	CombinedI    []float64          `json:"combined_i_samples"`
	CombinedQ    []float64          `json:"combined_q_samples"`
}

// AntennaTelemetry holds per-antenna data.
type AntennaTelemetry struct {
	AntennaID    string  `json:"antenna_id"`
	Name         string  `json:"name"`
	SNRDB        float64 `json:"snr_db"`
	PhaseOffset  float64 `json:"phase_offset_rad"`
	DelayUS      float64 `json:"delay_us"`
	PLLLocked    bool    `json:"pll_locked"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	AzimuthDeg   float64 `json:"azimuth_deg"`
	ElevationDeg float64 `json:"elevation_deg"`
	Active       bool    `json:"active"`
	DistanceKM   float64 `json:"distance_km"`
	DopplerHz    float64 `json:"doppler_hz"`
	PathLossDB   float64 `json:"path_loss_db"`
}

// ArrayController manages the antenna array and signal processing pipeline.
type ArrayController struct {
	mu              sync.RWMutex
	antennas        map[string]*receivers.AntennaNode
	activeIDs       []string
	plls            map[string]*dsp.PLL
	latestTelemetry *TelemetryData
	rdb             *redis.Client
	sampleRate      float64
	carrierFreqHz   float64
}

func NewArrayController() *ArrayController {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	ac := &ArrayController{
		antennas:      make(map[string]*receivers.AntennaNode),
		plls:          make(map[string]*dsp.PLL),
		rdb:           rdb,
		sampleRate:    10000.0,
		carrierFreqHz: 8.4e9,
	}

	// Initialize default stations
	stations := receivers.DefaultStations()
	defaultActive := []string{"DSS-14", "DSS-43", "DSS-63"} // One per complex
	for _, s := range stations {
		node := receivers.NewAntennaNode(s)
		ac.antennas[s.ID] = node
		ac.plls[s.ID] = dsp.NewPLL(ac.carrierFreqHz, ac.sampleRate, 50.0, 0.707)
	}
	ac.activeIDs = defaultActive

	return ac
}

// ProcessSignals runs the main DSP pipeline: PLL -> delay calc -> beamform -> demodulate
func (ac *ArrayController) ProcessSignals() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	var beamSignals []dsp.AntennaSignal
	var antennaTelemetry []AntennaTelemetry
	var snrs []float64

	// Reference station for delay calculation
	var refStation *receivers.AntennaNode
	if len(ac.activeIDs) > 0 {
		refStation = ac.antennas[ac.activeIDs[0]]
	}

	for _, id := range ac.activeIDs {
		ant, ok := ac.antennas[id]
		if !ok || !ant.Active {
			continue
		}

		iSamples, qSamples := ant.GetSignal()
		if len(iSamples) == 0 {
			continue
		}

		// Run PLL to track carrier and correct phase
		pll := ac.plls[id]
		iCorrected, qCorrected, _ := pll.ProcessBlock(iSamples, qSamples)

		// Calculate geometric delay relative to reference station
		var delaySamples float64
		if refStation != nil && id != ac.activeIDs[0] {
			refI, _ := refStation.GetSignal()
			if len(refI) > 0 {
				lagSamples := dsp.CrossCorrelationDelay(refI, iSamples, 50)
				delaySamples = lagSamples
			}
		}

		beamSignals = append(beamSignals, dsp.AntennaSignal{
			AntennaID:    id,
			ISamples:     iCorrected,
			QSamples:     qCorrected,
			PhaseOffset:  pll.CurrentPhase(),
			DelaySamples: delaySamples,
			Weight:       1.0,
		})

		snrs = append(snrs, ant.SNRDB)

		antennaTelemetry = append(antennaTelemetry, AntennaTelemetry{
			AntennaID:    id,
			Name:         ant.Name,
			SNRDB:        ant.SNRDB,
			PhaseOffset:  pll.CurrentPhase(),
			DelayUS:      delaySamples / ac.sampleRate * 1e6,
			PLLLocked:    pll.IsLocked(),
			Latitude:     ant.Latitude,
			Longitude:    ant.Longitude,
			AzimuthDeg:   ant.AzimuthDeg,
			ElevationDeg: ant.ElevationDeg,
			Active:       ant.Active,
			DistanceKM:   ant.DistanceKM,
			DopplerHz:    ant.DopplerShiftHz,
			PathLossDB:   ant.PathLossDB,
		})
	}

	if len(beamSignals) == 0 {
		return
	}

	// Maximum Ratio Combining beamforming
	result := dsp.MaxRatioBeamform(beamSignals, snrs)

	// Demodulate the combined signal
	samplesPerSymbol := int(ac.sampleRate / 1000.0) // 1000 symbols/sec
	demodResult := decoder.BPSKDemodulate(result.CombinedI, result.CombinedQ, samplesPerSymbol)

	// Estimate BER from combined SNR
	ber := decoder.EstimateBERFromSNR(result.CombinedSNR)

	// Limit samples sent to dashboard (downsample for visualization)
	maxVisSamples := 500
	visI := result.CombinedI
	visQ := result.CombinedQ
	if len(visI) > maxVisSamples {
		visI = visI[:maxVisSamples]
		visQ = visQ[:maxVisSamples]
	}

	ac.latestTelemetry = &TelemetryData{
		Timestamp:    float64(time.Now().UnixMilli()) / 1000.0,
		CombinedSNR:  result.CombinedSNR,
		BitErrorRate: ber,
		Antennas:     antennaTelemetry,
		CombinedI:    visI,
		CombinedQ:    visQ,
	}

	// Also publish to Redis for persistence
	ac.publishToRedis(ac.latestTelemetry, demodResult)
}

func (ac *ArrayController) publishToRedis(telemetry *TelemetryData, demod *decoder.DemodulationResult) {
	data, err := json.Marshal(telemetry)
	if err != nil {
		log.Printf("Error marshaling telemetry: %v", err)
		return
	}
	ac.rdb.Publish(ctx, "telemetry", data)
	ac.rdb.Set(ctx, "latest_telemetry", data, 30*time.Second)
}

// SimulateSignals generates simulated signals for all active antennas.
func (ac *ArrayController) SimulateSignals() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	t := float64(time.Now().UnixMilli()) / 1000.0
	numSamples := 1000

	for _, id := range ac.activeIDs {
		ant, ok := ac.antennas[id]
		if !ok {
			continue
		}

		// Generate a simulated BPSK signal with noise
		baseSNR := 5.0 + float64(len(ac.activeIDs))*2.0 // more antennas = better

		// Add dish-size bonus
		if ant.DishDiameterM >= 70.0 {
			baseSNR += 6.0
		}

		// Simulate signal with noise
		iSamples := make([]float64, numSamples)
		qSamples := make([]float64, numSamples)
		noisePower := 1.0 / math.Pow(10.0, baseSNR/10.0)
		noiseStd := math.Sqrt(noisePower)

		// Generate BPSK symbols
		symbolLen := 10 // samples per symbol
		for i := 0; i < numSamples; i++ {
			symbolIdx := i / symbolLen
			// Deterministic bit pattern based on time
			bit := (symbolIdx + int(t*100)) % 2
			if bit == 0 {
				iSamples[i] = 1.0
			} else {
				iSamples[i] = -1.0
			}
			// Add phase offset for this antenna
			phase := float64(symbolIdx) * 0.1 * float64(len(id))
			cos := math.Cos(phase)
			sin := math.Sin(phase)
			origI := iSamples[i]
			iSamples[i] = origI * cos
			qSamples[i] = origI * sin

			// Add noise
			iSamples[i] += noiseStd * (2.0*float64((i*7+13)%100)/100.0 - 1.0)
			qSamples[i] += noiseStd * (2.0*float64((i*11+7)%100)/100.0 - 1.0)
		}

		ant.UpdateSignal(iSamples, qSamples, baseSNR, 1000.0, 250.0, 1.5, 2.38e10, t)
		ant.AzimuthDeg = 180.0 + 30.0*math.Sin(t*0.01)
		ant.ElevationDeg = 45.0 + 15.0*math.Cos(t*0.01)
		ant.PLLLocked = true
	}
}

// SetActiveAntennas updates which antennas are active.
func (ac *ArrayController) SetActiveAntennas(ids []string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	validIDs := make([]string, 0)
	for _, id := range ids {
		if _, ok := ac.antennas[id]; ok {
			ac.antennas[id].Active = true
			validIDs = append(validIDs, id)
		}
	}

	// Deactivate removed antennas
	for _, id := range ac.activeIDs {
		found := false
		for _, vid := range validIDs {
			if id == vid {
				found = true
				break
			}
		}
		if !found {
			if ant, ok := ac.antennas[id]; ok {
				ant.Active = false
			}
		}
	}

	ac.activeIDs = validIDs
	log.Printf("Active antennas updated: %v", validIDs)
}

// GetAllAntennas returns info about all antennas.
func (ac *ArrayController) GetAllAntennas() []AntennaTelemetry {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	result := make([]AntennaTelemetry, 0, len(ac.antennas))
	for _, ant := range ac.antennas {
		active := false
		for _, id := range ac.activeIDs {
			if id == ant.ID {
				active = true
				break
			}
		}
		result = append(result, AntennaTelemetry{
			AntennaID: ant.ID,
			Name:      ant.Name,
			Latitude:  ant.Latitude,
			Longitude: ant.Longitude,
			Active:    active,
		})
	}
	return result
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func main() {
	log.Println("Starting DSN Array Controller...")

	controller := NewArrayController()

	// Start the simulation + processing loop
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			controller.SimulateSignals()
			controller.ProcessSignals()
		}
	}()

	// REST API endpoints
	http.HandleFunc("/api/telemetry", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		controller.mu.RLock()
		defer controller.mu.RUnlock()
		if controller.latestTelemetry != nil {
			json.NewEncoder(w).Encode(controller.latestTelemetry)
		} else {
			w.Write([]byte(`{"message":"no data yet"}`))
		}
	})

	http.HandleFunc("/api/antennas", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" {
			antennas := controller.GetAllAntennas()
			json.NewEncoder(w).Encode(antennas)
			return
		}

		if r.Method == "POST" {
			var config struct {
				ActiveIDs []string `json:"active_antenna_ids"`
			}
			if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			controller.SetActiveAntennas(config.ActiveIDs)
			w.Write([]byte(`{"success":true}`))
			return
		}
	})

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(`{"status":"healthy","service":"array-controller"}`))
	})

	// WebSocket endpoint for real-time telemetry
	http.HandleFunc("/ws/telemetry", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		log.Println("WebSocket client connected")

		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			controller.mu.RLock()
			telemetry := controller.latestTelemetry
			controller.mu.RUnlock()

			if telemetry != nil {
				if err := conn.WriteJSON(telemetry); err != nil {
					log.Printf("WebSocket write error: %v", err)
					return
				}
			}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Array Controller HTTP/WS server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

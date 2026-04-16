# 📡 DSN Array Simulator

An advanced simulation and visualization suite for Deep Space Network (DSN) arraying. This project models the physics of deep space communication, performing real-time digital signal processing (DSP), signal correlation, and beamforming across multiple simulated ground stations to extract clean telemetry from noisy environments.

## 📑 Table of Contents
- [Features](#-features)
- [Architecture](#-architecture)
- [Technologies Used](#-technologies-used)
- [Installation](#-installation)
- [Usage](#-usage)
- [Contributing](#-contributing)
- [License](#-license)

## 🚀 Features
* **Accurate Physics Modeling:** Utilizes Skyfield to calculate exact vectors between Earth stations and spacecraft, applying Free-Space Path Loss (FSPL), Doppler shifts, and random atmospheric attenuation.
* **Waveform Generation:** Simulates source transmissions using standard deep space modulation schemes (e.g., BPSK or QPSK).
* **High-Performance Beamforming:** A Go-based array controller calculates microsecond delays between Earth stations, utilizing Phase-Locked Loops (PLL) to phase-shift, sum signals, and maximize the Signal-to-Noise Ratio (SNR).
* **Signal Demodulation:** Translates the phased and summed analog waveforms back into digital telemetry bits.
* **3D Mission Dashboard:** An interactive React operations center featuring a 3D Earth map with tracking antennas, dynamic signal waterfall spectrograms, and live Link Budget metrics (SNR and Bit Error Rate).
* **Scalable Infrastructure:** Uses Redis as a high-speed message bus between the physics engine and correlator, and TimescaleDB for storing historical metrics.

## 🏗️ Architecture
The system operates as a distributed, microservice-based architecture:
1. **Physics Environment (Python):** The deep space and atmosphere model that streams degraded RF signals via gRPC.
2. **Array Controller (Go):** The high-CPU, math-heavy DSP brain responsible for signal correlation, delay calculations, and demodulation.
3. **Mission Dashboard (React/TypeScript):** The frontend UI that consumes WebSocket telemetry to visualize the array's performance in real-time.
4. **Infrastructure:** Supported by Redis and TimescaleDB, deployable via Kubernetes or Docker Compose.

## 🛠️ Technologies Used
* **Simulation & Physics:** Python, SciPy, NumPy, Skyfield, gRPC
* **DSP & Control:** Go (Golang)
* **Frontend UI:** React, TypeScript, WebSockets, Three.js (3D map), Plotly.js (spectrograms)
* **Data & Infrastructure:** Redis, TimescaleDB, Docker, Kubernetes

## 💻 Installation

### Prerequisites
* Docker and Docker Compose installed.
* Ensure you have `make` installed for build scripts and protobuf generation.

### Setup Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/mtepenner/dsn-array-simulator.git
   cd dsn-array-simulator
   ```
2. Generate necessary protobufs and build the suite:
   ```bash
   make build
   ```
3. Boot up the physics engine, array controller, infrastructure, and UI via Docker Compose:
   ```bash
   docker-compose up -d
   ```

## 🎮 Usage
Once the services are healthy:
1. Open your browser and navigate to the Mission Dashboard (typically `http://localhost:3000`).
2. Interact with the **Array Controls** to manually add or remove simulated antennas from the tracking array.
3. Observe the **ArrayMap3D** to watch ground stations physically rotate to track the deep-space target.
4. Monitor the **Signal Waterfall** to compare the noisy raw signals against the clean, combined beamformed signal.
5. Check the **Link Budget Table** for real-time improvements in SNR and BER as more antennas are brought online.

## 🤝 Contributing
Contributions are highly encouraged, especially in optimizing the DSP algorithms. Please ensure that any changes to the phase-alignment or PLL math pass the dedicated unit tests defined in `.github/workflows/test-dsp-math.yml`.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/NewModulation`)
3. Commit your Changes (`git commit -m 'Add support for QAM modulation'`)
4. Push to the Branch (`git push origin feature/NewModulation`)
5. Open a Pull Request

## 📄 License
Distributed under the MIT License. See `LICENSE` for more information.

// @ts-nocheck
import React, { useMemo } from 'react';
import Plot from 'react-plotly.js';

interface SignalWaterfallProps {
  combinedI: number[];
  combinedQ: number[];
  antennaSignals?: { id: string; snr: number }[];
}

const SignalWaterfall: React.FC<SignalWaterfallProps> = ({
  combinedI,
  combinedQ,
  antennaSignals,
}) => {
  // Compute magnitude spectrum via simple DFT approximation
  const spectrumData = useMemo(() => {
    if (!combinedI || combinedI.length === 0) return null;

    const n = combinedI.length;
    const magnitude = combinedI.map((val, i) =>
      Math.sqrt(val * val + (combinedQ[i] || 0) * (combinedQ[i] || 0))
    );

    // Time-domain waveform
    const timeAxis = Array.from({ length: n }, (_, i) => i / 10000); // assuming 10kHz

    // Simple power spectrum (magnitude squared, windowed)
    const fftSize = Math.min(n, 256);
    const spectrum: number[] = [];
    const freqAxis: number[] = [];
    for (let k = 0; k < fftSize / 2; k++) {
      let realSum = 0;
      let imagSum = 0;
      for (let t = 0; t < fftSize; t++) {
        const angle = (2 * Math.PI * k * t) / fftSize;
        realSum += (combinedI[t] || 0) * Math.cos(angle) + (combinedQ[t] || 0) * Math.sin(angle);
        imagSum += -(combinedI[t] || 0) * Math.sin(angle) + (combinedQ[t] || 0) * Math.cos(angle);
      }
      const power = 10 * Math.log10((realSum * realSum + imagSum * imagSum) / fftSize + 1e-20);
      spectrum.push(power);
      freqAxis.push((k * 10000) / fftSize);
    }

    return { timeAxis, magnitude, freqAxis, spectrum };
  }, [combinedI, combinedQ]);

  if (!spectrumData) {
    return (
      <div style={{ color: '#666', padding: '20px', textAlign: 'center' }}>
        Waiting for signal data...
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
      {/* Time-domain waveform */}
      <Plot
        data={[
          {
            x: spectrumData.timeAxis,
            y: combinedI.slice(0, spectrumData.timeAxis.length),
            type: 'scatter',
            mode: 'lines',
            name: 'I (In-phase)' as any,
            line: { color: '#00ff88', width: 1 },
          },
          {
            x: spectrumData.timeAxis,
            y: combinedQ.slice(0, spectrumData.timeAxis.length),
            type: 'scatter',
            mode: 'lines',
            name: 'Q (Quadrature)' as any,
            line: { color: '#4488ff', width: 1 },
          },
        ] as any}
        layout={{
          title: { text: 'Combined Signal Waveform', font: { color: '#00ff88', size: 14 } },
          paper_bgcolor: '#0a0a2e',
          plot_bgcolor: '#0d0d3a',
          font: { color: '#888' },
          xaxis: { title: 'Time (s)', gridcolor: '#1a1a4a' },
          yaxis: { title: 'Amplitude', gridcolor: '#1a1a4a' },
          margin: { t: 40, b: 40, l: 50, r: 20 },
          height: 200,
          showlegend: true,
          legend: { font: { size: 10 } },
        }}
        config={{ displayModeBar: false, responsive: true }}
        style={{ width: '100%' }}
      />

      {/* Power spectrum */}
      <Plot
        data={[
          {
            x: spectrumData.freqAxis,
            y: spectrumData.spectrum,
            type: 'scatter',
            mode: 'lines',
            fill: 'tozeroy',
            name: 'Power Spectrum' as any,
            line: { color: '#ff4488' },
            fillcolor: 'rgba(255,68,136,0.2)',
          },
        ] as any}
        layout={{
          title: { text: 'Signal Spectrogram', font: { color: '#00ff88', size: 14 } },
          paper_bgcolor: '#0a0a2e',
          plot_bgcolor: '#0d0d3a',
          font: { color: '#888' },
          xaxis: { title: 'Frequency (Hz)', gridcolor: '#1a1a4a' },
          yaxis: { title: 'Power (dB)', gridcolor: '#1a1a4a' },
          margin: { t: 40, b: 40, l: 50, r: 20 },
          height: 200,
        }}
        config={{ displayModeBar: false, responsive: true }}
        style={{ width: '100%' }}
      />
    </div>
  );
};

export default SignalWaterfall;

import React from 'react';
import { useTelemetry } from './hooks/useTelemetry';
import ArrayMap3D from './components/ArrayMap3D';
import SignalWaterfall from './components/SignalWaterfall';
import LinkBudgetTable from './components/LinkBudgetTable';
import ArrayControls from './components/ArrayControls';

const App: React.FC = () => {
  const { telemetry, connected, error, setActiveAntennas } = useTelemetry();

  return (
    <div style={containerStyle}>
      {/* Header */}
      <header style={headerStyle}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <span style={{ fontSize: '24px' }}>📡</span>
          <div>
            <h1 style={{ margin: 0, fontSize: '18px', letterSpacing: '3px' }}>
              DSN ARRAY SIMULATOR
            </h1>
            <div style={{ fontSize: '11px', color: '#666', letterSpacing: '1px' }}>
              DEEP SPACE NETWORK OPERATIONS CENTER
            </div>
          </div>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <span
            style={{
              display: 'inline-block',
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              backgroundColor: connected ? '#00ff88' : '#ff4444',
            }}
          />
          <span style={{ fontSize: '12px', color: connected ? '#00ff88' : '#ff4444' }}>
            {connected ? 'CONNECTED' : error || 'DISCONNECTED'}
          </span>
        </div>
      </header>

      {/* Main grid */}
      <div style={gridStyle}>
        {/* Left column: 3D Map + Controls */}
        <div style={panelStyle}>
          <div style={panelHeaderStyle}>GROUND STATION ARRAY</div>
          <ArrayMap3D antennas={telemetry?.antennas || []} />
          <ArrayControls
            antennas={telemetry?.antennas || []}
            onAntennaToggle={setActiveAntennas}
          />
        </div>

        {/* Right column: Signal + Metrics */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          <div style={panelStyle}>
            <div style={panelHeaderStyle}>SIGNAL ANALYSIS</div>
            <SignalWaterfall
              combinedI={telemetry?.combined_i_samples || []}
              combinedQ={telemetry?.combined_q_samples || []}
            />
          </div>

          <div style={panelStyle}>
            <div style={panelHeaderStyle}>LINK BUDGET</div>
            <LinkBudgetTable
              antennas={telemetry?.antennas || []}
              combinedSNR={telemetry?.combined_snr_db || 0}
              bitErrorRate={telemetry?.bit_error_rate || 1}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

const containerStyle: React.CSSProperties = {
  minHeight: '100vh',
  background: '#0a0a2e',
  color: '#e0e0e0',
  fontFamily: "'Courier New', monospace",
  padding: '16px',
};

const headerStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '12px 20px',
  background: '#0d0d3a',
  borderRadius: '8px',
  border: '1px solid #1a1a5a',
  marginBottom: '16px',
};

const gridStyle: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: '1fr 1fr',
  gap: '16px',
};

const panelStyle: React.CSSProperties = {
  background: '#0d0d3a',
  borderRadius: '8px',
  border: '1px solid #1a1a5a',
  overflow: 'hidden',
};

const panelHeaderStyle: React.CSSProperties = {
  padding: '10px 16px',
  fontSize: '12px',
  letterSpacing: '2px',
  color: '#00ff88',
  borderBottom: '1px solid #1a1a5a',
  background: '#080830',
};

export default App;

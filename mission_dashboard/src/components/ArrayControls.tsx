import React, { useEffect, useState } from 'react';
import { AntennaTelemetry } from '../hooks/useTelemetry';

interface ArrayControlsProps {
  antennas: AntennaTelemetry[];
  onAntennaToggle: (antennaIds: string[]) => void;
}

// DSN complexes grouping
const DSN_COMPLEXES: Record<string, string[]> = {
  'Goldstone (California)': ['DSS-14', 'DSS-24', 'DSS-25'],
  'Canberra (Australia)': ['DSS-34', 'DSS-35', 'DSS-43'],
  'Madrid (Spain)': ['DSS-54', 'DSS-55', 'DSS-63'],
};

const ArrayControls: React.FC<ArrayControlsProps> = ({ antennas, onAntennaToggle }) => {
  const [activeIds, setActiveIds] = useState<Set<string>>(new Set());

  useEffect(() => {
    const active = new Set(antennas.filter((a) => a.active).map((a) => a.antenna_id));
    setActiveIds(active);
  }, [antennas]);

  const toggleAntenna = (id: string) => {
    const newActive = new Set(activeIds);
    if (newActive.has(id)) {
      if (newActive.size > 1) {
        newActive.delete(id);
      }
    } else {
      newActive.add(id);
    }
    setActiveIds(newActive);
    onAntennaToggle(Array.from(newActive));
  };

  const activateComplex = (complex: string) => {
    const ids = DSN_COMPLEXES[complex] || [];
    const newActive = new Set(activeIds);
    ids.forEach((id) => newActive.add(id));
    setActiveIds(newActive);
    onAntennaToggle(Array.from(newActive));
  };

  const activateAll = () => {
    const allIds = antennas.map((a) => a.antenna_id);
    setActiveIds(new Set(allIds));
    onAntennaToggle(allIds);
  };

  const resetToDefault = () => {
    const defaults = new Set(['DSS-14', 'DSS-43', 'DSS-63']);
    setActiveIds(defaults);
    onAntennaToggle(Array.from(defaults));
  };

  return (
    <div style={{ padding: '16px' }}>
      <h3 style={{ color: '#00ff88', marginTop: 0, fontSize: '14px', letterSpacing: '2px' }}>
        ARRAY CONFIGURATION
      </h3>

      {/* Quick actions */}
      <div style={{ display: 'flex', gap: '8px', marginBottom: '16px', flexWrap: 'wrap' }}>
        <button onClick={activateAll} style={buttonStyle}>
          Activate All
        </button>
        <button onClick={resetToDefault} style={buttonStyle}>
          Reset Default
        </button>
        {Object.keys(DSN_COMPLEXES).map((complex) => (
          <button key={complex} onClick={() => activateComplex(complex)} style={buttonStyle}>
            + {complex.split(' ')[0]}
          </button>
        ))}
      </div>

      {/* Per-complex controls */}
      {Object.entries(DSN_COMPLEXES).map(([complexName, stationIds]) => (
        <div key={complexName} style={{ marginBottom: '12px' }}>
          <div style={{ color: '#888', fontSize: '11px', marginBottom: '4px', letterSpacing: '1px' }}>
            {complexName.toUpperCase()}
          </div>
          <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
            {stationIds.map((id) => {
              const antenna = antennas.find((a) => a.antenna_id === id);
              const isActive = activeIds.has(id);
              return (
                <button
                  key={id}
                  onClick={() => toggleAntenna(id)}
                  style={{
                    ...antennaButtonStyle,
                    backgroundColor: isActive ? '#00331a' : '#1a1a3a',
                    borderColor: isActive ? '#00ff88' : '#333366',
                    color: isActive ? '#00ff88' : '#666',
                  }}
                >
                  <div style={{ fontWeight: 'bold' }}>{id}</div>
                  <div style={{ fontSize: '10px' }}>
                    {antenna?.name || 'Unknown'}
                  </div>
                  {isActive && antenna?.pll_locked && (
                    <div style={{ fontSize: '10px', color: '#00ff88' }}>● LOCKED</div>
                  )}
                </button>
              );
            })}
          </div>
        </div>
      ))}

      <div style={{ color: '#555', fontSize: '11px', marginTop: '12px' }}>
        Active: {activeIds.size} antenna{activeIds.size !== 1 ? 's' : ''} |
        Array gain: +{(10 * Math.log10(Math.max(activeIds.size, 1))).toFixed(1)} dB
      </div>
    </div>
  );
};

const buttonStyle: React.CSSProperties = {
  background: '#1a1a3a',
  color: '#00ff88',
  border: '1px solid #333366',
  borderRadius: '4px',
  padding: '6px 12px',
  cursor: 'pointer',
  fontSize: '12px',
  fontFamily: 'monospace',
};

const antennaButtonStyle: React.CSSProperties = {
  background: '#1a1a3a',
  border: '1px solid #333366',
  borderRadius: '6px',
  padding: '8px 12px',
  cursor: 'pointer',
  fontFamily: 'monospace',
  minWidth: '100px',
  textAlign: 'left' as const,
};

export default ArrayControls;

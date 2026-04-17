import React from 'react';
import { AntennaTelemetry } from '../hooks/useTelemetry';

interface LinkBudgetTableProps {
  antennas: AntennaTelemetry[];
  combinedSNR: number;
  bitErrorRate: number;
}

const LinkBudgetTable: React.FC<LinkBudgetTableProps> = ({
  antennas,
  combinedSNR,
  bitErrorRate,
}) => {
  const formatSNR = (snr: number) => {
    if (!isFinite(snr)) return 'N/A';
    return snr.toFixed(2);
  };

  const formatBER = (ber: number) => {
    if (!isFinite(ber) || ber <= 0) return '< 1e-15';
    if (ber < 0.001) return ber.toExponential(2);
    return ber.toFixed(6);
  };

  const formatDistance = (km: number) => {
    if (!km || km <= 0) return 'N/A';
    if (km > 1e9) return `${(km / 1.496e8).toFixed(2)} AU`;
    if (km > 1e6) return `${(km / 1e6).toFixed(1)}M km`;
    return `${km.toFixed(0)} km`;
  };

  return (
    <div style={{ overflowX: 'auto' }}>
      {/* Summary metrics */}
      <div
        style={{
          display: 'flex',
          gap: '20px',
          marginBottom: '16px',
          flexWrap: 'wrap',
        }}
      >
        <div style={metricBoxStyle}>
          <div style={metricLabelStyle}>Combined SNR</div>
          <div style={{ ...metricValueStyle, color: combinedSNR > 10 ? '#00ff88' : combinedSNR > 3 ? '#ffaa00' : '#ff4444' }}>
            {formatSNR(combinedSNR)} dB
          </div>
        </div>
        <div style={metricBoxStyle}>
          <div style={metricLabelStyle}>Bit Error Rate</div>
          <div style={{ ...metricValueStyle, color: bitErrorRate < 1e-5 ? '#00ff88' : bitErrorRate < 1e-3 ? '#ffaa00' : '#ff4444' }}>
            {formatBER(bitErrorRate)}
          </div>
        </div>
        <div style={metricBoxStyle}>
          <div style={metricLabelStyle}>Active Antennas</div>
          <div style={metricValueStyle}>
            {antennas.filter((a) => a.active).length} / {antennas.length}
          </div>
        </div>
        <div style={metricBoxStyle}>
          <div style={metricLabelStyle}>PLL Locked</div>
          <div style={metricValueStyle}>
            {antennas.filter((a) => a.pll_locked).length} / {antennas.filter((a) => a.active).length}
          </div>
        </div>
      </div>

      {/* Per-antenna table */}
      <table style={tableStyle}>
        <thead>
          <tr>
            <th style={thStyle}>Antenna</th>
            <th style={thStyle}>Status</th>
            <th style={thStyle}>SNR (dB)</th>
            <th style={thStyle}>Phase (rad)</th>
            <th style={thStyle}>Delay (μs)</th>
            <th style={thStyle}>PLL</th>
            <th style={thStyle}>Distance</th>
            <th style={thStyle}>Doppler (Hz)</th>
            <th style={thStyle}>Az/El (°)</th>
          </tr>
        </thead>
        <tbody>
          {antennas.map((ant) => (
            <tr key={ant.antenna_id} style={{ opacity: ant.active ? 1 : 0.4 }}>
              <td style={tdStyle}>
                <span style={{ fontWeight: 'bold' }}>{ant.antenna_id}</span>
                <br />
                <span style={{ fontSize: '11px', color: '#666' }}>{ant.name}</span>
              </td>
              <td style={tdStyle}>
                <span
                  style={{
                    display: 'inline-block',
                    width: '10px',
                    height: '10px',
                    borderRadius: '50%',
                    backgroundColor: ant.active ? '#00ff88' : '#555',
                    marginRight: '6px',
                  }}
                />
                {ant.active ? 'ACTIVE' : 'STANDBY'}
              </td>
              <td style={{ ...tdStyle, color: ant.snr_db > 10 ? '#00ff88' : '#ffaa00' }}>
                {formatSNR(ant.snr_db)}
              </td>
              <td style={tdStyle}>{ant.phase_offset_rad?.toFixed(4) || 'N/A'}</td>
              <td style={tdStyle}>{ant.delay_us?.toFixed(3) || 'N/A'}</td>
              <td style={tdStyle}>
                <span style={{ color: ant.pll_locked ? '#00ff88' : '#ff4444' }}>
                  {ant.pll_locked ? '● LOCKED' : '○ UNLOCKED'}
                </span>
              </td>
              <td style={tdStyle}>{formatDistance(ant.distance_km)}</td>
              <td style={tdStyle}>{ant.doppler_hz?.toFixed(1) || 'N/A'}</td>
              <td style={tdStyle}>
                {ant.azimuth_deg?.toFixed(1) || '–'} / {ant.elevation_deg?.toFixed(1) || '–'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

const metricBoxStyle: React.CSSProperties = {
  background: '#0d0d3a',
  border: '1px solid #1a1a5a',
  borderRadius: '8px',
  padding: '12px 20px',
  minWidth: '140px',
  textAlign: 'center',
};

const metricLabelStyle: React.CSSProperties = {
  fontSize: '11px',
  color: '#888',
  textTransform: 'uppercase',
  letterSpacing: '1px',
  marginBottom: '4px',
};

const metricValueStyle: React.CSSProperties = {
  fontSize: '24px',
  fontWeight: 'bold',
  color: '#00ff88',
  fontFamily: 'monospace',
};

const tableStyle: React.CSSProperties = {
  width: '100%',
  borderCollapse: 'collapse',
  fontSize: '13px',
};

const thStyle: React.CSSProperties = {
  textAlign: 'left',
  padding: '8px 12px',
  borderBottom: '2px solid #1a1a5a',
  color: '#888',
  textTransform: 'uppercase',
  fontSize: '11px',
  letterSpacing: '1px',
};

const tdStyle: React.CSSProperties = {
  padding: '6px 12px',
  borderBottom: '1px solid #111144',
  fontFamily: 'monospace',
};

export default LinkBudgetTable;

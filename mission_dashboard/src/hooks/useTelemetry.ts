import { useState, useEffect, useRef, useCallback } from 'react';

export interface AntennaTelemetry {
  antenna_id: string;
  name: string;
  snr_db: number;
  phase_offset_rad: number;
  delay_us: number;
  pll_locked: boolean;
  latitude: number;
  longitude: number;
  azimuth_deg: number;
  elevation_deg: number;
  active: boolean;
  distance_km: number;
  doppler_hz: number;
  path_loss_db: number;
}

export interface TelemetryData {
  timestamp: number;
  combined_snr_db: number;
  bit_error_rate: number;
  antennas: AntennaTelemetry[];
  combined_i_samples: number[];
  combined_q_samples: number[];
}

const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws/telemetry';
const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export function useTelemetry() {
  const [telemetry, setTelemetry] = useState<TelemetryData | null>(null);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeout = useRef<NodeJS.Timeout | null>(null);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    try {
      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => {
        setConnected(true);
        setError(null);
        console.log('WebSocket connected to Array Controller');
      };

      ws.onmessage = (event) => {
        try {
          const data: TelemetryData = JSON.parse(event.data);
          setTelemetry(data);
        } catch (e) {
          console.error('Failed to parse telemetry:', e);
        }
      };

      ws.onclose = () => {
        setConnected(false);
        // Reconnect after 2 seconds
        reconnectTimeout.current = setTimeout(connect, 2000);
      };

      ws.onerror = (e) => {
        setError('WebSocket connection error');
        console.error('WebSocket error:', e);
      };
    } catch (e) {
      setError('Failed to connect to Array Controller');
      reconnectTimeout.current = setTimeout(connect, 5000);
    }
  }, []);

  useEffect(() => {
    connect();
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeout.current) {
        clearTimeout(reconnectTimeout.current);
      }
    };
  }, [connect]);

  const setActiveAntennas = useCallback(async (antennaIds: string[]) => {
    try {
      await fetch(`${API_URL}/api/antennas`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ active_antenna_ids: antennaIds }),
      });
    } catch (e) {
      console.error('Failed to update antenna config:', e);
    }
  }, []);

  const fetchAntennas = useCallback(async () => {
    try {
      const res = await fetch(`${API_URL}/api/antennas`);
      return await res.json();
    } catch (e) {
      console.error('Failed to fetch antennas:', e);
      return [];
    }
  }, []);

  return { telemetry, connected, error, setActiveAntennas, fetchAntennas };
}

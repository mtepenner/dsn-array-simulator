import React, { useRef, useMemo } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { OrbitControls, Stars, Html } from '@react-three/drei';
import * as THREE from 'three';
import { AntennaTelemetry } from '../hooks/useTelemetry';

interface ArrayMap3DProps {
  antennas: AntennaTelemetry[];
}

function Earth() {
  const meshRef = useRef<THREE.Mesh>(null);

  useFrame((_, delta) => {
    if (meshRef.current) {
      meshRef.current.rotation.y += delta * 0.05;
    }
  });

  return (
    <mesh ref={meshRef}>
      <sphereGeometry args={[2, 64, 64]} />
      <meshStandardMaterial
        color="#1a4a8a"
        roughness={0.8}
        metalness={0.2}
        wireframe={false}
      />
      {/* Continents approximation using additional geometry */}
      <mesh>
        <sphereGeometry args={[2.01, 32, 32]} />
        <meshStandardMaterial
          color="#2d6a2d"
          transparent
          opacity={0.4}
          wireframe={true}
        />
      </mesh>
    </mesh>
  );
}

function AntennaMarker({ antenna }: { antenna: AntennaTelemetry }) {
  const meshRef = useRef<THREE.Mesh>(null);

  // Convert lat/lon to 3D position on sphere
  const position = useMemo(() => {
    const lat = (antenna.latitude * Math.PI) / 180;
    const lon = (antenna.longitude * Math.PI) / 180;
    const r = 2.05;
    return new THREE.Vector3(
      r * Math.cos(lat) * Math.cos(lon),
      r * Math.sin(lat),
      -r * Math.cos(lat) * Math.sin(lon)
    );
  }, [antenna.latitude, antenna.longitude]);

  // Animate dish rotation
  useFrame((_, delta) => {
    if (meshRef.current && antenna.active) {
      meshRef.current.rotation.x += delta * 0.5;
    }
  });

  const color = antenna.active
    ? antenna.pll_locked
      ? '#00ff88'
      : '#ffaa00'
    : '#555555';

  return (
    <group position={position}>
      <mesh ref={meshRef}>
        <coneGeometry args={[0.08, 0.15, 8]} />
        <meshStandardMaterial color={color} emissive={color} emissiveIntensity={0.5} />
      </mesh>
      {/* Signal beam when active */}
      {antenna.active && (
        <mesh position={[0, 0.2, 0]}>
          <cylinderGeometry args={[0.01, 0.05, 0.3, 8]} />
          <meshStandardMaterial
            color="#00ff88"
            transparent
            opacity={0.3}
            emissive="#00ff88"
            emissiveIntensity={0.8}
          />
        </mesh>
      )}
      <Html distanceFactor={8} position={[0, 0.3, 0]}>
        <div
          style={{
            color: color,
            fontSize: '10px',
            fontFamily: 'monospace',
            whiteSpace: 'nowrap',
            textShadow: '0 0 4px rgba(0,0,0,0.8)',
          }}
        >
          {antenna.antenna_id}
        </div>
      </Html>
    </group>
  );
}

const ArrayMap3D: React.FC<ArrayMap3DProps> = ({ antennas }) => {
  return (
    <div style={{ width: '100%', height: '400px', background: '#050520', borderRadius: '8px' }}>
      <Canvas camera={{ position: [0, 2, 5], fov: 50 }}>
        <ambientLight intensity={0.3} />
        <directionalLight position={[5, 5, 5]} intensity={1} />
        <pointLight position={[-5, -5, -5]} intensity={0.5} color="#4488ff" />
        <Stars radius={100} depth={50} count={5000} factor={4} />
        <Earth />
        {antennas.map((ant) => (
          <AntennaMarker key={ant.antenna_id} antenna={ant} />
        ))}
        <OrbitControls enablePan={false} minDistance={3} maxDistance={10} />
      </Canvas>
    </div>
  );
};

export default ArrayMap3D;

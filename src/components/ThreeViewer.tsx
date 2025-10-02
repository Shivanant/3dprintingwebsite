"use client";
import { Canvas, useFrame } from "@react-three/fiber";
import { OrbitControls } from "@react-three/drei";
import { useEffect, useMemo, useState } from "react";
import * as THREE from "three";
import { STLLoader } from "three/examples/jsm/loaders/STLLoader.js";

function STLModel({ url }: { url: string }) {
  const [geom, setGeom] = useState<THREE.BufferGeometry | null>(null);

  useEffect(() => {
    if (!url) return;
    const loader = new STLLoader();
    loader.load(
      url,
      (g) => {
        // center + scale to fit box
        g.computeBoundingBox();
        const box = g.boundingBox!;
        const size = new THREE.Vector3();
        box.getSize(size);
        const maxDim = Math.max(size.x, size.y, size.z) || 1;
        const scale = 1.5 / maxDim; // fit into unit-ish box
        const mat = new THREE.Matrix4()
          .makeTranslation(
            -(box.min.x + size.x / 2),
            -(box.min.y + size.y / 2),
            -(box.min.z + size.z / 2)
          )
          .multiply(new THREE.Matrix4().makeScale(scale, scale, scale));
        g.applyMatrix4(mat);
        setGeom(g);
      },
      undefined,
      (err) => console.error("STL load error", err)
    );
  }, [url]);

  const mat = useMemo(
    () => new THREE.MeshStandardMaterial({ metalness: 0.2, roughness: 0.7 }),
    []
  );

  if (!geom) return null;
  return <mesh geometry={geom} material={mat} />;
}

export default function ThreeViewer({ url }: { url?: string }) {
  return (
    <div className="h-80 rounded-xl overflow-hidden border bg-white">
      <Canvas camera={{ position: [2, 1.5, 2.5], fov: 50 }}>
        <ambientLight intensity={0.9} />
        <directionalLight position={[5, 5, 5]} intensity={0.8} />
        <gridHelper args={[10, 10]} />
        {url ? (
          <STLModel url={url} />
        ) : (
          <mesh>
            <boxGeometry args={[1, 1, 1]} />
            <meshStandardMaterial />
          </mesh>
        )}
        <OrbitControls />
      </Canvas>
    </div>
  );
}

import { useRef, useEffect, useState, useCallback, Suspense } from "react";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { VRMModelAdapter, loadVRM } from "./vrm/VRMAdapter";
import type { ModelAdapter, StandardExpression } from "./model/types";

/** Props for the VRM scene (internal R3F component) */
interface VRMSceneProps {
  modelUrl: string;
  onLoad?: (adapter: ModelAdapter) => void;
  onProgress?: (percent: number) => void;
  onError?: (error: Error) => void;
}

function VRMScene({ modelUrl, onLoad, onProgress, onError }: VRMSceneProps) {
  const adapterRef = useRef<VRMModelAdapter | null>(null);
  const { scene } = useThree();

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        // Clean up previous model
        if (adapterRef.current) {
          scene.remove(adapterRef.current.sceneObject);
          adapterRef.current.dispose();
          adapterRef.current = null;
        }

        const adapter = await loadVRM(modelUrl, onProgress);
        if (cancelled) {
          adapter.dispose();
          return;
        }

        adapterRef.current = adapter;
        scene.add(adapter.sceneObject);
        onLoad?.(adapter);
      } catch (e) {
        if (!cancelled) {
          onError?.(e instanceof Error ? e : new Error(String(e)));
        }
      }
    }

    load();

    return () => {
      cancelled = true;
      if (adapterRef.current) {
        scene.remove(adapterRef.current.sceneObject);
        adapterRef.current.dispose();
        adapterRef.current = null;
      }
    };
  }, [modelUrl, scene, onLoad, onProgress, onError]);

  // Auto blink
  const blinkTimer = useRef(0);
  const isBlinking = useRef(false);

  useFrame((_, delta) => {
    if (!adapterRef.current) return;
    adapterRef.current.update(delta);

    // Simple auto-blink
    blinkTimer.current += delta;
    if (!isBlinking.current && blinkTimer.current > 3 + Math.random() * 4) {
      isBlinking.current = true;
      blinkTimer.current = 0;
    }
    if (isBlinking.current) {
      const t = blinkTimer.current;
      if (t < 0.06) {
        adapterRef.current.setExpressionWeight("blink", t / 0.06);
      } else if (t < 0.12) {
        adapterRef.current.setExpressionWeight("blink", 1 - (t - 0.06) / 0.06);
      } else {
        adapterRef.current.setExpressionWeight("blink", 0);
        isBlinking.current = false;
        blinkTimer.current = 0;
      }
    }
  });

  return null;
}

/** Props for the AvatarViewer component */
export interface AvatarViewerProps {
  modelUrl: string;
  width?: number | string;
  height?: number | string;
  className?: string;
  cameraPosition?: [number, number, number];
  cameraTarget?: [number, number, number];
  onAdapterReady?: (adapter: ModelAdapter) => void;
  /** When true, loading/error overlays use no background so the viewer blends into the page. */
  transparent?: boolean;
}

function CameraSetup({
  position,
  target,
}: {
  position: [number, number, number];
  target: [number, number, number];
}) {
  const { camera } = useThree();

  useEffect(() => {
    camera.position.set(...position);
    (camera as THREE.PerspectiveCamera).lookAt(new THREE.Vector3(...target));
  }, [camera, position, target]);

  return null;
}

export default function AvatarViewer({
  modelUrl,
  width = "100%",
  height = "100%",
  className = "",
  cameraPosition = [0, 1.1, 2.5],
  cameraTarget = [0, 1.0, 0],
  onAdapterReady,
  transparent = false,
}: AvatarViewerProps) {
  const [loading, setLoading] = useState(true);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const handleLoad = useCallback(
    (adapter: ModelAdapter) => {
      setLoading(false);
      setError(null);
      onAdapterReady?.(adapter);
    },
    [onAdapterReady],
  );

  const handleProgress = useCallback((percent: number) => {
    setProgress(percent);
  }, []);

  const handleError = useCallback((err: Error) => {
    setLoading(false);
    setError(err.message);
    console.error("AvatarViewer load error:", err);
  }, []);

  return (
    <div className={`relative ${className}`} style={{ width, height }}>
      <Canvas
        dpr={[1, 2]}
        shadows
        gl={{ alpha: true, antialias: true }}
        style={{ background: "transparent" }}
      >
        <CameraSetup position={cameraPosition} target={cameraTarget} />
        <ambientLight intensity={0.6} />
        <directionalLight
          position={[2, 3, 2]}
          intensity={1.2}
          castShadow
          shadow-mapSize-width={1024}
          shadow-mapSize-height={1024}
        />
        <Suspense fallback={null}>
          <VRMScene
            modelUrl={modelUrl}
            onLoad={handleLoad}
            onProgress={handleProgress}
            onError={handleError}
          />
        </Suspense>
        {/* Ground shadow plane */}
        <mesh
          rotation={[-Math.PI / 2, 0, 0]}
          position={[0, 0, 0]}
          receiveShadow
        >
          <planeGeometry args={[10, 10]} />
          <shadowMaterial opacity={0.15} />
        </mesh>
      </Canvas>

      {/* Loading overlay */}
      {loading && (
        <div
          className={`absolute inset-0 flex flex-col items-center justify-center ${transparent ? "" : "bg-black/20 backdrop-blur-sm rounded-2xl"}`}
        >
          <div className="w-8 h-8 border-2 border-white/20 border-t-blue-400 rounded-full animate-spin mb-3" />
          <p className="text-xs text-white/50">
            Loading model... {progress > 0 ? `${progress}%` : ""}
          </p>
        </div>
      )}

      {/* Error overlay */}
      {error && (
        <div
          className={`absolute inset-0 flex flex-col items-center justify-center ${transparent ? "" : "bg-black/20 backdrop-blur-sm rounded-2xl"}`}
        >
          <p className="text-xs text-red-400/70 text-center px-4">
            Failed to load model
          </p>
          <p className="text-[10px] text-white/30 mt-1 text-center px-4 max-w-[200px] truncate">
            {error}
          </p>
        </div>
      )}
    </div>
  );
}

/** Helper to trigger expressions from outside the component */
export function playExpression(
  adapter: ModelAdapter,
  expression: StandardExpression,
  weight = 1,
  duration = 2000,
): void {
  adapter.setExpressionWeight(expression, weight);
  setTimeout(() => {
    adapter.setExpressionWeight(expression, 0);
  }, duration);
}

/** Helper to play an animation from URL */
export async function playAnimation(
  adapter: ModelAdapter,
  url: string,
  fileType: string,
  loop = false,
): Promise<void> {
  const clip = await adapter.loadAnimationClip(url, fileType);
  if (!clip) return;

  const action = adapter.mixer.clipAction(clip);
  action.reset();
  action.setLoop(loop ? THREE.LoopRepeat : THREE.LoopOnce, loop ? Infinity : 1);
  action.clampWhenFinished = !loop;
  action.fadeIn(0.3);
  action.play();

  if (!loop) {
    adapter.mixer.addEventListener("finished", function onFinished(e) {
      if (e.action === action) {
        action.fadeOut(0.3);
        adapter.mixer.removeEventListener("finished", onFinished);
      }
    });
  }
}

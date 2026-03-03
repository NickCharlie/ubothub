import * as THREE from "three";
import { VRM, VRMHumanBoneName, VRMUtils } from "@pixiv/three-vrm";
import {
  createVRMAnimationClip,
  VRMAnimationLoaderPlugin,
} from "@pixiv/three-vrm-animation";
import { GLTFLoader } from "three/addons/loaders/GLTFLoader.js";
import type {
  ModelAdapter,
  StandardBoneName,
  StandardExpression,
} from "../model/types";
import { loadMixamoAnimation } from "./loadMixamoAnimation";

/** Map standard bone names to VRM humanoid bone names */
const boneNameMap: Record<StandardBoneName, VRMHumanBoneName> = {
  head: VRMHumanBoneName.Head,
  neck: VRMHumanBoneName.Neck,
  chest: VRMHumanBoneName.Chest,
  upperChest: VRMHumanBoneName.UpperChest,
  spine: VRMHumanBoneName.Spine,
  hips: VRMHumanBoneName.Hips,
  leftShoulder: VRMHumanBoneName.LeftShoulder,
  leftUpperArm: VRMHumanBoneName.LeftUpperArm,
  leftLowerArm: VRMHumanBoneName.LeftLowerArm,
  leftHand: VRMHumanBoneName.LeftHand,
  rightShoulder: VRMHumanBoneName.RightShoulder,
  rightUpperArm: VRMHumanBoneName.RightUpperArm,
  rightLowerArm: VRMHumanBoneName.RightLowerArm,
  rightHand: VRMHumanBoneName.RightHand,
  leftUpperLeg: VRMHumanBoneName.LeftUpperLeg,
  leftLowerLeg: VRMHumanBoneName.LeftLowerLeg,
  leftFoot: VRMHumanBoneName.LeftFoot,
  leftToes: VRMHumanBoneName.LeftToes,
  rightUpperLeg: VRMHumanBoneName.RightUpperLeg,
  rightLowerLeg: VRMHumanBoneName.RightLowerLeg,
  rightFoot: VRMHumanBoneName.RightFoot,
  rightToes: VRMHumanBoneName.RightToes,
};

/** Map standard expressions to VRM expression names */
const expressionMap: Record<StandardExpression, string> = {
  happy: "happy",
  sad: "sad",
  angry: "angry",
  surprised: "surprised",
  relaxed: "relaxed",
  blink: "blink",
  aa: "aa",
  ih: "ih",
  ou: "ou",
  ee: "ee",
  oh: "oh",
};

export class VRMModelAdapter implements ModelAdapter {
  readonly modelType = "vrm" as const;
  readonly sceneObject: THREE.Object3D;
  readonly mixer: THREE.AnimationMixer;
  private vrm: VRM;

  constructor(vrm: VRM) {
    this.vrm = vrm;
    this.sceneObject = vrm.scene;
    this.mixer = new THREE.AnimationMixer(vrm.scene);
  }

  update(delta: number): void {
    this.vrm.update(delta);
    this.mixer.update(delta);
  }

  getBoneNode(name: StandardBoneName): THREE.Object3D | null {
    const vrmBoneName = boneNameMap[name];
    if (!vrmBoneName) return null;
    return this.vrm.humanoid?.getNormalizedBoneNode(vrmBoneName) ?? null;
  }

  getExpressionTrackName(expr: StandardExpression): string | null {
    const vrmExprName = expressionMap[expr];
    if (!vrmExprName) return null;
    return (
      this.vrm.expressionManager?.getExpressionTrackName(vrmExprName) ?? null
    );
  }

  setExpressionWeight(expr: StandardExpression, weight: number): void {
    const vrmExprName = expressionMap[expr];
    if (!vrmExprName) return;
    this.vrm.expressionManager?.setValue(vrmExprName, weight);
  }

  getExpressionWeight(expr: StandardExpression): number {
    const vrmExprName = expressionMap[expr];
    if (!vrmExprName) return 0;
    return this.vrm.expressionManager?.getValue(vrmExprName) ?? 0;
  }

  async loadAnimationClip(
    url: string,
    fileType: string,
  ): Promise<THREE.AnimationClip | null> {
    try {
      switch (fileType.toLowerCase()) {
        case "fbx": {
          return await loadMixamoAnimation(url, this.vrm);
        }
        case "vrma": {
          const loader = new GLTFLoader();
          loader.register((parser) => new VRMAnimationLoaderPlugin(parser));
          const gltf = await loader.loadAsync(url);
          const vrmAnimations = gltf.userData.vrmAnimations;
          if (vrmAnimations && vrmAnimations.length > 0) {
            return createVRMAnimationClip(vrmAnimations[0], this.vrm);
          }
          return null;
        }
        case "glb":
        case "gltf": {
          const loader = new GLTFLoader();
          const gltf = await loader.loadAsync(url);
          if (gltf.animations.length > 0) {
            return gltf.animations[0] ?? null;
          }
          return null;
        }
        default:
          return null;
      }
    } catch (e) {
      console.error(`Failed to load animation: ${url}`, e);
      return null;
    }
  }

  dispose(): void {
    this.mixer.stopAllAction();
    VRMUtils.deepDispose(this.vrm.scene);
  }
}

/** Load a VRM model from URL and return a VRMModelAdapter */
export async function loadVRM(
  url: string,
  onProgress?: (percent: number) => void,
): Promise<VRMModelAdapter> {
  const loader = new GLTFLoader();
  loader.register((parser) => new VRMAnimationLoaderPlugin(parser));

  const gltf = await loader.loadAsync(url, (event) => {
    if (event.lengthComputable && onProgress) {
      onProgress(Math.round((event.loaded / event.total) * 100));
    }
  });

  const vrm = gltf.userData.vrm as VRM;
  if (!vrm) {
    throw new Error("Failed to load VRM from file");
  }

  // Optimize the VRM model
  VRMUtils.removeUnnecessaryVertices(vrm.scene);

  // Rotate model to face camera
  vrm.scene.rotation.y = Math.PI;

  return new VRMModelAdapter(vrm);
}

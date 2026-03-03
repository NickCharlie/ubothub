import * as THREE from "three";
import { FBXLoader } from "three/addons/loaders/FBXLoader.js";
import { VRM, VRMHumanBoneName } from "@pixiv/three-vrm";

/**
 * Mixamo rig name → VRM humanoid bone name mapping.
 * Ref: liynweb loadMixamoAnimation.ts
 */
const mixamoVRMRigMap: Record<string, string> = {
  mixamorigHips: "hips",
  mixamorigSpine: "spine",
  mixamorigSpine1: "chest",
  mixamorigSpine2: "upperChest",
  mixamorigNeck: "neck",
  mixamorigHead: "head",
  mixamorigLeftShoulder: "leftShoulder",
  mixamorigLeftArm: "leftUpperArm",
  mixamorigLeftForeArm: "leftLowerArm",
  mixamorigLeftHand: "leftHand",
  mixamorigLeftHandThumb1: "leftThumbMetacarpal",
  mixamorigLeftHandThumb2: "leftThumbProximal",
  mixamorigLeftHandThumb3: "leftThumbDistal",
  mixamorigLeftHandIndex1: "leftIndexProximal",
  mixamorigLeftHandIndex2: "leftIndexIntermediate",
  mixamorigLeftHandIndex3: "leftIndexDistal",
  mixamorigLeftHandMiddle1: "leftMiddleProximal",
  mixamorigLeftHandMiddle2: "leftMiddleIntermediate",
  mixamorigLeftHandMiddle3: "leftMiddleDistal",
  mixamorigLeftHandRing1: "leftRingProximal",
  mixamorigLeftHandRing2: "leftRingIntermediate",
  mixamorigLeftHandRing3: "leftRingDistal",
  mixamorigLeftHandPinky1: "leftLittleProximal",
  mixamorigLeftHandPinky2: "leftLittleIntermediate",
  mixamorigLeftHandPinky3: "leftLittleDistal",
  mixamorigRightShoulder: "rightShoulder",
  mixamorigRightArm: "rightUpperArm",
  mixamorigRightForeArm: "rightLowerArm",
  mixamorigRightHand: "rightHand",
  mixamorigRightHandThumb1: "rightThumbMetacarpal",
  mixamorigRightHandThumb2: "rightThumbProximal",
  mixamorigRightHandThumb3: "rightThumbDistal",
  mixamorigRightHandIndex1: "rightIndexProximal",
  mixamorigRightHandIndex2: "rightIndexIntermediate",
  mixamorigRightHandIndex3: "rightIndexDistal",
  mixamorigRightHandMiddle1: "rightMiddleProximal",
  mixamorigRightHandMiddle2: "rightMiddleIntermediate",
  mixamorigRightHandMiddle3: "rightMiddleDistal",
  mixamorigRightHandRing1: "rightRingProximal",
  mixamorigRightHandRing2: "rightRingIntermediate",
  mixamorigRightHandRing3: "rightRingDistal",
  mixamorigRightHandPinky1: "rightLittleProximal",
  mixamorigRightHandPinky2: "rightLittleIntermediate",
  mixamorigRightHandPinky3: "rightLittleDistal",
  mixamorigLeftUpLeg: "leftUpperLeg",
  mixamorigLeftLeg: "leftLowerLeg",
  mixamorigLeftFoot: "leftFoot",
  mixamorigLeftToeBase: "leftToes",
  mixamorigRightUpLeg: "rightUpperLeg",
  mixamorigRightLeg: "rightLowerLeg",
  mixamorigRightFoot: "rightFoot",
  mixamorigRightToeBase: "rightToes",
};

/**
 * Load a Mixamo FBX animation and retarget it to a VRM model.
 * Converts Mixamo rig bone rotations and positions to VRM humanoid bones
 * with proper rest-pose compensation and height scaling.
 */
export async function loadMixamoAnimation(
  url: string,
  vrm: VRM,
): Promise<THREE.AnimationClip> {
  const loader = new FBXLoader();
  const asset = await loader.loadAsync(url);
  const clip = THREE.AnimationClip.findByName(asset.animations, "mixamo.com")!;
  const tracks: THREE.KeyframeTrack[] = [];
  const restRotationInverse = new THREE.Quaternion();
  const parentRestWorldRotation = new THREE.Quaternion();
  const _quatA = new THREE.Quaternion();

  // Scale hips position to match VRM model height
  const motionHipsHeight = asset.getObjectByName("mixamorigHips")!.position.y;
  const vrmHipsHeight = vrm.humanoid.normalizedRestPose.hips!.position![1];
  const hipsPositionScale = vrmHipsHeight / motionHipsHeight;

  clip.tracks.forEach((track) => {
    const trackSplitted = track.name.split(".");
    const mixamoRigName = trackSplitted[0] ?? "";
    const vrmBoneName = mixamoVRMRigMap[mixamoRigName];
    if (!vrmBoneName) return;
    const vrmNodeName = vrm.humanoid?.getNormalizedBoneNode(
      vrmBoneName as VRMHumanBoneName,
    )?.name;
    const mixamoRigNode = asset.getObjectByName(mixamoRigName);

    if (vrmNodeName != null) {
      const propertyName = trackSplitted[1];

      mixamoRigNode!.getWorldQuaternion(restRotationInverse).invert();
      mixamoRigNode!.parent!.getWorldQuaternion(parentRestWorldRotation);

      if (track instanceof THREE.QuaternionKeyframeTrack) {
        for (let i = 0; i < track.values.length; i += 4) {
          const flatQuaternion = track.values.slice(i, i + 4);
          _quatA.fromArray(flatQuaternion);
          _quatA
            .premultiply(parentRestWorldRotation)
            .multiply(restRotationInverse);
          _quatA.toArray(flatQuaternion);
          flatQuaternion.forEach((v, index) => {
            track.values[index + i] = v;
          });
        }
        tracks.push(
          new THREE.QuaternionKeyframeTrack(
            `${vrmNodeName}.${propertyName}`,
            track.times,
            track.values.map((v, i) =>
              vrm.meta?.metaVersion === "0" && i % 2 === 0 ? -v : v,
            ),
          ),
        );
      } else if (track instanceof THREE.VectorKeyframeTrack) {
        const value = track.values.map(
          (v, i) =>
            (vrm.meta?.metaVersion === "0" && i % 3 !== 1 ? -v : v) *
            hipsPositionScale,
        );
        tracks.push(
          new THREE.VectorKeyframeTrack(
            `${vrmNodeName}.${propertyName}`,
            track.times,
            value,
          ),
        );
      }
    }
  });

  return new THREE.AnimationClip("vrmAnimation", clip.duration, tracks);
}

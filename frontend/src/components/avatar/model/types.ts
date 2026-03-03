import { AnimationClip, AnimationMixer, Object3D } from "three";

/** Supported model format types */
export type ModelType = "vrm";

/** Standardized bone names across model formats */
export type StandardBoneName =
  | "head"
  | "neck"
  | "chest"
  | "upperChest"
  | "spine"
  | "hips"
  | "leftShoulder"
  | "leftUpperArm"
  | "leftLowerArm"
  | "leftHand"
  | "rightShoulder"
  | "rightUpperArm"
  | "rightLowerArm"
  | "rightHand"
  | "leftUpperLeg"
  | "leftLowerLeg"
  | "leftFoot"
  | "leftToes"
  | "rightUpperLeg"
  | "rightLowerLeg"
  | "rightFoot"
  | "rightToes";

/** Standardized expression names for facial animation */
export type StandardExpression =
  | "happy"
  | "sad"
  | "angry"
  | "surprised"
  | "relaxed"
  | "blink"
  | "aa"
  | "ih"
  | "ou"
  | "ee"
  | "oh";

/** Animation state for the avatar */
export interface AnimationState {
  current: string;
  queue: AnimationTransform[];
  expression: string;
  expressionQueue: ExpressionTransform[];
}

export interface AnimationTransform {
  name: string;
  url?: string;
  fileType?: string;
  loop?: boolean;
  speed?: number;
  fadeIn?: number;
}

export interface ExpressionTransform {
  name: StandardExpression;
  weight: number;
  duration?: number;
}

/**
 * Unified model adapter interface.
 * Allows the avatar viewer to work with any model format
 * through a single abstraction. Ref: liynweb model/types.ts
 */
export interface ModelAdapter {
  readonly modelType: ModelType;
  readonly sceneObject: Object3D;
  readonly mixer: AnimationMixer;

  update(delta: number): void;
  getBoneNode(name: StandardBoneName): Object3D | null;
  getExpressionTrackName(expr: StandardExpression): string | null;
  setExpressionWeight(expr: StandardExpression, weight: number): void;
  getExpressionWeight(expr: StandardExpression): number;
  loadAnimationClip(
    url: string,
    fileType: string,
  ): Promise<AnimationClip | null>;
  dispose(): void;
}

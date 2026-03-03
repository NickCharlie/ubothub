import http from "./http";

export interface AvatarConfig {
  id: string;
  user_id: string;
  bot_id: string | null;
  name: string;
  description: string;
  render_type: "3d" | "live2d" | "sprite";
  scene_config: Record<string, any>;
  action_mapping: Record<string, any>;
  is_default: boolean;
  avatar_assets?: AvatarAsset[];
  bot?: { id: string; name: string } | null;
  created_at: string;
  updated_at: string;
}

export interface AvatarAsset {
  id: string;
  avatar_id: string;
  asset_id: string;
  role: string;
  config: Record<string, any>;
  sort_order: number;
  asset?: {
    id: string;
    name: string;
    category: string;
    format: string;
    thumbnail_key: string;
  };
}

export interface CreateAvatarParams {
  name: string;
  description?: string;
  render_type: "3d" | "live2d" | "sprite";
  scene_config?: Record<string, any>;
  action_mapping?: Record<string, any>;
}

export interface UpdateAvatarParams {
  name?: string;
  description?: string;
  scene_config?: Record<string, any>;
  action_mapping?: Record<string, any>;
}

export interface BindAssetParams {
  asset_id: string;
  role: string;
  config?: Record<string, any>;
  sort_order?: number;
}

export const avatarApi = {
  list: (page = 1, pageSize = 20) =>
    http.get("/avatars", { params: { page, page_size: pageSize } }),
  get: (id: string) => http.get(`/avatars/${id}`),
  create: (data: CreateAvatarParams) => http.post<{ data: AvatarConfig }>("/avatars", data),
  update: (id: string, data: UpdateAvatarParams) => http.put(`/avatars/${id}`, data),
  delete: (id: string) => http.delete(`/avatars/${id}`),
  bindBot: (id: string, botId: string) =>
    http.post(`/avatars/${id}/bind-bot`, { bot_id: botId }),
  bindAsset: (id: string, data: BindAssetParams) =>
    http.post(`/avatars/${id}/bind-asset`, data),
  unbindAsset: (id: string, assetId: string) =>
    http.delete(`/avatars/${id}/assets/${assetId}`),
  updateActionMapping: (id: string, actionMapping: Record<string, any>) =>
    http.put(`/avatars/${id}/action-mapping`, { action_mapping: actionMapping }),
};

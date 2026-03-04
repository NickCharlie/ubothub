import axios from "axios";

// Plaza uses a separate axios instance because public endpoints
// should NOT trigger a redirect to /auth/login on 401.
const publicHttp = axios.create({
  baseURL: "/api/v1",
  timeout: 15000,
  headers: { "Content-Type": "application/json" },
});

export interface PlazaBot {
  id: string;
  name: string;
  description: string;
  framework: string;
  visibility: string;
  status: string;
  config: string;
  last_active_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface PlazaAvatar {
  id: string;
  user_id: string;
  bot_id: string;
  name: string;
  description: string;
  render_type: string;
  scene_config: string;
  action_mapping: string;
  is_default: boolean;
  avatar_assets?: PlazaAvatarAsset[];
  created_at: string;
  updated_at: string;
}

export interface PlazaAvatarAsset {
  asset_id: string;
  asset_name: string;
  role: string;
  config: string;
  sort_order: number;
}

export const plazaApi = {
  listBots: (page = 1, pageSize = 20) =>
    publicHttp.get("/plaza/bots", { params: { page, page_size: pageSize } }),

  getBot: (id: string) => publicHttp.get(`/plaza/bots/${id}`),

  getAvatar: (id: string) => publicHttp.get(`/plaza/avatars/${id}`),

  getAssetDownloadURL: (assetId: string) =>
    publicHttp.get<{ data: { download_url: string; expires_in: number } }>(
      `/plaza/assets/${assetId}/download`,
    ),
};

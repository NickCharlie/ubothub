import http from "./http";

// AstrBot-specific configuration stored in the bot's JSONB config field.
export interface AstrBotConfig {
  api_key?: string;
  platform?: string;
}

export interface Bot {
  id: string;
  name: string;
  description: string;
  framework: string;
  webhook_url: string;
  config: string;
  visibility: string;
  status: string;
  last_active_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface BotWithToken extends Bot {
  access_token: string;
}

export interface CreateBotParams {
  name: string;
  description?: string;
  framework: string;
  webhook_url?: string;
  config?: string;
  visibility?: string;
}

export interface UpdateBotParams {
  name?: string;
  description?: string;
  webhook_url?: string;
  visibility?: string;
  config?: string;
}

export const botApi = {
  list: (page = 1, pageSize = 20) =>
    http.get("/bots", { params: { page, page_size: pageSize } }),
  get: (id: string) => http.get(`/bots/${id}`),
  create: (data: CreateBotParams) =>
    http.post<{ data: BotWithToken }>("/bots", data),
  update: (id: string, data: UpdateBotParams) => http.put(`/bots/${id}`, data),
  delete: (id: string) => http.delete(`/bots/${id}`),
  regenerateToken: (id: string) => http.post(`/bots/${id}/regenerate-token`),
};

export const adminApi = {
  dashboard: () => http.get("/admin/dashboard"),
  listUsers: (page = 1, pageSize = 20, status?: string, role?: string) =>
    http.get("/admin/users", {
      params: { page, page_size: pageSize, status, role },
    }),
  banUser: (id: string) => http.put(`/admin/users/${id}/ban`),
  unbanUser: (id: string) => http.put(`/admin/users/${id}/unban`),
  listBots: (page = 1, pageSize = 20) =>
    http.get("/admin/bots", { params: { page, page_size: pageSize } }),
  forceDeleteBot: (id: string) => http.delete(`/admin/bots/${id}`),
};

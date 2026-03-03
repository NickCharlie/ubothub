import http from "./http";

export interface Asset {
  id: string;
  user_id: string;
  name: string;
  description: string;
  category: "model_3d" | "model_live2d" | "motion" | "texture";
  format: string;
  file_key: string;
  file_size: number;
  thumbnail_key: string;
  metadata: Record<string, any>;
  tags: string[];
  is_public: boolean;
  download_count: number;
  status: "processing" | "ready" | "failed";
  created_at: string;
  updated_at: string;
}

export interface CompleteUploadParams {
  file_key: string;
  name: string;
  description?: string;
  category: string;
  format: string;
  file_size: number;
  is_public?: boolean;
  tags?: string[];
}

export interface UpdateAssetParams {
  name?: string;
  description?: string;
  is_public?: boolean;
  tags?: string[];
}

export const assetApi = {
  getPresignedUpload: (filename: string, category: string, fileSize: number) =>
    http.post<{ data: { upload_url: string; file_key: string } }>(
      "/assets/upload/presigned",
      { filename, category, file_size: fileSize },
    ),
  completeUpload: (data: CompleteUploadParams) =>
    http.post<{ data: Asset }>("/assets/upload/complete", data),
  list: (page = 1, pageSize = 20, category?: string, format?: string) =>
    http.get("/assets", { params: { page, page_size: pageSize, category, format } }),
  listPublic: (page = 1, pageSize = 20, category?: string, search?: string) =>
    http.get("/assets/public", { params: { page, page_size: pageSize, category, search } }),
  get: (id: string) => http.get(`/assets/${id}`),
  update: (id: string, data: UpdateAssetParams) => http.put(`/assets/${id}`, data),
  delete: (id: string) => http.delete(`/assets/${id}`),
  getDownloadURL: (id: string) =>
    http.get<{ data: { url: string } }>(`/assets/${id}/download`),
  getThumbnailURL: (id: string) =>
    http.get<{ data: { url: string } }>(`/assets/${id}/thumbnail`),
};

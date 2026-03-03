import http from "./http";

export interface LoginParams {
  email: string;
  password: string;
  captcha_id?: string;
  captcha_answer?: string;
}

export interface RegisterParams {
  email: string;
  username: string;
  password: string;
  captcha_id: string;
  captcha_answer: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: UserInfo;
}

export interface UserInfo {
  id: string;
  email: string;
  username: string;
  display_name: string;
  avatar_url: string;
  role: string;
  status: string;
  created_at: string;
}

export const authApi = {
  login: (data: LoginParams) => http.post<{ data: AuthResponse }>("/auth/login", data),
  register: (data: RegisterParams) => http.post("/auth/register", data),
  logout: () => http.post("/auth/logout"),
  refresh: (refreshToken: string) =>
    http.post<{ data: AuthResponse }>("/auth/refresh", { refresh_token: refreshToken }),
  getCaptcha: () => http.get<{ data: { captcha_id: string; captcha_image: string } }>("/auth/captcha"),
  forgotPassword: (email: string) => http.post("/auth/forgot-password", { email }),
  resetPassword: (token: string, password: string) =>
    http.post("/auth/reset-password", { token, password }),
  getMe: () => http.get<{ data: UserInfo }>("/users/me"),
};

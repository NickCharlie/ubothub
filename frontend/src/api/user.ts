import http from "./http";
import type { UserInfo } from "./auth";

export interface UpdateProfileParams {
  display_name?: string;
  avatar_url?: string;
}

export interface ChangePasswordParams {
  old_password: string;
  new_password: string;
}

export const userApi = {
  getProfile: () => http.get<{ data: UserInfo }>("/users/me"),
  updateProfile: (data: UpdateProfileParams) =>
    http.put<{ data: UserInfo }>("/users/me", data),
  changePassword: (data: ChangePasswordParams) =>
    http.put("/users/me/password", data),
};

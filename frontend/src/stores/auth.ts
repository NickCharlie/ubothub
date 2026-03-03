import { create } from "zustand";
import { authApi, type UserInfo } from "@/api/auth";

interface AuthState {
  user: UserInfo | null;
  accessToken: string | null;
  refreshToken: string | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchUser: () => Promise<void>;
  setTokens: (access: string, refresh: string) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: sessionStorage.getItem("access_token"),
  refreshToken: sessionStorage.getItem("refresh_token"),
  loading: false,

  login: async (email, password) => {
    set({ loading: true });
    try {
      const res = await authApi.login({ email, password });
      const data = res.data.data;
      sessionStorage.setItem("access_token", data.access_token);
      sessionStorage.setItem("refresh_token", data.refresh_token);
      set({
        user: data.user,
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        loading: false,
      });
    } catch (err) {
      set({ loading: false });
      throw err;
    }
  },

  logout: async () => {
    try {
      await authApi.logout();
    } catch {
      // ignore
    }
    get().clear();
  },

  fetchUser: async () => {
    if (!get().accessToken) return;
    try {
      const res = await authApi.getMe();
      set({ user: res.data.data });
    } catch {
      get().clear();
    }
  },

  setTokens: (access, refresh) => {
    sessionStorage.setItem("access_token", access);
    sessionStorage.setItem("refresh_token", refresh);
    set({ accessToken: access, refreshToken: refresh });
  },

  clear: () => {
    sessionStorage.removeItem("access_token");
    sessionStorage.removeItem("refresh_token");
    set({ user: null, accessToken: null, refreshToken: null });
  },
}));

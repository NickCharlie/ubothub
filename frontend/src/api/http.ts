import axios from "axios";

const http = axios.create({
  baseURL: "/api/v1",
  timeout: 15000,
  headers: {
    "Content-Type": "application/json",
  },
});

http.interceptors.request.use((config) => {
  const token = sessionStorage.getItem("access_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      sessionStorage.removeItem("access_token");
      window.location.href = "/auth/login";
    }
    return Promise.reject(error);
  },
);

export default http;

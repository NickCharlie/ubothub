import { useEffect } from "react";
import { BrowserRouter } from "react-router-dom";
import { ConfigProvider, theme } from "antd";
import AppRouter from "./router";
import { useAuthStore } from "@/stores/auth";

function App() {
  const { accessToken, fetchUser } = useAuthStore();

  useEffect(() => {
    if (accessToken) {
      fetchUser();
    }
  }, [accessToken, fetchUser]);

  return (
    <ConfigProvider
      theme={{
        algorithm: theme.darkAlgorithm,
        token: {
          colorPrimary: "#3b82f6",
          borderRadius: 10,
          colorBgContainer: "rgba(255, 255, 255, 0.06)",
          colorBgElevated: "rgba(20, 22, 30, 0.95)",
          colorBorder: "rgba(255, 255, 255, 0.1)",
          colorText: "#ededed",
          colorTextSecondary: "rgba(255, 255, 255, 0.5)",
        },
      }}
    >
      <BrowserRouter>
        <AppRouter />
      </BrowserRouter>
    </ConfigProvider>
  );
}

export default App;

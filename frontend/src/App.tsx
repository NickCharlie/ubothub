import { BrowserRouter } from "react-router-dom";
import { ConfigProvider } from "antd";
import AppRouter from "./router";

function App() {
  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: "#1677ff",
          borderRadius: 6,
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

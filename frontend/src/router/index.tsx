import { Routes, Route, Navigate } from "react-router-dom";
import DashboardPage from "../pages/dashboard/DashboardPage";

function AppRouter() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      <Route path="/dashboard" element={<DashboardPage />} />
    </Routes>
  );
}

export default AppRouter;

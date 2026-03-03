import { Routes, Route, Navigate } from "react-router-dom";
import { useAuthStore } from "@/stores/auth";
import AppLayout from "@/components/layout/AppLayout";
import LoginPage from "@/pages/auth/LoginPage";
import RegisterPage from "@/pages/auth/RegisterPage";
import DashboardPage from "@/pages/dashboard/DashboardPage";
import BotListPage from "@/pages/bot/BotListPage";
import BotDetailPage from "@/pages/bot/BotDetailPage";
import AvatarListPage from "@/pages/avatar/AvatarListPage";
import AssetListPage from "@/pages/asset/AssetListPage";
import WalletPage from "@/pages/wallet/WalletPage";
import AdminPage from "@/pages/admin/AdminPage";
import ProfilePage from "@/pages/profile/ProfilePage";
import PlazaPage from "@/pages/plaza/PlazaPage";
import PlazaChatPage from "@/pages/plaza/PlazaChatPage";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { accessToken } = useAuthStore();
  if (!accessToken) {
    return <Navigate to="/auth/login" replace />;
  }
  return <>{children}</>;
}

function GuestRoute({ children }: { children: React.ReactNode }) {
  const { accessToken } = useAuthStore();
  if (accessToken) {
    return <Navigate to="/dashboard" replace />;
  }
  return <>{children}</>;
}

export default function AppRouter() {
  return (
    <Routes>
      {/* Guest routes */}
      <Route
        path="/auth/login"
        element={
          <GuestRoute>
            <LoginPage />
          </GuestRoute>
        }
      />
      <Route
        path="/auth/register"
        element={
          <GuestRoute>
            <RegisterPage />
          </GuestRoute>
        }
      />

      {/* Plaza routes (public, with sidebar layout) */}
      <Route element={<AppLayout />}>
        <Route path="/plaza" element={<PlazaPage />} />
        <Route path="/plaza/:id" element={<PlazaChatPage />} />
      </Route>

      {/* Protected routes with layout */}
      <Route
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/bots" element={<BotListPage />} />
        <Route path="/bots/:id" element={<BotDetailPage />} />
        <Route path="/avatars" element={<AvatarListPage />} />
        <Route path="/assets" element={<AssetListPage />} />
        <Route path="/wallet" element={<WalletPage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/admin" element={<AdminPage />} />
      </Route>

      {/* Root redirect */}
      <Route path="/" element={<Navigate to="/plaza" replace />} />
      <Route path="*" element={<Navigate to="/plaza" replace />} />
    </Routes>
  );
}

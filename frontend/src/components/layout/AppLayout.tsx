import { useState } from "react";
import { Outlet, useNavigate, useLocation } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  LayoutDashboard,
  Bot,
  Box,
  User as UserIcon,
  Wallet,
  Shield,
  LogOut,
  ChevronLeft,
  ChevronRight,
  Sparkles,
  UserCog,
} from "lucide-react";
import { useAuthStore } from "@/stores/auth";

interface NavItem {
  key: string;
  label: string;
  icon: React.ReactNode;
  path: string;
  adminOnly?: boolean;
}

const navItems: NavItem[] = [
  {
    key: "dashboard",
    label: "Dashboard",
    icon: <LayoutDashboard size={20} />,
    path: "/dashboard",
  },
  { key: "bots", label: "Bots", icon: <Bot size={20} />, path: "/bots" },
  {
    key: "avatars",
    label: "Avatars",
    icon: <Sparkles size={20} />,
    path: "/avatars",
  },
  { key: "assets", label: "Assets", icon: <Box size={20} />, path: "/assets" },
  {
    key: "wallet",
    label: "Wallet",
    icon: <Wallet size={20} />,
    path: "/wallet",
  },
  {
    key: "profile",
    label: "Profile",
    icon: <UserCog size={20} />,
    path: "/profile",
  },
  {
    key: "admin",
    label: "Admin",
    icon: <Shield size={20} />,
    path: "/admin",
    adminOnly: true,
  },
];

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();

  const filteredNavItems = navItems.filter(
    (item) => !item.adminOnly || user?.role === "admin",
  );

  const activeKey = filteredNavItems.find((item) =>
    location.pathname.startsWith(item.path),
  )?.key;

  const handleLogout = async () => {
    await logout();
    navigate("/auth/login");
  };

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Sidebar */}
      <motion.aside
        className="glass-sidebar flex flex-col h-full z-30 relative"
        animate={{ width: collapsed ? 64 : 240 }}
        transition={{ type: "spring", damping: 25, stiffness: 300 }}
      >
        {/* Logo */}
        <div className="flex items-center h-16 px-4 border-b border-white/[0.06]">
          <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center flex-shrink-0">
            <Bot size={18} className="text-white" />
          </div>
          <AnimatePresence>
            {!collapsed && (
              <motion.span
                initial={{ opacity: 0, width: 0 }}
                animate={{ opacity: 1, width: "auto" }}
                exit={{ opacity: 0, width: 0 }}
                className="ml-3 text-base font-semibold whitespace-nowrap overflow-hidden"
              >
                UBotHub
              </motion.span>
            )}
          </AnimatePresence>
        </div>

        {/* Navigation */}
        <nav className="flex-1 py-4 px-2 space-y-1 overflow-y-auto scrollbar-hide">
          {filteredNavItems.map((item) => {
            const isActive = activeKey === item.key;
            return (
              <button
                key={item.key}
                onClick={() => navigate(item.path)}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200 ${
                  isActive
                    ? "bg-white/10 text-white"
                    : "text-white/50 hover:text-white/80 hover:bg-white/[0.04]"
                }`}
              >
                <span className="flex-shrink-0">{item.icon}</span>
                <AnimatePresence>
                  {!collapsed && (
                    <motion.span
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      className="whitespace-nowrap overflow-hidden"
                    >
                      {item.label}
                    </motion.span>
                  )}
                </AnimatePresence>
              </button>
            );
          })}
        </nav>

        {/* User section */}
        <div className="p-3 border-t border-white/[0.06]">
          {user && !collapsed && (
            <div className="flex items-center gap-3 px-2 py-2 mb-2">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-green-400 to-emerald-600 flex items-center justify-center flex-shrink-0">
                <UserIcon size={14} className="text-white" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">
                  {user.display_name || user.username}
                </p>
                <p className="text-xs text-white/40 truncate">{user.email}</p>
              </div>
            </div>
          )}
          <div className="flex items-center gap-2">
            <button
              onClick={() => setCollapsed(!collapsed)}
              className="flex-1 flex items-center justify-center p-2 rounded-lg text-white/40 hover:text-white/70 hover:bg-white/[0.04] transition-colors"
            >
              {collapsed ? (
                <ChevronRight size={16} />
              ) : (
                <ChevronLeft size={16} />
              )}
            </button>
            <button
              onClick={handleLogout}
              className="flex items-center justify-center p-2 rounded-lg text-white/40 hover:text-red-400 hover:bg-white/[0.04] transition-colors"
              title="Logout"
            >
              <LogOut size={16} />
            </button>
          </div>
        </div>
      </motion.aside>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">
        <div className="p-6 md:p-8 max-w-[1400px] mx-auto">
          <Outlet />
        </div>
      </main>
    </div>
  );
}

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import {
  Bot,
  Users,
  Box,
  Activity,
  ArrowUpRight,
  TrendingUp,
} from "lucide-react";
import { botApi, adminApi } from "@/api/bot";
import { assetApi } from "@/api/asset";
import { useAuthStore } from "@/stores/auth";

interface StatCard {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  change?: string;
  color: string;
}

export default function DashboardPage() {
  const { user } = useAuthStore();
  const [stats, setStats] = useState<any>(null);

  useEffect(() => {
    if (user?.role === "admin") {
      adminApi
        .dashboard()
        .then((res) => setStats(res.data.data))
        .catch(() => {});
    } else if (user) {
      // For regular users, fetch their own bot and asset counts.
      Promise.all([
        botApi.list(1, 1).catch(() => null),
        assetApi.list(1, 1).catch(() => null),
      ]).then(([botsRes, assetsRes]) => {
        setStats({
          total_bots: botsRes?.data?.data?.total ?? 0,
          total_users: null,
          total_assets: assetsRes?.data?.data?.total ?? 0,
          messages_today: null,
        });
      });
    }
  }, [user]);

  const isAdmin = user?.role === "admin";

  const cards: StatCard[] = isAdmin
    ? [
        {
          title: "Total Bots",
          value: stats?.total_bots ?? "--",
          icon: <Bot size={20} />,
          change: stats ? `${stats.total_bots ?? 0}` : undefined,
          color: "from-blue-500 to-blue-600",
        },
        {
          title: "Active Users",
          value: stats?.total_users ?? "--",
          icon: <Users size={20} />,
          color: "from-emerald-500 to-emerald-600",
        },
        {
          title: "Assets",
          value: stats?.total_assets ?? "--",
          icon: <Box size={20} />,
          color: "from-purple-500 to-purple-600",
        },
        {
          title: "Messages Today",
          value: stats?.messages_today ?? "--",
          icon: <Activity size={20} />,
          color: "from-amber-500 to-orange-600",
        },
      ]
    : [
        {
          title: "My Bots",
          value: stats?.total_bots ?? "--",
          icon: <Bot size={20} />,
          color: "from-blue-500 to-blue-600",
        },
        {
          title: "My Assets",
          value: stats?.total_assets ?? "--",
          icon: <Box size={20} />,
          color: "from-purple-500 to-purple-600",
        },
      ];

  const containerVariants = {
    hidden: {},
    show: { transition: { staggerChildren: 0.08 } },
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 16 },
    show: {
      opacity: 1,
      y: 0,
      transition: { type: "spring" as const, damping: 25, stiffness: 300 },
    },
  };

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">
          Welcome back, {user?.display_name || user?.username || "User"}
        </h1>
        <p className="text-sm text-white/40 mt-1">
          Here's what's happening on your platform.
        </p>
      </div>

      <motion.div
        className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8"
        variants={containerVariants}
        initial="hidden"
        animate="show"
      >
        {cards.map((card) => (
          <motion.div
            key={card.title}
            variants={itemVariants}
            className="glass rounded-2xl p-5 hover:bg-white/[0.08] transition-colors cursor-default"
          >
            <div className="flex items-center justify-between mb-4">
              <div
                className={`w-10 h-10 rounded-xl bg-gradient-to-br ${card.color} flex items-center justify-center`}
              >
                <span className="text-white">{card.icon}</span>
              </div>
              {card.change && (
                <span className="flex items-center gap-1 text-xs text-emerald-400">
                  <TrendingUp size={12} />
                  {card.change}
                </span>
              )}
            </div>
            <p className="text-2xl font-semibold">{card.value}</p>
            <p className="text-xs text-white/40 mt-1">{card.title}</p>
          </motion.div>
        ))}
      </motion.div>

      <motion.div
        className="glass rounded-2xl p-6"
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.3, type: "spring", damping: 25, stiffness: 300 }}
      >
        <h2 className="text-lg font-medium mb-4">Quick Actions</h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {[
            { label: "Create Bot", path: "/bots", icon: <Bot size={18} /> },
            { label: "Upload Asset", path: "/assets", icon: <Box size={18} /> },
            {
              label: "API Docs",
              path: "/swagger/index.html",
              icon: <ArrowUpRight size={18} />,
            },
          ].map((action) => (
            <a
              key={action.label}
              href={action.path}
              className="flex items-center gap-3 p-4 rounded-xl border border-white/[0.06] hover:bg-white/[0.04] transition-colors"
            >
              <span className="text-white/50">{action.icon}</span>
              <span className="text-sm">{action.label}</span>
            </a>
          ))}
        </div>
      </motion.div>
    </div>
  );
}

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import {
  Shield,
  Users,
  Bot,
  Box,
  Activity,
  Ban,
  CheckCircle,
  Trash2,
  Wifi,
} from "lucide-react";
import { adminApi } from "@/api/bot";
import { useAuthStore } from "@/stores/auth";
import { Modal, Tabs, Pagination, message } from "antd";

interface DashboardStats {
  total_users: number;
  total_bots: number;
  active_bots: number;
  total_assets: number;
  total_storage_bytes: number;
}

interface AdminUser {
  id: string;
  email: string;
  username: string;
  display_name: string;
  role: string;
  status: string;
  email_verified: boolean;
  created_at: string;
}

interface AdminBot {
  id: string;
  name: string;
  framework: string;
  status: string;
  visibility: string;
  user_id: string;
  created_at: string;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024)
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

export default function AdminPage() {
  const { user } = useAuthStore();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [userTotal, setUserTotal] = useState(0);
  const [userPage, setUserPage] = useState(1);
  const [bots, setBots] = useState<AdminBot[]>([]);
  const [botTotal, setBotTotal] = useState(0);
  const [botPage, setBotPage] = useState(1);

  useEffect(() => {
    if (user?.role !== "admin") return;
    adminApi
      .dashboard()
      .then((res) => setStats(res.data.data))
      .catch(() => {});
  }, [user]);

  const loadUsers = async () => {
    try {
      const res = await adminApi.listUsers(userPage, 20);
      setUsers(res.data.data?.items || []);
      setUserTotal(res.data.data?.total || 0);
    } catch {
      // ignore
    }
  };

  const loadBots = async () => {
    try {
      const res = await adminApi.listBots(botPage, 20);
      setBots(res.data.data?.items || []);
      setBotTotal(res.data.data?.total || 0);
    } catch {
      // ignore
    }
  };

  useEffect(() => {
    loadUsers();
  }, [userPage]);

  useEffect(() => {
    loadBots();
  }, [botPage]);

  const handleBan = (userId: string) => {
    Modal.confirm({
      title: "Ban User",
      content: "This user will no longer be able to access the platform.",
      okText: "Ban",
      okButtonProps: { danger: true },
      onOk: async () => {
        await adminApi.banUser(userId);
        message.success("User banned");
        loadUsers();
      },
    });
  };

  const handleUnban = async (userId: string) => {
    await adminApi.unbanUser(userId);
    message.success("User unbanned");
    loadUsers();
  };

  const handleForceDeleteBot = (botId: string) => {
    Modal.confirm({
      title: "Force Delete Bot",
      content: "This will permanently delete this bot and all associated data.",
      okText: "Delete",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await adminApi.forceDeleteBot(botId);
          message.success("Bot deleted");
          loadBots();
        } catch {
          message.error("Failed to delete bot");
        }
      },
    });
  };

  if (user?.role !== "admin") {
    return (
      <div className="flex items-center justify-center h-64 text-white/30">
        <Shield size={40} className="mr-3 opacity-30" />
        <p>Admin access required</p>
      </div>
    );
  }

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

  const statCards = [
    {
      title: "Users",
      value: stats?.total_users ?? "--",
      icon: <Users size={18} />,
      color: "from-blue-500 to-blue-600",
    },
    {
      title: "Total Bots",
      value: stats?.total_bots ?? "--",
      icon: <Bot size={18} />,
      color: "from-emerald-500 to-emerald-600",
    },
    {
      title: "Active Bots",
      value: stats?.active_bots ?? "--",
      icon: <Activity size={18} />,
      color: "from-purple-500 to-purple-600",
    },
    {
      title: "Assets",
      value: stats?.total_assets ?? "--",
      icon: <Box size={18} />,
      color: "from-amber-500 to-orange-600",
    },
  ];

  const tabItems = [
    {
      key: "users",
      label: "Users",
      children: (
        <div>
          <div className="space-y-1">
            {users.map((u) => (
              <div
                key={u.id}
                className="flex items-center gap-4 p-3 rounded-xl hover:bg-white/[0.04] transition-colors"
              >
                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500/30 to-purple-500/30 flex items-center justify-center flex-shrink-0">
                  <Users size={14} className="text-white/60" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm truncate">
                    {u.display_name || u.username}
                    <span className="text-white/30 ml-2 text-xs">
                      @{u.username}
                    </span>
                  </p>
                  <p className="text-xs text-white/30">{u.email}</p>
                </div>
                <div className="flex items-center gap-2 flex-shrink-0">
                  {u.role === "admin" && (
                    <span className="text-[10px] px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-400 border border-purple-500/20">
                      admin
                    </span>
                  )}
                  <span
                    className={`text-[10px] px-2 py-0.5 rounded-full border ${
                      u.status === "active"
                        ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                        : "bg-red-500/10 text-red-400 border-red-500/20"
                    }`}
                  >
                    {u.status}
                  </span>
                  {u.status === "active" && u.role !== "admin" ? (
                    <button
                      onClick={() => handleBan(u.id)}
                      className="p-1.5 rounded-lg hover:bg-red-500/10 text-white/30 hover:text-red-400 transition-colors"
                      title="Ban user"
                    >
                      <Ban size={14} />
                    </button>
                  ) : u.status === "banned" ? (
                    <button
                      onClick={() => handleUnban(u.id)}
                      className="p-1.5 rounded-lg hover:bg-emerald-500/10 text-white/30 hover:text-emerald-400 transition-colors"
                      title="Unban user"
                    >
                      <CheckCircle size={14} />
                    </button>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
          {userTotal > 20 && (
            <div className="flex justify-center mt-4">
              <Pagination
                current={userPage}
                total={userTotal}
                pageSize={20}
                onChange={setUserPage}
                size="small"
              />
            </div>
          )}
        </div>
      ),
    },
    {
      key: "bots",
      label: "Bots",
      children: (
        <div>
          <div className="space-y-1">
            {bots.map((bot) => (
              <div
                key={bot.id}
                className="flex items-center gap-4 p-3 rounded-xl hover:bg-white/[0.04] transition-colors"
              >
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center flex-shrink-0">
                  <Bot size={14} className="text-blue-400" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm truncate">{bot.name}</p>
                  <p className="text-xs text-white/30">
                    {bot.framework} ·{" "}
                    {new Date(bot.created_at).toLocaleDateString()}
                  </p>
                </div>
                <div className="flex items-center gap-2 flex-shrink-0">
                  <span
                    className={`text-[10px] px-2 py-0.5 rounded-full border ${
                      bot.status === "online"
                        ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                        : "bg-white/[0.06] text-white/40 border-white/[0.08]"
                    }`}
                  >
                    {bot.status}
                  </span>
                  <button
                    onClick={() => handleForceDeleteBot(bot.id)}
                    className="p-1.5 rounded-lg hover:bg-red-500/10 text-white/30 hover:text-red-400 transition-colors"
                    title="Force delete"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>
            ))}
          </div>
          {botTotal > 20 && (
            <div className="flex justify-center mt-4">
              <Pagination
                current={botPage}
                total={botTotal}
                pageSize={20}
                onChange={setBotPage}
                size="small"
              />
            </div>
          )}
        </div>
      ),
    },
  ];

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Admin Console</h1>

      {/* Stats */}
      <motion.div
        className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8"
        variants={containerVariants}
        initial="hidden"
        animate="show"
      >
        {statCards.map((card) => (
          <motion.div
            key={card.title}
            variants={itemVariants}
            className="glass rounded-2xl p-5"
          >
            <div className="flex items-center gap-3 mb-3">
              <div
                className={`w-10 h-10 rounded-xl bg-gradient-to-br ${card.color} flex items-center justify-center`}
              >
                <span className="text-white">{card.icon}</span>
              </div>
            </div>
            <p className="text-2xl font-semibold">{card.value}</p>
            <p className="text-xs text-white/40 mt-1">{card.title}</p>
          </motion.div>
        ))}
      </motion.div>

      {/* Storage info */}
      {stats && (
        <div className="glass rounded-2xl p-5 mb-6">
          <div className="flex items-center gap-3">
            <Wifi size={16} className="text-white/40" />
            <span className="text-sm text-white/50">
              Total Storage Usage: {formatBytes(stats.total_storage_bytes || 0)}
            </span>
          </div>
        </div>
      )}

      {/* Management tabs */}
      <div className="glass rounded-2xl p-6">
        <Tabs items={tabItems} />
      </div>
    </div>
  );
}

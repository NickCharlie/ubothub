import { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import {
  Bot,
  Search,
  Wifi,
  WifiOff,
  MessageSquare,
  Clock,
} from "lucide-react";
import { plazaApi, type PlazaBot } from "@/api/plaza";

export default function PlazaPage() {
  const navigate = useNavigate();
  const [bots, setBots] = useState<PlazaBot[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");

  const loadBots = useCallback(async () => {
    setLoading(true);
    try {
      const res = await plazaApi.listBots(page, 24);
      setBots(res.data.data?.items || []);
      setTotal(res.data.data?.total || 0);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    loadBots();
  }, [loadBots]);

  const filteredBots = bots.filter(
    (b) =>
      !search ||
      b.name.toLowerCase().includes(search.toLowerCase()) ||
      b.framework.toLowerCase().includes(search.toLowerCase()),
  );

  const containerVariants = {
    hidden: {},
    show: { transition: { staggerChildren: 0.04 } },
  };

  const cardVariants = {
    hidden: { opacity: 0, y: 20, scale: 0.95 },
    show: {
      opacity: 1,
      y: 0,
      scale: 1,
      transition: { type: "spring" as const, damping: 22, stiffness: 300 },
    },
  };

  const frameworkGradients: Record<string, string> = {
    astrbot: "from-blue-500/30 to-cyan-500/30",
    openai: "from-emerald-500/30 to-green-500/30",
    langchain: "from-purple-500/30 to-pink-500/30",
    custom: "from-amber-500/30 to-orange-500/30",
  };

  const getGradient = (framework: string) =>
    frameworkGradients[framework.toLowerCase()] ||
    "from-blue-500/30 to-purple-500/30";

  return (
    <div className="min-h-screen">
      {/* Hero header */}
      <motion.div
        className="text-center mb-10"
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ type: "spring", damping: 22, stiffness: 300 }}
      >
        <h1 className="text-3xl md:text-4xl font-bold mb-3 bg-gradient-to-r from-blue-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
          Bot Plaza
        </h1>
        <p className="text-white/40 text-sm md:text-base max-w-md mx-auto">
          Discover and chat with AI bots from the community
        </p>
      </motion.div>

      {/* Search bar */}
      <motion.div
        className="max-w-xl mx-auto mb-8"
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{
          type: "spring",
          damping: 22,
          stiffness: 300,
          delay: 0.08,
        }}
      >
        <div className="relative">
          <Search
            size={18}
            className="absolute left-4 top-1/2 -translate-y-1/2 text-white/30"
          />
          <input
            type="text"
            placeholder="Search bots by name or framework..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="glass-input w-full h-12 rounded-2xl pl-12 pr-4 text-sm"
          />
        </div>
      </motion.div>

      {/* Bot grid */}
      {loading ? (
        <div className="flex items-center justify-center h-48 text-white/30">
          <div className="flex items-center gap-3">
            <div className="w-5 h-5 border-2 border-white/20 border-t-blue-400 rounded-full animate-spin" />
            Loading bots...
          </div>
        </div>
      ) : filteredBots.length === 0 ? (
        <motion.div
          className="flex flex-col items-center justify-center h-48 text-white/30"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
        >
          <Bot size={48} className="mb-3 opacity-30" />
          <p>
            {search
              ? "No bots found matching your search"
              : "No public bots available yet"}
          </p>
        </motion.div>
      ) : (
        <>
          <motion.div
            className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4"
            variants={containerVariants}
            initial="hidden"
            animate="show"
          >
            {filteredBots.map((bot) => (
              <motion.div
                key={bot.id}
                variants={cardVariants}
                whileHover={{ y: -4, transition: { duration: 0.2 } }}
                onClick={() => navigate(`/plaza/${bot.id}`)}
                className="glass rounded-2xl p-5 cursor-pointer group hover:border-white/20 transition-colors"
              >
                {/* Bot icon + status */}
                <div className="flex items-start justify-between mb-4">
                  <div
                    className={`w-12 h-12 rounded-2xl bg-gradient-to-br ${getGradient(bot.framework)} border border-white/[0.08] flex items-center justify-center`}
                  >
                    <Bot
                      size={22}
                      className="text-white/80 group-hover:text-white transition-colors"
                    />
                  </div>
                  <div className="flex items-center gap-1.5">
                    {bot.status === "online" ? (
                      <Wifi size={12} className="text-emerald-400" />
                    ) : (
                      <WifiOff size={12} className="text-white/20" />
                    )}
                    <span
                      className={`text-[10px] ${bot.status === "online" ? "text-emerald-400" : "text-white/30"}`}
                    >
                      {bot.status}
                    </span>
                  </div>
                </div>

                {/* Bot info */}
                <h3 className="font-semibold text-sm mb-1 truncate group-hover:text-white transition-colors">
                  {bot.name}
                </h3>
                <p className="text-xs text-white/40 line-clamp-2 mb-4 min-h-[2.5rem]">
                  {bot.description || "No description provided"}
                </p>

                {/* Footer */}
                <div className="flex items-center justify-between">
                  <span className="text-[10px] px-2.5 py-1 rounded-full bg-white/[0.06] text-white/50 border border-white/[0.08]">
                    {bot.framework}
                  </span>
                  <div className="flex items-center gap-3 text-white/30">
                    {bot.last_active_at && (
                      <span className="flex items-center gap-1 text-[10px]">
                        <Clock size={10} />
                        {new Date(bot.last_active_at).toLocaleDateString()}
                      </span>
                    )}
                    <span className="flex items-center gap-1 text-[10px] opacity-0 group-hover:opacity-100 transition-opacity text-blue-400">
                      <MessageSquare size={10} />
                      Chat
                    </span>
                  </div>
                </div>
              </motion.div>
            ))}
          </motion.div>

          {/* Load more */}
          {total > page * 24 && (
            <div className="flex justify-center mt-8">
              <button
                onClick={() => setPage((p) => p + 1)}
                className="glass-btn-ghost h-10 px-6 rounded-xl text-sm"
              >
                Load More
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

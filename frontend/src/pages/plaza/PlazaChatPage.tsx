import {
  useEffect,
  useState,
  useRef,
  useCallback,
  lazy,
  Suspense,
} from "react";
import { useParams, useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  ArrowLeft,
  Bot,
  Send,
  Wifi,
  WifiOff,
  Sparkles,
  Info,
} from "lucide-react";
import { plazaApi, type PlazaBot, type PlazaAvatar } from "@/api/plaza";
import { useAuthStore } from "@/stores/auth";
import type { ModelAdapter } from "@/components/avatar/model/types";

// Lazy-load the 3D viewer to avoid loading Three.js on pages that don't need it
const AvatarViewer = lazy(() => import("@/components/avatar/AvatarViewer"));

interface ChatMessage {
  id: string;
  role: "user" | "bot" | "system";
  content: string;
  timestamp: number;
}

export default function PlazaChatPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { accessToken } = useAuthStore();

  const [bot, setBot] = useState<PlazaBot | null>(null);
  const [avatar, setAvatar] = useState<PlazaAvatar | null>(null);
  const [avatarModelUrl, setAvatarModelUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputMsg, setInputMsg] = useState("");
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [wsConnected, setWsConnected] = useState(false);
  const [showInfo, setShowInfo] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const adapterRef = useRef<ModelAdapter | null>(null);

  // Auto-scroll to latest message
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // Load bot details
  const loadBot = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await plazaApi.getBot(id);
      const botData = res.data.data;
      setBot(botData);

      // Try to load avatar if bot has one configured
      if (botData.config) {
        try {
          const config = JSON.parse(botData.config);
          if (config.avatar_id) {
            const avatarRes = await plazaApi.getAvatar(config.avatar_id);
            const avatarData = avatarRes.data.data;
            setAvatar(avatarData);

            // Find the model asset (role=primary_model for the main 3D model)
            if (avatarData.avatar_assets) {
              const bodyAsset = avatarData.avatar_assets.find(
                (a: { role: string }) =>
                  a.role === "primary_model" ||
                  a.role === "body" ||
                  a.role === "model",
              );
              if (bodyAsset) {
                // Fetch presigned download URL from the public plaza endpoint
                const dlRes = await plazaApi.getAssetDownloadURL(
                  bodyAsset.asset_id,
                );
                const downloadUrl = dlRes.data.data.download_url;
                setAvatarModelUrl(downloadUrl);
              }
            }
          }
        } catch {
          // config might not have avatar_id
        }
      }
    } catch {
      navigate("/plaza");
    } finally {
      setLoading(false);
    }
  }, [id, navigate]);

  useEffect(() => {
    loadBot();
  }, [loadBot]);

  // WebSocket connection
  useEffect(() => {
    if (!id || !accessToken) return;

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/api/v1/ws?token=${accessToken}&bot_id=${id}`;
    const socket = new WebSocket(wsUrl);

    socket.onopen = () => {
      setWsConnected(true);
      setMessages((prev) => [
        ...prev,
        {
          id: `sys-${Date.now()}`,
          role: "system",
          content: "Connected to real-time channel",
          timestamp: Date.now(),
        },
      ]);
    };

    socket.onmessage = async (event) => {
      try {
        const data = JSON.parse(event.data);

        if (data.type === "bot_reply" || data.type === "chat") {
          setMessages((prev) => [
            ...prev,
            {
              id: data.id || `msg-${Date.now()}`,
              role: "bot",
              content: data.content || data.text || "",
              timestamp: data.timestamp || Date.now(),
            },
          ]);

          // Trigger expression on bot reply
          if (adapterRef.current && data.data?.emotion) {
            const { playExpression } =
              await import("@/components/avatar/AvatarViewer");
            playExpression(adapterRef.current, data.data.emotion);
          }
        } else if (data.type === "avatar_action") {
          // Handle avatar action from WebSocket
          if (adapterRef.current && data.data?.animation_url) {
            const { playAnimation } =
              await import("@/components/avatar/AvatarViewer");
            playAnimation(
              adapterRef.current,
              data.data.animation_url,
              data.data.file_type || "vrma",
              data.data.loop || false,
            );
          }
          if (adapterRef.current && data.data?.expression) {
            const { playExpression } =
              await import("@/components/avatar/AvatarViewer");
            playExpression(adapterRef.current, data.data.expression);
          }
        } else if (data.type === "error") {
          setMessages((prev) => [
            ...prev,
            {
              id: `err-${Date.now()}`,
              role: "system",
              content: data.content || "An error occurred",
              timestamp: Date.now(),
            },
          ]);
        }
      } catch {
        setMessages((prev) => [
          ...prev,
          {
            id: `msg-${Date.now()}`,
            role: "bot",
            content: event.data,
            timestamp: Date.now(),
          },
        ]);
      }
    };

    socket.onclose = () => {
      setWsConnected(false);
    };

    socket.onerror = () => {
      setWsConnected(false);
    };

    setWs(socket);
    return () => {
      socket.close();
    };
  }, [id, accessToken]);

  const sendMessage = () => {
    if (!inputMsg.trim() || !ws || ws.readyState !== WebSocket.OPEN) return;

    const msg: ChatMessage = {
      id: `user-${Date.now()}`,
      role: "user",
      content: inputMsg.trim(),
      timestamp: Date.now(),
    };
    setMessages((prev) => [...prev, msg]);
    ws.send(
      JSON.stringify({ type: "chat", content: inputMsg.trim(), bot_id: id }),
    );
    setInputMsg("");
    inputRef.current?.focus();
  };

  const handleAdapterReady = useCallback((adapter: ModelAdapter) => {
    adapterRef.current = adapter;
  }, []);

  const messageVariants = {
    hidden: { opacity: 0, y: 8, scale: 0.96 },
    show: {
      opacity: 1,
      y: 0,
      scale: 1,
      transition: { type: "spring" as const, damping: 22, stiffness: 300 },
    },
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-[80vh] text-white/30">
        <div className="flex items-center gap-3">
          <div className="w-5 h-5 border-2 border-white/20 border-t-blue-400 rounded-full animate-spin" />
          Loading...
        </div>
      </div>
    );
  }

  if (!bot) return null;

  const hasAvatar = avatar && avatarModelUrl;

  return (
    <div className="flex flex-col h-[calc(100vh-3rem)] max-h-[calc(100vh-3rem)]">
      {/* Chat header */}
      <motion.div
        className="glass-strong rounded-2xl p-4 mb-4 flex-shrink-0"
        initial={{ opacity: 0, y: -16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ type: "spring", damping: 22, stiffness: 300 }}
      >
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate("/plaza")}
            className="p-2 rounded-xl hover:bg-white/[0.06] text-white/40 hover:text-white/70 transition-colors"
          >
            <ArrowLeft size={18} />
          </button>

          <div className="w-10 h-10 rounded-2xl bg-gradient-to-br from-blue-500/30 to-purple-500/30 border border-white/[0.08] flex items-center justify-center flex-shrink-0">
            <Bot size={18} className="text-blue-400" />
          </div>

          <div className="flex-1 min-w-0">
            <h2 className="font-semibold text-sm truncate">{bot.name}</h2>
            <div className="flex items-center gap-2">
              <span className="text-[10px] text-white/40">{bot.framework}</span>
              <span className="flex items-center gap-1">
                {wsConnected ? (
                  <>
                    <Wifi size={10} className="text-emerald-400" />
                    <span className="text-[10px] text-emerald-400">
                      Connected
                    </span>
                  </>
                ) : (
                  <>
                    <WifiOff size={10} className="text-red-400/60" />
                    <span className="text-[10px] text-red-400/60">
                      {accessToken ? "Disconnected" : "Login required"}
                    </span>
                  </>
                )}
              </span>
            </div>
          </div>

          {avatar && (
            <div className="flex items-center gap-2 mr-2">
              <Sparkles size={14} className="text-purple-400" />
              <span className="text-xs text-white/40">{avatar.name}</span>
            </div>
          )}

          <button
            onClick={() => setShowInfo(!showInfo)}
            className={`p-2 rounded-xl transition-colors ${showInfo ? "bg-white/[0.08] text-white/70" : "hover:bg-white/[0.06] text-white/40 hover:text-white/70"}`}
          >
            <Info size={16} />
          </button>
        </div>

        {/* Bot info panel */}
        <AnimatePresence>
          {showInfo && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: "auto", opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ type: "spring", damping: 22, stiffness: 300 }}
              className="overflow-hidden"
            >
              <div className="pt-4 mt-4 border-t border-white/[0.06]">
                <p className="text-xs text-white/50 leading-relaxed">
                  {bot.description || "No description provided."}
                </p>
                <div className="flex items-center gap-2 mt-3">
                  <span className="text-[10px] px-2.5 py-1 rounded-full bg-white/[0.06] text-white/50 border border-white/[0.08]">
                    {bot.framework}
                  </span>
                  <span
                    className={`text-[10px] px-2.5 py-1 rounded-full border ${bot.status === "online" ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" : "bg-white/[0.06] text-white/40 border-white/[0.08]"}`}
                  >
                    {bot.status}
                  </span>
                  {avatar && (
                    <span className="text-[10px] px-2.5 py-1 rounded-full bg-purple-500/10 text-purple-400 border border-purple-500/20">
                      {avatar.render_type} avatar
                    </span>
                  )}
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>

      {/* Main content: Avatar + Chat */}
      <div
        className={`flex-1 flex min-h-0 gap-4 ${hasAvatar ? "flex-row" : "flex-col"}`}
      >
        {/* 3D Avatar panel — transparent, blends into background */}
        {hasAvatar && (
          <motion.div
            className="flex-shrink-0 hidden lg:block pointer-events-none"
            style={{ width: 360, height: "100%" }}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ type: "spring", damping: 22, stiffness: 300 }}
          >
            <Suspense
              fallback={
                <div className="flex items-center justify-center h-full text-white/30">
                  <div className="w-5 h-5 border-2 border-white/20 border-t-purple-400 rounded-full animate-spin" />
                </div>
              }
            >
              <AvatarViewer
                modelUrl={avatarModelUrl}
                onAdapterReady={handleAdapterReady}
                cameraPosition={[0, 1.2, 2.0]}
                cameraTarget={[0, 1.0, 0]}
                transparent
              />
            </Suspense>
          </motion.div>
        )}

        {/* Chat column */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Messages area */}
          <div className="flex-1 overflow-y-auto min-h-0 px-2">
            <div className="space-y-3 py-4">
              {messages.length === 0 ? (
                <motion.div
                  className="flex flex-col items-center justify-center h-full min-h-[300px] text-white/20"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ delay: 0.3 }}
                >
                  <div className="w-16 h-16 rounded-3xl bg-gradient-to-br from-blue-500/10 to-purple-500/10 border border-white/[0.06] flex items-center justify-center mb-4">
                    <Bot size={28} className="text-white/20" />
                  </div>
                  <p className="text-sm mb-1">
                    Start a conversation with {bot.name}
                  </p>
                  <p className="text-xs text-white/15">
                    {accessToken
                      ? "Type a message below to begin"
                      : "Please login first to chat"}
                  </p>
                </motion.div>
              ) : (
                <AnimatePresence initial={false}>
                  {messages.map((msg) => (
                    <motion.div
                      key={msg.id}
                      variants={messageVariants}
                      initial="hidden"
                      animate="show"
                      className={`flex ${msg.role === "user" ? "justify-end" : msg.role === "system" ? "justify-center" : "justify-start"}`}
                    >
                      {msg.role === "system" ? (
                        <span className="text-[11px] text-white/25 px-3 py-1 bg-white/[0.03] rounded-full">
                          {msg.content}
                        </span>
                      ) : (
                        <div className="flex items-end gap-2 max-w-[75%]">
                          {msg.role === "bot" && (
                            <div className="w-7 h-7 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-white/[0.06] flex items-center justify-center flex-shrink-0 mb-0.5">
                              <Bot size={12} className="text-blue-400" />
                            </div>
                          )}
                          <div
                            className={`px-4 py-2.5 rounded-2xl text-sm leading-relaxed ${
                              msg.role === "user"
                                ? "bg-blue-500/20 border border-blue-500/25 text-blue-50 rounded-br-md"
                                : "glass border border-white/[0.08] text-white/80 rounded-bl-md"
                            }`}
                          >
                            <p className="whitespace-pre-wrap break-words">
                              {msg.content}
                            </p>
                            <span className="text-[9px] text-white/20 mt-1 block text-right">
                              {new Date(msg.timestamp).toLocaleTimeString([], {
                                hour: "2-digit",
                                minute: "2-digit",
                              })}
                            </span>
                          </div>
                        </div>
                      )}
                    </motion.div>
                  ))}
                </AnimatePresence>
              )}
              <div ref={messagesEndRef} />
            </div>
          </div>

          {/* Input area */}
          <motion.div
            className="glass-strong rounded-2xl p-3 mt-3 flex-shrink-0"
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{
              type: "spring",
              damping: 22,
              stiffness: 300,
              delay: 0.1,
            }}
          >
            {!accessToken ? (
              <div className="flex items-center justify-center gap-3 py-2">
                <span className="text-sm text-white/40">
                  Please login to start chatting
                </span>
                <button
                  onClick={() => navigate("/auth/login")}
                  className="glass-btn h-8 px-4 rounded-lg text-xs"
                >
                  Login
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-3">
                <input
                  ref={inputRef}
                  type="text"
                  value={inputMsg}
                  onChange={(e) => setInputMsg(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && !e.shiftKey) {
                      e.preventDefault();
                      sendMessage();
                    }
                  }}
                  placeholder={
                    wsConnected ? "Type a message..." : "Connecting..."
                  }
                  disabled={!wsConnected}
                  className="glass-input flex-1 h-10 rounded-xl px-4 text-sm disabled:opacity-40"
                />
                <button
                  onClick={sendMessage}
                  disabled={!wsConnected || !inputMsg.trim()}
                  className="glass-btn h-10 w-10 rounded-xl p-0 disabled:opacity-30 disabled:hover:transform-none flex-shrink-0"
                >
                  <Send size={16} />
                </button>
              </div>
            )}
          </motion.div>
        </div>
      </div>
    </div>
  );
}

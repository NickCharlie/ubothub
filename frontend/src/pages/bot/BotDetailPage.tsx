import { useEffect, useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import {
  Bot,
  ArrowLeft,
  Copy,
  RefreshCw,
  Settings,
  Send,
  Wifi,
  WifiOff,
  Trash2,
  Eye,
  EyeOff,
} from "lucide-react";
import { botApi, type Bot as BotType, type UpdateBotParams } from "@/api/bot";
import { useAuthStore } from "@/stores/auth";
import { Modal, Form, Input, Switch, message, Tabs } from "antd";

export default function BotDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { accessToken } = useAuthStore();
  const [bot, setBot] = useState<BotType | null>(null);
  const [loading, setLoading] = useState(true);
  const [showToken, setShowToken] = useState(false);
  const [editing, setEditing] = useState(false);
  const [form] = Form.useForm();

  // Chat state
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputMsg, setInputMsg] = useState("");
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [wsConnected, setWsConnected] = useState(false);

  interface ChatMessage {
    id: string;
    role: "user" | "bot" | "system";
    content: string;
    timestamp: number;
  }

  const loadBot = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await botApi.get(id);
      setBot(res.data.data);
    } catch {
      message.error("Failed to load bot");
      navigate("/bots");
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

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setMessages((prev) => [
          ...prev,
          {
            id: data.id || `msg-${Date.now()}`,
            role: "bot",
            content: data.content || data.text || JSON.stringify(data),
            timestamp: Date.now(),
          },
        ]);
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
      JSON.stringify({ type: "message", content: inputMsg.trim(), bot_id: id }),
    );
    setInputMsg("");
  };

  const handleUpdate = async () => {
    if (!bot) return;
    try {
      const values = await form.validateFields();
      const params: UpdateBotParams = {
        name: values.name,
        description: values.description,
        webhook_url: values.webhook_url,
        is_public: values.is_public,
      };
      await botApi.update(bot.id, params);
      message.success("Bot updated");
      setEditing(false);
      loadBot();
    } catch {
      // validation error
    }
  };

  const handleRegenerateToken = () => {
    if (!bot) return;
    Modal.confirm({
      title: "Regenerate Access Token",
      content:
        "The current token will be invalidated. All existing integrations will stop working.",
      okText: "Regenerate",
      okButtonProps: { danger: true },
      onOk: async () => {
        await botApi.regenerateToken(bot.id);
        message.success("Token regenerated");
        loadBot();
      },
    });
  };

  const handleDelete = () => {
    if (!bot) return;
    Modal.confirm({
      title: "Delete Bot",
      content:
        "This action cannot be undone. All data will be permanently removed.",
      okText: "Delete",
      okButtonProps: { danger: true },
      onOk: async () => {
        await botApi.delete(bot.id);
        message.success("Bot deleted");
        navigate("/bots");
      },
    });
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success("Copied to clipboard");
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-white/30">
        Loading...
      </div>
    );
  }

  if (!bot) return null;

  const tabItems = [
    {
      key: "chat",
      label: "Chat",
      children: (
        <div className="flex flex-col h-[500px]">
          {/* Messages area */}
          <div className="flex-1 overflow-y-auto space-y-3 mb-4 p-4 glass rounded-xl">
            {messages.length === 0 ? (
              <div className="flex items-center justify-center h-full text-white/30 text-sm">
                Send a message to start chatting with {bot.name}
              </div>
            ) : (
              messages.map((msg) => (
                <div
                  key={msg.id}
                  className={`flex ${msg.role === "user" ? "justify-end" : msg.role === "system" ? "justify-center" : "justify-start"}`}
                >
                  {msg.role === "system" ? (
                    <span className="text-xs text-white/30 px-3 py-1">
                      {msg.content}
                    </span>
                  ) : (
                    <div
                      className={`max-w-[70%] px-4 py-2.5 rounded-2xl text-sm ${
                        msg.role === "user"
                          ? "bg-blue-500/20 border border-blue-500/30 text-blue-100"
                          : "bg-white/[0.06] border border-white/[0.08] text-white/80"
                      }`}
                    >
                      {msg.content}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
          {/* Input area */}
          <div className="flex gap-3">
            <div className="flex items-center gap-2 mr-1">
              {wsConnected ? (
                <Wifi size={14} className="text-emerald-400" />
              ) : (
                <WifiOff size={14} className="text-red-400" />
              )}
            </div>
            <input
              type="text"
              value={inputMsg}
              onChange={(e) => setInputMsg(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && sendMessage()}
              placeholder={wsConnected ? "Type a message..." : "Disconnected"}
              disabled={!wsConnected}
              className="glass-input flex-1 h-10 rounded-xl px-4 text-sm"
            />
            <button
              onClick={sendMessage}
              disabled={!wsConnected || !inputMsg.trim()}
              className="glass-btn h-10 w-10 rounded-xl p-0 disabled:opacity-30"
            >
              <Send size={16} />
            </button>
          </div>
        </div>
      ),
    },
    {
      key: "settings",
      label: "Settings",
      children: (
        <div className="space-y-6">
          {/* Access Token */}
          <div className="glass rounded-xl p-5">
            <h3 className="text-sm font-medium mb-3 text-white/70">
              Access Token
            </h3>
            <div className="flex items-center gap-2">
              <code className="flex-1 text-xs bg-black/30 px-3 py-2 rounded-lg font-mono overflow-hidden text-white/60">
                {showToken
                  ? (bot as any).access_token || "••••••••"
                  : "••••••••••••••••••••••••"}
              </code>
              <button
                onClick={() => setShowToken(!showToken)}
                className="p-2 rounded-lg hover:bg-white/[0.06] text-white/40 hover:text-white/70 transition-colors"
              >
                {showToken ? <EyeOff size={14} /> : <Eye size={14} />}
              </button>
              <button
                onClick={() => copyToClipboard((bot as any).access_token || "")}
                className="p-2 rounded-lg hover:bg-white/[0.06] text-white/40 hover:text-white/70 transition-colors"
              >
                <Copy size={14} />
              </button>
              <button
                onClick={handleRegenerateToken}
                className="p-2 rounded-lg hover:bg-white/[0.06] text-white/40 hover:text-white/70 transition-colors"
              >
                <RefreshCw size={14} />
              </button>
            </div>
          </div>

          {/* Webhook URL */}
          <div className="glass rounded-xl p-5">
            <h3 className="text-sm font-medium mb-3 text-white/70">
              Webhook Endpoint
            </h3>
            <div className="flex items-center gap-2">
              <code className="flex-1 text-xs bg-black/30 px-3 py-2 rounded-lg font-mono text-white/60 overflow-x-auto">
                {`${window.location.origin}/api/v1/gateway/webhook/${(bot as any).access_token || "<token>"}`}
              </code>
              <button
                onClick={() =>
                  copyToClipboard(
                    `${window.location.origin}/api/v1/gateway/webhook/${(bot as any).access_token || ""}`,
                  )
                }
                className="p-2 rounded-lg hover:bg-white/[0.06] text-white/40 hover:text-white/70 transition-colors"
              >
                <Copy size={14} />
              </button>
            </div>
          </div>

          {/* Danger zone */}
          <div className="glass rounded-xl p-5 border-red-500/20">
            <h3 className="text-sm font-medium mb-3 text-red-400">
              Danger Zone
            </h3>
            <button
              onClick={handleDelete}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm hover:bg-red-500/20 transition-colors"
            >
              <Trash2 size={14} />
              Delete Bot
            </button>
          </div>
        </div>
      ),
    },
  ];

  return (
    <div>
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => navigate("/bots")}
          className="p-2 rounded-lg hover:bg-white/[0.06] text-white/40 hover:text-white/70 transition-colors"
        >
          <ArrowLeft size={20} />
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-white/[0.06] flex items-center justify-center">
              <Bot size={18} className="text-blue-400" />
            </div>
            <div>
              <h1 className="text-xl font-semibold">{bot.name}</h1>
              <p className="text-xs text-white/40">
                {bot.framework} · {bot.status}
              </p>
            </div>
          </div>
        </div>
        <button
          onClick={() => {
            form.setFieldsValue({
              name: bot.name,
              description: bot.description,
              webhook_url: bot.webhook_url,
              is_public: bot.is_public,
            });
            setEditing(true);
          }}
          className="glass-btn-ghost h-9 px-4 rounded-xl text-sm gap-2"
        >
          <Settings size={14} />
          Edit
        </button>
      </div>

      {/* Description */}
      {bot.description && (
        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="text-sm text-white/50 mb-6"
        >
          {bot.description}
        </motion.p>
      )}

      {/* Tabs */}
      <Tabs items={tabItems} defaultActiveKey="chat" />

      {/* Edit modal */}
      <Modal
        open={editing}
        title="Edit Bot"
        okText="Save"
        onOk={handleUpdate}
        onCancel={() => setEditing(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="webhook_url" label="Webhook URL">
            <Input placeholder="https://your-bot.example.com/webhook" />
          </Form.Item>
          <Form.Item name="is_public" label="Public" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

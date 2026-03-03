import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { Bot, Plus, Search, Trash2, Key, Globe, Lock } from "lucide-react";
import { botApi, type Bot as BotType, type CreateBotParams } from "@/api/bot";
import { Modal, Form, Input, Select, Switch, message } from "antd";

const FRAMEWORKS = [
  {
    value: "astrbot",
    label: "AstrBot",
    description: "Connect via AstrBot HTTP API",
  },
  {
    value: "custom",
    label: "Custom Webhook",
    description: "Generic webhook integration",
  },
] as const;

export default function BotListPage() {
  const navigate = useNavigate();
  const [bots, setBots] = useState<BotType[]>([]);
  const [total, setTotal] = useState(0);
  const [page] = useState(1);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [form] = Form.useForm();
  const selectedFramework = Form.useWatch("framework", form);

  const loadBots = async () => {
    setLoading(true);
    try {
      const res = await botApi.list(page, 20);
      setBots(res.data.data?.items || []);
      setTotal(res.data.data?.total || 0);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBots();
  }, [page]);

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();

      // Build framework-specific config JSON.
      const config: Record<string, string> = {};
      if (values.framework === "astrbot") {
        if (values.api_key) config.api_key = values.api_key;
        if (values.platform) config.platform = values.platform;
      }

      const params: CreateBotParams = {
        name: values.name,
        description: values.description,
        framework: values.framework,
        webhook_url: values.webhook_url,
        visibility: values.is_public ? "public" : "private",
        config:
          Object.keys(config).length > 0 ? JSON.stringify(config) : undefined,
      };
      const res = await botApi.create(params);
      const created = res.data.data;

      message.success("Bot created successfully");
      setShowCreate(false);
      form.resetFields();
      loadBots();

      // Show the access token to the user (only shown once on creation).
      if (created?.access_token) {
        Modal.info({
          title: "Bot Access Token",
          width: 520,
          content: (
            <div className="mt-3">
              <p
                className="text-sm mb-3"
                style={{ color: "rgba(255,255,255,0.5)" }}
              >
                Save this token securely. It won't be shown again.
              </p>
              <code
                className="block p-3 rounded-lg text-xs font-mono break-all select-all"
                style={{ background: "rgba(0,0,0,0.3)" }}
              >
                {created.access_token}
              </code>
            </div>
          ),
        });
      }
    } catch {
      // validation or API error
    }
  };

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: "Delete Bot",
      content:
        "Are you sure you want to delete this bot? This action cannot be undone.",
      okText: "Delete",
      okButtonProps: { danger: true },
      onOk: async () => {
        await botApi.delete(id);
        message.success("Bot deleted");
        loadBots();
      },
    });
  };

  const filteredBots = bots.filter(
    (b) =>
      !search ||
      b.name.toLowerCase().includes(search.toLowerCase()) ||
      b.framework.toLowerCase().includes(search.toLowerCase()),
  );

  const containerVariants = {
    hidden: {},
    show: { transition: { staggerChildren: 0.05 } },
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 12 },
    show: {
      opacity: 1,
      y: 0,
      transition: { type: "spring" as const, damping: 25, stiffness: 300 },
    },
  };

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-semibold">Bots</h1>
          <p className="text-sm text-white/40 mt-1">
            {total} bot{total !== 1 ? "s" : ""} registered
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="glass-btn h-10 px-4 rounded-xl text-sm gap-2"
        >
          <Plus size={16} />
          Create Bot
        </button>
      </div>

      {/* Search */}
      <div className="relative mb-6">
        <Search
          size={16}
          className="absolute left-3 top-1/2 -translate-y-1/2 text-white/30"
        />
        <input
          type="text"
          placeholder="Search bots..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="glass-input w-full h-10 rounded-xl pl-10 pr-4 text-sm"
        />
      </div>

      {/* Bot grid */}
      {loading ? (
        <div className="flex items-center justify-center h-48 text-white/30">
          Loading...
        </div>
      ) : filteredBots.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-48 text-white/30">
          <Bot size={40} className="mb-3 opacity-30" />
          <p>No bots yet. Create your first bot to get started.</p>
        </div>
      ) : (
        <motion.div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
          variants={containerVariants}
          initial="hidden"
          animate="show"
        >
          {filteredBots.map((bot) => (
            <motion.div
              key={bot.id}
              variants={itemVariants}
              onClick={() => navigate(`/bots/${bot.id}`)}
              className="glass rounded-2xl p-5 cursor-pointer hover:bg-white/[0.08] transition-all group"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-white/[0.06] flex items-center justify-center">
                  <Bot size={18} className="text-blue-400" />
                </div>
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDelete(bot.id);
                    }}
                    className="p-1.5 rounded-lg hover:bg-red-500/10 text-white/30 hover:text-red-400 transition-colors"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>
              <h3 className="font-medium text-sm mb-1 truncate">{bot.name}</h3>
              <p className="text-xs text-white/40 line-clamp-2 mb-3">
                {bot.description || "No description"}
              </p>
              <div className="flex items-center gap-2">
                <span className="text-[10px] px-2 py-0.5 rounded-full bg-white/[0.06] border border-white/[0.08] text-white/50">
                  {bot.framework}
                </span>
                <span
                  className={`text-[10px] px-2 py-0.5 rounded-full ${
                    bot.status === "active"
                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                      : "bg-white/[0.06] text-white/40 border border-white/[0.08]"
                  }`}
                >
                  {bot.status}
                </span>
                {bot.visibility === "public" ? (
                  <Globe size={12} className="text-white/30" />
                ) : (
                  <Lock size={12} className="text-white/30" />
                )}
              </div>
            </motion.div>
          ))}
        </motion.div>
      )}

      {/* Create modal */}
      <Modal
        open={showCreate}
        title="Create Bot"
        okText="Create"
        onOk={handleCreate}
        onCancel={() => {
          setShowCreate(false);
          form.resetFields();
        }}
        destroyOnClose
        width={480}
      >
        <Form
          form={form}
          layout="vertical"
          className="mt-4"
          initialValues={{ framework: "astrbot", platform: "ubothub" }}
        >
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: "Bot name is required" }]}
          >
            <Input placeholder="My Bot" />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input.TextArea placeholder="What does this bot do?" rows={2} />
          </Form.Item>
          <Form.Item
            name="framework"
            label="Framework"
            rules={[{ required: true, message: "Select a framework" }]}
          >
            <Select placeholder="Select framework">
              {FRAMEWORKS.map((f) => (
                <Select.Option key={f.value} value={f.value}>
                  <div>
                    <span>{f.label}</span>
                    <span
                      className="text-xs ml-2"
                      style={{ color: "rgba(255,255,255,0.35)" }}
                    >
                      {f.description}
                    </span>
                  </div>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          {/* AstrBot-specific fields */}
          {selectedFramework === "astrbot" && (
            <>
              <Form.Item
                name="webhook_url"
                label="AstrBot API Base URL"
                rules={[
                  { required: true, message: "AstrBot base URL is required" },
                  { type: "url", message: "Must be a valid URL" },
                ]}
                extra="The HTTP API base URL of your AstrBot instance (e.g. http://localhost:6185)"
              >
                <Input placeholder="http://localhost:6185" />
              </Form.Item>
              <Form.Item
                name="api_key"
                label="API Key"
                extra="AstrBot auth token (format: abk_...). Will be encrypted before storage."
              >
                <Input.Password
                  placeholder="abk_..."
                  prefix={<Key size={14} className="text-gray-400" />}
                />
              </Form.Item>
              <Form.Item
                name="platform"
                label="Platform ID"
                extra="Platform identifier sent to AstrBot (default: ubothub)"
              >
                <Input placeholder="ubothub" />
              </Form.Item>
            </>
          )}

          {/* Custom webhook fields */}
          {selectedFramework === "custom" && (
            <Form.Item
              name="webhook_url"
              label="Webhook URL"
              rules={[
                { required: true, message: "Webhook URL is required" },
                { type: "url", message: "Must be a valid URL" },
              ]}
              extra="The endpoint that will receive message payloads"
            >
              <Input placeholder="https://your-bot.example.com/webhook" />
            </Form.Item>
          )}

          <Form.Item name="is_public" label="Public" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

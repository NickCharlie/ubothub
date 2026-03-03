import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Sparkles, Plus, Trash2, Search, Link2, Unlink } from "lucide-react";
import {
  avatarApi,
  type AvatarConfig,
  type CreateAvatarParams,
} from "@/api/avatar";
import { botApi, type Bot } from "@/api/bot";
import { Modal, Form, Input, Select, message } from "antd";

export default function AvatarListPage() {
  const [avatars, setAvatars] = useState<AvatarConfig[]>([]);
  const [total, setTotal] = useState(0);
  const [page] = useState(1);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [showBind, setShowBind] = useState<string | null>(null);
  const [bots, setBots] = useState<Bot[]>([]);
  const [selectedBotId, setSelectedBotId] = useState<string>("");
  const [form] = Form.useForm();

  const loadAvatars = async () => {
    setLoading(true);
    try {
      const res = await avatarApi.list(page, 20);
      setAvatars(res.data.data?.items || []);
      setTotal(res.data.data?.total || 0);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  const loadBots = async () => {
    try {
      const res = await botApi.list(1, 100);
      setBots(res.data.data?.items || []);
    } catch {
      // ignore
    }
  };

  useEffect(() => {
    loadAvatars();
    loadBots();
  }, [page]);

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      const params: CreateAvatarParams = {
        name: values.name,
        description: values.description,
        render_type: values.render_type,
      };
      await avatarApi.create(params);
      message.success("Avatar created");
      setShowCreate(false);
      form.resetFields();
      loadAvatars();
    } catch {
      // validation error
    }
  };

  const handleDelete = (id: string) => {
    Modal.confirm({
      title: "Delete Avatar",
      content:
        "This will remove the avatar configuration and all asset bindings.",
      okText: "Delete",
      okButtonProps: { danger: true },
      onOk: async () => {
        await avatarApi.delete(id);
        message.success("Avatar deleted");
        loadAvatars();
      },
    });
  };

  const handleBindBot = async () => {
    if (!showBind || !selectedBotId) return;
    try {
      await avatarApi.bindBot(showBind, selectedBotId);
      message.success("Bot bound to avatar");
      setShowBind(null);
      setSelectedBotId("");
      loadAvatars();
    } catch (err: any) {
      message.error(err?.response?.data?.message || "Failed to bind bot");
    }
  };

  const filteredAvatars = avatars.filter(
    (a) =>
      !search ||
      a.name.toLowerCase().includes(search.toLowerCase()) ||
      a.render_type.toLowerCase().includes(search.toLowerCase()),
  );

  const renderTypeColors: Record<string, string> = {
    "3d": "bg-blue-500/10 text-blue-400 border-blue-500/20",
    live2d: "bg-purple-500/10 text-purple-400 border-purple-500/20",
    sprite: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  };

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
          <h1 className="text-2xl font-semibold">Avatars</h1>
          <p className="text-sm text-white/40 mt-1">
            {total} avatar configuration{total !== 1 ? "s" : ""}
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="glass-btn h-10 px-4 rounded-xl text-sm gap-2"
        >
          <Plus size={16} />
          Create Avatar
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
          placeholder="Search avatars..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="glass-input w-full h-10 rounded-xl pl-10 pr-4 text-sm"
        />
      </div>

      {/* Avatar grid */}
      {loading ? (
        <div className="flex items-center justify-center h-48 text-white/30">
          Loading...
        </div>
      ) : filteredAvatars.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-48 text-white/30">
          <Sparkles size={40} className="mb-3 opacity-30" />
          <p>No avatars yet. Create one to get started.</p>
        </div>
      ) : (
        <motion.div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
          variants={containerVariants}
          initial="hidden"
          animate="show"
        >
          {filteredAvatars.map((avatar) => (
            <motion.div
              key={avatar.id}
              variants={itemVariants}
              className="glass rounded-2xl p-5 group"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 border border-white/[0.06] flex items-center justify-center">
                  <Sparkles size={18} className="text-purple-400" />
                </div>
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={() => {
                      setShowBind(avatar.id);
                      setSelectedBotId(avatar.bot_id || "");
                    }}
                    className="p-1.5 rounded-lg hover:bg-white/[0.06] text-white/30 hover:text-white/70 transition-colors"
                    title={avatar.bot_id ? "Change bound bot" : "Bind bot"}
                  >
                    {avatar.bot_id ? <Unlink size={14} /> : <Link2 size={14} />}
                  </button>
                  <button
                    onClick={() => handleDelete(avatar.id)}
                    className="p-1.5 rounded-lg hover:bg-red-500/10 text-white/30 hover:text-red-400 transition-colors"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>

              <h3 className="font-medium text-sm mb-1 truncate">
                {avatar.name}
              </h3>
              <p className="text-xs text-white/40 line-clamp-2 mb-3">
                {avatar.description || "No description"}
              </p>

              <div className="flex items-center gap-2 flex-wrap">
                <span
                  className={`text-[10px] px-2 py-0.5 rounded-full border ${renderTypeColors[avatar.render_type] || "bg-white/[0.06] text-white/50 border-white/[0.08]"}`}
                >
                  {avatar.render_type}
                </span>
                {avatar.bot && (
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    {avatar.bot.name}
                  </span>
                )}
                {avatar.avatar_assets && avatar.avatar_assets.length > 0 && (
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-white/[0.06] text-white/50 border border-white/[0.08]">
                    {avatar.avatar_assets.length} asset
                    {avatar.avatar_assets.length !== 1 ? "s" : ""}
                  </span>
                )}
              </div>
            </motion.div>
          ))}
        </motion.div>
      )}

      {/* Create modal */}
      <Modal
        open={showCreate}
        title="Create Avatar"
        okText="Create"
        onOk={handleCreate}
        onCancel={() => setShowCreate(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: "Avatar name is required" }]}
          >
            <Input placeholder="My Avatar" />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input.TextArea placeholder="Describe this avatar..." rows={3} />
          </Form.Item>
          <Form.Item
            name="render_type"
            label="Render Type"
            rules={[{ required: true, message: "Select a render type" }]}
          >
            <Select placeholder="Select type">
              <Select.Option value="3d">3D (VRM / glTF / FBX)</Select.Option>
              <Select.Option value="live2d">Live2D (Cubism)</Select.Option>
              <Select.Option value="sprite">Sprite</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* Bind bot modal */}
      <Modal
        open={!!showBind}
        title="Bind Bot to Avatar"
        okText="Bind"
        onOk={handleBindBot}
        onCancel={() => {
          setShowBind(null);
          setSelectedBotId("");
        }}
      >
        <div className="mt-4">
          <p className="text-sm text-white/50 mb-4">
            Select a bot to bind with this avatar. Each bot can only be bound to
            one avatar.
          </p>
          <Select
            className="w-full"
            placeholder="Select a bot"
            value={selectedBotId || undefined}
            onChange={setSelectedBotId}
          >
            {bots.map((bot) => (
              <Select.Option key={bot.id} value={bot.id}>
                {bot.name} ({bot.framework})
              </Select.Option>
            ))}
          </Select>
        </div>
      </Modal>
    </div>
  );
}

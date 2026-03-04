import { useState, useEffect, useCallback } from "react";
import { Sparkles, Box, Trash2, Check, Loader2 } from "lucide-react";
import { Modal, message } from "antd";
import { botApi } from "@/api/bot";
import { assetApi, type Asset } from "@/api/asset";
import type { AvatarConfig } from "@/api/avatar";

interface BotAvatarSetupProps {
  botId: string;
  onAvatarChanged?: () => void;
}

export default function BotAvatarSetup({
  botId,
  onAvatarChanged,
}: BotAvatarSetupProps) {
  const [avatar, setAvatar] = useState<AvatarConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [assetsLoading, setAssetsLoading] = useState(false);
  const [showPicker, setShowPicker] = useState(false);
  const [settingUp, setSettingUp] = useState(false);
  const [selectedAssetId, setSelectedAssetId] = useState<string | null>(null);

  const loadAvatar = useCallback(async () => {
    setLoading(true);
    try {
      const res = await botApi.getAvatar(botId);
      setAvatar(res.data.data);
    } catch {
      setAvatar(null);
    } finally {
      setLoading(false);
    }
  }, [botId]);

  useEffect(() => {
    loadAvatar();
  }, [loadAvatar]);

  const loadAssets = useCallback(async () => {
    setAssetsLoading(true);
    try {
      const res = await assetApi.list(1, 100, "model_3d");
      setAssets(res.data.data?.items || []);
    } catch {
      setAssets([]);
    } finally {
      setAssetsLoading(false);
    }
  }, []);

  const openPicker = () => {
    setShowPicker(true);
    setSelectedAssetId(null);
    loadAssets();
  };

  const handleSetup = async () => {
    if (!selectedAssetId) return;
    setSettingUp(true);
    try {
      await botApi.setupAvatar(botId, { asset_id: selectedAssetId });
      message.success("Avatar linked successfully");
      setShowPicker(false);
      loadAvatar();
      onAvatarChanged?.();
    } catch {
      message.error("Failed to setup avatar");
    } finally {
      setSettingUp(false);
    }
  };

  const handleRemove = () => {
    Modal.confirm({
      title: "Remove Avatar",
      content:
        "This will remove the 3D avatar from this bot. The asset itself will not be deleted.",
      okText: "Remove",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await botApi.removeAvatar(botId);
          message.success("Avatar removed");
          setAvatar(null);
          onAvatarChanged?.();
        } catch {
          message.error("Failed to remove avatar");
        }
      },
    });
  };

  const primaryAsset = avatar?.avatar_assets?.find(
    (a) =>
      a.role === "primary_model" || a.role === "body" || a.role === "model",
  );

  if (loading) {
    return (
      <div className="glass rounded-xl p-5">
        <h3 className="text-sm font-medium mb-3 text-white/70 flex items-center gap-2">
          <Sparkles size={14} className="text-purple-400" />
          3D Avatar
        </h3>
        <div className="flex items-center justify-center py-4 text-white/30 text-xs">
          <Loader2 size={14} className="animate-spin mr-2" />
          Loading...
        </div>
      </div>
    );
  }

  return (
    <div className="glass rounded-xl p-5">
      <h3 className="text-sm font-medium mb-3 text-white/70 flex items-center gap-2">
        <Sparkles size={14} className="text-purple-400" />
        3D Avatar
      </h3>

      {avatar ? (
        <div className="space-y-3">
          <div className="flex items-center gap-3 p-3 rounded-lg bg-purple-500/5 border border-purple-500/15">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500/20 to-blue-500/20 border border-white/[0.06] flex items-center justify-center flex-shrink-0">
              <Box size={16} className="text-purple-400" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm text-white/80 truncate">{avatar.name}</p>
              <div className="flex items-center gap-2 mt-0.5">
                <span className="text-[10px] px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-400 border border-purple-500/20">
                  {avatar.render_type}
                </span>
                {primaryAsset && (
                  <span className="text-[10px] text-white/40 truncate">
                    {primaryAsset.asset?.name || "Model"}
                  </span>
                )}
              </div>
            </div>
          </div>

          <div className="flex gap-2">
            <button
              onClick={openPicker}
              className="flex-1 glass-btn h-8 rounded-lg text-xs gap-1.5"
            >
              <Box size={12} />
              Change Model
            </button>
            <button
              onClick={handleRemove}
              className="glass-btn h-8 px-3 rounded-lg text-xs text-red-400 hover:bg-red-500/10"
            >
              <Trash2 size={12} />
            </button>
          </div>
        </div>
      ) : (
        <div className="text-center py-4">
          <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-purple-500/10 to-blue-500/10 border border-white/[0.06] flex items-center justify-center mx-auto mb-3">
            <Box size={20} className="text-white/20" />
          </div>
          <p className="text-xs text-white/40 mb-3">
            No 3D avatar configured. Add a VRM/glTF model to bring your bot to
            life in the Plaza chat.
          </p>
          <button
            onClick={openPicker}
            className="glass-btn h-9 px-5 rounded-xl text-xs gap-2"
          >
            <Sparkles size={12} />
            Setup Avatar
          </button>
        </div>
      )}

      {/* Asset picker modal */}
      <Modal
        open={showPicker}
        title="Select 3D Model"
        okText={settingUp ? "Setting up..." : "Link to Bot"}
        onOk={handleSetup}
        onCancel={() => setShowPicker(false)}
        okButtonProps={{ disabled: !selectedAssetId || settingUp }}
        destroyOnClose
        width={560}
      >
        <p className="text-xs text-white/40 mb-4">
          Select a 3D model asset to use as this bot's avatar. The model will be
          displayed in the Plaza chat page.
        </p>

        {assetsLoading ? (
          <div className="flex items-center justify-center py-8 text-white/30">
            <Loader2 size={16} className="animate-spin mr-2" />
            Loading assets...
          </div>
        ) : assets.length === 0 ? (
          <div className="text-center py-8">
            <Box size={24} className="text-white/20 mx-auto mb-2" />
            <p className="text-sm text-white/40">No 3D model assets found</p>
            <p className="text-xs text-white/25 mt-1">
              Upload a VRM, glTF, or FBX model in the Assets page first
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-2 max-h-[360px] overflow-y-auto pr-1">
            {assets.map((asset) => (
              <button
                key={asset.id}
                onClick={() => setSelectedAssetId(asset.id)}
                className={`flex items-center gap-3 p-3 rounded-xl border text-left transition-all ${
                  selectedAssetId === asset.id
                    ? "border-purple-500/40 bg-purple-500/10"
                    : "border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.04]"
                }`}
              >
                <div
                  className={`w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 ${
                    selectedAssetId === asset.id
                      ? "bg-purple-500/20"
                      : "bg-white/[0.04]"
                  }`}
                >
                  {selectedAssetId === asset.id ? (
                    <Check size={14} className="text-purple-400" />
                  ) : (
                    <Box size={14} className="text-white/30" />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-white/80 truncate">{asset.name}</p>
                  <p className="text-[10px] text-white/30 mt-0.5">
                    {asset.format.toUpperCase()} ·{" "}
                    {(asset.file_size / 1024 / 1024).toFixed(1)}MB
                  </p>
                </div>
              </button>
            ))}
          </div>
        )}
      </Modal>
    </div>
  );
}

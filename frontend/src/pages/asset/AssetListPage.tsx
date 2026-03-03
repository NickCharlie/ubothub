import { useEffect, useState, useRef } from "react";
import { motion } from "framer-motion";
import {
  Box,
  Trash2,
  Search,
  Upload,
  Download,
  Globe,
  Lock,
  FileType,
} from "lucide-react";
import { assetApi, type Asset, type CompleteUploadParams } from "@/api/asset";
import { Modal, Form, Input, Select, Switch, message, Progress } from "antd";

const CATEGORY_OPTIONS = [
  {
    value: "model_3d",
    label: "3D Model",
    formats: ["vrm", "glb", "gltf", "fbx"],
  },
  { value: "model_live2d", label: "Live2D", formats: ["zip", "moc3"] },
  { value: "motion", label: "Motion", formats: ["bvh", "vmd", "fbx", "vrma"] },
  {
    value: "texture",
    label: "Texture",
    formats: ["png", "jpg", "jpeg", "webp", "ktx2"],
  },
];

const categoryColors: Record<string, string> = {
  model_3d: "bg-blue-500/10 text-blue-400 border-blue-500/20",
  model_live2d: "bg-purple-500/10 text-purple-400 border-purple-500/20",
  motion: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  texture: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
};

const statusColors: Record<string, string> = {
  processing: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  ready: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
  failed: "bg-red-500/10 text-red-400 border-red-500/20",
};

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024)
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

export default function AssetListPage() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [total, setTotal] = useState(0);
  const [page] = useState(1);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [filterCategory, setFilterCategory] = useState<string>("");
  const [showUpload, setShowUpload] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [form] = Form.useForm();

  const loadAssets = async () => {
    setLoading(true);
    try {
      const res = await assetApi.list(page, 20, filterCategory || undefined);
      setAssets(res.data.data?.items || []);
      setTotal(res.data.data?.total || 0);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAssets();
  }, [page, filterCategory]);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setSelectedFile(file);

    // Auto-detect category and format from extension
    const ext = file.name.split(".").pop()?.toLowerCase() || "";
    for (const cat of CATEGORY_OPTIONS) {
      if (cat.formats.includes(ext)) {
        form.setFieldsValue({
          category: cat.value,
          format: ext,
          name: file.name.replace(/\.[^.]+$/, ""),
        });
        break;
      }
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) {
      message.error("Please select a file");
      return;
    }
    try {
      const values = await form.validateFields();
      setUploading(true);
      setUploadProgress(10);

      // Step 1: Get presigned URL
      const presignRes = await assetApi.getPresignedUpload(
        selectedFile.name,
        values.category,
        selectedFile.size,
      );
      const { upload_url, file_key } = presignRes.data.data;
      setUploadProgress(30);

      // Step 2: Upload to object storage
      await fetch(upload_url, {
        method: "PUT",
        body: selectedFile,
        headers: { "Content-Type": "application/octet-stream" },
      });
      setUploadProgress(70);

      // Step 3: Complete upload
      const params: CompleteUploadParams = {
        file_key,
        name: values.name,
        description: values.description,
        category: values.category,
        format: values.format || selectedFile.name.split(".").pop() || "",
        file_size: selectedFile.size,
        is_public: values.is_public || false,
        tags: values.tags
          ? values.tags.split(",").map((t: string) => t.trim())
          : [],
      };
      await assetApi.completeUpload(params);
      setUploadProgress(100);

      message.success("Asset uploaded successfully");
      setShowUpload(false);
      setSelectedFile(null);
      form.resetFields();
      if (fileInputRef.current) fileInputRef.current.value = "";
      loadAssets();
    } catch (err: any) {
      message.error(err?.response?.data?.message || "Upload failed");
    } finally {
      setUploading(false);
      setUploadProgress(0);
    }
  };

  const handleDelete = (id: string) => {
    Modal.confirm({
      title: "Delete Asset",
      content: "This will permanently delete the file from storage.",
      okText: "Delete",
      okButtonProps: { danger: true },
      onOk: async () => {
        await assetApi.delete(id);
        message.success("Asset deleted");
        loadAssets();
      },
    });
  };

  const handleDownload = async (id: string) => {
    try {
      const res = await assetApi.getDownloadURL(id);
      window.open(res.data.data.url, "_blank");
    } catch {
      message.error("Failed to get download URL");
    }
  };

  const filteredAssets = assets.filter(
    (a) =>
      !search ||
      a.name.toLowerCase().includes(search.toLowerCase()) ||
      a.format.toLowerCase().includes(search.toLowerCase()),
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
          <h1 className="text-2xl font-semibold">Assets</h1>
          <p className="text-sm text-white/40 mt-1">
            {total} asset{total !== 1 ? "s" : ""} uploaded
          </p>
        </div>
        <button
          onClick={() => setShowUpload(true)}
          className="glass-btn h-10 px-4 rounded-xl text-sm gap-2"
        >
          <Upload size={16} />
          Upload Asset
        </button>
      </div>

      {/* Filters */}
      <div className="flex gap-3 mb-6">
        <div className="relative flex-1">
          <Search
            size={16}
            className="absolute left-3 top-1/2 -translate-y-1/2 text-white/30"
          />
          <input
            type="text"
            placeholder="Search assets..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="glass-input w-full h-10 rounded-xl pl-10 pr-4 text-sm"
          />
        </div>
        <select
          value={filterCategory}
          onChange={(e) => setFilterCategory(e.target.value)}
          className="glass-input h-10 rounded-xl px-3 text-sm min-w-[140px]"
        >
          <option value="">All Categories</option>
          {CATEGORY_OPTIONS.map((c) => (
            <option key={c.value} value={c.value}>
              {c.label}
            </option>
          ))}
        </select>
      </div>

      {/* Asset grid */}
      {loading ? (
        <div className="flex items-center justify-center h-48 text-white/30">
          Loading...
        </div>
      ) : filteredAssets.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-48 text-white/30">
          <Box size={40} className="mb-3 opacity-30" />
          <p>No assets yet. Upload your first 3D model or texture.</p>
        </div>
      ) : (
        <motion.div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
          variants={containerVariants}
          initial="hidden"
          animate="show"
        >
          {filteredAssets.map((asset) => (
            <motion.div
              key={asset.id}
              variants={itemVariants}
              className="glass rounded-2xl p-5 group"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-cyan-500/20 border border-white/[0.06] flex items-center justify-center">
                  <FileType size={18} className="text-emerald-400" />
                </div>
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={() => handleDownload(asset.id)}
                    className="p-1.5 rounded-lg hover:bg-white/[0.06] text-white/30 hover:text-white/70 transition-colors"
                    title="Download"
                  >
                    <Download size={14} />
                  </button>
                  <button
                    onClick={() => handleDelete(asset.id)}
                    className="p-1.5 rounded-lg hover:bg-red-500/10 text-white/30 hover:text-red-400 transition-colors"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>

              <h3 className="font-medium text-sm mb-1 truncate">
                {asset.name}
              </h3>
              <p className="text-xs text-white/40 line-clamp-2 mb-3">
                {asset.description || formatFileSize(asset.file_size)}
              </p>

              <div className="flex items-center gap-2 flex-wrap">
                <span
                  className={`text-[10px] px-2 py-0.5 rounded-full border ${categoryColors[asset.category] || "bg-white/[0.06] text-white/50 border-white/[0.08]"}`}
                >
                  {asset.category.replace("model_", "")}
                </span>
                <span className="text-[10px] px-2 py-0.5 rounded-full bg-white/[0.06] text-white/50 border border-white/[0.08]">
                  .{asset.format}
                </span>
                <span
                  className={`text-[10px] px-2 py-0.5 rounded-full border ${statusColors[asset.status] || ""}`}
                >
                  {asset.status}
                </span>
                {asset.is_public ? (
                  <Globe size={12} className="text-blue-400" />
                ) : (
                  <Lock size={12} className="text-white/30" />
                )}
              </div>
            </motion.div>
          ))}
        </motion.div>
      )}

      {/* Upload modal */}
      <Modal
        open={showUpload}
        title="Upload Asset"
        okText={uploading ? "Uploading..." : "Upload"}
        onOk={handleUpload}
        onCancel={() => {
          if (!uploading) {
            setShowUpload(false);
            setSelectedFile(null);
            form.resetFields();
          }
        }}
        okButtonProps={{ disabled: uploading }}
        cancelButtonProps={{ disabled: uploading }}
        closable={!uploading}
        destroyOnClose
      >
        <Form form={form} layout="vertical" className="mt-4">
          {/* File picker */}
          <div className="mb-4">
            <label className="block text-sm text-white/75 mb-2">File</label>
            <div
              onClick={() => fileInputRef.current?.click()}
              className="glass-input rounded-xl p-6 text-center cursor-pointer hover:border-blue-500/30 transition-colors"
            >
              {selectedFile ? (
                <div>
                  <p className="text-sm">{selectedFile.name}</p>
                  <p className="text-xs text-white/40 mt-1">
                    {formatFileSize(selectedFile.size)}
                  </p>
                </div>
              ) : (
                <div>
                  <Upload size={24} className="mx-auto mb-2 text-white/30" />
                  <p className="text-sm text-white/50">Click to select file</p>
                  <p className="text-xs text-white/30 mt-1">
                    VRM, glTF, FBX, Live2D, BVH, VMD, PNG, JPG (max 500MB)
                  </p>
                </div>
              )}
            </div>
            <input
              ref={fileInputRef}
              type="file"
              className="hidden"
              onChange={handleFileSelect}
              accept=".vrm,.glb,.gltf,.fbx,.zip,.moc3,.bvh,.vmd,.vrma,.png,.jpg,.jpeg,.webp,.ktx2"
            />
          </div>

          {uploading && <Progress percent={uploadProgress} className="mb-4" />}

          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: "Asset name is required" }]}
          >
            <Input placeholder="Asset name" />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input.TextArea placeholder="Describe this asset..." rows={2} />
          </Form.Item>
          <Form.Item
            name="category"
            label="Category"
            rules={[{ required: true, message: "Select a category" }]}
          >
            <Select placeholder="Select category">
              {CATEGORY_OPTIONS.map((c) => (
                <Select.Option key={c.value} value={c.value}>
                  {c.label}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="tags" label="Tags">
            <Input placeholder="Comma-separated tags" />
          </Form.Item>
          <Form.Item name="is_public" label="Public" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { User as UserIcon, Mail, Shield, Key, Save } from "lucide-react";
import { useAuthStore } from "@/stores/auth";
import {
  userApi,
  type UpdateProfileParams,
  type ChangePasswordParams,
} from "@/api/user";
import { Form, Input, message, Tabs } from "antd";

export default function ProfilePage() {
  const { user, fetchUser } = useAuthStore();
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [saving, setSaving] = useState(false);
  const [changingPassword, setChangingPassword] = useState(false);

  useEffect(() => {
    if (user) {
      profileForm.setFieldsValue({
        display_name: user.display_name,
        avatar_url: user.avatar_url,
      });
    }
  }, [user, profileForm]);

  const handleUpdateProfile = async () => {
    try {
      const values = await profileForm.validateFields();
      setSaving(true);
      const params: UpdateProfileParams = {
        display_name: values.display_name,
        avatar_url: values.avatar_url,
      };
      await userApi.updateProfile(params);
      await fetchUser();
      message.success("Profile updated");
    } catch {
      // validation error
    } finally {
      setSaving(false);
    }
  };

  const handleChangePassword = async () => {
    try {
      const values = await passwordForm.validateFields();
      if (values.new_password !== values.confirm_password) {
        message.error("Passwords do not match");
        return;
      }
      setChangingPassword(true);
      const params: ChangePasswordParams = {
        old_password: values.old_password,
        new_password: values.new_password,
      };
      await userApi.changePassword(params);
      message.success("Password changed successfully");
      passwordForm.resetFields();
    } catch (err: any) {
      message.error(
        err?.response?.data?.message || "Failed to change password",
      );
    } finally {
      setChangingPassword(false);
    }
  };

  if (!user) return null;

  const tabItems = [
    {
      key: "profile",
      label: "Profile",
      children: (
        <Form form={profileForm} layout="vertical" className="max-w-md">
          <Form.Item name="display_name" label="Display Name">
            <Input placeholder="Your display name" />
          </Form.Item>
          <Form.Item name="avatar_url" label="Avatar URL">
            <Input placeholder="https://example.com/avatar.png" />
          </Form.Item>
          <button
            onClick={handleUpdateProfile}
            disabled={saving}
            className="glass-btn h-10 px-5 rounded-xl text-sm gap-2"
          >
            <Save size={14} />
            {saving ? "Saving..." : "Save Changes"}
          </button>
        </Form>
      ),
    },
    {
      key: "security",
      label: "Security",
      children: (
        <Form form={passwordForm} layout="vertical" className="max-w-md">
          <Form.Item
            name="old_password"
            label="Current Password"
            rules={[{ required: true }]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="new_password"
            label="New Password"
            rules={[
              { required: true },
              { min: 8, message: "Password must be at least 8 characters" },
            ]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="confirm_password"
            label="Confirm New Password"
            rules={[{ required: true }]}
          >
            <Input.Password />
          </Form.Item>
          <button
            onClick={handleChangePassword}
            disabled={changingPassword}
            className="glass-btn h-10 px-5 rounded-xl text-sm gap-2"
          >
            <Key size={14} />
            {changingPassword ? "Changing..." : "Change Password"}
          </button>
        </Form>
      ),
    },
  ];

  return (
    <div>
      {/* User info header */}
      <motion.div
        className="glass rounded-2xl p-6 mb-6"
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ type: "spring", damping: 25, stiffness: 300 }}
      >
        <div className="flex items-center gap-4">
          <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center flex-shrink-0">
            {user.avatar_url ? (
              <img
                src={user.avatar_url}
                alt="avatar"
                className="w-16 h-16 rounded-2xl object-cover"
              />
            ) : (
              <UserIcon size={28} className="text-white" />
            )}
          </div>
          <div>
            <h1 className="text-xl font-semibold">
              {user.display_name || user.username}
            </h1>
            <div className="flex items-center gap-3 mt-1">
              <span className="flex items-center gap-1 text-xs text-white/40">
                <Mail size={12} />
                {user.email}
              </span>
              <span className="flex items-center gap-1 text-xs text-white/40">
                <Shield size={12} />
                {user.role}
              </span>
              <span
                className={`text-[10px] px-2 py-0.5 rounded-full border ${
                  user.status === "active"
                    ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                    : "bg-red-500/10 text-red-400 border-red-500/20"
                }`}
              >
                {user.status}
              </span>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Settings tabs */}
      <div className="glass rounded-2xl p-6">
        <Tabs items={tabItems} />
      </div>
    </div>
  );
}

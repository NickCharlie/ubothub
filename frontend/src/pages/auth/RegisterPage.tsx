import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { motion } from "framer-motion";
import { Bot, Eye, EyeOff, Mail, Lock, User } from "lucide-react";
import { authApi } from "@/api/auth";

export default function RegisterPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState({
    email: "",
    username: "",
    password: "",
    confirmPassword: "",
    captcha_id: "",
    captcha_answer: "",
  });
  const [captchaImg, setCaptchaImg] = useState("");
  const [showPwd, setShowPwd] = useState(false);
  const [agreed, setAgreed] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const loadCaptcha = async () => {
    try {
      const res = await authApi.getCaptcha();
      const data = res.data.data;
      setForm((prev) => ({ ...prev, captcha_id: data.captcha_id }));
      setCaptchaImg(data.captcha_image);
    } catch {
      // ignore
    }
  };

  useState(() => {
    loadCaptcha();
  });

  const update = (key: string, value: string) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!agreed) {
      setError("You must agree to the Terms of Service and Privacy Policy.");
      return;
    }
    if (form.password !== form.confirmPassword) {
      setError("Passwords do not match.");
      return;
    }
    if (form.password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }

    setLoading(true);
    try {
      await authApi.register({
        email: form.email,
        username: form.username,
        password: form.password,
        captcha_id: form.captcha_id,
        captcha_answer: form.captcha_answer,
      });
      navigate("/auth/login");
    } catch (err: any) {
      setError(err.response?.data?.message || "Registration failed.");
      loadCaptcha();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-[-20%] right-[-10%] w-[600px] h-[600px] rounded-full bg-purple-600/10 blur-[120px]" />
        <div className="absolute bottom-[-10%] left-[-10%] w-[500px] h-[500px] rounded-full bg-blue-600/10 blur-[120px]" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ type: "spring", damping: 25, stiffness: 300 }}
        className="glass rounded-2xl p-8 w-full max-w-[420px] relative z-10"
      >
        <div className="flex flex-col items-center mb-8">
          <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-purple-500 to-pink-600 flex items-center justify-center mb-4">
            <Bot size={28} className="text-white" />
          </div>
          <h1 className="text-2xl font-semibold">Create account</h1>
          <p className="text-sm text-white/40 mt-1">Join UBotHub platform</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="relative">
            <Mail
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-white/30"
            />
            <input
              type="email"
              placeholder="Email"
              value={form.email}
              onChange={(e) => update("email", e.target.value)}
              required
              className="glass-input w-full h-11 rounded-xl pl-10 pr-4 text-sm"
            />
          </div>

          <div className="relative">
            <User
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-white/30"
            />
            <input
              type="text"
              placeholder="Username"
              value={form.username}
              onChange={(e) => update("username", e.target.value)}
              required
              className="glass-input w-full h-11 rounded-xl pl-10 pr-4 text-sm"
            />
          </div>

          <div className="relative">
            <Lock
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-white/30"
            />
            <input
              type={showPwd ? "text" : "password"}
              placeholder="Password (8+ characters)"
              value={form.password}
              onChange={(e) => update("password", e.target.value)}
              required
              minLength={8}
              className="glass-input w-full h-11 rounded-xl pl-10 pr-10 text-sm"
            />
            <button
              type="button"
              onClick={() => setShowPwd(!showPwd)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-white/30 hover:text-white/60 transition-colors"
            >
              {showPwd ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>

          <div className="relative">
            <Lock
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-white/30"
            />
            <input
              type="password"
              placeholder="Confirm password"
              value={form.confirmPassword}
              onChange={(e) => update("confirmPassword", e.target.value)}
              required
              className="glass-input w-full h-11 rounded-xl pl-10 pr-4 text-sm"
            />
          </div>

          {captchaImg && (
            <div className="flex items-center gap-3">
              <img
                src={captchaImg}
                alt="captcha"
                className="h-11 rounded-lg cursor-pointer border border-white/10"
                onClick={loadCaptcha}
                title="Click to refresh"
              />
              <input
                type="text"
                placeholder="Captcha"
                value={form.captcha_answer}
                onChange={(e) => update("captcha_answer", e.target.value)}
                required
                className="glass-input flex-1 h-11 rounded-xl px-4 text-sm"
              />
            </div>
          )}

          {/* Agreement checkbox */}
          <label className="flex items-start gap-2.5 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={agreed}
              onChange={(e) => setAgreed(e.target.checked)}
              className="mt-0.5 w-4 h-4 rounded border-white/20 bg-white/5 accent-blue-500 cursor-pointer"
            />
            <span className="text-xs text-white/50 leading-relaxed">
              I have read and agree to the{" "}
              <a
                href="/api/v1/legal/terms"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-400 hover:text-blue-300 transition-colors"
              >
                Terms of Service
              </a>{" "}
              and{" "}
              <a
                href="/api/v1/legal/privacy"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-400 hover:text-blue-300 transition-colors"
              >
                Privacy Policy
              </a>
            </span>
          </label>

          {error && (
            <motion.p
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-red-400 text-xs text-center"
            >
              {error}
            </motion.p>
          )}

          <button
            type="submit"
            disabled={loading || !agreed}
            className="glass-btn w-full h-11 rounded-xl text-sm disabled:opacity-50"
          >
            {loading ? "Creating account..." : "Create account"}
          </button>
        </form>

        <p className="text-center mt-6 text-xs text-white/40">
          Already have an account?{" "}
          <Link
            to="/auth/login"
            className="text-blue-400 hover:text-blue-300 transition-colors"
          >
            Sign in
          </Link>
        </p>
      </motion.div>
    </div>
  );
}

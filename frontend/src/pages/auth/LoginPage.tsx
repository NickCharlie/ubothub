import { useEffect, useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { motion } from "framer-motion";
import { Bot, Eye, EyeOff, Mail, Lock, RefreshCw } from "lucide-react";
import { useAuthStore } from "@/stores/auth";
import { authApi } from "@/api/auth";

export default function LoginPage() {
  const navigate = useNavigate();
  const { login, loading } = useAuthStore();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPwd, setShowPwd] = useState(false);
  const [error, setError] = useState("");
  const [captchaId, setCaptchaId] = useState("");
  const [captchaImg, setCaptchaImg] = useState("");
  const [captchaAnswer, setCaptchaAnswer] = useState("");

  const loadCaptcha = async () => {
    try {
      const res = await authApi.getCaptcha();
      const data = res.data.data;
      setCaptchaId(data.captcha_id);
      setCaptchaImg(data.captcha_image);
      setCaptchaAnswer("");
    } catch {
      // ignore
    }
  };

  useEffect(() => {
    loadCaptcha();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      await login(email, password, captchaId, captchaAnswer);
      navigate("/dashboard");
    } catch (err: any) {
      setError(
        err.response?.data?.message || "Login failed, please try again.",
      );
      loadCaptcha();
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      {/* Background gradient */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-[-20%] left-[-10%] w-[600px] h-[600px] rounded-full bg-blue-600/10 blur-[120px]" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[500px] h-[500px] rounded-full bg-purple-600/10 blur-[120px]" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ type: "spring", damping: 25, stiffness: 300 }}
        className="glass rounded-2xl p-8 w-full max-w-[420px] relative z-10"
      >
        {/* Logo */}
        <div className="flex flex-col items-center mb-8">
          <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mb-4">
            <Bot size={28} className="text-white" />
          </div>
          <h1 className="text-2xl font-semibold">Welcome back</h1>
          <p className="text-sm text-white/40 mt-1">Sign in to UBotHub</p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="relative">
            <Mail
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-white/30"
            />
            <input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
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
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
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

          {/* Captcha */}
          {captchaImg && (
            <div className="flex items-center gap-2">
              <img
                src={captchaImg}
                alt="captcha"
                className="h-11 w-[120px] object-contain rounded-xl cursor-pointer border border-white/[0.08] flex-shrink-0"
                onClick={loadCaptcha}
                title="Click to refresh"
              />
              <button
                type="button"
                onClick={loadCaptcha}
                className="p-2 rounded-lg hover:bg-white/[0.06] text-white/40 hover:text-white/60 transition-colors flex-shrink-0"
              >
                <RefreshCw size={14} />
              </button>
              <input
                type="text"
                placeholder="Captcha"
                value={captchaAnswer}
                onChange={(e) => setCaptchaAnswer(e.target.value)}
                required
                className="glass-input flex-1 min-w-0 h-11 rounded-xl px-4 text-sm"
              />
            </div>
          )}

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
            disabled={loading}
            className="glass-btn w-full h-11 rounded-xl text-sm disabled:opacity-50"
          >
            {loading ? "Signing in..." : "Sign in"}
          </button>
        </form>

        {/* Links */}
        <div className="flex items-center justify-between mt-6 text-xs text-white/40">
          <Link
            to="/auth/forgot-password"
            className="hover:text-white/60 transition-colors"
          >
            Forgot password?
          </Link>
          <Link
            to="/auth/register"
            className="hover:text-white/60 transition-colors"
          >
            Create account
          </Link>
        </div>
      </motion.div>
    </div>
  );
}

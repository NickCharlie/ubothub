import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { DollarSign, Gift, Zap, Crown } from "lucide-react";
import { walletApi, type BotPricing } from "@/api/wallet";
import { Form, InputNumber, Select, message } from "antd";

interface BotPricingConfigProps {
  botId: string;
  isOwner: boolean;
}

const modeDescriptions: Record<string, string> = {
  free: "Anyone can use this bot for free",
  per_call: "Users pay per message/call to this bot",
  subscribe: "Users pay a monthly subscription to access this bot",
};

const modeIcons: Record<string, React.ReactNode> = {
  free: <Gift size={16} className="text-emerald-400" />,
  per_call: <Zap size={16} className="text-amber-400" />,
  subscribe: <Crown size={16} className="text-purple-400" />,
};

export default function BotPricingConfig({
  botId,
  isOwner,
}: BotPricingConfigProps) {
  const [pricing, setPricing] = useState<BotPricing | null>(null);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();

  const loadPricing = async () => {
    try {
      const res = await walletApi.getBotPricing(botId);
      setPricing(res.data.data);
    } catch {
      // no pricing set yet
    }
  };

  useEffect(() => {
    loadPricing();
  }, [botId]);

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);
      await walletApi.setBotPricing(botId, {
        mode: values.mode,
        price_per_call: values.price_per_call?.toString() || "0",
        monthly_price: values.monthly_price?.toString() || "0",
        free_calls_per_day: values.free_calls_per_day || 0,
      });
      message.success("Pricing updated");
      setEditing(false);
      loadPricing();
    } catch (err: any) {
      message.error(err?.response?.data?.message || "Failed to update pricing");
    } finally {
      setSaving(false);
    }
  };

  const mode = pricing?.mode || "free";

  return (
    <motion.div
      className="glass rounded-xl p-5"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ type: "spring", damping: 22, stiffness: 300 }}
    >
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <DollarSign size={16} className="text-white/50" />
          <h3 className="text-sm font-medium text-white/70">Pricing</h3>
        </div>
        {isOwner && !editing && (
          <button
            onClick={() => {
              form.setFieldsValue({
                mode: pricing?.mode || "free",
                price_per_call: pricing?.price_per_call
                  ? parseFloat(pricing.price_per_call)
                  : 0,
                monthly_price: pricing?.monthly_price
                  ? parseFloat(pricing.monthly_price)
                  : 0,
                free_calls_per_day: pricing?.free_calls_per_day || 0,
              });
              setEditing(true);
            }}
            className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
          >
            Configure
          </button>
        )}
      </div>

      {!editing ? (
        /* Display mode */
        <div className="space-y-3">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-white/[0.06] flex items-center justify-center">
              {modeIcons[mode]}
            </div>
            <div>
              <p className="text-sm font-medium capitalize">{mode}</p>
              <p className="text-xs text-white/40">{modeDescriptions[mode]}</p>
            </div>
          </div>

          {mode === "per_call" && pricing && (
            <div className="flex items-center gap-4 text-xs text-white/50">
              <span>
                Price: ¥{parseFloat(pricing.price_per_call).toFixed(4)}/call
              </span>
              {pricing.free_calls_per_day > 0 && (
                <span>Free quota: {pricing.free_calls_per_day}/day</span>
              )}
            </div>
          )}

          {mode === "subscribe" && pricing && (
            <div className="text-xs text-white/50">
              Monthly: ¥{parseFloat(pricing.monthly_price).toFixed(2)}
            </div>
          )}

          {pricing && (
            <div className="flex items-center gap-3 text-[10px] text-white/30 pt-2 border-t border-white/[0.06]">
              <span>
                Creator share:{" "}
                {(parseFloat(pricing.creator_rate) * 100).toFixed(0)}%
              </span>
              <span>
                Platform fee:{" "}
                {(parseFloat(pricing.platform_rate) * 100).toFixed(0)}%
              </span>
            </div>
          )}
        </div>
      ) : (
        /* Edit mode */
        <Form form={form} layout="vertical" size="small">
          <Form.Item
            name="mode"
            label="Pricing Mode"
            rules={[{ required: true }]}
          >
            <Select>
              <Select.Option value="free">Free</Select.Option>
              <Select.Option value="per_call">Per Call</Select.Option>
              <Select.Option value="subscribe">Subscription</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            noStyle
            shouldUpdate={(prev, curr) => prev.mode !== curr.mode}
          >
            {({ getFieldValue }) =>
              getFieldValue("mode") === "per_call" ? (
                <>
                  <Form.Item
                    name="price_per_call"
                    label="Price Per Call (CNY)"
                    rules={[{ required: true }]}
                  >
                    <InputNumber
                      className="w-full"
                      min={0.0001}
                      max={100}
                      step={0.01}
                      precision={4}
                      prefix="¥"
                    />
                  </Form.Item>
                  <Form.Item
                    name="free_calls_per_day"
                    label="Free Calls Per Day"
                  >
                    <InputNumber
                      className="w-full"
                      min={0}
                      max={10000}
                      precision={0}
                    />
                  </Form.Item>
                </>
              ) : getFieldValue("mode") === "subscribe" ? (
                <Form.Item
                  name="monthly_price"
                  label="Monthly Price (CNY)"
                  rules={[{ required: true }]}
                >
                  <InputNumber
                    className="w-full"
                    min={0.01}
                    max={9999}
                    step={1}
                    precision={2}
                    prefix="¥"
                  />
                </Form.Item>
              ) : null
            }
          </Form.Item>

          <div className="flex items-center gap-2 mt-4">
            <button
              onClick={handleSave}
              disabled={saving}
              className="glass-btn h-8 px-4 rounded-lg text-xs"
            >
              {saving ? "Saving..." : "Save"}
            </button>
            <button
              onClick={() => setEditing(false)}
              className="glass-btn-ghost h-8 px-4 rounded-lg text-xs"
            >
              Cancel
            </button>
          </div>
        </Form>
      )}
    </motion.div>
  );
}

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import {
  Wallet as WalletIcon,
  ArrowUpRight,
  ArrowDownLeft,
  CreditCard,
  TrendingUp,
  History,
  Plus,
} from "lucide-react";
import { walletApi, type Wallet, type Transaction } from "@/api/wallet";
import { Modal, Form, InputNumber, Select, message, Pagination } from "antd";

const txTypeIcons: Record<string, React.ReactNode> = {
  top_up: <ArrowDownLeft size={14} className="text-emerald-400" />,
  usage: <ArrowUpRight size={14} className="text-red-400" />,
  earning: <TrendingUp size={14} className="text-blue-400" />,
  withdrawal: <ArrowUpRight size={14} className="text-amber-400" />,
  refund: <ArrowDownLeft size={14} className="text-purple-400" />,
  subsidy: <ArrowDownLeft size={14} className="text-cyan-400" />,
};

const txTypeLabels: Record<string, string> = {
  top_up: "Top Up",
  usage: "Usage",
  earning: "Earning",
  withdrawal: "Withdrawal",
  refund: "Refund",
  subsidy: "Subsidy",
};

const txTypeColors: Record<string, string> = {
  top_up: "text-emerald-400",
  usage: "text-red-400",
  earning: "text-blue-400",
  withdrawal: "text-amber-400",
  refund: "text-purple-400",
  subsidy: "text-cyan-400",
};

export default function WalletPage() {
  const [wallet, setWallet] = useState<Wallet | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [txTotal, setTxTotal] = useState(0);
  const [txPage, setTxPage] = useState(1);
  const [txFilter, setTxFilter] = useState<string>("");
  const [loading, setLoading] = useState(false);
  const [showTopUp, setShowTopUp] = useState(false);
  const [topUpLoading, setTopUpLoading] = useState(false);
  const [form] = Form.useForm();

  const loadWallet = async () => {
    try {
      const res = await walletApi.get();
      setWallet(res.data.data);
    } catch {
      // ignore
    }
  };

  const loadTransactions = async () => {
    setLoading(true);
    try {
      const res = await walletApi.getTransactions(
        txPage,
        20,
        txFilter || undefined,
      );
      setTransactions(res.data.data?.items || []);
      setTxTotal(res.data.data?.total || 0);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadWallet();
  }, []);

  useEffect(() => {
    loadTransactions();
  }, [txPage, txFilter]);

  const handleTopUp = async () => {
    try {
      const values = await form.validateFields();
      setTopUpLoading(true);
      const res = await walletApi.topUp(values.amount, values.channel);
      const data = res.data.data;
      if (data?.pay_url) {
        window.open(data.pay_url, "_blank");
      }
      message.success(
        "Payment initiated. Complete the payment in the opened window.",
      );
      setShowTopUp(false);
      form.resetFields();
      // Refresh after a short delay
      setTimeout(() => {
        loadWallet();
        loadTransactions();
      }, 3000);
    } catch (err: any) {
      message.error(err?.response?.data?.message || "Top up failed");
    } finally {
      setTopUpLoading(false);
    }
  };

  const balance = parseFloat(wallet?.balance || "0");
  const frozen = parseFloat(wallet?.frozen || "0");
  const available = balance - frozen;

  const containerVariants = {
    hidden: {},
    show: { transition: { staggerChildren: 0.08 } },
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 16 },
    show: {
      opacity: 1,
      y: 0,
      transition: { type: "spring" as const, damping: 25, stiffness: 300 },
    },
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold">Wallet</h1>
        <button
          onClick={() => setShowTopUp(true)}
          className="glass-btn h-10 px-4 rounded-xl text-sm gap-2"
        >
          <Plus size={16} />
          Top Up
        </button>
      </div>

      {/* Balance cards */}
      <motion.div
        className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8"
        variants={containerVariants}
        initial="hidden"
        animate="show"
      >
        <motion.div variants={itemVariants} className="glass rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-emerald-600 flex items-center justify-center">
              <WalletIcon size={18} className="text-white" />
            </div>
            <span className="text-xs text-white/40">Available Balance</span>
          </div>
          <p className="text-2xl font-semibold">
            {wallet?.currency === "CNY" ? "¥" : "$"}
            {available.toFixed(2)}
          </p>
        </motion.div>

        <motion.div variants={itemVariants} className="glass rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center">
              <CreditCard size={18} className="text-white" />
            </div>
            <span className="text-xs text-white/40">Total Balance</span>
          </div>
          <p className="text-2xl font-semibold">
            {wallet?.currency === "CNY" ? "¥" : "$"}
            {balance.toFixed(2)}
          </p>
        </motion.div>

        <motion.div variants={itemVariants} className="glass rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-500 to-orange-600 flex items-center justify-center">
              <History size={18} className="text-white" />
            </div>
            <span className="text-xs text-white/40">Frozen</span>
          </div>
          <p className="text-2xl font-semibold">
            {wallet?.currency === "CNY" ? "¥" : "$"}
            {frozen.toFixed(2)}
          </p>
        </motion.div>
      </motion.div>

      {/* Transactions */}
      <div className="glass rounded-2xl p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-medium">Transaction History</h2>
          <select
            value={txFilter}
            onChange={(e) => {
              setTxFilter(e.target.value);
              setTxPage(1);
            }}
            className="glass-input h-8 rounded-lg px-3 text-xs"
          >
            <option value="">All Types</option>
            <option value="top_up">Top Up</option>
            <option value="usage">Usage</option>
            <option value="earning">Earning</option>
            <option value="withdrawal">Withdrawal</option>
            <option value="refund">Refund</option>
          </select>
        </div>

        {loading ? (
          <div className="flex items-center justify-center h-32 text-white/30">
            Loading...
          </div>
        ) : transactions.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-white/30 text-sm">
            No transactions yet
          </div>
        ) : (
          <div className="space-y-2">
            {transactions.map((tx) => (
              <div
                key={tx.id}
                className="flex items-center gap-4 p-3 rounded-xl hover:bg-white/[0.04] transition-colors"
              >
                <div className="w-8 h-8 rounded-lg bg-white/[0.06] flex items-center justify-center flex-shrink-0">
                  {txTypeIcons[tx.type] || <ArrowUpRight size={14} />}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm truncate">
                    {tx.description || txTypeLabels[tx.type]}
                  </p>
                  <p className="text-xs text-white/30">
                    {new Date(tx.created_at).toLocaleString()}
                  </p>
                </div>
                <div className="text-right flex-shrink-0">
                  <p
                    className={`text-sm font-medium ${txTypeColors[tx.type] || "text-white/70"}`}
                  >
                    {parseFloat(tx.amount) >= 0 ? "+" : ""}
                    {parseFloat(tx.amount).toFixed(2)}
                  </p>
                  <p className="text-[10px] text-white/30">
                    Balance: {parseFloat(tx.balance_after).toFixed(2)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}

        {txTotal > 20 && (
          <div className="flex justify-center mt-4">
            <Pagination
              current={txPage}
              total={txTotal}
              pageSize={20}
              onChange={setTxPage}
              size="small"
            />
          </div>
        )}
      </div>

      {/* Top Up modal */}
      <Modal
        open={showTopUp}
        title="Top Up Wallet"
        okText={topUpLoading ? "Processing..." : "Pay"}
        onOk={handleTopUp}
        onCancel={() => !topUpLoading && setShowTopUp(false)}
        okButtonProps={{ disabled: topUpLoading }}
        destroyOnClose
      >
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item
            name="amount"
            label="Amount (CNY)"
            rules={[
              { required: true, message: "Enter amount" },
              { type: "number", min: 1, message: "Minimum ¥1" },
            ]}
          >
            <InputNumber
              className="w-full"
              placeholder="100"
              min={1}
              max={50000}
              precision={2}
              prefix="¥"
            />
          </Form.Item>
          <Form.Item
            name="channel"
            label="Payment Method"
            rules={[{ required: true, message: "Select payment method" }]}
          >
            <Select placeholder="Select payment method">
              <Select.Option value="wechat">WeChat Pay</Select.Option>
              <Select.Option value="alipay">Alipay</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

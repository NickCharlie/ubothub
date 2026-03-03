import http from "./http";

export interface Wallet {
  id: string;
  user_id: string;
  balance: string;
  frozen: string;
  currency: string;
  created_at: string;
}

export interface Transaction {
  id: string;
  user_id: string;
  wallet_id: string;
  type: "top_up" | "usage" | "earning" | "withdrawal" | "refund" | "subsidy";
  amount: string;
  balance_before: string;
  balance_after: string;
  status: "pending" | "completed" | "failed" | "cancelled";
  bot_id: string | null;
  external_order_id: string | null;
  payment_channel: string | null;
  description: string;
  created_at: string;
}

export interface BotPricing {
  id: string;
  bot_id: string;
  mode: "free" | "per_call" | "subscribe";
  price_per_call: string;
  monthly_price: string;
  platform_rate: string;
  creator_rate: string;
  free_calls_per_day: number;
}

export const walletApi = {
  get: () => http.get<{ data: Wallet }>("/wallet"),
  topUp: (amount: number, channel: "wechat" | "alipay") =>
    http.post("/wallet/topup", { amount, channel }),
  getTransactions: (page = 1, pageSize = 20, type?: string) =>
    http.get("/wallet/transactions", { params: { page, page_size: pageSize, type } }),
  getBotPricing: (botId: string) =>
    http.get<{ data: BotPricing }>(`/bots/${botId}/pricing`),
  setBotPricing: (botId: string, data: Partial<BotPricing>) =>
    http.put(`/bots/${botId}/pricing`, data),
};

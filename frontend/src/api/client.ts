// Typed API client. All fetch calls go through here so auth headers and
// base URL are applied consistently.

const BASE = import.meta.env.VITE_API_URL ?? "";

export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  provider: string;
  role: string;
  paid_at: string | null;
  created_at: string;
}

export interface PaymentIntentResponse {
  client_secret: string;
  payment_id: string;
}

export interface UserListResponse {
  users: User[];
  total: number;
}

export interface AssumeResponse {
  token: string;
  target_user: User;
}

// ── Token Management ──────────────────────────────────────────────────────────

const TOKEN_KEY = "jwt";

export const getToken = (): string | null => localStorage.getItem(TOKEN_KEY);
export const setToken = (t: string) => localStorage.setItem(TOKEN_KEY, t);
export const clearToken = () => localStorage.removeItem(TOKEN_KEY);

// ── Core Fetch ────────────────────────────────────────────────────────────────

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${BASE}${path}`, { ...options, headers, credentials: "include" });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? `HTTP ${res.status}`);
  }

  if (res.status === 204) return undefined as unknown as T;
  return res.json();
}

// ── Auth ──────────────────────────────────────────────────────────────────────

export const getMe = () => request<User>("/api/v1/me");

export const logout = () =>
  request<void>("/auth/logout", { method: "POST" }).then(clearToken);

export const oauthURL = (provider: string) => `${BASE}/auth/${provider}`;

// ── Admin ─────────────────────────────────────────────────────────────────────

export const listUsers = (limit = 50, offset = 0) =>
  request<UserListResponse>(`/api/v1/admin/users?limit=${limit}&offset=${offset}`);

export const assumeUser = (id: string) =>
  request<AssumeResponse>(`/api/v1/admin/users/${id}/assume`, { method: "POST" });

// ── Payments ──────────────────────────────────────────────────────────────────

export const createPaymentIntent = (amount: number, currency = "usd") =>
  request<PaymentIntentResponse>("/api/v1/payments/intent", {
    method: "POST",
    body: JSON.stringify({ amount, currency }),
  });

// ── Agent ─────────────────────────────────────────────────────────────────────

export const sendPrompt = (prompt: string) =>
  request<{ response: string }>("/api/v1/agent/prompt", {
    method: "POST",
    body: JSON.stringify({ prompt }),
  });

export const streamURL = (prompt: string) =>
  `${BASE}/api/v1/agent/stream?prompt=${encodeURIComponent(prompt)}`;

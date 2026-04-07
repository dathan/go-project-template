import React, { createContext, useCallback, useContext, useEffect, useState } from "react";
import * as api from "../api/client";

interface AuthState {
  user: api.User | null;
  token: string | null;
  loading: boolean;
  // When an admin assumes a role, impersonatedUser is the target user and
  // user remains the admin (for UI context).
  impersonatedUser: api.User | null;
}

interface AuthContextValue extends AuthState {
  login: (token: string) => Promise<void>;
  logout: () => void;
  assume: (targetId: string) => Promise<void>;
  exitAssume: () => void;
  isAdmin: () => boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>({
    user: null,
    token: api.getToken(),
    loading: true,
    impersonatedUser: null,
  });

  // Fetch the current user on mount / token change
  useEffect(() => {
    if (!state.token) {
      setState((s) => ({ ...s, loading: false, user: null }));
      return;
    }
    api
      .getMe()
      .then((user) => setState((s) => ({ ...s, user, loading: false })))
      .catch(() => {
        api.clearToken();
        setState((s) => ({ ...s, token: null, user: null, loading: false }));
      });
  }, [state.token]);

  const login = useCallback(async (token: string) => {
    api.setToken(token);
    const user = await api.getMe();
    setState((s) => ({ ...s, token, user, loading: false }));
  }, []);

  const logout = useCallback(() => {
    api.logout();
    setState({ user: null, token: null, loading: false, impersonatedUser: null });
  }, []);

  const assume = useCallback(async (targetId: string) => {
    const { token, target_user } = await api.assumeUser(targetId);
    api.setToken(token);
    setState((s) => ({ ...s, token, impersonatedUser: target_user }));
  }, []);

  const exitAssume = useCallback(() => {
    // Restore original admin token — for simplicity, re-login is required.
    // In production you'd store the original token separately.
    logout();
  }, [logout]);

  const isAdmin = useCallback(
    () => state.user?.role === "admin",
    [state.user]
  );

  return (
    <AuthContext.Provider value={{ ...state, login, logout, assume, exitAssume, isAdmin }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside <AuthProvider>");
  return ctx;
}

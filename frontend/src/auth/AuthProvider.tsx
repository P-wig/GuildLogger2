/**

Application auth state provider.
Responsibilities:
Hydrate auth state from localStorage on app startup.
Persist login session with a fixed TTL (24 hours).
Validate and clear invalid/expired session data.
Keep all open tabs in sync via BroadcastChannel + storage events.
Expose auth state and actions through AuthContext:
user, isAuthenticated, loading, login, logout.
Validate stored JWT against backend on startup; set loading until resolved.
Notes:
Session data is base64-encoded for transport/storage convenience, not encryption.
localStorage is shared per browser profile and can be cleared by the user.
*/
import React, { useEffect, useMemo, useState } from "react";
import { AuthContext } from "./authContext";
import {
  authChannel,
  broadcastAuthChanged,
  DEFAULT_SESSION_EXPIRED_MESSAGE,
  SESSION_EXPIRED_EVENT,
  type SessionExpiredDetail,
} from "./authSync";
import type { User } from "../api/auth";
import { authApi } from "../api/auth";

const SESSION_KEY = "session_token";
const SESSION_TTL_MS = 24 * 60 * 60 * 1000; // 24 hours

interface StoredSession {
  user: User;
  expiresAt: number;
}

function encodeSession(session: StoredSession): string {
  return btoa(JSON.stringify(session));
}

function decodeSession(raw: string): StoredSession | null {
  try {
    return JSON.parse(atob(raw)) as StoredSession;
  } catch {
    return null;
  }
}

function persistSession(user: User) {
  const session: StoredSession = {
    user,
    expiresAt: Date.now() + SESSION_TTL_MS,
  };
  // This is shared across the same Chrome profile
  localStorage.setItem(SESSION_KEY, encodeSession(session));
}

function clearSession() {
  localStorage.removeItem(SESSION_KEY);
  localStorage.removeItem("authToken"); // clear JWT on logout
}

function loadSession(): User | null {
  const raw = localStorage.getItem(SESSION_KEY);
  if (!raw) return null;

  const session = decodeSession(raw);
  if (!session) {
    clearSession();
    return null;
  }

  if (Date.now() > session.expiresAt) {
    clearSession();
    return null;
  }

  return session.user;
}

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<User | null>(() => loadSession());
  const [loading, setLoading] = useState(true);
  const [sessionExpired, setSessionExpired] = useState<string | null>(null);

  // Validate stored token against backend on startup.
  // loading stays true until this resolves so ProtectedRoute doesn't flash-redirect.
  useEffect(() => {
    const token = localStorage.getItem("authToken");
    if (!token) {
      setLoading(false);
      return;
    }

    authApi
      .getSession()
      .then((res) => {
        persistSession(res.data.user);
        setUser(res.data.user);
      })
      .catch(() => {
        clearSession();
        setUser(null);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  // React to a 401 in this tab. http.ts has already cleared storage, but neither the
  // storage event nor BroadcastChannel fires in the tab that caused it, so without this
  // the user would sit on a protected page that silently fails every request.
  useEffect(() => {
    const onExpired = (e: Event) => {
      const detail = (e as CustomEvent<SessionExpiredDetail>).detail;
      clearSession();
      setUser(null);
      setSessionExpired(detail?.message ?? DEFAULT_SESSION_EXPIRED_MESSAGE);
      broadcastAuthChanged();
    };

    window.addEventListener(SESSION_EXPIRED_EVENT, onExpired);
    return () => window.removeEventListener(SESSION_EXPIRED_EVENT, onExpired);
  }, []);

  // Keep all tabs/windows in sync
  useEffect(() => {
    const sync = () => setUser(loadSession());

    const onStorage = (e: StorageEvent) => {
      if (e.key === SESSION_KEY || e.key === "__auth_changed_at__") sync();
    };

    window.addEventListener("storage", onStorage);

    const onMessage = (e: MessageEvent) => {
      if (e.data?.type === "AUTH_CHANGED") sync();
    };

    authChannel?.addEventListener("message", onMessage);

    return () => {
      window.removeEventListener("storage", onStorage);
      authChannel?.removeEventListener("message", onMessage);
    };
  }, []);

  const value = useMemo(() => {
    return {
      user,
      isAuthenticated: !!user,
      loading,
      sessionExpired,
      login: (u: User) => {
        persistSession(u);
        setUser(u);
        setSessionExpired(null);
        broadcastAuthChanged();
      },
      logout: () => {
        authApi.logout().catch(() => {}); // fire-and-forget — backend is stateless
        clearSession();
        setUser(null);
        setSessionExpired(null); // signing out deliberately is not an expiry
        broadcastAuthChanged();
      },
    };
  }, [user, loading, sessionExpired]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

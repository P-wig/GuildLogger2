/**

Application auth state provider.
Responsibilities:
Hydrate auth state from localStorage on app startup.
Persist login session with a fixed TTL (24 hours).
Validate and clear invalid/expired session data.
Keep all open tabs in sync via BroadcastChannel + storage events.
Expose auth state and actions through AuthContext:
user, isAuthenticated, loading, login, logout.
Notes:
Session data is base64-encoded for transport/storage convenience, not encryption.
localStorage is shared per browser profile and can be cleared by the user.
*/
import React, { useEffect, useMemo, useState } from "react";
import { AuthContext } from "./authContext";
import { authChannel, broadcastAuthChanged } from "./authSync";
import type { User } from "../api/auth";

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
  const loading = false;

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
      login: (u: User) => {
        persistSession(u);
        setUser(u);
        broadcastAuthChanged();
      },
      logout: () => {
        clearSession();
        setUser(null);
        broadcastAuthChanged();
      },
    };
  }, [user, loading]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

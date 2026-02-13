import React, { useMemo, useState } from "react";
import { AuthContext } from "./authContext";
import type { User } from "../api/users";

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
  localStorage.setItem(SESSION_KEY, encodeSession(session));
}

function clearSession() {
  localStorage.removeItem(SESSION_KEY);
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
  const [loading] = useState(false);

  const value = useMemo(() => {
    return {
      user,
      isAuthenticated: !!user,
      loading,
      login: (u: User) => {
        setUser(u);
        persistSession(u);
      },
      logout: () => {
        setUser(null);
        clearSession();
      },
    };
  }, [user, loading]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

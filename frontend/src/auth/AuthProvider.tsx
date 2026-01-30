import React, { useMemo, useState } from "react";
import { AuthContext } from "./authContext";

export type AuthUser = {
  id: string;
  email: string;
};

type AuthContextValue = {
  user: AuthUser | null;
  isAuthenticated: boolean;
  login: (email: string) => void;
  logout: () => void;
};

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  // TODO: replace with real auth token
  const [user, setUser] = useState<AuthUser | null>(null);

  const value = useMemo<AuthContextValue>(() => {
    return {
      user,
      isAuthenticated: !!user,
      login: (email: string) => setUser({ id: "testUser", email }),
      logout: () => setUser(null),
    };
  }, [user]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

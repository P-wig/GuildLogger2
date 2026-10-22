import { createContext } from "react";
import type { User } from "../api/auth";

export type AuthContextValue = {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  /** Message to show on the sign-in page when the session ended involuntarily; null otherwise. */
  sessionExpired: string | null;
  login: (user: User) => void;
  logout: () => void;
};

export const AuthContext = createContext<AuthContextValue | undefined>(
  undefined,
);

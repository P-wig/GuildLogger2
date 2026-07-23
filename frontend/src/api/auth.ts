import { api } from "./http";

export type User = {
  discordId: string;
  createdAt: string;
  updatedAt: string;
};

export type DiscordLoginRequest = {
  code: string;
  redirectUri: string;
};

export type DiscordAuthURLResponse = {
  ok: boolean;
  url: string;
};

export type LoginResponse = {
  ok: boolean;
  token: string;
  user: User;
};

export type SessionResponse = {
  ok: boolean;
  user: User;
};

export const authApi = {
  getDiscordAuthURL: (redirectUri: string, state?: string) =>
    api.get<DiscordAuthURLResponse>("/auth/discord/url", {
      params: { redirectUri, ...(state !== undefined ? { state } : {}) },
    }),

  discordLogin: (payload: DiscordLoginRequest) =>
    api.post<LoginResponse>("/auth/discord/login", payload),

  getSession: () => api.get<SessionResponse>("/auth/session"),

  logout: () => api.post<{ ok: boolean }>("/auth/logout"),
};
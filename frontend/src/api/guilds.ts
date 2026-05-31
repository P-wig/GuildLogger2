import { api } from "./http";

export type LinkedGuild = {
  guildId: string;
  name: string;
  ownerDiscordId: string;
  botInstalled: boolean;
  createdAt: string;
  updatedAt: string;
};

export type DiscordGuild = {
  id: string;
  name: string;
  icon: string;
};

export type ConnectGuildRequest = {
  guildId: string;
  name: string;
};

export type GuildDashboardData = {
  guild: LinkedGuild;
  memberCount: number;
  eventCount: number;
};

export type SyncStatus = {
  memberCount: number;
  synced: boolean;
};

export const guildsApi = {
  getMyGuilds: () =>
    api.get<{ ok: boolean; guilds: LinkedGuild[] }>("/guilds"),

  getDiscordGuilds: () =>
    api.get<{ ok: boolean; guilds: DiscordGuild[] }>("/guilds/discord"),

  connectGuild: (payload: ConnectGuildRequest) =>
    api.post<{ ok: boolean; guild: LinkedGuild }>("/guilds/connect", payload),
  getBotInviteUrl: (guildId: string, redirectUri: string) =>
    api.get<{ ok: boolean; url: string }>(`/guilds/${guildId}/bot/invite-url`, {
      params: { redirectUri },
    }),
  installBot: (guildId: string) =>
    api.post<{ ok: boolean }>(`/guilds/${guildId}/bot/install`),
  verifyBotInstall: (guildId: string) =>
    api.post<{ ok: boolean }>(`/guilds/${guildId}/bot/verify`),
  getDashboard: (guildId: string) =>
    api.get<{ ok: boolean; dashboard: GuildDashboardData }>(
      `/guilds/${guildId}/dashboard`
    ),
  getMemberSyncStatus: (guildId: string) =>
    api.get<{ ok: boolean; memberCount: number; synced: boolean }>(
      `/guilds/${guildId}/members/sync-status`
    ),
};
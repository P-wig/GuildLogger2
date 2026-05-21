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

export const guildsApi = {
  getMyGuilds: () =>
    api.get<{ ok: boolean; guilds: LinkedGuild[] }>("/guilds"),

  getDiscordGuilds: () =>
    api.get<{ ok: boolean; guilds: DiscordGuild[] }>("/guilds/discord"),

  connectGuild: (payload: ConnectGuildRequest) =>
    api.post<{ ok: boolean; guild: LinkedGuild }>("/guilds/connect", payload),
};
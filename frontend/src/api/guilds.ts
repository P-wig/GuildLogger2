import { api } from "./http";

export type GuildRole = {
  discordRoleId: string;
  name: string;
  position: number;
  type: string;
  appPermissions: string[];
  managed: boolean;
  isDefault: boolean;
};

export type LinkedGuild = {
  guildId: string;
  name: string;
  ownerDiscordId: string;
  botInstalled: boolean;
  roles: GuildRole[];
  notificationConfig: {
    statusRoles: {
      activeRoleId: string;
      inactiveRoleId: string;
    };
    milestoneNotifications: {
      enabled: boolean;
    };
  };
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

export type GuildDashboardLeaderboardEntry = {
  discordId: string;
  eventsHosted: number;
  eventsAttended: number;
  score: number;
  lastHostedDate?: string;
  lastAttendedDate?: string;
  rank: number;
};

export type GuildDashboardMemberRow = {
  discordId: string;
  username: string;
  avatarHash: string;
  rankedRoleId: string;
  status: "active" | "inactive";
  discordJoinedAt: string;
  roleIds: string[];
  eventsHosted: number;
  eventsAttended: number;
  lastHostedDate?: string;
  lastAttendedDate?: string;
  isInactiveByCutoff: boolean;
};

export type GuildDashboardInactiveMember = {
  discordId: string;
  rankedRoleId: string;
  status: "active" | "inactive";
  discordJoinedAt: string;
  lastHostedDate?: string;
  lastAttendedDate?: string;
  lastActivityDate?: string;
  daysSinceActivity?: number;
};

export type GuildDashboardStats = {
  totalMembers: number;
  activeMembers: number;
  inactiveMembers: number;
  totalEvents: number;
  closedEvents: number;
  participationRate: number;
};

export type GuildDashboardEventRow = {
  eventId: string;
  hostDiscordId: string;
  eventDate: string;
  participantIds: string[];
  summary: string;
};

export type GuildDashboardResponse = {
  guild: LinkedGuild;
  stats: GuildDashboardStats;
  leaderboard: GuildDashboardLeaderboardEntry[];
  members: GuildDashboardMemberRow[];
  inactiveMembers: GuildDashboardInactiveMember[];
  events: GuildDashboardEventRow[];
};

export type GuildDashboardData = GuildDashboardResponse;

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
    api.get<{ ok: boolean; dashboard: GuildDashboardResponse }>(
      `/guilds/${guildId}/dashboard`
    ),
  getMemberSyncStatus: (guildId: string) =>
    api.get<{ ok: boolean; memberCount: number; synced: boolean }>(
      `/guilds/${guildId}/members/sync-status`
    ),
  syncMembers: (guildId: string) =>
    api.post<{ ok: boolean; synced: number }>(`/guilds/${guildId}/members/sync`),
  setMemberRole: (guildId: string, roleId: string) =>
    api.put<{ ok: boolean }>(`/guilds/${guildId}/config/member-role`, { roleId }),
  updateGuildConfig: (guildId: string, config: { activeRoleId: string; inactiveRoleId?: string; rankedRoleIds?: string[] }) =>
    api.put<{ ok: boolean }>(`/guilds/${guildId}/config`, config),
};
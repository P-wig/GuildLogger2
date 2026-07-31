import { useEffect, useMemo, useState } from "react";
import type { AxiosError } from "axios";
import { useParams, Link as RouterLink, useNavigate } from "react-router";
import Alert from "@mui/material/Alert";
import Avatar from "@mui/material/Avatar";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import LoadingButton from "@mui/lab/LoadingButton";
import Card from "@mui/material/Card";
import CardActionArea from "@mui/material/CardActionArea";
import CardContent from "@mui/material/CardContent";
import CircularProgress from "@mui/material/CircularProgress";
import Chip from "@mui/material/Chip";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import Divider from "@mui/material/Divider";
import FormControl from "@mui/material/FormControl";
import IconButton from "@mui/material/IconButton";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Stack from "@mui/material/Stack";
import Tab from "@mui/material/Tab";
import Tabs from "@mui/material/Tabs";
import TextField from "@mui/material/TextField";
import InputAdornment from "@mui/material/InputAdornment";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import { useTheme } from "@mui/material/styles";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ExpandLessIcon from "@mui/icons-material/ExpandLess";
import InfoIcon from "@mui/icons-material/Info";
import SearchIcon from "@mui/icons-material/Search";
import SettingsIcon from "@mui/icons-material/Settings";
import {
  BarChart, Bar, XAxis, YAxis, Tooltip as RechartsTooltip,
  Legend, ResponsiveContainer, PieChart, Pie, Cell,
} from "recharts";
import { guildsApi, type GuildDashboardData, type GuildDashboardMemberRow, type LinkedGuild, type GuildRole, type SyncStatus } from "../api/guilds";
import { useAuth } from "../auth";

type ConfigSavedUpdates = {
  activeRoleId: string;
  inactiveRoleId: string;
  moderatorRoleIds: string[];
  rankedRoleIds: string[];
  eventsChannelId: string;
  eventTypes: string[];
};

const ConfigDialog = ({
  open,
  guild,
  guildId,
  selectableRoles,
  onClose,
  onSaved,
}: {
  open: boolean;
  guild: LinkedGuild | null;
  guildId: string;
  selectableRoles: GuildRole[];
  onClose: () => void;
  onSaved: (updates: ConfigSavedUpdates) => void;
}) => {
  const [cfgActiveRoleId, setCfgActiveRoleId] = useState("");
  const [cfgInactiveRoleId, setCfgInactiveRoleId] = useState("");
  const [cfgRankedRoleIds, setCfgRankedRoleIds] = useState<string[]>([]);
  const [cfgModeratorRoleIds, setCfgModeratorRoleIds] = useState<string[]>([]);
  const [cfgEventsChannelId, setCfgEventsChannelId] = useState("");
  const [cfgEventTypes, setCfgEventTypes] = useState<string[]>([]);
  const [cfgNewEventType, setCfgNewEventType] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Initialise state from guild every time the dialog opens.
  useEffect(() => {
    if (!open || !guild) return;
    setCfgActiveRoleId(guild.notificationConfig?.statusRoles?.activeRoleId ?? "");
    setCfgInactiveRoleId(guild.notificationConfig?.statusRoles?.inactiveRoleId ?? "");
    setCfgRankedRoleIds(guild.roles?.filter((r) => r.type === "ranked").map((r) => r.discordRoleId) ?? []);
    setCfgModeratorRoleIds(guild.notificationConfig?.statusRoles?.moderatorRoleIds ?? []);
    setCfgEventsChannelId(guild.eventConfig?.eventsChannelId ?? "");
    setCfgEventTypes(guild.eventConfig?.eventTypes ?? []);
    setCfgNewEventType("");
    setError(null);
  }, [open, guild]);

  const handleSave = async () => {
    if (!guildId || !cfgActiveRoleId) return;
    setSaving(true);
    setError(null);
    try {
      await Promise.all([
        guildsApi.updateGuildConfig(guildId, {
          activeRoleId: cfgActiveRoleId,
          inactiveRoleId: cfgInactiveRoleId,
          rankedRoleIds: cfgRankedRoleIds,
          moderatorRoleIds: cfgModeratorRoleIds,
        }),
        guildsApi.updateEventConfig(guildId, {
          eventsChannelId: cfgEventsChannelId,
          eventTypes: cfgEventTypes,
        }),
      ]);
      onSaved({
        activeRoleId: cfgActiveRoleId,
        inactiveRoleId: cfgInactiveRoleId,
        moderatorRoleIds: cfgModeratorRoleIds,
        rankedRoleIds: cfgRankedRoleIds,
        eventsChannelId: cfgEventsChannelId,
        eventTypes: cfgEventTypes,
      });
      onClose();
    } catch (err) {
      const axiosErr = err as AxiosError<{ error?: string }>;
      setError(axiosErr.response?.data?.error ?? "Failed to save configuration. Please try again.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Guild Configuration</DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          {selectableRoles.length === 0 && (
            <Alert severity="info">No roles available. Re-verify the bot to refresh the role list.</Alert>
          )}
          {selectableRoles.length > 0 && selectableRoles.every((r) => !r.name) && (
            <Alert severity="warning">
              Role names are missing — re-verify the bot from the Guilds page to refresh them.
            </Alert>
          )}

          {/* Active Member Role */}
          <Box>
            <Typography variant="subtitle2" gutterBottom>Active Member Role *</Typography>
            <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1 }}>
              Members holding this role are counted and synced. Required before syncing.
            </Typography>
            <FormControl size="small" fullWidth disabled={selectableRoles.length === 0}>
              <InputLabel>Active Role</InputLabel>
              <Select
                value={cfgActiveRoleId}
                label="Active Role"
                onChange={(e) => setCfgActiveRoleId(e.target.value)}
              >
                {selectableRoles.map((r) => (
                  <MenuItem key={r.discordRoleId} value={r.discordRoleId}>
                    {r.name || r.discordRoleId}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Box>

          {/* Inactive Member Role */}
          <Box>
            <Typography variant="subtitle2" gutterBottom>Inactive Member Role</Typography>
            <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1 }}>
              Role that marks members below the activity threshold. Optional.
            </Typography>
            <FormControl size="small" fullWidth disabled={selectableRoles.length === 0}>
              <InputLabel>Inactive Role</InputLabel>
              <Select
                value={cfgInactiveRoleId}
                label="Inactive Role"
                onChange={(e) => setCfgInactiveRoleId(e.target.value)}
              >
                <MenuItem value=""><em>None</em></MenuItem>
                {selectableRoles.map((r) => (
                  <MenuItem key={r.discordRoleId} value={r.discordRoleId}>
                    {r.name || r.discordRoleId}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Box>

          {/* Ranked Roles */}
          <Box>
            <Typography variant="subtitle2" gutterBottom>Ranked Roles</Typography>
            <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1 }}>
              Roles representing member ranks used for leaderboard sorting. Toggle all that apply.
            </Typography>
            {selectableRoles.length === 0 ? (
              <Typography variant="body2" color="text.secondary">No roles available.</Typography>
            ) : (
              <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                {selectableRoles.map((r) => (
                  <Chip
                    key={r.discordRoleId}
                    label={r.name || r.discordRoleId}
                    onClick={() =>
                      setCfgRankedRoleIds((prev) =>
                        prev.includes(r.discordRoleId)
                          ? prev.filter((id) => id !== r.discordRoleId)
                          : [...prev, r.discordRoleId]
                      )
                    }
                    color={cfgRankedRoleIds.includes(r.discordRoleId) ? "primary" : "default"}
                    variant={cfgRankedRoleIds.includes(r.discordRoleId) ? "filled" : "outlined"}
                  />
                ))}
              </Box>
            )}
          </Box>

          {/* Moderator Roles */}
          <Box>
            <Typography variant="subtitle2" gutterBottom>Moderator Roles</Typography>
            <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1 }}>
              Members holding any of these roles can create, edit, and delete event logs. Toggle all that apply.
            </Typography>
            {selectableRoles.length === 0 ? (
              <Typography variant="body2" color="text.secondary">No roles available.</Typography>
            ) : (
              <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                {selectableRoles.map((r) => (
                  <Chip
                    key={r.discordRoleId}
                    label={r.name || r.discordRoleId}
                    onClick={() =>
                      setCfgModeratorRoleIds((prev) =>
                        prev.includes(r.discordRoleId)
                          ? prev.filter((id) => id !== r.discordRoleId)
                          : [...prev, r.discordRoleId]
                      )
                    }
                    color={cfgModeratorRoleIds.includes(r.discordRoleId) ? "secondary" : "default"}
                    variant={cfgModeratorRoleIds.includes(r.discordRoleId) ? "filled" : "outlined"}
                  />
                ))}
              </Box>
            )}
          </Box>

          {/* Events Channel */}
          <Box>
            <Typography variant="subtitle2" gutterBottom>Events Channel</Typography>
            <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1 }}>
              Discord channel ID where the bot posts event RSVP announcements. Right-click the channel → Copy Channel ID.
            </Typography>
            <TextField
              size="small"
              fullWidth
              value={cfgEventsChannelId}
              onChange={(e) => setCfgEventsChannelId(e.target.value)}
              placeholder="e.g. 1234567890123456789"
            />
          </Box>

          {/* Event Types */}
          <Box>
            <Typography variant="subtitle2" gutterBottom>Event Types</Typography>
            <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1 }}>
              Types available in the /start event slash command. Press Enter or click Add.
            </Typography>
            <Stack direction="row" spacing={1} sx={{ mb: 1 }}>
              <TextField
                size="small"
                value={cfgNewEventType}
                onChange={(e) => setCfgNewEventType(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && cfgNewEventType.trim() && !cfgEventTypes.includes(cfgNewEventType.trim())) {
                    setCfgEventTypes((prev) => [...prev, cfgNewEventType.trim()]);
                    setCfgNewEventType("");
                  }
                }}
                placeholder="e.g. Raid Night"
                sx={{ flex: 1 }}
              />
              <Button
                size="small"
                variant="outlined"
                disabled={!cfgNewEventType.trim() || cfgEventTypes.includes(cfgNewEventType.trim())}
                onClick={() => {
                  setCfgEventTypes((prev) => [...prev, cfgNewEventType.trim()]);
                  setCfgNewEventType("");
                }}
              >
                Add
              </Button>
            </Stack>
            {cfgEventTypes.length > 0 && (
              <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                {cfgEventTypes.map((t) => (
                  <Chip
                    key={t}
                    label={t}
                    size="small"
                    onDelete={() => setCfgEventTypes((prev) => prev.filter((x) => x !== t))}
                  />
                ))}
              </Box>
            )}
          </Box>

          {error && <Alert severity="error">{error}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <LoadingButton
          variant="contained"
          loading={saving}
          onClick={handleSave}
          disabled={!cfgActiveRoleId}
        >
          Save
        </LoadingButton>
      </DialogActions>
    </Dialog>
  );
};

export const GuildDashboard = () => {
  const { guildId } = useParams<{ guildId: string }>();
  const { user } = useAuth();
  const theme = useTheme();

  const [dashboard, setDashboard] = useState<GuildDashboardData | null>(null);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [syncing, setSyncing] = useState(false);
  const [syncError, setSyncError] = useState<string | null>(null);
  const [syncCooldown, setSyncCooldown] = useState(0);
  const [statsModalMember, setStatsModalMember] = useState<GuildDashboardMemberRow | null>(null);
  const [configOpen, setConfigOpen] = useState(false);
  const [disconnectOpen, setDisconnectOpen] = useState(false);
  const [disconnecting, setDisconnecting] = useState(false);
  const [disconnectConfirm, setDisconnectConfirm] = useState("");
  const [memberSearch, setMemberSearch] = useState("");
  const [membersExpanded, setMembersExpanded] = useState(false);
  const [rankTab, setRankTab] = useState("");
  const navigate = useNavigate();

  // Restore sync cooldown from localStorage on mount
  useEffect(() => {
    if (!guildId) return;
    const until = Number(localStorage.getItem(`syncCooldownUntil_${guildId}`) ?? 0);
    const remaining = Math.max(0, Math.ceil((until - Date.now()) / 1000));
    if (remaining > 0) setSyncCooldown(remaining);
  }, [guildId]);

  // Countdown tick for sync cooldown
  useEffect(() => {
    if (syncCooldown <= 0) return;
    const timer = setInterval(() => setSyncCooldown((c) => c - 1), 1000);
    return () => clearInterval(timer);
  }, [syncCooldown]);

  // Member lookup map for leaderboard enrichment
  const memberMap = useMemo(() => {
    const m = new Map<string, GuildDashboardMemberRow>();
    (dashboard?.members ?? []).forEach((mbr) => m.set(mbr.discordId, mbr));
    return m;
  }, [dashboard?.members]);

  // Chart data: member status breakdown
  const memberStatusData = useMemo(() => {
    const s = dashboard?.stats;
    if (!s) return [];
    return [
      { name: "Active", value: s.activeMembers },
      { name: "Inactive", value: s.inactiveMembers },
    ].filter((d) => d.value > 0);
  }, [dashboard?.stats]);

  // Chart data: events per calendar month (from recent events list)
  const eventsByMonth = useMemo(() => {
    const evts = dashboard?.events ?? [];
    const byMonth: Record<string, { label: string; events: number; participants: number }> = {};
    evts.forEach((e) => {
      const key = e.eventDate.slice(0, 7);
      const label = new Date(e.eventDate).toLocaleDateString("en-US", {
        month: "short", year: "2-digit", timeZone: "UTC",
      });
      if (!byMonth[key]) byMonth[key] = { label, events: 0, participants: 0 };
      byMonth[key].events++;
      byMonth[key].participants += e.participantIds?.length ?? 0;
    });
    return Object.keys(byMonth).sort().map((k) => ({
      month: byMonth[k].label,
      Events: byMonth[k].events,
      "Avg Attendance": byMonth[k].events > 0 ? Math.round(byMonth[k].participants / byMonth[k].events) : 0,
    }));
  }, [dashboard?.events]);

  // Monthly top contributors grouped by rank (computed from recent dashboard events)
  const monthlyTopByRank = useMemo(() => {
    const now = new Date();
    const monthEvents = (dashboard?.events ?? []).filter((e) => {
      const d = new Date(e.eventDate);
      return d.getUTCFullYear() === now.getUTCFullYear() && d.getUTCMonth() === now.getUTCMonth();
    });
    if (monthEvents.length === 0) return [];

    const statsMap: Record<string, { hosted: number; attended: number }> = {};
    monthEvents.forEach((e) => {
      statsMap[e.hostDiscordId] = statsMap[e.hostDiscordId] ?? { hosted: 0, attended: 0 };
      statsMap[e.hostDiscordId].hosted++;
      (e.participantIds ?? []).forEach((pid) => {
        statsMap[pid] = statsMap[pid] ?? { hosted: 0, attended: 0 };
        statsMap[pid].attended++;
      });
    });

    const roles = dashboard?.guild.roles ?? [];
    const rolePositionMap = new Map(roles.map((r) => [r.discordRoleId, r.position]));
    const roleNameMap = new Map(roles.map((r) => [r.discordRoleId, r.name]));

    const grouped: Record<string, Array<{ discordId: string; hosted: number; attended: number; score: number }>> = {};
    Object.entries(statsMap).forEach(([discordId, s]) => {
      const roleId = memberMap.get(discordId)?.rankedRoleId || "__unranked__";
      grouped[roleId] = grouped[roleId] ?? [];
      grouped[roleId].push({ discordId, hosted: s.hosted, attended: s.attended, score: s.hosted * 2 + s.attended });
    });
    Object.values(grouped).forEach((arr) => arr.sort((a, b) => b.score - a.score));

    return Object.keys(grouped)
      .sort((a, b) => {
        if (a === "__unranked__") return 1;
        if (b === "__unranked__") return -1;
        return (rolePositionMap.get(b) ?? 0) - (rolePositionMap.get(a) ?? 0);
      })
      .map((roleId) => ({
        roleId,
        roleName: roleId === "__unranked__" ? "Unranked" : (roleNameMap.get(roleId) || roleId),
        entries: grouped[roleId],
      }));
  }, [dashboard?.events, dashboard?.guild.roles, memberMap]);

  // All-time lowest contributors (active members, score asc)
  const allTimeLow = useMemo(() => {
    const activeMembers = (dashboard?.members ?? []).filter((m) => m.status === "active");
    return [...activeMembers]
      .sort((a, b) => {
        const scoreA = a.eventsHosted * 2 + a.eventsAttended;
        const scoreB = b.eventsHosted * 2 + b.eventsAttended;
        if (scoreA !== scoreB) return scoreA - scoreB;
        return new Date(a.discordJoinedAt).getTime() - new Date(b.discordJoinedAt).getTime();
      })
      .slice(0, 10);
  }, [dashboard?.members]);

  // Monthly lowest contributors grouped by rank (all active members, score asc, zeroes included)
  const monthlyLowByRank = useMemo(() => {
    const allActive = dashboard?.members?.filter((m) => m.status === "active") ?? [];
    if (allActive.length === 0) return [];
    const now = new Date();
    const monthEvents = (dashboard?.events ?? []).filter((e) => {
      const d = new Date(e.eventDate);
      return d.getUTCFullYear() === now.getUTCFullYear() && d.getUTCMonth() === now.getUTCMonth();
    });
    const statsMap: Record<string, { hosted: number; attended: number }> = {};
    allActive.forEach((m) => { statsMap[m.discordId] = { hosted: 0, attended: 0 }; });
    monthEvents.forEach((e) => {
      if (statsMap[e.hostDiscordId] !== undefined) statsMap[e.hostDiscordId].hosted++;
      (e.participantIds ?? []).forEach((pid) => {
        if (statsMap[pid] !== undefined) statsMap[pid].attended++;
      });
    });
    const roles = dashboard?.guild.roles ?? [];
    const rolePositionMap = new Map(roles.map((r) => [r.discordRoleId, r.position]));
    const roleNameMap = new Map(roles.map((r) => [r.discordRoleId, r.name]));
    const grouped: Record<string, Array<{ discordId: string; hosted: number; attended: number; score: number }>> = {};
    Object.entries(statsMap).forEach(([discordId, s]) => {
      const roleId = memberMap.get(discordId)?.rankedRoleId || "__unranked__";
      grouped[roleId] = grouped[roleId] ?? [];
      grouped[roleId].push({ discordId, hosted: s.hosted, attended: s.attended, score: s.hosted * 2 + s.attended });
    });
    Object.values(grouped).forEach((arr) =>
      arr.sort((a, b) => {
        if (a.score !== b.score) return a.score - b.score;
        const mA = memberMap.get(a.discordId);
        const mB = memberMap.get(b.discordId);
        return new Date(mA?.discordJoinedAt ?? 0).getTime() - new Date(mB?.discordJoinedAt ?? 0).getTime();
      })
    );
    return Object.keys(grouped)
      .sort((a, b) => {
        if (a === "__unranked__") return 1;
        if (b === "__unranked__") return -1;
        return (rolePositionMap.get(b) ?? 0) - (rolePositionMap.get(a) ?? 0);
      })
      .map((roleId) => ({
        roleId,
        roleName: roleId === "__unranked__" ? "Unranked" : (roleNameMap.get(roleId) || roleId),
        entries: grouped[roleId],
      }));
  }, [dashboard?.events, dashboard?.members, dashboard?.guild.roles, memberMap]);

  // Members sorted by rank (highest role position first), then alphabetically
  const sortedMembers = useMemo(() => {
    const roles = dashboard?.guild.roles ?? [];
    const rolePositionMap = new Map(roles.map((r) => [r.discordRoleId, r.position]));
    return [...(dashboard?.members ?? [])].sort((a, b) => {
      const posA = a.rankedRoleId ? (rolePositionMap.get(a.rankedRoleId) ?? -1) : -1;
      const posB = b.rankedRoleId ? (rolePositionMap.get(b.rankedRoleId) ?? -1) : -1;
      if (posA !== posB) return posB - posA;
      return (a.username || a.discordId).localeCompare(b.username || b.discordId);
    });
  }, [dashboard?.members, dashboard?.guild.roles]);

  // Rank tabs derived from members — ordered highest role position first
  const rankTabs = useMemo(() => {
    const roles = dashboard?.guild.roles ?? [];
    const rolePositionMap = new Map(roles.map((r) => [r.discordRoleId, r.position]));
    const roleNameMap = new Map(roles.map((r) => [r.discordRoleId, r.name]));
    const mems = dashboard?.members ?? [];
    const countByRole: Record<string, number> = {};
    let unrankedCount = 0;
    mems.forEach((m) => {
      if (m.rankedRoleId) countByRole[m.rankedRoleId] = (countByRole[m.rankedRoleId] ?? 0) + 1;
      else unrankedCount++;
    });
    const tabs = Object.keys(countByRole)
      .sort((a, b) => (rolePositionMap.get(b) ?? 0) - (rolePositionMap.get(a) ?? 0))
      .map((id) => ({ id, name: roleNameMap.get(id) || id, count: countByRole[id] }));
    if (unrankedCount > 0) tabs.push({ id: "__unranked__", name: "Unranked", count: unrankedCount });
    return tabs;
  }, [dashboard?.members, dashboard?.guild.roles]);

  useEffect(() => {
    if (!guildId) return;

    setLoading(true);
    setError(null);

    Promise.all([
      guildsApi.getDashboard(guildId),
      guildsApi.getMemberSyncStatus(guildId).catch(() => null),
    ])
      .then(([dashRes, syncRes]) => {
        setDashboard(dashRes.data.dashboard);
        if (syncRes) {
          setSyncStatus({
            memberCount: syncRes.data.memberCount,
            synced: syncRes.data.synced,
          });
        }
      })
      .catch(() => setError("Failed to load guild dashboard."))
      .finally(() => setLoading(false));
  }, [guildId]);

  const handleSync = async () => {
    if (!guildId) return;
    setSyncing(true);
    setSyncError(null);
    try {
      await guildsApi.syncMembers(guildId);
      const [dashRes, syncRes] = await Promise.all([
        guildsApi.getDashboard(guildId),
        guildsApi.getMemberSyncStatus(guildId),
      ]);
      setDashboard(dashRes.data.dashboard);
      setSyncStatus({ memberCount: syncRes.data.memberCount, synced: syncRes.data.synced });
      const cooldownUntil = Date.now() + 60_000;
      localStorage.setItem(`syncCooldownUntil_${guildId}`, String(cooldownUntil));
      setSyncCooldown(60);
    } catch (err) {
      const axiosErr = err as AxiosError<{ error?: string }>;
      setSyncError(axiosErr.response?.data?.error ?? "Sync failed. Ensure the bot is installed and has permission to read members.");
    } finally {
      setSyncing(false);
    }
  };

  const handleOpenConfig = () => setConfigOpen(true);

  const handleDisconnect = async () => {
    if (!guildId) return;
    setDisconnecting(true);
    try {
      await guildsApi.deleteGuild(guildId);
      navigate("/app/guilds");
    } catch {
      setDisconnectOpen(false);
    } finally {
      setDisconnecting(false);
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", mt: 8 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !dashboard) {
    return (
      <Box sx={{ maxWidth: 600, mx: "auto", mt: 4 }}>
        <Alert severity="error">{error ?? "Guild not found."}</Alert>
        <Button component={RouterLink} to="/app/guilds" sx={{ mt: 2 }}>
          Back to Guilds
        </Button>
      </Box>
    );
  }

  const { guild, stats, members, events } = dashboard;
  const selectableRoles = (guild.roles ?? []).filter((r) => !r.managed && !r.isDefault).sort((a, b) => b.position - a.position);
  const isOwner = !!user && user.discordId === guild.ownerDiscordId;

  return (
    <Box sx={{ maxWidth: 1200, mx: "auto" }}>
      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 3 }}>
        <Button component={RouterLink} to="/app/guilds" size="small">
          ← Guilds
        </Button>
        <Typography variant="h5">{guild.name}</Typography>
        <Chip
          label={guild.botInstalled ? "Bot Installed" : "Bot Not Installed"}
          color={guild.botInstalled ? "success" : "warning"}
          size="small"
        />
        <Box sx={{ flexGrow: 1 }} />
        {isOwner && (
          <>
            <Button
              variant="outlined"
              size="small"
              color="error"
              onClick={() => { setDisconnectConfirm(""); setDisconnectOpen(true); }}
              sx={{ mr: 1 }}
            >
              Disconnect
            </Button>
            <Button
              variant="outlined"
              size="small"
              startIcon={<SettingsIcon />}
              onClick={handleOpenConfig}
              disabled={!guild.botInstalled}
            >
              Configure
            </Button>
          </>
        )}
      </Stack>

      {/* Stats Cards */}
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2} sx={{ mb: 3 }} flexWrap="wrap">
        <Card variant="outlined" sx={{ flex: 1, minWidth: 150 }}>
          <CardContent>
            <Typography color="text.secondary" variant="body2">
              Total Members
            </Typography>
            <Typography variant="h6">{stats.totalMembers}</Typography>
          </CardContent>
        </Card>
        <Card variant="outlined" sx={{ flex: 1, minWidth: 150 }}>
          <CardContent>
            <Typography color="text.secondary" variant="body2">
              Active Members
            </Typography>
            <Typography variant="h6">{stats.activeMembers}</Typography>
          </CardContent>
        </Card>
        <Card variant="outlined" sx={{ flex: 1, minWidth: 150 }}>
          <CardContent>
            <Typography color="text.secondary" variant="body2">
              Total Events
            </Typography>
            <Typography variant="h6">{stats.totalEvents}</Typography>
          </CardContent>
        </Card>
        <Card variant="outlined" sx={{ flex: 1, minWidth: 150 }}>
          <CardContent>
            <Typography color="text.secondary" variant="body2">
              Active Events
            </Typography>
            <Typography variant="h6">{stats.liveEvents}</Typography>
          </CardContent>
        </Card>
        <Card
          variant="outlined"
          sx={{ flex: 1, minWidth: 150, cursor: "pointer" }}
          component={RouterLink}
          to={`/app/guilds/${guildId}/events`}
        >
          <CardActionArea sx={{ height: "100%" }}>
            <CardContent>
              <Typography color="text.secondary" variant="body2">
                Event Logs
              </Typography>
              <Typography variant="h6">{stats.closedEvents}</Typography>
              <Typography variant="caption" color="primary">View →</Typography>
            </CardContent>
          </CardActionArea>
        </Card>
        <Card variant="outlined" sx={{ flex: 1, minWidth: 150 }}>
          <CardContent>
            <Typography color="text.secondary" variant="body2">
              Participation Rate
            </Typography>
            <Typography variant="h6">{stats.participationRate.toFixed(1)}%</Typography>
          </CardContent>
        </Card>
      </Stack>

      {/* Analytics Section */}
      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Analytics</Typography>
          <Divider sx={{ mb: 2 }} />
          <Stack direction={{ xs: "column", md: "row" }} spacing={3} alignItems="flex-start">

            {/* Member Status Donut */}
            <Box sx={{ flex: 1, minWidth: 0 }}>
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>Member Status</Typography>
              {memberStatusData.length > 0 ? (
                <Box sx={{ position: "relative" }}>
                  <ResponsiveContainer width="100%" height={220}>
                    <PieChart>
                      <Pie
                        data={memberStatusData}
                        cx="50%" cy="50%"
                        innerRadius={65} outerRadius={95}
                        startAngle={90} endAngle={-270}
                        dataKey="value"
                        paddingAngle={memberStatusData.length > 1 ? 3 : 0}
                      >
                        <Cell fill={theme.palette.success.main} />
                        <Cell fill={theme.palette.warning.main} />
                      </Pie>
                      <RechartsTooltip formatter={(v, n) => [`${v} members`, String(n)]} />
                      <Legend
                        formatter={(value) => (
                          <span style={{ fontSize: 12, color: theme.palette.text.secondary }}>{value}</span>
                        )}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                  <Box sx={{
                    position: "absolute", top: "45%", left: "50%",
                    transform: "translate(-50%, -50%)", textAlign: "center", pointerEvents: "none",
                  }}>
                    <Typography variant="h5" fontWeight="bold">{dashboard?.stats.totalMembers}</Typography>
                    <Typography variant="caption" color="text.secondary">Total</Typography>
                  </Box>
                </Box>
              ) : (
                <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>No member data.</Typography>
              )}
            </Box>

            {/* Events per Month */}
            <Box sx={{ flex: 2, minWidth: 0 }}>
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>Event Activity (Recent)</Typography>
              {eventsByMonth.length > 0 ? (
                <ResponsiveContainer width="100%" height={220}>
                  <BarChart data={eventsByMonth} barCategoryGap="30%">
                    <XAxis dataKey="month" tick={{ fontSize: 11 }} />
                    <YAxis allowDecimals={false} tick={{ fontSize: 11 }} width={28} />
                    <RechartsTooltip
                      contentStyle={{ fontSize: 12, borderRadius: 8, borderColor: theme.palette.divider }}
                    />
                    <Legend
                      formatter={(value) => (
                        <span style={{ fontSize: 12, color: theme.palette.text.secondary }}>{value}</span>
                      )}
                    />
                    <Bar dataKey="Events" fill={theme.palette.primary.main} radius={[3, 3, 0, 0]} />
                    <Bar dataKey="Avg Attendance" fill={theme.palette.secondary.main} radius={[3, 3, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>No event data to chart.</Typography>
              )}
            </Box>

          </Stack>
        </CardContent>
      </Card>

      {/* Leaderboard Section — two cards side by side */}
      <Stack direction={{ xs: "column", lg: "row" }} spacing={2} sx={{ mb: 3 }} alignItems="flex-start">

        {/* All-Time Top Contributors */}
        <Card variant="outlined" sx={{ flex: 1, minWidth: 0, width: "100%" }}>
          <CardContent>
            <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: 0.5 }}>
              <Typography variant="h6">All-Time Top Contributors</Typography>
              <Stack direction="row" spacing={1.5} alignItems="center">
                <Tooltip title="Points per hosted event">
                  <Typography variant="caption" color="primary" sx={{ fontWeight: 600, cursor: "default" }}>H&nbsp;=&nbsp;2pts</Typography>
                </Tooltip>
                <Tooltip title="Points per attended event">
                  <Typography variant="caption" color="success.main" sx={{ fontWeight: 600, cursor: "default" }}>A&nbsp;=&nbsp;1pt</Typography>
                </Tooltip>
              </Stack>
            </Stack>
            <Divider sx={{ mb: 1.5 }} />
            {(dashboard?.leaderboard ?? []).length > 0 ? (
              <Stack spacing={0}>
                <Stack direction="row" sx={{ px: 1, pb: 0.5 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ width: 28 }}>#</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ flex: 1 }}>Member</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ width: 52, textAlign: "right" }}>Hosted</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ width: 60, textAlign: "right" }}>Attended</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ width: 44, textAlign: "right" }}>Score</Typography>
                </Stack>
                <Divider />
                {(dashboard?.leaderboard ?? []).slice(0, 10).map((entry) => {
                  const mbr = memberMap.get(entry.discordId);
                  return (
                    <Stack
                      key={entry.discordId}
                      direction="row"
                      alignItems="center"
                      sx={{ px: 1, py: 0.75, borderBottom: "1px solid", borderColor: "divider" }}
                    >
                      <Typography variant="body2" color="text.secondary" sx={{ width: 28, fontWeight: entry.rank <= 3 ? 700 : 400 }}>
                        {entry.rank === 1 ? "🥇" : entry.rank === 2 ? "🥈" : entry.rank === 3 ? "🥉" : `#${entry.rank}`}
                      </Typography>
                      <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 1, minWidth: 0 }}>
                        <Avatar
                          src={mbr?.avatarHash ? `https://cdn.discordapp.com/avatars/${entry.discordId}/${mbr.avatarHash}.png?size=32` : undefined}
                          sx={{ width: 22, height: 22, fontSize: 9 }}
                        >
                          {mbr?.username?.[0]?.toUpperCase() ?? "?"}
                        </Avatar>
                        <Typography variant="body2" noWrap>
                          {mbr?.username ?? entry.discordId.slice(0, 10) + "…"}
                        </Typography>
                      </Stack>
                      <Typography variant="body2" sx={{ width: 52, textAlign: "right" }}>{entry.eventsHosted}</Typography>
                      <Typography variant="body2" sx={{ width: 60, textAlign: "right" }}>{entry.eventsAttended}</Typography>
                      <Typography variant="body2" fontWeight={600} color="primary" sx={{ width: 44, textAlign: "right" }}>{entry.score}</Typography>
                    </Stack>
                  );
                })}
              </Stack>
            ) : (
              <Typography color="text.secondary">No contributor data available.</Typography>
            )}
          </CardContent>
        </Card>

        {/* This Month by Rank */}
        <Card variant="outlined" sx={{ flex: 1, minWidth: 0, width: "100%" }}>
          <CardContent>
            <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: 0, pb: 0 }}>
              <Typography variant="h6">Top Contributors This Month</Typography>
              <Typography variant="caption" color="text.secondary">
                {new Date().toLocaleDateString("en-US", { month: "long", year: "numeric" })}
              </Typography>
            </Stack>
            <Divider sx={{ mt: 1, mb: 1.5 }} />
            {monthlyTopByRank.length === 0 ? (
              <Typography color="text.secondary">No events recorded this month.</Typography>
            ) : (
              <Stack spacing={2}>
                {monthlyTopByRank.map((group) => (
                  <Box key={group.roleId}>
                    <Chip label={group.roleName} size="small" variant="outlined" sx={{ mb: 1, fontWeight: 600 }} />
                    <Stack spacing={0}>
                      {group.entries.slice(0, 5).map((entry, idx) => {
                        const mbr = memberMap.get(entry.discordId);
                        return (
                          <Stack
                            key={entry.discordId}
                            direction="row"
                            alignItems="center"
                            sx={{ px: 1, py: 0.5, borderBottom: "1px solid", borderColor: "divider" }}
                          >
                            <Typography variant="caption" color="text.secondary" sx={{ width: 20 }}>
                              {idx === 0 ? "🥇" : `${idx + 1}.`}
                            </Typography>
                            <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 1, minWidth: 0 }}>
                              <Avatar
                                src={mbr?.avatarHash ? `https://cdn.discordapp.com/avatars/${entry.discordId}/${mbr.avatarHash}.png?size=32` : undefined}
                                sx={{ width: 20, height: 20, fontSize: 8 }}
                              >
                                {mbr?.username?.[0]?.toUpperCase() ?? "?"}
                              </Avatar>
                              <Typography variant="body2" noWrap>
                                {mbr?.username ?? entry.discordId.slice(0, 10) + "…"}
                              </Typography>
                            </Stack>
                            <Tooltip title="Hosted">
                              <Typography variant="caption" color="primary" sx={{ width: 36, textAlign: "right" }}>
                                {entry.hosted > 0 ? `${entry.hosted}H` : ""}
                              </Typography>
                            </Tooltip>
                            <Tooltip title="Attended">
                              <Typography variant="caption" color="success.main" sx={{ width: 36, textAlign: "right" }}>
                                {entry.attended > 0 ? `${entry.attended}A` : ""}
                              </Typography>
                            </Tooltip>
                          </Stack>
                        );
                      })}
                    </Stack>
                  </Box>
                ))}
              </Stack>
            )}
          </CardContent>
        </Card>

      </Stack>

      {/* Lowest Contributors Section — two cards side by side */}
      <Stack direction={{ xs: "column", lg: "row" }} spacing={2} sx={{ mb: 3 }} alignItems="flex-start">

        {/* All-Time Lowest Contributors */}
        <Card variant="outlined" sx={{ flex: 1, minWidth: 0, width: "100%" }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>All-Time Lowest Contributors</Typography>
            <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1 }}>
              Active members with the least recorded activity.
            </Typography>
            <Divider sx={{ mb: 1.5 }} />
            {allTimeLow.length === 0 ? (
              <Typography color="text.secondary">No active members found.</Typography>
            ) : (
              <Stack spacing={0}>
                <Stack direction="row" sx={{ px: 1, pb: 0.5 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ width: 28 }}>#</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ flex: 1 }}>Member</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ width: 52, textAlign: "right" }}>Hosted</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ width: 60, textAlign: "right" }}>Attended</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ width: 72, textAlign: "right" }}>Last Active</Typography>
                </Stack>
                <Divider />
                {allTimeLow.map((m, idx) => {
                  const lastActive = [m.lastHostedDate, m.lastAttendedDate]
                    .filter(Boolean)
                    .map((d) => new Date(d!))
                    .sort((a, b) => b.getTime() - a.getTime())[0];
                  return (
                    <Stack
                      key={m.discordId}
                      direction="row"
                      alignItems="center"
                      sx={{ px: 1, py: 0.75, borderBottom: "1px solid", borderColor: "divider" }}
                    >
                      <Typography variant="body2" color="text.secondary" sx={{ width: 28 }}>
                        {idx + 1}.
                      </Typography>
                      <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 1, minWidth: 0 }}>
                        <Avatar
                          src={m.avatarHash ? `https://cdn.discordapp.com/avatars/${m.discordId}/${m.avatarHash}.png?size=32` : undefined}
                          sx={{ width: 22, height: 22, fontSize: 9 }}
                        >
                          {!m.avatarHash && (m.username?.[0]?.toUpperCase() ?? "?")}
                        </Avatar>
                        <Typography variant="body2" noWrap>{m.username || m.discordId}</Typography>
                      </Stack>
                      <Typography variant="body2" sx={{ width: 52, textAlign: "right" }}>{m.eventsHosted}</Typography>
                      <Typography variant="body2" sx={{ width: 60, textAlign: "right" }}>{m.eventsAttended}</Typography>
                      <Typography variant="body2" sx={{ width: 72, textAlign: "right" }} color={lastActive ? "text.primary" : "text.disabled"}>
                        {lastActive ? lastActive.toLocaleDateString(undefined, { timeZone: "UTC" }) : "Never"}
                      </Typography>
                    </Stack>
                  );
                })}
              </Stack>
            )}
          </CardContent>
        </Card>

        {/* Lowest Contributors This Month by Rank */}
        <Card variant="outlined" sx={{ flex: 1, minWidth: 0, width: "100%" }}>
          <CardContent>
            <Stack direction="row" justifyContent="space-between" alignItems="baseline" sx={{ mb: 0, pb: 0 }}>
              <Typography variant="h6">Lowest Contributors This Month</Typography>
              <Typography variant="caption" color="text.secondary">
                {new Date().toLocaleDateString("en-US", { month: "long", year: "numeric" })}
              </Typography>
            </Stack>
            <Divider sx={{ mt: 1, mb: 1.5 }} />
            {monthlyLowByRank.length === 0 ? (
              <Typography color="text.secondary">No active members found.</Typography>
            ) : (
              <Stack spacing={2}>
                {monthlyLowByRank.map((group) => (
                  <Box key={group.roleId}>
                    <Chip label={group.roleName} size="small" variant="outlined" sx={{ mb: 1, fontWeight: 600 }} />
                    <Stack spacing={0}>
                      {group.entries.slice(0, 5).map((entry, idx) => {
                        const mbr = memberMap.get(entry.discordId);
                        return (
                          <Stack
                            key={entry.discordId}
                            direction="row"
                            alignItems="center"
                            sx={{ px: 1, py: 0.5, borderBottom: "1px solid", borderColor: "divider" }}
                          >
                            <Typography variant="caption" color="text.secondary" sx={{ width: 20 }}>
                              {idx + 1}.
                            </Typography>
                            <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 1, minWidth: 0 }}>
                              <Avatar
                                src={mbr?.avatarHash ? `https://cdn.discordapp.com/avatars/${entry.discordId}/${mbr.avatarHash}.png?size=32` : undefined}
                                sx={{ width: 20, height: 20, fontSize: 8 }}
                              >
                                {mbr?.username?.[0]?.toUpperCase() ?? "?"}
                              </Avatar>
                              <Typography variant="body2" noWrap>
                                {mbr?.username ?? entry.discordId.slice(0, 10) + "…"}
                              </Typography>
                            </Stack>
                            <Tooltip title="Hosted">
                              <Typography variant="caption" color={entry.hosted > 0 ? "primary" : "text.disabled"} sx={{ width: 36, textAlign: "right" }}>
                                {entry.hosted}H
                              </Typography>
                            </Tooltip>
                            <Tooltip title="Attended">
                              <Typography variant="caption" color={entry.attended > 0 ? "success.main" : "text.disabled"} sx={{ width: 36, textAlign: "right" }}>
                                {entry.attended}A
                              </Typography>
                            </Tooltip>
                          </Stack>
                        );
                      })}
                    </Stack>
                  </Box>
                ))}
              </Stack>
            )}
          </CardContent>
        </Card>

      </Stack>

      {/* Members Section */}
      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent>
          <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5 }}>
            <Typography variant="h6">Members ({members?.length ?? 0})</Typography>
          </Stack>
          <TextField
            size="small"
            fullWidth
            placeholder="Search by name or ID…"
            value={memberSearch}
            onChange={(e) => { setMemberSearch(e.target.value); setMembersExpanded(false); }}
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon fontSize="small" />
                  </InputAdornment>
                ),
              },
            }}
            sx={{ mb: 1.5 }}
          />
          {rankTabs.length > 0 && (
            <Tabs
              value={rankTab}
              onChange={(_, v) => { setRankTab(v); setMembersExpanded(false); }}
              variant="scrollable"
              scrollButtons="auto"
              sx={{ mb: 1.5, minHeight: 36, "& .MuiTab-root": { minHeight: 36, py: 0.5, fontSize: 12 } }}
            >
              <Tab label={`All (${sortedMembers.length})`} value="" />
              {rankTabs.map((rt) => (
                <Tab key={rt.id} label={`${rt.name} (${rt.count})`} value={rt.id} />
              ))}
            </Tabs>
          )}
          <Divider sx={{ mb: 2 }} />
          {(() => {
            const MEMBERS_PAGE = 10;
            const q = memberSearch.toLowerCase();
            const filtered = sortedMembers
              .filter((m) =>
                rankTab === "__unranked__" ? !m.rankedRoleId
                : rankTab ? m.rankedRoleId === rankTab
                : true
              )
              .filter((m) =>
                q ? (m.username?.toLowerCase().includes(q) || m.discordId.includes(memberSearch)) : true
              );
            // Truncate only on the All tab with no active search
            const showToggle = !memberSearch && !rankTab && filtered.length > MEMBERS_PAGE;
            const visible = showToggle && !membersExpanded ? filtered.slice(0, MEMBERS_PAGE) : filtered;
            if (filtered.length === 0) {
              return (
                <Typography color="text.secondary">
                  {memberSearch || rankTab ? "No members match your filter." : "No members to display."}
                </Typography>
              );
            }
            return (
              <>
                <Stack spacing={1}>
                  <Stack direction="row" justifyContent="space-between" sx={{ px: 1 }}>
                    <Typography variant="caption" color="text.secondary" sx={{ flex: 3 }}>Member</Typography>
                    <Typography variant="caption" color="text.secondary" sx={{ flex: 2, textAlign: "center" }}>Rank</Typography>
                    <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "center" }}>Status</Typography>
                    <Box sx={{ width: 36 }} />
                  </Stack>
                  <Divider />
                  {visible.map((m) => {
                    const rankName = m.rankedRoleId
                      ? (guild.roles?.find((r) => r.discordRoleId === m.rankedRoleId)?.name || m.rankedRoleId)
                      : null;
                    return (
                      <Stack key={m.discordId} direction="row" justifyContent="space-between" alignItems="center" sx={{ px: 1 }}>
                        <Stack direction="row" spacing={1} alignItems="center" sx={{ flex: 3 }}>
                          <Avatar
                            src={m.avatarHash ? `https://cdn.discordapp.com/avatars/${m.discordId}/${m.avatarHash}.png?size=32` : undefined}
                            sx={{ width: 28, height: 28, fontSize: 12 }}
                          >
                            {!m.avatarHash && (m.username?.[0]?.toUpperCase() ?? "?")}
                          </Avatar>
                          <Box>
                            <Typography variant="body2">{m.username || m.discordId}</Typography>
                            {m.username && (
                              <Typography variant="caption" color="text.secondary" sx={{ fontFamily: "monospace" }}>
                                {m.discordId}
                              </Typography>
                            )}
                          </Box>
                        </Stack>
                        <Box sx={{ flex: 2, display: "flex", justifyContent: "center" }}>
                          {rankName ? (
                            <Chip label={rankName} size="small" variant="outlined" />
                          ) : (
                            <Typography variant="body2" color="text.disabled">—</Typography>
                          )}
                        </Box>
                        <Box sx={{ flex: 1, display: "flex", justifyContent: "center" }}>
                          <Chip
                            label={m.status === "active" ? "Active" : "Retired"}
                            color={m.status === "active" ? "success" : "warning"}
                            size="small"
                          />
                        </Box>
                        <Tooltip title="View stats">
                          <IconButton size="small" onClick={() => setStatsModalMember(m)}>
                            <InfoIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </Stack>
                    );
                  })}
                </Stack>
                {showToggle && (
                  <Button
                    size="small"
                    onClick={() => setMembersExpanded((v) => !v)}
                    endIcon={membersExpanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
                    sx={{ mt: 1.5, width: "100%" }}
                  >
                    {membersExpanded ? "Collapse" : `Show all ${filtered.length} members`}
                  </Button>
                )}
              </>
            );
          })()}
        </CardContent>
      </Card>

      {/* Events Section */}
      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Recent Events ({events?.length ?? 0})
          </Typography>
          <Divider sx={{ mb: 2 }} />
          {events && events.length > 0 ? (
            <Stack spacing={1}>
              <Stack direction="row" justifyContent="space-between" sx={{ px: 1 }}>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 2 }}>Event ID</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 2 }}>Host</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "right" }}>Date</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "right" }}>Participants</Typography>
              </Stack>
              <Divider />
              {events.map((e) => (
                <Stack key={e.id || e.eventId} direction="row" justifyContent="space-between" alignItems="center" sx={{ px: 1 }}>
                  <Typography variant="body2" sx={{ flex: 2, fontFamily: "monospace" }}>
                    {e.id ? e.id.slice(0, 10) + "…" : e.eventId ? e.eventId.slice(0, 10) + "…" : "—"}
                  </Typography>
                  <Typography variant="body2" sx={{ flex: 2 }}>
                    {memberMap.get(e.hostDiscordId)?.username ?? e.hostDiscordId.slice(0, 10) + "…"}
                  </Typography>
                  <Typography variant="body2" sx={{ flex: 1, textAlign: "right" }}>{new Date(e.eventDate).toLocaleDateString(undefined, { timeZone: "UTC" })}</Typography>
                  <Typography variant="body2" sx={{ flex: 1, textAlign: "right" }}>{e.participantIds?.length ?? 0}</Typography>
                </Stack>
              ))}
            </Stack>
          ) : (
            <Typography color="text.secondary">No events to display.</Typography>
          )}
        </CardContent>
      </Card>

      {/* Member Role Configuration — summary card (full config via Configure button) */}
      {isOwner && (
      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent>
          <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1 }}>
            <Typography variant="h6">Configuration</Typography>
            {!guild.botInstalled && (
              <Typography variant="caption" color="text.secondary">Install bot to configure</Typography>
            )}
          </Stack>
          <Divider sx={{ mb: 2 }} />
          <Stack spacing={1}>
            <Stack direction="row" justifyContent="space-between">
              <Typography color="text.secondary" variant="body2">Member Role</Typography>
              <Typography variant="body2">
                {guild.notificationConfig?.statusRoles?.activeRoleId
                  ? (selectableRoles.find((r) => r.discordRoleId === guild.notificationConfig.statusRoles.activeRoleId)?.name || guild.notificationConfig.statusRoles.activeRoleId)
                  : <em style={{ color: "#f57c00" }}>Not set — required for sync</em>}
              </Typography>
            </Stack>
            <Stack direction="row" justifyContent="space-between">
              <Typography color="text.secondary" variant="body2">Inactive Role</Typography>
              <Typography variant="body2">
                {guild.notificationConfig?.statusRoles?.inactiveRoleId
                  ? (selectableRoles.find((r) => r.discordRoleId === guild.notificationConfig.statusRoles.inactiveRoleId)?.name || guild.notificationConfig.statusRoles.inactiveRoleId)
                  : "—"}
              </Typography>
            </Stack>
            <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
              <Typography color="text.secondary" variant="body2">Ranked Roles</Typography>
              <Box sx={{ display: "flex", flexWrap: "wrap", gap: 0.5, justifyContent: "flex-end", maxWidth: "60%" }}>
                {selectableRoles.filter((r) => r.type === "ranked").length > 0
                  ? selectableRoles.filter((r) => r.type === "ranked").map((r) => (
                      <Chip key={r.discordRoleId} label={r.name || r.discordRoleId} size="small" />
                    ))
                  : <Typography variant="body2">—</Typography>}
              </Box>
            </Stack>
            <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
              <Typography color="text.secondary" variant="body2">Moderator Roles</Typography>
              <Box sx={{ display: "flex", flexWrap: "wrap", gap: 0.5, justifyContent: "flex-end", maxWidth: "60%" }}>
                {(guild.notificationConfig?.statusRoles?.moderatorRoleIds ?? []).length > 0
                  ? (guild.notificationConfig.statusRoles.moderatorRoleIds).map((id) => (
                      <Chip
                        key={id}
                        label={selectableRoles.find((r) => r.discordRoleId === id)?.name || id}
                        size="small"
                        color="secondary"
                        variant="outlined"
                      />
                    ))
                  : <Typography variant="body2">—</Typography>}
              </Box>
            </Stack>
          </Stack>
        </CardContent>
      </Card>
      )}

      {isOwner && (
      <Card variant="outlined">
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Member Sync
          </Typography>
          <Divider sx={{ mb: 2 }} />
          <Stack spacing={2}>
            <Stack direction="row" justifyContent="space-between">
              <Typography color="text.secondary">Synced Members</Typography>
              <Typography>{syncStatus?.memberCount ?? 0}</Typography>
            </Stack>
            <Stack direction="row" justifyContent="space-between">
              <Typography color="text.secondary">Sync Status</Typography>
              <Chip
                label={syncStatus?.synced ? "Synced" : "Not Synced"}
                color={syncStatus?.synced ? "success" : "default"}
                size="small"
              />
            </Stack>
            {!dashboard?.guild.notificationConfig?.statusRoles?.activeRoleId && (
              <Alert severity="warning">Use the Configure button to set a member role before syncing.</Alert>
            )}
            {syncError && <Alert severity="error">{syncError}</Alert>}
            <LoadingButton
              variant="contained"
              size="small"
              loading={syncing}
              onClick={handleSync}
              disabled={syncCooldown > 0 || !dashboard?.guild.botInstalled || !dashboard?.guild.notificationConfig?.statusRoles?.activeRoleId}
            >
              {syncCooldown > 0 ? `Sync Members (${syncCooldown}s)` : "Sync Members"}
            </LoadingButton>
          </Stack>
        </CardContent>
      </Card>
      )}

      {/* Guild Configuration Modal */}
      <ConfigDialog
        open={configOpen}
        guild={dashboard?.guild ?? null}
        guildId={guildId ?? ""}
        selectableRoles={selectableRoles}
        onClose={() => setConfigOpen(false)}
        onSaved={(updates) => {
          setDashboard((prev) =>
            prev
              ? {
                  ...prev,
                  guild: {
                    ...prev.guild,
                    notificationConfig: {
                      ...prev.guild.notificationConfig,
                      statusRoles: {
                        activeRoleId: updates.activeRoleId,
                        inactiveRoleId: updates.inactiveRoleId,
                        moderatorRoleIds: updates.moderatorRoleIds,
                      },
                    },
                    roles: prev.guild.roles.map((r) => ({
                      ...r,
                      type: updates.rankedRoleIds.includes(r.discordRoleId) ? "ranked" : "default",
                    })),
                    eventConfig: {
                      eventsChannelId: updates.eventsChannelId,
                      eventTypes: updates.eventTypes,
                    },
                  },
                }
              : null
          );
        }}
      />

      {/* Member Stats Modal */}
      <Dialog open={statsModalMember !== null} onClose={() => setStatsModalMember(null)} maxWidth="xs" fullWidth>
        {statsModalMember && (
          <>
            <DialogTitle>
              <Stack direction="row" spacing={1.5} alignItems="center">
                <Avatar
                  src={statsModalMember.avatarHash ? `https://cdn.discordapp.com/avatars/${statsModalMember.discordId}/${statsModalMember.avatarHash}.png?size=64` : undefined}
                  sx={{ width: 36, height: 36 }}
                >
                  {statsModalMember.username?.[0]?.toUpperCase()}
                </Avatar>
                <Box>
                  <Typography variant="subtitle1" fontWeight="bold">{statsModalMember.username}</Typography>
                  <Typography variant="caption" color="text.secondary" fontFamily="monospace">{statsModalMember.discordId}</Typography>
                </Box>
              </Stack>
            </DialogTitle>
            <DialogContent>
              <Stack spacing={1.5}>
                <Stack direction="row" justifyContent="space-between">
                  <Typography color="text.secondary" variant="body2">Status</Typography>
                  <Chip label={statsModalMember.status === "active" ? "Active" : "Retired"} color={statsModalMember.status === "active" ? "success" : "warning"} size="small" />
                </Stack>
                <Divider />
                <Stack direction="row" justifyContent="space-between">
                  <Typography color="text.secondary" variant="body2">Events Hosted</Typography>
                  <Typography variant="body2">{statsModalMember.eventsHosted}</Typography>
                </Stack>
                <Stack direction="row" justifyContent="space-between">
                  <Typography color="text.secondary" variant="body2">Events Attended</Typography>
                  <Typography variant="body2">{statsModalMember.eventsAttended}</Typography>
                </Stack>
                <Divider />
                <Stack direction="row" justifyContent="space-between">
                  <Typography color="text.secondary" variant="body2">Joined Discord</Typography>
                  <Typography variant="body2">{statsModalMember.discordJoinedAt ? new Date(statsModalMember.discordJoinedAt).toLocaleDateString(undefined, { timeZone: "UTC" }) : "—"}</Typography>
                </Stack>
                <Stack direction="row" justifyContent="space-between">
                  <Typography color="text.secondary" variant="body2">Last Hosted</Typography>
                  <Typography variant="body2">{statsModalMember.lastHostedDate ? new Date(statsModalMember.lastHostedDate).toLocaleDateString(undefined, { timeZone: "UTC" }) : "Never"}</Typography>
                </Stack>
                <Stack direction="row" justifyContent="space-between">
                  <Typography color="text.secondary" variant="body2">Last Attended</Typography>
                  <Typography variant="body2">{statsModalMember.lastAttendedDate ? new Date(statsModalMember.lastAttendedDate).toLocaleDateString(undefined, { timeZone: "UTC" }) : "Never"}</Typography>
                </Stack>
              </Stack>
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setStatsModalMember(null)}>Close</Button>
            </DialogActions>
          </>
        )}
      </Dialog>

      {/* Disconnect Confirmation Dialog */}
      <Dialog open={disconnectOpen} onClose={() => !disconnecting && (setDisconnectOpen(false), setDisconnectConfirm(""))} maxWidth="xs" fullWidth>
        <DialogTitle>Disconnect Guild?</DialogTitle>
        <DialogContent>
          <Typography sx={{ mb: 2 }}>
            This will permanently remove <strong>{guild.name}</strong> and all its synced members from GuildLogger.
            Event history is preserved. This cannot be undone.
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
            Type the guild ID <strong style={{ fontFamily: "monospace" }}>{guild.guildId}</strong> to confirm.
          </Typography>
          <TextField
            size="small"
            fullWidth
            placeholder={guild.guildId}
            value={disconnectConfirm}
            onChange={(e) => setDisconnectConfirm(e.target.value)}
            disabled={disconnecting}
            autoComplete="off"
            slotProps={{ htmlInput: { spellCheck: false } }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => { setDisconnectOpen(false); setDisconnectConfirm(""); }} disabled={disconnecting}>
            Cancel
          </Button>
          <LoadingButton
            variant="contained"
            color="error"
            loading={disconnecting}
            onClick={handleDisconnect}
            disabled={disconnectConfirm !== guild.guildId}
          >
            Disconnect
          </LoadingButton>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
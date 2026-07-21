import { useEffect, useState } from "react";
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
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import SettingsIcon from "@mui/icons-material/Settings";
import { guildsApi, type GuildDashboardData, type SyncStatus } from "../api/guilds";

export const GuildDashboard = () => {
  const { guildId } = useParams<{ guildId: string }>();

  const [dashboard, setDashboard] = useState<GuildDashboardData | null>(null);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [syncing, setSyncing] = useState(false);
  const [syncError, setSyncError] = useState<string | null>(null);
  const [configOpen, setConfigOpen] = useState(false);
  const [configSaving, setConfigSaving] = useState(false);
  const [configError, setConfigError] = useState<string | null>(null);
  const [cfgActiveRoleId, setCfgActiveRoleId] = useState("");
  const [cfgInactiveRoleId, setCfgInactiveRoleId] = useState("");
  const [cfgRankedRoleIds, setCfgRankedRoleIds] = useState<string[]>([]);
  const [disconnectOpen, setDisconnectOpen] = useState(false);
  const [disconnecting, setDisconnecting] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    if (!guildId) return;

    setLoading(true);
    setError(null);

    Promise.all([
      guildsApi.getDashboard(guildId),
      guildsApi.getMemberSyncStatus(guildId),
    ])
      .then(([dashRes, syncRes]) => {
        setDashboard(dashRes.data.dashboard);
        setSyncStatus({
          memberCount: syncRes.data.memberCount,
          synced: syncRes.data.synced,
        });
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
    } catch (err) {
      const axiosErr = err as AxiosError<{ error?: string }>;
      setSyncError(axiosErr.response?.data?.error ?? "Sync failed. Ensure the bot is installed and has permission to read members.");
    } finally {
      setSyncing(false);
    }
  };

  const handleOpenConfig = () => {
    const g = dashboard?.guild;
    setCfgActiveRoleId(g?.notificationConfig?.statusRoles?.activeRoleId ?? "");
    setCfgInactiveRoleId(g?.notificationConfig?.statusRoles?.inactiveRoleId ?? "");
    setCfgRankedRoleIds(g?.roles?.filter((r) => r.type === "ranked").map((r) => r.discordRoleId) ?? []);
    setConfigError(null);
    setConfigOpen(true);
  };

  const handleSaveConfig = async () => {
    if (!guildId || !cfgActiveRoleId) return;
    setConfigSaving(true);
    setConfigError(null);
    try {
      await guildsApi.updateGuildConfig(guildId, {
        activeRoleId: cfgActiveRoleId,
        inactiveRoleId: cfgInactiveRoleId,
        rankedRoleIds: cfgRankedRoleIds,
      });
      setDashboard((prev) =>
        prev
          ? {
              ...prev,
              guild: {
                ...prev.guild,
                notificationConfig: {
                  ...prev.guild.notificationConfig,
                  statusRoles: {
                    activeRoleId: cfgActiveRoleId,
                    inactiveRoleId: cfgInactiveRoleId,
                  },
                },
                roles: prev.guild.roles.map((r) => ({
                  ...r,
                  type: cfgRankedRoleIds.includes(r.discordRoleId) ? "ranked" : "default",
                })),
              },
            }
          : null
      );
      setConfigOpen(false);
    } catch (err) {
      const axiosErr = err as AxiosError<{ error?: string }>;
      setConfigError(axiosErr.response?.data?.error ?? "Failed to save configuration. Please try again.");
    } finally {
      setConfigSaving(false);
    }
  };

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

  const { guild, stats, leaderboard, members, inactiveMembers, events } = dashboard;
  const selectableRoles = (guild.roles ?? []).filter((r) => !r.managed && !r.isDefault).sort((a, b) => b.position - a.position);

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
        <Button
          variant="outlined"
          size="small"
          color="error"
          onClick={() => setDisconnectOpen(true)}
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

      {/* Leaderboard Section */}
      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Top Contributors
          </Typography>
          <Divider sx={{ mb: 2 }} />
          {leaderboard && leaderboard.length > 0 ? (
            <Stack spacing={1}>
              {leaderboard.slice(0, 10).map((entry, idx) => (
                <Stack key={idx} direction="row" justifyContent="space-between" alignItems="center">
                  <Typography variant="body2">
                    #{entry.rank} - {entry.discordId.slice(0, 8)}...
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {entry.eventsHosted} hosted · {entry.eventsAttended} attended
                  </Typography>
                </Stack>
              ))}
            </Stack>
          ) : (
            <Typography color="text.secondary">No leaderboard data available.</Typography>
          )}
        </CardContent>
      </Card>

      {/* Members Section */}
      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Members ({members?.length ?? 0})
          </Typography>
          <Divider sx={{ mb: 2 }} />
          {members && members.length > 0 ? (
            <Stack spacing={1}>
              <Stack direction="row" justifyContent="space-between" sx={{ px: 1 }}>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 3 }}>Member</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 2, textAlign: "center" }}>Rank</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "center" }}>Status</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "right" }}>Hosted</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "right" }}>Attended</Typography>
              </Stack>
              <Divider />
              {members.map((m) => {
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
                  <Typography variant="body2" sx={{ flex: 1, textAlign: "right" }}>{m.eventsHosted}</Typography>
                  <Typography variant="body2" sx={{ flex: 1, textAlign: "right" }}>{m.eventsAttended}</Typography>
                </Stack>
                );
              })}
            </Stack>
          ) : (
            <Typography color="text.secondary">No members to display.</Typography>
          )}
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
                <Stack key={e.eventId} direction="row" justifyContent="space-between" alignItems="center" sx={{ px: 1 }}>
                  <Typography variant="body2" sx={{ flex: 2, fontFamily: "monospace" }}>{e.eventId.slice(0, 12)}…</Typography>
                  <Typography variant="body2" sx={{ flex: 2, fontFamily: "monospace" }}>{e.hostDiscordId}</Typography>
                  <Typography variant="body2" sx={{ flex: 1, textAlign: "right" }}>{new Date(e.eventDate).toLocaleDateString()}</Typography>
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
          </Stack>
        </CardContent>
      </Card>

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
              disabled={!dashboard?.guild.botInstalled || !dashboard?.guild.notificationConfig?.statusRoles?.activeRoleId}
            >
              Sync Members
            </LoadingButton>
          </Stack>
        </CardContent>
      </Card>

      {/* Guild Configuration Modal */}
      <Dialog open={configOpen} onClose={() => setConfigOpen(false)} maxWidth="sm" fullWidth>
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

            {configError && <Alert severity="error">{configError}</Alert>}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfigOpen(false)}>Cancel</Button>
          <LoadingButton
            variant="contained"
            loading={configSaving}
            onClick={handleSaveConfig}
            disabled={!cfgActiveRoleId}
          >
            Save
          </LoadingButton>
        </DialogActions>
      </Dialog>

      {/* Disconnect Confirmation Dialog */}
      <Dialog open={disconnectOpen} onClose={() => !disconnecting && setDisconnectOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Disconnect Guild?</DialogTitle>
        <DialogContent>
          <Typography>
            This will permanently remove <strong>{guild.name}</strong> and all its synced members from GuildLogger.
            Event history is preserved. This cannot be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDisconnectOpen(false)} disabled={disconnecting}>
            Cancel
          </Button>
          <LoadingButton
            variant="contained"
            color="error"
            loading={disconnecting}
            onClick={handleDisconnect}
          >
            Disconnect
          </LoadingButton>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
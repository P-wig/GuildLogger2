import { useEffect, useState } from "react";
import { useParams, Link as RouterLink } from "react-router";
import Alert from "@mui/material/Alert";
import Avatar from "@mui/material/Avatar";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import LoadingButton from "@mui/lab/LoadingButton";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import CircularProgress from "@mui/material/CircularProgress";
import Chip from "@mui/material/Chip";
import Divider from "@mui/material/Divider";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { guildsApi, type GuildDashboardData, type SyncStatus } from "../api/guilds";

export const GuildDashboard = () => {
  const { guildId } = useParams<{ guildId: string }>();

  const [dashboard, setDashboard] = useState<GuildDashboardData | null>(null);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [syncing, setSyncing] = useState(false);
  const [syncError, setSyncError] = useState<string | null>(null);

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
      const syncRes = await guildsApi.getMemberSyncStatus(guildId);
      setSyncStatus({ memberCount: syncRes.data.memberCount, synced: syncRes.data.synced });
    } catch {
      setSyncError("Sync failed. Ensure the bot is installed and has permission to read members.");
    } finally {
      setSyncing(false);
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
              Closed Events
            </Typography>
            <Typography variant="h6">{stats.closedEvents}</Typography>
          </CardContent>
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
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "center" }}>Status</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "right" }}>Hosted</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "right" }}>Attended</Typography>
              </Stack>
              <Divider />
              {members.map((m) => (
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
                  <Box sx={{ flex: 1, display: "flex", justifyContent: "center" }}>
                    <Chip label={m.status} color={m.status === "active" ? "success" : "default"} size="small" />
                  </Box>
                  <Typography variant="body2" sx={{ flex: 1, textAlign: "right" }}>{m.eventsHosted}</Typography>
                  <Typography variant="body2" sx={{ flex: 1, textAlign: "right" }}>{m.eventsAttended}</Typography>
                </Stack>
              ))}
            </Stack>
          ) : (
            <Typography color="text.secondary">No members to display.</Typography>
          )}
        </CardContent>
      </Card>

      {/* Inactive Members Section */}
      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Inactive Members ({inactiveMembers?.length ?? 0})
          </Typography>
          <Divider sx={{ mb: 2 }} />
          {inactiveMembers && inactiveMembers.length > 0 ? (
            <Stack spacing={1}>
              <Stack direction="row" justifyContent="space-between" sx={{ px: 1 }}>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 2 }}>Discord ID</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 1, textAlign: "right" }}>Days Inactive</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ flex: 2, textAlign: "right" }}>Last Activity</Typography>
              </Stack>
              <Divider />
              {inactiveMembers.map((m) => (
                <Stack key={m.discordId} direction="row" justifyContent="space-between" alignItems="center" sx={{ px: 1 }}>
                  <Typography variant="body2" sx={{ flex: 2, fontFamily: "monospace" }}>{m.discordId}</Typography>
                  <Typography variant="body2" sx={{ flex: 1, textAlign: "right" }}>{m.daysSinceActivity ?? "—"}</Typography>
                  <Typography variant="body2" sx={{ flex: 2, textAlign: "right" }}>
                    {m.lastActivityDate ? new Date(m.lastActivityDate).toLocaleDateString() : "—"}
                  </Typography>
                </Stack>
              ))}
            </Stack>
          ) : (
            <Typography color="text.secondary">No inactive members to display.</Typography>
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
            {syncError && <Alert severity="error">{syncError}</Alert>}
            <LoadingButton
              variant="contained"
              size="small"
              loading={syncing}
              onClick={handleSync}
              disabled={!dashboard?.guild.botInstalled}
            >
              Sync Members
            </LoadingButton>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
};
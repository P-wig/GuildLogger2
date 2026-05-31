import { useEffect, useState } from "react";
import { useParams, Link as RouterLink } from "react-router";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
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

  const { guild, memberCount, eventCount } = dashboard;

  return (
    <Box sx={{ maxWidth: 700, mx: "auto" }}>
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

      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Overview
          </Typography>
          <Divider sx={{ mb: 2 }} />
          <Stack spacing={1}>
            <Stack direction="row" justifyContent="space-between">
              <Typography color="text.secondary">Members</Typography>
              <Typography>{memberCount}</Typography>
            </Stack>
            <Stack direction="row" justifyContent="space-between">
              <Typography color="text.secondary">Events</Typography>
              <Typography>{eventCount}</Typography>
            </Stack>
            <Stack direction="row" justifyContent="space-between">
              <Typography color="text.secondary">Connected</Typography>
              <Typography>{new Date(guild.createdAt).toLocaleDateString()}</Typography>
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
          <Stack spacing={1}>
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
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
};
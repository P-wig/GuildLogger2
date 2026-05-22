import { useEffect, useState, useCallback } from "react";
import type { AxiosError } from "axios";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import LoadingButton from "@mui/lab/LoadingButton";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import CircularProgress from "@mui/material/CircularProgress";
import List from "@mui/material/List";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemText from "@mui/material/ListItemText";
import Typography from "@mui/material/Typography";
import { FormDialog } from "../components";
import { guildsApi, type DiscordGuild, type LinkedGuild } from "../api/guilds";

export const Guilds = () => {
  const [linkedGuilds, setLinkedGuilds] = useState<LinkedGuild[]>([]);
  const [linkedLoading, setLinkedLoading] = useState(true);
  const [linkedError, setLinkedError] = useState<string | null>(null);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [discordGuilds, setDiscordGuilds] = useState<DiscordGuild[]>([]);
  const [discordLoading, setDiscordLoading] = useState(false);
  const [discordError, setDiscordError] = useState<string | null>(null);
  const [selectedGuild, setSelectedGuild] = useState<DiscordGuild | null>(null);

  const [connectLoading, setConnectLoading] = useState(false);
  const [connectError, setConnectError] = useState<string | null>(null);

  const [inviteLoadingId, setInviteLoadingId] = useState<string | null>(null);
  const [pendingConfirmIds, setPendingConfirmIds] = useState<Set<string>>(new Set());
  const [confirmLoadingId, setConfirmLoadingId] = useState<string | null>(null);

  const fetchLinkedGuilds = useCallback(async () => {
    setLinkedLoading(true);
    setLinkedError(null);
    try {
      const res = await guildsApi.getMyGuilds();
      setLinkedGuilds(res.data.guilds ?? []);
    } catch {
      setLinkedError("Failed to load your guilds.");
    } finally {
      setLinkedLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchLinkedGuilds();
  }, [fetchLinkedGuilds]);

  const openDialog = async () => {
    setDialogOpen(true);
    setSelectedGuild(null);
    setConnectError(null);
    setDiscordError(null);
    setDiscordLoading(true);
    try {
      const res = await guildsApi.getDiscordGuilds();
      const alreadyLinked = new Set(linkedGuilds.map((g) => g.guildId));
      setDiscordGuilds(res.data.guilds.filter((g) => !alreadyLinked.has(g.id)));
    } catch {
      setDiscordError("Failed to fetch your Discord servers.");
    } finally {
      setDiscordLoading(false);
    }
  };

  const closeDialog = () => {
    setDialogOpen(false);
    setSelectedGuild(null);
    setConnectError(null);
  };

  const handleInstallBot = async (guildId: string) => {
    setInviteLoadingId(guildId);
    try {
      const res = await guildsApi.getBotInviteUrl(guildId);
      window.open(res.data.url, "_blank", "noopener,noreferrer");
      setPendingConfirmIds((prev) => new Set(prev).add(guildId));
    } catch {
      // invite URL fetch failed — button returns to normal state
    } finally {
      setInviteLoadingId(null);
    }
  };

  const handleConfirmBot = async (guildId: string) => {
    setConfirmLoadingId(guildId);
    try {
      await guildsApi.installBot(guildId);
      setPendingConfirmIds((prev) => {
        const next = new Set(prev);
        next.delete(guildId);
        return next;
      });
      fetchLinkedGuilds();
    } catch {
      // confirm failed — button returns to normal state
    } finally {
      setConfirmLoadingId(null);
    }
  };

  const handleConnect = async () => {
    if (!selectedGuild) return;
    setConnectLoading(true);
    setConnectError(null);
    try {
      await guildsApi.connectGuild({ guildId: selectedGuild.id, name: selectedGuild.name });
      closeDialog();
      fetchLinkedGuilds();
    } catch (err) {
      const axiosErr = err as AxiosError<{ error?: string }>;
      setConnectError(axiosErr.response?.data?.error ?? "Failed to connect guild.");
    } finally {
      setConnectLoading(false);
    }
  };

  return (
    <Box>
      <Box sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", mb: 3 }}>
        <Typography variant="h4">My Guilds</Typography>
        <Button variant="contained" onClick={openDialog} disabled={linkedLoading}>
          Connect a Guild
        </Button>
      </Box>

      {linkedLoading && (
        <Box sx={{ display: "flex", justifyContent: "center", mt: 4 }}>
          <CircularProgress />
        </Box>
      )}

      {linkedError && <Alert severity="error">{linkedError}</Alert>}

      {!linkedLoading && !linkedError && linkedGuilds.length === 0 && (
        <Typography color="text.secondary">
          No guilds connected yet. Click "Connect a Guild" to get started.
        </Typography>
      )}

      {!linkedLoading && linkedGuilds.length > 0 && (
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          {linkedGuilds.map((guild) => (
            <Card key={guild.guildId} variant="outlined">
              <CardContent>
                <Typography variant="h6">{guild.name}</Typography>
                <Typography variant="body2" color="text.secondary">
                  ID: {guild.guildId}
                </Typography>
                <Typography
                  variant="body2"
                  color={guild.botInstalled ? "success.main" : "text.secondary"}
                >
                  {guild.botInstalled ? "Bot installed" : "Bot not installed"}
                </Typography>
                {!guild.botInstalled && (
                  <Box sx={{ mt: 1, display: "flex", gap: 1 }}>
                    <LoadingButton
                      size="small"
                      variant="outlined"
                      loading={inviteLoadingId === guild.guildId}
                      onClick={() => handleInstallBot(guild.guildId)}
                    >
                      Install Bot
                    </LoadingButton>
                    {pendingConfirmIds.has(guild.guildId) && (
                      <LoadingButton
                        size="small"
                        variant="contained"
                        loading={confirmLoadingId === guild.guildId}
                        onClick={() => handleConfirmBot(guild.guildId)}
                      >
                        Confirm Installation
                      </LoadingButton>
                    )}
                  </Box>
                )}
              </CardContent>
            </Card>
          ))}
        </Box>
      )}

      <FormDialog
        open={dialogOpen}
        onClose={closeDialog}
        onSubmit={handleConnect}
        title="Connect a Discord Server"
        submitLabel="Connect"
        loading={connectLoading}
        error={connectError}
      >
        {discordLoading && (
          <Box sx={{ display: "flex", justifyContent: "center", py: 2 }}>
            <CircularProgress />
          </Box>
        )}

        {discordError && <Alert severity="error">{discordError}</Alert>}

        {!discordLoading && !discordError && discordGuilds.length === 0 && (
          <Typography color="text.secondary">
            All your Discord servers are already connected.
          </Typography>
        )}

        {!discordLoading && discordGuilds.length > 0 && (
          <>
            <Typography variant="body2" color="text.secondary">
              Select a server to connect:
            </Typography>
            <List disablePadding>
              {discordGuilds.map((g) => (
                <ListItemButton
                  key={g.id}
                  selected={selectedGuild?.id === g.id}
                  onClick={() => setSelectedGuild(g)}
                  sx={{ borderRadius: 1 }}
                >
                  <ListItemText primary={g.name} secondary={g.id} />
                </ListItemButton>
              ))}
            </List>
          </>
        )}
      </FormDialog>
    </Box>
  );
};
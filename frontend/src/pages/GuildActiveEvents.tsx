import { useEffect, useState } from "react";
import type { AxiosError } from "axios";
import { useParams, Link as RouterLink } from "react-router";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import LoadingButton from "@mui/lab/LoadingButton";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Chip from "@mui/material/Chip";
import CircularProgress from "@mui/material/CircularProgress";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import Divider from "@mui/material/Divider";
import IconButton from "@mui/material/IconButton";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import DeleteIcon from "@mui/icons-material/Delete";
import { guildsApi, type ActiveEvent } from "../api/guilds";

function formatScheduledAt(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

export const GuildActiveEvents = () => {
  const { guildId } = useParams<{ guildId: string }>();

  const [events, setEvents] = useState<ActiveEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [deleteTarget, setDeleteTarget] = useState<ActiveEvent | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  useEffect(() => {
    if (!guildId) return;
    setLoading(true);
    setError(null);
    guildsApi
      .listActiveEvents(guildId)
      .then((res) => setEvents(res.data.events ?? []))
      .catch(() => setError("Failed to load active events. You may not have moderator access."))
      .finally(() => setLoading(false));
  }, [guildId]);

  const handleDelete = async () => {
    if (!guildId || !deleteTarget) return;
    setDeleting(true);
    setDeleteError(null);
    try {
      await guildsApi.deleteActiveEvent(guildId, deleteTarget.id);
      setEvents((prev) => prev.filter((e) => e.id !== deleteTarget.id));
      setDeleteTarget(null);
    } catch (err) {
      const axiosErr = err as AxiosError<{ error?: string }>;
      setDeleteError(axiosErr.response?.data?.error ?? "Failed to delete event.");
    } finally {
      setDeleting(false);
    }
  };

  return (
    <Box sx={{ maxWidth: 900, mx: "auto" }}>
      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 3 }}>
        <Button component={RouterLink} to={`/app/guilds/${guildId}`} size="small">
          ← Dashboard
        </Button>
        <Typography variant="h5">Active Events</Typography>
        <Chip label={`${events.length} event${events.length !== 1 ? "s" : ""}`} size="small" />
      </Stack>

      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        These are events currently open or active in Discord. Moderators can delete stale events that were not
        cleaned up by the host.
      </Typography>

      {loading && (
        <Box sx={{ display: "flex", justifyContent: "center", mt: 6 }}>
          <CircularProgress />
        </Box>
      )}

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      {!loading && !error && events.length === 0 && (
        <Alert severity="info">No active events found.</Alert>
      )}

      {!loading && !error && events.length > 0 && (
        <Stack spacing={2}>
          {events
            .slice()
            .sort((a, b) => new Date(a.scheduledAt).getTime() - new Date(b.scheduledAt).getTime())
            .map((ev) => (
              <Card key={ev.id} variant="outlined">
                <CardContent>
                  <Stack direction="row" alignItems="flex-start" justifyContent="space-between" spacing={1}>
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      {/* Header row */}
                      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }} flexWrap="wrap">
                        <Typography variant="subtitle1" fontWeight="bold">
                          {ev.eventType}
                        </Typography>
                        <Chip
                          label={ev.status}
                          size="small"
                          color={ev.status === "active" ? "success" : "default"}
                          variant="outlined"
                        />
                        {ev.capacity > 0 && (
                          <Chip label={`Cap: ${ev.capacity}`} size="small" variant="outlined" />
                        )}
                      </Stack>

                      {/* Meta row */}
                      <Stack direction={{ xs: "column", sm: "row" }} spacing={2} sx={{ mb: 1 }}>
                        <Box>
                          <Typography variant="caption" color="text.secondary">Host</Typography>
                          <Typography variant="body2" fontFamily="monospace">{ev.hostDiscordId}</Typography>
                        </Box>
                        <Box>
                          <Typography variant="caption" color="text.secondary">Scheduled</Typography>
                          <Typography variant="body2">{formatScheduledAt(ev.scheduledAt)}</Typography>
                        </Box>
                        <Box>
                          <Typography variant="caption" color="text.secondary">Channel</Typography>
                          <Typography variant="body2" fontFamily="monospace">{ev.channelId || "—"}</Typography>
                        </Box>
                        <Box>
                          <Typography variant="caption" color="text.secondary">Event ID</Typography>
                          <Typography variant="body2" fontFamily="monospace" sx={{ fontSize: 11 }}>{ev.id}</Typography>
                        </Box>
                      </Stack>

                      {ev.description && (
                        <Typography variant="body2" color="text.secondary" sx={{ mb: 1, fontStyle: "italic" }}>
                          "{ev.description}"
                        </Typography>
                      )}

                      <Divider sx={{ my: 1 }} />

                      {/* RSVP counts */}
                      <Stack direction="row" spacing={3}>
                        <Box>
                          <Typography variant="caption" color="text.secondary">✅ Attending</Typography>
                          <Typography variant="body2" fontWeight="bold">{ev.attendingIds.length}</Typography>
                        </Box>
                        <Box>
                          <Typography variant="caption" color="text.secondary">❓ Maybe</Typography>
                          <Typography variant="body2" fontWeight="bold">{ev.maybeIds.length}</Typography>
                        </Box>
                        <Box>
                          <Typography variant="caption" color="text.secondary">❌ Not Attending</Typography>
                          <Typography variant="body2" fontWeight="bold">{ev.notAttendingIds.length}</Typography>
                        </Box>
                      </Stack>

                      {ev.attendingIds.length > 0 && (
                        <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, display: "block" }}>
                          {ev.attendingIds.join(", ")}
                        </Typography>
                      )}
                    </Box>

                    {/* Delete button */}
                    <Tooltip title="Delete this event (moderator action)">
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => { setDeleteTarget(ev); setDeleteError(null); }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </Stack>
                </CardContent>
              </Card>
            ))}
        </Stack>
      )}

      {/* Delete confirmation dialog */}
      <Dialog open={deleteTarget !== null} onClose={() => !deleting && setDeleteTarget(null)} maxWidth="xs" fullWidth>
        <DialogTitle>Delete Active Event</DialogTitle>
        <DialogContent>
          <Typography>
            Delete the <strong>{deleteTarget?.eventType}</strong> event scheduled for{" "}
            <strong>{deleteTarget ? formatScheduledAt(deleteTarget.scheduledAt) : ""}</strong>?
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            This permanently removes the event record. The Discord embed will remain but buttons will no longer respond.
          </Typography>
          {deleteError && <Alert severity="error" sx={{ mt: 1 }}>{deleteError}</Alert>}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)} disabled={deleting}>Cancel</Button>
          <LoadingButton variant="contained" color="error" loading={deleting} onClick={handleDelete}>
            Delete
          </LoadingButton>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

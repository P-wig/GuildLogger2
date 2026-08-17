import { useEffect, useState } from "react";
import { useSearchParams } from "react-router";
import Alert from "@mui/material/Alert";
import Autocomplete from "@mui/material/Autocomplete";
import Box from "@mui/material/Box";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Chip from "@mui/material/Chip";
import CircularProgress from "@mui/material/CircularProgress";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import LoadingButton from "@mui/lab/LoadingButton";
import { eventLogApi, type EventLogMember } from "../api/eventLog";
import type { ActiveEvent } from "../api/guilds";

type PageState =
  | { phase: "loading" }
  | { phase: "blocked"; reason: string }
  | {
      phase: "form";
      token: string;
      event: ActiveEvent;
      members: EventLogMember[];
      preSelectedIds: string[];
    }
  | { phase: "success" };

function blockedMessage(reason: string): string {
  switch (reason) {
    case "expired":
      return "This event log link has expired. Please use /event log in Discord to generate a new link.";
    case "already_submitted":
      return "This event has already been logged and is now closed.";
    case "not_found":
      return "The event associated with this link could not be found.";
    default:
      return "This link is invalid or has already been used.";
  }
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
}

export const EventLogSubmit = () => {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") ?? "";

  const [state, setState] = useState<PageState>({ phase: "loading" });
  const [summary, setSummary] = useState("");
  const [selectedMembers, setSelectedMembers] = useState<EventLogMember[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      setState({ phase: "blocked", reason: "missing_token" });
      return;
    }

    let cancelled = false;
    eventLogApi
      .validate(token)
      .then((res) => {
        if (cancelled) return;
        const data = res.data;
        if (!data.ok) {
          setState({ phase: "blocked", reason: data.reason });
          return;
        }
        // Pre-select members whose IDs are in the attending list.
        const preSelected = data.members.filter((m) =>
          data.preSelectedIds.includes(m.discordId)
        );
        setSelectedMembers(preSelected);
        setState({
          phase: "form",
          token,
          event: data.event,
          members: data.members,
          preSelectedIds: data.preSelectedIds,
        });
      })
      .catch(() => {
        if (!cancelled) setState({ phase: "blocked", reason: "expired" });
      });

    return () => {
      cancelled = true;
    };
  }, [token]);

  const handleSubmit = async () => {
    if (state.phase !== "form") return;
    if (!summary.trim()) {
      setSubmitError("Please provide a summary before submitting.");
      return;
    }

    setSubmitting(true);
    setSubmitError(null);

    try {
      const res = await eventLogApi.submit({
        token: state.token,
        summary: summary.trim(),
        participantIds: selectedMembers.map((m) => m.discordId),
      });
      if (res.data.ok) {
        setState({ phase: "success" });
      } else {
        setSubmitError("Submission failed. The link may have expired.");
      }
    } catch {
      setSubmitError("Submission failed. Please try again.");
    } finally {
      setSubmitting(false);
    }
  };

  // ── Loading ────────────────────────────────────────────────────────────────
  if (state.phase === "loading") {
    return (
      <Box display="flex" justifyContent="center" mt={8}>
        <CircularProgress />
      </Box>
    );
  }

  // ── Blocked ────────────────────────────────────────────────────────────────
  if (state.phase === "blocked") {
    return (
      <Box maxWidth={560} mx="auto" mt={8} px={2}>
        <Alert severity="warning" variant="filled">
          {blockedMessage(state.reason)}
        </Alert>
      </Box>
    );
  }

  // ── Success ────────────────────────────────────────────────────────────────
  if (state.phase === "success") {
    return (
      <Box maxWidth={560} mx="auto" mt={8} px={2}>
        <Alert severity="success" variant="filled">
          Event logged successfully! This link is now inactive.
        </Alert>
      </Box>
    );
  }

  // ── Form ───────────────────────────────────────────────────────────────────
  const { event, members } = state;

  return (
    <Box maxWidth={600} mx="auto" mt={6} px={2}>
      <Typography variant="h5" gutterBottom>
        Submit Event Log
      </Typography>

      {/* Event info header */}
      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent>
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
            <Chip label={event.eventType} color="primary" size="small" />
            <Typography variant="body2" color="text.secondary">
              {formatDate(event.scheduledAt)}
            </Typography>
          </Stack>
          {event.description && (
            <Typography variant="body2" mt={1}>
              {event.description}
            </Typography>
          )}
        </CardContent>
      </Card>

      <Stack spacing={3}>
        {/* Summary */}
        <TextField
          label="Event Summary"
          placeholder="What happened at this event?"
          multiline
          minRows={4}
          maxRows={12}
          inputProps={{ maxLength: 2000 }}
          value={summary}
          onChange={(e) => setSummary(e.target.value)}
          required
          fullWidth
        />

        {/* Participant picker */}
        <Autocomplete
          multiple
          options={members}
          value={selectedMembers}
          onChange={(_e, newValue) => setSelectedMembers(newValue)}
          getOptionLabel={(m) => m.username}
          isOptionEqualToValue={(a, b) => a.discordId === b.discordId}
          renderInput={(params) => (
            <TextField
              {...params}
              label="Participants"
              placeholder="Search members…"
            />
          )}
          renderTags={(value, getTagProps) =>
            value.map((m, index) => {
              const { key, ...tagProps } = getTagProps({ index });
              return <Chip key={key} label={m.username} size="small" {...tagProps} />;
            })
          }
          fullWidth
        />

        {submitError && <Alert severity="error">{submitError}</Alert>}

        <LoadingButton
          variant="contained"
          size="large"
          loading={submitting}
          onClick={handleSubmit}
          disabled={!summary.trim()}
        >
          Submit Log
        </LoadingButton>
      </Stack>
    </Box>
  );
};

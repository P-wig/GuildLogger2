import { useEffect, useMemo, useState } from "react";
import type { AxiosError } from "axios";
import { useParams, Link as RouterLink } from "react-router";
import Alert from "@mui/material/Alert";
import Autocomplete from "@mui/material/Autocomplete";
import Avatar from "@mui/material/Avatar";
import Badge from "@mui/material/Badge";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import LoadingButton from "@mui/lab/LoadingButton";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import CircularProgress from "@mui/material/CircularProgress";
import Chip from "@mui/material/Chip";
import Collapse from "@mui/material/Collapse";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import Divider from "@mui/material/Divider";
import IconButton from "@mui/material/IconButton";
import InputAdornment from "@mui/material/InputAdornment";
import ListItem from "@mui/material/ListItem";
import ListItemAvatar from "@mui/material/ListItemAvatar";
import ListItemText from "@mui/material/ListItemText";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import CloseIcon from "@mui/icons-material/Close";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import FilterListIcon from "@mui/icons-material/FilterList";
import { guildsApi, type EventLog, type GuildDashboardMemberRow, type LinkedGuild } from "../api/guilds";
import { useAuth } from "../auth";

const DISCORD_CDN = "https://cdn.discordapp.com";

function memberAvatar(member: GuildDashboardMemberRow): string | undefined {
  if (!member.avatarHash) return undefined;
  return `${DISCORD_CDN}/avatars/${member.discordId}/${member.avatarHash}.webp?size=64`;
}

function initials(username: string): string {
  return username.slice(0, 2).toUpperCase();
}

/** Shared Autocomplete option renderer for member picker dropdowns */
function MemberOption(props: React.HTMLAttributes<HTMLLIElement> & { key?: React.Key }, option: GuildDashboardMemberRow) {
  const { key, ...rest } = props as { key: React.Key } & React.HTMLAttributes<HTMLLIElement>;
  return (
    <ListItem key={key} {...rest} dense>
      <ListItemAvatar sx={{ minWidth: 40 }}>
        <Avatar src={memberAvatar(option)} alt={option.username} sx={{ width: 28, height: 28, fontSize: 11 }}>
          {initials(option.username)}
        </Avatar>
      </ListItemAvatar>
      <ListItemText
        primary={option.username}
        secondary={option.discordId}
        slotProps={{ secondary: { style: { fontSize: 11 } } }}
      />
    </ListItem>
  );
}

type LogDialogMode = "create" | "edit";

export const GuildEvents = () => {
  const { guildId } = useParams<{ guildId: string }>();
  const { user } = useAuth();

  const [logs, setLogs] = useState<EventLog[]>([]);
  const [members, setMembers] = useState<GuildDashboardMemberRow[]>([]);
  const [guild, setGuild] = useState<LinkedGuild | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // member lookup map: discordId → member
  const memberMap = useMemo(() => {
    const m = new Map<string, GuildDashboardMemberRow>();
    members.forEach((mbr) => m.set(mbr.discordId, mbr));
    return m;
  }, [members]);

  // Log dialog state
  const [logMode, setLogMode] = useState<LogDialogMode>("create");
  const [editingLogId, setEditingLogId] = useState<string | null>(null);
  const [logOpen, setLogOpen] = useState(false);
  const [logSaving, setLogSaving] = useState(false);
  const [logError, setLogError] = useState<string | null>(null);
  const [logSummary, setLogSummary] = useState("");
  const [logDate, setLogDate] = useState("");
  const [logHost, setLogHost] = useState<GuildDashboardMemberRow | null>(null);
  const [logSelectedMembers, setLogSelectedMembers] = useState<GuildDashboardMemberRow[]>([]);

  // Delete confirmation dialog state
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);

  // Filter state
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [filterAfter, setFilterAfter] = useState("");
  const [filterBefore, setFilterBefore] = useState("");
  const [filterHost, setFilterHost] = useState<GuildDashboardMemberRow | null>(null);
  const [filterParticipant, setFilterParticipant] = useState<GuildDashboardMemberRow | null>(null);
  const [filterLogId, setFilterLogId] = useState("");

  // Per-row summary expansion
  const [expandedLogIds, setExpandedLogIds] = useState<Set<string>>(new Set());
  const SUMMARY_LIMIT = 140;
  const toggleExpanded = (id: string) =>
    setExpandedLogIds((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });

  useEffect(() => {
    if (!guildId) return;
    setLoading(true);
    Promise.all([
      guildsApi.getEventLogs(guildId),
      guildsApi.getDashboard(guildId),
    ])
      .then(([logsRes, dashRes]) => {
        setLogs(logsRes.data.logs ?? []);
        setMembers(dashRes.data.dashboard.members ?? []);
        setGuild(dashRes.data.dashboard.guild);
      })
      .catch(() => setError("Failed to load event data."))
      .finally(() => setLoading(false));
  }, [guildId]);

  const handleOpenCreate = () => {
    setLogMode("create");
    setEditingLogId(null);
    setLogSummary("");
    setLogDate(new Date().toISOString().slice(0, 10));
    setLogHost(null);
    setLogSelectedMembers([]);
    setLogError(null);
    setLogOpen(true);
  };

  const handleOpenEdit = (log: EventLog) => {
    setLogMode("edit");
    setEditingLogId(log.id);
    setLogSummary(log.summary);
    setLogDate(log.eventDate ? new Date(log.eventDate).toISOString().slice(0, 10) : "");
    setLogHost(memberMap.get(log.hostDiscordId) ?? null);
    setLogSelectedMembers(
      (log.participantIds ?? []).map((id) => memberMap.get(id)).filter((m): m is GuildDashboardMemberRow => m !== undefined)
    );
    setLogError(null);
    setLogOpen(true);
  };

  const handleSaveLog = async () => {
    if (!guildId || !logSummary.trim() || !logDate || !logHost || logSelectedMembers.length === 0) return;
    setLogSaving(true);
    setLogError(null);

    const payload = {
      summary: logSummary.trim(),
      eventDate: new Date(logDate + "T12:00:00").toISOString(),
      hostDiscordId: logHost.discordId,
      participantIds: logSelectedMembers.map((m) => m.discordId),
    };

    try {
      if (logMode === "create") {
        const res = await guildsApi.createEventLog(guildId, payload);
        setLogs((prev) => [res.data.log, ...prev]);
      } else if (editingLogId) {
        await guildsApi.updateEventLog(guildId, editingLogId, payload);
        setLogs((prev) =>
          prev.map((l) =>
            l.id === editingLogId ? { ...l, ...payload, eventDate: payload.eventDate } : l
          )
        );
      }
      setLogOpen(false);
    } catch (err) {
      const axiosErr = err as AxiosError<{ error?: string }>;
      setLogError(axiosErr.response?.data?.error ?? "Failed to save event log.");
    } finally {
      setLogSaving(false);
    }
  };

  const handleConfirmDelete = async () => {
    if (!guildId || !deleteTargetId) return;
    setDeleting(true);
    try {
      await guildsApi.deleteEventLog(guildId, deleteTargetId);
      setLogs((prev) => prev.filter((l) => l.id !== deleteTargetId));
      setDeleteTargetId(null);
    } catch {
      // keep dialog open on error — user can try again
    } finally {
      setDeleting(false);
    }
  };

  const canSave = logSummary.trim().length > 0 && logDate.length > 0 && logHost !== null && logSelectedMembers.length > 0;

  // Active filter count
  const activeFilterCount = [filterAfter, filterBefore, filterHost, filterParticipant, filterLogId].filter(Boolean).length;

  const clearFilters = () => {
    setFilterAfter("");
    setFilterBefore("");
    setFilterHost(null);
    setFilterParticipant(null);
    setFilterLogId("");
  };

  // Filtered event logs — all filters are AND-combined
  const filteredLogs = useMemo(() => {
    return logs.filter((log) => {
      if (filterAfter) {
        if (new Date(log.eventDate) < new Date(filterAfter + "T00:00:00Z")) return false;
      }
      if (filterBefore) {
        if (new Date(log.eventDate) > new Date(filterBefore + "T23:59:59Z")) return false;
      }
      if (filterHost && log.hostDiscordId !== filterHost.discordId) return false;
      if (filterParticipant && !(log.participantIds ?? []).includes(filterParticipant.discordId)) return false;
      if (filterLogId) {
        const q = filterLogId.toLowerCase();
        if (!log.id.toLowerCase().includes(q) && !(log.eventId ?? "").toLowerCase().includes(q)) return false;
      }
      return true;
    });
  }, [logs, filterAfter, filterBefore, filterHost, filterParticipant, filterLogId]);

  // canWrite: true for guild owner or members holding a configured moderator role
  const canWrite = useMemo(() => {
    if (!guild || !user) return false;
    if (guild.ownerDiscordId === user.discordId) return true;
    const modRoleIds = new Set(guild.notificationConfig?.statusRoles?.moderatorRoleIds ?? []);
    if (modRoleIds.size === 0) return false;
    const me = members.find((m) => m.discordId === user.discordId);
    if (!me) return false;
    return me.roleIds.some((id) => modRoleIds.has(id));
  }, [guild, user, members]);

  return (
    <Box sx={{ maxWidth: 900, mx: "auto" }}>
      {/* Header */}
      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
        <Button component={RouterLink} to={`/app/guilds/${guildId}`} size="small">
          ← Dashboard
        </Button>
        <Typography variant="h5">Event Logs</Typography>
        <Box sx={{ flexGrow: 1 }} />
        <Badge badgeContent={activeFilterCount} color="primary" overlap="circular">
          <Button
            size="small"
            variant={filtersOpen ? "contained" : "outlined"}
            startIcon={<FilterListIcon />}
            onClick={() => setFiltersOpen((v) => !v)}
          >
            Filters
          </Button>
        </Badge>
        {canWrite && (
          <Button variant="contained" size="small" onClick={handleOpenCreate}>
            + Log Past Event
          </Button>
        )}
      </Stack>

      {/* Filter Panel */}
      <Collapse in={filtersOpen}>
        <Card variant="outlined" sx={{ mb: 2 }}>
          <CardContent sx={{ pb: "12px !important" }}>
            <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1.5 }}>
              <Typography variant="subtitle2">Filter Event Logs</Typography>
              {activeFilterCount > 0 && (
                <Button size="small" startIcon={<CloseIcon fontSize="small" />} onClick={clearFilters}>
                  Clear all ({activeFilterCount})
                </Button>
              )}
            </Stack>
            <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", sm: "1fr 1fr", md: "1fr 1fr 1fr" }, gap: 1.5 }}>
              <TextField
                label="After date"
                type="date"
                size="small"
                value={filterAfter}
                onChange={(e) => setFilterAfter(e.target.value)}
                slotProps={{
                  inputLabel: { shrink: true },
                  input: filterAfter
                    ? { endAdornment: <InputAdornment position="end"><IconButton size="small" edge="end" onClick={() => setFilterAfter("")}><CloseIcon fontSize="small" /></IconButton></InputAdornment> }
                    : undefined,
                }}
              />
              <TextField
                label="Before date"
                type="date"
                size="small"
                value={filterBefore}
                onChange={(e) => setFilterBefore(e.target.value)}
                slotProps={{
                  inputLabel: { shrink: true },
                  input: filterBefore
                    ? { endAdornment: <InputAdornment position="end"><IconButton size="small" edge="end" onClick={() => setFilterBefore("")}><CloseIcon fontSize="small" /></IconButton></InputAdornment> }
                    : undefined,
                }}
              />
              <TextField
                label="Log ID"
                size="small"
                value={filterLogId}
                onChange={(e) => setFilterLogId(e.target.value)}
                placeholder="Partial or full ID…"
                slotProps={{
                  input: filterLogId
                    ? { endAdornment: <InputAdornment position="end"><IconButton size="small" edge="end" onClick={() => setFilterLogId("")}><CloseIcon fontSize="small" /></IconButton></InputAdornment> }
                    : undefined,
                }}
              />
              <Autocomplete
                options={members}
                value={filterHost}
                onChange={(_, v) => setFilterHost(v)}
                getOptionLabel={(o) => o.username || o.discordId}
                isOptionEqualToValue={(a, b) => a.discordId === b.discordId}
                renderOption={MemberOption}
                renderInput={(params) => (
                  <TextField {...params} label="Host" size="small" placeholder="Any host…" />
                )}
              />
              <Autocomplete
                options={members}
                value={filterParticipant}
                onChange={(_, v) => setFilterParticipant(v)}
                getOptionLabel={(o) => o.username || o.discordId}
                isOptionEqualToValue={(a, b) => a.discordId === b.discordId}
                renderOption={MemberOption}
                renderInput={(params) => (
                  <TextField {...params} label="Participant" size="small" placeholder="Any participant…" />
                )}
              />
            </Box>
          </CardContent>
        </Card>
      </Collapse>

      {/* Content */}
      {loading ? (
        <Box sx={{ display: "flex", justifyContent: "center", mt: 8 }}>
          <CircularProgress />
        </Box>
      ) : error ? (
        <Alert severity="error">{error}</Alert>
      ) : (
        <Card variant="outlined">
          <CardContent>
            <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1 }}>
              <Typography variant="h6">
                Past Events ({activeFilterCount > 0 ? `${filteredLogs.length} of ${logs.length}` : logs.length})
                {activeFilterCount > 0 && (
                  <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}>
                    — {activeFilterCount} filter{activeFilterCount > 1 ? "s" : ""} active
                  </Typography>
                )}
              </Typography>
            </Stack>
            <Divider sx={{ mb: 2 }} />
            {filteredLogs.length === 0 ? (
              <Typography color="text.secondary">
                {logs.length === 0
                  ? "No event logs yet. Use \"Log Past Event\" to record completed events."
                  : "No logs match the active filters."}
              </Typography>
            ) : (
              <Stack spacing={0}>
                <Stack direction="row" sx={{ px: 1, pb: 0.5 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ flex: 2 }}>Date</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ flex: 2, fontFamily: "monospace" }}>Log ID</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ flex: 5 }}>Summary</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ flex: 2 }}>Host</Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ flex: 4 }}>Participants</Typography>
                  <Box sx={{ width: 72 }} />
                </Stack>
                <Divider />
                {filteredLogs.map((log) => {
                  const host = memberMap.get(log.hostDiscordId);
                  const resolvedParticipants = (log.participantIds ?? []).map((id) => ({
                    id,
                    member: memberMap.get(id),
                  }));
                  const isLong = log.summary.length > SUMMARY_LIMIT;
                  const isExpanded = expandedLogIds.has(log.id);
                  const displaySummary = isLong && !isExpanded
                    ? log.summary.slice(0, SUMMARY_LIMIT).trimEnd() + "…"
                    : log.summary;
                  return (
                    <Stack
                      key={log.id}
                      direction="row"
                      alignItems="flex-start"
                      sx={{ px: 1, py: 1, borderBottom: "1px solid", borderColor: "divider" }}
                    >
                      <Typography variant="body2" sx={{ flex: 2, pt: 0.5 }}>
                        {new Date(log.eventDate).toLocaleDateString(undefined, { timeZone: "UTC" })}
                      </Typography>
                      <Tooltip title={log.id} placement="top">
                        <Typography variant="body2" sx={{ flex: 2, pt: 0.5, fontFamily: "monospace", color: "text.secondary", cursor: "default" }}>
                          {log.id ? log.id.slice(0, 8) + "…" : "—"}
                        </Typography>
                      </Tooltip>
                      <Box sx={{ flex: 5 }}>
                        <Typography variant="body2">{displaySummary}</Typography>
                        {isLong && (
                          <Typography
                            variant="caption"
                            color="primary"
                            onClick={() => toggleExpanded(log.id)}
                            sx={{ cursor: "pointer", userSelect: "none", display: "block" }}
                          >
                            {isExpanded ? "Show less ▲" : "Show more ▼"}
                          </Typography>
                        )}
                        <Typography variant="caption" color="text.secondary">
                          Logged {new Date(log.submittedAt).toLocaleDateString()}
                        </Typography>
                      </Box>
                      <Box sx={{ flex: 2, pt: 0.25 }}>
                        {host ? (
                          <Chip
                            avatar={
                              <Avatar src={memberAvatar(host)} alt={host.username} sx={{ width: 18, height: 18, fontSize: 9 }}>
                                {initials(host.username)}
                              </Avatar>
                            }
                            label={host.username}
                            size="small"
                            variant="outlined"
                          />
                        ) : (
                          <Chip label={log.hostDiscordId.slice(-6)} size="small" variant="outlined" title={log.hostDiscordId} />
                        )}
                      </Box>
                      <Box sx={{ flex: 4, display: "flex", flexWrap: "wrap", gap: 0.5, pt: 0.25 }}>
                        {resolvedParticipants.map(({ id, member }) =>
                          member ? (
                            <Chip
                              key={id}
                              avatar={
                                <Avatar src={memberAvatar(member)} alt={member.username} sx={{ width: 18, height: 18, fontSize: 9 }}>
                                  {initials(member.username)}
                                </Avatar>
                              }
                              label={member.username}
                              size="small"
                              variant="outlined"
                            />
                          ) : (
                            <Chip key={id} label={id.slice(-6)} size="small" variant="outlined" title={id} />
                          )
                        )}
                      </Box>
                      <Stack direction="row" sx={{ width: 72, justifyContent: "flex-end" }}>
                        {canWrite && (
                          <>
                            <Tooltip title="Edit">
                              <IconButton size="small" onClick={() => handleOpenEdit(log)}>
                                <EditIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                            <Tooltip title="Delete">
                              <IconButton size="small" color="error" onClick={() => setDeleteTargetId(log.id)}>
                                <DeleteIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                          </>
                        )}
                      </Stack>
                    </Stack>
                  );
                })}
              </Stack>
            )}
          </CardContent>
        </Card>
      )}

      {/* Log / Edit Event Dialog */}
      <Dialog open={logOpen} onClose={() => !logSaving && setLogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{logMode === "edit" ? "Edit Event Log" : "Log Past Event"}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label="Event Date"
              type="date"
              value={logDate}
              onChange={(e) => setLogDate(e.target.value)}
              size="small"
              fullWidth
              required
              slotProps={{ inputLabel: { shrink: true } }}
            />
            <TextField
              label="Summary / Description"
              value={logSummary}
              onChange={(e) => setLogSummary(e.target.value)}
              size="small"
              fullWidth
              required
              multiline
              rows={4}
              placeholder="What happened at this event?"
            />
            <Autocomplete
              options={members}
              value={logHost}
              onChange={(_, val) => setLogHost(val)}
              getOptionLabel={(opt) => opt.username}
              isOptionEqualToValue={(a, b) => a.discordId === b.discordId}
              renderOption={MemberOption}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Host"
                  size="small"
                  required
                  placeholder="Search members by username…"
                  error={logHost === null && logSummary.length > 0}
                  helperText={logHost === null && logSummary.length > 0 ? "Host is required." : undefined}
                />
              )}
            />
            <Autocomplete
              multiple
              options={members}
              value={logSelectedMembers}
              onChange={(_, val) => setLogSelectedMembers(val)}
              getOptionLabel={(opt) => opt.username}
              isOptionEqualToValue={(a, b) => a.discordId === b.discordId}
              filterSelectedOptions
              renderOption={MemberOption}
              renderTags={(selected, getTagProps) =>
                selected.map((option, index) => {
                  const { key, ...tagProps } = getTagProps({ index });
                  return (
                    <Chip
                      key={key}
                      avatar={
                        <Avatar
                          src={memberAvatar(option)}
                          alt={option.username}
                          sx={{ width: 20, height: 20, fontSize: 9 }}
                        >
                          {initials(option.username)}
                        </Avatar>
                      }
                      label={option.username}
                      size="small"
                      {...tagProps}
                    />
                  );
                })
              }
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Participants"
                  size="small"
                  required
                  placeholder={logSelectedMembers.length === 0 ? "Search members by username\u2026" : undefined}
                  helperText={
                    logSelectedMembers.length > 0
                      ? `${logSelectedMembers.length} member${logSelectedMembers.length !== 1 ? "s" : ""} selected`
                      : "At least one participant is required."
                  }
                  error={logSelectedMembers.length === 0 && logSummary.length > 0}
                />
              )}
            />
            {logError && <Alert severity="error">{logError}</Alert>}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setLogOpen(false)} disabled={logSaving}>
            Cancel
          </Button>
          <LoadingButton
            variant="contained"
            loading={logSaving}
            onClick={handleSaveLog}
            disabled={!canSave}
          >
            {logMode === "edit" ? "Save Changes" : "Save Log"}
          </LoadingButton>
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteTargetId !== null} onClose={() => !deleting && setDeleteTargetId(null)} maxWidth="xs" fullWidth>
        <DialogTitle>Delete Event Log</DialogTitle>
        <DialogContent>
          <Typography>Are you sure you want to delete this event log? This cannot be undone.</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTargetId(null)} disabled={deleting}>Cancel</Button>
          <LoadingButton variant="contained" color="error" loading={deleting} onClick={handleConfirmDelete}>
            Delete
          </LoadingButton>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

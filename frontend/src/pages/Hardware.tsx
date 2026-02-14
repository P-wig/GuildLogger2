import { useCallback, useEffect, useState } from "react";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { ProjectButton } from "../components";
import {
  hardwareApi,
  type CreateHardwareRequest,
  type Hardware,
} from "../api/hardware";
import { HardwareList } from "../components";
import styles from "./homePage.module.css";

// ── Create Hardware Dialog ──────────────────────────────────────

type CreateDialogProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (hw: Hardware) => void;
};

const CreateHardwareDialog = ({
  open,
  onClose,
  onCreated,
}: CreateDialogProps) => {
  const [form, setForm] = useState({ hardwareName: "", capacity: "" });
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleChange = (field: string, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    setError(null);
  };

  const handleSubmit = async () => {
    if (!form.hardwareName.trim()) {
      setError("Hardware name is required.");
      return;
    }

    const capacity = parseInt(form.capacity, 10);
    if (isNaN(capacity) || capacity < 1) {
      setError("Capacity must be a positive number.");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const payload: CreateHardwareRequest = {
        hardwareName: form.hardwareName.trim(),
        capacity,
      };
      const { data } = await hardwareApi.create(payload);
      onCreated(data);
      setForm({ hardwareName: "", capacity: "" });
      onClose();
    } catch (err: unknown) {
      if (
        err &&
        typeof err === "object" &&
        "response" in err &&
        (err as { response?: { data?: { error?: string } } }).response?.data
          ?.error
      ) {
        setError(
          (err as { response: { data: { error: string } } }).response.data
            .error,
        );
      } else {
        setError("Failed to create hardware.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>Create New Hardware Set</DialogTitle>
      <DialogContent
        sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 1 }}
      >
        {error && <Alert severity="error">{error}</Alert>}
        <TextField
          label="Hardware Name"
          value={form.hardwareName}
          onChange={(e) => handleChange("hardwareName", e.target.value)}
          required
          helperText="Unique name for this hardware set"
          autoFocus
        />
        <TextField
          label="Capacity"
          type="number"
          value={form.capacity}
          onChange={(e) => handleChange("capacity", e.target.value)}
          required
          helperText="Total number of units in stock"
          slotProps={{ htmlInput: { min: 1 } }}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={submitting}>
          Cancel
        </Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={submitting}
        >
          {submitting ? "Creating…" : "Create"}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

// ── Hardware Card ───────────────────────────────────────────────

type HardwareCardProps = {
  hw: Hardware;
  onDelete: (hw: Hardware) => void;
};

const HardwareCard = ({ hw, onDelete }: HardwareCardProps) => {
  return (
    <Card
      variant="outlined"
      sx={{ width: "min(66vw, 1000px)", maxWidth: "100%", mx: "auto" }}
    >
      <CardContent>
        <Typography variant="h6" sx={{ mb: 1 }}>
          {hw.hardwareName}
        </Typography>
        <HardwareList
          hardware={[hw]}
          displayTitle={false}
          displayName={false}
        />
        {hw.assignedProjects.length > 0 && (
          <Typography
            variant="caption"
            color="text.secondary"
            sx={{ mt: 1, display: "block" }}
          >
            Assigned to {hw.assignedProjects.length} project
            {hw.assignedProjects.length !== 1 ? "s" : ""}
          </Typography>
        )}
      </CardContent>
      <CardActions sx={{ px: 2, pb: 2 }}>
        <ProjectButton
          label="Delete"
          onClick={() => onDelete(hw)}
          variant="outlined"
          color="error"
        />
      </CardActions>
    </Card>
  );
};

// ── Main Hardware Page ──────────────────────────────────────────

export const HardwarePage = () => {
  const [hardwareList, setHardwareList] = useState<Hardware[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);

  const fetchHardware = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const { data } = await hardwareApi.list();
      setHardwareList(data);
    } catch {
      setError("Failed to load hardware.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchHardware();
  }, [fetchHardware]);

  const handleCreated = (hw: Hardware) => {
    setHardwareList((prev) => [hw, ...prev]);
  };

  const handleDelete = async (hw: Hardware) => {
    try {
      await hardwareApi.delete(hw._id);
      setHardwareList((prev) => prev.filter((h) => h._id !== hw._id));
    } catch {
      setError("Failed to delete hardware.");
    }
  };

  return (
    <div className={styles.root}>
      <Stack
        direction="row"
        alignItems="baseline"
        justifyContent="space-between"
        gap={2}
      >
        <Button variant="contained" onClick={() => setCreateOpen(true)}>
          New Hardware Set
        </Button>
      </Stack>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Typography color="text.secondary">Loading hardware…</Typography>
      ) : hardwareList.length === 0 ? (
        <Card variant="outlined">
          <CardContent>
            <Typography color="text.secondary" align="center">
              No hardware sets yet. Create one to get started.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <div className={styles.grid}>
          {hardwareList.map((hw) => (
            <HardwareCard key={hw._id} hw={hw} onDelete={handleDelete} />
          ))}
        </div>
      )}

      <CreateHardwareDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={handleCreated}
      />
    </div>
  );
};

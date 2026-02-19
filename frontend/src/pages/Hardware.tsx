import { useMemo, useState } from "react";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import MenuItem from "@mui/material/MenuItem";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import ToggleButton from "@mui/material/ToggleButton";
import ToggleButtonGroup from "@mui/material/ToggleButtonGroup";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";
import { useAppData } from "../context/AppContext";
import { FormDialog, ProjectButton, HardwareList } from "../components";
import type { Hardware } from "../api/hardware";
import styles from "./homePage.module.css";

type CreateDialogProps = {
  open: boolean;
  onClose: () => void;
};

const CreateHardwareDialog = ({ open, onClose }: CreateDialogProps) => {
  const { createHardware } = useAppData();
  const [form, setForm] = useState({ hardwareName: "", capacity: "" });
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

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

    setLoading(true);
    setError(null);

    try {
      await createHardware({
        hardwareName: form.hardwareName.trim(),
        capacity,
      });
      setForm({ hardwareName: "", capacity: "" });
      onClose();
    } catch (err: unknown) {
      const msg =
        err &&
        typeof err === "object" &&
        "response" in err &&
        (err as { response?: { data?: { error?: string } } }).response?.data
          ?.error;
      setError((msg as string) || "Failed to create hardware.");
    } finally {
      setLoading(false);
    }
  };

  return (


    <FormDialog
      open={open}
      onClose={onClose}
      onSubmit={handleSubmit}
      title="Create New Hardware Set"
      submitLabel="Create"
      loading={loading}
      error={error}
    >

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
    </FormDialog>
  );
};

type ManageDialogProps = {
  open: boolean;
  onClose: () => void;
  hw: Hardware | null;
};

const ManageHardwareDialog = ({ open, onClose, hw }: ManageDialogProps) => {
  const { user } = useAuth();
  const { projects, checkoutHardware, checkinHardware } = useAppData();
  const userId = user?.userId ?? "";

  const [mode, setMode] = useState<"checkout" | "checkin">("checkout");
  const [selectedProjectId, setSelectedProjectId] = useState("");
  const [amount, setAmount] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // Only show projects the user is assigned to
  const userProjects = useMemo(
    () => projects.filter((p) => p.assignedUsers.includes(userId)),
    [projects, userId],
  );

  const handleSubmit = async () => {
    if (!hw) return;
    if (!selectedProjectId) {
      setError("Please select a project.");
      return;
    }
    const qty = parseInt(amount, 10);
    if (isNaN(qty) || qty < 1) {
      setError("Amount must be a positive number.");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      if (mode === "checkout") {
        await checkoutHardware(hw._id, {
          projectId: selectedProjectId,
          amount: qty,
        });
      } else {
        await checkinHardware(hw._id, {
          projectId: selectedProjectId,
          amount: qty,
        });
      }
      setSelectedProjectId("");
      setAmount("");
      onClose();
    } catch (err: unknown) {
      const msg =
        err &&
        typeof err === "object" &&
        "response" in err &&
        (err as { response?: { data?: { error?: string } } }).response?.data
          ?.error;
      setError(
        (msg as string | undefined) ||
        `Failed to ${mode === "checkout" ? "check out" : "check in"} hardware.`,
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <FormDialog
      open={open}
      onClose={onClose}
      onSubmit={handleSubmit}
      title={`Manage: ${hw?.hardwareName ?? ""}`}
      submitLabel={mode === "checkout" ? "Check Out" : "Check In"}
      loading={loading}
      error={error}
    >
      <ToggleButtonGroup
        value={mode}
        exclusive
        onChange={(_e, val) => {
          if (val) {
            setMode(val as "checkout" | "checkin");
            setError(null);
          }
        }}
        fullWidth
        size="small"
      >
        <ToggleButton value="checkout">Check Out</ToggleButton>
        <ToggleButton value="checkin">Check In</ToggleButton>
      </ToggleButtonGroup>

      <TextField
        select
        label="Project"
        value={selectedProjectId}
        onChange={(e) => {
          setSelectedProjectId(e.target.value);
          setError(null);
        }}
        required
        helperText="Select one of your assigned projects"
      >
        {userProjects.map((p) => (
          <MenuItem key={p.projectId} value={p.projectId}>
            {p.projectName} ({p.projectId})
          </MenuItem>
        ))}
      </TextField>

      <TextField
        label="Amount"
        type="number"
        value={amount}
        onChange={(e) => {
          setAmount(e.target.value);
          setError(null);
        }}
        required
        helperText={
          mode === "checkout"
            ? `Available: ${hw?.available ?? 0}`
            : "Units to return"
        }
        slotProps={{ htmlInput: { min: 1 } }}
      />
    </FormDialog>
  );
};

type HardwareCardProps = {
  hw: Hardware;
  onDelete: (hw: Hardware) => void;
  onManage: (hw: Hardware) => void;
};

const HardwareCard = ({ hw, onDelete, onManage }: HardwareCardProps) => {
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
          label="Manage"
          onClick={() => onManage(hw)}
          variant="contained"
        />
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

export const HardwarePage = () => {
  const { hardware, loadingHardware, deleteHardware } = useAppData();
  const [error, setError] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [manageHw, setManageHw] = useState<Hardware | null>(null);

  const handleDelete = async (hw: Hardware) => {
    try {
      await deleteHardware(hw._id);
    } catch {
      setError("Failed to delete hardware.");
    }
  };

  return (
    <div className={styles.root}>
      <Stack
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        alignItems={{ xs: "stretch", md: "center" }}
        spacing={2}
        sx={{ mb: 3 }}
      >
        <Typography variant="h4" sx={{ fontWeight: 600 }}>
          Hardware
        </Typography>

        <Button
          variant="contained"
          onClick={() => setCreateOpen(true)}
          sx={{ alignSelf: { xs: "stretch", md: "auto" } }}
        >
          New Hardware Set
        </Button>
      </Stack>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {loadingHardware ? (
        <Typography color="text.secondary">Loading hardware…</Typography>
      ) : hardware.length === 0 ? (
        <Card variant="outlined">
          <CardContent>
            <Typography color="text.secondary" align="center">
              No hardware sets yet. Create one to get started.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <div className={styles.grid}>
          {hardware.map((hw) => (
            <HardwareCard
              key={hw._id}
              hw={hw}
              onDelete={handleDelete}
              onManage={setManageHw}
            />
          ))}
        </div>
      )}

      <CreateHardwareDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />

      <ManageHardwareDialog
        open={!!manageHw}
        onClose={() => setManageHw(null)}
        hw={manageHw}
      />
    </div>
  );
};

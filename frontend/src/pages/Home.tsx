import { useMemo, useState } from "react";
import Alert from "@mui/material/Alert";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import MenuItem from "@mui/material/MenuItem";
import TextField from "@mui/material/TextField";
import ToggleButton from "@mui/material/ToggleButton";
import ToggleButtonGroup from "@mui/material/ToggleButtonGroup";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";
import { useAppData } from "../context/AppContext";
import { FormDialog, ProjectCard } from "../components";
import type { CardAction } from "../components/ProjectCard";
import type { Project } from "../api/projects";
import styles from "./homePage.module.css";

type ManageDialogProps = {
  open: boolean;
  onClose: () => void;
  project: Project | null;
};

const ManageProjectDialog = ({ open, onClose, project }: ManageDialogProps) => {
  const { hardware, checkoutHardware, checkinHardware } = useAppData();

  const [mode, setMode] = useState<"checkout" | "checkin">("checkout");
  const [selectedHwId, setSelectedHwId] = useState("");
  const [amount, setAmount] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const selectedHw = useMemo(
    () => hardware.find((h) => h._id === selectedHwId),
    [hardware, selectedHwId],
  );

  const handleSubmit = async () => {
    if (!project) return;
    if (!selectedHwId) {
      setError("Please select a hardware set.");
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
        await checkoutHardware(selectedHwId, {
          projectId: project.projectId,
          amount: qty,
        });
      } else {
        await checkinHardware(selectedHwId, {
          projectId: project.projectId,
          amount: qty,
        });
      }
      setSelectedHwId("");
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
      title={`Manage Hardware: ${project?.projectName ?? ""}`}
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
        label="Hardware"
        value={selectedHwId}
        onChange={(e) => {
          setSelectedHwId(e.target.value);
          setError(null);
        }}
        required
        helperText="Select a hardware set"
      >
        {hardware.map((h) => (
          <MenuItem key={h._id} value={h._id}>
            {h.hardwareName} ({h.available}/{h.capacity} available)
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
            ? `Available: ${selectedHw?.available ?? 0}`
            : "Units to return"
        }
        slotProps={{ htmlInput: { min: 1 } }}
      />
    </FormDialog>
  );
};

export const Home = () => {
  const { user } = useAuth();
  const {
    projects,
    hardware,
    loadingProjects,
    joinProject,
    leaveProject,
    deleteProject,
  } = useAppData();

  const [error, setError] = useState<string | null>(null);
  const [manageProject, setManageProject] = useState<Project | null>(null);
  const userId = user?.userId ?? "";

  const getHardwareForProject = (project: Project) => {
    const hwIds = project.assignedHardware.map((ah) => ah.hardwareId);
    return hardware.filter((h) => hwIds.includes(h._id));
  };

  const buildActions = (project: Project): CardAction[] => {
    const isOwner = project.ownerUserId === userId;
    const isAssigned = project.assignedUsers.includes(userId);
    const actions: CardAction[] = [];

    if (!isAssigned) {
      actions.push({
        label: "Join",
        onClick: async () => {
          try {
            await joinProject(project.projectId);
          } catch {
            setError("Failed to join project.");
          }
        },
        variant: "contained",
      });
    }

    if (isAssigned) {
      actions.push({
        label: "Manage",
        onClick: () => setManageProject(project),
        variant: "outlined",
      });
    }

    if (isAssigned) {
      actions.push({
        label: "Leave",
        onClick: async () => {
          try {
            await leaveProject(project.projectId);
          } catch {
            setError("Failed to leave project.");
          }
        },
        variant: "outlined",
      });
    }

    if (isOwner) {
      actions.push({
        label: "Delete",
        onClick: async () => {
          try {
            await deleteProject(project.projectId);
          } catch {
            setError("Failed to delete project.");
          }
        },
        variant: "outlined",
        color: "error",
      });
    }

    return actions;
  };

  return (
    <div className={styles.root}>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {loadingProjects ? (
        <Typography color="text.secondary">Loading projects…</Typography>
      ) : projects.length === 0 ? (
        <Card variant="outlined">
          <CardContent>
            <Typography color="text.secondary" align="center">
              No projects yet. Head to the Projects page to create or join one.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <div className={styles.grid}>
          {projects.map((project) => (
            <ProjectCard
              key={project._id}
              project={project}
              hardware={getHardwareForProject(project)}
              actions={buildActions(project)}
            />
          ))}
        </div>
      )}

      <ManageProjectDialog
        open={!!manageProject}
        onClose={() => setManageProject(null)}
        project={manageProject}
      />
    </div>
  );
};

import { useMemo, useState } from "react";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import MenuItem from "@mui/material/MenuItem";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import ToggleButton from "@mui/material/ToggleButton";
import ToggleButtonGroup from "@mui/material/ToggleButtonGroup";
import Typography from "@mui/material/Typography";

import { useAuth } from "../auth";
import { useAppData } from "../context/AppContext";
import { FormDialog, ProjectCard } from "../components";
import type { Project } from "../api/projects";
import type { CardAction } from "../components/ProjectCard";

import styles from "./homePage.module.css";

type ProjectFilter = "all" | "owner" | "assigned";

type ManageDialogProps = {
  open: boolean;
  onClose: () => void;
  project: Project | null;
};

type CreateDialogProps = {
  open: boolean;
  onClose: () => void;
};

type JoinDialogProps = {
  open: boolean;
  onClose: () => void;
};

/* ========================= */
/* Manage Hardware Dialog */
/* ========================= */

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
    } catch {
      setError(
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
        onChange={(_, val) => val && setMode(val)}
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
        onChange={(e) => setSelectedHwId(e.target.value)}
        required
      >
        {hardware.map((h) => (
          <MenuItem key={h._id} value={h._id}>
            {h.hardwareName} ({h.available}/{h.capacity})
          </MenuItem>
        ))}
      </TextField>

      <TextField
        label="Amount"
        type="number"
        value={amount}
        onChange={(e) => setAmount(e.target.value)}
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

/* ========================= */
/* Create Project Dialog */
/* ========================= */

const CreateProjectDialog = ({ open, onClose }: CreateDialogProps) => {
  const { createProject } = useAppData();
  const [form, setForm] = useState({
    projectId: "",
    projectName: "",
    description: "",
  });
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    if (!form.projectId.trim() || !form.projectName.trim()) {
      setError("Project ID and Project Name are required.");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await createProject({
        projectId: form.projectId.trim(),
        projectName: form.projectName.trim(),
        description: form.description.trim(),
        ownerUserId: "",
      });
      onClose();
    } catch {
      setError("Failed to create project.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <FormDialog
      open={open}
      onClose={onClose}
      onSubmit={handleSubmit}
      title="Create New Project"
      submitLabel="Create"
      loading={loading}
      error={error}
    >
      <TextField
        label="Project ID"
        value={form.projectId}
        onChange={(e) =>
          setForm({ ...form, projectId: e.target.value })
        }
        required
      />
      <TextField
        label="Project Name"
        value={form.projectName}
        onChange={(e) =>
          setForm({ ...form, projectName: e.target.value })
        }
        required
      />
      <TextField
        label="Description"
        value={form.description}
        onChange={(e) =>
          setForm({ ...form, description: e.target.value })
        }
        multiline
        rows={3}
      />
    </FormDialog>
  );
};

/* ========================= */
/* Join Dialog */
/* ========================= */

const JoinProjectDialog = ({ open, onClose }: JoinDialogProps) => {
  const { joinProject } = useAppData();
  const [projectId, setProjectId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    if (!projectId.trim()) {
      setError("Project ID is required.");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await joinProject(projectId.trim());
      onClose();
    } catch {
      setError("Failed to join project.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <FormDialog
      open={open}
      onClose={onClose}
      onSubmit={handleSubmit}
      title="Join Existing Project"
      submitLabel="Join"
      loading={loading}
      error={error}
    >
      <TextField
        label="Project ID"
        value={projectId}
        onChange={(e) => setProjectId(e.target.value)}
        required
      />
    </FormDialog>
  );
};

/* ========================= */
/* Main Page */
/* ========================= */

export const Projects = () => {
  const { user } = useAuth();
  const {
    projects,
    hardware,
    loadingProjects,
    joinProject,
    leaveProject,
    deleteProject,
  } = useAppData();

  const userId = user?.userId ?? "";

  const [filter, setFilter] = useState<ProjectFilter>("assigned");
  const [error, setError] = useState<string | null>(null);
  const [manageProject, setManageProject] = useState<Project | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [joinOpen, setJoinOpen] = useState(false);

  const filtered = useMemo(() => {
    switch (filter) {
      case "owner":
        return projects.filter((p) => p.ownerUserId === userId);
      case "assigned":
        return projects.filter((p) =>
          p.assignedUsers?.includes(userId),
        );
      default:
        return projects.filter(
          (p) =>
            p.ownerUserId === userId ||
            p.assignedUsers?.includes(userId),
        );
    }
  }, [projects, filter, userId]);

  const getHardwareForProject = (project: Project) => {
    const hwIds =
      project.assignedHardware?.map((ah) => ah.hardwareId) ?? [];
    return hardware.filter((h) => hwIds.includes(h._id));
  };

  const buildActions = (project: Project): CardAction[] => {
    const isOwner = project.ownerUserId === userId;
    const isAssigned = project.assignedUsers?.includes(userId);

    const actions: CardAction[] = [];

    if (!isAssigned) {
      actions.push({
        label: "Join",
        onClick: () => joinProject(project.projectId),
        variant: "contained",
      });
    }

    if (isAssigned) {
      actions.push({
        label: "Manage",
        onClick: () => setManageProject(project),
        variant: "outlined",
      });

      actions.push({
        label: "Leave",
        onClick: () => leaveProject(project.projectId),
        variant: "outlined",
      });
    }

    if (isOwner) {
      actions.push({
        label: "Delete",
        onClick: () => deleteProject(project.projectId),
        variant: "outlined",
        color: "error",
      });
    }

    return actions;
  };

  return (
    <div className={styles.root}>
      <Stack
        direction="row"
        justifyContent="space-between"
        flexWrap="wrap"
        gap={2}
      >
        <ToggleButtonGroup
          value={filter}
          exclusive
          onChange={(_, val) => val && setFilter(val)}
          size="small"
        >
          <ToggleButton value="all">All</ToggleButton>
          <ToggleButton value="owner">Owner</ToggleButton>
          <ToggleButton value="assigned">Assigned</ToggleButton>
        </ToggleButtonGroup>

        <Stack direction="row" gap={1}>
          <Button variant="outlined" onClick={() => setJoinOpen(true)}>
            Join Project
          </Button>
          <Button variant="contained" onClick={() => setCreateOpen(true)}>
            New Project
          </Button>
        </Stack>
      </Stack>

      {error && (
        <Alert severity="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {loadingProjects ? (
        <Typography>Loading projects…</Typography>
      ) : filtered.length === 0 ? (
        <Card>
          <CardContent>
            <Typography align="center">
              No projects found.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <div className={styles.grid}>
          {filtered.map((project) => (
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

      <CreateProjectDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />

      <JoinProjectDialog
        open={joinOpen}
        onClose={() => setJoinOpen(false)}
      />
    </div>
  );
};

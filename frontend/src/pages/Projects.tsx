import { useMemo, useState } from "react";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import ToggleButton from "@mui/material/ToggleButton";
import ToggleButtonGroup from "@mui/material/ToggleButtonGroup";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";
import { useAppData } from "../context/AppContext";
import { FormDialog, ProjectCard } from "../components";
import type { Project } from "../api/projects";
import styles from "./homePage.module.css";

type ProjectFilter = "all" | "owner" | "assigned";

type CreateDialogProps = {
  open: boolean;
  onClose: () => void;
};

const CreateProjectDialog = ({ open, onClose }: CreateDialogProps) => {
  const { createProject } = useAppData();
  const [form, setForm] = useState({
    projectId: "",
    projectName: "",
    description: "",
  });
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleChange = (field: string, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    setError(null);
  };

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
        ownerUserId: "", // AppContext fills this in
      });
      setForm({ projectId: "", projectName: "", description: "" });
      onClose();
    } catch (err: unknown) {
      const msg =
        err &&
        typeof err === "object" &&
        "response" in err &&
        (err as { response?: { data?: { error?: string } } }).response?.data
          ?.error;
      setError(
        (typeof msg === "string" ? msg : null) || "Failed to create project.",
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
      title="Create New Project"
      submitLabel="Create"
      loading={loading}
      error={error}
    >
      <TextField
        label="Project ID"
        value={form.projectId}
        onChange={(e) => handleChange("projectId", e.target.value)}
        required
        helperText="Unique identifier (cannot be changed later)"
        autoFocus
      />
      <TextField
        label="Project Name"
        value={form.projectName}
        onChange={(e) => handleChange("projectName", e.target.value)}
        required
      />
      <TextField
        label="Description"
        value={form.description}
        onChange={(e) => handleChange("description", e.target.value)}
        multiline
        rows={3}
      />
    </FormDialog>
  );
};

type JoinDialogProps = {
  open: boolean;
  onClose: () => void;
};

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
      setProjectId("");
      onClose();
    } catch {
      setError("Failed to join project. Check the ID and try again.");
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
        onChange={(e) => {
          setProjectId(e.target.value);
          setError(null);
        }}
        required
        helperText="Ask the project owner for this ID"
        autoFocus
      />
    </FormDialog>
  );
};

export const Projects = () => {
  const { user } = useAuth();
  const {
    projects,
    hardware,
    loadingProjects,
    leaveProject,
    deleteProject,
    joinProject,
  } = useAppData();

  const [error, setError] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [joinOpen, setJoinOpen] = useState(false);
  // Default filter: only show projects user is assigned to or owns
  const [filter, setFilter] = useState<ProjectFilter>("assigned");

  const userId = user?.userId ?? "";

  const filtered = useMemo(() => {
    switch (filter) {
      case "owner":
        return projects.filter((p) => p.ownerUserId === userId);
      case "assigned":
        return projects.filter((p) => p.assignedUsers.includes(userId));
      default:
        // Only show projects user owns or is assigned to
        return projects.filter(
          (p) => p.ownerUserId === userId || p.assignedUsers.includes(userId)
        );
    }
  }, [projects, filter, userId]);

  /** Get hardware objects assigned to a project */
  const getHardwareForProject = (project: Project) => {
    const hwIds = project.assignedHardware.map((ah) => ah.hardwareId);
    return hardware.filter((h) => hwIds.includes(h._id));
  };

  const handleLeave = async (project: Project) => {
    try {
      await leaveProject(project.projectId);
    } catch {
      setError("Failed to leave project.");
    }
  };

  const handleJoin = async (project: Project) => {
    try {
      await joinProject(project.projectId);
    } catch {
      setError("Failed to join project.");
    }
  };

  const handleDelete = async (project: Project) => {
    try {
      await deleteProject(project.projectId);
    } catch {
      setError("Failed to delete project.");
    }
  };

  return (
    <div className={styles.root}>
      <Stack
        direction="row"
        alignItems="center"
        justifyContent="space-between"
        gap={2}
        flexWrap="wrap"
      >
        {/* Filter buttons */}
        <ToggleButtonGroup
          value={filter}
          exclusive
          onChange={(_e, val) => val && setFilter(val as ProjectFilter)}
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
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {loadingProjects ? (
        <Typography color="text.secondary">Loading projects…</Typography>
      ) : filtered.length === 0 ? (
        <Card variant="outlined">
          <CardContent>
            <Typography color="text.secondary" align="center">
              {filter === "all"
                ? "No projects yet. Create one or join an existing project to get started."
                : `No projects match the "${filter}" filter.`}
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <div className={styles.grid}>
          {filtered.map((project) => {
            const isOwner = project.ownerUserId === userId;
            const isAssigned = project.assignedUsers.includes(userId);
            return (
              <ProjectCard
                key={project._id}
                project={project}
                hardware={getHardwareForProject(project)}
                actions={[
                  ...[
                    {
                      label: isAssigned ? "Leave" : "Join",
                      onClick: isAssigned
                        ? () => handleLeave(project)
                        : () => handleJoin(project),
                      variant: "contained" as const,
                    },
                  ],
                  ...(isOwner
                    ? [
                        {
                          label: "Delete",
                          onClick: () => handleDelete(project),
                          variant: "outlined" as const,
                          color: "error" as const,
                        },
                      ]
                    : []),
                ]}
              />
            );
          })}
        </div>
      )}

      <CreateProjectDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />

      <JoinProjectDialog open={joinOpen} onClose={() => setJoinOpen(false)} />
    </div>
  );
};

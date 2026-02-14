import { useCallback, useEffect, useState } from "react";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import Chip from "@mui/material/Chip";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import Divider from "@mui/material/Divider";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";
import {
  projectsApi,
  type CreateProjectRequest,
  type Project,
} from "../api/projects";

// ── Create Project Dialog ───────────────────────────────────────

type CreateDialogProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (project: Project) => void;
  ownerUserId: string;
};

const CreateProjectDialog = ({
  open,
  onClose,
  onCreated,
  ownerUserId,
}: CreateDialogProps) => {
  const [form, setForm] = useState({
    projectId: "",
    projectName: "",
    description: "",
  });
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleChange = (field: string, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    setError(null);
  };

  const handleSubmit = async () => {
    if (!form.projectId.trim() || !form.projectName.trim()) {
      setError("Project ID and Project Name are required.");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const payload: CreateProjectRequest = {
        projectId: form.projectId.trim(),
        projectName: form.projectName.trim(),
        description: form.description.trim(),
        ownerUserId,
      };
      const { data } = await projectsApi.create(payload);
      onCreated(data);
      setForm({ projectId: "", projectName: "", description: "" });
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
        setError("Failed to create project.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>Create New Project</DialogTitle>
      <DialogContent
        sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 1 }}
      >
        {error && <Alert severity="error">{error}</Alert>}
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

// ── Join Project Dialog ─────────────────────────────────────────

type JoinDialogProps = {
  open: boolean;
  onClose: () => void;
  onJoined: (project: Project) => void;
  userId: string;
};

const JoinProjectDialog = ({
  open,
  onClose,
  onJoined,
  userId,
}: JoinDialogProps) => {
  const [projectMongoId, setProjectMongoId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (!projectMongoId.trim()) {
      setError("Project ID is required.");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const { data } = await projectsApi.join(projectMongoId.trim(), userId);
      onJoined(data);
      setProjectMongoId("");
      onClose();
    } catch {
      setError("Failed to join project. Check the ID and try again.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>Join Existing Project</DialogTitle>
      <DialogContent
        sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 1 }}
      >
        {error && <Alert severity="error">{error}</Alert>}
        <TextField
          label="Project Mongo ID"
          value={projectMongoId}
          onChange={(e) => {
            setProjectMongoId(e.target.value);
            setError(null);
          }}
          required
          helperText="Ask the project owner for this ID"
          autoFocus
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
          {submitting ? "Joining…" : "Join"}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

// ── Project Card ────────────────────────────────────────────────

type ProjectCardProps = {
  project: Project;
  currentUserId: string;
  onLeave: (project: Project) => void;
  onDelete: (project: Project) => void;
};

const ProjectCard = ({
  project,
  currentUserId,
  onLeave,
  onDelete,
}: ProjectCardProps) => {
  const isOwner = project.ownerUserId === currentUserId;

  return (
    <Card variant="outlined">
      <CardContent>
        <Typography variant="h6">{project.projectName}</Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
          ID: {project.projectId}
        </Typography>
        {project.description && (
          <Typography variant="body2" sx={{ mb: 1 }}>
            {project.description}
          </Typography>
        )}
        <Divider sx={{ my: 1 }} />
        <Typography variant="caption" color="text.secondary">
          Members:
        </Typography>
        <Box sx={{ display: "flex", gap: 0.5, flexWrap: "wrap", mt: 0.5 }}>
          {project.assignedUsers.map((uid) => (
            <Chip
              key={uid}
              label={uid}
              size="small"
              color={uid === project.ownerUserId ? "primary" : "default"}
              variant={uid === currentUserId ? "filled" : "outlined"}
            />
          ))}
        </Box>
      </CardContent>
      <CardActions>
        {isOwner ? (
          <Button size="small" color="error" onClick={() => onDelete(project)}>
            Delete
          </Button>
        ) : (
          <Button size="small" color="warning" onClick={() => onLeave(project)}>
            Leave
          </Button>
        )}
      </CardActions>
    </Card>
  );
};

// ── Main Projects Page ──────────────────────────────────────────

export const Projects = () => {
  const { user } = useAuth();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [joinOpen, setJoinOpen] = useState(false);

  const userId = user?.userId ?? "";

  const fetchProjects = useCallback(async () => {
    if (!userId) return;
    setLoading(true);
    setError(null);
    try {
      const { data } = await projectsApi.list({ assignedUser: userId });
      setProjects(data);
    } catch {
      setError("Failed to load projects.");
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  const handleCreated = (project: Project) => {
    setProjects((prev) => [project, ...prev]);
  };

  const handleJoined = (project: Project) => {
    setProjects((prev) => {
      const exists = prev.find((p) => p._id === project._id);
      if (exists) {
        return prev.map((p) => (p._id === project._id ? project : p));
      }
      return [project, ...prev];
    });
  };

  const handleLeave = async (project: Project) => {
    try {
      await projectsApi.leave(project._id, userId);
      setProjects((prev) => prev.filter((p) => p._id !== project._id));
    } catch {
      setError("Failed to leave project.");
    }
  };

  const handleDelete = async (project: Project) => {
    try {
      await projectsApi.delete(project._id);
      setProjects((prev) => prev.filter((p) => p._id !== project._id));
    } catch {
      setError("Failed to delete project.");
    }
  };

  return (
    <Box sx={{ py: 3, px: 2, maxWidth: 900, mx: "auto" }}>
      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          mb: 3,
        }}
      >
        <Typography variant="h4">Projects</Typography>
        <Box sx={{ display: "flex", gap: 1 }}>
          <Button variant="outlined" onClick={() => setJoinOpen(true)}>
            Join Project
          </Button>
          <Button variant="contained" onClick={() => setCreateOpen(true)}>
            New Project
          </Button>
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Typography color="text.secondary">Loading projects…</Typography>
      ) : projects.length === 0 ? (
        <Card variant="outlined">
          <CardContent>
            <Typography color="text.secondary" align="center">
              No projects yet. Create one or join an existing project to get
              started.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <Grid container spacing={2}>
          {projects.map((project) => (
            <Grid size={{ xs: 12, sm: 6 }} key={project._id}>
              <ProjectCard
                project={project}
                currentUserId={userId}
                onLeave={handleLeave}
                onDelete={handleDelete}
              />
            </Grid>
          ))}
        </Grid>
      )}

      <CreateProjectDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={handleCreated}
        ownerUserId={userId}
      />

      <JoinProjectDialog
        open={joinOpen}
        onClose={() => setJoinOpen(false)}
        onJoined={handleJoined}
        userId={userId}
      />
    </Box>
  );
};

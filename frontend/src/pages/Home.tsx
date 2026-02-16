import { useState } from "react";
import Alert from "@mui/material/Alert";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";
import { useAppData } from "../context/AppContext";
import { ProjectCard } from "../components";
import type { Project } from "../api/projects";
import styles from "./homePage.module.css";

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
  const userId = user?.userId ?? "";

  const getHardwareForProject = (project: Project) => {
    const hwIds = project.assignedHardware.map((ah) => ah.hardwareId);
    return hardware.filter((h) => hwIds.includes(h._id));
  };

  const handleAction = async (project: Project) => {
    try {
      const isOwner = project.ownerUserId === userId;
      const isAssigned = project.assignedUsers.includes(userId);

      if (isOwner) {
        await deleteProject(project.projectId);
      } else if (isAssigned) {
        await leaveProject(project.projectId);
      } else {
        await joinProject(project.projectId);
      }
    } catch {
      setError("Action failed. Please try again.");
    }
  };

  const getButtonLabel = (project: Project) => {
    if (project.ownerUserId === userId) return "Delete";
    if (project.assignedUsers.includes(userId)) return "Leave";
    return "Join";
  };

  const getButtonColor = (
    project: Project,
  ): "error" | "warning" | "primary" | undefined => {
    if (project.ownerUserId === userId) return "error";
    return undefined;
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
          {projects.map((project) => {
            const label = getButtonLabel(project);
            const color = getButtonColor(project);
            return (
              <ProjectCard
                key={project._id}
                project={project}
                hardware={getHardwareForProject(project)}
                buttonLabel={label}
                onButtonClick={() => handleAction(project)}
                secondaryColor={color}
              />
            );
          })}
        </div>
      )}
    </div>
  );
};

import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import CardActions from "@mui/material/CardActions";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";

import type { Project } from "../api/projects";
import type { Hardware } from "../api/hardware";
import styles from "./projectCard.module.css";
import { UserList } from "./UserList";
import { HardwareList } from "./HardwareList";
import { ProjectButton } from "./ProjectButton";

export type CardAction = {
  label: string;
  onClick: () => void;
  variant?: "contained" | "outlined" | "text";
  color?: "primary" | "error" | "warning" | "inherit";
};

type ProjectCardProps = {
  project: Project;
  hardware?: Hardware[];
  actions?: CardAction[];
};

export const ProjectCard = ({
  project,
  hardware = [],
  actions = [],
}: ProjectCardProps) => {
  return (
    <Card className={styles.card} variant="outlined">
      <CardContent className={styles.content}>
        <div className={styles.header}>
          <Typography variant="h6" align="left">
            {project.projectName}
          </Typography>
          <Typography variant="body2" color="text.secondary" align="left">
            ID: {project.projectId}
          </Typography>
          {project.description && (
            <Typography variant="body2" align="left">
              {project.description}
            </Typography>
          )}
        </div>

        <Stack className={styles.sections} spacing={2}>
          <UserList
            users={project.assignedUsers}
            ownerUserId={project.ownerUserId}
          />
          {hardware.length > 0 && <HardwareList hardware={hardware} />}
        </Stack>
      </CardContent>

      {actions.length > 0 && (
        <CardActions className={styles.buttonContainer}>
          {actions.map((action) => (
            <ProjectButton
              key={action.label}
              label={action.label}
              onClick={action.onClick}
              variant={action.variant ?? "contained"}
              color={action.color}
            />
          ))}
        </CardActions>
      )}
    </Card>
  );
};

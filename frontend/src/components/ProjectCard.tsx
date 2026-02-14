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

type ProjectCardProps = {
  project: Project;
  hardware?: Hardware[];
  buttonLabel?: string;
  onButtonClick?: () => void;
  secondaryLabel?: string;
  onSecondaryClick?: () => void;
  secondaryColor?: "error" | "warning" | "primary";
};

export const ProjectCard = ({
  project,
  hardware = [],
  buttonLabel,
  onButtonClick,
  secondaryLabel,
  onSecondaryClick,
  secondaryColor = "error",
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

      {(buttonLabel || secondaryLabel) && (
        <CardActions className={styles.buttonContainer}>
          {buttonLabel && onButtonClick && (
            <ProjectButton
              label={buttonLabel}
              onClick={onButtonClick}
              variant="contained"
            />
          )}
          {secondaryLabel && onSecondaryClick && (
            <ProjectButton
              label={secondaryLabel}
              onClick={onSecondaryClick}
              variant="outlined"
              color={secondaryColor}
            />
          )}
        </CardActions>
      )}
    </Card>
  );
};

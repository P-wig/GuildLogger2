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
    <Card
      className={styles.card}
      variant="outlined"
      sx={{ borderRadius: 3 }}
    >
      <CardContent className={styles.content}>

        {/* ================= HEADER ================= */}
        <div className={styles.headerRow}>
          <Typography variant="h5" sx={{ fontWeight: 600 }}>
            {project.projectName}
          </Typography>

          <Typography
            variant="body2"
            color="text.secondary"
            sx={{ fontWeight: 500 }}
          >
            Project ID: {project.projectId}
          </Typography>
        </div>

        <div className={styles.divider} />

        {/* ================= DESCRIPTION ================= */}
        {project.description && (
          <>
            <div className={styles.section}>
              <Typography
                variant="subtitle2"
                className={styles.sectionTitle}
              >
                Project Description
              </Typography>

              <Typography
                variant="body2"
                color="text.secondary"
                className={styles.sectionText}
              >
                {project.description}
              </Typography>
            </div>

            <div className={styles.divider} />
          </>
        )}


        {/* ================= OWNER / USERS ================= */}
        <div className={styles.section}>
          <div className={styles.ownerUserRow}>
            <div>
              <Typography variant="subtitle2" className={styles.sectionTitle}>
                Owner
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {project.ownerUserId}
              </Typography>
            </div>

            <div>
              <Typography variant="subtitle2" className={styles.sectionTitle}>
                Users
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {project.assignedUsers?.join(", ")}
              </Typography>
            </div>
          </div>
        </div>

        <div className={styles.divider} />


        {/* ================= HARDWARE ================= */}
        {hardware.length > 0 && (
          <div className={styles.section}>
            <Typography
              variant="subtitle2"
              className={styles.sectionTitle}
            >
              Hardware
            </Typography>

            <HardwareList
              hardware={hardware}
              displayTitle={false}
            />
          </div>
        )}

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

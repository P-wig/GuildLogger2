import Button from "@mui/material/Button";
import styles from "./projectButton.module.css";

type ProjectButtonProps = {
  label: string;
  onClick: () => void;
  variant?: "text" | "outlined" | "contained";
  color?: "primary" | "error" | "warning" | "inherit";
  disabled?: boolean;
};

export const ProjectButton = ({
  label,
  onClick,
  variant = "text",
  color = "primary",
  disabled = false,
}: ProjectButtonProps) => {
  return (
    <Button
      className={styles.button}
      variant={variant}
      color={color}
      onClick={onClick}
      disabled={disabled}
      size="small"
    >
      {label}
    </Button>
  );
};

import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import LinearProgress from "@mui/material/LinearProgress";
import styles from "./hardwareList.module.css";
import type { Hardware } from "../api/hardware";

type HardwareListProps = {
  hardware: Hardware[];
};

function clampPercent(value: number) {
  return Math.max(0, Math.min(100, value));
}

export const HardwareList = ({ hardware }: HardwareListProps) => {
  return (
    <div className={styles.root}>
      <Typography variant="subtitle2">Hardware</Typography>

      <Stack spacing={1.25}>
        {hardware.map((hw) => {
          const percentAvailable =
            hw.capacity === 0
              ? 0
              : clampPercent((hw.available / hw.capacity) * 100);

          return (
            <div key={hw._id} className={styles.item}>
              <div className={styles.row}>
                <Typography variant="body2" className={styles.name}>
                  {hw.hardwareName}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {hw.available}/{hw.capacity} available
                </Typography>
              </div>

              <LinearProgress
                variant="determinate"
                value={percentAvailable}
                color={
                  percentAvailable < 20
                    ? "error"
                    : percentAvailable < 50
                      ? "warning"
                      : "primary"
                }
                className={styles.bar}
              />
            </div>
          );
        })}
      </Stack>
    </div>
  );
};

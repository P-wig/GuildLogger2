import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Chip from "@mui/material/Chip";
import styles from "./userList.module.css";

type UserListProps = {
  users: string[];
  /** Highlight the owner with a primary-colored chip */
  ownerUserId?: string;
};

export const UserList = ({ users, ownerUserId }: UserListProps) => {
  return (
    <div className={styles.root}>
      <Typography variant="subtitle2">Users</Typography>
      <Stack direction="row" flexWrap="wrap" gap={1}>
        {users.map((u) => (
          <Chip
            key={u}
            label={u}
            size="small"
            color={u === ownerUserId ? "primary" : "default"}
          />
        ))}
      </Stack>
    </div>
  );
};

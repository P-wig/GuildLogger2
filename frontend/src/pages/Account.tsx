import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";

export const Account = () => {
  const { user } = useAuth();

  return (
    <>
      <Typography variant="h4" gutterBottom>
        Account
      </Typography>
      <Typography>Signed in as: {user?.email}</Typography>
    </>
  );
};

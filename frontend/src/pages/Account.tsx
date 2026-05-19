import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import { useAuth } from "../auth";

export const Account = () => {
  const { user } = useAuth();

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Account
      </Typography>

      <Card variant="outlined">
        <CardContent>
          <Typography variant="h6" gutterBottom>
            User Information
          </Typography>
          <Typography>
            <strong>Discord ID:</strong> {user?.discordId}
          </Typography>
          <Typography>
            <strong>Member since:</strong>{" "}
            {user?.createdAt
              ? new Date(user.createdAt).toLocaleDateString()
              : "—"}
          </Typography>
        </CardContent>
      </Card>
    </Box>
  );
};

import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import { useAuth } from "../auth";

export const Account = () => {
  const { user } = useAuth();

  console.log("User data in Account component:", user);
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
            <strong>User ID:</strong> {user?.userId}
          </Typography>
        </CardContent>
      </Card>
    </Box>
  );
};

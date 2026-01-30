import { Link as RouterLink } from "react-router";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";

export const Home = () => {
  const { isAuthenticated, user } = useAuth();

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Home
      </Typography>

      <Typography color="text.secondary" sx={{ mb: 2 }}>
        Yoooooo.
      </Typography>

      <Card variant="outlined">
        <CardContent>
          <Stack spacing={1}>
            <Typography variant="h6">Status</Typography>

            {isAuthenticated ? (
              <Typography>
                Signed in as <strong>{user?.email}</strong>
              </Typography>
            ) : (
              <Typography>You’re not signed in.</Typography>
            )}

            <Typography color="text.secondary">
              Use the buttons below to navigate.
            </Typography>
          </Stack>
        </CardContent>

        <CardActions>
          {!isAuthenticated ? (
            <Button variant="contained" component={RouterLink} to="/auth">
              Sign in
            </Button>
          ) : (
            <Button variant="contained" component={RouterLink} to="/account">
              Go to account
            </Button>
          )}

          <Button variant="outlined" component={RouterLink} to="/account">
            Account (protected)
          </Button>
        </CardActions>
      </Card>
    </Box>
  );
};

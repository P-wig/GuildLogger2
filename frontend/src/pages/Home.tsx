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
    <Box sx={{ mt: 0, pt: 0 }}>
      <Typography variant="h4" gutterBottom sx={{ mt: 0 }}>
        Home
      </Typography>

      <Typography color="text.secondary" sx={{ mb: 3 }}>
        Welcome to the Cloud Native Team Project.
      </Typography>

      <Stack spacing={3}>
        {/* Quick Links Card - Only show if authenticated */}
        {isAuthenticated && (
          <Card variant="outlined">
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Quick Actions
              </Typography>
              <Typography color="text.secondary" sx={{ mb: 2 }}>
                Welcome back, <strong>{user?.userId}</strong>! Access your account and project features below.
              </Typography>
            </CardContent>

            <CardActions>
              <Button variant="contained" component={RouterLink} to="/account">
                Go to Account
              </Button>
              <Button variant="outlined" component={RouterLink} to="/projects">
                View Projects
              </Button>
            </CardActions>
          </Card>
        )}

        {/* Sign In Prompt - Only show if not authenticated */}
        {!isAuthenticated && (
          <Card variant="outlined">
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Get Started
              </Typography>
              <Typography color="text.secondary" sx={{ mb: 2 }}>
                Please sign in to access your account and manage projects.
              </Typography>
            </CardContent>

            <CardActions sx={{ justifyContent: 'center' }}>
              <Button variant="contained" component={RouterLink} to="/auth">
                Sign In
              </Button>
            </CardActions>
          </Card>
        )}
      </Stack>
    </Box>
  );
};
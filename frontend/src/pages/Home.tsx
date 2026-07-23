import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import { Link as RouterLink } from "react-router";
import { useAuth } from "../auth";

export const Home = () => {
  const { isAuthenticated } = useAuth();

  return (
    <Box sx={{ textAlign: "center" }}>
      <Typography variant="h4" gutterBottom>
        Dashboard
      </Typography>

      <Typography color="text.secondary" sx={{ mb: 4 }}>
        Manage your Discord guilds, events, and members.
      </Typography>

      {isAuthenticated ? (
        <Card variant="outlined" sx={{ maxWidth: 700, mx: "auto", p: 3 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Quick Actions
            </Typography>

            <Typography color="text.secondary" sx={{ mb: 3 }}>
              Welcome back!
            </Typography>

            <Grid container spacing={2} justifyContent="center">
              <Grid size={{ xs: 12, sm: "auto" }}>
                <Button
                  variant="contained"
                  fullWidth
                  component={RouterLink}
                  to="/app/guilds"
                  size="large"
                >
                  My Guilds
                </Button>
              </Grid>

            </Grid>
          </CardContent>
        </Card>
      ) : (
        <Card variant="outlined" sx={{ maxWidth: 400, mx: "auto", p: 3 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Get Started
            </Typography>

            <Typography color="text.secondary" sx={{ mb: 3 }}>
              Sign in with Discord to manage your guilds and events.
            </Typography>

            <Stack alignItems="center">
              <Button
                variant="contained"
                component={RouterLink}
                to="/auth"
                size="large"
              >
                Sign In
              </Button>
            </Stack>
          </CardContent>
        </Card>
      )}
    </Box>
  );
};
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";

import { Box } from "@mui/material";
import Grid from "@mui/material/Grid";
import Stack from "@mui/material/Stack";
import { Link as RouterLink } from "react-router";
import Button from "@mui/material/Button";

export const Home = () => {
  const { isAuthenticated, user } = useAuth();

  return (
    <Box sx={{ textAlign: "center" }}>
      <Typography variant="h4" gutterBottom>
        Home
      </Typography>

      <Typography color="text.secondary" sx={{ mb: 4 }}>
        Welcome to the Cloud Native Team Project.
      </Typography>

      {isAuthenticated ? (
        <Card
          variant="outlined"
          sx={{
            maxWidth: 700,
            mx: "auto",
            p: 3,
          }}
        >
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Quick Actions
            </Typography>

            <Typography color="text.secondary" sx={{ mb: 3 }}>
              Welcome back, <strong>{user?.userId}</strong>!
            </Typography>

            <Grid
              container
              spacing={2}
              justifyContent="center"
            >
              <Grid size= {{ xs: 12, sm: "auto" }}>
                <Button
                  variant="contained"
                  fullWidth
                  component={RouterLink}
                  to="/account"
                  size="large"
                >
                  ACCOUNT
                </Button>
              </Grid>

              <Grid size= {{ xs: 12, sm: "auto" }}>
                <Button
                  variant="outlined"
                  fullWidth
                  component={RouterLink}
                  to="/projects"
                  size="large"
                >
                  PROJECTS
                </Button>
              </Grid>

              <Grid size= {{ xs: 12, sm: "auto" }}>
                <Button
                  variant="outlined"
                  fullWidth
                  component={RouterLink}
                  to="/hardware"
                  size="large"
                >
                  HARDWARE
                </Button>
              </Grid>
            </Grid>

          </CardContent>
        </Card>
      ) : (
        <Card
          variant="outlined"
          sx={{
            maxWidth: 400,
            mx: "auto",
            p: 3,
          }}
        >
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Get Started
            </Typography>

            <Typography color="text.secondary" sx={{ mb: 3 }}>
              Please sign in to access your account and manage projects.
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
import { Link } from "react-router";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";

export const Welcome = () => {
  return (
    <Container maxWidth="md">
      <Box
        sx={{
          minHeight: "80vh",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
          textAlign: "center",
          py: 4,
        }}
      >
        <Typography
          variant="h2"
          component="h1"
          gutterBottom
          sx={{ fontWeight: 600, mb: 3 }}
        >
          Welcome to Your App
        </Typography>

        <Typography
          variant="h5"
          component="h2"
          color="text.secondary"
          sx={{ mb: 4, maxWidth: "600px" }}
        >
          Your full-stack application boilerplate with authentication, 
          modern UI, and everything you need to get started.
        </Typography>

        <Card variant="outlined" sx={{ maxWidth: 400, width: "100%" }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Get Started
            </Typography>
            <Typography color="text.secondary" sx={{ mb: 3 }}>
              Sign in to access your account and explore the features.
            </Typography>

            <Stack spacing={2}>
              <Button
                component={Link}
                to="/auth"
                variant="contained"
                size="large"
                fullWidth
              >
                Sign In
              </Button>
              
              <Typography variant="body2" color="text.secondary">
                New user? You can create an account on the sign-in page.
              </Typography>
            </Stack>
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
};
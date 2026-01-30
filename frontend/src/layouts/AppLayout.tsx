import { Link, Outlet } from "react-router";
import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";

export const AppLayout = () => {
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <Box>
      <AppBar position="static">
        <Toolbar>
          <Typography sx={{ flexGrow: 1 }} variant="h6">
            Hardware Checkout App
          </Typography>

          <Button color="inherit" component={Link} to="/">
            Home
          </Button>

          <Button color="inherit" component={Link} to="/account">
            Account
          </Button>

          {!isAuthenticated ? (
            <Button color="inherit" component={Link} to="/auth">
              Sign in
            </Button>
          ) : (
            <Button color="inherit" onClick={logout}>
              Sign out ({user?.email})
            </Button>
          )}
        </Toolbar>
      </AppBar>

      <Container sx={{ py: 3 }}>
        <Outlet />
      </Container>
    </Box>
  );
};

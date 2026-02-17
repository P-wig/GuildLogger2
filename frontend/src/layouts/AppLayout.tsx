import { Link, Outlet, useLocation } from "react-router";
import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";

const activeNavSx = {
  textShadow: "0 0 8px rgba(255,255,255,0.9), 0 0 16px rgba(255,255,255,0.5)",
  fontWeight: 700,
} as const;

export const AppLayout = () => {
  const { isAuthenticated, user, logout } = useAuth();
  const { pathname } = useLocation();

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
        minHeight: "100vh",
      }}
    >
      {/* Header AppBar */}
      <AppBar position="fixed" sx={{ top: 0, left: 0, right: 0, zIndex: 1200 }}>
        <Toolbar sx={{ justifyContent: "space-between" }}>
          {isAuthenticated && (
            <Box sx={{ display: "flex", gap: 2 }}>
              {[
                { to: "/home", label: "Home" },
                { to: "/account", label: "Account" },
                { to: "/projects", label: "Projects" },
                { to: "/hardware", label: "Hardware" },
              ].map(({ to, label }) => (
                <Button
                  key={to}
                  color="inherit"
                  component={Link}
                  to={to}
                  sx={pathname === to ? activeNavSx : undefined}
                >
                  {label}
                </Button>
              ))}
            </Box>
          )}

          {/* Center - App Title */}
          <Typography
            variant="h6"
            sx={{
              position: "absolute",
              left: "50%",
              transform: "translateX(-50%)",
              fontWeight: 600,
            }}
          >
            Hardware Checkout App
          </Typography>

          {/* Right side - Auth button */}
          <Box>
            {!isAuthenticated ? (
              <Button color="inherit" component={Link} to="/auth">
                Sign in
              </Button>
            ) : (
              <Button color="inherit" onClick={logout}>
                Sign out ({user?.userId})
              </Button>
            )}
          </Box>
        </Toolbar>
      </AppBar>

      {/* Main Content Area */}
      <Container
        component="main"
        sx={{
          flexGrow: 1,
          py: 3,
          mt: 8,
          mb: 8,
        }}
      >
        <Outlet />
      </Container>

      {/* Footer AppBar */}
      <AppBar
        component="footer"
        position="fixed"
        sx={{
          top: "auto",
          bottom: 0,
          backgroundColor: (theme) =>
            theme.palette.mode === "light"
              ? theme.palette.grey[800]
              : theme.palette.grey[900],
        }}
      >
        <Toolbar
          sx={{ justifyContent: "center", minHeight: "48px !important" }}
        >
          <Typography
            variant="body2"
            color="inherit"
            align="center"
            sx={{ opacity: 0.8 }}
          >
            © {new Date().getFullYear()} Cloud Native Team Project. All rights
            reserved.
          </Typography>
        </Toolbar>
      </AppBar>
    </Box>
  );
};

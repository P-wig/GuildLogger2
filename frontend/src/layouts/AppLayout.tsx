import { Link, Outlet, useLocation } from "react-router";
import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import { useAuth } from "../auth";

import { useState } from "react"; // local UI state
import Drawer from "@mui/material/Drawer"; // slide menu
import IconButton from "@mui/material/IconButton"; // hamburger button
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemText from "@mui/material/ListItemText";
import MenuIcon from "@mui/icons-material/Menu"; // hamburger icon

const activeNavSx = {
  textShadow: "0 0 8px rgba(255,255,255,0.9), 0 0 16px rgba(255,255,255,0.5)",
  fontWeight: 700,
} as const;

export const AppLayout = () => {
  const { isAuthenticated, user, logout } = useAuth();
  const { pathname } = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);

  const handleDrawerToggle = () => {
    setMobileOpen((prev) => !prev);
  };

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
        <Toolbar sx={{ display: "flex", alignItems: "center" }}>

          {/* LEFT: App Title (desktop) */}
          <Box sx={{ display: { xs: "none", md: "flex" }, alignItems: "center" }}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              Hardware Checkout App
            </Typography>
          </Box>

          {/* CENTER: Desktop Navigation */}
          {isAuthenticated && (
            <Box
              sx={{
                flexGrow: 1,
                display: { xs: "none", md: "flex" },
                justifyContent: "center",
                gap: 2,
              }}
            >
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

          {/* MOBILE CENTER TITLE */}
          <Box
            sx={{
              display: { xs: "flex", md: "none" },
              flexGrow: 1,
              justifyContent: "center",
            }}
          >
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              Hardware Checkout App
            </Typography>
          </Box>

          {/* RIGHT SECTION */}
          <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>

            {/* Auth button on desktop */}
            {isAuthenticated && (
              <Box sx={{ display: { xs: "none", md: "flex" } }}>
                <Button color="inherit" onClick={logout}>
                  Sign out ({user?.userId})
                </Button>
              </Box>
            )}

            {/* Mobile hamburger */}
            {isAuthenticated && (
              <Box sx={{ display: { xs: "flex", md: "none" } }}>
                <IconButton
                  color="inherit"
                  onClick={handleDrawerToggle}
                >
                  <MenuIcon />
                </IconButton>
              </Box>
            )}
          </Box>
        </Toolbar>
      </AppBar>

      <Drawer
        anchor="right"
        open={mobileOpen}
        onClose={handleDrawerToggle}
      >
        <Box
          sx={{ width: 250 }}
          role="presentation"
          onClick={handleDrawerToggle}
        >
          <List>
            {[
              { to: "/home", label: "Home" },
              { to: "/account", label: "Account" },
              { to: "/projects", label: "Projects" },
              { to: "/hardware", label: "Hardware" },
            ].map(({ to, label }) => (
              <ListItem key={to} disablePadding>
                <ListItemButton component={Link} to={to}>
                  <ListItemText primary={label} />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        </Box>
      </Drawer>

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

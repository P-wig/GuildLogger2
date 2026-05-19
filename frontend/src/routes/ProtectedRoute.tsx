import { Navigate, Outlet, useLocation } from "react-router";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import { useAuth } from "../auth";

/**
 * ProtectedRoute — client-side UX guard only.
 *
 * Checks whether the user has an active session in AuthContext (sourced from
 * localStorage) before rendering any child route. If no session exists, the
 * user is redirected to /auth immediately, before any API call is made.
 *
 * IMPORTANT: This guard provides no real security. A user could set localStorage
 * manually and bypass it. All real enforcement happens at the backend via JWT
 * middleware, which validates the token signature and expiry on every protected
 * API request.
 *
 * How the two layers interact:
 *
 *   User navigates to protected route
 *           │
 *           ▼
 *   ProtectedRoute checks localStorage
 *           │
 *      No token? ──────────────────► redirect to /auth (no API call made)
 *           │
 *      Token exists
 *           ▼
 *   Page renders and calls backend API
 *           │
 *           ▼
 *   Backend JWT middleware validates token
 *           │
 *      Invalid/expired? ──────────────► 401 → http.ts clears localStorage
 *           │
 *      Valid
 *           ▼
 *   Handler runs, returns data
 *
 * See: docs/DEVELOPER_GUIDE.md — "Authentication Architecture"
 */

export const ProtectedRoute = () => {
  const { isAuthenticated, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return (
      <Box
        sx={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "60vh" }}
      >
        <CircularProgress />
      </Box>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth" replace state={{ from: location }} />;
  }

  return <Outlet />;
};

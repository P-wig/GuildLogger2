import { Navigate, Outlet, useLocation, type Location } from "react-router";
import { useAuth } from "../auth";

type LocationState = {
  from?: Location;
};

export const AuthRoute = () => {
  const { isAuthenticated, loading } = useAuth();
  const location = useLocation();

  if (loading) return null;

  if (isAuthenticated) {
    const state = location.state as LocationState | null;
    const fromPath = state?.from?.pathname;
    // Redirect authenticated users to the app instead of root
    const target = fromPath && fromPath !== "/auth" ? fromPath : "/app";
    return <Navigate to={target} replace />;
  }

  return <Outlet />;
};

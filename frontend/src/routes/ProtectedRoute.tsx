import { Navigate, Outlet, useLocation } from "react-router";
import { useAuth } from "../auth";

export const ProtectedRoute = () => {
  const { isAuthenticated } = useAuth();
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/auth" replace state={{ from: location }} />;
  }

  return <Outlet />;
};

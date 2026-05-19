import { createBrowserRouter } from "react-router";
import { AuthRoute } from "./AuthRoute";
import { ProtectedRoute } from "./ProtectedRoute";
import { AppLayout } from "../layouts";
import { Account, Auth, Home, Welcome, Guilds, Events } from "../pages";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <AppLayout />,
    children: [
      // Public welcome page - no authentication required
      { index: true, element: <Welcome /> },

      // Auth routes - only for unauthenticated users
      {
        element: <AuthRoute />,
        children: [{ path: "auth", element: <Auth /> }],
      },

      // Protected routes - require authentication
      {
        path: "app",
        element: <ProtectedRoute />,
        children: [
          { index: true, element: <Home /> },
          { path: "account", element: <Account /> },
          { path: "guilds", element: <Guilds /> },
          { path: "events", element: <Events /> },
        ],
      },
    ],
  },
]);

import { createBrowserRouter } from "react-router";
import { AuthRoute } from "./AuthRoute";
import { ProtectedRoute } from "./ProtectedRoute";
import { AppLayout } from "../layouts";
import { Account, Auth, Home, Welcome, Guilds, GuildDashboard, GuildEvents, GuildActiveEvents, EventLogSubmit } from "../pages";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <AppLayout />,
    children: [
      // Public welcome page - no authentication required
      { index: true, element: <Welcome /> },

      // Public event-log submission page — accessible without a user session.
      // The one-time signed token in the URL provides access control.
      { path: "log-event", element: <EventLogSubmit /> },

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
          { path: "guilds/:guildId", element: <GuildDashboard /> },
          { path: "guilds/:guildId/events", element: <GuildEvents /> },
          { path: "guilds/:guildId/active-events", element: <GuildActiveEvents /> },
        ],
      },
    ],
  },
]);

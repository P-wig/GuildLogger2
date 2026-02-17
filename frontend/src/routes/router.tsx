import { createBrowserRouter } from "react-router";
import { ProtectedRoute } from "./ProtectedRoute";
import { AppLayout } from "../layouts";
import { Account, Auth, Home, HardwarePage } from "../pages";
import { Projects } from "../pages/Projects";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <AppLayout />,
    children: [
      { index: true, element: <Auth /> },
      { path: "auth", element: <Auth /> },

      // Authenticated routes
      {
        element: <ProtectedRoute />,
        children: [
          { path: "home", element: <Home /> },
          { path: "account", element: <Account /> },
          { path: "projects", element: <Projects /> },
          { path: "hardware", element: <HardwarePage /> },
        ],
      },
    ],
  },
]);

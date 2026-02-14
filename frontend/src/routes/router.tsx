import { createBrowserRouter } from "react-router";
import { ProtectedRoute } from "./ProtectedRoute";
import { AppLayout } from "../layouts";
import { Account, Auth, Home, } from "../pages";
import { Projects } from "../pages/Projects";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <AppLayout />,
    children: [
      { index: true, element: <Home /> },
      { path: "auth", element: <Auth /> },

      // Authenticated routes
      {
        element: <ProtectedRoute />,
        children: [
          { path: "account", element: <Account /> },
          { path: "projects", element: <Projects /> },
        ],
      },
    ],
  },
]);

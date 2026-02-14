import { RouterProvider } from "react-router";
import "./App.css";
import { AuthProvider } from "./auth";
import { router } from "./routes";
import { ProjectProvider } from "./projects/ProjectContext";

function App() {
  return (
    <AuthProvider>
      <ProjectProvider> // wrapped App with ProjectProvider
        <RouterProvider router={router} />
      </ProjectProvider>
    </AuthProvider>
  );
}

export default App;

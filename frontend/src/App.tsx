import { RouterProvider } from "react-router";
import "./App.css";
import { AuthProvider } from "./auth";
import { AppProvider } from "./context/AppContext";
import { router } from "./routes";

function App() {
  return (
    // manages user session
    <AuthProvider> 
      {/* manages app data (projects, hardware) */}
      <AppProvider>
        <RouterProvider router={router} />
      </AppProvider>
    </AuthProvider>
  );
}

export default App;

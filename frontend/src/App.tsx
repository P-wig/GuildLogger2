import { RouterProvider } from "react-router";
import "./App.css";
import { AuthProvider } from "./auth";
import { AppProvider } from "./context/AppContext";
import { router } from "./routes";

function App() {
  return (
    <AuthProvider>
      <AppProvider>
        <RouterProvider router={router} />
      </AppProvider>
    </AuthProvider>
  );
}

export default App;

import { RouterProvider } from "react-router";
import "./App.css";
import { AuthProvider } from "./auth";
import { router } from "./routes";

function App() {
  return (
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  );
}

export default App;

import { RouterProvider } from "react-router";
import "./App.css";
import { AuthProvider } from "./auth";
import { AppProvider } from "./context/AppContext";
import { router } from "./routes";

function App() {
  return (
    <AuthProvider> 
      {/* Clean boilerplate context - add your domain data here */}
      <AppProvider>
        <RouterProvider router={router} />
      </AppProvider>
    </AuthProvider>
  );
}

export default App;

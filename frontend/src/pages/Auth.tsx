import { useLocation, useNavigate } from "react-router";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { useMemo, useState } from "react";
import { useAuth } from "../auth";

export const Auth = () => {
  const { login } = useAuth();
  const [email, setEmail] = useState("test@example.com");

  const navigate = useNavigate();
  const location = useLocation();

  const redirectTo = useMemo(() => {
    const state = location.state as { from?: { pathname?: string } } | null;
    return state?.from?.pathname ?? "/account";
  }, [location.state]);

  return (
    <Box sx={{ maxWidth: 420 }}>
      <Typography variant="h4" gutterBottom>
        Sign in
      </Typography>

      <TextField
        fullWidth
        label="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        sx={{ mb: 2 }}
      />

      <Button
        variant="contained"
        onClick={() => {
          login(email);
          navigate(redirectTo, { replace: true });
        }}
      >
        Sign in
      </Button>
    </Box>
  );
};

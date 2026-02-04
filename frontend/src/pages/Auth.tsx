import React, { useState } from "react";
import { useLocation, useNavigate } from "react-router";
import Box from "@mui/material/Box";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import Alert from "@mui/material/Alert";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Stack from "@mui/material/Stack";
import LoadingButton from "@mui/lab/LoadingButton";
import LoginIcon from "@mui/icons-material/Login";
import { useMemo } from "react";
import { useAuth } from "../auth";
import { usersApi } from "../api/users";

interface LoginForm {
  userId: string;
  password: string;
}

export const Auth = () => {
  const { login } = useAuth();
  const [formData, setFormData] = useState<LoginForm>({
    userId: "",
    password: "",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const navigate = useNavigate();
  const location = useLocation();

  const redirectTo = useMemo(() => {
    const state = location.state as { from?: { pathname?: string } } | null;
    return state?.from?.pathname ?? "/account";
  }, [location.state]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    // Clear error when user starts typing
    if (error) {
      setError(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      // Basic validation
      if (!formData.userId.trim()) {
        throw new Error("User ID is required");
      }

      if (!formData.password.trim()) {
        throw new Error("Password is required");
      }

      // Call login API (POST /api/auth/login)
      const response = await usersApi.login(formData);

      // Update auth context
      login(response.data.user.userId);

      // Navigate to redirect destination
      navigate(redirectTo, { replace: true });
    } catch (err: any) {
      setError(err.response?.data?.message || err.message || "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ maxWidth: 420, mx: "auto", mt: 4 }}>
      <Typography variant="h4" gutterBottom>
        Sign In
      </Typography>

      <Typography color="text.secondary" sx={{ mb: 3 }}>
        Enter your credentials to access your account
      </Typography>

      <Card variant="outlined">
        <CardContent>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit}>
            <Stack spacing={2}>
              <TextField
                fullWidth
                label="User ID"
                name="userId"
                value={formData.userId}
                onChange={handleInputChange}
                disabled={loading}
                autoComplete="username"
                variant="outlined"
              />

              <TextField
                fullWidth
                label="Password"
                name="password"
                type="password"
                value={formData.password}
                onChange={handleInputChange}
                disabled={loading}
                autoComplete="current-password"
                variant="outlined"
              />

              <LoadingButton
                type="submit"
                variant="contained"
                loading={loading}
                startIcon={<LoginIcon />}
                fullWidth
                size="large"
              >
                Sign In
              </LoadingButton>
            </Stack>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};

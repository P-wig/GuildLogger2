import React, { useState } from "react";
import { useLocation, useNavigate } from "react-router";
import Box from "@mui/material/Box";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import Alert from "@mui/material/Alert";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import LoadingButton from "@mui/lab/LoadingButton";
import LoginIcon from "@mui/icons-material/Login";
import PersonAddIcon from "@mui/icons-material/PersonAdd";
import { useMemo } from "react";
import { useAuth } from "../auth";
import { usersApi } from "../api/users";

interface AuthForm {
  userId: string;
  password: string;
  confirmPassword?: string; // Only for register mode
}

type AuthMode = 'login' | 'register';

export const Auth = () => {
  const { login } = useAuth();
  const [mode, setMode] = useState<AuthMode>('login');
  const [formData, setFormData] = useState<AuthForm>({
    userId: "",
    password: "",
    confirmPassword: "",
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

      if (mode === 'register') {
        // Additional validation for registration
        if (formData.password !== formData.confirmPassword) {
          throw new Error("Passwords do not match");
        }
        
        if (formData.password.length < 6) {
          throw new Error("Password must be at least 6 characters");
        }

        // Call register API
        const response = await usersApi.register({
          userId: formData.userId,
          password: formData.password
        });

        // Auto-login after successful registration
        login(response.data.user.userId);
        navigate(redirectTo, { replace: true });

      } else {
        // Call login API
        const response = await usersApi.login({
          userId: formData.userId,
          password: formData.password
        });

        // Update auth context
        login(response.data.user.userId);
        navigate(redirectTo, { replace: true });
      }
    } catch (err: any) {
      setError(err.response?.data?.message || err.message || `${mode === 'login' ? 'Login' : 'Registration'} failed`);
    } finally {
      setLoading(false);
    }
  };

  const toggleMode = () => {
    setMode(mode === 'login' ? 'register' : 'login');
    setError(null);
    setFormData({
      userId: "",
      password: "",
      confirmPassword: "",
    });
  };

  return (
    <Box sx={{ maxWidth: 420, mx: "auto", mt: 4 }}>
      <Typography variant="h4" gutterBottom>
        {mode === 'login' ? 'Sign In' : 'Create Account'}
      </Typography>

      <Typography color="text.secondary" sx={{ mb: 3 }}>
        {mode === 'login' 
          ? 'Enter your credentials to access your account'
          : 'Create a new account to get started'
        }
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
                helperText={mode === 'register' ? "Choose a unique username" : ""}
              />

              <TextField
                fullWidth
                label="Password"
                name="password"
                type="password"
                value={formData.password}
                onChange={handleInputChange}
                disabled={loading}
                autoComplete={mode === 'login' ? "current-password" : "new-password"}
                variant="outlined"
                helperText={mode === 'register' ? "Minimum 6 characters" : ""}
              />

              {mode === 'register' && (
                <TextField
                  fullWidth
                  label="Confirm Password"
                  name="confirmPassword"
                  type="password"
                  value={formData.confirmPassword}
                  onChange={handleInputChange}
                  disabled={loading}
                  autoComplete="new-password"
                  variant="outlined"
                />
              )}

              <LoadingButton
                type="submit"
                variant="contained"
                loading={loading}
                startIcon={mode === 'login' ? <LoginIcon /> : <PersonAddIcon />}
                fullWidth
                size="large"
              >
                {mode === 'login' ? 'Sign In' : 'Create Account'}
              </LoadingButton>

              <Button
                variant="text"
                onClick={toggleMode}
                disabled={loading}
                fullWidth
              >
                {mode === 'login' 
                  ? "Don't have an account? Sign up"
                  : "Already have an account? Sign in"
                }
              </Button>
            </Stack>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};

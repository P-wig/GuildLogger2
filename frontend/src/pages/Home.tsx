import React, { useState } from 'react';
import { Link as RouterLink } from "react-router";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Alert from "@mui/material/Alert";
import LoadingButton from "@mui/lab/LoadingButton";
import LoginIcon from "@mui/icons-material/Login";
import { useAuth } from "../auth";
import { usersApi } from "../api/users";

interface LoginForm {
  userId: string;
  password: string;
}

export const Home = () => {
  const { isAuthenticated, user, login } = useAuth();
  const [formData, setFormData] = useState<LoginForm>({
    userId: '',
    password: ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
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
        throw new Error('User ID is required');
      }
      
      if (!formData.password.trim()) {
        throw new Error('Password is required');
      }

      // Call login API (POST /api/auth/login)
      const response = await usersApi.login(formData);
      
      // Update auth context
      if (login) {
        login(response.data.user.email || response.data.user.userId);
      }
      
      // Reset form
      setFormData({ userId: '', password: '' });
      
    } catch (err: any) {
      setError(err.response?.data?.message || err.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ mt: 0, pt: 0 }}>
      <Typography variant="h4" gutterBottom sx={{ mt: 0 }}>
        Home
      </Typography>

      <Typography color="text.secondary" sx={{ mb: 3 }}>
        Welcome to the Cloud Native Team Project.
      </Typography>

      <Stack spacing={3}>
        {/* Navigation Card - Only show if authenticated */}
        {isAuthenticated && (
          <Card variant="outlined">
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Navigation
              </Typography>
              <Typography color="text.secondary" sx={{ mb: 2 }}>
                Access your account and project features.
              </Typography>
            </CardContent>

            <CardActions>
              <Button variant="contained" component={RouterLink} to="/account">
                Go to account
              </Button>
              <Button variant="outlined" component={RouterLink} to="/projects">
                View Projects
              </Button>
            </CardActions>
          </Card>
        )}

        {/* Login Form Card - Only show if not authenticated */}
        {!isAuthenticated && (
          <Card variant="outlined">
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Sign In
              </Typography>
              
              <Typography color="text.secondary" sx={{ mb: 3 }}>
                Enter your credentials to access your account
              </Typography>

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
        )}
      </Stack>
    </Box>
  );
};

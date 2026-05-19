import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Alert from "@mui/material/Alert";
import LoadingButton from "@mui/lab/LoadingButton";
import { authApi } from "../api/auth";
import { useAuth } from "../auth";

export const Auth = () => {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const code = searchParams.get("code");
  const redirectUri = `${window.location.origin}/auth`;

  // Handle OAuth callback — exchange code automatically when Discord redirects back
  useEffect(() => {
    if (!code) return;

    const exchange = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await authApi.discordLogin({ code, redirectUri });
        localStorage.setItem("authToken", response.data.token);
        login(response.data.user);
        navigate("/app", { replace: true });
      } catch {
        setError("Login failed. Please try again.");
        setLoading(false);
      }
    };
    exchange();
  }, [code]);

  const handleDiscordSignIn = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await authApi.getDiscordAuthURL(redirectUri);
      window.location.href = response.data.url;
    } catch {
      setError("Failed to reach Discord. Please try again.");
      setLoading(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: "80vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <Card variant="outlined" sx={{ maxWidth: 400, width: "100%", p: 2 }}>
        <CardContent>
          <Typography variant="h5" gutterBottom fontWeight={600} align="center">
            Sign in to GuildLogger
          </Typography>
          <Typography color="text.secondary" align="center" sx={{ mb: 3 }}>
            {code
              ? "Completing sign-in…"
              : "Connect your Discord account to get started."}
          </Typography>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          {!code && (
            <LoadingButton
              loading={loading}
              onClick={handleDiscordSignIn}
              variant="contained"
              fullWidth
              size="large"
            >
              Continue with Discord
            </LoadingButton>
          )}

          {code && (
            <LoadingButton loading variant="contained" fullWidth size="large">
              Signing in…
            </LoadingButton>
          )}
        </CardContent>
      </Card>
    </Box>
  );
};

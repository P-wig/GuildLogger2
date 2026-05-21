import { useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Alert from "@mui/material/Alert";
import LoadingButton from "@mui/lab/LoadingButton";
import { authApi } from "../api/auth";
import { useAuth } from "../auth";

const OAUTH_STATE_KEY = "oauth_state";

const DISCORD_ERROR_MESSAGES: Record<string, string> = {
  access_denied: "You cancelled the Discord login. Click below to try again.",
  invalid_scope: "The requested Discord permissions are invalid.",
};

const DEFAULT_ERROR = "Login failed. Please try again.";

export const Auth = () => {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Stable ref — derived from window.location.origin which never changes
  const redirectUri = useRef(`${window.location.origin}/auth`).current;

  const code = searchParams.get("code");
  const discordError = searchParams.get("error");

  // Handle Discord denial/error before attempting exchange
  useEffect(() => {
    if (!discordError) return;
    const message = DISCORD_ERROR_MESSAGES[discordError] ?? DEFAULT_ERROR;
    setError(message);
  }, [discordError]);

  // Handle OAuth callback — verify state then exchange code
  useEffect(() => {
    if (!code) return;

    const exchange = async () => {
      setLoading(true);
      setError(null);

      // CSRF check — state returned by Discord must match what we stored
      const returnedState = searchParams.get("state");
      const storedState = sessionStorage.getItem(OAUTH_STATE_KEY);
      sessionStorage.removeItem(OAUTH_STATE_KEY);

      if (!returnedState || returnedState !== storedState) {
        setError("Login failed: state mismatch. Please try again.");
        setLoading(false);
        return;
      }

      try {
        const response = await authApi.discordLogin({ code, redirectUri });
        localStorage.setItem("authToken", response.data.token);
        login(response.data.user);
        navigate("/app", { replace: true });
      } catch {
        setError(DEFAULT_ERROR);
        setLoading(false);
      }
    };
    exchange();
  }, [code, redirectUri, login, navigate, searchParams]);

  const handleDiscordSignIn = async () => {
    setLoading(true);
    setError(null);
    try {
      const state = crypto.randomUUID();
      sessionStorage.setItem(OAUTH_STATE_KEY, state);
      const response = await authApi.getDiscordAuthURL(redirectUri, state);
      window.location.href = response.data.url;
    } catch {
      sessionStorage.removeItem(OAUTH_STATE_KEY);
      setError("Failed to reach Discord. Please try again.");
      setLoading(false);
    }
  };

  const showButton = !code && !discordError;

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

          {showButton && (
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

          {(discordError || (!code && error)) && (
            <LoadingButton
              loading={loading}
              onClick={handleDiscordSignIn}
              variant="outlined"
              fullWidth
              size="large"
              sx={{ mt: discordError ? 1 : 0 }}
            >
              Try Again
            </LoadingButton>
          )}

          {code && !discordError && (
            <LoadingButton loading variant="contained" fullWidth size="large">
              Signing in…
            </LoadingButton>
          )}
        </CardContent>
      </Card>
    </Box>
  );
};

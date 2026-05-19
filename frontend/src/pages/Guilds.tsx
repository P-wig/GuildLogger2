import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

export const Guilds = () => {
  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        My Guilds
      </Typography>
      <Typography color="text.secondary">
        Your connected Discord guilds will appear here.
      </Typography>
    </Box>
  );
};
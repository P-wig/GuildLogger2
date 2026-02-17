import type { ReactNode } from "react";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";

type FormDialogProps = {
  open: boolean;
  onClose: () => void;
  onSubmit: () => void;
  title: string;
  submitLabel: string;
  loading: boolean;
  error: string | null;
  children: ReactNode;
};

export const FormDialog = ({
  open,
  onClose,
  onSubmit,
  title,
  submitLabel,
  loading,
  error,
  children,
}: FormDialogProps) => {
  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>{title}</DialogTitle>
      <DialogContent
        sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 1 }}
      >
        {error && <Alert severity="error">{error}</Alert>}
        {children}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={loading}>
          Cancel
        </Button>
        <Button variant="contained" onClick={onSubmit} disabled={loading}>
          {loading ? "Loading..." : submitLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

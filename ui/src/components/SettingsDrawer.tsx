import { useEffect, useState } from "react";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import Divider from "@mui/material/Divider";
import Drawer from "@mui/material/Drawer";
import IconButton from "@mui/material/IconButton";
import InputAdornment from "@mui/material/InputAdornment";
import Switch from "@mui/material/Switch";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import CloseIcon from "@mui/icons-material/Close";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import { ServiceToggle } from "./ServiceToggle";
import type { Config, SecretsInfo, ExtensionInfo, ServiceConfig } from "../types";
import { SERVICE_INFO } from "../types";

interface SettingsDrawerProps {
  open: boolean;
  onClose: () => void;
  config: Config;
  secrets: SecretsInfo | null;
  info: ExtensionInfo | null;
  onSave: (config: Config) => Promise<void>;
  onReset: () => void;
}

export function SettingsDrawer({
  open,
  onClose,
  config,
  secrets,
  info,
  onSave,
  onReset,
}: SettingsDrawerProps) {
  const [draft, setDraft] = useState<Config>(config);
  const [saving, setSaving] = useState(false);
  const [confirmResetOpen, setConfirmResetOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (open) {
      setDraft(config);
    }
  }, [open, config]);

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(draft);
      onClose();
    } catch {
      // Error handled by parent
    } finally {
      setSaving(false);
    }
  };

  const handleCopyPassword = () => {
    if (secrets?.adminPassword) {
      navigator.clipboard.writeText(secrets.adminPassword);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleServiceToggle = (key: keyof ServiceConfig, enabled: boolean) => {
    setDraft((prev) => ({
      ...prev,
      services: { ...prev.services, [key]: enabled },
    }));
  };

  const lanAccessible = draft.bindAddress === "0.0.0.0";

  return (
    <>
      <Drawer
        anchor="right"
        open={open}
        onClose={onClose}
        PaperProps={{ sx: { width: 360 } }}
      >
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            height: "100%",
          }}
        >
          {/* Header */}
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              px: 2,
              py: 1.5,
            }}
          >
            <Typography variant="h6">Settings</Typography>
            <IconButton size="small" onClick={onClose} aria-label="Close settings">
              <CloseIcon fontSize="small" />
            </IconButton>
          </Box>
          <Divider />

          {/* Scrollable content */}
          <Box sx={{ flex: 1, overflow: "auto", px: 2, py: 2 }}>
            {/* Network section */}
            <Typography
              variant="overline"
              color="text.secondary"
              sx={{ mb: 1, display: "block" }}
            >
              Network
            </Typography>

            <TextField
              label="Port"
              type="number"
              size="small"
              fullWidth
              value={draft.port}
              onChange={(e) =>
                setDraft((prev) => ({
                  ...prev,
                  port: parseInt(e.target.value, 10) || 8080,
                }))
              }
              sx={{ mb: 2 }}
              inputProps={{ min: 1, max: 65535 }}
            />

            <Box
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                mb: 1,
              }}
            >
              <Box>
                <Typography variant="body2" sx={{ fontWeight: 600 }}>
                  LAN accessible
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Allow connections from other devices on the network
                </Typography>
              </Box>
              <Switch
                size="small"
                checked={lanAccessible}
                onChange={(_, checked) =>
                  setDraft((prev) => ({
                    ...prev,
                    bindAddress: checked ? "0.0.0.0" : "127.0.0.1",
                  }))
                }
              />
            </Box>

            <Box
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                mb: 2,
              }}
            >
              <Box>
                <Typography variant="body2" sx={{ fontWeight: 600 }}>
                  Expose PostgreSQL
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Make the database port accessible for external tools
                </Typography>
              </Box>
              <Switch
                size="small"
                checked={draft.exposePostgres}
                onChange={(_, checked) =>
                  setDraft((prev) => ({ ...prev, exposePostgres: checked }))
                }
              />
            </Box>

            <Divider sx={{ mb: 2 }} />

            {/* Services section */}
            <Typography
              variant="overline"
              color="text.secondary"
              sx={{ mb: 1, display: "block" }}
            >
              Services
            </Typography>

            {(
              Object.keys(SERVICE_INFO) as Array<keyof ServiceConfig>
            ).map((key) => (
              <ServiceToggle
                key={key}
                label={SERVICE_INFO[key].label}
                description={SERVICE_INFO[key].description}
                ram={SERVICE_INFO[key].ram}
                warning={SERVICE_INFO[key].warning}
                enabled={draft.services[key]}
                onChange={(enabled) => handleServiceToggle(key, enabled)}
              />
            ))}

            <Box sx={{ mt: 2 }} />
            <Divider sx={{ mb: 2 }} />

            {/* Info section */}
            <Typography
              variant="overline"
              color="text.secondary"
              sx={{ mb: 1, display: "block" }}
            >
              Info
            </Typography>

            <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
              <Box
                sx={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <Typography variant="caption" color="text.secondary">
                  Extension version
                </Typography>
                <Typography variant="caption">
                  {info?.extensionVersion ?? "..."}
                </Typography>
              </Box>
              <Box
                sx={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <Typography variant="caption" color="text.secondary">
                  Backend version
                </Typography>
                <Typography variant="caption">
                  {info?.backendVersion ?? "..."}
                </Typography>
              </Box>
              {secrets?.adminPassword && (
                <Box>
                  <Typography
                    variant="caption"
                    color="text.secondary"
                    sx={{ display: "block", mb: 0.5 }}
                  >
                    Admin password
                  </Typography>
                  <TextField
                    size="small"
                    fullWidth
                    value={secrets.adminPassword}
                    InputProps={{
                      readOnly: true,
                      sx: { fontFamily: "monospace", fontSize: 12 },
                      endAdornment: (
                        <InputAdornment position="end">
                          <IconButton
                            size="small"
                            onClick={handleCopyPassword}
                            aria-label="Copy admin password"
                          >
                            <ContentCopyIcon fontSize="small" />
                          </IconButton>
                        </InputAdornment>
                      ),
                    }}
                  />
                  {copied && (
                    <Typography
                      variant="caption"
                      color="success.main"
                      sx={{ mt: 0.5, display: "block" }}
                    >
                      Copied to clipboard
                    </Typography>
                  )}
                </Box>
              )}
            </Box>
          </Box>

          {/* Actions */}
          <Divider />
          <Box
            sx={{
              px: 2,
              py: 1.5,
              display: "flex",
              flexDirection: "column",
              gap: 1,
            }}
          >
            <Button
              variant="contained"
              fullWidth
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? "Saving..." : "Save and Restart"}
            </Button>
            <Button
              variant="outlined"
              color="error"
              fullWidth
              onClick={() => setConfirmResetOpen(true)}
            >
              Reset All Data
            </Button>
          </Box>
        </Box>
      </Drawer>

      {/* Reset confirmation dialog */}
      <Dialog
        open={confirmResetOpen}
        onClose={() => setConfirmResetOpen(false)}
      >
        <DialogTitle>Reset all data?</DialogTitle>
        <DialogContent>
          <DialogContentText>
            This will delete all artifacts, repositories, users, and
            configuration. The extension will be restored to its initial state.
            This action cannot be undone.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmResetOpen(false)}>Cancel</Button>
          <Button
            color="error"
            variant="contained"
            onClick={() => {
              setConfirmResetOpen(false);
              onReset();
            }}
          >
            Reset
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}

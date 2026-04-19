import { useCallback, useState } from "react";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import { createDockerDesktopClient } from "@docker/extension-api-client";
import { TopBar } from "./components/TopBar";
import { LoadingScreen } from "./components/LoadingScreen";
import { SettingsDrawer } from "./components/SettingsDrawer";
import { useHealth } from "./hooks/useHealth";
import { useConfig } from "./hooks/useConfig";
import type { Config } from "./types";

const WEB_URL = "http://localhost:3000";

const ddClient = createDockerDesktopClient();

export function App() {
  const { health, loading: healthLoading, refresh: refreshHealth } = useHealth();
  const { config, secrets, info, loading: configLoading, saveConfig, refresh: refreshConfig } = useConfig();
  const [settingsOpen, setSettingsOpen] = useState(false);

  const isHealthy = health?.overall === "healthy";
  const loading = healthLoading || configLoading;

  const handleSave = useCallback(
    async (newConfig: Config) => {
      await saveConfig(newConfig);
      refreshHealth();
      refreshConfig();
    },
    [saveConfig, refreshHealth, refreshConfig]
  );

  const handleReset = useCallback(async () => {
    try {
      await ddClient.extension.vm?.service?.post("/reset", {});
      refreshHealth();
      refreshConfig();
    } catch (err) {
      console.error("Failed to reset:", err);
    }
  }, [refreshHealth, refreshConfig]);

  const loadingMessage = loading
    ? "Starting Artifact Keeper..."
    : health === null
      ? "Waiting for backend to respond..."
      : `Services starting (${health.overall})...`;

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
        height: "100vh",
        width: "100vw",
        overflow: "hidden",
      }}
    >
      <TopBar
        services={health?.services ?? []}
        onSettingsClick={() => setSettingsOpen(true)}
      />

      <Box sx={{ flex: 1, position: "relative" }}>
        {!isHealthy || loading ? (
          <LoadingScreen message={loadingMessage} />
        ) : (
          <Stack
            spacing={3}
            alignItems="center"
            justifyContent="center"
            sx={{ height: "100%", px: 4, textAlign: "center" }}
          >
            <Typography variant="h4">Artifact Keeper is ready</Typography>
            <Typography variant="body1" color="text.secondary">
              The web UI runs at {WEB_URL}. Open it in your browser to sign in and manage repositories.
            </Typography>
            <Button
              variant="contained"
              size="large"
              startIcon={<OpenInNewIcon />}
              onClick={() => ddClient.host.openExternal(WEB_URL)}
            >
              Open Artifact Keeper
            </Button>
            <Stack spacing={0.5} alignItems="center">
              <Typography variant="caption" color="text.secondary">
                Initial credentials (please change on first login):
              </Typography>
              <Typography variant="body2" sx={{ fontFamily: "monospace" }}>
                admin / {secrets?.adminPassword ?? "…"}
              </Typography>
            </Stack>
          </Stack>
        )}
      </Box>

      <SettingsDrawer
        open={settingsOpen}
        onClose={() => setSettingsOpen(false)}
        config={config}
        secrets={secrets}
        info={info}
        onSave={handleSave}
        onReset={handleReset}
      />
    </Box>
  );
}

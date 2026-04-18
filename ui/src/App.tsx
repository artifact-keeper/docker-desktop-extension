import { useCallback, useState } from "react";
import Box from "@mui/material/Box";
import { createDockerDesktopClient } from "@docker/extension-api-client";
import { TopBar } from "./components/TopBar";
import { LoadingScreen } from "./components/LoadingScreen";
import { SettingsDrawer } from "./components/SettingsDrawer";
import { useHealth } from "./hooks/useHealth";
import { useConfig } from "./hooks/useConfig";
import type { Config } from "./types";

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
          <iframe
            src={`http://localhost:${config.port}`}
            title="Artifact Keeper"
            style={{
              position: "absolute",
              top: 0,
              left: 0,
              width: "100%",
              height: "100%",
              border: "none",
            }}
          />
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

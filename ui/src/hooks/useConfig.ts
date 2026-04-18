import { useCallback, useEffect, useState } from "react";
import { createDockerDesktopClient } from "@docker/extension-api-client";
import type { Config, SecretsInfo, ExtensionInfo } from "../types";
import { DEFAULT_CONFIG } from "../types";

const ddClient = createDockerDesktopClient();

export function useConfig() {
  const [config, setConfig] = useState<Config>(DEFAULT_CONFIG);
  const [secrets, setSecrets] = useState<SecretsInfo | null>(null);
  const [info, setInfo] = useState<ExtensionInfo | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      const [configRes, secretsRes, infoRes] = await Promise.all([
        ddClient.extension.vm?.service?.get("/config"),
        ddClient.extension.vm?.service?.get("/secrets"),
        ddClient.extension.vm?.service?.get("/info"),
      ]);
      if (configRes) setConfig(configRes as Config);
      if (secretsRes) setSecrets(secretsRes as SecretsInfo);
      if (infoRes) setInfo(infoRes as ExtensionInfo);
    } catch {
      // Keep defaults on failure
    } finally {
      setLoading(false);
    }
  }, []);

  const saveConfig = useCallback(async (newConfig: Config) => {
    try {
      await ddClient.extension.vm?.service?.put("/config", newConfig);
      setConfig(newConfig);
    } catch (err) {
      console.error("Failed to save config:", err);
      throw err;
    }
  }, []);

  const refresh = useCallback(() => {
    fetchAll();
  }, [fetchAll]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  return { config, secrets, info, loading, saveConfig, refresh };
}

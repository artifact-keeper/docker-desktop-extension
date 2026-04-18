import { useCallback, useEffect, useRef, useState } from "react";
import { createDockerDesktopClient } from "@docker/extension-api-client";
import type { HealthReport } from "../types";

const ddClient = createDockerDesktopClient();

export function useHealth(pollInterval = 10_000) {
  const [health, setHealth] = useState<HealthReport | null>(null);
  const [loading, setLoading] = useState(true);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchHealth = useCallback(async () => {
    try {
      const result = await ddClient.extension.vm?.service?.get("/health");
      setHealth(result as HealthReport);
    } catch {
      setHealth(null);
    } finally {
      setLoading(false);
    }
  }, []);

  const refresh = useCallback(() => {
    setLoading(true);
    fetchHealth();
  }, [fetchHealth]);

  useEffect(() => {
    fetchHealth();
    timerRef.current = setInterval(fetchHealth, pollInterval);
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, [fetchHealth, pollInterval]);

  return { health, loading, refresh };
}

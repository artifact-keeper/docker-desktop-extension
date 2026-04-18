import Box from "@mui/material/Box";
import Tooltip from "@mui/material/Tooltip";
import type { ServiceHealth } from "../types";

const STATUS_COLORS: Record<ServiceHealth["status"], string> = {
  healthy: "#4caf50",
  starting: "#ff9800",
  unhealthy: "#f44336",
  not_running: "#9e9e9e",
};

interface HealthDotsProps {
  services: ServiceHealth[];
}

export function HealthDots({ services }: HealthDotsProps) {
  return (
    <Box sx={{ display: "flex", gap: 0.75, alignItems: "center" }}>
      {services.map((svc) => (
        <Tooltip key={svc.name} title={`${svc.name}: ${svc.status}`} arrow>
          <Box
            sx={{
              width: 10,
              height: 10,
              borderRadius: "50%",
              backgroundColor: STATUS_COLORS[svc.status],
              flexShrink: 0,
            }}
          />
        </Tooltip>
      ))}
    </Box>
  );
}

import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import IconButton from "@mui/material/IconButton";
import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import SettingsIcon from "@mui/icons-material/Settings";
import { HealthDots } from "./HealthDots";
import type { ServiceHealth, ExtensionInfo } from "../types";

interface TopBarProps {
  services: ServiceHealth[];
  info: ExtensionInfo | null;
  onSettingsClick: () => void;
}

export function TopBar({ services, info, onSettingsClick }: TopBarProps) {
  return (
    <AppBar position="static" color="default" elevation={1}>
      <Toolbar variant="dense" sx={{ minHeight: 40 }}>
        <Typography variant="subtitle1" sx={{ fontWeight: 600, mr: 1 }}>
          Artifact Keeper
        </Typography>
        {info?.backendVersion && info.backendVersion !== "unknown" && (
          <Chip label={`backend v${info.backendVersion}`} size="small" variant="outlined" sx={{ fontSize: 10, height: 20, mr: 0.5 }} />
        )}
        {info?.webVersion && info.webVersion !== "unknown" && (
          <Chip label={`web v${info.webVersion}`} size="small" variant="outlined" sx={{ fontSize: 10, height: 20 }} />
        )}
        <Box sx={{ flexGrow: 1, display: "flex", justifyContent: "center" }}>
          <HealthDots services={services} />
        </Box>
        <IconButton
          size="small"
          onClick={onSettingsClick}
          aria-label="Settings"
        >
          <SettingsIcon fontSize="small" />
        </IconButton>
      </Toolbar>
    </AppBar>
  );
}

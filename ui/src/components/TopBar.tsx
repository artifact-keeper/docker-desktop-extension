import React from "react";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import IconButton from "@mui/material/IconButton";
import Box from "@mui/material/Box";
import SettingsIcon from "@mui/icons-material/Settings";
import { HealthDots } from "./HealthDots";
import type { ServiceHealth } from "../types";

interface TopBarProps {
  services: ServiceHealth[];
  onSettingsClick: () => void;
}

export function TopBar({ services, onSettingsClick }: TopBarProps) {
  return (
    <AppBar position="static" color="default" elevation={1}>
      <Toolbar variant="dense" sx={{ minHeight: 40 }}>
        <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
          Artifact Keeper
        </Typography>
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

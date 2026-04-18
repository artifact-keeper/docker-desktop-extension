import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import Divider from "@mui/material/Divider";
import Switch from "@mui/material/Switch";
import Typography from "@mui/material/Typography";
import WarningAmberIcon from "@mui/icons-material/WarningAmber";

interface ServiceToggleProps {
  label: string;
  description: string;
  ram: string;
  warning?: string;
  enabled: boolean;
  onChange: (enabled: boolean) => void;
}

export function ServiceToggle({
  label,
  description,
  ram,
  warning,
  enabled,
  onChange,
}: ServiceToggleProps) {
  return (
    <>
      <Box
        sx={{
          display: "flex",
          alignItems: "flex-start",
          justifyContent: "space-between",
          py: 1.5,
          px: 0,
        }}
      >
        <Box sx={{ flex: 1, mr: 2 }}>
          <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 0.5 }}>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {label}
            </Typography>
            <Chip label={ram} size="small" variant="outlined" />
            {warning && (
              <WarningAmberIcon
                fontSize="small"
                sx={{ color: "warning.main" }}
              />
            )}
          </Box>
          <Typography variant="caption" color="text.secondary">
            {description}
          </Typography>
          {enabled && warning && (
            <Typography
              variant="caption"
              color="warning.main"
              sx={{ display: "block", mt: 0.5 }}
            >
              {warning}
            </Typography>
          )}
        </Box>
        <Switch
          checked={enabled}
          onChange={(_, checked) => onChange(checked)}
          size="small"
        />
      </Box>
      <Divider />
    </>
  );
}

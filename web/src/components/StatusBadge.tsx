import type { HealthStatus, AppStatus, ProtectionStatus } from "@/api/types";

interface StatusBadgeProps {
  status: HealthStatus | AppStatus | ProtectionStatus;
  size?: "sm" | "md";
}

const statusConfig: Record<
  string,
  { label: string; color: string; bg: string }
> = {
  healthy: {
    label: "Healthy",
    color: "var(--color-status-healthy)",
    bg: "var(--color-status-healthy-bg)",
  },
  running: {
    label: "Running",
    color: "var(--color-status-healthy)",
    bg: "var(--color-status-healthy-bg)",
  },
  armed: {
    label: "Armed",
    color: "var(--color-status-healthy)",
    bg: "var(--color-status-healthy-bg)",
  },
  monitoring: {
    label: "Monitoring",
    color: "var(--color-status-info)",
    bg: "var(--color-status-info-bg)",
  },
  warning: {
    label: "Warning",
    color: "var(--color-status-warning)",
    bg: "var(--color-status-warning-bg)",
  },
  updating: {
    label: "Updating",
    color: "var(--color-status-info)",
    bg: "var(--color-status-info-bg)",
  },
  installing: {
    label: "Installing",
    color: "var(--color-status-info)",
    bg: "var(--color-status-info-bg)",
  },
  degraded: {
    label: "Degraded",
    color: "var(--color-status-warning)",
    bg: "var(--color-status-warning-bg)",
  },
  error: {
    label: "Error",
    color: "var(--color-status-error)",
    bg: "var(--color-status-error-bg)",
  },
  critical: {
    label: "Critical",
    color: "var(--color-status-error)",
    bg: "var(--color-status-error-bg)",
  },
  stopped: {
    label: "Stopped",
    color: "var(--color-status-neutral)",
    bg: "var(--color-status-neutral-bg)",
  },
  offline: {
    label: "Offline",
    color: "var(--color-status-neutral)",
    bg: "var(--color-status-neutral-bg)",
  },
  unknown: {
    label: "Unknown",
    color: "var(--color-status-neutral)",
    bg: "var(--color-status-neutral-bg)",
  },
};

export function StatusBadge({ status, size = "sm" }: StatusBadgeProps) {
  const config = statusConfig[status] ?? statusConfig.unknown;
  const isSm = size === "sm";

  return (
    <span
      role="status"
      aria-label={config.label}
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: "6px",
        padding: isSm ? "2px 8px" : "4px 12px",
        borderRadius: "var(--radius-full)",
        fontSize: isSm ? "var(--text-xs)" : "var(--text-sm)",
        fontWeight: "var(--weight-medium)" as unknown as number,
        color: config.color,
        backgroundColor: config.bg,
        lineHeight: "var(--leading-tight)",
        whiteSpace: "nowrap",
      }}
    >
      <span
        style={{
          width: isSm ? 6 : 8,
          height: isSm ? 6 : 8,
          borderRadius: "50%",
          backgroundColor: config.color,
          flexShrink: 0,
        }}
      />
      {config.label}
    </span>
  );
}

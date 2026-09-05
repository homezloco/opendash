import type { SourceTrust } from "@/api/types";

interface SourceTrustBadgeProps {
  trust?: SourceTrust;
}

const trustConfig: Record<
  SourceTrust,
  { label: string; color: string; bg: string }
> = {
  official: { label: "Official", color: "var(--color-status-healthy)", bg: "var(--color-status-healthy-bg)" },
  reviewed: { label: "Reviewed", color: "var(--color-status-info)", bg: "var(--color-status-info-bg)" },
  "user-trusted": { label: "User Trusted", color: "var(--color-status-info)", bg: "var(--color-status-info-bg)" },
  changed: { label: "Changed", color: "var(--color-status-warning)", bg: "var(--color-status-warning-bg)" },
  untrusted: { label: "Untrusted", color: "var(--color-status-error)", bg: "var(--color-status-error-bg)" },
};

export function SourceTrustBadge({ trust }: SourceTrustBadgeProps) {
  if (!trust) return null;
  const config = trustConfig[trust] ?? { label: trust, color: "var(--color-text-tertiary)", bg: "var(--color-bg-tertiary)" };
  return (
    <span
      role="status"
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: "6px",
        padding: "2px 8px",
        borderRadius: "var(--radius-full)",
        fontSize: "var(--text-xs)",
        fontWeight: "var(--weight-medium)" as unknown as number,
        color: config.color,
        backgroundColor: config.bg,
        lineHeight: "var(--leading-tight)",
        whiteSpace: "nowrap",
      }}
    >
      <span
        style={{
          width: 6,
          height: 6,
          borderRadius: "50%",
          backgroundColor: config.color,
          flexShrink: 0,
        }}
      />
      {config.label}
    </span>
  );
}

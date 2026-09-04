import { useEffect, useRef, useState } from "react";
import { api } from "@/api";
import type { InstalledApp } from "@/api/types";
import { StatusBadge } from "./StatusBadge";
import { IconClose, IconExternalLink } from "./Icons";
import styles from "./Drawer.module.css";

interface AppDetailDrawerProps {
  app: InstalledApp;
  onClose: () => void;
}

export function AppDetailDrawer({ app, onClose }: AppDetailDrawerProps) {
  const closeRef = useRef<HTMLButtonElement>(null);
  const [operation, setOperation] = useState<{ kind: string; status: string; error?: string } | null>(null);
  const [logs, setLogs] = useState<string | null>(null);
  const [showLogs, setShowLogs] = useState(false);
  const hasUpdate = app.version !== app.latestVersion;

  async function runAction(kind: string, fn: () => Promise<{ kind: string; status: string; error?: string }>) {
    setOperation({ kind, status: "pending" });
    try {
      const op = await fn();
      setOperation({ kind, status: op.status, error: op.error });
    } catch (e) {
      setOperation({ kind, status: "failed", error: e instanceof Error ? e.message : String(e) });
    }
  }

  async function fetchLogs() {
    setShowLogs(true);
    setLogs("Loading...");
    try {
      const res = await api.getAppLogs(app.id, 100);
      setLogs(res.output);
    } catch (e) {
      setLogs(e instanceof Error ? e.message : String(e));
    }
  }

  useEffect(() => {
    closeRef.current?.focus();

    function handleKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }

    const prev = document.activeElement as HTMLElement | null;
    document.addEventListener("keydown", handleKey);
    document.body.style.overflow = "hidden";

    return () => {
      document.removeEventListener("keydown", handleKey);
      document.body.style.overflow = "";
      prev?.focus();
    };
  }, [onClose]);

  return (
    <>
      <div
        className={styles.overlay}
        onClick={onClose}
        aria-hidden="true"
      />
      <div
        className={styles.drawer}
        role="dialog"
        aria-modal="true"
        aria-label={`${app.name} details`}
      >
        <div className={styles.header}>
          <div className={styles.headerIcon} aria-hidden="true">
            {app.icon}
          </div>
          <div className={styles.headerBody}>
            <div className={styles.headerTitle}>{app.name}</div>
            <div className={styles.headerMeta}>
              <StatusBadge status={app.status} size="md" />
              <StatusBadge status={app.health} size="md" />
            </div>
          </div>
          <button
            ref={closeRef}
            className={styles.closeButton}
            onClick={onClose}
            aria-label="Close detail panel"
          >
            <IconClose />
          </button>
        </div>

        <div className={styles.body}>
          {hasUpdate && (
            <div className={styles.updateBanner} role="alert">
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true">
                <path
                  d="M10 3L2 17h16L10 3z"
                  stroke="var(--color-status-warning)"
                  strokeWidth="1.5"
                  strokeLinejoin="round"
                />
                <path
                  d="M10 8v4M10 14.5v.5"
                  stroke="var(--color-status-warning)"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                />
              </svg>
              <span className={styles.updateBannerText}>
                Update available:{" "}
                <span className={styles.updateBannerVersion}>
                  v{app.version}
                </span>
                {" → "}
                <span className={styles.updateBannerVersion}>
                  v{app.latestVersion}
                </span>
              </span>
            </div>
          )}

          <section className={styles.section}>
            <h3 className={styles.sectionTitle}>Lifecycle</h3>
            <div style={{ display: "flex", gap: "var(--space-2)", flexWrap: "wrap" }}>
              <ActionButton label="Start" onClick={() => runAction("start", () => api.startApp(app.id))} disabled={app.status === "running"} />
              <ActionButton label="Stop" onClick={() => runAction("stop", () => api.stopApp(app.id))} disabled={app.status === "stopped"} />
              <ActionButton label="Restart" onClick={() => runAction("restart", () => api.restartApp(app.id))} />
              <ActionButton label="Logs" onClick={fetchLogs} />
            </div>
            {operation && (
              <div style={{ marginTop: "var(--space-2)" }}>
                {operation.kind}: {operation.status}
                {operation.error && <div className={styles.error}>{operation.error}</div>}
              </div>
            )}
          </section>

          {showLogs && (
            <section className={styles.section}>
              <h3 className={styles.sectionTitle}>Logs</h3>
              <pre className={styles.mono} style={{ whiteSpace: "pre-wrap", maxHeight: "300px", overflow: "auto", background: "var(--color-bg-secondary)", padding: "var(--space-3)", borderRadius: "var(--radius-md)" }}>
                {logs ?? "Loading..."}
              </pre>
            </section>
          )}

          <section className={styles.section}>
            <h3 className={styles.sectionTitle}>About</h3>
            <p className={styles.description}>{app.description}</p>
          </section>

          {app.endpoints.length > 0 && (
            <section className={styles.section}>
              <h3 className={styles.sectionTitle}>Endpoints</h3>
              <div className={styles.endpointList}>
                {app.endpoints.map((ep) => (
                  <div key={ep.url} className={styles.endpoint}>
                    <div>
                      <div className={styles.endpointLabel}>{ep.label}</div>
                      <a
                        href={ep.url.startsWith("http") ? ep.url : `https://${ep.url}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className={styles.endpointUrl}
                      >
                        {ep.url}
                        <IconExternalLink />
                      </a>
                    </div>
                    <span className={styles.endpointKind}>{ep.kind}</span>
                  </div>
                ))}
              </div>
            </section>
          )}

          <section className={styles.section}>
            <h3 className={styles.sectionTitle}>Services</h3>
            {app.services.map((svc) => (
              <div key={svc.name} className={styles.serviceRow}>
                <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)" }}>
                  <span className={styles.serviceName}>{svc.name}</span>
                  <StatusBadge status={svc.status} />
                </div>
                <div className={styles.serviceStats}>
                  <span>CPU {svc.cpuPercent}%</span>
                  <span>{svc.memoryMb} MB</span>
                </div>
              </div>
            ))}
          </section>

          <section className={styles.section}>
            <h3 className={styles.sectionTitle}>Storage</h3>
            {app.storage.map((vol) => {
              const pct = (vol.usedGb / vol.totalGb) * 100;
              const fillClass =
                pct > 90
                  ? styles.storageFillCritical
                  : pct > 75
                    ? styles.storageFillWarning
                    : "";
              return (
                <div key={vol.name} className={styles.storageRow}>
                  <div className={styles.storageHeader}>
                    <span className={styles.storageName}>{vol.name}</span>
                    <span className={styles.storageSize}>
                      {vol.usedGb} / {vol.totalGb} GB
                    </span>
                  </div>
                  <div
                    className={styles.storageBar}
                    role="progressbar"
                    aria-valuenow={Math.round(pct)}
                    aria-valuemin={0}
                    aria-valuemax={100}
                    aria-label={`${vol.name} storage usage`}
                  >
                    <div
                      className={`${styles.storageFill} ${fillClass}`}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <div className={styles.storagePath}>{vol.mountPath}</div>
                </div>
              );
            })}
          </section>
        </div>
      </div>
    </>
  );
}

function ActionButton({ label, onClick, disabled }: { label: string; onClick: () => void; disabled?: boolean }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      style={{
        padding: "var(--space-2) var(--space-4)",
        borderRadius: "var(--radius-md)",
        border: "1px solid var(--color-border-default)",
        background: "var(--color-bg-secondary)",
        color: "var(--color-text-primary)",
        cursor: disabled ? "not-allowed" : "pointer",
        opacity: disabled ? 0.6 : 1,
      }}
    >
      {label}
    </button>
  );
}

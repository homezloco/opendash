import type { InstalledApp } from "@/api/types";
import { StatusBadge } from "./StatusBadge";
import { IconChevronRight } from "./Icons";
import styles from "./AppCard.module.css";

interface AppCardProps {
  app: InstalledApp;
  onClick: (app: InstalledApp) => void;
}

export function AppCard({ app, onClick }: AppCardProps) {
  const hasUpdate = app.version !== app.latestVersion;

  return (
    <button
      className={styles.appCard}
      onClick={() => onClick(app)}
      aria-label={`${app.name} - ${app.status}${hasUpdate ? ", update available" : ""}`}
    >
      <div className={styles.icon} aria-hidden="true">
        {app.icon}
      </div>
      <div className={styles.body}>
        <div className={styles.name}>{app.name}</div>
        <div className={styles.meta}>
          <span className={styles.version}>v{app.version}</span>
          {hasUpdate && (
            <span
              className={styles.updateDot}
              title={`Update available: v${app.latestVersion}`}
              aria-label="Update available"
            />
          )}
        </div>
      </div>
      <div className={styles.trailing}>
        <StatusBadge status={app.status} />
        <span className={styles.chevron}>
          <IconChevronRight />
        </span>
      </div>
    </button>
  );
}

interface AppCardGridProps {
  apps: InstalledApp[];
  onSelect: (app: InstalledApp) => void;
}

export function AppCardGrid({ apps, onSelect }: AppCardGridProps) {
  return (
    <div className={styles.appGrid} role="list" aria-label="Installed applications">
      {apps.map((app) => (
        <div key={app.id} role="listitem">
          <AppCard app={app} onClick={onSelect} />
        </div>
      ))}
    </div>
  );
}

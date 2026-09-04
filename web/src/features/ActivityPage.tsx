import { api } from "@/api";
import type { ActivityEvent } from "@/api/types";
import { useAsync } from "@/hooks/use-async";
import { DataGuard } from "@/components/LoadingState";
import styles from "./ActivityPage.module.css";

function formatTimestamp(ts: string): string {
  const d = new Date(ts);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffMins = Math.floor(diffMs / 60_000);
  if (diffMins < 60) return `${diffMins}m ago`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  return d.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const dotClass: Record<ActivityEvent["severity"], string> = {
  info: styles.dotInfo,
  warning: styles.dotWarning,
  error: styles.dotError,
  success: styles.dotSuccess,
};

export function ActivityPage() {
  const { data: events, state, error } = useAsync(
    () => api.getActivity(),
    [],
  );

  return (
    <>
      <h1
        style={{
          fontSize: "var(--text-2xl)",
          fontWeight: "var(--weight-bold)" as unknown as number,
          color: "var(--color-text-primary)",
          marginBottom: "var(--space-6)",
          letterSpacing: "-0.02em",
        }}
      >
        Activity
      </h1>

      <DataGuard state={state} error={error}>
        <div className={styles.timeline} role="feed" aria-label="Activity feed">
          {events?.map((ev) => (
            <article key={ev.id} className={styles.event} aria-label={ev.title}>
              <span
                className={`${styles.dot} ${dotClass[ev.severity]}`}
                aria-hidden="true"
              />
              <div className={styles.eventBody}>
                <div className={styles.eventTitle}>{ev.title}</div>
                <div className={styles.eventDesc}>{ev.description}</div>
              </div>
              <div className={styles.eventMeta}>
                <span className={styles.eventType}>{ev.type}</span>
                <time className={styles.eventTime} dateTime={ev.timestamp}>
                  {formatTimestamp(ev.timestamp)}
                </time>
              </div>
            </article>
          ))}
        </div>
      </DataGuard>
    </>
  );
}

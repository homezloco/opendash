import { Link } from "react-router-dom";
import type { AttentionItem } from "@/api/types";
import { IconChevronRight } from "./Icons";
import styles from "./AttentionCard.module.css";

function formatRelativeTime(timestamp: string): string {
  const diff = Date.now() - new Date(timestamp).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function SeverityIcon({ severity }: { severity: AttentionItem["severity"] }) {
  if (severity === "critical") {
    return (
      <svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <circle cx="10" cy="10" r="8" stroke="currentColor" strokeWidth="1.5" />
        <path d="M10 6v5M10 13.5v.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      </svg>
    );
  }
  if (severity === "warning") {
    return (
      <svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path d="M10 3L2 17h16L10 3z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
        <path d="M10 8v4M10 14.5v.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      </svg>
    );
  }
  return (
    <svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <circle cx="10" cy="10" r="8" stroke="currentColor" strokeWidth="1.5" />
      <path d="M10 9v4M10 6.5v.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

interface AttentionCardProps {
  item: AttentionItem;
}

export function AttentionCard({ item }: AttentionCardProps) {
  return (
    <div
      className={`${styles.attention} ${styles[item.severity]}`}
      role="alert"
      aria-label={`${item.severity}: ${item.title}`}
    >
      <span className={styles.icon}>
        <SeverityIcon severity={item.severity} />
      </span>
      <div className={styles.body}>
        <div className={styles.title}>{item.title}</div>
        <div className={styles.description}>{item.description}</div>
        {item.actionLabel && item.actionRoute && (
          <Link to={item.actionRoute} className={styles.action}>
            {item.actionLabel}
            <IconChevronRight size={12} />
          </Link>
        )}
      </div>
      <span className={styles.timestamp}>
        {formatRelativeTime(item.timestamp)}
      </span>
    </div>
  );
}

interface AttentionListProps {
  items: AttentionItem[];
}

export function AttentionList({ items }: AttentionListProps) {
  const active = items.filter((i) => !i.dismissed);
  if (active.length === 0) return null;

  const criticalCount = active.filter((i) => i.severity === "critical").length;

  return (
    <section className={styles.list} aria-label="Attention items">
      <h2 className={styles.listTitle}>
        Needs Attention
        {criticalCount > 0 && (
          <span className={styles.badge} aria-label={`${criticalCount} critical`}>
            {criticalCount}
          </span>
        )}
      </h2>
      {active.map((item) => (
        <AttentionCard key={item.id} item={item} />
      ))}
    </section>
  );
}

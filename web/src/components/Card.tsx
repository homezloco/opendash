import type { ReactNode, HTMLAttributes } from "react";
import styles from "./Card.module.css";

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  interactive?: boolean;
}

export function Card({
  children,
  interactive = false,
  className = "",
  ...props
}: CardProps) {
  return (
    <div
      className={`${styles.card} ${interactive ? styles.cardInteractive : ""} ${className}`}
      tabIndex={interactive ? 0 : undefined}
      role={interactive ? "button" : undefined}
      onKeyDown={
        interactive
          ? (e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                e.currentTarget.click();
              }
            }
          : undefined
      }
      {...props}
    >
      {children}
    </div>
  );
}

export function CardHeader({ children }: { children: ReactNode }) {
  return <div className={styles.cardHeader}>{children}</div>;
}

export function CardTitle({ children }: { children: ReactNode }) {
  return <div className={styles.cardTitle}>{children}</div>;
}

export function CardValue({ children }: { children: ReactNode }) {
  return <div className={styles.cardValue}>{children}</div>;
}

export function CardSubtext({ children }: { children: ReactNode }) {
  return <div className={styles.cardSubtext}>{children}</div>;
}

export function CardGrid({ children }: { children: ReactNode }) {
  return <div className={styles.grid}>{children}</div>;
}

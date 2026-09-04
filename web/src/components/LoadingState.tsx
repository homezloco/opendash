import type { LoadingState as LoadingStateType } from "@/api/types";

interface LoadingProps {
  state: LoadingStateType;
  error: string | null;
  children: React.ReactNode;
}

export function DataGuard({ state, error, children }: LoadingProps) {
  if (state === "loading" || state === "idle") {
    return (
      <div
        role="status"
        aria-label="Loading"
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: "var(--space-12)",
        }}
      >
        <div
          style={{
            width: 32,
            height: 32,
            border: "3px solid var(--color-border-default)",
            borderTopColor: "var(--color-accent-primary)",
            borderRadius: "50%",
            animation: "spin 0.8s linear infinite",
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg) } }`}</style>
      </div>
    );
  }

  if (state === "error") {
    return (
      <div
        role="alert"
        style={{
          padding: "var(--space-8)",
          textAlign: "center",
          color: "var(--color-status-error)",
        }}
      >
        <div
          style={{
            fontSize: "var(--text-lg)",
            fontWeight: "var(--weight-semibold)" as unknown as number,
            marginBottom: "var(--space-2)",
          }}
        >
          Something went wrong
        </div>
        <div
          style={{
            fontSize: "var(--text-sm)",
            color: "var(--color-text-secondary)",
          }}
        >
          {error ?? "An unexpected error occurred"}
        </div>
      </div>
    );
  }

  return <>{children}</>;
}

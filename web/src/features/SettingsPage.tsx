export function SettingsPage() {
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
        Settings
      </h1>
      <div
        style={{
          padding: "var(--space-8)",
          textAlign: "center",
          color: "var(--color-text-tertiary)",
          fontSize: "var(--text-sm)",
        }}
      >
        Settings will be available in a future release.
      </div>
    </>
  );
}

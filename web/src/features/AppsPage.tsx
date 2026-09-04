import { useState, useCallback } from "react";
import { api } from "@/api";
import type { InstalledApp } from "@/api/types";
import { useAsync } from "@/hooks/use-async";
import { DataGuard } from "@/components/LoadingState";
import { AppCardGrid } from "@/components/AppCard";
import { AppDetailDrawer } from "@/components/AppDetailDrawer";

export function AppsPage() {
  const { data: apps, state, error } = useAsync(() => api.getApps(), []);
  const [selected, setSelected] = useState<InstalledApp | null>(null);

  const handleSelect = useCallback((app: InstalledApp) => {
    setSelected(app);
  }, []);

  const handleClose = useCallback(() => {
    setSelected(null);
  }, []);

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
        Apps
      </h1>

      <DataGuard state={state} error={error}>
        {apps && <AppCardGrid apps={apps} onSelect={handleSelect} />}
      </DataGuard>

      {selected && (
        <AppDetailDrawer app={selected} onClose={handleClose} />
      )}
    </>
  );
}

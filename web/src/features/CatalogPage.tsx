import { useState } from "react";
import { api } from "@/api";
import type { CatalogAppDetail } from "@/api/types";
import { useAsync } from "@/hooks/use-async";
import { DataGuard } from "@/components/LoadingState";
import { CatalogAppDetailDrawer } from "@/components/CatalogAppDetailDrawer";
import styles from "./CatalogPage.module.css";

export function CatalogPage() {
  const { data: catalog, state, error } = useAsync(() => api.getCatalog(), []);
  const [selected, setSelected] = useState<CatalogAppDetail | null>(null);

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
        Catalog
      </h1>

      <DataGuard state={state} error={error}>
        <div className={styles.grid}>
          {catalog?.map((app) => (
            <button
              key={app.id}
              className={styles.catalogCard}
              onClick={() => setSelected(app as CatalogAppDetail)}
              aria-label={`View details for ${app.name}`}
              style={{ textAlign: "left" }}
            >
              <div className={styles.cardTop}>
                <div className={styles.icon} aria-hidden="true">
                  {app.icon}
                </div>
                <div className={styles.cardInfo}>
                  <div className={styles.cardName}>{app.name}</div>
                  <div className={styles.cardCategory}>{app.category}</div>
                </div>
              </div>
              <div className={styles.cardDesc}>{app.description}</div>
              <div className={styles.cardFooter}>
                <span className={styles.cardVersion}>v{app.version}</span>
                {app.installed && (
                  <span className={styles.installedBadge}>Installed</span>
                )}
              </div>
            </button>
          ))}
        </div>
      </DataGuard>

      {selected && (
        <CatalogAppDetailDrawer app={selected} onClose={() => setSelected(null)} />
      )}
    </>
  );
}

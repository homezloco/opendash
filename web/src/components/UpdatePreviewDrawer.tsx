import { useEffect, useRef, useState } from "react";
import { api } from "@/api";
import { useAsync } from "@/hooks/use-async";
import { SourceTrustBadge } from "./SourceTrustBadge";
import { IconClose } from "./Icons";
import styles from "./Drawer.module.css";

interface UpdatePreviewDrawerProps {
  appId: string;
  onClose: () => void;
}

export function UpdatePreviewDrawer({ appId, onClose }: UpdatePreviewDrawerProps) {
  const closeRef = useRef<HTMLButtonElement>(null);
  const { data: preview, state, error } = useAsync(() => api.getUpdatePlan(appId), [appId]);
  const [confirmed, setConfirmed] = useState(false);
  const [acknowledged, setAcknowledged] = useState(false);
  const [operation, setOperation] = useState<{ id: string; status: string; error?: string } | null>(null);
  const [applyError, setApplyError] = useState("");

  useEffect(() => {
    if (!operation || !["pending", "running"].includes(operation.status)) return;
    const timer = window.setInterval(async () => {
      try { setOperation(await api.getOperation(operation.id)); } catch (e) { setApplyError(e instanceof Error ? e.message : "Operation polling failed"); }
    }, 1000);
    return () => window.clearInterval(timer);
  }, [operation]);

  async function apply() {
    if (!preview) return;
    setApplyError("");
    try {
      setOperation(await api.applyUpdate(appId, { commitSha: preview.newCommit, checksum: preview.newChecksum, confirmed, acknowledgedRisks: acknowledged }));
    } catch (e) { setApplyError(e instanceof Error ? e.message : "Update failed"); }
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
      <div className={styles.overlay} onClick={onClose} aria-hidden="true" />
      <div className={styles.drawer} role="dialog" aria-modal="true" aria-label="Update preview">
        <div className={styles.header}>
          <div className={styles.headerBody}>
            <div className={styles.headerTitle}>Update preview</div>
            <div className={styles.headerMeta}>
              v{preview?.currentVersion} → v{preview?.newVersion}
            </div>
          </div>
          <button ref={closeRef} className={styles.closeButton} onClick={onClose} aria-label="Close update preview">
            <IconClose />
          </button>
        </div>

        <div className={styles.body}>
          {state === "loading" && <div>Loading update preview...</div>}
          {error && <div role="alert" className={styles.error}>{error}</div>}

          {preview && (
            <>
              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Version & Trust</h3>
                <div className={styles.metaRow}>
                  <span className={styles.metaLabel}>Current commit</span>
                  <span className={styles.metaValue}>{preview.currentCommit.slice(0, 8)}</span>
                </div>
                <div className={styles.metaRow}>
                  <span className={styles.metaLabel}>New commit</span>
                  <span className={styles.metaValue}>{preview.newCommit.slice(0, 8)}</span>
                </div>
                <div className={styles.metaRow}>
                  <span className={styles.metaLabel}>Can update</span>
                  <span className={styles.metaValue}>{preview.canUpdate ? "Yes" : "No"}</span>
                </div>
                {preview.newManifest.trust && (
                  <div className={styles.metaRow}>
                    <span className={styles.metaLabel}>Trust</span>
                    <SourceTrustBadge trust={preview.newManifest.trust} />
                  </div>
                )}
              </section>

              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Changes</h3>
                <ul>
                  {preview.changes.map((c, i) => (
                    <li key={i} className={c.kind === "permission" ? styles.warning : undefined}>
                      [{c.kind}] {c.summary}
                    </li>
                  ))}
                </ul>
              </section>

              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Images</h3>
                <ul>
                  {preview.images.map((img) => (
                    <li key={img} className={styles.mono}>{img}</li>
                  ))}
                </ul>
                {preview.addedImages.length > 0 && (
                  <div className={styles.warning} style={{ marginTop: "var(--space-2)" }}>
                    Added: {preview.addedImages.join(", ")}
                  </div>
                )}
                {preview.removedImages.length > 0 && (
                  <div className={styles.warning} style={{ marginTop: "var(--space-2)" }}>
                    Removed: {preview.removedImages.join(", ")}
                  </div>
                )}
              </section>

              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Ports</h3>
                <ul>
                  {preview.ports.map((p, i) => (
                    <li key={i}>{p.label}: container {p.containerPort} {p.hostPort ? `→ host ${p.hostPort}` : "(dynamic)"}</li>
                  ))}
                </ul>
              </section>

              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Volumes</h3>
                <ul>
                  {preview.volumes.map((v) => (
                    <li key={v.name}>{v.name}: {v.hostPath} → {v.mountPath}</li>
                  ))}
                </ul>
              </section>

              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Permission Diff</h3>
                <ul>
                  {preview.permissionDiff.map((p, i) => (
                    <li key={i} className={p.state === "added" || p.state === "changed" ? styles.warning : undefined}>
                      [{p.state}] {p.kind} {p.required ? "(required)" : ""}
                    </li>
                  ))}
                </ul>
              </section>

              {preview.risks.length > 0 && (
                <section className={styles.section}>
                  <h3 className={styles.sectionTitle}>Risks</h3>
                  <ul>
                    {preview.risks.map((r, i) => (
                      <li key={i} className={r.severity === "critical" ? styles.error : styles.warning}>
                        [{r.severity}] {r.category}: {r.description}
                      </li>
                    ))}
                  </ul>
                </section>
              )}

              {preview.blockedReasons.length > 0 && (
                <section className={styles.section}>
                  <h3 className={styles.sectionTitle}>Blocked reasons</h3>
                  <ul>
                    {preview.blockedReasons.map((r, i) => (
                      <li key={i} className={styles.error}>{r}</li>
                    ))}
                  </ul>
                </section>
              )}

              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Migration</h3>
                <ul>
                  {preview.migrationDisclosures.map((d, i) => (
                    <li key={i}>{d}</li>
                  ))}
                </ul>
              </section>

              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Rollback</h3>
                <ul>
                  {preview.rollbackDisclosures.map((d, i) => (
                    <li key={i}>{d}</li>
                  ))}
                </ul>
              </section>
              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Confirm update</h3>
                <label><input type="checkbox" checked={confirmed} onChange={(e) => setConfirmed(e.target.checked)} /> I reviewed this exact commit and checksum and confirm the update.</label>
                {preview.requiresAcknowledgement && <label><input type="checkbox" checked={acknowledged} onChange={(e) => setAcknowledged(e.target.checked)} /> I acknowledge the permission, removed-volume, migration, and rollback disclosures above.</label>}
                <button type="button" onClick={apply} disabled={!preview.canUpdate || !confirmed || (preview.requiresAcknowledgement && !acknowledged) || Boolean(operation && ["pending", "running"].includes(operation.status))}>Apply reviewed update</button>
                {operation && <div role="status">Update operation: {operation.status}</div>}
                {(applyError || operation?.error) && <div role="alert" className={styles.error}>{applyError || operation?.error}</div>}
              </section>
            </>
          )}
        </div>
      </div>
    </>
  );
}

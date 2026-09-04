import { useEffect, useRef, useState } from "react";
import { api } from "@/api";
import type { CatalogAppDetail, InstallPlan, Operation } from "@/api/types";
import { useAsync } from "@/hooks/use-async";
import { IconClose } from "./Icons";
import styles from "./Drawer.module.css";

interface Props {
  app: CatalogAppDetail;
  onClose: () => void;
}

export function CatalogAppDetailDrawer({ app, onClose }: Props) {
  const closeRef = useRef<HTMLButtonElement>(null);
  const [confirmed, setConfirmed] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [installError, setInstallError] = useState<string | null>(null);
  const [operation, setOperation] = useState<Operation | null>(null);
  const { data: plan, state, error } = useAsync(() => api.getInstallPlan(app.id), [app.id]);

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

  useEffect(() => {
    if (!operation || operation.status === "completed" || operation.status === "failed") return;
    const timer = setInterval(() => {
      api.getOperation(operation.id).then((op) => setOperation(op));
    }, 1000);
    return () => clearInterval(timer);
  }, [operation]);

  const handleInstall = async () => {
    setInstalling(true);
    setInstallError(null);
    try {
      const op = await api.installApp(app.id);
      setOperation(op);
    } catch (e) {
      setInstallError(e instanceof Error ? e.message : "Install failed");
    } finally {
      setInstalling(false);
    }
  };

  const hasCriticalRisk = plan?.risks.some((r) => r.severity === "critical");

  return (
    <>
      <div className={styles.overlay} onClick={onClose} aria-hidden="true" />
      <div className={styles.drawer} role="dialog" aria-modal="true" aria-label={`${app.name} catalog details`}>
        <div className={styles.header}>
          <div className={styles.headerIcon} aria-hidden="true">
            {app.icon}
          </div>
          <div className={styles.headerBody}>
            <div className={styles.headerTitle}>{app.name}</div>
            <div className={styles.headerMeta}>v{app.version}</div>
          </div>
          <button ref={closeRef} className={styles.closeButton} onClick={onClose} aria-label="Close catalog details">
            <IconClose />
          </button>
        </div>

        <div className={styles.body}>
          <section className={styles.section}>
            <h3 className={styles.sectionTitle}>About</h3>
            <p className={styles.description}>{app.description}</p>
          </section>

          {state === "loading" && <div>Loading install preview...</div>}
          {error && <div role="alert" className={styles.error}>{error}</div>}

          {plan && <InstallPlanView plan={plan} />}

          {plan && (
            <section className={styles.section}>
              <h3 className={styles.sectionTitle}>Install</h3>
              {plan.conflicts.length > 0 && (
                <div role="alert" className={styles.error}>
                  This app cannot be installed because of conflicts. Review the conflicts above.
                </div>
              )}
              {plan.risks.length > 0 && (
                <div role="alert" className={hasCriticalRisk ? styles.error : styles.warning}>
                  This app requires elevated permissions or presents security risks. Review the risks above.
                </div>
              )}
              <label style={{ display: "flex", alignItems: "center", gap: "var(--space-2)", marginBottom: "var(--space-3)" }}>
                <input type="checkbox" checked={confirmed} onChange={(e) => setConfirmed(e.target.checked)} />
                <span>I have reviewed the install preview and want to install this app.</span>
              </label>
              <button
                disabled={!confirmed || plan.conflicts.length > 0 || installing}
                onClick={handleInstall}
                style={{ padding: "var(--space-2) var(--space-4)" }}
              >
                {installing ? "Installing..." : "Install"}
              </button>
              {installError && <div role="alert" className={styles.error}>{installError}</div>}
              {operation && (
                <div style={{ marginTop: "var(--space-3)" }}>
                  Operation: {operation.kind} — {operation.status}
                  {operation.error && <div className={styles.error}>{operation.error}</div>}
                </div>
              )}
            </section>
          )}
        </div>
      </div>
    </>
  );
}

function InstallPlanView({ plan }: { plan: InstallPlan }) {
  return (
    <>
      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Project</h3>
        <div className={styles.metaRow}>
          <span className={styles.metaLabel}>Project name</span>
          <span className={styles.metaValue}>{plan.projectName}</span>
        </div>
        <div className={styles.metaRow}>
          <span className={styles.metaLabel}>Path</span>
          <span className={styles.metaValue}>{plan.projectPath}</span>
        </div>
      </section>

      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Images</h3>
        <ul>
          {plan.images.map((img) => (
            <li key={img} className={styles.mono}>{img}</li>
          ))}
        </ul>
      </section>

      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Ports</h3>
        <ul>
          {plan.ports.map((p, i) => (
            <li key={i}>
              {p.label}: container {p.containerPort} {p.hostPort ? `→ host ${p.hostPort}` : "(dynamic host port)"}
            </li>
          ))}
        </ul>
      </section>

      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Volumes</h3>
        <ul>
          {plan.volumes.map((v) => (
            <li key={v.name}>
              {v.name}: {v.hostPath} → {v.mountPath}
            </li>
          ))}
        </ul>
      </section>

      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Environment</h3>
        <ul>
          {plan.environment.map((e) => (
            <li key={e.key}>
              {e.key}={e.secret ? "••••••••" : e.value}
            </li>
          ))}
        </ul>
      </section>

      {plan.risks.length > 0 && (
        <section className={styles.section}>
          <h3 className={styles.sectionTitle}>Privilege Risks</h3>
          <ul>
            {plan.risks.map((r, i) => (
              <li key={i} className={r.severity === "critical" ? styles.error : styles.warning}>
                [{r.severity}] {r.category}: {r.description}
              </li>
            ))}
          </ul>
        </section>
      )}

      {plan.conflicts.length > 0 && (
        <section className={styles.section}>
          <h3 className={styles.sectionTitle}>Conflicts</h3>
          <ul>
            {plan.conflicts.map((c, i) => (
              <li key={i} className={styles.error}>
                [{c.kind}] {c.target}: {c.description}
              </li>
            ))}
          </ul>
        </section>
      )}
    </>
  );
}

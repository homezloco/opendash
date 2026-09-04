import { api } from "@/api";
import { useAsync } from "@/hooks/use-async";
import { DataGuard } from "@/components/LoadingState";
import { StatusBadge } from "@/components/StatusBadge";
import styles from "./ProtectionPage.module.css";

const categoryIcons: Record<string, string> = {
  firewall: "FW",
  vpn: "VP",
  dns: "DN",
  ids: "ID",
  backup: "BK",
};

export function ProtectionPage() {
  const { data: protection, state, error } = useAsync(
    () => api.getProtection(),
    [],
  );

  return (
    <>
      <div className={styles.header}>
        <h1
          style={{
            fontSize: "var(--text-2xl)",
            fontWeight: "var(--weight-bold)" as unknown as number,
            color: "var(--color-text-primary)",
            letterSpacing: "-0.02em",
          }}
        >
          Protection
        </h1>
        {protection && <StatusBadge status={protection.status} size="md" />}
      </div>

      <DataGuard state={state} error={error}>
        {protection && (
          <>
            <div className={styles.statsGrid}>
              <div className={styles.statCard}>
                <div className={styles.statLabel}>Threats Blocked (24h)</div>
                <div className={styles.statValue}>
                  {protection.threatsBlocked24h.toLocaleString()}
                </div>
              </div>
              <div className={styles.statCard}>
                <div className={styles.statLabel}>Firewall</div>
                <div className={styles.statValue}>
                  {protection.firewallActive ? "Active" : "Inactive"}
                </div>
              </div>
              <div className={styles.statCard}>
                <div className={styles.statLabel}>VPN</div>
                <div className={styles.statValue}>
                  {protection.vpnConnected ? "Connected" : "Disconnected"}
                </div>
              </div>
              <div className={styles.statCard}>
                <div className={styles.statLabel}>DNS Filtering</div>
                <div className={styles.statValue}>
                  {protection.dnsFilteringActive ? "Active" : "Inactive"}
                </div>
              </div>
            </div>

            <h2 className={styles.rulesTitle}>Security Rules</h2>
            <div className={styles.ruleList} role="list" aria-label="Protection rules">
              {protection.rules.map((rule) => (
                <div key={rule.id} className={styles.ruleCard} role="listitem">
                  <div className={styles.ruleIcon} aria-hidden="true">
                    {categoryIcons[rule.category] ?? "??"}
                  </div>
                  <div className={styles.ruleBody}>
                    <div className={styles.ruleName}>{rule.name}</div>
                    <div className={styles.ruleDesc}>{rule.description}</div>
                  </div>
                  <div className={styles.ruleTrailing}>
                    <StatusBadge status={rule.status} />
                    <span className={styles.categoryLabel}>{rule.category}</span>
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </DataGuard>
    </>
  );
}

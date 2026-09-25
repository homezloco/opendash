import { useCallback, useState } from "react";
import { api } from "@/api";
import type { SystemSummary } from "@/api/types";
import { useAsync } from "@/hooks/use-async";
import { DataGuard } from "@/components/LoadingState";
import { Card, CardHeader, CardTitle, CardValue, CardSubtext, CardGrid } from "@/components/Card";
import { AttentionList } from "@/components/AttentionCard";
import { StatusBadge } from "@/components/StatusBadge";

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  return `${days}d ${hours}h`;
}

function SummaryCards({ summary }: { summary: SystemSummary }) {
  const storagePct = Math.round((summary.storageUsedGb / summary.storageTotalGb) * 100);

  return (
    <CardGrid>
      <Card>
        <CardHeader>
          <CardTitle>Apps</CardTitle>
        </CardHeader>
        <CardValue>
          {summary.appsRunning}/{summary.appsTotal}
        </CardValue>
        <CardSubtext>running</CardSubtext>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Protection</CardTitle>
        </CardHeader>
        <div style={{ marginTop: "var(--space-1)" }}>
          <StatusBadge status={summary.protectionStatus} size="md" />
        </div>
        <CardSubtext>all systems</CardSubtext>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Updates</CardTitle>
        </CardHeader>
        <CardValue>{summary.updatesAvailable}</CardValue>
        <CardSubtext>available</CardSubtext>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Storage</CardTitle>
        </CardHeader>
        <CardValue>{storagePct}%</CardValue>
        <CardSubtext>
          {summary.storageUsedGb} / {summary.storageTotalGb} GB
        </CardSubtext>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>CPU</CardTitle>
        </CardHeader>
        <CardValue>{summary.cpuPercent}%</CardValue>
        <CardSubtext>utilization</CardSubtext>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Memory</CardTitle>
        </CardHeader>
        <CardValue>{summary.memoryPercent}%</CardValue>
        <CardSubtext>used</CardSubtext>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Health</CardTitle>
        </CardHeader>
        <div style={{ marginTop: "var(--space-1)" }}>
          <StatusBadge status={summary.health} size="md" />
        </div>
        <CardSubtext>system status</CardSubtext>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Uptime</CardTitle>
        </CardHeader>
        <CardValue>{formatUptime(summary.uptimeSeconds)}</CardValue>
        <CardSubtext>since last reboot</CardSubtext>
      </Card>
    </CardGrid>
  );
}

export function OverviewPage() {
  const summary = useAsync<SystemSummary>(() => api.getSummary(), []);
  const attention = useAsync(() => api.getAttentionItems(), []);

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
        Overview
      </h1>

      <DataGuard state={attention.state} error={attention.error}>
        {attention.data && <AttentionList items={attention.data} />}
      </DataGuard>

      <DataGuard state={summary.state} error={summary.error}>
        {summary.data && <SummaryCards summary={summary.data} />}
      </DataGuard>
      <SupportBundleDownload />
    </>
  );
}

function SupportBundleDownload() {
  const [busy, setBusy] = useState(false);
  const download = useCallback(async () => {
    setBusy(true);
    try {
      const bundle = await api.getSupportBundle();
      const blob = new Blob([JSON.stringify(bundle, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `opendash-support-bundle-${new Date(bundle.generatedAt).toISOString()}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } finally {
      setBusy(false);
    }
  }, []);

  return (
    <section style={{ marginTop: "var(--space-6)" }}>
      <h2 style={{ fontSize: "var(--text-lg)", color: "var(--color-text-primary)", marginBottom: "var(--space-3)" }}>
        Diagnostics
      </h2>
      <p style={{ color: "var(--color-text-secondary)", marginBottom: "var(--space-3)" }}>
        Download a redacted support bundle with system state, app metadata, and
        recent activity. Secrets and configuration values are redacted.
      </p>
      <button onClick={download} disabled={busy} type="button">
        {busy ? "Generating..." : "Download support bundle"}
      </button>
    </section>
  );
}

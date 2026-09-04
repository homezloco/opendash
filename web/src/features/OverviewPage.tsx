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
    </>
  );
}

import { useCallback, useState } from "react";
import { api } from "@/api";
import type { SourceListItem, SourcePreview } from "@/api/types";
import { useAsync } from "@/hooks/use-async";
import { DataGuard } from "@/components/LoadingState";
import { SourceTrustBadge } from "@/components/SourceTrustBadge";

export function SourcesPage() {
  const { data: sources, state, error, refetch } = useAsync(() => api.getSources(), []);
  const [url, setUrl] = useState("");
  const [preview, setPreview] = useState<SourcePreview | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [previewing, setPreviewing] = useState(false);
  const [adding, setAdding] = useState(false);
  const [addError, setAddError] = useState<string | null>(null);
  const [confirmed, setConfirmed] = useState(false);

  const handlePreview = useCallback(async () => {
    setPreview(null);
    setPreviewError(null);
    setConfirmed(false);
    setPreviewing(true);
    try {
      const p = await api.previewSource(url);
      setPreview(p);
    } catch (e) {
      setPreviewError(e instanceof Error ? e.message : String(e));
    } finally {
      setPreviewing(false);
    }
  }, [url]);

  const handleAdd = useCallback(async () => {
    if (!confirmed) {
      setAddError("Confirm the preview before adding this source.");
      return;
    }
    setAddError(null);
    setAdding(true);
    try {
      await api.addSource(url, true);
      setPreview(null);
      setUrl("");
      setConfirmed(false);
      void refetch();
    } catch (e) {
      setAddError(e instanceof Error ? e.message : String(e));
    } finally {
      setAdding(false);
    }
  }, [url, confirmed, refetch]);

  const handleRemove = useCallback(async (id: string) => {
    try {
      await api.removeSource(id);
      void refetch();
    } catch (e) {
      alert(e instanceof Error ? e.message : String(e));
    }
  }, [refetch]);

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
        Sources
      </h1>

      <section
        style={{
          background: "var(--color-bg-secondary)",
          borderRadius: "var(--radius-md)",
          padding: "var(--space-4)",
          marginBottom: "var(--space-6)",
        }}
      >
        <h2 style={{ fontSize: "var(--text-lg)", marginBottom: "var(--space-3)" }}>
          Add a GitHub manifest source
        </h2>
        <div style={{ display: "flex", gap: "var(--space-2)", flexWrap: "wrap" }}>
          <input
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://github.com/owner/repo/blob/ref/path/app.json"
            style={{
              flex: 1,
              minWidth: "240px",
              padding: "var(--space-2)",
              borderRadius: "var(--radius-md)",
              border: "1px solid var(--color-border-default)",
              background: "var(--color-bg-primary)",
              color: "var(--color-text-primary)",
            }}
          />
          <button onClick={handlePreview} disabled={!url || previewing}>
            {previewing ? "Previewing..." : "Preview"}
          </button>
        </div>

        {previewError && (
          <div role="alert" style={{ color: "var(--color-status-error)", marginTop: "var(--space-3)" }}>
            {previewError}
          </div>
        )}

        {preview && (
          <div style={{ marginTop: "var(--space-4)", padding: "var(--space-3)", borderRadius: "var(--radius-md)", background: "var(--color-bg-primary)" }}>
            <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)", marginBottom: "var(--space-2)" }}>
              <strong>{preview.manifest.name}</strong>
              <span>v{preview.manifest.version}</span>
              <SourceTrustBadge trust={preview.trust} />
            </div>
            <div style={{ fontSize: "var(--text-sm)", color: "var(--color-text-tertiary)", marginBottom: "var(--space-2)" }}>
              Commit {preview.commitSha.slice(0, 8)} · Checksum {preview.checksum.slice(0, 12)}…
            </div>
            <label style={{ display: "flex", alignItems: "center", gap: "var(--space-2)", marginBottom: "var(--space-3)" }}>
              <input type="checkbox" checked={confirmed} onChange={(e) => setConfirmed(e.target.checked)} />
              <span>I reviewed the manifest and want to add this source.</span>
            </label>
            <button onClick={handleAdd} disabled={!confirmed || adding}>
              {adding ? "Adding..." : "Add source"}
            </button>
            {addError && <div role="alert" style={{ color: "var(--color-status-error)", marginTop: "var(--space-2)" }}>{addError}</div>}
          </div>
        )}
      </section>

      <DataGuard state={state} error={error}>
        {sources && (
          <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-3)" }}>
            {sources.map((src: SourceListItem) => (
              <div
                key={src.id}
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  padding: "var(--space-3)",
                  background: "var(--color-bg-secondary)",
                  borderRadius: "var(--radius-md)",
                }}
              >
                <div>
                  <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)", marginBottom: "var(--space-1)" }}>
                    <strong>{src.name}</strong>
                    <span style={{ color: "var(--color-text-tertiary)" }}>v{src.version}</span>
                    <SourceTrustBadge trust={src.trust} />
                  </div>
                  <div style={{ fontSize: "var(--text-sm)", color: "var(--color-text-tertiary)" }}>
                    {src.owner}/{src.repo} @ {src.ref} · {src.path}
                  </div>
                  <div style={{ fontSize: "var(--text-xs)", color: "var(--color-text-tertiary)", marginTop: "var(--space-1)" }}>
                    commit {src.commitSha.slice(0, 8)} · {src.sourceUrl}
                  </div>
                </div>
                <button onClick={() => handleRemove(src.id)}>Remove</button>
              </div>
            ))}
          </div>
        )}
      </DataGuard>
    </>
  );
}

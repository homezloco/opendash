import type { RestorePreview } from "@/api/types";

export function RestorePreviewDrawer({ preview, onClose }: { preview: RestorePreview | null; onClose: () => void }) {
  if (!preview) return null;
  return <div role="dialog" aria-modal="true" aria-label="Restore preview" style={{ position: "fixed", inset: 0, background: "rgba(0,0,0,.55)", padding: "10vh 10vw", zIndex: 20 }}>
    <section style={{ background: "var(--color-surface)", padding: 24, maxWidth: 720, margin: "auto" }}>
      <h2>Restore preview</h2>
      <p>{preview.manifest.name} {preview.manifest.version}</p>
      <p>Integrity: {preview.integrityOk ? "verified" : "mismatch"} · Source: {preview.sourceTrusted ? "trusted" : "untrusted"}</p>
      <h3>Volumes</h3><ul>{preview.volumes.map((v) => <li key={v}>{v}</li>)}</ul>
      <h3>Images</h3><ul>{preview.images.map((v) => <li key={v}>{v}</li>)}</ul>
      {preview.migrationRisk && <p><strong>Migration risk:</strong> {preview.migrationRisk}</p>}
      <button type="button" onClick={onClose}>Close</button>
    </section>
  </div>;
}

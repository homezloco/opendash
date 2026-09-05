import { useState } from "react";
import { api } from "@/api/client";
import type { RecoveryImportPreview, RecoveryInventory } from "@/api/types";

export function RecoveryPage() {
  const [preview, setPreview] = useState<RecoveryImportPreview | null>(null); const [error, setError] = useState("");
  async function load(file: File) { try { const inventory = JSON.parse(await file.text()) as RecoveryInventory; setPreview(await api.previewRecoveryImport(inventory)); } catch (e) { setError(e instanceof Error ? e.message : "Invalid inventory"); } }
  async function download() { try { const data = await api.exportRecovery(); const url = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: "application/json" })); const a = document.createElement("a"); a.href = url; a.download = "opendash-recovery.json"; a.click(); URL.revokeObjectURL(url); } catch (e) { setError(e instanceof Error ? e.message : "Export failed"); } }
  return <section><h1>Recovery</h1><p>Export app, source, and backup inventory or safely preview an import. Import does not execute source hooks.</p>{error && <p role="alert">{error}</p>}
    <button type="button" onClick={() => void download()}>Export inventory</button> <label>Preview inventory <input type="file" accept="application/json,.json" onChange={(e) => { const f = e.target.files?.[0]; if (f) void load(f); }} /></label>
    {preview && <div><h2>Import preview</h2><p>{preview.apps} apps, {preview.sources} sources, {preview.backups} backups</p>{preview.untrustedSources.length > 0 && <p>Untrusted sources: {preview.untrustedSources.join(", ")}</p>}<ul>{preview.conflicts.map((c) => <li key={`${c.appId}-${c.kind}`}>{c.message}</li>)}</ul></div>}</section>;
}

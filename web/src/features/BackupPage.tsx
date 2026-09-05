import { useEffect, useState } from "react";
import { api } from "@/api/client";
import type { BackupArchive, InstalledApp, RestorePreview } from "@/api/types";
import { RestorePreviewDrawer } from "@/components/RestorePreviewDrawer";

export function BackupPage() {
  const [apps, setApps] = useState<InstalledApp[]>([]); const [appId, setAppId] = useState("");
  const [backups, setBackups] = useState<BackupArchive[]>([]); const [preview, setPreview] = useState<RestorePreview | null>(null); const [error, setError] = useState("");
  useEffect(() => { void api.getApps().then((v) => { setApps(v); if (v[0]) setAppId(v[0].id); }).catch((e: unknown) => setError(e instanceof Error ? e.message : "Unable to load apps")); }, []);
  useEffect(() => { if (appId) void api.getBackups(appId).then(setBackups).catch((e: unknown) => setError(e instanceof Error ? e.message : "Unable to load backups")); }, [appId]);
  return <section><h1>Backups</h1><p>Create application-aware volume archives and inspect them before recovery.</p>
    {error && <p role="alert">{error}</p>}<label>Application <select value={appId} onChange={(e) => setAppId(e.target.value)}>{apps.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}</select></label>{" "}
    <button type="button" disabled={!appId} onClick={() => void api.createBackup(appId).then((v) => setBackups((old) => [v, ...old])).catch((e: unknown) => setError(e instanceof Error ? e.message : "Backup failed"))}>Create backup</button>
    <ul>{backups.map((b) => <li key={b.id}>{new Date(b.createdAt).toLocaleString()} · {(b.size / 1048576).toFixed(1)} MiB <button type="button" onClick={() => void api.previewRestore(b.id).then(setPreview)}>Preview restore</button></li>)}</ul>
    <RestorePreviewDrawer preview={preview} onClose={() => setPreview(null)} /></section>;
}

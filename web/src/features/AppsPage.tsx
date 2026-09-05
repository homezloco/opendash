import { useState, useCallback, useEffect, useMemo } from "react";
import { api } from "@/api";
import type { DashboardPreferences, InstalledApp } from "@/api/types";
import { useAsync } from "@/hooks/use-async";
import { DataGuard } from "@/components/LoadingState";
import { AppDetailDrawer } from "@/components/AppDetailDrawer";

const defaults: DashboardPreferences = { defaultView: "grid", tileDensity: "normal", tileSize: "medium", groupBy: "none", sortBy: "manual", tileOrder: [], favoriteAppIds: [], hiddenFields: { version: false, status: false, health: false, endpoints: false } };

export function AppsPage() {
  const { data: apps, state, error } = useAsync(() => api.getApps(), []);
  const [selected, setSelected] = useState<InstalledApp | null>(null);
  const [preferences, setPreferences] = useState(defaults);
  const [settings, setSettings] = useState(false);
  useEffect(() => { api.getDashboardPreferences().then((p) => setPreferences({ ...defaults, ...p, hiddenFields: { ...defaults.hiddenFields, ...p.hiddenFields } })).catch(() => undefined); }, []);
  const save = (next: DashboardPreferences) => { setPreferences(next); void api.updateDashboardPreferences(next); };
  const update = <K extends keyof DashboardPreferences>(key: K, value: DashboardPreferences[K]) => save({ ...preferences, [key]: value });
  const toggleFavorite = (id: string) => update("favoriteAppIds", preferences.favoriteAppIds.includes(id) ? preferences.favoriteAppIds.filter((item) => item !== id) : [...preferences.favoriteAppIds, id]);
  const moveTop = (id: string) => { update("tileOrder", [id, ...preferences.tileOrder.filter((item) => item !== id)]); };
  const displayed = useMemo(() => {
    const result = [...(apps ?? [])];
    const manual = new Map(preferences.tileOrder.map((id, index) => [id, index]));
    result.sort((a, b) => {
      if (preferences.sortBy === "manual") return (manual.get(a.id) ?? 1e9) - (manual.get(b.id) ?? 1e9) || a.name.localeCompare(b.name);
      return String(a[preferences.sortBy]).localeCompare(String(b[preferences.sortBy])) || a.name.localeCompare(b.name);
    });
    return result;
  }, [apps, preferences.sortBy, preferences.tileOrder]);
  const groups = useMemo(() => displayed.reduce<Record<string, InstalledApp[]>>((all, app) => { const key = preferences.groupBy === "none" ? "Applications" : String(app[preferences.groupBy]); (all[key] ??= []).push(app); return all; }, {}), [displayed, preferences.groupBy]);
  const handleSelect = useCallback((app: InstalledApp) => setSelected(app), []);

  return <>
    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "var(--space-6)" }}><h1 style={{ fontSize: "var(--text-2xl)", color: "var(--color-text-primary)" }}>Apps</h1><button onClick={() => setSettings(!settings)} aria-expanded={settings}>Dashboard settings</button></div>
    {settings && <section aria-label="Dashboard preferences" style={{ padding: 16, marginBottom: 20, border: "1px solid var(--color-border)", display: "grid", gap: 10 }}>
      {([ ["View", "defaultView", ["grid", "list"]], ["Density", "tileDensity", ["compact", "normal", "spacious"]], ["Tile size", "tileSize", ["small", "medium", "large", "wide"]], ["Group by", "groupBy", ["none", "category", "status", "health"]], ["Sort by", "sortBy", ["manual", "name", "status", "category"]] ] as const).map(([label, key, options]) => <label key={key}>{label} <select value={preferences[key]} onChange={(e) => update(key, e.target.value as never)}>{options.map((option) => <option key={option}>{option}</option>)}</select></label>)}
      <fieldset><legend>Hidden tile fields</legend>{(["version", "status", "health", "endpoints"] as const).map((field) => <label key={field} style={{ marginRight: 12 }}><input type="checkbox" checked={preferences.hiddenFields[field]} onChange={(e) => update("hiddenFields", { ...preferences.hiddenFields, [field]: e.target.checked })} /> {field}</label>)}</fieldset>
    </section>}
    <DataGuard state={state} error={error}>{apps && Object.entries(groups).map(([group, items]) => <section key={group}>{preferences.groupBy !== "none" && <h2>{group}</h2>}<div role="list" aria-label="Installed applications" style={{ display: "grid", gridTemplateColumns: preferences.defaultView === "list" ? "1fr" : `repeat(auto-fit,minmax(${preferences.tileSize === "small" ? 180 : preferences.tileSize === "wide" ? 360 : preferences.tileSize === "large" ? 280 : 230}px,1fr))`, gap: preferences.tileDensity === "compact" ? 6 : preferences.tileDensity === "spacious" ? 20 : 12 }}>{items.map((app) => <article role="listitem" key={app.id} style={{ padding: preferences.tileDensity === "compact" ? 8 : preferences.tileDensity === "spacious" ? 20 : 14, border: `1px solid ${preferences.favoriteAppIds.includes(app.id) ? "var(--color-warning)" : "var(--color-border)"}`, borderRadius: 8 }}><button aria-label={`${preferences.favoriteAppIds.includes(app.id) ? "Remove" : "Add"} ${app.name} favorite`} onClick={() => toggleFavorite(app.id)}>{preferences.favoriteAppIds.includes(app.id) ? "★" : "☆"}</button> <button onClick={() => handleSelect(app)}>{app.icon} {app.name}</button>{!preferences.hiddenFields.version && <span> v{app.version}</span>}{!preferences.hiddenFields.status && <div>Status: {app.status}</div>}{!preferences.hiddenFields.health && <div>Health: {app.health}</div>}{!preferences.hiddenFields.endpoints && <div>Endpoints: {app.endpoints.length}</div>}<button onClick={() => moveTop(app.id)} aria-label={`Move ${app.name} to top`}>Move to top</button></article>)}</div></section>)}</DataGuard>
    {selected && <AppDetailDrawer app={selected} onClose={() => setSelected(null)} />}
  </>;
}

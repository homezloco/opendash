import { describe, it, expect } from "vitest";
import { api } from "@/api";

describe("api client", () => {
  it("getSummary returns system summary", async () => {
    const summary = await api.getSummary();
    expect(summary).toHaveProperty("appsRunning");
    expect(summary).toHaveProperty("protectionStatus");
    expect(summary).toHaveProperty("health");
  });

  it("getApps returns array of installed apps", async () => {
    const apps = await api.getApps();
    expect(Array.isArray(apps)).toBe(true);
    expect(apps.length).toBeGreaterThan(0);
    expect(apps[0]).toHaveProperty("id");
    expect(apps[0]).toHaveProperty("status");
  });

  it("getApp returns a single app by id", async () => {
    const app = await api.getApp("nextcloud");
    expect(app.id).toBe("nextcloud");
    expect(app.name).toBe("Nextcloud");
  });

  it("getApp throws for unknown id", async () => {
    await expect(api.getApp("nonexistent")).rejects.toThrow("not found");
  });

  it("getAttentionItems returns array", async () => {
    const items = await api.getAttentionItems();
    expect(Array.isArray(items)).toBe(true);
  });

  it("getProtection returns protection overview", async () => {
    const protection = await api.getProtection();
    expect(protection).toHaveProperty("status");
    expect(protection).toHaveProperty("rules");
    expect(protection).toHaveProperty("threatsBlocked24h");
  });

  it("getCatalog returns catalog apps", async () => {
    const catalog = await api.getCatalog();
    expect(Array.isArray(catalog)).toBe(true);
  });

  it("getActivity returns activity events", async () => {
    const events = await api.getActivity();
    expect(Array.isArray(events)).toBe(true);
    expect(events[0]).toHaveProperty("type");
  });
});

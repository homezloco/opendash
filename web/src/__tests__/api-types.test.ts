import { describe, it, expect } from "vitest";
import { mockSummary, mockApps, mockAttentionItems, mockProtection, mockActivity, mockCatalogApps } from "@/api/mock-data";

describe("mock data integrity", () => {
  it("summary has valid ranges", () => {
    expect(mockSummary.appsRunning).toBeLessThanOrEqual(mockSummary.appsTotal);
    expect(mockSummary.cpuPercent).toBeGreaterThanOrEqual(0);
    expect(mockSummary.cpuPercent).toBeLessThanOrEqual(100);
    expect(mockSummary.memoryPercent).toBeGreaterThanOrEqual(0);
    expect(mockSummary.memoryPercent).toBeLessThanOrEqual(100);
    expect(mockSummary.storageUsedGb).toBeLessThanOrEqual(mockSummary.storageTotalGb);
  });

  it("all apps have required fields", () => {
    for (const app of mockApps) {
      expect(app.id).toBeTruthy();
      expect(app.name).toBeTruthy();
      expect(app.services.length).toBeGreaterThan(0);
      expect(app.storage.length).toBeGreaterThan(0);
    }
  });

  it("attention items have unique ids", () => {
    const ids = mockAttentionItems.map((i) => i.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it("protection has at least one rule", () => {
    expect(mockProtection.rules.length).toBeGreaterThan(0);
  });

  it("activity events are ordered by timestamp descending", () => {
    for (let i = 1; i < mockActivity.length; i++) {
      expect(new Date(mockActivity[i - 1].timestamp).getTime())
        .toBeGreaterThanOrEqual(new Date(mockActivity[i].timestamp).getTime());
    }
  });

  it("catalog apps have unique ids", () => {
    const ids = mockCatalogApps.map((a) => a.id);
    expect(new Set(ids).size).toBe(ids.length);
  });
});

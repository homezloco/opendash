import type {
  ActivityEvent,
  AttentionItem,
  AuthSession,
  BootstrapStatus,
  CatalogApp,
  CatalogAppDetail,
  ComposePreview,
  InstalledApp,
  InstallPlan,
  ManifestSource,
  Operation,
  ProtectionOverview,
  SourceListItem,
  SourcePreview,
  SystemSummary,
  UpdatePreview,
} from "./types";
import {
  mockActivity,
  mockApps,
  mockAttentionItems,
  mockCatalogApps,
  mockCatalogAppDetail,
  mockComposePreview,
  mockInstallPlan,
  mockOperation,
  mockProtection,
  mockSummary,
} from "./mock-data";

export const useMockApi = import.meta.env.VITE_USE_MOCK_API === "true";
const BASE = "/api/v1";
let csrfToken = "";

export class APIError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code?: string,
  ) {
    super(message);
    this.name = "APIError";
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  if (useMockApi) {
    await new Promise((resolve) => setTimeout(resolve, 80));
    return getMockData<T>(path);
  }

  const headers = new Headers(init?.headers);
  if (init?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const method = (init?.method ?? "GET").toUpperCase();
  if (csrfToken && !["GET", "HEAD", "OPTIONS"].includes(method)) {
    headers.set("X-CSRF-Token", csrfToken);
  }
  const response = await fetch(`${BASE}${path}`, {
    ...init,
    headers,
    credentials: "same-origin",
  });
  if (!response.ok) {
    if (response.status === 401) csrfToken = "";
    const contentType = response.headers.get("Content-Type") ?? "";
    const body = contentType.includes("application/json")
      ? ((await response.json().catch(() => null)) as {
          error?: string;
          code?: string;
        } | null)
      : null;
    throw new APIError(
      body?.error ?? `API request failed (${response.status})`,
      response.status,
      body?.code,
    );
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

function getMockData<T>(path: string): T {
  const routes: Record<string, unknown> = {
    "/summary": mockSummary,
    "/attention": mockAttentionItems,
    "/apps": mockApps,
    "/catalog": mockCatalogApps,
    "/protection": mockProtection,
    "/activity": mockActivity,
  };
  if (path.startsWith("/apps/")) {
    const rest = path.slice(6);
    if (rest.includes("/")) {
      const [_id, action] = rest.split("/");
      if (action === "logs") return { output: "mock log output" } as T;
      if (action === "uninstall-plan") return mockComposePreview as T;
      if (["start", "stop", "restart", "uninstall"].includes(action)) return mockOperation as T;
      throw new Error(`Unknown app action: ${action}`);
    }
    const app = mockApps.find((item) => item.id === rest);
    if (!app) throw new Error(`App not found: ${rest}`);
    return app as T;
  }
  if (path.startsWith("/catalog/")) {
    const rest = path.slice(9);
    const _id = rest.split("/")[0];
    void _id;
    if (rest.endsWith("/install-plan")) return mockInstallPlan as T;
    if (rest.includes("/install")) return mockOperation as T;
    return mockCatalogAppDetail as T;
  }
  if (path.startsWith("/operations/")) return mockOperation as T;
  const data = routes[path];
  if (data === undefined) throw new Error(`Unknown mock route: ${path}`);
  return data as T;
}

function saveSession(session: AuthSession): AuthSession {
  csrfToken = session.csrfToken;
  return session;
}

export const api = {
  getBootstrapStatus: () =>
    useMockApi
      ? Promise.resolve({ authEnabled: false, bootstrapRequired: false })
      : request<BootstrapStatus>("/bootstrap"),
  getSession: () => request<AuthSession>("/auth/session").then(saveSession),
  bootstrap: (username: string, password: string) =>
    request<AuthSession>("/bootstrap", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }).then(saveSession),
  login: (username: string, password: string) =>
    request<AuthSession>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }).then(saveSession),
  logout: async () => {
    if (!useMockApi) await request<void>("/auth/logout", { method: "POST" });
    csrfToken = "";
  },
  getSummary: () => request<SystemSummary>("/summary"),
  getAttentionItems: () => request<AttentionItem[]>("/attention"),
  getApps: () => request<InstalledApp[]>("/apps"),
  getApp: (id: string) => request<InstalledApp>(`/apps/${id}`),
  getCatalog: () => request<CatalogApp[]>("/catalog"),
  getCatalogApp: (id: string) => request<CatalogAppDetail>(`/catalog/${id}`),
  getInstallPlan: (id: string) => request<InstallPlan>(`/catalog/${id}/install-plan`),
  installApp: (id: string, config: Record<string, string> = {}) =>
    request<Operation>(`/catalog/${id}/install`, {
      method: "POST",
      body: JSON.stringify({ config }),
    }),
  getOperation: (id: string) => request<Operation>(`/operations/${id}`),
  startApp: (id: string) =>
    request<Operation>(`/apps/${id}/start`, { method: "POST" }),
  stopApp: (id: string) =>
    request<Operation>(`/apps/${id}/stop`, { method: "POST" }),
  restartApp: (id: string) =>
    request<Operation>(`/apps/${id}/restart`, { method: "POST" }),
  getAppLogs: (id: string, lines = 100) =>
    request<{ output: string }>(`/apps/${id}/logs?lines=${lines}`),
  getUninstallPlan: (id: string) => request<ComposePreview>(`/apps/${id}/uninstall-plan`),
  uninstallApp: (id: string) =>
    request<Operation>(`/apps/${id}/uninstall`, { method: "POST" }),
  getProtection: () => request<ProtectionOverview>("/protection"),
  getActivity: () => request<ActivityEvent[]>("/activity"),
  getSources: () => request<SourceListItem[]>("/sources"),
  previewSource: (url: string) =>
    request<SourcePreview>("/sources/preview", {
      method: "POST",
      body: JSON.stringify({ url }),
    }),
  addSource: (url: string, confirmed: boolean) =>
    request<ManifestSource>("/sources", {
      method: "POST",
      body: JSON.stringify({ url, confirmed }),
    }),
  removeSource: (id: string) =>
    request<void>(`/sources/${id}`, { method: "DELETE" }),
  getUpdatePlan: (id: string) => request<UpdatePreview>(`/apps/${id}/update-plan`),
  applyUpdate: (id: string, body: { commitSha: string; checksum: string; confirmed: boolean; acknowledgedRisks: boolean }) =>
    request<Operation>(`/apps/${id}/update`, { method: "POST", body: JSON.stringify(body) }),
};

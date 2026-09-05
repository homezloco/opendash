export interface BackupArchive {
  id: string; jobId: string; appId: string; size: number; sha256: string;
  manifestSha256?: string; createdAt: string; verifiedAt?: string; integrityOk?: boolean;
}
export interface RestorePreview {
  backupId: string; appId: string; volumes: string[]; images: string[];
  manifest: { id: string; name: string; version: string }; compatible: boolean;
  compatibility: string[]; migrationRisk?: string; rollbackSafe: boolean;
  sourceTrusted: boolean; integrityOk: boolean;
}
export interface VerificationResult { backupId: string; projectName: string; healthy: boolean; cleanedUp: boolean; verifiedAt: string; error?: string }
export interface RecoveryInventory { version: number; exportedAt: string; apps: unknown[]; sources: unknown[]; backups: BackupArchive[] }
export interface RecoveryImportPreview { valid: boolean; apps: number; sources: number; backups: number; untrustedSources: string[]; conflicts: { appId: string; kind: string; message: string }[]; strategy: string }

export interface BootstrapStatus {
  authEnabled: boolean;
  bootstrapRequired: boolean;
}

export interface AuthSession {
  username: string;
  csrfToken: string;
  expiresAt: string;
}

export type LoadingState = "idle" | "loading" | "success" | "error";

export interface ApiResult<T> {
  data: T | null;
  state: LoadingState;
  error: string | null;
}

export type ProtectionStatus = "armed" | "monitoring" | "degraded" | "offline";
export type HealthStatus = "healthy" | "warning" | "critical" | "unknown";
export type AppStatus = "running" | "stopped" | "updating" | "error" | "installing";

export interface SystemSummary {
  appsRunning: number;
  appsTotal: number;
  protectionStatus: ProtectionStatus;
  updatesAvailable: number;
  storageUsedGb: number;
  storageTotalGb: number;
  cpuPercent: number;
  memoryPercent: number;
  uptimeSeconds: number;
  health: HealthStatus;
}

export interface AttentionItem {
  id: string;
  severity: "critical" | "warning" | "info";
  title: string;
  description: string;
  timestamp: string;
  actionLabel?: string;
  actionRoute?: string;
  dismissed: boolean;
}

export interface AppEndpoint {
  label: string;
  url: string;
  kind: "web" | "api" | "tcp" | "udp";
}

export interface AppService {
  name: string;
  status: AppStatus;
  cpuPercent: number;
  memoryMb: number;
}

export interface AppStorageVolume {
  name: string;
  mountPath: string;
  usedGb: number;
  totalGb: number;
}

export interface InstalledApp {
  id: string;
  name: string;
  icon: string;
  description: string;
  status: AppStatus;
  health: HealthStatus;
  version: string;
  latestVersion: string;
  updateAvailable?: boolean;
  trust?: SourceTrust;
  category: string;
  endpoints: AppEndpoint[];
  services: AppService[];
  storage: AppStorageVolume[];
  updatedAt: string;
  installedAt: string;
}

export interface CatalogApp {
  id: string;
  name: string;
  icon: string;
  description: string;
  category: string;
  version: string;
  installed: boolean;
  trust?: SourceTrust;
}

export interface CatalogAppDetail extends CatalogApp {
  tags?: string[];
  website?: string;
  source?: string;
  maintainer?: string;
  license?: string;
  trust?: SourceTrust;
  compose: ManifestCompose;
  endpoints?: ManifestEndpoint[];
  storage?: ManifestStorage[];
  secrets?: ManifestSecret[];
  config?: ManifestConfig[];
  permissions?: ManifestPermission[];
}

export interface ManifestCompose {
  file?: string;
  inline?: ManifestInlineCompose;
  projectName?: string;
  mainService?: string;
}

export interface ManifestInlineCompose {
  services: Record<string, ManifestService>;
  networks?: Record<string, unknown>;
  volumes?: Record<string, unknown>;
}

export interface ManifestService {
  image: string;
  container_name?: string;
  restart?: string;
  ports?: string[];
  volumes?: string[];
  environment?: Record<string, string>;
  cap_add?: string[];
  cap_drop?: string[];
  privileged?: boolean;
  network_mode?: string;
  user?: string;
  depends_on?: string[];
}

export interface ManifestEndpoint {
  label: string;
  port: number;
  path?: string;
  kind: "web" | "api" | "tcp" | "udp";
}

export interface ManifestStorage {
  name: string;
  path: string;
  defaultSizeGb?: number;
}

export interface ManifestSecret {
  name: string;
  description?: string;
  required?: boolean;
}

export interface ManifestConfig {
  key: string;
  label?: string;
  description?: string;
  defaultValue?: string;
  required?: boolean;
  secret?: boolean;
}

export interface ManifestPermission {
  kind: string;
  description?: string;
  required?: boolean;
}

export interface InstallPlan {
  appId: string;
  name: string;
  version: string;
  images: string[];
  ports: PlannedPort[];
  volumes: PlannedVolume[];
  environment: PlannedConfig[];
  permissions: ManifestPermission[];
  risks: Risk[];
  conflicts: Conflict[];
  projectName: string;
  projectPath: string;
}

export interface PlannedPort {
  label: string;
  containerPort: number;
  hostPort?: number;
  kind: string;
}

export interface PlannedVolume {
  name: string;
  mountPath: string;
  hostPath: string;
}

export interface PlannedConfig {
  key: string;
  value: string;
  source: string;
  secret: boolean;
}

export interface Risk {
  severity: string;
  category: string;
  description: string;
}

export interface Conflict {
  kind: string;
  target: string;
  description: string;
}

export type OperationStatus = "pending" | "running" | "completed" | "failed" | "cancelled";
export type OperationKind = "install" | "start" | "stop" | "restart" | "uninstall";

export interface Operation {
  id: string;
  appId: string;
  kind: OperationKind;
  status: OperationStatus;
  error?: string;
  output?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ComposePreview {
  projectName: string;
  containers: string[];
  images: string[];
  volumes: string[];
  networks: string[];
  dataRetained: boolean;
}

export interface ProtectionRule {
  id: string;
  name: string;
  enabled: boolean;
  category: "firewall" | "vpn" | "dns" | "ids" | "backup";
  status: HealthStatus;
  description: string;
}

export interface ProtectionOverview {
  status: ProtectionStatus;
  rules: ProtectionRule[];
  threatsBlocked24h: number;
  lastScanAt: string;
  vpnConnected: boolean;
  firewallActive: boolean;
  dnsFilteringActive: boolean;
}

export interface ActivityEvent {
  id: string;
  type: "app" | "system" | "security" | "update" | "user";
  title: string;
  description: string;
  timestamp: string;
  severity: "info" | "warning" | "error" | "success";
}

export type SourceTrust =
  | "official"
  | "reviewed"
  | "user-trusted"
  | "changed"
  | "untrusted";

export interface ManifestSource {
  id: string;
  sourceUrl: string;
  owner: string;
  repo: string;
  path: string;
  ref: string;
  commitSha: string;
  checksum: string;
  trust: SourceTrust;
  createdAt: string;
  updatedAt: string;
}

export interface SourcePreview {
  sourceUrl: string;
  owner: string;
  repo: string;
  path: string;
  ref: string;
  commitSha: string;
  checksum: string;
  trust: SourceTrust;
  manifest: CatalogAppDetail;
}

export interface SourceListItem {
  id: string;
  sourceUrl: string;
  owner: string;
  repo: string;
  path: string;
  ref: string;
  commitSha: string;
  checksum: string;
  trust: SourceTrust;
  appId: string;
  name: string;
  version: string;
  updatedAt: string;
}

export interface PermissionDiff {
  kind: string;
  description?: string;
  required: boolean;
  state: "added" | "removed" | "changed" | "unchanged";
}

export interface UpdateChange {
  kind: string;
  target: string;
  before?: string;
  after?: string;
  summary: string;
}

export interface UpdatePreview {
  appId: string;
  currentVersion: string;
  newVersion: string;
  currentCommit: string;
  newCommit: string;
  newChecksum: string;
  currentManifest: CatalogAppDetail;
  newManifest: CatalogAppDetail;
  images: string[];
  addedImages: string[];
  removedImages: string[];
  ports: PlannedPort[];
  addedPorts: PlannedPort[];
  removedPorts: PlannedPort[];
  volumes: PlannedVolume[];
  addedVolumes: PlannedVolume[];
  removedVolumes: PlannedVolume[];
  environment: PlannedConfig[];
  permissions: ManifestPermission[];
  permissionDiff: PermissionDiff[];
  risks: Risk[];
  changes: UpdateChange[];
  canUpdate: boolean;
  blockedReasons: string[];
  migrationDisclosures: string[];
  rollbackDisclosures: string[];
  requiresAcknowledgement: boolean;
}

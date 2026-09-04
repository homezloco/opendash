import type {
  SystemSummary,
  AttentionItem,
  InstalledApp,
  CatalogApp,
  CatalogAppDetail,
  ComposePreview,
  InstallPlan,
  Operation,
  ProtectionOverview,
  ActivityEvent,
} from "./types";

export const mockSummary: SystemSummary = {
  appsRunning: 8,
  appsTotal: 10,
  protectionStatus: "armed",
  updatesAvailable: 3,
  storageUsedGb: 142,
  storageTotalGb: 500,
  cpuPercent: 23,
  memoryPercent: 61,
  uptimeSeconds: 1_296_000,
  health: "healthy",
};

export const mockAttentionItems: AttentionItem[] = [
  {
    id: "att-1",
    severity: "critical",
    title: "Nextcloud update failed",
    description:
      "Version 29.0.1 installation encountered a database migration error. Manual intervention may be required.",
    timestamp: "2026-09-04T08:15:00Z",
    actionLabel: "View Details",
    actionRoute: "/apps/nextcloud",
    dismissed: false,
  },
  {
    id: "att-2",
    severity: "warning",
    title: "3 updates available",
    description:
      "Immich, Jellyfin, and Vaultwarden have new versions ready to install.",
    timestamp: "2026-09-04T06:00:00Z",
    actionLabel: "Review Updates",
    actionRoute: "/apps",
    dismissed: false,
  },
  {
    id: "att-3",
    severity: "warning",
    title: "Storage usage above 80%",
    description:
      "Media volume is at 84% capacity. Consider expanding storage or cleaning up unused data.",
    timestamp: "2026-09-03T22:00:00Z",
    dismissed: false,
  },
  {
    id: "att-4",
    severity: "info",
    title: "Backup completed",
    description:
      "Nightly backup finished at 03:00. All 10 apps backed up (48.2 GB total).",
    timestamp: "2026-09-04T03:12:00Z",
    dismissed: false,
  },
];

export const mockApps: InstalledApp[] = [
  {
    id: "nextcloud",
    name: "Nextcloud",
    icon: "NC",
    description:
      "Self-hosted productivity platform with file sync, calendar, contacts, and collaboration tools.",
    status: "error",
    health: "critical",
    version: "28.0.4",
    latestVersion: "29.0.1",
    category: "Productivity",
    endpoints: [
      { label: "Web UI", url: "https://cloud.home.local", kind: "web" },
      { label: "WebDAV", url: "https://cloud.home.local/remote.php/dav", kind: "api" },
    ],
    services: [
      { name: "nextcloud-app", status: "error", cpuPercent: 0, memoryMb: 512 },
      { name: "nextcloud-db", status: "running", cpuPercent: 2, memoryMb: 256 },
      { name: "nextcloud-redis", status: "running", cpuPercent: 0.5, memoryMb: 64 },
    ],
    storage: [
      { name: "Data", mountPath: "/data", usedGb: 89, totalGb: 200 },
      { name: "Database", mountPath: "/var/lib/mysql", usedGb: 4.2, totalGb: 20 },
    ],
    updatedAt: "2026-09-04T08:15:00Z",
    installedAt: "2025-01-15T10:00:00Z",
  },
  {
    id: "immich",
    name: "Immich",
    icon: "IM",
    description:
      "High-performance self-hosted photo and video management with machine learning powered search.",
    status: "running",
    health: "healthy",
    version: "1.99.0",
    latestVersion: "1.100.0",
    category: "Media",
    endpoints: [
      { label: "Web UI", url: "https://photos.home.local", kind: "web" },
      { label: "API", url: "https://photos.home.local/api", kind: "api" },
    ],
    services: [
      { name: "immich-server", status: "running", cpuPercent: 8, memoryMb: 1024 },
      { name: "immich-ml", status: "running", cpuPercent: 15, memoryMb: 2048 },
      { name: "immich-db", status: "running", cpuPercent: 3, memoryMb: 384 },
    ],
    storage: [
      { name: "Photos", mountPath: "/photos", usedGb: 186, totalGb: 500 },
      { name: "Thumbnails", mountPath: "/cache", usedGb: 12, totalGb: 50 },
    ],
    updatedAt: "2026-08-28T14:00:00Z",
    installedAt: "2025-03-10T08:00:00Z",
  },
  {
    id: "jellyfin",
    name: "Jellyfin",
    icon: "JF",
    description:
      "Free software media system for streaming movies, TV shows, music, and live TV.",
    status: "running",
    health: "healthy",
    version: "10.9.8",
    latestVersion: "10.10.0",
    category: "Media",
    endpoints: [
      { label: "Web UI", url: "https://media.home.local", kind: "web" },
    ],
    services: [
      { name: "jellyfin", status: "running", cpuPercent: 5, memoryMb: 768 },
    ],
    storage: [
      { name: "Config", mountPath: "/config", usedGb: 0.8, totalGb: 5 },
      { name: "Media", mountPath: "/media", usedGb: 320, totalGb: 1000 },
    ],
    updatedAt: "2026-08-20T09:00:00Z",
    installedAt: "2025-02-01T12:00:00Z",
  },
  {
    id: "vaultwarden",
    name: "Vaultwarden",
    icon: "VW",
    description:
      "Lightweight Bitwarden-compatible password manager server written in Rust.",
    status: "running",
    health: "healthy",
    version: "1.31.0",
    latestVersion: "1.32.0",
    category: "Security",
    endpoints: [
      { label: "Web Vault", url: "https://vault.home.local", kind: "web" },
      { label: "API", url: "https://vault.home.local/api", kind: "api" },
    ],
    services: [
      { name: "vaultwarden", status: "running", cpuPercent: 0.5, memoryMb: 48 },
    ],
    storage: [
      { name: "Data", mountPath: "/data", usedGb: 0.1, totalGb: 5 },
    ],
    updatedAt: "2026-08-15T10:00:00Z",
    installedAt: "2025-01-10T09:00:00Z",
  },
  {
    id: "home-assistant",
    name: "Home Assistant",
    icon: "HA",
    description:
      "Open-source home automation platform with support for thousands of integrations.",
    status: "running",
    health: "healthy",
    version: "2026.9.1",
    latestVersion: "2026.9.1",
    category: "Automation",
    endpoints: [
      { label: "Web UI", url: "https://hass.home.local", kind: "web" },
      { label: "API", url: "https://hass.home.local/api", kind: "api" },
    ],
    services: [
      { name: "homeassistant", status: "running", cpuPercent: 4, memoryMb: 512 },
      { name: "mosquitto", status: "running", cpuPercent: 0.3, memoryMb: 16 },
      { name: "zigbee2mqtt", status: "running", cpuPercent: 1, memoryMb: 64 },
    ],
    storage: [
      { name: "Config", mountPath: "/config", usedGb: 2.5, totalGb: 10 },
    ],
    updatedAt: "2026-09-01T06:00:00Z",
    installedAt: "2025-01-05T08:00:00Z",
  },
  {
    id: "wireguard",
    name: "WireGuard",
    icon: "WG",
    description:
      "Modern, high-performance VPN tunnel with minimal attack surface.",
    status: "running",
    health: "healthy",
    version: "1.0.20210914",
    latestVersion: "1.0.20210914",
    category: "Network",
    endpoints: [
      { label: "VPN", url: "vpn.home.local:51820", kind: "udp" },
    ],
    services: [
      { name: "wireguard", status: "running", cpuPercent: 0.1, memoryMb: 8 },
    ],
    storage: [
      { name: "Config", mountPath: "/etc/wireguard", usedGb: 0.001, totalGb: 1 },
    ],
    updatedAt: "2026-07-01T10:00:00Z",
    installedAt: "2025-01-05T08:00:00Z",
  },
  {
    id: "grafana",
    name: "Grafana",
    icon: "GR",
    description:
      "Analytics and interactive visualization platform for metrics, logs, and traces.",
    status: "running",
    health: "warning",
    version: "11.1.0",
    latestVersion: "11.1.0",
    category: "Monitoring",
    endpoints: [
      { label: "Dashboard", url: "https://grafana.home.local", kind: "web" },
    ],
    services: [
      { name: "grafana", status: "running", cpuPercent: 2, memoryMb: 256 },
      { name: "prometheus", status: "running", cpuPercent: 6, memoryMb: 512 },
    ],
    storage: [
      { name: "Data", mountPath: "/var/lib/grafana", usedGb: 8, totalGb: 20 },
      { name: "Prometheus", mountPath: "/prometheus", usedGb: 45, totalGb: 100 },
    ],
    updatedAt: "2026-08-25T09:00:00Z",
    installedAt: "2025-02-20T14:00:00Z",
  },
  {
    id: "adguard",
    name: "AdGuard Home",
    icon: "AG",
    description:
      "Network-wide ad and tracker blocking DNS server with parental controls.",
    status: "running",
    health: "healthy",
    version: "0.107.52",
    latestVersion: "0.107.52",
    category: "Network",
    endpoints: [
      { label: "Admin", url: "https://dns.home.local", kind: "web" },
      { label: "DNS", url: "dns.home.local:53", kind: "udp" },
    ],
    services: [
      { name: "adguardhome", status: "running", cpuPercent: 1, memoryMb: 96 },
    ],
    storage: [
      { name: "Config", mountPath: "/opt/adguardhome/conf", usedGb: 0.05, totalGb: 1 },
      { name: "Work", mountPath: "/opt/adguardhome/work", usedGb: 2.1, totalGb: 10 },
    ],
    updatedAt: "2026-08-10T10:00:00Z",
    installedAt: "2025-01-05T08:00:00Z",
  },
  {
    id: "gitea",
    name: "Gitea",
    icon: "GT",
    description: "Lightweight self-hosted Git service with issue tracking and CI/CD.",
    status: "stopped",
    health: "unknown",
    version: "1.22.0",
    latestVersion: "1.22.0",
    category: "Development",
    endpoints: [
      { label: "Web UI", url: "https://git.home.local", kind: "web" },
      { label: "SSH", url: "git.home.local:2222", kind: "tcp" },
    ],
    services: [
      { name: "gitea", status: "stopped", cpuPercent: 0, memoryMb: 0 },
      { name: "gitea-db", status: "stopped", cpuPercent: 0, memoryMb: 0 },
    ],
    storage: [
      { name: "Repositories", mountPath: "/data/gitea/repositories", usedGb: 3.2, totalGb: 50 },
      { name: "Database", mountPath: "/var/lib/postgresql", usedGb: 0.5, totalGb: 10 },
    ],
    updatedAt: "2026-08-01T10:00:00Z",
    installedAt: "2025-04-10T10:00:00Z",
  },
  {
    id: "paperless",
    name: "Paperless-ngx",
    icon: "PL",
    description:
      "Document management system that transforms physical documents into a searchable online archive.",
    status: "running",
    health: "healthy",
    version: "2.9.0",
    latestVersion: "2.9.0",
    category: "Productivity",
    endpoints: [
      { label: "Web UI", url: "https://docs.home.local", kind: "web" },
      { label: "API", url: "https://docs.home.local/api", kind: "api" },
    ],
    services: [
      { name: "paperless-web", status: "running", cpuPercent: 2, memoryMb: 384 },
      { name: "paperless-worker", status: "running", cpuPercent: 1, memoryMb: 256 },
      { name: "paperless-db", status: "running", cpuPercent: 0.5, memoryMb: 128 },
    ],
    storage: [
      { name: "Documents", mountPath: "/usr/src/paperless/media", usedGb: 15, totalGb: 100 },
    ],
    updatedAt: "2026-08-30T09:00:00Z",
    installedAt: "2025-05-01T10:00:00Z",
  },
];

export const mockCatalogAppDetail: CatalogAppDetail = {
  id: "uptime-kuma",
  name: "Uptime Kuma",
  icon: "UK",
  description: "Self-hosted monitoring tool.",
  category: "Monitoring",
  version: "1.23.15",
  installed: false,
  tags: ["monitoring", "uptime"],
  website: "https://uptime.kuma.pet",
  source: "https://github.com/louislam/uptime-kuma",
  maintainer: "OpenDash",
  license: "MIT",
  compose: {
    projectName: "opendash-uptime-kuma",
    mainService: "uptime-kuma",
    inline: {
      services: {
        "uptime-kuma": {
          image:
            "docker.io/louislam/uptime-kuma:1.23.15@sha256:8eea954901871d3e33f2dbddf9d6be07d362b0fc01c4ec613f390d4bd3d99cf2",
          container_name: "opendash-uptime-kuma",
          restart: "unless-stopped",
          volumes: ["/app/data"],
        },
      },
      volumes: {
        "opendash-uptime-kuma-Data": {},
      },
    },
  },
  endpoints: [
    { label: "Web UI", port: 3001, path: "/", kind: "web" },
    { label: "TCP Ping", port: 3001, kind: "tcp" },
  ],
  storage: [{ name: "Data", path: "/app/data", defaultSizeGb: 5 }],
  secrets: [],
  config: [
    {
      key: "TZ",
      label: "Timezone",
      description: "Container timezone",
      defaultValue: "UTC",
      required: false,
      secret: false,
    },
  ],
  permissions: [],
};

export const mockInstallPlan: InstallPlan = {
  appId: "uptime-kuma",
  name: "Uptime Kuma",
  version: "1.23.15",
  images: [
    "docker.io/louislam/uptime-kuma:1.23.15@sha256:8eea954901871d3e33f2dbddf9d6be07d362b0fc01c4ec613f390d4bd3d99cf2",
  ],
  ports: [
    { label: "Web UI", containerPort: 3001, hostPort: 0, kind: "web" },
    { label: "TCP Ping", containerPort: 3001, hostPort: 0, kind: "tcp" },
  ],
  volumes: [
    {
      name: "opendash-uptime-kuma-Data",
      mountPath: "/app/data",
      hostPath: "/data/apps/uptime-kuma/volumes/Data",
    },
  ],
  environment: [{ key: "TZ", value: "UTC", source: "Timezone", secret: false }],
  permissions: [],
  risks: [],
  conflicts: [],
  projectName: "opendash-uptime-kuma",
  projectPath: "/data/apps/uptime-kuma",
};

export const mockOperation: Operation = {
  id: "op-1",
  appId: "uptime-kuma",
  kind: "install",
  status: "completed",
  output: "installed",
  error: "",
  createdAt: "2026-09-04T12:00:00Z",
  updatedAt: "2026-09-04T12:01:00Z",
};

export const mockComposePreview: ComposePreview = {
  projectName: "opendash-uptime-kuma",
  containers: ["opendash-uptime-kuma"],
  images: [
    "docker.io/louislam/uptime-kuma:1.23.15@sha256:8eea954901871d3e33f2dbddf9d6be07d362b0fc01c4ec613f390d4bd3d99cf2",
  ],
  volumes: ["opendash-uptime-kuma-Data"],
  networks: [],
  dataRetained: true,
};

export const mockCatalogApps: CatalogApp[] = [
  {
    id: "audiobookshelf",
    name: "Audiobookshelf",
    icon: "AB",
    description: "Self-hosted audiobook and podcast server.",
    category: "Media",
    version: "2.10.0",
    installed: false,
  },
  {
    id: "freshrss",
    name: "FreshRSS",
    icon: "FR",
    description: "Self-hosted RSS feed aggregator.",
    category: "Productivity",
    version: "1.24.1",
    installed: false,
  },
  {
    id: "mealie",
    name: "Mealie",
    icon: "ML",
    description: "Recipe management and meal planning.",
    category: "Productivity",
    version: "1.12.0",
    installed: false,
  },
  {
    id: "nextcloud",
    name: "Nextcloud",
    icon: "NC",
    description: "Self-hosted productivity platform.",
    category: "Productivity",
    version: "29.0.1",
    installed: true,
  },
  {
    id: "immich",
    name: "Immich",
    icon: "IM",
    description: "Photo and video management with ML search.",
    category: "Media",
    version: "1.100.0",
    installed: true,
  },
];

export const mockProtection: ProtectionOverview = {
  status: "armed",
  rules: [
    {
      id: "fw-1",
      name: "Inbound Firewall",
      enabled: true,
      category: "firewall",
      status: "healthy",
      description: "Block all unsolicited inbound connections except allowed ports.",
    },
    {
      id: "vpn-1",
      name: "WireGuard VPN",
      enabled: true,
      category: "vpn",
      status: "healthy",
      description: "Encrypted tunnel for remote access. 3 peers connected.",
    },
    {
      id: "dns-1",
      name: "DNS Filtering",
      enabled: true,
      category: "dns",
      status: "healthy",
      description: "AdGuard Home blocking 142,000+ domains. 12,489 queries blocked today.",
    },
    {
      id: "ids-1",
      name: "Intrusion Detection",
      enabled: true,
      category: "ids",
      status: "warning",
      description: "Suricata monitoring active. 2 suspicious events flagged in the last hour.",
    },
    {
      id: "backup-1",
      name: "Encrypted Backups",
      enabled: true,
      category: "backup",
      status: "healthy",
      description: "Nightly encrypted backups to off-site storage. Last backup: 6 hours ago.",
    },
  ],
  threatsBlocked24h: 12489,
  lastScanAt: "2026-09-04T03:00:00Z",
  vpnConnected: true,
  firewallActive: true,
  dnsFilteringActive: true,
};

export const mockActivity: ActivityEvent[] = [
  {
    id: "ev-1",
    type: "update",
    title: "Nextcloud update failed",
    description: "Update to v29.0.1 failed during database migration.",
    timestamp: "2026-09-04T08:15:00Z",
    severity: "error",
  },
  {
    id: "ev-2",
    type: "system",
    title: "Backup completed",
    description: "Nightly backup of all apps completed (48.2 GB).",
    timestamp: "2026-09-04T03:12:00Z",
    severity: "success",
  },
  {
    id: "ev-3",
    type: "security",
    title: "Suspicious DNS query blocked",
    description: "Query to known C2 domain blocked by AdGuard Home.",
    timestamp: "2026-09-04T02:45:00Z",
    severity: "warning",
  },
  {
    id: "ev-4",
    type: "app",
    title: "Immich ML indexing complete",
    description: "Face detection and object recognition finished for 1,247 new photos.",
    timestamp: "2026-09-03T23:30:00Z",
    severity: "info",
  },
  {
    id: "ev-5",
    type: "system",
    title: "System updated",
    description: "Host OS packages updated. No reboot required.",
    timestamp: "2026-09-03T22:00:00Z",
    severity: "success",
  },
  {
    id: "ev-6",
    type: "user",
    title: "New VPN peer connected",
    description: "Peer 'laptop-shane' connected via WireGuard.",
    timestamp: "2026-09-03T18:42:00Z",
    severity: "info",
  },
  {
    id: "ev-7",
    type: "app",
    title: "Gitea stopped",
    description: "Gitea was manually stopped by admin.",
    timestamp: "2026-09-03T16:00:00Z",
    severity: "info",
  },
  {
    id: "ev-8",
    type: "security",
    title: "Failed login attempt",
    description: "3 failed login attempts from 192.168.1.105 to Nextcloud.",
    timestamp: "2026-09-03T14:22:00Z",
    severity: "warning",
  },
  {
    id: "ev-9",
    type: "update",
    title: "Jellyfin update available",
    description: "Jellyfin v10.10.0 is available (current: v10.9.8).",
    timestamp: "2026-09-03T12:00:00Z",
    severity: "info",
  },
  {
    id: "ev-10",
    type: "system",
    title: "Certificate renewed",
    description: "Let's Encrypt certificates renewed for *.home.local.",
    timestamp: "2026-09-03T04:00:00Z",
    severity: "success",
  },
];

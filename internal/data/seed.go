package data

import (
	"time"

	"github.com/opendash-project/opendash/internal/models"
)

var Summary = &models.SystemSummary{
	AppsRunning:      8,
	AppsTotal:        10,
	ProtectionStatus: models.ProtectionArmed,
	UpdatesAvailable: 3,
	StorageUsedGb:    142,
	StorageTotalGb:   500,
	CPUPercent:       23,
	MemoryPercent:    61,
	UptimeSeconds:    1296000,
	Health:           models.HealthHealthy,
	DemoData:         true,
}

var AttentionItems = []models.AttentionItem{
	{
		ID:          "att-1",
		Severity:    models.SeverityCritical,
		Title:       "Nextcloud update failed",
		Description: "Version 29.0.1 installation encountered a database migration error. Manual intervention may be required.",
		Timestamp:   mustTime("2026-09-04T08:15:00Z"),
		ActionLabel: "View Details",
		ActionRoute: "/apps/nextcloud",
		Dismissed:   false,
	},
	{
		ID:          "att-2",
		Severity:    models.SeverityWarning,
		Title:       "3 updates available",
		Description: "Immich, Jellyfin, and Vaultwarden have new versions ready to install.",
		Timestamp:   mustTime("2026-09-04T06:00:00Z"),
		ActionLabel: "Review Updates",
		ActionRoute: "/apps",
		Dismissed:   false,
	},
	{
		ID:          "att-3",
		Severity:    models.SeverityWarning,
		Title:       "Storage usage above 80%",
		Description: "Media volume is at 84% capacity. Consider expanding storage or cleaning up unused data.",
		Timestamp:   mustTime("2026-09-03T22:00:00Z"),
		Dismissed:   false,
	},
	{
		ID:          "att-4",
		Severity:    models.SeverityInfo,
		Title:       "Backup completed",
		Description: "Nightly backup finished at 03:00. All 10 apps backed up (48.2 GB total).",
		Timestamp:   mustTime("2026-09-04T03:12:00Z"),
		Dismissed:   false,
	},
}

var Apps = []models.InstalledApp{
	{
		ID:            "nextcloud",
		Name:          "Nextcloud",
		Icon:          "NC",
		Description:   "Self-hosted productivity platform with file sync, calendar, contacts, and collaboration tools.",
		Status:        models.AppError,
		Health:        models.HealthCritical,
		Version:       "28.0.4",
		LatestVersion: "29.0.1",
		Category:      "Productivity",
		Endpoints: []models.AppEndpoint{
			{Label: "Web UI", URL: "https://cloud.home.local", Kind: models.EndpointWeb},
			{Label: "WebDAV", URL: "https://cloud.home.local/remote.php/dav", Kind: models.EndpointAPI},
		},
		Services: []models.AppService{
			{Name: "nextcloud-app", Status: models.AppError, CPUPercent: 0, MemoryMb: 512},
			{Name: "nextcloud-db", Status: models.AppRunning, CPUPercent: 2, MemoryMb: 256},
			{Name: "nextcloud-redis", Status: models.AppRunning, CPUPercent: 0.5, MemoryMb: 64},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Data", MountPath: "/data", UsedGb: 89, TotalGb: 200},
			{Name: "Database", MountPath: "/var/lib/mysql", UsedGb: 4.2, TotalGb: 20},
		},
		UpdatedAt:   mustTime("2026-09-04T08:15:00Z"),
		InstalledAt: mustTime("2025-01-15T10:00:00Z"),
	},
	{
		ID:            "immich",
		Name:          "Immich",
		Icon:          "IM",
		Description:   "High-performance self-hosted photo and video management with machine learning powered search.",
		Status:        models.AppRunning,
		Health:        models.HealthHealthy,
		Version:       "1.99.0",
		LatestVersion: "1.100.0",
		Category:      "Media",
		Endpoints: []models.AppEndpoint{
			{Label: "Web UI", URL: "https://photos.home.local", Kind: models.EndpointWeb},
			{Label: "API", URL: "https://photos.home.local/api", Kind: models.EndpointAPI},
		},
		Services: []models.AppService{
			{Name: "immich-server", Status: models.AppRunning, CPUPercent: 8, MemoryMb: 1024},
			{Name: "immich-ml", Status: models.AppRunning, CPUPercent: 15, MemoryMb: 2048},
			{Name: "immich-db", Status: models.AppRunning, CPUPercent: 3, MemoryMb: 384},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Photos", MountPath: "/photos", UsedGb: 186, TotalGb: 500},
			{Name: "Thumbnails", MountPath: "/cache", UsedGb: 12, TotalGb: 50},
		},
		UpdatedAt:   mustTime("2026-08-28T14:00:00Z"),
		InstalledAt: mustTime("2025-03-10T08:00:00Z"),
	},
	{
		ID:            "jellyfin",
		Name:          "Jellyfin",
		Icon:          "JF",
		Description:   "Free software media system for streaming movies, TV shows, music, and live TV.",
		Status:        models.AppRunning,
		Health:        models.HealthHealthy,
		Version:       "10.9.8",
		LatestVersion: "10.10.0",
		Category:      "Media",
		Endpoints: []models.AppEndpoint{
			{Label: "Web UI", URL: "https://media.home.local", Kind: models.EndpointWeb},
		},
		Services: []models.AppService{
			{Name: "jellyfin", Status: models.AppRunning, CPUPercent: 5, MemoryMb: 768},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Config", MountPath: "/config", UsedGb: 0.8, TotalGb: 5},
			{Name: "Media", MountPath: "/media", UsedGb: 320, TotalGb: 1000},
		},
		UpdatedAt:   mustTime("2026-08-20T09:00:00Z"),
		InstalledAt: mustTime("2025-02-01T12:00:00Z"),
	},
	{
		ID:            "vaultwarden",
		Name:          "Vaultwarden",
		Icon:          "VW",
		Description:   "Lightweight Bitwarden-compatible password manager server written in Rust.",
		Status:        models.AppRunning,
		Health:        models.HealthHealthy,
		Version:       "1.31.0",
		LatestVersion: "1.32.0",
		Category:      "Security",
		Endpoints: []models.AppEndpoint{
			{Label: "Web Vault", URL: "https://vault.home.local", Kind: models.EndpointWeb},
			{Label: "API", URL: "https://vault.home.local/api", Kind: models.EndpointAPI},
		},
		Services: []models.AppService{
			{Name: "vaultwarden", Status: models.AppRunning, CPUPercent: 0.5, MemoryMb: 48},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Data", MountPath: "/data", UsedGb: 0.1, TotalGb: 5},
		},
		UpdatedAt:   mustTime("2026-08-15T10:00:00Z"),
		InstalledAt: mustTime("2025-01-10T09:00:00Z"),
	},
	{
		ID:            "home-assistant",
		Name:          "Home Assistant",
		Icon:          "HA",
		Description:   "Open-source home automation platform with support for thousands of integrations.",
		Status:        models.AppRunning,
		Health:        models.HealthHealthy,
		Version:       "2026.9.1",
		LatestVersion: "2026.9.1",
		Category:      "Automation",
		Endpoints: []models.AppEndpoint{
			{Label: "Web UI", URL: "https://hass.home.local", Kind: models.EndpointWeb},
			{Label: "API", URL: "https://hass.home.local/api", Kind: models.EndpointAPI},
		},
		Services: []models.AppService{
			{Name: "homeassistant", Status: models.AppRunning, CPUPercent: 4, MemoryMb: 512},
			{Name: "mosquitto", Status: models.AppRunning, CPUPercent: 0.3, MemoryMb: 16},
			{Name: "zigbee2mqtt", Status: models.AppRunning, CPUPercent: 1, MemoryMb: 64},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Config", MountPath: "/config", UsedGb: 2.5, TotalGb: 10},
		},
		UpdatedAt:   mustTime("2026-09-01T06:00:00Z"),
		InstalledAt: mustTime("2025-01-05T08:00:00Z"),
	},
	{
		ID:            "wireguard",
		Name:          "WireGuard",
		Icon:          "WG",
		Description:   "Modern, high-performance VPN tunnel with minimal attack surface.",
		Status:        models.AppRunning,
		Health:        models.HealthHealthy,
		Version:       "1.0.20210914",
		LatestVersion: "1.0.20210914",
		Category:      "Network",
		Endpoints: []models.AppEndpoint{
			{Label: "VPN", URL: "vpn.home.local:51820", Kind: models.EndpointUDP},
		},
		Services: []models.AppService{
			{Name: "wireguard", Status: models.AppRunning, CPUPercent: 0.1, MemoryMb: 8},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Config", MountPath: "/etc/wireguard", UsedGb: 0.001, TotalGb: 1},
		},
		UpdatedAt:   mustTime("2026-07-01T10:00:00Z"),
		InstalledAt: mustTime("2025-01-05T08:00:00Z"),
	},
	{
		ID:            "grafana",
		Name:          "Grafana",
		Icon:          "GR",
		Description:   "Analytics and interactive visualization platform for metrics, logs, and traces.",
		Status:        models.AppRunning,
		Health:        models.HealthWarning,
		Version:       "11.1.0",
		LatestVersion: "11.1.0",
		Category:      "Monitoring",
		Endpoints: []models.AppEndpoint{
			{Label: "Dashboard", URL: "https://grafana.home.local", Kind: models.EndpointWeb},
		},
		Services: []models.AppService{
			{Name: "grafana", Status: models.AppRunning, CPUPercent: 2, MemoryMb: 256},
			{Name: "prometheus", Status: models.AppRunning, CPUPercent: 6, MemoryMb: 512},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Data", MountPath: "/var/lib/grafana", UsedGb: 8, TotalGb: 20},
			{Name: "Prometheus", MountPath: "/prometheus", UsedGb: 45, TotalGb: 100},
		},
		UpdatedAt:   mustTime("2026-08-25T09:00:00Z"),
		InstalledAt: mustTime("2025-02-20T14:00:00Z"),
	},
	{
		ID:            "adguard",
		Name:          "AdGuard Home",
		Icon:          "AG",
		Description:   "Network-wide ad and tracker blocking DNS server with parental controls.",
		Status:        models.AppRunning,
		Health:        models.HealthHealthy,
		Version:       "0.107.52",
		LatestVersion: "0.107.52",
		Category:      "Network",
		Endpoints: []models.AppEndpoint{
			{Label: "Admin", URL: "https://dns.home.local", Kind: models.EndpointWeb},
			{Label: "DNS", URL: "dns.home.local:53", Kind: models.EndpointUDP},
		},
		Services: []models.AppService{
			{Name: "adguardhome", Status: models.AppRunning, CPUPercent: 1, MemoryMb: 96},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Config", MountPath: "/opt/adguardhome/conf", UsedGb: 0.05, TotalGb: 1},
			{Name: "Work", MountPath: "/opt/adguardhome/work", UsedGb: 2.1, TotalGb: 10},
		},
		UpdatedAt:   mustTime("2026-08-10T10:00:00Z"),
		InstalledAt: mustTime("2025-01-05T08:00:00Z"),
	},
	{
		ID:            "gitea",
		Name:          "Gitea",
		Icon:          "GT",
		Description:   "Lightweight self-hosted Git service with issue tracking and CI/CD.",
		Status:        models.AppStopped,
		Health:        models.HealthUnknown,
		Version:       "1.22.0",
		LatestVersion: "1.22.0",
		Category:      "Development",
		Endpoints: []models.AppEndpoint{
			{Label: "Web UI", URL: "https://git.home.local", Kind: models.EndpointWeb},
			{Label: "SSH", URL: "git.home.local:2222", Kind: models.EndpointTCP},
		},
		Services: []models.AppService{
			{Name: "gitea", Status: models.AppStopped, CPUPercent: 0, MemoryMb: 0},
			{Name: "gitea-db", Status: models.AppStopped, CPUPercent: 0, MemoryMb: 0},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Repositories", MountPath: "/data/gitea/repositories", UsedGb: 3.2, TotalGb: 50},
			{Name: "Database", MountPath: "/var/lib/postgresql", UsedGb: 0.5, TotalGb: 10},
		},
		UpdatedAt:   mustTime("2026-08-01T10:00:00Z"),
		InstalledAt: mustTime("2025-04-10T10:00:00Z"),
	},
	{
		ID:            "paperless",
		Name:          "Paperless-ngx",
		Icon:          "PL",
		Description:   "Document management system that transforms physical documents into a searchable online archive.",
		Status:        models.AppRunning,
		Health:        models.HealthHealthy,
		Version:       "2.9.0",
		LatestVersion: "2.9.0",
		Category:      "Productivity",
		Endpoints: []models.AppEndpoint{
			{Label: "Web UI", URL: "https://docs.home.local", Kind: models.EndpointWeb},
			{Label: "API", URL: "https://docs.home.local/api", Kind: models.EndpointAPI},
		},
		Services: []models.AppService{
			{Name: "paperless-web", Status: models.AppRunning, CPUPercent: 2, MemoryMb: 384},
			{Name: "paperless-worker", Status: models.AppRunning, CPUPercent: 1, MemoryMb: 256},
			{Name: "paperless-db", Status: models.AppRunning, CPUPercent: 0.5, MemoryMb: 128},
		},
		Storage: []models.AppStorageVolume{
			{Name: "Documents", MountPath: "/usr/src/paperless/media", UsedGb: 15, TotalGb: 100},
		},
		UpdatedAt:   mustTime("2026-08-30T09:00:00Z"),
		InstalledAt: mustTime("2025-05-01T10:00:00Z"),
	},
}

var CatalogApps = []models.CatalogApp{
	{ID: "audiobookshelf", Name: "Audiobookshelf", Icon: "AB", Description: "Self-hosted audiobook and podcast server.", Category: "Media", Version: "2.10.0", Installed: false},
	{ID: "freshrss", Name: "FreshRSS", Icon: "FR", Description: "Self-hosted RSS feed aggregator.", Category: "Productivity", Version: "1.24.1", Installed: false},
	{ID: "mealie", Name: "Mealie", Icon: "ML", Description: "Recipe management and meal planning.", Category: "Productivity", Version: "1.12.0", Installed: false},
	{ID: "nextcloud", Name: "Nextcloud", Icon: "NC", Description: "Self-hosted productivity platform.", Category: "Productivity", Version: "29.0.1", Installed: true},
	{ID: "immich", Name: "Immich", Icon: "IM", Description: "Photo and video management with ML search.", Category: "Media", Version: "1.100.0", Installed: true},
}

var Protection = &models.ProtectionOverview{
	Status: models.ProtectionArmed,
	Rules: []models.ProtectionRule{
		{ID: "fw-1", Name: "Inbound Firewall", Enabled: true, Category: models.ProtectionCategoryFirewall, Status: models.HealthHealthy, Description: "Block all unsolicited inbound connections except allowed ports."},
		{ID: "vpn-1", Name: "WireGuard VPN", Enabled: true, Category: models.ProtectionCategoryVPN, Status: models.HealthHealthy, Description: "Encrypted tunnel for remote access. 3 peers connected."},
		{ID: "dns-1", Name: "DNS Filtering", Enabled: true, Category: models.ProtectionCategoryDNS, Status: models.HealthHealthy, Description: "AdGuard Home blocking 142,000+ domains. 12,489 queries blocked today."},
		{ID: "ids-1", Name: "Intrusion Detection", Enabled: true, Category: models.ProtectionCategoryIDS, Status: models.HealthWarning, Description: "Suricata monitoring active. 2 suspicious events flagged in the last hour."},
		{ID: "backup-1", Name: "Encrypted Backups", Enabled: true, Category: models.ProtectionCategoryBackup, Status: models.HealthHealthy, Description: "Nightly encrypted backups to off-site storage. Last backup: 6 hours ago."},
	},
	ThreatsBlocked24h:  12489,
	LastScanAt:         mustTime("2026-09-04T03:00:00Z"),
	VpnConnected:       true,
	FirewallActive:     true,
	DnsFilteringActive: true,
}

var ActivityEvents = []models.ActivityEvent{
	{ID: "ev-1", Type: models.ActivityUpdate, Title: "Nextcloud update failed", Description: "Update to v29.0.1 failed during database migration.", Timestamp: mustTime("2026-09-04T08:15:00Z"), Severity: models.SeverityError},
	{ID: "ev-2", Type: models.ActivitySystem, Title: "Backup completed", Description: "Nightly backup of all apps completed (48.2 GB).", Timestamp: mustTime("2026-09-04T03:12:00Z"), Severity: models.SeveritySuccess},
	{ID: "ev-3", Type: models.ActivitySecurity, Title: "Suspicious DNS query blocked", Description: "Query to known C2 domain blocked by AdGuard Home.", Timestamp: mustTime("2026-09-04T02:45:00Z"), Severity: models.SeverityWarning},
	{ID: "ev-4", Type: models.ActivityApp, Title: "Immich ML indexing complete", Description: "Face detection and object recognition finished for 1,247 new photos.", Timestamp: mustTime("2026-09-03T23:30:00Z"), Severity: models.SeverityInfo},
	{ID: "ev-5", Type: models.ActivitySystem, Title: "System updated", Description: "Host OS packages updated. No reboot required.", Timestamp: mustTime("2026-09-03T22:00:00Z"), Severity: models.SeveritySuccess},
	{ID: "ev-6", Type: models.ActivityUser, Title: "New VPN peer connected", Description: "Peer 'laptop-shane' connected via WireGuard.", Timestamp: mustTime("2026-09-03T18:42:00Z"), Severity: models.SeverityInfo},
	{ID: "ev-7", Type: models.ActivityApp, Title: "Gitea stopped", Description: "Gitea was manually stopped by admin.", Timestamp: mustTime("2026-09-03T16:00:00Z"), Severity: models.SeverityInfo},
	{ID: "ev-8", Type: models.ActivitySecurity, Title: "Failed login attempt", Description: "3 failed login attempts from 192.168.1.105 to Nextcloud.", Timestamp: mustTime("2026-09-03T14:22:00Z"), Severity: models.SeverityWarning},
	{ID: "ev-9", Type: models.ActivityUpdate, Title: "Jellyfin update available", Description: "Jellyfin v10.10.0 is available (current: v10.9.8).", Timestamp: mustTime("2026-09-03T12:00:00Z"), Severity: models.SeverityInfo},
	{ID: "ev-10", Type: models.ActivitySystem, Title: "Certificate renewed", Description: "Let's Encrypt certificates renewed for *.home.local.", Timestamp: mustTime("2026-09-03T04:00:00Z"), Severity: models.SeveritySuccess},
}

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

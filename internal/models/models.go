package models

import "time"

type ProtectionStatus string

const (
	ProtectionArmed      ProtectionStatus = "armed"
	ProtectionMonitoring ProtectionStatus = "monitoring"
	ProtectionDegraded   ProtectionStatus = "degraded"
	ProtectionOffline    ProtectionStatus = "offline"
)

type HealthStatus string

const (
	HealthHealthy  HealthStatus = "healthy"
	HealthWarning  HealthStatus = "warning"
	HealthCritical HealthStatus = "critical"
	HealthUnknown  HealthStatus = "unknown"
)

type AppStatus string

const (
	AppRunning    AppStatus = "running"
	AppStopped    AppStatus = "stopped"
	AppUpdating   AppStatus = "updating"
	AppError      AppStatus = "error"
	AppInstalling AppStatus = "installing"
)

type EndpointKind string

const (
	EndpointWeb EndpointKind = "web"
	EndpointAPI EndpointKind = "api"
	EndpointTCP EndpointKind = "tcp"
	EndpointUDP EndpointKind = "udp"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
	SeveritySuccess  Severity = "success"
	SeverityError    Severity = "error"
)

type ActivityType string

const (
	ActivityApp      ActivityType = "app"
	ActivitySystem   ActivityType = "system"
	ActivitySecurity ActivityType = "security"
	ActivityUpdate   ActivityType = "update"
	ActivityUser     ActivityType = "user"
)

type ProtectionCategory string

const (
	ProtectionCategoryFirewall ProtectionCategory = "firewall"
	ProtectionCategoryVPN      ProtectionCategory = "vpn"
	ProtectionCategoryDNS      ProtectionCategory = "dns"
	ProtectionCategoryIDS      ProtectionCategory = "ids"
	ProtectionCategoryBackup   ProtectionCategory = "backup"
)

type SystemSummary struct {
	AppsRunning      int              `json:"appsRunning"`
	AppsTotal        int              `json:"appsTotal"`
	ProtectionStatus ProtectionStatus `json:"protectionStatus"`
	UpdatesAvailable int              `json:"updatesAvailable"`
	StorageUsedGb    float64          `json:"storageUsedGb"`
	StorageTotalGb   float64          `json:"storageTotalGb"`
	CPUPercent       float64          `json:"cpuPercent"`
	MemoryPercent    float64          `json:"memoryPercent"`
	UptimeSeconds    int64            `json:"uptimeSeconds"`
	Health           HealthStatus     `json:"health"`
	DemoData         bool             `json:"demoData,omitempty"`
}

type AttentionItem struct {
	ID          string    `json:"id"`
	Severity    Severity  `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	ActionLabel string    `json:"actionLabel,omitempty"`
	ActionRoute string    `json:"actionRoute,omitempty"`
	Dismissed   bool      `json:"dismissed"`
}

type AppEndpoint struct {
	Label string       `json:"label"`
	URL   string       `json:"url"`
	Kind  EndpointKind `json:"kind"`
}

type AppService struct {
	Name       string    `json:"name"`
	Status     AppStatus `json:"status"`
	CPUPercent float64   `json:"cpuPercent"`
	MemoryMb   float64   `json:"memoryMb"`
}

type AppStorageVolume struct {
	Name      string  `json:"name"`
	MountPath string  `json:"mountPath"`
	UsedGb    float64 `json:"usedGb"`
	TotalGb   float64 `json:"totalGb"`
}

type InstalledApp struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Icon            string             `json:"icon"`
	Description     string             `json:"description"`
	Status          AppStatus          `json:"status"`
	Health          HealthStatus       `json:"health"`
	Version         string             `json:"version"`
	LatestVersion   string             `json:"latestVersion"`
	UpdateAvailable bool               `json:"updateAvailable"`
	Trust           SourceTrust        `json:"trust,omitempty"`
	Category        string             `json:"category"`
	Endpoints       []AppEndpoint      `json:"endpoints"`
	Services        []AppService       `json:"services"`
	Storage         []AppStorageVolume `json:"storage"`
	UpdatedAt       time.Time          `json:"updatedAt"`
	InstalledAt     time.Time          `json:"installedAt"`
}

type CatalogApp struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Icon        string      `json:"icon"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	Version     string      `json:"version"`
	Installed   bool        `json:"installed"`
	Trust       SourceTrust `json:"trust,omitempty"`
}

type CatalogAppDetail struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Description string               `json:"description"`
	Category    string               `json:"category"`
	Tags        []string             `json:"tags,omitempty"`
	Icon        string               `json:"icon"`
	Website     string               `json:"website,omitempty"`
	Source      string               `json:"source,omitempty"`
	Maintainer  string               `json:"maintainer,omitempty"`
	License     string               `json:"license,omitempty"`
	Trust       SourceTrust          `json:"trust,omitempty"`
	Compose     ManifestCompose      `json:"compose"`
	Endpoints   []ManifestEndpoint   `json:"endpoints,omitempty"`
	Storage     []ManifestStorage    `json:"storage,omitempty"`
	Secrets     []ManifestSecret     `json:"secrets,omitempty"`
	Config      []ManifestConfig     `json:"config,omitempty"`
	Permissions []ManifestPermission `json:"permissions,omitempty"`
}

type ProtectionRule struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Enabled     bool               `json:"enabled"`
	Category    ProtectionCategory `json:"category"`
	Status      HealthStatus       `json:"status"`
	Description string             `json:"description"`
}

type ProtectionOverview struct {
	Status             ProtectionStatus `json:"status"`
	Rules              []ProtectionRule `json:"rules"`
	ThreatsBlocked24h  int              `json:"threatsBlocked24h"`
	LastScanAt         time.Time        `json:"lastScanAt"`
	VpnConnected       bool             `json:"vpnConnected"`
	FirewallActive     bool             `json:"firewallActive"`
	DnsFilteringActive bool             `json:"dnsFilteringActive"`
}

type ActivityEvent struct {
	ID          string       `json:"id"`
	Type        ActivityType `json:"type"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Timestamp   time.Time    `json:"timestamp"`
	Severity    Severity     `json:"severity"`
}

type Manifest struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Description string               `json:"description"`
	Category    string               `json:"category"`
	Tags        []string             `json:"tags,omitempty"`
	Icon        string               `json:"icon"`
	Website     string               `json:"website,omitempty"`
	Source      string               `json:"source,omitempty"`
	Maintainer  string               `json:"maintainer,omitempty"`
	License     string               `json:"license,omitempty"`
	Compose     ManifestCompose      `json:"compose"`
	Endpoints   []ManifestEndpoint   `json:"endpoints,omitempty"`
	Storage     []ManifestStorage    `json:"storage,omitempty"`
	Secrets     []ManifestSecret     `json:"secrets,omitempty"`
	Config      []ManifestConfig     `json:"config,omitempty"`
	Permissions []ManifestPermission `json:"permissions,omitempty"`
	Upgrade     *ManifestUpgrade     `json:"upgrade,omitempty"`
}

// ManifestUpgrade contains declarative update safety disclosures. It never
// executes migration hooks.
type ManifestUpgrade struct {
	MigrationRisk string `json:"migrationRisk,omitempty"`
	RollbackSafe  bool   `json:"rollbackSafe,omitempty"`
	RollbackNotes string `json:"rollbackNotes,omitempty"`
}

type ManifestCompose struct {
	File        string                 `json:"file,omitempty"`
	Inline      *ManifestInlineCompose `json:"inline,omitempty"`
	ProjectName string                 `json:"projectName,omitempty"`
	MainService string                 `json:"mainService,omitempty"`
}

type ManifestInlineCompose struct {
	Services map[string]ManifestService `json:"services"`
	Networks map[string]any             `json:"networks,omitempty"`
	Volumes  map[string]any             `json:"volumes,omitempty"`
}

type ManifestService struct {
	Image         string               `json:"image"`
	ContainerName string               `json:"container_name,omitempty"`
	Restart       string               `json:"restart,omitempty"`
	Ports         []string             `json:"ports,omitempty"`
	Volumes       []string             `json:"volumes,omitempty"`
	Environment   map[string]string    `json:"environment,omitempty"`
	CapAdd        []string             `json:"cap_add,omitempty"`
	CapDrop       []string             `json:"cap_drop,omitempty"`
	Privileged    bool                 `json:"privileged,omitempty"`
	NetworkMode   string               `json:"network_mode,omitempty"`
	User          string               `json:"user,omitempty"`
	DependsOn     []string             `json:"depends_on,omitempty"`
	HealthCheck   *ManifestHealthCheck `json:"healthcheck,omitempty"`
}

type ManifestHealthCheck struct {
	Test     []string `json:"test,omitempty"`
	Interval string   `json:"interval,omitempty"`
	Timeout  string   `json:"timeout,omitempty"`
	Retries  int      `json:"retries,omitempty"`
	Disable  bool     `json:"disable,omitempty"`
}

type ManifestEndpoint struct {
	Label string       `json:"label"`
	Port  int          `json:"port"`
	Path  string       `json:"path,omitempty"`
	Kind  EndpointKind `json:"kind"`
}

type ManifestStorage struct {
	Name          string  `json:"name"`
	Path          string  `json:"path"`
	DefaultSizeGb float64 `json:"defaultSizeGb,omitempty"`
}

type ManifestSecret struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type ManifestConfig struct {
	Key          string `json:"key"`
	Label        string `json:"label,omitempty"`
	Description  string `json:"description,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Required     bool   `json:"required,omitempty"`
	Secret       bool   `json:"secret,omitempty"`
}

type ManifestPermission struct {
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type CatalogIndex struct {
	Version     string         `json:"version"`
	GeneratedAt string         `json:"generatedAt"`
	SchemaURL   string         `json:"schemaUrl,omitempty"`
	Apps        []CatalogEntry `json:"apps"`
	Categories  []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"categories,omitempty"`
}

type CatalogEntry struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Description  string `json:"description,omitempty"`
	Category     string `json:"category"`
	Icon         string `json:"icon,omitempty"`
	Installed    bool   `json:"installed,omitempty"`
	ManifestPath string `json:"manifestPath"`
}

type InstallPlan struct {
	AppID       string               `json:"appId"`
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Images      []string             `json:"images"`
	Ports       []PlannedPort        `json:"ports"`
	Volumes     []PlannedVolume      `json:"volumes"`
	Environment []PlannedConfig      `json:"environment"`
	Permissions []ManifestPermission `json:"permissions"`
	Risks       []Risk               `json:"risks"`
	Conflicts   []Conflict           `json:"conflicts"`
	ProjectName string               `json:"projectName"`
	ProjectPath string               `json:"projectPath"`
}

type PlannedPort struct {
	Label         string `json:"label"`
	ContainerPort int    `json:"containerPort"`
	HostPort      int    `json:"hostPort,omitempty"`
	Kind          string `json:"kind"`
}

type PlannedVolume struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	HostPath  string `json:"hostPath"`
}

type PlannedConfig struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Source string `json:"source"`
	Secret bool   `json:"secret"`
}

type Risk struct {
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type Conflict struct {
	Kind        string `json:"kind"`
	Target      string `json:"target"`
	Description string `json:"description"`
}

type OperationStatus string

const (
	OperationPending   OperationStatus = "pending"
	OperationRunning   OperationStatus = "running"
	OperationCompleted OperationStatus = "completed"
	OperationFailed    OperationStatus = "failed"
	OperationCancelled OperationStatus = "cancelled"
)

type OperationKind string

const (
	OperationInstall   OperationKind = "install"
	OperationUpdate    OperationKind = "update"
	OperationStart     OperationKind = "start"
	OperationStop      OperationKind = "stop"
	OperationRestart   OperationKind = "restart"
	OperationUninstall OperationKind = "uninstall"
)

type Operation struct {
	ID        string          `json:"id"`
	AppID     string          `json:"appId"`
	Kind      OperationKind   `json:"kind"`
	Status    OperationStatus `json:"status"`
	Error     string          `json:"error,omitempty"`
	Output    string          `json:"output,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type AppInstanceStatus string

const (
	AppInstanceInstalling AppInstanceStatus = "installing"
	AppInstanceRunning    AppInstanceStatus = "running"
	AppInstanceStopped    AppInstanceStatus = "stopped"
	AppInstanceError      AppInstanceStatus = "error"
)

type AppInstance struct {
	ID          string             `json:"id"`
	CatalogID   string             `json:"catalogId"`
	SourceID    string             `json:"sourceId,omitempty"`
	Trust       SourceTrust        `json:"trust,omitempty"`
	Name        string             `json:"name"`
	Version     string             `json:"version"`
	ProjectName string             `json:"projectName"`
	InstallPath string             `json:"installPath"`
	ComposePath string             `json:"composePath"`
	Status      AppInstanceStatus  `json:"status"`
	Health      HealthStatus       `json:"health"`
	Endpoints   []AppEndpoint      `json:"endpoints"`
	Storage     []AppStorageVolume `json:"storage"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

type Revision struct {
	ID          string    `json:"id"`
	AppID       string    `json:"appId"`
	Manifest    Manifest  `json:"manifest"`
	ConfigJSON  string    `json:"configJson"`
	ComposePath string    `json:"composePath"`
	CreatedAt   time.Time `json:"createdAt"`
}

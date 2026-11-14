package catalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/opendash-project/opendash/internal/models"
)

var (
	idPattern      = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	semVerPattern  = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`)
	portMapPattern = regexp.MustCompile(`^((\d{1,3}\.)?\d{1,3}\.\d{1,3}\.\d{1,3}:)?(\d+):(\d+)$`)
)

const projectNamePrefix = "opendash-"

func LoadIndex(root string) (*models.CatalogIndex, error) {
	data, err := os.ReadFile(filepath.Join(root, "index.json"))
	if err != nil {
		return nil, fmt.Errorf("read catalog index: %w", err)
	}
	var idx models.CatalogIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse catalog index: %w", err)
	}
	return &idx, nil
}

func LoadManifest(path string) (*models.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m models.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &m, nil
}

func ValidateManifest(m *models.Manifest) error {
	if !idPattern.MatchString(m.ID) {
		return fmt.Errorf("invalid manifest id: %q", m.ID)
	}
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("manifest name is required")
	}
	if !semVerPattern.MatchString(m.Version) {
		return fmt.Errorf("invalid version %q", m.Version)
	}
	if m.Compose.File == "" && m.Compose.Inline == nil {
		return errors.New("compose.file or compose.inline is required")
	}
	if m.Compose.File != "" && m.Compose.Inline != nil {
		return errors.New("compose.file and compose.inline are mutually exclusive")
	}
	if m.Compose.Inline != nil {
		if len(m.Compose.Inline.Services) == 0 {
			return errors.New("compose.inline.services must contain at least one service")
		}
		for name, svc := range m.Compose.Inline.Services {
			if svc.Image == "" {
				return fmt.Errorf("service %q image is required", name)
			}
			for _, p := range svc.Ports {
				if _, _, err := ParsePortMapping(p); err != nil {
					return fmt.Errorf("service %q invalid port mapping %q: %w", name, p, err)
				}
			}
		}
	}
	seenStorage := make(map[string]struct{})
	for _, st := range m.Storage {
		if _, ok := seenStorage[st.Name]; ok {
			return fmt.Errorf("duplicate storage name %q", st.Name)
		}
		seenStorage[st.Name] = struct{}{}
	}
	seenConfig := make(map[string]struct{})
	for _, cfg := range m.Config {
		if _, ok := seenConfig[cfg.Key]; ok {
			return fmt.Errorf("duplicate config key %q", cfg.Key)
		}
		seenConfig[cfg.Key] = struct{}{}
	}
	seenEndpoints := make(map[int]struct{})
	for _, ep := range m.Endpoints {
		if _, ok := seenEndpoints[ep.Port]; ok {
			return fmt.Errorf("duplicate endpoint port %d", ep.Port)
		}
		seenEndpoints[ep.Port] = struct{}{}
	}
	return nil
}

func ParsePortMapping(s string) (hostPort, containerPort int, err error) {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 1:
		cp, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid port %q", s)
		}
		return 0, cp, nil
	case 2:
		hp, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil || hp < 1 || hp > 65535 {
			return 0, 0, fmt.Errorf("invalid host port in %q", s)
		}
		cp, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || cp < 1 || cp > 65535 {
			return 0, 0, fmt.Errorf("invalid container port in %q", s)
		}
		return hp, cp, nil
	case 3:
		hp, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || hp < 1 || hp > 65535 {
			return 0, 0, fmt.Errorf("invalid host port in %q", s)
		}
		cp, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil || cp < 1 || cp > 65535 {
			return 0, 0, fmt.Errorf("invalid container port in %q", s)
		}
		return hp, cp, nil
	default:
		return 0, 0, fmt.Errorf("invalid port mapping %q", s)
	}
}

func DefaultProjectName(m *models.Manifest) string {
	if m.Compose.ProjectName != "" {
		return m.Compose.ProjectName
	}
	return projectNamePrefix + m.ID
}

func InstallPath(appsRoot string, m *models.Manifest) string {
	return filepath.Join(appsRoot, "apps", m.ID)
}

func ApplyConfig(m *models.Manifest, overrides map[string]string) (map[string]string, []models.PlannedConfig, error) {
	result := make(map[string]string, len(m.Config))
	var planned []models.PlannedConfig
	for _, cfg := range m.Config {
		value := cfg.DefaultValue
		if v, ok := overrides[cfg.Key]; ok {
			value = v
		}
		if cfg.Required && strings.TrimSpace(value) == "" {
			return nil, nil, fmt.Errorf("config %q is required", cfg.Key)
		}
		result[cfg.Key] = value
		planned = append(planned, models.PlannedConfig{
			Key:    cfg.Key,
			Value:  value,
			Source: cfg.Label,
			Secret: cfg.Secret,
		})
	}
	return result, planned, nil
}

func BuildInstallPlan(m *models.Manifest, appsRoot string, overrides map[string]string, existing []models.AppInstance) (*models.InstallPlan, error) {
	if err := ValidateManifest(m); err != nil {
		return nil, err
	}
	configValues, plannedConfig, err := ApplyConfig(m, overrides)
	if err != nil {
		return nil, err
	}
	projectName := DefaultProjectName(m)
	installPath := InstallPath(appsRoot, m)
	plan := &models.InstallPlan{
		AppID:       m.ID,
		Name:        m.Name,
		Version:     m.Version,
		ProjectName: projectName,
		ProjectPath: installPath,
		Environment: plannedConfig,
		Permissions: m.Permissions,
	}

	images := make(map[string]struct{})
	if m.Compose.Inline != nil {
		for _, svc := range m.Compose.Inline.Services {
			images[svc.Image] = struct{}{}
		}
	}
	for img := range images {
		plan.Images = append(plan.Images, img)
	}
	sort.Strings(plan.Images)

	for _, ep := range m.Endpoints {
		plan.Ports = append(plan.Ports, models.PlannedPort{
			Label:         ep.Label,
			ContainerPort: ep.Port,
			HostPort:      0,
			Kind:          string(ep.Kind),
		})
	}
	for _, st := range m.Storage {
		name := volumeName(m.ID, st.Name)
		plan.Volumes = append(plan.Volumes, models.PlannedVolume{
			Name:      name,
			MountPath: st.Path,
			HostPath:  name,
		})
	}

	plan.Risks = evaluateRisks(m, configValues)
	plan.Conflicts = evaluateConflicts(m, projectName, installPath, existing)

	// Include implicit risks for required permissions not yet granted.
	for _, perm := range m.Permissions {
		if perm.Required {
			plan.Risks = append(plan.Risks, models.Risk{
				Severity:    "warning",
				Category:    perm.Kind,
				Description: perm.Description,
			})
		}
	}

	// Add secrets to environment preview with placeholder.
	for _, sec := range m.Secrets {
		value := configValues[sec.Name]
		if value == "" {
			value = ""
		}
		plan.Environment = append(plan.Environment, models.PlannedConfig{
			Key:    sec.Name,
			Value:  value,
			Source: sec.Description,
			Secret: true,
		})
	}

	return plan, nil
}

func evaluateRisks(m *models.Manifest, configValues map[string]string) []models.Risk {
	var risks []models.Risk
	if m.Compose.Inline != nil {
		for name, svc := range m.Compose.Inline.Services {
			if svc.Privileged {
				risks = append(risks, models.Risk{
					Severity:    "critical",
					Category:    "privileged",
					Description: fmt.Sprintf("service %q runs privileged", name),
				})
			}
			if svc.NetworkMode == "host" {
				risks = append(risks, models.Risk{
					Severity:    "critical",
					Category:    "network",
					Description: fmt.Sprintf("service %q uses host network namespace", name),
				})
			}
			for _, cap := range svc.CapAdd {
				if isHighRiskCapability(cap) {
					risks = append(risks, models.Risk{
						Severity:    "warning",
						Category:    "capability",
						Description: fmt.Sprintf("service %q adds capability %q", name, cap),
					})
				}
			}
			if svc.User == "root" || svc.User == "0" {
				risks = append(risks, models.Risk{
					Severity:    "warning",
					Category:    "user",
					Description: fmt.Sprintf("service %q runs as root", name),
				})
			}
			if !imagePinned(svc.Image) {
				risks = append(risks, models.Risk{
					Severity:    "warning",
					Category:    "image",
					Description: fmt.Sprintf("service %q image %q is not pinned to a digest", name, svc.Image),
				})
			}
		}
	}
	return risks
}

func isHighRiskCapability(cap string) bool {
	switch strings.ToLower(cap) {
	case "net_admin", "sys_admin", "sys_ptrace", "sys_module", "dac_read_search", "dac_override":
		return true
	}
	return false
}

func imagePinned(image string) bool {
	return strings.Contains(image, "@sha256:")
}

func evaluateConflicts(m *models.Manifest, projectName, installPath string, existing []models.AppInstance) []models.Conflict {
	var conflicts []models.Conflict
	seenProjects := make(map[string]bool)
	seenPaths := make(map[string]bool)
	seenPorts := make(map[int]bool)
	for _, inst := range existing {
		if inst.CatalogID == m.ID {
			conflicts = append(conflicts, models.Conflict{
				Kind:        "installed",
				Target:      inst.CatalogID,
				Description: fmt.Sprintf("%q is already installed as %q", m.ID, inst.Name),
			})
		}
		seenProjects[inst.ProjectName] = true
		seenPaths[inst.InstallPath] = true
		for _, ep := range inst.Endpoints {
			// Best-effort parse host port from endpoint URL.
			if port := urlPort(ep.URL); port > 0 {
				seenPorts[port] = true
			}
		}
	}
	if seenProjects[projectName] {
		conflicts = append(conflicts, models.Conflict{
			Kind:        "project",
			Target:      projectName,
			Description: fmt.Sprintf("project name %q is already in use", projectName),
		})
	}
	if seenPaths[installPath] {
		conflicts = append(conflicts, models.Conflict{
			Kind:        "path",
			Target:      installPath,
			Description: fmt.Sprintf("install path %q is already in use", installPath),
		})
	}
	// If manifest exposes fixed host ports, check conflicts.
	if m.Compose.Inline != nil {
		for _, svc := range m.Compose.Inline.Services {
			for _, p := range svc.Ports {
				hp, _, _ := ParsePortMapping(p)
				if hp > 0 && seenPorts[hp] {
					conflicts = append(conflicts, models.Conflict{
						Kind:        "port",
						Target:      strconv.Itoa(hp),
						Description: fmt.Sprintf("host port %d is already in use", hp),
					})
				}
			}
		}
	}
	return conflicts
}

func urlPort(raw string) int {
	u := strings.TrimPrefix(raw, "https://")
	u = strings.TrimPrefix(u, "http://")
	if _, port, err := net.SplitHostPort(u); err == nil {
		if n, err := strconv.Atoi(port); err == nil {
			return n
		}
	}
	return 0
}

func volumeName(appID, name string) string {
	return projectNamePrefix + appID + "-" + strings.ToLower(name)
}

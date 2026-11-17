package catalog

import (
	\"errors\"
	\"fmt\"
	\"path/filepath\"
	\"regexp\"
	\"strconv\"
	\"strings\"

	\"github.com/opendash-project/opendash/internal/models\"
)

const projectNamePrefix = \"opendash-\"

var (
	idPattern      = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	semVerPattern  = regexp.MustCompile(`^\\d+\\.\\d+\\.\\d+(-[a-zA-Z0-9.]+)?(\\+[a-zA-Z0-9.]+)?$`)
	ghURLPattern   = regexp.MustCompile(`^https://github\\.com/[^/]+/[^/]+/blob/[^/]+/.+$`)
)

func ValidateManifest(m *models.Manifest) error {
	if !idPattern.MatchString(m.ID) {
		return fmt.Errorf(\"invalid manifest id: %q\", m.ID)
	}
	if strings.TrimSpace(m.Name) == \"\" {
		return errors.New(\"manifest name is required\")
	}
	if !semVerPattern.MatchString(m.Version) {
		return fmt.Errorf(\"invalid version %q\", m.Version)
	}
	if m.Source != \"\" {
		if !strings.HasPrefix(m.Source, \"https://github.com/\") {
			return fmt.Errorf(\"unsupported manifest source: %q\", m.Source)
		}
		if !ghURLPattern.MatchString(m.Source) {
			return fmt.Errorf(\"invalid github manifest source: %q\", m.Source)
		}
	}
	if m.Compose.File == \"\" && m.Compose.Inline == nil {
		return errors.New(\"compose.file or compose.inline is required\")
	}
	if m.Compose.File != \"\" && m.Compose.Inline != nil {
		return errors.New(\"compose.file and compose.inline are mutually exclusive\")
	}
	if m.Compose.Inline != nil {
		if len(m.Compose.Inline.Services) == 0 {
			return errors.New(\"compose.inline.services must contain at least one service\")
		}
		for name, svc := range m.Compose.Inline.Services {
			if svc.Image == \"\" {
				return fmt.Errorf(\"service %q image is required\", name)
			}
			for _, p := range svc.Ports {
				if _, _, err := ParsePortMapping(p); err != nil {
					return fmt.Errorf(\"service %q invalid port mapping %q: %w\", name, p, err)
				}
			}
		}
	}
	seenStorage := make(map[string]struct{})
	for _, st := range m.Storage {
		if _, ok := seenStorage[st.Name]; ok {
			return fmt.Errorf(\"duplicate storage name %q\", st.Name)
		}
		seenStorage[st.Name] = struct{}{}
	}
	seenConfig := make(map[string]struct{})
	for _, cfg := range m.Config {
		if _, ok := seenConfig[cfg.Key]; ok {
			return fmt.Errorf(\"duplicate config key %q\", cfg.Key)
		}
		seenConfig[cfg.Key] = struct{}{}
	}
	seenEndpoints := make(map[int]struct{})
	for _, ep := range m.Endpoints {
		if _, ok := seenEndpoints[ep.Port]; ok {
			return fmt.Errorf(\"duplicate endpoint port %d\", ep.Port)
		}
		seenEndpoints[ep.Port] = struct{}{}
	}
	return nil
}

func ParsePortMapping(s string) (hostPort, containerPort int, err error) {
	parts := strings.Split(s, \":\")
	switch len(parts) {
	case 1:
		cp, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return 0, 0, fmt.Errorf(\"invalid port %q\", s)
		}
		return 0, cp, nil
	case 2:
		hp, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil || hp < 1 || hp > 65535 {
			return 0, 0, fmt.Errorf(\"invalid host port in %q\", s)
		}
		cp, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || cp < 1 || cp > 65535 {
			return 0, 0, fmt.Errorf(\"invalid container port in %q\", s)
		}
		return hp, cp, nil
	case 3:
		hp, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || hp < 1 || hp > 65535 {
			return 0, 0, fmt.Errorf(\"invalid host port in %q\", s)
		}
		cp, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || cp < 1 || cp > 65535 {
			return 0, 0, fmt.Errorf(\"invalid container port in %q\", s)
		}
		return hp, cp, nil
	default:
		return 0, 0, fmt.Errorf(\"invalid port mapping %q\", s)
	}
}

func DefaultProjectName(m *models.Manifest) string {
	if m.Compose.ProjectName != \"\" {
		return m.Compose.ProjectName
	}
	return projectNamePrefix + m.ID
}

func InstallPath(appsRoot string, m *models.Manifest) string {
	return filepath.Join(appsRoot, \"apps\", m.ID)
}

func ApplyConfig(m *models.Manifest, overrides map[string]string) (map[string]string, []models.PlannedConfig, error) {
    return nil, nil, nil
}
\",path:
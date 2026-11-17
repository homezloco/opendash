package catalog

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/homezloco/opendash/internal/models"
)

var (
	idPattern     = regexp.MustCompile(`^[a-z0-9-]+$`)
	semVerPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?$`)
	ghURLPattern  = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/blob/[^/]+/.+$`)
)

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
	if m.Source != "" {
		if !strings.HasPrefix(m.Source, "https://github.com/") {
			return fmt.Errorf("unsupported manifest source: %q", m.Source)
		}
		if !ghURLPattern.MatchString(m.Source) {
			return fmt.Errorf("invalid github manifest source: %q", m.Source)
		}
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
	// ... (rest of the file preserved)
	return 0, 0, nil
}

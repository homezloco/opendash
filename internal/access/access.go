package access

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/opendash-project/opendash/internal/models"
)

func DirectLink(endpoint models.AppEndpoint) (string, error) {
	if strings.TrimSpace(endpoint.URL) == "" {
		return "", fmt.Errorf("empty endpoint URL")
	}
	if endpoint.Kind != models.EndpointWeb && endpoint.Kind != models.EndpointAPI && endpoint.Kind != models.EndpointTCP && endpoint.Kind != models.EndpointUDP {
		return "", fmt.Errorf("unsupported endpoint kind: %s", endpoint.Kind)
	}

	raw := endpoint.URL
	if !strings.Contains(raw, "://") {
		switch endpoint.Kind {
		case models.EndpointWeb, models.EndpointAPI:
			raw = "https://" + raw
		case models.EndpointTCP:
			raw = "tcp://" + raw
		case models.EndpointUDP:
			raw = "udp://" + raw
		}
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		return "", fmt.Errorf("missing URL scheme")
	}
	if u.Host == "" && u.Path == "" {
		return "", fmt.Errorf("missing host")
	}
	return u.String(), nil
}

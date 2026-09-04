package access_test

import (
	"testing"

	"github.com/opendash-project/opendash/internal/access"
	"github.com/opendash-project/opendash/internal/models"
)

func TestDirectLinkWeb(t *testing.T) {
	endpoint := models.AppEndpoint{Label: "Web UI", URL: "https://cloud.home.local", Kind: models.EndpointWeb}
	got, err := access.DirectLink(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://cloud.home.local" {
		t.Fatalf("expected https URL, got %s", got)
	}
}

func TestDirectLinkNoScheme(t *testing.T) {
	endpoint := models.AppEndpoint{Label: "Web UI", URL: "cloud.home.local", Kind: models.EndpointWeb}
	got, err := access.DirectLink(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://cloud.home.local" {
		t.Fatalf("expected https://cloud.home.local, got %s", got)
	}
}

func TestDirectLinkTCP(t *testing.T) {
	endpoint := models.AppEndpoint{Label: "SSH", URL: "git.home.local:2222", Kind: models.EndpointTCP}
	got, err := access.DirectLink(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if got != "tcp://git.home.local:2222" {
		t.Fatalf("expected tcp URL, got %s", got)
	}
}

func TestDirectLinkUDP(t *testing.T) {
	endpoint := models.AppEndpoint{Label: "DNS", URL: "dns.home.local:53", Kind: models.EndpointUDP}
	got, err := access.DirectLink(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if got != "udp://dns.home.local:53" {
		t.Fatalf("expected udp URL, got %s", got)
	}
}

func TestDirectLinkEmptyURL(t *testing.T) {
	endpoint := models.AppEndpoint{Label: "X", URL: "", Kind: models.EndpointWeb}
	if _, err := access.DirectLink(endpoint); err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestDirectLinkInvalidKind(t *testing.T) {
	endpoint := models.AppEndpoint{Label: "X", URL: "x", Kind: models.EndpointKind("ftp")}
	if _, err := access.DirectLink(endpoint); err == nil {
		t.Fatal("expected error for unsupported kind")
	}
}

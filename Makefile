.PHONY: all build test vet security docker-build validate-schemas dist clean \
        fmt web-typecheck web-test web-build go-test go-vet go-build go-build-trimpath go-test-docker

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BINARY_NAME = opendash
DIST_DIR = dist
DIST_NAME = opendash-$(VERSION)-$(GOOS)-$(GOARCH)
DIST_TARBALL = $(DIST_DIR)/$(DIST_NAME).tar.gz

GO = go
NPM = npm
NODE = node
GOFLAGS = -trimpath -ldflags "-s -w -X main.version=$(VERSION)"
CGO_ENABLED = 0

all: build

build: web-build go-build

test: go-test web-test

vet: go-vet

fmt:
	gofmt -w cmd internal

web-typecheck:
	cd web && $(NPM) run typecheck

web-test:
	cd web && $(NPM) run test

web-build:
	cd web && $(NPM) run build

go-test:
	$(GO) test ./...

go-vet:
	$(GO) vet ./...

go-build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/opendash

go-build-trimpath:
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME) ./cmd/opendash

go-test-docker:
	docker run --rm -v $(shell pwd):/src -w /src golang:1.22 go test ./...

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

validate-schemas:
	$(NODE) validate-schemas.mjs

security: go-vet go-build-trimpath
	@echo "Checking docker.sock permissions..."
	@if [ -S /var/run/docker.sock ]; then \
		perms=$$(stat -c '%a' /var/run/docker.sock 2>/dev/null || echo "unknown"); \
		if [ "$$perms" = "666" ] || [ "$$perms" = "777" ]; then \
			echo "WARNING: docker.sock is world-writable ($$perms); restrict access to the opendash service account"; \
		fi; \
	else \
		echo "docker.sock not present; skipping permission check"; \
	fi
	cd web && $(NPM) audit --audit-level=high || echo "WARNING: npm audit reported high-severity findings; review before release"
	@if command -v gosec >/dev/null 2>&1; then \
		echo "Running gosec..."; \
		gosec ./...; \
	else \
		echo "WARNING: gosec is not installed; skipping gosec scan"; \
	fi

dist: build
	@rm -rf $(DIST_DIR)/$(DIST_NAME)
	@mkdir -p $(DIST_DIR)/$(DIST_NAME)/web
	cp $(BINARY_NAME) $(DIST_DIR)/$(DIST_NAME)/
	cp -R web/dist $(DIST_DIR)/$(DIST_NAME)/web/
	cp -R deploy/systemd $(DIST_DIR)/$(DIST_NAME)/
	cp README.md $(DIST_DIR)/$(DIST_NAME)/
	tar -czf $(DIST_TARBALL) --sort=name --mtime='2024-01-01 00:00:00Z' --owner=0 --group=0 --numeric-owner -C $(DIST_DIR) $(DIST_NAME)
	@echo "Distribution: $(DIST_TARBALL)"

clean:
	rm -rf $(BINARY_NAME) $(DIST_DIR)
	cd web && rm -rf dist node_modules *.tsbuildinfo

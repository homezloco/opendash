.PHONY: all web-typecheck web-test web-build go-test go-vet go-build go-test-docker docker-build validate-schemas

all: web-build go-build

web-typecheck:
	cd web && npm run typecheck

web-test:
	cd web && npm test

web-build:
	cd web && npm run build

go-test:
	go test ./...

go-vet:
	go vet ./...

go-build:
	go build -o opendash ./cmd/opendash

go-test-docker:
	docker run --rm -v $(shell pwd):/src -w /src golang:1.23 go test ./...

docker-build:
	docker build -t opendash:latest .

validate-schemas:
	node validate-schemas.mjs

# syntax=docker/dockerfile:1

# Stage 1: build the React SPA.
FROM node:22-alpine AS web-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: build the Go binary with the SPA available for serving.
ARG VERSION=0.1.0
FROM golang:1.22-alpine AS go-builder
ARG VERSION
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o opendash ./cmd/opendash

# Stage 3: minimal runtime image.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /var/lib/opendash
COPY --from=go-builder --chown=nonroot:nonroot /src/opendash /opendash
COPY --from=go-builder --chown=nonroot:nonroot /src/web/dist /var/lib/opendash/web/dist
USER nonroot:nonroot
EXPOSE 8080
ENV OPENDASH_WEB_ROOT=/var/lib/opendash/web/dist
ENTRYPOINT ["/opendash"]

FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o opendash ./cmd/opendash && mkdir /data

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /src/opendash /opendash
COPY --from=builder --chown=nonroot:nonroot /data /var/lib/opendash

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/opendash"]

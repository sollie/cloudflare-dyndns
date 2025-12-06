FROM golang:1.25.5-alpine AS builder

WORKDIR /builder
RUN apk update && apk --no-cache add ca-certificates tzdata

COPY source /builder
RUN CGO_ENABLED=0 go build -o cloudflare-dyndns

FROM ghcr.io/sollie/docker-upx:v5.0.1 AS upx
WORKDIR /upx
COPY --from=builder /builder/cloudflare-dyndns /upx/cloudflare-dyndns
RUN upx --best cloudflare-dyndns

FROM scratch AS final
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY --from=upx /upx/cloudflare-dyndns /app/

ENTRYPOINT ["/app/cloudflare-dyndns"]

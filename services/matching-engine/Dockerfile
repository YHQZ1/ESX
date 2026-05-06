FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata &&     adduser -D -g "" appuser

WORKDIR /app
COPY . .

WORKDIR /app/services/matching-engine

RUN --mount=type=cache,target=/go/pkg/mod     --mount=type=cache,target=/root/.cache/go-build     CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/server ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /bin/server /bin/server

USER appuser
EXPOSE 9094
ENTRYPOINT ["/bin/server"]

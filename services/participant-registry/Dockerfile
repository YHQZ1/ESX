FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY packages/ packages/
COPY services/ services/
COPY tests/ tests/

WORKDIR /app/services/participant-registry
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/server /bin/server
EXPOSE 8081 9091
ENTRYPOINT ["/bin/server"]
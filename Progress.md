# ESX — Progress Tracker

## Status: Phase 1 — Complete. Phase 2 — Up Next.

---

## Completed

### Infrastructure

- [x] `docker-compose.yml` — 7 containers on named `esx` network
  - Postgres 16 → `localhost:5433`
  - Redis 7 → `localhost:6379`
  - Kafka KRaft (confluentinc/cp-kafka:7.6.0) → `kafka:9092` (internal), `localhost:9092` (host via /etc/hosts)
  - Kafka UI → `localhost:9080`
  - Prometheus → `localhost:9090`
  - Grafana → `localhost:3001` (admin/admin)
  - Jaeger → `localhost:16686`
- [x] `infra/postgres/init.sql` — creates all 8 service databases on first boot
- [x] `infra/observability/prometheus/prometheus.yml` — scrapes all 8 services via host.docker.internal
- [x] `/etc/hosts` — `127.0.0.1 kafka` added so host services can resolve kafka broker

### Kafka Topics (all created manually)

- [x] `trade.executed`
- [x] `trade.cleared`
- [x] `trade.settled`
- [x] `order.submitted`
- [x] `order.cancelled`
- [x] `order.partially_filled`
- [x] `risk.rejected`
- [x] `circuit.breaker.triggered`
- [x] `circuit.breaker.lifted`

### Shared Packages

- [x] `packages/logger` — zerolog structured JSON logger
  - `logger.New("service-name")` returns a logger
  - Methods: Trace, Debug, Info, Warn, Error, Fatal
  - Field helpers: Str, Int, Int64, Bool, Any
  - Log level controlled via `LOG_LEVEL` env var
- [x] `packages/kafka` — segmentio/kafka-go producer/consumer wrapper
  - `NewProducer(brokers, topic, log)` → `Publish(ctx, key, payload)`
  - `NewConsumer(brokers, topic, groupID, log)` → `RegisterHandler` + `Start(ctx)`
  - `Decode[T](msg)` generic helper for unmarshalling messages
  - All 9 topic constants defined in `topics.go`
- [x] `packages/proto` — gRPC proto definitions + generated Go stubs
  - `participant.proto` → `ValidateAPIKey` RPC
  - `risk.proto` → `CheckAndLock`, `ReleaseLock` RPCs
  - `matching.proto` → `SubmitOrder`, `CancelOrder` RPCs
  - Generated stubs at `packages/proto/participant/`, `risk/`, `matching/`
  - `packages/proto/go.mod` exists with grpc + protobuf dependencies

### Services

- [x] `services/participant-registry` — identity, API keys, accounts `:8081` + gRPC `:9091`
  - REST: `POST /participants/register`, `POST /participants/:id/deposit`, `GET /participants/:id`
  - gRPC: `ValidateAPIKey` — hashes incoming key, looks up participant, returns participant_id + status
  - Schema: `participants`, `api_keys`, `cash_accounts`, `securities_accounts`
  - API keys: `crypto/rand` generated, SHA-256 hashed, raw key returned once only
  - `.env` loaded via `godotenv`
  - All endpoints tested and working

- [x] `services/risk-engine` — pre-trade collateral validation, locking gRPC `:9093`
  - gRPC: `CheckAndLock` — validates and locks cash (BUY) or shares (SELL)
  - gRPC: `ReleaseLock` — releases lock on cancel, partial release on partial fill
  - Schema: `locks` table in `risk_engine` database
  - BUY: verifies `cash_accounts.balance - locked >= price * quantity`, locks cash
  - SELL: verifies `securities_accounts.quantity - locked >= quantity`, locks shares
  - Reads directly from `participant_registry` database for balance checks
  - gRPC reflection enabled for grpcurl testing
  - All lock operations atomic — no partial state possible
  - Tested: approved and rejected cases both verified

- [x] `services/matching-engine` — order book, price-time priority matching gRPC `:9094`
  - gRPC: `SubmitOrder`, `CancelOrder`
  - Redis-backed order book per symbol using sorted sets
  - Price-time priority: best price first, FIFO within same price level
  - Handles full fills, partial fills, IOC orders
  - Circuit breaker: halts symbol on 10% price move, emits `circuit.breaker.triggered`
  - Emits `trade.executed` to Kafka on every match
  - Schema: `orders` table in `matching_engine` database
  - gRPC reflection enabled
  - Full end-to-end trade tested and verified: sell order rested, buy order matched, status `filled`

---

## Up Next

### Clearing House `:8084`

Pure Kafka consumer + producer service. Consumes `trade.executed`, steps in as central counterparty (novation), verifies locked collateral on both sides, emits `trade.cleared`.

Build order:

1. `go.mod` — kafka, logger, lib/pq, godotenv, uuid
2. Database schema — `cleared_trades` table
3. `internal/db/models.go` + `internal/db/queries.go`
4. `internal/novation/novation.go` — CCP logic, verifies both sides have active locks
5. `internal/netting/netting.go` — multilateral netting engine (aggregates positions end of day)
6. `internal/kafka/consumer.go` — consumes `trade.executed`, calls novation, publishes `trade.cleared`
7. `cmd/server/main.go` — wire everything, Kafka consumer only, no REST, no gRPC
8. `.env` — DATABASE_URL, KAFKA_BROKERS, LOG_LEVEL

Key logic:

- On every `trade.executed` event: verify buyer lock and seller lock both exist and are active in risk_engine db
- Create a cleared_trade record with ESX Clearing House as central counterparty
- Emit `trade.cleared` with all four IDs: buyer, seller, buy_lock_id, sell_lock_id

---

## Service Port Map

| Service              | HTTP  | gRPC  |
| -------------------- | ----- | ----- |
| Order Gateway        | :8080 | —     |
| Participant Registry | :8081 | :9091 |
| Risk Engine          | —     | :9093 |
| Matching Engine      | —     | :9094 |
| Clearing House       | —     | —     |
| Market Data Feed     | :8085 | —     |
| Settlement Engine    | —     | —     |
| Ledger Service       | :8087 | —     |

---

## Remaining Phases

| Phase    | Focus                                                             | Status      |
| -------- | ----------------------------------------------------------------- | ----------- |
| Phase 1  | Participant Registry, Risk Engine, Matching Engine                | ✅ Complete |
| Phase 2  | Clearing House, Settlement Engine, Ledger Service, full Kafka e2e | In Progress |
| Phase 3  | Order Gateway — FIX Protocol 4.2 parser                           | Planned     |
| Phase 4  | Market Data Feed — WebSocket streaming, circuit breakers          | Planned     |
| Phase 5  | Integration tests, k6 load testing                                | Planned     |
| Phase 6  | Dockerize all services, Kubernetes manifests, Helm charts         | Planned     |
| Phase 7  | AWS with Terraform — EKS, RDS, MSK, ElastiCache                   | Planned     |
| Phase 8  | CI/CD — GitHub Actions + ArgoCD                                   | Planned     |
| Phase 9  | Observability — Prometheus, Grafana, OpenTelemetry, Jaeger        | Planned     |
| Phase 10 | Frontend — Next.js + TradingView Lightweight Charts               | Planned     |

---

## Key Decisions Made

| Decision                | Choice                                            | Reason                                                              |
| ----------------------- | ------------------------------------------------- | ------------------------------------------------------------------- |
| Kafka mode              | KRaft (no Zookeeper)                              | Modern, no extra infra overhead                                     |
| Kafka image             | confluentinc/cp-kafka:7.6.0                       | Bitnami tags were unavailable                                       |
| Kafka client            | segmentio/kafka-go                                | Closest to KafkaJS, pure Go, no CGo                                 |
| Kafka broker resolution | `127.0.0.1 kafka` in /etc/hosts                   | Single broker address works both inside Docker and on host          |
| Log format              | JSON everywhere                                   | Standard for Prometheus, Grafana, log aggregators                   |
| Observability           | Self-hosted Prometheus + Grafana + Jaeger         | Production signal, free, what real exchanges run                    |
| Postgres port           | 5433 (host) → 5432 (container)                    | Avoid conflict with local Postgres                                  |
| Docker network          | named `esx` bridge network                        | Clean service isolation                                             |
| Amounts                 | Stored as integers in smallest unit (paise/cents) | Avoids floating point rounding bugs                                 |
| .env loading            | godotenv in all services                          | Simple, no external config service needed in dev                    |
| proto go.mod            | separate go.mod in packages/proto                 | Required for replace directives in service go.mod files             |
| Risk engine DB access   | Direct connection to participant_registry DB      | Avoids gRPC round-trip on hot path, risk checks must be synchronous |
| gRPC reflection         | Enabled on all gRPC services                      | Allows grpcurl testing without proto files                          |

---

## Conventions

- No comments in code
- No comments in yaml/config files
- All monetary amounts as integers in smallest currency unit (paise for INR)
- One Postgres database per service
- Services communicate via Kafka (async) or gRPC (sync, blocking)
- gRPC only for paths that must block: auth, risk checks, order forwarding
- Kafka for everything else
- Each service loads its own .env via godotenv
- Internal db layer is hand-written sqlc-style (models.go + queries.go) — no codegen step needed
- Service structure: cmd/server/main.go, internal/{domain}/_, db/migrations/_.sql
- All gRPC services have reflection enabled
- No REST on pure backend services (risk-engine, matching-engine, clearing-house, settlement-engine)

## Test Participants (local dev)

| Name       | participant_id                       | Notes                        |
| ---------- | ------------------------------------ | ---------------------------- |
| Trader One | 86303d8d-4429-41de-9a03-66c72d3fe06e | buyer, 1000000 paise cash    |
| Trader Two | 7864c733-a3ff-4d9e-a36c-1af290216c1d | seller, 1000 RELIANCE shares |

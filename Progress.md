# ESX — Progress Tracker

## Status: Phase 1 — Core Services (In Progress)

---

## Completed

### Infrastructure

- [x] `docker-compose.yml` — 7 containers on named `esx` network
  - Postgres 16 → `localhost:5433`
  - Redis 7 → `localhost:6379`
  - Kafka KRaft (confluentinc/cp-kafka:7.6.0) → `localhost:9092`
  - Kafka UI → `localhost:9080`
  - Prometheus → `localhost:9090`
  - Grafana → `localhost:3001` (admin/admin)
  - Jaeger → `localhost:16686`
- [x] `infra/postgres/init.sql` — creates all 8 service databases on first boot
- [x] `infra/observability/prometheus/prometheus.yml` — scrapes all 8 services via host.docker.internal

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

- [x] `services/risk-engine` — pre-trade collateral validation, locking `:9092`
  - gRPC: `CheckAndLock` — validates and locks cash (BUY) or shares (SELL)
  - gRPC: `ReleaseLock` — releases lock on cancel, consumes on fill
  - Schema: `locks` table with atomic DB transactions
  - BUY: verifies `cash_accounts.balance - locked >= price * quantity`, locks cash
  - SELL: verifies `securities_accounts.quantity - locked >= quantity`, locks shares
  - All lock operations are atomic — no partial state possible
  - `.env` loaded via `godotenv`

---

## Up Next

### Matching Engine `:8083`

Pure gRPC + Kafka service. Implements `SubmitOrder`, `CancelOrder` from `packages/proto/matching.proto`. Redis-backed order book per symbol, price-time priority matching, circuit breaker, emits `trade.executed` to Kafka.

Build order:

1. `go.mod` — grpc, kafka-go, redis, lib/pq, zerolog, godotenv, uuid
2. Database schema — `orders` table
3. `internal/db/models.go` + `internal/db/queries.go`
4. `internal/orderbook/orderbook.go` — Redis-backed order book per symbol
5. `internal/matching/matching.go` — price-time priority matching logic
6. `internal/circuit/circuit.go` — 10% price movement in 60s triggers halt
7. `internal/kafka/producer.go` — emits `trade.executed`
8. `internal/grpc/server.go` — implements SubmitOrder + CancelOrder
9. `cmd/server/main.go` — wire everything, gRPC on `:9093`, no REST
10. `.env` — DATABASE_URL, REDIS_URL, KAFKA_BROKERS, LOG_LEVEL

Key logic:

- Order book maintained in Redis per symbol — sorted sets for bids and asks
- Price-time priority: best price first, FIFO within same price level
- MARKET orders walk the book until filled or book is exhausted
- LIMIT orders rest in the book if not immediately matchable
- Partial fills emit `order.partially_filled` to Kafka, remainder stays in book
- Circuit breaker monitors 60s rolling window — halts symbol if price moves >10%
- Every match emits `trade.executed` with both participant IDs, symbol, price, quantity

---

## Remaining Phases

| Phase    | Focus                                                             | Status      |
| -------- | ----------------------------------------------------------------- | ----------- |
| Phase 1  | Participant Registry, Risk Engine, Matching Engine                | In Progress |
| Phase 2  | Clearing House, Settlement Engine, Ledger Service, full Kafka e2e | Planned     |
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

| Decision       | Choice                                            | Reason                                                  |
| -------------- | ------------------------------------------------- | ------------------------------------------------------- |
| Kafka mode     | KRaft (no Zookeeper)                              | Modern, no extra infra overhead                         |
| Kafka image    | confluentinc/cp-kafka:7.6.0                       | Bitnami tags were unavailable                           |
| Kafka client   | segmentio/kafka-go                                | Closest to KafkaJS, pure Go, no CGo                     |
| Log format     | JSON everywhere                                   | Standard for Prometheus, Grafana, log aggregators       |
| Observability  | Self-hosted Prometheus + Grafana + Jaeger         | Production signal, free, what real exchanges run        |
| Postgres port  | 5433 (host) → 5432 (container)                    | Avoid conflict with local Postgres                      |
| Docker network | named `esx` bridge network                        | Clean service isolation                                 |
| Amounts        | Stored as integers in smallest unit (paise/cents) | Avoids floating point rounding bugs                     |
| .env loading   | godotenv in all services                          | Simple, no external config service needed in dev        |
| proto go.mod   | separate go.mod in packages/proto                 | Required for replace directives in service go.mod files |

---

## Conventions

- No comments in code
- No comments in yaml/config files
- All monetary amounts as integers in smallest currency unit
- One Postgres database per service
- Services communicate via Kafka (async) or gRPC (sync, blocking)
- gRPC only for paths that must block: auth, risk checks, order forwarding
- Kafka for everything else
- Each service loads its own .env via godotenv
- Internal db layer is hand-written sqlc-style (models.go + queries.go) — no codegen step needed

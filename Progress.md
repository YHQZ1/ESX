# ESX — Progress Tracker

## Status: Phase 2 — Complete. Phase 3 — Up Next.

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
- [x] `packages/kafka` — segmentio/kafka-go producer/consumer wrapper
- [x] `packages/proto` — gRPC proto definitions + generated Go stubs
  - `participant.proto` → `ValidateAPIKey` RPC
  - `risk.proto` → `CheckAndLock`, `ReleaseLock` RPCs
  - `matching.proto` → `SubmitOrder`, `CancelOrder` RPCs

### Services

- [x] `services/participant-registry` — identity, API keys, accounts `:8081` + gRPC `:9091`
  - REST: `POST /participants/register`, `POST /participants/:id/deposit`, `GET /participants/:id`
  - gRPC: `ValidateAPIKey`
  - Schema: `participants`, `api_keys`, `cash_accounts`, `securities_accounts`
  - API keys: `crypto/rand` generated, SHA-256 hashed, raw key returned once only
  - All endpoints tested and working

- [x] `services/risk-engine` — pre-trade collateral validation, locking gRPC `:9093`
  - gRPC: `CheckAndLock`, `ReleaseLock`
  - BUY: verifies `cash_accounts.balance - locked >= price * quantity`, locks cash
  - SELL: verifies `securities_accounts.quantity - locked >= quantity`, locks shares
  - Reads directly from `participant_registry` database
  - All cases tested and verified

- [x] `services/matching-engine` — order book, price-time priority matching gRPC `:9094`
  - gRPC: `SubmitOrder`, `CancelOrder`
  - Redis-backed order book per symbol using sorted sets
  - Price-time priority: best price first, FIFO within same price level
  - Handles full fills, partial fills, IOC orders
  - Circuit breaker: halts symbol on 10% price move
  - Emits `trade.executed` to Kafka on every match
  - Full end-to-end trade tested and verified

- [x] `services/clearing-house` — novation, trade guarantee `:8084`
  - Consumes `trade.executed` from Kafka
  - Verifies both buy and sell locks are active in risk_engine db
  - Creates `cleared_trade` record as central counterparty
  - Emits `trade.cleared` to Kafka
  - All trades verified in cleared_trades table

- [x] `services/settlement-engine` — atomic DvP settlement `:8086`
  - Consumes `trade.cleared` from Kafka
  - Atomic 4-way transaction: debit buyer cash, credit seller cash, debit seller shares, credit buyer shares
  - Updates lock status to `consumed` in risk_engine db
  - Emits `trade.settled` to Kafka
  - Cash and shares verified moving correctly in participant_registry db

- [x] `services/ledger-service` — double-entry bookkeeping `:8087`
  - Consumes `trade.settled` from Kafka
  - Writes 4 journal entries per trade (2 cash + 2 securities)
  - REST: `GET /ledger/:id/balance`, `/positions`, `/cash-transactions`, `/securities-transactions`
  - All endpoints tested and returning correct data

---

## Up Next

### Order Gateway `:8080`

The public entry point for all market participants. Implements FIX Protocol 4.2 parser alongside a REST API. Authenticates participants via gRPC to Participant Registry, runs pre-trade risk checks via gRPC to Risk Engine, forwards validated orders to Matching Engine.

Build order:

1. `go.mod` — Gin, grpc, godotenv, lib/pq, uuid
2. `internal/fix/parser.go` — FIX 4.2 message parser (tag=value format)
3. `internal/fix/types.go` — FIX message types and constants
4. `internal/middleware/auth.go` — API key extraction and validation via gRPC to Participant Registry
5. `internal/client/registry.go` — gRPC client for Participant Registry
6. `internal/client/risk.go` — gRPC client for Risk Engine
7. `internal/client/matching.go` — gRPC client for Matching Engine
8. `internal/handlers/orders.go` — REST order submission handler
9. `internal/handlers/fix.go` — FIX order submission handler
10. `cmd/server/main.go` — wire everything, HTTP on `:8080`
11. `.env` — DATABASE_URL, PORT, PARTICIPANT_REGISTRY_ADDR, RISK_ENGINE_ADDR, MATCHING_ENGINE_ADDR

Key logic:

- Every inbound order (REST or FIX) goes through the same pipeline:
  1. Parse and validate format
  2. Authenticate API key via gRPC to Participant Registry → get participant_id
  3. Run risk check via gRPC to Risk Engine → get lock_id or rejection
  4. Forward to Matching Engine via gRPC → get order_id and status
  5. Return execution report to caller
- FIX 4.2 message format: `8=FIX.4.2|9=<len>|35=<type>|...|10=<checksum>`
- Support order types: MARKET, LIMIT, STOP
- Support time in force: GTC, IOC

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

| Phase    | Focus                                                      | Status      |
| -------- | ---------------------------------------------------------- | ----------- |
| Phase 1  | Participant Registry, Risk Engine, Matching Engine         | ✅ Complete |
| Phase 2  | Clearing House, Settlement Engine, Ledger Service          | ✅ Complete |
| Phase 3  | Order Gateway — FIX Protocol 4.2 parser                    | In Progress |
| Phase 4  | Market Data Feed — WebSocket streaming, circuit breakers   | Planned     |
| Phase 5  | Integration tests, k6 load testing                         | Planned     |
| Phase 6  | Dockerize all services, Kubernetes manifests, Helm charts  | Planned     |
| Phase 7  | AWS with Terraform — EKS, RDS, MSK, ElastiCache            | Planned     |
| Phase 8  | CI/CD — GitHub Actions + ArgoCD                            | Planned     |
| Phase 9  | Observability — Prometheus, Grafana, OpenTelemetry, Jaeger | Planned     |
| Phase 10 | Frontend — Next.js + TradingView Lightweight Charts        | Planned     |

---

## Key Decisions Made

| Decision                | Choice                                                                                     | Reason                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------- |
| Kafka mode              | KRaft (no Zookeeper)                                                                       | Modern, no extra infra overhead                                                 |
| Kafka image             | confluentinc/cp-kafka:7.6.0                                                                | Bitnami tags were unavailable                                                   |
| Kafka client            | segmentio/kafka-go                                                                         | Closest to KafkaJS, pure Go, no CGo                                             |
| Kafka broker resolution | `127.0.0.1 kafka` in /etc/hosts + `KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092` | Host services use localhost:9092, Docker containers use kafka:9092 via Kafka UI |
| Log format              | JSON everywhere                                                                            | Standard for Prometheus, Grafana, log aggregators                               |
| Observability           | Self-hosted Prometheus + Grafana + Jaeger                                                  | Production signal, free, what real exchanges run                                |
| Postgres port           | 5433 (host) → 5432 (container)                                                             | Avoid conflict with local Postgres                                              |
| Docker network          | named `esx` bridge network                                                                 | Clean service isolation                                                         |
| Amounts                 | Stored as integers in smallest unit (paise/cents)                                          | Avoids floating point rounding bugs                                             |
| .env loading            | godotenv in all services                                                                   | Simple, no external config service needed in dev                                |
| proto go.mod            | separate go.mod in packages/proto                                                          | Required for replace directives in service go.mod files                         |
| Risk engine DB access   | Direct connection to participant_registry DB                                               | Avoids gRPC round-trip on hot path                                              |
| Settlement DB access    | Direct connection to participant_registry + risk_engine DB                                 | Atomic DvP requires direct DB access                                            |
| gRPC reflection         | Enabled on all gRPC services                                                               | Allows grpcurl testing without proto files                                      |

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

| Name       | participant_id                       | Notes                      |
| ---------- | ------------------------------------ | -------------------------- |
| Trader One | 86303d8d-4429-41de-9a03-66c72d3fe06e | buyer, cash balance varies |
| Trader Two | 7864c733-a3ff-4d9e-a36c-1af290216c1d | seller, RELIANCE shares    |

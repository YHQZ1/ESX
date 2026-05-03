# ESX — Progress Tracker

## Status: Phase 4 — Complete. Phase 5 — Up Next.

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
- [x] `KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092` — host services use localhost:9092

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

- [x] `services/participant-registry` `:8081` + gRPC `:9091`
  - REST: `POST /participants/register`, `POST /participants/:id/deposit`, `GET /participants/:id`
  - gRPC: `ValidateAPIKey`
  - Schema: `participants`, `api_keys`, `cash_accounts`, `securities_accounts`
  - API keys: `crypto/rand` generated, SHA-256 hashed, raw key returned once only

- [x] `services/risk-engine` gRPC `:9093`
  - gRPC: `CheckAndLock`, `ReleaseLock`
  - BUY: verifies `cash_accounts.balance - locked >= price * quantity`, locks cash
  - SELL: verifies `securities_accounts.quantity - locked >= quantity`, locks shares
  - Reads directly from `participant_registry` database

- [x] `services/matching-engine` gRPC `:9094`
  - gRPC: `SubmitOrder`, `CancelOrder`
  - Redis-backed order book per symbol using sorted sets
  - Price-time priority: best price first, FIFO within same price level
  - Handles full fills, partial fills, IOC orders
  - Circuit breaker: halts symbol on 10% price move
  - Emits `trade.executed` to Kafka on every match

- [x] `services/clearing-house`
  - Consumes `trade.executed` from Kafka
  - Verifies both buy and sell locks are active in risk_engine db
  - Creates `cleared_trade` record as central counterparty
  - Emits `trade.cleared` to Kafka

- [x] `services/settlement-engine`
  - Consumes `trade.cleared` from Kafka
  - Atomic 4-way DvP: debit buyer cash, credit seller cash, debit seller shares, credit buyer shares
  - Updates lock status to `consumed` in risk_engine db
  - Emits `trade.settled` to Kafka

- [x] `services/ledger-service` `:8087`
  - Consumes `trade.settled` from Kafka
  - Writes 4 double-entry journal entries per trade (2 cash + 2 securities)
  - REST: `GET /ledger/:id/balance`, `/positions`, `/cash-transactions`, `/securities-transactions`

- [x] `services/order-gateway` `:8080`
  - REST: `POST /orders`, `DELETE /orders/:id` with `x-api-key` auth middleware
  - FIX Protocol 4.2: `POST /fix` — full message parser and execution report builder
  - Authenticates via gRPC to Participant Registry
  - Risk checks via gRPC to Risk Engine
  - Order routing via gRPC to Matching Engine
  - Both REST and FIX tested end to end with full fill

- [x] `services/market-data-feed` `:8085`
  - WebSocket server — clients connect to `ws://localhost:8085/ws`
  - Subscribe/unsubscribe via JSON: `{"action": "subscribe", "channel": "trades.RELIANCE"}`
  - Consumes `trade.executed` from Kafka
  - Broadcasts to `trades.{symbol}` and `ticker.{symbol}` channels
  - Live broadcast verified: trade received in real time via wscat

---

## Up Next

### Phase 5 — Integration Tests + k6 Load Testing

Build order:

1. `infra/k6/orderflow.js` — full order lifecycle load test
   - Register participants, fund accounts, submit buy/sell pairs
   - Measure: order submission latency, matching throughput, end-to-end trade latency
2. `infra/k6/matching.js` — matching engine stress test
   - Flood the book with resting orders at multiple price levels
   - Measure: orders/sec, fill latency under load
3. Go integration tests per service
   - `participant-registry`: register, deposit, validate API key
   - `risk-engine`: CheckAndLock approved + rejected cases
   - `matching-engine`: full fill, partial fill, IOC cancel, resting order

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
| Phase 3  | Order Gateway — FIX Protocol 4.2 parser                    | ✅ Complete |
| Phase 4  | Market Data Feed — WebSocket streaming                     | ✅ Complete |
| Phase 5  | Integration tests, k6 load testing                         | Up Next     |
| Phase 6  | Dockerize all services, Kubernetes manifests, Helm charts  | Planned     |
| Phase 7  | AWS with Terraform — EKS, RDS, MSK, ElastiCache            | Planned     |
| Phase 8  | CI/CD — GitHub Actions + ArgoCD                            | Planned     |
| Phase 9  | Observability — Prometheus, Grafana, OpenTelemetry, Jaeger | Planned     |
| Phase 10 | Frontend — Next.js + TradingView Lightweight Charts        | Planned     |

---

## Key Decisions Made

| Decision                | Choice                                                     | Reason                                                  |
| ----------------------- | ---------------------------------------------------------- | ------------------------------------------------------- |
| Kafka mode              | KRaft (no Zookeeper)                                       | Modern, no extra infra overhead                         |
| Kafka image             | confluentinc/cp-kafka:7.6.0                                | Bitnami tags were unavailable                           |
| Kafka client            | segmentio/kafka-go                                         | Closest to KafkaJS, pure Go, no CGo                     |
| Kafka broker resolution | `KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092`   | Host services and Docker UI both work on localhost:9092 |
| Log format              | JSON everywhere                                            | Standard for Prometheus, Grafana, log aggregators       |
| Observability           | Self-hosted Prometheus + Grafana + Jaeger                  | Production signal, free, what real exchanges run        |
| Postgres port           | 5433 (host) → 5432 (container)                             | Avoid conflict with local Postgres                      |
| Docker network          | named `esx` bridge network                                 | Clean service isolation                                 |
| Amounts                 | Stored as integers in smallest unit (paise/cents)          | Avoids floating point rounding bugs                     |
| .env loading            | godotenv in all services                                   | Simple, no external config service needed in dev        |
| proto go.mod            | separate go.mod in packages/proto                          | Required for replace directives in service go.mod files |
| Risk engine DB access   | Direct connection to participant_registry DB               | Avoids gRPC round-trip on hot path                      |
| Settlement DB access    | Direct connection to participant_registry + risk_engine DB | Atomic DvP requires direct DB access                    |
| gRPC reflection         | Enabled on all gRPC services                               | Allows grpcurl testing without proto files              |
| WebSocket subscriptions | JSON subscribe/unsubscribe messages                        | Simple, works with any WS client                        |

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

| Name       | participant_id                       | API Key                                                          | Notes                       |
| ---------- | ------------------------------------ | ---------------------------------------------------------------- | --------------------------- |
| Trader One | 86303d8d-4429-41de-9a03-66c72d3fe06e | 9e214dc45c6db75645a1598bd60ec875db9192ed81f9da23c4093c4ea3af96bd | buyer                       |
| Trader Two | 7864c733-a3ff-4d9e-a36c-1af290216c1d | d2f36fa3c1544c57e16c024fd48435b8bb627950f8727d06cb6ba168c0e50cd6 | seller, has RELIANCE shares |

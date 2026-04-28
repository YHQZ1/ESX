<div align="center">

<img src="https://cdn.brandfetch.io/idOOGS0oFC/idN3Ai6Nko.svg?c=1bxid64Mup7aczewSAYMX&t=1651825673631" height="52" alt="NASDAQ" style="vertical-align: middle;" />
&nbsp;&nbsp;&nbsp;&nbsp;
<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/5/50/BSE_logo.svg/3840px-BSE_logo.svg.png" height="72" alt="BSE" style="vertical-align: middle;" />
&nbsp;&nbsp;&nbsp;&nbsp;
<img src="https://cdn.brandfetch.io/idLb1SWTJA/theme/dark/logo.svg?c=1bxid64Mup7aczewSAYMX&t=1746508392065" height="72" alt="NSE" style="vertical-align: middle;" />

<h1>ESX — Escrow Stock Exchange</h1>

<p><strong>A production-grade securities exchange infrastructure, built from scratch.</strong><br />
The engine that powers trade matching, clearing, and settlement —<br />
the same core systems that run NASDAQ, NSE, and BSE, rebuilt from the ground up.</p>

<br />

<img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go&logoColor=white" />
<img src="https://img.shields.io/badge/Apache_Kafka-KRaft-231F20?style=flat-square&logo=apachekafka&logoColor=white" />
<img src="https://img.shields.io/badge/Redis-7-DC382D?style=flat-square&logo=redis&logoColor=white" />
<img src="https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat-square&logo=postgresql&logoColor=white" />
<img src="https://img.shields.io/badge/gRPC-Protocol_Buffers-244c5a?style=flat-square" />
<img src="https://img.shields.io/badge/Kubernetes-326CE5?style=flat-square&logo=kubernetes&logoColor=white" />
<img src="https://img.shields.io/badge/AWS-EKS_·_RDS_·_MSK-FF9900?style=flat-square&logo=amazonaws&logoColor=white" />
<img src="https://img.shields.io/badge/Terraform-IaC-7B42BC?style=flat-square&logo=terraform&logoColor=white" />
<img src="https://img.shields.io/badge/FIX_Protocol-4.2-000000?style=flat-square" />

<br />

<p>
  <a href="#what-esx-is">What ESX is</a> &nbsp;·&nbsp;
  <a href="#architecture">Architecture</a> &nbsp;·&nbsp;
  <a href="#trade-lifecycle">Trade Lifecycle</a> &nbsp;·&nbsp;
  <a href="#services">Services</a> &nbsp;·&nbsp;
  <a href="#engineering-concepts">Engineering Concepts</a> &nbsp;·&nbsp;
  <a href="#kafka-events">Kafka Events</a> &nbsp;·&nbsp;
  <a href="#getting-started">Getting Started</a> &nbsp;·&nbsp;
  <a href="#project-structure">Project Structure</a> &nbsp;·&nbsp;
  <a href="#roadmap">Roadmap</a>
</p>

</div>

---

<h2 id="what-esx-is">What ESX is</h2>

ESX is a **securities exchange infrastructure** — not a trading app, not a brokerage wrapper. There are no calls to any external exchange or market data provider. ESX owns the full trade lifecycle end to end.

Almost every developer has interacted with the surface of a stock exchange — placing orders through an app, watching prices move in real time. Almost none have seen what runs underneath. ESX is that underneath: the matching engine that pairs buyers with sellers in microseconds, the clearing house that guarantees every trade, the settlement engine that atomically transfers cash and shares, and the ledger that records every movement with double-entry precision.

This is the same class of system that powers **NASDAQ**, **NSE**, **BSE**, and every other major exchange in the world — built from scratch, documented openly, and deployed on production-grade cloud infrastructure. The design is exchange-agnostic: the same architecture applies whether the underlying market is equities, derivatives, or currency.

**ESX owns the full stack:**

- Order intake, validation, and routing via FIX Protocol 4.2
- Pre-trade risk checks and real-time collateral locking
- Order book management and trade matching via price-time priority
- Real-time market data dissemination over WebSockets
- Trade guarantee via central counterparty clearing (novation)
- Atomic delivery-versus-payment settlement (DvP)
- Double-entry bookkeeping across cash and securities dimensions
- Nightly reconciliation and ledger integrity verification

---

<h2 id="architecture">Architecture</h2>

ESX is a fully event-driven microservices system. Eight independent services communicate over **Kafka** and **gRPC**, each owning its own PostgreSQL database and bounded domain. No shared state. No coupled services. Every component can be scaled, deployed, and reasoned about independently.

```
                    +-------------------------------------------------+
                    |             Market Participants                 |
                    |   Brokers · Traders · Algorithms · HFT Firms    |
                    +------------------------+------------------------+
                                             |  FIX Protocol 4.2 / REST
                                    +--------v---------+
                                    |   Order Gateway  |  :8080
                                    |   FIX · REST     |
                                    +--+------+-----+--+
                               gRPC    |      |     |   gRPC
                  +--------------------+      |     +-------------------+
                  v                           v                         v
   +---------------------+    +---------------------+    +---------------------+
   | Participant Registry|    |    Risk Engine      |    |  Matching Engine    |
   |        :8081        |    |       :8082         |    |      :8083          |
   +---------------------+    +---------------------+    +----------+----------+
                                                                    |
                                                    Kafka ----------+
                                             trade.executed         |
                                    +-------------------------------+
                                    |                               |
                       +------------v-----------+      +------------v----------+
                       |     Clearing House     |      |   Market Data Feed    |
                       |         :8084          |      |        :8085          |
                       +------------+-----------+      +------------+----------+
                                    | Kafka                          | WebSocket
                            trade.cleared                            v
                       +------------v-----------+         +-------------------+
                       |   Settlement Engine    |         |  Trading Terminal |
                       |        :8086           |         |    (Frontend)     |
                       +------------+-----------+         +-------------------+
                                    | Kafka
                            trade.settled
                       +------------v-----------+
                       |     Ledger Service     |
                       |         :8087          |
                       +------------------------+
```

<br />

**Communication patterns:**

| From              | To                   | Protocol            | Purpose                                  |
| ----------------- | -------------------- | ------------------- | ---------------------------------------- |
| External clients  | Order Gateway        | FIX 4.2 / REST      | Order submission                         |
| Order Gateway     | Participant Registry | gRPC (sync)         | API key authentication                   |
| Order Gateway     | Risk Engine          | gRPC (sync)         | Pre-trade collateral check               |
| Order Gateway     | Matching Engine      | gRPC (sync)         | Validated order forwarding               |
| Matching Engine   | Kafka                | Event               | `trade.executed` on every match          |
| Clearing House    | Kafka                | Consumer + Producer | Trade guarantee and multilateral netting |
| Settlement Engine | Kafka                | Consumer + Producer | Atomic DvP asset transfer                |
| Ledger Service    | Kafka                | Consumer            | Double-entry bookkeeping                 |
| Market Data Feed  | WebSocket            | Push                | Live prices and order book depth         |

---

<h2 id="trade-lifecycle">The Full Trade Lifecycle</h2>

This is what ESX does every time a participant places an order — from submission to final settlement. Every step is a service with its own database and its own bounded responsibility.

```
  PARTICIPANT SUBMITS ORDER (FIX Protocol / REST)
          |
          v
  +------------------+
  |  Order Gateway   |  Parses FIX message. Validates format and fields.
  +--------+---------+
           | gRPC (sync)
           v
  +----------------------+
  | Participant Registry |  Verifies API key. Returns participant_id.
  +--------+-------------+  Rejects unknown or suspended participants.
           | gRPC (sync)
           v
  +------------------+
  |   Risk Engine    |  BUY  -> verify cash_balance >= price x quantity. Lock cash.
  |                  |  SELL -> verify share_position >= quantity. Lock shares.
  +--------+---------+  Rejects if collateral is insufficient.
           | gRPC (sync)
           v
  +------------------+
  | Matching Engine  |  Order enters the order book.
  |                  |  Price-time priority matching fires against resting orders.
  |                  |  Full fill, partial fill, or order rests in book.
  |                  |  Emits trade.executed (Kafka)
  +--------+---------+
           |
    +------+------------------+
    |                         |
    v                         v
+--------+         +--------------------+
| Market |         |  Clearing House    |  Steps in as central counterparty (novation).
|  Data  |         |                    |  Verifies locked collateral on both sides.
|  Feed  |         |                    |  Performs multilateral netting end-of-day.
|        |         |                    |  Emits trade.cleared (Kafka)
+--------+         +--------------------+
(broadcasts                  |
 new price                   v
 over WS)        +--------------------+
                 | Settlement Engine  |  Delivery versus Payment (DvP).
                 |                    |  Atomic 4-way DB transaction:
                 |                    |    DEBIT  buyer cash      -X
                 |                    |    CREDIT seller cash     +X
                 |                    |    DEBIT  seller shares   -N
                 |                    |    CREDIT buyer shares    +N
                 |                    |  Emits trade.settled (Kafka)
                 +--------+-----------+
                          |
                          v
                 +--------------------+
                 |   Ledger Service   |  Writes 4 double-entry journal entries.
                 |                    |  cash_journal + securities_journal.
                 |                    |  All entries net to zero. Always.
                 |                    |  Nightly reconciliation verifies integrity.
                 +--------------------+

  TRADE COMPLETE
```

---

<h2 id="services">Services</h2>

### Order Gateway &nbsp;`:8080`

The public entry point for all market participants. Implements a **FIX Protocol 4.2 parser** alongside a REST API. Every inbound request is authenticated and risk-checked synchronously before the order is forwarded to the Matching Engine. The gateway is the only service exposed to external traffic — all other services are internal.

**Responsibilities:**

- Parse and validate FIX 4.2 messages
- Authenticate participants via gRPC to Participant Registry
- Run synchronous pre-trade risk checks via gRPC to Risk Engine
- Forward validated orders to the Matching Engine
- Return execution reports back to the submitting participant
- Support `MARKET`, `LIMIT`, and `STOP` order types
- Support `Immediate-or-Cancel (IOC)` execution instructions

---

### Participant Registry &nbsp;`:8081`

The identity authority for the entire system. Every other service traces authentication back here. No participant interacts with any ESX service without first being validated through the registry.

**Responsibilities:**

- Participant registration and account provisioning
- API key generation (cryptographically random, hashed before storage)
- Cash account and securities account management
- Margin and collateral account tracking
- Internal API key validation endpoint called on every inbound order

Internal endpoints (`/internal/*`) are protected by a shared `x-internal-token` header in Phase 1. In production, mTLS between services replaces this entirely.

---

### Risk Engine &nbsp;`:8082`

The pre-trade gatekeeper. No order reaches the order book without explicit approval from the Risk Engine. Runs synchronously over gRPC — the Order Gateway blocks until a response is received. There is intentionally no Kafka in this path. A failed risk check must stop the order immediately, not eventually.

**Responsibilities:**

- **Buy orders**: verify `cash_balance >= (price x quantity)`, lock required cash
- **Sell orders**: verify `share_position >= quantity`, lock required shares
- Release locks on order cancellation or expiry
- Partial lock release on partial fills
- Reject insufficient collateral before orders enter the book

---

### Matching Engine &nbsp;`:8083`

The heart of ESX. Written in Go for microsecond-level performance. Maintains a live order book per listed security in Redis and applies strict price-time priority matching against every incoming order.

**Responsibilities:**

- Maintain a live order book per security (Redis-backed for in-memory speed)
- Match incoming orders against resting orders by price-time priority
- Handle full fills, partial fills, and resting limit orders
- Walk the book across price levels for market orders
- Trigger circuit breakers when price moves more than 10% in 60 seconds
- Emit `trade.executed` to Kafka on every match

**Order book structure:**

```
ASK (sellers — ascending price)
  505.00 | 200 shares  <- best ask
  502.00 | 150 shares
  501.00 |  80 shares
  ---------------------- spread (2.00)
  499.00 | 120 shares
  498.00 | 300 shares
  495.00 | 500 shares  <- worst bid
BID (buyers — descending price)
```

---

### Market Data Feed &nbsp;`:8085`

Real-time broadcast of all market activity to connected clients over WebSockets. Consumes `trade.executed` from Kafka and pushes structured updates within milliseconds of every execution.

**WebSocket channels:**

| Channel                             | Description                                               |
| ----------------------------------- | --------------------------------------------------------- |
| `orderbook.{symbol}`                | Full order book depth — all bids and asks with quantities |
| `trades.{symbol}`                   | Live trade feed — price, quantity, side, timestamp        |
| `ticker.{symbol}`                   | Best bid, best ask, last traded price, 24h volume         |
| `candles.{symbol}.{1m\|5m\|1h\|1d}` | OHLCV candlestick data                                    |

---

### Clearing House &nbsp;`:8084`

Guarantees every trade through **novation** — stepping in as the central counterparty to both sides of every executed trade. This is how real clearing houses operate: DTCC in the US, NSCCL in India (NSE's clearing arm), and ICCL for BSE. Bilateral counterparty risk is eliminated entirely.

**Responsibilities:**

- Become central counterparty to both sides on every `trade.executed` event
- Verify locked collateral exists on both sides before clearing
- Perform **multilateral netting** at end of day, drastically reducing settlement volume
- Absorb and unwind failed or defaulted trades
- Emit `trade.cleared` to Kafka once the trade is guaranteed

**Novation:**

```
Before:   Buyer <------------------------------> Seller
After:    Buyer <--> ESX Clearing House <--> Seller
```

**Multilateral netting:**

```
Without:  3 buys + 2 sells for same security  =  5 settlements
With:     net position calculated once         =  1 settlement
```

---

### Settlement Engine &nbsp;`:8086`

The final step. Implements **Delivery versus Payment (DvP)** — cash and securities transfer simultaneously or not at all. All four ledger movements execute inside a single atomic database transaction. It is structurally impossible for a buyer to pay without receiving shares, or a seller to deliver shares without receiving payment.

**Settlement modes:**

- `T+1` — standard, settles next business day (the convention used by NSE, BSE, and most global markets post-2024)
- `INSTANT` — immediate settlement for testing and simulation

**The four atomic operations per trade:**

```
1. DEBIT  buyer cash account      -50,000
2. CREDIT seller cash account     +50,000
3. DEBIT  seller share position   -100 shares
4. CREDIT buyer share position    +100 shares
----------------------------------------------
   Net cash:    0
   Net shares:  0
```

---

### Ledger Service &nbsp;`:8087`

The financial source of truth for ESX. Implements **double-entry bookkeeping** across both cash and securities dimensions. Every financial event produces balanced journal entries that net to zero. The ledger cannot be in an inconsistent state.

**Double-entry example:**

```
trade.settled — 100 shares @ 500 — Participant A (buyer) vs Participant B (seller)

cash_journal:
  DEBIT  | cash | participant_a | -50,000
  CREDIT | cash | participant_b | +50,000
  ----------------------------------------
  Net: 0

securities_journal:
  DEBIT  | shares | participant_b | -100
  CREDIT | shares | participant_a | +100
  ----------------------------------------
  Net: 0
```

**APIs:**

```
GET /ledger/balance          — current cash balance
GET /ledger/positions        — all securities holdings
GET /ledger/transactions     — cursor-paginated trade and transfer history
GET /ledger/locked           — currently reserved collateral breakdown
```

A nightly reconciliation cron validates: all debits match credits, every settled trade has exactly four entries, no balance is negative, no orphaned locks exist.

---

<h2 id="engineering-concepts">Key Engineering Concepts</h2>

### Price-Time Priority

The universal matching rule of all order-driven exchanges. Orders are matched by best price first — lowest ask for buyers, highest bid for sellers. When two orders share the same price, the one submitted earliest is matched first (strict FIFO). This ensures fairness and prevents front-running within a price level. Every major exchange — NASDAQ, NSE, BSE, LSE — uses this rule.

### FIX Protocol

Financial Information eXchange (FIX) is a 40-year-old message standard that still carries the majority of global electronic trading volume. Every institutional broker, trading algorithm, and exchange worldwide speaks FIX. By implementing a FIX 4.2 parser in the Order Gateway, any real-world trading system can connect to ESX without modification.

### Novation

When the Clearing House steps between buyer and seller after a trade executes, it replaces the original bilateral contract with two new contracts — one between the buyer and the Clearing House, and one between the seller and the Clearing House. The original relationship is extinguished. Neither party has exposure to the other — only to the Clearing House. This is how DTCC operates in the US, NSCCL operates for NSE, and ICCL operates for BSE.

### Delivery versus Payment (DvP)

The settlement principle that cash and securities must transfer simultaneously. The atomic database transaction in the Settlement Engine enforces this at the application layer — all four movements commit together or all four roll back. There is no state in which a buyer pays without receiving shares, or a seller delivers shares without receiving payment. This eliminates principal risk entirely.

### Double-Entry Bookkeeping

Every financial movement creates two equal and opposite journal entries that sum to zero. The `cash_journal` and `securities_journal` can never be in a state where debits do not match credits. This is enforced at write time and verified mathematically by a nightly reconciliation job that scans every trade and every entry.

### Circuit Breakers

When the Matching Engine detects a security's price has moved more than 10% within a 60-second rolling window, it halts trading for that security and emits a `circuit.breaker.triggered` event to Kafka. All services consume this and halt related processing. This mirrors the mechanism NSE uses (dynamic price bands), BSE uses (price band filters), and NASDAQ uses (Limit Up-Limit Down) to prevent flash crashes.

### Multilateral Netting

Rather than settling each trade individually, the Clearing House aggregates all of a participant's trades across a session and calculates a net position. A participant who bought 300 shares and sold 200 shares of the same security settles only 100 shares net — one settlement instead of five. This reduces settlement volume by orders of magnitude at scale and is how every real clearing house operates.

### Cursor-Based Pagination

All list endpoints in the Ledger Service use cursor-based pagination. Offset pagination degrades as the dataset grows and produces inconsistent results under concurrent writes — on a high-volume financial ledger this is unacceptable. Cursor pagination is stable, efficient at any dataset size, and correct under concurrent writes.

### Amounts in Smallest Currency Unit

All monetary values in ESX are stored and transmitted as integers in the smallest currency unit (paise for INR, cents for USD). There are no floating-point amounts anywhere in the system. `50000` means ₹500.00. This prevents the class of floating-point rounding bugs that have caused real financial losses in production systems.

---

<h2 id="kafka-events">Kafka Event Schema</h2>

| Topic                       | Producer          | Consumers                        | Trigger                                           |
| --------------------------- | ----------------- | -------------------------------- | ------------------------------------------------- |
| `order.submitted`           | Order Gateway     | Matching Engine                  | Valid order accepted by gateway                   |
| `order.cancelled`           | Matching Engine   | Risk Engine, Ledger              | Order cancelled or expired                        |
| `order.partially_filled`    | Matching Engine   | Risk Engine                      | Partial fill — remainder stays in book            |
| `trade.executed`            | Matching Engine   | Clearing House, Market Data Feed | Order matched, trade executed                     |
| `trade.cleared`             | Clearing House    | Settlement Engine                | Trade guaranteed by clearing house                |
| `trade.settled`             | Settlement Engine | Ledger Service, Market Data Feed | Cash and shares atomically transferred            |
| `risk.rejected`             | Risk Engine       | Order Gateway                    | Pre-trade check failed — insufficient collateral  |
| `circuit.breaker.triggered` | Matching Engine   | All services                     | Price moved more than 10% in 60s — trading halted |
| `circuit.breaker.lifted`    | Matching Engine   | All services                     | Cooling-off expired — trading resumed             |

All events include `participant_id`, `symbol`, `timestamp`, and event-specific fields.

---

<h2 id="getting-started">Getting Started</h2>

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- `protoc` — Protocol Buffers compiler
- `protoc-gen-go` and `protoc-gen-go-grpc` plugins

### 1. Clone the repository

```bash
git clone https://github.com/YHQZ1/esx.git
cd esx
```

### 2. Set up environment variables

Each service has a `.env.example`. Copy it to `.env`:

```bash
for svc in order-gateway participant-registry risk-engine matching-engine \
           market-data-feed clearing-house settlement-engine ledger-service; do
  cp services/$svc/.env.example services/$svc/.env
done
```

Generate a secure internal secret and set it across all `.env` files. All services must share the same value:

```bash
openssl rand -hex 32
```

### 3. Generate gRPC stubs

```bash
protoc --go_out=. --go-grpc_out=. packages/proto/*.proto
```

### 4. Start infrastructure

```bash
docker compose up -d
```

Starts PostgreSQL (one database per service), Redis, Kafka (KRaft), Kafka UI at `:9080`, Prometheus, and Grafana. Kafka takes approximately 30 seconds to become ready.

```bash
docker compose ps   # all containers should show healthy
```

### 5. Run migrations

```
coming in Phase 1
```

### 6. Start services

```
coming in Phase 1
```

### 7. Verify the full flow

**Register a participant:**

```bash
curl -X POST http://localhost:8081/participants/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Trader One", "email": "trader@example.com"}'
```

**Fund the account:**

```bash
curl -X POST http://localhost:8081/participants/{participant_id}/deposit \
  -H "x-api-key: {your_api_key}" \
  -H "Content-Type: application/json" \
  -d '{"amount": 1000000, "currency": "INR"}'
```

**Submit a limit buy order:**

```bash
curl -X POST http://localhost:8080/orders \
  -H "x-api-key: {your_api_key}" \
  -H "Content-Type: application/json" \
  -d '{"symbol": "RELIANCE", "side": "BUY", "type": "LIMIT", "quantity": 100, "price": 50000}'
```

**Observe Kafka events:**
Open `http://localhost:9080` (Kafka UI). You should see `order.submitted`, `trade.executed`, `trade.cleared`, and `trade.settled` events flowing through their topics as trades execute.

**Check ledger:**

```bash
curl http://localhost:8087/ledger/balance \
  -H "x-api-key: {your_api_key}"

curl http://localhost:8087/ledger/positions \
  -H "x-api-key: {your_api_key}"
```

---

<h2 id="project-structure">Project Structure</h2>

```
esx/
├── docker-compose.yml                        # Postgres x8, Redis, Kafka, Kafka UI, Prometheus, Grafana
├── go.work                                   # Go workspace
│
├── infra/
│   ├── postgres/
│   │   └── init.sql                          # Creates all 8 databases on first run
│   ├── terraform/                            # AWS: EKS, RDS, MSK, ElastiCache, VPC, IAM
│   └── k8s/
│       ├── manifests/                        # Raw Kubernetes manifests per service
│       └── helm/                             # Helm charts per service
│
├── packages/
│   ├── proto/                                # Shared gRPC .proto definitions
│   │   ├── participant.proto
│   │   ├── risk.proto
│   │   └── matching.proto
│   ├── kafka/                                # Shared Kafka producer/consumer wrappers
│   └── logger/                               # Structured JSON logger (zerolog)
│
└── services/
    ├── order-gateway/                        # :8080 — FIX parser, REST API, order routing
    │   ├── cmd/server/
    │   └── internal/
    │       ├── fix/                          # FIX 4.2 protocol parser
    │       ├── handlers/                     # REST order endpoints
    │       ├── middleware/                   # auth, rate limiting
    │       └── client/                       # gRPC clients for registry, risk, matching
    │
    ├── participant-registry/                 # :8081 — identity, API keys, accounts
    │   ├── cmd/server/
    │   └── internal/
    │       ├── handlers/                     # registration, deposit, account management
    │       ├── grpc/                         # internal gRPC server
    │       └── lib/                          # key generation, hashing, validation
    │
    ├── risk-engine/                          # :8082 — pre-trade validation, collateral locking
    │   ├── cmd/server/
    │   └── internal/
    │       ├── checks/                       # buy/sell validation logic
    │       ├── locks/                        # collateral locking and releasing
    │       └── grpc/                         # gRPC server
    │
    ├── matching-engine/                      # :8083 — order book, trade execution
    │   ├── cmd/server/
    │   └── internal/
    │       ├── orderbook/                    # order book data structure (Redis-backed)
    │       ├── matching/                     # price-time priority matching logic
    │       ├── circuit/                      # circuit breaker — price band monitoring
    │       └── kafka/                        # trade.executed producer
    │
    ├── market-data-feed/                     # :8085 — real-time WebSocket streaming
    │   ├── cmd/server/
    │   └── internal/
    │       ├── ws/                           # WebSocket hub and client management
    │       ├── channels/                     # orderbook, trades, ticker, candles
    │       └── kafka/                        # trade.executed consumer
    │
    ├── clearing-house/                       # :8084 — novation, netting, trade guarantee
    │   ├── cmd/server/
    │   └── internal/
    │       ├── novation/                     # central counterparty logic
    │       ├── netting/                      # multilateral netting engine
    │       └── kafka/                        # trade.executed consumer, trade.cleared producer
    │
    ├── settlement-engine/                    # :8086 — atomic DvP settlement
    │   ├── cmd/server/
    │   └── internal/
    │       ├── settlement/                   # DvP atomic transaction logic
    │       ├── scheduler/                    # T+1 settlement scheduling
    │       └── kafka/                        # trade.cleared consumer, trade.settled producer
    │
    └── ledger-service/                       # :8087 — double-entry bookkeeping
        ├── cmd/server/
        └── internal/
            ├── handlers/                     # balance, positions, transactions APIs
            ├── journal/                      # double-entry write logic
            ├── reconciliation/               # nightly cron — integrity verification
            └── kafka/                        # trade.settled consumer
```

---

## Environment Variables

### Shared (all services)

| Variable                  | Description                                                                       |
| ------------------------- | --------------------------------------------------------------------------------- |
| `INTERNAL_SERVICE_SECRET` | Shared secret for service-to-service auth. Must be identical across all services. |
| `KAFKA_BROKERS`           | Comma-separated Kafka broker addresses. Default: `localhost:9092`                 |
| `REDIS_URL`               | Redis connection URL. Default: `redis://localhost:6379`                           |

### Per-service

| Service           | Variable                         | Description                                                    |
| ----------------- | -------------------------------- | -------------------------------------------------------------- |
| All               | `PORT`                           | HTTP port the service listens on                               |
| All               | `DATABASE_URL`                   | PostgreSQL connection string for this service's database       |
| order-gateway     | `PARTICIPANT_REGISTRY_ADDR`      | gRPC address for participant registry                          |
| order-gateway     | `RISK_ENGINE_ADDR`               | gRPC address for risk engine                                   |
| order-gateway     | `MATCHING_ENGINE_ADDR`           | gRPC address for matching engine                               |
| matching-engine   | `CIRCUIT_BREAKER_THRESHOLD`      | Price movement percentage that triggers a halt (default: `10`) |
| matching-engine   | `CIRCUIT_BREAKER_WINDOW_SECONDS` | Rolling window for circuit breaker check (default: `60`)       |
| settlement-engine | `SETTLEMENT_MODE`                | `T1` or `INSTANT` (default: `T1`)                              |

---

## Tech Stack

| Layer                  | Technology                       |
| ---------------------- | -------------------------------- |
| Language               | Go 1.22                          |
| HTTP Framework         | Gin                              |
| Service Communication  | gRPC + Protocol Buffers          |
| Exchange Protocol      | FIX Protocol 4.2                 |
| Message Broker         | Apache Kafka (KRaft)             |
| Order Book State       | Redis 7                          |
| Database               | PostgreSQL 16                    |
| Query Layer            | sqlc                             |
| Containerization       | Docker + Docker Compose          |
| Orchestration          | Kubernetes + Helm                |
| GitOps                 | ArgoCD                           |
| Cloud                  | AWS (EKS, RDS, MSK, ElastiCache) |
| Infrastructure as Code | Terraform                        |
| CI/CD                  | GitHub Actions                   |
| Metrics                | Prometheus                       |
| Dashboards             | Grafana                          |
| Tracing                | OpenTelemetry + Jaeger           |
| Load Testing           | k6                               |
| Frontend               | Next.js + Tailwind CSS           |
| Charts                 | TradingView Lightweight Charts   |

---

<h2 id="roadmap">Development Roadmap</h2>

| Phase        | Focus                                                                        | Status      |
| ------------ | ---------------------------------------------------------------------------- | ----------- |
| **Phase 1**  | Participant Registry, Risk Engine, Matching Engine, Docker Compose infra     | In Progress |
| **Phase 2**  | Clearing House, Settlement Engine, Ledger Service, full Kafka event flow e2e | Planned     |
| **Phase 3**  | Order Gateway — FIX Protocol 4.2 parser, full order lifecycle integration    | Planned     |
| **Phase 4**  | Market Data Feed — WebSocket streaming, circuit breakers                     | Planned     |
| **Phase 5**  | Integration tests, k6 load testing, latency benchmarking on matching engine  | Planned     |
| **Phase 6**  | Dockerize all services, Kubernetes manifests, Helm charts                    | Planned     |
| **Phase 7**  | AWS with Terraform — EKS, RDS, MSK, ElastiCache                              | Planned     |
| **Phase 8**  | CI/CD — GitHub Actions pipelines + ArgoCD GitOps                             | Planned     |
| **Phase 9**  | Observability — Prometheus, Grafana, OpenTelemetry, Jaeger                   | Planned     |
| **Phase 10** | Frontend trading terminal — Next.js + TradingView Lightweight Charts         | Planned     |

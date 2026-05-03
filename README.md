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

<pre>
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
         |        :8081        |    |       :9093         |    |      :9094          |
         +---------------------+    +---------------------+    +----------+----------+
                                                                          |
                                                          Kafka ----------+
                                                   trade.executed         |
                                          +-------------------------------+
                                          |                               |
                             +------------v-----------+      +------------v----------+
                             |     Clearing House     |      |   Market Data Feed    |
                             |      (Kafka only)      |      |        :8085          |
                             +------------+-----------+      +------------+----------+
                                          | Kafka                          | WebSocket
                                  trade.cleared                            v
                             +------------v-----------+         +-------------------+
                             |   Settlement Engine    |         |  Trading Terminal |
                             |      (Kafka only)      |         |    (Frontend)     |
                             +------------+-----------+         +-------------------+
                                          | Kafka
                                  trade.settled
                             +------------v-----------+
                             |     Ledger Service     |
                             |         :8087          |
                             +------------------------+
</pre>

<br />

**Communication patterns:**

| From              | To                   | Protocol            | Purpose                         |
| ----------------- | -------------------- | ------------------- | ------------------------------- |
| External clients  | Order Gateway        | FIX 4.2 / REST      | Order submission                |
| Order Gateway     | Participant Registry | gRPC (sync)         | API key authentication          |
| Order Gateway     | Risk Engine          | gRPC (sync)         | Pre-trade collateral check      |
| Order Gateway     | Matching Engine      | gRPC (sync)         | Validated order forwarding      |
| Matching Engine   | Kafka                | Event               | `trade.executed` on every match |
| Clearing House    | Kafka                | Consumer + Producer | Trade guarantee via novation    |
| Settlement Engine | Kafka                | Consumer + Producer | Atomic DvP asset transfer       |
| Ledger Service    | Kafka                | Consumer            | Double-entry bookkeeping        |
| Market Data Feed  | WebSocket            | Push                | Live prices and trade feed      |

---

<h2 id="trade-lifecycle">The Full Trade Lifecycle</h2>

This is what ESX does every time a participant places an order — from submission to final settlement. Every step is a service with its own database and its own bounded responsibility.

<pre>
  PARTICIPANT SUBMITS ORDER (FIX Protocol / REST)
          |
          v
  +------------------+
  |  Order Gateway   |  Parses FIX message or REST body. Validates format.
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
|  Feed  |         |                    |  Emits trade.cleared (Kafka)
|        |         +--------------------+
+--------+                    |
(broadcasts                   v
 live trade          +--------------------+
 over WS)            | Settlement Engine  |  Delivery versus Payment (DvP).
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
                     +--------------------+

  TRADE COMPLETE
</pre>

---

<h2 id="services">Services</h2>

### Order Gateway &nbsp;`:8080`

The public entry point for all market participants. Implements a **FIX Protocol 4.2 parser** alongside a REST API. Every inbound request is authenticated and risk-checked synchronously before the order is forwarded to the Matching Engine.

**REST endpoints:**

<pre>
POST   /orders          Submit a new order
DELETE /orders/:id      Cancel an open order
POST   /fix             Submit a FIX 4.2 message
</pre>

**REST order submission:**

```bash
curl -X POST http://localhost:8080/orders \
  -H "x-api-key: {your_api_key}" \
  -H "Content-Type: application/json" \
  -d '{"symbol": "RELIANCE", "side": "BUY", "type": "LIMIT", "quantity": 10, "price": 50000}'
```

**FIX order submission:**

```bash
curl -X POST http://localhost:8080/fix \
  -H "x-api-key: {your_api_key}" \
  -H "Content-Type: text/plain" \
  -d "8=FIX.4.2|9=100|35=D|49=CLIENT|56=ESX|34=1|52=20240101-10:00:00|11=ORD001|55=RELIANCE|54=1|38=10|40=2|44=50000|10=000|"
```

---

### Participant Registry &nbsp;`:8081`

The identity authority for the entire system. Every other service traces authentication back here.

**REST endpoints:**

<pre>
POST  /participants/register        Register a new participant
POST  /participants/:id/deposit     Deposit cash into account
GET   /participants/:id             Get account details
</pre>

---

### Risk Engine &nbsp;`gRPC :9093`

The pre-trade gatekeeper. No order reaches the order book without explicit approval from the Risk Engine. Runs synchronously over gRPC — the Order Gateway blocks until a response is received.

- **Buy orders**: verify `cash_balance >= (price x quantity)`, lock required cash
- **Sell orders**: verify `share_position >= quantity`, lock required shares

---

### Matching Engine &nbsp;`gRPC :9094`

The heart of ESX. Maintains a live order book per listed security in Redis and applies strict **price-time priority** matching against every incoming order.

- Redis sorted sets as the order book data structure
- Full fills, partial fills, and resting limit orders
- Circuit breaker: halts symbol on 10% price move in 60 seconds
- Emits `trade.executed` to Kafka on every match

---

### Clearing House &nbsp;`Kafka consumer`

Guarantees every trade through **novation** — stepping in as the central counterparty to both sides of every executed trade.

- Consumes `trade.executed`, verifies both locks are active
- Creates a `cleared_trade` record with ESX as central counterparty
- Emits `trade.cleared` to Kafka

---

### Settlement Engine &nbsp;`Kafka consumer`

Implements **Delivery versus Payment (DvP)** — cash and securities transfer simultaneously or not at all inside a single atomic database transaction.

- Consumes `trade.cleared`
- Atomic 4-way transaction across buyer and seller accounts
- Emits `trade.settled` to Kafka

---

### Ledger Service &nbsp;`:8087`

The financial source of truth. Implements **double-entry bookkeeping** across cash and securities dimensions.

**REST endpoints:**

<pre>
GET /ledger/:id/balance                  Current cash balance
GET /ledger/:id/positions                All securities holdings
GET /ledger/:id/cash-transactions        Cash journal entries
GET /ledger/:id/securities-transactions  Securities journal entries
</pre>

---

### Market Data Feed &nbsp;`:8085`

Real-time broadcast of all market activity over **WebSockets**.

**Connect:**

```bash
wscat -c ws://localhost:8085/ws
```

**Subscribe to a channel:**

```json
{ "action": "subscribe", "channel": "trades.RELIANCE" }
```

**Channels:**

| Channel                             | Description                                  |
| ----------------------------------- | -------------------------------------------- |
| `trades.{symbol}`                   | Live trade feed — price, quantity, timestamp |
| `ticker.{symbol}`                   | Last traded price, volume                    |
| `orderbook.{symbol}`                | Order book depth                             |
| `candles.{symbol}.{1m\|5m\|1h\|1d}` | OHLCV candlestick data                       |

---

<h2 id="engineering-concepts">Key Engineering Concepts</h2>

### Price-Time Priority

The universal matching rule of all order-driven exchanges. Orders are matched by best price first — lowest ask for buyers, highest bid for sellers. When two orders share the same price, the one submitted earliest is matched first (strict FIFO). Every major exchange — NASDAQ, NSE, BSE, LSE — uses this rule.

### FIX Protocol

Financial Information eXchange (FIX) is a 40-year-old message standard that still carries the majority of global electronic trading volume. Every institutional broker, trading algorithm, and exchange worldwide speaks FIX. ESX implements a FIX 4.2 parser — any real-world trading system can connect without modification.

### Novation

When the Clearing House steps between buyer and seller after a trade executes, it replaces the original bilateral contract with two new contracts. Neither party has exposure to the other — only to the Clearing House. This is how DTCC operates in the US, NSCCL operates for NSE, and ICCL operates for BSE.

### Delivery versus Payment (DvP)

The settlement principle that cash and securities must transfer simultaneously. The atomic database transaction in the Settlement Engine enforces this — all four movements commit together or all four roll back. There is no state in which a buyer pays without receiving shares, or a seller delivers shares without receiving payment.

### Double-Entry Bookkeeping

Every financial movement creates two equal and opposite journal entries that sum to zero. The `cash_journal` and `securities_journal` can never be in a state where debits do not match credits.

### Circuit Breakers

When the Matching Engine detects a security's price has moved more than 10% within a 60-second rolling window, it halts trading for that security. This mirrors the mechanism NSE uses (dynamic price bands), BSE uses (price band filters), and NASDAQ uses (Limit Up-Limit Down) to prevent flash crashes.

### Amounts in Smallest Currency Unit

All monetary values in ESX are stored and transmitted as integers in the smallest currency unit (paise for INR, cents for USD). There are no floating-point amounts anywhere in the system. `50000` means ₹500.00.

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

### 2. Add kafka to /etc/hosts

```bash
echo "127.0.0.1 kafka" | sudo tee -a /etc/hosts
```

### 3. Generate gRPC stubs

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --go_out=./packages/proto --go-grpc_out=./packages/proto \
  --go_opt=module=github.com/YHQZ1/esx/packages/proto \
  --go-grpc_opt=module=github.com/YHQZ1/esx/packages/proto \
  packages/proto/*.proto
```

### 4. Start infrastructure

```bash
docker compose up -d
docker compose ps   # all containers should show healthy
```

### 5. Create Kafka topics

```bash
for topic in trade.executed trade.cleared trade.settled order.submitted order.cancelled \
             order.partially_filled risk.rejected circuit.breaker.triggered circuit.breaker.lifted; do
  docker exec esx-kafka kafka-topics --bootstrap-server localhost:9092 \
    --create --topic $topic --partitions 3 --replication-factor 1
done
```

### 6. Run migrations

```bash
psql postgres://esx:esx@localhost:5433/participant_registry -f services/participant-registry/db/migrations/001_init.sql
psql postgres://esx:esx@localhost:5433/risk_engine -f services/risk-engine/db/migrations/001_init.sql
psql postgres://esx:esx@localhost:5433/matching_engine -f services/matching-engine/db/migrations/001_init.sql
psql postgres://esx:esx@localhost:5433/clearing_house -f services/clearing-house/db/migrations/001_init.sql
psql postgres://esx:esx@localhost:5433/settlement_engine -f services/settlement-engine/db/migrations/001_init.sql
psql postgres://esx:esx@localhost:5433/ledger_service -f services/ledger-service/db/migrations/001_init.sql
```

### 7. Start all services

```bash
cd services/participant-registry && go run cmd/server/main.go &
cd services/risk-engine && go run cmd/server/main.go &
cd services/matching-engine && go run cmd/server/main.go &
cd services/clearing-house && go run cmd/server/main.go &
cd services/settlement-engine && go run cmd/server/main.go &
cd services/ledger-service && go run cmd/server/main.go &
cd services/order-gateway && go run cmd/server/main.go &
cd services/market-data-feed && go run cmd/server/main.go &
```

### 8. Register participants and trade

**Register a participant:**

```bash
curl -X POST http://localhost:8081/participants/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Trader One", "email": "trader@example.com"}'
```

**Fund the account:**

```bash
curl -X POST http://localhost:8081/participants/{participant_id}/deposit \
  -H "Content-Type: application/json" \
  -d '{"amount": 1000000}'
```

**Submit a limit buy order:**

```bash
curl -X POST http://localhost:8080/orders \
  -H "x-api-key: {your_api_key}" \
  -H "Content-Type: application/json" \
  -d '{"symbol": "RELIANCE", "side": "BUY", "type": "LIMIT", "quantity": 10, "price": 50000}'
```

**Subscribe to live trades:**

```bash
wscat -c ws://localhost:8085/ws
# then send: {"action": "subscribe", "channel": "trades.RELIANCE"}
```

**Check ledger:**

```bash
curl http://localhost:8087/ledger/{participant_id}/balance
curl http://localhost:8087/ledger/{participant_id}/positions
```

---

<h2 id="project-structure">Project Structure</h2>

<pre>
esx/
├── docker-compose.yml
├── go.work
│
├── infra/
│   ├── postgres/init.sql
│   ├── observability/
│   │   ├── prometheus/prometheus.yml
│   │   └── grafana/
│   ├── k8s/manifests/
│   ├── k8s/helm/
│   └── terraform/
│
├── packages/
│   ├── proto/
│   │   ├── participant.proto + generated stubs
│   │   ├── risk.proto + generated stubs
│   │   └── matching.proto + generated stubs
│   ├── kafka/
│   └── logger/
│
└── services/
    ├── order-gateway/          :8080  FIX + REST, auth, risk, matching
    ├── participant-registry/   :8081  identity, API keys, accounts
    ├── risk-engine/            :9093  pre-trade validation, collateral locking
    ├── matching-engine/        :9094  order book, price-time priority matching
    ├── clearing-house/                novation, trade guarantee
    ├── settlement-engine/             atomic DvP settlement
    ├── ledger-service/         :8087  double-entry bookkeeping
    └── market-data-feed/       :8085  WebSocket streaming
</pre>

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

| Phase        | Focus                                                                        | Status   |
| ------------ | ---------------------------------------------------------------------------- | -------- |
| **Phase 1**  | Participant Registry, Risk Engine, Matching Engine, Docker Compose infra     | Complete |
| **Phase 2**  | Clearing House, Settlement Engine, Ledger Service, full Kafka event flow e2e | Complete |
| **Phase 3**  | Order Gateway — FIX Protocol 4.2 parser, full order lifecycle integration    | Complete |
| **Phase 4**  | Market Data Feed — WebSocket streaming, circuit breakers                     | Complete |
| **Phase 5**  | Integration tests, k6 load testing, latency benchmarking on matching engine  | Up Next  |
| **Phase 6**  | Dockerize all services, Kubernetes manifests, Helm charts                    | Planned  |
| **Phase 7**  | AWS with Terraform — EKS, RDS, MSK, ElastiCache                              | Planned  |
| **Phase 8**  | CI/CD — GitHub Actions pipelines + ArgoCD GitOps                             | Planned  |
| **Phase 9**  | Observability — Prometheus, Grafana, OpenTelemetry, Jaeger                   | Planned  |
| **Phase 10** | Frontend trading terminal — Next.js + TradingView Lightweight Charts         | Planned  |

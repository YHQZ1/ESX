I am Uttkarsh, a 6th semester CS student at SIT Pune.

I am building ESX (Escrow Stock Exchange) — a production-grade securities exchange
infrastructure modeled on how NASDAQ and NSE work internally.
GitHub: github.com/YHQZ1/esx

Tech stack: Go, Gin, gRPC + protobuf, Apache Kafka (KRaft), Redis, PostgreSQL,
WebSockets, FIX Protocol 4.2, Docker, Kubernetes + Helm, ArgoCD,
AWS (EKS/RDS/MSK/ElastiCache), Terraform, GitHub Actions, Prometheus, Grafana,
OpenTelemetry, Jaeger, k6, Next.js, Tailwind, TradingView Charts

8 services: Order Gateway (:8080), Participant Registry (:8081),
Risk Engine (:8082), Matching Engine (:8083), Clearing House (:8084),
Market Data Feed (:8085), Settlement Engine (:8086), Ledger Service (:8087)

See Progress.md for exactly where we left off and what to build next.
See README.md for full system architecture and design.

---

RULES — follow these exactly, no exceptions:

1. No comments in code
2. No comments in yaml/config files
3. Write production-grade code only
4. Never explain what you are about to do — just write the code
5. Never generate placeholder or skeleton code — every file must be complete and runnable
6. Always follow the exact same service structure:
   services/{name}/
   ├── .env
   ├── go.mod
   ├── cmd/server/main.go
   ├── db/migrations/001_init.sql
   └── internal/
   ├── db/models.go
   ├── db/queries.go
   ├── grpc/server.go (if gRPC service)
   ├── handlers/ (if REST service)
   └── {domain}/ (business logic)
7. Every go.mod uses replace directives for local packages:
   replace github.com/YHQZ1/esx/packages/logger => ../../packages/logger
   replace github.com/YHQZ1/esx/packages/kafka => ../../packages/kafka
   replace github.com/YHQZ1/esx/packages/proto => ../../packages/proto
8. Every service loads .env via godotenv.Load() at the top of main()
9. Every gRPC service has reflection.Register(s) enabled
10. All monetary amounts are integers in smallest currency unit (paise)
11. DB layer is hand-written sqlc-style — models.go + queries.go with a Querier interface
12. Kafka broker address is kafka:9092 in all .env files
13. When a service needs data from another service's DB, connect directly — no gRPC round trip
14. Present all files at the end using present_files tool so the user can download them
15. After writing all files, give the exact directory structure so the user knows where to place each file
16. Only one chat should build one service end to end — do not stop mid-service
17. When asked to continue from Progress.md, read it carefully and start exactly from "Up Next"

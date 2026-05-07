#!/bin/bash
set -euo pipefail

BIN_DIR="/tmp/esx-bins"
LOG_DIR="/tmp/esx-logs"
PID_DIR="/tmp/esx-pids"

SERVICES=(
  "participant-registry"
  "risk-engine"
  "matching-engine"
  "clearing-house"
  "settlement-engine"
  "ledger-service"
  "order-gateway"
  "market-data-feed"
)

# ESX-owned ports only — never include 9092/9093 (Kafka)
ESX_PORTS="8080,8081,8085,8087,9091,9094"

log()  { echo "[$(date '+%H:%M:%S')] $*"; }
ok()   { echo "[$(date '+%H:%M:%S')] ✓ $*"; }
err()  { echo "[$(date '+%H:%M:%S')] ✗ $*" >&2; }

wait_for_infra() {
  log "Waiting for Postgres (5433)..."
  attempts=0
  until pg_isready -h localhost -p 5433 -U esx > /dev/null 2>&1; do
    attempts=$((attempts + 1))
    if [ $attempts -ge 30 ]; then
      err "Postgres not ready after 30s. Is Docker running?"
      exit 1
    fi
    printf '.'
    sleep 1
  done
  echo; ok "Postgres ready."

  log "Waiting for Kafka (9092)..."
  attempts=0
  until nc -z localhost 9092 > /dev/null 2>&1; do
    attempts=$((attempts + 1))
    if [ $attempts -ge 60 ]; then
      err "Kafka not ready after 60s. Is Docker running?"
      exit 1
    fi
    printf '.'
    sleep 1
  done
  echo; ok "Kafka ready."

  log "Waiting for Redis (6379)..."
  attempts=0
  until nc -z localhost 6379 > /dev/null 2>&1; do
    attempts=$((attempts + 1))
    if [ $attempts -ge 30 ]; then
      err "Redis not ready after 30s. Is Docker running?"
      exit 1
    fi
    printf '.'
    sleep 1
  done
  echo; ok "Redis ready."
}

start_all() {
  stop_all
  wait_for_infra

  mkdir -p "$BIN_DIR" "$LOG_DIR" "$PID_DIR"
  log "Building and starting all ESX services..."

  for svc in "${SERVICES[@]}"; do
    log "  Building $svc..."
    if ! (cd "services/$svc" && go build -o "$BIN_DIR/$svc" ./cmd/server); then
      err "Failed to build $svc"
      exit 1
    fi

    (cd "services/$svc" && "$BIN_DIR/$svc") > "$LOG_DIR/$svc.log" 2>&1 &
    echo $! > "$PID_DIR/$svc.pid"
    ok "  Started $svc (pid=$!)"
  done

  log "All services started. Logs in $LOG_DIR/"
}

stop_all() {
  log "Stopping all ESX services..."

  for svc in "${SERVICES[@]}"; do
    pid_file="$PID_DIR/$svc.pid"
    if [ -f "$pid_file" ]; then
      pid=$(cat "$pid_file")
      if kill "$pid" 2>/dev/null; then
        ok "  Stopped $svc (pid=$pid)"
      fi
      rm -f "$pid_file"
    fi
  done

  # Safety net for anything that slipped through — ESX ports only
  rm -f "$BIN_DIR"/*
  log "All services stopped."
}

logs_all() {
  log_files=()
  for svc in "${SERVICES[@]}"; do
    log_files+=("$LOG_DIR/$svc.log")
  done
  tail -f "${log_files[@]}"
}

status_all() {
  echo "=== ESX Service Status ==="
  for svc in "${SERVICES[@]}"; do
    pid_file="$PID_DIR/$svc.pid"
    if [ -f "$pid_file" ]; then
      pid=$(cat "$pid_file")
      if kill -0 "$pid" 2>/dev/null; then
        echo "  ✓ $svc (pid=$pid)"
      else
        echo "  ✗ $svc (DEAD — stale pid=$pid)"
      fi
    else
      echo "  - $svc (not started)"
    fi
  done
}

seed_sellers() {
  log "Registering fixed load test sellers..."
  for i in $(seq 0 19); do
    RESULT=$(curl -s -X POST http://localhost:8081/participants/register \
      -H "Content-Type: application/json" \
      -d "{\"name\":\"k6 Seller $i\",\"email\":\"k6_seller_fixed_$i@esx.com\"}")
    SELLER_ID=$(echo "$RESULT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('participant_id',''))")
    SELLER_KEY=$(echo "$RESULT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('api_key',''))")
    if [ -n "$SELLER_ID" ]; then
      psql postgres://esx:esx@localhost:5433/participant_registry -c \
        "INSERT INTO securities_accounts (participant_id, symbol, quantity) VALUES ('$SELLER_ID', 'RELIANCE', 100000000) ON CONFLICT (participant_id, symbol) DO UPDATE SET quantity = 100000000;" > /dev/null
      echo "SELLER_KEY_$i=$SELLER_KEY"
    fi
  done
}

seed_buyers() {
  log "Registering fixed load test buyers..."
  for i in $(seq 0 19); do
    RESULT=$(curl -s -X POST http://localhost:8081/participants/register \
      -H "Content-Type: application/json" \
      -d "{\"name\":\"k6 Buyer $i\",\"email\":\"k6_buyer_fixed_$i@esx.com\"}")
    PARTICIPANT_ID=$(echo "$RESULT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('participant_id',''))")
    API_KEY=$(echo "$RESULT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('api_key',''))")
    if [ -n "$PARTICIPANT_ID" ]; then
      curl -s -X POST "http://localhost:8081/participants/$PARTICIPANT_ID/deposit" \
        -H "Content-Type: application/json" \
        -d '{"amount":100000000000}' > /dev/null
      echo "BUYER_KEY_$i=$API_KEY"
    fi
  done
}

# ── entrypoint ────────────────────────────────────────────────────────────────
case "${1:-help}" in
  start)   start_all ;;
  stop)    stop_all ;;
  restart) stop_all && start_all ;;
  logs)    logs_all ;;
  status)  status_all ;;
  seed-sellers) seed_sellers ;;
  seed-buyers)  seed_buyers ;;
  *)
    echo "Usage: $0 {start|stop|restart|logs|status|seed-sellers|seed-buyers}"
    exit 1
    ;;
esac
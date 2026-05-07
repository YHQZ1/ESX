.PHONY: run-all stop-all logs-all seed-sellers seed-buyers

SERVICES := participant-registry risk-engine matching-engine clearing-house settlement-engine ledger-service order-gateway market-data-feed
BIN_DIR  := /tmp/esx-bins
LOG_DIR  := /tmp/esx-logs
PID_DIR  := /tmp/esx-pids

run-all: stop-all
	@echo "Starting all ESX services..."
	@mkdir -p $(BIN_DIR) $(LOG_DIR) $(PID_DIR)
	@echo "  Building & starting participant-registry..."
	@cd services/participant-registry  && go build -o $(BIN_DIR)/participant-registry  ./cmd/server && $(BIN_DIR)/participant-registry  > $(LOG_DIR)/participant-registry.log  2>&1 & echo $$! > $(PID_DIR)/participant-registry.pid
	@echo "  Building & starting risk-engine..."
	@cd services/risk-engine           && go build -o $(BIN_DIR)/risk-engine           ./cmd/server && $(BIN_DIR)/risk-engine           > $(LOG_DIR)/risk-engine.log           2>&1 & echo $$! > $(PID_DIR)/risk-engine.pid
	@echo "  Building & starting matching-engine..."
	@cd services/matching-engine       && go build -o $(BIN_DIR)/matching-engine       ./cmd/server && $(BIN_DIR)/matching-engine       > $(LOG_DIR)/matching-engine.log       2>&1 & echo $$! > $(PID_DIR)/matching-engine.pid
	@echo "  Building & starting clearing-house..."
	@cd services/clearing-house        && go build -o $(BIN_DIR)/clearing-house        ./cmd/server && $(BIN_DIR)/clearing-house        > $(LOG_DIR)/clearing-house.log        2>&1 & echo $$! > $(PID_DIR)/clearing-house.pid
	@echo "  Building & starting settlement-engine..."
	@cd services/settlement-engine     && go build -o $(BIN_DIR)/settlement-engine     ./cmd/server && $(BIN_DIR)/settlement-engine     > $(LOG_DIR)/settlement-engine.log     2>&1 & echo $$! > $(PID_DIR)/settlement-engine.pid
	@echo "  Building & starting ledger-service..."
	@cd services/ledger-service        && go build -o $(BIN_DIR)/ledger-service        ./cmd/server && $(BIN_DIR)/ledger-service        > $(LOG_DIR)/ledger-service.log        2>&1 & echo $$! > $(PID_DIR)/ledger-service.pid
	@echo "  Building & starting order-gateway..."
	@cd services/order-gateway         && go build -o $(BIN_DIR)/order-gateway         ./cmd/server && $(BIN_DIR)/order-gateway         > $(LOG_DIR)/order-gateway.log         2>&1 & echo $$! > $(PID_DIR)/order-gateway.pid
	@echo "  Building & starting market-data-feed..."
	@cd services/market-data-feed      && go build -o $(BIN_DIR)/market-data-feed      ./cmd/server && $(BIN_DIR)/market-data-feed      > $(LOG_DIR)/market-data-feed.log      2>&1 & echo $$! > $(PID_DIR)/market-data-feed.pid
	@echo "All services started. Logs in $(LOG_DIR)/"

stop-all:
	@echo "Stopping all ESX services..."
	@for pid_file in $(PID_DIR)/participant-registry.pid $(PID_DIR)/risk-engine.pid \
		$(PID_DIR)/matching-engine.pid $(PID_DIR)/clearing-house.pid \
		$(PID_DIR)/settlement-engine.pid $(PID_DIR)/ledger-service.pid \
		$(PID_DIR)/order-gateway.pid $(PID_DIR)/market-data-feed.pid; do \
		if [ -f $$pid_file ]; then \
			kill $$(cat $$pid_file) 2>/dev/null || true; \
			rm -f $$pid_file; \
		fi \
	done
	@# Nuke anything still holding the ports, covers orphaned go run processes from old sessions
	@lsof -ti:8080,8081,8085,8087,9091,9092,9093,9094 | xargs kill -9 2>/dev/null || true
	@rm -f $(BIN_DIR)/*
	@echo "All services stopped."

logs-all:
	@tail -f \
		$(LOG_DIR)/participant-registry.log \
		$(LOG_DIR)/risk-engine.log \
		$(LOG_DIR)/matching-engine.log \
		$(LOG_DIR)/clearing-house.log \
		$(LOG_DIR)/settlement-engine.log \
		$(LOG_DIR)/ledger-service.log \
		$(LOG_DIR)/order-gateway.log \
		$(LOG_DIR)/market-data-feed.log

status:
	@echo "=== ESX Service Status ==="
	@for svc in $(SERVICES); do \
		pid_file=$(PID_DIR)/$$svc.pid; \
		if [ -f $$pid_file ]; then \
			pid=$$(cat $$pid_file); \
			if kill -0 $$pid 2>/dev/null; then \
				echo "  ✓ $$svc (pid=$$pid)"; \
			else \
				echo "  ✗ $$svc (DEAD — stale pid=$$pid)"; \
			fi \
		else \
			echo "  - $$svc (not started)"; \
		fi \
	done

seed-sellers:
	@echo "Registering fixed load test sellers..."
	@for i in $$(seq 0 19); do \
		RESULT=$$(curl -s -X POST http://localhost:8081/participants/register \
			-H "Content-Type: application/json" \
			-d "{\"name\":\"k6 Seller $$i\",\"email\":\"k6_seller_fixed_$$i@esx.com\"}"); \
		SELLER_ID=$$(echo $$RESULT | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('participant_id',''))"); \
		SELLER_KEY=$$(echo $$RESULT | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('api_key',''))"); \
		if [ -n "$$SELLER_ID" ]; then \
			psql postgres://esx:esx@localhost:5433/participant_registry -c \
			"INSERT INTO securities_accounts (participant_id, symbol, quantity) VALUES ('$$SELLER_ID', 'RELIANCE', 100000000) ON CONFLICT (participant_id, symbol) DO UPDATE SET quantity = 100000000;" > /dev/null; \
			echo "SELLER_KEY_$$i=$$SELLER_KEY"; \
		fi \
	done

seed-buyers:
	@echo "Registering fixed load test buyers..."
	@for i in $$(seq 0 19); do \
		RESULT=$$(curl -s -X POST http://localhost:8081/participants/register \
			-H "Content-Type: application/json" \
			-d "{\"name\":\"k6 Buyer $$i\",\"email\":\"k6_buyer_fixed_$$i@esx.com\"}"); \
		PARTICIPANT_ID=$$(echo $$RESULT | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('participant_id',''))"); \
		API_KEY=$$(echo $$RESULT | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('api_key',''))"); \
		if [ -n "$$PARTICIPANT_ID" ]; then \
			curl -s -X POST http://localhost:8081/participants/$$PARTICIPANT_ID/deposit \
				-H "Content-Type: application/json" \
				-d '{"amount":100000000000}' > /dev/null; \
			echo "BUYER_KEY_$$i=$$API_KEY"; \
		fi \
	done
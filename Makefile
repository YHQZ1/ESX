.PHONY: run-all stop-all logs-all

run-all:
	@echo "Starting all ESX services..."
	@cd services/participant-registry && go run cmd/server/main.go > /tmp/participant-registry.log 2>&1 & echo $$! > /tmp/participant-registry.pid
	@cd services/risk-engine && go run cmd/server/main.go > /tmp/risk-engine.log 2>&1 & echo $$! > /tmp/risk-engine.pid
	@cd services/matching-engine && go run cmd/server/main.go > /tmp/matching-engine.log 2>&1 & echo $$! > /tmp/matching-engine.pid
	@cd services/clearing-house && go run cmd/server/main.go > /tmp/clearing-house.log 2>&1 & echo $$! > /tmp/clearing-house.pid
	@cd services/settlement-engine && go run cmd/server/main.go > /tmp/settlement-engine.log 2>&1 & echo $$! > /tmp/settlement-engine.pid
	@cd services/ledger-service && go run cmd/server/main.go > /tmp/ledger-service.log 2>&1 & echo $$! > /tmp/ledger-service.pid
	@cd services/order-gateway && go run cmd/server/main.go > /tmp/order-gateway.log 2>&1 & echo $$! > /tmp/order-gateway.pid
	@cd services/market-data-feed && go run cmd/server/main.go > /tmp/market-data-feed.log 2>&1 & echo $$! > /tmp/market-data-feed.pid
	@echo "All services started. Logs in /tmp/*.log"

stop-all:
	@echo "Stopping all ESX services..."
	@for pid in /tmp/participant-registry.pid /tmp/risk-engine.pid /tmp/matching-engine.pid \
		/tmp/clearing-house.pid /tmp/settlement-engine.pid /tmp/ledger-service.pid \
		/tmp/order-gateway.pid /tmp/market-data-feed.pid; do \
		if [ -f $$pid ]; then \
			kill $$(cat $$pid) 2>/dev/null; \
			rm -f $$pid; \
		fi \
	done
	@pkill -f "go run.*cmd/server/main.go" 2>/dev/null || true
	@echo "All services stopped."

logs-all:
	@tail -f /tmp/participant-registry.log /tmp/risk-engine.log /tmp/matching-engine.log \
		/tmp/clearing-house.log /tmp/settlement-engine.log /tmp/ledger-service.log \
		/tmp/order-gateway.log /tmp/market-data-feed.log

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
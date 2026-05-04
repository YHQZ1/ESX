.PHONY: dev down

dev:
	tmux new-session -d -s esx -n infra
	tmux send-keys -t esx 'docker-compose up -d' Enter
	sleep 5
	tmux new-window -t esx -n participant-registry
	tmux send-keys -t esx:participant-registry 'cd services/participant-registry && go run cmd/server/main.go' Enter
	tmux new-window -t esx -n risk-engine
	tmux send-keys -t esx:risk-engine 'cd services/risk-engine && go run cmd/server/main.go' Enter
	tmux new-window -t esx -n matching-engine
	tmux send-keys -t esx:matching-engine 'cd services/matching-engine && go run cmd/server/main.go' Enter
	tmux new-window -t esx -n order-gateway
	tmux send-keys -t esx:order-gateway 'cd services/order-gateway && go run cmd/server/main.go' Enter
	tmux new-window -t esx -n clearing-house
	tmux send-keys -t esx:clearing-house 'cd services/clearing-house && go run cmd/server/main.go' Enter
	tmux new-window -t esx -n settlement-engine
	tmux send-keys -t esx:settlement-engine 'cd services/settlement-engine && go run cmd/server/main.go' Enter
	tmux new-window -t esx -n ledger-service
	tmux send-keys -t esx:ledger-service 'cd services/ledger-service && go run cmd/server/main.go' Enter
	tmux new-window -t esx -n market-data-feed
	tmux send-keys -t esx:market-data-feed 'cd services/market-data-feed && go run cmd/server/main.go' Enter
	tmux attach -t esx

down:
	tmux kill-session -t esx
	docker-compose down
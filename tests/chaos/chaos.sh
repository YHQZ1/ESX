#!/bin/bash

echo "Initiating ESX Fault Tolerance and Recovery Test"
echo "------------------------------------------------"

echo "[1/5] Launching k6 load test in the background..."
k6 run infra/k6/orderflow.js > /dev/null 2>&1 &
K6_PID=$!

echo "Waiting 15 seconds for traffic to reach peak load..."
sleep 15

echo ""
echo "[2/5] Simulating a localized service outage..."
echo "🚨 ACTION REQUIRED: Please switch to your Settlement Engine Tmux pane and press Ctrl+C to stop it."
read -p "Press [Enter] here to continue once you have stopped it..."
echo ""

echo "[3/5] Settlement layer is offline. Verifying Kafka event queuing..."
echo "Waiting 15 seconds while trades pile up in the Kafka broker..."
sleep 15

echo ""
echo "[4/5] Initiating state recovery..."
echo "🚨 ACTION REQUIRED: Please switch back to your Settlement Engine Tmux pane and restart it (go run cmd/server/main.go)."
read -p "Press [Enter] here to continue once it is running again..."
echo ""

echo "Waiting for the load test duration to complete..."
wait $K6_PID
echo "Load test execution finished."

echo "[5/5] Executing Ledger Reconciliation Audit..."
echo "Verifying absolute data integrity and zero message loss during the outage window."
echo "------------------------------------------------"
go test -v tests/integration/reconciliation_test.go
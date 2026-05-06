package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const (
	gatewayURL  = "http://localhost:8080"
	registryURL = "http://localhost:8081"
	ledgerURL   = "http://localhost:8087"
	marketData  = "ws://localhost:8085/ws"
	dbDSN       = "postgres://esx:esx@localhost:5433/participant_registry?sslmode=disable"
)

type Participant struct {
	ID     string `json:"participant_id"`
	APIKey string `json:"api_key"`
}

type OrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type WSMessage struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

func TestEndToEndTradeLifecycle(t *testing.T) {
	// 1. Setup Participants
	t.Log("Registering participants...")
	buyer := registerParticipant(t, "Buyer Alpha", fmt.Sprintf("buyer_%d@esx.com", time.Now().UnixNano()))
	seller := registerParticipant(t, "Seller Beta", fmt.Sprintf("seller_%d@esx.com", time.Now().UnixNano()))

	// 2. Fund Accounts
	t.Log("Funding accounts...")
	initialCash := int64(10000000) // 100,000.00 INR
	tradeQty := int64(15)
	tradePrice := int64(50000) // 500.00 INR
	totalTradeValue := tradeQty * tradePrice

	depositCash(t, buyer.ID, initialCash)
	injectSharesDB(t, seller.ID, "RELIANCE", 50)

	// 3. Connect to Market Data WebSocket
	t.Log("Connecting to Market Data Feed...")
	u, _ := url.Parse(marketData)
	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.NoError(t, err)
	defer wsConn.Close()

	subMsg, _ := json.Marshal(map[string]string{"action": "subscribe", "channel": "trades.RELIANCE"})
	require.NoError(t, wsConn.WriteMessage(websocket.TextMessage, subMsg))

	// 4. Submit Orders
	// Orders are queued asynchronously — both return ACCEPTED immediately
	t.Log("Submitting orders...")
	sellOrder := submitOrder(t, seller.APIKey, "RELIANCE", "SELL", tradeQty, tradePrice)
	require.Equal(t, "ACCEPTED", sellOrder.Status)

	buyOrder := submitOrder(t, buyer.APIKey, "RELIANCE", "BUY", tradeQty, tradePrice)
	require.Equal(t, "ACCEPTED", buyOrder.Status)

	// 5. Await WebSocket Broadcast
	t.Log("Waiting for trade execution broadcast over WebSocket...")
	wsConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		_, msg, err := wsConn.ReadMessage()
		require.NoError(t, err)

		var m WSMessage
		require.NoError(t, json.Unmarshal(msg, &m))
		if m.Channel == "trades.RELIANCE" {
			t.Logf("Received real-time trade event: %s", string(m.Data))
			break
		}
	}

	// 6. Poll Ledger for final settlement reconciliation
	// Settlement and Ledger operate asynchronously via Kafka so we poll
	t.Log("Polling Ledger Service for final settlement reconciliation...")
	require.Eventually(t, func() bool {
		buyerBal := getLedgerBalance(t, buyer.ID)
		sellerBal := getLedgerBalance(t, seller.ID)
		buyerPos := getLedgerPositions(t, buyer.ID)
		sellerPos := getLedgerPositions(t, seller.ID)

		buyerCashCorrect := buyerBal == (initialCash - totalTradeValue)
		sellerCashCorrect := sellerBal == totalTradeValue
		buyerSharesCorrect := buyerPos["RELIANCE"] == tradeQty
		sellerSharesCorrect := sellerPos["RELIANCE"] == (50 - tradeQty)

		return buyerCashCorrect && sellerCashCorrect && buyerSharesCorrect && sellerSharesCorrect
	}, 15*time.Second, 500*time.Millisecond, "Ledger did not reconcile correctly within timeout")

	t.Log("SUCCESS: Full Trade Lifecycle Reconciled across 8 services!")
}

func registerParticipant(t *testing.T, name, email string) Participant {
	body, _ := json.Marshal(map[string]string{"name": name, "email": email})
	resp, err := http.Post(registryURL+"/participants/register", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var p Participant
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&p))
	return p
}

func depositCash(t *testing.T, participantID string, amount int64) {
	body, _ := json.Marshal(map[string]int64{"amount": amount})
	resp, err := http.Post(fmt.Sprintf("%s/participants/%s/deposit", registryURL, participantID), "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func injectSharesDB(t *testing.T, participantID, symbol string, quantity int64) {
	db, err := sql.Open("postgres", dbDSN)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(
		`INSERT INTO securities_accounts (participant_id, symbol, quantity)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (participant_id, symbol)
		 DO UPDATE SET quantity = securities_accounts.quantity + $3`,
		participantID, symbol, quantity,
	)
	require.NoError(t, err)
}

func submitOrder(t *testing.T, apiKey, symbol, side string, qty, price int64) OrderResponse {
	body, _ := json.Marshal(map[string]any{
		"symbol": symbol, "side": side, "type": "LIMIT", "quantity": qty, "price": price,
	})
	req, _ := http.NewRequest("POST", gatewayURL+"/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result OrderResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

func getLedgerBalance(t *testing.T, participantID string) int64 {
	resp, err := http.Get(fmt.Sprintf("%s/ledger/%s/balance", ledgerURL, participantID))
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0
	}

	var res struct {
		Balance int64 `json:"balance"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	return res.Balance
}

func getLedgerPositions(t *testing.T, participantID string) map[string]int64 {
	resp, err := http.Get(fmt.Sprintf("%s/ledger/%s/positions", ledgerURL, participantID))
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]int64{}
	}

	var res struct {
		Positions map[string]int64 `json:"positions"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	return res.Positions
}

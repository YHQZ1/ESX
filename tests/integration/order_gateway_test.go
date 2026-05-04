package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

const gatewayURL = "http://localhost:8080"

const (
	buyerAPIKey  = "9e214dc45c6db75645a1598bd60ec875db9192ed81f9da23c4093c4ea3af96bd"
	sellerAPIKey = "d2f36fa3c1544c57e16c024fd48435b8bb627950f8727d06cb6ba168c0e50cd6"
	buyerID      = "86303d8d-4429-41de-9a03-66c72d3fe06e"
)

type orderResponse struct {
	OrderID  string `json:"order_id"`
	Status   string `json:"status"`
	Symbol   string `json:"symbol"`
	Side     string `json:"side"`
	Quantity int64  `json:"quantity"`
	Price    int64  `json:"price"`
}

func fundBuyer(t *testing.T) {
	t.Helper()
	body, _ := json.Marshal(map[string]int64{"amount": 10000000})
	resp, err := http.Post(
		fmt.Sprintf("%s/participants/%s/deposit", registryURL, buyerID),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatalf("failed to fund buyer: %v", err)
	}
	resp.Body.Close()
}

func submitOrder(t *testing.T, apiKey, symbol, side, orderType string, quantity, price int64) orderResponse {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"symbol":   symbol,
		"side":     side,
		"type":     orderType,
		"quantity": quantity,
		"price":    price,
	})

	req, _ := http.NewRequest("POST", gatewayURL+"/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("order request failed: %v", err)
	}
	defer resp.Body.Close()

	var result orderResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func TestOrderSubmissionRequiresAPIKey(t *testing.T) {
	body, _ := json.Marshal(map[string]any{
		"symbol": "RELIANCE", "side": "BUY", "type": "LIMIT", "quantity": 1, "price": 50000,
	})
	resp, err := http.Post(gatewayURL+"/orders", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestInvalidAPIKeyRejected(t *testing.T) {
	body, _ := json.Marshal(map[string]any{
		"symbol": "RELIANCE", "side": "BUY", "type": "LIMIT", "quantity": 1, "price": 50000,
	})
	req, _ := http.NewRequest("POST", gatewayURL+"/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", "invalid-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLimitOrderRestingInBook(t *testing.T) {
	price := int64(30000 + time.Now().UnixNano()%10000)
	result := submitOrder(t, sellerAPIKey, "RELIANCE", "SELL", "LIMIT", 1, price)

	if result.OrderID == "" {
		t.Fatal("expected order_id in response")
	}
	if result.Status != "open" {
		t.Fatalf("expected status open, got %s", result.Status)
	}
}

func TestFullFillBothSides(t *testing.T) {
	fundBuyer(t)

	price := int64(45000 + time.Now().UnixNano()%1000)

	sell := submitOrder(t, sellerAPIKey, "RELIANCE", "SELL", "LIMIT", 1, price)
	if sell.Status != "open" {
		t.Fatalf("sell order should be open, got %s", sell.Status)
	}

	buy := submitOrder(t, buyerAPIKey, "RELIANCE", "BUY", "LIMIT", 1, price)
	if buy.Status != "filled" {
		t.Fatalf("buy order should be filled, got %s", buy.Status)
	}

	t.Logf("trade executed: order_id=%s price=%d", buy.OrderID, price)
}

func TestInsufficientFundsRejected(t *testing.T) {
	body, _ := json.Marshal(map[string]any{
		"symbol": "RELIANCE", "side": "BUY", "type": "LIMIT", "quantity": 999999, "price": 999999,
	})
	req, _ := http.NewRequest("POST", gatewayURL+"/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", buyerAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 insufficient funds, got %d", resp.StatusCode)
	}
}

func TestFIXOrderSubmission(t *testing.T) {
	price := fmt.Sprintf("%d", 40000+time.Now().UnixNano()%5000)
	clOrdID := fmt.Sprintf("TEST%d", time.Now().UnixNano())

	fixMsg := fmt.Sprintf(
		"8=FIX.4.2|9=100|35=D|49=CLIENT|56=ESX|34=1|52=20240101-10:00:00|11=%s|55=RELIANCE|54=2|38=1|40=2|44=%s|10=000|",
		clOrdID, price,
	)

	req, _ := http.NewRequest("POST", gatewayURL+"/fix", bytes.NewBufferString(fixMsg))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("x-api-key", sellerAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("FIX request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	if len(body) == 0 {
		t.Fatal("expected FIX execution report in response")
	}
	if !strings.Contains(body, "8=FIX.4.2") {
		t.Fatalf("expected FIX execution report containing 8=FIX.4.2, got: %s", body[:20])
	}

	t.Logf("FIX execution report received: %s", body[:50])
}

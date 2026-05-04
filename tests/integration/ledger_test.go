package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const ledgerURL = "http://localhost:8087"

type balanceResponse struct {
	ParticipantID string `json:"participant_id"`
	Balance       int64  `json:"balance"`
	Currency      string `json:"currency"`
}

type positionsResponse struct {
	ParticipantID string           `json:"participant_id"`
	Positions     map[string]int64 `json:"positions"`
}

type cashTransactionsResponse struct {
	ParticipantID string `json:"participant_id"`
	Entries       []struct {
		EntryType    string `json:"entry_type"`
		Amount       int64  `json:"amount"`
		BalanceAfter int64  `json:"balance_after"`
	} `json:"entries"`
}

func TestLedgerBalance(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/ledger/%s/balance", ledgerURL, buyerID))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result balanceResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.ParticipantID != buyerID {
		t.Fatalf("expected participant_id %s, got %s", buyerID, result.ParticipantID)
	}
	if result.Currency != "INR" {
		t.Fatalf("expected currency INR, got %s", result.Currency)
	}

	t.Logf("buyer balance: %d paise", result.Balance)
}

func TestLedgerPositions(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/ledger/%s/positions", ledgerURL, buyerID))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result positionsResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.ParticipantID != buyerID {
		t.Fatalf("expected participant_id %s, got %s", buyerID, result.ParticipantID)
	}

	t.Logf("buyer positions: %v", result.Positions)
}

func TestLedgerCashTransactions(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/ledger/%s/cash-transactions", ledgerURL, buyerID))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result cashTransactionsResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.ParticipantID != buyerID {
		t.Fatalf("expected participant_id %s, got %s", buyerID, result.ParticipantID)
	}

	for _, entry := range result.Entries {
		if entry.EntryType != "DEBIT" && entry.EntryType != "CREDIT" {
			t.Fatalf("unexpected entry type: %s", entry.EntryType)
		}
		if entry.Amount <= 0 {
			t.Fatalf("expected positive amount, got %d", entry.Amount)
		}
	}

	t.Logf("cash journal entries: %d", len(result.Entries))
}

func TestLedgerSecuritiesTransactions(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/ledger/%s/securities-transactions", ledgerURL, buyerID))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	t.Logf("securities-transactions endpoint responded 200")
}

func TestLedgerAfterTrade(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("%s/ledger/%s/balance", ledgerURL, buyerID))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	var before balanceResponse
	json.NewDecoder(resp.Body).Decode(&before)
	resp.Body.Close()

	price := int64(44000 + time.Now().UnixNano()%1000)
	submitOrder(t, sellerAPIKey, "RELIANCE", "SELL", "LIMIT", 1, price)
	submitOrder(t, buyerAPIKey, "RELIANCE", "BUY", "LIMIT", 1, price)

	time.Sleep(2 * time.Second)

	resp, err = http.Get(fmt.Sprintf("%s/ledger/%s/balance", ledgerURL, buyerID))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	var after balanceResponse
	json.NewDecoder(resp.Body).Decode(&after)
	resp.Body.Close()

	if after.Balance >= before.Balance {
		t.Fatalf("expected balance to decrease after buy trade: before=%d after=%d", before.Balance, after.Balance)
	}

	t.Logf("balance before: %d, after: %d, diff: %d", before.Balance, after.Balance, before.Balance-after.Balance)
}

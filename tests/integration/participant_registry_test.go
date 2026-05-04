package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const registryURL = "http://localhost:8081"

type registerResponse struct {
	ParticipantID string `json:"participant_id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	APIKey        string `json:"api_key"`
}

type depositResponse struct {
	ParticipantID string `json:"participant_id"`
	Balance       int64  `json:"balance"`
	Currency      string `json:"currency"`
}

type accountResponse struct {
	Participant struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Status string `json:"status"`
	} `json:"participant"`
	Cash struct {
		Balance  int64  `json:"balance"`
		Currency string `json:"currency"`
	} `json:"cash"`
	Securities []any `json:"securities"`
}

func TestParticipantRegistration(t *testing.T) {
	email := fmt.Sprintf("test_%d@esx.com", time.Now().UnixNano())

	body, _ := json.Marshal(map[string]string{
		"name":  "Test Trader",
		"email": email,
	})

	resp, err := http.Post(registryURL+"/participants/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result registerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.ParticipantID == "" {
		t.Fatal("expected participant_id in response")
	}
	if result.APIKey == "" {
		t.Fatal("expected api_key in response")
	}
	if result.Email != email {
		t.Fatalf("expected email %s, got %s", email, result.Email)
	}

	t.Logf("registered participant: %s", result.ParticipantID)
}

func TestDuplicateEmailRejected(t *testing.T) {
	email := fmt.Sprintf("dupe_%d@esx.com", time.Now().UnixNano())
	body, _ := json.Marshal(map[string]string{"name": "Trader", "email": email})

	http.Post(registryURL+"/participants/register", "application/json", bytes.NewBuffer(body))

	body, _ = json.Marshal(map[string]string{"name": "Trader", "email": email})
	resp, err := http.Post(registryURL+"/participants/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected duplicate to fail, got %d", resp.StatusCode)
	}
}

func TestDeposit(t *testing.T) {
	email := fmt.Sprintf("deposit_%d@esx.com", time.Now().UnixNano())
	body, _ := json.Marshal(map[string]string{"name": "Depositor", "email": email})

	resp, _ := http.Post(registryURL+"/participants/register", "application/json", bytes.NewBuffer(body))
	var reg registerResponse
	json.NewDecoder(resp.Body).Decode(&reg)
	resp.Body.Close()

	body, _ = json.Marshal(map[string]int64{"amount": 500000})
	resp, err := http.Post(
		fmt.Sprintf("%s/participants/%s/deposit", registryURL, reg.ParticipantID),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		t.Fatalf("deposit request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result depositResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Balance != 500000 {
		t.Fatalf("expected balance 500000, got %d", result.Balance)
	}

	t.Logf("deposited 500000 paise, balance: %d", result.Balance)
}

func TestGetAccount(t *testing.T) {
	email := fmt.Sprintf("account_%d@esx.com", time.Now().UnixNano())
	body, _ := json.Marshal(map[string]string{"name": "Account Trader", "email": email})

	resp, _ := http.Post(registryURL+"/participants/register", "application/json", bytes.NewBuffer(body))
	var reg registerResponse
	json.NewDecoder(resp.Body).Decode(&reg)
	resp.Body.Close()

	resp, err := http.Get(fmt.Sprintf("%s/participants/%s", registryURL, reg.ParticipantID))
	if err != nil {
		t.Fatalf("get account request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result accountResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Participant.ID != reg.ParticipantID {
		t.Fatalf("expected participant_id %s, got %s", reg.ParticipantID, result.Participant.ID)
	}
	if result.Participant.Status != "active" {
		t.Fatalf("expected status active, got %s", result.Participant.Status)
	}
	if result.Cash.Balance != 0 {
		t.Fatalf("expected balance 0, got %d", result.Cash.Balance)
	}
}

func TestInvalidParticipantID(t *testing.T) {
	resp, err := http.Get(registryURL + "/participants/not-a-uuid")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

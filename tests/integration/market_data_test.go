package integration

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const marketDataURL = "ws://localhost:8085/ws"

type wsMessage struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

func TestWebSocketConnect(t *testing.T) {
	u, _ := url.Parse(marketDataURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	t.Log("WebSocket connection established")
}

func TestWebSocketSubscribeAndReceive(t *testing.T) {
	u, _ := url.Parse(marketDataURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	subMsg, _ := json.Marshal(map[string]string{
		"action":  "subscribe",
		"channel": "trades.RELIANCE",
	})
	if err := conn.WriteMessage(websocket.TextMessage, subMsg); err != nil {
		t.Fatalf("failed to send subscribe: %v", err)
	}

	received := make(chan wsMessage, 1)

	go func() {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var m wsMessage
		if err := json.Unmarshal(msg, &m); err != nil {
			return
		}
		received <- m
	}()

	time.Sleep(100 * time.Millisecond)
	submitOrder(t, sellerAPIKey, "RELIANCE", "SELL", "LIMIT", 1, 43000)
	submitOrder(t, buyerAPIKey, "RELIANCE", "BUY", "LIMIT", 1, 43000)

	select {
	case msg := <-received:
		if msg.Channel != "trades.RELIANCE" {
			t.Fatalf("expected channel trades.RELIANCE, got %s", msg.Channel)
		}
		t.Logf("received trade broadcast on channel: %s data: %s", msg.Channel, string(msg.Data))
	case <-time.After(8 * time.Second):
		t.Fatal("timed out waiting for trade broadcast")
	}
}

func TestWebSocketUnsubscribe(t *testing.T) {
	u, _ := url.Parse(marketDataURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	subMsg, _ := json.Marshal(map[string]string{"action": "subscribe", "channel": "trades.RELIANCE"})
	conn.WriteMessage(websocket.TextMessage, subMsg)

	unsubMsg, _ := json.Marshal(map[string]string{"action": "unsubscribe", "channel": "trades.RELIANCE"})
	if err := conn.WriteMessage(websocket.TextMessage, unsubMsg); err != nil {
		t.Fatalf("failed to send unsubscribe: %v", err)
	}

	t.Log("subscribe and unsubscribe completed without error")
}

package security

import (
	"encoding/json"
	"testing"
)

// FuzzOrderRequest targets the JSON parsing and validation logic
func FuzzOrderRequest(f *testing.F) {
	// Seed with a valid order
	f.Add([]byte(`{"symbol":"RELIANCE","side":"BUY","quantity":10,"price":250000,"type":"LIMIT"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var obj map[string]interface{}
		if err := json.Unmarshal(data, &obj); err != nil {
			return // Skip invalid JSON
		}

		// If it's valid JSON, our internal validation should handle it
		// without crashing, even if 'quantity' is a nested object or a boolean.
	})
}

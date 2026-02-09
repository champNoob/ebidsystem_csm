package matching

import (
	"testing"
	"time"
)

func TestEngine_MultiSymbolIsolation(t *testing.T) {
	engine := NewEngine()
	engine.Start()

	// AAPL
	engine.Submit(&Order{
		ID:       1,
		Symbol:   "AAPL",
		Side:     OrderSideBuy,
		Price:    10,
		Quantity: 10,
	})
	engine.Submit(&Order{
		ID:       2,
		Symbol:   "AAPL",
		Side:     OrderSideSell,
		Price:    9,
		Quantity: 5,
	})

	// TSLA
	engine.Submit(&Order{
		ID:       3,
		Symbol:   "TSLA",
		Side:     OrderSideBuy,
		Price:    20,
		Quantity: 7,
	})
	engine.Submit(&Order{
		ID:       4,
		Symbol:   "TSLA",
		Side:     OrderSideSell,
		Price:    19,
		Quantity: 7,
	})

	received := 0
	timeout := time.After(time.Second)

	for received < 2 {
		select {
		case <-engine.Events():
			received++
		case <-timeout:
			t.Fatalf("expected 2 events, got %d", received)
		}
	}
}

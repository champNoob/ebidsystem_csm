package matching

import (
	"testing"
	"time"
)

func TestEngine_MultiSymbolIsolation(t *testing.T) {
	engine := NewEngine()
	engine.Start()
	defer engine.Stop()

	// AAPL
	if err := engine.Submit(&Order{
		ID:       1,
		Symbol:   "AAPL",
		Side:     OrderSideBuy,
		Price:    10,
		Quantity: 10,
	}); err != nil {
		t.Fatalf("failed to submit order: %v", err)
	}
	if err := engine.Submit(&Order{
		ID:       2,
		Symbol:   "AAPL",
		Side:     OrderSideSell,
		Price:    9,
		Quantity: 5,
	}); err != nil {
		t.Fatalf("failed to submit order: %v", err)
	}

	// TSLA
	if err := engine.Submit(&Order{
		ID:       3,
		Symbol:   "TSLA",
		Side:     OrderSideBuy,
		Price:    20,
		Quantity: 7,
	}); err != nil {
		t.Fatalf("failed to submit order: %v", err)
	}
	if err := engine.Submit(&Order{
		ID:       4,
		Symbol:   "TSLA",
		Side:     OrderSideSell,
		Price:    19,
		Quantity: 7,
	}); err != nil {
		t.Fatalf("failed to submit order: %v", err)
	}

	receivedAAPL := 0
	receivedTSLA := 0
	timeout := time.After(time.Second)

	for receivedAAPL < 1 || receivedTSLA < 1 {
		select {
		case event := <-engine.Events():
			switch event.Symbol {
			case "AAPL":
				receivedAAPL++
			case "TSLA":
				receivedTSLA++
			default:
				t.Fatalf("unexpected symbol: %s", event.Symbol)
			}
		case <-timeout:
			t.Fatalf("expected 2 events, got %d", receivedAAPL+receivedTSLA)
		}
	}
}

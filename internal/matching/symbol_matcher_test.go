package matching

import (
	"testing"
	"time"
)

func TestSymbolMatcher_MatchFlow(t *testing.T) {
	eventCh := make(chan MatchEvent, 10)
	sm := NewSymbolMatcher("AAPL", eventCh)
	sm.Start()
	defer sm.Stop()

	buy := &Order{
		ID:       1,
		Symbol:   "AAPL",
		Side:     OrderSideBuy,
		Price:    10,
		Quantity: 10,
	}
	sell := &Order{
		ID:       2,
		Symbol:   "AAPL",
		Side:     OrderSideSell,
		Price:    9,
		Quantity: 5,
	}

	sm.Submit(buy)
	sm.Submit(sell)

	select {
	case ev := <-eventCh:
		if ev.Quantity != 5 {
			t.Fatalf("expected qty=5, got %d", ev.Quantity)
		}
	case <-time.After(time.Second):
		t.Fatal("no match event received")
	}
}

func TestSymbolMatcher_Remove(t *testing.T) {
	eventCh := make(chan MatchEvent, 1)
	sm := NewSymbolMatcher("AAPL", eventCh)
	sm.Start()
	defer sm.Stop()

	order := &Order{
		ID:       1,
		Symbol:   "AAPL",
		Side:     OrderSideBuy,
		Price:    10,
		Quantity: 10,
	}

	sm.Submit(order)
	sm.Remove(order.ID)
}

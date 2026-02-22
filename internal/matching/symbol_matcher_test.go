package matching

import (
	"context"
	"ebidsystem_csm/internal/pkg/logger"
	"testing"
	"time"
)

func TestSymbolMatcher_MatchFlow(t *testing.T) {
	eventCh := make(chan MatchEvent, 10)
	eventLogger, _ := logger.NewLogger(50000, "engine/symbol_matcher_match.log", true, false)
	obMatchLogger, _ := logger.NewLogger(50000, "engine/orderbook_match.log", true, false)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sm := NewSymbolMatcher(ctx, "AAPL", eventCh, eventLogger, obMatchLogger)
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

func TestSymbolMatcher_Remove(t *testing.T) { //#
	eventCh := make(chan MatchEvent, 1)
	eventLogger, _ := logger.NewLogger(50000, "engine/symbol_matcher_match.log", true, false)
	obMatchLogger, _ := logger.NewLogger(50000, "engine/orderbook_match.log", true, false)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sm := NewSymbolMatcher(ctx, "AAPL", eventCh, eventLogger, obMatchLogger)
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

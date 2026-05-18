package matching

import (
	"context"
	"ebidsystem_csm/internal/pkg/logger"
	"testing"
	"time"
)

func newSymbolMatcherTestLoggers(t *testing.T) (*logger.Logger, *logger.Logger) {
	t.Helper()
	eventLogger, err := logger.NewLogger(50000, "engine/symbol_matcher_match_test.log", true, false)
	if err != nil {
		t.Fatalf("create event logger failed: %v", err)
	}
	obMatchLogger, err := logger.NewLogger(50000, "engine/orderbook_match_test.log", true, false)
	if err != nil {
		eventLogger.Close()
		t.Fatalf("create orderbook logger failed: %v", err)
	}
	return eventLogger, obMatchLogger
}

func TestSymbolMatcher_MatchFlow(t *testing.T) {
	eventCh := make(chan MatchEvent, 10)
	eventLogger, obMatchLogger := newSymbolMatcherTestLoggers(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sm := NewSymbolMatcher(ctx, "AAPL", eventCh, eventLogger, obMatchLogger)
	sm.Start()
	defer func() {
		sm.Stop()
		eventLogger.Close()
		obMatchLogger.Close()
	}()

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
		if ev.Symbol != "AAPL" || ev.BuyOrderID != 1 || ev.SellOrderID != 2 {
			t.Fatalf("unexpected event: %+v", ev)
		}
		if ev.Quantity != 5 {
			t.Fatalf("expected qty=5, got %d", ev.Quantity)
		}
	case <-time.After(time.Second):
		t.Fatal("no match event received")
	}
}

func TestSymbolMatcher_Remove_PreventsFutureMatch(t *testing.T) {
	eventCh := make(chan MatchEvent, 10)
	eventLogger, obMatchLogger := newSymbolMatcherTestLoggers(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sm := NewSymbolMatcher(ctx, "AAPL", eventCh, eventLogger, obMatchLogger)
	sm.Start()
	defer func() {
		sm.Stop()
		eventLogger.Close()
		obMatchLogger.Close()
	}()

	buy := &Order{ID: 1, Symbol: "AAPL", Side: OrderSideBuy, Price: 10, Quantity: 10}
	sm.Submit(buy)

	// 等待订单进入SymbolMatcher内部订单簿，避免Remove早于AddOrder执行：
	deadline := time.After(time.Second)
	for {
		if len(sm.book.buyOrders) == 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("buy order was not added to order book in time")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	sm.Remove(buy.ID)
	time.Sleep(20 * time.Millisecond)

	sm.Submit(&Order{
		ID:       2,
		Symbol:   "AAPL",
		Side:     OrderSideSell,
		Price:    9,
		Quantity: 10,
	})

	select {
	case ev := <-eventCh:
		t.Fatalf("expected no match event after remove, got %+v", ev)
	case <-time.After(200 * time.Millisecond):
		// pass
	}
}

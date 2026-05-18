package matching

import (
	"ebidsystem_csm/internal/pkg/logger"
	"testing"
)

func newOrderBookTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	l, err := logger.NewLogger(LOG_BUFFER_SIZE, "engine/orderbook_match.log", true, false)
	if err != nil {
		t.Fatalf("create logger failed: %v", err)
	}
	t.Cleanup(l.Close)
	return l
}

// 测试订单簿的常规撮合：
func TestOrderBook_Match_Simple(t *testing.T) {
	ob := NewOrderBook()
	obMatchLogger := newOrderBookTestLogger(t)

	buy := &Order{
		ID:       1,
		Side:     OrderSideBuy,
		Price:    10,
		Quantity: 10,
	}
	sell := &Order{
		ID:       2,
		Side:     OrderSideSell,
		Price:    9,
		Quantity: 7,
	}
	ob.AddOrder(buy)
	ob.AddOrder(sell)

	events := ob.Match(obMatchLogger)
	if len(events) != 1 {
		t.Fatalf("expected 1 match event, got %d", len(events))
	}

	ev := events[0]
	if ev.BuyOrderID != 1 || ev.SellOrderID != 2 {
		t.Fatalf("unexpected order IDs: %+v", ev)
	}
	if ev.Price != 9 {
		t.Fatalf("expected match price=sell price 9, got %.2f", ev.Price)
	}
	if ev.Quantity != 7 {
		t.Fatalf("expected match qty=7, got %d", ev.Quantity)
	}
	if buy.Remaining != 3 {
		t.Fatalf("expected buy remaining=3, got %d", buy.Remaining)
	}
	if sell.Remaining != 0 {
		t.Fatalf("expected sell remaining=0, got %d", sell.Remaining)
	}
}

// 验证最优买价低于最优卖价时不成交，买卖双方Remaining不变化：
func TestOrderBook_Match_NoMatchWhenBuyPriceBelowSellPrice(t *testing.T) {
	ob := NewOrderBook()
	obMatchLogger := newOrderBookTestLogger(t)

	buy := &Order{
		ID:       1,
		Symbol:   "AAPL",
		Side:     OrderSideBuy,
		Price:    9,
		Quantity: 10,
	}
	sell := &Order{
		ID:       2,
		Symbol:   "AAPL",
		Side:     OrderSideSell,
		Price:    10,
		Quantity: 10,
	}

	ob.AddOrder(buy)
	ob.AddOrder(sell)

	events := ob.Match(obMatchLogger)
	if len(events) != 0 {
		t.Fatalf("expected no match event, got %d", len(events))
	}
	if buy.Remaining != 10 || sell.Remaining != 10 {
		t.Fatalf("remaining quantity should not change, buy=%d sell=%d", buy.Remaining, sell.Remaining)
	}
}

// 验证同价位下Seq较小的订单优先成交：
func TestOrderBook_Match_TimePriorityBySeq(t *testing.T) {
	ob := NewOrderBook()
	obMatchLogger := newOrderBookTestLogger(t)

	lateBuy := &Order{ID: 2,
		Symbol:   "AAPL",
		Side:     OrderSideBuy,
		Price:    10,
		Quantity: 1,
		Seq:      2,
	}
	earlyBuy := &Order{
		ID:       1,
		Symbol:   "AAPL",
		Side:     OrderSideBuy,
		Price:    10,
		Quantity: 1,
		Seq:      1,
	}
	sell := &Order{
		ID:       3,
		Symbol:   "AAPL",
		Side:     OrderSideSell,
		Price:    9,
		Quantity: 1,
		Seq:      3,
	}

	// 刻意先插入Seq较大的订单，再插入Seq较小的订单，验证同价位下按Seq排序。
	ob.AddOrder(lateBuy)
	ob.AddOrder(earlyBuy)
	ob.AddOrder(sell)

	events := ob.Match(obMatchLogger)
	if len(events) != 1 {
		t.Fatalf("expected 1 match event, got %d", len(events))
	}
	if events[0].BuyOrderID != earlyBuy.ID {
		t.Fatalf("expected early buy order %d to match first, got %d", earlyBuy.ID, events[0].BuyOrderID)
	}
}

// 验证一个大买单连续吃掉多个小卖单时，事件数量和成交数量正确：
func TestOrderBook_Match_LargeOrderAgainstMultipleSmallOrders(t *testing.T) {
	ob := NewOrderBook()
	obMatchLogger := newOrderBookTestLogger(t)

	buy := &Order{
		ID:       1,
		Symbol:   "AAPL",
		Side:     OrderSideBuy,
		Price:    10,
		Quantity: 3,
		Seq:      1,
	}
	sell1 := &Order{
		ID:       2,
		Symbol:   "AAPL",
		Side:     OrderSideSell,
		Price:    9,
		Quantity: 1,
		Seq:      2,
	}
	sell2 := &Order{
		ID:       3,
		Symbol:   "AAPL",
		Side:     OrderSideSell,
		Price:    9,
		Quantity: 2,
		Seq:      3,
	}

	ob.AddOrder(buy)
	ob.AddOrder(sell1)
	ob.AddOrder(sell2)

	events := ob.Match(obMatchLogger)
	if len(events) != 2 {
		t.Fatalf("expected 2 match events, got %d", len(events))
	}
	if events[0].Quantity != 1 || events[1].Quantity != 2 {
		t.Fatalf("unexpected quantities: %+v", events)
	}
	if buy.Remaining != 0 || sell1.Remaining != 0 || sell2.Remaining != 0 {
		t.Fatalf("expected all orders fully filled, buy=%d sell1=%d sell2=%d",
			buy.Remaining, sell1.Remaining, sell2.Remaining)
	}
}

package matching

import "testing"

func TestOrderBook_Match_Simple(t *testing.T) {
	ob := NewOrderBook()

	buy := &Order{
		ID:        1,
		Side:      OrderSideBuy,
		Price:     10,
		Quantity:  10,
		Remaining: 10,
	}
	sell := &Order{
		ID:        2,
		Side:      OrderSideSell,
		Price:     9,
		Quantity:  7,
		Remaining: 7,
	}

	ob.AddOrder(buy)
	ob.AddOrder(sell)

	events := ob.Match()

	if len(events) != 1 {
		t.Fatalf("expected 1 match event, got %d", len(events))
	}

	ev := events[0]
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

package matching

import (
	"context"
	"log"

	"github.com/google/uuid"
)

type SymbolMatcher struct {
	symbol   string
	orderCh  chan *Order
	removeCh chan uint64
	book     *OrderBook

	eventCh chan<- MatchEvent

	ctx    context.Context
	cancel context.CancelFunc
}

func NewSymbolMatcher(
	symbol string,
	eventCh chan<- MatchEvent,
) *SymbolMatcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &SymbolMatcher{
		symbol:   symbol,
		orderCh:  make(chan *Order, 1024),
		removeCh: make(chan uint64, 1024),
		book:     NewOrderBook(),
		eventCh:  eventCh,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (sm *SymbolMatcher) Start() {
	go func() {
		for {
			select {
			case <-sm.ctx.Done():
				log.Printf("[SYMBOL_MATCHER_STOP] symbol=%s", sm.symbol)
				return

			case order := <-sm.orderCh:
				sm.book.AddOrder(order)
				sm.matchAndEmit()

			case orderID := <-sm.removeCh:
				sm.book.Remove(orderID)
			}
		}
	}()
}

func (sm *SymbolMatcher) Stop() {
	sm.cancel()
}

func (sm *SymbolMatcher) Submit(order *Order) {
	sm.orderCh <- order
}

func (sm *SymbolMatcher) Remove(orderID uint64) {
	sm.removeCh <- orderID
}

func (sm *SymbolMatcher) matchAndEmit() {
	events := sm.book.Match()

	for _, ev := range events {
		log.Printf(
			"[MATCH] symbol=%s buy=%d sell=%d qty=%d price=%.2f",
			sm.symbol,
			ev.BuyOrderID,
			ev.SellOrderID,
			ev.Quantity,
			ev.Price,
		)
		// 生成事件 ID：
		ev.EventID = uuid.NewString() // 或 snowflake
		// 事件输出（由 Engine fan-in）：
		sm.eventCh <- ev
	}
}

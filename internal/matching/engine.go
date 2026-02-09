package matching

import (
	"log"
)

type Engine struct {
	orderCh  chan *Order
	eventCh  chan MatchEvent
	matchers map[string]*SymbolMatcher
}

func NewEngine() *Engine {
	return &Engine{
		orderCh:  make(chan *Order, 1024),
		eventCh:  make(chan MatchEvent, 1024),
		matchers: make(map[string]*SymbolMatcher),
	}
}

func (e *Engine) Start() {
	go func() {
		for order := range e.orderCh {
			matcher, ok := e.matchers[order.Symbol]
			if !ok {
				matcher = NewSymbolMatcher(order.Symbol, e.eventCh)
				matcher.Start()
				e.matchers[order.Symbol] = matcher
			}
			matcher.Submit(order)
			log.Printf(
				"[MATCH] symbol=%s buySideID=%d sellSideID=%d price=%.2f quantity=%d",
				order.Symbol,
				order.ID,
				order.ID,
				order.Price,
				order.Quantity,
			)
		}
	}()
}

func (e *Engine) Submit(order *Order) error {
	if order.Type == OrderTypeMarket {
		return ErrMarketOrderNotSupported
	}
	e.orderCh <- order
	return nil
}

func (e *Engine) Remove(orderID uint64, symbol string) {
	sm, ok := e.matchers[symbol]
	if !ok {
		return
	}
	sm.Remove(orderID)
}

func (e *Engine) Events() <-chan MatchEvent {
	return e.eventCh
}

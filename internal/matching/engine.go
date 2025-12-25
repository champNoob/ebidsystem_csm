package matching

import (
	"log"
)

type Engine struct {
	orderCh chan *Order
	books   map[string]*OrderBook
}

func NewEngine() *Engine {
	return &Engine{
		orderCh: make(chan *Order, 1024),
		books:   make(map[string]*OrderBook),
	}
}

func (e *Engine) Start() {
	go func() {
		for order := range e.orderCh {
			ob, ok := e.books[order.Symbol]
			if !ok {
				ob = NewOrderBook()
				e.books[order.Symbol] = ob
			}

			ob.AddOrder(order)
			events := ob.Match()

			for _, ev := range events {
				log.Printf(
					"[MATCH] symbol=%s buy=%d sell=%d price=%.2f qty=%d",
					order.Symbol,
					ev.BuyOrderID,
					ev.SellOrderID,
					ev.Price,
					ev.Quantity,
				)
			}
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

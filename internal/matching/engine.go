package matching

import (
	"context"
	"ebidsystem_csm/internal/middleware"
	"fmt"
	"log"
	"sync"
)

type Engine struct {
	orderCh  chan *Order
	eventCh  chan MatchEvent
	matchers map[string]*SymbolMatcher

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
}

func NewEngine() *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		orderCh:  make(chan *Order, 1024),
		eventCh:  make(chan MatchEvent, 1024),
		matchers: make(map[string]*SymbolMatcher),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (e *Engine) Start() {
	e.wg.Add(1)

	go func() {
		defer e.wg.Done()

		for {
			select {
			case <-e.ctx.Done():
				log.Printf("[MATCHING_ENGINE_STOP]")
				return

			case order, ok := <-e.orderCh:
				if !ok {
					return //channel 已关闭，退出循环
				}
				matcher, ok := e.matchers[order.Symbol]
				if !ok {
					matcher = NewSymbolMatcher(e.ctx, order.Symbol, e.eventCh)
					matcher.Start()
					e.matchers[order.Symbol] = matcher
				}
				matcher.Submit(order)
				message := fmt.Sprintf(
					"[ENGINE_SUBMIT] symbol=%s side=%s ID=%d price=%.2f remaining=%d",
					order.Symbol,
					order.Side,
					order.ID,
					order.Price,
					order.Remaining,
				)
				middleware.PrintLog(false, message, "log/orders/engine_submit.txt")
			}
		}
	}()
}

func (e *Engine) Stop() {
	// 先关闭订单通道：
	close(e.orderCh)
	// 再停止接受新订单：
	e.cancel()
	// 再等待引擎协程退出：
	e.wg.Wait()
	// 然后关闭所有撮合器：
	for _, sm := range e.matchers {
		sm.Stop()
	}
	// 最后关闭事件通道：
	close(e.eventCh)
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

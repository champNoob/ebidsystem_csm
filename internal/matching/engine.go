package matching

import (
	"context"
	"ebidsystem_csm/internal/pkg/logger"
	"fmt"
	"sync"
)

const (
	LOG_BUFFER_SIZE = 100000
)

type Engine struct {
	orderCh  chan *Order
	eventCh  chan MatchEvent
	matchers map[string]*SymbolMatcher

	submitLogger  *logger.Logger
	eventLogger   *logger.Logger
	obMatchLogger *logger.Logger

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
}

func NewEngine() *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	// 创建订单提交日志实例：
	submitLogger, err := logger.NewLogger(
		LOG_BUFFER_SIZE,
		"engine/engine_submit.log",
		true,
		false,
	)
	if err != nil {
		panic(fmt.Errorf("撮合引擎：创建订单提交日志失败: %w", err))
	}
	// 创建撮合事件日志实例：
	eventLogger, err := logger.NewLogger(
		LOG_BUFFER_SIZE,
		"engine/symbol_matcher_match.log",
		true,
		false,
	)
	if err != nil {
		panic(fmt.Errorf("撮合引擎：创建撮合事件日志失败: %w", err))
	}
	// 创建订单簿撮合日志实例：
	obMatchLogger, err := logger.NewLogger(
		LOG_BUFFER_SIZE,
		"engine/orderbook_match.log",
		true,
		false,
	)
	if err != nil {
		panic(fmt.Errorf("撮合引擎：创建订单簿撮合日志失败: %w", err))
	}

	return &Engine{
		orderCh:       make(chan *Order, 1024),
		eventCh:       make(chan MatchEvent, 1024),
		matchers:      make(map[string]*SymbolMatcher),
		ctx:           ctx,
		cancel:        cancel,
		submitLogger:  submitLogger,
		eventLogger:   eventLogger,
		obMatchLogger: obMatchLogger,
	}
}

func (e *Engine) Start() {
	e.wg.Add(1)

	go func() {
		defer e.wg.Done()

		for {
			select {
			case <-e.ctx.Done():
				e.submitLogger.Log("[MATCHING_ENGINE_STOP]")
				return

			case order, ok := <-e.orderCh:
				if !ok {
					return //channel 已关闭，退出循环
				}
				matcher, ok := e.matchers[order.Symbol]
				if !ok {
					matcher = NewSymbolMatcher(
						e.ctx,
						order.Symbol,
						e.eventCh,
						e.eventLogger,
						e.obMatchLogger,
					)
					matcher.Start()
					e.matchers[order.Symbol] = matcher
				}
				matcher.Submit(order)
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
	// 向各日志实例发送停止信号：
	e.submitLogger.Close()
	e.eventLogger.Close()
	e.obMatchLogger.Close()
}

func (e *Engine) Submit(order *Order) error {
	if order.Type == OrderTypeMarket {
		return ErrMarketOrderNotSupported
	}
	// 输出日志：
	message := fmt.Sprintf(
		"[ENGINE_SUBMIT] symbol=%s side=%s ID=%d price=%.2f quantity=%d",
		order.Symbol,
		order.Side,
		order.ID,
		order.Price,
		order.Quantity,
	)
	e.submitLogger.Log(message)

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

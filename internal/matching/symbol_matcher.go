package matching

import (
	"context"
	"ebidsystem_csm/internal/pkg/logger"
	"fmt"
	"sync"
	"time"
)

var (
	droppedOrderNum int64
)

type SymbolMatcher struct {
	symbol   string
	orderCh  chan *Order
	removeCh chan uint64
	book     *OrderBook

	eventCh       chan<- MatchEvent
	eventLogger   *logger.Logger
	obMatchLogger *logger.Logger

	seq      int64  //订单时间优先级（用于FIFO排序）
	eventSeq uint64 //撮合事件编号（递增）

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewSymbolMatcher(
	parentCtx context.Context,
	symbol string,
	eventCh chan<- MatchEvent,
	eventLogger *logger.Logger,
	obMatchLogger *logger.Logger,
) *SymbolMatcher {
	ctx, cancel := context.WithCancel(parentCtx)

	return &SymbolMatcher{
		symbol:        symbol,
		book:          NewOrderBook(),
		orderCh:       make(chan *Order, 1024),
		removeCh:      make(chan uint64, 1024),
		eventCh:       eventCh,
		eventLogger:   eventLogger,
		obMatchLogger: obMatchLogger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (sm *SymbolMatcher) Start() {
	sm.wg.Add(1)

	go func() {
		defer sm.wg.Done()

		for {
			select {
			case <-sm.ctx.Done():
				sm.eventLogger.Log(fmt.Sprintf("[SYMBOL_MATCHER_STOP] symbol=%s", sm.symbol))
				return

			case orderID, ok := <-sm.removeCh: //优先响应撤单
				if !ok {
					return
				}
				sm.book.Remove(orderID)

			default: //#公平竞争撤单和下单
				select {
				case order, ok := <-sm.orderCh:
					if !ok {
						return
					}
					sm.book.AddOrder(order)
					sm.matchAndEmit()
				case orderID, ok := <-sm.removeCh:
					if !ok {
						return
					}
					sm.book.Remove(orderID)
				}
			}
		}
	}()
}

func (sm *SymbolMatcher) Stop() {
	sm.eventLogger.Log(fmt.Sprintf(
		"[SYMBOL_MATCHER_STOP] symbol=%s droppedOrderNum=%d",
		sm.symbol,
		droppedOrderNum,
	))
	sm.cancel()
	close(sm.orderCh)
	close(sm.removeCh)
	sm.wg.Wait()
}

func (sm *SymbolMatcher) Submit(order *Order) {
	sm.seq++
	order.Seq = sm.seq
	/* 非阻塞写法会丢单（约30%）：
	select {
	case sm.orderCh <- order:
	case <-sm.ctx.Done():
		return
	default:
		atomic.AddInt64(&droppedOrderNum, 1) //统计丢单
	}
	*/
	sm.orderCh <- order //阻塞架构
}

func (sm *SymbolMatcher) Remove(orderID uint64) {
	sm.removeCh <- orderID
}

func (sm *SymbolMatcher) matchAndEmit() {
	events := sm.book.Match(sm.obMatchLogger)

	for _, ev := range events {
		ev.EventID = sm.nextEventID() //撮合引擎直接生成事件ID
		ev.Symbol = sm.symbol

		// 输出日志：
		message := fmt.Sprintf("[SM_MATCH] symbol=%s buyID=%d sellID=%d matchQty=%d price=%.2f",
			sm.symbol,
			ev.BuyOrderID,
			ev.SellOrderID,
			ev.Quantity,
			ev.Price,
		)
		sm.eventLogger.Log(message)
		// 事件输出（由 Engine fan-in）：
		sm.eventCh <- ev
	}
}

func (sm *SymbolMatcher) nextEventID() string {
	sm.eventSeq++
	return fmt.Sprintf("%s-%d-%d",
		sm.symbol,
		time.Now().UnixNano(),
		sm.eventSeq,
	)
}

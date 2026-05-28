package service

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/apperror"
	"ebidsystem_csm/internal/matching"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/pkg/logger"
	"ebidsystem_csm/internal/repository"
	"fmt"
	"log"
)

type OrderService struct {
	repo             repository.OrderRepository
	matcher          *matching.Engine
	matchEventLogger *logger.Logger
	ctx              context.Context
	cancelFunc       context.CancelFunc
}

func NewOrderService(repo repository.OrderRepository, matcher *matching.Engine) *OrderService {
	matchEventLogger, err := logger.NewLogger(
		10000,
		"order/match_event_error.log",
		true,
		false,
	)
	if err != nil {
		matchEventLogger = nil //日志初始化失败不应阻止订单服务创建
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &OrderService{
		repo:             repo,
		matcher:          matcher,
		matchEventLogger: matchEventLogger,
		ctx:              ctx,
		cancelFunc:       cancel,
	}
}

func (s *OrderService) Close() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	if s.matchEventLogger != nil {
		s.matchEventLogger.Close()
	}
}

// CreateOrder 下单
func (s *OrderService) CreateOrder(
	ctx context.Context,
	userID int64,
	role model.UserRole,
	symbol string,
	orderType model.OrderType,
	orderSide model.OrderSide,
	price *float64,
	quantity int64,
) error {
	// 角色×方向 校验：
	if err := validateRoleSide(role, orderSide); err != nil {
		return err
	}
	// 判断订单类型：
	switch orderType {
	case model.OrderTypeLimit:
		if price == nil {
			return apperror.ErrOrderLimitWithoutPrice
		}
	case model.OrderTypeMarket:
		if price != nil {
			return apperror.ErrOrderMarketWithPrice
		}
	default:
		return apperror.ErrOrderInvalidType
	}

	order := &model.Order{
		UserID:   userID,
		Symbol:   symbol,
		Type:     orderType,
		Side:     orderSide,
		Price:    price,
		Quantity: quantity,
		Status:   model.OrderStatusPending,
	}

	orderID, err := s.repo.Create(ctx, order)
	if err != nil {
		return err
	}
	// 向撮合引擎递交订单
	matchingOrder := &matching.Order{
		ID:       orderID,
		UserID:   uint64(userID),
		Symbol:   order.Symbol,
		Type:     matching.OrderType(order.Type),
		Side:     matching.OrderSide(order.Side),
		Price:    *order.Price,
		Quantity: order.Quantity,
	}
	if err := s.matcher.Submit(matchingOrder); err != nil {
		return err
	}

	return nil
}

// ListOrders 查询订单
func (s *OrderService) ListOrders(
	ctx context.Context,
	userID int64,
	role string,
	status string,
) ([]*model.Order, error) {

	statuses, err := parseOrderQueryStatus(status)
	if err != nil {
		return nil, err
	}

	switch role {
	case "admin":
		return s.repo.FindAll(ctx, statuses)

	case "trader":
		if status == "" || status == "current" {
			return s.repo.FindAll(ctx, []model.OrderStatus{
				model.OrderStatusPending,
				model.OrderStatusPartial,
			})
		}
		return nil, apperror.ErrPermissionDenied

	case "sales":
		return s.repo.FindByUserID(ctx, userID, statuses) //#未来需要区分userID和tgtID

	case "client", "seller":
		return s.repo.FindByUserID(ctx, userID, statuses)

	default:
		return nil, apperror.ErrPermissionDenied
	}
}

func parseOrderQueryStatus(s string) ([]model.OrderStatus, error) {
	switch s {
	case "", "all":
		return nil, nil // nil = 不加过滤
	case "current":
		return []model.OrderStatus{
			model.OrderStatusPending,
			model.OrderStatusPartial,
		}, nil
	case "history":
		return []model.OrderStatus{
			model.OrderStatusFilled,
			model.OrderStatusCanceled,
		}, nil
	default:
		return nil, apperror.ErrInvalidOrderStatusQuery
	}
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(
	ctx context.Context,
	orderID int64,
	userID int64,
	role string,
) error {
	// 0. 查询订单：
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return apperror.ErrOrderNotFound
	}

	// 1. 权限校验：
	if role != "admin" && order.UserID != userID {
		return apperror.ErrPermissionDenied
	}

	// 2. 状态校验：
	if !order.Status.CanCancel() { // 订单强类型
		return apperror.ErrOrderNotCancellable
	}

	// 3. 执行撤单（原子）：
	if err := s.repo.CancelOrder(ctx, uint64(orderID)); err != nil {
		return err
	}

	// 4. 通知撮合引擎：
	s.matcher.Remove(uint64(orderID), order.Symbol)
	return nil
}

/*
	撮合引擎部分
*/

// 启动撮合事件监听器：
func (s *OrderService) StartMatchEventListener() {
	go func() {
		for {
			select {
			case ev, ok := <-s.matcher.Events():
				if !ok {
					s.logMatchEvent("[MATCH_EVENT_LISTENER_STOP] event channel closed")
					return
				}
				if err := s.handleMatchEvent(s.ctx, ev); err != nil {
					if be, ok := err.(*apperror.BusinessError); ok {
						s.logMatchEvent(fmt.Sprintf(
							"[MATCH_EVENT_ERROR] eventID=%s symbol=%s buyID=%d sellID=%d code=%s msg=%s",
							ev.EventID,
							ev.Symbol,
							ev.BuyOrderID,
							ev.SellOrderID,
							be.Code,
							be.Message,
						))
					} else {
						s.logMatchEvent(fmt.Sprintf(
							"[MATCH_EVENT_ERROR] eventID=%s symbol=%s buyID=%d sellID=%d err=%v",
							ev.EventID,
							ev.Symbol,
							ev.BuyOrderID,
							ev.SellOrderID,
							err,
						))
					}
				}
			case <-s.ctx.Done():
				s.logMatchEvent("[MATCH_EVENT_LISTENER_STOP] context canceled")
				return
			}
		}
	}()
}

// 处理撮合事件（幂等）：
func (s *OrderService) handleMatchEvent(
	ctx context.Context,
	ev matching.MatchEvent,
) error {

	return s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		// 0. 幂等门闸
		ok, err := s.repo.InsertMatchEventTx(ctx, tx, ev.EventID, ev.Symbol, ev.BuyOrderID, ev.SellOrderID, ev.Quantity, ev.Price)
		if err != nil {
			return err
		}
		if !ok { //幂等命中：整个事件已经处理过
			return nil
		}
		// 1. 买单
		if err := s.repo.FillOrderTx(
			ctx, tx, ev.Symbol, ev.BuyOrderID, ev.Quantity,
		); err != nil {
			return err
		}

		// 2. 卖单
		if err := s.repo.FillOrderTx(
			ctx, tx, ev.Symbol, ev.SellOrderID, ev.Quantity,
		); err != nil {
			return err
		}

		// 3. 成交（幂等）
		trade := &model.Trade{
			EventID:     ev.EventID,
			Symbol:      ev.Symbol,
			BuyOrderID:  ev.BuyOrderID,
			SellOrderID: ev.SellOrderID,
			Price:       ev.Price,
			Quantity:    ev.Quantity,
		}
		return s.repo.CreateTradeTx(ctx, tx, trade)
	})
}

// 引擎重启恢复订单：
func (s *OrderService) RecoverActiveOrders(ctx context.Context) error {
	// 检查脏订单：
	dirtyOrders, err := s.repo.FindDirtyOrdersForRecovery(ctx)
	if err != nil {
		return err
	}

	for _, o := range dirtyOrders {
		s.logMatchEvent(fmt.Sprintf(
			"[RECOVERY_DIRTY_ORDER] id=%d symbol=%s status=%s quantity=%d filled=%d reason=%s",
			o.ID,
			o.Symbol,
			o.Status,
			o.Quantity,
			o.FilledQuantity,
			o.Reason,
		))
	}

	// 恢复活跃订单：
	orders, err := s.repo.FindActiveOrdersForRecovery(ctx)
	if err != nil {
		return err
	}

	recovered := 0
	skipped := 0

	for _, o := range orders {
		if o.ID <= 0 || o.UserID <= 0 { //跳过无效订单
			skipped++
			continue
		}

		if o.Price == nil { //跳过价格为空的订单
			skipped++
			continue
		}

		remaining := o.Quantity - o.FilledQuantity
		if remaining <= 0 {
			skipped++
			continue
		}

		matchingOrder := &matching.Order{
			ID:        uint64(o.ID),
			UserID:    uint64(o.UserID),
			Symbol:    o.Symbol,
			Type:      matching.OrderType(o.Type),
			Side:      matching.OrderSide(o.Side),
			Price:     *o.Price,
			Quantity:  o.Quantity,
			Remaining: remaining,
		}

		if err := s.matcher.Submit(matchingOrder); err != nil {
			s.logMatchEvent(fmt.Sprintf(
				"[RECOVER_ORDER_ERROR] orderID=%d symbol=%s side=%s remaining=%d err=%v",
				o.ID,
				o.Symbol,
				o.Side,
				remaining,
				err,
			))
			return err
		}

		recovered++
	}

	s.logMatchEvent(fmt.Sprintf(
		"[RECOVER_ORDERS_DONE] recovered=%d skipped=%d dirty=%d",
		recovered,
		skipped,
		len(dirtyOrders),
	))

	return nil
}

// 防止 logger 初始化失败，提供备用方案：
func (s *OrderService) logMatchEvent(msg string) {
	if s.matchEventLogger != nil {
		s.matchEventLogger.Log(msg)
		return
	}
	log.Println(msg)
}

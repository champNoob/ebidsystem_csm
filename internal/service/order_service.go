package service

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/matching"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/repository"
	"log"
)

type OrderService struct {
	repo    repository.OrderRepository
	matcher *matching.Engine
}

func NewOrderService(repo repository.OrderRepository, matcher *matching.Engine) *OrderService {
	return &OrderService{repo: repo, matcher: matcher}
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
			return ErrOrderLimitWithoutPrice
		}
	case model.OrderTypeMarket:
		if price != nil {
			return ErrOrderMarketWithPrice
		}
	default:
		return ErrOrderInvalidType
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
	s.matcher.Submit(matchingOrder)

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
	case "admin", "trader":
		return s.repo.FindAll(ctx, statuses)

	case "client", "seller":
		return s.repo.FindByUserID(ctx, userID, statuses)

	default:
		return nil, ErrPermissionDenied
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
		return nil, ErrInvalidOrderStatusQuery
	}
}

func (s *OrderService) CancelOrder(
	ctx context.Context,
	orderID int64,
	userID int64,
	role string,
) error {
	// 0. 查询订单：
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return ErrOrderNotFound
	}

	// 1. 权限校验：
	if role != "admin" && order.UserID != userID {
		return ErrPermissionDenied
	}

	// 2. 状态校验：
	if !order.Status.CanCancel() { // 订单强类型
		return ErrOrderNotCancellable
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
		ctx := context.Background()
		for {
			select {
			case ev := <-s.matcher.Events():
				log.Print("matching event catched") //--
				if err := s.handleMatchEvent(ctx, ev); err != nil {
					log.Printf("[MATCH_EVENT_ERROR] %v", err)
				}
			case <-ctx.Done():
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
		ok, err := s.repo.InsertMatchEventTx(ctx, tx, ev.EventID)
		if err != nil {
			return err
		}
		if !ok { //幂等命中：整个事件已经处理过
			return nil
		}
		// 1. 买单
		if err := s.repo.FillOrderTx(
			ctx, tx, ev.BuyOrderID, ev.Quantity,
		); err != nil {
			return err
		}

		// 2. 卖单
		if err := s.repo.FillOrderTx(
			ctx, tx, ev.SellOrderID, ev.Quantity,
		); err != nil {
			return err
		}

		// 3. 成交（幂等）
		trade := &model.Trade{
			EventID:     ev.EventID,
			BuyOrderID:  ev.BuyOrderID,
			SellOrderID: ev.SellOrderID,
			Price:       ev.Price,
			Quantity:    ev.Quantity,
		}
		return s.repo.CreateTradeTx(ctx, tx, trade)
	})
}

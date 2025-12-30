package model

// 订单状态（强类型）
type OrderStatus string

const (
	OrderStatusPending  OrderStatus = "pending"  // 已下单，未成交
	OrderStatusFilled   OrderStatus = "filled"   // 已完全成交
	OrderStatusCanceled OrderStatus = "canceled" // 已撤单
	OrderStatusPartial  OrderStatus = "partial"  // 部分成交
)

func (s OrderStatus) CanCancel() bool {
	return s == OrderStatusPending || s == OrderStatusPartial
}

func (s OrderStatus) IsFinal() bool {
	return s == OrderStatusFilled || s == OrderStatusCanceled
}

package matching

// 简化：使用 slice，后期可换 heap / skiplist
type OrderBook struct {
	buyOrders  []*Order // 按价格 desc
	sellOrders []*Order // 按价格 asc
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		buyOrders:  make([]*Order, 0),
		sellOrders: make([]*Order, 0),
	}
}

func (ob *OrderBook) AddOrder(o *Order) {
	if o.Side == Buy {
		ob.buyOrders = append(ob.buyOrders, o)
	} else {
		ob.sellOrders = append(ob.sellOrders, o)
	}
}

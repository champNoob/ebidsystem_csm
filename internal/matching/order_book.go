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
	if o.Side == OrderSideBuy {
		ob.buyOrders = append(ob.buyOrders, o)
	} else {
		ob.sellOrders = append(ob.sellOrders, o)
	}
	o.Remaining = o.Quantity
}

func (ob *OrderBook) Remove(orderID uint64) {
	ob.buyOrders = filterOrders(ob.buyOrders, orderID)
	ob.sellOrders = filterOrders(ob.sellOrders, orderID)
}

func filterOrders(orders []*Order, id uint64) []*Order {
	res := orders[:0]
	for _, o := range orders {
		if o.ID != id {
			res = append(res, o)
		}
	}
	return res
}

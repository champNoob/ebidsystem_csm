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
	if o.Remaining <= 0 {
		o.Remaining = o.Quantity
	}

	if o.Side == OrderSideBuy {
		ob.buyOrders = insertBuy(ob.buyOrders, o)
	} else {
		ob.sellOrders = insertSell(ob.sellOrders, o)
	}
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

func insertBuy(orders []*Order, o *Order) []*Order {
	idx := 0
	for idx < len(orders) {
		if o.Price > orders[idx].Price {
			break
		}
		if o.Price == orders[idx].Price && o.Seq < orders[idx].Seq {
			break
		}
		idx++
	}

	orders = append(orders, nil)
	copy(orders[idx+1:], orders[idx:])
	orders[idx] = o
	return orders
}

func insertSell(orders []*Order, o *Order) []*Order {
	idx := 0
	for idx < len(orders) {
		if o.Price < orders[idx].Price {
			break
		}
		if o.Price == orders[idx].Price && o.Seq < orders[idx].Seq {
			break
		}
		idx++
	}

	orders = append(orders, nil)
	copy(orders[idx+1:], orders[idx:])
	orders[idx] = o
	return orders
}

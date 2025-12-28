package matching

func (ob *OrderBook) Match() []MatchEvent {
	events := make([]MatchEvent, 0)

	for len(ob.buyOrders) > 0 && len(ob.sellOrders) > 0 {
		buy := ob.buyOrders[0]
		sell := ob.sellOrders[0]
		// 若价格不满足，则停止撮合：
		if buy.Price < sell.Price {
			break
		}
		// 本次成交量 = min(剩余量)：
		qty := min(buy.Remaining, sell.Remaining)
		events = append(events, MatchEvent{
			BuyOrderID:  buy.ID,
			SellOrderID: sell.ID,
			Price:       sell.Price,
			Quantity:    qty,
		})
		// 扣减剩余量：
		buy.Remaining -= qty
		sell.Remaining -= qty
		// 若买单或卖单吃完，则移出订单簿：
		if buy.Remaining == 0 {
			ob.buyOrders = ob.buyOrders[1:]
		}
		if sell.Remaining == 0 {
			ob.sellOrders = ob.sellOrders[1:]
		}
	}

	return events
}

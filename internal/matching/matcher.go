package matching

func (ob *OrderBook) Match() []MatchEvent {
	events := make([]MatchEvent, 0)

	for len(ob.buyOrders) > 0 && len(ob.sellOrders) > 0 {
		buy := ob.buyOrders[0]
		sell := ob.sellOrders[0]

		if buy.Price < sell.Price {
			break
		}

		// 成交（简化：全量成交）
		events = append(events, MatchEvent{
			BuyOrderID:  buy.ID,
			SellOrderID: sell.ID,
			Price:       sell.Price,
			Quantity:    buy.Quantity,
		})

		ob.buyOrders = ob.buyOrders[1:]
		ob.sellOrders = ob.sellOrders[1:]
	}

	return events
}

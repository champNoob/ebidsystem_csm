package matching

type MatchEvent struct {
	BuyOrderID  uint64
	SellOrderID uint64
	Price       float64
	Quantity    int64
}

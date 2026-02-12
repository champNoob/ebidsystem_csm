package matching

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplayDeterministic(t *testing.T) {

	engine := NewEngine()
	engine.Start()

	orders := []*Order{
		{ID: 1, Symbol: "AAPL", Side: OrderSideBuy, Price: 100, Quantity: 10},
		{ID: 2, Symbol: "AAPL", Side: OrderSideSell, Price: 99, Quantity: 5},
		{ID: 3, Symbol: "AAPL", Side: OrderSideSell, Price: 100, Quantity: 5},
	}

	for _, o := range orders {
		require.NoError(t, engine.Submit(o))
	}

	events := make([]MatchEvent, 0)
	for i := 0; i < 2; i++ {
		ev := <-engine.Events()
		events = append(events, ev)
	}

	require.Equal(t, uint64(1), events[0].BuyOrderID)
	require.Equal(t, uint64(2), events[0].SellOrderID)
	require.Equal(t, int64(5), events[0].Quantity)

	require.Equal(t, uint64(1), events[1].BuyOrderID)
	require.Equal(t, uint64(3), events[1].SellOrderID)
	require.Equal(t, int64(5), events[1].Quantity)
}

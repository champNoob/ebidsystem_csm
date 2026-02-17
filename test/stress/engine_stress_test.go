package stress

import (
	"sync"
	"testing"
	"time"

	"ebidsystem_csm/internal/matching"
)

// 卖单并发提交压力测试：
func TestEngine_ConcurrentSubmit(t *testing.T) {
	engine := matching.NewEngine()
	engine.Start()
	defer engine.Stop()

	totalOrders := 100000
	concurrency := 100

	wg := sync.WaitGroup{}
	wg.Add(concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(worker int) {
			defer wg.Done()

			for j := 0; j < totalOrders/concurrency; j++ {
				engine.Submit(&matching.Order{
					ID:        uint64(worker*1000000 + j),
					Symbol:    "AAPL",
					Price:     10,
					Remaining: 1,
					Side:      matching.OrderSideBuy,
				})
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(start)

	t.Logf("Submitted %d orders in %v", totalOrders, elapsed)
}


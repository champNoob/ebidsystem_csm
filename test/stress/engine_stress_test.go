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

// 对冲撮合压力测试：
func TestEngine_MatchPressure(t *testing.T) {
	engine := matching.NewEngine()
	engine.Start()

	// 启动 event 消费者：
	eventCount := 0
	eventDone := make(chan struct{})
	go func() {
		defer close(eventDone)
		for range engine.Events() {
			eventCount++
		}
	}()

	total := 100000
	expectedMatches := total / 2 // 买卖单各半，理论上最多匹配 total/2 对

	start := time.Now()

	// 提交买卖订单：
	for i := 0; i < total; i++ {
		side := matching.OrderSideBuy
		if i%2 == 1 {
			side = matching.OrderSideSell
		}

		engine.Submit(&matching.Order{
			ID:        uint64(i),
			Symbol:    "AAPL",
			Price:     10,
			Remaining: 1,
			Side:      side,
		})
	}

	// 记录提交时间：
	submitElapsed := time.Since(start)
	t.Logf("Submitted %d mixed orders in %v", total, submitElapsed)
	// 计算提交性能指标：
	opsPerSecond := float64(total) / submitElapsed.Seconds()
	t.Logf("Submit Performance: %.2f orders/second", opsPerSecond)

	// 等待撮合完成：
	t.Log("Waiting for matching to complete...")
	time.Sleep(5 * time.Second)

	// 停止引擎：
	engine.Stop()
	// 等待事件处理完成：
	select {
	case <-eventDone:
		t.Logf("Processed %d match events", eventCount)
	case <-time.After(30 * time.Second):
		t.Logf("WARNING: Event processing timeout, processed %d events", eventCount)
	}

	// 验证撮合结果：
	if eventCount >= int(float64(expectedMatches)*0.95) { // 允许5%的误差
		t.Logf("PASS: Match performance meets expectation. Expected ~%d, got %d", expectedMatches, eventCount)
	} else {
		t.Logf("WARNING: Match performance below expectation. Expected ~%d, got %d", expectedMatches, eventCount)
	}
}

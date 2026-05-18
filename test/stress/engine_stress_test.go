package stress

import (
	"log"
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
				err := engine.Submit(&matching.Order{
					ID:       uint64(worker*10*totalOrders + j),
					Symbol:   "AAPL",
					Price:    10,
					Quantity: 1,
					Side:     matching.OrderSideSell,
				})
				if err != nil {
					log.Println(err)
				}
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(start)
	t.Logf("Submitted %d orders in %v", totalOrders, elapsed)
}

// 对冲撮合压力测试：
func TestEngine_MatchPressure(t *testing.T) {
	// 添加撮合引擎：
	engine := matching.NewEngine()
	engine.Start()

	var (
		// eventCount int
		events    []matching.MatchEvent
		mu        sync.Mutex
		eventDone = make(chan struct{})
	)

	// 启动 event 消费者：
	go func() {
		defer close(eventDone)
		for ev := range engine.Events() {
			mu.Lock()
			events = append(events, ev)
			// eventCount++
			mu.Unlock()
		}
	}()

	total := 100000
	expectedMatches := total / 2 //买卖单各半，理论上最多匹配 total/2 对

	start := time.Now()

	// 提交买卖订单：
	for i := 0; i < total; i++ {
		side := matching.OrderSideBuy
		if i%2 == 1 {
			side = matching.OrderSideSell
		}

		if err := engine.Submit(&matching.Order{
			ID:       uint64(i),
			Symbol:   "AAPL",
			Price:    10,
			Quantity: 1,
			Side:     side,
		}); err != nil {
			t.Fatalf("submit order failed: %v", err)
		}
	}

	// 记录提交时间：
	submitElapsed := time.Since(start)
	t.Logf("Submitted %d mixed orders in %v", total, submitElapsed)
	// 计算提交性能指标：
	opsPerSecond := float64(total) / submitElapsed.Seconds()
	t.Logf("Submit Performance: %.2f orders/second", opsPerSecond)

	// 等待撮合完成：
	t.Log("Waiting for matching to complete...")
	waitForEventCount(t, &mu, &events, expectedMatches, 30*time.Second)

	// 停止引擎：
	engine.Stop()
	// 等待事件处理完成：
	select {
	case <-eventDone:
		// t.Logf("Processed %d match events", eventCount)
	case <-time.After(30 * time.Second):
		// t.Logf("Warning: Event processing timeout, processed %d events", eventCount)
		t.Fatalf("Event processing timeout")
	}

	/* --- 验证撮合结果 --- */

	// 1. 数量验证：
	/*
		if eventCount >= int(float64(expectedMatches)*0.95) { // 允许5%的误差
			t.Logf("PASS: Match performance meets expectation. Expected ~%d, got %d", expectedMatches, eventCount)
		} else {
			t.Logf("WARNING: Match performance below expectation. Expected ~%d, got %d", expectedMatches, eventCount)
		}
	*/
	gotEvents := snapshotEvents(&mu, events)
	if len(gotEvents) != expectedMatches {
		t.Fatalf("match count incorrect: expected %d, got %d", expectedMatches, len(gotEvents))
	}
	// 2. Quantity 校验：
	for _, ev := range gotEvents {
		if ev.Quantity != 1 {
			t.Fatalf("invalid match quantity: %d", ev.Quantity)
		}
	}
	// 3. ID 唯一性校验：
	seenBuy := make(map[uint64]bool)
	seenSell := make(map[uint64]bool)
	for _, ev := range events {
		if seenBuy[ev.BuyOrderID] {
			t.Fatalf("Duplicate buy order matched: %d", ev.BuyOrderID)
		}
		if seenSell[ev.SellOrderID] {
			t.Fatalf("Duplicate sell order matched: %d", ev.SellOrderID)
		}
		seenBuy[ev.BuyOrderID] = true
		seenSell[ev.SellOrderID] = true
	}
	// 4. 配对合理性校验：买单 ID 必须是偶数、卖单 ID 必须是奇数
	for _, ev := range events {
		if ev.BuyOrderID%2 != 0 {
			t.Fatalf("Invalid buy order ID: %d", ev.BuyOrderID)
		}
		if ev.SellOrderID%2 != 1 {
			t.Fatalf("Invalid sell order ID: %d", ev.SellOrderID)
		}
	}

	t.Log("All correctness checks passed")
}

// 多代码撮合压力测试：
func TestEngine_MultiSymbolPressure(t *testing.T) {
	engine := matching.NewEngine()
	engine.Start()

	symbols := []string{"AAPL", "GOOG", "TSLA", "AMZN"}
	symbolCount := len(symbols)

	totalPerSymbol := 50000
	totalOrders := totalPerSymbol * symbolCount
	expectedMatchesPerSymbol := totalPerSymbol / 2
	expectedTotalMatches := expectedMatchesPerSymbol * symbolCount

	var (
		events    []matching.MatchEvent
		eventDone = make(chan struct{})
		mu        sync.Mutex
	)

	// 事件消费者：
	go func() {
		defer close(eventDone)
		for ev := range engine.Events() {
			mu.Lock()
			events = append(events, ev)
			mu.Unlock()
		}
	}()

	start := time.Now()

	// 并发提交每个 symbol：
	var submitWg sync.WaitGroup
	for _, sym := range symbols {
		symbol := sym
		submitWg.Add(1)

		go func() {
			defer submitWg.Done()

			for i := 0; i < totalPerSymbol; i++ {
				side := matching.OrderSideBuy
				idBase := hashSymbol(symbol) //保证 ID 不冲突

				if i%2 == 1 {
					side = matching.OrderSideSell
				}

				if err := engine.Submit(&matching.Order{
					ID:       uint64(idBase + i),
					Symbol:   symbol,
					Price:    10,
					Quantity: 1,
					Side:     side,
				}); err != nil {
					t.Errorf("submit order failed: %v", err)
				}
			}
		}()
	}

	submitWg.Wait()

	submitElapsed := time.Since(start)
	t.Logf("Submitted %d orders across %d symbols in %v",
		totalOrders, symbolCount, submitElapsed)

	t.Log("Waiting for matching...")
	waitForEventCount(t, &mu, &events, expectedTotalMatches, 30*time.Second)

	engine.Stop()

	select {
	case <-eventDone:
	case <-time.After(30 * time.Second):
		t.Fatal("Event consumer timeout")
	}

	gotEvents := snapshotEvents(&mu, events)
	totalMatches := len(events)
	t.Logf("Total matches: %d", totalMatches)

	/* --- 验证撮合结果 --- */

	// 1. 总量验证：
	if totalMatches != expectedTotalMatches {
		t.Fatalf("Match count mismatch: expected %d, got %d",
			expectedTotalMatches, totalMatches)
	}
	// 2. 每个 symbol 匹配数验证：
	perSymbolCount := make(map[string]int)
	for _, ev := range gotEvents {
		perSymbolCount[ev.Symbol]++
	}
	for _, sym := range symbols {
		if perSymbolCount[sym] != expectedMatchesPerSymbol {
			t.Fatalf("Symbol %s match count mismatch: expected %d, got %d",
				sym, expectedMatchesPerSymbol, perSymbolCount[sym])
		}
	}
	// 3. 唯一性校验：
	seenBuy := make(map[uint64]bool)
	seenSell := make(map[uint64]bool)
	for _, ev := range gotEvents {
		if ev.Quantity != 1 {
			t.Fatalf("Invalid quantity: %d", ev.Quantity)
		}
		if seenBuy[ev.BuyOrderID] {
			t.Fatalf("Duplicate buy ID: %d", ev.BuyOrderID)
		}
		if seenSell[ev.SellOrderID] {
			t.Fatalf("Duplicate sell ID: %d", ev.SellOrderID)
		}
		seenBuy[ev.BuyOrderID] = true
		seenSell[ev.SellOrderID] = true
	}
	t.Log("Multi-symbol correctness checks passed")
	// 4. 吞吐评估：
	matchesPerSecond := float64(totalMatches) / submitElapsed.Seconds()
	t.Logf("Match throughput: %.2f matches/sec", matchesPerSecond)
}

func hashSymbol(sym string) int {
	var h int
	for i := 0; i < len(sym); i++ {
		h += int(sym[i])
	}
	return h * 1000000
}

func waitForEventCount(
	t *testing.T,
	mu *sync.Mutex,
	events *[]matching.MatchEvent,
	expected int,
	timeout time.Duration,
) {
	t.Helper()

	deadline := time.After(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		mu.Lock()
		count := len(*events)
		mu.Unlock()

		if count >= expected {
			return
		}

		select {
		case <-ticker.C:
		case <-deadline:
			t.Fatalf("timeout waiting for events: expected %d, got %d", expected, count)
		}
	}
}

func snapshotEvents(mu *sync.Mutex, events []matching.MatchEvent) []matching.MatchEvent {
	mu.Lock()
	defer mu.Unlock()

	copied := make([]matching.MatchEvent, len(events))
	copy(copied, events)
	return copied
}

# 新增撮合引擎

```txt
internal/
├── matching/              # 撮合引擎（核心）
│   ├── engine.go          # 撮合引擎入口
│   ├── order_book.go      # 订单簿
│   ├── matcher.go         # 撮合逻辑
│   ├── event.go           # 撮合结果事件
│   └── types.go           # 核心类型
```

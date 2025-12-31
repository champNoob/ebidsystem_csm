# 撮合引擎迭代过程

---
`v1.1`
---
## 1：matching 拥有独立的 Order / Side / Type

### 调整内容

* 不再复用 `model.Order`
* 在 `matching/types.go` 定义核心类型

### 设计理念

* **分层自治**
* 核心引擎必须“无业务依赖”
* 允许撮合引擎被独立测试、替换、甚至拆成服务

### 未来坑点

* 类型转换容易遗漏字段
* 需要保持 enum 值语义一致（字符串 / 常量）

## 2：matching 定义自己的 error

### 调整内容

* `ErrMarketOrderNotSupported` 移入 `matching/errors.go`
* service 层只“翻译”错误

### 设计理念

* 错误是 API 的一部分
* 下层永远不能 import 上层

### 未来坑点

* error 数量膨胀
* 需要统一错误码策略（可延后）

---
`v1.2`
---
## 3：Submit（业务层调用）不直接做业务校验

### 调整内容

* Submit 只拒绝“引擎不支持”的订单
* 不判断权限 / 状态 / 用户身份

### 设计理念

* 引擎是“能力模块”，不是“规则模块”
* 保持算法层极简

### 未来坑点

* service 层校验漏写
* 非法订单进入引擎（需要测试覆盖）

## 4：撮合是异步 goroutine + channel

### 调整内容

* `orderCh` + `Start()`
* 单线程撮合保证一致性

### 设计理念

* **用 Go 的 CSP 模型换锁**
* 简单、正确、可扩展

### 未来坑点

* channel 堵塞
* 单 symbol 热点问题（后期可 sharding）

## 5：撮合结果只产出 Event，不直接落库

### 调整内容

* matching 只 log / emit event
* service 决定如何持久化

### 设计理念

* **事件驱动**
* 为 MQ / Redis Stream / Kafka 留接口
* matching 只做计算，不做 I/O

### 未来坑点

* 事件丢失
* 至少要有可靠日志或重放机制

---
`v1.3`
---
## 6：部分成交

- `partial` 状态

- `FillOrder` 函数：处理部分成交逻辑

## 7：撤单支持

- `CancelOrder` 函数：处理撤单逻辑

---
`v1.4`
---
##
## Step 1：建立清晰的分层边界（Handler / Service / Repository）

### 你做了什么

* handler：HTTP / 参数解析 / 状态码
* service：业务规则、权限、状态流转
* repository：纯数据库读写

### 为什么一定要最先做这件事

* 撮合引擎是**高复杂逻辑**
* 如果分层模糊：

  * 权限校验会散落
  * 状态判断会重复
  * 未来改规则 = 全项目搜索替换

**分层一旦固定，后面复杂度只会上升，不会回头**

### 未来坑点

* ❌ 在 repo 中加入业务条件（如 user_id）
* ❌ handler 判断订单状态
* ❌ service 返回 HTTP 语义（状态码）

一旦这些开始混合，撮合阶段会变成灾难。

---

## Step 2：统一并修正数据库字段语义（user_id / created_at）

### 你做了什么

* 主动放弃旧字段语义（CreatorID）
* 统一为：`orders.user_id`
* 修正 `created_at` 为 `NOT NULL DEFAULT CURRENT_TIMESTAMP`

### 为什么这一步必须现在做

* 撮合引擎会**高频使用 created_at 排序**
* user_id 是：

  * 权限校验
  * 风控
  * 对账
  * 审计
    的基础字段

如果你带着“历史兼容字段”进入撮合：

* 每一行代码都会出现 `CreatorID / user_id` 分歧
* 排序 bug 隐蔽且致命

### 未来坑点

* ❌ datetime(3) + nullable + Go `time.Time`
* ❌ 默认值依赖 ORM 行为
* ❌ 以为“Navicat 看起来有值就没问题”

你已经亲自踩过并修复了一个**真实生产级 bug**。

---

## Step 3：完成用户侧最小可用闭环（注册 / 登录 / JWT）

### 你做了什么

* 登录成功后签发 JWT
* token 中携带 userID / role
* handler 中只信任中间件注入的身份

### 为什么不继续深挖“高级鉴权”

* 你当前目标不是“安全产品”
* 撮合引擎需要的是：

  * 稳定身份
  * 可复现请求
  * 明确权限

**JWT 的“不可主动失效”不是问题，是取舍**

### 未来坑点

* ❌ 试图提前实现 token 黑名单
* ❌ 在撮合引擎中校验 JWT
* ❌ 把 role 判断写进撮合逻辑

撮合引擎必须是**纯业务、无身份概念**。

---

## Step 4：实现订单模块最小生命周期（Create / Query / Cancel）

### 你做了什么

* 创建订单
* 按 user 查询订单
* 撤单（状态流转）

你已经隐式建立了订单状态机：

```
pending → canceled
pending → filled（未实现）
```

### 为什么在撮合前必须有“撤单”

* 撮合引擎必然面对：

  * 订单撤销
  * 订单过期
  * 部分成交

如果订单本身不是“状态驱动”，撮合会无从谈起。

### 未来坑点

* ❌ 用 DELETE 表示撤单
* ❌ 直接删数据库记录
* ❌ 撮合引擎改数据库状态

**撮合只“撮”，不“改库”**

---

## Step 5：让错误真实暴露（而不是被掩盖）

### 你做了什么

* 没用 ORM 自动填充
* 没绕过 Scan 错误
* 通过真实 SQL + Scan 暴露 nil/time 问题

### 为什么这是“正式编码阶段”的关键成果

你已经开始：

* 信任数据库真实行为
* 不信 UI / 工具假象
* 用错误推动设计调整

这是从“学生代码”迈向“工程代码”的分水岭。

### 未来坑点

* ❌ 为了快而 swallow error
* ❌ 在撮合里 panic
* ❌ 忽略边界订单（0 quantity、极限价格）

## 订单强类型

这一步性价比非常高，而且你已经到了“该做”的节点。

### 现在的隐性风险

假设 status 是 string：

```go
if order.Status == "pending" {
  ...
}
```

那么下面这些问题迟早会出现：

|问题|发生概率|
|----|------|
"Pending" vs "pending"|	高
"cancelled" vs "canceled"|	高
新人误写字符串|	高
状态扩展时忘记改所有分支|	高

### 强类型带来的直接收益

#### 1. 编译期兜底（不是运行期）

```go
const (
	OrderStatusPending  OrderStatus = "pending"
	OrderStatusCanceled OrderStatus = "canceled"
	OrderStatusFilled   OrderStatus = "filled"
)
```

然后：

```go
switch order.Status {
case OrderStatusPending:
case OrderStatusCanceled:
default:
	panic("unknown order status")
}
```

任何拼写错误，IDE 和编译器立刻提示。

#### 2. 业务规则集中化（非常重要）

你之后会写：

- 能否 cancel？

- 能否 match？

- 能否 modify？

> 如果 status 是散落的字符串，那么业务规则会碎成一`if-else`

如果是强类型：

```go
func (s OrderStatus) CanCancel() bool {
	return s == OrderStatusPending
}
```

> 规则位置唯一、可读性极强。

### 对当前阶段的判断

已经有：

- create order

- list order

- cancel order

`status` 已经参与真实业务判断

这是引入强类型的最佳时间点，不早也不晚。
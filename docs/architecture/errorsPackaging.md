# 错误封装

## 一、错误集中定义：消除“散弹式 `errors.New`”

### 做法

- 项目中所有 `errors.New(...)` **只存在于** `internal/service/errors.go`

- 服务层业务代码**只引用已定义的** `error` 变量

### 优势

这样的错误语义**单一来源（Single Source of Truth）** 避免了：

- 同一业务错误，不同模块文案不一致
- 搜索、重构、国际化困难

为后续：

- 错误码（error code）
- 多语言
- 统一错误映射

提供了天然基础er

## 二、区分“业务错误”与“技术错误”

### 做法

- `errors.go` 中的错误全部是**业务语义错误**，例如：
  - `ErrUserAlreadyExists`
  - `ErrPermissionDenied`
  - `ErrOrderNotCancellable`
- 技术性错误（SQL、Scan、ctx 等）：
  - 不直接暴露给前端
  - 只用于日志或被包装为业务错误

### 优势

- 前端 / 用户 **不应该知道数据库、字段名、校验标签**
- 用户只关心能不能做、为什么不能做（业务原因）

从而避免：

```
Error 1062 (23000): Duplicate entry ...
Key: 'CreateOrderRequest.OrderSide' Error:Field validation ...
```

这种“后端裸奔式错误”

## 三、错误分层：校验错误 ≠ 权限错误 ≠ 系统错误

### 当前已形成的隐式分层

| 层级        | 示例                     | 是否对前端可见 |
| --------- | ---------------------- | ------- |
| 参数 / 输入错误 | ErrInvalidInput        | ✅       |
| 权限 / 角色错误 | ErrPermissionDenied    | ✅       |
| 状态机错误     | ErrOrderNotCancellable | ✅       |
| 系统内部错误    | ErrInternal            | ✅（兜底）   |
| 技术细节错误    | SQL / Scan error       | ❌       |

### 关键点

- 业务规则失败 ≠ 程序出错
- 程序出错 ≠ 用户输入错

你已经在 **Service 层承担了“错误翻译器”的角色**，这是正确的。

## 四、校验函数返回“内部错误语义”，Service 层统一映射

### 典型例子：`validateRoleSide`

- 校验函数内部：

  - 返回具体、可诊断的错误（或枚举）
- Service 层：

  - **统一映射为稳定业务错误**
  - 同时记录日志保留细节

```go
log.Printf("[ORDER_VALIDATE] role=%s side=%s err=%v", ...)
return ErrPermissionDenied
```

### 设计思想

- 内部：详细、可调试
- 对外：稳定、可控、不泄漏规则

这是**企业级服务端的标准模式**


## 五、Handler 层不再“创造错误”，只做错误映射

### 做法

Handler 层的职责被严格控制为：

1. 参数解析
2. 调用 Service
3. 根据返回 error 决定 HTTP 行为

Handler 不再：

* `errors.New(...)`
* 拼装错误文案
* 判断复杂业务规则

> 创建 `Helper` 函数：`respondError`，统一调用服务层的自定义错误

### 结果

* Handler 极度“薄”
* Service 成为唯一业务真相
* API 行为一致、可预测

## 六、错误裁剪三原则

### 原则 1：错误是否“面向用户”？

- 是 → 保留独立错误码
- 否 → 可内部合并

### 原则 2：是否跨层传播？

- 跨 handler/service → 粗
- 仅 service 内 → 细

### 原则 3：是否有独立处理价值？

- 会单独统计 / 监控 / 告警 → 不合并
- 永远同样处理 → 可合并

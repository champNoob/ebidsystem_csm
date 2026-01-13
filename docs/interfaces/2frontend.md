# 一、前端接口定义总览

| 模块 | 接口     | 方法                          | 请求体                               | 响应体                                                                 | 权限                              | 备注                      |
| -- | ------ | --------------------------- | --------------------------------- | ------------------------------------------------------------------- | ------------------------------- | ----------------------- |
| 用户 | 注册     | POST /users                 | { username, password, role }      | { message }                                                         | admin / 注册用户                    | admin 才能创建新用户           |
| 用户 | 登录     | POST /login                 | { username, password }            | { token }                                                           | 所有                              | 返回 JWT，前端存 localStorage |
| 用户 | 当前用户信息 | GET /api/me                 | -                                 | { id, username, role }                                              | JWT                             | 用于显示用户名、权限              |
| 订单 | 下单     | POST /api/orders            | { symbol, side, price, quantity } | { message }                                                         | JWT: client/seller/trader       | 前端禁用非角色操作               |
| 订单 | 查询订单列表 | GET /api/orders             | query: ?symbol=&side=&status=     | [ { id, symbol, side, price, quantity, filled_quantity, status } ]  | JWT: client/seller/trader/admin | admin 可看全部，普通用户仅看自己     |
| 订单 | 撤单     | POST /api/orders/:id/cancel | -                                 | { message } 或 { error }                                             | JWT: client/seller/trader/admin | 根据状态机决定是否可撤单            |
| 管理 | 用户查询   | GET /api/admin/users/:id    | -                                 | { id, username, role, status }                                      | JWT admin                       | 后端权限校验                  |
| 管理 | 用户创建   | POST /api/admin/users       | { username, password, role }      | { message }                                                         | JWT admin                       | 可重复使用 /users 创建用户接口     |
| 成交 | 查询成交记录 | GET /api/trades             | query: ?symbol=&orderID=          | [ { id, buy_order_id, sell_order_id, price, quantity, timestamp } ] | JWT: 根据角色                       | 可选（MVP阶段可先不实现）          |

---

# 二、接口说明（前端重点字段）

## 1️⃣ 用户模块

### 1.1 登录

* **URL**: `/login`
* **方法**: POST
* **请求体**:

```json
{
  "username": "alice",
  "password": "12345678"
}
```

* **返回**:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

* **前端关注**:

  * 保存 token（localStorage）
  * 后续 Axios 默认加 `Authorization: Bearer <token>`

---

### 1.2 当前用户信息

* **URL**: `/api/me`
* **方法**: GET
* **请求头**: `Authorization: Bearer <token>`
* **返回**:

```json
{
  "id": 12,
  "username": "alice",
  "role": "client"
}
```

* **前端用途**:

  * 显示用户名
  * 控制页面操作权限（下单 / 撤单按钮禁用）

---

## 2️⃣ 订单模块

### 2.1 下单

* **URL**: `/api/orders`
* **方法**: POST
* **请求体**:

```json
{
  "symbol": "AAPL",
  "side": "buy",
  "price": 150.5,  // limit 单必填
  "quantity": 10
}
```

* **返回**:

```json
{
  "message": "order created"
}
```

* **前端关注**:

  * 表单验证
  * price 对 limit 单必须有值
  * side = buy/sell
  * quantity > 0

---

### 2.2 查询订单列表

* **URL**: `/api/orders`
* **方法**: GET
* **请求头**: `Authorization: Bearer <token>`
* **查询参数**（可选）: `symbol`, `side`, `status`
* **返回**:

```json
[
  {
    "id": 101,
    "symbol": "AAPL",
    "side": "buy",
    "price": 150.5,
    "quantity": 10,
    "filled_quantity": 3,
    "status": "partial"
  },
  {
    "id": 102,
    "symbol": "AAPL",
    "side": "sell",
    "price": 151.0,
    "quantity": 5,
    "filled_quantity": 0,
    "status": "pending"
  }
]
```

* **前端关注**:

  * 根据 `status` 控制撤单按钮可用性
  * 可以轮询更新订单状态

---

### 2.3 撤单

* **URL**: `/api/orders/:id/cancel`
* **方法**: POST
* **请求头**: `Authorization: Bearer <token>`
* **返回**:

**成功**:

```json
{
  "message": "order cancelled"
}
```

**失败**:

```json
{
  "error": "order cannot be cancelled"
}
```

* **前端注意**:

  * 根据 `status` 决定按钮禁用
  * 成功后刷新订单列表
  * 错误提示给用户

---

## 3️⃣ 管理模块

* `/api/admin/users/:id` GET → 查询用户
* `/api/admin/users` POST → 创建用户

前端只需在 admin 角色下渲染对应页面即可，其他普通用户隐藏。

---

## 4️⃣ 成交记录（可选）

* `/api/trades` GET
* 返回每条交易：`buy_order_id`, `sell_order_id`, `price`, `quantity`, `timestamp`
* 前端用于交易历史表格

---

# 三、前端使用 Axios 封装

```ts
// src/api/axios.ts
import axios from 'axios'

const api = axios.create({
  baseURL: 'http://localhost:8080',
  timeout: 5000
})

api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

export default api
```
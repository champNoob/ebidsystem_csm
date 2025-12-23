# 鉴权设计说明（Authentication & Authorization）

## 1. 设计目标

本系统采用基于 JWT 的无状态鉴权机制，目标是：

- 支持多角色用户（admin / trader / seller / client 等）
- 保证接口访问安全
- 结构清晰，便于后期扩展
- 控制系统复杂度，避免过度设计

当前系统定位为：
- 教学 / 课程 / 个人项目级证券交易系统
- 用户规模有限
- 不追求企业级 SSO / 单点登出复杂方案

## 2. 鉴权模型概览

### 2.1 用户模型

所有用户（包括 admin）统一存储于 `users` 表：

| 字段 | 说明 |
|-|-|
| id | 用户唯一标识 |
| username | 登录名 |
| password_hash | 密码哈希 |
| role | 用户角色 |
| is_deleted | 逻辑删除 |

**admin 并非特殊账号，而是 role = admin 的普通用户。**

### 2.2 登录流程

1. 用户调用 `POST /login`
2. 服务端校验用户名与密码
3. 校验成功后生成 JWT
4. JWT 返回给客户端
5. 客户端在后续请求中通过 Header 携带：

```http
Authorization: Bearer <token>
````

## 3. JWT 行为定义（非常重要）

### 3.1 JWT 的内容（Claims）

JWT 中包含以下核心信息：

* user_id：用户唯一标识
* role：用户角色
* exp：过期时间
* iat：签发时间

JWT 是 **自包含的（self-contained）**，服务端不会为每次请求访问数据库。

### 3.2 JWT 的校验方式

服务端对每个受保护接口：

1. 校验 token 是否存在
2. 校验签名是否合法
3. 校验 token 是否过期
4. 从 token 中解析 user_id、role
5. 将信息写入 Gin Context，供后续使用

### 3.3 JWT 行为约定（重要设计决策）

* JWT 在 **未过期前始终有效**
* 同一用户可同时持有多个有效 JWT
* 系统 **不支持 token 主动失效 / 登出**

这是一个**有意的设计选择，而非缺陷**。

## 4. 为什么不支持 Token 主动失效？

### 4.1 采用该策略的原因

* JWT 是无状态的，天然不支持服务端主动失效
* 若支持主动失效，必须引入：

  * Redis 黑名单
  * token version
  * refresh token 体系
* 会显著提高系统复杂度

### 4.2 本系统的取舍

| 方案              | 是否采用 |
|  | - |
| JWT + exp       | ✅    |
| Redis token 黑名单 | ❌    |
| Refresh Token   | ❌    |
| 单点登录            | ❌    |

本系统选择：

> **短有效期 JWT + 到期自动失效**

这是在当前项目阶段 **性价比最高、可维护性最好的方案**。

## 5. 授权（Authorization）模型

### 5.1 基于角色的访问控制（RBAC）

接口按角色划分访问权限，例如：

| 接口                    | 角色              |
|  |  |
| GET /api/me           | 所有登录用户          |
| POST /users           | 匿名 / admin      |
| POST /api/admin/users | admin           |
| GET /api/orders       | trader / seller |

### 5.2 实现方式

* JWT 中携带 role
* 中间件 `RequireRole(...)` 校验当前用户角色
* 未通过校验直接返回 403



## 6. admin 用户设计说明

* admin 不是硬编码账号
* admin 与普通用户共用同一套登录、鉴权机制
* 区别仅在 role 字段

### admin 的创建方式：

* 初始 admin 通过数据库手动插入
* 后续可通过后台接口调整用户角色（预留）



## 7. 未实现但已规划的功能（TODO）

* PUT /users/:id/role（admin）
* 用户禁用 / 启用
* 操作审计日志
* Token 刷新机制（如未来需要）

## 8. 设计总结

当前鉴权方案遵循以下原则：

* 简单优先
* 明确边界
* 可解释、可扩展
* 避免为当前阶段引入不必要复杂度
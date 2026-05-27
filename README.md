# 在线证券交易系统（接口层-服务层-仓储层架构）

## 项目功能

## 项目结构

### 文件结构

```txt
ebidsystem_csm/
├── bin/                        # 编译后的二进制文件（.gitignore，暂未实现）
│
├── cmd/                        # 应用程序入口（支持多程序入口）
│   ├── cli/                      # 命令行工具（暂未实现）
│   │   └── main.go
│   └── server/                   # HTTP服务器
│       ├── main.go                 # 主入口文件
│       ├── wire.go                 # Wire依赖定义（暂未实现）
│       └── wire_gen.go             # Wire生成的代码（.gitignore，暂未实现）
│
├── deployments/                # 部署配置（暂未实现）
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── k8s/
│   └── nginx/
│
├── docs/                       # 文档
│   ├── architecture/             # 架构设计文档
│   ├── lessons/                  # 学习笔记
│   ├── notes_of_every_stage/     # 各阶段开发笔记
│   ├── questions/                # 问题与反思记录
│   └── versions/                 # 版本记录
│
├── internal/                   # 私有应用程序代码（外部不可导入）
│   ├── api/                      # HTTP接口层
│   │   ├── dto/                    # 数据传输对象
│   │   │   ├── request/              # 请求对象
│   │   │   │   ├── create_order.go     # 创建订单请求
│   │   │   │   ├── create_user.go      # 创建用户请求
│   │   │   │   ├── list_orders.go      # 列出订单请求
│   │   │   │   └── login.go            # 登录请求
│   │   │   └── response/             # 响应对象
│   │   │       ├── errors.go           # 错误响应
│   │   │       └── user.go             # 用户响应
│   │   ├── handler/                # 请求控制器（Controller）
│   │   │   ├── admin_handler.go      # 管理员控制器
│   │   │   ├── auth_errors.go        # 控制层内部错误
│   │   │   ├── error_handler.go      # 错误处理
│   │   │   ├── order_handler.go      # 订单控制器
│   │   │   └── user_handler.go       # 用户控制器
│   │   └── route/                  # 路由定义
│   │       ├── router.go             # 路由配置
│   │       └── middleware.go         # 路由级中间件配置（暂未实现）
│   │
│   ├── apperror/                   # 全局错误定义
│   │   ├── business_error.go         # 业务错误类型定义（服务层）
|   |   ├── catalog.go                # 错误目录
|   |   └── http_status.go            # HTTP状态码映射
│   │
│   ├── cache/                    # 缓存实现（需复杂缓存逻辑，暂未实现）
│   │
│   ├── config/                   # 配置管理
│   │   ├── config.go               # 配置结构体定义
│   │   ├── config.yaml             # 配置文件示例
│   │   ├── env.go                  # 环境变量处理（暂未实现）
│   │   └── loader.go               # 配置加载器
│   │
│   ├── constant/                 # 全局常量定义（暂未实现）
│   │
│   ├── matching/                 # 撮合引擎
│   │   ├── engine_test.go          # 撮合引擎测试
│   │   ├── engine.go               # 撮合引擎主入口
│   │   ├── errors.go               # 错误类型定义
│   │   ├── event.go                # 撮合事件定义
│   │   ├── matcher_test.go         # 核心撮合逻辑测试（暂未实现）
│   │   ├── matcher.go              # 核心撮合逻辑
│   │   ├── order_book_test.go      # 订单簿测试
│   │   ├── order_book.go           # 订单簿实现
│   │   ├── replay_test.go          # 幂等性检查
│   │   ├── symbol_matcher_test.go  # 代码级撮合器测试
│   │   ├── symbol_matcher.go       # 代码级撮合器实现
│   │   └── types.go                # 撮合引擎相关类型定义
│   │
│   ├── middleware/               # HTTP中间件
│   │   ├── auth/                   # 认证中间件
│   │   │   ├── casbin.go             # RBAC权限控制（暂未实现）
│   │   │   ├── jwt.go                # JWT认证
│   │   │   └── role.go               # 鉴权中间件
│   │   ├── logging/                # 日志中间件（暂未实现）
│   │   │   ├── access_log.go         # 访问日志
│   │   │   └── request_log.go        # 请求日志
│   │   ├── recovery/               # 恢复中间件（暂未实现）
│   │   │   └── panic_recover.go      # 恐慌恢复
│   │   ├── tracing/                # 跟踪中间件（暂未实现）
│   │   │   └── request_id.go         # 请求ID追踪
│   │   ├── rate_limit/             # 限流中间件（暂未实现）
│   │   └── cors/                   # 跨域中间件（暂未实现）
│   │
│   ├── model/                    # 数据模型
│   │   ├── model_side.go           # 买卖方向类型
│   │   ├── model_status.go         # 订单状态类型
│   │   ├── order_type.go           # 订单类型
│   │   ├── order.go                # 订单模型
│   │   ├── trade.go                # 交易模型
│   │   ├── user_role.go            # 用户角色模型
│   │   └── user.go                 # 用户模型
│   │
│   ├── pkg/                      # 内部共享包（可被internal内其他包引用）
│   │   ├── cache/                  # 缓存封装（暂未实现）
│   │   │   └── redis_cache.go
│   │   ├── database/               # 数据库连接池
│   │   │   ├── mysql.go              # MySQL连接池
│   │   │   ├── redis.go              # Redis连接池
│   │   │   └── migrate.go            # 数据库迁移（暂未实现）
│   │   ├── logger/                 # 日志系统
│   │   │   ├── logger_test.go        # 日志测试
│   │   │   └── logger.go             # 日志记录器实现
│   │   ├── queue/                  # 消息队列（暂未实现）
│   │   │   ├── producer.go
│   │   │   └── consumer.go
│   │   ├── security/               # 安全相关（暂未实现）
│   │   |   ├── password.go
│   │   |   └── encryption.go
│   │   └── validator/              # 参数验证器（暂未实现）
│   │       └── custom_validator.go
│   │
│   ├── repository/               # 仓储层（为演进到DDD做准备）
│   │   ├── dto/                    # 数据传输对象（仓储层内部）
|   |   |   └── admin_dto.go          # 管理员对象
│   │   ├── mysql/                  # MySQL仓储
│   │   │   ├── admin_repo.go         # 管理员接口
│   │   │   ├── errors.go             # 仓储层错误管理
│   │   │   ├── event_repo.go         # 事件管理
│   │   │   ├── match_event_repo.go   # 撮合事件接口
│   │   │   ├── order_repo_tx_test.go # 订单接口事务测试
│   │   │   ├── order_repo.go         # 订单接口
│   │   │   ├── tx.go                 # 事务管理
│   │   │   └── user_repo.go          # 用户接口
│   │   ├── redis/                  # Redis仓储（暂未实现）
│   │   │   └── cache_repo.go
│   │   └── interface.go            # 仓储层接口
│   │
│   ├── service/                  # 业务层
│   │   ├── admin_service.go        # 订单服务
│   │   ├── order_service.go        # 订单服务
│   │   ├── order_validation.go     # 订单校验
│   │   └── user_service.go         # 用户服务
│   │
│   └── utils/                    # 工具函数（静态方法，暂未实现）
│       ├── time_util.go
│       ├── string_util.go
│       ├── json_util.go
│       └── conv_util.go            # 类型转换
│    
├── log/                        # 日志文件目录（.gitignore）
│   ├── engine/                   # 撮合引擎日志
│   ├── logger/                   # 日志系统日志
│   └── order/                    # 订单相关日志
│
├── pkg/                        # 公共库包（可被外部项目引用，暂未实现）
│   ├── errors/                   # 自定义错误类型
│   ├── pagination/               # 分页组件
│   └── idgenerator/              # ID生成器
│
├── scripts/                    # 脚本目录（暂未实现）
│   ├── deploy/
│   ├── migration/
│   └── build.sh
│
├── storage/                    # 存储目录（上传文件等，.gitignore，暂未实现）
│   ├── uploads/
│   └── cache/
│
├── temp/                       # 临时文件（.gitignore）
│
├── test/                       # 测试文件
│   ├── e2e/                      # 端到端测试（暂未实现）
│   ├── integration/              # 集成测试（暂未实现）
│   ├── performance/              # 性能测试（暂未实现）
│   └── stress/                   # 压力测试
|       └── engine_stress_test.go   # 引擎压力测试（含多代码撮合压力测试）
|
├── .env.example                # 环境变量示例（暂未实现）
├── .env.local                  # 本地环境配置（.gitignore，暂未实现）
├── .gitattributes              # Git属性配置
├── .gitignore                  # Git忽略文件
├── go.mod                      # Go模块文件
├── go.sum                      # Go模块校验和
├── Makefile                    # 构建命令（暂未实现）
└── README.md                   # 项目说明
```

### 分层架构

![分层架构图](https://www.plantuml.com/plantuml/png/bP9DJiCm48NtFiMecxQBbXkW_bI2I4L08rR0mdAcgLN7YSOEA0As780ZSX9AdBgkKHPidz_CfyoNcR6SR5qePgLL2BYG6QIDiLZ019Ob8QomGfsX5Wsi9C-95uoPlTGL9rv0nSMUvvZQHY4G2ijrhZ0ec1tFobUfSzXoPP2nRW86yxi4rhn1UjAZAwdXckjC8Pdn0DuO0C3ZWq7gqcUNt58MH_FQxdpo4UnFaLwaGCzOr4PgD0RMPIx5EQNhXXGVUOfFGie6gz98MrBZGOcsI5ikq7Y8F4RmIplDH8yjE7Zj0IL5fGwScoQzNyC5R32JuTKwWdiV1z_K5o-vsE78hJd_kaHlYHjBjR0reKSuychHBMadsy7XqI9yVtv1Tp1s9X8caylew5whByGE0ikk41bXmoEi1GUtyZ9Oo7Gx8XUUUMh4GWwSo4FB2o7HzM4wVUyXnRL_u48hVvdkG5vIXEjl1F-1gtYrCzHPNQWV
)

系统采用分层架构，Handler 仅依赖 Service，Service 仅依赖 Repository 接口，具体存储实现可替换。

## 撮合引擎

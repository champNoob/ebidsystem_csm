# 在线证券交易系统（接口层-服务层-仓储层架构）

## 项目功能

## 项目结构

```txt
ebidsystem_csm/
├── cmd/                        # 应用程序入口（支持多程序入口）
│   ├── cli/                      # 命令行工具（暂不实现）
│   │   └── main.go
│   └── server/                   # HTTP服务器
│       ├── main.go                 # 主入口文件
│       ├── wire.go                 # Wire依赖定义（暂不实现）
│       └── wire_gen.go             # Wire生成的代码（.gitignore）
│
├── internal/                   # 私有应用程序代码（外部不可导入）
│   ├── api/                      # HTTP接口层
│   │   ├── handler/                # 请求处理器（Controller）
│   │   │   ├── user_handler.go       # 用户处理器
│   │   │   ├── auth_handler.go       # 认证处理器
│   │   │   └── base_handler.go       # 基础处理器
│   │   ├── route/                  # 路由定义
│   │   │   ├── router.go             # 路由配置
│   │   │   └── middleware.go         # 路由级中间件配置
│   │   └── dto/                    # 数据传输对象
│   │       ├── request/              # 请求参数
│   │       └── response/             # 响应参数
│   │
│   ├── service/                  # 业务逻辑层
│   │   ├── user_service.go         # 用户服务
│   │   ├── auth_service.go         # 认证服务
│   │   ├── order_service.go        # 订单服务
│   │   └── interface.go            # 服务接口定义
│   │
│   ├── repository/               # 数据访问层（为演进到DDD做准备）
│   │   ├── mysql/                  # MySQL仓储
│   │   │   ├── user_repo.go
│   │   │   └── base_repo.go
│   │   ├── redis/                  # Redis仓储
│   │   │   └── cache_repo.go
│   │   └── interface.go            # 仓储接口
│   │
│   ├── model/                    # 数据模型
│   │   ├── user.go                 # 用户模型
│   │   ├── order.go                # 订单模型
│   │   └── base.go                 # 基础模型
│   │
│   ├── config/                   # 配置管理
│   │   ├── config.go               # 配置结构体定义
│   │   ├── loader.go               # 配置加载器
│   │   ├── env.go                  # 环境变量处理
│   │   └── config.yaml             # 配置文件示例
│   │
│   ├── middleware/               # HTTP中间件
│   │   ├── auth/                   # 认证中间件
│   │   │   ├── jwt.go                # JWT认证
│   │   │   └── casbin.go             # RBAC权限控制
│   │   ├── logging/                # 日志中间件
│   │   │   ├── access_log.go         # 访问日志
│   │   │   └── request_log.go        # 请求日志
│   │   ├── recovery/               # 恢复中间件
│   │   │   └── panic_recover.go      # 恐慌恢复
│   │   ├── tracing/                # 跟踪中间件
│   │   │   └── request_id.go         # 请求ID追踪
│   │   ├── rate_limit/             # 限流中间件
│   │   └── cors/                   # 跨域中间件
│   │
│   ├── pkg/                      # 内部共享包（可被internal内其他包引用）
│   │   ├── database/               # 数据库连接池
│   │   │   ├── mysql.go              # MySQL连接池
│   │   │   ├── redis.go              # Redis连接池
│   │   │   └── migrate.go            # 数据库迁移
│   │   ├── cache/                  # 缓存封装
│   │   │   └── redis_cache.go
│   │   ├── queue/                  # 消息队列
│   │   │   ├── producer.go
│   │   │   └── consumer.go
│   │   ├── validator/              # 参数验证器
│   │   │   └── custom_validator.go
│   │   └── security/               # 安全相关
│   │       ├── password.go
│   │       └── encryption.go
│   │
│   ├── utils/                    # 工具函数（静态方法）
│   │   ├── time_util.go
│   │   ├── string_util.go
│   │   ├── json_util.go
│   │   └── conv_util.go            # 类型转换
│   │
│   ├── constant/                 # 常量定义
│   │   ├── error_code.go
│   │   ├── cache_key.go
│   │   └── status_code.go
│   │
│   └── cache/                    # 缓存实现（如需复杂缓存逻辑）
│       ├── local_cache.go          # 本地缓存
│       └── distributed_lock.go     # 分布式锁
│
├── pkg/                        # 公共库包（可被外部项目引用）
│   ├── errors/                   # 自定义错误类型
│   ├── pagination/               # 分页组件
│   └── idgenerator/              # ID生成器
│
├── scripts/                    # 脚本目录
│   ├── deploy/
│   ├── migration/
│   └── build.sh
│
├── test/                       # 测试文件
│   ├── e2e/                      # 端到端测试
│   └── integration/              # 集成测试
│
├── web/                        # Web资源（可选）
│   └── static/
│
├── deployments/                # 部署配置
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── k8s/
│   └── nginx/
│
├── docs/                       # 文档
│   ├── api/
│   ├── architecture/
│   └── database/
│
├── bin/                        # 编译后的二进制文件（.gitignore）
│   └── myapp
│
├── log/                        # 日志文件目录（.gitignore）
│   ├── app/
│   ├── access/
│   └── error/
│
├── tmp/                        # 临时文件（.gitignore）
│
├── storage/                    # 存储目录（上传文件等，.gitignore）
│   ├── uploads/
│   └── cache/
│
├── .env.example                # 环境变量示例
├── .env.local                  # 本地环境配置（.gitignore）
├── go.mod
├── go.sum
├── Makefile                    # 构建命令
├── README.md
└── .gitignore
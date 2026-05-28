# 错误封装机制演进总结

本项目最初的错误处理方式较为分散。早期代码中，handler 层会直接返回 `err.Error()`，仓储层可能直接向上返回 MySQL 或 SQL 原始错误，service 层也存在若干通过 `errors.New` 临时创建的错误。这种做法虽然可以快速定位问题，但会带来两个明显缺陷：第一，前端可能直接看到数据库错误、SQL 错误、Go 校验库错误等内部实现细节；第二，项目各层对错误的表达不统一，后续维护和前后端联调成本较高。

## 统一业务错误结构

为了改善这一问题，项目首先引入了统一业务错误结构 `BusinessError`。该结构包含 `Code` 和 `Message` 两个核心字段，其中 `Code` 用作前后端稳定识别错误类型，`Message` 用于向用户展示友好的错误提示。此时，handler 层不再直接返回 `err.Error()`，而是通过统一的 `respondError` 函数将业务错误转换为标准 JSON 响应，例如：

```json
{
  "code": "ORDER_NOT_FOUND",
  "message": "找不到该订单"
}
```

这一阶段实现了错误响应格式的统一，也避免了大部分后端内部错误直接暴露给前端。

## 错误封装 2.0

但随着项目扩展，错误定义继续放在 `internal/service` 下开始暴露出新的结构问题。认证中间件需要使用认证错误，仓储层需要返回订单不存在、唯一键冲突、订单状态异常等错误，handler 层也需要根据错误码生成 HTTP 响应。如果所有错误都定义在 service 层，就会导致 middleware 或 repository 反向依赖 service，从而破坏项目原有的分层结构。因此，项目进入“错误封装 2.0”阶段，将错误定义从 service 层抽离为独立的横切基础设施模块。

错误封装 2.0 中，项目新增 `internal/apperror` 包，用于集中管理全项目的权威错误定义。该包由三个核心文件组成：`business_error.go`、`catalog.go` 和 `http_status.go`。

`business_error.go` 负责定义错误结构和错误工具函数。新的 `BusinessError` 不仅包含 `Code` 和 `Message`，还增加了 `Cause` 字段，用于保存底层错误原因。通过实现 `Unwrap()` 方法，项目可以利用 Go 标准错误链机制保留底层错误。例如，当数据库查询失败时，仓储层可以返回 `apperror.Wrap(apperror.ErrInternal, err)`。这样，前端仍然只能看到统一的内部错误提示，而后端日志或调试过程中仍然可以追踪到底层 MySQL 错误原因。

`catalog.go` 负责定义项目中所有权威错误对象。错误按照领域分为通用错误、认证鉴权错误、用户错误、订单错误、撮合事件错误等。例如：`ErrInvalidInput`、`ErrInvalidToken`、`ErrUserAlreadyExists`、`ErrOrderNotFound`、`ErrOrderNotFillable`、`ErrMatchEventInvalid` 等。这样可以避免不同包重复定义语义相同的错误，也让错误码体系具备清晰的领域边界。

`http_status.go` 负责维护错误码到 HTTP 状态码的映射关系。例如，认证失败映射为 401，权限不足映射为 403，资源不存在映射为 404，状态冲突映射为 409，内部错误映射为 500。handler 层不需要关心每个错误具体应该返回哪个 HTTP 状态码，只需要调用 `apperror.HTTPStatusOf(err)`、`apperror.CodeOf(err)` 和 `apperror.MessageOf(err)` 即可生成统一响应。

错误封装 2.0 后，项目各层的错误职责更加清晰。仓储层负责识别数据库错误和业务状态异常。对于查询不到数据的情况，返回明确的业务错误，例如 `ErrUserNotFound` 或 `ErrOrderNotFound`；对于唯一键冲突，MySQL 仓储层通过 `isMySQLDuplicateEntry` 识别错误号 1062，并转换为 `ErrUserAlreadyExists` 或幂等命中；对于普通数据库错误、扫描错误、事务提交错误，则统一使用 `wrapDBError(err)` 包装为 `ErrInternal` 并保留底层 cause。

service 层负责业务规则校验和业务流程编排，例如校验密码长度、角色合法性、订单方向权限、订单类型规则等。service 层直接返回 `apperror.ErrXXX`，而不再创建临时错误。对于调用仓储层得到的错误，如果已经是业务错误则原样返回；如果是内部错误，则由仓储层或 service 层包装为统一内部错误。

handler 层负责 HTTP 入参解析和响应生成，不再直接拼接或返回 `err.Error()`。所有错误统一通过 `respondError(c, err)` 返回。该函数内部根据 `apperror` 提供的状态码、错误码和错误消息生成标准响应。因此，即使底层发生 SQL 错误，前端也只会看到 `INTERNAL_ERROR` 和友好提示，而不会看到数据库表名、字段名或 SQL 语句。

middleware 层不依赖 handler，也不依赖 service，而是直接使用 `apperror` 中的认证鉴权错误。由于 handler 的 `respondError` 是非导出的包内函数，middleware 中可定义自己的 `abortWithError` 辅助函数，基于 `apperror.HTTPStatusOf`、`CodeOf` 和 `MessageOf` 返回认证失败响应。这样可以避免 middleware 反向依赖 handler 或 service，保持分层清晰。

在 MySQL 仓储包中，`errors.go` 负责处理 MySQL 方言相关的错误识别和包装。例如，`isMySQLDuplicateEntry` 用于识别唯一键冲突，`isMySQLDeadlock` 用于识别死锁，`isMySQLLockWaitTimeout` 用于识别锁等待超时，`wrapDBError` 用于将普通数据库错误包装为统一内部错误。这些函数属于 MySQL 基础设施细节，因此保留在 `internal/repository/mysql` 下，而不是放进通用的 `apperror` 包中。

通过错误封装 2.0，项目完成了从“能返回错误”到“错误可分类、可追踪、可隐藏内部细节、可维护”的升级。前端接收稳定的错误码和用户友好的错误信息；后端保留底层原因链用于排查；仓储层不再向前端泄露数据库错误；service 层不再承担所有错误定义；middleware、handler、repository 都可以共享同一套权威错误体系。这使系统的错误处理更加符合分层架构和工程化要求。

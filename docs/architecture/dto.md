# DTO 设计

项目中有两套 DTO：
```txt
internal/api/dto
internal/repository/dto
```

它们的区别是：

| DTO 类型                    | 所属层     | 面向对象            | 用途                          |
| --------------------------- | --------- | ------------------ | ----------------------------- |
| `internal/api/dto/request`  | API 层    | 前端 HTTP 请求      | 绑定 JSON / query / form 参数  |
| `internal/api/dto/response` | API 层    | 前端 HTTP 响应      | 规定接口返回格式                |
| `internal/repository/dto`   | 仓储接口层 | service/repository | 表示查询条件、统计投影、分页结果 |

- 完成 Issue-E2-S2-I16
-- 功能总结
实现 Token 刷新 API（Refresh）：从 HttpOnly Cookie 读取 refresh_token，校验存在性、是否撤销、是否过期、用户状态（disabled 拒绝），通过后签发新的 AccessToken。
-- 涉及文件
src/internal/auth/handler.go,
src/internal/auth/refresh_test.go

- 完成 Issue-E2-S3-I18
-- 功能总结
实现 AuthMiddleware：Bearer Token 提取、JWT 验签、黑名单校验、错误码映射（TOKEN_EXPIRED/TOKEN_REVOKED/TOKEN_INVALID/AUTH_REQUIRED）、并将 UserContext 注入 request context。
-- 涉及文件
src/internal/auth/middleware.go,
src/internal/auth/middleware_test.go

- 完成 Issue-E2-S3-I19
-- 功能总结
实现 RequireRole 中间件：基于 Viewer/Operator/Admin 权重比较进行权限拦截；权限不足返回 403 且包含 required_role。
-- 涉及文件
src/internal/auth/middleware.go,
src/internal/auth/middleware_test.go

- 完成 Issue-E2-S2-I17
-- 功能总结
实现注销 API（Logout）：校验 AccessToken 后将 jti 加入黑名单；若存在 refresh_token Cookie 则撤销对应 RefreshToken；同时下发过期 Cookie 清除客户端 refresh_token。
-- 涉及文件
src/internal/auth/handler.go,
src/internal/auth/logout_test.go

- 完成 Issue-E2-S4-I20
-- 功能总结
实现用户管理核心接口：GET /users/me、PUT /users/me/password、GET /users（Admin）、PUT /users/{id}/role、DELETE /users/{id}、POST /users/{id}/disable；实现 Admin 自操作保护与最后 Admin 保护。
-- 涉及文件
src/internal/auth/user_management.go,
src/internal/auth/user_management_test.go

- 完成 Issue-E3-S1-I21
-- 功能总结
实现 SystemctlService（Start/Stop/Restart/Status），支持 30s 超时控制、service 名称安全校验、systemctl show 输出解析，并提供可注入 Executor 便于单元测试 Mock。
-- 涉及文件
src/internal/gateway/systemctl.go,
src/internal/gateway/systemctl_test.go

- 完成 Issue-E8-S1-I41
-- 功能总结
实现任务 Repository：Create/FindByID/UpdateStatus/UpdateResult/List；支持按 status/task_type/created_by 过滤；状态流转时自动写 started_at/finished_at。
-- 涉及文件
src/internal/task/model.go,
src/internal/task/repo.go,
src/internal/task/repo_test.go

- 完成 Issue-E8-S3-I44
-- 功能总结
实现任务查询与取消接口：GET 任务列表、GET 任务详情、POST 取消任务。实现 Viewer 仅可查看自己任务，Admin 可查看全部，Operator/Admin 可取消且仅允许取消 PENDING 任务。
-- 涉及文件
src/internal/task/handler.go,
src/internal/task/handler_test.go

- 完成 Issue-E3-S2-I24
-- 功能总结
实现 Gateway 日志读取 API：支持 file/journald 双源，lines 参数校验与上限 1000，file 源日志路径强制白名单（/tmp/openclaw）。
-- 涉及文件
src/internal/gateway/logs.go,
src/internal/gateway/logs_test.go

- 验证结果
-- 执行命令
cd src && go test ./...
-- 结果
全部通过。

- 完成 Issue-E3-S1-I22
-- 功能总结
实现 Gateway 深度状态查询：并发执行 systemctl 状态查询与 `openclaw gateway status --deep`，解析绑定地址/端口、日志路径、Node 路径，并基于 Node 路径检测 `NVMWarning`；当 openclaw 命令超时时返回超时错误且保留 systemctl 部分结果。
-- 涉及文件
src/internal/gateway/systemctl.go,
src/internal/gateway/systemctl_test.go

- 验证结果
-- 执行命令
cd src && "C:\Program Files\Go\bin\go.exe" test ./internal/gateway/...
-- 结果
通过。

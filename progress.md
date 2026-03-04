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

- 验证结果
-- 执行命令
cd src && go test ./...
-- 结果
全部通过。

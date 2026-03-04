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

- 验证结果
-- 执行命令
cd src && go test ./...
-- 结果
全部通过。

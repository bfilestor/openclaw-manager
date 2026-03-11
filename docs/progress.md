- 完成 Issue-E1-S1-I1
-- 功能总结
Issue-E1-S1-I1，完成 Go 模块与目录骨架初始化，包含 `cmd/server` 启动入口、默认配置路径解析、配置路径存在性校验、基础 Makefile 与 golangci 配置。
-- 涉及文件
src/go.mod,
src/cmd/server/main.go,
src/cmd/server/main_test.go,
src/Makefile,
src/.golangci.yml,
src/.gitignore,
doc/todo-list.md

- 完成 Issue-E1-S1-I2
-- 功能总结
Issue-E1-S1-I2，完成 config.toml 加载模块：支持 TOML 解析、默认值填充、`OPENCLAW_JWT_SECRET` 环境变量覆盖、`~` 路径展开、`jwt_secret` 长度校验与 `listen` 地址合法性校验。
-- 涉及文件
src/internal/config/config.go,
src/internal/config/config_test.go,
src/go.mod,
src/go.sum,
doc/todo-list.md

- 完成 Issue-E1-S1-I3
-- 功能总结
Issue-E1-S1-I3，完成 SQLite 存储初始化与迁移模块：自动创建数据库目录、设置 WAL/外键约束、内置 migration 执行与幂等记录（schema_migrations），并创建 users/refresh_tokens/token_blacklist/tasks/revisions/backups 六张核心表。
-- 涉及文件
src/internal/storage/storage.go,
src/internal/storage/storage_test.go,
src/internal/storage/migrations/000001_init.sql,
src/go.mod,
src/go.sum,
doc/todo-list.md

- 完成 Issue-E1-S2-I4
-- 功能总结
Issue-E1-S2-I4，完成 HTTP 服务器与路由基础框架：提供 `/api/v1/health`，新增 API 404 统一 JSON 返回，集成请求日志、CORS、Panic Recovery 中间件，并提供 30 秒优雅关闭能力及信号启动入口。
-- 涉及文件
src/internal/server/server.go,
src/internal/server/signal.go,
src/internal/server/server_test.go,
doc/todo-list.md

- 完成 Issue-E1-S2-I5
-- 功能总结
Issue-E1-S2-I5，完成统一错误响应与请求校验基础框架：新增 `AppError` 结构、标准错误码常量、统一 JSON 错误输出函数 `WriteAppError`，并补充 `BindJSON` 作为统一请求体绑定入口（空体/未知字段/多对象统一返回 VALIDATION_ERROR）。
-- 涉及文件
src/internal/middleware/error.go,
src/internal/middleware/error_test.go,
src/internal/middleware/bind.go,
src/internal/middleware/bind_test.go,
doc/todo-list.md

- 完成 Issue-E1-S2-I6
-- 功能总结
Issue-E1-S2-I6，完成路径白名单安全模块：实现 `PathValidator.Validate` 与 `JoinAndValidate`，支持 `filepath.Clean`、符号链接解析（`EvalSymlinks`）、`~` 路径展开、空字节/空路径拦截，确保路径只能落在允许 base 目录内。
-- 涉及文件
src/internal/storage/pathvalidator.go,
src/internal/storage/pathvalidator_test.go,
doc/todo-list.md

- 完成 Issue-E1-S3-I7
-- 功能总结
Issue-E1-S3-I7，完成测试基础设施增强：引入 `testify` 断言库、新增 `NewTestDB(t)` 测试数据库工厂，并扩展 Makefile 测试目标（`test-unit` / `test-integration` / `test-coverage`）。
-- 涉及文件
src/internal/storage/test_helper.go,
src/internal/storage/test_helper_test.go,
src/Makefile,
src/go.mod,
src/go.sum,
doc/todo-list.md

- 完成 Issue-E1-S3-I8
-- 功能总结
Issue-E1-S3-I8，完成压缩包安全解压工具 `SafeExtract`：支持 zip/tar.gz，具备 zip-slip 路径穿越拦截、绝对路径拦截、单文件与总解压大小限制。
-- 涉及文件
src/internal/storage/extract.go,
src/internal/storage/extract_test.go,
doc/todo-list.md

- 完成 Issue-E1-S3-I9
-- 功能总结
Issue-E1-S3-I9，完成原子写入工具 `AtomicWriteFile`：使用同目录临时文件写入后 `os.Rename` 原子替换，并在失败场景清理临时文件；补充并发写测试确保不会出现半写入内容。
-- 涉及文件
src/internal/storage/atomic.go,
src/internal/storage/atomic_test.go,
doc/todo-list.md

- 完成 Issue-E2-S1-I10
-- 功能总结
Issue-E2-S1-I10，完成用户数据模型与 Repository：实现 users 表的 Create/FindByID/FindByUsername/Update/Delete/List/Count/ExistsAdmin，并补充唯一约束、分页、NotFound 等测试覆盖。
-- 涉及文件
src/internal/user/model.go,
src/internal/user/repo.go,
src/internal/user/repo_test.go,
doc/todo-list.md

- 完成 Issue-E2-S1-I11
-- 功能总结
Issue-E2-S1-I11，完成密码服务：基于 bcrypt(cost=12) 实现 Hash/Verify/ValidateStrength，支持空密码、长度不足、纯数字/纯字母弱口令拦截。
-- 涉及文件
src/internal/auth/password.go,
src/internal/auth/password_test.go,
src/go.mod,
src/go.sum,
doc/todo-list.md

- 完成 Issue-E2-S1-I12
-- 功能总结
Issue-E2-S1-I12，完成注册接口处理器 `POST /api/v1/auth/register`：支持用户名格式校验、密码强度校验、public_registration 开关、首位用户自动 Admin、后续用户默认 Viewer、重复用户名冲突处理。
-- 涉及文件
src/internal/auth/handler.go,
src/internal/auth/handler_test.go,
doc/todo-list.md

- 完成 Issue-E2-S2-I13
-- 功能总结
Issue-E2-S2-I13，完成 JWT 服务：实现 HS256 AccessToken 签发与验证，Claims 包含 `sub/role/jti/exp/iat`，支持过期校验、签名校验、黑名单回调校验，并提供 RefreshToken 生成入口。
-- 涉及文件
src/internal/auth/jwt.go,
src/internal/auth/jwt_test.go,
src/go.mod,
src/go.sum,
doc/todo-list.md

- 完成 Issue-E2-S2-I14
-- 功能总结
Issue-E2-S2-I14，完成 RefreshToken 与黑名单仓储：支持 SHA-256 token 哈希、保存/查询/撤销/按用户全撤销/过期清理；新增黑名单写入、查询、过期清理能力。
-- 涉及文件
src/internal/auth/token_repo.go,
src/internal/auth/token_repo_test.go,
doc/todo-list.md

- 完成 Issue-E2-S2-I15
-- 功能总结
Issue-E2-S2-I15，完成登录处理器：实现用户名密码校验、禁用账号拦截、签发 AccessToken+RefreshToken、RefreshToken 落库、HttpOnly+Secure+SameSite=Strict Cookie 下发、更新 last_login_at，且响应体不返回 refresh_token 明文。
-- 涉及文件
src/internal/auth/handler.go,
src/internal/auth/login_test.go,
doc/todo-list.md

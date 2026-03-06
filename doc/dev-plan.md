# OpenClaw Web 管理平台 — 开发规划文档

> **方法论**：TDD（测试驱动开发）+ 敏捷开发（Scrum）  
> **文档版本**：v1.0 | 2026-03-04  
> **对应需求**：openclaw_requirements_v1.1.docx  
> **技术栈**：Go 1.22+ / Vue 3 + Element Plus / SQLite / JWT / SSE

---

## 目录

1. [总览与约定](#总览与约定)
2. [Epic 索引](#epic-索引)
3. [Epic E1 — 项目基础骨架与工程化](#epic-e1--项目基础骨架与工程化)
4. [Epic E2 — 用户体系与认证授权](#epic-e2--用户体系与认证授权)
5. [Epic E3 — Gateway 生命周期管理](#epic-e3--gateway-生命周期管理)
6. [Epic E4 — 配置管理与版本历史](#epic-e4--配置管理与版本历史)
7. [Epic E5 — Agent 管理与 Channel Binding](#epic-e5--agent-管理与-channel-binding)
8. [Epic E6 — Skills 管理](#epic-e6--skills-管理)
9. [Epic E7 — 备份与还原](#epic-e7--备份与还原)
10. [Epic E8 — 任务系统与实时日志](#epic-e8--任务系统与实时日志)
11. [Epic E9 — 前端框架与通用组件](#epic-e9--前端框架与通用组件)
12. [Epic E10 — 集成测试与部署](#epic-e10--集成测试与部署)
13. [Issue 依赖关系总图](#issue-依赖关系总图)
14. [Sprint 规划建议](#sprint-规划建议)

---

## 总览与约定

### TDD 开发规范

每个 Issue 遵循以下开发顺序：
1. **Red**：先写失败的单元测试/集成测试
2. **Green**：写最少代码让测试通过
3. **Refactor**：重构代码保持测试通过

### Issue 字段说明

| 字段 | 说明 |
|------|------|
| **ID** | 唯一标识，格式 `E{epic编号}-S{story编号}-I{issue编号}` |
| **Story Points** | 复杂度估算（1=半天, 2=1天, 3=1.5天, 5=2.5天, 8=4天） |
| **优先级** | P0=阻塞型必做 / P1=核心功能 / P2=增强功能 |
| **依赖** | 必须先完成的 Issue ID 列表 |
| **测试类型** | Unit=单元测试 / Integration=集成测试 / E2E=端到端测试 |

### 角色权重约定（贯穿全文）

```
Viewer=1 / Operator=2 / Admin=3
权限检查：user.roleWeight >= required.roleWeight
```

### 目录结构约定

```
openclaw-manager/
├── cmd/server/          # Go 入口
├── internal/
│   ├── auth/            # JWT、bcrypt、中间件
│   ├── user/            # 用户 CRUD
│   ├── gateway/         # systemctl 封装
│   ├── config/          # 配置文件读写
│   ├── agent/           # Agent & Binding
│   ├── skills/          # Skills 管理
│   ├── backup/          # 备份还原
│   ├── task/            # 任务引擎
│   ├── storage/         # SQLite 封装
│   └── middleware/      # HTTP 中间件
├── frontend/            # Vue 3 SPA
│   ├── src/
│   │   ├── views/
│   │   ├── components/
│   │   ├── stores/      # Pinia
│   │   └── api/         # axios 封装
└── tests/
    ├── unit/
    ├── integration/
    └── e2e/
```

---

## Epic 索引

| Epic ID | 名称 | Story 数 | Issue 数 | Sprint |
|---------|------|----------|----------|--------|
| E1 | 项目基础骨架与工程化 | 3 | 12 | Sprint 1 |
| E2 | 用户体系与认证授权 | 4 | 18 | Sprint 1-2 |
| E3 | Gateway 生命周期管理 | 3 | 12 | Sprint 2 |
| E4 | 配置管理与版本历史 | 3 | 11 | Sprint 3 |
| E5 | Agent 管理与 Channel Binding | 3 | 13 | Sprint 3-4 |
| E6 | Skills 管理 | 2 | 9 | Sprint 4 |
| E7 | 备份与还原 | 3 | 12 | Sprint 4-5 |
| E8 | 任务系统与实时日志 | 3 | 11 | Sprint 2 |
| E9 | 前端框架与通用组件 | 3 | 10 | Sprint 1-2 |
| E10 | 集成测试与部署 | 2 | 7 | Sprint 5-6 |
| **合计** | | **29** | **115** | **6 Sprints** |

---

## Epic E1 — 项目基础骨架与工程化

> **目标**：建立 Go 后端框架、SQLite 初始化、配置加载、路由骨架、CI 基础，是所有后续 Epic 的前置依赖。

---

### Story E1-S1：Go 后端框架初始化

**描述**：初始化 Go 模块、选定 HTTP 框架（gin/chi）、建立项目目录结构、配置 Makefile 与 linter。

---

#### Issue E1-S1-I1：Go 模块与目录结构初始化

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：无  
- **测试类型**：Unit

**功能描述**：
- 执行 `go mod init`，选定 HTTP 框架（推荐 `chi` 或 `gin`）
- 建立 `cmd/server/main.go` 入口，支持 `--config` 参数指定配置文件路径
- 建立 `internal/` 各子包骨架（空文件占位）
- 配置 `golangci-lint`、`.gitignore`、`Makefile`（targets: build/test/lint/run）

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E1S1I1-001 | 默认配置路径启动 | 无 `--config` 参数 | 使用 `~/.openclaw-manager/config.toml` 默认路径，启动不 panic | Unit |
| TC-E1S1I1-002 | 自定义配置路径 | `--config /tmp/test.toml` | 读取指定路径 | Unit |
| TC-E1S1I1-003 | 配置文件不存在 | `--config /nonexistent.toml` | 打印明确错误并 exit(1)，不 panic | Unit |
| TC-E1S1I1-004 | `make build` 编译成功 | — | 生成二进制文件，无编译错误 | Unit |
| TC-E1S1I1-005 | `make lint` 无报错 | — | golangci-lint 通过 | Unit |

**验证方法**：
```bash
go build ./cmd/server/
./server --config /nonexistent.toml  # 期望 exit 1
make lint
make test
```

---

#### Issue E1-S1-I2：config.toml 配置加载模块

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E1-S1-I1  
- **测试类型**：Unit

**功能描述**：
- 使用 `viper` 或手写解析，加载 `config.toml`
- 支持环境变量覆盖（`OPENCLAW_JWT_SECRET` 优先于 toml 文件）
- 配置结构体：`ServerConfig{Listen}`, `AuthConfig{JwtSecret, AccessTokenTTL, RefreshTokenTTL, PublicRegistration, PasswordMinLength}`, `PathsConfig{OpenclawHome, ManagerHome}`
- 配置校验：`jwt_secret` 不可为空或少于 32 字节；`listen` 必须是合法 `host:port`

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E1S1I2-001 | 正常加载完整 config.toml | 合法 toml 文件 | 所有字段正确解析 | Unit |
| TC-E1S1I2-002 | jwt_secret 为空 | `jwt_secret = ""` | 返回 `ErrConfigInvalid`，消息含"jwt_secret" | Unit |
| TC-E1S1I2-003 | jwt_secret 少于 32 字节 | `jwt_secret = "short"` | 返回 `ErrConfigInvalid` | Unit |
| TC-E1S1I2-004 | 环境变量覆盖 jwt_secret | env `OPENCLAW_JWT_SECRET=validlongkey32bytes` | 使用环境变量值而非 toml 值 | Unit |
| TC-E1S1I2-005 | listen 地址非法 | `listen = "invalid"` | 返回 `ErrConfigInvalid`，消息含"listen" | Unit |
| TC-E1S1I2-006 | access_token_ttl 使用默认值 | toml 中省略该字段 | 默认 15min | Unit |
| TC-E1S1I2-007 | 路径中 `~` 展开 | `manager_home = "~/.openclaw-manager"` | 展开为绝对路径 | Unit |

**验证方法**：
```bash
go test ./internal/config/... -v -count=1
# 覆盖率要求 ≥ 90%
go test ./internal/config/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

---

#### Issue E1-S1-I3：SQLite 存储层初始化与迁移

- **Story Points**：3  
- **优先级**：P0  
- **依赖**：E1-S1-I2  
- **测试类型**：Unit + Integration

**功能描述**：
- 使用 `database/sql` + `mattn/go-sqlite3`（或 `modernc.org/sqlite`，纯 Go 可选）
- 实现数据库初始化：自动创建数据目录、执行建表 SQL
- 实现版本化迁移（embedded SQL 文件，migrate 顺序执行，记录已执行版本）
- 建表：`users`、`refresh_tokens`、`token_blacklist`、`tasks`、`revisions`、`backups`
- 所有表启用 WAL 模式（`PRAGMA journal_mode=WAL`）、外键约束（`PRAGMA foreign_keys=ON`）

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E1S1I3-001 | 首次初始化（数据库文件不存在） | 指定新路径 | 自动创建文件，所有表存在 | Integration |
| TC-E1S1I3-002 | 重复初始化（数据库已存在） | 已有数据库 | 不报错，数据保留，不重复建表 | Integration |
| TC-E1S1I3-003 | WAL 模式已启用 | 初始化后 | `PRAGMA journal_mode` 返回 `wal` | Integration |
| TC-E1S1I3-004 | 外键约束生效 | 插入 tasks 时 created_by 引用不存在 user_id | 返回 foreign key 约束错误 | Integration |
| TC-E1S1I3-005 | 迁移幂等性 | 执行迁移两次 | 第二次无副作用，版本号不重复 | Integration |
| TC-E1S1I3-006 | 数据目录不存在时自动创建 | `manager_home` 指向不存在目录 | 自动创建目录 + 数据库文件 | Integration |
| TC-E1S1I3-007 | 数据库文件无写权限 | `chmod 400` 数据库文件 | 返回明确错误，不 panic | Integration |

**验证方法**：
```bash
go test ./internal/storage/... -v -count=1 -tags integration
# 使用 :memory: 数据库进行单元测试
```

---

### Story E1-S2：HTTP 路由骨架与中间件基础

**描述**：建立 HTTP 服务器、路由组织结构、基础中间件（日志、Recovery、CORS），为后续 API 开发提供框架。

---

#### Issue E1-S2-I4：HTTP 服务器与路由注册

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E1-S1-I1  
- **测试类型**：Unit + Integration

**功能描述**：
- 启动 HTTP 服务监听 `config.Listen`（默认 `127.0.0.1:18790`）
- 路由分组：`/api/v1/auth/**`（公开）、`/api/v1/**`（受保护）、`/`（静态文件）
- 中间件：请求日志（含耗时）、Panic Recovery（返回 500 JSON）、CORS（限制 Origin）
- 优雅关闭：接收 SIGTERM/SIGINT，等待进行中请求完成（超时 30s）

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E1S2I4-001 | 健康检查接口 | `GET /api/v1/health` | `200 {"status":"ok","version":"..."}` | Integration |
| TC-E1S2I4-002 | 未知路由 | `GET /api/v1/nonexistent` | `404 {"error":"not found","code":"NOT_FOUND"}` | Integration |
| TC-E1S2I4-003 | Panic Recovery | Handler 内 panic | `500 {"error":"internal server error"}` 不崩溃 | Integration |
| TC-E1S2I4-004 | 优雅关闭 | 有进行中请求时发 SIGTERM | 等待请求完成后退出，不丢弃响应 | Integration |
| TC-E1S2I4-005 | CORS 头正确 | 请求含 `Origin: http://localhost:5173` | 响应含 `Access-Control-Allow-Origin` | Integration |
| TC-E1S2I4-006 | 静态文件服务 | `GET /` | 返回 Vue SPA 的 `index.html` | Integration |

**验证方法**：
```bash
go test ./internal/server/... -v -count=1
curl -i http://127.0.0.1:18790/api/v1/health
```

---

#### Issue E1-S2-I5：统一错误响应与请求校验框架

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E1-S2-I4  
- **测试类型**：Unit

**功能描述**：
- 定义统一错误类型 `AppError{Code, Message, StatusCode}`
- 实现 `ErrorHandler` 中间件：将 `AppError` 转换为标准 JSON 响应
- 错误码常量：`AUTH_REQUIRED`、`PERMISSION_DENIED`、`NOT_FOUND`、`VALIDATION_ERROR`、`CONFLICT`、`INTERNAL_ERROR`
- 请求体绑定与校验辅助函数（基于 `go-playground/validator`）
- 错误响应格式：`{"error":"message","code":"ERROR_CODE","required_role":"Operator"}`（required_role 仅权限错误时附加）

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E1S2I5-001 | 401 错误格式 | 触发 AUTH_REQUIRED | `{"error":"unauthorized","code":"AUTH_REQUIRED"}` | Unit |
| TC-E1S2I5-002 | 403 含 required_role | 触发 PERMISSION_DENIED | `{"error":"forbidden","code":"PERMISSION_DENIED","required_role":"Operator"}` | Unit |
| TC-E1S2I5-003 | 400 校验错误 | 必填字段缺失 | `{"error":"validation failed","code":"VALIDATION_ERROR","fields":{"username":"required"}}` | Unit |
| TC-E1S2I5-004 | 500 不泄露内部信息 | 内部 error | 响应不含 stack trace，只有通用消息 | Unit |
| TC-E1S2I5-005 | 409 冲突错误 | 触发 CONFLICT | `{"error":"...","code":"CONFLICT"}` | Unit |

**验证方法**：
```bash
go test ./internal/middleware/... -v -count=1
```

---

#### Issue E1-S2-I6：路径白名单安全模块

- **Story Points**：3  
- **优先级**：P0  
- **依赖**：E1-S1-I2  
- **测试类型**：Unit

**功能描述**：
- 实现 `PathValidator` 结构体，接收白名单 base 路径列表
- 核心方法 `Validate(inputPath string) (safePath string, err error)`：
  1. `filepath.Clean(inputPath)`
  2. `filepath.EvalSymlinks(inputPath)`（若目标存在）
  3. 校验 realpath 是否在某个 base 下（`strings.HasPrefix`）
- 提供 `JoinAndValidate(base, subPath string) (string, error)` 用于拼接后校验
- 白名单配置：`~/.openclaw/`、`~/.config/systemd/user/openclaw-gateway.service`、`/tmp/openclaw/`、`~/.openclaw-manager/`

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E1S2I6-001 | 合法路径通过 | `~/.openclaw/openclaw.json` | 返回绝对路径，无错误 | Unit |
| TC-E1S2I6-002 | 路径穿越攻击（../） | `~/.openclaw/../../../etc/passwd` | 返回 `ErrPathNotAllowed` | Unit |
| TC-E1S2I6-003 | 绝对路径穿越 | `/etc/shadow` | 返回 `ErrPathNotAllowed` | Unit |
| TC-E1S2I6-004 | 符号链接指向白名单外 | 符号链接 `~/.openclaw/link -> /etc/passwd` | EvalSymlinks 后返回 `ErrPathNotAllowed` | Unit |
| TC-E1S2I6-005 | base 路径本身合法 | 传入白名单 base 路径 | 通过校验 | Unit |
| TC-E1S2I6-006 | 空路径 | `""` | 返回 `ErrPathEmpty` | Unit |
| TC-E1S2I6-007 | 路径包含空字节 | `"/tmp/\x00evil"` | 返回错误（filepath.Clean 处理后也不合法） | Unit |
| TC-E1S2I6-008 | JoinAndValidate 正常用例 | base=`~/.openclaw/skills/`, sub=`my-skill` | 返回拼接后安全路径 | Unit |
| TC-E1S2I6-009 | JoinAndValidate 子路径穿越 | sub=`../../../etc/cron.d/evil` | 返回 `ErrPathNotAllowed` | Unit |

**验证方法**：
```bash
go test ./internal/storage/pathvalidator_test.go -v -count=1
# 所有 9 个用例必须 PASS
```

---

### Story E1-S3：CI/CD 与测试基础设施

---

#### Issue E1-S3-I7：单元测试框架与覆盖率配置

- **Story Points**：1  
- **优先级**：P0  
- **依赖**：E1-S1-I1  
- **测试类型**：Unit

**功能描述**：
- 配置 `testify/assert` 与 `testify/mock`
- 建立测试 Helper：内存 SQLite 数据库工厂函数 `NewTestDB(t)`
- 配置 `go test` 覆盖率报告，目标 ≥ 80%
- `Makefile` 新增 `make test-unit`、`make test-integration`、`make test-coverage`

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E1S3I7-001 | `NewTestDB(t)` 返回可用内存数据库 | 数据库可读写，测试结束后自动清理 | Unit |
| TC-E1S3I7-002 | `make test-unit` 运行所有单元测试 | 全部通过，输出覆盖率报告 | Unit |
| TC-E1S3I7-003 | `make test-coverage` 不低于 80% | coverage ≥ 80% | Unit |

---

#### Issue E1-S3-I8：Zip-slip 防护工具函数

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E1-S2-I6  
- **测试类型**：Unit

**功能描述**：
- 实现 `SafeExtract(archivePath, destBase string) error`
- 支持 `.zip` 和 `.tar.gz` 格式
- 逐 entry 校验：禁止绝对路径、禁止 `..` 路径段、禁止符号链接指向 destBase 外
- 文件大小限制：单文件不超过 50MB，总解压不超过 200MB

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E1S3I8-001 | 正常 zip 解压 | 合法 zip 文件 | 文件解压到 destBase 内 | Unit |
| TC-E1S3I8-002 | zip-slip 攻击（zip 内含 `../evil`） | 含 `../../etc/passwd` 的 zip | 返回 `ErrZipSlip`，不创建任何文件 | Unit |
| TC-E1S3I8-003 | 绝对路径 entry | entry 路径为 `/etc/passwd` | 返回 `ErrZipSlip` | Unit |
| TC-E1S3I8-004 | tar.gz 正常解压 | 合法 tar.gz | 正确解压 | Unit |
| TC-E1S3I8-005 | tar.gz zip-slip | 含 `../` 的 tar.gz | 返回 `ErrZipSlip` | Unit |
| TC-E1S3I8-006 | 单文件超过 50MB | 含 60MB 文件的 zip | 返回 `ErrFileTooLarge` | Unit |
| TC-E1S3I8-007 | 总解压超过 200MB | 多文件总计 201MB | 返回 `ErrExtractTooLarge` | Unit |
| TC-E1S3I8-008 | 空压缩包 | 0 entry zip | 成功，不报错 | Unit |
| TC-E1S3I8-009 | 符号链接指向解压目录外 | symlink -> `/tmp/evil` | 返回 `ErrZipSlip` | Unit |

**验证方法**：
```bash
go test ./internal/storage/extract_test.go -v -count=1
```

---

#### Issue E1-S3-I9：原子文件写入工具函数

- **Story Points**：1  
- **优先级**：P0  
- **依赖**：E1-S2-I6  
- **测试类型**：Unit

**功能描述**：
- 实现 `AtomicWriteFile(path string, data []byte, perm os.FileMode) error`
- 流程：写临时文件（同目录 `.tmp_RANDOM`） → `os.Rename` 原子替换
- 若写入过程失败，自动删除临时文件

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E1S3I9-001 | 正常写入 | 文件内容正确，权限正确 | Unit |
| TC-E1S3I9-002 | 写入中途模拟失败 | 原文件不被破坏，临时文件已清理 | Unit |
| TC-E1S3I9-003 | 目标目录不存在 | 返回明确 IO 错误 | Unit |
| TC-E1S3I9-004 | 并发写入同一文件 | 最终文件内容为某一次写入的完整内容，不出现半写状态 | Unit |

---

## Epic E2 — 用户体系与认证授权

> **目标**：实现用户注册、登录、JWT 认证、角色权限中间件、用户管理接口，是所有业务接口的安全前置依赖。

---

### Story E2-S1：用户注册与密码管理

---

#### Issue E2-S1-I10：用户数据模型与 Repository

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E1-S1-I3  
- **测试类型**：Unit + Integration

**功能描述**：
- 定义 `User` 结构体：`UserID(uuid)`, `Username`, `PasswordHash`, `Role(Viewer/Operator/Admin)`, `Status(active/disabled)`, `CreatedAt`, `LastLoginAt`, `UpdatedAt`
- 实现 `UserRepository` 接口及 SQLite 实现：
  - `Create(user) error`
  - `FindByID(id) (*User, error)`
  - `FindByUsername(username) (*User, error)`
  - `Update(user) error`
  - `Delete(id) error`
  - `List(offset, limit int) ([]*User, int, error)`
  - `Count() (int, error)`（用于检测首位用户）
  - `ExistsAdmin() (bool, error)`（至少有一个 Admin 存在）

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E2S1I10-001 | 创建用户成功 | 合法用户对象 | 插入成功，返回无错误 | Integration |
| TC-E2S1I10-002 | 用户名重复 | 同 username 插入两次 | 第二次返回唯一约束错误 | Integration |
| TC-E2S1I10-003 | FindByUsername 存在 | 已插入的 username | 返回正确 User 对象 | Integration |
| TC-E2S1I10-004 | FindByUsername 不存在 | 未插入的 username | 返回 `ErrNotFound` | Integration |
| TC-E2S1I10-005 | Count() 空表 | 空数据库 | 返回 0 | Integration |
| TC-E2S1I10-006 | ExistsAdmin() 无 Admin | 只有 Viewer | 返回 false | Integration |
| TC-E2S1I10-007 | ExistsAdmin() 有 Admin | 有 Admin 用户 | 返回 true | Integration |
| TC-E2S1I10-008 | Delete 用户存在 | 已有 user_id | 删除成功 | Integration |
| TC-E2S1I10-009 | Delete 用户不存在 | 不存在 user_id | 返回 `ErrNotFound` | Integration |
| TC-E2S1I10-010 | List 分页正确 | limit=2, offset=1 | 返回第 2-3 条，total 正确 | Integration |

---

#### Issue E2-S1-I11：密码哈希服务

- **Story Points**：1  
- **优先级**：P0  
- **依赖**：E1-S1-I1  
- **测试类型**：Unit

**功能描述**：
- 实现 `PasswordService`：
  - `Hash(plain string) (hash string, err error)`：bcrypt cost=12
  - `Verify(plain, hash string) bool`：bcrypt 比对
- 密码强度校验 `ValidateStrength(plain string) error`：最短长度（配置项，默认 8）、必须同时含字母和数字

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E2S1I11-001 | 哈希结果不是明文 | `"P@ssw0rd"` | 返回值不等于输入，以 `$2a$` 开头 | Unit |
| TC-E2S1I11-002 | 哈希结果每次不同（salt） | 同一密码哈希两次 | 两次结果不同 | Unit |
| TC-E2S1I11-003 | Verify 正确密码 | 密码与其哈希 | 返回 true | Unit |
| TC-E2S1I11-004 | Verify 错误密码 | 不同密码与哈希 | 返回 false | Unit |
| TC-E2S1I11-005 | 密码长度不足 | `"abc123"` (6位) | 返回 `ErrPasswordTooShort` | Unit |
| TC-E2S1I11-006 | 密码全数字 | `"12345678"` | 返回 `ErrPasswordWeak` | Unit |
| TC-E2S1I11-007 | 密码全字母 | `"abcdefgh"` | 返回 `ErrPasswordWeak` | Unit |
| TC-E2S1I11-008 | 密码为空 | `""` | 返回 `ErrPasswordEmpty` | Unit |
| TC-E2S1I11-009 | 合法密码通过校验 | `"Pass1234"` | 返回 nil | Unit |
| TC-E2S1I11-010 | bcrypt cost 正确 | 哈希后解析 | cost 字段 = 12 | Unit |

---

#### Issue E2-S1-I12：用户注册 API — POST /api/v1/auth/register

- **Story Points**：3  
- **优先级**：P0  
- **依赖**：E2-S1-I10, E2-S1-I11, E1-S2-I5  
- **测试类型**：Unit + Integration

**功能描述**：
- 请求体：`{"username":"alice","password":"P@ss1234"}`
- 业务逻辑：
  1. 校验用户名格式（3-32位，字母数字下划线）
  2. 校验密码强度
  3. 检查用户名是否已存在
  4. `Count()==0` 时赋予 Admin，否则赋予 Viewer
  5. bcrypt 哈希密码，插入 users 表
- 成功响应：`201 {"user_id":"...","username":"alice","role":"Admin","created_at":"..."}`
- `public_registration=false` 时返回 `403 {"code":"REGISTRATION_DISABLED"}`

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E2S1I12-001 | 首位用户注册成功 | 合法请求，空用户表 | 201，role=Admin | Integration |
| TC-E2S1I12-002 | 第二位用户默认 Viewer | 合法请求，已有用户 | 201，role=Viewer | Integration |
| TC-E2S1I12-003 | 用户名已存在 | 重复 username | 409，code=USERNAME_EXISTS | Integration |
| TC-E2S1I12-004 | 用户名太短（2位） | `{"username":"ab","password":"Pass1234"}` | 400，code=VALIDATION_ERROR | Integration |
| TC-E2S1I12-005 | 用户名太长（33位） | 33 字符 username | 400，code=VALIDATION_ERROR | Integration |
| TC-E2S1I12-006 | 用户名含非法字符 | `{"username":"alice!","password":"Pass1234"}` | 400，code=VALIDATION_ERROR | Integration |
| TC-E2S1I12-007 | 密码强度不足（纯数字） | `{"password":"12345678"}` | 400，code=PASSWORD_WEAK | Integration |
| TC-E2S1I12-008 | 密码太短 | `{"password":"P1"}` | 400，code=PASSWORD_TOO_SHORT | Integration |
| TC-E2S1I12-009 | 公开注册关闭 | `public_registration=false` | 403，code=REGISTRATION_DISABLED | Integration |
| TC-E2S1I12-010 | 请求体为空 | `{}` | 400，code=VALIDATION_ERROR | Integration |
| TC-E2S1I12-011 | username 包含 SQL 注入尝试 | `"alice'; DROP TABLE users;--"` | 400 校验拒绝（含非法字符） | Integration |
| TC-E2S1I12-012 | 超长密码（1000字符） | 1000 字符密码 | 400，code=PASSWORD_TOO_LONG（建议上限 128） | Integration |

---

### Story E2-S2：JWT 认证与 Token 管理

---

#### Issue E2-S2-I13：JWT 签发与验证服务

- **Story Points**：3  
- **优先级**：P0  
- **依赖**：E2-S1-I10, E1-S1-I2  
- **测试类型**：Unit

**功能描述**：
- 使用 `golang-jwt/jwt` 库，HS256 算法
- `JWTService` 接口：
  - `SignAccessToken(userID, role string) (token string, jti string, err error)`：有效期 AccessTokenTTL，payload 含 `sub`(userID), `role`, `jti`(uuid), `exp`, `iat`
  - `VerifyAccessToken(token string) (*Claims, error)`：验证签名、过期、`jti` 黑名单
  - `SignRefreshToken() (tokenID, rawToken string, err error)`：有效期 RefreshTokenTTL
- `Claims` 结构体：`UserID`, `Role`, `JTI`, `StandardClaims`

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E2S2I13-001 | SignAccessToken 返回合法 JWT | JWT 可被 VerifyAccessToken 成功解析 | Unit |
| TC-E2S2I13-002 | Claims 字段正确 | sub=userID, role 正确, jti 非空 | Unit |
| TC-E2S2I13-003 | 过期 Token 验证失败 | TTL=1ns 的 Token，验证返回 `ErrTokenExpired` | Unit |
| TC-E2S2I13-004 | 错误密钥验证失败 | 用不同密钥签发的 Token | 返回 `ErrTokenInvalid` | Unit |
| TC-E2S2I13-005 | 篡改 payload | 修改 JWT payload 后验证 | 返回 `ErrTokenInvalid` | Unit |
| TC-E2S2I13-006 | 黑名单 jti 验证失败 | jti 已在黑名单 | 返回 `ErrTokenRevoked` | Unit |
| TC-E2S2I13-007 | 空 Token | `""` | 返回 `ErrTokenMissing` | Unit |
| TC-E2S2I13-008 | 格式错误 Token | `"notajwt"` | 返回 `ErrTokenInvalid` | Unit |
| TC-E2S2I13-009 | 两次 Sign 的 jti 不同 | 连续调用两次 | 两个 jti UUID 不同 | Unit |

---

#### Issue E2-S2-I14：RefreshToken Repository 与黑名单

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E1-S1-I3, E2-S2-I13  
- **测试类型**：Integration

**功能描述**：
- `RefreshTokenRepository`：`Save`, `FindByHash`, `Revoke(tokenID)`, `RevokeAllByUser(userID)`, `DeleteExpired`
- 存储时对 rawToken 做 SHA-256 哈希，不存明文
- `TokenBlacklistRepository`：`Add(jti, expiresAt)`, `Exists(jti) bool`, `CleanExpired()`

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E2S2I14-001 | Save 后 FindByHash 成功 | 正确返回 RefreshToken 记录 | Integration |
| TC-E2S2I14-002 | Revoke 后 FindByHash 返回 revoked=true | 已撤销标记生效 | Integration |
| TC-E2S2I14-003 | RevokeAllByUser 撤销该用户所有 token | 该用户 tokens 全部 revoked=true | Integration |
| TC-E2S2I14-004 | DeleteExpired 清理过期记录 | 已过期的 token 被删除 | Integration |
| TC-E2S2I14-005 | 黑名单 Add + Exists | jti 加入后 Exists 返回 true | Integration |
| TC-E2S2I14-006 | 黑名单不存在的 jti | Exists 返回 false | Integration |
| TC-E2S2I14-007 | CleanExpired 清理黑名单 | 过期 jti 被清除，有效 jti 保留 | Integration |
| TC-E2S2I14-008 | RefreshToken 哈希存储 | 数据库中 token_hash 非原始值 | Integration |

---

#### Issue E2-S2-I15：登录 API — POST /api/v1/auth/login

- **Story Points**：3  
- **优先级**：P0  
- **依赖**：E2-S2-I13, E2-S2-I14, E2-S1-I11  
- **测试类型**：Integration

**功能描述**：
- 请求体：`{"username":"alice","password":"P@ss1234"}`
- 业务逻辑：
  1. FindByUsername，不存在返回 401（不区分"用户不存在"和"密码错误"，防信息泄露）
  2. 检查账号 status=active，disabled 返回 403
  3. bcrypt Verify，失败返回 401
  4. 签发 AccessToken + RefreshToken
  5. RefreshToken 存入数据库，通过 `Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict` 设置
  6. 更新 `last_login_at`
- 成功响应：`200 {"access_token":"...","expires_in":900,"token_type":"Bearer","user":{...}}`

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E2S2I15-001 | 正常登录 | 正确用户名密码 | 200，含 access_token，Set-Cookie 含 refresh_token | Integration |
| TC-E2S2I15-002 | 用户不存在 | 不存在的 username | 401，code=INVALID_CREDENTIALS（不区分） | Integration |
| TC-E2S2I15-003 | 密码错误 | 正确 username + 错误密码 | 401，code=INVALID_CREDENTIALS | Integration |
| TC-E2S2I15-004 | 账号禁用 | status=disabled 用户 | 403，code=ACCOUNT_DISABLED | Integration |
| TC-E2S2I15-005 | Cookie 属性正确 | 正常登录 | Set-Cookie 含 HttpOnly、SameSite=Strict | Integration |
| TC-E2S2I15-006 | 响应不含 refresh_token 明文 | 正常登录 | JSON 响应体中无 refresh_token 字段 | Integration |
| TC-E2S2I15-007 | access_token 有效期正确 | 正常登录 | JWT exp - iat = 配置的 AccessTokenTTL | Integration |
| TC-E2S2I15-008 | last_login_at 已更新 | 登录后查库 | last_login_at 被更新为当前时间 | Integration |
| TC-E2S2I15-009 | 请求体缺少 password | `{"username":"alice"}` | 400，code=VALIDATION_ERROR | Integration |

---

#### Issue E2-S2-I16：Token 刷新 API — POST /api/v1/auth/refresh

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E2-S2-I14, E2-S2-I13  
- **测试类型**：Integration

**功能描述**：
- 从请求 Cookie 中取 `refresh_token`
- SHA-256 哈希后查 refresh_tokens 表
- 验证：存在、未撤销、未过期、关联用户状态 active
- 签发新 AccessToken（jti 更新）
- 可选（安全增强）：旋转 RefreshToken（撤销旧的，签发新的）

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E2S2I16-001 | 合法 RefreshToken 换取新 AccessToken | 200，返回新 access_token | Integration |
| TC-E2S2I16-002 | 已撤销 RefreshToken | 401，code=TOKEN_REVOKED | Integration |
| TC-E2S2I16-003 | 过期 RefreshToken | 401，code=TOKEN_EXPIRED | Integration |
| TC-E2S2I16-004 | Cookie 中无 refresh_token | 401，code=AUTH_REQUIRED | Integration |
| TC-E2S2I16-005 | 伪造 RefreshToken | 401，code=TOKEN_INVALID | Integration |
| TC-E2S2I16-006 | 新 AccessToken jti 与旧 AccessToken jti 不同 | 新旧 jti 不同 | Integration |
| TC-E2S2I16-007 | 用户已被禁用时刷新 | 403，code=ACCOUNT_DISABLED | Integration |

---

#### Issue E2-S2-I17：注销 API — POST /api/v1/auth/logout

- **Story Points**：1  
- **优先级**：P1  
- **依赖**：E2-S2-I14, E2-S3-I18  
- **测试类型**：Integration

**功能描述**：
- 需要有效 AccessToken（经 AuthMiddleware）
- 将当前 AccessToken 的 jti 加入 token_blacklist
- 撤销 Cookie 中的 RefreshToken
- 响应：`200 {"message":"logged out"}`，清除 Cookie

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E2S2I17-001 | 正常注销 | 200，jti 进入黑名单，RefreshToken 已撤销 | Integration |
| TC-E2S2I17-002 | 注销后原 AccessToken 无法使用 | 再次调用 /overview 返回 401 | Integration |
| TC-E2S2I17-003 | 注销后 RefreshToken 无法刷新 | 调用 /refresh 返回 401 | Integration |
| TC-E2S2I17-004 | Cookie 被清除 | 响应 Set-Cookie 将 refresh_token 过期 | Integration |
| TC-E2S2I17-005 | 无 Token 注销 | 401，code=AUTH_REQUIRED | Integration |

---

### Story E2-S3：权限中间件

---

#### Issue E2-S3-I18：AuthMiddleware — JWT 验证与用户注入

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E2-S2-I13, E2-S2-I14  
- **测试类型**：Unit + Integration

**功能描述**：
- 从 `Authorization: Bearer <token>` 提取 JWT
- 调用 `JWTService.VerifyAccessToken`
- 查 token_blacklist，若存在则 401
- 将 `UserContext{UserID, Role}` 注入 request context
- 失败统一返回 `401 {"code":"AUTH_REQUIRED"}`

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E2S3I18-001 | 合法 Token，context 注入正确 | Handler 能从 context 取到正确 UserID 和 Role | Unit |
| TC-E2S3I18-002 | 无 Authorization 头 | 401，code=AUTH_REQUIRED | Unit |
| TC-E2S3I18-003 | Bearer 格式错误 | `Authorization: Token xxx` | 401 | Unit |
| TC-E2S3I18-004 | Token 已过期 | 过期 JWT | 401，code=TOKEN_EXPIRED | Unit |
| TC-E2S3I18-005 | Token 在黑名单中 | jti 已加入黑名单 | 401，code=TOKEN_REVOKED | Unit |
| TC-E2S3I18-006 | Token 签名错误 | 篡改的 JWT | 401，code=TOKEN_INVALID | Unit |
| TC-E2S3I18-007 | 中间件不查数据库（黑名单除外） | 合法 Token | 不查 users 表（性能要求） | Unit |

---

#### Issue E2-S3-I19：RequireRole 中间件

- **Story Points**：1  
- **优先级**：P0  
- **依赖**：E2-S3-I18  
- **测试类型**：Unit

**功能描述**：
- `RequireRole(minRole Role) Middleware`
- 从 context 取 `UserContext`，比较 `user.Role.Weight() >= minRole.Weight()`
- 不足时返回 `403 {"code":"PERMISSION_DENIED","required_role":"Operator"}`

**测试用例**：

| 用例编号 | 描述 | 输入 | 预期输出 | 测试类型 |
|----------|------|------|----------|----------|
| TC-E2S3I19-001 | Viewer 访问 Viewer 接口 | role=Viewer, minRole=Viewer | 通过，调用 Handler | Unit |
| TC-E2S3I19-002 | Viewer 访问 Operator 接口 | role=Viewer, minRole=Operator | 403，required_role=Operator | Unit |
| TC-E2S3I19-003 | Operator 访问 Admin 接口 | role=Operator, minRole=Admin | 403，required_role=Admin | Unit |
| TC-E2S3I19-004 | Admin 访问 Operator 接口 | role=Admin, minRole=Operator | 通过（Admin ⊇ Operator） | Unit |
| TC-E2S3I19-005 | context 中无 UserContext（未经 AuthMiddleware） | 无注入 | panic 或返回 500（防御性处理） | Unit |
| TC-E2S3I19-006 | Admin 访问所有级别接口 | role=Admin, minRole=各级 | 全部通过 | Unit |

---

### Story E2-S4：用户管理接口

---

#### Issue E2-S4-I20：用户管理 API（Admin）

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E2-S3-I18, E2-S3-I19, E2-S1-I10  
- **测试类型**：Integration

**功能描述**：
- `GET /api/v1/users/me` — 返回自身信息（任意登录用户）
- `PUT /api/v1/users/me/password` — 修改自身密码，需验证旧密码
- `GET /api/v1/users` — 用户列表，支持分页（Admin）
- `PUT /api/v1/users/{user_id}/role` — 修改角色（Admin），禁止修改自身
- `DELETE /api/v1/users/{user_id}` — 删除用户（Admin），禁止删除自身，禁止删除最后一个 Admin
- `POST /api/v1/users/{user_id}/disable` — 禁用/启用（Admin），禁止禁用自身

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E2S4I20-001 | GET /me 返回正确用户信息 | 200，不含 password_hash | Integration |
| TC-E2S4I20-002 | 修改密码，旧密码正确 | 200，密码已更新，旧密码无法再登录 | Integration |
| TC-E2S4I20-003 | 修改密码，旧密码错误 | 400，code=INVALID_PASSWORD | Integration |
| TC-E2S4I20-004 | 修改密码，新密码强度不足 | 400，code=PASSWORD_WEAK | Integration |
| TC-E2S4I20-005 | Viewer 访问 GET /users | 403，code=PERMISSION_DENIED | Integration |
| TC-E2S4I20-006 | Admin 修改他人角色 | 200，角色已变更 | Integration |
| TC-E2S4I20-007 | Admin 修改自身角色 | 400，code=CANNOT_MODIFY_SELF | Integration |
| TC-E2S4I20-008 | Admin 删除最后一个 Admin | 400，code=LAST_ADMIN_PROTECTED | Integration |
| TC-E2S4I20-009 | Admin 删除自身 | 400，code=CANNOT_DELETE_SELF | Integration |
| TC-E2S4I20-010 | Admin 禁用自身 | 400，code=CANNOT_DISABLE_SELF | Integration |
| TC-E2S4I20-011 | 禁用用户后该用户无法登录 | 403，code=ACCOUNT_DISABLED | Integration |
| TC-E2S4I20-012 | 分页查询 limit=5 offset=0 | total 正确，返回 ≤5 条 | Integration |
| TC-E2S4I20-013 | user_id 不存在 | 404，code=NOT_FOUND | Integration |
| TC-E2S4I20-014 | 响应体不含 password_hash | 任意用户接口 | JSON 中无 password_hash 字段 | Integration |

---

## Epic E3 — Gateway 生命周期管理

> **目标**：实现 Gateway 启停、状态查询、日志查看接口，封装 systemctl --user 命令。

---

### Story E3-S1：systemctl 命令封装

---

#### Issue E3-S1-I21：SystemctlService — user service 封装

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E1-S1-I1, E8-S1-I66（任务引擎，可并行开发后集成）  
- **测试类型**：Unit（Mock）+ Integration

**功能描述**：
- 实现 `SystemctlService` 接口：
  - `Start(service string) error`
  - `Stop(service string) error`
  - `Restart(service string) error`
  - `Status(service string) (*ServiceStatus, error)`
- 所有命令均为 `systemctl --user {action} {service}`
- `ServiceStatus`：`ActiveState`, `SubState`, `MainPID`, `ExecStart`, `FragmentPath`, `ActiveEnterTimestamp`
- 命令执行超时：30s
- 提供 Mock 实现用于单元测试

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E3S1I21-001 | Status 解析 active/running | mock 返回 running 输出 | 解析正确 | Unit |
| TC-E3S1I21-002 | Status 解析 inactive/dead | mock 返回 inactive 输出 | 解析正确 | Unit |
| TC-E3S1I21-003 | Start 命令超时 | mock 命令 hang 超 30s | 返回 `ErrCommandTimeout` | Unit |
| TC-E3S1I21-004 | systemctl 不存在 | 系统无 systemctl | 返回明确错误，不 panic | Unit |
| TC-E3S1I21-005 | service 名称包含特殊字符 | `service = "../evil"` | 拒绝执行，返回 `ErrInvalidServiceName` | Unit |
| TC-E3S1I21-006 | 真实 Status 集成测试 | 系统有 openclaw-gateway.service | 返回真实状态 | Integration |

---

#### Issue E3-S1-I22：Gateway 深度状态查询

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E3-S1-I21  
- **测试类型**：Unit

**功能描述**：
- 聚合查询：`systemctl --user status` + `openclaw gateway status --deep`（子进程调用）
- 解析 openclaw 输出，提取：绑定地址、端口、日志路径、Node 路径
- 检测 Node 路径是否包含 `.nvm`，返回 `NVMWarning bool` 字段
- 并发执行两个命令（goroutine + timeout），合并结果

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E3S1I22-001 | Node 路径包含 .nvm | Mock 输出含 nvm 路径 | NVMWarning=true | Unit |
| TC-E3S1I22-002 | Node 路径系统 node | Mock 输出含 /usr/bin/node | NVMWarning=false | Unit |
| TC-E3S1I22-003 | openclaw 命令超时 | Mock 超时 | 返回超时错误，systemctl 部分结果仍返回 | Unit |
| TC-E3S1I22-004 | 绑定地址解析正确 | Mock 含 bind=127.0.0.1:18789 | BindAddr=127.0.0.1, Port=18789 | Unit |

---

#### Issue E3-S1-I23：Gateway API 接口

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E3-S1-I22, E2-S3-I19, E8-S1-I66  
- **测试类型**：Integration

**功能描述**：
- `GET /api/v1/gateway/status`（Viewer）：调用深度状态查询，返回结构化 JSON
- `POST /api/v1/gateway/start`（Operator）：创建 `gateway.start` 任务
- `POST /api/v1/gateway/stop`（Operator）：创建 `gateway.stop` 任务
- `POST /api/v1/gateway/restart`（Operator）：创建 `gateway.restart` 任务
- 启停任务互斥：运行中的 gateway 任务存在时返回 `409 {"code":"TASK_CONFLICT","running_task_id":"..."}`

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E3S1I23-001 | Viewer 获取 status | 200，含状态字段 | Integration |
| TC-E3S1I23-002 | Viewer 尝试 start | 403，code=PERMISSION_DENIED | Integration |
| TC-E3S1I23-003 | Operator start 创建任务 | 202，含 task_id | Integration |
| TC-E3S1I23-004 | 并发两次 start | 第二次 409，code=TASK_CONFLICT，含 running_task_id | Integration |
| TC-E3S1I23-005 | 无 Token 访问 | 401，code=AUTH_REQUIRED | Integration |
| TC-E3S1I23-006 | status 含 NVMWarning 字段 | NVMWarning 字段存在 | Integration |

---

### Story E3-S2：日志查看

---

#### Issue E3-S2-I24：日志文件读取 API

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E1-S2-I6, E2-S3-I19  
- **测试类型**：Unit + Integration

**功能描述**：
- `GET /api/v1/gateway/logs?lines=200&source=file|journald`（Viewer）
- `source=file`：读取 `/tmp/openclaw/openclaw-{DATE}.log`，取最后 N 行（先找今日日志，不存在则找最近日期）
- `source=journald`：执行 `journalctl --user -u openclaw-gateway.service -n {lines} --no-pager --output=json-seq`
- 路径白名单校验：日志文件必须在 `/tmp/openclaw/` 下

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E3S2I24-001 | 读取今日日志文件 | 返回最后 200 行 | Integration |
| TC-E3S2I24-002 | 日志文件不存在 | 返回空数组，不报错 | Integration |
| TC-E3S2I24-003 | lines 参数超过 1000 | 截断至 1000 | Unit |
| TC-E3S2I24-004 | lines=0 | 返回空数组 | Unit |
| TC-E3S2I24-005 | lines 为负数 | 400，code=VALIDATION_ERROR | Unit |
| TC-E3S2I24-006 | source=journald 调用 journalctl | 返回 journald 日志行 | Integration |
| TC-E3S2I24-007 | 路径穿越尝试 | 伪造 source 路径参数 | 400，路径校验拒绝 | Unit |

---

### Story E3-S3：Doctor 功能

---

#### Issue E3-S3-I25：Doctor Run/Repair 任务接口

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E8-S1-I66, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `POST /api/v1/doctor/run`（Operator）：创建 `doctor.run` 任务，执行 `openclaw doctor`
- `POST /api/v1/doctor/repair`（Operator）：创建 `doctor.repair` 任务，执行 `openclaw doctor --repair`
- 解析 doctor 输出，识别 nvm Node 路径风险，输出结构化建议

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E3S3I25-001 | run 创建任务成功 | 202，task_id | Integration |
| TC-E3S3I25-002 | repair 需要 Operator | Viewer 调用 → 403 | Integration |
| TC-E3S3I25-003 | 任务超时 5min | 超时后任务状态=FAILED，exit_code=-1 | Integration |
| TC-E3S3I25-004 | doctor 输出解析 nvm 风险 | 输出含 nvm 时 NVMDetected=true | Unit |

---

## Epic E4 — 配置管理与版本历史

> **目标**：实现 openclaw.json 与 Agent IDENTITY.md 的可视化编辑、原子写入、Revision 历史管理。

---

### Story E4-S1：Revision 系统

---

#### Issue E4-S1-I26：Revision Repository

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E1-S1-I3  
- **测试类型**：Integration

**功能描述**：
- `RevisionRepository`：`Save(rev Revision) error`，`List(targetType, targetID string, limit int) ([]*Revision, error)`，`FindByID(revID string) (*Revision, error)`
- 保留最近 50 条（每次保存后裁剪超出部分，按 created_at 降序）
- `Revision` 含：`RevisionID`, `TargetType(openclaw_json|agent_identity)`, `TargetID`, `Content`, `SHA256`, `CreatedAt`, `CreatedBy`

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E4S1I26-001 | Save 后可通过 FindByID 查到 | 内容正确 | Integration |
| TC-E4S1I26-002 | List 按时间倒序 | 最新的排第一 | Integration |
| TC-E4S1I26-003 | 超过 50 条时自动裁剪 | 插入第 51 条后，List 返回最多 50 条 | Integration |
| TC-E4S1I26-004 | SHA256 计算正确 | 手动计算与存储值一致 | Unit |
| TC-E4S1I26-005 | TargetType 不同时互不影响 | openclaw_json 的 Revision 不出现在 agent_identity 列表中 | Integration |

---

#### Issue E4-S1-I27：openclaw.json 读写 API

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E4-S1-I26, E1-S3-I9, E1-S2-I6, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `GET /api/v1/config/openclaw-json`（Viewer）：读取文件，返回内容 + 最后修改时间 + 文件大小
- `PUT /api/v1/config/openclaw-json`（Operator）：
  1. 校验请求体 `{"content":"..."}` 为合法 JSON
  2. 创建 Revision
  3. AtomicWriteFile 写入
  4. 可选：触发 gateway.restart 任务（`restart_gateway: true`）
- `GET /api/v1/config/openclaw-json/revisions`：Revision 列表
- `POST /api/v1/config/openclaw-json/revisions/{rev_id}/restore`（Operator）：还原到指定版本

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E4S1I27-001 | 读取文件成功 | 200，content 为合法 JSON | Integration |
| TC-E4S1I27-002 | 文件不存在时读取 | 200，content=null 或 404 | Integration |
| TC-E4S1I27-003 | 写入合法 JSON | 200，文件内容已更新，Revision 已创建 | Integration |
| TC-E4S1I27-004 | 写入非法 JSON | 400，code=INVALID_JSON | Integration |
| TC-E4S1I27-005 | 写入空内容 | 400，code=VALIDATION_ERROR | Integration |
| TC-E4S1I27-006 | 还原到指定 Revision | 文件内容回滚，新 Revision 记录被创建 | Integration |
| TC-E4S1I27-007 | 还原不存在的 Revision ID | 404 | Integration |
| TC-E4S1I27-008 | Viewer 尝试写入 | 403 | Integration |
| TC-E4S1I27-009 | restart_gateway=true 触发任务 | 写入成功后 task 表出现 gateway.restart 任务 | Integration |
| TC-E4S1I27-010 | 并发写入（两个请求同时 PUT） | 文件最终内容为完整的某一次写入，不出现混合 | Integration |

---

### Story E4-S2：Agent Identity 编辑

---

#### Issue E4-S2-I28：Agent Identity 读写 API

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E4-S1-I26, E1-S3-I9, E5-S1-I32（Agent 路径解析，可部分依赖）  
- **测试类型**：Integration

**功能描述**：
- `GET /api/v1/agents/{id}/identity`（Viewer）：读取 `<workspace>/IDENTITY.md`
- `PUT /api/v1/agents/{id}/identity`（Operator）：写入 + Revision
- `GET /api/v1/agents/{id}/identity/revisions`：历史版本列表
- 路径通过 `AgentRepository.GetWorkspacePath(id)` 解析，严格白名单校验

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E4S2I28-001 | 读取存在的 Identity | 200，markdown 内容 | Integration |
| TC-E4S2I28-002 | Agent ID 不存在 | 404 | Integration |
| TC-E4S2I28-003 | 写入 Identity，Revision 创建 | 200，文件更新 | Integration |
| TC-E4S2I28-004 | Viewer 读取 Identity | 200 | Integration |
| TC-E4S2I28-005 | Viewer 尝试写入 Identity | 403 | Integration |
| TC-E4S2I28-006 | agent_id 含路径穿越字符 | 400，白名单拒绝 | Integration |
| TC-E4S2I28-007 | Identity 内容超过 1MB | 400，code=CONTENT_TOO_LARGE | Integration |

---

## Epic E5 — Agent 管理与 Channel Binding

> **目标**：实现 Agent 列表、新建/删除、Channel Binding 可视化管理。

---

### Story E5-S1：Agent CRUD

---

#### Issue E5-S1-I29：Agent Repository 与路径解析

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E1-S2-I6  
- **测试类型**：Unit + Integration

**功能描述**：
- `AgentRepository`：通过解析 `openclaw agents list --bindings` 输出构建 Agent 列表（非 DB 存储，以 openclaw CLI 为数据源）
- `GetWorkspacePath(agentID string) (string, error)`：从 openclaw 配置或 agent 输出解析 workspace 路径，严格白名单校验
- 缓存机制：60s TTL，避免频繁调用 CLI

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E5S1I29-001 | 解析 CLI 输出，返回 Agent 列表 | 字段正确（id、workspace、bindings 数量） | Unit |
| TC-E5S1I29-002 | GetWorkspacePath 合法 agent | 返回绝对路径，在白名单内 | Unit |
| TC-E5S1I29-003 | GetWorkspacePath 不存在 agent | 返回 ErrNotFound | Unit |
| TC-E5S1I29-004 | agent_id 含路径穿越 | ErrInvalidAgentID | Unit |
| TC-E5S1I29-005 | 缓存 TTL 内不重复调用 CLI | 60s 内第二次调用不触发子进程 | Unit |

---

#### Issue E5-S1-I30：Agent 列表与详情 API

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E5-S1-I29, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `GET /api/v1/agents`（Viewer）：返回所有 Agent，含 Binding 数量统计
- `GET /api/v1/agents/{id}`（Viewer）：返回单个 Agent 详情

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E5S1I30-001 | 列表返回正确数量 | agents 数组长度与 CLI 输出一致 | Integration |
| TC-E5S1I30-002 | 无 Agent 时返回空数组 | `{"agents":[]}` 而非 null | Integration |
| TC-E5S1I30-003 | 无 Token 访问 | 401 | Integration |
| TC-E5S1I30-004 | Viewer 可访问 | 200 | Integration |
| TC-E5S1I30-005 | 不存在的 agent_id | 404 | Integration |

---

#### Issue E5-S1-I31：Agent 新建 API（Admin，任务化）

- **Story Points**：3  
- **优先级**：P2  
- **依赖**：E5-S1-I29, E8-S1-I66, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `POST /api/v1/agents`（Admin）：请求体 `{"agent_id":"sales","workspace":"","identity":"..."}`
- 任务执行：`openclaw agents create {id} [--workspace {path}]` + 可选写 IDENTITY.md
- agent_id 格式校验：字母数字下划线，不超过 32 字符

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E5S1I31-001 | 合法请求创建任务 | 202，task_id | Integration |
| TC-E5S1I31-002 | Operator 尝试创建 | 403 | Integration |
| TC-E5S1I31-003 | agent_id 重复 | 任务 FAILED，stderr 含重复提示 | Integration |
| TC-E5S1I31-004 | agent_id 含空格 | 400，code=VALIDATION_ERROR | Integration |
| TC-E5S1I31-005 | workspace 路径穿越 | 400，白名单拒绝 | Integration |

---

#### Issue E5-S1-I32：Agent 删除 API（Admin，任务化）

- **Story Points**：2  
- **优先级**：P2  
- **依赖**：E5-S1-I29, E8-S1-I66  
- **测试类型**：Integration

**功能描述**：
- `DELETE /api/v1/agents/{id}`（Admin）
- 删除前列出该 Agent 的所有 Binding 规则（响应中包含）
- 任务执行：先 unbind 所有规则，再 delete

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E5S1I32-001 | 删除存在的 Agent | 202，task_id | Integration |
| TC-E5S1I32-002 | 删除不存在的 Agent | 404 | Integration |
| TC-E5S1I32-003 | Operator 尝试删除 | 403 | Integration |
| TC-E5S1I32-004 | 有 Binding 的 Agent 删除 | 任务内先 unbind，再 delete，任务 SUCCEEDED | Integration |

---

### Story E5-S2：Channel Binding 管理

---

#### Issue E5-S2-I33：Binding 列表与批量应用 API

- **Story Points**：4  
- **优先级**：P1  
- **依赖**：E5-S1-I29, E8-S1-I66, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `GET /api/v1/bindings`（Viewer）：返回所有 Binding 规则
- `POST /api/v1/bindings/apply`（Operator）：批量添加/删除
  - 请求体：`{"add":[{"agent_id":"sales","channel":"telegram","account":"default","peer":"@xxx"}],"remove":[...]}`
  - 任务内逐条执行 `openclaw agents bind/unbind`
  - 任一条失败则记录到 stderr_tail，继续执行其余条目（不回滚），任务最终状态为 FAILED

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E5S2I33-001 | 获取 Binding 列表 | 200，数组 | Integration |
| TC-E5S2I33-002 | 批量添加 Binding | 202，task_id，任务 SUCCEEDED | Integration |
| TC-E5S2I33-003 | 批量删除 Binding | 202，task_id，任务 SUCCEEDED | Integration |
| TC-E5S2I33-004 | add 和 remove 同时存在 | 同一任务内先 add 再 remove | Integration |
| TC-E5S2I33-005 | add 中 agent_id 不存在 | 任务 FAILED，错误信息含 agent_id | Integration |
| TC-E5S2I33-006 | Viewer 尝试 apply | 403 | Integration |
| TC-E5S2I33-007 | add 列表为空 | 400，code=VALIDATION_ERROR | Integration |
| TC-E5S2I33-008 | peer 字段为空 | 400，code=VALIDATION_ERROR | Integration |
| TC-E5S2I33-009 | 部分失败场景 | 3 条中 1 条失败，task FAILED，成功 2 条已执行 | Integration |

---

## Epic E6 — Skills 管理

> **目标**：实现 Global/Agent 粒度的 Skills 列表、安装（上传包）、删除。

---

### Story E6-S1：Skills 列表与删除

---

#### Issue E6-S1-I34：Skills 扫描与列表 API

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E1-S2-I6, E5-S1-I29, E2-S3-I19  
- **测试类型**：Unit + Integration

**功能描述**：
- `GET /api/v1/skills?scope=global|agent&agent_id=xxx`（Viewer）
- Global：扫描 `~/.openclaw/skills/` 子目录
- Agent：扫描 `<workspace>/skills/` 子目录
- 每个 skill 返回：name（目录名）、scope、agent_id（agent 时）、size_bytes、has_meta（是否有 skill.json/README）

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E6S1I34-001 | 列出 Global Skills | 200，数组，字段正确 | Integration |
| TC-E6S1I34-002 | 列出 Agent Skills | 200，agent_id 字段正确 | Integration |
| TC-E6S1I34-003 | scope=agent 但无 agent_id | 400，code=VALIDATION_ERROR | Unit |
| TC-E6S1I34-004 | Skills 目录为空 | 返回空数组，不报错 | Integration |
| TC-E6S1I34-005 | Skills 目录不存在 | 返回空数组，不报错 | Integration |
| TC-E6S1I34-006 | 无效 scope 参数 | 400，code=VALIDATION_ERROR | Unit |

---

#### Issue E6-S1-I35：Skills 删除 API（任务化）

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E6-S1-I34, E8-S1-I66, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `DELETE /api/v1/skills/{name}?scope=global|agent&agent_id=xxx`（Operator）
- 任务内：路径校验 → `os.RemoveAll(skillPath)`

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E6S1I35-001 | 删除存在的 Global Skill | 202，任务 SUCCEEDED，目录已删除 | Integration |
| TC-E6S1I35-002 | 删除不存在的 Skill | 404，任务不创建 | Integration |
| TC-E6S1I35-003 | Viewer 尝试删除 | 403 | Integration |
| TC-E6S1I35-004 | skill 名称含路径穿越 | 400，白名单拒绝 | Integration |
| TC-E6S1I35-005 | 删除 Agent Skill | 202，对应 agent workspace/skills/{name} 已删除 | Integration |

---

### Story E6-S2：Skills 安装

---

#### Issue E6-S2-I36：Skills 上传安装 API（任务化）

- **Story Points**：4  
- **优先级**：P1  
- **依赖**：E1-S3-I8, E6-S1-I34, E8-S1-I66, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `POST /api/v1/skills/install`（Operator）：multipart/form-data
  - 字段：`file`（zip 或 tar.gz），`scope`，`agent_id`（可选），`skill_name`（可选，覆盖目录名）
- 上传限制：100MB
- 任务内：`SafeExtract` → 校验目标目录不存在（或覆盖选项）→ 移动到目标位置
- 目标路径：`{base}/skills/{skill_name}/`，严格白名单

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E6S2I36-001 | 上传合法 zip 安装到 Global | 202，任务 SUCCEEDED，skill 目录存在 | Integration |
| TC-E6S2I36-002 | 上传合法 tar.gz | 202，任务 SUCCEEDED | Integration |
| TC-E6S2I36-003 | 上传含 zip-slip 的 zip | 任务 FAILED，无文件泄露到白名单外 | Integration |
| TC-E6S2I36-004 | 文件大小超过 100MB | 400，code=FILE_TOO_LARGE（上传阶段拒绝） | Integration |
| TC-E6S2I36-005 | 不支持的文件格式（.rar） | 400，code=UNSUPPORTED_FORMAT | Integration |
| TC-E6S2I36-006 | scope=agent 无 agent_id | 400，code=VALIDATION_ERROR | Integration |
| TC-E6S2I36-007 | skill 同名已存在 | 任务 FAILED，code=SKILL_EXISTS（或覆盖，视配置） | Integration |
| TC-E6S2I36-008 | Viewer 尝试安装 | 403 | Integration |
| TC-E6S2I36-009 | skill_name 含路径穿越 | 400，白名单拒绝 | Integration |

---

## Epic E7 — 备份与还原

> **目标**：实现完整的备份创建、列表查看、还原（含 dry_run）功能，是高风险操作，Admin 授权保护。

---

### Story E7-S1：备份创建

---

#### Issue E7-S1-I37：备份服务核心逻辑

- **Story Points**：4  
- **优先级**：P1  
- **依赖**：E1-S3-I9, E1-S2-I6, E1-S1-I3  
- **测试类型**：Unit + Integration

**功能描述**：
- `BackupService.Create(scope []string, label string, createdBy string) (backupID string, err error)`
- scope 枚举：`openclaw_json`, `global_skills`, `workspaces`, `user_systemd_unit`, `manager_revisions`
- 流程：生成 backup_id → 按 scope 归档（tar.gz） → 计算 SHA-256 → 写 manifest.json → 写 backups 表
- manifest.json 内容：`{backup_id, label, scope, paths, sha256, created_at, created_by}`

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E7S1I37-001 | 全量备份成功 | backup 目录存在，manifest.json 合法，数据库记录插入 | Integration |
| TC-E7S1I37-002 | 部分 scope 备份 | 只有指定 scope 的文件被打包 | Integration |
| TC-E7S1I37-003 | openclaw.json 不存在时备份 | 该文件 scope 跳过，不报错，manifest 记录缺失文件 | Integration |
| TC-E7S1I37-004 | SHA-256 校验和正确 | 手动计算与 manifest 一致 | Integration |
| TC-E7S1I37-005 | backup_id 唯一 | 连续创建 3 个备份，backup_id 各不相同 | Unit |
| TC-E7S1I37-006 | 无效 scope 枚举值 | 返回 ErrInvalidScope | Unit |

---

#### Issue E7-S1-I38：备份 API 接口

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E7-S1-I37, E8-S1-I66, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `POST /api/v1/backups`（Operator）：创建备份任务
- `GET /api/v1/backups`（Viewer）：列表（分页，按 created_at 降序）
- `GET /api/v1/backups/{backup_id}`（Viewer）：详情（含 manifest 内容）
- `DELETE /api/v1/backups/{backup_id}`（Admin）：删除备份记录+文件
- `GET /api/v1/backups/{backup_id}/download`（Operator）：流式返回 tar.gz 文件

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E7S1I38-001 | Operator 创建备份 | 202，task_id | Integration |
| TC-E7S1I38-002 | Viewer 创建备份 | 403 | Integration |
| TC-E7S1I38-003 | 列表返回正确数量和分页 | 200，total 正确 | Integration |
| TC-E7S1I38-004 | 详情含 manifest | 200，manifest 字段存在 | Integration |
| TC-E7S1I38-005 | Viewer 下载备份 | 403（下载需 Operator） | Integration |
| TC-E7S1I38-006 | Operator 下载备份 | 200，Content-Type: application/gzip，文件完整 | Integration |
| TC-E7S1I38-007 | Admin 删除备份 | 200，记录已删除，文件已删除 | Integration |
| TC-E7S1I38-008 | 删除不存在的备份 | 404 | Integration |

---

### Story E7-S2：备份还原

---

#### Issue E7-S2-I39：还原服务（含 dry_run）

- **Story Points**：5  
- **优先级**：P1  
- **依赖**：E7-S1-I37, E1-S3-I9, E1-S2-I6  
- **测试类型**：Unit + Integration

**功能描述**：
- `BackupService.Restore(backupID string, dryRun bool, restartGateway bool, createdBy string) error`
- 验证步骤：backup_id 存在 → 读 manifest → 校验 SHA-256
- dry_run=true：扫描将被覆盖的文件列表，不执行任何写操作，返回预览报告
- 正式还原：① 自动创建「还原前快照」备份 → ② 解压覆盖各 scope 路径 → ③ 可选 restart gateway

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E7S2I39-001 | dry_run 不修改文件 | 返回 will_overwrite 列表，原文件未变 | Integration |
| TC-E7S2I39-002 | dry_run 报告内容正确 | 将被覆盖的路径与实际路径一致 | Integration |
| TC-E7S2I39-003 | 正式还原成功 | 文件内容回到备份时的状态 | Integration |
| TC-E7S2I39-004 | 还原前自动创建快照 | backups 表出现新记录 label=pre-restore | Integration |
| TC-E7S2I39-005 | backup SHA-256 不匹配 | 返回 ErrBackupCorrupted，拒绝还原 | Integration |
| TC-E7S2I39-006 | backup_id 不存在 | 返回 ErrNotFound | Integration |
| TC-E7S2I39-007 | restart_gateway=true | 还原后触发 gateway.restart 任务 | Integration |
| TC-E7S2I39-008 | dry_run=false 正式还原后文件变更 | 还原文件与备份内容一致（字节级比较） | Integration |

---

#### Issue E7-S2-I40：还原 API

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E7-S2-I39, E8-S1-I66, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `POST /api/v1/backups/{backup_id}/restore`（Admin）
- 请求体：`{"dry_run":true,"restart_gateway":false}`
- dry_run=true：同步返回 will_overwrite 列表（非任务化）
- dry_run=false：任务化执行，返回 task_id

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E7S2I40-001 | dry_run=true 同步返回预览 | 200，will_overwrite 数组 | Integration |
| TC-E7S2I40-002 | dry_run=false 创建任务 | 202，task_id | Integration |
| TC-E7S2I40-003 | Operator 尝试还原 | 403，code=PERMISSION_DENIED，required_role=Admin | Integration |
| TC-E7S2I40-004 | 不存在的 backup_id | 404 | Integration |
| TC-E7S2I40-005 | dry_run 默认为 true（未传参数） | 同步预览，不执行还原 | Integration |

---

## Epic E8 — 任务系统与实时日志

> **目标**：实现统一任务引擎（异步执行、状态机、超时）、SSE/WS 实时日志流，是 E3-E7 所有"任务化"操作的基础。

---

### Story E8-S1：任务引擎

---

#### Issue E8-S1-I41：任务 Repository

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E1-S1-I3  
- **测试类型**：Integration

**功能描述**：
- `TaskRepository`：`Create`, `FindByID`, `UpdateStatus`, `UpdateResult`, `List(filter)`
- filter 支持：status、task_type、created_by、limit、offset
- `UpdateStatus` 同时更新 `started_at`（RUNNING）或 `finished_at`（终态）

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E8S1I41-001 | Create 插入 PENDING 任务 | status=PENDING，task_id 不为空 | Integration |
| TC-E8S1I41-002 | PENDING→RUNNING，started_at 更新 | started_at 不为空 | Integration |
| TC-E8S1I41-003 | RUNNING→SUCCEEDED，finished_at 更新 | finished_at 不为空 | Integration |
| TC-E8S1I41-004 | List 按 status 过滤 | 只返回指定 status 的任务 | Integration |
| TC-E8S1I41-005 | List 按 created_by 过滤 | 只返回指定用户的任务 | Integration |

---

#### Issue E8-S1-I42：任务执行引擎（Worker Pool）

- **Story Points**：5  
- **优先级**：P0  
- **依赖**：E8-S1-I41, E1-S1-I2  
- **测试类型**：Unit + Integration

**功能描述**：
- `TaskEngine`：接收任务，在 goroutine pool（最大并发 3，可配置）中执行
- 每个 task 注册 `Handler func(ctx context.Context, task *Task) error`
- 执行过程：PENDING → RUNNING → 执行 Handler → SUCCEEDED/FAILED
- 超时控制：基于 task_type 超时配置，`context.WithTimeout`
- 互斥控制：gateway 类型任务加全局互斥锁
- 日志收集：实时捕获 stdout/stderr，写入 log_path 文件，同时推送 SSE/WS
- 取消支持：`Cancel(taskID)` 对 PENDING 任务有效，RUNNING 任务不可取消

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E8S1I42-001 | 提交任务，Handler 执行成功 | 任务最终 SUCCEEDED | Integration |
| TC-E8S1I42-002 | Handler 返回 error | 任务 FAILED，stderr_tail 含错误信息 | Integration |
| TC-E8S1I42-003 | 任务超时 | 超时后 FAILED，exit_code=-1 | Unit |
| TC-E8S1I42-004 | 并发超过 maxWorkers=3 | 第 4 个任务在 PENDING 等待，不报错 | Integration |
| TC-E8S1I42-005 | Gateway 互斥：第二个 gateway 任务 | 返回 409，running_task_id 正确 | Integration |
| TC-E8S1I42-006 | 取消 PENDING 任务 | 任务变为 CANCELED，Handler 不执行 | Unit |
| TC-E8S1I42-007 | 取消 RUNNING 任务 | 返回 400（不支持取消运行中任务） | Unit |
| TC-E8S1I42-008 | 日志实时写入 log_path | 执行过程中 log_path 文件有内容 | Integration |

---

### Story E8-S2：实时日志流（SSE / WebSocket）

---

#### Issue E8-S2-I43：SSE 任务日志流

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E8-S1-I42, E2-S3-I18  
- **测试类型**：Integration

**功能描述**：
- `GET /api/v1/tasks/{task_id}/events`（Viewer）：SSE 流
- Token 从 query param `?token=` 或 `Authorization` 头获取（SSE 浏览器不支持自定义头）
- 事件格式：`data: {"seq":1,"ts":"...","stream":"stdout","line":"..."}\n\n`
- 任务结束发送 done 事件：`data: {"type":"done","status":"SUCCEEDED","exit_code":0}\n\n`
- 任务已结束时：直接回放历史日志 + 立即发送 done 事件

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E8S2I43-001 | 订阅运行中任务 | 实时收到日志行，任务完成收到 done | Integration |
| TC-E8S2I43-002 | 订阅已完成任务 | 立即收到历史日志 + done | Integration |
| TC-E8S2I43-003 | 无效 task_id | 404 | Integration |
| TC-E8S2I43-004 | 无 Token | 401 | Integration |
| TC-E8S2I43-005 | 客户端断开连接 | 服务端检测到断开，停止推送（不 panic） | Integration |
| TC-E8S2I43-006 | seq 编号递增不重复 | 收到的 seq 严格递增 | Integration |

---

### Story E8-S3：任务查询 API

---

#### Issue E8-S3-I44：任务列表与详情 API

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E8-S1-I41, E2-S3-I19  
- **测试类型**：Integration

**功能描述**：
- `GET /api/v1/tasks`（Viewer）：列表，支持 `?status=&type=&limit=50&offset=0`
- `GET /api/v1/tasks/{task_id}`（Viewer）：详情（含 stdout_tail/stderr_tail）
- `POST /api/v1/tasks/{task_id}/cancel`（Operator）：取消 PENDING 任务
- Viewer 只能看到自己的任务；Admin 可看所有任务

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E8S3I44-001 | Viewer 只看到自己的任务 | 不含其他用户的任务 | Integration |
| TC-E8S3I44-002 | Admin 看到所有任务 | 含所有用户的任务 | Integration |
| TC-E8S3I44-003 | 按 status 过滤 | 只返回指定 status | Integration |
| TC-E8S3I44-004 | 取消 PENDING 任务（Operator） | 200，状态变 CANCELED | Integration |
| TC-E8S3I44-005 | Viewer 尝试取消任务 | 403 | Integration |
| TC-E8S3I44-006 | 取消不存在的任务 | 404 | Integration |

---

## Epic E9 — 前端框架与通用组件

> **目标**：建立 Vue 3 SPA 框架、路由守卫、认证 Store、权限指令，以及各业务页面。

---

### Story E9-S1：前端基础框架

---

#### Issue E9-S1-I45：Vue 3 项目初始化与路由守卫

- **Story Points**：3  
- **优先级**：P0  
- **依赖**：E1-S2-I4  
- **测试类型**：E2E

**功能描述**：
- Vite + Vue 3 + TypeScript + Element Plus + Pinia + Vue Router 4
- 路由定义：`/login`、`/register`、`/dashboard`、`/gateway`、`/agents`、`/skills`、`/config`、`/backups`、`/tasks`、`/admin/users`
- 路由守卫：未登录重定向 `/login`；`/admin/users` 非 Admin 重定向 `/dashboard`
- Layout：全局 NavBar（含用户名、角色标签、退出按钮）+ Sidebar + Content

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E9S1I45-001 | 未登录访问 /dashboard | 重定向到 /login | E2E |
| TC-E9S1I45-002 | Viewer 访问 /admin/users | 重定向到 /dashboard | E2E |
| TC-E9S1I45-003 | 登录后导航正常 | 各页面可正常访问 | E2E |
| TC-E9S1I45-004 | 退出后清除 Token | 再访问 /dashboard 重定向 /login | E2E |

---

#### Issue E9-S1-I46：认证 Store（Pinia）与 Token 管理

- **Story Points**：3  
- **优先级**：P0  
- **依赖**：E9-S1-I45  
- **测试类型**：Unit（Vitest）

**功能描述**：
- `useAuthStore`：state: `accessToken`, `user{id,username,role}`, `isAuthenticated`
- AccessToken 存内存（不存 localStorage）
- axios 拦截器：自动添加 `Authorization: Bearer {token}`；收到 401 时自动调用 `/auth/refresh`（使用 Cookie 中的 RefreshToken）；refresh 失败则清除 store 跳转 `/login`
- 防并发刷新：多个 401 同时发生时只发一次 refresh 请求（Promise 复用）

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E9S1I46-001 | 登录后 store 有正确 user 信息 | username、role 正确 | Unit |
| TC-E9S1I46-002 | 401 触发自动 refresh | refresh 成功后原请求自动重试 | Unit |
| TC-E9S1I46-003 | refresh 失败跳转 /login | store 清空，路由跳转 | Unit |
| TC-E9S1I46-004 | 并发 401 只发一次 refresh | mock 3 个 401 同时发生，refresh 只调用 1 次 | Unit |

---

#### Issue E9-S1-I47：权限指令 v-permission 与工具函数

- **Story Points**：2  
- **优先级**：P0  
- **依赖**：E9-S1-I46  
- **测试类型**：Unit

**功能描述**：
- Vue 自定义指令 `v-permission="'Operator'"`：角色不足时禁用按钮并添加 tooltip
- `usePermission()` composable：`canAccess(minRole: Role): boolean`
- 全局 403 响应处理：toast 显示 "权限不足：需要 {required_role} 角色"

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E9S1I47-001 | Viewer 角色，v-permission='Operator' 按钮 | 按钮 disabled，hover 有 tooltip | Unit |
| TC-E9S1I47-002 | Operator 角色，v-permission='Operator' | 按钮正常可点击 | Unit |
| TC-E9S1I47-003 | canAccess('Admin') — Operator 用户 | 返回 false | Unit |
| TC-E9S1I47-004 | API 返回 403，toast 显示 required_role | toast 内容含 required_role 字段 | Unit |

---

### Story E9-S2：认证页面

---

#### Issue E9-S2-I48：登录页与注册页

- **Story Points**：3  
- **优先级**：P0  
- **依赖**：E9-S1-I46  
- **测试类型**：E2E

**功能描述**：
- 登录页：用户名+密码表单，登录按钮，错误提示，"前往注册"链接
- 注册页：用户名+密码+确认密码，密码强度条（弱/中/强），注册按钮
- 密码强度条：≤8位=弱，含大写字母=中，含特殊字符=强（纯前端 UI）
- 注册成功跳转登录并 toast "注册成功，请登录"

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E9S2I48-001 | 正常登录 | 跳转 /dashboard | E2E |
| TC-E9S2I48-002 | 密码错误 | 页面显示错误提示，不跳转 | E2E |
| TC-E9S2I48-003 | 两次密码不一致 | 注册按钮禁用或提交时报错 | E2E |
| TC-E9S2I48-004 | 密码强度显示 | 输入不同密码显示对应强度 | E2E |
| TC-E9S2I48-005 | 注册成功 | 跳转 /login，toast 显示 | E2E |

---

### Story E9-S3：业务页面

---

#### Issue E9-S3-I49：Dashboard 页

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E9-S1-I47, E3-S1-I23（后端 API）  
- **测试类型**：E2E

**功能描述**：
- Gateway 状态卡（30s 轮询）、Channels 健康卡、Agents 汇总卡、Doctor 建议卡（NVMWarning 时显示橙色横幅）、最近任务卡
- Operator+ 可见 Gateway 快捷操作按钮；Viewer 按钮 disabled
- NVM 警告横幅全局展示（可关闭，session 内不重复显示）

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E9S3I49-001 | NVMWarning=true 显示橙色横幅 | 横幅可见，含修复按钮 | E2E |
| TC-E9S3I49-002 | Viewer 看不到启停按钮（disabled） | 按钮 disabled 状态 | E2E |
| TC-E9S3I49-003 | 30s 后状态自动刷新 | 状态卡更新（mock 状态变化） | E2E |

---

#### Issue E9-S3-I50：用户管理页（Admin）

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E9-S1-I47, E2-S4-I20（后端）  
- **测试类型**：E2E

**功能描述**：
- 用户列表（含分页）、角色 badge、状态标签
- 操作列：修改角色（下拉）、禁用/启用（开关）、删除（带二次确认）
- 自身行操作全部禁用，tooltip 提示"不可修改自身"
- Viewer/Operator 角色访问此页重定向

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E9S3I50-001 | 自身行操作全 disabled | 无法点击 | E2E |
| TC-E9S3I50-002 | 修改他人角色 | API 调用成功，列表即时刷新 | E2E |
| TC-E9S3I50-003 | 删除用户二次确认 | 弹窗出现，取消则不删除 | E2E |
| TC-E9S3I50-004 | Operator 访问此页 | 重定向 /dashboard | E2E |

---

#### Issue E9-S3-I51：Tasks 页与实时日志面板

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E9-S1-I47, E8-S2-I43（后端 SSE）  
- **测试类型**：E2E

**功能描述**：
- 任务列表（状态色块、类型、创建时间、操作人）
- 点击任务行打开右侧日志面板（SSE 实时流）
- 日志面板：支持自动滚动到底部（开关）、搜索过滤（前端）
- FAILED 任务高亮显示 stderr_tail

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E9S3I51-001 | 点击运行中任务，日志实时追加 | 日志面板实时更新 | E2E |
| TC-E9S3I51-002 | 任务完成后日志面板显示 done | "任务完成" 标记出现 | E2E |
| TC-E9S3I51-003 | 自动滚动关闭时不跳到底部 | 关闭开关后日志不自动滚动 | E2E |

---

## Epic E10 — 集成测试与部署

---

### Story E10-S1：端到端集成测试

---

#### Issue E10-S1-I52：关键业务流程 E2E 测试

- **Story Points**：5  
- **优先级**：P1  
- **依赖**：E1~E9 全部完成  
- **测试类型**：E2E（Playwright）

**功能描述**：
覆盖以下完整业务流程的自动化 E2E 测试：

**测试用例**：

| 用例编号 | 流程名称 | 步骤描述 | 通过标准 |
|----------|----------|----------|----------|
| TC-E10S1I52-001 | 首位用户注册并获得 Admin | 注册 → 登录 → 查看 /me | role=Admin |
| TC-E10S1I52-002 | Admin 邀请 Operator | 注册第二用户 → Admin 修改角色为 Operator | 第二用户登录后角色=Operator |
| TC-E10S1I52-003 | Operator 编辑配置并重启 Gateway | 编辑 JSON → 保存 → 确认重启 → 任务 SUCCEEDED | 文件已更新，Gateway 重启任务完成 |
| TC-E10S1I52-004 | 全量备份后还原 | 备份 → 修改配置 → dry_run 预览 → 还原 → 验证文件 | 文件回到备份时内容 |
| TC-E10S1I52-005 | Viewer 无法执行写操作 | 用 Viewer 账号尝试启停/写配置/删备份 | 全部 403 |
| TC-E10S1I52-006 | Token 过期自动刷新 | 登录后等 access token 过期，发请求 | 自动刷新，用户无感知 |
| TC-E10S1I52-007 | 注销后 Token 失效 | 注销后立即用旧 Token 请求 | 401 |

---

#### Issue E10-S1-I53：安全测试用例

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E1~E9  
- **测试类型**：Security

**测试用例**：

| 用例编号 | 描述 | 攻击向量 | 通过标准 |
|----------|------|----------|----------|
| TC-E10S1I53-001 | 路径穿越攻击（配置写入） | PUT config 的 content 字段含路径穿越 | 400 或文件不在白名单外 |
| TC-E10S1I53-002 | zip-slip 上传攻击 | 上传含 ../ 的 zip | 任务 FAILED，无文件泄露 |
| TC-E10S1I53-003 | 未认证访问所有业务接口 | 无 Token 请求各接口 | 全部 401 |
| TC-E10S1I53-004 | 越权访问（Viewer 调用 Admin 接口） | Viewer Token 调用 DELETE /users | 全部 403 |
| TC-E10S1I53-005 | 使用已注销的 Token | 注销后用旧 Token | 401 |
| TC-E10S1I53-006 | JWT 篡改（修改 role 字段） | 修改 payload 的 role 为 Admin | 401（签名不匹配） |
| TC-E10S1I53-007 | 密码暴力猜测响应时间 | 正确/错误密码响应时间对比 | 时间差 < 50ms（bcrypt 固定耗时防侧信道） |
| TC-E10S1I53-008 | 响应体不含敏感字段 | 任何用户接口响应 | 无 password_hash、jwt_secret、refresh_token 明文 |

---

### Story E10-S2：部署与运维

---

#### Issue E10-S2-I54：systemd user service 配置文件

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E1-S1-I1  
- **测试类型**：Integration

**功能描述**：
- 提供 `openclaw-manager.service` systemd user service 模板
- 配置：`ExecStart`、`Restart=on-failure`、`RestartSec=5`、`Environment=OPENCLAW_JWT_SECRET=`（提示填写）
- 提供安装脚本 `scripts/install.sh`：复制 service 文件、`systemctl --user daemon-reload`、`enable`、`start`

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E10S2I54-001 | install.sh 执行成功 | service 文件已复制，状态 active | Integration |
| TC-E10S2I54-002 | 服务 crash 后自动重启 | kill 进程后 5s 内自动重启 | Integration |
| TC-E10S2I54-003 | 缺少 jwt_secret 时启动失败 | 启动时配置校验失败，service inactive | Integration |

---

#### Issue E10-S2-I55：README 与开发者文档

- **Story Points**：2  
- **优先级**：P2  
- **依赖**：所有 Epic 完成  
- **测试类型**：无

**功能描述**：
- README.md：系统架构图、快速部署步骤、config.toml 配置说明
- API 文档：使用 swaggo/swag 生成 OpenAPI 3.0 规范，部署到 `/api/v1/docs`
- CHANGELOG.md：记录 v1.0 和 v1.1 变更

---

## Issue 依赖关系总图

```
E1-S1-I1 (Go初始化)
    ├── E1-S1-I2 (config.toml)
    │       └── E1-S1-I3 (SQLite初始化)
    │               ├── E2-S1-I10 (User Repository)
    │               ├── E8-S1-I41 (Task Repository)
    │               └── E7-S1-I37 (Backup Service)
    ├── E1-S2-I4 (HTTP服务器)
    │       └── E1-S2-I5 (统一错误响应)
    └── E1-S2-I6 (路径白名单)
            ├── E1-S3-I8 (zip-slip防护)
            └── E1-S3-I9 (原子写入)

E2-S1-I10 → E2-S1-I11 (密码服务) → E2-S1-I12 (注册API)
E2-S2-I13 (JWT) + E2-S2-I14 (Token存储) → E2-S2-I15 (登录API)
                                          → E2-S2-I16 (刷新API)
E2-S3-I18 (AuthMiddleware) → E2-S3-I19 (RequireRole)
    │                              └─ 所有业务API的前置依赖
    └── E2-S2-I17 (注销API)

E8-S1-I41 → E8-S1-I42 (任务引擎)
    ├── E3-S1-I23 (Gateway API)
    ├── E4-S1-I27 (Config API)
    ├── E5-S1-I31 (Agent新建)
    ├── E6-S2-I36 (Skills安装)
    ├── E7-S1-I38 (备份API)
    └── E7-S2-I40 (还原API)

E8-S1-I42 → E8-S2-I43 (SSE日志流)
          → E8-S3-I44 (任务查询API)
```

### 关键路径（Critical Path）

```
E1-S1-I1 → E1-S1-I2 → E1-S1-I3 → E2-S1-I10 → E2-S2-I13
         → E1-S2-I4 → E2-S3-I18 → E2-S3-I19
         → E8-S1-I41 → E8-S1-I42 → [所有业务功能]
```

**关键路径总时长估算**：E1-S1-I1(2) → E1-S1-I2(2) → E1-S1-I3(3) → 同步执行认证+任务引擎 ≈ **12 Story Points ≈ 6 工作日**，后续业务功能可并行。

---

## Sprint 规划建议

### Sprint 1（2 周）— 基础骨架 + 认证

**目标**：跑通注册→登录→权限检查完整链路

| Issue ID | 名称 | SP |
|----------|------|----|
| E1-S1-I1 | Go 模块初始化 | 2 |
| E1-S1-I2 | config.toml 加载 | 2 |
| E1-S1-I3 | SQLite 初始化 | 3 |
| E1-S2-I4 | HTTP 服务器 | 2 |
| E1-S2-I5 | 统一错误响应 | 2 |
| E1-S2-I6 | 路径白名单 | 3 |
| E1-S3-I7 | 测试框架 | 1 |
| E1-S3-I8 | zip-slip 防护 | 2 |
| E1-S3-I9 | 原子写入 | 1 |
| E2-S1-I10 | User Repository | 2 |
| E2-S1-I11 | 密码哈希服务 | 1 |
| E2-S1-I12 | 注册 API | 3 |
| E9-S1-I45 | Vue 项目初始化 | 3 |
| E9-S2-I48 | 登录注册页 | 3 |
| **合计** | | **30 SP** |

---

### Sprint 2（2 周）— 认证完整 + 任务引擎 + Gateway

**目标**：认证全链路（JWT/刷新/注销/权限）+ 任务引擎 + Gateway 基础

| Issue ID | 名称 | SP |
|----------|------|----|
| E2-S2-I13 | JWT 服务 | 3 |
| E2-S2-I14 | RefreshToken 存储 | 2 |
| E2-S2-I15 | 登录 API | 3 |
| E2-S2-I16 | Token 刷新 | 2 |
| E2-S2-I17 | 注销 API | 1 |
| E2-S3-I18 | AuthMiddleware | 2 |
| E2-S3-I19 | RequireRole | 1 |
| E2-S4-I20 | 用户管理 API | 3 |
| E8-S1-I41 | Task Repository | 2 |
| E8-S1-I42 | 任务执行引擎 | 5 |
| E8-S2-I43 | SSE 日志流 | 3 |
| E8-S3-I44 | 任务查询 API | 2 |
| E9-S1-I46 | 认证 Store | 3 |
| E9-S1-I47 | 权限指令 | 2 |
| **合计** | | **34 SP** |

---

### Sprint 3（2 周）— Gateway + 配置管理

**目标**：Gateway 完整管理 + 配置文件编辑 + Revision 历史

| Issue ID | 名称 | SP |
|----------|------|----|
| E3-S1-I21 | SystemctlService | 3 |
| E3-S1-I22 | 深度状态查询 | 2 |
| E3-S1-I23 | Gateway API | 3 |
| E3-S2-I24 | 日志读取 API | 2 |
| E3-S3-I25 | Doctor API | 2 |
| E4-S1-I26 | Revision Repository | 2 |
| E4-S1-I27 | openclaw.json API | 3 |
| E4-S2-I28 | Identity API | 2 |
| E9-S3-I49 | Dashboard 页 | 3 |
| E9-S3-I50 | 用户管理页 | 3 |
| E9-S3-I51 | Tasks 日志页 | 3 |
| **合计** | | **28 SP** |

---

### Sprint 4（2 周）— Agent + Skills

**目标**：Agent 管理、Binding 配置、Skills 安装删除

| Issue ID | 名称 | SP |
|----------|------|----|
| E5-S1-I29 | Agent Repository | 2 |
| E5-S1-I30 | Agent 列表 API | 2 |
| E5-S1-I31 | Agent 新建 API | 3 |
| E5-S1-I32 | Agent 删除 API | 2 |
| E5-S2-I33 | Binding API | 4 |
| E6-S1-I34 | Skills 列表 API | 2 |
| E6-S1-I35 | Skills 删除 API | 2 |
| E6-S2-I36 | Skills 安装 API | 4 |
| **合计** | | **21 SP** |

---

### Sprint 5（2 周）— 备份还原 + 前端完善

**目标**：备份还原完整链路，前端各业务页面

| Issue ID | 名称 | SP |
|----------|------|----|
| E7-S1-I37 | 备份核心逻辑 | 4 |
| E7-S1-I38 | 备份 API | 2 |
| E7-S2-I39 | 还原服务 | 5 |
| E7-S2-I40 | 还原 API | 2 |
| 前端页面 | Config/Agents/Skills/Backups 页 | 12 |
| **合计** | | **25 SP** |

---

### Sprint 6（2 周）— 集成测试 + 部署加固

**目标**：E2E 测试、安全测试、部署文档，达到上线标准

| Issue ID | 名称 | SP |
|----------|------|----|
| E10-S1-I52 | E2E 业务流程测试 | 5 |
| E10-S1-I53 | 安全测试 | 3 |
| E10-S2-I54 | systemd service 配置 | 2 |
| E10-S2-I55 | README 与 API 文档 | 2 |
| 缓冲 | Bug 修复与性能优化 | 8 |
| **合计** | | **20 SP** |

---

## 附录：测试覆盖率要求

| 模块 | 单元测试覆盖率要求 | 备注 |
|------|--------------------|------|
| internal/auth | ≥ 95% | 安全核心，要求最高 |
| internal/user | ≥ 90% | 含权限逻辑 |
| internal/storage/pathvalidator | ≥ 100% | 安全关键路径 |
| internal/storage/extract | ≥ 100% | 安全关键路径 |
| internal/storage/atomic | ≥ 95% | 数据完整性 |
| internal/task | ≥ 85% | 并发逻辑 |
| internal/gateway | ≥ 80% | 依赖外部命令 |
| internal/config | ≥ 90% | |
| frontend/stores | ≥ 85% | |
| **整体** | **≥ 80%** | |

---

## 附录：命名与编码规范

- **Go**：遵循 `Effective Go`；接口名以 `er` 结尾（`UserRepository`）；错误变量以 `Err` 开头（`ErrNotFound`）
- **API**：资源名用复数小写（`/users`，`/agents`）；任务化接口返回 202；同步接口返回 200/201
- **Git Commit**：`[E2-S1-I12] feat: implement user registration API`（含 Issue ID 前缀）
- **测试文件**：与被测文件同包，`xxx_test.go`；集成测试用 `//go:build integration` tag
- **数据库迁移**：文件名 `000001_init.sql`，按序执行，只增不改

---

## Epic E11 — 多 Agent Workspace 可视化与备份增强

> **目标**：补齐多 Agent 场景下的 Workspace 可视化与备份能力，确保 `workspaces` scope 能覆盖主 Agent 与非主 Agent 工作区。

### Story E11-S1：Agent Workspace 可视化补强

#### Issue E11-S1-I56：Agent 列表 Workspace 位置展示与兜底

- **Story Points**：2  
- **优先级**：P1  
- **依赖**：E5-S1-I29, E9-S3-I52  
- **测试类型**：Unit + Integration

**功能描述**：
- Agent 列表页明确展示 `workspace_path` 列
- 当 `openclaw agents list --bindings` 未返回 workspace 时，后端从 `openclaw.json` 兜底解析
- 兜底规则：
  - `main` 使用 `agents.defaults.workspace`（缺失时回退 `~/.openclaw/workspace`）
  - 其他 Agent 缺省按 `workspace-<agentId>` 推导

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E11S1I56-001 | CLI workspace 为空时兜底 | 列表返回每个 agent 的 workspace_path | Unit |
| TC-E11S1I56-002 | main 缺省 workspace 解析 | 返回 defaults.workspace | Unit |
| TC-E11S1I56-003 | 非 main 缺省 workspace 解析 | 返回 workspace-<agentId> | Unit |
| TC-E11S1I56-004 | 前端列表展示 workspace 列 | 页面可见 workspace 位置文本 | Integration |

### Story E11-S2：多 Agent Workspace 备份

#### Issue E11-S2-I57：`workspaces` scope 按 openclaw.json 扩展多目录归档

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E7-S1-I37, E4-S1-I27  
- **测试类型**：Unit + Integration

**功能描述**：
- `backup.create` 在 `scope=["workspaces"]` 时，从 `openclaw.json` 解析全部 Agent workspace
- 支持显式配置与缺省推导混合场景
- 去重后归档，保持 Manifest 中 `paths` 与实际归档范围一致
- 读取 `openclaw.json` 失败时回退旧行为（仅主 workspace）

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E11S2I57-001 | 多 Agent workspace 解析 | Manifest paths 包含全部工作区 | Unit |
| TC-E11S2I57-002 | 显式 workspace + 缺省推导混合 | 路径完整且无重复 | Unit |
| TC-E11S2I57-003 | openclaw.json 缺失/非法 | 回退主 workspace 并可成功备份 | Unit |
| TC-E11S2I57-004 | 创建备份接口联调 | 返回 202，任务状态可追踪 | Integration |

## Epic E12 — Agent Workspace 迁移

> **目标**：提供可视化 Workspace 迁移流程，支持目录内容搬迁、配置同步更新与 Gateway 自动重启。

### Story E12-S1：迁移入口与页面

#### Issue E12-S1-I58：Agent 列表迁移入口与迁移页

- **Story Points**：3  
- **优先级**：P1  
- **依赖**：E11-S1-I56, E9-S3-I52  
- **测试类型**：Integration

**功能描述**：
- 在 Agent 列表每条记录后增加“迁移”按钮
- 点击跳转 Workspace 迁移页面
- 页面展示旧目录地址（只读），用户填写新目录地址后提交

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E12S1I58-001 | 列表页显示迁移按钮 | 每个 Agent 行可见“迁移” | Integration |
| TC-E12S1I58-002 | 跳转迁移页 | 路由跳转到 `/agents/{id}/workspace-migrate` | Integration |
| TC-E12S1I58-003 | 迁移页加载旧目录 | 页面展示旧 workspace 路径 | Integration |

### Story E12-S2：迁移后端执行链路

#### Issue E12-S2-I59：Workspace 迁移 API（移动目录 + 更新配置 + 重启）

- **Story Points**：5  
- **优先级**：P1  
- **依赖**：E5-S1-I29, E4-S1-I27, E3-S1-I23  
- **测试类型**：Unit + Integration

**功能描述**：
- 新增 `POST /api/v1/agents/{id}/workspace/migrate`
- 执行顺序：
  1. 读取旧 workspace
  2. 将旧目录下全部文件/目录迁移到新目录
  3. 更新 `openclaw.json` 对应 Agent workspace 配置
  4. 重启 `openclaw gateway`
- 返回迁移结果（旧路径、新路径、重启状态）

**测试用例**：

| 用例编号 | 描述 | 预期输出 | 测试类型 |
|----------|------|----------|----------|
| TC-E12S2I59-001 | 成功迁移 | 200，文件已迁移，新路径生效 | Integration |
| TC-E12S2I59-002 | 目标目录文件冲突 | 409，迁移中止 | Unit |
| TC-E12S2I59-003 | agent 不存在 | 404 | Unit |
| TC-E12S2I59-004 | 迁移完成后配置更新 | openclaw.json 对应 workspace 更新 | Unit |
| TC-E12S2I59-005 | 迁移完成后网关重启 | 调用 gateway restart 成功 | Unit |

*文档结束 — 总计 119 个 Issue，覆盖 12 个 Epic，建议 8 个 Sprint（约 14 周）完成 MVP+。*

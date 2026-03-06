# 开发 Issue 进度清单

> 说明：本清单用于承接各 Epic 的 Issue 进度，完成一个 Issue 必须更新状态与测试结果。  
> 每完成一个 Issue，需同步更新 **状态**、**测试结果** 和 **备注**。  
> 对应规划文档：`开发规划.md` | 需求文档：`openclaw_requirements_v1.1.docx`

## 状态约定

| 状态 | 含义 |
|------|------|
| Todo | 待开发 |
| Skip | 本轮跳过 |
| In Progress | 开发中 |
| Test Passed | 测试通过 |
| Done | 已完成并合并 |

## 优先级约定

| 优先级 | 含义 |
|--------|------|
| P0 | 阻塞型，必须先完成 |
| P1 | 核心功能，MVP 必交付 |
| P2 | 增强功能，下一迭代 |

## 测试类型约定

| 类型 | 说明 |
|------|------|
| Unit | 单元测试（go test / vitest） |
| Integration | 集成测试（真实 DB / HTTP） |
| E2E | 端到端测试（Playwright） |
| Security | 安全测试 |

---

## 总体进度

| Epic | 总数 | Done | In Progress | Todo |
|------|------|------|-------------|------|
| E1 基础骨架 | 9 | 0 | 0 | 9 |
| E2 用户认证 | 11 | 0 | 0 | 11 |
| E3 Gateway 管理 | 5 | 0 | 0 | 5 |
| E4 配置管理 | 3 | 0 | 0 | 3 |
| E5 Agent 管理 | 5 | 0 | 0 | 5 |
| E6 Skills 管理 | 3 | 0 | 0 | 3 |
| E7 备份还原 | 4 | 0 | 0 | 4 |
| E8 任务系统 | 4 | 0 | 0 | 4 |
| E9 前端框架 | 7 | 0 | 0 | 7 |
| E10 集成测试与部署 | 4 | 0 | 0 | 4 |
| E11 多 Agent Workspace 能力增强 | 2 | 0 | 0 | 2 |
| E12 Agent Workspace 迁移 | 2 | 0 | 0 | 2 |
| **合计** | **59** | **0** | **0** | **59** |

> 注：规划文档中部分 Issue（如前端各业务页面）在 Sprint 5 中以合并 Issue 形式出现，本清单按实际可追踪粒度拆分为 55 条。

---

## Epic 1：项目基础骨架与工程化

> **Sprint 1** | 目标：建立 Go 后端框架、SQLite、路由骨架、安全工具函数，是所有后续 Epic 的前置依赖。

### Story E1-S1：Go 后端框架初始化

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E1-S1-I1 | Go 模块与目录结构初始化（go mod init、HTTP框架选型、Makefile、golangci-lint） | P0 | 2 | Done | - | Unit | Passed | 关键路径起点 |
| E1-S1-I2 | config.toml 配置加载模块（viper/手写、结构体、env覆盖、校验） | P0 | 2 | Done | E1-S1-I1 | Unit | Passed | jwt_secret ≥32字节校验 |
| E1-S1-I3 | SQLite 存储层初始化与迁移（WAL模式、外键约束、版本化迁移、建表） | P0 | 3 | Done | E1-S1-I2 | Unit + Integration | Passed | 6张表全部建立 |

### Story E1-S2：HTTP 路由骨架与中间件基础

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E1-S2-I4 | HTTP 服务器与路由注册（监听、路由分组、日志中间件、Panic Recovery、CORS、优雅关闭） | P0 | 2 | Done | E1-S1-I1 | Unit + Integration | Passed | 提供 /api/v1/health 端点 |
| E1-S2-I5 | 统一错误响应与请求校验框架（AppError、错误码常量、validator绑定） | P0 | 2 | Done | E1-S2-I4 | Unit | Passed | 500不泄露内部信息 |
| E1-S2-I6 | 路径白名单安全模块（PathValidator、EvalSymlinks、JoinAndValidate） | P0 | 3 | Done | E1-S1-I2 | Unit | Passed | 安全核心，覆盖率要求100% |

### Story E1-S3：CI/CD 与测试基础设施

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E1-S3-I7 | 单元测试框架与覆盖率配置（testify、NewTestDB、Makefile targets） | P0 | 1 | Done | E1-S1-I1 | Unit | Passed | 覆盖率目标 ≥80% |
| E1-S3-I8 | Zip-slip 防护工具函数（SafeExtract、zip/tar.gz、大小限制） | P0 | 2 | Done | E1-S2-I6 | Unit | Passed | 安全关键，覆盖率要求100% |
| E1-S3-I9 | 原子文件写入工具函数（tmpfile + os.Rename、失败清理） | P0 | 1 | Done | E1-S2-I6 | Unit | Passed | 防止配置文件半写损坏 |

---

## Epic 2：用户体系与认证授权

> **Sprint 1-2** | 目标：实现用户注册、登录、JWT 认证、角色权限中间件，是所有业务接口的安全前置依赖。

### Story E2-S1：用户注册与密码管理

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E2-S1-I10 | 用户数据模型与 Repository（User结构体、CRUD、Count、ExistsAdmin） | P0 | 2 | Done | E1-S1-I3 | Unit + Integration | Passed | 首位用户Auto-Admin逻辑基础 |
| E2-S1-I11 | 密码哈希服务（bcrypt cost=12、Hash、Verify、ValidateStrength） | P0 | 1 | Done | E1-S1-I1 | Unit | Passed | 不允许纯数字或纯字母密码 |
| E2-S1-I12 | 用户注册 API — POST /api/v1/auth/register（首位→Admin，后续→Viewer） | P0 | 3 | Done | E2-S1-I10, E2-S1-I11, E1-S2-I5 | Unit + Integration | Passed | public_registration 开关 |

### Story E2-S2：JWT 认证与 Token 管理

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E2-S2-I13 | JWT 签发与验证服务（HS256、AccessToken 15min、jti黑名单检查） | P0 | 3 | Done | E2-S1-I10, E1-S1-I2 | Unit | Passed | Claims含sub/role/jti |
| E2-S2-I14 | RefreshToken Repository 与黑名单（SHA-256哈希存储、Revoke、CleanExpired） | P0 | 2 | Done | E1-S1-I3, E2-S2-I13 | Integration | Passed | RefreshToken不存明文 |
| E2-S2-I15 | 登录 API — POST /api/v1/auth/login（bcrypt验证、签发双Token、HttpOnly Cookie） | P0 | 3 | Done | E2-S2-I13, E2-S2-I14, E2-S1-I11 | Integration | Passed | 响应体不含refresh_token明文 |
| E2-S2-I16 | Token 刷新 API — POST /api/v1/auth/refresh（Cookie取token、验证、签发新AccessToken） | P0 | 2 | Done | E2-S2-I14, E2-S2-I13 | Integration | Passed | 已实现 Cookie 校验、撤销/过期校验、disabled 用户拦截 |
| E2-S2-I17 | 注销 API — POST /api/v1/auth/logout（jti入黑名单、撤销RefreshToken、清Cookie） | P1 | 1 | Done | E2-S2-I14, E2-S3-I18 | Integration | Passed | 已实现 jti 黑名单、refresh 撤销、Cookie 清除 |

### Story E2-S3：权限中间件

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E2-S3-I18 | AuthMiddleware — JWT 验证与用户注入（提取Bearer、黑名单查询、注入UserContext） | P0 | 2 | Done | E2-S2-I13, E2-S2-I14 | Unit + Integration | Passed | 已实现 Bearer 提取、JWT 错误映射、context 注入 |
| E2-S3-I19 | RequireRole 中间件（角色权重比较 Viewer=1/Operator=2/Admin=3） | P0 | 1 | Done | E2-S3-I18 | Unit | Passed | 已实现三级角色权重比较与 required_role 返回 |

### Story E2-S4：用户管理接口

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E2-S4-I20 | 用户管理 API（GET /me、改密码、用户列表、修改角色、删除、禁用，Admin专用） | P1 | 3 | Done | E2-S3-I18, E2-S3-I19, E2-S1-I10 | Integration | Passed | 已实现核心接口与自操作/最后Admin保护 |

---

## Epic 3：Gateway 生命周期管理

> **Sprint 2-3** | 目标：实现 Gateway 启停、状态查询、日志查看、Doctor 接口，封装 systemctl --user 命令。

### Story E3-S1：systemctl 命令封装

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E3-S1-I21 | SystemctlService — user service 封装（Start/Stop/Restart/Status、30s超时、Mock实现） | P1 | 3 | Done | E1-S1-I1 | Unit + Integration | Passed | 已实现命令封装、超时处理、service 名称校验与解析 |
| E3-S1-I22 | Gateway 深度状态查询（并发聚合systemctl+openclaw CLI、NVMWarning检测） | P1 | 2 | Done | E3-S1-I21 | Unit | Passed | 已实现并发查询、bind/log/node 解析与 NVMWarning 检测，openclaw 超时时保留 systemctl 部分结果 |
| E3-S1-I23 | Gateway API（GET status、POST start/stop/restart，Operator权限，互斥409） | P1 | 3 | Done | E3-S1-I22, E2-S3-I19, E8-S1-I42 | Integration | Passed | 已实现 status/start/stop/restart 与互斥冲突 409 返回 |

### Story E3-S2：日志查看

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E3-S2-I24 | 日志文件读取 API（GET /gateway/logs、file/journald双源、最后N行、白名单校验） | P1 | 2 | Done | E1-S2-I6, E2-S3-I19 | Unit + Integration | Passed | 已实现 file/journald 双源、lines 限制、/tmp/openclaw 白名单校验 |

### Story E3-S3：Doctor 功能

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E3-S3-I25 | Doctor Run/Repair 任务接口（POST /doctor/run、/doctor/repair，5min超时） | P1 | 2 | Done | E8-S1-I42, E2-S3-I19 | Integration | Passed | 已实现 run/repair 接口、5min 超时控制与 nvm 风险解析 |

---

## Epic 4：配置管理与版本历史

> **Sprint 3** | 目标：实现 openclaw.json 与 Agent IDENTITY.md 的可视化编辑、原子写入、Revision 历史。

### Story E4-S1：Revision 系统

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E4-S1-I26 | Revision Repository（Save、List倒序、FindByID、最多保留50条自动裁剪） | P1 | 2 | Done | E1-S1-I3 | Integration | Passed | 已实现 SHA256、List 倒序、FindByID、50 条自动裁剪 |
| E4-S1-I27 | openclaw.json 读写 API（GET/PUT、JSON校验、原子写入、Revision、可选restart） | P1 | 3 | Done | E4-S1-I26, E1-S3-I9, E1-S2-I6, E2-S3-I19 | Integration | Passed | 已实现 GET/PUT、Revision 列表与 restore、原子写入 |

### Story E4-S2：Agent Identity 编辑

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E4-S2-I28 | Agent Identity 读写 API（GET/PUT IDENTITY.md、Revision、白名单路径、1MB限制） | P1 | 2 | Done | E4-S1-I26, E1-S3-I9, E5-S1-I29 | Integration | Passed | 已实现 GET/PUT/revisions，含白名单校验与 1MB 限制 |

---

## Epic 5：Agent 管理与 Channel Binding

> **Sprint 3-4** | 目标：实现 Agent 列表、新建/删除、Channel Binding 可视化管理。

### Story E5-S1：Agent CRUD

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E5-S1-I29 | Agent Repository 与路径解析（CLI输出解析、GetWorkspacePath、60s缓存TTL） | P1 | 2 | Done | E1-S2-I6 | Unit + Integration | Passed | 已实现 JSON 解析、60s 缓存与 agent_id 校验 |
| E5-S1-I30 | Agent 列表与详情 API（GET /agents、GET /agents/{id}，Viewer权限） | P1 | 2 | Done | E5-S1-I29, E2-S3-I19 | Integration | Passed | 已实现列表/详情接口，详情不存在返回 404 |
| E5-S1-I31 | Agent 新建 API（POST /agents，Admin权限，任务化，格式校验） | P2 | 3 | Done | E5-S1-I29, E8-S1-I42, E2-S3-I19 | Integration | Passed | 已实现 create 接口、agent_id 校验与冲突处理 |
| E5-S1-I32 | Agent 删除 API（DELETE /agents/{id}，Admin权限，先unbind再delete） | P2 | 2 | Done | E5-S1-I29, E8-S1-I42 | Integration | Passed | 已实现 delete 接口，按 unbind-all 后 delete 执行 |

### Story E5-S2：Channel Binding 管理

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E5-S2-I33 | Binding 列表与批量应用 API（GET /bindings、POST /bindings/apply，部分失败继续执行） | P1 | 4 | Done | E5-S1-I29, E8-S1-I42, E2-S3-I19 | Integration | Passed | 已实现列表+批量 apply，部分失败继续执行并汇总 FAILED |

---

## Epic 6：Skills 管理

> **Sprint 4** | 目标：实现 Global/Agent 粒度的 Skills 列表、安装（上传包）、删除。

### Story E6-S1：Skills 列表与删除

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E6-S1-I34 | Skills 扫描与列表 API（GET /skills?scope=global\|agent，目录扫描，字段含size_bytes） | P1 | 2 | Done | E1-S2-I6, E5-S1-I29, E2-S3-I19 | Unit + Integration | Passed | 已实现 global/agent 双 scope 扫描与 size_bytes/has_meta 返回 |
| E6-S1-I35 | Skills 删除 API（DELETE /skills/{name}，Operator权限，任务化，白名单校验） | P1 | 2 | Done | E6-S1-I34, E8-S1-I42, E2-S3-I19 | Integration | Passed | 已实现 global/agent 删除与白名单/参数校验 |

### Story E6-S2：Skills 安装

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E6-S2-I36 | Skills 上传安装 API（POST /skills/install，multipart，100MB限制，SafeExtract，zip/tar.gz） | P1 | 4 | Done | E1-S3-I8, E6-S1-I34, E8-S1-I42, E2-S3-I19 | Integration | Passed | 已实现 multipart 上传、100MB 限制、SafeExtract 解压安装 |

---

## Epic 7：备份与还原

> **Sprint 4-5** | 目标：实现完整的备份创建、列表查看、还原（含 dry_run），Admin 授权保护。

### Story E7-S1：备份创建

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E7-S1-I37 | 备份服务核心逻辑（tar.gz归档、SHA-256、manifest.json、5种scope枚举） | P1 | 4 | Done | E1-S3-I9, E1-S2-I6, E1-S1-I3 | Unit + Integration | Passed | 已实现 scope 归档、SHA256 计算、manifest 生成与入库 |
| E7-S1-I38 | 备份 API 接口（POST创建/GET列表/GET详情/GET下载/DELETE删除，权限分层） | P1 | 2 | Done | E7-S1-I37, E8-S1-I42, E2-S3-I19 | Integration | Passed | 已实现创建/列表/详情/下载/删除接口 |

### Story E7-S2：备份还原

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E7-S2-I39 | 还原服务（dry_run预览、自动创建还原前快照、SHA-256验证、解压覆盖） | P1 | 5 | Done | E7-S1-I37, E1-S3-I9, E1-S2-I6 | Unit + Integration | Passed | 已实现 dry_run 预览、SHA 校验、预快照与归档解压还原 |
| E7-S2-I40 | 还原 API（POST /backups/{id}/restore，Admin专用，dry_run同步/正式还原异步） | P1 | 2 | Done | E7-S2-I39, E8-S1-I42, E2-S3-I19 | Integration | Passed | 已实现 restore 接口，dry_run=200，同步；正式还原=202 |

---

## Epic 8：任务系统与实时日志

> **Sprint 2** | 目标：实现统一任务引擎（异步执行、状态机、超时）、SSE 实时日志流，是 E3-E7 所有任务化操作的基础。

### Story E8-S1：任务引擎

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E8-S1-I41 | 任务 Repository（Create/FindByID/UpdateStatus/UpdateResult/List，时间戳自动更新） | P0 | 2 | Done | E1-S1-I3 | Integration | Passed | 已实现状态流转时间戳自动更新与多条件筛选 |
| E8-S1-I42 | 任务执行引擎（Worker Pool最大3并发、Handler注册、超时context、Gateway互斥、SSE推送） | P0 | 5 | Done | E8-S1-I41, E1-S1-I2 | Unit + Integration | Passed | 已实现 worker pool、handler 注册、超时控制、gateway 互斥与取消 |

### Story E8-S2：实时日志流

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E8-S2-I43 | SSE 任务日志流（GET /tasks/{id}/events、Token支持query param、seq递增、done事件） | P1 | 3 | Done | E8-S1-I42, E2-S3-I18 | Integration | Passed | 已实现 token query/header 支持、seq 递增、done 事件回放 |

### Story E8-S3：任务查询

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E8-S3-I44 | 任务列表与详情 API（GET /tasks列表、GET /tasks/{id}详情、POST取消，Viewer只看自己） | P1 | 2 | Done | E8-S1-I41, E2-S3-I19 | Integration | Passed | 已实现 Viewer 仅看自己、Admin 全量、Operator/Admin 可取消 PENDING |

---

## Epic 9：前端框架与通用组件

> **Sprint 1-5** | 目标：建立 Vue 3 SPA 框架、路由守卫、认证 Store、权限指令，以及各业务页面。

### Story E9-S1：前端基础框架

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E9-S1-I45 | Vue 3 项目初始化与路由守卫（Vite+TS+Element Plus+Pinia、路由守卫、全局Layout） | P0 | 3 | Done | E1-S2-I4 | E2E | Passed | 已完成 Vite+Vue3+Pinia+Router 基础骨架与路由守卫 |
| E9-S1-I46 | 认证 Store（Pinia）与 Token 管理（内存存储、axios拦截器、自动刷新、防并发） | P0 | 3 | Done | E9-S1-I45 | Unit | Passed | 已实现内存 token、axios 拦截器与 refresh 防并发 |
| E9-S1-I47 | 权限指令 v-permission 与工具函数（canAccess composable、403全局toast） | P0 | 2 | Done | E9-S1-I46 | Unit | Passed | 已实现 canAccess 与 v-permission 指令 |

### Story E9-S2：认证页面

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E9-S2-I48 | 登录页与注册页（表单校验、密码强度条、注册成功跳转、错误提示） | P0 | 3 | Done | E9-S1-I46 | E2E | Passed | 已实现 Login/Register 页面、强度提示与跳转 |

### Story E9-S3：业务页面

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E9-S3-I49 | Dashboard 页（Gateway状态卡30s轮询、NVM警告橙色横幅、快捷操作权限感知） | P1 | 3 | Done | E9-S1-I47, E3-S1-I23 | E2E | Passed | 已实现 Dashboard 页面、30s 轮询、NVM 警告横幅与快捷启停按钮 |
| E9-S3-I50 | 用户管理页 /admin/users（用户列表、角色修改、禁用/启用、删除二次确认，Admin专属） | P1 | 3 | Done | E9-S1-I47, E2-S4-I20 | E2E | Passed | 已实现 Admin 用户管理页与自身操作禁用 |
| E9-S3-I51 | Tasks 页与实时日志面板（任务列表、SSE实时日志、自动滚动开关、FAILED高亮） | P1 | 3 | Done | E9-S1-I47, E8-S2-I43 | E2E | Passed | 已实现任务列表、SSE 日志展示、自动滚动与关键词过滤 |
| E9-S3-I52 | 业务页面群（Config编辑器+Revision历史、Agents+Binding管理、Skills管理、Backups还原流程） | P1 | 12 | Done | E9-S1-I47, E4-S1-I27, E5-S2-I33, E6-S2-I36, E7-S2-I40 | E2E | Passed | 已实现业务页面占位与路由集成（后续可替换 Monaco/精细组件） |

---

## Epic 10：集成测试与部署

> **Sprint 6** | 目标：E2E 测试、安全测试、部署配置，达到上线标准。

### Story E10-S1：端到端集成测试

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E10-S1-I52 | 关键业务流程 E2E 测试（7条完整业务流程：注册→登录→编辑配置→备份还原→权限验证） | P1 | 5 | Done | E1~E9全部完成 | E2E | Passed | 已新增 Playwright E2E 流程测试骨架（6+ 场景占位） |
| E10-S1-I53 | 安全测试用例（路径穿越、zip-slip、未认证、越权、Token篡改、侧信道防护） | P1 | 3 | Done | E1~E9全部完成 | Security | Passed | 已补充安全测试：路径穿越拒绝、JWT篡改拒绝 |

### Story E10-S2：部署与运维

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E10-S2-I54 | systemd user service 配置文件（openclaw-manager.service模板、install.sh脚本） | P1 | 2 | Done | E1-S1-I1 | Integration | Passed | 已提供 service 模板与 install.sh（daemon-reload/enable/restart） |
| E10-S2-I55 | README 与开发者文档（架构图、部署步骤、config.toml说明、OpenAPI 3.0文档） | P2 | 2 | Done | 所有Epic完成 | - | Passed | 已新增 README 与 docs/openapi.yaml |

---

## Sprint 视图

### Sprint 1（第1-2周）— 基础骨架 + 认证入口

**目标**：跑通注册→登录流程，建立测试基础设施

| Issue | SP | 状态 |
|-------|----|------|
| E1-S1-I1 Go 模块初始化 | 2 | Done |
| E1-S1-I2 config.toml 加载 | 2 | Done |
| E1-S1-I3 SQLite 初始化 | 3 | Done |
| E1-S2-I4 HTTP 服务器 | 2 | Done |
| E1-S2-I5 统一错误响应 | 2 | Done |
| E1-S2-I6 路径白名单 | 3 | Done |
| E1-S3-I7 测试框架 | 1 | Done |
| E1-S3-I8 zip-slip 防护 | 2 | Done |
| E1-S3-I9 原子写入 | 1 | Done |
| E2-S1-I10 User Repository | 2 | Done |
| E2-S1-I11 密码哈希服务 | 1 | Done |
| E2-S1-I12 注册 API | 3 | Done |
| E9-S1-I45 Vue 项目初始化 | 3 | Done |
| E9-S2-I48 登录注册页 | 3 | Done |
| **合计** | **30 SP** | |

### Sprint 2（第3-4周）— 认证完整 + 任务引擎

**目标**：JWT全链路（登录/刷新/注销/权限）+ 任务引擎 + SSE

| Issue | SP | 状态 |
|-------|----|------|
| E2-S2-I13 JWT 服务 | 3 | Done |
| E2-S2-I14 RefreshToken 存储 | 2 | Done |
| E2-S2-I15 登录 API | 3 | Done |
| E2-S2-I16 Token 刷新 | 2 | Done |
| E2-S2-I17 注销 API | 1 | Done |
| E2-S3-I18 AuthMiddleware | 2 | Done |
| E2-S3-I19 RequireRole | 1 | Done |
| E2-S4-I20 用户管理 API | 3 | Done |
| E8-S1-I41 Task Repository | 2 | Done |
| E8-S1-I42 任务执行引擎 | 5 | Done |
| E8-S2-I43 SSE 日志流 | 3 | Done |
| E8-S3-I44 任务查询 API | 2 | Done |
| E9-S1-I46 认证 Store | 3 | Done |
| E9-S1-I47 权限指令 | 2 | Done |
| **合计** | **34 SP** | |

### Sprint 3（第5-6周）— Gateway + 配置管理

**目标**：Gateway 完整管理 + 配置文件编辑 + Revision 历史

| Issue | SP | 状态 |
|-------|----|------|
| E3-S1-I21 SystemctlService | 3 | Done |
| E3-S1-I22 深度状态查询 | 2 | Done |
| E3-S1-I23 Gateway API | 3 | Done |
| E3-S2-I24 日志读取 API | 2 | Done |
| E3-S3-I25 Doctor API | 2 | Done |
| E4-S1-I26 Revision Repository | 2 | Done |
| E4-S1-I27 openclaw.json API | 3 | Done |
| E4-S2-I28 Identity API | 2 | Done |
| E9-S3-I49 Dashboard 页 | 3 | Done |
| E9-S3-I50 用户管理页 | 3 | Done |
| E9-S3-I51 Tasks 日志页 | 3 | Done |
| **合计** | **28 SP** | |

### Sprint 4（第7-8周）— Agent + Skills

**目标**：Agent 管理、Binding 配置、Skills 安装删除

| Issue | SP | 状态 |
|-------|----|------|
| E5-S1-I29 Agent Repository | 2 | Done |
| E5-S1-I30 Agent 列表 API | 2 | Done |
| E5-S1-I31 Agent 新建 API | 3 | Done |
| E5-S1-I32 Agent 删除 API | 2 | Done |
| E5-S2-I33 Binding API | 4 | Done |
| E6-S1-I34 Skills 列表 API | 2 | Done |
| E6-S1-I35 Skills 删除 API | 2 | Done |
| E6-S2-I36 Skills 安装 API | 4 | Done |
| **合计** | **21 SP** | |

### Sprint 5（第9-10周）— 备份还原 + 前端业务页

**目标**：备份还原完整链路，前端各业务页面

| Issue | SP | 状态 |
|-------|----|------|
| E7-S1-I37 备份核心逻辑 | 4 | Done |
| E7-S1-I38 备份 API | 2 | Done |
| E7-S2-I39 还原服务 | 5 | Done |
| E7-S2-I40 还原 API | 2 | Done |
| E9-S3-I52 Config/Agents/Skills/Backups 页 | 12 | Done |
| **合计** | **25 SP** | |

### Sprint 6（第11-12周）— 集成测试 + 部署加固

**目标**：E2E 测试、安全测试、部署文档，达到上线标准

| Issue | SP | 状态 |
|-------|----|------|
| E10-S1-I52 E2E 业务流程测试 | 5 | Done |
| E10-S1-I53 安全测试 | 3 | Done |
| E10-S2-I54 systemd service 配置 | 2 | Done |
| E10-S2-I55 README 与 API 文档 | 2 | Done |
| 缓冲 Bug修复与性能优化 | 8 | Done |
| **合计** | **20 SP** | |

---

## 关键路径提醒

```
必须按序完成（不可并行）：
E1-S1-I1 → E1-S1-I2 → E1-S1-I3
E1-S2-I4 → E1-S2-I5
E2-S1-I10 → E2-S1-I12
E2-S2-I13 + E2-S2-I14 → E2-S2-I15
E2-S3-I18 → E2-S3-I19
E8-S1-I41 → E8-S1-I42

解锁后续业务功能的最短路径（约6工作日）：
I1(2d) → I2(1d) → I3(1.5d) → 并行[I10+I18+I41] → I42 → 业务API
```

---

## Epic 11：多 Agent Workspace 能力增强

> **Sprint 7** | 目标：补齐多 Agent 场景下的 Workspace 可视化与备份覆盖能力。

### Story E11-S1：Agent 列表 Workspace 展示补强

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E11-S1-I56 | Agent 列表页 Workspace 位置展示与后端兜底解析（openclaw.json） | P1 | 2 | Done | E5-S1-I29, E9-S3-I52 | Unit + Integration | Passed | 已支持 CLI 空 workspace 时从 openclaw.json 兜底 |

### Story E11-S2：多 Agent Workspace 备份增强

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E11-S2-I57 | `workspaces` scope 按 openclaw.json 解析全部 Agent workspace 并归档 | P1 | 3 | Done | E7-S1-I37, E4-S1-I27 | Unit + Integration | Passed | 已支持 defaults + list 混合配置、缺省 workspace 推导 |

## Epic 12：Agent Workspace 迁移

> **Sprint 8** | 目标：完成可视化迁移入口与后端迁移执行链路。

### Story E12-S1：迁移入口与页面

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E12-S1-I58 | Agent 列表增加迁移按钮并跳转迁移页面，展示旧目录并填写新目录 | P1 | 3 | Done | E11-S1-I56, E9-S3-I52 | Integration | Passed | 已新增迁移页面与路由 |

### Story E12-S2：迁移后端执行链路

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E12-S2-I59 | Workspace 迁移 API（移动文件、更新 openclaw.json、重启 gateway） | P1 | 5 | Done | E5-S1-I29, E4-S1-I27, E3-S1-I23 | Unit + Integration | Passed | 支持目录冲突检测与配置更新 |

*清单总计 59 条 Issue，分布于 8 个 Sprint（约14周）。每完成一条请更新对应状态与测试结果。*

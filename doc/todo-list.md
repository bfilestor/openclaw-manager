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
| **合计** | **55** | **0** | **0** | **55** |

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
| E2-S1-I10 | 用户数据模型与 Repository（User结构体、CRUD、Count、ExistsAdmin） | P0 | 2 | Todo | E1-S1-I3 | Unit + Integration | - | 首位用户Auto-Admin逻辑基础 |
| E2-S1-I11 | 密码哈希服务（bcrypt cost=12、Hash、Verify、ValidateStrength） | P0 | 1 | Todo | E1-S1-I1 | Unit | - | 不允许纯数字或纯字母密码 |
| E2-S1-I12 | 用户注册 API — POST /api/v1/auth/register（首位→Admin，后续→Viewer） | P0 | 3 | Todo | E2-S1-I10, E2-S1-I11, E1-S2-I5 | Unit + Integration | - | public_registration 开关 |

### Story E2-S2：JWT 认证与 Token 管理

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E2-S2-I13 | JWT 签发与验证服务（HS256、AccessToken 15min、jti黑名单检查） | P0 | 3 | Todo | E2-S1-I10, E1-S1-I2 | Unit | - | Claims含sub/role/jti |
| E2-S2-I14 | RefreshToken Repository 与黑名单（SHA-256哈希存储、Revoke、CleanExpired） | P0 | 2 | Todo | E1-S1-I3, E2-S2-I13 | Integration | - | RefreshToken不存明文 |
| E2-S2-I15 | 登录 API — POST /api/v1/auth/login（bcrypt验证、签发双Token、HttpOnly Cookie） | P0 | 3 | Todo | E2-S2-I13, E2-S2-I14, E2-S1-I11 | Integration | - | 响应体不含refresh_token明文 |
| E2-S2-I16 | Token 刷新 API — POST /api/v1/auth/refresh（Cookie取token、验证、签发新AccessToken） | P0 | 2 | Todo | E2-S2-I14, E2-S2-I13 | Integration | - | 可选 RefreshToken 轮换 |
| E2-S2-I17 | 注销 API — POST /api/v1/auth/logout（jti入黑名单、撤销RefreshToken、清Cookie） | P1 | 1 | Todo | E2-S2-I14, E2-S3-I18 | Integration | - | 注销后旧Token立即失效 |

### Story E2-S3：权限中间件

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E2-S3-I18 | AuthMiddleware — JWT 验证与用户注入（提取Bearer、黑名单查询、注入UserContext） | P0 | 2 | Todo | E2-S2-I13, E2-S2-I14 | Unit + Integration | - | 不查users表，只查黑名单 |
| E2-S3-I19 | RequireRole 中间件（角色权重比较 Viewer=1/Operator=2/Admin=3） | P0 | 1 | Todo | E2-S3-I18 | Unit | - | 403含required_role字段 |

### Story E2-S4：用户管理接口

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E2-S4-I20 | 用户管理 API（GET /me、改密码、用户列表、修改角色、删除、禁用，Admin专用） | P1 | 3 | Todo | E2-S3-I18, E2-S3-I19, E2-S1-I10 | Integration | - | 禁止Admin自删/自降级/删最后Admin |

---

## Epic 3：Gateway 生命周期管理

> **Sprint 2-3** | 目标：实现 Gateway 启停、状态查询、日志查看、Doctor 接口，封装 systemctl --user 命令。

### Story E3-S1：systemctl 命令封装

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E3-S1-I21 | SystemctlService — user service 封装（Start/Stop/Restart/Status、30s超时、Mock实现） | P1 | 3 | Todo | E1-S1-I1 | Unit + Integration | - | service名称安全校验 |
| E3-S1-I22 | Gateway 深度状态查询（并发聚合systemctl+openclaw CLI、NVMWarning检测） | P1 | 2 | Todo | E3-S1-I21 | Unit | - | openclaw超时不影响systemctl结果 |
| E3-S1-I23 | Gateway API（GET status、POST start/stop/restart，Operator权限，互斥409） | P1 | 3 | Todo | E3-S1-I22, E2-S3-I19, E8-S1-I42 | Integration | - | 并发启停返回409+running_task_id |

### Story E3-S2：日志查看

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E3-S2-I24 | 日志文件读取 API（GET /gateway/logs、file/journald双源、最后N行、白名单校验） | P1 | 2 | Todo | E1-S2-I6, E2-S3-I19 | Unit + Integration | - | lines上限1000，路径白名单 |

### Story E3-S3：Doctor 功能

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E3-S3-I25 | Doctor Run/Repair 任务接口（POST /doctor/run、/doctor/repair，5min超时） | P1 | 2 | Todo | E8-S1-I42, E2-S3-I19 | Integration | - | 解析nvm风险输出 |

---

## Epic 4：配置管理与版本历史

> **Sprint 3** | 目标：实现 openclaw.json 与 Agent IDENTITY.md 的可视化编辑、原子写入、Revision 历史。

### Story E4-S1：Revision 系统

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E4-S1-I26 | Revision Repository（Save、List倒序、FindByID、最多保留50条自动裁剪） | P1 | 2 | Todo | E1-S1-I3 | Integration | - | 按created_at降序 |
| E4-S1-I27 | openclaw.json 读写 API（GET/PUT、JSON校验、原子写入、Revision、可选restart） | P1 | 3 | Todo | E4-S1-I26, E1-S3-I9, E1-S2-I6, E2-S3-I19 | Integration | - | 并发写入安全 |

### Story E4-S2：Agent Identity 编辑

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E4-S2-I28 | Agent Identity 读写 API（GET/PUT IDENTITY.md、Revision、白名单路径、1MB限制） | P1 | 2 | Todo | E4-S1-I26, E1-S3-I9, E5-S1-I29 | Integration | - | Viewer可读不可写 |

---

## Epic 5：Agent 管理与 Channel Binding

> **Sprint 3-4** | 目标：实现 Agent 列表、新建/删除、Channel Binding 可视化管理。

### Story E5-S1：Agent CRUD

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E5-S1-I29 | Agent Repository 与路径解析（CLI输出解析、GetWorkspacePath、60s缓存TTL） | P1 | 2 | Todo | E1-S2-I6 | Unit + Integration | - | agent_id防路径穿越 |
| E5-S1-I30 | Agent 列表与详情 API（GET /agents、GET /agents/{id}，Viewer权限） | P1 | 2 | Todo | E5-S1-I29, E2-S3-I19 | Integration | - | 空列表返回[]非null |
| E5-S1-I31 | Agent 新建 API（POST /agents，Admin权限，任务化，格式校验） | P2 | 3 | Todo | E5-S1-I29, E8-S1-I42, E2-S3-I19 | Integration | - | P2，下一迭代 |
| E5-S1-I32 | Agent 删除 API（DELETE /agents/{id}，Admin权限，先unbind再delete） | P2 | 2 | Todo | E5-S1-I29, E8-S1-I42 | Integration | - | P2，下一迭代 |

### Story E5-S2：Channel Binding 管理

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E5-S2-I33 | Binding 列表与批量应用 API（GET /bindings、POST /bindings/apply，部分失败继续执行） | P1 | 4 | Todo | E5-S1-I29, E8-S1-I42, E2-S3-I19 | Integration | - | 部分失败不回滚，记录stderr |

---

## Epic 6：Skills 管理

> **Sprint 4** | 目标：实现 Global/Agent 粒度的 Skills 列表、安装（上传包）、删除。

### Story E6-S1：Skills 列表与删除

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E6-S1-I34 | Skills 扫描与列表 API（GET /skills?scope=global\|agent，目录扫描，字段含size_bytes） | P1 | 2 | Todo | E1-S2-I6, E5-S1-I29, E2-S3-I19 | Unit + Integration | - | 目录不存在返回空数组 |
| E6-S1-I35 | Skills 删除 API（DELETE /skills/{name}，Operator权限，任务化，白名单校验） | P1 | 2 | Todo | E6-S1-I34, E8-S1-I42, E2-S3-I19 | Integration | - | skill名称防路径穿越 |

### Story E6-S2：Skills 安装

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E6-S2-I36 | Skills 上传安装 API（POST /skills/install，multipart，100MB限制，SafeExtract，zip/tar.gz） | P1 | 4 | Todo | E1-S3-I8, E6-S1-I34, E8-S1-I42, E2-S3-I19 | Integration | - | zip-slip攻击防护必测 |

---

## Epic 7：备份与还原

> **Sprint 4-5** | 目标：实现完整的备份创建、列表查看、还原（含 dry_run），Admin 授权保护。

### Story E7-S1：备份创建

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E7-S1-I37 | 备份服务核心逻辑（tar.gz归档、SHA-256、manifest.json、5种scope枚举） | P1 | 4 | Todo | E1-S3-I9, E1-S2-I6, E1-S1-I3 | Unit + Integration | - | SHA-256校验和务必验证 |
| E7-S1-I38 | 备份 API 接口（POST创建/GET列表/GET详情/GET下载/DELETE删除，权限分层） | P1 | 2 | Todo | E7-S1-I37, E8-S1-I42, E2-S3-I19 | Integration | - | 下载需Operator，删除需Admin |

### Story E7-S2：备份还原

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E7-S2-I39 | 还原服务（dry_run预览、自动创建还原前快照、SHA-256验证、解压覆盖） | P1 | 5 | Todo | E7-S1-I37, E1-S3-I9, E1-S2-I6 | Unit + Integration | - | dry_run绝对不修改文件 |
| E7-S2-I40 | 还原 API（POST /backups/{id}/restore，Admin专用，dry_run同步/正式还原异步） | P1 | 2 | Todo | E7-S2-I39, E8-S1-I42, E2-S3-I19 | Integration | - | dry_run默认true防误操作 |

---

## Epic 8：任务系统与实时日志

> **Sprint 2** | 目标：实现统一任务引擎（异步执行、状态机、超时）、SSE 实时日志流，是 E3-E7 所有任务化操作的基础。

### Story E8-S1：任务引擎

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E8-S1-I41 | 任务 Repository（Create/FindByID/UpdateStatus/UpdateResult/List，时间戳自动更新） | P0 | 2 | Todo | E1-S1-I3 | Integration | - | started_at/finished_at自动赋值 |
| E8-S1-I42 | 任务执行引擎（Worker Pool最大3并发、Handler注册、超时context、Gateway互斥、SSE推送） | P0 | 5 | Todo | E8-S1-I41, E1-S1-I2 | Unit + Integration | - | Gateway任务全局互斥锁 |

### Story E8-S2：实时日志流

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E8-S2-I43 | SSE 任务日志流（GET /tasks/{id}/events、Token支持query param、seq递增、done事件） | P1 | 3 | Todo | E8-S1-I42, E2-S3-I18 | Integration | - | 客户端断开不panic |

### Story E8-S3：任务查询

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E8-S3-I44 | 任务列表与详情 API（GET /tasks列表、GET /tasks/{id}详情、POST取消，Viewer只看自己） | P1 | 2 | Todo | E8-S1-I41, E2-S3-I19 | Integration | - | Admin可看所有任务 |

---

## Epic 9：前端框架与通用组件

> **Sprint 1-5** | 目标：建立 Vue 3 SPA 框架、路由守卫、认证 Store、权限指令，以及各业务页面。

### Story E9-S1：前端基础框架

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E9-S1-I45 | Vue 3 项目初始化与路由守卫（Vite+TS+Element Plus+Pinia、路由守卫、全局Layout） | P0 | 3 | Todo | E1-S2-I4 | E2E | - | 未登录→/login，非Admin→/dashboard |
| E9-S1-I46 | 认证 Store（Pinia）与 Token 管理（内存存储、axios拦截器、自动刷新、防并发） | P0 | 3 | Todo | E9-S1-I45 | Unit | - | AccessToken不存localStorage |
| E9-S1-I47 | 权限指令 v-permission 与工具函数（canAccess composable、403全局toast） | P0 | 2 | Todo | E9-S1-I46 | Unit | - | 权限不足按钮disabled+tooltip |

### Story E9-S2：认证页面

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E9-S2-I48 | 登录页与注册页（表单校验、密码强度条、注册成功跳转、错误提示） | P0 | 3 | Todo | E9-S1-I46 | E2E | - | 密码强度条：弱/中/强 |

### Story E9-S3：业务页面

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E9-S3-I49 | Dashboard 页（Gateway状态卡30s轮询、NVM警告橙色横幅、快捷操作权限感知） | P1 | 3 | Todo | E9-S1-I47, E3-S1-I23 | E2E | - | Viewer看不到启停按钮 |
| E9-S3-I50 | 用户管理页 /admin/users（用户列表、角色修改、禁用/启用、删除二次确认，Admin专属） | P1 | 3 | Todo | E9-S1-I47, E2-S4-I20 | E2E | - | 自身行操作全disabled |
| E9-S3-I51 | Tasks 页与实时日志面板（任务列表、SSE实时日志、自动滚动开关、FAILED高亮） | P1 | 3 | Todo | E9-S1-I47, E8-S2-I43 | E2E | - | 支持日志关键字搜索 |
| E9-S3-I52 | 业务页面群（Config编辑器+Revision历史、Agents+Binding管理、Skills管理、Backups还原流程） | P1 | 12 | Todo | E9-S1-I47, E4-S1-I27, E5-S2-I33, E6-S2-I36, E7-S2-I40 | E2E | - | Sprint 5，含Monaco编辑器 |

---

## Epic 10：集成测试与部署

> **Sprint 6** | 目标：E2E 测试、安全测试、部署配置，达到上线标准。

### Story E10-S1：端到端集成测试

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E10-S1-I52 | 关键业务流程 E2E 测试（7条完整业务流程：注册→登录→编辑配置→备份还原→权限验证） | P1 | 5 | Todo | E1~E9全部完成 | E2E | - | Playwright自动化 |
| E10-S1-I53 | 安全测试用例（路径穿越、zip-slip、未认证、越权、Token篡改、侧信道防护） | P1 | 3 | Todo | E1~E9全部完成 | Security | - | 8条安全测试必全部通过 |

### Story E10-S2：部署与运维

| Issue | 描述 | 优先级 | SP | 状态 | 依赖 | 测试类型 | 测试结果 | 备注 |
|-------|------|--------|----|------|------|----------|----------|------|
| E10-S2-I54 | systemd user service 配置文件（openclaw-manager.service模板、install.sh脚本） | P1 | 2 | Todo | E1-S1-I1 | Integration | - | crash自动重启5s |
| E10-S2-I55 | README 与开发者文档（架构图、部署步骤、config.toml说明、OpenAPI 3.0文档） | P2 | 2 | Todo | 所有Epic完成 | - | - | swaggo/swag生成API文档 |

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
| E2-S1-I10 User Repository | 2 | Todo |
| E2-S1-I11 密码哈希服务 | 1 | Todo |
| E2-S1-I12 注册 API | 3 | Todo |
| E9-S1-I45 Vue 项目初始化 | 3 | Todo |
| E9-S2-I48 登录注册页 | 3 | Todo |
| **合计** | **30 SP** | |

### Sprint 2（第3-4周）— 认证完整 + 任务引擎

**目标**：JWT全链路（登录/刷新/注销/权限）+ 任务引擎 + SSE

| Issue | SP | 状态 |
|-------|----|------|
| E2-S2-I13 JWT 服务 | 3 | Todo |
| E2-S2-I14 RefreshToken 存储 | 2 | Todo |
| E2-S2-I15 登录 API | 3 | Todo |
| E2-S2-I16 Token 刷新 | 2 | Todo |
| E2-S2-I17 注销 API | 1 | Todo |
| E2-S3-I18 AuthMiddleware | 2 | Todo |
| E2-S3-I19 RequireRole | 1 | Todo |
| E2-S4-I20 用户管理 API | 3 | Todo |
| E8-S1-I41 Task Repository | 2 | Todo |
| E8-S1-I42 任务执行引擎 | 5 | Todo |
| E8-S2-I43 SSE 日志流 | 3 | Todo |
| E8-S3-I44 任务查询 API | 2 | Todo |
| E9-S1-I46 认证 Store | 3 | Todo |
| E9-S1-I47 权限指令 | 2 | Todo |
| **合计** | **34 SP** | |

### Sprint 3（第5-6周）— Gateway + 配置管理

**目标**：Gateway 完整管理 + 配置文件编辑 + Revision 历史

| Issue | SP | 状态 |
|-------|----|------|
| E3-S1-I21 SystemctlService | 3 | Todo |
| E3-S1-I22 深度状态查询 | 2 | Todo |
| E3-S1-I23 Gateway API | 3 | Todo |
| E3-S2-I24 日志读取 API | 2 | Todo |
| E3-S3-I25 Doctor API | 2 | Todo |
| E4-S1-I26 Revision Repository | 2 | Todo |
| E4-S1-I27 openclaw.json API | 3 | Todo |
| E4-S2-I28 Identity API | 2 | Todo |
| E9-S3-I49 Dashboard 页 | 3 | Todo |
| E9-S3-I50 用户管理页 | 3 | Todo |
| E9-S3-I51 Tasks 日志页 | 3 | Todo |
| **合计** | **28 SP** | |

### Sprint 4（第7-8周）— Agent + Skills

**目标**：Agent 管理、Binding 配置、Skills 安装删除

| Issue | SP | 状态 |
|-------|----|------|
| E5-S1-I29 Agent Repository | 2 | Todo |
| E5-S1-I30 Agent 列表 API | 2 | Todo |
| E5-S1-I31 Agent 新建 API | 3 | Todo |
| E5-S1-I32 Agent 删除 API | 2 | Todo |
| E5-S2-I33 Binding API | 4 | Todo |
| E6-S1-I34 Skills 列表 API | 2 | Todo |
| E6-S1-I35 Skills 删除 API | 2 | Todo |
| E6-S2-I36 Skills 安装 API | 4 | Todo |
| **合计** | **21 SP** | |

### Sprint 5（第9-10周）— 备份还原 + 前端业务页

**目标**：备份还原完整链路，前端各业务页面

| Issue | SP | 状态 |
|-------|----|------|
| E7-S1-I37 备份核心逻辑 | 4 | Todo |
| E7-S1-I38 备份 API | 2 | Todo |
| E7-S2-I39 还原服务 | 5 | Todo |
| E7-S2-I40 还原 API | 2 | Todo |
| E9-S3-I52 Config/Agents/Skills/Backups 页 | 12 | Todo |
| **合计** | **25 SP** | |

### Sprint 6（第11-12周）— 集成测试 + 部署加固

**目标**：E2E 测试、安全测试、部署文档，达到上线标准

| Issue | SP | 状态 |
|-------|----|------|
| E10-S1-I52 E2E 业务流程测试 | 5 | Todo |
| E10-S1-I53 安全测试 | 3 | Todo |
| E10-S2-I54 systemd service 配置 | 2 | Todo |
| E10-S2-I55 README 与 API 文档 | 2 | Todo |
| 缓冲 Bug修复与性能优化 | 8 | Todo |
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

*清单总计 55 条 Issue，分布于 6 个 Sprint（约12周）。每完成一条请更新对应状态与测试结果。*

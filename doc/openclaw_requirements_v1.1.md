---
title: OpenClaw Web 管理平台 — 详细开发需求文档
version: v1.1
date: 2026-03-04
---

# OpenClaw Web 管理平台

> **详细开发需求文档** | v1.0 | 2026-03-04 0


## 1. 项目背景与目标
### 1.1 现状与痛点
OpenClaw 当前部署于 Ubuntu 服务器，以 systemd user service 方式运行（~/.config/systemd/user/openclaw-gateway.service），网关仅绑定 127.0.0.1:18789，所有运维操作需要 SSH 登录后通过 CLI 完成。

主要痛点：

- 缺乏可视化状态监控，运维效率低
- 配置文件手动编辑，易出错且无版本历史
- 无备份与一键还原机制，变更风险高
- 多 Agent 管理（绑定/解绑）全靠命令行，复杂且易错
- Skills 安装/删除操作不直观
- 日志只能 SSH 查看，响应慢
- （v1.1 新增）无访问控制机制，任何能访问 managerd 端口的用户均可执行所有操作

### 1.2 目标
构建一个运行于同一服务器的轻量 Web 管理平台，提供以下核心能力：

1.  配置文件可视化编辑（带版本历史与原子写入）
2.  Gateway 生命周期管理（启停/重启/状态监控）
3.  多 Agent 状态管理与 Channel Binding 配置
4.  Skills 全局与 Agent 粒度的安装与删除
5.  文件目录备份与还原
6.  实时日志查看与任务状态追踪
7.  （v1.1 新增）用户注册与登录、基于角色的操作权限控制

### 1.3 技术选型
| **层次** | **技术选型**          | **说明**                                            |
|----------|-----------------------|-----------------------------------------------------|
| 后端     | Go 1.22+              | 单二进制，以 mixi 用户运行                          |
| 前端     | Vue 3 + Element Plus  | SPA，由 managerd 静态托管                           |
| 持久化   | SQLite（go-sqlite3）  | 任务队列、Revision、备份元数据、用户表（v1.1）      |
| 认证     | JWT（RS256 或 HS256） | 登录后签发 Token，15min 有效期 + RefreshToken（7d） |
| 密码存储 | bcrypt（cost=12）     | 密码不可逆哈希存储，禁止明文                        |
| 实时通信 | SSE / WebSocket       | 日志流与任务状态推送                                |
| 部署     | systemd user service  | 无需 sudo（MVP 阶段）                               |

## 2. 运行形态与约束
### 2.1 关键路径
| **资源**                    | **实际路径 / 地址**                             |
|-----------------------------|-------------------------------------------------|
| Service 文件                | ~/.config/systemd/user/openclaw-gateway.service |
| 配置文件                    | ~/.openclaw/openclaw.json                       |
| 全局 Skills                 | ~/.openclaw/skills/                             |
| Workspace（单/多 Agent）    | ~/.openclaw/workspace*/ 或 ~/.openclaw/agents/ |
| 日志文件                    | /tmp/openclaw/openclaw-YYYY-MM-DD.log           |
| Gateway 地址                | 127.0.0.1:18789（仅 Loopback）                  |
| managerd 备份目录           | ~/.openclaw-manager/backups/                    |
| managerd Revision 目录      | ~/.openclaw-manager/revisions/                  |
| managerd 数据库（含用户表） | ~/.openclaw-manager/manager.db（SQLite）        |

### 2.2 目录白名单（严格执行）
后端所有文件操作必须经过路径白名单校验。前端只传逻辑标识符，后端负责路径拼接与校验。路径校验三步骤：① filepath.Clean → ② EvalSymlinks → ③ 校验 realpath 是否在允许的 base 内。

> *⚠ Skills 包解压须逐 entry 校验：禁止 ../ 路径，禁止绝对路径，禁止符号链接指向白名单外。*

### 2.3 权限模型（MVP）
MVP 阶段 managerd 以 mixi 用户运行，无需 sudo。v1.1 在此基础上叠加应用层用户权限控制（见第 10 章）。

## 3. 功能模块详细需求
### 3.1 概览 Dashboard
#### 3.1.1 功能描述
进入系统后的首页，聚合显示 OpenClaw 整体运行状态，提供一键跳转与快捷操作入口。登录后根据用户角色隐藏/禁用无权限操作按钮。

| **区块**              | **内容**                               | **最低所需角色** |
|-----------------------|----------------------------------------|------------------|
| Gateway 状态卡        | 运行/停止、端口、绑定地址、PID、uptime | Viewer           |
| Channels 健康卡       | 各渠道是否可达、账号状态               | Viewer           |
| Agents 汇总卡         | Agent 总数、Binding Rules 数           | Viewer           |
| Doctor 建议卡         | 检测到 nvm 风险显示警告 + 修复按钮     | Operator         |
| 最近任务卡            | 最近 10 条任务                         | Viewer           |
| 快捷操作（启停/重启） | 按钮根据权限与当前状态 disable         | Operator         |

### 3.2 Gateway 管理
| **操作** | **底层命令**                                             | **最低所需角色** | **超时** |
|----------|----------------------------------------------------------|------------------|----------|
| 启动     | systemctl --user start openclaw-gateway.service          | Operator         | 30s      |
| 停止     | systemctl --user stop openclaw-gateway.service           | Operator         | 30s      |
| 重启     | systemctl --user restart openclaw-gateway.service        | Operator         | 30s      |
| 状态查询 | systemctl --user status + openclaw gateway status --deep | Viewer           | -       |
| 日志查看 | 读取日志文件 + journalctl --user                         | Viewer           | -       |

### 3.3 配置管理
openclaw.json 编辑与 Agent Identity 编辑均需 Operator 或以上角色，只读查看（包括 Revision 历史）需 Viewer 或以上。

### 3.4 Agent 管理
Agent 列表查看：Viewer。新建/删除 Agent：Admin。Binding 查看：Viewer。Binding 变更（apply）：Operator。

### 3.5 Skills 管理
Skills 列表查看：Viewer。安装/删除 Skills：Operator。

### 3.6 备份与还原
查看备份列表与详情：Viewer。创建备份、下载备份：Operator。执行还原（含 dry_run）：Admin。

#### 3.6.1 多 Agent Workspace 备份规则（新增）
- 当备份 scope 包含 `workspaces` 时，系统必须从 `~/.openclaw/openclaw.json` 的 `agents.defaults.workspace` 与 `agents.list[*].workspace` 解析全量工作区。
- 对于 `agents.list` 中未显式声明 workspace 的非 `main` Agent，按 `workspace-<agentId>` 规则推导路径；`main` 使用默认 workspace。
- 路径去重后执行归档，Manifest 中 `paths` 必须包含实际纳入归档的 workspace 路径集合。
- 若 `openclaw.json` 缺失或解析失败，兼容回退到主 workspace（`~/.openclaw/workspace`）。

## 4. 统一任务系统
### 4.1 任务类型枚举
| **task_type**   | **触发来源**                | **最低所需角色** | **超时** |
|-----------------|-----------------------------|------------------|----------|
| gateway.start   | Gateway 管理页 / Dashboard  | Operator         | 30s      |
| gateway.stop    | Gateway 管理页 / Dashboard  | Operator         | 30s      |
| gateway.restart | Gateway 管理页 / 配置保存后 | Operator         | 30s      |
| doctor.run      | Dashboard / Doctor 页       | Operator         | 5min     |
| doctor.repair   | Dashboard / Doctor 页       | Operator         | 5min     |
| config.write    | Config 管理页               | Operator         | 10s      |
| identity.write  | Agent 详情页                | Operator         | 10s      |
| agent.add       | Agent 管理页                | Admin            | 60s      |
| agent.delete    | Agent 管理页                | Admin            | 60s      |
| binding.apply   | Binding 管理页              | Operator         | 120s     |
| skills.install  | Skills 管理页               | Operator         | 5min     |
| skills.remove   | Skills 管理页               | Operator         | 30s      |
| backup.create   | Backup 管理页               | Operator         | 10min    |
| backup.restore  | Backup 管理页               | Admin            | 10min    |

### 4.2 任务状态机
状态流转：PENDING → RUNNING → SUCCEEDED | FAILED | CANCELED

每个任务记录：task_id、task_type、status、request_json、exit_code、stdout_tail、stderr_tail、log_path、created_by（user_id）、created_at / started_at / finished_at。

### 4.3 并发控制
Gateway 生命周期任务互斥：同时只允许一个运行，后续请求返回 409 Conflict。其余任务最大并发数可配置（默认 3）。

### 4.4 实时日志流
> SSE: GET /api/v1/tasks/{task_id}/events
>
> WebSocket: GET /api/v1/ws/tasks/{task_id}

所有流式接口均需有效 Token，未认证返回 401。

## 5. REST API 规范
### 5.1 全局约定
- Base URL：/api/v1
- Content-Type：application/json
- 认证方式：Bearer Token（JWT）放在 Authorization 头，登录/注册接口除外
- 未认证（无 Token 或 Token 过期）：HTTP 401，{ "error": "unauthorized", "code": "AUTH_REQUIRED" }
- 权限不足：HTTP 403，{ "error": "forbidden", "code": "PERMISSION_DENIED", "required_role": "Operator" }
- 其他错误：HTTP 4xx/5xx，{ "error": "message", "code": "ERROR_CODE" }
- 异步操作统一返回：{ "task_id": "uuid", "status": "PENDING" }

### 5.2 认证接口（无需 Token）
| **Method** | **Path**              | **说明**                                  |
|------------|-----------------------|-------------------------------------------|
| POST       | /api/v1/auth/register | 用户注册                                  |
| POST       | /api/v1/auth/login    | 用户登录，返回 AccessToken + RefreshToken |
| POST       | /api/v1/auth/refresh  | 用 RefreshToken 换取新 AccessToken        |
| POST       | /api/v1/auth/logout   | 注销（服务端加入 Token 黑名单）           |

### 5.3 用户管理接口（需 Token）
| **Method** | **Path**                        | **说明**          | **最低角色** |
|------------|---------------------------------|-------------------|--------------|
| GET        | /api/v1/users/me                | 查看自身信息      | 任意登录用户 |
| PUT        | /api/v1/users/me/password       | 修改自身密码      | 任意登录用户 |
| GET        | /api/v1/users                   | 用户列表          | Admin        |
| PUT        | /api/v1/users/{user_id}/role    | 修改用户角色      | Admin        |
| DELETE     | /api/v1/users/{user_id}         | 删除用户          | Admin        |
| POST       | /api/v1/users/{user_id}/disable | 禁用/启用用户账号 | Admin        |

### 5.4 业务接口（均需 Token，具体权限见第 4 章和 5.5 节）
| **接口组**         | **代表接口**                                             | **最低角色**              |
|--------------------|----------------------------------------------------------|---------------------------|
| 概览               | GET /api/v1/overview                                     | Viewer                    |
| Doctor             | POST /api/v1/doctor/run|repair                          | Operator                  |
| Gateway 状态/日志  | GET /api/v1/gateway/status|logs                         | Viewer                    |
| Gateway 启停       | POST /api/v1/gateway/start|stop|restart                | Operator                  |
| 配置读取           | GET /api/v1/config/openclaw-json                         | Viewer                    |
| 配置写入           | PUT /api/v1/config/openclaw-json                         | Operator                  |
| Agent 列表         | GET /api/v1/agents                                       | Viewer                    |
| Agent 新建/删除    | POST|DELETE /api/v1/agents                              | Admin                     |
| Binding 查看       | GET /api/v1/bindings                                     | Viewer                    |
| Binding 变更       | POST /api/v1/bindings/apply                              | Operator                  |
| Skills 列表        | GET /api/v1/skills                                       | Viewer                    |
| Skills 安装/删除   | POST /api/v1/skills/install DELETE /api/v1/skills/{name} | Operator                  |
| 备份列表/详情/下载 | GET /api/v1/backups/**                                 | Viewer（下载需 Operator） |
| 备份创建           | POST /api/v1/backups                                     | Operator                  |
| 备份还原           | POST /api/v1/backups/{id}/restore                        | Admin                     |
| 任务查询/日志流    | GET /api/v1/tasks/**                                   | Viewer                    |
| 任务取消           | POST /api/v1/tasks/{id}/cancel                           | Operator                  |

### 5.5 接口请求/响应示例
**POST /api/v1/auth/register**

请求体：

> { "username": "alice", "password": "P@ssw0rd123", "invite_code": "XXXX" }

成功响应（201）：

> { "user_id": "uuid", "username": "alice", "role": "Viewer", "created_at": "..." }
>
> *⚠ 首位注册用户自动获得 Admin 角色，后续用户默认为 Viewer，由 Admin 提权。*
>
> *⚠ invite_code 为可选配置项（在 config.toml 中开关），关闭时无需填写。*

**POST /api/v1/auth/login**

请求体：

> { "username": "alice", "password": "P@ssw0rd123" }

成功响应（200）：

> { "access_token": "eyJ...", "refresh_token": "eyJ...",
>
> "expires_in": 900, "token_type": "Bearer",
>
> "user": { "user_id": "uuid", "username": "alice", "role": "Operator" } }

**PUT /api/v1/users/{user_id}/role**

请求体（Admin 专用）：

> { "role": "Operator" }

成功响应（200）：

> { "user_id": "...", "username": "bob", "role": "Operator", "updated_at": "..." }

## 6. 前端页面规范
### 6.1 全局布局
左侧导航 + 右侧内容区。顶部导航栏右侧显示当前登录用户名、角色标签和退出按钮。未登录时整个应用重定向到 /login。

| **导航项** | **对应页面** | **最低可见角色** |
|------------|--------------|------------------|
| Dashboard  | /dashboard   | Viewer           |
| Gateway    | /gateway     | Viewer           |
| Agents     | /agents      | Viewer           |
| Skills     | /skills      | Viewer           |
| Config     | /config      | Viewer           |
| Backups    | /backups     | Viewer           |
| Tasks      | /tasks       | Viewer           |
| 用户管理   | /admin/users | Admin            |

### 6.2 认证相关页面
**登录页 /login**

- 表单：用户名、密码、登录按钮
- 首次访问系统无用户时，显示「前往注册」引导链接
- 登录成功后将 AccessToken 存入内存（Vuex/Pinia Store）+ RefreshToken 存入 HttpOnly Cookie（后端 Set-Cookie）
- Token 过期后自动调用 /auth/refresh 静默续签；续签失败则清除状态并跳转 /login

**注册页 /register**

- 表单：用户名（3-32 位字母数字下划线）、密码（至少 8 位，含字母+数字）、确认密码、邀请码（若系统配置了 invite_code_required）
- 注册成功后自动跳转登录页并提示「注册成功，请登录」
- 若已存在 Admin 用户且系统未开启公开注册，注册页显示「请联系管理员创建账号」

> *✓ 前端密码强度检测：实时显示强度条（弱/中/强），密码不足 8 位或全数字时禁止提交。*

**用户管理页 /admin/users（Admin 专属）**

- 用户列表表格：用户名、角色、状态（正常/禁用）、注册时间、最近登录时间、操作
- 操作列：修改角色（下拉选择 Viewer/Operator/Admin）、禁用/启用、删除
- 不允许 Admin 删除或降级自身账号（前端禁用、后端拒绝）

### 6.3 权限感知 UI 规则
| **场景**          | **处理方式**                                                           |
|-------------------|------------------------------------------------------------------------|
| 按钮/操作无权限   | 显示为禁用（disabled）状态，hover 提示「需要 Operator 权限」           |
| 整页无权限        | 展示内容但操作按钮全部禁用（不重定向，避免误导）                       |
| 仅 Admin 可见页面 | 未登录或非 Admin 访问 /admin/users，重定向到 /dashboard 并提示         |
| Token 过期        | 后台静默 refresh；refresh 失败则 toast 提示「登录已过期」并跳转 /login |
| API 返回 403      | toast 显示「权限不足：需要 {required_role} 角色」                      |

### 6.4 其余页面交互规范
- 异步操作反馈：提交后立即显示 task_id，跳转任务详情或内嵌进度条
- 实时日志：任务运行中实时展示 stdout/stderr，支持自动滚动切换
- 危险操作确认：删除 Agent / 还原备份 / 删除 Skill 必须二次确认
- 状态刷新：Gateway 状态卡自动轮询（间隔 30s），其余页面手动刷新为主
- nvm 警告横幅：检测到 nvm Node 路径时，顶部橙色横幅提示（Operator+ 可见一键修复）

## 7. 数据模型（SQLite）
### 7.1 users 表（v1.1 新增）
> CREATE TABLE users (
>
> user_id TEXT PRIMARY KEY, -- UUID v4
>
> username TEXT NOT NULL UNIQUE, -- 3-32 字符，字母数字下划线
>
> password_hash TEXT NOT NULL, -- bcrypt hash, cost=12
>
> role TEXT NOT NULL DEFAULT 'Viewer', -- Viewer | Operator | Admin
>
> status TEXT NOT NULL DEFAULT 'active', -- active | disabled
>
> created_at TEXT NOT NULL,
>
> last_login_at TEXT,
>
> updated_at TEXT
>
> );

### 7.2 refresh_tokens 表（v1.1 新增）
> CREATE TABLE refresh_tokens (
>
> token_id TEXT PRIMARY KEY, -- UUID v4
>
> user_id TEXT NOT NULL REFERENCES users(user_id),
>
> token_hash TEXT NOT NULL UNIQUE, -- SHA-256(refresh_token)
>
> expires_at TEXT NOT NULL, -- 7 天后
>
> revoked INTEGER NOT NULL DEFAULT 0,
>
> created_at TEXT NOT NULL,
>
> user_agent TEXT, -- 记录来源客户端
>
> ip_address TEXT
>
> );

### 7.3 token_blacklist 表（v1.1 新增）
> CREATE TABLE token_blacklist (
>
> jti TEXT PRIMARY KEY, -- JWT jti claim
>
> expires_at TEXT NOT NULL, -- 与 JWT exp 一致，用于清理
>
> created_at TEXT NOT NULL
>
> );

### 7.4 tasks 表（v1.1 更新：新增 created_by）
> CREATE TABLE tasks (
>
> task_id TEXT PRIMARY KEY,
>
> task_type TEXT NOT NULL,
>
> status TEXT NOT NULL DEFAULT 'PENDING',
>
> request_json TEXT,
>
> exit_code INTEGER,
>
> stdout_tail TEXT,
>
> stderr_tail TEXT,
>
> log_path TEXT,
>
> created_by TEXT REFERENCES users(user_id), -- v1.1 新增
>
> created_at TEXT NOT NULL,
>
> started_at TEXT,
>
> finished_at TEXT
>
> );

### 7.5 revisions 表
> CREATE TABLE revisions (
>
> revision_id TEXT PRIMARY KEY,
>
> target_type TEXT NOT NULL,
>
> target_id TEXT,
>
> content TEXT NOT NULL,
>
> sha256 TEXT NOT NULL,
>
> created_at TEXT NOT NULL,
>
> created_by TEXT REFERENCES users(user_id) -- 关联用户
>
> );

### 7.6 backups 表
> CREATE TABLE backups (
>
> backup_id TEXT PRIMARY KEY,
>
> label TEXT,
>
> scope_json TEXT NOT NULL,
>
> manifest_path TEXT NOT NULL,
>
> size_bytes INTEGER,
>
> sha256 TEXT NOT NULL,
>
> verified INTEGER DEFAULT 0,
>
> created_by TEXT REFERENCES users(user_id), -- 关联用户
>
> created_at TEXT NOT NULL
>
> );

## 8. 非功能需求
| **类别**    | **要求**                                                                                                                                                 |
|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| 安全 - 认证 | JWT HS256（密钥 ≥ 32 字节，存于 config.toml，不进代码库）；AccessToken 有效期 15min；RefreshToken 有效期 7 天存于 HttpOnly Cookie；注销时 jti 加入黑名单 |
| 安全 - 密码 | bcrypt cost=12；不允许空密码；最短 8 位；登录失败连续 5 次锁定账号 15 分钟（可配置）                                                                     |
| 安全 - 权限 | 每个 API Handler 在业务逻辑前执行 RequireRole(role) 中间件；中间件失败立即返回 403，不执行后续逻辑                                                       |
| 安全 - 文件 | 所有文件操作路径白名单校验；上传文件限 100MB；zip-slip 防护                                                                                              |
| 可靠性      | 原子写入防配置损坏；备份前自动快照；任务超时自动终止子进程                                                                                               |
| 性能        | 日志流式传输；Overview 并发调用 CLI 命令；认证中间件不做数据库查询（JWT 自包含，仅黑名单用 DB）                                                          |
| 可观测性    | managerd 日志写入 ~/.openclaw-manager/manager.log，记录操作人（user_id）；审计日志：写操作（config.write、binding.apply 等）单独记录到 audit.log         |
| 可维护性    | managerd 以 systemd user service 管理；配置通过 config.toml 指定；JWT 密钥、invite_code 等敏感配置环境变量优先于 config.toml                             |

## 9. 开发计划（建议阶段划分）
| **阶段**                      | **内容**                                                                                     | **预估周期** |
|-------------------------------|----------------------------------------------------------------------------------------------|--------------|
| Phase 1 – 基础骨架 + 用户体系 | Go 框架 + SQLite 初始化 + 用户注册/登录/JWT + 角色中间件 + Vue 登录/注册页                   | 1.5 周       |
| Phase 2 – Gateway 核心        | Gateway start/stop/restart/status/logs API + Vue Gateway 页 + Dashboard 基础卡片（权限感知） | 1 周         |
| Phase 3 – 配置管理            | openclaw.json 编辑 + Revision 系统 + Identity 编辑器（Operator 鉴权）                        | 1 周         |
| Phase 4 – Agent & Binding     | Agent CRUD + Binding 可视化管理（Admin/Operator 分权）                                       | 1.5 周       |
| Phase 5 – Skills & Backup     | Skills install/remove + Backup create/restore/dry_run（Admin 还原）                          | 1.5 周       |
| Phase 6 – 用户管理 + 集成加固 | 用户管理页 + 角色修改/禁用 + Doctor 集成 + 审计日志 + 完整测试                               | 1 周         |

### 9.1 MVP 阶段最小交付范围
| **功能**                      | **MVP** | **P1** |
|-------------------------------|---------|--------|
| 用户注册（首位自动 Admin）    | ✓       |        |
| 用户登录 + JWT                | ✓       |        |
| 角色权限中间件（三级角色）    | ✓       |        |
| 用户管理页（Admin）           | ✓       |        |
| Token 刷新与注销              | ✓       |        |
| 邀请码注册限制                |         | ✓      |
| 登录失败锁定                  |         | ✓      |
| Gateway 启停/状态/日志        | ✓       |        |
| openclaw.json 编辑 + Revision | ✓       |        |
| 备份与还原（含 dry_run）      | ✓       |        |
| Agent 列表 + Binding 管理     | ✓       |        |
| Skills 安装与删除             | ✓       |        |
| 新建/删除 Agent               |         | ✓      |
| 审计日志                      |         | ✓      |

## 10. 用户体系与权限管理（v1.1 新增）
### 10.1 角色定义
系统采用三级固定角色模型（RBAC-Flat），简单、直观，满足小团队运维场景。

| **角色** | **英文标识** | **定位**                                                      | **典型人员**  |
|----------|--------------|---------------------------------------------------------------|---------------|
| 管理员   | Admin        | 最高权限，可管理用户、执行还原、创建/删除 Agent               | 系统负责人    |
| 操作员   | Operator     | 日常运维权限，可启停 Gateway、编辑配置、管理 Skills、创建备份 | 运维人员      |
| 观察者   | Viewer       | 只读权限，可查看所有状态、日志、配置（只读）、备份列表        | 审计/监控人员 |

> *⚠ 角色为线性包含关系：Admin ⊇ Operator ⊇ Viewer。高角色自动拥有低角色全部权限。*

### 10.2 权限矩阵（完整）
| **操作**                           | **Viewer** | **Operator** | **Admin** |
|------------------------------------|------------|--------------|-----------|
| 查看 Dashboard / Overview          | ✓          | ✓            | ✓         |
| 查看 Gateway 状态与日志            | ✓          | ✓            | ✓         |
| 启动 / 停止 / 重启 Gateway         | ✗          | ✓            | ✓         |
| 运行 Doctor / Doctor Repair        | ✗          | ✓            | ✓         |
| 读取 openclaw.json / Revision 列表 | ✓          | ✓            | ✓         |
| 编辑 openclaw.json / 还原 Revision | ✗          | ✓            | ✓         |
| 读取 / 编辑 Agent Identity         | ✗（读）✓   | ✓            | ✓         |
| 查看 Agent 列表                    | ✓          | ✓            | ✓         |
| 新建 / 删除 Agent                  | ✗          | ✗            | ✓         |
| 查看 Channel Binding               | ✓          | ✓            | ✓         |
| 新增 / 删除 Binding Rule           | ✗          | ✓            | ✓         |
| 查看 Skills 列表                   | ✓          | ✓            | ✓         |
| 安装 / 删除 Skill                  | ✗          | ✓            | ✓         |
| 查看备份列表与详情                 | ✓          | ✓            | ✓         |
| 创建备份 / 下载备份                | ✗          | ✓            | ✓         |
| 执行还原（含 dry_run）             | ✗          | ✗            | ✓         |
| 删除备份记录                       | ✗          | ✗            | ✓         |
| 查看任务列表与日志流               | ✓          | ✓            | ✓         |
| 取消任务                           | ✗          | ✓            | ✓         |
| 查看自身信息 / 修改自身密码        | ✓          | ✓            | ✓         |
| 查看用户列表                       | ✗          | ✗            | ✓         |
| 修改用户角色 / 禁用用户            | ✗          | ✗            | ✓         |
| 删除用户                           | ✗          | ✗            | ✓         |

注：Agent Identity 读取权限说明 — Viewer 可读取但不可修改（GET 接口返回 200，PUT 接口返回 403）。

### 10.3 用户注册流程
#### 10.3.1 首位用户（自动 Admin）
8.  访问系统，检测到 users 表为空，前端展示注册引导页
9.  用户填写用户名 + 密码，提交 POST /api/v1/auth/register
10. 后端检测 users 表为空，将该用户 role 设为 Admin
11. 注册成功，自动跳转登录页

#### 10.3.2 后续用户
12. 访问 /register（若系统配置允许公开注册）
13. 提交注册，后端默认分配 Viewer 角色
14. 注册成功后通知 Admin 提权（系统不主动提权）

若系统配置 invite_code_required: true，注册时须填写邀请码（Admin 在用户管理页生成，P1 功能）。若配置 public_registration: false，则 /register 接口返回 403，新用户只能由 Admin 在用户管理页创建。

> *⚠ 系统默认配置 public_registration: true，以降低首次部署门槛。生产环境建议设为 false。*

### 10.4 认证流程
#### 10.4.1 登录
15. POST /api/v1/auth/login，后端验证用户名密码（bcrypt 比对）
16. 验证通过：签发 AccessToken（JWT，15min，payload 含 user_id/role/jti）+ RefreshToken（UUID，7d）
17. RefreshToken 存储：后端 Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict，同时 SHA-256 哈希后存入 refresh_tokens 表
18. 前端接收 AccessToken 存入 Pinia/Vuex 内存，不存 localStorage
19. 前端后续请求携带 Authorization: Bearer {access_token}

#### 10.4.2 Token 刷新
20. AccessToken 过期（前端检测 exp 或收到 401）
21. 前端调用 POST /api/v1/auth/refresh（Cookie 自动携带 RefreshToken）
22. 后端验证 RefreshToken 有效且未撤销，签发新 AccessToken
23. 若 RefreshToken 也过期，返回 401，前端清除状态跳转 /login

#### 10.4.3 注销
24. 前端调用 POST /api/v1/auth/logout
25. 后端将当前 AccessToken 的 jti 写入 token_blacklist 表，撤销 RefreshToken（refresh_tokens.revoked=1）
26. 前端清除内存中的 AccessToken，跳转 /login

### 10.5 权限中间件（后端实现要求）
#### 10.5.1 中间件链
每个受保护路由按顺序执行以下中间件：

27. AuthMiddleware：从 Authorization 头提取并验证 JWT → 检查 jti 黑名单 → 将 User{user_id, role} 注入 context
28. RequireRole(minRole) 中间件：从 context 取 User，校验 role \>= minRole，不足则返回 403
29. 业务 Handler（此时 context 中已有可信的 User 信息）

#### 10.5.2 角色比较
角色权重定义（用于 \>= 比较）：

> Viewer = 1
>
> Operator = 2
>
> Admin = 3

RequireRole(minRole) 等价于：if user.RoleWeight() \< minRole.Weight() → 403

#### 10.5.3 示例路由注册（Go 伪代码）
> r.GET("/api/v1/overview", Auth(), RequireRole(Viewer), overviewHandler)
>
> r.POST("/api/v1/gateway/restart", Auth(), RequireRole(Operator), gatewayRestartHandler)
>
> r.POST("/api/v1/agents", Auth(), RequireRole(Admin), agentCreateHandler)
>
> r.POST("/api/v1/backups/:id/restore", Auth(), RequireRole(Admin), backupRestoreHandler)

### 10.6 账号安全规则
| **规则**                 | **默认值**  | **是否可配置**                              |
|--------------------------|-------------|---------------------------------------------|
| 密码最短长度             | 8 位        | 可（config.toml: auth.password_min_length） |
| 密码必须包含             | 字母 + 数字 | 可                                          |
| 登录失败锁定阈值         | 连续 5 次   | 可（P1 实现）                               |
| 锁定持续时间             | 15 分钟     | 可（P1 实现）                               |
| AccessToken 有效期       | 15 分钟     | 可（config.toml: auth.access_token_ttl）    |
| RefreshToken 有效期      | 7 天        | 可（config.toml: auth.refresh_token_ttl）   |
| JWT 签名算法             | HS256       | 不可更改（MVP）                             |
| 禁止自删/自降级（Admin） | 是          | 否                                          |
| 系统最少保留一个 Admin   | 是          | 否                                          |

> *⚠ 系统始终保持至少一个 Admin 账号存活。若 Admin 试图删除/降级自身，或删除后将导致无 Admin 存在，后端返回 400 并说明原因。*

### 10.7 config.toml 认证相关配置项
> \[auth\]
>
> jwt_secret = "" \# 必填，≥32 字节随机字符串，也可用环境变量 OPENCLAW_JWT_SECRET
>
> access_token_ttl = "15m" \# AccessToken 有效期
>
> refresh_token_ttl = "168h" \# RefreshToken 有效期（7 天）
>
> public_registration = true \# false 时禁止公开注册
>
> invite_code_required = false \# true 时注册需邀请码（P1）
>
> password_min_length = 8
>
> \[server\]
>
> listen = "127.0.0.1:18790" \# managerd 监听地址
>
> \[paths\]
>
> openclaw_home = "~/.openclaw"
>
> manager_home = "~/.openclaw-manager"

*文档结束 | v1.1*

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

- 完成 Issue-E5-S1-I29
-- 功能总结
实现 Agent Repository：调用 openclaw agents list --bindings --json 解析 Agent 列表；提供 GetWorkspacePath；加入 60s TTL 缓存；对 agent_id 做格式与路径穿越防护。
-- 涉及文件
src/internal/agent/repo.go,
src/internal/agent/repo_test.go

- 完成 Issue-E5-S1-I30
-- 功能总结
实现 Agent 列表与详情 API：GET /agents、GET /agents/{id}；支持空列表返回 []；详情不存在返回 404。
-- 涉及文件
src/internal/agent/handler.go,
src/internal/agent/handler_test.go

- 完成 Issue-E6-S1-I34
-- 功能总结
实现 Skills 列表 API：GET /skills?scope=global|agent；支持 global 与 agent scope 扫描；返回 name/scope/agent_id/size_bytes/has_meta 字段。
-- 涉及文件
src/internal/skills/handler.go,
src/internal/skills/handler_test.go

- 完成 Issue-E4-S1-I26
-- 功能总结
实现 Revision Repository：支持 Save/List/FindByID，保存时计算 SHA256；按 target_type/target_id 维度维护历史，自动裁剪仅保留最新 50 条。
-- 涉及文件
src/internal/config/revision_repo.go,
src/internal/config/revision_repo_test.go

- 完成 Issue-E4-S1-I27
-- 功能总结
实现 openclaw.json 管理 API：GET 读取、PUT 写入（JSON 校验+原子写入+保存 Revision）、GET revisions、POST restore（按 revision_id 恢复并再生成一条 revision）。
-- 涉及文件
src/internal/config/openclaw_json_handler.go,
src/internal/config/openclaw_json_handler_test.go

- 完成 Issue-E6-S1-I35
-- 功能总结
实现 Skills 删除 API：DELETE /skills/{name}，支持 global/agent 两种 scope；实现 name 参数校验、防路径穿越、白名单校验、不存在返回 404。
-- 涉及文件
src/internal/skills/delete_handler.go,
src/internal/skills/delete_handler_test.go

- 完成 Issue-E3-S1-I23
-- 功能总结
实现 Gateway API Handler：GET status、POST start/stop/restart；增加 gateway 生命周期操作互斥，冲突时返回 409 与 running_task_id。
-- 涉及文件
src/internal/gateway/api_handler.go,
src/internal/gateway/api_handler_test.go

- 完成 Issue-E4-S2-I28
-- 功能总结
实现 Agent Identity 读写 API：GET/PUT IDENTITY.md、GET revisions；写入采用原子写入，加入 1MB 内容限制与路径白名单校验。
-- 涉及文件
src/internal/config/identity_handler.go,
src/internal/config/identity_handler_test.go

- 完成 Issue-E3-S1-I22（补充收口）
-- 功能总结
深度状态查询模块完成收口并在本轮保持回归通过：并发聚合 systemctl 与 openclaw 状态，支持 NVMWarning 检测，openclaw 超时场景保留 systemctl 结果。
-- 涉及文件
src/internal/gateway/systemctl.go,
src/internal/gateway/systemctl_test.go

- 完成 Issue-E3-S3-I25
-- 功能总结
实现 Doctor Run/Repair API：POST /doctor/run 与 /doctor/repair；执行 openclaw doctor 命令并提供 5 分钟超时控制；解析输出中的 .nvm 路径风险标记。
-- 涉及文件
src/internal/gateway/doctor_handler.go,
src/internal/gateway/doctor_handler_test.go

- 完成 Issue-E5-S2-I33
-- 功能总结
实现 Binding 列表与批量应用 API：GET /bindings、POST /bindings/apply；支持 add/remove 批量执行，部分失败继续执行并汇总失败计数，最终状态可为 FAILED。
-- 涉及文件
src/internal/agent/binding_handler.go,
src/internal/agent/binding_handler_test.go

- 完成 Issue-E8-S2-I43
-- 功能总结
实现任务 SSE 日志流：GET /tasks/{id}/events；支持 token query/header 鉴权、stdout/stderr 回放、seq 递增与 done 事件输出。
-- 涉及文件
src/internal/task/sse_handler.go,
src/internal/task/sse_handler_test.go

- 完成 Issue-E5-S1-I31
-- 功能总结
实现 Agent 新建 API：POST /agents，校验 agent_id 合法性，调用 openclaw agents create，返回任务化响应。
-- 涉及文件
src/internal/agent/manage_handler.go,
src/internal/agent/manage_handler_test.go

- 完成 Issue-E5-S1-I32
-- 功能总结
实现 Agent 删除 API：DELETE /agents/{id}，先执行 unbind-all，再执行 delete，返回任务化响应。
-- 涉及文件
src/internal/agent/manage_handler.go,
src/internal/agent/manage_handler_test.go

- 完成 Issue-E6-S2-I36
-- 功能总结
实现 Skills 上传安装 API：支持 multipart 上传（100MB 限制），并使用 SafeExtract 解压安装到目标技能目录。
-- 涉及文件
src/internal/skills/install_handler.go,
src/internal/skills/install_handler_test.go

- 完成 Issue-E7-S1-I37
-- 功能总结
实现备份核心服务：按 scope 归档 tar.gz、计算 SHA-256、生成 manifest.json，并写入 backups 表。
-- 涉及文件
src/internal/backup/service.go,
src/internal/backup/service_test.go

- 完成 Issue-E7-S1-I38
-- 功能总结
实现备份 API：POST 创建、GET 列表、GET 详情、GET 下载、DELETE 删除。
-- 涉及文件
src/internal/backup/api_handler.go,
src/internal/backup/api_handler_test.go

- 完成 Issue-E7-S2-I39
-- 功能总结
实现还原服务 Restore：支持 dry_run 预览将覆盖文件列表；校验备份 SHA-256；正式还原前自动创建 pre-restore 快照，并执行归档解压恢复。
-- 涉及文件
src/internal/backup/restore.go,
src/internal/backup/restore_test.go

- 完成 Issue-E7-S2-I40
-- 功能总结
实现还原 API：POST /backups/{id}/restore；dry_run=true 返回同步预览（200），dry_run=false 返回异步风格响应（202）。
-- 涉及文件
src/internal/backup/api_handler.go,
src/internal/backup/restore_api_test.go

- 完成 Issue-E10-S2-I54
-- 功能总结
新增 systemd user service 部署模板与安装脚本：openclaw-manager.service + scripts/install.sh，支持 daemon-reload/enable/restart。
-- 涉及文件
openclaw-manager.service,
scripts/install.sh

- 完成 Issue-E10-S2-I55
-- 功能总结
补充 README 与 API 文档：新增项目部署说明、config.toml 示例以及 docs/openapi.yaml（OpenAPI 3.0 简版）。
-- 涉及文件
README.md,
docs/openapi.yaml

- 完成 Issue-E10-S1-I53
-- 功能总结
补充安全测试用例：路径穿越拒绝、JWT 篡改拒绝，形成基础安全回归测试。
-- 涉及文件
src/internal/security/security_test.go

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

# OpenClaw Manager

轻量 Web 管理平台（Go + Vue + SQLite），用于管理 OpenClaw Gateway 与 Agent。

## 架构

- 后端：Go 1.22+
- 前端：Vue 3 + Element Plus
- 数据：SQLite
- 认证：JWT + Refresh Token
- 日志：SSE

## 快速部署

1. 构建后端：
```bash
cd src
make build
```
2. 准备配置：

- `~/.openclaw-manager/config.toml`
- 设置 `OPENCLAW_JWT_SECRET`（>=32 字节）
3. 安装 systemd user service：
```bash
./scripts/install.sh
```
4. 查看状态：
```bash
systemctl --user status openclaw-manager.service
```

## config.toml 关键配置

```toml
[server]
listen = "127.0.0.1:18790"

[auth]
jwt_secret = "replace-with-strong-secret-32bytes-min"
access_token_ttl = "15m"
refresh_token_ttl = "168h"
public_registration = true
password_min_length = 8

[paths]
openclaw_home = "~/.openclaw"
manager_home = "~/.openclaw-manager"
```

## 核心功能及特色

- 安全修改openclaw.json，支持历史版本管理，安全回滚
- Agent快速创建，bot与agent快捷绑定
- 快速添加多个QQbot
- openclaw手动与自动备份
- 通过web方式执行openclaw相关命令
- 简单的用户体系

可以任意折腾你的小龙虾，主打一个改不死。
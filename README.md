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

## API 文档（简版 OpenAPI）

见：`docs/openapi.yaml`

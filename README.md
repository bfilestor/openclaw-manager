# OpenClaw Manager

轻量OpenClaw Gateway管理平台（Go + Vue + SQLite），用于管理和备份 OpenClaw Gateway 各种配置文件 与 Agent。

## 核心功能及特色

- 安全修改openclaw.json，agent各个配置文件（AGENTS.md/SOUL.md/）
- 支持修改文件历史版本管理、修改对比、安全回滚
- Agent快速创建，bot与agent快捷绑定
- openclaw配置手动与自动备份
- 通过web方式执行openclaw相关命令
- 快速添加多个QQbot
- 简单的多用户体系,支持多个用户使用一个openclaw
- token使用统计，总体费用预估，单用户费用预估
- 自主注册开关，用户与bot关联

## 20260317更新
- 新增龙虾守护功能，后台开启设置后，定时检查龙虾运行状态，如果失败则回滚配置并重启龙虾
- 新增龙虾升级功能，前台可升级龙虾，如果升级失败支持回滚

## 20260313更新
- 新增多用户支持：支持用户和bot绑定，一只龙虾多人食用
- token使用统计：支持总体token统计，分用户token用量统计
- 支持插件备份
- 管理员重设密码，入口 /resetpwd,使用reset_super_token作为超级

## TODO
- LLM API接入新增与修改
- 更多bot快速接入

可以任意折腾你的小龙虾，主打一个改不死。

## 架构

- 后端：Go 1.22+
- 前端：Vue 3 + Element Plus
- 数据：SQLite
- 认证：JWT + Refresh Token
- 日志：SSE

## 快速部署

1. 下载文件创建目录解压文件：

```bash
mkdir ~/.openclaw-manager
tar -xzf openclaw-manager-xxxx.tar.gz -C ~/.openclaw-manager --strip-components=1
cd ~/.openclaw-manager
```
2. 准备配置：

- `~/.openclaw-manager/config.toml`
- 设置 `jwt_secret`,`reset_super_token`（>=32 字节）
3. 安装 systemd user service：

```bash
./scripts/install.sh
```
4. 查看状态：
```bash
systemctl --user status openclaw-manager.service
```
5. 重启服务
```
systemctl --user restart openclaw-manager.service
```

## config.toml 关键配置

```toml
[server]
listen = "0.0.0.0:18799"

[auth]
jwt_secret = "replace-with-strong-secret-32bytes-min"
reset_super_token = "replace-with-another-strong-secret-32bytes-min"
access_token_ttl = "15m"
refresh_token_ttl = "168h"
public_registration = true
password_min_length = 8

[paths]
openclaw_home = "~/.openclaw"
manager_home = "~/.openclaw-manager"
```

## Windows / macOS 预览支持

> 当前处于跨平台适配阶段（持续完善中），建议优先在 Linux/WSL2 使用完整功能。

Windows 本地可先使用以下脚本完成最小运行：

```powershell
# 1) 构建（前后端）
powershell -ExecutionPolicy Bypass -File .\scripts\build-windows.ps1 -All

# 2) 安装/更新服务
powershell -ExecutionPolicy Bypass -File .\scripts\install-service.ps1

# 3) 卸载服务（可选）
powershell -ExecutionPolicy Bypass -File .\scripts\uninstall-service.ps1
```

macOS 本地可先使用以下脚本完成最小运行：

```bash
# 1) 构建（前后端）
bash ./scripts/build-macos.sh --all

# 2) 安装/更新 launchd 服务
bash ./scripts/install-launchd.sh

# 3) 卸载 launchd 服务（可选）
bash ./scripts/uninstall-launchd.sh
```

## Tauri 桌面客户端（Windows/macOS 最小版）

已提供最小可用 Tauri 桌面壳：`desktop/tauri`。

打包前准备资源：

```bash
# Linux/macOS
bash ./desktop/tauri/scripts/prepare-assets.sh

# Windows PowerShell
powershell -ExecutionPolicy Bypass -File .\desktop\tauri\scripts\prepare-assets.ps1
```

构建桌面应用：

```bash
cd desktop/tauri
pnpm install
pnpm run build
```

> 桌面应用会尝试拉起 `managerd`，并加载本地 `http://127.0.0.1:18799` 管理面板。
> 当前桌面最小版**不集成 openclaw CLI**，仅打包 managerd 与前端资源。
> 关闭窗口时会最小化到系统托盘，可通过托盘菜单恢复窗口或退出应用。

## 注意事项
- reset_super_token是重设管理员密码的重要凭证，请妥善保管。一旦泄露，危害巨大，请定期修改！
- Linux 仍是当前最稳定运行环境，Windows/macOS 功能正在逐步对齐。
- 用户角色四种 admin/operator/viewer/user,admin全部权限，viewer只有查看权限，无修改权限,user普通使用者，只用户统计token使用情况
- 第一次运行，注册的第一个用户默认为管理员权限。后续注册用户默认为普通使用者

## 截图
### 主页
![Main UI](screenshots/main.png)

### 配置页
![Config Page](screenshots/config-2.png)

### QQBot
![QQBot](screenshots/QQBot.png)

### 多用户和token 预估
![token](screenshots/token.png)
![普通用户界面](screenshots/ordinal-user.png)

### 沟通群
<img src="screenshots/qq.jpg" width="200"/>

## License

本项目采用 **非商业许可协议（NCL）**。

允许：
- 个人学习
- 研究用途
- 非商业项目使用
- 修改与二次开发

禁止：
- 商业产品使用
- SaaS 服务
- 企业内部商业系统
- 未授权商业盈利

商业授权请联系作者：<bfilestor@gmail.com>

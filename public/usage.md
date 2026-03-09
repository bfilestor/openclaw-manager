# OpenClaw Manager 部署使用说明

本文对应 `openclaw-manager-*.tar.gz` 发布包。

## 1) 获取并解压

```bash
cd /tmp
# 把 openclaw-manager-*.tar.gz 下载到当前目录后执行
tar -xzf openclaw-manager-*.tar.gz
cd openclaw-manager
```

解压后目录结构：

- `bin/managerd`
- `web/dist/`
- `config/config.toml`
- `scripts/install.sh`
- `service/openclaw-manager.service`

## 2) 需要修改的文件（部署前）

请优先修改：

### `config/config.toml`

根据实际环境调整（示例）：

```toml
# 监听地址（示例）
listen = "0.0.0.0:18080"

# OpenClaw 配置目录（示例）
openclaw_config = "/home/<user>/.openclaw"

# 数据目录（示例）
data_dir = "/var/lib/openclaw-manager"
```

> 说明：字段名以你包内现有 `config.toml` 为准，按实际主机路径与端口改。

### `service/openclaw-manager.service`

至少检查并按需修改：

- `User=`：运行服务的系统用户
- `WorkingDirectory=`：部署目录
- `ExecStart=`：`managerd` 的绝对路径
- `Environment=`：必要环境变量

## 3) 快速安装（推荐）

```bash
chmod +x scripts/install.sh
sudo ./scripts/install.sh
```

若安装脚本包含 systemd 操作，完成后可验证：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now openclaw-manager
sudo systemctl status openclaw-manager --no-pager
```

## 4) 手动部署（备用）

如果不使用安装脚本，可手动执行：

```bash
# 1) 拷贝文件到目标目录
sudo mkdir -p /opt/openclaw-manager
sudo cp -r bin web config scripts /opt/openclaw-manager/

# 2) 安装 systemd service
sudo cp service/openclaw-manager.service /etc/systemd/system/

# 3) 启动
sudo systemctl daemon-reload
sudo systemctl enable --now openclaw-manager
```

## 5) 验证与排障

```bash
# 查看服务状态
sudo systemctl status openclaw-manager --no-pager

# 查看实时日志
journalctl -u openclaw-manager -f
```

如果启动失败，优先检查：

1. `ExecStart` 路径是否正确
2. `config.toml` 路径与权限
3. 端口是否被占用
4. 服务运行用户是否有读写权限

## 6) 后续重新打包

在项目根目录执行：

```bash
./script/publish.sh
```

产物会自动放到 `public/`：

- `openclaw-manager-YYYYmmdd-HHMMSS.tar.gz`
- `openclaw-manager-YYYYmmdd-HHMMSS.tar.gz.sha256`

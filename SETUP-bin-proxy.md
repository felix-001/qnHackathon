# bin-proxy 快速部署指南

## 1. bin-manager 端配置

### 1.1 创建二进制文件存储目录

```bash
mkdir -p ./bins
chmod 755 ./bins
```

### 1.2 配置 manager.json

编辑 `internal/config/manager.json`，确保包含 `binDir` 配置：

```json
{
  "binDir": "./bins",
  ...
}
```

### 1.3 启动 bin-manager

```bash
go run cmd/manager/main.go -f internal/config/manager.json
```

服务将在 `http://localhost:8081` 启动。

### 1.4 部署二进制文件

将需要分发的二进制文件放入 `./bins` 目录：

```bash
cp /path/to/your/app ./bins/your-app
chmod +x ./bins/your-app
```

## 2. 节点端 (bin-proxy) 配置

### 2.1 安装依赖

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y curl jq supervisor

# CentOS/RHEL
sudo yum install -y curl jq supervisor
```

### 2.2 部署 bin-proxy

```bash
# 复制脚本
sudo cp bin-proxy.sh /usr/local/bin/bin-proxy
sudo chmod +x /usr/local/bin/bin-proxy

# 创建配置目录
sudo mkdir -p /etc/bin-proxy

# 创建日志目录
sudo mkdir -p /var/log
sudo touch /var/log/bin-proxy.log
```

### 2.3 配置 bin-manifests.json

创建 `/etc/bin-proxy/bin-manifests.json`：

```json
{
  "node": {
    "name": "web-server-01",
    "cpuArch": "x86_64",
    "osRelease": "Ubuntu 20.04.6 LTS",
    "proxyVersion": "1.0.0"
  },
  "binaries": [
    {
      "name": "your-app",
      "version": "latest",
      "hash": "",
      "path": "/usr/local/bin/your-app"
    }
  ]
}
```

**注意**: 
- `name`: 必须与 bin-manager 的 `./bins/` 目录中的文件名一致
- `hash`: 初始值可以为空字符串，首次运行会自动更新
- `path`: 二进制文件在本地的安装路径

### 2.4 配置环境变量

编辑 `/etc/environment` 或在脚本中直接修改：

```bash
export BIN_MANAGER_API="http://your-manager-host:8081/api/v1"
export BIN_DOWNLOAD_URL="http://your-manager-host:8081/api/v1/download"
```

### 2.5 配置 supervisor (可选但推荐)

创建 `/etc/supervisor/conf.d/your-app.conf`：

```ini
[program:your-app]
command=/usr/local/bin/your-app
directory=/var/lib/your-app
user=www-data
autostart=true
autorestart=true
stderr_logfile=/var/log/your-app.err.log
stdout_logfile=/var/log/your-app.out.log
```

重启 supervisor：

```bash
sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl start your-app
```

### 2.6 配置 crontab

```bash
# 编辑 root 用户的 crontab
sudo crontab -e

# 添加以下行，每分钟执行一次
* * * * * BIN_MANAGER_API="http://your-manager-host:8081/api/v1" BIN_DOWNLOAD_URL="http://your-manager-host:8081/api/v1/download" /usr/local/bin/bin-proxy >> /var/log/bin-proxy.log 2>&1
```

## 3. 验证部署

### 3.1 手动测试 bin-proxy

```bash
sudo BIN_MANAGER_API="http://your-manager-host:8081/api/v1" \
     BIN_DOWNLOAD_URL="http://your-manager-host:8081/api/v1/download" \
     /usr/local/bin/bin-proxy
```

检查输出和日志：

```bash
tail -f /var/log/bin-proxy.log
```

### 3.2 验证 API 连接

```bash
# 测试 keepalive
curl -v "http://your-manager-host:8081/api/v1/keepalive?node=$(hostname)"

# 测试获取 hash
curl -v "http://your-manager-host:8081/api/v1/bins/your-app"

# 测试下载
curl -v "http://your-manager-host:8081/api/v1/download/your-app" -o /tmp/test-download
```

### 3.3 验证自动升级

1. 在 bin-manager 端更新二进制文件：

```bash
cp /path/to/new-version/your-app ./bins/your-app
```

2. 等待最多 1 分钟，bin-proxy 会自动检测并升级

3. 查看日志确认升级：

```bash
tail -f /var/log/bin-proxy.log
```

4. 检查服务状态：

```bash
sudo supervisorctl status your-app
```

## 4. 故障排查

### 4.1 bin-proxy 未执行

检查 crontab：

```bash
sudo crontab -l | grep bin-proxy
```

检查 cron 日志：

```bash
sudo tail -f /var/log/syslog | grep CRON
```

### 4.2 无法连接 API

检查网络：

```bash
telnet your-manager-host 8081
curl -v http://your-manager-host:8081/api/v1/keepalive?node=test
```

检查防火墙：

```bash
sudo ufw status
sudo iptables -L
```

### 4.3 下载失败

检查权限：

```bash
ls -la /usr/local/bin/
ls -la /tmp/
```

检查磁盘空间：

```bash
df -h
```

### 4.4 服务重启失败

检查 supervisor：

```bash
sudo supervisorctl status
sudo tail -f /var/log/supervisor/supervisord.log
```

检查服务配置：

```bash
sudo supervisorctl avail
```

## 5. 最佳实践

### 5.1 安全

- 使用 HTTPS 连接 bin-manager API
- 为 API 添加认证机制
- 限制 bin-proxy 运行权限
- 定期审查日志

### 5.2 监控

- 监控 `/var/log/bin-proxy.log` 的错误
- 设置告警规则
- 记录升级历史

### 5.3 回滚

bin-proxy 会自动备份旧版本到 `<bin-path>.backup`，需要回滚时：

```bash
sudo cp /usr/local/bin/your-app.backup /usr/local/bin/your-app
sudo supervisorctl restart your-app
```

### 5.4 多环境部署

不同环境使用不同的 bin-manager API：

```bash
# 开发环境
BIN_MANAGER_API="http://dev-manager:8081/api/v1"

# 生产环境
BIN_MANAGER_API="http://prod-manager:8081/api/v1"
```

## 6. 常见问题

### Q1: 如何管理多个二进制文件？

在 `bin-manifests.json` 中添加多个条目：

```json
{
  "binaries": [
    {"name": "app1", "version": "latest", "hash": "", "path": "/usr/local/bin/app1"},
    {"name": "app2", "version": "latest", "hash": "", "path": "/usr/local/bin/app2"},
    {"name": "app3", "version": "latest", "hash": "", "path": "/usr/local/bin/app3"}
  ]
}
```

### Q2: 如何跳过某个版本的升级？

临时方案：在 bin-manager 端保留旧版本，或在节点上手动修改 hash 值。

### Q3: 升级失败后如何处理？

1. 查看日志定位问题
2. 使用备份文件回滚
3. 修复问题后等待下次自动升级

### Q4: 如何批量部署到多台机器？

使用配置管理工具（如 Ansible）：

```yaml
- hosts: all
  tasks:
    - name: Deploy bin-proxy
      copy:
        src: bin-proxy.sh
        dest: /usr/local/bin/bin-proxy
        mode: '0755'
    
    - name: Configure bin-manifests
      template:
        src: bin-manifests.json.j2
        dest: /etc/bin-proxy/bin-manifests.json
    
    - name: Setup crontab
      cron:
        name: "bin-proxy"
        job: "/usr/local/bin/bin-proxy >> /var/log/bin-proxy.log 2>&1"
```

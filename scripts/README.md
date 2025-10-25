# bin-proxy

bin-proxy 是一个用于自动维护和更新宿主机二进制文件的工具。

## 功能特性

1. 基于 crontab 定时执行（每分钟一次）
2. 与 bin-manager API 交互获取最新二进制文件的 MD5 哈希值
3. 维护 bin-manifests 文件，记录需要管理的二进制文件列表
4. 自动对比 MD5 值并执行升级
5. 升级后通过 supervisor 自动重启服务
6. 支持自动回滚功能

## 前置要求

- `curl` - 用于 API 请求
- `jq` - 用于 JSON 处理
- `md5sum` - 用于计算 MD5 哈希值
- `supervisorctl` - 用于服务管理（可选）

## 安装步骤

### 1. 安装依赖

```bash
# Ubuntu/Debian
sudo apt-get install -y curl jq coreutils supervisor

# CentOS/RHEL
sudo yum install -y curl jq coreutils supervisor
```

### 2. 配置脚本

复制脚本到系统目录：

```bash
sudo cp bin-proxy.sh /usr/local/sbin/
sudo chmod +x /usr/local/sbin/bin-proxy.sh
```

创建配置目录并复制配置文件：

```bash
sudo mkdir -p /etc/bin-proxy
sudo cp bin-manifests.json /etc/bin-proxy/
```

### 3. 配置环境变量

创建配置文件 `/etc/bin-proxy/config`：

```bash
# bin-manager API 地址
export BIN_MANAGER_API="http://your-bin-manager-host:8080/api/v1"

# 二进制文件安装目录
export BIN_DIR="/usr/local/bin"

# manifests 文件路径
export BIN_MANIFESTS="/etc/bin-proxy/bin-manifests.json"

# 日志文件路径
export LOG_FILE="/var/log/bin-proxy.log"
```

### 4. 配置 bin-manifests.json

编辑 `/etc/bin-proxy/bin-manifests.json`，添加需要管理的二进制文件：

```json
{
  "binaries": [
    {
      "name": "manager",
      "version": "latest",
      "currentMd5": ""
    },
    {
      "name": "your-service-name",
      "version": "latest",
      "currentMd5": ""
    }
  ]
}
```

### 5. 配置 crontab

添加 crontab 任务，每分钟执行一次：

```bash
sudo crontab -e
```

添加以下内容：

```cron
# bin-proxy: 每分钟检查并更新二进制文件
* * * * * . /etc/bin-proxy/config && /usr/local/sbin/bin-proxy.sh
```

### 6. 配置 Supervisor（可选）

如果使用 supervisor 管理服务，需要为每个服务创建配置文件。

示例 `/etc/supervisor/conf.d/manager.conf`：

```ini
[program:manager]
command=/usr/local/bin/manager
directory=/opt/app
user=appuser
autostart=true
autorestart=true
stderr_logfile=/var/log/manager.err.log
stdout_logfile=/var/log/manager.out.log
```

重新加载 supervisor 配置：

```bash
sudo supervisorctl reread
sudo supervisorctl update
```

## 使用方法

### 手动执行

```bash
# 使用默认配置
sudo /usr/local/sbin/bin-proxy.sh

# 使用自定义配置
sudo BIN_MANAGER_API="http://custom-host:8080/api/v1" \
     BIN_MANIFESTS="/custom/path/bin-manifests.json" \
     /usr/local/sbin/bin-proxy.sh
```

### 查看日志

```bash
sudo tail -f /var/log/bin-proxy.log
```

## API 接口要求

bin-proxy 需要 bin-manager 提供以下 API 接口：

1. **获取最新版本 MD5**
   ```
   GET /api/v1/releases/latest/{bin_name}/md5
   
   Response:
   {
     "code": 0,
     "message": "success",
     "data": {
       "md5": "5d41402abc4b2a76b9719d911017c592"
     }
   }
   ```

2. **下载最新版本二进制**
   ```
   GET /api/v1/releases/latest/{bin_name}/download
   
   Response: 二进制文件流
   ```

## 工作流程

1. bin-proxy 每分钟通过 crontab 执行
2. 读取 bin-manifests.json 获取需要管理的二进制列表
3. 对每个二进制文件：
   - 调用 API 获取最新版本的 MD5 值
   - 与当前运行版本的 MD5 对比
   - 如果不同，下载新版本
   - 验证下载文件的 MD5
   - 备份当前版本
   - 替换二进制文件
   - 通过 supervisor 重启服务
   - 更新 bin-manifests.json 中的 MD5 值
4. 如果重启失败，自动回滚到备份版本

## 故障排查

### 查看日志

```bash
sudo tail -f /var/log/bin-proxy.log
```

### 测试 API 连接

```bash
curl -s "http://your-bin-manager-host:8080/api/v1/releases/latest/manager/md5"
```

### 检查 crontab 是否运行

```bash
sudo grep CRON /var/log/syslog | grep bin-proxy
```

### 验证文件权限

```bash
ls -la /usr/local/sbin/bin-proxy.sh
ls -la /etc/bin-proxy/
```

## 安全建议

1. 使用 HTTPS 连接 bin-manager API
2. 限制脚本执行权限（仅 root 用户）
3. 定期审计日志文件
4. 备份重要的二进制文件
5. 实施 API 认证机制

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT

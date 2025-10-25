# bin-proxy

bin-proxy 是一个用于自动维护和更新宿主机二进制文件的工具。

## 功能特性

1. 基于 crontab 定时执行（每分钟一次）
2. 与 bin-manager API 交互获取最新二进制文件的 SHA-256 哈希值
3. 维护 bin-manifests 文件，记录需要管理的二进制文件列表
4. 自动对比 SHA-256 值并执行升级
5. 升级后通过 supervisor 自动重启服务
6. 支持自动回滚功能
7. **文件锁机制** - 防止多个实例同时处理同一任务
8. **进程管理** - 自动清理旧的下载进程，避免版本冲突
9. **进度追踪** - 实时上报任务处理时间到 API
10. **节点信息管理** - 自动收集和上报节点信息（CPU架构、系统版本等）
11. **Keepalive 机制** - 定期向 API 上报节点状态

## 前置要求

- `curl` - 用于 API 请求
- `jq` - 用于 JSON 处理
- `sha256sum` - 用于计算 SHA-256 哈希值
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

# 锁文件目录
export LOCK_DIR="/var/run/bin-proxy"

# 下载基础 URL（可选，默认使用 API 路径）
export DOWNLOAD_BASE_URL="http://your-bin-manager-host:8080/api/v1/download"
```

创建锁文件目录：

```bash
sudo mkdir -p /var/run/bin-proxy
```

### 4. 配置 bin-manifests.json

编辑 `/etc/bin-proxy/bin-manifests.json`，添加需要管理的二进制文件：

```json
{
  "binaries": [
    {
      "serviceName": "manager",
      "binaryName": "manager",
      "version": "",
      "previousVersion": "",
      "port": 8080
    },
    {
      "serviceName": "your-service",
      "binaryName": "your-service-name",
      "version": "",
      "previousVersion": "",
      "port": 9090
    }
  ],
  "nodeInfo": {
    "cpuArch": "",
    "osRelease": "",
    "nodeName": "",
    "binProxyVersion": ""
  }
}
```

注意：`nodeInfo` 部分会在脚本首次运行时自动填充。

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

### 1. Keepalive 接口

**检查节点状态**
```
GET /api/v1/keepalive

Response:
{
  "code": 0,
  "message": "success",
  "data": {
    "nodeName": "host-001",
    "status": "active"
  }
}
```

**注册/更新节点信息**
```
POST /api/v1/keepalive
Content-Type: application/json

Request Body:
{
  "cpuArch": "x86_64",
  "osRelease": "Ubuntu 20.04.3 LTS",
  "nodeName": "host-001",
  "binProxyVersion": "1.1.0"
}

Response:
{
  "code": 0,
  "message": "success"
}
```

### 2. 二进制信息接口

**获取最新版本 SHA-256**
```
GET /api/v1/bins/{bin_name}

Response:
{
  "code": 0,
  "message": "success",
  "data": {
    "sha256": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
    "version": "latest"
  }
}
```

**上报更新后的版本信息**
```
POST /api/v1/bins/{bin_name}
Content-Type: application/json

Request Body:
{
  "nodeName": "host-001",
  "binName": "manager",
  "sha256": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
  "version": "latest"
}

Response:
{
  "code": 0,
  "message": "success"
}
```

### 3. 进度上报接口

**上报任务处理进度**
```
POST /api/v1/bins/{bin_name}/progress
Content-Type: application/json

Request Body:
{
  "nodeName": "host-001",
  "binName": "manager",
  "targetHash": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
  "processingTime": 45,
  "status": "in_progress"  // 可选值: in_progress, success, failed
}

Response:
{
  "code": 0,
  "message": "success"
}
```

说明：
- `targetHash`: 目标版本的 SHA-256 哈希值，用于标识正在处理的具体版本
- `processingTime`: 从获取锁开始到当前的处理时间（秒）
- `status`: 任务状态
  - `in_progress`: 正在处理中（下载/安装阶段）
  - `success`: 成功完成
  - `failed`: 处理失败

### 4. 下载接口

**下载二进制文件**
```
GET /api/v1/download/{bin_name}

Response: 二进制文件流
```

**备用下载接口**
```
GET /api/v1/releases/latest/{bin_name}/download

Response: 二进制文件流
```

## 工作流程

1. **初始化阶段**
   - bin-proxy 每分钟通过 crontab 执行
   - 创建锁文件目录
   - 更新 bin-manifests.json 中的节点信息
   - 向 API 发送 keepalive 请求（检查/注册节点）

2. **处理每个二进制文件**
   - 读取 bin-manifests.json 获取需要管理的二进制列表
   - 对每个二进制文件：
     
     a. **获取锁**
        - 尝试获取该二进制的文件锁
        - 如果锁已存在且未超时（10分钟），跳过处理
        - 如果锁已超时，清理旧锁并继续
     
     b. **版本检查**
        - 调用 API 获取最新版本的 SHA-256 值
        - 与当前运行版本的 SHA-256 对比
        - 如果相同，释放锁并跳过
     
     c. **下载更新**（如果需要）
        - 清理该二进制的旧下载进程（避免并发下载冲突）
        - 上报处理进度（in_progress 状态）
        - 下载新版本二进制文件
        - 验证下载文件的 SHA-256
     
     d. **应用更新**
        - 备份当前版本
        - 替换二进制文件
        - 通过 supervisor 重启服务
        - 如果重启失败，自动回滚到备份版本
     
     e. **上报结果**
        - 上报任务完成状态和处理时间（success/failed）
        - 向 API POST 更新后的版本信息
        - 更新 bin-manifests.json 中的 SHA-256 值
        - 释放文件锁

3. **并发控制机制**
   - **版本特定的文件锁** - 锁文件命名格式为 `{bin_name}-{hash}.lock`，不同版本使用不同的锁文件
   - **自动清理旧版本锁** - 获取新版本锁时，自动清理该二进制的旧版本锁文件
   - **进程清理** - 清理该二进制的旧下载进程，防止多版本并发下载冲突
   - **锁超时机制** - 10分钟超时，防止死锁
   - **实时进度上报** - 基于锁文件创建时间计算处理时长并上报给 API

## 故障排查

### 查看日志

```bash
sudo tail -f /var/log/bin-proxy.log
```

### 测试 API 连接

```bash
curl -s "http://your-bin-manager-host:8080/api/v1/releases/latest/manager/sha256"
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

### 检查锁文件状态

```bash
ls -la /var/run/bin-proxy/
```

查看锁文件内容（Unix 时间戳）：
```bash
cat /var/run/bin-proxy/manager-abc123def456.lock
```

手动清理锁文件（如果需要）：
```bash
sudo rm /var/run/bin-proxy/*.lock
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

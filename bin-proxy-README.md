# bin-proxy 使用文档

## 概述

bin-proxy 是一个自动化的二进制文件升级工具，通过与 bin-manager API 交互，自动检测并升级宿主机上的二进制文件。

## 功能特性

1. **自动版本检测**: 定期检查二进制文件的 MD5 hash 值，与远程版本对比
2. **自动下载升级**: 发现新版本时自动下载并替换本地文件
3. **服务重启**: 升级完成后通过 supervisor 自动重启服务
4. **节点注册**: 自动向 bin-manager 注册节点信息
5. **配置化管理**: 通过 bin-manifests.json 配置需要管理的二进制文件

## 系统要求

- Bash shell
- curl
- jq (JSON 处理工具)
- supervisorctl (可选，用于服务重启)
- md5sum

## 安装步骤

### 1. 安装依赖

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y curl jq supervisor

# CentOS/RHEL
sudo yum install -y curl jq supervisor
```

### 2. 安装 bin-proxy

```bash
# 复制脚本到系统目录
sudo cp bin-proxy.sh /usr/local/bin/bin-proxy
sudo chmod +x /usr/local/bin/bin-proxy

# 创建配置目录
sudo mkdir -p /etc/bin-proxy
```

### 3. 配置 bin-manifests.json

```bash
# 复制示例配置文件
sudo cp bin-manifests.json.example /etc/bin-proxy/bin-manifests.json

# 编辑配置文件，添加需要管理的二进制文件
sudo vim /etc/bin-proxy/bin-manifests.json
```

配置文件示例：

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
      "name": "myapp",
      "version": "latest",
      "hash": "d41d8cd98f00b204e9800998ecf8427e",
      "path": "/usr/local/bin/myapp"
    }
  ]
}
```

### 4. 配置环境变量（可选）

```bash
# 在 /etc/environment 或 ~/.bashrc 中添加
export BIN_MANAGER_API="http://your-manager-api:8081/api/v1"
export BIN_DOWNLOAD_URL="http://your-manager-api:8081/api/v1/download"
```

### 5. 配置 crontab

```bash
# 编辑 crontab
sudo crontab -e

# 添加以下行，每分钟执行一次
* * * * * /usr/local/bin/bin-proxy >> /var/log/bin-proxy.log 2>&1
```

## bin-manager API 端点

bin-proxy 依赖以下 bin-manager API 端点：

### 1. Keepalive - 节点心跳

**GET** `/api/v1/keepalive?node=<node-name>`

查询节点信息。如果节点不存在，返回 404。

**POST** `/api/v1/keepalive?node=<node-name>`

注册新节点。请求体：

```json
{
  "name": "node-name",
  "cpuArch": "x86_64",
  "osRelease": "Ubuntu 20.04",
  "proxyVersion": "1.0.0",
  "binaries": [...],
  "lastSeen": "2024-01-01T00:00:00Z"
}
```

### 2. 获取二进制 Hash

**GET** `/api/v1/bins/<bin-name>`

返回示例：

```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "name": "myapp",
    "version": "latest",
    "hash": "abc123..."
  }
}
```

### 3. 更新二进制 Hash

**POST** `/api/v1/bins/<bin-name>`

请求体：

```json
{
  "hash": "abc123..."
}
```

### 4. 下载二进制文件

**GET** `/api/v1/download/<bin-name>`

直接返回二进制文件内容。

## 工作流程

1. **启动**: bin-proxy 通过 crontab 每分钟执行一次
2. **心跳**: 向 bin-manager 发送 keepalive 请求，首次运行会注册节点
3. **读取配置**: 从 bin-manifests.json 读取需要管理的二进制列表
4. **检查更新**: 
   - 对每个二进制文件，向 API 查询最新的 hash 值
   - 与本地配置的 hash 值对比
5. **执行升级**:
   - 如果 hash 不同，下载新版本到临时文件
   - 验证下载文件的 hash 值
   - 备份旧文件
   - 替换为新文件
   - 通过 supervisor 重启服务
6. **同步状态**:
   - 向 API 报告新的 hash 值
   - 更新本地 bin-manifests.json

## 日志

日志文件位置：`/var/log/bin-proxy.log`

日志包含：
- 执行时间戳
- 检查和升级操作
- 下载进度
- 错误信息

查看日志：

```bash
tail -f /var/log/bin-proxy.log
```

## 故障排查

### 1. bin-proxy 没有运行

检查 crontab 配置：

```bash
sudo crontab -l
```

检查日志：

```bash
tail -100 /var/log/bin-proxy.log
```

### 2. 无法连接 bin-manager API

检查网络连接：

```bash
curl -v http://your-manager-api:8081/api/v1/keepalive?node=test
```

检查环境变量：

```bash
echo $BIN_MANAGER_API
```

### 3. 下载失败

检查权限：

```bash
ls -l /usr/local/bin/
```

检查磁盘空间：

```bash
df -h /tmp
```

### 4. 服务重启失败

检查 supervisor 状态：

```bash
sudo supervisorctl status
```

确保服务已在 supervisor 中配置：

```bash
ls /etc/supervisor/conf.d/
```

## 安全建议

1. **使用 HTTPS**: 在生产环境中，bin-manager API 应使用 HTTPS
2. **认证**: 考虑为 API 添加认证机制
3. **权限控制**: bin-proxy 脚本应以适当的用户权限运行
4. **备份**: 升级前自动备份旧版本，出问题时可以快速回滚
5. **Hash 验证**: 始终验证下载文件的 hash 值

## 扩展功能

### 自定义下载 URL

可以为每个二进制文件指定不同的下载源：

```json
{
  "binaries": [
    {
      "name": "myapp",
      "version": "latest",
      "hash": "",
      "path": "/usr/local/bin/myapp",
      "downloadUrl": "http://custom-server/myapp"
    }
  ]
}
```

### 钩子脚本

可以在升级前后执行自定义脚本：

```bash
# 在 bin-proxy.sh 中添加钩子函数
pre_upgrade_hook() {
    # 升级前执行
}

post_upgrade_hook() {
    # 升级后执行
}
```

## bin-manager 配置

在 bin-manager 的配置文件中设置二进制文件存储目录：

```json
{
  "binDir": "/var/lib/bin-manager/bins",
  ...
}
```

确保该目录存在并有适当的权限：

```bash
sudo mkdir -p /var/lib/bin-manager/bins
sudo chmod 755 /var/lib/bin-manager/bins
```

## 维护

### 更新 bin-proxy 自身

```bash
# 下载新版本
sudo cp new-bin-proxy.sh /usr/local/bin/bin-proxy
sudo chmod +x /usr/local/bin/bin-proxy
```

### 添加新的二进制文件

编辑 `/etc/bin-proxy/bin-manifests.json`，添加新条目，下次执行时会自动处理。

### 删除二进制文件管理

从 `bin-manifests.json` 中移除对应条目即可。

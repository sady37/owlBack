# wisefido-radar 服务管理脚本

## 启动脚本
- `./start-radar.sh` - 启动 wisefido-radar 服务

## 停止脚本
- `./stop-radar.sh` - 停止 wisefido-radar 服务

## 停止脚本选项

### 基本用法
```bash
# 优雅停止（默认）
./stop-radar.sh

# 强制停止
./stop-radar.sh --force
./stop-radar.sh -f

# 显示帮助
./stop-radar.sh --help
./stop-radar.sh -h

# 详细输出
./stop-radar.sh --verbose
./stop-radar.sh -v
```

### 停止模式说明

#### 1. 优雅停止（默认）
- 发送 SIGTERM 信号给进程
- 给进程时间进行清理工作
- 等待最多10秒让进程自行退出
- 如果进程没有退出，会自动转为强制停止

#### 2. 强制停止
- 直接发送 SIGKILL (kill -9) 信号
- 立即终止进程
- 清理所有相关进程和端口占用

## 脚本功能

### 停止脚本 (`stop-radar.sh`) 功能：
1. **检查运行状态**：检查 wisefido-radar 服务和 decode-track 工具是否在运行
2. **多种停止模式**：支持优雅停止和强制停止
3. **端口清理**：清理 HTTPS 端口 8443 的占用
4. **验证停止**：验证所有相关进程是否已停止
5. **详细日志**：显示停止过程的详细信息

### 停止的进程包括：
1. `wisefido-radar` 主服务
2. `decode-track` 工具
3. 所有占用端口 8443 的 wisefido-radar 相关进程

## 使用示例

### 示例 1：正常停止服务
```bash
cd wisefido-radar
./stop-radar.sh
```

输出示例：
```
🛑 Stopping wisefido-radar service...

📊 Current running services:
  - wisefido-radar service (PID: 12345)

🛑 Stop Options:
  Mode: Graceful stop (SIGTERM)

Are you sure you want to stop wisefido-radar services? (Y/n): Y

🔄 Gracefully stopping services...
  Sending SIGTERM to wisefido-radar (PID: 12345)
  Waiting for wisefido-radar to stop... (9s remaining)

🔍 Verifying services are stopped...
✅ wisefido-radar stopped
✅ Port 8443 is free

========================================
✅ All wisefido-radar services stopped successfully
========================================

📝 Log files:
  - /tmp/owl_radar_startup.log

💡 To restart services:
  ./start-radar.sh
```

### 示例 2：强制停止服务
```bash
./stop-radar.sh --force
```

### 示例 3：停止并查看详细输出
```bash
./stop-radar.sh --verbose
```

## 注意事项

1. **优雅停止优先**：默认使用优雅停止，给进程时间进行清理
2. **自动降级**：如果优雅停止失败，会自动转为强制停止
3. **端口检查**：脚本会检查并清理 wisefido-radar 占用的端口
4. **安全确认**：停止前会要求用户确认
5. **日志保留**：停止服务不会删除日志文件

## 故障排除

### 问题 1：服务无法停止
```bash
# 尝试强制停止
./stop-radar.sh --force

# 手动检查进程
ps aux | grep wisefido-radar

# 手动停止
pkill -f "wisefido-radar"
pkill -f "decode-track"
```

### 问题 2：端口仍被占用
```bash
# 检查端口占用
lsof -i :8443

# 手动清理端口
lsof -ti :8443 | xargs kill -9
```

### 问题 3：脚本权限问题
```bash
# 添加执行权限
chmod +x stop-radar.sh

# 检查权限
ls -la stop-radar.sh
```

## 相关文件

- `start-radar.sh` - 启动脚本
- `stop-radar.sh` - 停止脚本
- `/tmp/owl_radar_startup.log` - 服务日志文件
- `generate-cert.sh` - 证书生成脚本（如果需要HTTPS）

## 开发说明

停止脚本设计为与启动脚本 (`start-radar.sh`) 配合使用，确保服务可以完整地启动和停止循环。
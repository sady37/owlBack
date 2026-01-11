# HTTPS 服务检测指南

## 当前状态

### 服务状态
- ✅ wisefido-radar 进程存在
- ✅ HTTPS 服务器启动日志显示已启动
- ⚠️  端口 8443 未监听（可能启动失败）

### 证书状态
- ✅ 证书文件存在: `server.crt` (1.3K)
- ✅ 私钥文件存在: `server.key` (1.7K)
- ✅ 证书文件有效
- ✅ 私钥文件有效

## 远端检测方法

### 方法 1: 使用 curl 测试（推荐）

```bash
# 测试 HTTPS 连接
curl -k -v https://47.77.194.143:8443/prod-api/thirdmqtt/v2/auth/device \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "uid": "test-device-001",
    "type": 1,
    "mcu": {
      "hw": "HC2-2.0",
      "sw": "20240101",
      "mac": "AA:BB:CC:DD:EE:FF"
    },
    "radar": {
      "hw": "1.0",
      "sw": "20240101",
      "cap": "T"
    }
  }'
```

**预期结果**:
- ✅ 如果服务正常: 返回 JSON 响应（成功或失败）
- ❌ 如果服务未启动: `Connection refused` 或 `Connection timed out`

### 方法 2: 使用测试脚本

```bash
cd owlBack/wisefido-radar
./test-https-remote.sh 47.77.194.143 8443
```

脚本会测试：
1. 端口连通性
2. HTTPS 连接
3. 证书信息

### 方法 3: 使用 telnet 测试端口

```bash
telnet 47.77.194.143 8443
```

**预期结果**:
- ✅ 如果端口开放: 显示 `Connected to 47.77.194.143`
- ❌ 如果端口关闭: `Connection refused` 或超时

### 方法 4: 使用 nc (netcat) 测试

```bash
nc -zv 47.77.194.143 8443
```

**预期结果**:
- ✅ 如果端口开放: `Connection to 47.77.194.143 8443 port [tcp/*] succeeded!`
- ❌ 如果端口关闭: `Connection refused`

### 方法 5: 使用 openssl 测试证书

```bash
echo | openssl s_client -connect 47.77.194.143:8443 -servername 47.77.194.143
```

**预期结果**:
- ✅ 如果服务正常: 显示证书信息和连接成功
- ❌ 如果服务未启动: `connect: Connection refused`

### 方法 6: 使用 nmap 扫描端口

```bash
nmap -p 8443 47.77.194.143
```

**预期结果**:
- ✅ 如果端口开放: `8443/tcp open  https-alt`
- ❌ 如果端口关闭: `8443/tcp closed https-alt` 或 `8443/tcp filtered https-alt`

## 本地检测方法

### 检查服务是否运行

```bash
# 检查进程
ps aux | grep wisefido-radar | grep -v grep

# 检查端口监听
sudo lsof -i :8443
# 或
sudo netstat -tlnp | grep 8443
# 或
sudo ss -tlnp | grep 8443
```

### 检查启动日志

```bash
tail -f /tmp/owl_radar_startup.log
```

查找以下信息：
- `Starting HTTPS server` - 服务器启动
- `HTTPS server error` - 启动错误
- `addr":":8443"` - 监听地址

### 本地测试 HTTPS

```bash
curl -k -v https://127.0.0.1:8443/prod-api/thirdmqtt/v2/auth/device \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"test":"test"}'
```

## 故障排查

### 如果端口未监听

1. **检查服务是否真的在运行**
   ```bash
   ps aux | grep wisefido-radar
   ```

2. **检查启动日志中的错误**
   ```bash
   tail -100 /tmp/owl_radar_startup.log | grep -i error
   ```

3. **检查证书文件**
   ```bash
   ls -l owlBack/wisefido-radar/server.crt server.key
   openssl x509 -in server.crt -text -noout
   ```

4. **重启服务**
   ```bash
   # 停止服务
   pkill -f wisefido-radar
   
   # 重新启动
   cd owlBack/wisefido-radar
   ./start-radar.sh
   ```

### 如果远端无法访问

1. **检查防火墙规则**
   ```bash
   sudo iptables -L -n | grep 8443
   sudo ufw status | grep 8443
   ```

2. **检查服务器 IP 地址**
   ```bash
   ip addr show | grep 47.77.194.143
   ```

3. **检查服务监听地址**
   - 服务应该监听 `0.0.0.0:8443` 或 `:8443`（所有接口）
   - 不应该只监听 `127.0.0.1:8443`（仅本地）

## 快速检测命令

### 一键检测脚本

```bash
#!/bin/bash
SERVER_IP="47.77.194.143"
PORT="8443"

echo "检测 HTTPS 服务..."
echo "服务器: $SERVER_IP:$PORT"
echo ""

# 端口测试
if nc -zv -w 5 "$SERVER_IP" "$PORT" 2>&1 | grep -q "succeeded"; then
    echo "✅ 端口 $PORT 可访问"
else
    echo "❌ 端口 $PORT 不可访问"
fi

# HTTPS 测试
HTTP_CODE=$(curl -k -s -o /dev/null -w "%{http_code}" \
    --max-time 5 \
    "https://$SERVER_IP:$PORT/prod-api/thirdmqtt/v2/auth/device" \
    -X POST -H "Content-Type: application/json" \
    -d '{"test":"test"}')

if [ "$HTTP_CODE" != "000" ] && [ -n "$HTTP_CODE" ]; then
    echo "✅ HTTPS 服务响应 (HTTP $HTTP_CODE)"
else
    echo "❌ HTTPS 服务无响应"
fi
```

## 总结

### 当前状态
- ⚠️  HTTPS 服务可能未正常启动（端口未监听）
- ✅ 证书文件存在且有效
- ✅ 服务进程存在

### 建议操作
1. 重启 wisefido-radar 服务
2. 检查启动日志确认 HTTPS 服务器是否成功启动
3. 使用远端测试方法验证服务可访问性

# PIN 验证调试指南

## 1. 检查数据库中的 PIN Hash

### 方法 1: 使用 SQL 脚本

运行提供的 SQL 脚本：

```bash
psql -U your_user -d your_database -f check_pin_hash.sql
```

或者直接在 psql 中执行：

```sql
-- 查找 demo 租户 admin 用户的 PIN hash
SELECT 
    u.user_id::text as user_id,
    u.user_account,
    u.nickname,
    encode(u.pin_hash, 'hex') as pin_hash_hex,
    length(u.pin_hash) as pin_hash_length_bytes
FROM users u
JOIN tenants t ON u.tenant_id = t.tenant_id
WHERE t.tenant_name = 'demo' 
  AND u.user_account = 'admin';

-- 计算 PIN 1212 和 1234 的 SHA256 hash
SELECT 
    '1212' as pin_value,
    encode(digest('1212', 'sha256'), 'hex') as pin_hash_hex;

SELECT 
    '1234' as pin_value,
    encode(digest('1234', 'sha256'), 'hex') as pin_hash_hex;
```

### 方法 2: 使用 Go 代码计算

在 Go 中，PIN hash 的计算方式：

```go
import (
    "crypto/sha256"
    "encoding/hex"
)

func hashPIN(pin string) string {
    sum := sha256.Sum256([]byte(pin))
    return hex.EncodeToString(sum[:])
}

// PIN 1212 的 hash
hash1212 := hashPIN("1212")
// PIN 1234 的 hash
hash1234 := hashPIN("1234")
```

## 2. Service 方法输入/输出

### Service 方法: `authService.VerifyPIN`

**位置**: `wisefido-data/internal/service/auth_service.go`

**输入** (`VerifyPINRequest`):
```go
type VerifyPINRequest struct {
    PinHash string // SHA256(pin) 的 hex 编码，必填
    UserID  string // 用户 ID（从 request header X-User-Id 获取），必填
}
```

**输出** (`VerifyPINResponse`):
```go
type VerifyPINResponse struct {
    Success bool `json:"success"`
}
```

**调用流程**:
1. HTTP Handler (`AuthHandler.VerifyPIN`) 接收请求
2. 从 request body 解析 `pin_hash`
3. 从 request header 获取 `X-User-Id`
4. 调用 `authService.VerifyPIN(ctx, req)`
5. Service 方法：
   - 解析 `pin_hash` (hex 字符串) 为字节数组
   - 从数据库查询用户的 `pin_hash` (字节数组)
   - 使用 constant-time comparison 比较两个字节数组
   - 返回 `Success: true/false`

## 3. 查看日志输出

### 日志位置

所有日志都会输出到**标准输出 (STD)**，可以在后端服务的控制台/终端中直接查看。

### 日志格式

当调用 PIN 验证时，会输出以下日志：

```
[AuthHandler.VerifyPIN] ========== HTTP REQUEST ==========
[AuthHandler.VerifyPIN] Method: POST
[AuthHandler.VerifyPIN] URL: /auth/api/v1/verify-pin
[AuthHandler.VerifyPIN] Request payload: map[pin_hash:...]
[AuthHandler.VerifyPIN] X-User-Id header: ...
[AuthHandler.VerifyPIN] pin_hash from payload: ...

[VerifyPIN] ========== INPUT ==========
[VerifyPIN] UserID: ...
[VerifyPIN] PinHash (hex string): ...
[VerifyPIN] PinHash length: ...
[VerifyPIN] PinHash (decoded bytes): ...
[VerifyPIN] PinHash (decoded length): ... bytes

[VerifyPIN] ========== DATABASE VALUE ==========
[VerifyPIN] StoredPinHash (bytes): ...
[VerifyPIN] StoredPinHash (hex string): ...
[VerifyPIN] StoredPinHash length: ... bytes

[VerifyPIN] ========== COMPARISON ==========
[VerifyPIN] Input PinHash (hex): ...
[VerifyPIN] Stored PinHash (hex): ...
[VerifyPIN] Length match: true/false
[VerifyPIN] Hash match: true/false

[VerifyPIN] ========== OUTPUT ==========
[VerifyPIN] Success: true/false

[AuthHandler.VerifyPIN] ========== HTTP RESPONSE ==========
[AuthHandler.VerifyPIN] Response: Success=true/false
```

### 如何查看日志

1. **如果后端服务在终端运行**:
   - 直接在终端查看输出

2. **如果后端服务在后台运行**:
   - 查看服务日志文件
   - 使用 `journalctl` (systemd) 或 `docker logs` (Docker)

3. **如果使用 Docker**:
   ```bash
   docker logs -f <container_name>
   ```

## 4. 常见问题排查

### 问题 1: PIN 验证失败，但 PIN 是正确的

**可能原因**:
- 数据库中存储的 PIN hash 格式不正确
- PIN hash 计算方式不一致

**排查步骤**:
1. 查看日志中的 `Input PinHash (hex)` 和 `Stored PinHash (hex)`
2. 对比两个值是否相同
3. 如果不同，检查数据库中存储的值是否正确

### 问题 2: 数据库中 PIN hash 为 NULL

**可能原因**:
- 用户从未设置过 PIN
- PIN 设置失败

**解决方法**:
- 通过侧边栏 Account Settings 重新设置 PIN
- 或通过管理员界面重置 PIN

### 问题 3: PIN hash 长度不匹配

**可能原因**:
- SHA256 hash 应该是 32 字节 (64 个 hex 字符)
- 如果长度不对，可能是存储或传输过程中出现问题

**检查**:
- 查看日志中的 `PinHash (decoded length)` 和 `StoredPinHash length`
- 两者都应该是 32 字节


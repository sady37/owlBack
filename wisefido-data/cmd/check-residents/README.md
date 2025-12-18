# Check Residents Tool

数据库查询工具，用于排查 resident 的密码 hash 和其他信息。

## 用途

- 查询 resident 的密码 hash 值
- 验证密码是否正确
- 排查登录问题
- 检查多个 resident 的记录

## 使用方法

### 基本用法（查询 r1, r2, r3）

```bash
cd wisefido-data
go run cmd/check-residents/main.go
```

### 查询指定的 resident

```bash
# 通过 resident_account 查询
go run cmd/check-residents/main.go -ids r1,r2,r3

# 查询 nickname 包含 "done" 的 resident
go run cmd/check-residents/main.go -ids done

# 查询多个 resident
go run cmd/check-residents/main.go -ids r1,smith,done
```

### 验证密码 hash

```bash
# 计算密码的 hash 值
go run cmd/check-residents/main.go -password Ts123@123

# 同时查询 resident 并验证密码
go run cmd/check-residents/main.go -ids r1 -password Ts123@123
```

### 显示所有 resident（限制 100 条）

```bash
go run cmd/check-residents/main.go -all
```

### 指定数据库名称

```bash
go run cmd/check-residents/main.go -db wisefido_data
```

## 参数说明

- `-ids`: 逗号分隔的 resident ID 或 account 列表（例如：`r1,r2,r3`）
- `-password`: 要验证的密码（例如：`Ts123@123`）
- `-all`: 显示所有 resident（限制 100 条）
- `-db`: 指定数据库名称（默认：尝试 `wisefido_data` 然后 `owlrd`）

## 环境变量

工具会使用项目的配置文件（`internal/config/config.go`），支持以下环境变量：

- `DB_HOST`: 数据库主机（默认：`localhost`）
- `DB_PORT`: 数据库端口（默认：`5432`）
- `DB_USER`: 数据库用户（默认：`postgres`）
- `DB_PASSWORD`: 数据库密码（默认：`postgres`）
- `DB_NAME`: 数据库名称（默认：`owlrd`）
- `DB_SSLMODE`: SSL 模式（默认：`disable`）

## 输出示例

```
Connected to database: owlrd

Password Hash Comparison:
Password: Ts123@123 -> Hash: 552691efde472982464368f31c2c7684ef4a8a25ce7911a4cb34ce802df5b2c0

Resident Records:
ID | Account | Nickname | Status | Password Hash (hex)
---|--------|----------|--------|-------------------
e01355f6-96ee-4056-9775-542954b0e325 | r1 | smith | active | 483f4d4fb0fd46c623dd75840677c73f0f412c1f885605a5bf2f5ede7efd6165
cc607170-bd83-4a74-9b64-5aa246d32e8c | r2 | Done | active | 552691efde472982464368f31c2c7684ef4a8a25ce7911a4cb34ce802df5b2c0
02c551f1-b000-45f0-b095-3a9b92015188 | r3 | test1 | active | 552691efde472982464368f31c2c7684ef4a8a25ce7911a4cb34ce802df5b2c0

Password Hash: 552691efde472982464368f31c2c7684ef4a8a25ce7911a4cb34ce802df5b2c0
(Compare with password_hash_hex above to verify password)
```

## 常见问题排查

### 1. 验证密码是否正确

```bash
# 查看 r1 的密码 hash
go run cmd/check-residents/main.go -ids r1

# 计算密码的 hash
go run cmd/check-residents/main.go -password Ts123@123

# 比较两个 hash 值是否一致
```

### 2. 检查多个 resident 的密码

```bash
go run cmd/check-residents/main.go -ids r1,r2,r3 -password Ts123@123
```

### 3. 查找特定 nickname 的 resident

```bash
go run cmd/check-residents/main.go -ids done
```

## 注意事项

- 工具会自动尝试连接 `wisefido_data` 和 `owlrd` 数据库
- 密码 hash 使用 SHA256 算法计算
- 查询结果按 `resident_id` 排序


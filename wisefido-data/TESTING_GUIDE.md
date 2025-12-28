# 端到端测试指南：Vue 前端 → PostgreSQL 后端

本指南帮助您配置和测试从 Vue 前端到 PostgreSQL 后端的完整业务流程。

## 前置条件

1. **PostgreSQL 数据库已安装并运行**
   ```bash
   # 检查 PostgreSQL 是否运行
   psql -U postgres -c "SELECT version();"
   ```

2. **Redis 已安装并运行**
   ```bash
   # 检查 Redis 是否运行
   redis-cli ping
   ```

3. **数据库已初始化**
   - 确保 `owlrd` 数据库已创建
   - 确保所有迁移脚本已执行（`owlRD/db/*.sql`）

## 步骤 1: 配置数据库连接

### 方式 A: 使用环境变量（推荐）

在启动后端服务前，设置以下环境变量：

```bash
export DB_ENABLED=true
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres  # 修改为您的数据库密码
export DB_NAME=owlrd
export DB_SSLMODE=disable

export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=
export HTTP_ADDR=:8080
export LOG_LEVEL=info
```

### 方式 B: 使用启动脚本

使用项目提供的启动脚本：

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
./scripts/run_local.sh
```

或使用 `start_server.sh`：

```bash
./scripts/start_server.sh
```

## 步骤 2: 验证数据库连接

在启动服务前，先验证数据库连接：

```bash
# 测试数据库连接
psql -h localhost -U postgres -d owlrd -c "SELECT COUNT(*) FROM tenants;"
```

如果连接成功，应该能看到 tenants 表的记录数。

## 步骤 3: 启动后端服务

### 方式 A: 直接运行

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
go run cmd/wisefido-data/main.go
```

### 方式 B: 编译后运行

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
go build -o wisefido-data cmd/wisefido-data/main.go
./wisefido-data
```

### 验证服务启动

服务启动后，您应该看到类似输出：

```
DB enabled for wisefido-data
Starting server on :8080
```

测试健康检查：

```bash
curl http://localhost:8080/health
```

应该返回 `{"status":"ok"}`

## 步骤 4: 配置前端连接

### 检查前端 API 配置

确保 Vue 前端的 API 基础 URL 指向后端服务：

```typescript
// owlFront/src/utils/http/axios/index.ts
// 确保 baseURL 指向后端服务
baseURL: 'http://localhost:8080'
```

### 启动前端服务

```bash
cd /Users/sady3721/project/owlFront
npm run dev
```

前端通常运行在 `http://localhost:5173` 或 `http://localhost:3000`

## 步骤 5: 测试业务流程

### 5.1 测试登录

1. 打开浏览器访问前端地址
2. 使用系统管理员账号登录：
   - **账号**: `sysadmin`
   - **密码**: `ChangeMe123!`
   - **租户**: System

### 5.2 测试侧边栏 Account 设置

1. 登录后，点击侧边栏的 **Account Settings** 按钮（锁图标）
2. 验证以下功能：

#### Resident 账户设置测试

**前提条件**: 使用 Resident 角色登录

1. **获取账户设置**
   - 前端调用: `GET /admin/api/v1/residents/:id/account-settings`
   - 验证返回字段：
     - `account`: 住户账号
     - `nickname`: 昵称
     - `email`: 邮箱（可为空）
     - `phone`: 电话（可为空）
     - `save_email`: 是否保存 email
     - `save_phone`: 是否保存 phone

2. **更新账户设置**
   - 修改密码、邮箱、电话
   - 前端调用: `PUT /admin/api/v1/residents/:id/account-settings`
   - 验证更新是否成功

#### Staff 账户设置测试

**前提条件**: 使用 Staff 角色（Admin/Nurse/Caregiver 等）登录

1. **获取账户设置**
   - 前端调用: `GET /admin/api/v1/users/:id/account-settings`
   - 验证返回字段

2. **更新账户设置**
   - 前端调用: `PUT /admin/api/v1/users/:id/account-settings`
   - 验证更新是否成功

### 5.3 测试 Resident 管理功能

1. **创建住户**
   - 前端调用: `POST /admin/api/v1/residents`
   - 验证三层结构：
     - `InherentAttributes`: 固有属性
     - `UnitRelation`: Unit 关系
     - `CaregiverRelation`: Caregiver 关系

2. **更新住户**
   - 前端调用: `PUT /admin/api/v1/residents/:id`
   - 验证 `domain.UpdateX` 类型是否正确处理

3. **查询住户列表**
   - 前端调用: `GET /admin/api/v1/residents`
   - 验证 `branch_id` 权限过滤是否正常工作

### 5.4 测试权限检查

验证 `branch_id` 基于权限检查：

1. **创建测试用户和院区关联**
   ```sql
   -- 在 user_branches 表中创建关联
   INSERT INTO user_branches (tenant_id, user_id, branch_id, is_primary)
   VALUES ('tenant-id', 'user-id', 'branch-id', true);
   ```

2. **测试 ListResidents 权限过滤**
   - 用户无关联院区：只显示 `branch_id IS NULL` 的住户
   - 用户关联单个院区：只显示该院区的住户
   - 用户关联多个院区：显示所有关联院区的住户

## 步骤 6: 检查数据库数据

### 查看账户设置数据

```sql
-- 查看住户账户设置
SELECT 
    r.resident_id,
    r.resident_account,
    r.nickname,
    rp.resident_email,
    rp.resident_phone
FROM residents r
LEFT JOIN resident_phi rp ON rp.resident_id = r.resident_id
WHERE r.tenant_id = 'your-tenant-id'
LIMIT 10;

-- 查看用户账户设置
SELECT 
    user_id,
    user_account,
    nickname,
    email,
    phone
FROM users
WHERE tenant_id = 'your-tenant-id'
LIMIT 10;
```

### 查看权限关联

```sql
-- 查看用户-院区关联
SELECT 
    ub.user_id,
    ub.branch_id,
    ub.is_primary,
    b.branch_name
FROM user_branches ub
LEFT JOIN branches b ON b.branch_id = ub.branch_id
WHERE ub.tenant_id = 'your-tenant-id'
LIMIT 10;
```

## 步骤 7: 调试和日志

### 查看后端日志

后端服务会输出详细的日志，包括：
- 数据库连接状态
- API 请求和响应
- 错误信息

### 查看数据库日志

如果 PostgreSQL 配置了日志，可以查看：

```bash
# macOS (Homebrew)
tail -f /usr/local/var/log/postgres.log

# Linux
tail -f /var/log/postgresql/postgresql-*.log
```

### 使用 Doctor 端点（如果启用）

```bash
# 健康检查
curl http://localhost:8080/health

# 就绪检查
curl http://localhost:8080/ready

# 性能分析（如果启用 pprof）
curl http://localhost:8080/debug/pprof/
```

## 常见问题排查

### 1. 数据库连接失败

**错误**: `DB enabled but connection failed`

**解决方案**:
- 检查 PostgreSQL 是否运行: `pg_isready`
- 检查数据库配置是否正确
- 检查防火墙设置
- 验证数据库用户权限

### 2. 前端无法连接后端

**错误**: CORS 错误或连接超时

**解决方案**:
- 检查后端服务是否运行: `curl http://localhost:8080/health`
- 检查前端 API 配置的 baseURL
- 检查浏览器控制台的网络请求

### 3. 权限检查失败

**错误**: `permission denied`

**解决方案**:
- 检查 `user_branches` 表中是否有正确的关联
- 检查 `branch_id` 是否正确设置
- 查看后端日志中的权限检查详情

### 4. 账户设置更新失败

**错误**: 更新后数据未保存

**解决方案**:
- 检查数据库事务是否提交
- 查看后端日志中的错误信息
- 验证字段类型是否正确（`domain.UpdateX` vs 普通指针）

## 测试检查清单

- [ ] PostgreSQL 数据库已启动
- [ ] Redis 已启动
- [ ] 数据库已初始化（所有迁移脚本已执行）
- [ ] 后端服务已启动并连接数据库
- [ ] 前端服务已启动并配置正确的 API 地址
- [ ] 可以成功登录系统
- [ ] 侧边栏 Account 设置功能正常
- [ ] Resident 账户设置可以正常获取和更新
- [ ] Staff 账户设置可以正常获取和更新
- [ ] Resident 管理功能（创建/更新/查询）正常
- [ ] 权限检查（branch_id 过滤）正常工作
- [ ] 数据库中的数据正确保存和更新

## 下一步

完成基本测试后，可以：
1. 测试更多业务场景
2. 进行性能测试
3. 测试错误处理
4. 编写自动化测试用例


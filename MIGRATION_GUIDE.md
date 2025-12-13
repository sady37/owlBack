# 迁移指南

## 从 v1.0 (wisefido-backend) 迁移到 v1.5 (owlBack)

### 主要变化

1. **数据库**: MySQL → PostgreSQL + TimescaleDB
2. **雷达对接**: TCP Socket + MQTT → 仅 MQTT
3. **新增功能**: OTA升级、数据转换、传感器融合、AI报警
4. **架构**: 微服务化，服务职责更清晰

### 迁移步骤

#### 1. 数据库迁移

```bash
# 1. 创建PostgreSQL数据库
psql -U postgres -c "CREATE DATABASE owlrd;"

# 2. 启用TimescaleDB扩展
psql -U postgres -d owlrd -f ../owlRD/db/00_extensions.sql

# 3. 执行所有表结构脚本（按顺序）
for file in ../owlRD/db/*.sql; do
  psql -U postgres -d owlrd -f "$file"
done

# 4. 数据迁移（从MySQL导出，导入PostgreSQL）
# 需要编写迁移脚本
```

#### 2. 服务迁移

##### wisefido-radar
- ✅ 保留MQTT对接
- ❌ 移除TCP Socket相关代码
- ✅ 新增OTA功能模块
- ✅ 适配PostgreSQL

##### wisefido-sleepace
- ✅ 保留MQTT对接
- ✅ 适配PostgreSQL
- ✅ 数据标准化（SNOMED CT）

##### wisefido-data-transformer (新增)
- 从设备服务接收原始数据
- 进行SNOMED CT映射
- 写入iot_timeseries表

##### wisefido-sensor-fusion (新增)
- 从iot_timeseries读取数据
- 执行多传感器融合
- 更新Redis缓存

##### wisefido-alarm (增强)
- 从wisefido-data提取报警逻辑
- 增加AI评估模块
- 增加自动巡检功能

##### wisefido-card-aggregator (新增)
- 聚合卡片数据
- 更新Redis缓存

##### wisefido-data
- 移除数据融合逻辑
- 移除报警处理逻辑
- 改为从Redis读取聚合数据

### 配置迁移

#### v1.0 配置
```yaml
database:
  type: mysql
  host: localhost
  port: 3306
```

#### v1.5 配置
```yaml
database:
  type: postgresql
  host: localhost
  port: 5432
  sslmode: disable
```

### 数据迁移脚本

需要编写数据迁移工具，将MySQL数据迁移到PostgreSQL：

```go
// tools/migrate/main.go
// 1. 连接MySQL（源）
// 2. 连接PostgreSQL（目标）
// 3. 迁移表数据
// 4. 数据转换（字段映射、类型转换）
```

### 测试清单

- [ ] 数据库连接正常
- [ ] MQTT消息接收正常
- [ ] 数据转换正确
- [ ] 传感器融合正确
- [ ] 报警处理正常（规则 + AI）
- [ ] API接口正常
- [ ] 性能测试通过

### 回滚计划

如果迁移失败，需要回滚：

1. 停止v1.5服务
2. 恢复v1.0服务
3. 恢复MySQL数据库（如有备份）


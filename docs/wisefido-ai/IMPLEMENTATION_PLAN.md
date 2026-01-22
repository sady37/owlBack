# wisefido-ai 服务实现计划

## 📋 当前状态

### ✅ 已完成
- 项目结构已创建
- Config 已定义
- Models 已定义（realtime_data, alarm_event, alarm_config）

### ⏳ 待实现
- Repository 层（数据库操作）
- Consumer 层（Redis 缓存读取）
- Evaluator 层（报警规则评估）
- Service 层（主逻辑）
- Main 入口

## 🎯 实现步骤

### 阶段 1：Repository 层
- [ ] `alarm_cloud.go` - 读取租户级别报警策略
- [ ] `alarm_device.go` - 读取设备级别报警配置
- [ ] `alarm_events.go` - 写入报警事件到 PostgreSQL

### 阶段 2：Consumer 层
- [ ] `cache_consumer.go` - 轮询 Redis 缓存（`vital-focus:card:{card_id}:realtime`）
- [ ] `cache_manager.go` - 更新报警缓存（`vital-focus:card:{card_id}:alarms`）

### 阶段 3：Evaluator 层
- [ ] `vital_signs.go` - 生命体征评估（HR/RR 异常、呼吸暂停）
- [ ] `behavior.go` - 行为事件评估（跌倒、离床）
- [ ] `device_status.go` - 设备状态评估（离线、低电量、故障）

### 阶段 4：Service 层
- [ ] `alarm.go` - 报警服务主逻辑，整合各层

### 阶段 5：Main 入口
- [ ] `main.go` - 服务启动入口

## 📝 实现顺序

建议按以下顺序实现：
1. Repository 层（数据访问基础）
2. Consumer 层（数据读取）
3. Evaluator 层（核心业务逻辑）
4. Service 层（整合）
5. Main 入口（启动服务）

## 🔗 相关文档

- `docs/13_Alarm_Fusion_Implementation.md` - 详细设计文档
- `docs/alarm_rule.md` - 报警规则详细说明
- `docs/system_architecture_complete.md` - 系统架构文档


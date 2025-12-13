# 下一步行动计划

## 📋 当前状态

### ✅ 已完成
1. **卡片创建**（wisefido-card-aggregator）
   - ✅ 卡片创建规则实现
   - ✅ 轮询模式（每60秒全量更新）
   - ✅ 定时任务（每天上午9点）
   - ✅ 事件驱动模式（代码已实现，待 wisefido-data 服务发布事件）

2. **传感器融合**（wisefido-sensor-fusion）
   - ✅ 代码已实现
   - ✅ 已检查并验证：符合前端绑定规则和卡片需求定义
   - ✅ GetCardByDeviceID 查询逻辑正确
   - ⏳ **需要功能验证**：实际运行测试

## 🎯 下一步选项

### 选项 1：验证传感器融合功能（推荐）

**目标**：确保 wisefido-sensor-fusion 能正常工作

**步骤**：
1. 运行 `wisefido-card-aggregator` 创建卡片
2. 运行 `wisefido-sensor-fusion` 服务
3. 发送测试数据到 `iot:data:stream`
4. 验证：
   - 能正确查询到卡片
   - 能正确融合数据
   - Redis 缓存正确更新

**优点**：
- 验证已实现的功能是否正常工作
- 确保数据流畅通
- 为后续功能开发打好基础

### 选项 2：实现报警评估层（wisefido-alarm）

**目标**：实现报警规则评估功能

**功能**：
- 读取融合后的实时数据（`vital-focus:card:{card_id}:realtime`）
- 评估报警规则（HR/RR异常、跌倒、离床等）
- 写入报警事件到 PostgreSQL（`alarm_events` 表）
- 更新报警缓存到 Redis（`vital-focus:card:{card_id}:alarms`）

**优点**：
- 继续完善数据流
- 实现核心业务功能

### 选项 3：实现卡片数据聚合（wisefido-card-aggregator）

**目标**：实现卡片数据聚合功能（与卡片创建不同）

**功能**：
- 从 PostgreSQL 读取基础信息（cards, devices, residents）
- 从 Redis 读取实时数据（`vital-focus:card:{card_id}:realtime`）
- 从 Redis 读取报警数据（`vital-focus:card:{card_id}:alarms`）
- 组装完整的 VitalFocusCard 对象
- 更新 Redis 缓存（`vital-focus:card:{card_id}:full`）

**优点**：
- 为 API 服务提供完整数据
- 完成数据流的关键环节

## 📊 数据流顺序

```
1. 卡片创建 ✅ 已完成
   ↓
2. 传感器融合 ✅ 已实现，待验证
   ↓
3. 报警评估 ⏳ 待实现
   ↓
4. 卡片聚合 ⏳ 待实现
   ↓
5. API 服务 ⏳ 待实现
```

## 💡 建议

**推荐顺序**：
1. **先验证传感器融合功能**（选项 1）
   - 确保已实现的功能正常工作
   - 验证数据流是否畅通
   
2. **再实现报警评估层**（选项 2）
   - 继续完善数据流
   - 实现核心业务功能

3. **然后实现卡片数据聚合**（选项 3）
   - 为 API 服务提供完整数据

4. **最后实现 API 服务**（选项 4）
   - 提供 HTTP 接口给前端

## 🔗 相关文档

- `docs/system_architecture_complete.md` - 系统架构文档
- `docs/NEXT_STEPS_VERIFICATION.md` - 传感器融合验证步骤
- `wisefido-sensor-fusion/CHECK_SUMMARY.md` - 传感器融合检查总结
- `wisefido-alarm/IMPLEMENTATION_PLAN.md` - 报警服务实现计划


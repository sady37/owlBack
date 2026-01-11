# wisefido-ai 问题分析（TODO、权限、分层）

## 📋 检查结果总结

### ✅ 架构分层

**当前分层结构**：
```
main.go
    ↓
AlarmService (service/alarm.go)
    ├─ CacheConsumer (consumer/cache_consumer.go)
    ├─ Evaluator (evaluator/evaluator.go)
    │   ├─ Event1Evaluator
    │   ├─ Event2Evaluator
    │   ├─ Event3Evaluator
    │   └─ Event4Evaluator
    └─ Repositories
        ├─ AlarmEventsRepository
        ├─ CardRepository
        ├─ DeviceRepository
        └─ ...
```

**分层问题**：
- ✅ **Repository 层**：已实现，结构清晰
- ✅ **Consumer 层**：已实现，负责读取 Redis 缓存
- ✅ **Evaluator 层**：已实现，负责报警规则评估
- ✅ **Service 层**：已实现，整合各层
- ⚠️ **Handler 层**：**缺失** - `wisefido-ai` 没有 HTTP Handler

**关键发现**：
- `wisefido-ai` 是一个**后台服务**（轮询模式），不是 HTTP API 服务
- `AlarmEventService` 在 `wisefido-data` 中实现，用于 HTTP API
- `wisefido-ai` 的 `AlarmService` 是后台评估服务，不是 HTTP 服务

---

## ⏳ TODO 问题

### 1. Evaluator 层的事件评估逻辑（高优先级）

#### 事件1：床上跌落检测 (`event1_bed_fall.go`)

**TODO 项**：
- ⚠️ **TODO: 实现完整的事件1逻辑**（第32行）
- ⚠️ **TODO: 实现完整的状态管理和定时器逻辑**（第54行）
- ⚠️ **TODO: 需要检查 postures 中是否有移动的 track_id**（第79行）
- ⚠️ **TODO: 检查位置是否变化（需要历史位置数据）**（第82行）

**当前状态**：
- ✅ 基础框架已实现
- ✅ 退出条件检查已实现（简化版）
- ✅ 状态管理函数已实现（`getEvent1State`, `setEvent1State`）
- ❌ 完整的状态管理和定时器逻辑未实现
- ❌ 位置变化检测未实现

**需要实现**：
1. 状态管理（lying基线、离床时间、track_id状态）
2. 定时器（T0+5秒、T0+60秒、T0+120秒）
3. 退出条件检查（持续检查）
4. 危险情况检测（track_id突然消失）

---

#### 事件2：Sleepad可靠性判断 (`event2_sleepad_reliability.go`)

**TODO 项**：
- ⚠️ **TODO: 实现完整的事件2逻辑**（第26行）
- ⚠️ **TODO: 需要查询卡片绑定的设备，检查是否有 Radar 设备**（第42行）

**当前状态**：
- ✅ 基础框架已实现
- ❌ 核查1（前置条件检查）未实现
- ❌ 分支判断（分支A和分支B）未实现
- ❌ 核查2和核查3未实现

**需要实现**：
1. 核查1（前置条件检查）
2. 分支判断（分支A和分支B）
3. 核查2和核查3

---

#### 事件3：Bathroom可疑跌倒检测 (`event3_bathroom_fall.go`)

**TODO 项**：
- ⚠️ **TODO: 实现完整的事件3逻辑**（第27行）
- ⚠️ **TODO: 需要检查 postures 中的姿态，判断是否是站立状态**（第46行）

**当前状态**：
- ✅ 基础框架已实现
- ✅ 卫生间检测已实现（`roomRepo.IsBathroom`）
- ❌ 站立状态检测未实现
- ❌ 位置变化检测未实现
- ❌ 单人检测未实现

**需要实现**：
1. 站立状态检测（检查 postures 中的姿态）
2. 位置变化检测（位置变化小于10cm，超过10分钟）
3. 单人检测（雷达检测范围内仅1人）

---

#### 事件4：雷达检测到人突然消失 (`event4_sudden_disappear.go`)

**TODO 项**：
- ⚠️ **TODO: 实现完整的事件4逻辑**（第25行）
- ⚠️ **TODO: 需要维护 track_id 的历史状态，检测是否突然消失**（第30行）
- ⚠️ **TODO: 需要维护历史高度数据，检测高度变化**（第34行）

**当前状态**：
- ✅ 基础框架已实现
- ❌ track_id 历史状态管理未实现
- ❌ 质心降低检测未实现
- ❌ 5分钟无活动检测未实现

**需要实现**：
1. track_id 历史状态管理
2. 质心降低检测（高度降低超过60cm）
3. 5分钟无活动检测

---

## 🔐 权限问题

### 1. AlarmEventService 权限检查

**位置**：`wisefido-data/internal/service/alarm_event_service.go`

**当前状态**：
- ✅ `HandleAlarmEvent` 方法有权限检查（`checkHandlePermission`）
- ✅ 权限检查逻辑：
  - Facility 类型卡片：只有 Nurse 或 Caregiver 可以处理
  - Home 类型卡片：所有角色都可以处理

**问题**：
- ⚠️ **权限检查不完整**：只检查了 Facility/Home 类型，没有检查：
  - `assigned_only` 权限（Caregiver/Nurse 只能处理分配的住户）
  - `branch_only` 权限（Manager 只能处理同分支的住户）

**需要改进**：
1. 添加 `assigned_only` 权限检查（参考 `ResidentService`）
2. 添加 `branch_only` 权限检查（参考 `ResidentService`）
3. 统一权限检查逻辑（与 `SleepaceReportHandler` 保持一致）

---

### 2. wisefido-ai 服务权限

**位置**：`wisefido-ai` 服务本身

**当前状态**：
- ✅ `wisefido-ai` 是后台服务，不需要 HTTP 权限检查
- ✅ 服务级别的权限通过 `TENANT_ID` 环境变量控制

**结论**：
- ✅ **无权限问题** - 后台服务不需要 HTTP 权限检查

---

## 🏗️ 分层问题

### 1. Service 层职责不清

**问题**：
- ⚠️ `AlarmService` (service/alarm.go) 和 `AlarmEventService` (service/alarm_event_service.go) 职责重叠
- ⚠️ `AlarmService` 是后台评估服务，`AlarmEventService` 是 HTTP API 服务

**当前架构**：
```
wisefido-ai/
  └─ service/
      ├─ alarm.go              ← 后台评估服务（轮询模式）
      └─ alarm_event_service.go ← HTTP API 服务（查询、处理报警事件）
```

**问题分析**：
- `AlarmEventService` 应该在 `wisefido-data` 中（已实现）
- `AlarmService` 应该在 `wisefido-ai` 中（已实现）
- ✅ **架构正确** - 两个服务职责分离

**建议**：
- ✅ **保持现状** - 两个服务职责分离是正确的

---

### 2. Repository 层跨服务使用

**问题**：
- ⚠️ `AlarmEventService` 在 `wisefido-data` 中，但使用了 `wisefido-ai` 的 Repository
- ⚠️ 这可能导致跨模块依赖问题

**当前实现**：
- `wisefido-data` 中的 `AlarmEventService` 使用自己的 Repository（`wisefido-data/internal/repository/alarm_events_repo.go`）
- `wisefido-ai` 中的 `AlarmEventsRepository` 是独立的实现

**结论**：
- ✅ **无问题** - 两个服务各自维护自己的 Repository 实现

---

### 3. Handler 层缺失

**问题**：
- ⚠️ `wisefido-ai` 没有 HTTP Handler 层
- ⚠️ HTTP API 在 `wisefido-data` 中实现

**当前架构**：
```
wisefido-ai/
  └─ 后台服务（轮询模式，无 HTTP API）

wisefido-data/
  └─ HTTP API（查询、处理报警事件）
```

**结论**：
- ✅ **架构正确** - `wisefido-ai` 是后台服务，不需要 HTTP Handler
- ✅ HTTP API 在 `wisefido-data` 中实现是正确的

---

## 📝 待处理问题清单

### 高优先级

1. **Evaluator 层事件评估逻辑**
   - [ ] 事件1：完整的状态管理和定时器逻辑
   - [ ] 事件2：核查1、分支判断、核查2和核查3
   - [ ] 事件3：站立状态检测、位置变化检测、单人检测
   - [ ] 事件4：track_id 历史状态管理、质心降低检测、5分钟无活动检测

2. **AlarmEventService 权限检查**
   - [ ] 添加 `assigned_only` 权限检查
   - [ ] 添加 `branch_only` 权限检查
   - [ ] 统一权限检查逻辑（与 `SleepaceReportHandler` 保持一致）

### 中优先级

3. **配置和优化**
   - [ ] 添加更多配置项（事件1-4的阈值、时间窗口等）
   - [ ] 优化 `GetAllCardIDs` 方法（从 PostgreSQL 查询，而不是扫描 Redis）
   - [ ] 添加指标监控（处理速度、报警数量等）

### 低优先级

4. **测试**
   - [ ] 添加 Evaluator 层单元测试
   - [ ] 添加 Service 层集成测试
   - [ ] 添加端到端测试

---

## 🎯 建议的修复顺序

### 第一步：修复权限检查（最重要）

**文件**：`wisefido-data/internal/service/alarm_event_service.go`

**需要修改**：
1. `HandleAlarmEvent` 方法
2. 添加 `assigned_only` 权限检查
3. 添加 `branch_only` 权限检查
4. 参考 `SleepaceReportHandler` 的权限检查实现

---

### 第二步：完善 Evaluator 层事件评估逻辑

**文件**：
- `wisefido-ai/internal/evaluator/event1_bed_fall.go`
- `wisefido-ai/internal/evaluator/event2_sleepad_reliability.go`
- `wisefido-ai/internal/evaluator/event3_bathroom_fall.go`
- `wisefido-ai/internal/evaluator/event4_sudden_disappear.go`

**优先级**：
1. 事件1（床上跌落检测）- 最复杂，最重要
2. 事件3（Bathroom可疑跌倒检测）- 安全相关
3. 事件4（人突然消失）- 安全相关
4. 事件2（Sleepad可靠性判断）- 相对简单

---

## ✅ 总结

### 架构分层

- ✅ **Repository 层**：结构清晰，无问题
- ✅ **Consumer 层**：职责明确，无问题
- ✅ **Evaluator 层**：框架已实现，逻辑待完善
- ✅ **Service 层**：职责分离正确（后台服务 vs HTTP API 服务）
- ✅ **Handler 层**：HTTP API 在 `wisefido-data` 中实现，架构正确

### 权限问题

- ⚠️ **AlarmEventService 权限检查不完整**：
  - 缺少 `assigned_only` 权限检查
  - 缺少 `branch_only` 权限检查
  - 需要与 `SleepaceReportHandler` 保持一致

### TODO 问题

- ⚠️ **Evaluator 层事件评估逻辑未完成**：
  - 事件1-4 都有 TODO，需要实现完整的评估逻辑
  - 优先级：事件1 > 事件3 > 事件4 > 事件2

---

## 🔗 相关文档

- `SERVICE_LAYER_COMPLETE_DESIGN.md` - Service 层完整设计
- `IMPLEMENTATION_SUMMARY.md` - 实现总结
- `REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结


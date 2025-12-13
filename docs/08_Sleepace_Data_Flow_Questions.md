# Sleepace 数据流问题清单

## ❓ 需要确认的问题

### 1. Sleepace 数据来源

**问题**: Sleepace 数据如何进入系统？

**选项**:
- [ ] Sleepace 厂家服务直接写入 PostgreSQL `iot_timeseries` 表
- [ ] Sleepace 厂家服务发布到 Redis Streams (`sleepace:data:stream`)
- [ ] Sleepace 厂家服务通过 HTTP API 提供数据
- [ ] 其他方式（请说明）

### 2. Sleepace 数据格式

**问题**: Sleepace 数据格式是什么？

**选项**:
- [ ] 已经是标准格式（SNOMED CT 编码、FHIR Category）
- [ ] v1.0 格式（需要转换）
- [ ] 其他格式（请说明）

### 3. wisefido-data-transformer 是否需要处理 Sleepace 数据

**问题**: `wisefido-data-transformer` 是否需要处理 Sleepace 数据？

**如果 Sleepace 数据直接写入数据库**:
- ✅ 不需要在 `wisefido-data-transformer` 中处理
- ✅ 只需要处理 Radar 数据

**如果 Sleepace 数据进入 Redis Streams**:
- ⚠️ 需要在 `wisefido-data-transformer` 中处理
- ⚠️ 需要实现 Sleepace 数据转换器

### 4. 数据一致性

**问题**: 如何确保 Sleepace 和 Radar 数据的一致性？

**考虑**:
- 数据格式是否统一？
- 时间戳是否同步？
- 位置信息是否一致？

---

## 📝 当前实现状态

### wisefido-data-transformer

**当前实现**:
- ✅ 消费 `radar:data:stream`
- ⚠️ `sleepace:data:stream` 消费代码已注释（待确认）

**需要调整**:
- 根据确认结果，决定是否启用 Sleepace 数据消费
- 如果不需要，可以完全移除相关代码

---

## 🔧 建议的调整方案

### 方案 A：Sleepace 数据直接写入数据库

**调整**:
1. 移除 `wisefido-data-transformer` 中的 Sleepace 相关代码
2. 只保留 Radar 数据处理
3. 更新文档说明

### 方案 B：Sleepace 数据进入 Redis Streams

**调整**:
1. 启用 `sleepace:data:stream` 消费代码
2. 实现 Sleepace 数据转换器
3. 处理 Sleepace 数据格式转换

---

## 📋 待确认事项清单

- [ ] Sleepace 数据来源（直接写入数据库 / Redis Streams / HTTP API / 其他）
- [ ] Sleepace 数据格式（标准格式 / v1.0 格式 / 其他）
- [ ] `wisefido-data-transformer` 是否需要处理 Sleepace 数据
- [ ] 数据一致性要求

---

## 🎯 下一步

1. **与 Sleepace 厂家确认**数据提供方式
2. **根据确认结果调整** `wisefido-data-transformer`
3. **更新文档**和代码注释


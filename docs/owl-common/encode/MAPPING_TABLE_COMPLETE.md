# 完整映射表（供检查）

## 一、床上坐起映射（统一 SNOMED: 225698008）

| 设备 | 原始字段/事件 | 原始值 | 标准字段名 | SNOMED 编码 | SNOMED 显示 | Category | 映射路径 |
|------|------------|--------|----------|------------|------------|----------|---------|
| **Sleepace** | `sitUp` | `0` | `sitUp` | `109030009` | "Lying position" | `activity` | `realtime.sitUp` |
| **Sleepace** | `sitUp` | `> 0` | `sitUp` | `225698008` | "Sitting up in bed" | `activity` | `realtime.sitUp` |
| **Radar** | `BED_SIT_UP` | 事件类型 | `BED_SIT_UP` | `225698008` | "Sitting up in bed" | `activity` | `event.BED_SIT_UP` |
| **Radar** | `pose` | `9` | `pose` | `225698008` | "Sitting up in bed" | `activity` | `monitor.track.pose` |
| **Radar** | `pose` | `10` | `pose` | `225698008` | "Sitting up in bed" | `activity` | `monitor.track.pose` |
| **Radar** | `pose` | `11` | `pose` | `225698008` | "Sitting up in bed" | `activity` | `monitor.track.pose` |

**说明**：
- Sleepace: `sitUp = 0` 时映射为 "Lying position" (109030009) - 床上躺下
- Radar pose 9: 普通床上坐起 (display_en: "Sitting up in bed")
- Radar pose 10: 疑似床上坐起 (display_en: "Sitting up in bed:suspected")
- Radar pose 11: 确认床上坐起 (display_en: "Sitting up in bed:confirmed")

## 二、离床事件映射（统一 SNOMED: 248570008）

| 设备 | 原始字段/事件 | 原始值 | 标准字段名 | SNOMED 编码 | SNOMED 显示 | Category | 映射路径 |
|------|------------|--------|----------|------------|------------|----------|---------|
| **Sleepace** | `bedStatus` | `1` | `bedStatus` | `248570008` | "Not in bed" | `activity` | `realtime.bedStatus` |
| **Radar** | `event` | `6` | `event` | `248570008` | "Not in bed" | `activity` | `monitor.track.event` |
| **通用** | `LEFT_BED` | 事件类型 | `LEFT_BED` | `248570008` | "Not in bed" | `activity` | `event.LEFT_BED` |

**说明**：
- Sleepace: `bedStatus = 0` 时映射为 `248569007` - "In bed"
- 告警 Category: `behavioral`（在云端计算告警时）

## 三、Radar pose 值完整映射表

| pose 值 | 含义 | SNOMED 编码 | SNOMED 显示 | Category | display_en | 说明 |
|---------|------|------------|------------|----------|------------|------|
| `0` | 初始化 | `null` | "Initialization" | `activity` | "Initialization" | - |
| `1` | 行走 | `129006008` | "Walking" | `activity` | "Walking" | - |
| `2` | 疑似跌倒 | `129839007` | "At risk for falls" | `safety` | "At risk for falls" | - |
| `3` | 蹲坐 | `402120000` | "Sitting position" | `activity` | "Sitting position" | - |
| `4` | 站立 | `383370001` | "Standing position" | `activity` | "Standing position" | - |
| `5` | 确认跌倒 | `161898004` | "Fall" | `safety` | "Fall" | - |
| `6` | 卧床 | `109030009` | "Lying position" | `activity` | "Lying position" | - |
| `7` | 疑似坐地 | `129839007` | "At risk for falls" | `safety` | "At risk for falls" | - |
| `8` | 确认坐地 | `161898004` | "Fall" | `safety` | "Fall" | - |
| **`9`** | **普通床上坐起** | **`225698008`** | **"Sitting up in bed"** | **`activity`** | **"Sitting up in bed"** | **✅ 已修正** |
| **`10`** | **疑似床上坐起** | **`225698008`** | **"Sitting up in bed"** | **`activity`** | **"Sitting up in bed:suspected"** | **✅ 已修正** |
| **`11`** | **确认床上坐起** | **`225698008`** | **"Sitting up in bed"** | **`activity`** | **"Sitting up in bed:confirmed"** | **✅ 已修正** |

**映射位置**：`owlBack/wisefido-qinglan/internal/decode/config/radar_convert_table.json` → `monitor.track.pose`

## 四、Radar event 值映射表

| event 值 | 含义 | SNOMED 编码 | SNOMED 显示 | Category | display_en | 映射路径 |
|---------|------|------------|------------|----------|------------|---------|
| `0` | 无事件 | `null` | "No event" | `activity` | "No event" | `monitor.track.event` |
| `1` | 进入房间 | `null` | "Enter room" | `activity` | "Enter room" | `event.ENTER_ROOM` |
| `2` | 离开房间 | `null` | "Leave room" | `activity` | "Leave room" | `event.LEAVE_ROOM` |
| `3` | 进入区域 | `null` | "Enter area" | `activity` | "Enter area" | `event.ENTER_AREA` |
| `4` | 离开区域 | `null` | "Leave area" | `activity` | "Leave area" | `event.LEAVE_AREA` |
| `5` | 进入床 | `248569007` | "In bed" | `activity` | "Enter bed" | `event.ENTER_BED` |
| `6` | 离开床 | `248570008` | "Not in bed" | `activity` | "Left bed" | `event.LEFT_BED` |

**映射位置**：
- `monitor.track.event` → `owlBack/wisefido-qinglan/internal/decode/config/radar_convert_table.json`
- `event.*` → `owlRD/db/24_snomed_mapping.sql`

## 五、Sleepace bedStatus 映射表

| bedStatus 值 | 含义 | SNOMED 编码 | SNOMED 显示 | Category | display_en | 映射路径 |
|-------------|------|------------|------------|----------|------------|---------|
| `0` | 在床 | `248569007` | "In bed" | `activity` | "In bed" | `realtime.bedStatus` |
| `1` | 离床 | `248570008` | "Not in bed" | `activity` | "Left bed" | `realtime.bedStatus` |

**映射位置**：`owlBack/owl-common/encode/config/sleepace_convert_table.json` → `realtime.bedStatus`

## 六、Sleepace sitUp 映射表

| sitUp 值 | 含义 | SNOMED 编码 | SNOMED 显示 | Category | display_en | 映射路径 |
|---------|------|------------|------------|----------|------------|---------|
| `0` | 床上躺下 | `109030009` | "Lying position" | `activity` | "Lying position" | `realtime.sitUp` |
| `> 0` | 床上坐起 | `225698008` | "Sitting up in bed" | `activity` | "Sitting up in bed" | `realtime.sitUp` |

**映射位置**：`owlBack/owl-common/encode/config/sleepace_convert_table.json` → `realtime.sitUp`

## 七、通用事件映射表（数据库）

| 事件类型 | SNOMED 编码 | SNOMED 显示 | Category | display_en | 映射位置 |
|---------|------------|------------|----------|------------|---------|
| `LEFT_BED` | `248570008` | "Not in bed" | `activity` | "Left bed" | `24_snomed_mapping.sql` |
| `ENTER_BED` | `248569007` | "In bed" | `activity` | "Enter bed" | `24_snomed_mapping.sql` |
| `BED_SIT_UP` | `225698008` | "Sitting up in bed" | `activity` | "Bed sit up" | `24_snomed_mapping.sql` |

**映射位置**：`owlRD/db/24_snomed_mapping.sql` → `snomed_mapping` 表

## 八、告警类型统一命名表

| 告警类型 | 统一名称 | Category | SNOMED 编码 | 说明 |
|---------|---------|----------|------------|------|
| 离床告警 | `LeftBed` | `behavioral` | `248570008` | 统一，不分 Radar_/SleepPad_ 前缀 |
| 床上坐起告警 | `SitUp` / `BED_SIT_UP` | `behavioral` | `225698008` | 统一，不分 SleepPad_ 前缀 |
| 呼吸暂停告警 | `ApneaHypopnea` | `clinical` | `67905006` | 统一，不分 Radar_ 前缀 |
| 心率异常告警 | `AbnormalHeartRate` | `clinical` | - | 统一，不分 Radar_ 前缀 |
| 呼吸频率异常告警 | `AbnormalRespiratoryRate` | `clinical` | - | 统一，不分 Radar_ 前缀 |
| 异常体动告警 | `AbnormalBodyMovement` | `behavioral` | - | 统一，不分 SleepPad_ 前缀 |
| 跌倒告警 | `Fall` | `safety` | `161898004` | 统一，不分设备 |
| 疑似跌倒告警 | `SuspectedFall` | `safety` | `129839007` | 统一，不分设备 |

## 九、关键修正说明

### 9.1 Radar pose 9-11 修正
- ❌ **之前错误**：
  - pose 9 → `109030009` - "Lying position" (卧位)
  - pose 10 → `129839007` - "At risk for falls" (疑似跌倒)
  - pose 11 → `161898004` - "Fall" (跌倒)
- ✅ **现在正确**：
  - pose 9, 10, 11 → `225698008` - "Sitting up in bed" (床上坐起，SNOMED CT 标准编码)

### 9.2 Radar BED_SIT_UP 事件修正
- ❌ **之前错误**：`40199007` - "Bed sitting position" 或 `422256002` - "Chronic obstructive lung disease"
- ✅ **现在正确**：`225698008` - "Sitting up in bed"（SNOMED CT 标准编码）

### 9.3 告警命名统一
- ❌ **之前**：`Radar_LeftBed`, `SleepPad_LeftBed`, `Radar_ApneaHypopnea`, `SleepPad_SitUp`
- ✅ **现在**：`LeftBed`, `ApneaHypopnea`, `SitUp`（统一，不分设备前缀）

## 十、映射文件位置

| 映射类型 | 文件路径 | 说明 |
|---------|---------|------|
| Radar pose/event | `owlBack/wisefido-qinglan/internal/decode/config/radar_convert_table.json` | Radar 设备字段映射 |
| Sleepace 字段 | `owlBack/owl-common/encode/config/sleepace_convert_table.json` | Sleepace 设备字段映射 |
| 通用事件 | `owlRD/db/24_snomed_mapping.sql` | 数据库 SNOMED 映射表 |

## 十一、检查要点

1. ✅ **床上坐起统一**：所有床上坐起相关映射都使用 `225698008`（SNOMED CT 标准编码）
2. ✅ **显示名称统一**：`snomed_display` 使用标准名称 "Sitting up in bed"
3. ✅ **display_en 格式**：pose 10 使用 "Sitting up in bed:suspected"，pose 11 使用 "Sitting up in bed:confirmed"
2. ✅ **离床事件统一**：所有离床相关映射都使用 `248570008`
3. ✅ **pose 9-11 修正**：已从卧位/跌倒修正为床上坐起
4. ✅ **告警命名统一**：移除所有设备前缀（Radar_/SleepPad_）
5. ✅ **Category 分类正确**：床上坐起为 `activity`，离床为 `activity`（告警时为 `behavioral`）

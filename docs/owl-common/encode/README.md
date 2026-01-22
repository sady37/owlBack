# Encode 模块文档说明

## 最终版本文档

### 核心文档
1. **MAPPING_TABLE_COMPLETE.md** - 完整映射表（所有字段和事件的 SNOMED 映射，快速查询参考）

### 规范文档
2. **config/RADAR_SNOMED_CATEGORY_MAPPING.md** - Radar SNOMED 分类映射规范（Category 分类体系、实现原理）

### 字段文档
3. **FIELD_MAPPING_COMPLETE_LIST.md** - 完整字段映射列表（包含字段命名标准、所有设备的字段映射）

### 实现文档
4. **SLEEPACE_ENCODER_IMPLEMENTATION.md** - Sleepace 编码器实现说明

### 配置文档（config/）
5. **config/README_RADAR_CONVERT.md** - Radar 转换表说明
6. **config/06_FHIR_Simple_Conversion_Guide.md** - FHIR 转换指南
7. **config/Radar_HTTPS_MQTT_Protocol_Formatted.md** - Radar 协议文档

## 文档使用指南

- **快速查询映射关系**：查看 `MAPPING_TABLE_COMPLETE.md`（包含所有具体的 SNOMED 编码和映射值）
- **了解分类体系**：查看 `config/RADAR_SNOMED_CATEGORY_MAPPING.md`（了解 Category 分类规则和实现原理）
- **查询字段列表**：查看 `FIELD_MAPPING_COMPLETE_LIST.md`（所有字段的详细映射）
- **了解实现细节**：查看 `SLEEPACE_ENCODER_IMPLEMENTATION.md` 和 config 目录下的文档

## 文档关系说明

- **MAPPING_TABLE_COMPLETE.md**：快速查询表，包含所有具体映射值（Radar + Sleepace）
- **config/RADAR_SNOMED_CATEGORY_MAPPING.md**：系统规范文档，解释 Radar 的 Category 分类体系和实现原理
- 两者侧重点不同：一个侧重快速查询，一个侧重系统规范

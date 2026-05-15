# 美国 Elder Care / Hospital Bed 规格参考

> WiseFido 床矩形生成算法的先验数据源
> Version 1.0 · 2026-05-14
> 适用市场：USA

---

## 1. 标准家用护理床 (Home Hospital Bed)

适用场景：家庭老人护理、轻度需求

| 维度 | 尺寸 |
|---|---|
| **宽度** | 36 英寸（91 cm）|
| **长度** | 80 英寸（203 cm）|
| **承重** | 350-400 磅 |
| **等效床型** | Twin XL（38×80"）|

**市场占比**：覆盖 70%+ 的家庭护理 / 入院护理场景。

---

## 2. 标准病房床 (Hospital Bed)

适用场景：医院、Skilled Nursing Facility、Assisted Living

**尺寸与家用护理床相同**：

| 维度 | 尺寸 |
|---|---|
| **宽度** | 36 英寸（91 cm）|
| **长度** | 80 英寸（203 cm）|

**核心差异在高度可调**——见下文。

---

## 3. 床高度可调范围

这是病房床与家用普通床的本质区别，直接影响 WiseFido 的 fall alarm 逻辑。

| 床类型 | 最低高度 | 最高高度 | Travel Range |
|---|---|---|---|
| **Standard Adjustable** | 21" (53 cm) | 29" (74 cm) | ~20 cm |
| **Hi-Low Bed** | 7-9" (18-23 cm) | 30-32" (76-81 cm) | ~60 cm |
| **Ultra-Low** | 7" (18 cm) | 30" (76 cm) | ~58 cm |
| **Premium (SonderCare 等)** | 17" (43 cm) | 39" (99 cm) | 56 cm |

**Hi-Low 床的设计用途**：
- **低位（7-12"）**：Fall Prevention（防跌落）——夜间或老人独处时降到最低
- **高位（30-32"）**：Caregiver Access（方便护理人员）——护理操作时升高

---

## 4. 加长床型 (Extended Length)

适用：高个子患者

| 床型 | 长度 |
|---|---|
| Standard | 80" (203 cm) |
| Extended | 84" (213 cm) |
| Long | 88" (224 cm) |

**长度变化范围**：203-224 cm，差异约 ±13 cm。

---

## 5. 加宽床型 (Bariatric)

适用：肥胖型患者，美国市场占比 10-20%

| 床型 | 宽度 |
|---|---|
| Standard | 36" (91 cm) |
| Bariatric (基础) | 42" (107 cm) |
| Bariatric (中等) | 48" (122 cm) |
| Bariatric (高级) | 54" (137 cm) |

**关键差异**：宽度从 91cm 到 137cm，**变化最高接近 50cm**。

---

## 6. 对 WiseFido 的具体影响

### 6.1 长度先验仍然可用

```python
BED_LENGTH_PRIOR = 203  # cm (80 英寸标准)
BED_LENGTH_EXTENDED = 213  # cm (84 英寸加长)
BED_LENGTH_LONG = 224     # cm (88 英寸超长)

# 实践：默认使用 203，对 90%+ 场景准确
# 加长版让用户手动选择
```

### 6.2 宽度必须实测，不能用先验

```python
# 床尾两张蓝色 ArUco 实测宽度
bed_width = distance(blue_aruco_1, blue_aruco_2)

# 自动分类
if 86 <= bed_width <= 96:
    bed_type = "standard"      # 91 ± 5 cm
elif 100 <= bed_width <= 145:
    bed_type = "bariatric"     # 107-137 cm 范围
    # 进一步细分（可选）：
    # 107-117: bariatric_basic (42")
    # 117-127: bariatric_medium (48")
    # 127-145: bariatric_premium (54")
else:
    bed_type = "unknown"
    # 提示用户确认
```

### 6.3 床高度必须忽略

**核心原则**：Hi-Low 床高度可变（18-81cm），**绝对高度不可靠**。Fall alarm 判断只能用**水平投影**。

```python
def evaluate_fall_alarm(target_position, target_state):
    if target_state != "Lying":
        return False
    
    # 只用 2D 水平投影判断，不考虑 z 坐标
    for bed_zone in room.lying_zones:
        if bed_zone.polygon_2d.contains(target_position.xy):
            return False  # 在床的水平投影内，正常 Lying
    
    return True  # 不在任何 Lying 区，触发 fall alarm
```

### 6.4 ArUco 卡片放置位置（关键）

由于 Hi-Low 床面会升降，蓝色 ArUco **不能贴在床面上**：

```
❌ 不放：
  - 床面（床垫上方）
  - 任何随床面升降的部位

✅ 推荐：
  - 床尾底座外侧两角（最低处的固定结构）
  - 床头墙面（如果床贴墙，完全静止）
  - 床的金属侧轨底部（不被 sheets 遮挡的位置）
```

---

## 7. APP 标定流程加强（针对 hi-low 床）

### Step：床类型确认

```
APP 标定后显示：
  "识别到床尺寸：91 × 203 cm (标准型)"

如果是病房 / 护理机构场景：
  "这是可升降的医疗床吗？"
  [是] → 启用宽 Lying tolerance 模式
        （考虑床面可能在 18-81cm 之间任何高度）
  [否] → 标准模式
```

### Step：床尺寸异常提示

```
if bed_width > 100 cm:
    show: "检测到加宽床（{bed_width} cm）
          这是 Bariatric 床吗？"
    [是] → bed_type = "bariatric"
           扩大 Lying tolerance
    [否] → 让用户拖动调整尺寸
```

---

## 8. JSON 输出示例

```json
{
  "lying_zones": [
    {
      "type": "bed",
      "subtype": "standard",
      "dimensions": {
        "width": 91,
        "length": 203,
        "unit": "cm"
      },
      "height_class": "variable",
      "height_range": {
        "min": 18,
        "max": 81,
        "comment": "hi-low adjustable, footprint fixed"
      },
      "polygon": [[x1,y1], [x2,y2], [x3,y3], [x4,y4]],
      "polygon_basis": "blue_aruco_at_bed_foot_base",
      "confidence": 0.95
    }
  ]
}
```

---

## 9. 完整尺寸参考表

```
床型              宽度         长度         市场占比
─────────────────────────────────────────────────
Standard         91 cm        203 cm        70%+
Extended         91 cm        213 cm        ~10%
Long             91 cm        224 cm        <5%
Bariatric 42"    107 cm       203 cm        ~10%
Bariatric 48"    122 cm       203 cm        ~5%
Bariatric 54"    137 cm       203 cm        <5%

高度（Hi-Low 床）：
  最低：18 cm（fall prevention）
  最高：81 cm（caregiver access）
  Travel：~60 cm
```

---

## 10. 国际市场参考（未来扩展）

注意：本文档主要数据为美国市场。若 WiseFido 未来扩展到其他市场，床长度先验需要调整：

| 市场 | 标准床长 |
|---|---|
| 美国 | 203 cm (80") |
| 欧洲 | 200 cm |
| 中国 | 200 cm 或 210 cm |
| 日本 | 195 cm（典型）|

宽度方面，36" (91cm) 是国际通用的医疗床标准，bariatric 范围各市场略有不同。

---

## 11. 数据来源

数据综合自网络搜索（2026-05），主要参考：

- Drive Medical / Invacare / Hill-Rom 等主流厂商规格
- SonderCare 等高端家用护理床品牌
- Medical Bed 行业标准（USA market）
- Assisted Living Facility 采购规范

---

## 12. 集成到 WiseFido Spec 的位置

这些数据已经体现在 `WiseFido_Engineering_Spec_v1.1.1` 的以下章节：

- **§5.4 床矩形生成**：算法逻辑已包含 standard/bariatric 识别
- **§8.1 JSON Schema**：`lying_zones` 字段已支持 `subtype` 和 `dimensions`
- **§9.1 Fall Alarm 逻辑**：明确使用水平投影，忽略高度
- **§12 验收标准**：Pilot 需验证 bariatric 场景识别率

---

## 13. Pilot 测试建议

在 Pilot 阶段，建议至少覆盖以下床型场景：

| 优先级 | 床型 | 测试目的 |
|---|---|---|
| 高 | Standard 91×203 | 验证 90% 主流场景 |
| 高 | Hi-Low Bed (Assisted Living) | 验证高度变化下 fall alarm 不误报 |
| 中 | Bariatric 107+ | 验证宽度自动识别 |
| 低 | Extended Length | 验证长度先验 fallback |

---

**文档版本**：1.0
**最后更新**：2026-05-14
**配套文档**：WiseFido Engineering Spec v1.1.1

---

**END**

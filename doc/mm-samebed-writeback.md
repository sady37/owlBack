# MM `samebed` 回写设计（先验→后验，回写权威关系）

状态：**设计已定调（2026-06-21，架构师拍）**，未落码。推翻 `owl-common/spatial/relmatrix.go` 旧纪律 line 14「运行时学习不回写本矩阵」——见 §1 为何这条 specifically 可推翻。
关联 memory [[mm_relationship_matrix]] / [[mm_per_device_covers_ownership]] / [[dbn_zone_room_joint_obc]] / [[cell_dbn_timescales_stillbox_single_source]]。

---

## 0. 模型

MM 走 **per-(雷达 × 床-index) 空间**（不强造 netip 前缀，复用 covers 现状那套，[[mm_per_device_covers_ownership]]）。

```
MM.samebed[radar r][bed j]:
  初值（冷启）= 纯几何"平均概率" = covers(r,j) ⊗ onbed(j)
                几何有歧义 → 候选 0.5（没分清）
  ↓ 30s「同向 × 时间优向」事件 EMA（吸收新信息；两事件越近越频繁 → 关系越强）
  收敛 = 真实概率（后验）：真同床 → 1 / 不同床 → 0
  ↓ 相对稳定（layout 不变、关系是物理事实，不逐帧抖）
  → 回写进 MM 作权威关系
       下游（吸纳 / belief κ / AreaBed）读这个稳定值
```

**界线（关键）**：

| 进 MM（稳定关系，回写） | 不进 MM（瞬时态，留 DBN / per-frame） |
|---|---|
| `samebed`（30s EMA 后的真实概率）、`covers`、`onbed` | 现在谁在床 = N（M×N / FwAreaID）、新鲜度 `fresh`（解吸纳用） |

---

## 1. 为何这条可推翻 line 14（而 cell→DBN 单向只读不能动）

旧纪律（[[cell_dbn_timescales_stillbox_single_source]]）：慢学地图不该被瞬时判断污染 → 单向只读、不回写。`samebed` 回写**不违反其精神**，因为三条都成立：

1. **可恢复 → 回写不丢信息**。`covers`/`onbed`（几何原子）单独留在 MM；`samebed` 只是 `covers⊗onbed` 冷启 + EMA，任何时候能从几何原子重算冷启。覆盖 `samebed` ≠ 抹掉先验。cell 没有可恢复的几何原子，故 cell 那条不能动。
2. **慢学 ≠ 瞬时态**。EMA 学的是"这台雷达和这张垫是不是同一个人的探测器"（结构事实），**不是**"现在谁在床"（DBN per-frame 后验，那个不进 MM）。时标对，在回写线正确一侧。
3. **结构稳定**。layout 不变时关系是物理事实，不逐帧抖 → 名正言顺作权威关系。

---

## 2. 开码前必须钉死的边界（会反咬）

### ① 反馈环 —— 最危险

EMA 的「事件」**必须是上游原始探测**（radar InBed 事件 ∧ sleepad InBed 事件的物理共现），**绝不能**是已消费 `samebed` 的下游（吸纳路由后的归属）。否则 `samebed 高 → 吸纳路由证据到此床 → 制造更多共现 → samebed 更高` = 自我强化，假关系焊死、错了回不来。
> 大白话：学习的证据得是传感器原话，不能是"我认定他俩同床之后产生的二手证据"。

**实现约束**：事件源取 firmware/raw bed-occupancy（radar FwAreaID 命中床 + sleepad bed_status），在吸纳路由**之前**采样。

### ② `samebed` 只在「互活」事件更新；一侧静默 = 不更新 = 保持原值

这是 `fresh` 与 `samebed` 分家的命门（"无 max 互活门控"，[[dbn_zone_room_joint_obc]]）：
- sleepad 掉线 **不该**把学到的配对清零——那是 `fresh` 管"现在还活着吗"。
- `samebed` 只在**两侧都活跃但矛盾**时才往 0 走。
> 大白话：垫子哑了不代表"它跟这床没关系"了，关系还在；只有它在响却跟雷达对不上，才扣分。

### ③ layout 变 → `samebed` 失效重置成新几何冷启

绑 `covers`/`onbed` 变更失效：床挪了/重画了，旧后验立刻作废，不拿陈旧配对跨 layout 续命。

### ④ 未收敛的 0.5（几何歧义）必须 FN-safe

收敛前下游读到 0.5 = 没分清，此时**绝不能**让"不确定的 `samebed`"去压摔（同 Λ→1 不可判走风险不对称）。冷启期宁可少耦合、多报。

---

## 3. 持久化：内存态（MatrixCache），不落库

按 MM maintainer 纪律（[[mm_relationship_matrix]]「不微服务不落库派生数据」）：`samebed` 回写**进内存 MatrixCache**，不落 DB。重启 → 几何冷启重收敛 30s = **FN-safe**（丢暖启，不丢安全）。

---

## 4. 范围与时序

碰三处：
- **belief κ**：`samebed` 来源从 belief-内 per-Room 瞬时上移为 MM 权威值（κ 读 MM.samebed）。
- **吸纳（B 域）**：吸纳判定读 MM.samebed。
- **MM maintainer**：回写 EMA 结果。

**超出当前 A/B 的 N_r/sleepad 切片**（契约 §1 A/B 都不碰 belief）。→ **另立"MM samebed 回写"任务，排在融合切片 commit 之后**，不塞进当前 A+B 混合未提交工作树。

---

## 5. 验证（真 case，禁 unit test [[validate_real_case_no_unit_tests]]）

- 含 sleepad + radar 同床的 case：观察 `samebed[r][j]` 从冷启（0.5 或几何值）经 30s EMA 收敛到 1（真同床）/0（不同床）。
- 反馈环防护验证：构造"吸纳后路由"不得反喂 EMA（事件源采样点在路由前）——LOG 事件源时戳/类型确认。
- FN-safe：冷启期（未收敛）真摔不被未定 `samebed` 压制。

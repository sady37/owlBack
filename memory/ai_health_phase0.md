---
name: AI_health Phase 0 校准 + 建表 + 双角色评审
description: AI_health 设计文档全量校准 + 13 张表 SQL 已落（未 deploy）+ 护理/精算双角色评审落地的 4 类边界原则与字段重命名；下次接续从 Phase 1 服务骨架开始
type: project
originSessionId: 9b5f78d1-d7cd-46b5-938d-70618577fd85
---
AI_health 项目 Phase 0 完成（2026-05-02）。

## 三条最高级架构原则（[doc/AI_health.md](../../../owl/owlBack/doc/AI_health.md) §0.3）

1. **事后批处理 / 零改实时通路**：仅从 PG 表读（iot_timeseries / alarm_events / sleepace_report /
   roomengine_grid_snapshot / cards / devices），不修改 cardagg / sleepace consumer，不要 RPC
2. **Care, not Treatment**：产品是护理监测（PACE / SNF / HCBS care manager 看的趋势）；
   急性医学信号（cardio / respiratory / SpO2 / Apnea）只能进 monthly trend report，
   **绝不进 realtime alert 通道**；一旦走"医治流程"产品要 FDA 510(k)，研发成本 5-10 倍
3. **Sensor-derived signals only**：只产雷达+床垫可观测的特征；**不持有也不索取**用药 / 体重 /
   实验室结果 / 病种诊断 / cost / claim；outcome / treatment 数据所有权属机构（data controller），
   Wisefido 是 processor。同时解决 HIPAA 最低必要 + 商业边界 + 监管复杂度

## 关键设计决策（用户 2026-05-02 定）

**Phase 0 校准（D1-D4）**：
- **D1 long-sit = C 路线**：iot_timeseries 自算 episode（dx/dy ≤ 15cm 静止 + 8min），默认 sit；
  仅 AreaBed cell → lying；与 verified Fall 时间窗 [-10min, +5min] 重叠的 episode 整段排除（避免与
  fall_count 重复计数）
- **D2 cell tag**：DB 直读 [`roomengine_grid_snapshot`](../../../owl/owlRD/db/36_roomengine_grid_snapshot.sql)
  .payload，解析 cells[i].b[0].Type ∈ AreaType 枚举（[cell.go:17-27](../../../owl/owlBack/wisefido-ai/internal/roomengine/cell.go#L17-L27)）
- **D3 双源融合**：iot_timeseries device_type='Sleepad'/'Radar' AND category='track'；HR/RR/vital_confidence
  都在 data_value JSONB 里（用 jsonb_array_elements 解）
- **D4 seIndex**：ETL `(report::jsonb->0->'summary'->>'seIndex')::numeric`（不动 sleepace parser）
- **facility_id → branch_id**：cards 表只有 branch_id，无 facility_id；全设计 13 处替换

**双角色评审决策（护理专家 + 保险精算师视角，2026-05-02）**：
- **bedsore 改读 sleepace alarm**（撤销自算 turn_over < 5）：
  直接 COUNT alarm_events.event_type IN ('NoBodyMove','NoTurnOver','AbnormalBodyMovement')；
  sleepace 厂家算法已含个体校准，自算反而引入误判；NoTurnOver 直接对应"2H 翻身护士提醒"
- **晨起步速 15min = "最差时刻" by design**：故意捕老人晨起最不稳定时刻（ankle stiffness）；
  衰弱越早期越在最不稳定时段先暴露；主指标改 `gait_morning_p10_cms` (P10)，p50 降为辅助；
  全天稳态步速彻底退出 frailty 评分
- **acute_watch_score 不进 realtime**：护理 ≠ 医疗；只在 monthly trend 报表呈现，
  绝不触发 RN 24h 响应 webhook；40 SQL 加 CHECK 约束 `chk_alert_type_non_medical` 在 DB 层强制
- **`health_score` → `vitality_preservation_score`**（合规决策）：
  从"FICO 风格 health score"改为"个人 baseline 偏移度 / 活力保持度"；避开 NAIC / HIPAA Title I
  pre-existing condition discrimination；attestation bundle 显示名 = "Vitality Preservation"；
  scope_metric_keys 严禁 `health_score` 字符串
- **outcome 保留原 4 类大 enum + 加 outcome_metadata JSONB**：撤销扩展 readmission_30d /
  snf_admission / 病种细分；机构在回填时自由塞 `{"icd10":"...","readmit_30d":true}`，
  我们存而不解析

## 已撤销的字段建议（用户判定为不可实现 / 越界）

- ❌ `outcome_cost_usd / outcome_los_days`（保险公司内部数据，HIPAA 最低必要）
- ❌ `meds_change_*`（用药 = PHI 最敏感类，sensor 推不出，elder care 主动给会违规）
- ❌ `weight_kg / pain_score`（雷达+床垫物理上测不出体重，秤不在产品规划）
- ❌ 新建 `health_score_weights` 表 → 简化为单列 `stability_weights_version VARCHAR(20) DEFAULT 'equal_v1'`
  （calibration 调权需要 ground truth outcomes，归机构 / academic partner 做）
- ❌ `adl_indicators JSONB` 占位（保险公司内部事，我们不掺和）

## 关键校准发现（doc/AI_health.md §10.2 / §10.3）

- alarm_events.event_type 实际只 13 种，**无 LongSit / ProlongedStay / Radar_Fall**
- alarm_events.trigger_data 无 duration_sec；用 `(event_end - event_since)/1000`（毫秒）
- alarm_events.operation 字段：`verified` / `false_alarm` / `test` / `''`
- alarm_events 有 `iot_timeseries_id` 列直接关联触发 timeseries 行
- iot_timeseries.timestamp 是 bigint (Unix ms) 不是 tz；分区表（42 child）
- iot_timeseries device_type 列直接区分 `Sleepad` / `Radar`，免 JOIN
- POSE 编码：walking=1 / sit=3 / stand=4 / lie=6
- AreaType 枚举：Unknown=0 Enter=1 Bed=2 Sit=3 Active=4 Deny=5 Shower=6 Toilet=7

## 已落地 SQL（已 deploy 到 owlrd 库 2026-05-03）— 13 张表

- [`owlRD/db/38_health_metrics.sql`](../../../owl/owlRD/db/38_health_metrics.sql) — 7 张：daily / monthly /
  baseline / monthly_trend / event_recovery + etl_state / etl_errors
- [`owlRD/db/39_health_cohort.sql`](../../../owl/owlRD/db/39_health_cohort.sql) — 4 张：health_signing_keys（最前，
  cohort/audit 引用） + cohort_health_metrics + institution_contract + data_access_audit
  （后者带 append-only trigger `prevent_audit_mutation`）
- [`owlRD/db/40_health_realtime.sql`](../../../owl/owlRD/db/40_health_realtime.sql) — 2 张：
  realtime_liveness_alert（带 CHECK `chk_alert_type_non_medical`） + visitor_episode
- 方案 2 分文件，无 TimescaleDB（普通表 + 索引）；不加跨文件 FK 保持文件独立性

## 下一步

1. ~~deploy 三个 SQL 到 demo 库~~ ✅ 已 deploy 到 owlrd 库 2026-05-03（13 张表 + 2 trigger 全部建立）
2. ~~创建 `owlBack/wisefido-ai-health/` 服务骨架（Go）~~ ✅ 2026-05-03 完成
3. 按 §9 Phase 1 实现 daily_etl 跑测试 card 30 天验证

DB 连接：`DB_HOST=127.0.0.1 DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=owlrd`（见 owlBack/.env）

## 服务骨架（2026-05-03 落地）— `owlBack/wisefido-ai-health/`

独立 Go 服务，进程隔离于 wisefido-ai（实时通路），module=`wisefido-ai-health`，go 1.25.0。

**目录**：
```
wisefido-ai-health/
├── cmd/health-etl/main.go     # 入口；支持 --once daily|monthly
├── internal/
│   ├── config/                # yaml.v3 + owl-common/config 共享 DatabaseConfig/LogConfig
│   ├── repo/                  # PG 连接池 + GetWatermark/UpsertWatermark/LogError
│   ├── etl/                   # RunDaily / RunMonthly 空 runner（log + watermark）
│   └── scheduler/             # stdlib time.Timer 实现 daily HH:MM + monthly D 号
├── config.yaml                # Denver tz / 02:00 daily / 1 号 03:00 monthly
├── start-health.sh / stop-health.sh
└── go.mod (replace owl-common => ../owl-common)
```

**关键决策**：
- **不引入 robfig/cron**——调度需求只有 daily+monthly 两条，stdlib timer 自实现 ~50 行更易审计
- **--once daily/monthly 模式**：CI / 回填用，立即跑一次然后退出
- **dry_run flag**：true 时只读不写 watermark（本地开发安全网）
- **per_card_timeout=60s / parallel=8**：按设计文档 §8.5 估算（1s/card/天 × buffer）
- **cmd/health-etl/main.go 风格**（与 wisefido-data 对齐），不用根目录 main.go
- **conflict target 用 partial index 形式**：`ON CONFLICT (task_name) WHERE card_id IS NULL` /
  `ON CONFLICT (task_name, card_id) WHERE card_id IS NOT NULL`，对应 38_health_metrics.sql 两个唯一索引

**冒烟测试结果（2026-05-03）**：
- `--once daily`：成功，daily_etl watermark 写入 last_complete_date=2026-05-02 last_status=success
- `--once monthly`：成功，monthly_etl watermark 写入
- 守护模式：daily 算下次 = Denver tz 2026-05-04 02:00；monthly 算下次 = 2026-06-01 03:00；SIGTERM 优雅退出

**下次接续 Phase 1 入口**：[etl/etl.go:54 RunDaily 的 TODO 注释块](../../../owl/owlBack/wisefido-ai-health/internal/etl/etl.go#L54)
列出 4 步：拉 cards → 8 路并行 → 各指标子模块 → INSERT daily_health_metrics ON CONFLICT。

## 剩余 P1/P2（不阻塞建表）

- frailty_velocity 6 月窗口 → 12 月（季节性污染）
- sundowning_index Phase 2（早期痴呆独家信号）
- cohort PMPY normalization Phase 4（精算 KPI 标准）
- peer reference Phase 5（精算定价对照）
- chronic_resonance raw counter 透明（精算可解释）

**Why**：用户偏好讨论对齐再实施 + 不过度工程化，护理/精算双角色评审帮助锁定边界（不越界做用药 /
weight / cost 这些不可实现或越界的字段），合规命名（vitality_preservation_score）+ 边界 trigger
（audit append-only / alert_type CHECK）让 schema 一开始就符合 HIPAA / NAIC / FDA 边界。

**How to apply**：下次接续 AI_health 工作时，先 read 此 memory + doc/AI_health.md；
不要重新设计 schema；如果 SQL 还没 deploy，先确认是否 deploy 再写 ETL；
**任何新字段加之前先用三条原则过滤**（事后批处理 / Care-not-Treatment / Sensor-only），
否则容易踩 PHI 或监管雷区。

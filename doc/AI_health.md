# AI Health — Wisefido 长期健康趋势 + 集团健康认证平台设计

> 内部医学分析（§1–§10）+ 对外集团认证（§11）双层架构。
> 从养老/独居老人存量数据中识别**长期下行趋势** + **独居生活力异常**，输出**风险提示（非诊断）** + 可解释贡献项；以**签名认证报告**形式交付给集团客户。

---

## 0. 概览

### 0.1 公司与市场

- **Wisefido, Inc.** — Denver, CO（美国注册）
- **仅 USA 市场** — HIPAA / HITRUST / SOC 2 / CMS / 州层合规为主线；不考虑 GDPR / 中国 / 国际化
- **完全 B2B / B2G** — 不做 to-C；**老人/家属对评分、报告、数据零可见**
- **三类目标客户**：
  1. 大型养老连锁 + PACE 运营商（首选 Denver 总部 InnovAge）
  2. 州 Medicaid HCBS Waiver 项目（首选 Colorado HCPF）
  3. 商业健康险 / Medicare Advantage / 再保 / LTC（post-VC）

### 0.2 双层架构

```
┌──────────────────────────────────────────────────────┐
│  §11  集团认证层（B2B Disclosure Layer，搁置加注）    │
│       - cohort + member 签名报告（Phase 5+ 实现）    │
│       - institution_contract + BAA                   │
│       - audit log + provenance + anti-fraud          │
└──────────────────────────────────────────────────────┘
                       ▲
┌──────────────────────────────────────────────────────┐
│  §1–§10  内部医学分析层（MVP 主体）                  │
│  - 数据/特征/基线/评分/解释                          │
│  - Tier S 锚点：recovery / frailty velocity /        │
│    stability / chronic resonance / solo-living       │
└──────────────────────────────────────────────────────┘
                       ▲
   原始：iot_timeseries / alarm_events / sleepace_report
```

### 0.3 设计哲学

| 原则 | 含义 | 反例 |
|---|---|---|
| **事后批处理 / 零改实时通路** | 仅从 PG 表读（iot_timeseries / alarm_events / sleepace_report / roomengine_grid_snapshot / cards / devices）；所有特征、cell tag、排除规则都在 ETL 阶段完成；不修改 cardagg / sleepace consumer / 不要 RPC | 为 long-sit 在 cardagg 加新 alarm；为 cell tag 加 RPC 接口 |
| **Care, not Treatment** | 产品是**护理监测**（PACE / SNF / HCBS care manager 看的健康趋势 / 风险提示），不是**医疗决策**（ICU / 急诊 / 用药调整）；急性医学信号（cardio / respiratory / SpO2 / Apnea）只能进 monthly trend report，**绝不进 realtime alert 通道**；一旦走"医治流程"产品就要 FDA 510(k)，研发成本 5-10 倍 | acute_watch_score 触发 RN 24h 响应 webhook；Apnea 计数 → 实时告警；HR drop → ER routing |
| **Sensor-derived signals only** | 我们只产**雷达 + 床垫可观测的特征**（mobility / sleep / vital / restlessness / liveness）；**不持有也不索取**：用药 / 体重 / 实验室结果 / 病种诊断 / cost / claim。outcome / treatment 数据所有权属机构（data controller），Wisefido 是 processor。同时解决 HIPAA 最低必要 + 商业边界 + 监管复杂度 | 让 elder care 主动给我们 meds 列表；从机构 EHR 拉 ICD-10 / lab；建 weight_kg 字段（雷达测不出） |
| **长趋势 > 单点阈值** | 多周一致漂移才报黄；单次跳变只入波动指标 | 心率今天 95，立刻报警 |
| **个人基线 > 人群参考** | 每人 90–180 天滚动 median/IQR 作为参照系 | 用 60–100 bpm 一刀切 |
| **多信号共振 > 单维度** | mobility + sleep + vital 同向恶化，置信度才高 | 只看心率波动就出风险分 |
| **可解释 > 黑盒** | 任何风险分都必须能拆出 top-N 贡献项与方向 | 直接给"心血管风险 0.78" |
| **风险提示 ≠ 诊断** | 输出永远是 hint / drift / suggestive，不出疾病名 | "高血压风险" |
| **数据稀疏可接受** | 缺失 → 降低 confidence，不要假填补 | 用全院平均替代该 card 的缺失日 |
| **B2B-only, no to-C** | 老人/家属看不到任何评分/原始/导出 | App 让老人查自己的健康分 |
| **集团内可见、不可导** | 集团 dashboard 只能查阅，导出由 Wisefido 渲染签发 | 机构按钮一键导 CSV |
| **vital_confidence ≥ 36** | sleepace `signal_quality × 18`，≥ 36 等价 signal_quality ≥ 2 | 全收 raw 数据进 baseline |

### 0.4 MVP 边界（融资前 / 融资后）

| **Pre-VC（必做）** | **Post Series A** |
|---|---|
| §1–§10 数据 / 特征 / 评分公式 | OAuth 数据中介层（Plaid 风格） |
| 集团内部 dashboard | 公开 verify endpoint |
| 签名 PDF 报告（cohort + member） | 完整 token / scope / revocation 系统 |
| 机构合同 + BAA 轻量管理 | Databus / SDK / embedded UI |
| 审计日志（HIPAA 7 年） | 行为级 anti-fraud / ML |
| 反欺诈最小集（设备 fingerprint + provenance hash） | 多州扩展 |
| Denver 三角 1–3 pilot（Phase 5） | 再保 RWE 数据池（Phase 7） |

### 0.5 Tier S 价值锚点（这套系统的护城河）

设计必须 double-down 在这五个**别人复制不了**的信号上：

| 锚点 | 独家性来源 | 客户价值 | 设计位置 |
|---|---|---|---|
| **1. 恢复力曲线** | 家居 24×365 连续轻负荷恢复，医院只能瞬时态测 | 心血管事件 6–12 月前兆；护理等级依据 | §3.6 |
| **2. 衰弱速度** | 步速年化下降率，月级连续监测全行业空白 | LTC 精算硬指标；机构护理升级触发器 | §3.7 |
| **3. 稳定性指纹** | 个人 baseline 偏离 zone time（不是均值） | 续保/加保定价；机构续费证据 | §3.8 |
| **4. 慢性多信号共振** | 多模态同向漂移，单维度数据保险买不到 | 大病预测最值钱信号；科学照护卖点 | §3.9 |
| **5. 独居生活力** | 独居老人无家属/无可穿戴，被动雷达是唯一可获得 | 州 Medicaid HCBS 评估、APS 干预、未发现死亡预警 | §3.11 |

### 0.6 7 个补充洞察（同等重要，非 Tier S 但显著拉宽护城河）

| # | 洞察 | 临床/商业价值 | Phase | 设计位置 |
|---|---|---|---|---|
| A | **体动 / 翻身利用** | 翻身 < 5 次/夜 = 褥疮高风险（机构 #1 担责事件） | 0 | §3.12 |
| B | **Restlessness** （pose 转换 + room 过渡） | 谵妄 / 慢性疼痛 / 抑郁早期 — 家居场景独家 | 1 | §3.13 |
| C | **数据缺失模式** | "老人是否在屋"作为行为信号 | 1 | §3.14 |
| D | **Chronotropic Competence** （HR + Mobility 同步） | 走路时 HR 不升 = 心衰 / deconditioning 强信号；心血管储备金标准 | 1–2 | §3.15 |
| E | **HR 突降事件** （pre-syncope） | 跌倒前兆独家信号；fall 责任溯源（vasovagal vs 真摔倒） | 2 | §3.16 |
| F | **Postprandial Nap HR** （午餐后小睡） | autonomic neuropathy / 代谢储备代理；糖尿病老人命中率高 | 2 | §3.17 |
| G | **Calibration Loop + Outcome 字段** | 用真实结局回喂权重；架构性护城河，越久越深 | 3 + 持续 | §5 + §10 |

---

## 1. 数据源与维度映射

### 1.1 四类原始数据

| 类别 | 来源表 | 关键字段 | 现状 |
|---|---|---|---|
| **轨迹/姿态** | [iot_timeseries](../../owlRD/db/18_iot_timeseries.sql) | `data_value` JSONB 数组（含 position_x/y/z, pose, ts_ms, heart_rate, respiratory_rate, vital_confidence, bed_status, body_move, turn_over）；`category` 区分 `vital-signs` / `activity` / `track` | TimescaleDB 候选超表，按 `timestamp` 1 天 chunk |
| **报警事件** | [alarm_events](../../owlRD/db/22_alarm_events.sql) | `event_type`, `category` (safety/clinical/behavioral/device), `triggered_at`, `hand_time`, `trigger_data` JSONB | 已成熟，**无 card_id 列**——需 `device_id JOIN devices.bed_id JOIN cards` 反查 |
| **睡眠报告** | [sleepace_report](../../owlRD/db/29_sleepace_report.sql) | `date` (YYYYMMDD), `start_time`/`end_time` (Unix s), `sleep_state` TEXT(JSON int 数组), `report` TEXT(JSON) | 每晚一条；周聚合 `WeeklySleepEfficiency` 已实现 |
| **生命体征流** | iot_timeseries（同表，sleepace 写 `category=track` / 雷达写各自） | `data_value[i].heart_rate`, `respiratory_rate`, `vital_confidence` | sleepace 默认 10s（adaptive sampling 后夜 2s + 午睡 2s）；雷达 vital 仅床上 2s |

### 1.2 数据采样频率（最终）

| 源 | 信号 | 频率 | 时间窗 |
|---|---|---|---|
| Qinglan 雷达 | Track（pos + pose） | 1s（有人）/ 30s（无人 ID=88） | 全天 |
| Qinglan 雷达 | HR / RR | **2s** | **仅床上时段** |
| Sleepace 床垫 | 实时 HR / RR | 默认 10s（adaptive sampling 工单后夜睡 2s + 午睡 2s） | 床上时段 |
| Sleepace 床垫 | 整夜 report | 1 次/晚 | 整夜聚合 |

**双源融合策略（HR / RR 共用）**：
- 主源 sleepace；备源 radar
- ETL 按 2s bucket：`COALESCE(sleepace_value, radar_value)`
- 不做双源 cross-validation（保留简单可调试）

### 1.3 核心维度：card_id 统一

| 表 | card_id 来源 | 处理 |
|---|---|---|
| `iot_timeseries.card_id` | VARCHAR(100) | ETL 入聚合表前 cast/lookup 到 UUID |
| `alarm_events`（无列） | `device_id → devices.bed_id → cards` JOIN | 在 ETL SQL 显式 JOIN |
| `sleepace_report.card_id` | UUID | 直接用 |

> 所有聚合表 / 报告 / 评分 → `card_id UUID` 唯一维度。次级维度 `tenant_id, resident_id, branch_id, cohort_id` 冗余存储。

### 1.4 数据可获得性矩阵

| 信号 | 必须 | 缺失策略 |
|---|---|---|
| 步速（gait_speed） | radar 在线 + 老人在 unit 内活动 | 缺失日 → 7 日 median 占位仅用于趋势线，confidence 减分 |
| HR / RR（夜间） | sleepace 在床（主）/ 雷达床上（备） | 双源都缺 → 当晚 vital 全 NULL |
| sleep_efficiency | sleepace 出报告 | 缺失 → 该日不参与 sleep block；月内合格日 < 18 → confidence < 0.6 |
| fall / long-sit count | alarm_events 写入 | 默认 0；无需占位 |
| HRV (RMSSD-style) | 需 IBI（**厂家不给**） | 自算 RMSSD-proxy（HR 序列相邻差分），标 `source='computed_proxy'` |
| 完全停呼吸 (apnea) | alarm_events `ApneaHypopnea` | 默认 0；CheckAH 算法依赖 sleepace ≤2s 采样（adaptive sampling 工单上线前结果偏低） |
| Liveness（独居） | radar 至少在场探测 | 设备掉线 ≥ 30min → 单独"设备离线"事件，不计入 inactive_streak |
| Pattern drift（独居） | 30+ 天数据训练个体节律 | 冷启动期不出 pattern drift 字段 |

> **vital_confidence 阈值 ≥ 36**（= signal_quality ≥ 2 / 0–5 评分中等以上质量）；Phase 1 实测后调整。

### 1.5 独居场景额外数据需求

PACE / state Medicaid HCBS 客户专属：

| 信号 | 来源 | 用途 |
|---|---|---|
| 设备在线/离线状态 | iot_timeseries 心跳 + qinglan/sleepace 健康检查 | 区分"老人静止"vs"设备故障" |
| Cell-level 在场 | roomengine 输出 cell 占用 | 区分卫生间滞留 / 厨房久坐 / 沙发久坐 |
| 多人 vs 单人在场 | radar track 数 | 访客识别（caregiver / 家属） |
| 入睡/起床节律 | sleepace_report.start_time / end_time | 每日规律性偏差（核心 #1 子项） |
| 用餐时段在厨房 cell 在场 | roomengine cell tag = kitchen | 用餐节律（认知衰退最早期信号之一） |

---

## 2. 数据层：聚合表 schema

### 2.1 整体分层

```
原始：iot_timeseries / alarm_events / sleepace_report
   │
   ├── 实时 │ liveness_streamer ──▶ realtime_liveness_alert    (T 秒级)
   │
   ▼ 日 ETL（T+1 02:00）
daily_health_metrics                  ← 每天每 card 一行
health_event_recovery                 ← 每个 bathroom episode 一行
visitor_episode (INTERNAL_ONLY)       ← 每次访客 episode 一行
   │
   ▼ 月 ETL（每月 1 号 03:00）
monthly_health_metrics                ← 月度聚合
health_baseline                       ← 滚动 90/180 天 baseline
   │
   ▼ 月评分 + 池子聚合（每月 1 号 04:00）
monthly_health_trend                  ← 个人级风险分 + 解释
cohort_health_metrics                 ← 集团池子级分布

每次 API 拉取 → data_access_audit (HIPAA 7 年保留)
```

### 2.2 `daily_health_metrics`

```sql
CREATE TABLE daily_health_metrics (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    card_id         UUID NOT NULL,
    resident_id     UUID,
    branch_id     UUID,
    date_key        DATE NOT NULL,

    -- ============ Mobility block ============
    -- 主指标：晨起步速（首 15 min）
    gait_morning_p50_cms             REAL,
    gait_morning_p10_cms             REAL,
    gait_morning_sample_count        INTEGER,
    -- 起床速度（Tier S）— 仅 sit-up→leftbed (sleep_state 6→1)
    sit_to_stand_sec_p50             REAL,
    sit_to_stand_max_sec             REAL,
    sit_to_stand_attempts_p50        REAL,
    -- 全天活动量（辅助 / 季节性参考）
    walking_duration_sec             INTEGER,
    walking_distance_m               INTEGER,
    walking_step_count_estimate      INTEGER,
    gait_speed_p50_cms_full_day      REAL,
    gait_speed_p10_cms_full_day      REAL,
    -- Fall
    fall_count                       SMALLINT DEFAULT 0,
    fall_total_duration_sec          INTEGER  DEFAULT 0,
    -- 卫生间（日夜分桶）
    bathroom_visit_count_day         SMALLINT DEFAULT 0,
    bathroom_visit_count_night       SMALLINT DEFAULT 0,    -- 22:00-06:00 nocturia
    bathroom_visit_duration_p50_day  INTEGER,
    bathroom_visit_duration_p50_night INTEGER,
    bathroom_visit_max_duration_sec  INTEGER  DEFAULT 0,
    -- Long-sit (病理) — 仅日间，夜间睡觉不计
    long_sit_count_day               SMALLINT DEFAULT 0,
    long_sit_max_duration_sec        INTEGER  DEFAULT 0,
    long_sit_total_sec_day           INTEGER  DEFAULT 0,
    -- Cell-tagged long-sit（路径 A/B 待校准；见 §3.1）
    long_sit_sofa_sec                INTEGER  DEFAULT 0,
    long_sit_bed_sec                 INTEGER  DEFAULT 0,
    long_sit_bathroom_sec            INTEGER  DEFAULT 0,
    long_sit_kitchen_sec             INTEGER  DEFAULT 0,
    long_sit_other_sec               INTEGER  DEFAULT 0,

    -- ============ Sleep Architecture block ============
    sleep_efficiency_pct             REAL,                   -- sleepace seIndex
    sleep_duration_min               INTEGER,
    sleep_onset_latency_sec          INTEGER,                -- sleep_state 2→3/4 转换耗时
    waso_count                       SMALLINT,               -- 夜间觉醒（state 2）次数
    waso_total_sec                   INTEGER,
    deep_sleep_pct                   REAL,                   -- state 3 占比
    light_sleep_pct                  REAL,                   -- state 4 占比
    sleep_fragmentation_idx          REAL,
    sit_up_count                     SMALLINT,               -- state==6 计数
    situp_to_leftbed_median_sec      INTEGER,
    -- Daily Rhythm Drift（合并自原 #4 Cognition 缩水版）
    sleep_onset_time_iqr_days_30     REAL,                   -- 过去 30 天入睡时间 IQR (h)
    wake_time_iqr_days_30            REAL,
    sleep_duration_iqr_min_30        REAL,

    -- ============ Cardiovascular Vital block (双源 fallback / 锁夜窗) ============
    hr_night_p50_bpm                 REAL,
    hr_night_cv                      REAL,
    hr_rmssd_proxy_bpm               REAL,
    hr_rmssd_source                  VARCHAR(20) DEFAULT 'computed_proxy',
    hr_spike_count                   SMALLINT,
    hr_drop_event_count              SMALLINT DEFAULT 0,     -- §3.16 pre-syncope 候选
    rr_night_p50_rpm                 REAL,
    rr_night_cv                      REAL,
    rr_spike_count                   SMALLINT,
    hr_rr_corr_night                 REAL,
    sleep_period_hr_amp_bpm          REAL,                   -- 入睡前 30min vs 深睡 HR 落差
    -- 双源覆盖统计
    hr_primary_minutes_night         INTEGER,                -- sleepace 直供分钟
    hr_fallback_minutes_night        INTEGER,                -- radar 兜底分钟
    hr_total_coverage_pct            REAL,
    rr_primary_minutes_night         INTEGER,
    rr_fallback_minutes_night        INTEGER,
    rr_total_coverage_pct            REAL,

    -- ============ Respiratory & SDB block ============
    apnea_complete_count_night       SMALLINT,               -- alarm_events ApneaHypopnea
    apnea_complete_total_duration_sec INTEGER,
    apnea_complete_max_duration_sec  INTEGER,
    cs_suspect_minutes_night         INTEGER,                -- Phase 2: FFT 周期检测
    cs_episode_count_night           SMALLINT,

    -- ============ Recovery — 夜间 bathroom only ============
    bathroom_episode_count_night     SMALLINT,
    bathroom_recovery_hr_pre_p50     REAL,                   -- 离床前 5min baseline HR
    bathroom_recovery_hr_peak_p50    REAL,                   -- 离床后 60s HR peak
    bathroom_recovery_hr_recovery_sec_p50  INTEGER,          -- 回床→回 baseline 耗时
    bathroom_recovery_rr_recovery_sec_p50  INTEGER,
    bathroom_recovery_quality        REAL,                   -- 完整 episode 比例

    -- ============ Postprandial Nap block (洞察 F) ============
    -- 12:15–13:30 时段 sleepace adaptive sampling 切到 2s
    nap_detected_today               BOOLEAN DEFAULT FALSE,
    nap_started_at                   TIMESTAMPTZ,
    nap_duration_min                 INTEGER,
    nap_hr_p50_bpm                   REAL,                   -- 上床 5-30min HR median
    nap_rr_p50_rpm                   REAL,
    postprandial_hr_lift_bpm         REAL,                   -- = nap_hr_p50 - hr_night_p50
    nap_data_quality                 VARCHAR(20),            -- 'captured' / 'too_short' / 'no_nap_today'

    -- ============ Restlessness block (洞察 B) ============
    pose_transitions_per_hour_night  REAL,                   -- 夜间 pose 序列状态变化频率
    room_transitions_per_day         SMALLINT,               -- cell 序列穿越次数
    pre_sleep_restlessness_idx       REAL,                   -- 入睡前 1h pose 转换次数
    -- 体动 / 翻身（洞察 A）— 主信号 = sleepace 已检测的 alarm（直接 COUNT alarm_events）
    no_body_move_alarm_count_night        SMALLINT DEFAULT 0,  -- alarm_events.event_type='NoBodyMove'
    no_turn_over_alarm_count_night        SMALLINT DEFAULT 0,  -- alarm_events.event_type='NoTurnOver' — 直接对应"2H 未翻身护士提醒"
    abnormal_body_move_alarm_count_night  SMALLINT DEFAULT 0,  -- alarm_events.event_type='AbnormalBodyMovement'
    -- 辅助 — sleepace iot_timeseries 字段（sanity check / 不进风险评分）
    turn_over_count_night            SMALLINT,               -- 直接 SUM(turn_over field)；不再用绝对阈值（信任 NoTurnOver alarm）
    body_move_total_count_night      SMALLINT,

    -- ============ Solo-living block (Tier S #5，瘦身) ============
    last_seen_active_at              TIMESTAMPTZ,
    inactive_streak_hours_max        REAL,
    night_at_home_pct                REAL,
    daily_pattern_drift_score        REAL,                   -- 当日规律 vs 30 天 baseline

    -- ============ 数据缺失模式（洞察 C） ============
    daytime_activity_data_pct        REAL,                   -- 白天 activity 数据完整度
    nighttime_vital_data_pct         REAL,                   -- 夜间 vital 数据完整度
    device_intentional_avoidance     BOOLEAN DEFAULT FALSE,  -- 设备拔/避开监测的可疑信号

    -- ============ Visitor block (基础部分对外可见) ============
    visitor_episode_count_day        SMALLINT DEFAULT 0,
    visitor_total_minutes_day        INTEGER DEFAULT 0,
    visitor_max_episode_minutes      INTEGER DEFAULT 0,
    visitor_unique_track_count       SMALLINT DEFAULT 0,

    -- ============ Visitor Intimacy (INTERNAL_ONLY — 不出对外报告/attestation) ============
    intimacy_index_day_p50           REAL,                   -- INTERNAL_ONLY
    intimacy_index_day_max           REAL,                   -- INTERNAL_ONLY
    intimate_minutes_day             INTEGER,                -- INTERNAL_ONLY: < 50 cm 累计
    close_minutes_day                INTEGER,                -- INTERNAL_ONLY: 50-150 cm

    -- ============ 元信息 ============
    samples_hr                       INTEGER,
    samples_rr                       INTEGER,
    samples_track                    INTEGER,
    has_sleep_report                 BOOLEAN,
    device_uptime_pct                REAL,
    data_quality_score               REAL,

    -- ============ Provenance（§11.4） ============
    source_event_count               INTEGER,
    source_device_uids               TEXT[],
    row_hash                         VARCHAR(64),
    prev_row_hash                    VARCHAR(64),

    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),

    UNIQUE (card_id, date_key)
);
CREATE INDEX idx_daily_health_card_date     ON daily_health_metrics(card_id, date_key DESC);
CREATE INDEX idx_daily_health_branch_date   ON daily_health_metrics(branch_id, date_key DESC);
CREATE INDEX idx_daily_health_tenant_date   ON daily_health_metrics(tenant_id, date_key DESC);

-- 字段注释：所有 intimacy_* / intimate_minutes_* / close_minutes_* 字段
-- INTERNAL_ONLY: do not include in attestation bundle / cohort report / PDF / public API
COMMENT ON COLUMN daily_health_metrics.intimacy_index_day_p50 IS 'INTERNAL_ONLY';
COMMENT ON COLUMN daily_health_metrics.intimacy_index_day_max IS 'INTERNAL_ONLY';
COMMENT ON COLUMN daily_health_metrics.intimate_minutes_day   IS 'INTERNAL_ONLY';
COMMENT ON COLUMN daily_health_metrics.close_minutes_day      IS 'INTERNAL_ONLY';
```

### 2.3 `monthly_health_metrics`

```sql
CREATE TABLE monthly_health_metrics (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    card_id         UUID NOT NULL,
    resident_id     UUID,
    branch_id     UUID,
    month_key       DATE NOT NULL,

    days_total      SMALLINT NOT NULL,
    days_with_data  SMALLINT NOT NULL,
    days_quality_ok SMALLINT NOT NULL,

    -- ===== Mobility 月聚合 =====
    gait_morning_p50_month           REAL,
    gait_morning_p10_month           REAL,
    sit_to_stand_sec_p50_month       REAL,
    walking_distance_p50_month_m     INTEGER,
    fall_count_month                 SMALLINT,
    fall_total_duration_sec_month    INTEGER,
    long_sit_mean_duration_sec       REAL,
    long_sit_count_month             SMALLINT,
    long_sit_sofa_pct                REAL,
    long_sit_bathroom_total_sec      INTEGER,
    -- Frailty velocity (Tier S #2)
    gait_decline_annualized_pct      REAL,
    gait_decline_r2                  REAL,

    -- ===== Sleep 月聚合 =====
    sleep_efficiency_month_p50       REAL,
    sleep_onset_latency_p50_month    INTEGER,
    waso_count_p50_month             REAL,
    deep_sleep_pct_p50_month         REAL,
    sit_up_count_month_total         SMALLINT,
    situp_to_leftbed_p50_month       INTEGER,
    sleep_fragmentation_month_p50    REAL,

    -- ===== Cardiovascular Vital 月聚合 =====
    hr_night_p50_month               REAL,
    rr_night_p50_month               REAL,
    hr_cv_month                      REAL,
    hr_rmssd_proxy_month             REAL,
    hr_spike_density_month           REAL,
    hr_drop_event_density_month      REAL,                  -- §3.16
    rr_cv_month                      REAL,
    rr_spike_density_month           REAL,
    sleep_period_hr_amp_month_p50    REAL,                  -- 替代原 circadian
    hr_rr_corr_month                 REAL,

    -- ===== Respiratory 月聚合 =====
    apnea_complete_count_month       SMALLINT,
    apnea_per_hour_night_month       REAL,
    cs_suspect_minutes_month         INTEGER,                -- Phase 2

    -- ===== Recovery 月聚合 (Tier S #1) =====
    bathroom_recovery_hr_p50_month_sec INTEGER,
    bathroom_recovery_episode_count_month SMALLINT,
    recovery_drift_pct               REAL,                   -- vs 180d baseline
    recovery_quality_month           REAL,

    -- ===== Stability signature (Tier S #3) =====
    stability_hr_night_oob_pct       REAL,
    stability_rr_night_oob_pct       REAL,
    stability_gait_p50_oob_pct       REAL,
    stability_sleep_eff_oob_pct      REAL,
    stability_composite_score        REAL,                   -- 默认 4 信号等权 mean
    -- Wisefido 不做 calibration 调权（需 ground truth outcomes，归机构 / academic partner）；
    -- 等权 baseline 由我们出，未来合作方提供训练后权重时升版本号
    stability_weights_version        VARCHAR(20) DEFAULT 'equal_v1',

    -- ===== Chronic resonance (Tier S #4) =====
    chronic_resonance_index          REAL,

    -- ===== 洞察 D — Chronotropic competence =====
    chronotropic_response_p50_month  REAL,                   -- (hr_walk - hr_rest) / hr_rest

    -- ===== 洞察 F — Postprandial =====
    nap_days_count_month             SMALLINT,
    postprandial_hr_lift_month_p50   REAL,
    postprandial_response_decline_pct REAL,                  -- vs 90d baseline

    -- ===== 洞察 B — Restlessness 月聚合 =====
    pose_transitions_per_hour_month_p50 REAL,
    room_transitions_per_day_month_p50  REAL,
    turn_over_count_month_p50        REAL,                   -- 辅助；不进风险评分
    -- 体动 / 翻身月聚合 — 直接来自 sleepace alarm（不再用 turn_over < 5 自算）
    no_body_move_alarm_count_month   SMALLINT DEFAULT 0,
    no_turn_over_alarm_count_month   SMALLINT DEFAULT 0,     -- 替代旧 bedsore_risk_days_count；frailty 评分主输入

    -- ===== Solo-living 月聚合 (Tier S #5) =====
    inactive_streak_month_p95        REAL,
    night_at_home_pct_month          REAL,
    pattern_drift_days               SMALLINT,
    bathroom_visits_per_day_p50      REAL,
    nocturia_episodes_per_night_p50  REAL,
    solo_living_risk_score           REAL,

    -- ===== Social Engagement 月聚合 =====
    visitor_days_count_month         SMALLINT,
    no_visitor_streak_max_days       SMALLINT,
    visitor_episode_total_month      INTEGER,
    -- INTERNAL_ONLY:
    intimacy_index_month_p50         REAL,                   -- INTERNAL_ONLY
    intimate_minutes_month           INTEGER,                -- INTERNAL_ONLY
    care_quality_score               REAL,                   -- INTERNAL_ONLY (依赖 visitor_role 标注)
    family_support_score             REAL,                   -- INTERNAL_ONLY

    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE (card_id, month_key)
);

COMMENT ON COLUMN monthly_health_metrics.intimacy_index_month_p50 IS 'INTERNAL_ONLY';
COMMENT ON COLUMN monthly_health_metrics.intimate_minutes_month IS 'INTERNAL_ONLY';
COMMENT ON COLUMN monthly_health_metrics.care_quality_score IS 'INTERNAL_ONLY';
COMMENT ON COLUMN monthly_health_metrics.family_support_score IS 'INTERNAL_ONLY';
```

### 2.4 `health_baseline`

```sql
CREATE TABLE health_baseline (
    id              BIGSERIAL PRIMARY KEY,
    card_id         UUID NOT NULL,
    snapshot_date   DATE NOT NULL,
    window_days     SMALLINT NOT NULL,            -- 90 / 180

    -- 每个核心指标 median + IQR
    gait_morning_p50_med  REAL,  gait_morning_p50_iqr  REAL,
    sit_to_stand_med      REAL,  sit_to_stand_iqr      REAL,
    walking_distance_med  REAL,  walking_distance_iqr  REAL,
    hr_night_med          REAL,  hr_night_iqr          REAL,
    rr_night_med          REAL,  rr_night_iqr          REAL,
    hr_cv_med             REAL,  hr_cv_iqr             REAL,
    rr_cv_med             REAL,  rr_cv_iqr             REAL,
    sleep_eff_med         REAL,  sleep_eff_iqr         REAL,
    long_sit_mean_med     REAL,  long_sit_mean_iqr     REAL,
    bathroom_recovery_med REAL,  bathroom_recovery_iqr REAL,    -- 180d
    -- 独居场景
    inactive_streak_med   REAL,  inactive_streak_iqr   REAL,
    bathroom_visit_med    REAL,  bathroom_visit_iqr    REAL,
    pattern_baseline      JSONB,                                 -- 个人节律向量
    -- Chronotropic
    chronotropic_response_med REAL, chronotropic_response_iqr REAL,
    -- Postprandial
    postprandial_hr_lift_med REAL, postprandial_hr_lift_iqr REAL,
    -- Restlessness / 体动
    pose_trans_per_hour_med REAL, pose_trans_per_hour_iqr REAL,
    turn_over_count_med   REAL,  turn_over_count_iqr   REAL,

    sample_days     SMALLINT NOT NULL,
    is_cold_start   BOOLEAN  NOT NULL DEFAULT FALSE,
    metadata        JSONB DEFAULT '{}'::JSONB,    -- e.g., {"drift": true}

    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE (card_id, snapshot_date, window_days)
);
```

### 2.5 `monthly_health_trend`（个人级输出，对内）

```sql
CREATE TABLE monthly_health_trend (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    card_id         UUID NOT NULL,
    resident_id     UUID,
    branch_id     UUID,
    month_key       DATE NOT NULL,

    cardio_stress_score    REAL,
    frailty_score          REAL,
    acute_watch_score      REAL,
    solo_living_risk_score REAL,
    vitality_preservation_score SMALLINT,                            -- 0–100 个人 baseline 偏移度（不是绝对健康水平 / 不是核保依据，§5.3）
    sub_scores             JSONB,                     -- {"stability":82,"recovery":75,...}

    confidence_score       REAL,
    top_contributors       JSONB,
    risk_color             VARCHAR(10),               -- green / yellow / red
    risk_persistence_weeks SMALLINT,
    short_clinical_hint    TEXT,

    -- 洞察 G — Calibration loop 字段（等待真实 outcome 回填）
    -- §0.3 Sensor-only 边界：Wisefido 不索取也不解析病种细节 / 成本 / 用药；
    -- 机构（PACE / SNF / 保险）作为 data controller 通过 outcome_backfill 接口回填，
    -- outcome_metadata JSONB 由机构自由扩展（ICD-10 / readmit_30d / cost / etc.），
    -- 我们仅存储不解析、不输出，作为合作方 RWE 训练原材料。
    outcome_event_type     VARCHAR(40),               -- 'hospitalization' / 'er_visit' / 'fall_with_injury' / 'death' / NULL
    outcome_event_at       TIMESTAMPTZ,
    outcome_severity       VARCHAR(20),               -- 'minor' / 'major' / 'critical'
    outcome_lookback_days  SMALLINT,                  -- 此 trend 行预测了"下 N 天"的事件
    outcome_metadata       JSONB,                     -- 机构自由扩展槽位（不解析、不索引、不出对外报表）

    -- Provenance（signing_key_id 引用 §2.12 health_signing_keys 表）
    row_hash               VARCHAR(64),
    bundle_signature       TEXT,
    signing_key_id         UUID,

    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE (card_id, month_key)
);
CREATE INDEX idx_trend_card_month     ON monthly_health_trend(card_id, month_key DESC);
CREATE INDEX idx_trend_branch_color   ON monthly_health_trend(branch_id, risk_color, month_key DESC);
CREATE INDEX idx_trend_outcome        ON monthly_health_trend(outcome_event_type, outcome_event_at) WHERE outcome_event_type IS NOT NULL;
```

### 2.6 `health_event_recovery`（仅夜间 bathroom）

```sql
CREATE TABLE health_event_recovery (
    id              BIGSERIAL PRIMARY KEY,
    card_id         UUID NOT NULL,
    event_type      VARCHAR(40) NOT NULL DEFAULT 'bathroom_round',  -- MVP 只此一种
    event_started_at TIMESTAMPTZ NOT NULL,
    event_ended_at   TIMESTAMPTZ,

    hr_pre_p50      REAL,  hr_peak  REAL,  hr_recovery_sec INTEGER,
    rr_pre_p50      REAL,  rr_peak  REAL,  rr_recovery_sec INTEGER,
    quality_flag    VARCHAR(20),       -- 'complete' / 'no_baseline' / 'no_recovery'

    raw_alarm_event_id UUID,
    cell_tag        VARCHAR(40),
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_event_recovery_card_event ON health_event_recovery(card_id, event_started_at DESC);
```

### 2.7 `cohort_health_metrics`（集团池子级 — §11 输出核心）

```sql
CREATE TABLE cohort_health_metrics (
    id              BIGSERIAL PRIMARY KEY,
    cohort_id       UUID NOT NULL,
    cohort_type     VARCHAR(40) NOT NULL,    -- facility_chain / pace / state_hcbs / ma_plan / aco / ltc_carrier
    cohort_name     VARCHAR(200),
    month_key       DATE NOT NULL,

    member_count                    INTEGER,
    member_active_count             INTEGER,

    -- Health Score 分布
    vitality_preservation_score_p10                SMALLINT,
    vitality_preservation_score_p25                SMALLINT,
    vitality_preservation_score_p50                SMALLINT,
    vitality_preservation_score_p75                SMALLINT,
    vitality_preservation_score_p90                SMALLINT,
    vitality_preservation_score_avg                REAL,

    -- 关注名单 — "low vitality" 表示个人偏移大，**不是**疾病严重程度
    attention_count_below_40        INTEGER,    -- vitality_preservation_score < 40 的成员数
    yellow_color_count              INTEGER,    -- §5.5 RAG 仪表盘标准（trend caution，不是 disease severity）
    red_color_count                 INTEGER,    -- §5.5 RAG（attention required，不是 critical condition）

    -- 关键指标
    fall_count_per_1000_days        REAL,
    acute_alert_count_month         INTEGER,
    frailty_velocity_avg            REAL,
    solo_living_risk_avg            REAL,
    bedsore_risk_member_count       INTEGER,    -- 月内 no_turn_over_alarm_count_month > 0 的人数（直接读 sleepace alarm）

    -- 资源建议
    recommended_escalation_count    INTEGER,
    recommended_home_visit_count    INTEGER,
    top_drivers_in_cohort           JSONB,

    -- Provenance
    cohort_hash                     VARCHAR(64),
    bundle_signature                TEXT,

    created_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE (cohort_id, month_key)
);
CREATE INDEX idx_cohort_month ON cohort_health_metrics(cohort_type, month_key DESC);
```

### 2.8 `institution_contract`（轻量合同表 — §11.2）

```sql
CREATE TABLE institution_contract (
    contract_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    consumer_org_id      UUID NOT NULL,
    consumer_org_name    VARCHAR(200) NOT NULL,
    consumer_type        VARCHAR(40) NOT NULL,
    contract_started_at  TIMESTAMPTZ NOT NULL,
    contract_expires_at  TIMESTAMPTZ NOT NULL,

    scope_cohort_ids     UUID[],
    scope_metric_keys    TEXT[],                  -- 白名单：禁用 INTERNAL_ONLY 字段
    scope_geographic     VARCHAR(100),
    output_cadence       VARCHAR(20),

    baa_signed_at        TIMESTAMPTZ,
    baa_doc_uri          TEXT,
    api_credential_id    UUID,

    realtime_webhook_url TEXT,
    realtime_alert_types TEXT[],

    status               VARCHAR(20) DEFAULT 'active',
    created_at           TIMESTAMPTZ DEFAULT now(),
    updated_at           TIMESTAMPTZ DEFAULT now()
);
```

### 2.9 `realtime_liveness_alert`（独居场景实时告警通道）

```sql
CREATE TABLE realtime_liveness_alert (
    alert_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id         UUID NOT NULL,
    branch_id     UUID,
    alert_type      VARCHAR(40) NOT NULL,
    severity        VARCHAR(20) NOT NULL,
    detected_at     TIMESTAMPTZ NOT NULL,
    snapshot        JSONB,
    delivered_to    UUID[],
    delivery_status JSONB,
    acked_by        UUID,
    acked_at        TIMESTAMPTZ,
    resolution      VARCHAR(60),
    row_hash        VARCHAR(64),
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_liveness_card_time ON realtime_liveness_alert(card_id, detected_at DESC);
CREATE INDEX idx_liveness_unacked    ON realtime_liveness_alert(branch_id, severity, detected_at DESC) WHERE acked_at IS NULL;
```

### 2.10 `data_access_audit`（HIPAA 7 年保留）

```sql
CREATE TABLE data_access_audit (
    audit_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id       UUID NOT NULL REFERENCES institution_contract(contract_id),
    consumer_user_id  UUID,
    api_endpoint      TEXT NOT NULL,
    request_method    VARCHAR(10),
    request_payload   JSONB,
    cohort_id         UUID,
    card_id           UUID,
    bundle_id         UUID,
    request_ip        INET,
    request_signature TEXT,
    response_status   SMALLINT,
    response_signature TEXT,
    responded_at      TIMESTAMPTZ NOT NULL,
    retention_until   DATE NOT NULL
);
CREATE INDEX idx_audit_contract_time ON data_access_audit(contract_id, responded_at DESC);
CREATE INDEX idx_audit_card_time     ON data_access_audit(card_id, responded_at DESC) WHERE card_id IS NOT NULL;
CREATE INDEX idx_audit_retention     ON data_access_audit(retention_until);
```

### 2.11 `visitor_episode`（**INTERNAL_ONLY** — 不出对外报告）

```sql
CREATE TABLE visitor_episode (
    episode_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID NOT NULL,
    card_id               UUID NOT NULL,
    branch_id           UUID,
    started_at            TIMESTAMPTZ NOT NULL,
    ended_at              TIMESTAMPTZ,
    duration_sec          INTEGER,

    resident_track_id     INT,
    visitor_track_ids     INT[],
    visitor_count_max     SMALLINT,

    -- 亲密度时序聚合（INTERNAL_ONLY）
    intimacy_index        REAL,
    min_distance_cm       INTEGER,
    intimate_seconds      INTEGER,        -- < 50 cm
    close_seconds         INTEGER,        -- 50-150 cm
    present_seconds       INTEGER,        -- 150-300 cm
    distant_seconds       INTEGER,        -- > 300 cm
    cell_distribution     JSONB,

    -- 机构标注
    visitor_role          VARCHAR(20),    -- 'caregiver' / 'family' / 'unknown'
    notes                 TEXT,

    created_at            TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_visitor_episode_card_time ON visitor_episode(card_id, started_at DESC);

COMMENT ON TABLE visitor_episode IS 'INTERNAL_ONLY: All fields excluded from attestation bundle / cohort report / public API';
```

### 2.12 `health_signing_keys`（ed25519 密钥轮换历史 — §11.4）

```sql
CREATE TABLE health_signing_keys (
    key_id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    algorithm           VARCHAR(20) NOT NULL DEFAULT 'ed25519',
    public_key_pem      TEXT NOT NULL,
    kms_key_arn         TEXT,                   -- AWS / GCP KMS 引用（私钥位置）
    activated_at        TIMESTAMPTZ NOT NULL,
    retired_at          TIMESTAMPTZ,            -- NULL = 当前活跃
    rotation_reason     VARCHAR(60),            -- 'annual_rotation' / 'compromise' / 'algorithm_upgrade'
    notes               TEXT,
    created_at          TIMESTAMPTZ DEFAULT now()
);
-- 当前活跃 key 全表唯一
CREATE UNIQUE INDEX ux_signing_keys_active
    ON health_signing_keys((true)) WHERE retired_at IS NULL;
```

设计目的：
- HIPAA / SOC 2 audit 要求"哪个 key 签了哪个 bundle / response"5-7 年可追溯
- `monthly_health_trend.signing_key_id` 和 `data_access_audit.signing_key_id` 引用本表
- 私钥永远在 KMS（GCP/AWS），本表只存公钥 + 元数据；retire 后**永不删除**

---

## 3. 特征公式

### 3.1 Mobility（4 子项 — 用户定稿）

#### 晨起步速（首 15 min）— 主指标 ⚠️ "最差时刻" by design

> **设计意图**（用户 2026-05-02 确认）：15 min 短窗 = **故意**捕"最差时刻"步速，
> **不是**稳态步速。物理逻辑：老人晨起 ankle stiffness / 平衡未恢复 / 找拐杖 →
> 这是个体每天最不稳定的时刻，**衰弱越早期越能在最差时刻先暴露征兆**（稳态步速对早期 frailty 不敏感，类似 TUG test 思想）。
>
> 因此：**`gait_morning_p10_cms` (10 分位) 是核心信号**（更尾部 → 更早期暴露）；
> `gait_morning_p50_cms` 是辅助；`gait_speed_p50_cms_full_day` 仅做季节性 / sanity check 参考，
> **不进 frailty 评分**。

**起床时刻识别**：
- 主：`alarm_events.event_type='SleepPad_LeftBed'` 的 `triggered_at`
- 兜底：sleep_state 数组从睡眠态(2/3/4)→离床态(1) 的转换 = `start_time + i × time_step`

**步速计算**（基于 iot_timeseries position 差分，`Track` struct 无 vx/vy）：
- 关键参数：POSE_WALK=1（[cell.go:46](../../owlBack/wisefido-ai/internal/roomengine/cell.go#L46)）、MIN_DT 200ms、MAX_DT 2000ms、SPEED 范围 [15, 200] cm/s
- 取 `[起床时刻, 起床时刻+15min]` 窗口内 walking 样本
- `gait_morning_p10_cms` = 该窗 P10（**主指标 — 最差时刻**）
- `gait_morning_p50_cms` = 该窗 P50（辅助 — 中位数代表当时整体状态）

**全天步速作辅助**（不进风险评分，仅作季节性 / sanity check 参考）：
- `gait_speed_p50_cms_full_day` / `gait_speed_p10_cms_full_day`

**临床区间**（参考 [walking_speed_health.md](walking_speed_health.md)）：< 60 危险 / 60–80 衰弱 / > 100 健康——前端着色和评分输入。配合"最差时刻"语义：
- p10 < 60 cm/s → 当日最不稳定时段已落入危险区，frailty 早期信号
- p10 vs 个人 baseline 漂移 → frailty velocity 主输入（§3.7）

#### 起床速度（Sit-to-Stand）— Tier S 独立指标

**MVP 仅做 A 版**（sleepace 床垫数据）：
- sit_up = sleep_state 状态 6
- left_bed = sleep_state 状态 1（或 alarm_events SleepPad_LeftBed）
- 单次配对：每个 sit_up 找其后 30 min 内最近的 left_bed
- `sit_to_stand_sec_p50` = 当日多次起床的中位时长
- `sit_to_stand_max_sec` = 当日最长一次（最差时刻指标）
- `sit_to_stand_attempts_p50` = 单次起床中 sit_up 状态切回又重启的次数（"尝试-停顿"代理）

**B 版**（雷达 pose 序列细化）：Phase 2+ 实现，当前 sleepad/radar 精度不够。

**临床判读**：
- 总时长 > 30s → 衰弱 / 帕金森
- 多次"尝试-停顿-重试" → 痴呆早期

#### 卫生间使用（日 / 夜分桶）

数据来源：路径 A（`alarm_events.trigger_data.area_type` 已写入）/ 路径 B（roomengine RPC 查 cell tag）— 待 §1.4 校准 #5。

夜间 = 22:00–06:00（院区时区）。临床阈值：
- `bathroom_visit_count_night ≥ 3` → 心衰 / 前列腺 / 糖尿病强信号
- `bathroom_visit_max_duration_sec > 1800`（30 min）→ 卫生间跌倒 / 失能预警
- `bathroom_visit_count_day < baseline - 2σ`（脱水）老人尤其危险

#### 累积步长（距离主 + 步数估算）

**距离**（精确）：
```
walking_distance_m = SUM(distance_between_adjacent_walking_samples) / 100
                    -- 仅在 valid_speeds（速度过滤后）样本上累加
```

**步数估算**（粗略）：
```
walking_step_count_estimate = walking_distance_m × 100 / avg_step_length_cm
                              avg_step_length_cm = 55  (老人经验值，Phase 1 个体化校准)
```

#### Fall（来自 alarm_events）

实际 schema 校准（§10.2）：
- alarm_events 实际 event_type 只有 `'Fall'`（没有 `'Radar_Fall'`）
- trigger_data 没有 `duration_sec` 字段；时长用 `event_end - event_since`（毫秒）
- `operation` 字段值：`verified` / `false_alarm` / `test` / `''`

```sql
fall_count = COUNT(*) FILTER (WHERE event_type = 'Fall')
fall_total_duration_sec = SUM(
    GREATEST(0,
      ((trigger_data->>'event_end')::bigint - (trigger_data->>'event_since')::bigint) / 1000
    )
  ) FILTER (WHERE event_type = 'Fall' AND trigger_data ? 'event_end')
```

#### Long-sit — iot_timeseries 自算（不依赖 alarm_events）

实际 alarm_events 没有 LongSit / ProlongedStay 事件类型，且 long-sit 是事后批处理特征，**完全从 iot_timeseries 自算**。

**Episode 检测算法**（按 card / day / track_id 分组）：
```
1. 取 iot_timeseries WHERE card_id=... AND device_type='Radar' AND category='track'
   字段：position_x, position_y, position_z, pose, ts (timestamp ms)

2. 1Hz 滑窗（5s 窗）判定"静止"：
   max(position_x in window) - min(position_x in window) <= 15 cm
   AND max(position_y in window) - min(position_y in window) <= 15 cm
   注：dx/dy ≤ 15cm 容忍小幅扰动（雷达多径 / 微动），不要求完全不动

3. 同 track_id 连续静止段 ≥ 8 min → 一个 long_sit episode
   episode = {start_ts, end_ts, duration_sec, center_x, center_y}

4. Pose 推断（默认 sit；查 grid_snapshot 例外）：
   episode_pose = 'sit'                       -- 默认
   IF cell_tag_at(unit_id, center_x, center_y) == AreaBed:
       episode_pose = 'lying'                  -- 仅床区例外
   -- 沙发(AreaSit) / 卫浴(AreaToilet/Shower) / 其他 都标 'sit'

5. Verified Fall 排除（避免与 fall_count 重复计数）：
   排除任何与 verified Fall 时间窗 [triggered_at - 10min, triggered_at + 5min]
   重叠的 episode（整段 episode 不计入 long_sit_*）
   verified 定义：alarm_events.event_type='Fall' AND operation='verified'

6. 仅日间统计（夜间睡觉时段不计 long-sit）：
   day_window = 06:00–22:00（院区时区）
   episode 起点落在 day_window 内才计入 long_sit_count_day
```

**字段产出**：
```
long_sit_count_day        = COUNT(episodes WHERE pose='sit' AND start IN day_window)
long_sit_max_duration_sec = MAX(episode.duration_sec WHERE start IN day_window)
long_sit_total_sec_day    = SUM(episode.duration_sec WHERE start IN day_window)
```

#### Cell-tagged long-sit（DB 直接读 grid_snapshot，不要 RPC）

数据源：[`roomengine_grid_snapshot`](../../owlRD/db/36_roomengine_grid_snapshot.sql) — 每房间一行，每 5min engine 全量 dump 学习状态。

**`cell_tag_at(unit_id, x, y)` 的 ETL 实现**（Go batch 查询）：
```go
// 1. 加载 unit_id 关联的所有房间 grid_snapshot.payload
//    (cards.unit_id → rooms.unit_id 反查)
// 2. payload.grid {w, h, ox, oy} 决定 cell index = col*W + row
//    col = (x - ox) / 10,  row = (y - oy) / 10  (10cm 量化)
// 3. 取 cells[i].b[0] = [type, conf, source]
//    type ∈ {0=Unknown, 1=Enter, 2=Bed, 3=Sit, 4=Active, 5=Deny, 6=Shower, 7=Toilet}
// 4. cell_tag 映射（只取 Belief[0]，conf < 30 的退化为 'other'）：
//    AreaBed     → 'bed'
//    AreaSit     → 'sofa'        (Sit 区学到的是沙发/椅子)
//    AreaToilet  → 'bathroom'
//    AreaShower  → 'bathroom'
//    AreaActive  → 'walk'
//    AreaEnter   → 'door'
//    AreaDeny    → 'deny'         (track 不应出现)
//    AreaUnknown → 'other'
```

**字段产出**（按 episode 中心点 cell tag GROUP BY）：
```
long_sit_sofa_sec     = SUM(episode.duration WHERE cell_tag='sofa')      -- 不进 frailty
long_sit_bed_sec      = SUM(episode.duration WHERE cell_tag='bed')
long_sit_bathroom_sec = SUM(episode.duration WHERE cell_tag='bathroom')  -- 卫生间久留 → still-fall 信号
long_sit_kitchen_sec  = 0  -- (kitchen 在 grid_snapshot 没有专属 AreaType，需 layout_config 补 cell.tag='kitchen' 元信息；P2 再做)
long_sit_other_sec    = SUM(episode.duration WHERE cell_tag IN ('other','walk','door','deny'))
```

> 沙发久坐 (`long_sit_sofa_sec`) **不计入** "病理 long-sit burden"——只 `long_sit_bed_sec / long_sit_bathroom_sec` 进入 frailty 评分。
> Kitchen tag 当前 `roomengine_grid_snapshot` 不区分（AreaActive 涵盖一切活动区），需 layout_config 单独标，Phase 2 再做。

### 3.2 Sleep Architecture（核心 #1）

数据源：sleepace_report.sleep_state（int 数组，状态码：0=未监测，1=离床，2=清醒，3=深睡，4=浅睡，6=坐起；时间 = `start_time + i × time_step`，默认 60s）。

| 字段 | 公式 |
|---|---|
| `sleep_onset_latency_sec` | 从 sleep_state 首个 2(清醒) → 首个 3/4(深/浅睡) 的耗时 |
| `waso_count` | 入睡后状态 2(清醒) 出现次数（不含入睡前后） |
| `waso_total_sec` | 状态 2 累计时长 × time_step |
| `deep_sleep_pct` | COUNT(state==3) / COUNT(all valid) |
| `light_sleep_pct` | COUNT(state==4) / COUNT(all valid) |
| `sleep_fragmentation_idx` | `1 - longest_continuous_3_or_4_run / total_sleep_sec` |
| `sit_up_count` | COUNT(state==6) |
| `situp_to_leftbed_median_sec` | 同夜 sit-up→left-bed 配对差中位值 |
| `sleep_efficiency_pct` | ETL 直接 jsonb 解：`(report::jsonb->0->'summary'->>'seIndex')::numeric`（[report_service.go:50-62](../../owlBack/wisefido-sleepace/internal/service/report_service.go#L50-L62) 当前 parser 未读 seIndex；不动业务服务，ETL 自取） |

**Daily Rhythm Drift**（合并自原 #4 Cognition 缩水版）：
```
sleep_onset_time_iqr_days_30 = IQR(过去 30 日 sleepace.start_time of-day, in hours)
wake_time_iqr_days_30        = IQR(过去 30 日 sleepace.end_time of-day, in hours)
sleep_duration_iqr_min_30    = IQR(过去 30 日 sleep_duration_min)
```
节律 IQR > 1.5h 多周持续 → 认知衰退 / 抑郁早期信号。

### 3.3 Cardiovascular Vital — 双源 fallback / 锁夜窗（核心 #2）

**夜间窗**：`sleepace_report.start_time / end_time` 圈定（精准）；缺报告则 22:00–06:00 院区时区。

**实际 schema**（§10.2 #6/#9 校准）：
- iot_timeseries 有 `device_type` 列直接区分（`'Sleepad'` / `'Radar'`），免 JOIN
- 两源 HR/RR/vital_confidence **都写在 `category='track'`** 里（同一行 data_value JSONB 包含 `heart_rate / respiratory_rate / vital_confidence` 字段）
- `category='heart'` 仅极少量纯 vital 行，本设计忽略

**双源融合**（按 2s bucket）：
```sql
WITH sleepad_2s AS (
  SELECT
    (timestamp / 2000) * 2000           AS bucket_ms,
    AVG((d->>'heart_rate')::numeric)    AS hr,
    AVG((d->>'respiratory_rate')::numeric) AS rr,
    AVG((d->>'vital_confidence')::numeric) AS vital_confidence
  FROM iot_timeseries it,
       jsonb_array_elements(it.data_value) AS d
  WHERE it.card_id = $1::text
    AND it.device_type = 'Sleepad'
    AND it.category    = 'track'
    AND it.timestamp BETWEEN $night_start_ms AND $night_end_ms
    AND (d ? 'heart_rate' OR d ? 'respiratory_rate')
  GROUP BY bucket_ms
),
radar_2s AS (
  SELECT
    (timestamp / 2000) * 2000           AS bucket_ms,
    AVG((d->>'heart_rate')::numeric)    AS hr,
    AVG((d->>'respiratory_rate')::numeric) AS rr,
    AVG((d->>'vital_confidence')::numeric) AS vital_confidence
  FROM iot_timeseries it,
       jsonb_array_elements(it.data_value) AS d
  WHERE it.card_id = $1::text
    AND it.device_type = 'Radar'
    AND it.category    = 'track'
    AND it.timestamp BETWEEN $night_start_ms AND $night_end_ms
    AND (d ? 'heart_rate' OR d ? 'respiratory_rate')
  GROUP BY bucket_ms
),
merged AS (
  SELECT
    COALESCE(s.bucket_ms, r.bucket_ms) AS bucket_ms,
    COALESCE(s.hr, r.hr)               AS hr,
    COALESCE(s.rr, r.rr)               AS rr,
    COALESCE(s.vital_confidence, r.vital_confidence) AS vital_confidence,
    CASE WHEN s.hr IS NOT NULL THEN 'Sleepad'
         WHEN r.hr IS NOT NULL THEN 'Radar' END AS source
  FROM sleepad_2s s
  FULL OUTER JOIN radar_2s r USING (bucket_ms)
)
```

**字段**：
```
hr_night_p50_bpm  = MEDIAN(merged.hr WHERE vital_confidence >= 36)
rr_night_p50_rpm  = MEDIAN(merged.rr WHERE vital_confidence >= 36)
hr_primary_minutes_night    = COUNT(source = 'Sleepad') × 2 / 60
hr_fallback_minutes_night   = COUNT(source = 'Radar')   × 2 / 60
hr_total_coverage_pct       = (primary + fallback) / total_night_minutes
```

### 3.4 Cardiovascular Vital — 波动 / 节律

| 指标 | 公式 |
|---|---|
| `hr_cv` | `stddev(hr) / mean(hr)`（夜窗内 merged 数据） |
| `hr_rmssd_proxy_bpm` | `sqrt(mean((hr[i+1]-hr[i])^2))`，标 `source='computed_proxy'`（厂家不给真 HRV） |
| `hr_spike_count` | `COUNT(\|hr - baseline.hr_night_med\| > 2 × baseline.hr_night_iqr)` |
| `sleep_period_hr_amp_bpm` | `hr_pre_sleep_30min_p50 - hr_deep_sleep_p50`（替代原 circadian_amp） |

### 3.5 HR-RR 耦合

```
hr_rr_corr_night = Pearson(hr_series_aligned, rr_series_aligned)
```
健康老人正相关；`|corr|` 显著下降（月 < 0.3 且 baseline > 0.6）→ 自主神经失衡线索。

### 3.6 Recovery Kinetics — 仅夜间 bathroom（Tier S #1）

**事件链时刻定义**：

| 时刻 | 信号 | 来源 |
|---|---|---|
| t0 baseline 窗 | 离床前 5 min | sleep_state 状态 3/4 |
| t1 离床 | 状态 6→1 转换 OR alarm_events `SleepPad_LeftBed` | sleepace |
| t2 peak | 离床后 60s 内 HR 最高 | iot_timeseries merged |
| t3 回床 | 状态 1→2/3/4 OR alarm_events `OnBed` | sleepace |
| t4 恢复 | 回床后 HR 首次落入 baseline ± IQR/2 | iot_timeseries merged |
| **recovery_sec** | t4 − t3 | 计算 |

**月度衍生**：
```
bathroom_recovery_hr_p50_month_sec = MEDIAN(hr_recovery_sec WHERE event IN month)
recovery_drift_pct                 = (current_month - baseline_180d) / baseline_180d
```
`recovery_drift_pct > +20%` 多周持续 → 心血管储备下降信号。

### 3.7 Frailty Velocity（Tier S #2）

每月对 `gait_morning_p50_month` 做过去 6 个月线性回归：
```
slope_per_month             = linear_regression(month_index, gait_morning_p50_month).slope
gait_decline_annualized_pct = (slope_per_month × 12) / current_month_p50
gait_decline_r2             = R²（< 0.3 标 unreliable）
```
**临床阈值**：年化下降 > 5% 衰弱进展期；> 10% 快速进展。LTC carrier / state Medicaid 升级评估的核心字段。

### 3.8 Stability Signature（Tier S #3）

每个核心指标月度 out-of-band 比例：
```
stability_<metric>_oob_pct = COUNT(daily_value 在 [baseline.med - baseline.iqr, baseline.med + baseline.iqr] 之外) / days_with_data
```
合成稳定性：
```
stability_composite_score = 1 - mean(stability_*_oob_pct of {hr_night, rr_night, gait_p50, sleep_eff})
                            ∈ [0, 1]
```

### 3.9 Chronic Multi-signal Resonance（Tier S #4）

```
mobility_z_month = (gait_morning_p50_month - baseline.med) / sigma_eq
sleep_z_month    = (sleep_efficiency_month_p50 - baseline.med) / sigma_eq
vital_z_month    = combined z of {hr_night, hr_rmssd_proxy, hr_rr_corr}

resonance_count_3mo     = COUNT(month in last 3 months WHERE all three z < -1)
chronic_resonance_index = sigmoid(resonance_count_3mo / 3 × 4 - 2)
```

### 3.10 Respiratory & SDB（核心 #3）

**完全停呼吸（Apnea）— MVP**：
```
apnea_complete_count_night = COUNT(alarm_events.event_type = 'ApneaHypopnea' IN night_window)
apnea_per_hour_night       = apnea_complete_count_night / sleep_duration_hours
```
> ⚠️ CheckAH 算法依赖 sleepace ≤ 2s 采样；adaptive sampling 工单上线前结果偏低。

**Cheyne-Stokes 周期性呼吸 — Phase 2**：
```python
# 输入：夜间 RR 时序（双源 merged，2s 采样）
# 算法：
#   1) 滑动窗 60s，FFT 找主导周期
#   2) 若 30s ≤ period ≤ 120s 且振幅 > baseline_RR × 0.4 → 标 CS-suspect
#   3) 月度累加：cs_suspect_minutes_month
```
临床意义：心衰 / 中风后金信号。

### 3.11 独居生活力（Tier S #5，瘦身）

#### Liveness（活着 + 在屋）
- `last_seen_active_at` = 当日最后一次 motion / track / vital 信号
- `inactive_streak_hours_max` = 当日最长连续无任何信号时段

**校正**：设备离线时段不计入；已知睡眠时段不计入；入住前 30 天为校准期不出 alert。

**实时告警阈值**（独居场景）：
| 严重度 | 触发 | 路由 |
|---|---|---|
| info | inactive > 6h（非夜间） | 集团 dashboard 提醒 |
| warning | inactive > 12h（非夜间） | case manager webhook |
| critical | inactive > 18h or 跨整夜无任何信号 | case manager + APS hotline 集成 |

#### Daily Pattern Drift
30 天后建立个人节律向量 `pattern_baseline` JSONB：
```
[wake_time_p50, sleep_time_p50, kitchen_visit_count_p50, bathroom_visit_count_p50, ...]
```
当日向量与基线计算归一化欧式距离 → `daily_pattern_drift_score ∈ [0, 1]`；> 0.4 累计天数 → `pattern_drift_days`。

> Phase 2+ 增强：餐次规律 / wandering / 重复路径（依赖 cell tag RPC + ≥3 月数据）。

### 3.12 Restlessness + 体动（洞察 A + B）

**主信号 — 直接读 sleepace 已检测的 alarm 事件**（不再单字段自算，避免误判）：

sleepace consumer 已经把这两类事件落入 alarm_events（[mqtt_consumer.go:739-741](../../owlBack/wisefido-sleepace/internal/consumer/mqtt_consumer.go#L739-L741) → [alarm.go:74-75](../../owl-common/alarm/alarm.go#L74-L75)）：

| event_type | 含义 | sleepace 触发逻辑 | 临床用途 |
|---|---|---|---|
| `NoBodyMove`            | 睡眠时期身体无微动持续 N 分钟 | 厂家算法（adaptive） | 极静止状态信号（呼吸 / 脉动微动也低 → 数据可靠度参考） |
| `NoTurnOver`            | 久卧未翻身 ≥ 2H | 厂家算法 | **直接对应"褥疮高风险护士提醒"** — 临床唯一可操作信号 |
| `AbnormalBodyMovement`  | 异常体动量 | 厂家算法 | 不安腿 / 抽搐 / 谵妄候选信号（慎用） |

**daily 字段产出**（直接 COUNT alarm_events 行）：
```sql
no_body_move_alarm_count_night = COUNT(*) FILTER (WHERE event_type='NoBodyMove'           AND triggered_at IN night_window)
no_turn_over_alarm_count_night = COUNT(*) FILTER (WHERE event_type='NoTurnOver'           AND triggered_at IN night_window)
abnormal_body_move_count_night = COUNT(*) FILTER (WHERE event_type='AbnormalBodyMovement' AND triggered_at IN night_window)
```

**辅助 — sleepace iot_timeseries 字段**（参考统计，不进风险评分）：
```
turn_over_count_night       = SUM(turn_over field from data_value in night_window)
body_move_total_count_night = SUM(body_move ...)
```
作用：sanity check（若 alarm_events 长期 0 但 turn_over_count 也低 → 可能 sleepace 离线 / 老人不在床）；不再用 `< 5` 这类绝对阈值（已被 #10 修正撤销）。

> ⚠️ **不要从 turn_over 字段自己重算 bedsore 风险**——sleepace 厂家算法已包含个体校准，自算会引入误判。直接信任 NoTurnOver alarm 即可。

**Restlessness 指标**：
```
pose_transitions_per_hour_night = COUNT(pose state changes in night window) / night_hours
room_transitions_per_day        = COUNT(cell tag changes through day)
pre_sleep_restlessness_idx      = pose_transitions in [sleep_onset - 1h, sleep_onset]
```
**临床信号**：
- 谵妄（delirium）最早期信号——医院 ICU 用，家居场景独家
- room_transitions 太低（< 5/天）= 静止；太高（> 50）= 焦虑 / wandering

### 3.13 数据缺失模式（洞察 C）

**信号化的"缺失"**：
```
daytime_activity_data_pct  = (有 activity 数据的小时数 / 16h, 06:00-22:00)
nighttime_vital_data_pct   = (有 vital 数据的小时数 / sleepace 报告时长)
```
- 健康老人：daytime_activity > 0.6 + nighttime_vital > 0.85
- 长期卧床：daytime_activity < 0.3
- 故意避开监测：device_uptime_pct 高（设备没坏）但数据塌陷 → `device_intentional_avoidance = true`

### 3.14 Chronotropic Competence（洞察 D — 心血管储备金标准）

**关键约束**：HR 主要在床上时段，walking 主要在白天非床时段——两者**时间重叠少**。可行的近似：

```
# 取床上但清醒态（state==2 之前 or sit_up=6 之后）的 HR 作为"近 walking 时段 HR"
# 取深睡态 HR 作为 hr_resting

hr_active_p50_window = MEDIAN(hr WHERE bed_status=0 AND walking_duration_in_5min > 60)
                        # 雷达 vital 仅床上，所以这个数据稀疏；Phase 2 后用 sleepace adaptive 后扩
                        # MVP 用：hr_pre_sleep（清醒态在床）作为 active 代理
hr_resting_p50       = hr_night_p50_bpm  -- 已有
chronotropic_response_p50_month = (hr_active_p50_window - hr_resting_p50) / hr_resting_p50
```
**临床判读**：
- 健康：> 0.10
- 衰弱：0.05–0.10
- 异常（chronotropic incompetence）：< 0.05 或反向

> Phase 1–2 实现，需要 sleepace adaptive sampling 上线后床上 HR 数据更丰富。

### 3.15 HR Drop Event — Pre-syncope / Vasovagal（洞察 E）

```
hr_drop_event = HR 在 30s 内下降 > 15 bpm 且 baseline > 60 bpm
hr_drop_event_count = COUNT per day
```

**与 fall 事件配对**（独立分析，不入主评分）：
```
IF fall_event AND EXISTS hr_drop_event WITHIN [-60s, 0]:
    fall.metadata.root_cause_hint = 'syncope_likely'
```
**临床/商业价值**：
- 老人晕厥前兆独家信号
- fall 责任溯源（vasovagal vs 真摔倒）→ 家属解释 / 保险定责

### 3.16 Postprandial Nap HR（洞察 F）

依赖 sleepace adaptive sampling 工单的"nap_window"规则：12:15–13:30 切到 2s 采样。

```
nap_detected_today = (12:15–13:30 期间 bed_status=0 持续 ≥ 15 min)
nap_started_at     = 上床时刻
nap_hr_p50_bpm     = MEDIAN(hr in [上床后 5min, 上床后 30min])
postprandial_hr_lift_bpm = nap_hr_p50_bpm - hr_night_p50_bpm
```

**临床判读**（postprandial autonomic carryover）：
- 健康：lift > 8 bpm
- 钝化（糖尿病/衰弱/心衰）：lift < 3 bpm 或反向

**月度衍生**：
```
postprandial_hr_lift_month_p50    = MEDIAN(daily lift WHERE nap_detected)
postprandial_response_decline_pct = (current_month - baseline_90d) / baseline_90d
```

**Confidence 处理**：月内 nap captured 天数 < 8 → 不输出，confidence < 0.5。

### 3.17 Visitor Episode + Intimacy（核心 #6 — INTERNAL_ONLY）

**Episode 边界**（cardagg 现有逻辑）：multi_person_duration ≥ 30s 启动；多人消失结束。

**亲密度算法**（在 wisefido-ai/roomengine 新模块）：
```python
# 每秒（雷达 1Hz）算两 track 距离
distance_t = sqrt((x_resident - x_visitor)^2 + (y_resident - y_visitor)^2)

# 距离桶
if   distance < 50:   bucket = 'intimate'    # 拉手 / 喂食
elif distance < 150:  bucket = 'close'       # 床边 / 协助
elif distance < 300:  bucket = 'present'     # 同房间但有距离
else:                 bucket = 'distant'     # 走过场

# 加权
intimacy_index = (
    1.0 × intimate_seconds +
    0.6 × close_seconds +
    0.3 × present_seconds +
    0.0 × distant_seconds
) / total_episode_seconds
```

**判读**（INTERNAL_ONLY，仅集团内 HR / care manager 看）：
| intimacy_index | 含义 |
|---|---|
| > 0.6 | 高质量陪伴（家属真心 / 关怀型护工） |
| 0.3–0.6 | 任务型护理 |
| < 0.3 | 走过场 |

**对外可出**（衍生指标）：
- `visitor_days_count_month` / `no_visitor_streak_max_days` / `visitor_total_minutes_day` — social isolation proxy
- 不出：`intimacy_*` / `intimate_minutes_*` / `care_quality_score` / `family_support_score`

### 3.18 缺失/离群通用规则

| 情况 | 处理 |
|---|---|
| 单日采样 < 最小样本数 | 当日该指标 NULL，`data_quality_score` 减分 |
| 极端值（HR > 180 / RR > 40 / speed > 200 cm/s） | 入流前过滤，标 outlier |
| 设备离线全天 | 当日整行 device_uptime_pct = 0；solo-living 字段全 NULL |
| 月内有效天 < 18 | 月度 confidence < 0.6，不出黄/红色 |
| 多设备同 card 部分离线 | 按权重计算 device_uptime_pct |
| `vital_confidence < 36` | 该样本不入 baseline / 静态指标，但仍可入 acute spike 检测 |

---

## 4. 个人基线建模

### 4.1 滚动窗口

- **主基线 90 天**：月度对比
- **稳态基线 180 天**：recovery / circadian / 节律向量等慢变指标
- **独居 pattern 基线 30 天**：节律学习冷启动期
- 每月 1 号刷新

### 4.2 Robust 统计

```
median  = percentile_cont(0.5)
iqr     = percentile_cont(0.75) - percentile_cont(0.25)
sigma_eq ≈ iqr / 1.349
```

### 4.3 冷启动

| 累计有效天数 | 策略 |
|---|---|
| < 14 | 不出基线，月度 confidence=0 |
| 14–29 | 14 天 short baseline，`is_cold_start=true`，confidence ≤ 0.4 |
| 30–89 | 实际窗口，confidence 按 `actual_days / 90` 缩放 |
| ≥ 90 | 标准 90 日基线 |
| 独居 pattern | < 30 天不出 pattern_drift |

### 4.4 数据质量与基线漂移

`data_quality_score`（日级）：
```
quality = 0.25 × (samples_hr / 720)
        + 0.25 × (samples_track / 1000)
        + 0.15 × (1 if has_sleep_report else 0)
        + 0.20 × confidence_avg_of_vitals
        + 0.15 × device_uptime_pct
```

**基线漂移**：每月新基线 vs 上月，若 `|new_med - prev_med| > 1.5 × prev_iqr` → 写 `metadata.drift=true`。

---

## 5. 风险评分

### 5.1 三大医学风险轨迹（0–1 子分）

#### Cardio Stress
| Contributor | z（坏方向） |
|---|---|
| `hr_night_drift` | `(month - baseline_90d) / sigma_eq`，正向坏 |
| `hrv_drop` (rmssd_proxy) | `-(month - baseline) / sigma`，下降坏 |
| `recovery_lengthening` | `(bathroom_recovery_p50 - baseline_180d) / sigma`，变长坏 |
| `rr_night_volatility_up` | `(month_rr_cv - baseline) / sigma` |
| `sleep_period_amp_decline` | `-(month - baseline) / sigma`，下降坏 |
| `hr_rr_decoupling` | `-(month_corr - baseline) / sigma` |
| `chronotropic_response_decline` | `-(month - baseline) / sigma`，下降坏（洞察 D） |
| `hr_drop_event_density_up` | `(month - baseline) / sigma`（洞察 E） |
| `postprandial_response_decline` | `-(month - baseline) / sigma`（洞察 F） |

```
cardio_raw   = sigmoid(Σ wᵢ · clip(zᵢ, -3, 3))
cardio_stress_score = cardio_raw × confidence_score
```
权重默认等权（向量长度归一），Phase 4 后用 calibration loop 调（§G）。

#### Frailty / Cerebrovascular
| Contributor | z |
|---|---|
| `gait_decline_annualized` | 直接用 §3.7 输出，归一化到 z |
| `gait_morning_decline_30d` | `-(month - baseline) / sigma` |
| `sit_to_stand_lengthening` | `(month - baseline) / sigma` |
| `walking_distance_decline` | `-(month - baseline) / sigma` |
| `long_sit_burden_up` (排除 sofa) | `(month - baseline) / sigma` |
| `situp_latency_up` | `(month - baseline) / sigma` |
| `sleep_fragmentation_up` | `(month - baseline) / sigma` |
| `fall_burden` | `min(fall_count_month / 3, 1) × 3` |
| `no_turn_over_alarm_burden` | `no_turn_over_alarm_count_month / month_days`（直接读 sleepace alarm，§3.12） |
| `restlessness_up` | `(pose_transitions - baseline) / sigma`（洞察 B 部分进 frailty） |

#### Acute Watch（72h 滚动） — **趋势提示，不是实时告警**

> **§0.3 Care-not-Treatment 原则**：此分仅在 monthly_health_trend 报表中呈现，给 PACE / SNF
> case manager 看"近 72h 是否进入需要更多关注的状态"。**绝不进 realtime_liveness_alert 通道**，
> 也不触发 RN 24h 响应 webhook。任何医学决策（用药调整 / ER 转运 / 急救）需医生独立判断 —
> AI_health 不替代临床流程。

用 baseline 14d。
| Contributor | 触发 |
|---|---|
| `rr_volatility_jump_72h` | 近 72h rr_cv > baseline_14d + 2σ |
| `hr_spikes_low_activity` | HR spike 升 + walking_duration 显著降 |
| `sleep_eff_drop_72h` | 近 3 晚 sleep_eff < baseline_14d - 1.5σ |
| `cs_suspect_spike` | CS minutes 突增（Phase 2 后） |
| `multi_signal_resonance` | 上述 ≥ 2 项同时触发，加 0.2 权重 |

### 5.2 Solo Living Risk Score（独居专属）

仅对 `cohort_type IN ('pace', 'state_hcbs')` 计算：
```
solo_living_risk_score (0–1) =
    0.30 × liveness_concern         # inactive_streak_p95_month / 24h
  + 0.25 × frailty_velocity_norm    # |gait_decline_annualized_pct| / 10%
  + 0.20 × acute_72h
  + 0.15 × pattern_drift_norm
  + 0.10 × bathroom_anomaly         # nocturia / 频次偏离
```

### 5.3 Vitality Preservation Score 0–100

> **合规定位**（用户 2026-05-02 定）：此分数语义是"**个人 baseline 偏移度 / 稳定保持程度**"，
> **不是**"绝对健康水平 / 既存疾病评级 / 核保依据"。
>
> - 数值高 = 该 card 当月相对自己 90 日 baseline 偏移小，活力指标稳定
> - 数值低 = 该 card 偏移大，提示 care manager 加强观察
>
> 故意避开 NAIC / HIPAA Title I 合规雷区（FICO 风格 health_score 会被认定为风险定价输入 →
> pre-existing condition discrimination → 多州 algorithmic underwriting 法规约束）。
> 以"trend / preservation"立意属于 wellness 范畴，不属核保。

```
vitality_preservation_score = round(100 × (
    0.25 × (1 - cardio_stress_score)         # cardio 漂移程度 → 反向（漂移越小 = 保持越好）
  + 0.25 × (1 - frailty_score)               # frailty velocity 漂移
  + 0.10 × (1 - acute_watch_score)           # 急性偏离
  + 0.20 × stability_composite_score         # 稳定性正向（baseline 内时间占比）
  + 0.10 × (1 - chronic_resonance_index)     # 慢性多信号共振 → 反向
  + 0.10 × recovery_capacity_score           # = 1 - sigmoid(recovery_drift_pct)
))
```

约束：`confidence_score >= 0.5` 时输出，否则 NULL。

**子分 sub_scores**（各分量都是"个人 baseline 偏移度"，不是绝对水平）：
```json
{
  "stability": 82,
  "recovery_capacity": 75,
  "frailty_velocity": -3.2,
  "acute_offset_72h": 15,
  "solo_living_offset": 38
}
```

**对外字段命名约束**：
- 内部表 / API 响应 JSON 字段名：`vitality_preservation_score`（不要 `health_score`）
- attestation bundle / cohort PDF 显示名：`Vitality Preservation` / `活力保持度`
- `institution_contract.scope_metric_keys` 白名单严禁出现 `health_score` 字符串

### 5.4 Confidence

```
confidence_score =
    0.4 × min(days_quality_ok / 21, 1)
  + 0.3 × min(baseline.sample_days / 90, 1)
  + 0.2 × (1 - cold_start_factor)
  + 0.1 × signal_coverage_ratio
```

### 5.5 风险颜色（持续性 gating）

```
color_now = bucket(vitality_preservation_score):  green ≥ 70  /  yellow 40–69  /  red < 40
color_emit:
  yellow 仅当连续 ≥ 2 周保持
  red    仅当连续 ≥ 1 周保持 且 confidence ≥ 0.6
  否则降级到上一稳定颜色
```

### 5.6 Calibration Loop（洞察 G — 架构性护城河）

每个发出的 risk_color → 6 个月后回看实际 outcome：
```
monthly_health_trend.outcome_event_type IN ('hospitalization', 'er_visit', 'fall_with_injury', 'death', NULL)
monthly_health_trend.outcome_event_at
monthly_health_trend.outcome_severity
monthly_health_trend.outcome_lookback_days
```
- Phase 0 已埋字段（见 §2.5），等待真实事件回填
- Phase 5 试点跑 6 月后开始有数据
- Phase 6+ 用加权回归（不必上 ML）调权重
- 12 月后：估计 AUC > 0.75（同行 ~0.6）
- 24 月后：RWE 论文 + 卖给再保 = Phase 7 现金流

**这是把 RWE 沉淀变现金的设计闭环**——也是融资估值锚点。

---

## 6. 可解释性层

### 6.1 Top Contributors

每个轨迹独立选 top-N（默认 N=3）：
1. 按 `|wᵢ · zᵢ|` 排序
2. 同向（向坏）的优先
3. 输出结构：
```json
{
  "key": "hr_night_p50",
  "direction": "up",
  "delta_value": 8.2,
  "delta_unit": "bpm",
  "z": 1.8,
  "baseline_window_days": 90,
  "contribution": 0.42,
  "narrative_en": "Resting nighttime heart rate is up ~8 bpm vs the 90-day personal baseline."
}
```

### 6.2 Narrative 模板（英文为主，USA 市场）

```
hr_night_p50 ↑:           "Resting nighttime heart rate up {delta} bpm vs {window}-day baseline"
hrv_drop:                 "Heart rate variability declined ~{pct}% this month, suggestive of reduced autonomic regulation"
gait_morning_p50 ↓:       "Morning gait speed median {value} cm/s, down {pct}% vs baseline"
sit_to_stand ↑:           "Sit-to-stand time extended from {prev}s to {now}s on average"
recovery ↑:               "Post-arousal heart rate recovery time extended from {prev}s to {now}s"
chronotropic_decline:     "Reduced heart rate response to activity, suggestive of cardiovascular deconditioning"
postprandial_decline:     "Reduced postprandial heart rate response, possible autonomic involvement"
multi_signal:             "Sleep, nighttime HR, and gait declined concurrently this month — recommend closer observation"
solo_inactive:            "Maximum daily inactive streak reached {hours} hours on {date}"
pattern_drift:            "Daily routine deviated from personal baseline on {N} days this month"
turn_over_low:            "Reduced overnight turnover frequency — pressure injury risk monitoring suggested"
```

### 6.3 防诊断 Guardrails

| 禁用词（拦截/替换） | 替换为 |
|---|---|
| hypertension / 高血压 | cardiovascular load indicator |
| heart failure / 心衰 | reduced cardiovascular reserve indicator |
| AFib / 房颤 | rhythm instability signal |
| Parkinson / 帕金森 | gait pattern change |
| dementia / 痴呆 | cognitive-behavioral pattern change |
| delirium / 谵妄 | acute restlessness pattern |
| pressure ulcer / 褥疮 | pressure injury risk indicator |

---

## 7. REST API 契约（集团 dashboard 内部消费）

> Pre-VC 阶段 API key + IP allowlist；OAuth/databus post-VC。

### 7.1 Cohort metrics

```
GET /api/v1/health/cohorts/{cohort_id}/monthly?from=2026-02&to=2026-04
```
返回：`cohort_health_metrics` 行 + 同窗口下属 high-risk member 列表（不含 INTERNAL_ONLY 字段）。

### 7.2 Member drilldown

```
GET /api/v1/health/cohorts/{cohort_id}/members/{card_id}/monthly?from=2026-02&to=2026-04
```
返回 `monthly_health_trend` + 子分 + top_contributors + short_clinical_hint。**不含 INTERNAL_ONLY 字段**。每次写 `data_access_audit`。

### 7.3 单指标时序

```
GET /api/v1/health/cohorts/{cohort_id}/members/{card_id}/metric/{metric_key}
    ?granularity=day&from=2026-02-01&to=2026-04-30
```
`metric_key` 仅允许白名单字段（非 INTERNAL_ONLY）。

### 7.4 事件恢复曲线

```
GET /api/v1/health/cohorts/{cohort_id}/members/{card_id}/recovery
    ?from=2026-04-01&to=2026-04-30
```

### 7.5 Liveness alert webhook

```
POST {realtime_webhook_url}
X-Wisefido-Signature: ed25519:...
```
- 失败按指数退避重试 5 次

### 7.6 报告 PDF

```
GET /api/v1/health/cohorts/{cohort_id}/reports/{month_key}.pdf
GET /api/v1/health/cohorts/{cohort_id}/members/{card_id}/reports/{month_key}.pdf
```
后端渲染 + ed25519 签名 + 字段白名单（不含 INTERNAL_ONLY）。

### 7.7 错误/空数据

| 状态 | HTTP | body |
|---|---|---|
| cohort 无数据 | 200 | `{"months":[], "status":"no_data"}` |
| 冷启动 | 200 | `vitality_preservation_score=null, status='building_baseline'` |
| cohort 不在合同 | 403 | `{"error":"out_of_contract_scope"}` |
| 合同到期 | 410 | `{"error":"contract_expired"}` |
| 合同暂停 | 423 | `{"error":"contract_suspended"}` |

---

## 8. 服务架构

### 8.1 组件划分

| 组件 | 职责 | 选型 |
|---|---|---|
| **新服务 `wisefido-ai-health`** | ETL 调度 + 风险评分 + REST API + signed report 渲染 | Go，独立进程 |
| **新服务 `wisefido-liveness-streamer`** | 实时活动信号监听 + alert 推送 | Go，订阅 Redis stream |
| **新模块 `wisefido-data` / sleepace_sampling_scheduler**（独立工单） | sleepace adaptive sampling（夜睡 + 午睡两窗口） | Go，per-card scheduler |
| **新模块 `wisefido-ai/roomengine` / intimacy_calculator**（独立工单） | 多 track 距离 → 亲密度 → visitor_episode | Go，订阅多 track 状态 |
| **新模块 `wisefido-cardagg` / visitor_episode_persistence**（独立工单） | TargetState episode 落 PostgreSQL | Go，扩展现有 state_service |
| 复用 `wisefido-data` | API gateway 反代 | 仅路由 |
| 复用 `wisefido-cardagg` | 只读 alarm_events | 不改主逻辑 |
| 复用 `wisefido-ai/roomengine` | cell tag 查询接口（路径 B） | RPC |

### 8.2 ETL 调度

| 任务 | 频率 | 输入 | 输出 |
|---|---|---|---|
| `daily_etl` | T+1 02:00（院区时区） | iot_timeseries / alarm_events / sleepace_report / visitor_episode | daily_health_metrics |
| `daily_event_recovery` | T+1 02:30 | iot_timeseries（事件前后窗口）+ alarm_events | health_event_recovery |
| `monthly_etl` | 每月 1 号 03:00 + on-demand | daily_health_metrics（近 90/180d） | monthly_health_metrics, health_baseline |
| `monthly_scoring` | 每月 1 号 04:00 + on-demand | 上面 | monthly_health_trend |
| `cohort_aggregation` | 每月 1 号 05:00 | monthly_health_trend | cohort_health_metrics |
| `acute_watch_refresh` | 每日 06:00 | 近 72h | 仅刷新当月 acute_watch_score |
| `liveness_streamer` | 实时（Redis stream） | iot:* signals + 设备心跳 | realtime_liveness_alert + webhook |
| `outcome_backfill` | 每月 | 外部 outcome 上报（机构) | 回填 monthly_health_trend.outcome_* |
| `audit_purge` | 每日 03:00 | data_access_audit | 删除 retention_until < today（HIPAA） |

**幂等**：所有 ETL `INSERT ... ON CONFLICT DO UPDATE`。
**增量**：`daily_etl` 用 `triggered_at >= last_run_watermark`；watermark 存 `health_etl_state`。
**Provenance**：`row_hash = sha256(card_id || date_key || canonical_json(metrics) || prev_row_hash)`。

### 8.3 物化视图 vs 实表

- 核心表：实表
- 可选 MV：`v_card_latest_trend` / `v_cohort_high_risk_today`

### 8.4 性能预算

- 1000 cards × 90 天 × 24h × 60 样本/h ≈ 1.3 亿行 vital
- TimescaleDB hypertable + GIN(data_value)，单 card+day 查询 < 50ms
- `daily_etl` ~1s/card/天；并行 8 路 1000 cards ≤ 3min
- 评分秒级
- `liveness_streamer` 检测延迟 < 30s

### 8.5 错误处理

- 单 card 失败不阻塞批次：写 `health_etl_errors` + 跳过
- `pg_advisory_lock` 防 ETL 冲突
- 监控：cards_processed_per_run / error_rate / latest_complete_date_per_card

---

## 9. 实施路径

### Phase 0 — 地基（1–2 周）
- [ ] **校准 Phase 0 前置 TODOs**（见 §10.2 清单）
- [ ] owlRD/db 新增 `30_health_*.sql`（11 张表）
- [ ] `wisefido-ai-health` 服务骨架
- [ ] 1 个数据最完整的 card 端到端跑通 30 天 daily_etl，肉眼校验
- [ ] 落 outcome 字段（calibration loop 准备）

### Phase 1 — Mobility 全量 + Sleep + 体动（1.5 周）
- [ ] gait morning + sit_to_stand + bathroom 日夜分桶 + walking distance
- [ ] Sleep Architecture 全字段（onset latency / WASO / 占比 / 节律 IQR）
- [ ] 体动 / 翻身字段（褥疮预警）
- [ ] Restlessness 日聚合
- [ ] frailty_velocity 公式上线
- [ ] cell-tagged long-sit 路径 A/B 决策（依赖 area_type 校准结果）

### Phase 1.5 — sleepace adaptive sampling 工单（独立）
- [ ] 新模块 `wisefido-data/sleepace_sampling_scheduler/`
- [ ] 个人入睡时间预测（过去 14d sleepace.start_time P50）
- [ ] 双时段规则（夜睡 reset_time-5min / 午睡 12:15-13:30）
- [ ] sleepace API 调用 + 重试 + 状态记录

### Phase 2 — Vital + Recovery + Respiratory（1.5 周）
- [ ] 双源 fallback HR/RR
- [ ] hr_rmssd_proxy / sleep_period_hr_amp / hr_rr_corr
- [ ] 夜间 bathroom recovery ETL（5 时刻链路）
- [ ] Apnea 事件计数
- [ ] HR drop event（pre-syncope）
- [ ] Postprandial Nap（依赖 1.5 完成）
- [ ] Chronotropic competence 实现

### Phase 2.5 — Visitor / Intimacy 工单（独立，与 Phase 2 并行）
- [ ] cardagg `visitor_episode` 落库
- [ ] wisefido-ai 亲密度算法模块
- [ ] daily / monthly visitor 字段聚合
- [ ] INTERNAL_ONLY 字段隔离 + 注释

### Phase 3 — 基线 + 评分（1 周）
- [ ] health_baseline 月刷新（含全部新字段）
- [ ] cardio / frailty / acute / chronic_resonance / stability 全部公式
- [ ] Vitality Preservation Score 0–100 + 子分
- [ ] **Cheyne-Stokes FFT 检测**（核心 #3 旗舰）
- [ ] 等权先跑（calibration loop 等数据）

### Phase 3.5 — 独居生活力 + 实时告警（1 周）
- [ ] solo_living_block 字段
- [ ] `wisefido-liveness-streamer` 服务
- [ ] solo_living_risk_score
- [ ] webhook 推送 + retry

### Phase 4 — 集团 dashboard + 签名报告（1.5 周）
- [ ] cohort_health_metrics 月聚合
- [ ] dashboard MVP（INTERNAL_ONLY 字段权限隔离）
- [ ] PDF 渲染 + ed25519 签名（字段白名单）
- [ ] institution_contract + BAA 后台
- [ ] data_access_audit 全链路埋点
- [ ] 解释层 narrative + guardrails

### Phase 5 — Denver 三角试点（**Pre-VC milestone**，3–6 月）

| 试点对象 | 切入策略 | 关键指标 |
|---|---|---|
| **InnovAge**（Denver 总部 PACE 巨头） | PACE capitation ROI：减 ER / 住院 = 直接省钱 | ER ↓%, hospitalization ↓%, capitation savings $ |
| **Colorado HCPF / HCBS Waiver** | 独居老人 solo_living_risk + APS 集成 | inactive alert 响应率 / HCBS 评估准确度 |
| **UC Anschutz Geriatrics** | 学术合作 + IRB + RWE 论文 | 估值加成 / 学术背书 |

**目标产出（融资材料）**：
- 3 池子 ≥ 200 老人 × ≥ 6 个月数据
- 至少 1 个量化 outcome
- 1 篇 working paper / pre-print
- 完整产品演示
- 7 年 audit log 一致性证明

> **Series A 融资发生在此 milestone 后**

### Phase 6（**Post Series A**）— OAuth/databus + 多州扩展（6–12 月）
- OAuth 2.0 数据中介（Plaid 风格）
- 公开 verify endpoint + SDK
- 第二批州（CA / NY / TX / FL）
- MA plans / ACO / 大型养老连锁
- **季节去趋势 baseline** 上线（12 月数据后）
- **Calibration loop 第一轮调权**（6 个月 outcome 数据）

### Phase 7（**Post Series A**）— 再保 + RWE 数据池
- 去标识化 RWE 数据池
- 再保（Munich Re / Swiss Re / RGA）合作
- LTC carrier 精算输入
- "Quiet Quitting" Detection 上线（mobility ↓ + visitor ↓ + sleep ↑ + intimacy ↓ → 临终前 6-12 月放弃信号）

---

## 10. 开放问题与风险

### 10.1 开放问题

| 问题 | 风险 | 当前处理 |
|---|---|---|
| **无 IBI/RR-interval** | HRV 用 RMSSD-proxy | 标 `source='computed_proxy'`；Phase 6+ 向硬件厂家询问 IBI 输出 |
| **iot_timeseries.card_id 是 VARCHAR** | ETL JOIN 慢 | ETL 内部 cast；远期推动 schema 改 UUID |
| **alarm_events 无 card_id** | 4 表 JOIN | 评估增冗余列 |
| **沙发误报 long-sit** | 老人沙发 2h 不应病理化 | §3.1 cell tagging 拆分 |
| **多设备/多床覆盖同老人** | 数据归属复杂 | card_id 聚合；前端按 resident 合并 |
| **冬季 HR 整体上行（生理性）** | 季节性误判 | Phase 6 引入"季节去趋势" |
| **PHI 加密** | resident 信息 | 复用 owlBack PHI（resident_phi AES-256-GCM） |
| **roomengine 依赖** | cell tag 不稳定时 mobility 退化 | DB 直读 grid_snapshot；缺 snapshot / 低 confidence cell 兜底 'other' |
| **InnovAge 接触不足** | first pilot 无对象 | 同步推 UC Anschutz |
| **HIPAA SOC 2 合规初期成本** | Phase 4 之前必须达标 | Phase 0 引入合规咨询 |
| **设备 attestation 缺失** | 反欺诈最弱环 | 先做 source-row hash chain 兜底 |
| **sleepace adaptive sampling 工单依赖** | nap / 夜窗细粒度数据延迟 | 独立工单 Phase 1.5；本设计公式中标 confidence 降级 |
| **CheckAH 算法依赖采样频率** | adaptive sampling 上线前 apnea 检测偏低 | 接受 MVP 偏低；adaptive 上线后自动校准 |
| **RR 数据来源依赖 sleepace** | sleepace 离线时 RR 全断 | radar 床上 2s 兜底；双源 fallback |
| **outcome 回填依赖机构上报** | calibration loop 数据稀疏 | Phase 5 试点机构合同明确 outcome 上报义务 |

### 10.2 Phase 0 校准结果（2026-05-02 完成）

| # | 校准项 | 实际结果 | 设计落点 |
|---|---|---|---|
| 1 | POSE 编码 | walking=1 / sit=3 / stand=4 / lie=6（[cell.go:46-49](../../owlBack/wisefido-ai/internal/roomengine/cell.go#L46-L49)） | 直接落 §3.1 / §3.6 |
| 2 | facility_id 列 | **不存在** — cards 用 `branch_id`（cards.branch_id REFERENCES branches） | 全设计 facility_id → branch_id |
| 3 | cards→devices JOIN | `cards.devices JSONB` 已预计算包含 `[{device_id, ...}]` 列表（ETL 直读） | 不必 JOIN devices |
| 4 | alarm_events.event_type | 实际 13 种：`Fall/BedSitUp/LeftBed/HeartRateAlert.High/NoBodyMove/ApneaHypopnea/NightAbsence/BedNightAbsence/Stay/SensorDetached/AbnormalBodyMovement/HeartRateAlert.Low/RespRateAlert.High`。**无 LongSit / ProlongedStay / Radar_Fall** | §3.1 long-sit 改自算（D1.c）；Fall 公式只查 `'Fall'` |
| 5 | `area_type` 入 trigger_data | `area_type` 是雷达**区域枚举 0-6**（none/custom/bed/interfer/door/monitor_bed/sensing），**不带 cell tag (sofa/kitchen) 语义** | 路径 A 不可行；走 D2 — DB 直读 [`roomengine_grid_snapshot`](../../owlRD/db/36_roomengine_grid_snapshot.sql)，不要 RPC |
| 6 | iot_timeseries.category 实际值 | 30+ 种：监控数据用 `track / heart / activity / sleep-stage / number_people`；事件用 `EnterRoom/ExitRoom/InBed/LeftBed/Walking/Fall/Initialization/...`；设备健康用 `Offline/OfflineRecover/SignalPoor/AngleException/...` | 双源融合 SQL 改 `category='track'` |
| 7 | sleepace report.summary 解析 | [report_service.go:50-62](../../owlBack/wisefido-sleepace/internal/service/report_service.go#L50-L62) 当前 parser 只解析 5 字段（recordCount/startTime/stopMode/timeStep/timezone），**未读 seIndex** | D4.b — ETL 直接 jsonb 解（不动业务服务） |
| 8 | body_move/turn_over 入库 | ✅ [mqtt_consumer.go:366-367](../../owlBack/wisefido-sleepace/internal/consumer/mqtt_consumer.go#L366-L367) 写 FieldBodyMove/FieldTurnOver 在 `category=track` 行 | §3.12 字段直接读 |
| 9 | 雷达 vital category | **写在 `category='track'`**（同行 data_value 包含 heart_rate/respiratory_rate/vital_confidence），不是单独 category | 双源融合 SQL 改 `device_type='Radar' AND category='track'` |
| 10 | device 来源区分 | iot_timeseries 直接有 `device_type` 列（`'Sleepad'` / `'Radar'`），免 JOIN | 双源融合不需 JOIN device_store |

### 10.3 校准过程的额外发现

| 发现 | 说明 | 设计落点 |
|---|---|---|
| **alarm_events.trigger_data 无 `duration_sec`** | 实际结构：`{event_id, track_id, event_since, event_end, event_status, event_payload}`（毫秒时间戳） | duration 用 `(event_end - event_since)/1000`（§3.1 已改） |
| **alarm_events 有 `iot_timeseries_id` 列** | 直接关联触发该 alarm 的 timeseries 行，不必按 timestamp 找 | recovery / cell tag 关联可走此 ID |
| **alarm_events 无 card_id 列**（设计已知）| 实际通过 `device_id → cards.devices @> [{device_id}]` 反查；性能可接受 | §10.1 已记 |
| **alarm_events.operation 字段值** | `verified` / `false_alarm` / `test` / `''` | "verified fall" = `event_type='Fall' AND operation='verified'`（§3.1 排除规则） |
| **iot_timeseries.timestamp 是 bigint (Unix ms)** | 不是 timestamp tz；JOIN alarm_events.triggered_at 时需转换 | ETL 显式 cast |
| **iot_timeseries 是 PG 分区表（42 child）** | 按时间分区 | TimescaleDB 候选；查询性能由分区裁剪保证 |
| **iot_timeseries 已冗余 `branch_name/unit_name/room_name/bed_name`**（VARCHAR 字符串非 UUID）| 不能直接用作 JOIN，仅供 ETL 错位时 sanity check | 主路径走 cards → branches |
| **roomengine_grid_snapshot.payload schema** | `{grid:{w,h,ox,oy}, cells:[{i, b:[[type,conf,source]×3], c:{at:[Move,Stand,Sit,Lie],...}}]}`；AreaType 枚举 0-7（[cell.go:17-27](../../owlBack/wisefido-ai/internal/roomengine/cell.go#L17-L27)） | §3.1 cell_tag_at 算法已落 |

---

## 11. Wisefido B2B Disclosure Layer

> **搁置加注**：本设计 MVP（§1–§10）不实现 §11；Phase 5+ post-VC 才做。当前章节仅作架构占位，避免后期返工。

### 11.1 设计哲学（保留）
- B2B-only / B2G-only
- 集团内可见、不可导
- 机构作为数据控制方，Wisefido 作为 processor
- USA-only
- 审计友好

### 11.2 institution_contract + BAA（Pre-VC 轻量）

数据库见 §2.8。合同签订流程人工，无自动化。

### 11.3 输出形态

#### Cohort + Member 签名 PDF
- 后端渲染（**字段白名单**：仅非 INTERNAL_ONLY 字段）
- 末页 QR + ed25519 签名 + bundle_id + valid_until + issuer=Wisefido
- 集团端只能下载，不能"导出 raw"

#### Attestation Bundle（API JSON）
```json
{
  "bundle_id": "uuid",
  "issuer": "wisefido",
  "issued_at": "2026-05-01T...",
  "valid_until": "2026-06-01T...",
  "subject": { "type": "member|cohort", "id": "uuid" },
  "contract_id": "uuid",
  "vitality_preservation_score": 78,
  "sub_scores": { "stability": 82, "recovery": 75, "frailty_velocity": -3.2, "acute_risk_72h": 15 },
  "coverage": { "monitored_days": 187, "coverage_pct": 0.94 },
  "evidence_hash": "sha256:...",
  "signature": "ed25519:..."
}
```

### 11.4 Provenance + ed25519 签名

链式 hash：
```
row_hash      = sha256(canonical_json(content) || prev_row_hash)
bundle_hash   = sha256(canonical_json(bundle_payload))
signature     = ed25519_sign(wisefido_private_key, bundle_hash)
```
- 公钥：`https://wisefido.com/.well-known/keys.json`（post-VC 公开）
- 私钥：GCP/AWS KMS 中签名，不下放应用进程
- 公钥按年轮换（旧公钥保留 7 年）

### 11.5 Audit Log

见 §2.10 + §8.2 audit_purge。7 年保留。集团客户可通过 `GET /api/v1/health/contracts/{contract_id}/audit-log` 查自家访问历史。

### 11.6 Anti-fraud 最小集

| 控制点 | 实现 |
|---|---|
| **Source hash chain** | row_hash + prev_row_hash |
| **设备 fingerprint** | source_device_uids 写入 daily 行 |
| **数据连续度门槛** | 月有效天 < 18 → 不出 vitality_preservation_score |
| **Coverage 检查** | bundle 必带 monitored_days / coverage_pct |
| **设备替换检测** | 月内设备指纹突变 → metadata.device_swap=true |
| **数据"完美无瑕"检测** | 长期所有指标在 baseline ±0.5σ → 标可疑 |

### 11.7 Post-Series-A 路线（占位）

| 模块 | 方向 |
|---|---|
| OAuth 2.0 数据中介层（Plaid 风格） | 集团 dashboard → Wisefido OAuth → 直接给集团 |
| 公开 verify endpoint | `GET /verify/{bundle_id}` |
| Databus / SDK / embed | 嵌入集团客户系统 |
| 行为级 anti-fraud（ML） | 跨机构数据关联 |
| 自动化合同 + 计费系统 | self-serve |
| 多州 / 多产品扩展 | CA / NY / TX / FL |

---

## 附录 A — 与现有系统的交互边界

> **核心原则（§0.3）**：AI_health 仅从 PG 表读，零修改实时通路。

| 触点 | 关系 |
|---|---|
| `cardagg` | **零修改**。事后批处理从 alarm_events 表读 |
| `wisefido-ai/roomengine` | **零修改**。DB 直读 `roomengine_grid_snapshot.payload` 解析 cell tag（见 §3.1 cell_tag_at 算法），不要 RPC |
| `wisefido-sleepace` | **零修改**。从 sleepace_report 读；seIndex 等未解析字段 ETL 自取 jsonb（D4.b） |
| `wisefido-data` | 路由 reverse proxy；**独立工单**：sleepace_sampling_scheduler（adaptive sampling，影响 nap / vital 细粒度） |
| visitor / intimacy | **独立工单**：cardagg `visitor_episode` 持久化 + roomengine `intimacy_calculator`（§2.5 / 附录 D #2-3） |
| 前端 | 新内部 React app，session auth，无 to-C UI |
| PHI 加密 | 复用 resident_phi AES-256-GCM；聚合表只存 card_id/resident_id |

---

## 附录 B — 示例 SQL（Phase 0 验证，已校准）

**Fall 部分**（仅靠 alarm_events，event_type='Fall' 单值，duration 自算）：

```sql
INSERT INTO daily_health_metrics (tenant_id, card_id, branch_id, date_key,
                                  fall_count, fall_total_duration_sec)
SELECT
    c.tenant_id,
    c.card_id::uuid,
    c.branch_id,
    DATE(ae.triggered_at AT TIME ZONE u.timezone) AS date_key,
    COUNT(*) FILTER (WHERE ae.event_type = 'Fall')                          AS fall_count,
    COALESCE(SUM(
        GREATEST(0,
          ((ae.trigger_data->>'event_end')::bigint
           - (ae.trigger_data->>'event_since')::bigint) / 1000
        )
      ) FILTER (WHERE ae.event_type = 'Fall'
                  AND ae.trigger_data ? 'event_end'),
      0)                                                                     AS fall_total_duration_sec
FROM cards c
JOIN units  u ON c.unit_id = u.unit_id
JOIN alarm_events ae
       ON ae.tenant_id = c.tenant_id
      AND c.devices @> jsonb_build_array(jsonb_build_object('device_id', ae.device_id::text))
      AND ae.triggered_at >= now() - interval '30 days'
WHERE c.card_id = '<TEST_CARD_UUID>'
GROUP BY c.tenant_id, c.card_id, c.branch_id, date_key
ON CONFLICT (card_id, date_key) DO UPDATE SET
    fall_count               = EXCLUDED.fall_count,
    fall_total_duration_sec  = EXCLUDED.fall_total_duration_sec,
    updated_at = now();
```

**Long-sit 部分**：episode 检测 + cell tag 在 ETL Go 层算（参见 §3.1 算法），SQL 仅写入。完整 SQL 在 Phase 0 实现时补充到 `wisefido-ai-health/sql/`。

---

## 附录 C — 标准对接占位（Phase 5 时细化）

### C.1 InterRAI HC（家居评估国际标准）字段映射

| InterRAI 项 | Wisefido 字段映射 |
|---|---|
| Section G — ADL Self-Performance（移动） | `gait_morning_p50_month`, `walking_distance_p50_month_m`, `gait_decline_annualized_pct`, `sit_to_stand_sec_p50_month` |
| Section H — Continence | `bathroom_visit_count_day`, `bathroom_visit_count_night`, `bathroom_visit_duration_p50_*` |
| Section J — Health Conditions | `hr_night_p50_month`, `acute_watch_score`, `chronotropic_response_p50_month` |
| Section K — Disease Diagnosis | （不出诊断；contributor + clinical hint） |
| Section P — Sleep | `sleep_efficiency_month_p50`, `waso_count_p50_month`, `sleep_fragmentation_month_p50` |
| Section Q — Falls | `fall_count_month`, `fall_total_duration_sec_month` |
| Section R — Cognition | `daily_pattern_drift_score`, `pattern_drift_days`, `sleep_onset_time_iqr_days_30` |
| Section T — Social Functioning | `visitor_days_count_month`, `no_visitor_streak_max_days` (不含 intimacy) |

### C.2 OASIS-E 字段映射（CMS Home Health）— Phase 5 定终版

### C.3 CPT 98975–98981（RTM 报销字段）— Phase 5 定终版

---

## 附录 D — 独立工单清单

按依赖顺序：

| # | 工单 | 位置 | 阻塞 | 优先级 |
|---|---|---|---|---|
| 1 | **Sleepace Adaptive Sampling Scheduler** | wisefido-data 新模块 | Phase 2（Postprandial / 夜间细粒度 vital） | 高 |
| 2 | **cardagg visitor_episode 持久化** | wisefido-cardagg 扩展 state_service | Phase 2.5 | 中 |
| 3 | **roomengine intimacy_calculator** | wisefido-ai 新模块 | Phase 2.5（依赖 #2） | 中 |
| 4 | ~~roomengine cell_tag_at RPC~~ | ~~wisefido-ai 新接口~~ | **已撤销** — DB 直读 grid_snapshot.payload（§3.1 / §0.3 架构原则） | — |
| 5 | **outcome_backfill 接口** | wisefido-ai-health 新 endpoint | Phase 5 末（calibration loop 第一轮） | 中 |

---

**下一步起步入口**（Phase 0 校准已完成 2026-05-02）：
1. ✅ 校准 §10.2 Phase 0 前置 TODOs（结果在 §10.2 / §10.3）
2. 建表（11 张表，对应 §2 → `owlRD/db/38_health_*.sql`）
3. 启动 sleepace adaptive sampling scheduler 工单（独立工单 #1，并行 Phase 0）
4. 按 §9 Phase 0 → 7 顺序推进

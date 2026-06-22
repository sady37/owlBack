# owlBack 项目规则

本文件是 owlBack 后端工作的硬约束。每条规则后跟一行 **Why** 注脚。违反必须先改规则再写代码。

新规则在末尾追加，编号只增不复用；规则废弃用删除线并保留编号占位。

---

## 规则 #1 — 编码风格

### 1.1 命名

- **Go identifier**：PascalCase（导出）/ camelCase（包内）。包名单字小写无下划线（`alarm`, `observation`, `card`）。
- **文件名**：snake_case (`alarm_service.go`)。
- **JSON / DB 字段**：snake_case (`heart_rate`, `device_addr`, `event_kind`)。Go struct tag 显式指定，禁靠默认。
- **Redis key**：冒号分隔 `domain:purpose:scope` (`iot:event:stream`, `card:status:<addr>`)。新 key 写下时检查存量同 domain 的命名一致性。
- **事件 / 类别字面量**：唯一来源 = `owl-common/observation` 包常量（`observation.EnterRoom`、`observation.InBed` 等）。**禁止**字面量重复（`"EnterRoom"` 写进业务代码 = 违规）、**禁止** alias re-export、**禁止** 在 alarm 包或其他包复定义同名常量。新增事件必须先在 observation 落常量再用。
- **新事件命名风格**：dot-namespaced（`fall.suspect` / `night.absence_onset`），向 CloudEvents 收敛。旧的 PascalCase（`EnterRoom`/`Fall`）保留不改名。

**Why**：drift 是最贵的 bug — 改一处忘改另一处的字面量复写就会 silent miss。常量化让编译器替你 grep。

### 1.2 不向后兼容

- 删除即删除。**不留** stub / no-op shim / NULL 占位 / `if v2 else v1` 双路 / deprecated alias / re-export 兼容层。
- **不写** `TODO: 迁移` / `TODO(战役 X): audit` 之类签字注释。要做现在做，做不了就别动 schema。
- schema 改完直接让旧调用点编译报错，由编译器驱动 fix。
- migration PR 合并前必须 grep `已退役|已废弃|deprecated|TODO.*迁移|no-op stub|短路|兜底` 自查，命中 = 没扫干净。

**Why**：每一个"先留着别人改"的位置都会变成下一个会话的清理负担。屎山是这么累积的。

### 1.3 单一入口 / 单源真相

- 多源风险字段（layout / device / spatial_addr / event_kind / alarm_config）必须有 **唯一** unifiedSave 入口；其他路径不能绕过去写。
- 常量在 canonical package 定义，引用必须 `import`；禁止字面量复写。
- 派生数据有唯一权威源（zone state 由 sensor zoneengine 写 / card status 由 cardagg 投影 / device meta 由 owl-common ipam）；下游只读。

**Why**：分歧不在数据本身，分歧在多个写入口对"现在是什么"的不同理解。砍掉次要写入口 = 砍掉一类 bug。

### 1.4 错误处理只在边界

- 系统边界（用户输入 / 外部 API / firmware 上报 / 文件 IO）：必须 validate + 显式错误响应。
- 内部调用（同进程同 package / 跨自家服务）：trust caller / trust framework guarantee；不写 nil check / 类型断言兜底 / 空值默认。
- **禁止** `if v == nil { return defaultValue }` 这类"防御性"代码 — 它把 bug 变成静默错误。
- panic 用在"不可能到达"的分支（编译器还推不出来的不变量）；不要用 panic 替代正常错误传播。

**Why**：兜底是把 bug 推到下游的隐藏成本。让 nil pointer panic 暴露 wiring 错，比让它静默走默认值好得多。

### 1.5 注释纪律

- **默认不写**。命名要好到让 reader 不需要注释。
- 只在 WHY 不显然时写一行（隐藏约束 / 微妙不变量 / 特定 bug workaround / 反直觉决策）。
- **禁止** WHAT 注释（`// increment counter`）。
- **禁止** 引用当前 task / fix / 调用方（`// used by X` / `// for the Y flow` / `// fixes issue #123`）— 那是 commit message 的事，写进代码会随重构腐烂。
- **禁止** 多段 docstring；package doc 一行说清职责即可。

**Why**：注释是债务。错的注释比没有注释更危险——它会误导未来读者。

### 1.6 提交前自检

每次写完代码合并前：
1. `grep -rE '已退役|已废弃|deprecated|TODO.*迁移|no-op stub|短路|兜底' <changed_paths>` — 命中 = 没扫干净
2. `grep -rE '"(EnterRoom|InBed|LeftBed|Fall|...)"' <changed_paths>` — 事件名字面量 = 没用常量
3. `go vet ./... && go build ./...` 全绿才提交

**Why**：每条规则都有人破。grep 是机械化的诚实。

---

## 规则 #2 — 流分配纪律：持久化事件 vs 瞬态状态

### 2.1 两类东西，两种流

| 类别 | 流 | 持久化 |
|---|---|---|
| **离散持久化事件**（业务/审计/HIPAA 需要）<br>`EnterRoom` / `fall.suspect` / `night.absence_onset` / `resident.detected` | `iot:event:stream`（广播总线）| iot → `event_log` 90d |
| **离散持久化报警**（actionable，需要事后追责）<br>confirmed `Fall` / `LeftBed` / `Offline` | `iot:alarm:stream`（广播总线）| iot → `alarm_events` 7y HIPAA |
| **高频 raw 监控**<br>1Hz track XY / HR / RR | `iot:monitor:stream`（广播总线）| iot → `monitor_stream` 90d |
| **瞬态状态翻转**（投影 / 中间产物，无审计价值）<br>`zone.bed.occupied/leaving/vacant` / `zone.room.people=N` | 专用流 (e.g., `sensor:zone:state:stream`)| **不入库** |
| **AI 派生判定 + 决策审计**<br>track ghost verdict / lost-fall 进-取消-触发-抑制 | 专用流 `ai:track:verdict:stream`（realtime → cardagg override 内存 cache）| **旁路入库** `sensor_decision_log`（66_，30d，iot 第二消费者写）<br>2026-06-03 反转原"不入库"：DBN 复盘需"决策+特征"层 trail（负决策"为什么没报"原仅在 journal 会轮转）。仍是旁路审计，不在热路径 |
| **维护型实时流**（producer 持续 maintain）<br>card:realtime 1s snapshot | 专用流 `card:realtime:stream` | 不入库 |

### 2.2 分配决策树

- 这条信号要不要入业务库 / 审计库？
  - **要** → 走广播总线 `iot:event/alarm/monitor:stream`，让 iot 模块负责持久化；多 consumer 各取所需
  - **不要** → 开专用流，零持久化开销，consumer 自己订阅
- 这条信号是不是有"持续维护态"（TTL / 分组 / 累积）？
  - **是** → producer 是 maintainer，consumer 只读不重做维护工作（rule 2.4）

### 2.3 不开专用流的反例

不要为了"性能"或"隔离"擅自开专用流：
- 同 category filter 即可（`if msg.Category == "fall.suspect"`）成本 = 几 ns 每条
- 多开一条流 = consumer group 配置 + 监控 + 重启 wiring + 文档同步 — 维护成本远大于过滤

开专用流的合法理由只有 2 个：
1. **不入库**（避免污染 event_log / alarm_events）
2. **producer 是 maintainer + 已 grouped product**（如 card:realtime:stream，cardagg 已按 cardID 分组并 1s snapshot）

### 2.4 Producer/Maintainer vs Consumer 分工

某些流是 producer 持续 maintain 的产品（不只是 transport）：
- producer 做 TTL / dedup / 分组 / snapshot
- consumer 只读，**不重做维护工作**

让 consumer 重做 = 维护工作双写 = drift 风险（如 cardagg 维护 12s TTL，data 也 cache 12s，两边 TTL 不一致就 silent bug）。

**判定**：流名是 `<domain>:realtime:*` / `<domain>:projection:*` / `<domain>:state:*` 之类暗示"已加工"的，按 maintainer 模式；流名是 `iot:*:stream` 之类原始总线的，按 raw transport 模式。

**Why**：分错流 = 要么 event_log 爆掉瞬态状态（90d retention 浪费）、要么真业务事件丢失审计、要么 consumer 重做维护工作 drift。流分配是架构层第一刀，砍错很难回头。

---

## 规则 #3 — replay/真 case 验证的重点 = 运行机制，**与 fire 无关**

跑 replay / 真 case 排查时，**验证对象是"机制有没有按设计执行"**，不是"这一帧报没报警（fire）"。

- **看机制**：lost / exit / hand-off / blind-faller hold / drop / 占用聚合 / 跨雷达对账 / SLeft 注入 / lostAt 登记 / rhoFor 解析 …… 这些状态转移与判据**有没有正确触发、按正确顺序走、得出正确中间结论**。
- **不看 fire**：某条轨 belief 到没到 0.85、floor StillSec 够不够阈、这次报没报——**是结果不是机制**。fire 阈值/曲线全靠人为测试数据（铁律 [[fall_data_is_artificial_test]]，禁标定），所以 fire 与否**证明不了对错**；机制是否正确才是 replay 能验、也该验的。
- **判据**：发现"0.23 < 0.85 所以不报"这类时，停下——那是 fire 层，问的应是"机制把它 hold 成 blind-faller 了吗 / lost 登记了吗 / hand-off 查了吗 / 跨雷达对账了吗"。机制对、fire 不对 = 留 oracle 调阈；机制错 = 真 bug 要修。

**Why**：人为测试数据下 fire 阈无法标定（规则隐含铁律），盯 fire 会把"阈值没调"误当 bug、把"机制断裂"漏过。机制是结构正确性（可证伪），fire 是标定产物（数据不足证伪）——replay 的价值在前者。


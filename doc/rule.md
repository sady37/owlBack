# owlBack 统一规则书 — Single Source of Truth

**定位**：项目最终唯一权威。code 注释 / 设计 doc / memory 与本文件冲突时**本文件胜**。
**编辑**：每条 = 编号 + 一句陈述 + 一行 `Why:` 或 `Case:`；废止用 `~~删除线~~` 保留编号占位；编号只增不复用；新规追加不重排。

`revision: 1` ・ `last-audit: 2026-05-21`

---

## §0 — Top-7 任何动作前先脑过

> 这 7 条破坏力最大 / 出现最频繁。新会话开工、新 PR 提交、新设计落地前先脑过一遍。

- **0-1**　[S1] schema 改前先改 `owlRD/dbv2/<NN>_<table>.sql` CREATE 给用户审。`Why:` 不审 = 后续 rollback 难，drift 早期不暴露。
- **0-2**　[W1] 多步 plan 主线 / 支线物理隔离，走支线不动主线条目。`Case:` 历史上从支线返回后主线全丢。
- **0-3**　[F5] 跨服务 PR producer 先彻底完工，downstream / FE 再做。`Why:` 顺序错则 downstream 对死字段空跑。
- **0-4**　[B3] 禁防御性 `if v == nil { return default }`。`Why:` 让 nil panic 暴露 wiring 错，比静默走默认值好。
- **0-5**　[S2] 删除即删除：不留 stub / shim / 双路 fallback / deprecated alias / re-export 兼容层。`Why:` 屎山从"先留着"开始累积。
- **0-6**　[I2] API ID 全 IPv6 canonical text；UUID 仅对外集成（sleepace SDK 等）。`Case:` 半迁移产生 UUID↔IPv6 mismatch。
- **0-7**　[N4] 事件 / 类别字面量唯一来源 = `owl-common/observation` 常量。`Case:` 字面量复写一处忘改 = silent miss。

---

## §C — Card

- **C1**　`card_id ≡ spatial_prefix`（INET），mask 即卡型：**/80 unit·public + /88 room + /96 bed_card 三种实卡**（/48 tenant、/56 branch、/64 site 只用作权限 scope，不出 card 行；/128 device 永不出 card）。`Why:` 2026-05-23 重订 — /88 room 卡为保留 same-room 拓扑关联（OOB / room radar 与床位 device 同卡显示）必不可少；split 规则按 C9 决定具体出哪几层。
- **C2**　card 寿命跟 space 寿命；与 device、resident 双解耦。具体哪些 mask 层级出独立卡由 C9 split rule 决定。**/80 unitCard 仅在 unit 有 room 时出**（merge 模式必出，split 模式按 M 与 lazy 规则）；/88 roomCard 在 split 模式且 Room.N≤1 时与对应 room space 同生死；/96 bedCard 仅在某 room 内 Room.N>1 时与对应 bed space 同生死。`Why:` device/resident 任一缺位都不该让卡消失；"card 下无空间则删"（unit 无 room 不出 /80）。
- **C3**　`has_bed` = ActiveBed 语义：scope 下任一 bed 上**已绑 device** 才 TRUE；空 bed_card has_bed=FALSE。`Why:` has_bed 是给 sensor 消费的运行时 flag（"scope 内是否有可用 bed 设备"），不是空间结构标记；structural"是否有 bed 空间"由 split 规则的 Bed count 隐含表达。
- **C4**　card 唯一写入口在 `wisefido-data`；其他模块只读。`Why:` 多写源 → 状态发散；见 §X。
- **C5**　alarm_events / event_log **不存** card_id；归属反查走 `device_addr + ts ∈ episode`。`Why:` card 重组不污染历史审计；episode 时段反查即得。
- **C6**　device-card 1:1 绑定；公共区域用 public 卡（不属任何住户）；不做 fan-out；多卡命中 = 数据异常 warn 兜底。`Why:` fan-out 复杂度不值；public 卡承"无主"语义。
- **C7**　未绑卡 device 用 device_id 兜底（`owl-common.StreamHeadCardID`）；0 卡分支不再 drop。`Why:` 卡缺失时不能 drop 实时流。
- ~~**C8**　`has_bed` = structural：scope 下有 bed 结构即 TRUE；与 device 绑定状态无关。~~（2026-05-21 当日撤销 — has_bed 是 sensor 消费的运行时 flag，应为 ActiveBed 语义，方向定反；恢复 C3）
- **C9**　card 是纯空间集合，shape 由 unit 内**结构性 bed 总数**一次定型（不是 ActiveBed runtime）。规则（2026-05-23 拍板）：

  ```
  N = bedCount_in_unit          (整 unit 床数 /96 候选)
  M = noBedRoomCount_in_unit    (不含 bed 的 room 数 /88 候选，含 0-device 空 room)

  Step 0: unit 无 room → 无卡

  Step 1: Unit 层
    N ≤ 1  → /80 only (merge，吸所有 device + alarm)
    N > 1  → split：
              若 M > 0 → 预创建 /80 (装 noBed room 的 device)
              若 M = 0 → /80 不预创建（可能 Step 2 lazy create）

  Step 2: 每个含 bed 的 room (split 模式下)
    Room.N ≤ 1 → /88 卡 (吸 bed + room device)
    Room.N > 1 → N 张 /96 + room device → /80
                 若 M = 0 → lazy 创建 /80
  ```

  device / alarm 路由 = PG GiST `<<=` LPM 自动（device <<= 现存 /96 落 /96；<<= 现存 /88 落 /88；否则 /80）。无 /96 时 alarm 自然归更高层卡，`alarm.producer` 保留发起设备身份。`Why:` 卡型 = spatial mask（C1），结构由 space 决定（C2），不由设备绑定状态翻转；/88 保留以维系 same-room 拓扑（Bedroom 内 radar 与 BedA device 同卡可见，sensor OOB 即便不依赖 cards 也对 FE 显示有用）；旧 `alreadySplit` latching / activeBed 触发的动态 split 模型全部砍掉。
- **C10**　Card 创建/删除触发 = space DDL（按 C9 重算 unit scope expected set）：`unit INSERT` 首 room 加入时出 /80；`room INSERT/DELETE` 影响 M 和 split 路径，可能新增/删 /88 或 /80（含 lazy create）；`bed INSERT/DELETE` 跨 unit 阈值 N=1↔N=2 触发 /80→/96 切换，跨 room 阈值 Room.N=1↔N=2 触发 /88→/96 切换；`unit DELETE` 级联删本 unit 全部卡。device 绑定/解绑、resident_unit 变更**只更新 snapshot 字段**（card_name/resident_id/has_*），不创建不删除卡。Card delete 时同 card_id 上非终态 alarm 自动 → `expired`（见 [[alarm_anchored_to_card]] 规则）。`Why:` 落实 C2+C9，card 寿命跟 space 不跟 device/resident。
- **C11**　Public 卡 = `units.unit_type=3` 的 /80 卡：`card_name='public'`、`resident_id=NULL`；公共区域 device 落 public 卡（C6）。普通 unit /80 卡 `card_name`=resident nickname。`Why:` public 是 unit 属性不是独立卡型，简化模型。
- **C12**　Card 字段填充：INSERT 时 `card_id`=spatial_prefix（C1）/ `card_name`=resident nickname or `'public'` / `card_dns`=`card.ShortCodeOf(card_id)` / `resident_id`=LPM(`resident_unit.spatial_prefix`) or NULL。`has_bed` (C3 ActiveBed 语义) / `has_bathroom` / `has_kitchen` 在 **device 绑定/解绑** 事件上更新（room/bed 内有无绑 device），不在 INSERT 时一次性 EXISTS。`Why:` 字段单源；snapshot 读快免 JOIN；has_* 是运行时 flag 跟绑定状态走（C3），不是结构标记；LPM 由 IPv6 prefix 直接得出（I1）。

---

## §D — Device

- **D1**　device 全栈两个命名收敛：`device_uid`（UUID，硬件 identity / vendor 烧录终身不变） + `device_addr`（INET IPv6，业务侧寻址 / spatial 可重分配）。`Why:` identity vs address 是网络协议自带分层；slot 0/0xFF 重分配（IPAM）证明 IPv6 是 addr 不是 id，命名跟上语义。
- **D2**　使用边界：**静态 / 与业务无关**（OTA 升级 / auth / 出入库 inventory）一律 `device_uid`；**业务内**所有模块（wisefido-data / card / cardagg / alarm / sensor / iot 等）一律 `device_addr`。`Why:` 静态侧不需要层级关系（I1 prefix-match 不适用），UID 不变量最简；业务侧需要 INET <<= 寻址必须 IPv6。
- **D3**　过渡期遗留字段 `device_id`（业务侧现存 UUID）/ `device_ipv6`（业务侧现存 INET）保留不动，直到 wisefido-data + card + device 重写窗口一次性收口到 `device_addr`，遵守 single-trip 纪律（不双发 / 不 fallback / 不观察期）+ 同 PR 同部署窗（参 0-6 / I2 / Z-001）。`Why:` 同名异型 (UUID→INET) 是 Go silent drift 风险；只在重写窗内 cutover 才安全。
- **D4**　Phase 1（现行）：OTA / auth 等静态业务模块所有现用 UID 的 callsite 显式改名 `device_uid`，与业务侧字段物理隔离；做完后 grep 全仓 `device_id` 应 100% 落在"即将重写"的业务模块内（漂出的归 device_uid 或纳入 Phase 2 范围）。`Why:` Phase 1 idempotent 可独立部署零业务风险，并为 Phase 2 提供 audit 入口。
- ~~**D5**　Phase 2 决策门闩：wisefido-data + card + device 重写**动工前**必须先拍 Phase 2（业务侧 device_id / device_ipv6 → device_addr 收口范围 + 时间窗）；错过此窗口禁开工。~~（2026-05-21 Phase 2 完成；保留编号占位作历史）
- **D6**　Phase 2 收口落地（2026-05-21 cutover）：dbv2 schema + PG live data + Go 全栈代码同 PR 同部署窗完成；终态 = `device_uid VARCHAR(50)` (dfm PK / logMAC) + `device_addr INET /128` (devices PK / 业务寻址)；`device_id` UUID 列 + `device_ipv6` 列名全栈退役；migration 入 `owlRD/dbv2/migrations/2026-05-21_phase2_device_uid_collapse.sql`，含 19+224 行数据 backfill + DROP COLUMN device_id (alarm_events) + RENAME device_ipv6→device_addr。`Why:` 一刀切 + 编译器驱动；§S5 v1 service 含 3+ schema 引用立即整段重写已落地；R-003 retention 退役。
- **D7**　Sleepace SDK userId 现 = `device_uid` logMAC string（如 "BM87224700864"），不再是合成 UUID。`Case:` 2026-05-22 实测 sleepace 厂家 API 对 UUID 和 logMAC 两种 userId 格式返回相同数据（{status:0, reportUploadTime:12}），证明 vendor 实际按 deviceId (platform code) 键，userId 只是 opaque tag，任何 string 都接受；既存 8 设备无需 rebind，直接切换零冲突。
- **D8**　设备 identity FK 链：FK→`dfm(device_uid)` 是唯一标识引用路径；`devices.device_uid` UNIQUE 也接受为 FK target；任何业务表需要"哪个物理设备"用 device_uid，需要"设备地址"用 device_addr，**禁混用** identity 和 address 字段。`Why:` D6 一刀切后两条命名平行，混用 = 重新引入语义漂移。

---

## §E — SSE

- **E1**　SSE 是 transport，不做业务过滤；订阅集合 = 用户权限内全部 card（含 0 设备卡）。`Why:` 保持 SSE 单一职责。
- **E2**　业务过滤（隐藏 monitor=off 卡等）放 FE 视图层。`Why:` 后端不用知道 FE 视图态变化即时同步。
- **E3**　`card_change` 事件 → FE re-POST `/stream/watch` 重订阅。`Why:` FE 自动维持 watch 集与权限/状态同步。

---

## §F — 流分配（events vs state）

- **F1**　持久化事件（业务/审计/HIPAA）走广播总线 `iot:event:stream` / `iot:alarm:stream` / `iot:monitor:stream`，iot 模块负责入库。`Why:` 单点入库，多 consumer 各取所需。
- **F2**　瞬态状态翻转走专用流，**不入库**。`Why:` event_log 90d 不该被状态翻转污染。
- **F3**　Maintainer 模式：流名带 `realtime` / `projection` / `state` 的，producer 负责 TTL / dedup / 分组 / snapshot；consumer 只读不重做维护。`Why:` 双写 TTL/dedup 必 drift。
- **F4**　开专用流的合法理由只有 2 个：① 不入库；② producer 是 maintainer + 已 grouped product。`Why:` 防"图清净"随手乱开流。
- **F5**　Producer-first：跨服务多 PR producer 先彻底完工。`Why:` 见 0-3。

---

## §A — Alarm

- **A1**　派生信号（WeakBio / 趋势 / 聚合等）**禁进** alarm 决策路径（verdict / severity / routing / dedup / timing 任一）；fall_verify 是已知唯一例外不再扩。`Case:` 4 次历史违规变种。
- **A2**　设备类报警（Offline / SignalPoor / AngleException / SensorDetached）一律走 alarm 流 + auto-recover；不走 event 流。`Why:` 与事件流分工清晰。
- **A3**　`alarm_status` 5 值，无 `acked_auto_resolved`：用 `operation='auto_resolved' AND handler IS NOT NULL` 复合表达等价语义；Critical auto_resolved → 直接 resolved。`Case:` 2026-05-21 砍掉第 6 值决定。
- **A4**　Radar HR 不能 CRITICAL：FE select 收窄 + load 时 CRITICAL→WARNING；Sleepad 接触式不受此限。`Why:` mmWave HR ±5-10 BPM 不达临床精度。

---

## §I — IPv6 / 身份

- **I1**　判 Unit / Room / Bed / Device 任意层包含关系，首选 `netip.Contains` / SQL `INET <<=`；禁用结构 JOIN（bed.room_id 等）。`Why:` 自描述路径，零结构假设。
- **I2**　API ID 全 IPv6 canonical text。`Case:` 见 0-6。
- **I3**　短码 alias（`dns_short_name` 等）只在 perm check resolve；handler 用 `resolvedID` 不用 raw param。`Case:` resolve 不到位 = silent cache miss。
- **I4**　Server 内部全 UTC；TZ 转换仅在 API 边界。`Why:` cron / 比较 / 算术只用 UTC，少一类时区 bug。

---

## §S — Schema

- **S1**　schema 改先 dbv2 CREATE 提审。`Why:` 见 0-1。
- **S2**　删除即删除。`Why:` 见 0-5。
- **S3**　v2 cutover R-001：不动 v1 业务逻辑 / 功能，只动管道层。`Why:` cutover 限定爆破半径。
- **S4**　v2 cutover R-002：禁硬删（HIPAA + 业务恢复），全用 `status='deleted'` 软删。`Why:` 法规 + 可恢复。
- **S5**　v1 service 含 3+ 处已删/改 schema 引用时立即整段 v2 重写，禁逐函数 short-circuit。`Case:` 短路变种 3+ 后不可维护。

---

## §N — 命名

- **N1**　Go identifier：PascalCase 导出 / camelCase 包内；包名单字小写无下划线。`Why:` 标准 Go 风格。
- **N2**　文件名 snake_case；JSON / DB 字段 snake_case；Go struct tag 显式指定。`Why:` 一致性。
- **N3**　Redis key 冒号分隔 `domain:purpose:scope`。`Why:` 同 domain 命名一致便 grep。
- **N4**　事件字面量唯一来源 = `owl-common/observation` 常量。`Case:` 见 0-7。
- **N5**　新事件 dot-namespaced（`fall.suspect`）向 CloudEvents 收敛；旧 PascalCase（`EnterRoom` / `Fall`）保留不改名。`Why:` 不破坏既有，但新增收敛。

---

## §X — 单一入口 / 单源真相

- **X1**　多源风险字段（layout / device / spatial_addr / event_kind / alarm_config）必须有**唯一** unifiedSave 入口。`Why:` 分歧不在数据，在多写入口对"现在是什么"理解不同。
- **X2**　常量在 canonical package 定义，引用必须 import；禁字面量复写。`Why:` 同 N4。
- **X3**　派生数据唯一权威源（zone state ← sensor zoneengine / card status ← cardagg / device meta ← owl-common ipam）；下游只读。`Why:` 下游重做维护 = 双写 drift。
- **X4**　Radar Layout / Device 三项（InstallMod / Height / Boundary）必须 == firmware；不允许漂移；唯一入口 `unifiedSave`。`Case:` 漂移 = sensor 算出错位。

---

## §B — Boundary 错误处理

- **B1**　系统边界（用户输入 / 外部 API / firmware 上报 / 文件 IO）：必须 validate + 显式错误响应。`Why:` 不可信源必查。
- **B2**　内部调用：trust caller / trust framework；不写 nil check / 类型断言兜底 / 空值默认。`Why:` 兜底把 bug 推到下游隐藏。
- **B3**　禁防御性 `if v == nil { return default }`。`Why:` 见 0-4。
- **B4**　panic 仅用于"不可能到达"分支；不替代正常错误传播。`Why:` 让"应该 panic"和"应该错误"清楚分开。

---

## §M — 注释纪律

- **M1**　默认不写注释；命名要好到不需要注释。`Why:` 注释是债务，错的比无更危险。
- **M2**　仅在 WHY 不显然时写一行。`Why:` 命名 + 代码应自解释。
- **M3**　禁 WHAT 注释（`// increment counter`）。`Why:` 重复 code。
- **M4**　禁引用当前 task / fix / 调用方（`// used by X` / `// fixes issue #123`）。`Why:` commit message 的事，写代码里会腐烂。
- **M5**　禁多段 docstring；package doc 一行说清职责。`Why:` 多段必腐烂。

---

## §V — 提交前自检

- **V1**　`grep -rE '已退役|已废弃|deprecated|TODO.*迁移|no-op stub|短路|兜底' <changed>` 命中 = 没扫干净。`Why:` 机械化诚实。
- **V2**　`grep -rE '"(EnterRoom|InBed|LeftBed|Fall|...)"' <changed>` 命中 = 没用常量。`Why:` 见 N4。
- **V3**　`go vet ./... && go build ./...` 全绿才提交。`Why:` 基本闸。
- **V4**　改 v2 service 必 grep `main.go` capability interface。`Case:` Z-004 type-assert 静默吞掉。

---

## §UI

- **UI1**　执行用 ID 展示用 name：INET CIDR / UUID 不直接渲染到屏幕；a-select 用 `:options` 显式 label，a-table dataIndex 指向 name 字段。`Why:` 用户认 name 不认 CIDR/UUID。
- **UI2**　Transfer 双栏：Current 在左 / Available 在右（与 antd 默认相反）。`Why:` 本项目约定。

---

## §R — 改前读 / 改后复查

- **R1**　任何代码改动前必读对应业务设计 doc；改后枚举边界场景复查。`Case:` bed_presence_fusion dedup 没读 engine Vacant 行为 → bathroom 卡死 6min。
- **R2**　Memory 中包含具体函数 / 文件 / flag 名的条目使用前先 verify 存在。`Why:` 重命名 / 删除 / 未合并都可能；memory 是某时点快照。

---

## §Dev — 开发阶段

- **Dev1**　允许直接 stop 整个 owlback；cutover 时必整套停。`Why:` 冷启动暴露跨服务 wiring 问题。
- **Dev2**　改完代码无需问授权，直接 `sudo systemctl restart owlback.<name>`。`Why:` dev 阶段已授权。
- **Dev3**　本机 dev 环境：所有 Bash 命令一律预先授权（含删除）；**唯一附加纪律 = 删除文件/数据（`rm` / `DROP` / `DELETE` / `TRUNCATE`）执行前我主动做一次二次确认**。`Why:` 2026-05-31 用户授权全部命令免反复弹窗；删除不可逆故保留一次人工确认。连接：DB `127.0.0.1:5432` postgres/postgres/owl_v2；Redis `127.0.0.1:6379` 密码 `TeLunSu-36kr`。

---

## §W — 工作流

- **W1**　多步 plan 主线 / 支线物理隔离。`Case:` 见 0-2。
- **W2**　完成项打 `~~删除线~~` 保留审计痕迹，不直接删条目。`Why:` 审计回溯。
- **W3**　新规则必有依据（memory entry / CLAUDE.md / 本会话决议）。`Why:` 防凭空捏造。

---

## §Z — Anti-pattern 反例库

> 每次踩坑命名后加进来；后续讨论引用 `Z-XXX` 不用复述。

- **Z-001 Mom unbind PK collision** — `device_ipv6` 当 PK + `reset_device_prefix` 在 unbind 时重写 → fixture MAC 重复 → PK 冲突。修复方向：device_id 转稳定 IPv6（data_v2_todo #2）。2026-05-21。
- **Z-002 has_bed producer/consumer drift** — `card_reconcile.go` 写 structural / cardagg `CardMeta.HasBed()` 读 ActiveBed 语义，6mo+ 反向 drift；workaround = `HasBedBoundRadar` 兜底。2026-05-21 解决方向（修订）：**producer 对齐 consumer**（has_bed = ActiveBed，见 C3）—— consumer 的 sensor 使用语义是对的，producer 写 structural 是 bug；改 `card_reconcile.go` has_bed 填充走 device 绑定状态、device 绑/解绑事件触发 snapshot 更新（C10+C12），删 `HasBedBoundRadar` workaround（入 data_v2_todo）。
- **Z-003 upsertSpaceCard 整段 stub** — v1→v2 cutover 把 `card_sync_service.upsertSpaceCard` 短路成 `SELECT 1`，真实写卡转到 `card_reconcile.go`，但 stub 路径未删 → 两套 writer 并存。规则违反：S5。
- **Z-004 v2 cutover type-assert silent drop** — `main.go` `if-ok` 注入 v2 缺方法，整段下发不执行无报错。检测：V4 grep。

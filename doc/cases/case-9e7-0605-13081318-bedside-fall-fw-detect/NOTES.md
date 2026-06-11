# bedtest-0605-2 — 床边跌倒 / 固件 pose2→5 判出 / EMERG / sleepad 矛盾待确认

- 设备: radar 9D8A326309E7 (9e7) + sleepad BM87224700978 (978)
- 窗口: 2026-06-05 13:08:30–13:18:45 MDT = 19:08:30–19:18:45 UTC
- 数据: `test_record.txt`（脚本 `scripts/export_bed_test_record.sh` 生成）

## 时间线要点
- 19:08:36 radar 固件 InBed（mqtt rx 原生，早于 sleepad ~3min）
- 19:11:34 sleepad InBed（接触式权威晚锁，要躯干压稳+锁 HR）
- 19:17:07 Walking → **19:17:36 Fall pose=5 → 告警 EMERG(level=0) firmware_radar_fall**
- 19:17:58 Fall end → 告警 **auto_resolved**（持续 22s）
- 19:18:19 radar LeftBed / 19:18:26 sleepad LeftBed / 19:18:28 ExitRoom

## 关键结论
1. **固件这次判出 pose 2→5** → 走固件直通报警（非段4c）。对照 #1 固件 pose 全程=3 没判出。
2. **firmware InBed/LeftBed 成对、均固件原生**：19:08:36 InBed（trace 9e7.9999）、19:18:19 LeftBed
   （trace 9e7.11081），qinglan `mqtt rx` 实证。固件看躺下即判 InBed，比 sleepad 早 ~3min。
   （注：sensor 派生的是 LeftBed *告警*，与事件流 event_log 的 InBed/LeftBed 是两条流，勿混。）
3. **★硬矛盾（待物理确认）**：跌倒全程（19:17:07–19:18:23）sleepad 铁证 bed=0 + HR 稳 69 +
   body_move=0，LeftBed 迟至 19:18:26（躯干压垫上，真离床才发，8–15s 去抖在压力离开后起算）；
   radar 同刻 pose2→5 摔地 + 走出房门。一个稳躺垫上的身体不可能同时摔地再走人。
   - **单人半挂床沿**（躯干在垫→sleepad 读，身体垂地→radar 摔）最自洽；或**两目标**。
4. **待拍板**：
   - Fall 告警 `auto_resolved`（22s 自动消解）对老人护理是否合适，要不要改人工 ack。
   - level=0 EMERG（比 CRITICAL 还高）确认是否预期。
   - 固件 fall 直通**未对账 sleepad** 的"床上人稳着"反证（passthrough 不交叉验证）；若是单人误报情形，
     sleepad 本可压制/降级这条 EMERG。

# Xsensor-replay.md — 真 case 重放手册 + 本会话摘要

> 验证哲学（铁律 [[validate_real_case_no_unit_tests]]）：**禁止 unit test**，用真实 case 解剖真生产系统
> （NASA 地面放 100% 真航天飞机排障）。验证 = 跑真 `cmd/xsensor` on `tools/replay` 重放的真 case，看
> `xsensor_xray` 全景日志逐 tick 切片。

## 1. 重放操作手册

```bash
# ① 导出真 case（PG → fixture window.json+meta.json+sleepad）。d5f7 用 Denver，d523 用户的时间是 MDT
./tools/export/export.sh case-<last4>-<MMDD>-<startHHMM><endHHMM> --tz America/Denver --year 2026
#    或原始 ms：./tools/export/export.sh <uid> <start_ms> <end_ms> <case_name>

# ② 建 xsensor（最新含 forensic/FN-safe/镜像窗）
go build -C tools/Xsensorv1 -o /tmp/xsensor ./cmd/xsensor
# 建 replay（owl-tools module，在 tools/ 下建）
cd tools && go build -o /tmp/replay ./replay && cd ..

# ③ 跑 xsensor（消费 test:*，只读不写生产；日志 → /home/wisefido/owl/log/Xsensor.log）
set -a && source .env && set +a
for s in monitor event alarm; do redis-cli -a "$REDIS_PASSWORD" -n 0 DEL "test:iot:$s:stream" >/dev/null; done
: > /home/wisefido/owl/log/Xsensor.log
nohup /tmp/xsensor > /tmp/xsensor.stdout 2>&1 &   # sleep 6 确认存活再放

# ④ 重放（4x）
nohup /tmp/replay -fixture doc/cases/<case_dir> -speed 4 -stream-prefix "test:" -streams monitor,event > /tmp/replay.out 2>&1 &

# ⑤ 分析：grep 房间 room_id 从 Xsensor.log，解析 xray JSON（s_dist/dbn/target/walls）
grep "<room_id>::/88" /home/wisefido/owl/log/Xsensor.log | python3 -c "..."
```

**安全 + 坑**：
- 生产 wisefido-sensor 在 Redis **DB0** 跑（PID 长驻）。xsensor 硬接 `test:iot:*:stream`（与生产 `iot:*` 隔离）；
  xsensor 对生产**只读**（SaveToRedis 从未被调用）。`tools/replay` 100% 忠实（仅平移 time、全帧不丢）。
- **xsensor/replay 的 Redis DB 硬编码 0**（不读 env），靠 `test:` 前缀隔离。
- 杀进程用 `pkill -9 xsensor`（按 comm）；`pkill -f /tmp/xsensor` 会匹配到自己命令行误判。
- 起 xsensor 必从 owlBack 根 `source .env`（cwd 在 tools/Xsensorv1 时 source .env 失败 → REDIS 认证断 → 静默退出）。
- 不能 `sleep 100`（被 block）；用 `until grep -q 完成 /tmp/replay.out; do sleep 3; done`。
- **track_id==88 = 固件 no-target 心跳**（全零帧），track_manager 真路径丢之；它是 no-people 信号，DBN 清 logicID 用。
- **多 radar 房**（如 d523 fd00:0:3:111:3:100::/88，2 雷达坐标系未 co-register）→ 几何（reflectSep）confound，sep 值不可信。

## 2. 本会话的 test cases（全在 doc/cases/）

| case | 场景 | 重放结论 |
|---|---|---|
| `case-d5f7-0524-13351357` | bathroom 真摔，pose 误判 Sit | ✅ belief 自爬 Sit→Fallen(0.016→0.997) fire——**silent fall 经 belief 检出**，pose 错不漏 |
| `case-d5f7-metal-0128` | metal 反射 2-track（01:28 UTC，192 个 2-track 时刻）| ✅ 反射 lid2 收敛 Mirror(PM=0.97 via WallMargin)排出 N_r；真人 lid1 PR=1.0；**ghost filter 真数据验证通过** |
| `case-d5f7-0616-07400750` | metal + track_id swap（早上）| metal 先到独存 60s→被锚 real，真人后到判 ghost（**无解**）；揪出 **§79 FN：ghost 谷底否决 fall**→催生 FN-safe 修复 |
| `case-d523-0526-21402145` | mirror，1 人起 +40s 出现 mirror（later-born）| mirror(lid2)经几何被判（PM 升）但**揭出 radar-on-wall + 多radar geometry confound**（sep=405 假值）|
| `case-d523-0611-22262238` | **生产误报** Fall，radar border + 2 track（22:31 MDT）| ❌ **xsensor 也误报**——近场静物伪迹(生在 radar 原点 z=0 pose=4)静止→dwell(still≥60s)→SFallen→fire。**根因=silent 阈太短** |
| `case-d523-0612-22052220` | 走进来→7min 报 fall（22:05-22:12 MDT）| **未重放**（下个会话）|
| `doc/cases/d523-mirror-ghost-0526/` | README 源（P0 ghost 38min/P1 flicker/P2 real）| 40min 全段重放过：P0(先到持久 mirror)判 Real 没 flag（born-first 无解 + README 的长寿命/空间范围信号没接）|

## 3. xsensor 本会话改动（全 commit+push）

| commit | 内容 |
|---|---|
| `98ff1de` | realness 重写为 2 态转移矩阵（零 gate）+ 删 aScore 超速 |
| `ce0b8cf` | 删全部 unit test + replay 旁路 harness（验证走真 case） |
| `9ddd0aa`/`1f8f4cd` | 6A：realness 模型 + ghost 判定原因 R1-R6 |
| `4582898` | forensic 入 xray（sep/wallM/rho/later/canvas x,y/walls/radar）|
| `e0dec8f` | **FN-safe 治本 §79：realness 绝不否决 fall**（pFallReal≡1 + eligible≡true，只喂 N_r）|
| `7879e1f` | 镜像/sync 判定只出生 5s 算+冻结省算力 + 孤轨永 Real override；6A 交 C |

## 4. 关键架构结论

- **realness 绝不否决 fall**（[[realness_never_vetoes_fall]]）：漏报≫误报 + 有 ghost 永凑不齐 95% 把握；realness 只经 N_r→C_FN 折扣。
- **镜像几何 = radar→track 连线穿任何墙(外/侧)→反射，取距 track 最近交点**（reflectSep 现有逻辑**正确**，point-in-polygon 被否——侧墙 mirror 时 real/ghost 全在多边形内会漏）。radar 壁挂/坐标粗画不 special-case。
- **latch 否决**（分不出真伪就不该 latch，z 不准）。
- **metal-born-first / ghost-born-first 无解但 FN-safe**（fall 不靠 PR 否决）。

## 5. 待办（下个会话主线）：silent/moving fall 参数移植

**0611 误报根因 = Xsensorv1 silent 阈太短**：
- Xsensorv1 emission `stillTau = 60s`（统一一把尺）→ dwell 60s 就喂 SFallen → 近场静物伪迹 2.7min 即误火。
- **产线 `wisefido-sensor/internal/roomengine/fall_rules_param.go` 有分区/分钟级阈值**：
  - **Still Fall（silent）`stillFallParam`**：Toilet/Shower **15min** / Deny **5min** / Default **8min** / 非风险 ×1.2。按 cell areaType 分。
  - **Lost Fall（moving）`lostFallParam`**：`MovingPreconditionMs=60s`（消失前 still-box <60s=走动中才算 lost-fall）、StillBoxCm=50、ExitDistMinCm=30、DistanceGateCm=500。
  - **`cellHistoryParam`** 自适应阈（fake 高发区收紧 tolerance）。
- **方向**：把 fall_rules_param.go 的差异化阈值（silent 分区分钟级 + moving precondition）**移植**进 Xsensorv1 decide/emission（"移植 production 非另造"），治本 0611 这类 silent 误报、不动 moving 真摔灵敏度。
- ⚠️ 注：Xsensorv1 现 decide.go 只有单阈 `pFireHi=0.55`+Λ+T_hold，**无 silent/moving 分支**；emission 只有 silent 路径（dwell），**无 moving-fall**（kinematics Δz 已删 P2.1）。

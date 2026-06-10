---
name: feedback-store-raw-not-derived
description: 缓存层只存 DB raw 值，转换/分类/桶映射在 query 时现算；建 cache 时禁止 load-time 派生
metadata: 
  node_type: memory
  type: feedback
  originSessionId: c9715c78-74f8-4b78-864e-2103493a53a0
---

**规则**：缓存层存的是 DB raw 值（与 SQL SELECT 返回的列类型对齐），任何"text → enum / size → bucket / 状态 → ts 派生"都在公共方法 query 时现算。

**禁忌**：
- ❌ `RoomType int` 字段 — 把 raw `room_type` text（"bathroom"/"kitchen"/"restroom"）在 LoadAll 时跑 `classifyRoomType()` 转 enum 存进去
- ❌ `BedSize string` 字段 — 把 raw `size_kind` text 在 LoadAll 时跑 `BedSizeBucket()` 转 small/large 存进去
- ❌ `lastAloneAnchorByCIDR map[string]int64` — 镜像 engine ZoneState.AloneContinuousTs 到 publisher 本地
- ❌ `AloneSinceTs` 当 RoomState 字段（FE 拿 ts 算分钟）— 应该 sensor 现算 `AloneContinuousMin` 整数推 FE

**正确**：
- ✅ `RoomTypeText string` 存 raw text；`IsBathroom()` query 时调 classifier
- ✅ `SizeKindText string` 存 raw text；`BedSizeBucket()` query 时调 bucket fn
- ✅ engine.GetState 直查 ZoneState；不在 publisher 缓存镜像
- ✅ sensor 60s tick 算好 `AloneContinuousMin`/`StandingContinuousMin` 推 cardagg；FE 直读不算 ts

**Why**：
1. 转换规则改 → 只改公共方法不需重 load cache（cache stay valid）
2. 中间态镜像 = 漂移源（engine 单源 vs publisher 副本两边不一致）
3. RoomState 暴露 ts 给 FE → client 时钟不一致问题，FE 跨 timezone 算错

**2026-05-25 实测教训**：连续 3 次同一错误模式——
1. publisher.lastAloneAnchorByCIDR cache 镜像 engine anchor → 砍掉用 engine.GetState
2. RoomState.AloneSinceTs（ts） → 改 AloneContinuousMin（int，sensor 60s tick 算好）
3. SpatialCache.Entry.RoomType (int) + BedSize (string) load 时派生 → 改 RoomTypeText / SizeKindText raw 字段

**How to apply**：
- 写 cache struct 字段前问："这值 DB 列存什么类型？" — 跟 DB 列对齐
- 看到 `LoadAll/refresh` 里调 `classify*` / `bucket*` / `derive*` 等转换函数 → red flag，应该挪到 query 入口
- 镜像 / 副本 / cache 同一份数据存两处 → 必有漂移；统一指向单源（engine / aggregator / DB）
- 对外接口（推给 FE / 跨服务 publish）的 ts 字段 → 多想一遍 client 是否需要 ts，能不能 server 端算好再推

**变体**：
- 派生分类（text → enum）：raw text 存 cache
- 派生桶（数值 → 区间）：raw 数值存 cache
- 派生时长（ts → min）：ts 留 producer 端，consumer 接收 min 整数
- 镜像状态（engine anchor → publisher cache）：永远 query 单源

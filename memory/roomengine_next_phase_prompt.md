---
name: RoomEngine 下一阶段 prompt
description: feat/radar-fall-verifier 分支 PR-15.x 后的工作清单（重新启动 conversation 时用）
type: project
originSessionId: 08872106-789f-48c2-8a3f-bcb3bd7092ef
---
继续 wisefido-ai roomengine 在 `feat/radar-fall-verifier` 分支的开发。
基线 commit = `0e0a44b`（PR-15.5 落地 + Kitchen 误报文档化）。

## 最新已落地（PR-13 → PR-15.5）

- **PR-13**: AreaSit 区域静止+Z 突变自学习（2min static + |dz|≥10cm + 90% 容忍）
- **PR-14**: silent fall 仅留 LeftBed 矛盾路径 + 双源 InBed 入场门控（删除旧 Path 1）
- **PR-15.0/.1**: Auto-Deny 5-cell 软共识 + RealDecay<5 容忍偶发轻触
- **PR-15.3**: 软共识 + 15 天时间门控 + Schema v6 持久化
- **PR-15.4**: Auto-Deny 改 BFS 距离场（depth≥2 或不可达）+ 自纠错（Deny 被走过→回退 Unknown）
- **PR-15.5**: 时间门控默认 10 天（实测 vs 15 天加速 5 天，结果一致 200 cells）
- **daily layout reload**: 每天 22:00 local 自动重读 layout，hash 变即重置 grid
- **playback --cycles N**: 多循环 + `--start/--end` 人类可读时间戳

## 下一阶段重点（按优先级）

1. **PR-16 prod 部署观察** — 重启 wisefido-ai 服务应用 PR-15.x，观察 D523 /
   Kitchen / LivingRoom 实测 14-21 天的 Auto-Deny 学习效果。验证：
   - 10 天后 furniture 内核被自动学到 AreaDeny
   - Walk 区不被误标
   - 自纠错（家具搬走→Deny 回退 Unknown）

2. **PR-17 自纠错 prod 验证** — 测试场景：手动改 layout 移除 Furniture 矩形，
   等 daily reload 22:00 触发，验证旧 SourceLearned Deny cell 在被走过后正确
   回退 Unknown 并被 Walk 规则重学

3. **PR-18 前端 layout 工具增强** — Wall 边变虚线 = 自动生成 Enter 矩形（30cm
   厚 inward）。后端 schema 不变，仅前端 UX 简化。详见会话讨论中的混合方案

4. **PR-19 Kitchen 类房间智能 stay-alarm** — 区分场景：
   - 老人独居 + 自己做饭：Kitchen 需 still-fall（人在 counter 前真倒下）
   - 养老机构集中厨房：不部署 radar，避免做饭误报
   - 已记 doc/AI_fall_detect.md §16.1

## 参考文件

- **完整设计**: `owlBack/doc/AI_fall_detect.md`（含 §16 已知误报）
- **D523 实测**: `owlBack/doc/playback-D523-3d-PR15.html`
- **Kitchen-min 13 天实测**: `owlBack/doc/playback-Kitchen-min-4cycles-PR15.5.html`
  （200 DenyAI cells 在 cycle 4 触发）
- **playback 命令模板**:
  ```
  go run ./cmd/roomengine-playback \
    --layout LAYOUT_PATH \
    --room "ROOM_NAME" \
    --uid DEVICE_UID \
    --start "2026-04-26 16:00:00" \
    --end "2026-04-29 22:00:00" \
    --cycles N \
    --snap 30 \
    --out ../doc/OUT.html
  ```

## 关键参数（默认值）

```go
AutoDenyMinPersistDays:    10  // 持续 10 天才升 Deny
AutoDenyTraverseTolerate:  3   // 5-cell 软共识 < 3 算未走过
AutoDenyTraverseReset:     5   // >= 5 重置时间计时
realDecayDenyTolerance:    5   // RealDecay < 5 容忍偶发
NearTraverseDeny:          20  // 已弃用，PR-15.4 BFS 取代
```

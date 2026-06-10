---
name: deploy-mode-by-domain
description: 部署级约定：test.wisefido.com 跑 sandbox / owl.wisefido.com 跑 release；通用 sandbox/release 行为开关（AI override 等）按域名/部署单元配
metadata: 
  node_type: memory
  type: reference
  originSessionId: 0566c6c8-ddbd-4e2d-b8b6-79fef2a5a715
---

## 域名 → 部署 mode 映射

| 域名 | 部署 mode | 含义 |
|---|---|---|
| `test.wisefido.com` | **sandbox** | 演示 / 测试 / 开发环境；副作用类开关默认旁路（仅 log，不动 prod 决策） |
| `owl.wisefido.com`  | **release** | 生产；副作用类开关全开（写入 / 覆写 / 派生 alarm 走真路径） |

## 当前应用此约定的开关

| 服务 | 配置字段 | env override | 默认 | sandbox 行为 | release 行为 |
|---|---|---|---|---|---|
| cardagg | `ai.override_mode` (config.yaml) | `CARDAGG_AI_OVERRIDE_MODE` | sandbox | AIVerdictHandler 收到 verdict 仅 log；monitor.Apply 不动 track_confidence | Apply 覆写 track_confidence + ai_source → FE 按 AI 调整后值渲染 |

## 实施约定

新加 sandbox/release 双模式开关时：
1. **代码侧**：config struct 加 `mode` 字段；main 用 `cfg.<x>.Mode` 不写硬编码；env override 用 `<SERVICE>_<FEATURE>_MODE` 命名
2. **默认 sandbox**：保护 test 环境（默认不开副作用）；release 必须明确配置
3. **deploy 侧**：每环境 deploy manifest 注入对应 env var；test 不动（用默认）/ prod 显式 release
4. **memory**：新开关同步进本表

## 关联

- [[cardagg_v1_to_v2_migration_audit]] §S5a — AIOverrideCache sandbox/release 双模式起源
- v1 `wisefido-cardaggv1/internal/service/ai_overrides.go` — 模式 enum 设计原型

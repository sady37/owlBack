---
name: feedback-dev-restart-services
description: 开发阶段可直接 sudo systemctl restart owlback.* 服务，不必每次询问授权
metadata: 
  node_type: memory
  type: feedback
  originSessionId: b9c43444-5a87-4a7c-9a57-e98422659a2d
---

开发阶段（owlBack 各 owlback.* 服务）可以直接 `sudo systemctl restart owlback.<service>` 重启，不必每次问。

**Why:** 用户明确指示 2026-05-13；当前 owlback 各服务（cardagg/qinglan/sleepace/data/sensor/iot 等）都在 dev/test 阶段，重启无外部用户影响。每次问会拖慢迭代节奏。

**How to apply:**
- 改完代码 → build 通过 → 直接 `sudo systemctl restart owlback.<name>` → 看 journalctl 验证
- 适用范围：所有 `owlback.*.service`（单一开发机器）
- 不适用：生产部署 / 数据库 drop / 共享 BIND/kea 系统服务（仍需先确认）
- 若 classifier 拦截，向用户说明并请用户给出 Bash 权限规则

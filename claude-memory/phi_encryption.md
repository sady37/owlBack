---
name: PHI Encryption Architecture
description: resident_phi全字段加密已部署，K服务架构，MW双因素恢复机制
type: project
originSessionId: 81929375-86e8-4d80-919d-0a8cc5b3c58d
---
PHI加密于2026-04-14部署完成。

**架构**: K服务(独立进程, unix socket /tmp/owl-kms.sock) → wisefido-data通过MASTER_PIN获取tenant_key → AES-256-GCM全字段加密

**关键决策**:
- 全表加密（含bool/int/float字段），不区分敏感/非敏感
- MW = Matrix Wallet（非MV），12×7查询表用于K重启恢复
- master_pin单独交付，不放在MW打印件中
- plus_code限流5次/天/用户
- K不在加解密热路径，只在启动/新增租户时调用一次

**Why:** HIPAA合规要求，用户要求简化加密范围判断（全部加密更省事）

**How to apply:** 修改resident_phi相关代码时，所有字段必须通过_enc列读写；新增字段也需加_enc列

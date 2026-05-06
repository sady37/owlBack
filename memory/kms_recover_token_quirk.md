---
name: KMS recover 双 token 模式
description: KMS recover 的 --mw-token 取值取决于 archive 来源；master_pin 永久不变；详见 owlBack/doc/kms.md
type: project
originSessionId: 8c59ce7c-75cf-4a83-aa46-6e62300e572e
---
**master_pin 是 init 一次性生成的永久凭证**（写 .env，跨重启从不变化）。`--recover` 不重新生成。本机值 = `FtjPuGB8`。

**两类 archive，两种 unwrap 口令**：

| archive 来源 | 加密用什么 | recover 时 `--mw-token` 传什么 |
|---|---|---|
| `--init` 写 (master_key-20260417.json) | mwToken(seed, 当日 month, 当日 isoWd) | 查 mw.pgp 找 init 当日单元格 = `FCKE7K` (Apr×Fri) |
| `--recover` 写 (后续 master_key-*.json) | master_pin | 直接传 `.env` 里的 MASTER_PIN = `FtjPuGB8` |

**Why:** init archive 故意用 MW 封装 = 灾难恢复跳板（.env 丢失也能用 MW 纸质表 + GPG PIN 救回）；recover archive 用 master_pin 封装 = 日常运维钥匙（master_pin 已是常驻口令，无需再依赖 MW）。这是设计意图，不是 bug。

**陷阱**：cleanupArchives(vaultDir, keep=2) 在第 2 次 recover 之后会删 init archive！必须在 init 后立刻把 `master_key-<init日期>.json` 异地备份，否则 MW 灾备能力失效。

**How to apply:** 重启 KMS 时先看用哪份 archive：还是 init 那份就查 mw.pgp，否则直接用 .env 的 MASTER_PIN 当 mw-token。`--vault-dir` 必须传 `owlBack/kms`（与 init 时一致），否则新 archive 散到默认 `owlBack/vault/`。

# KMS (owl-kms) — PHI 加密密钥服务

> 本文基于源码 [owlBack/kms/main.go](../kms/main.go) + [owl-common/crypto/wrap.go](../owl-common/crypto/wrap.go)。
> 原始设计文档 `owl/kms.md (v2)`（commit 10f8e16 引用）已丢失，本文替代之。
>
> **2026-06-10 设计复位**：早前 `doRecover` 用 master_pin 重封盘上 archive，使 master_pin（盘上常驻）单独即可解封 KMS——离线第二因子被架空（偷盘=拿 PHI）。已改回原始设计：**盘上 archive 始终用 MW 格子封，每次重启必须人工查离线表解封，master_pin 退回纯在线 `/tenant-key` 认证、绝不碰盘**。同日完成数据迁移（archive 重封为 MW，销毁所有 master_pin 封残档 + 自动解封旁路 + 明文 MW 表）。

---

## 0. 原理与心智模型（先读这节，后面就不会记反）

### 0.1 解密链：PHI 是被谁解开的

```
   ── 解封（每次重启，离线人工）────────────────────────────────
   盘上 archive ──[ MW 格子 解封 ]──▶  masterKey
   master_key-*.json                 （32B，只在 K 内存，永不落盘、永不变）
                                            │
   ── 取钥（serve 时，在线，每 tenant 一次）──┤
   业务请求带 master_pin ─[ 解开内存 masterKeyPGP ]┘
                                            │
                          + deploymentSalt + tenant_id ,  HKDF-SHA256
                                            ▼
                                       tenant_key（每 tenant 一把，32B）
                                            │
                                       AES-256-GCM
                                            ▼
                                  resident_phi.*_enc（库里密文）
```

**两道闸守着同一个 masterKey**：离线闸（MW 格子，重启时把 masterKey 载入内存）、在线闸（master_pin，serve 时凭它取用）。PHI 始终由 **masterKey**（经 tenant_key）解开——archive 本身不解 PHI，它只是 masterKey 的保险箱。

### 0.2 四件套：什么变、什么不变、在线还是离线

| 东西 | 变不变 | 在哪 | 角色 |
|---|---|---|---|
| **masterKey** | **永不变**（变=全库 PHI 报废） | 只在 K 内存 | 解所有 PHI 的根钥（经 tenant_key） |
| **deploymentSalt** | **永不变** | 随 archive 存（`arc.Salt`） | tenant_key 派生的固定 salt |
| **archive** (`master_key-*.json`) | 文件名/封印格**每次重启轮换**；**里面 masterKey 永远同一个** | 盘上 + **异地备一份** | 锁着 masterKey 的保险箱 |
| **MW 格子** | **每次重启轮换 → 近一次性** | 离线纸表 / `mw.pgp` | 开保险箱的**离线**钥匙 |
| **master_pin** | **静态常驻、不过期、可重置** | `.env`（盘上） | **在线**认证：业务取 tenant_key + 封内存副本 |

### 0.3 三条反直觉但关键（最易记反）

1. **master_pin 是"外层口令"不是 salt**。它是 Argon2id 的 passphrase（[wrap.go:21](../owl-common/crypto/wrap.go#L21)），派生出 AES key 去**封内存里的 masterKey**；不参与 tenant_key 派生。改它 = 用新口令重封同一个 masterKey，**数据一字不动** → 任意一次重启都能换（代价：同步各业务 `.env`）。

2. **任一份 archive 永久够用**。任何一份（哪怕旧的）archive + 它**文件名日期**那格 MW = 解出**同一个** masterKey = 解**全部** PHI（含以后才写入的）。所以离线**备一份就一劳永逸**，不必追最新那份。

3. **静态的是 master_pin，轮换/一次性的是 MW 格子**——别对调。master_pin 跨重启永远同一个字符串、不会"失效"；每次重启要重查的是 MW 格子。

> **双因子真义**：在线=master_pin（盘上）、离线解封=MW 格子（盘外）。**盘上没有任何能自动解封的凭证 → 偷整盘 ≠ 拿 PHI**。这就是 2026-06-10 设计复位守住的核心不变量。

---

## 1. 架构

```
   K 服务（独立进程）                  wisefido-data（业务进程）
   ┌──────────────────────┐          ┌──────────────────────┐
   │ 内存:                │          │ 内存:                │
   │   masterKeyPGP       │◄──unix──►│   tenantKey (per     │
   │   deploymentSalt     │  socket  │   tenant, 派生缓存)  │
   └──────────────────────┘          └──────────────────────┘
            │                                   │
            ▼                                   ▼
   owlBack/kms/master_key-YYYYMMDD.json     PG: resident_phi.*_enc
   owlBack/kms/mw.pgp                       (AES-256-GCM 字段加密)
   .env: MASTER_PIN, KMS_SOCKET
```

**特性**

- K 独立进程，监听 unix socket（默认 `/tmp/owl-kms.sock`，权限 0660）
- K **不在加解密热路径上**，仅在业务进程启动时**每个 tenant 调用一次** `/tenant-key`，业务拿到 tenant_key 后自行 HKDF + AES-GCM
- masterKey 永不落盘明文，永不离开 K 进程内存
- HIPAA 合规：resident_phi 全字段加密（含 bool/int/float），不区分敏感/非敏感

---

## 2. 三个凭证与其角色

| 凭证 | 长度/来源 | 保管位置 | 生命周期 | 角色 |
|---|---|---|---|---|
| **master_pin** | 8 字符随机 ([main.go:83 `randAlphanumeric(8)`](../kms/main.go#L83)) | `.env: MASTER_PIN` | init 生成；**可重置**（只是在线口令，不封 masterKey） | **仅在线认证**：业务进程调 `/tenant-key` 的口令、封 K 内存里的 masterKeyPGP。**绝不解封盘上 archive** |
| **MW 表** | 12 行 × 7 列，每格 6 字符 base32 | `kms/mw.pgp` (GPG 加密) + 客户纸质件 | **init 一次生成，永不变** | **唯一解封钥匙**：每次 KMS/服务器重启，用离线表查格子解盘上 archive；每次重启轮换格子 |
| **GPG PIN** | 用户自定义口令（`--init --pin`） | 运维记忆 / 密码管理器 | 永久 | 解 `mw.pgp` 看到 MW 表 |

**职责分离（非"盘上双因子"）**：
- **在线服务**（K 活着）：业务凭 `master_pin` 取 tenant_key。master_pin 的效力**绑定在"K 已解封"的活会话上**。
- **解封**（K 死/重启）：K 内存清空 → masterKey 没了 → **master_pin 单独无能为力**，必须人工查 **MW 表** 解盘上 archive，把 masterKey 重新载入内存。这是离线人工因子，盘上没有任何能自动解封的东西。
- **致命口令排序**：丢 master_pin = 不致命（重置一个新的，更新 `.env` + recover 时 `--master-pin` 传新值即可，masterKey 不动）。丢 **MW 表 + GPG PIN** 且无 archive 异地备份 = **masterKey 永久无法解封，全部 PHI 报废**。

---

## 3. 三种运行模式（源码行为）

### 3.1 `--init` （首次部署，**只跑一次**）

[main.go:80-129 doInit()](../kms/main.go#L80-L129)：

```
1. 随机 32B masterKey
2. 随机 8 字符 masterPin              ← master_pin 唯一一次产生
3. 随机 16B deploymentSalt
4. 随机 24B mwSeed → 渲染 12×7 MW 表（mwToken 函数 hash 出每格）
5. masterKeyPGP = WrapMasterKey(masterKey, masterPin)             ← 内存（masterPin 包装）
6. archive      = WrapMasterKey(masterKey, mwToken(今日))          ← 写盘（MW 当日 token 包装）
7. mw.pgp       = GPG对称加密(MW表, GPG-PIN)                       ← 写盘
8. zero(masterKey) + 删 mw.md 明文
9. 打印 master_pin / MW_today / archive 路径到 stdout
```

**输出文件**：
- `<vault-dir>/master_key-YYYYMMDD.json` — archive，**用当日 MW token 加密**
- `<out-dir>/mw.pgp` — 12×7 MW 表，**用 GPG PIN 加密**
- stdout 提示把 `MASTER_PIN=<8字符>` 与 `KMS_SOCKET=/tmp/owl-kms.sock` 写入 `.env`

**为何盘上 archive 用 MW token 而非 master_pin？**
解封必须是离线人工因子。master_pin 在 `.env`（盘上），若它能解盘上 archive 就等于偷盘=拿 PHI。所以盘上 archive 一律 MW 封（init 与之后每次 recover 同理，详 §3.2）；master_pin 仅封内存副本作在线认证，且**可重置**——`.env` 丢了只是重新定一个新 master_pin，masterKey 不受影响。

### 3.2 `--recover` （K 重启 / 服务器重启）

[main.go doRecover()](../kms/main.go#L131)。**两个 MW 格子,由运维当场查离线表提供**:`--mw-token` 解旧、`--reseal-token` 封新。

```
1. masterPin: --master-pin flag 或 fallback env MASTER_PIN   (仅用于步骤 4 的内存封装)
2. 读 archive 文件 → 反序列化（失败即 fatal）
3. masterKey = UnwrapMasterKey(arc.Wrapped, unsealToken)
   ↑ unsealToken(--mw-token) = 该 archive **文件名日期**对应的 MW 格子
     GCM 带认证：格子错 → "unseal failed: message authentication failed" 当场报错
4. masterKeyPGP = WrapMasterKey(masterKey, masterPin)        ← 内存（在线认证闸，master_pin 封）
5. reWrapped    = WrapMasterKey(masterKey, resealToken)      ← 盘上新 archive（**MW 格子封**）
   ↑ resealToken(--reseal-token) = **今日**那格 → 每次重启轮换封印
6. 写 <vault-dir>/master_key-<今日>.json
7. cleanupArchives(vaultDir, keep=2)                         ← 按文件名（=日期）字典序留最新 2 份
8. zero(masterKey)
```

**为何盘上 archive 始终 MW 封、且每次轮换？**
- **始终 MW 封**：解封是离线人工因子。master_pin 在盘上（`.env`），若它能解盘上 archive，就等于"偷盘=拿 PHI"，离线因子失效。所以 master_pin 只封步骤 4 的**内存**副本（在线认证用），**绝不**封盘上 archive。
- **每次轮换（用今日格而非复用一格）**：让"抄到一次格子"的人在下次重启后失效——能持续恢复必须持续能查离线表，而非一次性记忆。固定一格则一次泄露=永久可解。

> **masterKey 永不变**：步骤 3 解出、步骤 5 原样重封，deploymentSalt 复用 `arc.Salt`。"轮换"换的只是外层封印格子，masterKey 与 salt 一字不动——否则全库 PHI 解不开。

> ⚠️ **单点 + cleanup**：archive 每次轮换、`keep=2` 只留最新 2 份本地。整盘丢失即 masterKey 没了 → **必须异地备份至少一份 MW 封 archive + 记下它的日期（=解它的格子）+ mw.pgp + GPG PIN**。这是唯一的灾难恢复底牌。

### 3.3 `serve` 模式（持续运行）

[main.go:153-200](../kms/main.go#L153-L200)：

监听 unix socket，提供：

| Endpoint | 用途 | 请求 | 响应 |
|---|---|---|---|
| `GET /health` | 健康检查 | — | `{"status":"ok"}` |
| `POST /tenant-key` | 业务进程取 tenant_key | `{tenant_id, master_pin}` | `{tenant_key: base64}` 或 401 |

`/tenant-key` 实现：
1. 用请求里的 `master_pin` 解开 `masterKeyPGP` → 拿 masterKey
2. `tenant_key = HKDF-SHA256(masterKey, deploymentSalt, tenant_id)` (32B)
3. 返回 base64
4. 错误 master_pin → 401，audit log 记录

---

## 4. MW 表查表规则

每格由 [main.go mwToken()](../kms/main.go#L219) 算出（KMS 自己不持种子、不算格子——只有持 mw.pgp/纸表的人能查）：

```
mwToken(seed, month, isoWd) = upper(base32(SHA256(seed || month_byte || isoWd_byte))[:6])
```

- `month`: 1~12 ；`isoWeekday`: 周一=1 … **周日=7**（不是 0）
- 只有 12×7 = 84 个格子，按 (月 × 周几) 定位。

**每次 recover 查两格**（运维从离线 mw.pgp / 纸表读）：

| 参数 | 查哪一格 | 用途 |
|---|---|---|
| `--mw-token`（解） | **现有 archive 文件名日期**对应的 (月×周几) | 解开盘上 archive |
| `--reseal-token`（封） | **今天**的 (月×周几) | 重封新 archive（轮换） |

> 因为 recover 总用"今日格"重封、文件名也用今日日期，所以**不变量成立**：任何 archive 的封印格 == 它文件名日期的格。下次解它,照文件名日期查表即可。
>
> **禁止把真实格子值写进任何受版本控制的文件**（含本 doc）。格子值只在离线纸表 / GPG 封的 mw.pgp 里。示例一律用占位 `<XXXXXX>`。

---

## 5. 文件清单

| 路径 | 来源 | 说明 |
|---|---|---|
| [owlBack/kms/owl-kms](../kms/owl-kms) | `go build` | 二进制 |
| [owlBack/kms/main.go](../kms/main.go) | 源码 | 单文件 |
| owlBack/kms/master_key-YYYYMMDD.json | `--recover` 每次产生 | archive，**MW 格子封**（封它的格 = 文件名日期那格），`keep=2` 滚动 |
| owlBack/kms/mw.pgp | `--init` | GPG PIN 加密的 12×7 MW 表（**git-ignored，不入库**） |
| /tmp/owl-kms.sock | `serve` 时创建 | unix socket，0660 |
| log/unseal_audit.log | `serve` 全程 | init/recover/tenant-key 审计 |
| .env | 人工 | `MASTER_PIN=...`（仅在线认证）, `KMS_SOCKET=/tmp/owl-kms.sock` |

> **已废除（2026-06-10 移除）**：`kms/.kms-secrets`（盘上存 token 供自动解封）、`scripts/systemd/owl-kms-run.sh`（自动喂 token 的 wrapper）。二者把 MW 离线因子架空，与原设计冲突，已删。`owl-kms.service` 因此**不再能自动解封**——重启后需人工跑 recover（场景 B）。

> **vault-dir 一致性**：本机 `--vault-dir owlBack/kms`（非默认 `vault/`）。recover 必须传同一 `--vault-dir owlBack/kms`，否则新 archive 散到别处、cleanupArchives 也管不到。

---

## 6. 操作场景

### 场景 A：日常 owlback.data 重启（K 仍在运行）

无需运维介入。owlback.data 启动 → 读 `.env` 拿 `MASTER_PIN` → socket 调 `/tenant-key` → 拿 tenant_key → 业务正常。

```bash
sudo systemctl restart owlback.data
```

### 场景 B：K 服务重启 / 服务器 reboot（**人工 MW 解封，无自动恢复**）

K 内存清空 → 必须人工查离线 MW 表解封。需要：
- 能读 MW 表（手头纸表，或用 **GPG PIN** 解 `mw.pgp`）
- `.env` 的 `MASTER_PIN`（仅用于封内存、让业务能认证）
- 最新一份 archive（看文件名日期）

```bash
cd /home/wisefido/owl/owlBack

# 1. 取 MW 表（若无纸表）——读完即焚
gpg --batch --yes -o /tmp/mw.md -d kms/mw.pgp && cat /tmp/mw.md && shred -u /tmp/mw.md

# 2. 查两格：
#    ① 解格 = 最新 archive 文件名日期的 (月×周几)
#    ② 封格 = 今天的 (月×周几)
LATEST=$(ls -1 kms/master_key-*.json | tail -1)   # 文件名字典序=日期序

# 3. 解封 + 重封（master_pin 只封内存）
nohup kms/owl-kms --recover --archive "$LATEST" --vault-dir kms \
  --mw-token     <①解格 XXXXXX> \
  --reseal-token <②今日格 XXXXXX> \
  --master-pin   <.env MASTER_PIN> \
  > log/kms.out 2>&1 & disown
sleep 1

# 4. 验证在线
curl -s --unix-socket /tmp/owl-kms.sock http://localhost/health   # {"status":"ok"}

# 5. 业务重连（reboot 后 owlback 已自启但 KMS 当时还没起，必须重启让它重连）
sudo systemctl restart owlback
```

> **解格错** → 步骤 3 当场 `unseal failed: message authentication failed`，不会有任何破坏，查对再来。
> **封格错（typo）** → 本次能起，但下次按"今日格"解不开新档；靠 `keep=2` 退回上一份档（用它文件名日期的格解）。所以**封格务必从离线表照抄**。
> **顺序**：先 KMS（场景 B），再 `restart owlback`。reboot 时 owlback 自启、owl-kms 不自启，业务会先起来连不上 socket。

### 场景 C：首次部署（一次性）

```bash
cd /home/wisefido/owl/owlBack/kms && go build -o owl-kms . && cd ..

./kms/owl-kms --init --pin <你定的 GPG 口令> --vault-dir kms --out-dir kms

# stdout 打印的 master_pin 写 .env（仅在线认证口令）
echo "MASTER_PIN=<打印的8字符>" >> .env
echo "KMS_SOCKET=/tmp/owl-kms.sock" >> .env

# 解出 mw.md 打印 12×7 表交付客户纸质件，读完即焚
gpg --batch --yes -o /tmp/mw.md -d kms/mw.pgp && cat /tmp/mw.md && shred -u /tmp/mw.md

# 异地备份（灾备必需）：一份 MW 封 archive + mw.pgp，并记下 archive 文件名日期
cp kms/master_key-$(date +%Y%m%d).json kms/mw.pgp /backup/异地/

# init 完自动 serve；之后每次重启走场景 B
```

### 场景 D：灾难恢复（全盘重装 / 本地 archive 全丢）

需要异地备份的（三选其备齐）：
1. 一份 **MW 封 archive** + **它的文件名日期**（决定解它用哪格）
2. **MW 表**：纸质件，或 `mw.pgp` + **GPG PIN**
3. ~~master_pin~~ **不需要**——它只是在线口令，恢复时**现取一个新值**即可

```bash
# 1. 把异地 archive 放回 owlBack/kms/，记下它的日期 D
# 2. 重建 .env（master_pin 随便定一个新的，与下面 --master-pin 一致即可）
echo "MASTER_PIN=<新定一个8字符>" >> .env
echo "KMS_SOCKET=/tmp/owl-kms.sock" >> .env
# 3. 查表：--mw-token = 日期 D 的格；--reseal-token = 今日格；跑场景 B 的 recover
```

> **致命损失边界**:archive **与** (MW 表/mw.pgp+GPG PIN) **同时全丢** = masterKey 永久无法解封 → **全部 PHI 报废**，只能新 init（旧密文成废数据）。所以异地必须同时备 archive + mw.pgp，且 GPG PIN 有可靠离线副本。
> **反之 master_pin 丢了不致命**——它不封 masterKey，重置即可。这正是设计复位后的关键改进：盘上常驻口令不再是单点解密钥匙。

---

## 7. 安全约束

- **盘上无自动解封凭证**：archive 始终 MW 格子封；master_pin 只封内存副本（在线认证），单独解不开盘上任何 archive。验证：用 master_pin 当 `--mw-token` 解档应 `message authentication failed`。
- **MW 格子值禁止落任何受控文件**（含本 doc）；只在离线纸表 / GPG 封的 mw.pgp 里。
- `.env`、`mw.pgp`、`master_key-*.json` 均 `.gitignore`，不入 git
- master_pin 永不落日志、永不落审计记录；masterKey 永不以任何形式落盘
- tenant_key 仅在业务进程内存，进程退出即清除
- unix socket 0660，仅同组可读写
- 审计日志 `log/unseal_audit.log` 记录所有 init / recover / tenant-key 调用

---

## 8. 日常验证流程

每次 KMS 重启后、定期巡检、或怀疑加解密通路异常时，跑 `verify_phi` 工具确认端到端链路活的。

### 工具

[wisefido-data/cmd/verify_phi](../wisefido-data/cmd/verify_phi/main.go) — 纯只读：DB 只 SELECT，KMS 只调 `/tenant-key` 派生，`MASTER_PIN` 仅从 env 读、不入日志。

```bash
cd owlBack/wisefido-data

# 模式 A: round-trip 自检（任选一行 first_name + first_name_enc，解密对比）
MASTER_PIN=$(awk -F= '/^MASTER_PIN=/{print $2}' ../.env) go run ./cmd/verify_phi
# 期望: "OK — END-TO-END decrypt matches plaintext"

# 模式 B: 指定 resident（按 nickname 或 UUID），dump 所有非空 PHI 字段
MASTER_PIN=$(awk -F= '/^MASTER_PIN=/{print $2}' ../.env) go run ./cmd/verify_phi Arthur.S
# 期望: "fields: N decrypted OK, 0 failed, M empty/null"
```

### 标准验证 resident

| 项 | 值 |
|---|---|
| tenant | `demo` |
| nickname | `Arthur.S` |
| resident_id | `876f6663-5dad-47b1-a204-ee8920fd35ed` |
| 标记字段 | `resident_phi.medical_history`（PHI 加密字段，写入会强制走 KMS） |
| 标记内容格式 | `KMS YYYYMMDDHHMMSS`（带时分秒，多次验证不冲突）<br>例：`KMS 20260502194930` |

### 标准验证步骤

1. **UI 写入**：登录 demo tenant → 找到 Arthur.S → 编辑 `medical_history` → 写入 `KMS <当前时分秒>` → 保存
2. **后台读出**：跑 `verify_phi Arthur.S`
3. **比对**：输出里 `medical_history = "KMS <你刚写的时间戳>"` ✓

### 验证什么

| 输出 | 验证了什么 |
|---|---|
| 工具能跑通 | KMS socket 在线 + master_pin 正确 + tenant_key 派生成功 |
| 显示真明文（非乱码） | tenant_key 解出来跟历史 init 时一致（masterKey + deploymentSalt 没变） |
| `medical_history` 等于你刚写的时间戳 | 写入路径调用了 PHICryptor 加密、读出路径用同一 tenant_key 解密成功 |
| `[YYYY-MM-DD HH:MM:SS]` 时间戳 | 工具自身的执行时间，便于回看"上次验证在何时" |

### 副作用

- KMS audit log 增加一条 `action=tenant-key tenant=<demo> detail=issued`（这是预期，不是问题）
- 不修改任何 DB / KMS 状态

---

## 9. Git 历史

KMS 代码在 owlBack 仓库**只有一个 commit**：

```
10f8e16  feat: PHI encryption — K service + AES-256-GCM full-field encryption
         Author: sady37   Date: 2026-04-17 00:26:46 -0700
         Files: kms/{go.mod, go.sum, main.go}
```

**2026-06-10 设计复位**（doRecover）：盘上 archive 改回 MW 格子封、`--mw-token`/`--reseal-token` 分离、master_pin 退出解封路径。同日迁移既有 archive + 销毁 master_pin 封残档/自动解封旁路/明文 MW 表。本文档替代丢失的 `owl/kms.md (v2)`，作为权威设计 + 操作参考。

# KMS (owl-kms) — PHI 加密密钥服务

> 本文基于源码 [owlBack/kms/main.go](../kms/main.go) + [owl-common/crypto/wrap.go](../owl-common/crypto/wrap.go) 重建。
> 原始设计文档 `owl/kms.md (v2)`（commit 10f8e16 引用）已丢失，本文替代之。

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
| **master_pin** | 8 字符随机 ([main.go:83 `randAlphanumeric(8)`](../kms/main.go#L83)) | `.env: MASTER_PIN` | **init 一次生成，永不变** | 日常运维凭证：业务进程认证、archive 加解密 |
| **MW 表** | 12 行 × 7 列，每格 6 字符 base32 | `kms/mw.pgp` (GPG 加密) + 客户纸质件 | **init 一次生成，永不变** | 灾难恢复跳板：解开 init archive |
| **GPG PIN** | 用户自定义口令 | 运维记忆 / 密码管理器 | 永久 | 解 mw.pgp |

**双因素互锁**：日常重启需 `master_pin + 任一 archive`；灾难恢复需 `MW 表 + GPG PIN + init archive 备份`。

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

**为何 init archive 用 MW token 而非 master_pin？**
让 init archive 成为"灾难恢复跳板"：即便 `.env` 丢失（master_pin 没了），只要还有 init archive + MW 纸质表 + GPG PIN，就能解出 masterKey、重建 master_pin 体系。

### 3.2 `--recover` （K 重启 / 服务器重启）

[main.go:131-151 doRecover()](../kms/main.go#L131-L151)：

```
1. masterPin: 来自 --master-pin flag 或 fallback 到 env MASTER_PIN  (不重新生成)
2. 读 archive 文件 → 反序列化
3. masterKey = UnwrapMasterKey(archive.Wrapped, mwOldToken)
   ↑ mwOldToken 必须是 archive 创建那天对应的 MW 单元格
     — init archive 用 MW 加密 → 传 init 当日 MW token
     — recover archive 用 master_pin 加密 → 传 master_pin（参数名虽叫 mw-token）
4. masterKeyPGP = WrapMasterKey(masterKey, masterPin)       ← 内存
5. newArchive   = WrapMasterKey(masterKey, masterPin)       ← 写新 archive（master_pin 包装）
6. 写 <vault-dir>/master_key-<今日>.json
7. cleanupArchives(vaultDir, keep=2)                        ← 按文件名字典序保留最新 2 份
8. zero(masterKey)
```

**为何 recover archive 用 master_pin？**
recover 之后的 archive 是"日常重启钥匙"——只要 `.env` 在，master_pin 永远不变，就能解开任何后续 archive。**MW 表只在 init archive 上有用**，recover 写出的新 archive 不再绑 MW，是因为 master_pin 已经是"运维知道的常驻口令"。

> ⚠️ **cleanupArchives 副作用**：保留最新 2 份是按文件名（=日期）字典序。第 2 次 recover 之后，**init archive 会被删除**！如果想保留 MW 灾备能力，必须在 init 后立刻把 `master_key-<init日期>.json` **异地备份**，然后允许本地 cleanup 删它也无妨。

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

[main.go:219-226 mwToken()](../kms/main.go#L219-L226)：

```
mwToken(seed, month, isoWd) = upper(base32(SHA256(seed || month_byte || isoWd_byte))[:6])
```

- `month`: 1~12
- `isoWeekday`: 周一=1, 周二=2, …, **周日=7**（不是 0）

例：archive 创建于 2026-04-17 (Friday) → 查 MW 表 `Apr × Fri` 单元格 = `FCKE7K`。

---

## 5. 文件清单

| 路径 | 来源 | 说明 |
|---|---|---|
| [owlBack/kms/owl-kms](../kms/owl-kms) | `go build` | 二进制 |
| [owlBack/kms/main.go](../kms/main.go) | 源码 | 248 行单文件 |
| owlBack/kms/master_key-20260417.json | `--init` (本机首次) | init archive (用 FCKE7K = Apr×Fri MW token 加密) |
| owlBack/kms/master_key-YYYYMMDD.json | `--recover` 后续产生 | recover archive (用 master_pin 加密) |
| owlBack/kms/mw.pgp | `--init` | GPG 加密的 12×7 MW 表 |
| /tmp/owl-kms.sock | `serve` 时创建 | unix socket，0660 |
| log/unseal_audit.log | `serve` 全程 | init/recover/tenant-key 审计 |
| .env | 人工 | `MASTER_PIN=...`, `KMS_SOCKET=/tmp/owl-kms.sock` |

> **vault-dir 一致性**：本机 init 时人工传了 `--vault-dir owlBack/kms`，所以 init archive 在 `kms/` 而不是默认的 `vault/`。recover 时务必也传 `--vault-dir owlBack/kms`，否则新 archive 散到 `vault/` 目录、cleanupArchives 也作用不到 init archive 那个目录。

---

## 6. 操作场景

### 场景 A：日常 owlback.data 重启（K 仍在运行）

无需运维介入。owlback.data 启动 → 读 `.env` 拿 `MASTER_PIN` → socket 调 `/tenant-key` → 拿 tenant_key → 业务正常。

```bash
sudo systemctl restart owlback.data
```

### 场景 B：K 服务重启 / 服务器 reboot

需要：
- `.env` 里的 `MASTER_PIN`（一直在）
- 最新一份 archive 文件
- **archive 创建那天**对应的"解锁口令"：
  - 如果还在用 init archive → 查 mw.pgp 找 init 当日 MW token
  - 如果是 recover archive → 直接用 `MASTER_PIN`（master_pin 既是运维认证、又是 archive 解锁）

```bash
# 1. (仅 init archive 时需要) 解 mw.pgp 查 token
gpg --batch --yes --output /tmp/mw.md --decrypt owlBack/kms/mw.pgp
cat /tmp/mw.md      # 找到对应 month × isoWeekday 单元格
shred -u /tmp/mw.md

# 2. 启动恢复（注意 --vault-dir 与 init 时一致）
nohup owlBack/kms/owl-kms --recover \
  --archive owlBack/kms/master_key-20260417.json \
  --vault-dir owlBack/kms \
  --mw-token <init当日MW单元格 或 .env的MASTER_PIN> \
  --master-pin <.env的MASTER_PIN> \
  > log/kms.out 2>&1 &
disown

# 3. 验证 socket 在线
curl -s --unix-socket /tmp/owl-kms.sock http://localhost/health
# {"status":"ok"}

# 4. 业务进程重连
sudo systemctl restart owlback
```

> 本机首次 reboot 后的 token = `FCKE7K` (Apr 17 = Apr × Fri)；
> 之后用 recover archive 时 token = `FtjPuGB8` (`.env` MASTER_PIN)。

### 场景 C：首次部署（一次性）

```bash
cd /home/wisefido/owl/owlBack
go build -o kms/owl-kms ./kms

./kms/owl-kms --init --pin <你定的 GPG 口令> \
  --vault-dir kms --out-dir kms

# 把 stdout 输出的 master_pin 写 .env
echo "MASTER_PIN=<打印的8字符>" >> .env
echo "KMS_SOCKET=/tmp/owl-kms.sock" >> .env

# 解出 mw.md 打印 12×7 表交付客户纸质件
gpg --batch --yes --output /tmp/mw.md --decrypt kms/mw.pgp
cat /tmp/mw.md   # 客户打印归档
shred -u /tmp/mw.md

# 把 init archive 异地备份（重要！cleanup 会在第 2 次 recover 后删它）
cp kms/master_key-$(date +%Y%m%d).json /backup/异地/

# K 进入 serve 模式（init 完会自动 serve；如已退出，用场景 B 的 recover 流程）
```

### 场景 D：灾难恢复（`.env` 丢 / 全盘重装）

需要异地备份的：
1. `kms/master_key-<init日期>.json` (init archive)
2. 客户提供的纸质 MW 表
3. 运维记忆/密码管理器里的 master_pin

```bash
# 1. 把 init archive 拿回来放回 owlBack/kms/
# 2. 重建 .env
echo "MASTER_PIN=<记忆中的8字符>" >> .env
echo "KMS_SOCKET=/tmp/owl-kms.sock" >> .env
# 3. 跑场景 B 的 recover，--mw-token 用 init 当日的 MW 单元格
```

> 如果 master_pin 也忘了——那就只能新 init 了，**所有已加密的 PHI 数据无法恢复**。所以 master_pin 必须有可靠的离线副本（密码管理器导出加密份）。

---

## 7. 安全约束

- `.env` 在 `.gitignore`，master_pin 不入 git
- mw.pgp 入 git 是可接受的（GPG 对称加密，PIN 不在仓库）
- master_pin 永不落日志、永不落审计记录
- masterKey 永不以任何形式落盘
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
         Files: kms/{go.mod, go.sum, main.go} (264 行)
```

之后无任何修改。本文档替代丢失的 `owl/kms.md (v2)`，作为权威设计 + 操作参考。

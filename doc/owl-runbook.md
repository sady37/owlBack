# Owl 全栈运维手册（停服 / 重启 / 验证 / 长期任务）

适用范围：owlBack 业务进程、sleepace（Docker MySQL + 宿主机 Java）、owlFront（Vite dev）、Docker 基础设施（PG/Redis/MQTT/MySQL）、owl-kms（PHI 加密 unix-socket 服务）。

> 单租户单机部署。所有服务在同一台 Linux 机器上，user=`wisefido`，根目录 `/home/wisefido/owl`。
>
> KMS 设计与原理详见 [kms.md](kms.md)；本文 §4 只列操作命令。

---

## 0. 系统拓扑速查

```
                ┌──────────────── owlFront (Vite :3100) ─────────┐
                │                                                 │
端口 8080 ───►  wisefido-data ──────► PostgreSQL :5432 (Docker)  │
                  │                    Redis :6379    (Docker)   │
                  │                    KMS unix:/tmp/owl-kms.sock│
端口 8081/8443 ► wisefido-qinglan ─► MQTT :1883/8883 (Docker)    │
端口 8083 ───►  wisefido-sleepace ─► Java :8090 (host) ──────► sleepace-mysql :3306 (Docker)
端口 8085 ───►  wisefido-iot ──────► Redis Streams                                
(无端口) ───►   wisefido-cardagg ──► PG / Redis                                    
(无端口) ───►   wisefido-sensor ───► Redis / PG                                    
```

关键依赖链（启停顺序的依据）：

```
owl-kms (unix socket)              ← wisefido-data 启动时连接
PostgreSQL / Redis / MQTT (Docker) ← 所有 Go 服务都依赖
sleepace-mysql (Docker)            ← Java sleepace-service 依赖
Java sleepace-service :8090        ← wisefido-sleepace :8083 依赖
wisefido-data :8080                ← cardagg 启动时同步 card 表（先起 data 等 ~10s）
```

---

## 1. 端口 / 进程 / Unit / 启停脚本对照表

| 模块 | 端口 | 进程二进制 | systemd unit | 启动脚本 | 停止脚本 |
|---|---|---|---|---|---|
| wisefido-data | 8080 | `wisefido-data/.bin/wisefido-data` | `owlback.data` | `wisefido-data/start-data.sh` | `wisefido-data/stop-data.sh` |
| wisefido-cardagg | — | `wisefido-cardagg/.bin/wisefido-cardagg` | `owlback.cardagg` | `wisefido-cardagg/start-cardagg.sh` | `wisefido-cardagg/stop-cardagg.sh` |
| wisefido-qinglan | 8081 / 8443 | `wisefido-qinglan/.bin/wisefido-qinglan` | `owlback.qinglan` | `wisefido-qinglan/start-qinglan.sh` | `wisefido-qinglan/stop-qinglan.sh` |
| wisefido-sleepace | 8083 | `wisefido-sleepace/.bin/wisefido-sleepace` | `owlback.sleepace` | (在 owlback-run-service.sh) | `wisefido-sleepace/stop-sleepace.sh` |
| wisefido-iot | 8085 | `wisefido-iot/.bin/wisefido-iot` | `owlback.iot` | (同上) | `wisefido-iot/stop-iot.sh` |
| wisefido-sensor | — | `wisefido-sensor/.bin/wisefido-sensor` | `owlback.sensor` | (同上) | `wisefido-sensor/stop-sensor.sh` |
| owlback (编排) | — | (oneshot) | `owlback` | `start-owlback-full.sh` | `stop-owlback-full.sh` |
| sleepace-service (Java) | 8090 | `java`（Tomcat 内嵌） | `sleepace` | `sleepace/sleepace.sh start` | `sleepace/sleepace.sh stop` |
| owlFront (Vite dev) | 3100 | `node`（npm run dev） | `owlfront` | `npm run dev` | `systemctl stop owlfront` |
| owl-kms | unix `/tmp/owl-kms.sock` | `owlBack/kms/owl-kms` | **无 unit（手动）** | 见 §4 | `pkill -TERM owl-kms` |
| PostgreSQL | 5432 | `owl-postgresql` 容器 | docker | `docker compose up -d postgresql` | `docker compose stop postgresql` |
| Redis | 6379 | `owl-redis` 容器 | docker | 同上（redis） | 同上 |
| MQTT (mosquitto) | 1883/8883/9001 | `owl-mqtt` 容器 | docker | 同上（mqtt） | 同上 |
| sleepace-mysql | 3306 | `sleepace-mysql` 容器 | docker | 同上 | 同上 |

> Docker 4 个业务容器都是 `restart=unless-stopped`，宿主机重启后会自动恢复。
>
> **当前 Docker 版本（2026-05-02）**：`docker 28.2.2` + 两套 compose 共存：
> - `docker compose` v2.40.3（plugin，**推荐**——官方持续维护）
> - `docker-compose` v1.29.2（独立 Python 二进制，已 EOL，留作兜底）
>
> 仓库里 `sleepace.sh` 等脚本会优先探测 plugin、失败回退到 v1，所以两种都装着没冲突。新写脚本/手敲命令统一用 `docker compose`（带空格）。

```bash
docker --version          # Docker version 28.2.2
docker compose version    # Docker Compose version 2.40.3
docker-compose --version  # docker-compose version 1.29.2 (legacy)
```

`enable` 状态（`systemctl is-enabled`）：

```
owlback              enabled   ← 开机拉起 owlback.* 全套
owlback.{data,cardagg,qinglan,iot,ai,sleepace}  static  ← 跟随 owlback
sleepace             enabled   ← 开机拉起 sleepace-mysql + Java
owlfront             enabled   ← 开机拉起 Vite dev
docker               enabled   ← 容器按 restart policy 自恢复
owl-kms              （无 unit，需人工）
```

---

## 2. 完整停服流程（从外到内）

> 目的：彻底释放端口和资源，准备做深度运维（升级 Docker、补丁内核、整机迁移等）。

### Step 1 — 停前端

```bash
sudo systemctl stop owlfront
```

### Step 2 — 停业务进程（owlback 全栈）

```bash
# 推荐：通过编排 unit 一次性停所有 owlback.* 子模块
sudo systemctl stop owlback

# 等价于 / 校验：
sudo /home/wisefido/owl/owlBack/stop-owlback-full.sh
```

`stop-owlback-full.sh` 会按反向顺序停 `ai → iot → sleepace → qinglan → cardagg → data`，再跑 `stop-owlback.sh` 做端口/进程兜底清场（lsof + ss + fuser 三重定位 PID）。

### Step 3 — 停 Java sleepace-service

```bash
sudo systemctl stop sleepace
# 等价: /home/wisefido/owl/sleepace/sleepace.sh stop
```

验证：

```bash
ss -lntp | grep :8090     # 应空
```

> ⚠️ **历史坑（2026-05-02 已修）**：`sleepace.sh stop_java_service` 原本只调 Tomcat 自带 `bin/shutdown.sh`，且末尾 `|| true` 吞失败——CATALINA_PID 文件丢失或 Tomcat shutdown port (8005) 不响应时，shutdown 静默失败但 Java **仍占着 8090**，systemctl stop 却返回成功。
>
> 现在 `stop_java_service` 加了三段式兜底：(1) Tomcat shutdown → (2) 等 10s 看 8090 是否释放 → (3) 还在就 SIGTERM 5s → 再不死 SIGKILL。所以 `systemctl stop sleepace` 现在能保证真把 Java 杀干净。
>
> 如果用的是旧版 sleepace.sh 又遇到 8090 不释放，手动兜底：
> ```bash
> pid=$(lsof -ti :8090 -sTCP:LISTEN | head -1)
> kill -TERM $pid; sleep 3; kill -0 $pid 2>/dev/null && kill -9 $pid
> ```

注意：`sleepace.sh stop` **不会**停 `sleepace-mysql` 容器（Docker 由 restart policy 管理）。

### Step 4 — 停 KMS

```bash
pkill -TERM -f owlBack/kms/owl-kms
# 或用进程号
ps -ef | grep owl-kms | grep -v grep
kill -TERM <pid>

# 验证 socket 已删除
ls /tmp/owl-kms.sock 2>&1   # 应报 No such file
```

### Step 5 — 停 Docker 容器（基础设施）

```bash
cd /home/wisefido/owl/owlBack
docker compose down               # 仅停业务容器，不删 volume
# docker compose down -v          # ⚠️ 会删数据卷，仅在确实要清空数据时用
```

### Step 6 — 校验全部停妥

```bash
bash /home/wisefido/owl/owlBack/ServiceStatus.sh
# 期望: 8080/8081/8083/8085/8090/3100/3306/5432/6379/1883 全部 [DOWN]
docker ps                        # 应空（或仅剩与 owl 无关的容器）
```

---

## 3. 服务器重启全流程

> 假设：服务器刚 `reboot` 完成，登录到 `wisefido` 账户。

### 3.1 自动恢复部分（systemd 处理）

| 组件 | 恢复方式 | 验证命令 |
|---|---|---|
| Docker daemon | `docker.service` enabled | `systemctl is-active docker` |
| 4 个业务容器 | restart=unless-stopped | `docker ps` 应见 4 个 healthy |
| owlback.* | `owlback.service` enabled (oneshot 顺序拉) | `systemctl status owlback` |
| sleepace (Java) | `sleepace.service` enabled | `systemctl status sleepace` |
| owlfront | `owlfront.service` enabled | `systemctl status owlfront` |

**但 owlback.data 启动时如未连上 KMS，加密相关代码路径会失败。**所以人工先恢复 KMS，再让 owlback 真正可用。

### 3.2 必须人工干预：KMS 恢复

KMS 故意**不 enable** auto-start，且 `--init` 会**生成新 master key**（导致已加密 PHI 数据无法解密）。重启后必须人工触发 `--recover`。

**推荐路径**（首次 reboot 之后都用这条）：

```bash
sudo systemctl start owl-kms        # 不 enable，仅手动启动
sudo systemctl status owl-kms       # 应为 active (running)
curl -s --unix-socket /tmp/owl-kms.sock http://localhost/health   # {"status":"ok"}
```

`owl-kms.service` 调用 [`owlBack/scripts/systemd/owl-kms-run.sh`](../scripts/systemd/owl-kms-run.sh)，从 `kms/.kms-secrets`（chmod 600）读 `MW_TOKEN` / `MASTER_PIN`，自动挑最新的 `master_key-YYYYMMDD.json`，硬编码 `--recover`（脚本里禁掉了 `--init`）。

> ⚠️ **已知妥协**：`ps -ef` 仍能看到 token——owl-kms 只接收 CLI 参数，无法消除。改善方向是 owl-kms 支持从 fd / env 读 token，下个迭代再做。
>
> ⚠️ **首次 reboot 例外**：如果 `.kms-secrets` 还没建（或 archive 仍是 init 那份），不要走 systemctl，照旧手敲（下面步骤）。`--mw-token` 用 init 当日 MW 单元格（本机 = `FCKE7K`，Apr×Fri；查 `mw.pgp`）。

**两种 archive 阶段**：

| 阶段 | archive | `--mw-token` |
|---|---|---|
| **首次 reboot**（archive = init 那份）| `kms/master_key-20260417.json` | `FCKE7K`（查 mw.pgp）|
| **后续 reboot**（archive = 上次 recover 写的）| 最新 `kms/master_key-YYYYMMDD.json` | `FtjPuGB8`（== MASTER_PIN）|

> 为什么两种 token 不一样？详见 [kms.md §3.2](kms.md)。简言之：
> - init archive 用 MW token 加密（灾难恢复跳板）
> - recover 写出的 archive 用 master_pin 加密（日常运维钥匙）
> - cleanupArchives 保留最新 2 份，**第 2 次 recover 之后 init archive 会被删**，必须异地备份

**首次 reboot 手敲步骤**（仅在 `.kms-secrets` 尚未建立时用）：

```bash
cd /home/wisefido/owl

# (1) 解 mw.pgp 查 token
gpg --batch --yes --output /tmp/mw.md --decrypt owlBack/kms/mw.pgp
cat /tmp/mw.md
# 找 "Apr | Fri" = "FCKE7K"
shred -u /tmp/mw.md

# (2) 启动 KMS recover
nohup ./owlBack/kms/owl-kms \
  --recover \
  --archive owlBack/kms/master_key-20260417.json \
  --vault-dir owlBack/kms \
  --mw-token FCKE7K \
  --master-pin FtjPuGB8 \
  > log/kms.out 2> log/kms.err &
disown

# (3) 验证
ls -la /tmp/owl-kms.sock           # srw-rw---- wisefido wisefido
curl -s --unix-socket /tmp/owl-kms.sock http://localhost/health
# {"status":"ok"}

# (4) 后续 reboot 准备：建 .kms-secrets（之后用 systemctl start owl-kms 即可）
umask 077
cat > owlBack/kms/.kms-secrets <<'EOF'
MW_TOKEN=FtjPuGB8
MASTER_PIN=FtjPuGB8
EOF
```

> ⚠️ **`--vault-dir` 必须传 `owlBack/kms`**——否则默认 `owlBack/vault` 会让新 archive 散到另一目录，cleanupArchives 也作用不到 init archive 那个目录。

> ⚠️ **绝不要用 `--init`**——会生成新 master key + 新 salt，导致 resident_phi 表里所有 `*_enc` 列无法解密，业务面会大面积 401 / 解密失败。wrapper 脚本会拒绝 `--init`，但直接调 owl-kms 不会。

### 3.3 KMS 起来后重拉 owlback

```bash
sudo systemctl restart owlback
# 或仅 restart 受影响模块
sudo systemctl restart owlback.data owlback.cardagg owlback.sensor
```

### 3.4 完整重启后理想顺序

```
系统 reboot
   │
   ├─► docker.service 自起 → owl-postgresql / owl-redis / owl-mqtt / sleepace-mysql / kea-* 自起
   ├─► owlback.service 自起 (oneshot)
   │     └─► start-owlback-full.sh
   │           ├─► 30s 等基础设施 ready
   │           ├─► owlback.data → 等 10s（card 表同步） → cardagg/qinglan/sleepace/iot/ai
   ├─► sleepace.service 自起 → 启 Java（端口 8090）
   ├─► owlfront.service 自起 → Vite :3100
   │
人工: sudo systemctl start owl-kms     # 见 §3.2 推荐路径
人工: sudo systemctl restart owlback   # 让 owlback 重新挂上 KMS
```

### 3.5 已知 reboot 陷阱（历史踩过的坑）

| 陷阱 | 症状 | 修复 | 是否一次性 |
|---|---|---|---|
| **Ubuntu 自带 postgresql.service 占 5432** | docker `owl-postgresql` 一直 Exited，wisefido-data 连不上 `owl_v2` DB | `sudo systemctl disable --now postgresql.service`（系统 PG 没业务数据，禁掉即可）| **永久**（已 disable）|
| **kea 容器 stale PID** | `owl-kea-dhcp6` / `owl-kea-ctrl` 卡 `Restarting` 循环，日志 `DHCP6_ALREADY_RUNNING ... PID: 1` | docker-compose 已把 `/var/run/kea` 改成 tmpfs（[commit](../docker-compose.yml)）—— PID 不再持久化 | **永久**（compose 改完）|
| **KMS recover 手敲参数易错** | 端口起来但 PHI 解密失败 / KMS 不接 socket | 用 `sudo systemctl start owl-kms`（见 §3.2）；首次 reboot 仍需手敲一次拿 init MW token | 首次手敲一次后永久 |

> 都是已修复或固化的问题。如果未来又有新 reboot 翻车点，往这张表里加一行 + 写下永久修复方式。

---

## 4. KMS 操作速查

> 完整设计与原理见 [kms.md](kms.md)。

### 4.1 文件清单

```
owlBack/kms/owl-kms                       二进制（go build）
owlBack/kms/master_key-YYYYMMDD.json      archive（init 用 MW 封；recover 用 master_pin 封）
owlBack/kms/mw.pgp                        GPG 加密的 12×7 MW 表
owlBack/kms/main.go / go.mod              源码
log/unseal_audit.log                      审计日志（init / tenant-key / recover）
.env: KMS_SOCKET / MASTER_PIN             业务进程读取
```

### 4.2 MW token 查表规则

`mwToken(seed, month, isoWeekday)` 取 `SHA256(seed || month || iso_wd)` 的 base32 前 6 字符大写。

- `month`：1~12
- `isoWeekday`：周一=1，周二=2，…，**周日=7**（不是 0）

### 4.3 验证 KMS 在响应

```bash
curl -s --unix-socket /tmp/owl-kms.sock http://localhost/health
# {"status":"ok"}
```

### 4.4 长期改进 TODO

- [ ] 写 `kms.service`（Type=forking）+ EnvironmentFile 拉 MASTER_PIN 自动 recover
- [ ] init 后立刻自动把 init archive 异地备份（避免 cleanupArchives 后丢失灾备能力）

---

## 5. 全流程验证清单

> 所有"应"的预期都基于：今天有数据、MQTT 设备在线。设备离线时 §5.4 之后的项可能空。

### 5.1 端口与进程（基础）

```bash
bash /home/wisefido/owl/owlBack/ServiceStatus.sh
```

期望 ALL UP：

```
=== OwlBack Services ===
  [UP]   wisefido-data            port:8080
  [UP]   wisefido-cardagg
  [UP]   wisefido-qinglan         port:8081
  [UP]   wisefido-sleepace        port:8083
  [UP]   wisefido-iot             port:8085
  [UP]   wisefido-sensor

=== Infrastructure ===
  [UP]   PostgreSQL  port:5432
  [UP]   Redis       port:6379
  [UP]   MQTT        port:1883
  [UP]   MySQL       port:3306

=== Sleepace vendor ===
  [UP]   sleepace-service (Java)  port:8090

=== Frontend ===
  [UP]   owlFront (Vite)          port:3100
```

### 5.2 Docker 容器健康

```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
# 应显示 4 个 owl-/sleepace- 容器，状态 "Up X (healthy)"
```

### 5.3 KMS 通路

```bash
ls -la /tmp/owl-kms.sock
curl -s --unix-socket /tmp/owl-kms.sock http://localhost/health
# 期望: {"status":"ok"}
```

### 5.4 数据库 / Redis / MQTT 三件套

```bash
# PG
PGPASSWORD=postgres psql -h 127.0.0.1 -U postgres -d owlrd -c "SELECT NOW(), count(*) FROM resident_phi;"

# Redis
redis-cli -a 'TeLunSu-36kr' ping        # PONG
redis-cli -a 'TeLunSu-36kr' XLEN iot:monitor:stream

# MQTT 订阅一会儿（看雷达/手环数据是否在推）
mosquitto_sub -h 127.0.0.1 -p 1883 -t '#' -v -W 5 2>&1 | head -20
```

### 5.5 业务接口（前后端联通）

```bash
curl -s http://127.0.0.1:8080/health
curl -s http://127.0.0.1:8080/api/v1/cards | head -c 200

curl -s http://127.0.0.1:8085/health 2>&1 | head -3

# wisefido-qinglan：HTTPS 8443 自签证书，HTTP 8081 是设备上行
curl -sk https://127.0.0.1:8443/health 2>&1 | head -3

# 前端
curl -s http://127.0.0.1:3100/ -o /dev/null -w "HTTP %{http_code}\n"
```

### 5.6 数据流端到端

```bash
# 雷达 → MQTT → iot → PG/Redis 整链
tail -50 /home/wisefido/owl/log/wisefido-iot.log | grep -iE "monitor|stream|written"
tail -50 /home/wisefido/owl/log/wisefido-qinglan.log | grep -iE "track|fall|enter"

# Sleepace 报告链
tail -50 /home/wisefido/owl/log/wisefido-sleepace.log
tail -50 /home/wisefido/owl/log/sleepace.out      # Java 端

# Sensor / RoomEngine 告警
tail -50 /home/wisefido/owl/log/wisefido-sensor.log | grep -iE "fall|alarm|pending"

# CardAgg 心跳/聚合
tail -50 /home/wisefido/owl/log/wisefido-cardagg.log | grep -iE "card|aggregat|alarm"
```

### 5.7 PHI 加解密通路

调一个会读 `resident_phi` 的接口（例如卡片详情），看响应是否含明文姓名/性别/年龄。

```bash
curl -s -H "Authorization: Bearer <token>" http://127.0.0.1:8080/api/v1/residents/<id>
```

或 `grep -i "kms\|tenant_key\|decrypt" /home/wisefido/owl/log/wisefido-data.log | tail -20`，应见 `tenant_key fetched` / 无 `unwrap failed`。

---

## 6. 故障速查

| 症状 | 第一步排查 |
|---|---|
| 8080 起不来 | `journalctl -u owlback.data -n 50` + 确认 KMS socket 在 |
| 8083 起来但无数据 | 看 sleepace-mysql 容器状态 + `tail log/sleepace.err` Java 是否报错 |
| 雷达数据停了 | `mosquitto_sub` 看 MQTT 是否在推；不在 → 设备侧；在 → wisefido-iot 日志 |
| owlback.* restart loop | 端口冲突。`ss -lntp \| grep 808x` 看是否旧进程残留，跑 `stop-owlback-full.sh` 兜底清场 |
| KMS 报 unwrap failed | mw-token 输错 / 取错单元格（注意 ISO 周日=7）/ 用了 init 的 token 解 recover archive（反之亦然） |
| Vite 起不来 | `journalctl -u owlfront`，常见 npm 缺包 → `cd owlFront && npm i` 再 `systemctl restart owlfront` |
| Docker 容器丢了数据 | `docker volume ls`，重要 volume：`owlback_postgres_data` `owlback_sleepace_mysql_data` `owlback_redis_data` |
| `owl-postgresql` Exited / 占不到 5432 | Ubuntu system postgresql 抢端口；本机已 `systemctl disable postgresql.service`，若重新出现先 `systemctl is-enabled postgresql` 确认 |
| `owl-kea-*` 卡 Restarting / 日志 `DHCP6_ALREADY_RUNNING PID:1` | stale PID。`/var/run/kea` 已改 tmpfs（compose 内），若复发先 `docker volume rm owlback_kea_run`（若存在）再 `docker compose up -d --force-recreate kea-dhcp6 kea-ctrl` |

---

## 7. PHI 数据库快速操作（测试 / 验证 KMS 通路用）

> ⚠️ 这些 SQL 直接对生产 PG 操作。请确认环境再执行；clear-all 那条不可逆。

### 7.1 看一行 resident_phi（验证 `*_enc` 列结构 / 排查解密）

```bash
PGPASSWORD=postgres psql -h 127.0.0.1 -U postgres -d owlrd -x \
  -c "SELECT * FROM resident_phi OFFSET 1 LIMIT 1"
```

`-x` = expanded 输出，按列名换行——很容易看出哪些列已加密（`*_enc`）、哪些还是明文。

### 7.2 清空所有 PHI 字段（仅测试环境！）

> 用途：演练 K 服务恢复后"重新写入加密数据"的流程；或测试场景从干净状态起步。**生产环境绝对不要跑。**

```bash
PGPASSWORD=postgres psql -h 127.0.0.1 -U postgres -d owlrd -c "
UPDATE resident_phi SET
  first_name=NULL,last_name=NULL,gender=NULL,date_of_birth=NULL,
  resident_phone=NULL,resident_email=NULL,weight_lb=NULL,height_ft=NULL,
  height_in=NULL,mobility_level=NULL,tremor_status=NULL,mobility_aid=NULL,
  adl_assistance=NULL,comm_status=NULL,has_hypertension=NULL,has_hyperlipaemia=NULL,
  has_hyperglycaemia=NULL,has_stroke_history=NULL,has_paralysis=NULL,has_alzheimer=NULL,
  medical_history=NULL,home_address_street=NULL,home_address_city=NULL,
  home_address_state=NULL,home_address_postal_code=NULL,plus_code=NULL;"
```

> 注意：列表只覆盖**明文列**。`*_enc` 二进制列同步靠业务代码下次写入时重新加密填充——若需彻底清零，再跑一遍把所有 `*_enc` 列也置 NULL。

---

## 8. 后台运维任务

> 与核心在线服务（cardagg / iot / sleepace / ai 等）解耦的运维任务集合。

### 背景

2026-04 发现 NightAbsence 批量误报，追溯后发现两个独立问题：
1. **Sleepace 时区 / 报告时间下发路径**：已修——改用 per-device alarm_params（随 monitor settings save 自动下发），DST 由 IANA 库在调用时换算。
2. **NightAbsence 判定逻辑**：仍待修——还在查有 bug 的 sleepace_report 表（见 §8.2）。

架构结论：
- **厂家按设备本地时钟**触发 `reportUploadTime`（不是服务器 OS 时钟）
- 设备本地时钟 = bind 时传的 `timezone` 秒数；**厂家不做 DST**，所以我们每次 save 用 IANA 当前 offset 重算
- Sleepace 服务器的 `config.properties.timeZone=-25200` 是 channel 默认（仅影响 report summary 字段），与设备行为无关

### 8.1 Sleepace 时区 / 报告时间（已完成）

**数据模型**：

```
alarm_cloud.metadata.tenant_sleepreport_time   (int 1-24)  ← tenant 默认
alarm_cloud.metadata.TenantResetTime           (ResetTime/NapTime)
alarm_device.monitor_config.items[SleepadSetting].alarm_params:
    timezone           (IANA 字符串，"" 表示跟随 unit.timezone)
    sleep_report_time  (int 1-24)
```

**save 流程**（`device_monitor_settings_service.pushDeviceSettings` SleepadSetting 分支）：

1. `normalizeSleepadTimezone`：若 `alarm_params.timezone == unit.timezone` → 清空（让 unit.tz 变更自动跟随）
2. `BindDevice(tzSec, gender, age)` 下发 `/sleepace/bind`：
   - tzSec = `IANAToOffsetSeconds(timezone)`，IANA 库按 call-time 处理 DST
   - gender/age = 查 resident_phi 解密（无 resident → 1 / 65；age 向上取整 5 倍数，最小 60）
3. `SetReportUploadTime(hour)` 下发 `/sleepace/setReportUploadTime`

**DST 触发入口**（外部脚本调，service 内部自行查当前 effective 值下发）：

- `POST /internal/sleepace/device/{device_addr}/resync-timezone`
- `POST /internal/sleepace/device/{device_addr}/resync-report-time`

典型用法：DST 切换日 cron 遍历所有 Sleepad 设备各调一次这两个端点即可。

### 8.2 NightAbsence 逻辑改造（P1，待做）

依赖 sleepace_report 对齐，但当前 cardagg 查询层有两处 bug，必须先修。

**8.2.1 sleepace_report 查询 bug（cardagg）**

[alarm_service.go:1051-1083](../wisefido-cardagg/internal/service/alarm_service.go#L1051-L1083)：

- `sleepace_report.date` 列是 INT（`20260422` 形式），但 cardagg 查询用 `nightStart.Format("20060102")` 字符串 → 类型不匹配
- `date` 写入时是 **UTC**（`timeToDate()` 用 `.UTC()`），cardagg 查询用 **Denver 本地日期** → 时区错位
- 修：改 int + 改用 UTC 日期换算

**8.2.2 改用实时流为主、报告为辅**

- 主判据：`iot_timeseries` 在 `nightStart~nightEnd` 窗口内有 `bed_status=0` 的 monitor track（或 HR>0 / RR>0）
- 兜底判据：`alarm_events` 同窗口有 `InBed` 事件
- `sleepace_report` 仅作交叉验证（报告存在且 `sleep_state>=1` → 有人）
- 三者 OR = 有人；三者全无 → NightAbsence

**8.2.3 Sleepad 灵敏度异常独立事件（P2）**

如 202 Room 整夜只有 ~2 分钟在床信号——不是"整夜未归"而是传感器问题。加新事件 `DeviceInsensitivity`（或并入 `DeviceFailure`）：

- 夜间窗口 bed_status=0 累计 < N 分钟（默认 30）
- 但 monitor 在线
- → 灵敏度/摆放问题

**8.2.4 ResetTime 两层 resolver（仍需做）**

当前 `cardagg/alarm_service.go:512 getResetTimeParamsForTenant` 只读 `alarm_cloud.metadata`；`alarm_device.metadata.ResetTime` 没人读也没人写。要补：

- wisefido-data 补 device-level ResetTime CRUD + 初建从 tenant 拷贝
- owl-common 加 `EffectiveResetTime(tenantID, deviceID)` resolver（device > tenant）
- cardagg 切换为 consumer

### 8.3 健康快照（P2，未实现）

每日一份 `/home/wisefido/owl/log/maintenance-YYYYMMDD.json`：

- Sleepad 灵敏度异常候选（像 202 Room）
- NightAbsence 报警日/周趋势
- sleepace_report 最新 createTime 漂移（> N 小时未上传的设备）
- Sleepace analysis MQ 事件延迟

### 8.4 备份（未实现，占位）

**Target 待定：**

- 本机冷备 `/var/backups/owl/` 还是推 S3/NAS
- 保留策略（日备 14 天 / 周备 8 周）
- 加密层

**范围待定：**

- `owlrd` PG（核心）
- `sleepace_tb_data` MySQL（历史报告，厂家也存，可选）
- `sleepace_pro_user` MySQL（设备映射）
- Redis AOF（cardStatus 可重建，可选）
- `/home/wisefido/owl/log/*.log`
- 各服务 `config.yaml`
- **owlBack/kms/master_key-<init日期>.json**（init archive，cleanupArchives 之后会删）

### 8.5 日志轮转（未实现，占位）

- 归档 `/home/wisefido/owl/log/*.log.YYYYMMDD`，压缩
- 14 天解压 + 90 天压缩 + 之后删除
- 关键错误 tail 到 `maintenance-YYYYMM.log`

### 8.6 DST 自动化（建议简化版）

不需要常驻 Python scheduler，只需 **2 个 DST 切换日**的 cron：

```cron
# 每年 3 月第二个周日 + 11 月第一个周日 03:30 local 触发
# （或直接每天凌晨 03:30 跑一次——幂等，每日重复无副作用）
30 3 * * *  /usr/local/bin/owl-dst-resync.sh
```

脚本伪代码：

```bash
#!/bin/bash
# 遍历所有 Sleepad 设备，各调一次 resync endpoint。
# /internal/* 跳过 auth，无需 token。
for device_addr in $(psql -At -c "SELECT host(d.device_addr) FROM devices d JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid WHERE dfm.device_type='Sleepad' AND dfm.device_code<>''"); do
  curl -X POST "http://127.0.0.1:8080/internal/sleepace/device/$device_addr/resync-timezone"
done
```

**新装设备不需要这个脚本**——device_store 创建时 `DeviceStoreService.InitializeDevice` 静默自动调 bind + monitor save 路径都会推 timezone。

---

## 9. 实施优先级

| P | 任务 | 状态 |
|---|---|---|
| ~~P0~~ | Sleepace 时区 / 报告时间下发 | ✅ 已完成（§8.1） |
| **P1** | 8.2.1 cardagg sleepace_report 查询 bug | 待做 |
| **P1** | 8.2.2 NightAbsence 改用实时流 | 待做（依赖 8.2.1） |
| P2 | 8.2.3 灵敏度异常事件 | 待做 |
| P2 | 8.2.4 ResetTime device-level resolver | 待做 |
| P2 | 4.4 KMS systemd unit + init archive 异地备份 | 待做 |
| P2 | 8.3 健康快照 | 待做 |
| P3 | 8.4 备份 | 明确 target 后 |
| P3 | 8.5 日志轮转 | 空间告急前 |

---

## 10. 一页速查

```bash
# 状态
bash /home/wisefido/owl/owlBack/ServiceStatus.sh

# 全停 / 全起
sudo systemctl stop  owlback owlfront sleepace
sudo systemctl start owlback owlfront sleepace

# 单模块重启
sudo systemctl restart owlback.data        # 或 cardagg/qinglan/sleepace/iot/ai

# KMS 恢复（重启后必做）
#   首次 reboot:    archive=master_key-20260417.json  --mw-token=FCKE7K (查 mw.pgp 得)
#   后续 reboot:    archive=master_key-<最新>.json    --mw-token=FtjPuGB8 (即 .env MASTER_PIN)
gpg --batch --yes --output /tmp/mw.md --decrypt /home/wisefido/owl/owlBack/kms/mw.pgp
cat /tmp/mw.md && shred -u /tmp/mw.md
nohup /home/wisefido/owl/owlBack/kms/owl-kms \
  --recover \
  --archive /home/wisefido/owl/owlBack/kms/master_key-20260417.json \
  --vault-dir /home/wisefido/owl/owlBack/kms \
  --mw-token <FCKE7K 或 FtjPuGB8> --master-pin FtjPuGB8 \
  > /home/wisefido/owl/log/kms.out 2>&1 &
disown
sudo systemctl restart owlback   # 让 data/cardagg/ai 重连 KMS

# DST 切换日
/usr/local/bin/owl-dst-resync.sh   # 遍历 Sleepad 设备 resync timezone

# 日志
journalctl -u owlback.data -f
tail -f /home/wisefido/owl/log/wisefido-*.log
```

---

## 11. 开放问题

1. 多 branch 全球化后：DST resync cron 用 UTC 触发还是各 branch 本地？当前单 branch（Denver）简单直接
2. `setReportUploadTime` 冬/夏切换当天的报告覆盖不足 24h（-1h）或重叠（+1h），UI 如何展示？（低优先，厂家行为）
3. KMS init archive 异地备份策略（手动 vs 自动 push 到 S3/NAS）

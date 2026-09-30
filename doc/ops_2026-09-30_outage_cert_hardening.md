# 2026-09-30 运维变更记录：nginx 宕机根因、证书体系改造、容器与安全加固

> 变更全部落在 `/etc`、`/root` 与 Docker 运行时，不在本仓库内，故以本文档留痕。
> 每项均附回滚路径与备份位置。

## 0. 起因

`pilot.wisefido.com` / `www.wisefido.com` 等全部站点无法访问，ping 通但 HTTP 超时。
排查发现 **nginx 自 2026-09-25 15:49:01 起已停止 4 天**，80/443 无监听。

---

## 1. nginx 宕机根因：acme.sh 为一张孤儿证书停机且未起回

### 根因链

`ota.wisefido.work`（注意是 `.work`，与 certbot 管的 `.com` 无关）用 acme.sh 的
**tls-alpn** 验证（需独占 443），配了停机钩子：

| 钩子 | 内容 |
|---|---|
| `Le_PreHook` | `systemctl stop nginx` |
| `Le_PostHook` | `systemctl start nginx` |
| `Le_ReloadCmd` | `systemctl reload nginx` ← 致命 |

1. `15:49:01` cron 触发 acme.sh，PreHook 停 nginx 腾出 443
2. `15:49:08` 证书签发**成功**
3. install-cert 阶段执行 ReloadCmd → `nginx.service is not active, cannot reload`
4. acme.sh 判 `Install cert retry failed` / `Error renewing` → **走 error 分支退出，
   PostHook 的 `start` 从未执行** → nginx 保持 down
5. `Le_InstallCertSuccessTime` 因此不更新，之后每天/每 6 小时重试 → reload 再失败，
   死循环刷了 4 天错误日志，**零告警**

**reload 一个自己刚亲手停掉的服务** —— 这个自相矛盾是全部病根。

### 二次伤害：webroot 证书被连累

nginx 停机期间，三张 webroot 证书进入续期窗口后全部失败（journal 实录）：

```
Sep 29 05:10  Failed to renew certificate ota.wisefido.com  (http-01 challenge failed)
Sep 29 18:33  Failed to renew certificate demo.wisefido.com
Sep 29 18:33  Failed to renew certificate wisefido.com
Sep 29 18:33  3 renew failure(s), 0 parse failure(s)
```

webroot 依赖 nginx 活着提供 `/.well-known/acme-challenge/`。若未及时发现，
10-29 三张证书集体过期，会是比 502 严重得多的二次故障。

### 处置

追查该证书服务于谁，发现是**彻底的孤儿**，四条证据：

- `nginx -T` 的 `ssl_certificate` 全集中无 `/etc/nginx/ssl/ota_wisefido/`
- 无 `server_name ota.wisefido.work` 的 vhost
- 实测 SNI 访问该域名返回的是 `CN=app.wisefido.com`（default server 兜底）
- `/opt` + `/etc/systemd` + `/home/wisefido` 全盘零引用

真正的 OTA 容器 `ota-ql` 挂载 `/opt/ota-ql/certs`，内含 **`ota.wisefido.com`**
（certbot 签发），由 root cron `OTA-QL-CERT-SYNC` 同步，**与 `.work` 毫无关系**。

即：为一张没有任何进程读过的证书配了停 nginx 的钩子，搞挂全站 4 天。

**已执行**（备份 `/root/acme-work-cleanup-backup-20260930071750/`）：

- `acme.sh --remove --domain ota.wisefido.work --ecc`
- 删 root crontab 中 acme.sh 行（**保留了必须留的 `OTA-QL-CERT-SYNC`**）
- 删 `/etc/cron.d/ota-wisefido-acme-renew`
- 删 `/root/.acme.sh/ota.wisefido.work_ecc`、`/etc/nginx/ssl/ota_wisefido`

acme.sh 现零证书，全站停机钩子归零。域名 `wisefido.work`（注册商 Dynadot，
DNS 非 Cloudflare）于 2026-09-30 18:42 UTC 到期，经确认零使用，任其过期。

**附加兜底**：`/etc/systemd/system/nginx.service.d/override.conf`
加 `Restart=on-failure` / `RestartSec=5s`。注意**拦不住主动 `systemctl stop`**，
那一层靠上面的钩子清理解决。

---

## 2. 四张证书统一切换到 dns-cloudflare

### 动机

斩断「证书续期 ↔ nginx / 80 端口」的级联依赖（见 §1 二次伤害）。
dns-01 不需要 nginx 活着，也不需要 80 端口可达。

### 变更

| 证书 | 之前 | 之后 |
|---|---|---|
| `app.wisefido.com`（含 pilot/test） | dns-cloudflare | 不变 |
| `ota.wisefido.com` | webroot | **dns-cloudflare** |
| `wisefido.com`（含 www） | webroot | **dns-cloudflare** |
| `demo.wisefido.com` | webroot | **dns-cloudflare** |

前提：`wisefido.com` 权威 NS 在 Cloudflare（`craig/adele.ns.cloudflare.com`）。

### 做法（零配额、零中断）

直接改 `/etc/letsencrypt/renewal/<name>.conf` 的 `[renewalparams]`：

```ini
authenticator = dns-cloudflare
dns_cloudflare_propagation_seconds = 30
dns_cloudflare_credentials = /etc/letsencrypt/cloudflare.ini
```

并删 `webroot_path` 行与 `[[webroot_map]]` 段。
**不要用 `--force-renewal` 重签**（消耗 LE 配额且触发 deploy hook 重启服务）。

注意：`ota` 那张的 `[[webroot_map]]` 是空段删一行即可；`wisefido.com` / `demo`
两张段内有映射行，须 `sed '/^\[\[webroot_map\]\]/,$d'` 连段删到文件末尾。

删 `[[webroot_map]]` **不会丢 SAN 域名** —— certbot 从证书本身读域名列表，
该段只是 webroot 插件的路径映射。dry-run 输出
`Simulating renewal ... for wisefido.com and www.wisefido.com` 已证。

验证标志：`certbot renew --cert-name <name> --dry-run` 输出
`Waiting 30 seconds for DNS changes to propagate` 即证明真走了 DNS-01。

**切换不影响下游**：deploy hooks 与 ota-ql 的证书同步 cron 都按
`/etc/letsencrypt/live/<name>/` 路径工作，验证方式变了路径不变。
nginx 中的 `/.well-known/acme-challenge/` location 切后成死代码，
**故意保留作回滚余地**。

回滚：各 conf 的 `.bak.*` 备份。

### 补做续期

跑了一次 `certbot renew`，补上 9-29 失败的三张。到期从 10-29 推到 **12-29**。
全程走 dns-01 未触碰 nginx（nginx 启动时间未变）。

---

## 3. certbot deploy hooks 修复

### 问题

`certbot` 对**每张续期的证书各调用一次** deploy-hook（探针实测确认），
而原 hooks 无任何 lineage 过滤，导致：

1. **无差别重启**：demo / ota / www 这些与后端无关的证书续期，
   也会把 `owlback` + `owlback.qinglan` + `owl-mqtt` 全重启。
   9-30 那次续了 3 张证书 → 三个服务各被重启 **3 遍**。
2. **`fix-permissions.sh` 是空壳**：它操作的
   `/etc/letsencrypt/archive/test.wisefido.com/` 根本不存在（`test` 只是 app
   证书的一个 SAN，非独立 lineage，certbot 不会为它建 archive 目录）。
   5 条 `chgrp`/`chmod` 全部失败，而脚本**没有 `set -e`**，最后那行
   `restart owlback.qinglan` 照跑 → 本职（修权限）完全失效，只剩多余的重启。
3. **nginx 被 reload 三遍**：`10-nginx-reload`、`reload-nginx.sh`、
   `ota-ql-ota.wisefido.com.sh` 三个脚本干同一件事。

### 变更（5 个脚本 → 2 个）

| 脚本 | 处置 |
|---|---|
| `10-nginx-reload` | 保留 —— 所有证书都该 reload nginx，无需过滤 |
| `renew_owlback.sh` | **重写** —— 加 lineage 过滤 + 接管 qinglan 重启 |
| ~~`reload-nginx.sh`~~ | 删（重复） |
| ~~`ota-ql-ota.wisefido.com.sh`~~ | 删（重复） |
| ~~`fix-permissions.sh`~~ | 删（死代码，功能并入 `renew_owlback.sh`） |

过滤逻辑：

```bash
LE_DIR="/etc/letsencrypt/live/app.wisefido.com"
if [ "${RENEWED_LINEAGE:-}" != "$LE_DIR" ]; then
  exit 0
fi
```

只有 `app.wisefido.com`（SAN = app + pilot + test，这三个服务实际加载的证书）
续期时才分发证书并重启服务。

### 验证方式

1. `bash -n` 语法检查
2. mock 掉 `systemctl`/`docker` 等副作用命令，验证三种分支：
   demo 续期 → 完全跳过；app 续期 → 重启 owlback + qinglan；
   无 `RENEWED_LINEAGE`（手动误跑）→ 安全跳过
3. **探针实测 `RENEWED_LINEAGE` 的真实值**（最关键，否则格式不符会导致
   11-26 app 续期后服务不重启、加载不到新证书，且要到那时才暴露）：
   临时移开真脚本、放入只记录的探针，跑 `certbot renew --dry-run --run-deploy-hooks`。
   结果 `/etc/letsencrypt/live/<certname>`，**无尾斜杠，与脚本一致**。

注：`--run-deploy-hooks` 单独用时证书未到期不触发，须配合 `--dry-run`。

回滚：`/root/renewal-hooks-backup-20260930090952/`

---

## 4. ota-ql 容器 PID 僵尸耗尽

### 根因（经典 Docker PID 1 问题）

容器 `Up 2 months (unhealthy)`，但**不是服务坏了**，是容器内 PID 耗尽导致
`fork` 失败 —— healthcheck 自身和 `docker exec` 都报
`Resource temporarily unavailable`。

镜像 `ghcr.io/hhtbing/ota-ql:latest` 的 entrypoint `./server` 作为 PID 1
**不调用 `wait()` 回收孤儿子进程**。镜像自带 healthcheck 每 30s 跑
`wget -q --spider --no-check-certificate https://localhost:10088/api/health`,
busybox 的 wget 处理 **https** 时 fork 出 `ssl_client` 辅助进程，
wget 退出后它成孤儿 reparent 到 PID 1 → 无人回收 → 僵尸每周期堆一个。

实测 `pids.current=18303` / `pids.max=18304`，`ps` 确认 **18292 个 `Z` 状态的
`ssl_client`**，父进程全是容器 PID 1。**健康检查亲手杀死了它要检查的容器。**

速率：最老僵尸约 59.6 天前 → 跑满约 **60 天**必然复发，重启只是缓解。
注意 `docker ps` 的 "Up 2 months" 是运行时长，不是 unhealthy 时长。

### 变更

以原参数重建容器，**唯一增加 `--init`**（Docker 用 tini 作 PID 1 自动 reap）。

验证：PID 1 由 `server` 变 `docker-init`，`server` 降为 PID 7，状态 healthy，
僵尸 18292 → 0（跨 5 个 healthcheck 周期持续为 0，对照旧容器每周期必留一个），
六端口（1060/10086/10088/10089/31883/38883）全恢复，`ota.wisefido.com` 200。

**重建要点**：容器是手工 `docker run` 创建（无 compose 标签），参数须从
`docker inspect` 完整提取。坑 = 镜像自带 Env 只有 `PATH`/`TZ`，那 5 个 `OTA_*`
全是 run 时传入必须原样带上；`--health-start-period 10s` 也是 run 时加的。
数据全在 bind mount（`/opt/ota-ql/{data,certs:ro,logs,firmware}`），重建不丢。

回滚：旧容器改名 `ota-ql-broken-20260930` 保留未删。

---

## 5. 安全加固：pilot 的 vite dev server 暴露面

### 问题

`pilot.wisefido.com` 的前端以 **vite dev server 直接对公网服务**
（`owlfront.service`，`ExecStart=/usr/bin/npm run dev`，`NODE_ENV=development`）。

**住户隐私数据可匿名下载（已封堵）**：`/roomengine/*` 可匿名列出并下载
playback 回放 —— 住户轨迹 / 姿态 / 生命体征，7 个文件共 **246MB**，涉 3 台设备。

链路：`公网 → nginx(只转发) → vite dev server(只转发) → vite proxy →
roomengine-api(127.0.0.1:7788，dev 工具自身无认证)`，**三环无一鉴权**。

`vite.config.ts` 原注释直说了动机：
`用 /roomengine/* 前缀避开 owlBack 的 /api/* auth middleware`。
本地开发合理（该 API 只绑 127.0.0.1），错在这份 dev 配置被原样搬到公网，
vite proxy 成了穿透内网的桥。

补充：**前端登录/锁屏只是浏览器里的路由守卫**，管"页面显不显示"，
`curl` 直接请求根本不经过它。**前端登录 ≠ 后端鉴权。**

### 变更

pilot vhost 在 SPA 兜底 `location /` **之前**插入拦截块：

```nginx
location ^~ /roomengine/ { return 404; }
location ^~ /@fs/        { return 404; }
location = /vite.config.ts    { return 404; }
location = /package.json      { return 404; }
location = /package-lock.json { return 404; }
location = /tsconfig.json     { return 404; }
```

验证：泄露路径全 404；`/` + `/src/main.ts` + `/login` 仍 200（应用正常）；
`127.0.0.1:7788` 本地照常（开发者走 SSH 隧道不受影响）。

回滚：`/etc/nginx/sites-available/pilot.wisefido.com.bak.20260930094402`

### 已核实未泄露

vite 的 `server.fs.allow` 限死在 `/home/wisefido/owl/owlFront`：
`owlBack/.env` 与 `/etc/passwd` 均返回 `403 Restricted`。
**owlBack Go 源码、wisefido-sensor 算法、任何 `.env` 与 DB 凭据都没暴露**；
owlFront 下无 `.env` 文件；扫 `src/` 无真实硬编码凭据
（唯一命中是 mock 假密码 `mock-admin-pass`）。

### 排查中两次差点误报（教训）

- **`200 + 412B` 是 SPA fallback 的 `index.html`，不是文件泄露。**
  `/.env`、`/@fs/etc/letsencrypt/cloudflare.ini` 都属此类 ——
  **判泄露必须看响应体开头，不能只看状态码。**
- grep `3100` 会匹配 `33100` 子串，一度误判 demo 也连着 vite。
  实际 **pilot 是唯一通向 `:3100` 的 vhost**；
  demo → `:33100`(owl-monitor-mock)、www/wisefido.com → `:8680`(owl-website)。

---

## 6. 监控：owl-watchdog 黑盒探活 + 自愈

### 动机

本日四次故障（nginx 停机 4 天、certbot 连挂 3 张证书、acme.sh 死循环、
ota-ql PID 耗尽）**全部零告警**，靠人工发现网页打不开。

关键认识：**问题不是"检测不到"，是"有信号但没人看"** ——

```
systemd 知道  nginx.service inactive
systemd 知道  certbot.service Failed with result 'exit-code'   ← 连报 4 天
docker  知道  ota-ql (unhealthy)
acme.sh 写了  /var/log/ota-wisefido-acme-renew.log 几十条 Error
```

信号齐全，缺的是送达。

### 部署

| | |
|---|---|
| 脚本 | `/usr/local/bin/owl-watchdog.sh`（版本管理副本见 `scripts/owl-watchdog.sh`） |
| 调度 | root cron `*/2 * * * *`，`flock -n /run/owl-watchdog.lock` 防重叠 |
| 日志 | `/var/log/owl-watchdog.log`（超 5MB 自截半） |
| 状态 | `/var/lib/owl-watchdog/*.state` |

覆盖 **16 项**：4 站点 HTTP 探活（pilot/www/ota/demo）+ 4 systemd 单元
（nginx/owlback/owlback.qinglan/owlfront）+ 4 容器（ota-ql/mqtt/redis/postgresql）
+ 4 证书到期（<14 天告警）。

### 四个设计决定（均由本日故障反推）

1. **自愈优先于告警** —— 凌晨三点没人看告警。探活失败先拉起 nginx。
   但 `nginx -t` 先校验，**配置坏则拒绝自动拉起转人工**，避免把坏配置反复推上线。
   09-25 那次若有此机制，4 天可缩到一个探测周期。
2. **状态机只在翻转时记一条** —— acme.sh 死循环 4 天里失败几十次，
   每次都记会淹没真信号。实测稳定态日志零增长。
3. **自愈节流**（10min 内最多 2 次）—— 防对着起不来的服务无限重启。
4. **容器 `running` 但 `unhealthy` 只告警不重启** —— 专门针对 ota-ql 那类：
   进程在、端口通，但健康检查持续失败。盲目重启可能丢状态，需人工判断。

`notify()` 是预留的通知出口，当前仅落日志。以后接钉钉/Telegram/webhook
**只改这一个函数**，16 项检查逻辑一行不用动。

### 验证（真停一次 nginx）

```
[CRIT] http:pilot   故障: HTTP 000
[HEAL] nginx        检测到 inactive(由 pilot 触发)，已自动拉起
[OK  ] http:pilot   已恢复: HTTP 200
```

完整故障-自愈-恢复闭环，稳定后不再新增。

### 排查中修掉的一个 bug

初版漏了 mqtt/redis/postgresql 三个容器——它们的实际名带 compose 前缀哈希
（`a2437c59f1e9_owl-mqtt`），匹配写成了 `_mqtt$` 而实际是 `_owl-mqtt`。
靠核对状态文件清单发现（只生成 1 个容器状态而非 4 个）。
**若只看"脚本跑通、退出码 0"就上线，这三个容器会永远处于监控盲区而毫无迹象。**

---

## 7. certbot.timer：恢复并固定触发时段

### 问题

早前为避开续期时的服务重启临时 `stop` 了 timer，但它仍是 `enabled`，
机器重启会自己回来 —— 属半停不停、行为不一致的状态。

发行版默认 `OnCalendar=*-*-* 00,12:00:00` 且 **`RandomizedDelaySec=43200`
（12 小时随机）**，实际触发点完全随机（09-29 那两次落在 05:10 与 18:33）。

续期本身无害（走 dns-01 不碰 nginx），但 `app.wisefido.com` 那张证书续期会经
deploy hook 重启 owlback / owlback.qinglan / owl-mqtt（该证书是这三个服务实际
加载的），随机时刻重启对 7x24 监护业务不可控。

### 变更

`/etc/systemd/system/certbot.timer.d/override.conf`（drop-in，certbot 升级不丢）：

```ini
[Timer]
OnCalendar=
OnCalendar=*-*-* 04:00:00
RandomizedDelaySec=30m
```

**空 `OnCalendar=` 是必须的** —— systemd 中该指令可累加，不先清空会变成
"原随机时间 + 04:00" 两组都触发。

现状：`active` + `enabled`，下次触发 10-01 04:25。

### 效果

| 时间 | 行为 |
|---|---|
| 每天 04:00±30min | 检查续期；证书未到期则秒退，无副作用 |
| **~10-26** | app 证书续期 → **会重启 owlback/qinglan/mqtt**，但落在凌晨 4 点 |
| ~11-28 | 另三张续期 → 仅 reload nginx，**不重启服务**（§3 lineage 过滤之功） |

### 一个非预期副作用

`Persistent=true` 使 timer 一启动就判定"错过了今天 04:00"，**立即补跑一次**。
本次无害（证书均未到期，`Finished certbot.service` 秒退，服务 uptime 未变），
但若启动 timer 时恰有证书处于续期窗口内，**它会立刻续期并触发服务重启**，
不会等到凌晨。

### 可选优化（未做）

`owlback` 代码中无 SIGHUP 处理，**不支持热重载**，必须重启才能加载新证书。
但 **mosquitto 支持 SIGHUP 重载证书**，可把 hook 里的 `docker restart` 改为
`docker kill -s HUP`，**设备不会断连重连**，把重启面从三个缩到两个。

---

## 8. 遗留待办

| 项 | 说明 |
|---|---|
| owlFront 源码仍公网可下载 | `/src/*`、`/node_modules/*`。nginx 拦不了 —— dev 模式下 `index.html` 直接 `<script src="/src/main.ts">`，封了站点就打不开。根治 = `npm run build` + nginx serve `dist/` + 停 `owlfront.service`。**迁移工作量比初估小得多**：nginx 已有 11 条 API location（`/auth/api/`、`/admin/api/`、`/data/`、`/radar-device/`、`/qinglan/`、`/ota/`、`/settings/api/`、`/device/api/`、`/sleepace/api/`、`/api/`、`/internal/`），已覆盖前端源码实际用到的绝大部分前缀，仅需核实 `/sleepad/api`、`/sleepace` 等少数边界是否对齐。风险中等：前端代码本就下发浏览器，dev 模式只是未压缩、带注释、原文件名，降低攻击门槛但无直接凭据。 |
| 监控 Layer 2：systemd `OnFailure` | 零成本，用原生机制覆盖定时任务类失败（certbot 那次 systemd 早知道但没人接）。写一个 `notify@.service` 模板，各 unit 加一行 `OnFailure=notify@%n.service` 即可。当前 watchdog 不查 `certbot.service` 退出码。 |
| 监控：反向心跳 | watchdog 自己挂了无人知晓。解法是 dead man's switch——定期向外部上报"我还活着"，超时未收到才告警（如 healthchecks.io 或自建）。 |
| mqtt 改 SIGHUP | 见 §7 可选优化，可消除证书续期时的设备断连重连。 |
| `ota.wisefido.work` DNS 记录 | 托管在 Dynadot（非 Cloudflare，现有凭据够不着）。域名 9-30 过期后自然消失，无需处理。 |

---

## 附：备份与回滚索引

| 变更 | 备份位置 |
|---|---|
| acme.sh 孤儿证书清理 | `/root/acme-work-cleanup-backup-20260930071750/`（含 crontab 原文） |
| certbot renewal conf | `/etc/letsencrypt/renewal/*.conf.bak.*` |
| deploy hooks | `/root/renewal-hooks-backup-20260930090952/` |
| pilot vhost | `/etc/nginx/sites-available/pilot.wisefido.com.bak.20260930094402` |
| ota-ql 旧容器 | Docker 容器 `ota-ql-broken-20260930`（已停，未删） |
| owl-watchdog | 新增无需回滚；停用 = 删 root crontab 中 `# OWL-WATCHDOG` 行 |
| certbot.timer 时段 | 新增 drop-in；回滚 = 删 `/etc/systemd/system/certbot.timer.d/override.conf` 后 `daemon-reload`（恢复发行版默认的 12h 随机） |

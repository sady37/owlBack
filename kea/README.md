# owl_v2 kea + BIND infrastructure

owl_v2 IPv6 寻址体系运行时支撑：
- **kea-dhcp6** + **kea-ctrl-agent**：IPAM，prefix delegation REST API
- **BIND9**：DNS 短名解析 + ip6.arpa 反向

## 角色定位

owl 设备**不直接**走 DHCPv6 协议（设备是 IPv4 + MQTT，owl 地址是逻辑标识）。kea 在 owl 体系里是 **IPAM allocator**：

```
owl Go service (wisefido-data)
     ↓ POST /command (kea REST API)
kea-ctrl-agent (port 8000)
     ↓ unix socket
kea-dhcp6
     ↓ DHCPv6-PD 协议算法
返回分配的 prefix
     ↓
owl 写入 owl_v2.tenants/branches/units/...
     ↓ DDNS update
BIND9 zone files
```

## 数据库策略：仅 owl_v2

旧库 owlrd 不再启用。docker-compose 配置 `POSTGRES_DB=owl_v2` + `initdb.d=../owlRD/dbv2`，**首次启动自动初始化 dbv2 schema**。

- 历史 owlrd 数据全 dump 已存档：`/home/wisefido/owl/backup1.0/db-rebuild-20260508-080819/full.sql`
- 配置 v1 已存档：`/home/wisefido/owl/backup1.0/owlBack-config-v1/`
- 真要回退，restore from `full.sql` 单独重建 owlrd 即可

**首次启动注意**：如果 `postgres_data` volume 已有 owlrd 旧数据，需先 `docker volume rm` 才能让 init 跑。

## 备份位置

```
/home/wisefido/owl/backup1.0/
├── db-rebuild-20260508-080819/        full owlrd dump (2.5 GB)
└── owlBack-config-v1/
    ├── docker-compose.yml._v1          (kea+BIND 加入前)
    ├── .env._v1                        (DB_NAME=owlrd 时刻)
    └── env.example._v1
```

## 环境变量约定（.env）

```
# DB
DB_NAME=owl_v2
DB_HOST / PORT / USER / PASSWORD

# kea REST API
KEA_API_HOST / PORT / USER / PASSWORD

# DDNS TSIG（kea-ddns + bind + owl service nsupdate 共用）
DDNS_TSIG_NAME / ALGORITHM / SECRET

# BIND DNS
BIND_HOST / PORT
```

## bring-up 顺序（首次部署 owl_v2）

```bash
cd /home/wisefido/owl/owlBack

# 1. 检查并清理旧 volume（如果存在 owlrd 旧数据）
docker volume ls | grep postgres_data
docker volume rm owlBack_postgres_data 2>/dev/null || true

# 2. 启全部 owl_v2 服务（含 kea/bind）
docker compose up -d

# 3. 等 postgres 初始化完成（首次约 30s 跑 dbv2 全套 schema）
docker compose logs -f postgresql | grep "ready to accept connections"

# 4. 验证
psql -h localhost -U postgres -d owl_v2 -c "\dt"   # 应见 dbv2 全部表
curl http://localhost:8000/                         # kea-ctrl-agent
dig @localhost tenant1.owl AAAA                     # BIND

# 5. 启 wisefido-* services（host-run）
sudo systemctl start owlback.cardagg owlback.data owlback.iot owlback.qinglan owlback.sensor owlback.sleepace
```

## 回退到 owlrd（紧急情况）

```bash
docker compose down
# 恢复旧 volume
docker volume create owlBack_postgres_data
docker run --rm -v owlBack_postgres_data:/restore postgres:15 \
  pg_restore -d postgres /backup1.0/db-rebuild-20260508-080819/full.sql

# 改 docker-compose.yml: POSTGRES_DB=owlrd, initdb.d=../owlRD/dbv1
# 改 .env: DB_NAME=owlrd
docker compose up -d
```

## owl Go service 调用 kea REST 范例

```go
// 创建新 tenant：分配 /48 prefix
type Cmd struct {
    Command string                 `json:"command"`
    Service []string               `json:"service"`
    Args    map[string]interface{} `json:"arguments"`
}

cmd := Cmd{
    Command: "subnet6-add",
    Service: []string{"dhcp6"},
    Args: map[string]interface{}{
        "subnet6": []map[string]interface{}{
            {"subnet": "fd00:0:0001::/48", "id": 1},
        },
    },
}
resp, _ := http.Post("http://localhost:8000/", "application/json",
                     bytes.NewReader(jsonMarshal(cmd)))
```

## 配置文件清单

```
kea/
  dhcp6/
    kea-dhcp6.conf       DHCPv6 daemon 配置（subnet 定义 + DDNS hooks）
  ctrl/
    kea-ctrl-agent.conf  REST API 配置（port 8000）

bind/
  named.conf             BIND 主配置
  zones/
    tenant1.owl.zone     tenant 1 正向 zone（AAAA records）
    d.0.0.f.ip6.arpa.zone  反向 zone (fd00::/8 ULA 范围)
  keys/
    ddns-update.key      DDNS TSIG 密钥（与 kea-ddns 共享）

scripts/
  setup_owl_v2.sh        手动 setup 脚本（默认首次 docker compose up
                         自动跑 init.d，无需手动；保留作灾备 / 重跑用）
```

## 待 Phase B

当前是基础设施骨架。Phase B 完整对接 owl service 时需要：
1. owl-data 加 kea-ctrl-agent client（HTTP wrapper）
2. CreateTenant / Branch / Unit / Room / Bed handler 调用 kea REST
3. kea-ddns 集成 BIND 自动维护 zone
4. seed scripts 90_-93_ 写完

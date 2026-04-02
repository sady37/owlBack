# OwlBack 系统架构图

## 1. 整体系统架构图

```mermaid
graph TB
    subgraph DeviceLayer["🔌 设备层"]
        Radar["📡 Radar 设备<br/>生命信号传感器"]
        Sleepace["🛌 Sleepace 设备<br/>睡眠监测设备"]
    end

    subgraph TransportLayer["🔗 传输层"]
        MQTT["MQTT Broker<br/>1883"]
    end

    subgraph GatewayLayer["🌐 接入网关层"]
        Qinglan["wisefido-qinglan<br/>HTTP:8081/HTTPS:8443<br/>Radar网关"]
        Sleepace_GW["wisefido-sleepace<br/>HTTP:8083<br/>Sleepace网关"]
    end

    subgraph EventBusLayer["🚀 事件总线 Redis"]
        IotMonitor["iot:monitor:stream"]
        IotStat["iot:stat:stream"]
        IotEvent["iot:event:stream"]
        IotAlarm["iot:alarm:stream"]
        IotAuth["iot:auth:stream"]
        ConfigCard["config:card:stream"]
        ConfigAlarmDevice["config:alarmDevice:stream"]
        ConfigAlarmProcess["config:alarmProcess:stream"]
        CardRealtime["card:realtime:stream"]
        CardStatus["card:status:stream"]
        CardState["card:state:*<br/>Hash"]
    end

    subgraph ComputeLayer["⚙️ 计算聚合层"]
        CardAgg["wisefido-cardagg<br/>卡片聚合引擎"]
        IoT["wisefido-iot<br/>时序数据落库"]
        AI["wisefido-ai<br/>告警推理引擎"]
    end

    subgraph APILayer["🌍 业务API层"]
        Data["wisefido-data<br/>HTTP:8080<br/>主业务API/SSE"]
    end

    subgraph StorageLayer["💾 存储层"]
        PostgreSQL["PostgreSQL/Timescale DB<br/>5432"]
        RedisCache["Redis Cache/KV<br/>6379"]
    end

    subgraph ClientLayer["👥 客户端"]
        Frontend["前端应用<br/>SSE实时推送"]
        Admin["管理后台"]
    end

    %% 设备 -> MQTT
    Radar -->|MQTT| MQTT
    Sleepace -->|MQTT| MQTT

    %% MQTT -> 网关
    MQTT -->|MQTT订阅| Qinglan
    MQTT -->|MQTT订阅| Sleepace_GW

    %% 网关 -> Redis Streams (IoT事件)
    Qinglan -->|发布| IotMonitor
    Qinglan -->|发布| IotStat
    Qinglan -->|发布| IotEvent
    Qinglan -->|发布| IotAlarm
    Qinglan -->|发布| IotAuth
    Sleepace_GW -->|发布| IotMonitor
    Sleepace_GW -->|发布| IotEvent
    Sleepace_GW -->|发布| IotAlarm

    %% Data API -> Redis Streams (配置事件)
    Data -->|发布| ConfigCard
    Data -->|发布| ConfigAlarmDevice
    Data -->|发布| ConfigAlarmProcess

    %% 计算层订阅 IoT 和配置流
    IotMonitor -->|消费| CardAgg
    IotStat -->|消费| CardAgg
    IotEvent -->|消费| CardAgg
    IotAlarm -->|消费| CardAgg
    ConfigCard -->|消费| CardAgg
    ConfigAlarmDevice -->|消费| CardAgg
    ConfigAlarmProcess -->|消费| CardAgg

    IotMonitor -->|消费| IoT
    IotStat -->|消费| IoT
    IotEvent -->|消费| IoT
    IotAlarm -->|消费| IoT
    IotAuth -->|消费| IoT

    IotMonitor -->|轮询读取| AI
    IotAlarm -->|轮询读取| AI

    %% 计算层 -> Redis Streams (卡片事件)
    CardAgg -->|发布| CardRealtime
    CardAgg -->|发布| CardStatus
    CardAgg -->|写| CardState

    %% Data API 订阅卡片流
    CardRealtime -->|消费| Data
    CardStatus -->|消费| Data

    %% 数据API -> 存储
    Data <-->|CRUD| PostgreSQL
    Data <-->|缓存| RedisCache
    CardAgg <-->|写入| PostgreSQL
    IoT -->|写入时序数据| PostgreSQL
    AI <-->|读写| RedisCache
    AI <-->|查询规则| PostgreSQL

    %% API -> 网关调用
    Data <-->|HTTP调用<br/>设备控制| Qinglan
    Data <-->|HTTP调用<br/>设备状态| Sleepace_GW

    %% API -> 客户端
    Data -->|HTTP REST/SSE| Frontend
    Data -->|HTTP REST| Admin

    style DeviceLayer fill:#e1f5ff
    style TransportLayer fill:#fff3e0
    style GatewayLayer fill:#f3e5f5
    style EventBusLayer fill:#e8f5e9
    style ComputeLayer fill:#fce4ec
    style APILayer fill:#f1f8e9
    style StorageLayer fill:#ede7f6
    style ClientLayer fill:#e0f2f1
```

## 2. 数据流时序图

```mermaid
sequenceDiagram
    actor Device as 设备
    participant MQTT as MQTT Broker
    participant GW as 接入网关<br/>qinglan/sleepace
    participant RS as Redis Streams<br/>事件总线
    participant CardAgg as CardAgg<br/>聚合引擎
    participant API as wisefido-data<br/>业务API
    participant Client as 前端客户端

    Device->>MQTT: 1. 发送传感器数据<br/>(生命信号/睡眠数据)
    MQTT->>GW: 2. 网关消费MQTT消息
    GW->>RS: 3. 发布到iot:*:stream<br/>(monitor/stat/event/alarm/auth)

    par 并行处理
        RS->>CardAgg: 消费iot事件
        RS->>API: wisefido-iot消费<br/>写入TimescaleDB
    end

    CardAgg->>CardAgg: 聚合卡片状态<br/>应用告警规则
    CardAgg->>RS: 发布card:*:stream<br/>与card:state:*

    RS->>API: wisefido-data<br/>订阅card流
    API->>API: 聚合API数据<br/>维护缓存

    API->>Client: SSE推送实时更新
    Client->>Client: 前端刷新显示

    Note over Device,Client: 用户配置变更流程
    Client->>API: 更新配置(告警/设备等)
    API->>RS: 发布config:*:stream
    RS->>CardAgg: CardAgg消费配置<br/>刷新告警规则
    API->>GW: HTTP调用设备控制
```

## 3. 模块依赖关系图

```mermaid
graph LR
    subgraph owl-common["owl-common<br/>共享库"]
        RedisHelpers["Redis Stream定义<br/>Card读写器"]
        DB["DB/Config/Log<br/>工具函数"]
    end

    subgraph Gateways["接入网关"]
        Qinglan["wisefido-qinglan"]
        Sleepace["wisefido-sleepace"]
    end

    subgraph Services["服务模块"]
        CardAgg["wisefido-cardagg"]
        IoT["wisefido-iot"]
        Data["wisefido-data"]
        AI["wisefido-ai"]
    end

    Qinglan -->|依赖| owl-common
    Sleepace -->|依赖| owl-common
    CardAgg -->|依赖| owl-common
    IoT -->|依赖| owl-common
    Data -->|依赖| owl-common
    AI -->|依赖| owl-common

    %% 服务间通信
    CardAgg -->|Redis Streams| IoT
    CardAgg -->|Redis Streams| Data
    CardAgg -->|Redis Streams| AI

    style owl-common fill:#fff9c4
    style Gateways fill:#f3e5f5
    style Services fill:#bbdefb
```

## 4. Redis Streams 消息流向图

```mermaid
graph LR
    subgraph Source["🔹 消息源"]
        Q["wisefido-qinglan<br/>iot:monitor/stat/event/alarm/auth"]
        S["wisefido-sleepace<br/>iot:monitor/event/alarm"]
        D["wisefido-data<br/>config:alarmDevice<br/>config:alarmProcess<br/>config:card"]
    end

    subgraph Streams["Redis Streams"]
        IotStreams["iot:* streams<br/>(monitor/stat/event/alarm/auth)"]
        ConfigStreams["config:* streams<br/>(card/alarmDevice/alarmProcess)"]
        CardStreams["card:* streams<br/>(realtime/status)"]
    end

    subgraph Consumer["⚡ 消费者"]
        CardAgg["CardAgg<br/>GROUP: cardagg-group"]
        IoT["IoT Service"]
        Data["Data API<br/>GROUP: wisefido-data-consumer"]
        AI["AI Service<br/>轮询模式"]
    end

    Q -->|发布| IotStreams
    S -->|发布| IotStreams
    D -->|发布| ConfigStreams

    IotStreams -->|消费| CardAgg
    ConfigStreams -->|消费| CardAgg
    IotStreams -->|消费| IoT

    CardAgg -->|发布| CardStreams
    CardStreams -->|消费| Data

    IotStreams -->|轮询读取| AI

    style Source fill:#ffccbc
    style Streams fill:#c8e6c9
    style Consumer fill:#bbdefb
```

## 5. 部署架构图

```mermaid
graph TB
    subgraph Infrastructure["基础设施"]
        MQTT["MQTT Broker<br/>Port 1883"]
        PG["PostgreSQL/Timescale<br/>Port 5432"]
        Redis["Redis<br/>Port 6379"]
        MySQL["MySQL<br/>Port 3306"]
    end

    subgraph Microservices["微服务"]
        subgraph GW["网关层"]
            Q["wisefido-qinglan<br/>:8081/:8443"]
            S["wisefido-sleepace<br/>:8083"]
        end

        subgraph Compute["计算层"]
            C["wisefido-cardagg"]
            I["wisefido-iot"]
            A["wisefido-ai<br/>:8085"]
        end

        API["wisefido-data<br/>:8080"]
    end

    subgraph Clients["客户端"]
        WEB["Web Frontend"]
        ADMIN["Admin Dashboard"]
        MOBILE["Mobile App"]
    end

    WEB -->|HTTP/SSE| API
    ADMIN -->|HTTP/SSE| API
    MOBILE -->|HTTP/SSE| API

    Q -->|MQTT| MQTT
    S -->|MQTT| MQTT

    Q -->|Redis Streams| Redis
    S -->|Redis Streams| Redis
    API -->|Redis Streams| Redis
    C -->|Redis Streams| Redis
    I -->|Redis Streams| Redis
    A -->|Redis Cache| Redis

    API -->|SQL| PG
    C -->|SQL| PG
    I -->|SQL| PG
    A -->|SQL| PG

    C -->|依赖| Redis
    I -->|依赖| Redis
    A -->|依赖| Redis

    style Infrastructure fill:#fff9c4
    style GW fill:#f3e5f5
    style Compute fill:#fce4ec
    style API fill:#c8e6c9
    style Clients fill:#bbdefb
```

## 6. 告警处理流程图

```mermaid
stateDiagram-v2
    [*] --> DataCollection: 设备数据通过网关入站

    DataCollection --> IotStream: 发布到iot:*:stream

    IotStream --> CardAggEnrich: CardAgg消费并聚合

    CardAggEnrich --> RuleEval: 应用告警规则<br/>从config:alarmDevice/Process读取

    RuleEval --> AlarmGen: 生成告警事件
    AlarmGen --> CardStream: 发布到card:realtime:stream<br/>card:status:stream

    CardStream --> DataAPI: wisefido-data订阅<br/>更新卡片状态

    DataAPI --> Notification: 通过SSE推送前端

    Notification --> Frontend: 前端实时显示告警

    Frontend --> [*]

    note right of RuleEval
        支持规则类型:
        - 阈值告警
        - 异常检测
        - 规则推理
    end
```

---

## 核心特性总结

| 特性 | 说明 |
|-----|------|
| **实时数据流** | Redis Streams 实现毫秒级消息传递 |
| **微服务架构** | 6个独立服务，松耦合高内聚 |
| **事件驱动** | 基于Redis Stream的发布-订阅模式 |
| **时序存储** | TimescaleDB用于设备数据长期存储 |
| **消费者组** | Redis Consumer Group支持分布式处理 |
| **SSE实时推送** | wisefido-data通过SSE向前端推送实时数据 |
| **多设备支持** | Radar + Sleepace双设备接入 |
| **告警处理** | CardAgg+AI双引擎告警决策 |


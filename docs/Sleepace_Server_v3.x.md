# 服务架构

1. Sleepace 设备将监测数据上传到 Sleepace 服务
2. Sleepace 服务通过消息队列发布数据
3. 第三方服务向 Sleepace 消息队列订阅感兴趣的数据主题，当消息队列收到 Sleepace 服务发布该主题消息时，向第三方服务推送消息
4. 第三方可以通过 Sleepace 提供的 HTTP 接口向 Sleepace 服务发送设备控制指令

## 交互流程

1. 厂商服务连接到 MQTT 服务，并订阅数据
2. **BM8701-2（Wi-Fi版）**：使用 Sleepace 提供的设备配网工具为设备配置 Wi-Fi 与服务器地址
3. **M901L（4G版）**：根据产品使用教程插上 SIM 卡后，设备将会自动连接服务器
4. 为设备绑定用户
5. 设备会开始上报实时数据，并通过 MQTT service 转发给厂商

# 接入前必读

请仔细阅读以下内容，若有疑问，请联系我司商务或技术支持团队~

## 技术文档组成部分

### BM8701-2（Wi-Fi版）、Z400TWP-3、M8701W

- **云端 API 文档**，包含：API-实时数据订阅、API-控制和报告查询
- **配网工具（Android、iOS App）**
  - 通过 BLE 给设备配置 Wi-Fi 和需要连接的服务器
- **配网原生 SDK**
  - iOS SDK 下载地址
  - Android SDK 下载地址
- **配网 JS SDK**：JS SDK 下载地址
  - 如果贵司没有自己的 App 或小程序，可以不用集成本 SDK，可用配网工具
- **配置信息**：包含渠道、安全码、TOPIC 等
  - 如果没有，请先与接收文档的同事确认，其次再找我司商务同事沟通

### M901L（4G版）

- **云端 API 文档**，包含：API-实时数据订阅、API-控制和报告查询
- **配置信息**：包含渠道、安全码、TOPIC 等
  - 如果没有，请先与接收文档的同事确认，其次再找我司商务同事沟通

## 设备使用前必做事宜

### Sleepace 测试服务器

1. 提供设备 ID 给我们，我们需要录入到测试环境下厂商的渠道内（一般送样前会录入并提供设备 ID 密文，若绑定设备时，提示状态码 5 和 10，请联系我们）
2. **BM8701-2（Wi-Fi版）**：使用配网工具给设备配置到 Sleepace 测试服务器
3. **M901L（4G版）**：对接前联系我们将设备切到 Sleepace 测试服务器。我们切换完成后，请按照产品说明书插入 SIM 卡，重启设备，设备将会自动连接 Sleepace 测试服务器
4. 对接前请一定先看文档了解设备功能和对接方案，其次有问题先查看常见问题

### 商用服务器

**前提：使用厂商自己的服务器**

1. 联系我们提供设备 ID 明文和密文文件
2. 贵司录入到自己的数据库
3. **BM8701-2（Wi-Fi版）**：使用配网工具给设备配置到对应的服务器
4. **M901L（4G版）**：提供贵司服务器 IP 和端口给我们，我们会把设备连接的服务 IP 和端口改成贵司的。按照产品说明书插入 SIM 卡，设备将会自动连接贵司商用服务器

## 2022/7/5 之前已接入设备的厂商必读

Sleepace 养老对接方案 3.x 更新新的报告算法，并将数据进行了迁移。

适用于在 2022/7/5 之前已经对接了 BM8701-2，且想要升级到 Sleepace 养老对接方案 3.x 的厂商（无论是要更新 BM8701-2（Wi-Fi版），还是新接入 M901L（4G版））。若设备已在使用，原来产生的历史数据将会丢失，若需要使用旧数据，需要调用接口重新生成报告至新的数据库。

## 关于样机测试和开发

一般在贵司接入前，我司会提供测试样机和我司的养老系统（https://xgh.sleepace.com/）使用，而后贵司才会确定开始接入到贵司的系统。我们需要强调，我司的养老系统和对接的测试环境是不通的，两者没有任何关系，请不要混用。

## 关于固件升级

请一定要实现固件升级的业务，原因如下：

1. 固件内置本地算法，解析采集到的信号数据可以分析出心率、呼吸率、睡眠状态等等，我司一直长足于提高算法的准确率和稳定性，每次有较大提高，我们都会更新固件
2. 如果贵司订单量小或者出货时间紧急，我们会采用库存产品出库，库存产品的固件不一定是最新的，所以最好要升级到最新的固件

## 关于商用服务器部署

贵司开发测试完成后，需联系我司商务提供服务部署包，贵司部署到你们的服务器。服务器要求如下：

- 最好是 Ubuntu（20），或者是其他 Linux 系统
- 四核 CPU，8G 内存
- 必须是 root 权限，因为需要安装服务
- 数据库是 MySQL 数据库（V 5.7）

**注：** 请不用连接 Sleepace 的测试服务器作为商用使用，其一，我们的测试环境不稳定，经常重启，会影响到贵司的使用；其二，太多设备接入会对我司服务器造成压力，我们会不定期关闭一些测试渠道

# 开始对接

## Sleepace 测试环境配置信息

| 项目 | 值 |
|------|-----|
| MQTT(tcp) 地址 | 120.24.68.136:1888 |
| HTTP 地址 | 120.77.233.171:8090 |
| 接入设备连接的服务器地址（BM8701-2(Wi-Fi版)需要） | 120.77.233.171:29012 |
| 账号(appId)/渠道号(channelId) | Sleepace 提供 |
| 密码(secureKey) | Sleepace 提供 |
| TOPIC | Sleepace 提供 |

**注：** 开发完成上线后，则是厂商商用服务器的信息

## 第一步：订阅数据

厂商服务器连接 Sleepace 消息队列订阅数据

### 推送方式

1. **MQTT 推送**：默认为 MQTT 推送，测试过程可以使用第三方客户端测试，如：mqtt.fx，请自行百度搜索下载
2. **HTTP 推送**：可调用设置数据订阅方式设置为 HTTP 推送

# **API-实时数据订阅：** https://www.yuque.com/ysss/elder/eoxu4n

**版本：** v3.11

### 订阅方式

Sleepace 实时数据可以通过消息队列（MQTT 协议）的订阅

#### MQTT 消息订阅

**Sleepace 测试环境消息队列基本信息**

| 项目 | 值 |
|------|-----|
| MQTT(tcp) 地址 | 120.24.68.136:1888 |
| 账号(appId)/渠道号(channelId) | Sleepace 提供 |
| 密码(secureKey) | Sleepace 提供 |
| TOPIC | Sleepace 提供 |

#### HTTP 接口方式订阅

接入厂商根据规范实现接口，并通过设置数据订阅方式设置使用 HTTP 接口方式推送。

**Method:** POST

**Body:**

```json
[
  {
    "dataKey": "<<推送消息类型的key>>",
    "timeStamp": "<<事件发生时间>>",
    "data": "<<消息实体>>",
    "deviceId": "设备密文id"
  }
]
```

**字段说明**

| 字段 | 类型 | 描述 |
|------|------|------|
| dataKey | String | 消息类型 |
| timeStamp | long | 消息时间戳，毫秒 |
| data | json | 消息 json 实体，详见各推送内容 |
| deviceId | String | 设备密文 id |

**响应**

```json
{
  "status": 0,
  "msg": ""
}
```

| 字段 | 类型 | 描述 |
|------|------|------|
| status | int | 消息类型，0:接收成功 其他:失败 |
| msg | String | 失败原因 |

**注：** 部分消息类型返回失败状态码服务器会启动重试机制

## 设备上下线事件

**dataKey:** `connectionStatus`

**消息内容：**

```json
[
  {
    "dataKey": "connectionStatus",
    "deviceId": "<<设备Id>>",
    "data": {
      "connectionStatus": "<<设备在线状态，0:不在线，1:在线>>"
    },
    "timeStamp": "<<数据发布时间>>"
  }
]
```

## 实时数据

**dataKey:** `realtime`

**消息内容：**

每条消息可能包含多条心率、呼吸率、翻身、温度和湿度记录，结构如下：

```json
[
  {
    "dataKey": "realtime",
    "deviceId": "<<设备Id>>",
    "data": {
      "leftRight": "<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>",
      "breath": "<<呼吸率>>",
      "heart": "<<心率>>",
      "trunOver": "<<翻身>>",
      "bodyMove": "<<体动>>",
      "sitUp": "<<坐起>>",
      "initStatus": "<<初始化>>",
      "temp": "<<温度>>",
      "hum": "<<湿度>>",
      "bedStatus": "<<在床状态>>",
      "signalQuality": "<<信号质量，仅针对特定设备>>"
    },
    "timeStamp": "<<数据发布时间>>"
  }
]
```

**字段说明**

| 字段 | 类型 | 描述 |
|------|------|------|
| breath | int | 呼吸率，255 时为初始化状态 |
| heart | int | 心率，255 时为初始化状态 |
| turnOver | int | 翻身次数（适用于：BM8701-2 固件版本≤2.x 的设备） |
| bodyMove | int | 体动次数（适用于：BM8701-2 固件版本≥5.x 的设备及 M901L） |
| sitUp | int | 坐起<br>0：非坐起；1：坐起；<br>8：情况一：sitUp、bedStatus/inbedStatus、sleepStage 都为 8 时，说明算法在初始化；<br>情况二：sitUp 为 8，bedStatus/inbedStatus、sleepStage 不全为 8，说明当前状态未发生改变，保持前一个有效状态的值<br>其他：无效 |
| initStatus | int | 初始化<br>0：正常；1：设备初始化中<br>**注：** 值为 1 时，其他参数 breath、heart、turnOver、bodyMove、sitUp、temp、hum、bedStatus 的值都是无效的，请注意处理此种状态 |
| temp | int | 温度（BM8701-2 有此数据）<br>**注：** 2022 年 2 月后出货产品无此功能，接口保留，请不要使用 |
| hum | int | 湿度（BM8701-2 有此数据）<br>**注：** 2022 年 2 月后出货产品无此功能，接口保留，请不要使用 |
| bedStatus | int | 0: 未离床<br>1: 离床<br>8：情况一：sitUp、bedStatus/inbedStatus、sleepStage 都为 8 时，说明算法在初始化；<br>情况二：bedStatus/inbedStatus 为 8，sitUp、sleepStage 不全为 8，说明当前状态未发生改变，保持前一个有效状态的值<br>其他：无效<br>**注：** 若接入的设备为 BM8701 和 M901L，请使用本接口查询在离床，不建议使用 inBedStatus |
| signalQuality | int | 信号质量，描述采集到的原始信号的质量好坏<br>取值范围：0~5（其中 5 是最好，1 是最差，4 和 5 的信号是比较好，可信度高，1~3 信号较差，可信度低）<br>**注：** BM8701-2 固件版本≥v6.22 才支持 |

**例子：**

```json
[
  {
    "dataKey": "realtime",
    "deviceId": "002s7dbcssyxc",
    "data": {
      "leftRight": 0,
      "breath": 15,
      "heart": 65,
      "turnOver": 0,
      "temp": 25,
      "hum": 32,
      "bedStatus": 0
    },
    "timeStamp": 1502069838
  },
  {
    "dataKey": "realtime",
    "deviceId": "0010y81hgof88",
    "data": {
      "leftRight": 0,
      "breath": 16,
      "heart": 71,
      "turnOver": 1,
      "temp": 25,
      "hum": 32,
      "bedStatus": 0
    },
    "timeStamp": 1502069356
  }
]
```

## 在床状态改变事件

用户上床、离床时会产生在床状态事件

**dataKey:** `inBedStatus`

**消息内容：**

每条消息可能包含多条在床状态记录，结构如下：

```json
[
  {
    "dataKey": "inBedStatus",
    "deviceId": "<<设备Id>>",
    "data": {
      "leftRight": "<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>",
      "inbedStatus": "<<在床状态>>"
    },
    "timeStamp": "<<数据发布时间>>"
  }
]
```

**字段说明**

| 字段 | 类型 | 描述 |
|------|------|------|
| inbedStatus | int | 0: 未离床<br>1: 离床<br>8：情况一：sitUp、bedStatus/inbedStatus、sleepStage 都为 8 时，说明算法在初始化；<br>情况二：inbedStatus 为 8，sitUp、sleepStage 不全为 8，说明当前状态未发生改变，保持前一个有效状态的值<br>其他：无效<br>**注：** 本接口为 Z400TWP（睡眠带）的旧接口，若要使用在离床状态，请使用 realtime>bedStatus |

**例子：**

```json
[
  {
    "dataKey": "inBedStatus",
    "deviceId": "002s7dbcssyxc",
    "data": {
      "leftRight": 0,
      "inbedStatus": 1
    },
    "timeStamp": 1502069838
  },
  {
    "dataKey": "inBedStatus",
    "deviceId": "0010y81hgof88",
    "data": {
      "leftRight": 0,
      "inbedStatus": 0
    },
    "timeStamp": 1502069356
  }
]
```

## 睡眠状态事件

用于上报用户睡眠状态（清醒、浅睡、深睡），事件产生的时机：

1. 当用户睡眠状态发生改变时
2. 设备重新联网时，也会上报一次用户当前的睡眠状态

**dataKey:** `sleepStage`

**消息内容：**

```json
[
  {
    "dataKey": "sleepStage",
    "deviceId": "<<设备Id>>",
    "data": {
      "leftRight": "<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>",
      "sleepStage": "<<睡眠状态>>"
    },
    "timeStamp": "<<数据发布时间>>"
  }
]
```

**字段说明**

| 字段 | 类型 | 描述 |
|------|------|------|
| sleepStage | int | 1: 清醒<br>2: 浅睡<br>4: 深睡<br>8：情况一：sitUp、bedStatus/inbedStatus、sleepStage 都为 8 时，说明算法在初始化；<br>情况二：sleepStage 为 8，sitUp、bedStatus/inbedStatus 不全为 8，说明当前状态未发生改变，保持前一个有效状态的值<br>其他：无效 |

**例子：**

```json
[
  {
    "dataKey": "sleepStage",
    "deviceId": "002s7dbcssyxc",
    "data": {
      "leftRight": 0,
      "sleepStage": 1
    },
    "timeStamp": 1502069838
  },
  {
    "dataKey": "sleepStage",
    "deviceId": "0010y81hgof88",
    "data": {
      "leftRight": 0,
      "sleepStage": 3
    },
    "timeStamp": 1502069356
  }
]
```

## 设备监测垫脱落事件（BM8701-2&M901L 专用）

设备传感器脱落产生的事件

**dataKey:** `deviceSenSor`

**消息内容：**

每条消息可能包含多条在床状态记录，结构如下：

```json
[
  {
    "dataKey": "deviceSenSor",
    "deviceId": "<<设备Id>>",
    "data": {
      "status": "<<离床传感器状态0,脱落，1、已经插上>>"
    }
  }
]
```

**字段说明**

| 字段 | 类型 | 描述 |
|------|------|------|
| status | int | 0: 脱落<br>1: 已经插上 |

**例子：**

```json
[
  {
    "dataKey": "deviceSenSor",
    "deviceId": "<<设备Id>>",
    "data": {
      "status": "<<离床传感器状态0,脱落，1、已经插上>>"
    }
  }
]
```

## 设备监测垫初始化事件

设备传感器初始化的事件

**dataKey:** `pressureSenSor`

**消息内容：**

每条消息可能包含多条在床状态记录，结构如下：

```json
[
  {
    "dataKey": "pressureSenSor",
    "deviceId": "<<设备Id>>",
    "leftRight": "<<左右侧>>",
    "data": {
      "status": "<<0:未初始化1:已初始化>>"
    }
  }
]
```

**字段说明**

| 字段 | 类型 | 描述 |
|------|------|------|
| status | int | 0: 未初始化<br>1: 已初始化 |

## 睡眠报告生成事件

**dataKey:** `analysis`

**消息内容：**

第三方收到时睡眠报告事件后，可以通过"查询睡眠报告"接口获取到详细睡眠报告。每条消息可能包含多条睡眠报告记录，结构如下：

```json
[
  {
    "dataKey": "analysis",
    "deviceId": "<<设备Id>>",
    "data": {
      "userId": "<<用户UID>>",
      "startTime": "<<监测开始时间>>"
    },
    "timeStamp": "<<数据发布时间>>"
  }
]
```

**字段说明**

| 字段 | 类型 | 描述 |
|------|------|------|
| userId | String | 用户 UID |
| startTime | int | 监测开始时间 |

**例子：**

```json
[
  {
    "dataKey": "analysis",
    "deviceId": "002s7dbcssyxc",
    "data": {
      "userId": "",
      "startTime": 1556240282
    },
    "timeStamp": 1502069838
  },
  {
    "dataKey": "analysis",
    "deviceId": "0010y81hgof88",
    "data": {
      "userId": "",
      "startTime": 1556242507
    },
    "timeStamp": 1502069356
  }
]
```

## 设备升级进度

**dataKey:** `upgradeProgress`

**消息内容：**

```json
[
  {
    "dataKey": "upgradeProgress",
    "deviceId": "<<设备Id>>",
    "data": {
      "length": "<<本次通信要下载长度>>",
      "offset": "<<已下载的固件长度>>"
    },
    "timeStamp": 1502069838
  }
]
```

**字段说明**

| 字段 | 类型 | 描述 |
|------|------|------|
| length | int | 本次通信要下载长度 |
| offset | int | 已下载的固件长度 |

**例子：**

```json
[
  {
    "dataKey": "upgradeProgress",
    "deviceId": "002s7dbcssyxc",
    "data": {
      "length": 56213,
      "offset": 1024
    },
    "timeStamp": 1502069838
  }
]
```

## 报警事件

**dataKey:** `alarmNotify`

**消息内容：**

默认以单个报警事件的数据结构推送，你可以通过设置数据订阅方式接口将数据结构修改以数组的方式推送

**单个事件数据结构：**

```json
{
  "dataKey": "alarmNotify",
  "deviceId": "xxx",
  "timeStamp": 1660620197,
  "data": {
    "id": 1,
    "type": "xxx",
    "timestamp": 1660620197,
    "userId": "25",
    "deviceId": "xxx",
    "status": 0,
    "relieveReason": "xxx",
    "relieveTime": 0
  }
}
```

**数组结构：**

```json
[
  {
    "dataKey": "alarmNotify",
    "deviceId": "xxx",
    "timeStamp": 1660620197,
    "data": {
      "id": 1,
      "type": "xxx",
      "timestamp": 1660620197,
      "userId": "25",
      "deviceId": "xxx",
      "status": 0,
      "relieveReason": "xxx",
      "relieveTime": 0
    }
  }
]
```

**字段说明**

- `id`: 报警事件 id
- `type`: 报警类型
- `timestamp`: 时间戳，单位秒
- `userId`: 用户 id
- `deviceId`: 设备密文 id
- `status`: 报警状态，0 新产生的报警事件（未处理），1 系统自动解除的报警事件
- `relieveReason`: 解除原因
- `relieveTime`: 解除报警的时间戳，单位秒

### 报警类型

| 类型 | 描述 |
|------|------|
| alarmSensorFall | 传感器脱落报警（BM8701-2、M901L 支持，SDC100、M8701W、Z400TWP-3 不支持） |
| alarmLeftBed | 离床报警 |
| alarmHeartRateFast | 心率过速报警 |
| alarmHeartRateSlow | 心率过缓报警 |
| alarmBreathRateFast | 呼吸过速报警 |
| alarmBreathRateSlow | 呼吸过缓报警 |
| alarmBreathRatePause | 呼吸暂停报警 |
| alarmBodymove | 频繁体动报警 |
| alarmNoBodymove | 无体动报警 |
| alarmNoTurnOver | 久未翻身报警 |
| alarmSitup | 疑似坐起报警（BM8701-2/M901L 硬板传感器、M8701W 支持，SDC100、BM8701-2/M901L 压电传感器、Z400TWP-3 不支持） |
| alarmOnBed | 在床报警 |



# API-控制和报告查询

**文档链接：** https://www.yuque.com/ysss/elder/xig49g

## 修改记录

| 版本号 | 修订人 | 修订日期 | 修改内容 |
|--------|--------|----------|----------|
| 3.0 | 黄朝维 | 2022-07-01 | 文档创建 |
| 3.1 | | 2022-9-2 | 增加报警接口 |
| 3.2 | | 2022-10-31 | 增加数据订阅方式设置接口 |
| 3.3 | | 2023-03-28 | 增加4g设备通讯信息查询接口 |
| 3.4 | | 2023-04-10 | 1、增加"设备绑定信息获取"；<br>2、增加设备原始数据上传相关接口 |
| 3.5 | | 2023-05-23 | 报告查询接口（get24HourDailyWithMaxReport）的MaxReport增加：体动指数、体动指数扣分、睡眠连续性、睡眠连续性扣分和呼吸风险等 |
| 3.51 | | 2023-10-27 | 增加报警设置的体动报警（频繁体动报警、无体动报警）设置时长说明 |
| 3.52 | | 2024-1-22 | get24HourDailyWithMaxReport接口对特定设备增加signalQuality字段 |
| 3.53 | | 2024-3-12 | 增加心率呼吸率模式设置和心率呼吸率模式获取接口 |
| 3.54 | | 2024-6-15 | 1.增加历史数据存储方式，支持设置定时生成报告或离床1小时自动生成报告<br>2.自动生成报告模式下，支持调接口生成报告 |
| 3.55 | | 2025-9-15 | 1.离床灵敏度这支持设置高、中、低<br>2.设备传感器支持手动校准 |

第三方可以通过 Sleepace 服务提供的 HTTP 接口发送指令消息。

## HTTP 接口基本描述

- **请求协议接口数据格式支持：** Content-Type: application/json
- **响应协议接口数据格式：** Content-Type: application/json
- **协议接口数据字符编码：** UTF-8
- **Method:** POST 

## 设置数据订阅方式

通过该接口设置事件推送方式（未设置则默认推送方式为 MQTT 推送）

**URL:** `http(s)://domain{:port}/sleepace/system/pushType/set`
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
pushType:<<推送方式>>,
pushUrl:<<HTTP推送地址>>	,
retryConf:<<推送失败重试此事>>
}
}

**字段说明**

| 字段 | 类型 | 是否必传 | 描述 |
|------|------|----------|------|
| appId | string | 是 | 与消息队列的账号相同 |
| secureKey | string | 是 | 与消息队列的密码相同 |
| pushType | string | 是 | 推送方式：MQTT、HTTP |
| pushUrl | string | pushType 为 HTTP 时必传 | HTTP 推送地址 |
| retryConf | Map<String,Integer> | 否 | 各消息类型数据在 HTTP 推送失败时重试次数，默认推送一次 |
| alarmDataType | string | 否 | 之前报警数据结构为单个报警事件：{}；与其他数据结构不一致（数组：[]）。为规范报警数据结构数据，可以通过该字段进行指定（默认保持单个）。<br>取值：<br>- single：单个<br>- array：数组 |
### retryConf 详解

| 字段 | 类型 | 描述 |
|------|------|------|
| key | string | 需要设置重试的消息，值为各推送消息 datakey，详细 datakey 请参考实时数据订阅 |
| value | Integer | 对应失败重试次数，单位：次 |

**响应：**

```json
{
  "status": 0,
  "msg": null,
  "data": null
}
```

| 字段 | 类型 | 描述 |
|------|------|------|
| status | int | 状态码，0 表示成功，其他失败，详见状态码 |
| msg | string | 失败原因 |
| data | json | 最新实时数据，详见实时数据推送内容 |
## 查询数据订阅配置

通过该接口获取数据推送方式

**URL:** `http(s)://domain{:port}/sleepace/system/pushType/get`
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
}
}
字段	类型	是否必传	描述
appId	string	是	与消息队列的账号相同
secureKey			string	是	与消息队列的密码相同
响应
{
    "status": 0,
    "msg": null,
    "data": null
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg			string	失败原因
data	json	最新实时数据，详见实时数据推送内容

示例：
{
    "status": 0,
    "msg": null,
    "data": {
        "channel": 56952,
        "pushType": "HTTP",
        "pushUrl": "http://127.0.0.1:8091/sleepace/http/publishList",
        "createTime": 1666579500119,
        "updateTime": 1666579500119
    }
}
## 设备在线状态查询

**URL:** `http(s)://domain{:port}/sleepace/connectioStatus`
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备密文

响应
{
	"status": 0
	"msg": "",
data:{
connectionStatus:0,
}

}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
connectionStatus	int	设备在线状态，0:不在线， 1:在线

设备信息查询1
通过设备明文查询设备信息
URL
http(s)://domain{:port}/sleepace/deviceInfo/plaintextId
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
plaintextId:<<设备ID>>,
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
plaintextId	string	设备明文ID

响应
{
    "status": 0,
    "msg": “”,
    "data": {
        "devicePlaintextId": "设备明文ID",
        "deviceId": "设备ID",
        "version": "设备版本"
    }
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
plaintextId	string	设备明文ID
deviceId	string	设备密文
version	string	设备版本

设备信息查询2
通过设备密文查询设备信息
URL
http(s)://domain{:port}/sleepace/deviceInfo/deviceId
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备ID

响应
{
    "status": 0,
    "msg": “”,
    "data": {
        "devicePlaintextId": "设备明文ID",
        "deviceId": "设备ID",
        "version": "设备版本"
    }
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
plaintextId	string	设备明文ID
deviceId	string	设备ID
version	string	设备版本
用户绑定设备
Sleepace设备需要和用户进行关联，所以操作设备前需要对设备与用户进行绑定
URL
http(s)://domain{:port}/sleepace/bind

参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
leftRight:<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>,
userId:<<用户ID>>
gender:<<性别>>
age:<<年龄>>,
timezone:<<时区>>

}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备id（设备id密文）
leftRight	int	为0，养老设备都是单人设备
userId	string	合作方用户的唯一标识
gender	int	性别 1：男，2：女
age	int	年龄
timezone	int	用户所在时区，单位：秒，如：8时区=28800

响应
{
	"status": 0
"msg": ""

}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子

请求
Post   http(s)://domain{:port}/sleepace/bind
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:”ecfabc0e1297”,
leftRight:0,
userId:”user-1”
gender:1,
age:66,
timezone:28800

}
}

响应
{
	"status": 0
"msg": ""

}


用户解绑设备
解绑设备
URL
http(s)://domain{:port}/sleepace/unbind  

参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
leftRight:<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备id（设备id密文）
leftRight	int	为0，养老设备都是单人设备
响应
{
	"status": 0
"msg": ""

}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子

Request
Post   http(s)://domain{:port}/sleepace/unbind
Parameter {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:”ecfabc0e1297”,
leftRight:0
}
}
Response
{
"status": 0
"msg": ""
}

用户绑定信息获取
URL
http(s)://domain{:port}/sleepace/bindInfo

参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
userId:<<用户ID>>
}
}
字段	类型	描述
userId	string	合作方用户的唯一标识

响应
{
"status": 0
"msg": "",
“data”:[
{
“deviceId”:<<>>,
“deviceName”:<<>>,
“deviceType”:<<>>,
“leftRight”:<<>>,
“deviceVersion”:<<>>
},
...
]

}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
deviceId	string	设备密文
deviceName	string	设备明文
deviceType	int	设备类型
leftRight	int	养老设备都是0，表示为单人设备
deviceVersion	float	设备当前版本

例子
Request
Post http(s)://domain{:port}/sleepace/bindInfo
Parameter {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
userId:<<用户ID>>,
}
}

Response
{
"status": 0
"msg": "",
data:[{
“deviceId”:”hpjbv26gwbzee”,
“deviceName”:”Z4TWP18110003”,
“deviceType”:27,
“leftRight”:0,
“deviceVersion”:1.79
}
]

}

设备绑定信息获取
URL
http(s)://domain{:port}/sleepace/bindInfoByDevice

参数
{
token:{
appId:'',<<与消息队列的账号相同>>
secureKey:'',<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备密文ID>>
}
}
字段	类型	描述
deviceId	string	设备id（设备id密文)

响应
{
"status": 0
"msg": "",
“data”:[
{
"deviceId":'':<<>>,
"deviceName":<<>>,
"deviceType":<<>>,
"leftRight":<<>>,
"deviceVersion":<<>>
},
...
]

}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
deviceId	string	设备密文
deviceName	string	设备明文
deviceType	int	设备类型
leftRight	int	养老设备都是0，表示为单人设备
deviceVersion	float	设备当前版本

例子
Request
Post http(s)://domain{:port}/sleepace/bindInfoByDevice
Parameter {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<>>,
}
}

Response
{
"status": 0
"msg": "",
data:[{
“deviceId”:”hpjbv26gwbzee”,
“deviceName”:”Z4TWP18110003”,
“deviceType”:27,
“leftRight”:0,
“deviceVersion”:1.79
}
]

}

离床监测垫配置设置(M901L、BM8701-2)
URL
http(s)://domain{:port}/sleepace/leftBedDeviceStatus/set

参数
{
token:{
appId:'',<<与消息队列的账号相同>>
secureKey:'',<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备密文ID>>,
userId:<<用户ID>>,
leftBedDeviceStatus:<<离床脚垫配置>>
}
}
字段	类型	描述
deviceId	string	设备id（设备id密文)
userId	string	合作方用户的唯一标识
leftBedDeviceStatus	int	否连接离床监测垫  0.否 1.是

响应
{
"status": 0
"msg": "",
"data":null
}

离床监测垫配置获取(M901L、BM8701-2)
URL
http(s)://domain{:port}/sleepace/leftBedDeviceStatus/get

参数
{
token:{
appId:'',<<与消息队列的账号相同>>
secureKey:'',<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备密文ID>>,
userId:<<用户ID>>
}
}
字段	类型	描述
deviceId	string	设备id（设备id密文)
userId	string	合作方用户的唯一标识

响应
{
"status": 0
"msg": "",
"data":[
{
"deviceId":'':<<>>,
"deviceName":<<>>,
"deviceType":<<>>,
"leftBedDeviceStatus":<<>>
},
...
]

}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
deviceId	string	设备密文
deviceName	string	设备明文
deviceType	int	设备类型
leftBedDeviceStatus	int	否连接离床监测垫  0.否 1.是

用户睡眠状态查询
URL
http(s)://domain{:port}/sleepace/sleepStage
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
leftRight:<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>,

}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
leftRight	int	为0，养老设备都是单人设备

响应
{
	"status": 0
	"msg": "",
data:{
sleepStage:0,
}

}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
sleepStage	int	1: 清醒
2: 浅睡
4: 深睡


例子

请求
Post   http(s)://domain{:port}/sleepace/search
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
}
}

响应
{
	"status": 0
	"msg": "",
    “data”:{
      sleepStage:1
}

}


设备安装环境配置设置(BM8701-2&M901L&M8701W专用)
URL
http(s)://domain{:port}/sleepace/updateSetting
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
leftRight:<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>,
thickness:<<床垫厚度>>,
material:<<床垫>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
leftRight	int	为0，养老设备都是单人设备
thickness	int	床垫厚度, 厚度值1-3
 (1:5-10cm，2:11-20cm，3:21-30cm)
material	int	床垫材质, 值为1-5
1、海绵，2、弹簧3、乳胶，4、气垫5、其他

响应
{
	"status": 0
	"msg": "",
data：null
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子

请求
Post   http(s)://domain{:port}/sleepace/updateSetting
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
}，
data:{
deviceId:<<设备ID>>,
leftRight:<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>,
thickness:<<床垫厚度>>,
material:<<床垫>>
}
}

响应
{
	"status": 0
	"msg": "",
    “data”:null
}

设备安装环境配置获取(BM8701-2&M901L&M8701W专用)
URL
http(s)://domain{:port}/sleepace/getSetting
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
leftRight:<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
leftRight	int	为0，养老设备都是单人设备

响应
{
	"status": 0
	"msg": "",
"data":{
"thickness":1,
"material":1
}
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	json	
thickness	int	床垫厚度, 厚度值1-3
 (1:5-10cm，2:11-20cm，3:21-30cm)
material	int	床垫材质, 值为1-5
1、海绵，2、弹簧3、乳胶，4、气垫5、其他

鼾声干预开关设置（SDC100专用）
URL
http(s)://domain{:port}/sleepace/snoreConfig/set
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
flag:<<鼾声开关>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
flag	int	0：关闭 1：开启

响应
{
	"status": 0
	"msg": "",
data：null
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因

鼾声干预开关获取（SDC100专用）
URL
http(s)://domain{:port}/sleepace/snoreConfig/get
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID

响应
{
	"status": 0
	"msg": "",
"data"：{
"flag"：1
}
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	Object	设备配置
flag	int	鼾声开关：0关 1开
单双人模式设置（SDC100专用）
URL
http(s)://domain{:port}/sleepace/userMode/set
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
mode:<<单双人模式>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
mode	int	0：单人模式 1：双人模式

响应
{
	"status": 0
	"msg": "",
data：null
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
单双人模式获取（SDC100专用）
URL
http(s)://domain{:port}/sleepace/userMode/get
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID

响应
{
	"status": 0
	"msg": "",
"data"：{
"mode"：1
}
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	Object	设备配置
mode	int	单双人模式：0：单人模式 1：双人模式
实时数据上报模式设置
注：仅BM8701-2支持，固件版本需≥v6.67
URL
http(s)://domain{:port}/sleepace/realtimeMode/set
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
mode:<<实时数据上报模式>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
mode	int	离床后实时数据上报模式
0:上报:  1:不上报
		

响应
{
	"status": 0
	"msg": "",
data：null
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
实时数据上报模式获取
注：仅BM8701-2支持，固件版本需≥v6.67
URL
http(s)://domain{:port}/sleepace/realtimeMode/get
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID

响应
{
	"status": 0
	"msg": "",
"data"：{
"mode"：1
}
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	Object	设备配置
mode	int	离床后实时数据上报模式
0:上报:  1:不上报
指示灯状态设置
注：仅BM8701-2支持，固件版本需≥v6.67
URL
http(s)://domain{:port}/sleepace/deviceLightConf/set
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
status:<<指示灯状态>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
status	int	指示灯状态配置
0:开启:  1:关闭
		

响应
{
	"status": 0
	"msg": "",
data：null
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
指示灯状态配置获取
注：仅BM8701-2支持，固件版本需≥v6.67
URL
http(s)://domain{:port}/sleepace/deviceLightConf/get
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID

响应
{
	"status": 0
	"msg": "",
"data"：{
"status"：1
}
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	Object	设备配置
status	int	指示灯状态配置
0:开启:  1:关闭
实时数据上报间隔
注：请将固件升级到5.x 版本以上，否则不支持
4G 设备，或者关注流量的设备，设置上报间隔不要小于30秒，否则100MB/月流量不够。
查询监测设备实时数据上报间隔
http(s)://domain{:port}/sleepace/device/getconfig
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx",
"deviceId":"xxx",
"leftRight":0
}
}

参数说明
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
userid	string	合作方用户的唯一标识
deviceId	string	设备密文id
leftRight	int	左右侧，0左，1右
响应
{
"status": 0,
"msg": null,
"data": {
	"realtimeDataInterval": 10
}
}

字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	Object	设备配置
realtimeDataInterval	int	实时数据上报间隔，单位秒。（修改后报警设置中的频繁体动时长和无体动时长需要修改改为实时数据上报间隔的倍数，如：上报间隔设置为2分钟,则频繁体动时长和无体动时长需要设置2*n分钟=2分钟、4分钟等）
修改监测设备实时数据上报间隔
http(s)://domain{:port}/sleepace/device/updateconfig
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx",
"deviceId":"xxx",
"leftRight":0,
"interval":10
}
}

参数说明
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
userid	string	合作方用户的唯一标识
deviceId	string	设备密文id
leftRight	int	左右侧，0左，1右
interval	int	实时数据上报间隔，单位秒
范围：1~254
响应
{
"status": 0,
"msg": null,
"data": null
}

字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	Object	此接口返回null
设备压力传感器初始化状态
设备压力传感器状态查询
URL
http(s)://domain{:port}/sleepace/pressureSensorStatus
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>，
leftRight:<<左右侧>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备密文
leftRight	int	<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>

响应
{
	"status": 0
	"msg": "",
data:{
pressureSensorStatus:0,
}

}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
pressureSensorStatus	int	0：未初始化
1：已初始化（压力传感器初始化完成的情况下，才能监测到坐起，其次离床速度相对会快些）
2：设备不在线

传感器校准
设备版本需求：BM8701-2 ≥6.70，M901L ≥v1.65，M8701W ≥v8.39
http(s)://domain{:port}/sleepace/device/forceInitSensor
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx",
"deviceId":"xxx",
"leftRight":0
}
}

参数说明
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
userid	string	合作方用户的唯一标识
deviceId	string	设备密文id
leftRight	int	左右侧，0左，1右
响应
{
"status": 0,
"msg": null,
}

字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
离床灵敏度
设备固件版本要求：BM8701-2 ≥v6.70，M901L ≥v1.65，M8701W ≥v8.39
离床灵敏度说明
离床灵敏度	离床检测速度	适用场景
高	3~5秒	适用于需即时响应任何离床动作的场景。
使用者体重较轻、床头抬高、翻身、侧翻至床边等情况会存在概率性造成误判离床。
中	5~8秒	适用于需即时响应任何离床动作的场景。
误判离床概率低于“高”灵敏度。
低	15~20秒	适用于监控或记录离床超时的场景。离床判断准确度高
注：设置为高、中灵敏度时，若实测不如预期，请先查询设备是否已初始化完成，调用接口“查询设备初始化状态”，若未初始化完成，在床上无人且设备安装规范、床为放平状态时，可调用接口“传感器校准”手动校准设备。

查询监测设备离床灵敏度
http(s)://domain{:port}/sleepace/device/getAlgMode
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx",
"deviceId":"xxx",
"leftRight":0
}
}

参数说明
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
userid	string	合作方用户的唯一标识
deviceId	string	设备密文id
leftRight	int	左右侧，0左，1右
响应
{
"status": 0,
"msg": null,
"data": {
	"aloMode": 0
}
}

字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	Object	设备配置
aloMode	int	离床灵敏度设置
0:离床灵敏度低(默认);1:离床灵敏度中:2:离床灵敏度高

修改监测设备离床灵敏度
http(s)://domain{:port}/sleepace/device/updateAlgMode
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx",
"deviceId":"xxx",
"leftRight":0,
"mode":1
}
}

参数说明
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
userid	string	合作方用户的唯一标识
deviceId	string	设备密文id
leftRight	int	左右侧，0左，1右
mode	int	离床灵敏度设置
0:离床灵敏度低(默认);1:离床灵敏度中:2:离床灵敏度高
响应
{
"status": 0,
"msg": null,
"data": null
}

字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	Object	此接口返回null
设备最新固件版本获取
获取服务器已部署最新固件版本信息
URL
http(s)://domain{:port}/sleepace/deviceVersions

参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
“channelId”:<<渠道号>>
}
}
字段	类型	描述
channelId	int	渠道号

响应
{
"status": 0
"msg": "",
"data": [
{
"deviceType": <<device type>>,
"deviceVersion": <<version>>,
"description": <<description>>,
"url": <<download url>>
"fileLen": <<file`s length>>,
"crcBin": <<crcBin>>,
"crcDes": <<crcDes>>,
"lan": <<lang>>
},
....
]
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
deviceType	int	设备类型
deviceVersion	float	设备当前版本
description	string	升级描述，支持多语言
url	string	固件下载地址
fileLen	int	固件长度
crcBin	int	固件校验信息
crcDes	int	固件校验信息
lan	string	“description”使用的语言：


例子
Request
Post http(s)://domain{:port}/sleepace/deviceVersions
Parameter {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:[ {
"deviceType": 1,
"deviceVersion": "3.09",
"description": "Aktualisierung des Inhalts",
"url": "http://download.sleepace.net/sleepace/firmwave/RestOn/TestVersion/Z100_20160704.3.09_Release.des",
"fileLen": 70656,
"crcBin": 797092797,
"crcDes": 4158032917,
"lan": "de"
},
{
"deviceType": 1,
"deviceVersion": "3.09",
"description": " Update content ",
"url": "http://download.sleepace.net/sleepace/firmwave/RestOn/TestVersion/Z100_20160704.3.09_Release.des",
"fileLen": 70656,
"crcBin": 797092797,
"crcDes": 4158032917,
"lan": "en"
}

]
}

Response
{
"status": 0
"msg": ""

}

设备升级
设备升级有两种方式：
1. 自动升级：当设备被绑定或设备生成报告时，Sleepace服务器会判断设备是否为最新版本，如果不是将对设备进行升级；
2. 手动升级：厂商可通过调用设备升级接口对设备进行升级
升级方式可通过调用设置设备自动升级接口设定，默认为自动升级
设置设备升级方式
设置设备升级方式，如果不设置，默认为自动升级
URL
http(s)://domain{:port}/sleepace/setDeviceUpgradeMode
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
upgradeMode:<<升级方式>>
}
}
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
upgradeMode	int	0：自动升级
      升级时机：
a、设备绑定用户时，且设备在线
b、设备生成报告后

1：手动升级

响应
{
"status": 0
"msg": ""
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子
请求
Post http(s)://domain{:port}/sleepace/setDeviceUpgradeMode
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
upgradeMode:<<升级模式>>
}
}

响应
{
"status": 0
"msg": ""

}


查询升级方式
查询渠道下的设备的升级方式
URL
http(s)://domain{:port}/sleepace/getDeviceUpgradeMode
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
}
}
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
响应
{
"status": 0
"msg": ""
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子
请求
Post http(s)://domain{:port}/sleepace/setDeviceAutoUpgrade
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
}
}

响应
{
"status": 0
"msg": "",
“data”:{
“upgradeMode‘’:0
}

}

设备固件上传
上传设备固件 
URL
http(s)://domain{:port}/sleepace/firmware/uploadFile
参数

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
file	File	固件压缩包

响应
{
	"status": 0
	"msg": ""
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子

设备固件删除
调用设备升级接口成功后，可通过升级进度查看进度
URL
http(s)://domain{:port}/sleepace/firmware/delete
参数
{
token:{
“appId:’’,<<与消息队列的账号相同>>
“secureKey:’’,<<与消息队列的密码相同>>
},
data:{
“deviceType”:<<设备类型>>，
“deviceVersion”:<<设备固件版本>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceType	int	设备类型
deviceVersion	string	需要升级到的固件版本

响应
{
	"status": 0
	"msg": ""
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子

设备升级
调用设备升级接口成功后，可通过设备升级进度接口查看进度
URL
http(s)://domain{:port}/sleepace/upgrade/device
参数
{
token:{
“appId:’’,<<与消息队列的账号相同>>
“secureKey:’’,<<与消息队列的密码相同>>
},
data:{
“deviceId”:<<设备ID>>，
“deviceVerison”:<<设备固件版本>>
}
}
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	需要升级设备
deviceVerison	string	需要升级到的固件版本

响应
{
"status": 0
"msg": ""
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子
请求
Post http(s)://domain{:port}/sleepace/upgrade/device
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
Data{
“deviceId”:<<设备ID>>，
“deviceVerison”:<<设备固件版本>>
}
}

响应
{
"status": 0
"msg": ""

}

历史数据存储方式（BM8701-2专用）
历史数据存储方式设置
适用于BM8701-2，固件版本≥ 6.60
用途：设置设备生成报告的方式：
1）每天定时上传报告（设备出厂默认此项）：在不需要及时查看报告，想要查看整日作息情况的使用场景，如养老机构、疗养院，建议此方式
2）离床1小时后自动生成报告或调用“结束监测”接口生成报告：在需要及时查看报告，如康养酒店，顾客起床或者退住后就需要生成报告，建议此方式
请注意，设置后，在设置之前产生的数据将会清空！！！
URL
http(s)://domain{:port}/sleepace/reportUploadType/set
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
reportUploadType:<<历史数据存储方式设置>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
reportUploadType	int	0:24小时;1:自动结束（离床1小时后自动生成）

响应
{
	"status": 0
	"msg": "",
data：null
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
 历史数据存储方式获取
URL
http(s)://domain{:port}/sleepace/reportUploadType/get
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>
}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID

响应
{
	"status": 0
	"msg": "",
"data"：{
"reportUploadType"：1
}
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
data	Object	设备配置
reportUploadType	int	0:24小时;1:自动结束（离床1小时后自动生成）
报告
结束监测（BM8701-2专用）
仅支持在“历史数据存储方式”为离床1小时后自动生成报告
URL
http(s)://domain{:port}/sleepace/stopMonitor
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
leftRight:<<单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准>>

}
}

字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
leftRight	int	单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准

响应
{
	"status": 0
	"msg": ""
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子
请求
Post   http(s)://domain{:port}/sleepace/stopMonitor
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>
}
}

响应
{
	"status": 0
	"msg": ""

}
设置设备生成报告时间
设置后，设备将会在设置的时间点后15分钟内生成报告。
备注：
BM8701-2：
1）固件版本≥v2.01 才能支持，请先调用固件升级接口升级到最新的固件版本
2）固件版本≥v6.39 仅支持在“历史数据存储方式”为每天定时上传报告模式时使用此接口
URL
http(s)://domain{:port}/sleepace/setReportUploadTime
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
leftRight:<<左右侧>>,
reportUploadTime:<<设备报告上传时间，请设置整点，否则无法生效>>
}
}
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
leftRight	int	左右侧0左侧 1 右侧（不区分左右侧设备该参数非必填）
reportUploadTime	int	报告上传时间，请设置整点（1~24），否则无法生效

响应up
{
"status": 0
"msg": ""
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子
请求
Post http(s)://domain{:port}/sleepace/ setReportUploadTime
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
reportUploadTime:<<设备报告上传时间>> 
}
}

响应
{
"status": 0
"msg": ""

}


查询设备的报告生成时间
设备将会在设置的时间点后15分钟内生成报告
URL
http(s)://domain{:port}/sleepace/getReportUploadTime
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
leftRight:<<左右侧>>
}
}
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
deviceId	string	设备加密ID
leftRight	int	左右侧0左侧 1 右侧（不区分左右侧设备该参数非必填）
reportUploadTime	int	报告上传时间
响应
{
"status": 0
"msg": ""
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因


例子
请求
Post http(s)://domain{:port}/sleepace/ getReportUploadTime
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
reportUploadTime:<<设备报告上传时间>> 
}
}

响应
{
"status": 0
"msg": "",
“data”:{
“reportUploadTime”:10
}
}

报告查询(旧接口，针对24小时报告的设备,无最长一段睡眠分析报告)
查询指定时间范围内产生的历史报告
注：本接口为2.x对接方案的旧接口，2022/7/5之前对接的可以沿用此接口，但是已产生的历史数据需要重新调用接口跑数据，否则将会丢失；2022/7/5 新对接的厂商请不用使用本接口
URL
http(s)://domain{:port}/sleepace/get24HourDailyReport
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
userId:<<用户ID>>,
startTime:<<开始时间>>,
endTime:<<结束时间>>
}
}
字段	类型	描述
appId	string	与消息队列的账号相同
secureKey	string	与消息队列的密码相同
userId	string	用户ID
startTime	int	开始时间
endTime	int	结束时间
startTime和endTime如何设置？
只要报告在查询时间的区间内，都会返回。比如,下图有report1、report2、report3  三份报告，在startTime和endTime查询范围内，report2和report3都会返回 

响应
{
"status": 0
"msg": "",
“data”:[{
summary:{<<概要数据>>},
detail:{<<监测详细数据>>},
analysis:{<<分析结果>>}
},{
summary:{<<概要数据>>},
detail:{<<监测详细数据>>},
analysis:{<<分析结果>>}
},
...
...
]
}
字段	类型	描述
status	int	状态码，0表示成功，其他失败，详见状态码
msg	string	失败原因
summary	obj	概要数据
detail	obj	详细数据
analysis	obj	分析结果

Summary
字段	类型	描述
recordCount	int	监测数据长度
startTime	int	开始时间，单位，秒
stopMode	int	停止检测方式：0正常停止，APP发送停止采集，1离床一小时,自动停止,2异常停止(a监测超过24小时，b、reston关机 c、升级),3设备异常重启,错误停止
timeStep 	int	记录间隔时间 (默认60s，即：1分钟一个时间一个点)
timezone	int	时区
Detail
字段	类型	描述
breathRate	int[]	呼吸率
heartRate	int[]	心率
status	int[]	状态标识
statusValue	int[]	状态值
eHumidity	int[]	环境湿度（BM8701-2 有此数据，不建议使用，因为准确度不高）
eTemp	int[]	环境温度（BM8701-2 有此数据，不建议使用，因为准确度不高）
Analysis
字段	类型	描述
avgBreathRate	int	平均呼吸率
avgHeartRate	int	平均心率
startTime	int	报告开始时间
duration	int	睡眠时间
wake	int	清醒时间
lightSleepDuration	int	浅睡时间
deepSleepDuration	int	深睡时间
outOfBedDuration	int	不在床时间
sleepStateStr	byte[]	分析出来的睡眠状态（0:设备未在监测，1、离床，2、清醒，3、深睡，4、浅睡，6、坐起
）

例子：
{
"status": 0,
"msg": null,
"data": [
{
"summary": {
"userId": 401563,
"deviceId": "ch3a9o5ujtwqk",
"startTime": 1634547600,
"timezone": 28800,
"timeStep": 60,
"dstOff": 0,
"recordCount": 1440,
"stopMode": 0,
"plat": "wifi_49",
"source": 49,
"arithmeticVer": "1.0.1",
"firmwareVers":"02.03.56",
"originStartTime": 1634547600,
"updateTime": 1634634022000,
"dataMd5": "",
"flaginvalid": 0
},
"detail": {
"userId": 401563,
"startTime": 1634547600,
"breathRate": "[16,11,13,0,0,0,0,0,0,0,0,0,0,0,0,13,15,12,12,13,17,16,15,13,13,13,15,14,15,14,12,16,13,16,14,12,12,13,16,17,16,14,14,14,14,16,14,14,13,12,11,14,15,16,12,16,16,13,11,14,13,16,16,16,12,12,11,15,15,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,14,11,14,17,16,17,16,16,15,17,15,15,16,14,14,12,16,16,12,15,13,16,18,15,0,0,0,12,14,12,16,16,16,16,15,16,13,12,12,14,0,0,0,0,0,0,0,0,0,0,16,0,0,13,18,14,14,18,15,13,14,14,17,18,14,17,18,18,17,18,18,18,14,18,16,16,17,17,16,18,18,17,17,18,15,16,17,17,18,17,12,15,14,17,17,16,16,17,16,17,18,15,13,17,14,18,17,18,18,15,15,14,19,15,14,18,13,14,13,12,18,14,14,10,11,13,14,16,15,12,14,11,14,15,13,12,13,12,13,12,10,15,13,14,15,14,13,14,16,16,13,14,14,14,16,14,13,14,14,14,12,12,18,12,11,13,12,10,11,12,10,13,13,18,15,14,15,16,14,17,12,17,17,14,11,12,13,12,13,14,12,16,15,11,13,13,11,18,18,14,15,16,15,0,0,17,14,13,16,15,13,15,13,20,16,15,13,13,12,15,16,19,17,13,14,16,15,12,13,0,15,12,12,12,14,14,14,11,11,12,13,14,14,12,13,12,13,12,11,14,10,16,14,11,15,13,11,10,11,12,13,0,0,0,0,0,0,15,11,12,12,15,13,13,14,14,10,16,14,12,13,12,13,13,11,0,0,0,0,18,0,0,12,12,15,12,12,12,15,11,12,15,12,12,12,13,13,13,14,15,12,16,12,13,0,0,0,0,13,13,12,14,17,17,13,11,12,13,14,15,15,10,13,11,11,14,11,13,12,13,9,13,13,16,15,12,11,14,-1]",
"heartRate": "[76,78,74,0,0,0,0,0,0,0,0,0,0,0,0,0,81,81,80,79,81,83,81,83,79,82,81,84,82,81,83,83,81,81,85,84,82,85,83,83,79,82,85,85,79,80,82,80,82,82,82,82,81,83,77,81,83,79,73,77,78,81,79,78,84,79,79,78,79,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,82,78,79,76,78,84,82,84,81,84,80,82,82,82,80,82,81,80,85,84,84,84,84,85,0,0,0,78,85,87,82,82,88,85,85,83,86,83,85,82,0,0,0,0,0,0,0,0,0,0,51,0,0,0,0,0,0,0,0,0,0,60,62,61,55,53,58,60,56,54,55,55,59,56,57,57,60,60,60,59,58,58,60,61,56,62,57,52,57,63,63,58,60,57,56,59,58,57,58,55,52,57,66,59,52,58,64,60,62,67,60,57,58,56,60,54,56,56,59,58,57,60,61,59,57,61,58,59,60,60,59,60,58,63,67,60,63,65,59,61,63,61,63,59,57,57,56,56,55,57,59,62,63,61,57,55,54,55,56,58,60,61,63,65,62,62,62,58,61,63,64,65,66,63,63,65,66,68,66,64,63,62,58,55,61,58,56,57,60,48,57,57,59,61,64,64,56,55,57,57,58,59,57,0,0,49,79,83,87,86,88,86,86,87,85,86,84,79,82,77,84,90,86,82,82,84,82,84,85,0,93,49,49,65,61,61,64,61,57,61,62,62,61,54,60,65,66,58,59,61,61,60,65,65,58,60,60,63,64,63,63,0,0,0,0,0,0,88,88,86,90,87,89,85,87,83,74,80,82,86,84,79,80,81,79,0,0,0,0,86,0,0,0,0,0,0,88,80,88,78,85,87,88,85,88,86,83,82,86,85,85,87,86,83,0,0,0,0,81,80,82,82,86,82,82,82,82,84,85,88,88,81,80,80,78,77,77,82,78,77,77,84,80,85,82,80,78,80,-1]",
"status": "[0,0,4,5,5,5,5,5,5,5,5,5,5,5,5,0,4,4,0,0,0,0,0,0,4,0,0,4,4,4,4,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,4,0,4,0,0,0,0,0,0,0,0,0,0,0,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,5,4,0,0,0,4,0,0,0,0,0,0,0,0,0,0,4,4,0,0,4,4,0,0,0,5,5,5,4,4,0,0,0,0,4,0,0,0,0,0,0,5,5,5,5,5,5,5,5,5,5,4,5,5,4,4,4,4,4,4,4,4,4,4,4,0,0,4,0,0,4,0,0,0,4,0,0,4,0,0,4,0,0,0,4,0,0,0,0,0,4,0,0,0,0,0,0,0,0,0,0,0,0,0,4,4,0,0,0,0,4,0,0,0,0,4,4,4,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,4,0,0,0,4,0,4,0,0,0,0,0,0,0,0,0,0,4,4,4,0,0,0,0,0,0,0,0,0,0,0,4,0,0,0,0,0,0,4,4,0,0,0,0,0,0,0,0,0,0,0,0,4,0,0,0,0,0,0,4,0,0,4,0,4,0,0,0,0,0,5,5,4,0,0,0,0,0,0,0,4,0,0,0,0,0,0,0,0,0,0,4,0,0,0,4,5,4,4,0,4,4,0,4,0,4,0,0,0,4,4,4,0,4,0,0,0,0,0,0,0,0,0,4,0,0,4,0,5,5,5,5,5,5,0,4,0,0,0,4,4,0,4,0,4,4,0,0,4,0,4,0,5,5,5,5,0,5,5,0,0,0,0,0,4,4,0,0,0,0,0,0,0,0,0,0,0,0,4,0,0,5,5,5,5,4,0,0,0,0,0,0,0,0,0,0,0,4,4,0,0,0,0,0,0,0,0,0,0,0,0,4,0,0,0,0]",
"statusValue": "[0,0,1,31,60,60,60,47,46,60,60,60,60,60,35,0,1,1,0,0,0,0,0,0,1,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,34,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,60,33,54,60,60,36,1,0,0,0,2,0,0,0,0,0,0,0,0,0,0,2,2,0,0,1,1,0,0,0,7,60,28,1,1,0,0,0,0,1,0,0,0,0,0,0,17,60,60,60,60,60,60,60,60,49,4,50,55,6,7,8,9,3,7,8,7,1,1,1,0,0,1,0,0,1,0,0,0,2,0,0,1,0,0,1,0,0,0,1,0,0,0,0,0,3,0,0,0,0,0,0,0,0,0,0,0,0,0,2,3,0,0,0,0,1,0,0,0,0,1,3,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0,0,2,1,1,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,1,0,0,2,0,1,0,0,0,0,0,32,49,3,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,2,0,0,0,1,36,1,1,0,2,1,0,1,0,1,0,0,0,2,1,2,0,1,0,0,0,0,0,0,0,0,0,1,0,0,1,0,25,60,60,60,60,17,0,1,0,0,0,3,1,0,1,0,1,1,0,0,1,0,1,0,25,60,60,45,0,2,21,0,0,0,0,0,2,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,29,60,60,3,1,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,-1]",
"eTemp": "[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]",
"eHumidity": "[64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,65,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,64,65,65,64,64,64,64,64,64,64,64,64,64,65,65,65,64,64,64,64,64,65,65,65,65,64,65,65,65,64,65,65,65,65,65,65,65,64,65,65,65,65,65,65,65,65,65,65,65,65,65,64,65,65,65,65,65,65,65,65,65,65,65,65,65,65,65,65,65,65,65,65,65,65,66,65,65,65,65,65,65,65,65,67,66,67,67,67,66,67,67,67,67,67,67,67,67,67,66,67,67,67,67,67,67,67,67,67,67,67,67,67,67,67,67,67,66,67,67,67,67,67,66,67,67,67,68,67,67,67,67,67,67,67,68,67,67,67,67,67,67,67,67,67,67,67,67,67,67,67,67,67,66,67,68,67,67,67,67,67,67,67,67,67,67,67,67,67,66,67,67,68,68,62,62,62,61,61,62,62,62,61,61,61,61,61,61,61,61,61,62,61,62,62,62,61,62,62,62,62,62,62,61,62,62,62,62,62,62,61,61,61,62,62,61,62,61,62,62,62,62,61,62,62,62,62,62,61,62,61,62,62,61,62,61,62,62,61,61,61,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,61,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,61,62,62,62,62,62,62,62,62,62,62,62,62,63,62,62,62,62,62,62,62,62,62,62,62,63,62,62,62,62,62,62,62,63,62,63,62,63,62,63,63,63,62,62,63,63,62,63,62,63,63,63,62,62,64,63,63,63,63,63,63,63,62,62,63,63,63,62,63,62,62,62,62,65,65,65,65,65,64,65,64,64,65,63,65,65,65,64,64,65,64,64,65,65,64,65,64,64,64,64,64,64,65,65,65,65,64,65,64,64,64,65,65,65,65,64,65,64,65,64,65,64,64,63,65,65,65,65,65,65,64,64,64,64,64,65,65,65,65,64,64,64,65,65,65,65,65,65,64,64,64,65,65,65,65,65,64,64,65,65,65,65,65,63,63,63,62,62,62,62,63,63,63,62,61,63,62,63,63,63,63,62,63,63,62,62,62,63,63,63,63,62,63,62,62,62,63,63,63,63,63,62,61,62,62,62,63,63,63,63,63,63,63,63,62,62,61,63,63,63,63,63,63,62,63,63,63,62,63,63,63,62,62,63,62,62,62,63,62,63,63,63,63,63,63,63,62,63,63,63,63,63,63,63,63,63,63,63,63,62,63,63,62,63,63,62,63,63,62,63,63,63,63,62,63,60,62,63,62,62,63,63,62,62,63,63,63,63,63,63,62,62,63,63,62,62,63,62,60,63,63,63,63,62,63,63,63,63,63,63,63,62,62,62,63,63,63,63,63,63,62,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,62,62,63,63,63,62,63,63,63,62,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,61,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,60,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,62,63,63,63,63,63,63,63,63,61,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,64,64,63,63,63,63,63,63,64,63,63,63,63,63,63,64,63,63,63,64,63,63,64,63,63,63,63,64,63,64,63,63,63,63,63,64,64,64,64,63,64,64,63,63,63,63,64,64,63,64,64,64,64,64,64,64,63,64,64,64,64,64,64,64,63,64,64,64,64,63,64,63,63,64,63,63,63,63,63,64,63,63,63,64,63,63,64,63,63,64,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,62,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,63,62,63,62,63,63,62,62,62,63,63,63,62,62,62,62,63,62,62,62,62,62,62,63,62,62,62,62,62,62,62,62,62,62,62,61,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,60,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,61,62,62,61,62,62,61,62,61,61,61,61,62,62,62,62,61,61,61,61,61,61,61,61,61,63,63,63,63,63,63,63,62,63,63,63,63,63,64,63,63,64,63,63,64,63,63,62,64,64,64,64,64,64,64,64,63,64,63,64,63,64,64,63,63,64,64,64,64,63,64,64,64,64,64,64,63,64,64,63,63,63,63,63,64,64,63,64,64,64,64,46,46,46,46,46,46,46,46,47,47,47,47,47,47,47,47,47,47,48,48,48,48,48,48,48,48,48,48,49,49,49,49,49,49,49,49,49,49,51,51,51,51,51,51,51,51,51,51,51,51,51,51,51,51,51,50,50,51,50,50,51,51,51,50,50,50,50,50,50,50,50,51,50,50,50,50,51,51,51,51,50,51,51,51,51,51,51,51,51,51,50,50,51,51,50,51,50,50,50,51,50,51,50,51,50,51,51,50,51,51,51,50,50,50,50,50,51,51,51,50,50,50,50,50,51,51,51,50,51,53,53,53,52,52,53,52,53,53,53,53,53,53,52,52,53,53,53,53,52,53,53,53,52,53,52,53,53,53,53,-1]",
"eLight": null,
"eCO2": null,
"eNoise": null,
"mTemp": null,
"mHumidity": null
},
"analysis": {
"startTime": 1634547600,
"duration": 1439,
"wake": 1068,
"outOfBedDuration": 1,
"avgHeartRate": 71.0,
"avgBreathRate": 14.0,
"sleepStateStr": "[3,2,2,1,1,1,1,1,1,1,1,1,1,1,1,2,2,2,2,2,2,2,2,2,2,2,3,3,3,3,3,2,2,3,3,3,3,3,3,3,3,3,3,2,2,2,2,2,2,2,3,3,3,2,2,2,2,2,2,2,2,2,2,3,3,2,2,2,2,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,2,2,2,2,2,2,2,2,2,2,3,3,3,3,3,2,2,2,2,2,2,2,2,2,1,1,1,2,2,2,2,2,2,2,3,3,3,3,3,3,1,1,1,1,1,1,1,1,1,1,2,1,1,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,2,2,3,3,3,3,3,2,2,3,3,3,3,3,3,3,3,3,2,2,2,2,2,2,2,2,3,2,2,2,2,2,2,2,2,2,2,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,2,2,3,3,3,3,3,3,3,3,3,2,2,2,2,3,3,3,3,3,3,3,3,3,3,3,2,2,3,3,2,2,2,2,2,2,2,2,3,3,3,3,2,2,3,3,3,3,3,2,2,2,2,2,2,2,2,2,2,2,2,2,2,1,1,2,2,2,2,2,2,2,2,2,2,3,3,3,3,3,2,2,2,3,2,2,2,2,2,1,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,3,3,3,3,3,3,3,3,3,1,1,1,1,1,1,2,2,2,2,2,2,2,2,2,2,2,2,2,2,3,3,3,3,1,1,1,1,2,1,1,2,2,2,2,2,2,2,2,2,2,2,3,3,3,3,3,3,3,3,2,2,2,1,1,1,1,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,0]",
"lightSleepDuration": 231,
"deepSleepDuration": 140
}
}
]
}
报告查询(新接口，24小时+最长一段睡眠分析报告)
查询指定时间范围内产生的历史报告
URL
http(s)://domain{:port}/sleepace/get24HourDailyWithMaxReport
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
userId:<<用户ID>>,
startTime:<<开始时间>>,
endTime:<<结束时间>>
}
}
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
userId
string
用户ID
startTime
int
开始时间,unix time
endTime
int
结束时间,unix time
startTime和endTime如何设置？
只要报告在查询时间的区间内，都会返回。如下图有report1、report2、report3  三份报告，在startTime和endTime查询范围内，report1、2、3都会返回。例：有3份报告，分别是 7/1 10:00~7/2 10:00、7/2 10:00~7/3 10:00、7/3 10:00~7/4 10:00，查询时间为 7/2 8:00~7/4 8:00，那么这3份报告都会返回。
响应
{
"status": 0
"msg": "",
“data”:[{
summary:{<<概要数据>>},
detail:{<<监测详细数据>>},
analysis:{<<分析结果>>}
},{
summary:{<<概要数据>>},
detail:{<<监测详细数据>>},
analysis:{<<分析结果>>}
},
...
...
]
}
字段
类型
描述
status
int
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
summary
obj
概要数据
detail
obj
详细数据
analysis
obj
分析结果
Summary
字段
类型
描述
recordCount
int
监测数据长度
startTime
int
开始时间，单位，秒
stopMode
int
停止检测方式：0正常停止，APP发送停止采集，1离床一小时,自动停止,2异常停止(a监测超过24小时，b、reston关机 c、升级),3设备异常重启,错误停止
timeStep 
int
记录间隔时间 (默认60s，即：1分钟一个时间一个点)
timezone
int
时区
Detail
字段
类型
描述
breathRate
int[]
呼吸率
heartRate
int[]
心率
status
int[]
状态标识
0x00 离床
0x01 在床
0x11 呼吸暂停
0x22 体动
0x30 躺着
0x31 坐起
statusValue
int[]
 status==0x00 的情况下，statusValue==x，表示这一分钟内离床时长为 x 秒，若x==62,则表示这一分钟为未监测状态；
status==0x01 的情况下，statusValue == 0，是固定值，无意义
status==0x11 的情况下，statusValue== x，表示这一分钟内呼吸暂停的时长为 x 秒；
status==0x22 的情况下，statusValue==x,表示这一分钟内体动的次数为 x 次  
eHumidity
int[]
环境湿度（BM8701-2 有此数据）
eTemp
int[]
环境温度（BM8701-2 有此数据）
signalQuality
int[]
信号质量，仅对特定设备开发
Analysis
字段
类型
描述
avgBreathRate
int
平均呼吸率
avgHeartRate
int
平均心率
startTime
int
报告开始时间
duration
int
睡眠时长（分钟）
wake
int
清醒时长（分钟）
lightSleepDuration
int
浅睡时长（分钟）
deepSleepDuration
int
深睡时长（分钟）
outOfBedDuration
int
不在床时长（分钟）
sleepStateStr
byte[]
分析出来的睡眠状态（0:设备未在监测，1、离床，2、清醒，3、深睡，4、浅睡，6、坐起
）
sleepArray
List<SleepEvent>
睡眠分段,记录每段睡眠的开始时间及持续时间,详见SleepEvent
分段逻辑：
将24小时的数据分成几段，划分规则：
1.离床或设备未监测间隔≥1小时
2.＜10分钟的数据不统计
maxReport
MaxReport
分析得到的最长一段睡眠报告,详见MaxReport,注意：如果maxReport中的scale(睡眠评分)=0时，报告为短报告。
SleepEvent
字段
类型
描述
startTime
Int
事件开始时间戳
duration
Int
事件持续时长(分钟)
MaxReport
字段
类型
描述
inBedTime
int
第一帧在床的时刻
getupTime
int
最末帧在床的时刻
onsetTime
int
睡着时刻,第一个入睡的时间点
wakeupTime
int
醒来时刻,最后一个醒来的时间点
tsTime
int
睡眠时长,最后一个醒来时刻 - 第一个入睡时刻
单位：分钟
参考评价：
＜5小时：偏少
5≤睡眠时长≤9 小时：正常
＞9小时：偏多
asleepDur
int
睡着时长=浅睡时长+深睡时长
单位：分钟
seIndex
int
睡眠效率=睡着时长/总记录时长x100%
参考评价：
≥80%：优秀
≥70%，<80%：正常
<70%：偏少
awakeDur
int
清醒时长,单位：分钟
awakePct
int
清醒占比,清醒时长/睡眠时长x100%
lightDur
int
浅睡时长,单位：分钟
lightPct
int
浅睡占比,浅睡时长/睡眠时长x100%
deepDur
int
深睡时长,单位：分钟
deepPct
int
深睡比例,深睡时长/睡眠时长x100%
参考评价：
≥20%：优秀
≥15%，<20%：正常
<15%：偏少
offBedDur
int
离床时长,单位：分钟
offBedPct
int
离床占比,离床时长/睡眠时长x100%
discDur
int
未监测时长,单位：分钟
discPct
int
未监测占比,未监测时长/睡眠时长x100%
awakeCnt
int
清醒次数
awakeEvent
List<SleepEvent>
清醒事件统计,发生时间点、持续时长
注：离床前后的清醒统计为1次
offBedCnt
int
离床次数
注：间隔小于5分钟的2次离床统计为1次
leftBedEvent
List<SleepEvent>
离床事件统计,发生时间点、持续时长
bmCnt
int
体动次数
hrMean
int
平均心率,单位：次/分
hrMin
int
心率最低值,单位：次/分
hrMax
int
心率最高值,单位：次/分
hrHiDur
int
累计心率过速时长,心率≥120次/分 时长总和
单位：分钟
hrHiPct
int
累计心率过速占比,心率过速时长/睡眠时长x100%
hrLoDur
int
累计心率过缓时长,心率≤45次/分 时长总和
单位：分钟
hrLoPct
int
累计心率过缓占比,心率过缓时长/睡眠时长x100%
brMean
int
平均呼吸率,每分钟呼吸率平均值
brMin
int
呼吸率最低值
brMax
int
呼吸率最高值
brHiDur
int
累计呼吸率过速时长,呼吸率≥26次/分 时长总和
单位：分钟
brHiPct
int
累计呼吸率过速占比,呼吸率过速时长/睡眠时长x100%
brLoDur
int
累计呼吸率过缓时长,呼吸率≤12次/分 时长总和
单位：分钟
brLoPct
int
累计呼吸率过缓占比,呼吸率过缓时长/睡眠时长x100%
ahiDur
int
累计呼吸暂停时长,全部呼吸暂停时长之和
单位：秒
注：统计的是睡着（浅睡、深睡）状态下数据
ahiCnt
int
累计呼吸暂停次数
注：统计的是睡着（浅睡、深睡）状态下数据
ahiMaxDur
int
最长呼吸暂停时长,呼吸暂停时长的最大值
单位：秒
注：统计的是睡着（浅睡、深睡）状态下数据
csaDur
int
中枢性呼吸暂停时长之和
单位：秒
注：统计的是睡着（浅睡、深睡）状态下数据
csaCnt
int
中枢性呼吸暂停次数
注：统计的是睡着（浅睡、深睡）状态下数据
osaDur
int
阻塞性呼吸暂停或低通气时长之和
单位：秒
注：统计的是睡着（浅睡、深睡）状态下数据
osaCnt
int
阻塞性呼吸暂停或低通气次数
注：统计的是睡着（浅睡、深睡）状态下数据
bmIndex
int
体动指数，统计入睡后平均每小时的体动次数
单位：次/小时
参考评价：
≤5：优秀
＞5，≤15：正常
＞15：偏多
contIndex
int
睡眠连续性指数，统计入睡后平均每小时的清醒次数（仅统计清醒时长大于5分钟的次数，且离床前后算作一次清醒）
单位：次/小时
前端显示时，需除以100； 举例：服务器拿到的该字段值为40，则实际的睡眠连续性指数应为40/100 = 0.4次/小时
参考评价：
≤0.4：优秀
＞0.4，≤1.25：正常
＞1.25：偏多
ahIndex
int
呼吸暂停/低通气风险（AHI值）,睡眠期内呼吸风险指数
统计睡眠期间每小时出现呼吸暂停/低通气的平均次数
参考评价：
＜5：无风险
≥5，<15：低风险
≥15，<30：中风险
≥30：高风险
scale
int
睡眠评分（满分100分，扣分制）
markLong
int
睡眠总时间过长扣分
markShort
int
睡眠总时间过短扣分
markContinuity
int
睡眠连续性扣分
markEfficiency
int
睡眠效率扣分
markDeep
int
深睡比例扣分
markAHI
int
睡眠呼吸风险扣分
markBM
int
体动指数扣分
osaMaxDur
int
睡眠期内最长阻塞性呼吸暂停或低通气持续时长，单位：秒
hrvIndex
int
房颤(atrial fibrillation)风险等级指数,{0:未知(hrv数目过少，不具统计意义);1:无风险(hrvPct < 30%);2:低风险(30% <= hrvPct < 50% );3:中风险(50% <= hrvPct < 70% );4:高风险(70% <= hrvPct < 100% )}
hrvPct
int
睡眠期间相邻心跳间隔时长大于120ms的比例
hrvRmssd
int
心率变异性的RMSSD值，单位：毫秒
avgHrv
int
近30天hrv均值
hrvStress
int
压力指数,{0x00:压力过载;0x01:注意压力;0x02:状态正常;0x03:状态优秀}
hrvStart
int
hrv开始时间(时刻，单位分钟，如980 = 16*60+20，即开始时间为16:20)
hrvEnd
int
hrv结束时间(时刻，单位分钟，如1070 = 17*60+50，即开始时间为17:50)
hrvArray
short[]
每5分钟一次分析得出的Hrv值数组
snoreAnalysis
SnoreAnalysis
鼾声分析结果（仅电动床SDC100支持）
SnoreAnalysis
sn_index
int
鼾声程度，{0:无；1:轻度；2:中度；3:严重}
snore_number
int
鼾声总次数
snore_duration
int
鼾声总时长(单位:分钟)
snore_high_proportion
int
鼾声程度为高的比例
snore_mid_proportion
int
鼾声程度为中的比例
snore_low_proportion
int
鼾声程度为低的比例
snore_non_proportion
int
鼾声程度为无的比例
snore_level
String
用于返回鼾声分布状态，{0x00:无鼾声,0x01:鼾声强度低,0x02:鼾声强度中,0x03:鼾声强度高}
例子：
{
    "status":0,
    "msg":null,
    "data":[
        {
            "summary":{
                "userId":402000,
                "deviceId":"6kf53xjpnfrtk",
                "startTime":1643515200,
                "timezone":28800,
                "timeStep":60,
                "dstOff":0,
                "recordCount":1440,
                "stopMode":0,
                "plat":"wifi_wifi_49",
                "source":49,
                "arithmeticVer":"2.0.13",
                "firmwareVers":"02.03.42\u0000\u0000\u0000\u0000",
                "originStartTime":1643515200,
                "updateTime":1656489820000,
                "dataMd5":"",
                "flaginvalid":0
            },
            "detail":{
                "userId":402000,
                "startTime":1643515200,
                "breathRate":"[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,12,11,13,11,16,14,14,14,13,13,13,11,12,12,13,16,13,13,11,15,12,14,12,12,14,13,12,12,13,12,11,13,11,11,11,12,13,12,12,13,12,13,11,18,15,13,14,13,13,16,12,12,13,14,14,13,13,12,11,11,14,14,15,15,14,11,11,10,11,13,12,13,11,13,12,12,11,11,12,14,13,11,13,12,13,15,12,10,11,12,13,11,12,14,13,12,14,14,13,15,16,16,13,12,14,14,14,13,13,14,14,14,14,14,0,0,11,10,12,13,13,12,18,13,12,11,12,12,10,11,12,13,13,14,16,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,12,12,13,13,15,14,14,15,15,15,15,15,15,15,15,15,15,16,16,16,16,14,12,15,16,16,14,13,13,14,15,15,15,14,15,15,14,15,14,15,15,15,15,15,16,16,16,15,14,13,14,14,14,15,15,14,14,15,15,14,15,12,12,12,14,15,14,16,15,15,15,14,14,15,15,15,15,15,13,14,14,13,13,15,13,15,12,13,11,12,14,11,12,14,10,12,13,11,13,13,11,11,12,13,12,14,14,15,11,13,11,10,13,11,10,13,13,11,14,14,12,11,10,12,14,14,14,12,12,11,14,11,12,12,12,13,14,13,15,13,12,12,12,12,14,12,13,12,13,12,15,12,12,13,13,16,14,15,15,16,15,13,12,11,12,13,15,12,13,13,11,11,11,12,13,14,15,12,16,15,16,15,15,16,15,16,16,16,16,16,16,15,15,12,12,11,13,13,14,13,14,15,15,13,15,13,15,15,16,15,15,15,15,16,16,16,13,14,15,11,12,11,12,14,15,16,14,11,13,11,13,13,11,13,15,15,15,15,15,15,15,16,15,15,15,15,15,15,12,11,11,11,13,14,12,12,14,14,13,13,14,13,13,15,15,14,15,15,15,14,14,15,16,16,16,15,17,17,16,17,16,12,13,12,12,16,14,14,14,14,14,15,16,15,16,16,14,16,15,14,14,13,14,15,15,14,16,16,14,14,15,14,15,13,14,17,14,15,13,16,14,11,12,12,14,15,12,13,11,12,12,12,11,12,15,14,16,12,11,10,17,16,10,9,11,10,13,13,14,10,13,13,11,13,11,13,13,15,15,15,11,12,12,11,16,16,14,11,12,13,15,15,12,13,16,14,12,14,14,14,14,11,13,13,12,12,15,16,16,15,15,15,15,15,15,15,15,15,16,15,16,16,15,17,16,14,16,16,15,16,16,16,15,16,15,15,15,16,16,16,16,16,16,15,15,15,14,13,12,13,13,11,14,12,14,14,14,11,14,12,12,12,11,16,15,14,12,12,12,11,13,16,11,11,11,11,11,12,14,12,12,13,13,15,13,13,12,14,14,14,15,15,14,15,11,11,11,12,14,15,13,13,13,15,13,12,11,14,14,14,16,15,16,12,12,14,13,12,13,11,12,11,12,13,12,11,16,15,12,15,12,12,12,16,15,15,15,15,15,14,15,15,15,11,11,13,11,11,12,12,11,12,14,10,16,16,15,14,12,10,13,11,16,14,15,16,15,15,12,14,15,14,12,11,12,11,12,11,12,12,12,13,13,11,10,10,12,18,16,13,14,12,13,13,14,12,11,13,12,13,12,14,14,12,15,14,12,13,16,12,12,13,14,12,11,12,15,16,16,16,14,16,15,17,16,16,16,17,15,16,16,16,17,15,16,16,16,15,15,15,15,16,16,15,17,16,15,16,16,16,16,16,13,16,16,16,16,16,16,15,16,15,15,15,16,14,15,15,15,16,15,15,16,16,13,15,15,13,15,12,12,12,11,12,12,13,14,12,14,11,16,14,15,12,15,14,13,11,11,11,14,12,12,16,14,13,11,12,14,11,11,14,12,13,13,17,12,10,15,10,12,14,13,10,14,15,12,14,12,14,14,14,11,11,12,12,11,20,14,10,12,12,15,15,16,16,15,16,17,17,15,16,16,16,15,13,12,11,12,16,10,9,12,11,12,15,13,12,12,12,12,12,13,12,15,15,15,15,15,16,15,15,15,15,16,15,16,15,15,13,15,15,15,14,16,16,15,16,15,15,17,16,16,16,16,16,17,17,16,16,15,15,16,15,14,15,15,13,14,13,15,16,16,16,16,16,12,13,14,14,14,15,15,15,16,15,15,14,15,14,13,13,15,14,12,11,12,15,13,13,14,12,12,13,13,12,12,12,16,0,0,0,0,0,10,12,12,15,13,14,11,14,15,14,14,15,11,13,13,15,14,15,12,15,12,13,12,14,11,11,12,10,11,13,13,12,14,15,14,13,13,13,11,12,14,12,10,14,12,12,16,11,13,10,12,12,14,14,11,14,15,13,13,12,13,13,11,15,13,12,12,13,12,12,13,10,11,12,12,11,13,16,13,12,18,18,17,18,18,12,14,12,12,14,17,13,11,11,12,11,14,10,13,14,13,18,17,17,17,14,15,15,15,14,14,14,14,14,16,15,15,15,16,15,15,15,16,15,14,13,16,12,14,15,17,15,15,14,15,15,16,15,16,16,16,12,15,15,15,14,15,13,14,19,17,13,14,14,14,16,13,14,13,16,13,13,17,16,18,18,17,13,16,15,11,11,14,11,13,12,16,16,16,15,0]",
                "heartRate":"[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,65,65,61,65,65,64,65,66,63,65,64,63,66,64,66,64,62,61,61,60,63,64,65,59,63,65,60,59,64,63,60,66,65,61,64,61,63,65,64,62,65,61,61,62,64,65,65,64,62,64,63,64,67,66,63,59,61,64,59,60,62,65,66,61,65,65,63,64,64,60,65,67,63,67,64,61,59,60,62,60,61,66,66,64,63,63,59,59,65,61,61,58,60,63,63,60,62,62,62,62,60,68,76,80,78,79,78,76,75,74,77,75,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,74,74,74,72,72,72,72,72,72,72,72,72,72,72,72,76,77,77,78,79,76,78,78,77,75,74,74,73,72,72,72,72,71,72,74,72,74,70,72,73,71,70,75,75,74,72,71,72,70,70,70,69,70,70,70,73,75,74,73,73,72,73,71,72,72,72,71,71,72,71,72,80,74,74,75,75,75,73,73,74,81,79,80,80,74,75,76,75,76,79,75,75,73,74,75,76,76,78,81,78,76,73,77,81,80,81,78,80,79,82,79,78,84,85,83,79,81,78,82,79,85,83,74,79,79,79,85,85,82,81,79,79,81,77,77,76,79,81,80,78,75,80,79,77,76,84,78,79,84,86,85,85,86,85,85,84,76,79,79,81,81,80,75,78,79,83,80,83,75,85,83,82,86,86,86,85,86,85,85,85,83,84,84,85,84,83,82,79,82,77,80,80,80,78,80,78,77,79,80,79,85,86,86,86,86,85,84,84,84,83,79,83,81,79,80,80,81,78,76,79,76,78,78,75,82,80,84,86,88,87,86,86,85,85,85,84,84,84,83,83,84,83,83,77,77,78,80,79,76,80,81,79,79,78,77,82,87,88,88,86,86,86,85,85,85,83,83,83,82,83,83,82,82,82,83,77,78,80,80,84,85,85,85,85,84,85,84,84,83,84,84,84,84,84,83,85,83,85,84,84,84,85,84,84,85,84,84,84,84,85,83,83,87,88,87,87,88,81,87,79,77,77,74,74,74,74,75,78,77,83,79,84,83,76,74,75,82,85,78,83,83,86,78,80,83,85,80,77,79,78,91,91,92,91,84,80,85,80,90,76,79,76,78,87,92,90,85,87,88,81,79,80,79,88,86,78,85,87,87,89,92,91,91,88,90,90,90,88,88,85,83,85,85,84,81,82,87,87,81,82,81,79,77,79,78,78,78,81,81,82,82,80,84,85,80,83,84,77,78,84,83,82,75,83,83,81,85,88,91,92,88,81,73,83,80,80,82,78,78,75,75,75,75,74,78,76,76,78,88,80,86,84,78,79,76,72,79,81,81,78,83,84,74,87,80,79,80,81,83,75,80,82,84,90,84,81,81,80,77,76,78,83,85,77,80,78,74,78,82,82,80,76,77,78,76,80,77,74,76,79,86,88,86,86,84,83,89,89,88,84,84,83,81,81,78,79,77,74,75,78,78,80,77,82,84,90,83,84,83,76,73,77,76,80,84,78,86,82,84,80,90,83,78,76,76,78,79,81,75,80,78,75,79,75,76,82,83,81,80,70,74,76,78,75,74,81,74,80,72,74,76,75,75,75,75,79,78,75,74,73,76,81,76,76,76,82,78,80,81,75,78,91,92,92,89,75,89,91,89,88,88,88,87,86,86,85,86,85,86,84,87,88,88,87,87,86,86,86,86,85,85,84,84,85,84,84,84,84,82,83,83,82,80,84,81,83,82,82,82,82,82,81,82,81,82,81,81,82,84,81,80,76,81,79,79,71,82,79,73,80,80,79,80,78,83,82,78,78,80,82,83,77,77,74,80,79,76,75,73,77,70,72,74,75,76,82,81,81,79,82,76,74,73,74,75,84,80,73,81,78,84,74,77,75,79,80,81,76,78,76,80,79,84,77,76,88,86,85,88,87,85,85,83,85,87,86,87,85,85,82,79,78,78,80,79,79,74,76,76,75,78,76,80,75,78,77,76,83,88,88,88,87,86,86,85,85,85,84,84,83,83,83,78,83,86,84,83,84,86,80,84,83,85,82,84,84,83,83,82,81,84,80,82,82,77,80,81,82,83,83,82,80,75,77,84,83,77,79,76,78,82,86,80,81,83,84,82,83,82,80,82,84,82,83,84,84,82,84,81,77,76,79,86,82,84,77,79,73,78,74,81,83,84,0,0,0,0,0,0,0,0,0,0,0,0,60,64,63,57,64,60,59,60,65,69,69,67,62,64,68,63,58,58,59,63,64,66,62,66,62,58,59,64,67,63,59,63,66,63,62,61,56,62,64,66,66,61,63,62,60,63,62,58,59,62,60,62,58,63,60,66,66,63,63,62,64,64,65,62,64,66,65,66,66,61,61,63,60,63,62,62,62,62,65,60,62,59,57,60,61,60,62,64,67,61,62,62,64,65,63,65,60,67,62,72,80,78,79,79,78,78,76,78,77,77,76,76,77,77,76,76,76,77,80,77,77,76,76,77,75,75,75,75,74,74,74,74,75,77,77,77,75,74,73,74,74,78,83,77,83,84,83,83,82,79,83,83,80,81,81,84,79,76,79,77,75,75,76,83,75,80,82,78,78,76,79,76,75,0]",
                "status":"[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,34,34,1,34,34,1,1,1,1,1,1,1,1,1,1,1,34,34,1,1,1,1,34,1,1,1,1,34,34,34,1,1,1,1,1,1,34,1,1,1,1,34,34,34,34,34,34,34,34,34,34,1,34,34,34,1,1,1,34,1,1,34,34,34,34,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,34,1,1,1,1,1,1,1,1,1,34,1,1,1,1,1,1,1,1,1,1,1,0,0,1,1,34,1,34,34,1,1,1,1,1,34,34,1,34,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,34,34,1,34,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,34,1,1,1,1,34,34,34,34,34,34,1,1,34,1,34,34,34,34,34,34,34,34,34,34,34,1,1,1,1,1,34,34,34,34,34,1,1,1,1,1,1,1,1,34,1,1,1,34,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,34,1,1,1,1,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,1,1,1,1,34,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,34,34,34,1,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,34,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,34,34,34,1,34,1,1,34,34,34,1,34,1,1,34,34,1,1,1,1,1,34,34,34,34,1,34,34,1,34,34,34,1,34,34,34,34,1,1,1,1,1,34,1,34,1,1,34,34,1,34,1,1,1,1,1,1,1,1,1,34,34,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,1,1,1,1,1,1,1,1,1,1,1,34,34,34,1,1,1,34,1,34,34,1,34,34,34,34,34,34,34,34,34,34,34,34,34,34,34,34,1,34,34,34,34,1,1,1,1,34,1,1,1,34,1,34,34,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,34,34,1,1,34,34,1,34,34,1,34,34,1,1,1,1,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,34,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,1,1,1,1,1,34,34,1,34,34,34,1,34,1,1,1,1,1,1,1,34,34,34,34,34,34,1,34,1,34,34,34,1,34,34,34,1,1,1,1,1,34,1,1,1,1,34,1,1,1,34,34,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,1,34,34,1,34,34,34,1,34,34,1,1,1,1,34,34,34,1,34,34,34,34,34,34,34,1,34,1,34,34,34,34,1,34,1,1,1,1,34,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,34,34,34,34,34,34,34,1,34,34,34,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,34,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,1,1,1,1,1,1,34,34,1,1,1,1,1,1,1,1,1,1,1,1,34,1,34,34,34,34,34,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,0,34,34,34,34,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,34,34,34,34,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,1,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,1,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,34,1,1,1,34,1,1,1,1,1,1,1,1,34,1,1,1,1,1,1,1,34,34,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,34,34,1,1,1,1,1,1,1,1,0]",
                "statusValue":"[62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,5,1,1,0,1,1,0,0,0,0,0,0,0,0,0,0,0,2,1,0,0,0,0,1,0,0,0,0,1,1,1,0,0,0,0,0,0,1,0,0,0,0,1,1,3,2,1,2,1,2,1,1,0,1,1,1,0,0,0,1,0,0,1,1,1,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,4,3,0,0,1,0,1,1,0,0,0,0,0,2,1,0,1,0,0,0,0,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,62,5,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,1,1,0,2,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,3,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,0,0,0,0,1,1,1,2,1,1,0,0,3,0,2,1,1,1,1,2,1,3,2,1,3,0,0,0,0,0,1,3,2,1,1,0,0,0,0,0,0,0,0,1,0,0,0,1,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,2,0,1,0,0,0,0,0,0,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,3,1,0,1,1,1,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,3,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,1,2,0,1,0,0,1,2,1,0,1,0,0,1,2,0,0,0,0,0,1,2,1,2,0,1,1,0,1,1,1,0,1,1,2,2,0,0,0,0,0,2,0,1,0,0,1,2,0,1,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,1,2,0,0,0,1,0,1,1,0,1,1,3,1,1,2,2,1,2,3,3,2,1,1,1,1,0,2,2,1,1,0,0,0,0,1,0,0,0,2,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,1,0,0,1,2,0,1,1,0,1,1,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,2,1,0,1,3,1,0,1,0,0,0,0,0,0,0,1,2,1,2,1,2,0,1,0,1,1,1,0,1,1,1,0,0,0,0,0,1,0,0,0,0,2,0,0,0,1,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,3,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,1,1,0,2,1,1,0,2,1,0,0,0,0,1,1,3,0,2,1,1,2,2,1,2,0,1,0,1,1,1,2,0,1,0,0,0,0,2,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,3,1,1,1,1,3,1,2,0,1,2,1,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2,1,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,3,1,1,2,1,0,0,0,0,0,0,0,0,0,0,0,0,0,62,62,62,62,5,1,1,3,1,0,0,0,0,0,0,0,0,0,0,0,0,0,3,1,1,1,1,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,2,1,0,0,0,1,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,1,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,2,0,0,0,0,0,0,0,0,62]",
                "eTemp":"[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]",
                "eHumidity":"[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]",
                "eLight":null,
                "eCO2":null,
                "eNoise":null,
                "mTemp":null,
                "mHumidity":null
            },
            "analysis":{
                "startTime":1643515200,
                "duration":935,
                "wake":217,
                "outOfBedDuration":5,
                "avgHeartRate":77,
                "avgBreathRate":14,
                "sleepStateStr":"[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,2,2,2,2,2,2,2,2,4,4,3,3,3,3,3,3,4,4,4,4,4,2,2,2,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,2,2,2,2,2,2,2,2,2,3,3,3,3,2,2,1,1,2,2,3,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,2,2,2,2,2,2,2,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,4,3,3,3,3,3,3,4,4,4,3,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,2,2,2,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,4,2,2,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,3,3,2,2,2,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,3,3,3,3,3,3,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,2,2,2,2,2,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,4,4,4,4,4,4,4,4,4,4,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,4,4,4,4,4,3,3,3,3,3,3,3,3,3,4,4,4,4,4,2,2,2,2,2,2,2,2,2,2,2,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,4,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,4,4,4,4,4,4,4,4,4,4,4,2,2,4,4,4,4,4,4,2,2,0,0,0,0,1,2,2,2,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,2,4,4,4,4,4,4,2,2,2,2,2,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,2,2,2,2,2,2,2,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,2,2,0]",
                "lightSleepDuration":713,
                "deepSleepDuration":222,
                "offBedAllTime":5,
                "sleepArray":[
                    {
                        "startTime":1643530740,
                        "duration":1179
                    }
                ],
                "maxReport":{
                    "inBedTime":1643530740,
                    "getupTime":1643601480,
                    "onsetTime":1643531220,
                    "wakeupTime":1643601420,
                    "scale":65,
                    "tsTime":1170,
                    "sleepRate":2,
                    "asleepDur":935,
                    "seIndex":79,
                    "markEfficiency":2,
                    "awakeDur":207,
                    "awakePct":18,
                    "lightDur":713,
                    "lightPct":61,
                    "deepDur":222,
                    "deepPct":19,
                    "markDeep":3,
                    "offBedDur":4,
                    "offBedPct":0,
                    "discDur":24,
                    "discPct":2,
                    "awakeCnt":22,
                    "awakeEvent":[
                        {
                            "startTime":1643532000,
                            "duration":6
                        },
                        {
                            "startTime":1643533320,
                            "duration":28
                        },
                        {
                            "startTime":1643536560,
                            "duration":11
                        },
                        {
                            "startTime":1643537460,
                            "duration":6
                        },
                        {
                            "startTime":1643537880,
                            "duration":16
                        },
                        {
                            "startTime":1643546220,
                            "duration":8
                        },
                        {
                            "startTime":1643553840,
                            "duration":2
                        },
                        {
                            "startTime":1643555340,
                            "duration":3
                        },
                        {
                            "startTime":1643559420,
                            "duration":4
                        },
                        {
                            "startTime":1643566920,
                            "duration":10
                        },
                        {
                            "startTime":1643570040,
                            "duration":4
                        },
                        {
                            "startTime":1643574180,
                            "duration":33
                        },
                        {
                            "startTime":1643579640,
                            "duration":3
                        },
                        {
                            "startTime":1643581320,
                            "duration":2
                        },
                        {
                            "startTime":1643582040,
                            "duration":22
                        },
                        {
                            "startTime":1643584500,
                            "duration":11
                        },
                        {
                            "startTime":1643587920,
                            "duration":3
                        },
                        {
                            "startTime":1643589840,
                            "duration":2
                        },
                        {
                            "startTime":1643590320,
                            "duration":2
                        },
                        {
                            "startTime":1643593620,
                            "duration":3
                        },
                        {
                            "startTime":1643594160,
                            "duration":8
                        },
                        {
                            "startTime":1643599800,
                            "duration":9
                        }
                    ],
                    "offBedCnt":3,
                    "leftBedEvent":[
                        {
                            "startTime":1643537580,
                            "duration":2
                        },
                        {
                            "startTime":1643538840,
                            "duration":21
                        },
                        {
                            "startTime":1643590440,
                            "duration":5
                        }
                    ],
                    "bmCnt":0,
                    "markBM":10,
                    "hrMean":77,
                    "hrMin":56,
                    "hrMax":92,
                    "hrHiDur":0,
                    "hrHiPct":0,
                    "hrLoDur":0,
                    "hrLoPct":0,
                    "brMean":14,
                    "brMin":9,
                    "brMax":20,
                    "brHiDur":0,
                    "brHiPct":0,
                    "brLoDur":345,
                    "brLoPct":29,
                    "ahiDur":0,
                    "ahiCnt":0,
                    "ahiMaxDur":0
                }
            }
        }
    ]
}
重运算旧的睡眠数据
Sleepace养老对接方案3.x更新新的报告算法，并将数据进行了迁移。
适用于在2022/7/5之前对接了BM8701-2，且想要升级到Sleepace养老对接方案3.x的厂商 。若设备已在使用，原来产生的历史数据将会丢失，若需要使用旧数据，需要调用接口重新生成报告至新的数据库。  
URL
http(s)://domain{:port}/sleepace/analysisOldData
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
}
}
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
响应
{
	"status": 0
	"msg": "",
data:
}
字段
类型
描述
status
int
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
例子
请求
Post   http(s)://domain{:port}/sleepace/analysisOldData
参数 {
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
}
}
响应
{
	"status": 0
	"msg": "",
    “data”:null
}
报警
注：请将固件升级到5.x以上版本，否则无法支持
获取用户的报警配置信息
http(s)://domain{:port}/sleepace/getalarmnotifyconfig
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx"
}
}
参数说明
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
userid
string
合作方用户的唯一标识
响应
{
"status": 0,
"msg": null,
"data": {
"fallFlag": 0,
"leftBedFlag": 1,
"leftBedDuration": 10,
"leftBedStartHour": 17,
"leftBedStartMinute": 0,
"leftBedEndHour": 21,
"leftBedEndMinute": 0,
"heartRateFastFlag": 0,
"heartRateFastDuration": 600,
"maxHeartRate": 120,
"heartRateSlowFlag": 0,
"heartRateSlowDuration": 1200,
"minHeartRate": 45,
"breathRateFastFlag": 0,
"breathRateFastDuration": 1200,
"maxBreathRate": 26,
"breathRateSlowFlag": 0,
"breathRateSlowDuration": 1200,
"minBreathRate": 10,
"breathPauseFlag": 0,
"breathPauseDuration": 60,
"bodyMoveFlag": 0,
"bodyMoveDuration": 10,
"nobodyMoveFlag": 0,
"nobodyMoveDuration": 60,
"noTurnOverFlag":0,
"noTurnOverDuration":60,
"situpFlag": 0,
"onbedFlag": 0,
"onbedDuration": 600
}
}
字段
类型
描述
status
int
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
fallFlag
int
传感器脱落报警开关，0关1开
leftBedFlag
int
离床报警开关，0关1开
leftBedDuration
int
离床时长，单位秒
leftBedStartHour
int
离床报警范围，开始小时[0,23]
leftBedStartMinute
int
离床报警范围，开始分钟[0,59]
leftBedEndHour
int
离床报警范围，结束小时[0,23]
leftBedEndMinute
int
离床报警范围，结束分钟[0,59]
heartRateFastFlag
int
心率过速报警开关，0关 1开
heartRateFastDuration
int
心率过速持续时长，单位秒
maxHeartRate
int
最大心率值，>=该值触发报警
heartRateSlowFlag
int
心率过缓报警开关，0关 1开
heartRateSlowDuration
int
心率过缓持续时长，单位秒
minHeartRate
int
最小心率值，<=该值触发报警
breathRateFastFlag
int
呼吸过速报警开关，0关 1开
breathRateFastDuration
int
呼吸过速持续时长，单位秒
maxBreathRate
int
最大呼吸值，>=该值触发报警
breathRateSlowFlag
int
呼吸过缓报警开关，0关 1开
breathRateSlowDuration
int
呼吸过缓持续时长，单位秒
minBreathRate
int
最小呼吸值，<=该值触发报警
breathPauseFlag
int
呼吸暂停报警开关，0关 1开
breathPauseDuration
int
呼吸暂停持续时长，单位秒
bodyMoveFlag
int
频繁体动报警开关，0关 1开
bodyMoveDuration
int
频繁体动持续时长，单位分钟，（时长必须是实时数据上报间隔的倍数，如：实时数据上报间隔设置为2分钟，则体动持续时长可以设置为：2分钟，4分钟，6分钟，... ...）
nobodyMoveFlag
int
无体动报警开关，0关 1开
nobodyMoveDuration
int
无体动持续时长，单位分钟
（时长必须是实时数据上报间隔的倍数，如：实时数据上报间隔设置为2分钟，则体动持续时长可以设置为：2分钟，4分钟，6分钟，... ...）
noTurnOverFlag
int
无翻身报警开关，0关 1开
noTurnOverDuration
int
无翻身持续时长，单位分钟
（时长必须是实时数据上报间隔的倍数，如：实时数据上报间隔设置为2分钟，则体动持续时长可以设置为：2分钟，4分钟，6分钟，... ...）
situpFlag
int
坐起报警开关，0关 1开
onbedFlag
int
在床报警开关，0关 1开
onbedDuration
int
在床持续时长，单位秒
修改用户报警配置
http(s)://domain{:port}/sleepace/updatealarmnotifyconfig
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx",
"deviceId":"xxx",
		" fallFlag":1,
		…
}
}
参数说明
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
userId
string
合作方用户的唯一标识
deviceId
string
设备密文id
fallFlag
int
传感器脱落报警开关，0关1开
…
int
参考上面获取接口的描述
报警逻辑（包含可设置的阈值范围、报警逻辑等）可查阅以下文档进行深入了解
Sleepace 养老对接方案报警业务逻辑参考及应用场景说明-20241206.xls
(27 kB)
特别说明：参数deviceId后面的报警参数，可以是一个或多个
响应
{
"status": 0,
"msg": null,
"data": null
}
字段
类型
描述
status
int
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
data
object
更新接口，data为null
查询用户某个时间区间内的所有报警记录
http(s)://domain{:port}/sleepace/getalarmrecordlist
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx",
"startTime":1660190319,
"endTime":1660290319
}
}
参数说明
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
userid
string
合作方用户的唯一标识
startTime
int
报警记录开始时间戳，单位秒，包含该时间点
endTime
int
报警记录截止时间戳，单位秒，包含该时间点
特别说明：时间区间最大为7天，即: endTime – startTime <= 7 * 24 * 60 * 60;
响应
{
"status": 0,
"msg": null,
"data": [
{
"id": 1,
"type": "alarmLeftBed",
"triggerTime": 1660190319,
"userId": 33934,
"deviceId": "1znz6fkqjgk6w",
"status": 0,
"relieveReason": null,
"relieveTime": 0
}
]
}
字段
类型
描述
status
int
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
data
array
报警记录
id
long
记录id
type
string
报警类型，详见报警类型说明
triggerTime
int
报警时间戳，单位秒
userId
long
用户id
deviceId
string
设备密文id
status
int
报警状态0未自动解除，1系统自动解除
relieveReason
string
解除报警的原因
relieveTime
int
解除报警的时间戳，单位秒
根据报警状态查询报警记录
http(s)://domain{:port}/sleepace/getalarmrecordlistbystatus
参数
{
"token": {
"appId": "xxx",
"secureKey":"xxx"
},
"data": {
"userId":"xxx",
"status":0,
"startTime":1660190319,
"endTime":1660290319
}
}
参数说明
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
userid
string
合作方用户的唯一标识
status
int
0:未自动解除的报警记录，1:已解除的报警记录
startTime
int
报警记录开始时间戳，单位秒，包含该时间点
endTime
int
报警记录截止时间戳，单位秒，包含该时间点
特别说明：时间区间最大为7天，即: endTime – startTime <= 7 * 24 * 60 * 60;
响应
{
"status": 0,
"msg": null,
"data": []
}
字段
类型
描述
status
int
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
data
array
报警记录，结构参考上面的接口
4G设备通讯信息（M901L专用）
通过设备密文查询设备网络信息
URL
http(s)://domain{:port}/sleepace/device/net
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
}
}
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
deviceId
string
设备ID
响应
{
    "status": 0,
    "msg": "",
    "data": {
      "deviceId": "u84ykskr2rvng",
      "iccid": "89860465011981033249",
      "imei": "864578068280536"
    }
}
字段
类型
描述
status
int
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
deviceId
string
设备ID
iccid
string
iccid
imei
string
imei
原始数据上传功能
原始数据指的是设备采集到的最原始的数据，用于析数据出现与预期不符的情况。一般情况下不要打开此功能，只有需要定位睡眠数据异常问题时才打开。设备在线前提下，设备每秒上传原始数据，所以若要定位问题，一定需要保证设备一直在线。另外，一个小时的原始数据将消耗1MB左右，请谨慎打开此功能
原始数据上传设置
通过设备密文开启原始数据功能及开启时段
URL
http(s)://domain{:port}/sleepace/origin/set
参数
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
deviceId
string
设备ID
leftRight
number
单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准
startTime
number
开始时间戳，毫秒，
endTime
number
结束时间戳，毫秒
注意：设置成功后，服务器什么时候通知设备上传原始数据？
 1、调用接口后服务器会判断当前时间是否在上传的时间范围，如果在则通知设备上传原始数据，不在则不通知
 2、服务器会在设备上传报告或设备重连服务器时通知设备上传原始数据，如果这个时间不在设置的上传时间范围，则要等到下次设备上传报告或设备重连服务器的时候。
响应
{
    "status": 0,
    "msg": ""
}
字段
类型
描述
status
number
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
删除原始数据上传设置
通过设备密文关闭原始数据上传功能
URL
http(s)://domain{:port}/sleepace/origin/del
参数
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
deviceId
string
设备ID
leftRight
number
单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准
响应
{
    "status": 0,
    "msg": ""
}
字段
类型
描述
status
number
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
查询指定设备原始数据上传设置
通过设备密文查询设备原始数据上传设置
URL
http(s)://domain{:port}/sleepace/origin/get
参数
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
deviceId
string
设备ID
leftRight
number
单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准
响应
字段
类型
描述
status
number
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
deviceId
string
设备ID
leftRight
number
单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准
startTime
number
开始时间戳，毫秒
endTime
number
结束时间戳，毫秒
查询所有设备原始数据上传设置
查询所有设备原始数据上传设置
URL
http(s)://domain{:port}/sleepace/origin/all
参数
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
响应
字段
类型
描述
status
number
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
deviceId
string
设备ID
leftRight
number
单人设备值为0；双人设备表示左右侧(0/1),具体对应关系以实际产品为准
startTime
number
开始时间戳，毫秒
endTime
number
结束时间戳，毫秒
原始数据记录查询
按照时间段查询设备上报的原始数据记录列表
URL
http(s)://domain{:port}/sleepace/origin/recordList
参数
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
deviceId
string
设备ID
startTime
number
记录最早开始时间戳，毫秒
endTime
number
记录最晚开始时间戳，毫秒
响应
{
    "status": 0,
    "msg": ""，
"data": [
        {
            "originUrl": "",
            "weightUrl": "t",
            "startTime": ,
            "createTime": 
        }
    ]
}
字段
类型
描述
status
number
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
originUrl
string
原始数据下载地址
weightUrl
string
重力信号原始数据下载地址
startTime
number
开始时间戳，毫秒
createTime
number
创建时间戳，毫秒
心率呼吸率模式设置（BM8701-2专用）
仅支持固件版本≥6.30
URL
http(s)://domain{:port}/sleepace/heartModeSet 
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>,
 mode:<<模式>>
}
}
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
deviceId
string
设备加密ID
mode
int
0：模式0(默认)，心率、呼吸率一直有值，在信号质量不好或体动较多时，准确率会比平稳躺着时低一些；
1：模式，心率、呼吸率计算不出(信号质量不好或体动较多)，上报无效值 255
注：两个模式在信号质量好且体动较少时，准确率是一致的
响应
{
	"status": 0
	"msg": "",
data：null
}
字段
类型
描述
status
int
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
心率呼吸率模式获取（BM8701-2专用）
仅支持固件版本≥6.30
URL
http(s)://domain{:port}/sleepace/heartModeGet 
参数
{
token:{
appId:’’,<<与消息队列的账号相同>>
secureKey:’’,<<与消息队列的密码相同>>
},
data:{
deviceId:<<设备ID>>
}
}
字段
类型
描述
appId
string
与消息队列的账号相同
secureKey
string
与消息队列的密码相同
deviceId
string
设备加密ID
响应
Plain Text
复制代码
1
2
3
4
5
6
7
{
    "status": 0,
    "msg": null,
    "data": {
        "mode": 0
    }
}
字段
类型
描述
status
int
状态码，0表示成功，其他失败，详见状态码
msg
string
失败原因
mode
int
0：模式0(默认)，心率/呼吸率 计算不出时，保留之前的值不变或者加一定的波动
1：模式，心率/呼吸率计算不出，上报无效值 255
状态码
状态码
描述
-1
未登录或未认证
0
成功
1
设备离线
2
服务器错误
3
超时
4
设备未绑定
5
设备不存在
8
操作命令不存在
9
参数错误
10
device not belong your channel
该设备不在你的渠道下
11
认证错误



# 实时数据常见问题
1.实时数据中的心率呼吸率返还数据为“255”是什么情况呢？
A： 255是无效值，分两种情况：1.还没计算出心率；2.因为体动导致心率暂时不给出。  
 255会出现在设备重启时或监测到有体动发生时 。
实时数据的更新频率是多久？

2.实时数据的更新频率是多久？
A： BM8701-2（Wi-Fi版）：1秒更新1次，支持调用接口配置上报间隔；
M901L（4G版）：
心率、呼吸率正常时，30秒更新1次；
当心率、呼吸率达到报警设置的取值时，呼吸率=0rpm时，立即上报实时数据，下一次上报数据还是以首次上报实时数据的时间为起点+30秒*N次，数据恢复正常时又立即上报实时数据，见下图：

3.为什么心率和呼吸率为“255”？
1）由于设备的监测原理是通过传感器捕捉监测范围内动作信号，然后通过算法分析出心率、呼吸率。如果你躺在床上体动较大，设备捕捉不到有效信号，心率和呼吸率暂时计算不出来，待躺的比较平稳时，则可计算出心率和呼吸率
2）设备适合于安装在床板和床垫之间，如果放置在沙发上方、床垫上方，心率和呼吸率也有可能会计算不出来
3）前端处理可以显示为“--”或者“数据计算中”
躺在/坐在床上没有睡着，但是设备上报睡眠状态为“浅睡/深睡”？
由于设备的监测原理是通过传感器捕捉监测范围内动作信号，然后通过算法分析出心率、呼吸率、睡眠状态和时间。如果躺在设备上并且不怎么动，身体状态接近睡着状态，存在一定概率被判为睡着。


4.躺在/坐在床上没有睡着，但是设备上报睡眠状态为“浅睡/深睡”？
由于设备的监测原理是通过传感器捕捉监测范围内动作信号，然后通过算法分析出心率、呼吸率、睡眠状态和时间。如果躺在设备上并且不怎么动，身体状态接近睡着状态，存在一定概率被判为睡着。

5. 躺在床上睡着了，但是设备上报睡眠状态为“清醒”？
问题现象：
躺在设备上方，实际上是睡着的，但是设备上报睡眠状态为“清醒”
原因和解决方案：
由于设备的监测原理是通过传感器捕捉监测范围内动作信号，然后通过算法分析出心率、呼吸率、睡眠状态和时间。如果周遭有电磁干扰或振动干扰，存在一定概率会误判为清醒。

6.坐在床上，但是设备上报为“离床”？
由于设备的监测原理是通过传感器捕捉监测范围内动作信号，如果使用者没有躺在设备上方，有可能是采集不到有效信号，设备则认为不在床。
本设备的正确使用方式如下：


7. 躺在床上睡着了，但是设备上报睡眠状态为“清醒”？
问题现象：
躺在设备上方，实际上是睡着的，但是设备上报睡眠状态为“清醒”
原因和解决方案：
由于设备的监测原理是通过传感器捕捉监测范围内动作信号，然后通过算法分析出心率、呼吸率、睡眠状态和时间。如果周遭有电磁干扰或振动干扰，存在一定概率会误判为清醒。
坐在床上，但是设备上报为“离床”？
由于设备的监测原理是通过传感器捕捉监测范围内动作信号，如果使用者没有躺在设备上方，有可能是采集不到有效信号，设备则认为不在床。
本设备的正确使用方式如下：

睡眠过程中没有离床，但是报告中出现“离床”状态？
8.床上无人，为什么会有心率/呼吸率？
问题现象：
设备安装规范的情况下，床上无人，实时数据/睡眠报告出现了心率、呼吸率
原因和解决方案：
电磁干扰：
尽量不要与其他的电子设备共用一个排插，特别是大功率的适配器（如电脑适配器）
使用标配的5V1A或5V 2A适配器，不要使用快充适配器
大功率电器如冰箱、空气净化器、洗衣机与设备保持2米以上的距离
床板没有放平，或有微小震动干扰
请将床板放平
请先查询设备是否已初始化完成，调用接口“查询设备初始化状态”，若未初始化完成，在床上无人且设备安装规范、床为放平状态时，可调用接口“传感器校准”手动校准设备。

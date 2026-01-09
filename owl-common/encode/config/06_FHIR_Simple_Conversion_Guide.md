  Observation = 原始数据/状态上传 (IoT 实时监控)
  Flag = 基于数据判断后的告警事件 (IoT 报警事件)

## 一、FHIR 与 SNOMED 的区别

### 1.1 LOINC vs SNOMED：谁管什么？

- **LOINC 解决"这一列是什么项目？"**  
  - 例如：`8867-4` = Heart rate（这是"心率"这个测量项目）  
  - 在 FHIR 里一般放在 `Observation.code`，表示"测的是什么"。

- **SNOMED 解决"这个值说明了什么临床含义？"**  
  - 例如：`271636001` = Tachycardia（心动过速，这个结果的医学意义）  
  - 在 FHIR 里放在 `Observation.valueCodeableConcept` / `interpretation`、`Flag.code` 等，表示"结果是什么状态/事件"。

## 二、Observation ↔ Alarm ↔ 映射总表

### 2.1 统一分类映射表（FHIR Category）

**统一原则**：所有表统一使用 FHIR Category，不再使用 TDP Tag Category。

| 旧名称（已废弃）              | 新 FHIR Category  | 用于资源类型 | 数据库表字段           | 说明                                    |
|:-----------------------------|:------------------|:-------------|:-----------------------|:----------------------------------------|
| **Physiological** (基础测量) | `vital-signs`     | Observation  | `iot_timeseries.category` | 生命体征基础测量值（心率、呼吸率原始数值） |
| **Physiological** (异常状态) | `clinical`        | Flag         | `alarm_events.category`   | 生命体征异常告警（心动过速、心动过缓、呼吸暂停） |
| **Behavioral** (基础行为)    | `activity`        | Observation  | `iot_timeseries.category` | 行为观测（离床、上床、床上坐起，不触发告警） |
| **Behavioral** (健康风险)    | `behavioral`      | Flag         | `alarm_events.category`   | 行为健康告警（超2H未翻身、2H无体动） |
| **HealthCondition**          | `behavioral`      | Flag         | `alarm_events.category`   | 健康风险告警（统一归为 behavioral） |
| **Posture**                  | `activity`        | Observation  | `iot_timeseries.category` | 姿态观测（站立、坐位、卧位） |
| **MotionState**              | `activity`        | Observation  | `iot_timeseries.category` | 运动状态观测（行走、静止） |
| **SleepState**               | `activity`        | Observation  | `iot_timeseries.category` | 睡眠状态观测（清醒、浅睡眠、深睡眠） |
| **Safety**                   | `safety`          | Flag         | `alarm_events.category`   | 安全告警（跌倒、疑似跌倒、滞留、24H无人） |
| **DeviceError**              | `device`          | Flag         | `alarm_events.category`   | 设备技术告警（设备故障、传感器断线） |

**数据库字段说明**：
- `iot_timeseries.category`: 统一使用 FHIR Category（`vital-signs`, `activity`, `safety`, `clinical`, `behavioral`, `device`）
- `alarm_events.category`: 统一使用 FHIR Flag Category（`safety`, `clinical`, `behavioral`, `device`）
- `event_mapping.category`: 统一使用 FHIR Category
- `posture_mapping.category`: 统一使用 FHIR Category（`activity`, `safety`）


============================================================================================================================================
## 三、FHIR Flag（报警标志）分为 **4 大类**，对应不同的 `flag-category`：

| 领域 | FHIR flag-category | 说明 |
|------|-------------------|------|
| **临床安全** | `safety` | 跌倒、滞留、危险姿态等直接威胁生命安全的事件 |
| **生命体征** | `clinical` / `vital-signs` | 心率/呼吸率异常等生命体征告警（FHIR推荐使用 `clinical`） |
| **行为健康** | `behavioral` / `behavioral-health` | 长时间无活动、异常行为模式等行为健康告警 |
| **设备技术** | `device` / `technical-alert` | 设备故障、网络异常等技术告警 |


### 3.1 cloud_alarm_policies 报警项分类映射表

根据 `cloud_alarm_policies` 表中的报警项，按照 FHIR Category 分类：

#### 按 FHIR Category 分类

| FHIR Category      | 报警项                              | 设备类型   | 说明                              | 默认 DangerLevel |
|:-------------------|:-----------------------------------|:----------|:----------------------------------|:-----------------|
| **`safety`**       | `Fall`                             | Radar     | 跌倒                              | EMERGENCY        |
|                    | `SuspectedFall`                    | Radar     | 可疑跌倒                           | WARNING          |
|                    | `Stay`                             | Radar     | 滞留（卫生间/浴室）                  | WARNING          |
|                    | `NoActivity24h`                    | Radar     | 24小时无人                         | EMERGENCY        |
| **`clinical`**     | `Radar_ApneaHypopnea`              | Radar     | 呼吸暂停                           | DISABLE        |
|                    | `Radar_AbnormalHeartRate`          | Radar     | 心率异常（心动过速/心动过缓）         | EMERGENCY        |
|                    | `Radar_AbnormalRespiratoryRate`    | Radar     | 呼吸频率异常（呼吸急促/呼吸缓慢）      | EMERGENCY        |
|                    | `VitalsWeak`                       | Radar     | 生命体征微弱                        | WARNING          |
|                    | `SleepPad_ApneaHypopnea`           | SleepPad  | 呼吸暂停                           | EMERGENCY        |
|                    | `SleepPad_AbnormalHeartRate`       | SleepPad  | 心率异常（心动过速/心动过缓）          | EMERGENCY        |
|                    | `SleepPad_AbnormalRespiratoryRate` | SleepPad  | 呼吸频率异常（呼吸急促/呼吸缓慢）      | EMERGENCY        |
| **`behavioral`**   | `Radar_LeftBed`                    | Radar     | 离床                              | WARNING          |
|                    | `SleepPad_LeftBed`                 | SleepPad  | 离床                              | WARNING          |
|                    | `SleepPad_SitUp`                   | SleepPad  | 床上坐起                           | WARNING          |
|                    | `SleepPad_AbnormalBodyMovement`    | SleepPad  | 异常体动（超2H未翻身/2H无体动）       | WARNING          |
|                    | `SleepPad_InBed`                   | SleepPad  | 上床（取决于住户service_level）     | DISABLE          |
| **`device`**       | `OfflineAlarm`                     | 所有设备   | 设备离线                           | WARNING          |
|                    | `LowBattery`                       | 所有设备   | 低电量                             | WARNING          |
|                    | `DeviceFailure`                    | 所有设备   | 设备故障                           | EMERGENCY        |
|                    | `AngleException`                   | Radar     | 角度异常                           | WARNING          |

#### 按设备类型分类

| 设备类型            | 报警项                              | FHIR Category | 说明                              | 默认 DangerLevel |
|:-------------------|:-----------------------------------|:--------------|:----------------------------------|:-----------------|
| **所有设备（通用）** | `OfflineAlarm`                      | `device`      | 设备离线                          | WARNING            |
|                    | `LowBattery`                       | `device`      | 低电量                            | WARNING          |
|                    | `DeviceFailure`                    | `device`      | 设备故障                           | EMERGENCY        |
| **Radar**          | `Fall`                             | `safety`      | 跌倒                              | EMERGENCY        |
|                    | `SuspectedFall`                    | `safety`      | 可疑跌倒                           | WARNING          |
|                    | `Stay`                             | `safety`      | 滞留（卫生间/浴室）                  | WARNING          |
|                    | `NoActivity24h`                    | `safety`      | 24小时无人                         | EMERGENCY        |
|                    | `Radar_ApneaHypopnea`              | `clinical`    | 呼吸暂停                           | DISABLE        |
|                    | `Radar_AbnormalHeartRate`          | `clinical`    | 心率异常（心动过速/心动过缓）          | EMERGENCY        |
|                    | `Radar_AbnormalRespiratoryRate`    | `clinical`    | 呼吸频率异常（呼吸急促/呼吸缓慢）      | EMERGENCY        |
|                    | `VitalsWeak`                       | `clinical`    | 生命体征微弱                        | WARNING          |
|                    | `Radar_LeftBed`                    | `behavioral`  | 离床                              | WARNING          |
|                    | `AngleException`                   | `device`      | 角度异常                           | WARNING          |
| **SleepPad**       | `SleepPad_ApneaHypopnea`           | `clinical`    | 呼吸暂停                           | EMERGENCY        |
|                    | `SleepPad_AbnormalHeartRate`       | `clinical`    | 心率异常（心动过速/心动过缓）         | EMERGENCY        |
|                    | `SleepPad_AbnormalRespiratoryRate` | `clinical`    | 呼吸频率异常（呼吸急促/呼吸缓慢）      | EMERGENCY        |
|                    | `SleepPad_LeftBed`                 | `behavioral`  | 离床                              | WARNING          |
|                    | `SleepPad_SitUp`                   | `behavioral`  | 床上坐起                           | WARNING          |
|                    | `SleepPad_AbnormalBodyMovement`    | `behavioral`  | 异常体动（超2H未翻身/2H无体动）       | WARNING          |
|                    | `SleepPad_InBed`                   | `behavioral`  | 上床（取决于住户service_level）      | DISABLE          |

**分类说明**：
- **`safety`**: 直接威胁生命安全的事件（跌倒、可疑跌倒、滞留、24H无人）
- **`clinical`**: 生命体征异常告警（心率/呼吸率异常、呼吸暂停、生命体征微弱）
- **`behavioral`**: 行为健康告警（离床、床上坐起、异常体动、上床等）。注意：某些事件（如上床）是否触发告警取决于住户的 `service_level`。当 `service_level` 要求关注此类行为时，应归类为 `behavioral` 并使用 Flag 资源；否则可作为 `activity` 使用 Observation 资源记录
- **`device`**: 设备技术告警（离线、低电量、故障、角度异常）

============================================================================================================================================
## 四、Observation 映射

### 4.1 生命体征（Vital Signs）映射

| 项目      | Observation（测量） | Alarm（报警）        | LOINC   | SNOMED（Observation） | SNOMED（Alarm） | FHIR Category          |
|:----------|:-------------------|:---------------------|:--------|:----------------------|:---------------|:-----------------------|
| 心率 HR   | 实时心率             | 心动过缓（<=44）      | 8867-4   | 364075005             | 342400002      | vital-signs / clinical |
|          |                     | 心动过速（>=116）     |          |                       | 271636001      |                        |
|          |                     | 中度异常             |          |                       | 364400009      |                        |
| 呼吸 RR   | 实时呼吸率           | 呼吸过缓（<=7）       | 9279-1   | 86290005              | 301273002      | vital-signs / clinical |
|           |                    | 呼吸急促（>=27）      |          |                       | 301279008      |                        |
|           |                    | 呼吸暂停             |          |                       | 67905006       |                        |

### 4.2 姿态（Body Position）映射

| 姿态                   | Observation（状态） | Alarm（报警） | SNOMED（姿态） | SNOMED（报警） | Category   |
|:-----------------------|:-------------------|:--------------|:---------------|:---------------|:-----------|
| 仰卧                 | 102538003          | -             | 102538003       | -                  | activity   |
| 左侧卧               | 102536004          | -             | 102536004       | -                  | activity   |
| 右侧卧               | 102535000          | -             | 102535000       | -                  | activity   |
| 坐位                 | 33586001           | -             | 33586001        | -                  | activity   |
| 站立                 | 10904000           | -             | 10904000        | -                  | activity   |
| 长时间保持异常姿态   | 连续 Observation   | 异常姿态报警  | 原姿态编码      | 43029002           | safety     |

### 4.3 行为与活动（Activity）映射

| 行为                   | Observation（行为） | Alarm（报警）         | SNOMED（行为） | SNOMED（报警） | Category   |
|:-----------------------|:-------------------|:----------------------|:---------------|:---------------|:-----------|
| 上床                   | 370998004               | -                       | 370998004       | -              | activity   |
| 离床                   | 424287000               | 离床过久（wandering）     | 424287000       | 129838004      | behavioral |
| 行走                   | 229065009               | -                       | 229065009       | -              | activity   |
| 静止                   | 248257009               | -                       | 248257009       | -              | activity   |
| 床上坐起               | 422256002                | -                       | 422256002       | -              | activity   |
| 摔倒特征动作（瞬时）     | Observation 组合事件      | 跌倒报警                  | -               | 161898004      | safety     |
| 可疑跌倒               | -                        | 129839007               | -               | 129839007      | safety     |

### 4.4 睡眠（Sleep Stage）映射

| 阶段                     | Observation（睡眠阶段） | Alarm（报警）     | SNOMED（睡眠阶段） | Category   |
|:------------------------|:----------------------|:------------------|:-------------------|:-----------|
| 清醒                    | 248220008             | -                 | 248220008          | activity   |
| 浅睡                    | 248221007             | -                 | 248221007          | activity   |
| 深睡                    | 248222000             | -                 | 248222000          | activity   |
| 睡眠中异常姿态/无动作过久  | Observation           | 异常姿态/休克风险 | 姿态SNOMED         | safety     |

### 4.5 环境类（Environment / Behavior）

| 项目           | Observation              | Alarm       | SNOMED（Observation）              | SNOMED（Alarm） | Category        |
|:--------------|:------------------------|:------------|:-----------------------------------|:----------------|:---------------|
| 多人存在        | 多人检测（radar）          | -           | 无官方 SNOMED（自定义 coding）        | -               | social-history  |
| 人员距离 >0.3m  | Observation              | -           | 自定义                              | -               | activity        |
| 卫生间滞留      | 位置不动                   | 424547005   | -                                  | 424547005       | safety          |

### 4.6 设备类（Device）映射

| 项目     | Observation            | Alarm         | SNOMED（Observation） | SNOMED（Alarm） | Category   |
|:---------|:----------------------|:--------------|:---------------------|:----------------|:-----------|
| 雷达在线 | device-status          | -             | 706689003            | -               | device     |
| 设备断线 | device-status=offline  | 397942008     | -                    | 397942008       | device     |
| 配置更新 | informational          | 配置失败报警  | -                    | 4661009         | device     |
| 电量状态 | Observation            | 703507001     | 129545001            | 703507001       | device     |

### 4.7 DangerLevel（严重等级）位置映射

你的 DangerLevel（0–7）→ 对应 FHIR extension：

| DangerLevel | 含义         | FHIR 字段              | 说明                                 |
|:------------|:-------------|:----------------------|:-------------------------------------|
| 0           | UNKNOWN      | extension.valueInteger| 未知（通常不使用）                    |
| 1           | EMERGENCY    | extension.valueInteger| 紧急，高风险，高置信（如跌倒、心率/呼吸率严重异常，持续≥1分钟） |
| 2           | ALERT        | extension.valueInteger| 警报，高危事件（已映射到 WARNING(5)） |
| 3           | CRITICAL     | extension.valueInteger| 严重，需关注（保留，内部使用）         |
| 4           | ERROR        | extension.valueInteger| 错误，设备故障（如传感器断线、角度错误） |
| 5           | WARNING      | extension.valueInteger| 警告，预警（如可疑跌倒、心率/呼吸率中度异常，持续≥5分钟，原 ALERT(2) 映射到此） |
| 6           | NOTICE       | extension.valueInteger| 通知，正常但重要的事件（如配置指令下发） |
| 7           | INFORMATIONAL| extension.valueInteger| 信息，一般信息性消息（如设备上线、状态变化） |

**写法（所有报警共用）**：

```json
"extension": [{
  "url": "http://wisefido.io/fhir/danger-level",
  "valueInteger": 1
}]
```


## 五、常见环境（Environment / Location）类的 SNOMED CT 代码清单

### 5.1 家庭/住宅类环境（Home Environment）

| 环境                     | SNOMED Code | 描述                        |
|:------------------------|:------------|:---------------------------|
| 家 / 家庭环境            | 314767009   | Home environment           |
| 卧室                     | 257667001   | Bedroom                    |
| 客厅（Living room）      | 257669003   | Living room                |
| 浴室 / 卫生间（Bathroom）| 77605003    | Bathroom                   |
| 厕所（Toilet room）      | 257914005   | Toilet                     |
| 厨房（Kitchen）          | 257670002   | Kitchen                    |
| 餐厅（Dining room）      | 257671003   | Dining room                |
| 走廊（Corridor / Hallway）| 257915006  | Corridor                   |
| 玄关 / 门厅（Entrance hall）| 257916007 | Entrance hall             |
| 楼梯（Stairway）         | 257917003   | Stairway                   |
| 洗衣房（Laundry room）   | 257663009   | Laundry room               |
| 车库（Garage）           | 257664003   | Garage                     |
| 院子（Yard）             | 257673000   | Yard                       |
| 床边（Bedside）          | 257880003   | At bedside（医疗也可用）    |

### 5.2 医疗类环境（Hospital / Clinical Environment）

| 环境                         | SNOMED Code | 描述                    |
|:----------------------------|:------------|:-----------------------|
| 医院 (Hospital)             | 22232009    | Hospital               |
| 病房 (Ward)                 | 309900005   | Hospital ward          |
| 病房区（Inpatient unit）    | 309915005   | Inpatient unit         |
| 急诊科（Emergency department）| 113858008  | Emergency department   |
| ICU（Intensive care unit）  | 309904001   | ICU                    |
| 观察室（Observation room）  | 309912008   | Observation room       |
| 诊室（Consulting room）     | 309898003   | Consulting room        |
| 候诊区（Waiting room）      | 309898003   | Waiting room           |

### 5.3 养老院 / 护理环境（Long-term care）

| 环境                             | SNOMED Code | 描述                    |
|:--------------------------------|:------------|:-----------------------|
| 养老院 (Nursing home)           | 225368008   | Nursing home           |
| 长期护理机构（Long term care facility）| 42665001 | Long-term care facility|
| 护理房间（Care home room）      | 257667001   | 可重复使用 Bedroom      |
| 生活区（Lounge / Communal area）| 257669003   | Living/Lounge          |
| 活动区（Activity room）         | 257918008   | Activity area          |

### 5.4 户外 / 公共环境

| 环境               | SNOMED Code | 描述                |
|:-------------------|:------------|:-------------------|
| 室外（Outdoor）    | 260787004   | Outdoor environment|
| 公园（Park）       | 257672007   | Park               |
| 街道（Street）     | 257914005   | Street             |
| 商店（Store/Shop） | 27250006    | Shop               |
| 建筑物（Building） | 257914005   | Building（泛用）    |

### 5.5 常用场景位置编码

| 场景                     | SNOMED Code | 说明                |
|:------------------------|:------------|:-------------------|
| 床上（In bed）          | 248569007   | 用于上床状态        |
| 不在床上（Not in bed）  | 248570008   | 用于离床状态        |
| 床边（At bedside）      | 257880003   | 床附近活动区域      |
| 浴室（Bathroom）        | 77605003    | 卫生间滞留事件      |
| 厕所（Toilet room）     | 257914005   | 更精确              |
| 走廊（Corridor）        | 257915006   | 跌倒检测常用        |
| 客厅（Living room）     | 257669003   | 日常活动区域        |
| 卧室（Bedroom）         | 257667001   | 夜间状况监测        |
| 厨房（Kitchen）         | 257670002   | 烟雾/活动监测       |

---
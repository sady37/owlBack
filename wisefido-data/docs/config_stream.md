1. 基础事件格式
```json
{
  "tenant_id": "租户ID",
  "topic_type": "config",
  "category": "alarm/device/address/user/resident",
  "operate": "create/update/delete",
  "data_value": {
    // 具体变更数据，根据 category 不同而不同
  }
}
```

2. 各类型具体格式
2.1 Device 配置变更 (category: "device")
```Json

{
  "tenant_id": "xxx",
  "topic_type": "config",
  "category": "device",
  "operate": "update",
  "data_value": {
    "device_id": "设备ID",
    "device_uid": "设备UID",
    "device_serial": "设备序列号"
  }
}

```

2.2 Alarm 配置变更 (category: "alarm")

````Json

{
  "tenant_id": "xxx",
  "topic_type": "config",
  "category": "alarm",
  "operate": "update",
  "data_value": {
    "type": "alarm_cloud/alarm_device",
    "affected_devices": [
      {
        "device_id": "设备ID",
        "device_uid": "设备UID",
        "device_serial": "设备序列号"
      }
    ]
  }
}

```


2.3 Address 配置变更 (category: "address")

```Json


{
  "tenant_id": "xxx",
  "topic_type": "config",
  "category": "address",
  "operate": "update",
  "data_value": {
    "address_type": "branch/building/unit/room/bed",
    "address_id": "地址ID",
  }
}

```


2.4 User 配置变更 (category: "user")

```Json

Apply
{
  "tenant_id": "xxx",
  "topic_type": "config",
  "category": "user",
  "operate": "update",

  "data_value": {
  "user_ids": ["用户ID1", "用户ID2"]
  }
}

```


2.5 Resident 配置变更 (category: "resident")

```Json

Apply
{
  "tenant_id": "xxx",
  "topic_type": "config",
  "category": "resident",
  "operate": "update",

  "data_value": {
    "resident_idsis": ["住户ID1", "住户ID2"
      }
    
  }

```


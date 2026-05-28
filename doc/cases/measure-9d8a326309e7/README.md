# Measure Session Fixture: 9D8A326309E7

**设备**：9D8A326309E7（单圈走法测试设备）
**用途**：S1+S2+S3+S4 单圈数据回归测试

## 录制 sessions

| intent | walking pattern | file |
|---|---|---|
| wall | 单圈绕房间一周 | `wall_single.json` |
| bed | 单圈绕床一周 | `bed_single.json` |

## Dot 格式 (v2)

参见 [owlFront/docs/measure_fit_algorithm.md § Dot 数据格式 (v2)](../../../../owlFront/docs/measure_fit_algorithm.md#dot-数据格式-v2)。

每条 dot：
```json
{ "time": 1791234567, "x": 0, "y": 0, "z": 0, "d": 0, "E": 0 }
```

- `time`: UTC 秒（整数）
- `x, y, z`: canvas 坐标 cm
- `d`: 距上一秒空间距离 cm；`-1` = 仍在盲区
- `E`: 0 / 1，盲区边界标记

## Top-level JSON schema

```json
{
  "device": "9D8A326309E7",
  "captured_at": "ISO 8601",
  "note": "...",
  "intent": "wall" | "bed",
  "dots": [Dot, Dot, ...]
}
```

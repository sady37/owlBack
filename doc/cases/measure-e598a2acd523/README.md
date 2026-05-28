# Measure Session Fixture: E598A2ACD523

**设备**：E598A2ACD523（多圈走法测试设备，含 enter/Aisle 复杂 case）
**用途**：S1+S2+S3+S4 多圈/折返/盲区数据回归测试

## 录制 sessions

| intent | walking pattern | file |
|---|---|---|
| wall | 多圈绕房间，可能含盲区段 | `wall_multi.json` |
| other (table) | 多圈绕 table | `table_multi.json` |
| enter | 多次穿门（含盲区） | `enter_multi.json` |

## Dot 格式 (v2)

同 [9D8A326309E7 README](../measure-9d8a326309e7/README.md)。

特别注意：本设备数据可能含**盲区段**：
- 连续若干秒 `d=-1, x=y=z=0, E=1` 表示走出 FoV
- 进入盲区前最后一个 dot 的 `E=1`（边界）
- 返回 FoV 第一个 dot 的 `E=1`，其 `d` = 距进入盲区前最后有效 dot 的空间距离

## Top-level JSON schema

```json
{
  "device": "E598A2ACD523",
  "captured_at": "ISO 8601",
  "note": "...",
  "intent": "wall" | "other" | "enter",
  "dots": [Dot, Dot, ...]
}
```

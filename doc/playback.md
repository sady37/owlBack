# Room Playback

一条命令出 HTML，浏览器打开看 cell 染色 + 时间轴。脚本自动启动 dev API、拉数据、统计颜色。

## 用法

```bash
scripts/playback.sh <device_uid> <start> <end> [snap_min]
```

时间格式可选 RFC3339（带时区）或 `'YYYY-MM-DD HH:MM'`（当前时区）。`snap_min` 默认 30。

## 示例

```bash
# Radar_333B 今晨 1:00-8:00，每 30min 一帧
scripts/playback.sh 25A859B8333B "2026-05-05 01:00" "2026-05-05 08:00"

# 带时区的明确写法
scripts/playback.sh 25A859B8333B \
  "2026-05-05T01:00:00-07:00" "2026-05-05T08:00:00-07:00" 10
```

## 输出

文件落到 `owlBack/out/playback_<uid>_<yyyymmdd-hhmm>.html`（已 gitignore）。

```
[playback] uid=25A859B8333B  2026-05-05T01:00:00-07:00 → 2026-05-05T08:00:00-07:00  snap_min=30
[playback] saved /home/wisefido/owl/owlBack/out/playback_25A859B8333B_20260505-0100.html  (3794010 bytes)
  cells x snaps =   52350
  !InRoom (Out)    14235   27%
  InRoom           38115   73%
  Walk (#ffffff)     793   ← cell_learning 累积出的活动区
  Enter (#44cc66)    345
  Bed (#4488dd)       15
  Sit (#ff9933)       15

[playback] open in browser:
  xdg-open /home/wisefido/owl/owlBack/out/playback_25A859B8333B_20260505-0100.html
  file:///home/wisefido/owl/owlBack/out/playback_25A859B8333B_20260505-0100.html
[playback] list all:  ls -lh /home/wisefido/owl/owlBack/out/playback_*.html
```

通过环境变量改输出位置：`OUT_DIR=/some/path scripts/playback.sh ...`

## 颜色快速对照

| 色 | 含义 |
|---|---|
| `#2a2a2a` 深灰 | **!InRoom**（边界外） |
| `#c8c8c8` 浅灰 | 雷达盲区 / 还没学到（AreaUnknown） |
| `#ffffff` 白 | **Walk**（cell_learning 学到的活动区） |
| `#44cc66` 绿 | Enter（门洞） |
| `#4488dd` 蓝 | Bed |
| `#ff9933` 橙 | Sit（沙发/椅子/马桶） |
| `#4a4a4a` 中灰 | Furniture（layout 标的家具 Deny） |

线条层：黑实线=Wall（无 wall 时降级 boundary 兜底） / 蓝虚线=雷达 FOV / 绿虚线=Enter 矩形。

## 故障排查

| 现象 | 原因 |
|---|---|
| `playback timeout (60s)` | 时间窗太长 → 缩窄 / 加大 `snap_min` |
| `(no fills found ...)` | 时间窗内 monitor 数据为空，设备离线 |
| 全图 100% Out | layout 没画 wall **且** 没 boundary 兜底 → 检查 `boundaryPolygonForStamp` |
| 0 个 Walk cells | layout/InRoom 失效，cell_learning 没跑 |

## 内部细节（按需查阅）

- 脚本: [scripts/playback.sh](../scripts/playback.sh)
- HTTP API: [cmd/roomengine-api](../wisefido-sensor/cmd/roomengine-api/main.go) 监听 `:7788`
- 渲染: [internal/roomengine/room_svg.go](../wisefido-sensor/internal/roomengine/room_svg.go)
- Wall fallback: [internal/roomengine/layout_parser.go](../wisefido-sensor/internal/roomengine/layout_parser.go) `boundaryPolygonForStamp`

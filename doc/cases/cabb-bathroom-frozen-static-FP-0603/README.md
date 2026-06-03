# CABB 浴室 lost_track 误报 — frozen-static-target（每日，2026-06-03）

## 一句话

人进浴室、走到雷达近旁某点站定（洗手台/镜前），**firmware 把轨迹冻结在一个精确不变的坐标**（丢了真信号还续报最后位置），~27s 后 drop（np=0）；人没走出门（**无 ExitRoom**），engine 把"轨迹消失"当跌倒 → lost_track CRITICAL。**与 D5F7（反射/换ID）、MoM（走出残影）都不同**——这是**静止目标被丢锁的冻结轨迹**。

## 设备

CABB `4D8710D5CABB` = `fd00:0:3:411:1:200:10d5:cabb`，**ceiling 吊顶装 z=240cm**（geom x=-50,y=250,z=240）。

## 最近 3 次 fall（全 lost_track，全近距）

| 触发(Shanghai) | reason | 锚点 | 距雷达 |
|---|---|---|---|
| 06-03 15:58:16 | lost_track | (-60,10) | 61cm |
| 06-03 12:09:24 | lost_track | (0,-20) | 20cm |
| 06-02 14:34:33 | lost_track | (-10,-30) | 32cm |

## 文件

| 文件 | 内容 |
|---|---|
| `timeline_FP_1558.txt` | FP 完整时间线（62 行，07:55–08:04 UTC）|
| `timeline_normal_0643.txt` | 正常访问对照（263 行，人一直被追 + ExitRoom）|
| `room_layout.json` | CABB /128 layout（ceiling z=240）|
| `alarm_last3.json` | 最近 3 条 fall payload |

## FP 铁证（timeline_FP_1558.txt）

```
15:55:30  EnterRoom np=1   track0 (0,-30,z62)→(-20,-10,z46)   z=高度,正常
15:55:31  track0 (-60,10,z=0)
15:55:31–57  逐帧 (-60,10,0) p4 c80 **精确一模一样 27 秒**     ← firmware 冻结轨迹
15:55:58  np=0
15:56:00–16:04  纯 88 空心跳，无 ExitRoom，无复现             → lost_track 误报
```

## 判别器（FP vs 正常）—— event_log 已验证

| | EnterRoom | 轨迹 | ExitRoom | 结果 |
|---|---|---|---|---|
| **正常访问**（几十条）| 有 | 位置**持续抖动变化**（密集帧）| **有** | 无 fall |
| **FP**（3 条）| 有 | 位置**精确冻结**（零方差 ≥20s）+ z 掉 0 | **无** | lost_track |

近 3 天 EnterRoom/ExitRoom 列表里：FP 的 enter 全部**没有配对 ExitRoom**（人被丢锁、没记录到过门离开）；正常访问全是 Enter→Exit 配对。

## 根因

ceiling 雷达，人走到近旁站定（洗漱），**毫米波丢静止目标**（无微动 Doppler）→ firmware 冻结轨迹在精确坐标（z 掉 0）→ 续报 ~27s 后 drop（np=0）。人**没过门**（无 ExitRoom），但 engine 的 lost_track 凭"轨迹消失 + 无 ExitRoom + 无 recovery" 推断跌倒。pose 全程 4（firmware **从没**给跌倒 pose 2/5）。

## 处置决定（2026-06-03，用户拍板）：**不抑制、不降级，保留告警**

**雷达突然丢信号，真跌倒与信号丢失在数据上不可分**——真摔倒同样表现为突然丢锁。人也没法判断是真 Fall 还是信号丢失，故标成 fall 是诚实且安全的（宁可误报不可漏报）。

因此**故意不做** frozen-static-target 抑制闸（曾提议"零方差冻结→抑制"，但那会把一类**可能的真跌倒**也压掉 = 漏报，不可接受）；也不降级。这类"近距 + 冻结 + 无 ExitRoom 的纯丢锁"**保留 CRITICAL 告警**。

**这是雷达硬件/安装固有局限**（ceiling 雷达丢近旁静止目标），不是软件可判的逻辑 bug。真正缓解只能走硬件层：调安装位/参数避开易丢区，或加门磁/sleepad 多源佐证。本 case 作"已知不可分 FP"归档，**不要再提抑制**。

> 与 MoM（走出残影，**已被空房账闸解决**）、D5F7（反射换ID，已解）不同：CABB 这类是**人没走 + 静止丢锁**，无 ExitRoom 可依、无法证伪 → 安全起见保留报警。
> 关联 [[cabb_ghost_fall_cases]]、[[fall_rules_three_classes]]、memory `lost_track_fall_detection_envelope_gate`。

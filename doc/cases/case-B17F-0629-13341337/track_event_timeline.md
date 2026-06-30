# case-B17F-0629-13341337 — 每 tick belief 时间线 (room fd00:0:3:111:2:300, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
13:34:00 B17F.1   B17F13400714  walk    80   NoReport walk               trk  0.81 Empty      2   0     0.00  0.02  0.26  0.00  0.69  0.03
13:34:00 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.50 Empty      2   0     0.00  0.03  0.08  0.00  0.85  0.04
13:34:01 B17F.1   B17F13400714  walk    68   NoReport walk               trk  0.82 Empty      2   0     0.00  0.02  0.52  0.00  0.40  0.01
13:34:01 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.51 Empty      2   0     0.00  0.04  0.09  0.00  0.78  0.01
13:34:02 B17F.1   B17F13400714  walk    78   NoReport walk               trk  0.96 Empty      2   0     0.00  0.02  0.70  0.00  0.18  0.02
13:34:02 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.52 Empty      2   0     0.00  0.04  0.10  0.01  0.62  0.01
13:34:03 B17F.1   B17F13400714  walk    85   NoReport walk               trk  0.99 Sit        2   0     0.00  0.02  0.80  0.00  0.07  0.02
13:34:03 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.53 Sit        2   0     0.00  0.03  0.08  0.01  0.39  0.01
13:34:04 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.54 Sit        2   0     0.00  0.02  0.06  0.02  0.11  0.01
13:34:04 B17F.1   B17F13400714  walk    99   NoReport walk               trk  1.00 Sit        2   0     0.00  0.03  0.80  0.00  0.03  0.02
13:34:05 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 Sit        2   0     0.00  0.01  0.04  0.01  0.04  0.01
13:34:05 B17F.1   B17F13400714  walk    88   NoReport walk               trk  1.00 Sit        2   0     0.00  0.02  0.83  0.00  0.02  0.02
13:34:06 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 Sit        2   0     0.00  0.01  0.04  0.01  0.01  0.01
13:34:06 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 Sit        2   0     0.00  0.02  0.84  0.00  0.01  0.02
13:34:07 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.01  0.04  0.01  0.01  0.01
13:34:07 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:08 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:08 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:09 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 Sit        2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:09 B17F.1   B17F13400714  stand   106  NoReport stand              trk  1.00 Sit        2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:10 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 Sit        2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:10 B17F.1   B17F13400714  stand   98   NoReport stand              trk  1.00 Sit        2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:11 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 Sit        2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:11 B17F.1   B17F13400714  stand   96   NoReport stand              trk  1.00 Sit        2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:12 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 Sit        2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:12 B17F.1   B17F13400714  stand   83   NoReport stand              trk  1.00 Sit        2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:13 B17F.0   B17F03400714  sit     84   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.13  0.01  0.00  0.01
13:34:13 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:14 B17F.0   B17F03400714  walk    83   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.01  0.46  0.02  0.01  0.02
13:34:14 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:15 B17F.0   B17F03400714  walk    80   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.58  0.01  0.01  0.02
13:34:15 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:16 B17F.0   B17F03400714  walk    61   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.55  0.01  0.01  0.02
13:34:16 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:17 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:17 B17F.0   B17F03400714  walk    77   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.64  0.01  0.01  0.02
13:34:18 B17F.1   B17F13400714  stand   76   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:18 B17F.0   B17F03400714  walk    0    NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.60  0.01  0.01  0.02
13:34:19 B17F.0   B17F03400714  sit     93   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.01  0.27  0.00  0.00  0.01
13:34:19 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:20 B17F.0   B17F03400714  walk    84   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.01  0.29  0.01  0.00  0.01
13:34:20 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:21 B17F.0   B17F03400714  walk    91   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.01  0.31  0.01  0.00  0.01
13:34:21 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:22 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.32  0.01  0.00  0.01
13:34:22 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:23 B17F.0   B17F03400714  stand   89   NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.33  0.01  0.00  0.01
13:34:23 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:24 B17F.0   B17F03400714  stand   91   NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.34  0.01  0.00  0.01
13:34:24 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:25 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.35  0.01  0.00  0.01
13:34:25 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:26 B17F.0   B17F03400714  stand   87   NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.35  0.01  0.00  0.01
13:34:26 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:27 B17F.0   B17F03400714  stand   78   NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.50  0.01  0.01  0.02
13:34:27 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:28 B17F.0   B17F03400714  stand   66   NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.71  0.01  0.01  0.02
13:34:28 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:29 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.66  0.01  0.01  0.02
13:34:29 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:30 B17F.0   B17F03400714  stand   73   NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.61  0.00  0.01  0.02
13:34:30 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:31 B17F.0   B17F03400714  stand   85   NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.57  0.01  0.01  0.02
13:34:31 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:32 B17F.0   B17F03400714  stand   70   NoReport stand              trk  0.55 OpenFloor  2   13    0.00  0.02  0.65  0.01  0.01  0.02
13:34:32 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
13:34:33 B17F.0   B17F03400714  stand   77   NoReport stand              trk  0.55 OpenFloor  2   14    0.00  0.02  0.78  0.01  0.01  0.02
13:34:33 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
13:34:34 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   15    0.00  0.02  0.72  0.00  0.01  0.02
13:34:34 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
13:34:35 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   16    0.00  0.02  0.66  0.00  0.01  0.02
13:34:35 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
13:34:36 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   17    0.00  0.02  0.61  0.00  0.01  0.02
13:34:36 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
13:34:37 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
13:34:37 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   18    0.00  0.02  0.57  0.00  0.01  0.02
13:34:38 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
13:34:38 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   19    0.00  0.01  0.53  0.01  0.01  0.01
13:34:39 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:34:39 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   20    0.00  0.01  0.49  0.01  0.01  0.01
13:34:40 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
13:34:40 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   21    0.00  0.01  0.47  0.01  0.01  0.01
13:34:41 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
13:34:41 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   22    0.00  0.01  0.45  0.01  0.01  0.01
13:34:42 B17F.0   B17F03400714  stand   76   NoReport stand              trk  0.55 OpenFloor  2   23    0.00  0.02  0.57  0.01  0.01  0.02
13:34:42 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
13:34:43 B17F.0   B17F03400714  stand   85   NoReport stand              trk  0.55 OpenFloor  2   24    0.00  0.01  0.53  0.01  0.01  0.02
13:34:43 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   24    0.00  0.02  0.85  0.00  0.01  0.02
13:34:44 B17F.0   B17F03400714  stand   71   NoReport stand              trk  0.55 OpenFloor  2   25    0.00  0.01  0.50  0.01  0.01  0.01
13:34:44 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
13:34:45 B17F.0   B17F03400714  stand   77   NoReport stand              trk  0.55 OpenFloor  2   26    0.00  0.01  0.47  0.01  0.01  0.01
13:34:45 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:34:46 B17F.0   B17F03400714  stand   81   NoReport stand              trk  0.55 OpenFloor  2   27    0.00  0.02  0.70  0.01  0.01  0.02
13:34:46 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
13:34:47 B17F.0   B17F03400714  stand   83   NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.80  0.01  0.01  0.02
13:34:47 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:48 B17F.0   B17F03400714  walk    89   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.83  0.00  0.01  0.02
13:34:48 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:49 B17F.0   B17F03400714  walk    87   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
13:34:49 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:50 B17F.0   B17F03400714  walk    78   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:50 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:51 B17F.0   B17F03400714  walk    62   NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.78  0.00  0.01  0.02
13:34:51 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:52 B17F.0   B17F03400714  walk    0    NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.72  0.00  0.01  0.02
13:34:52 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:53 B17F.0   B17F03400714  walk    0    NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.66  0.00  0.01  0.02
13:34:53 B17F.1   B17F13400714  stand   92   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:54 B17F.0   B17F03400714  sit     81   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.01  0.32  0.00  0.00  0.01
13:34:54 B17F.1   B17F13400714  stand   82   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:55 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.03  0.00  0.00  0.00
13:34:55 B17F.1   B17F13400714  stand   71   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:56 B17F.0   B17F03400714  sit     79   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.02  0.00  0.00  0.00
13:34:56 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:57 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:57 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.03  0.01  0.00  0.01
13:34:58 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:34:58 B17F.0   B17F03400714  sit     77   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.01  0.00  0.00  0.00
13:34:59 B17F.0   B17F03400714  sit     0    NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.01  0.00  0.00  0.00
13:34:59 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:00 B17F.0   B17F03400714  sit     76   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.03  0.01  0.00  0.01
13:35:00 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:01 B17F.0   B17F03400714  sit     78   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.04  0.01  0.00  0.01
13:35:01 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:02 B17F.0   B17F03400714  sit     74   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.02  0.01  0.00  0.00
13:35:02 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:03 B17F.0   B17F03400714  sit     69   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.01  0.00  0.00  0.00
13:35:03 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:04 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.00  0.02  0.02
13:35:04 B17F.0   B17F03400714  sit     79   NoReport sit                trk  0.55 OpenFloor  2   0     0.00  0.00  0.01  0.00  0.00  0.00
13:35:05 B17F.1   B17F13400714  stand   63   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
13:35:05 B17F.0   B17F03400714  walk    0    NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.00  0.06  0.01  0.00  0.01
13:35:06 B17F.1   B17F13400714  walk    35   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:06 B17F.0   B17F03400714  walk    0    NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.01  0.38  0.03  0.01  0.02
13:35:07 B17F.1   B17F13400714  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:07 B17F.0   B17F03400714  walk    0    NoReport walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.65  0.02  0.01  0.02
13:35:08 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:08 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.78  0.01  0.01  0.02
13:35:09 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:09 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.83  0.01  0.01  0.02
13:35:10 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:10 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
13:35:11 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:11 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:12 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:12 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:13 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:13 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:14 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:14 B17F.0   B17F03400714  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:15 B17F.E   -             -       0    NoReport np=3               room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:35:15 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.92  0.00  0.01  0.01
13:35:15 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.92  0.00  0.01  0.01
13:35:15 B17F.2   B17F23515081  stand   79   NoReport stand              trk  0.50 OpenFloor  3   0     0.00  0.03  0.15  0.00  0.79  0.04
13:35:16 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:35:16 B17F.2   B17F23515081  stand   0    NoReport stand              trk  0.51 OpenFloor  3   0     0.00  0.02  0.54  0.00  0.36  0.01
13:35:16 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:35:17 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:35:17 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:35:17 B17F.2   B17F23515081  stand   82   NoReport stand              trk  0.52 OpenFloor  3   0     0.00  0.01  0.71  0.00  0.08  0.01
13:35:18 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   11    0.00  0.01  0.93  0.00  0.00  0.01
13:35:18 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   11    0.00  0.01  0.93  0.00  0.00  0.01
13:35:18 B17F.2   B17F23515081  stand   0    NoReport stand              trk  0.53 OpenFloor  3   11    0.00  0.01  0.86  0.00  0.02  0.01
13:35:19 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   12    0.00  0.01  0.93  0.00  0.00  0.01
13:35:19 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   12    0.00  0.01  0.93  0.00  0.00  0.01
13:35:19 B17F.2   B17F23515081  stand   0    NoReport stand              trk  0.54 OpenFloor  3   12    0.00  0.01  0.80  0.00  0.00  0.01
13:35:20 B17F.2   B17F23515081  stand   0    NoReport stand              trk  0.55 OpenFloor  3   13    0.00  0.01  0.73  0.00  0.00  0.01
13:35:20 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   13    0.00  0.01  0.93  0.00  0.00  0.01
13:35:20 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   13    0.00  0.01  0.93  0.00  0.00  0.01
13:35:21 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.01  0.67  0.00  0.00  0.01
13:35:21 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.01  0.93  0.00  0.00  0.01
13:35:21 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.01  0.93  0.00  0.00  0.01
13:35:22 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.01  0.61  0.00  0.00  0.01
13:35:22 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.01  0.93  0.00  0.00  0.01
13:35:22 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.01  0.93  0.00  0.00  0.01
13:35:23 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.01  0.93  0.00  0.00  0.01
13:35:23 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.01  0.93  0.00  0.00  0.01
13:35:23 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.01  0.56  0.00  0.00  0.01
13:35:23 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.01  0.52  0.00  0.00  0.01
13:35:23 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.01  0.93  0.00  0.00  0.01
13:35:23 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.01  0.93  0.00  0.00  0.01
13:35:24 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.01  0.49  0.00  0.00  0.01
13:35:24 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.01  0.93  0.00  0.00  0.01
13:35:24 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.01  0.93  0.00  0.00  0.01
13:35:25 B17F.2   B17F23515081  stand   76   NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.47  0.00  0.00  0.01
13:35:25 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.93  0.00  0.00  0.01
13:35:25 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.93  0.00  0.00  0.01
13:35:26 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.01  0.45  0.00  0.00  0.01
13:35:26 B17F.1   B17F13400714  stand   64   NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.01  0.93  0.00  0.00  0.01
13:35:26 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.01  0.93  0.00  0.00  0.01
13:35:27 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.00  0.43  0.00  0.00  0.01
13:35:27 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.01  0.93  0.00  0.00  0.01
13:35:27 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.01  0.93  0.00  0.00  0.01
13:35:28 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.00  0.42  0.00  0.00  0.01
13:35:28 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.01  0.93  0.00  0.00  0.01
13:35:28 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.01  0.93  0.00  0.00  0.01
13:35:29 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.01  0.93  0.00  0.00  0.01
13:35:29 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.01  0.93  0.00  0.00  0.01
13:35:29 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.00  0.41  0.00  0.00  0.01
13:35:30 B17F.2   B17F23515081  stand   81   NoReport stand              trk  1.00 OpenFloor  3   24    0.00  0.00  0.40  0.00  0.00  0.01
13:35:30 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   24    0.00  0.01  0.93  0.00  0.00  0.01
13:35:30 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   24    0.00  0.01  0.93  0.00  0.00  0.01
13:35:31 B17F.2   B17F23515081  stand   76   NoReport stand              trk  1.00 OpenFloor  3   25    0.00  0.00  0.39  0.00  0.00  0.01
13:35:31 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   25    0.00  0.01  0.93  0.00  0.00  0.01
13:35:31 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   25    0.00  0.01  0.93  0.00  0.00  0.01
13:35:32 B17F.2   B17F23515081  stand   79   NoReport stand              trk  1.00 OpenFloor  3   26    0.00  0.01  0.69  0.01  0.00  0.01
13:35:32 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
13:35:32 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
13:35:33 B17F.2   B17F23515081  stand   83   NoReport stand              trk  1.00 OpenFloor  3   27    0.00  0.01  0.64  0.00  0.00  0.01
13:35:33 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   27    0.00  0.01  0.93  0.00  0.00  0.01
13:35:33 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   27    0.00  0.01  0.93  0.00  0.00  0.01
13:35:34 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   28    0.00  0.01  0.82  0.00  0.00  0.01
13:35:34 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   28    0.00  0.01  0.93  0.00  0.00  0.01
13:35:34 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   28    0.00  0.01  0.93  0.00  0.00  0.01
13:35:35 B17F.2   B17F23515081  stand   76   NoReport stand              trk  1.00 OpenFloor  3   29    0.00  0.01  0.75  0.00  0.00  0.01
13:35:35 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   29    0.00  0.01  0.93  0.00  0.00  0.01
13:35:35 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   29    0.00  0.01  0.93  0.00  0.00  0.01
13:35:36 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   30    0.00  0.01  0.68  0.00  0.00  0.01
13:35:36 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   30    0.00  0.01  0.93  0.00  0.00  0.01
13:35:36 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   30    0.00  0.01  0.93  0.00  0.00  0.01
13:35:37 B17F.2   B17F23515081  stand   86   NoReport stand              trk  1.00 OpenFloor  3   31    0.00  0.01  0.75  0.00  0.00  0.01
13:35:37 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   31    0.00  0.01  0.93  0.00  0.00  0.01
13:35:37 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   31    0.00  0.01  0.93  0.00  0.00  0.01
13:35:38 B17F.2   B17F23515081  stand   78   NoReport stand              trk  1.00 OpenFloor  3   32    0.00  0.01  0.86  0.00  0.00  0.01
13:35:38 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   32    0.00  0.01  0.93  0.00  0.00  0.01
13:35:38 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   32    0.00  0.01  0.93  0.00  0.00  0.01
13:35:39 B17F.2   B17F23515081  stand   74   NoReport stand              trk  1.00 OpenFloor  3   33    0.00  0.01  0.79  0.00  0.00  0.01
13:35:39 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   33    0.00  0.01  0.93  0.00  0.00  0.01
13:35:39 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   33    0.00  0.01  0.93  0.00  0.00  0.01
13:35:40 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.01  0.91  0.00  0.00  0.01
13:35:40 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.01  0.93  0.00  0.00  0.01
13:35:40 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.01  0.93  0.00  0.00  0.01
13:35:41 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   35    0.00  0.01  0.93  0.00  0.00  0.01
13:35:41 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   35    0.00  0.01  0.93  0.00  0.00  0.01
13:35:41 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   35    0.00  0.01  0.91  0.00  0.00  0.01
13:35:42 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   36    0.00  0.01  0.93  0.00  0.00  0.01
13:35:42 B17F.2   B17F23515081  stand   84   NoReport stand              trk  1.00 OpenFloor  3   36    0.00  0.01  0.84  0.00  0.00  0.01
13:35:42 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   36    0.00  0.01  0.93  0.00  0.00  0.01
13:35:43 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   37    0.00  0.01  0.89  0.00  0.00  0.01
13:35:43 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   37    0.00  0.01  0.93  0.00  0.00  0.01
13:35:43 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   37    0.00  0.01  0.93  0.00  0.00  0.01
13:35:44 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   38    0.00  0.01  0.82  0.00  0.00  0.01
13:35:44 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   38    0.00  0.01  0.93  0.00  0.00  0.01
13:35:44 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   38    0.00  0.01  0.93  0.00  0.00  0.01
13:35:45 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   39    0.00  0.01  0.93  0.00  0.00  0.01
13:35:45 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   39    0.00  0.01  0.93  0.00  0.00  0.01
13:35:45 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   39    0.00  0.01  0.75  0.00  0.00  0.01
13:35:46 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   40    0.00  0.01  0.68  0.00  0.00  0.01
13:35:46 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   40    0.00  0.01  0.93  0.00  0.00  0.01
13:35:46 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   40    0.00  0.01  0.93  0.00  0.00  0.01
13:35:47 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   41    0.00  0.01  0.63  0.00  0.00  0.01
13:35:47 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   41    0.00  0.01  0.93  0.00  0.00  0.01
13:35:47 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   41    0.00  0.01  0.93  0.00  0.00  0.01
13:35:48 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   42    0.00  0.01  0.58  0.00  0.00  0.01
13:35:48 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   42    0.00  0.01  0.93  0.00  0.00  0.01
13:35:48 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   42    0.00  0.01  0.93  0.00  0.00  0.01
13:35:49 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   43    0.00  0.01  0.53  0.00  0.00  0.01
13:35:49 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   43    0.00  0.01  0.93  0.00  0.00  0.01
13:35:49 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   43    0.00  0.01  0.93  0.00  0.00  0.01
13:35:50 B17F.2   B17F23515081  stand   82   NoReport stand              trk  1.00 OpenFloor  3   44    0.00  0.01  0.65  0.00  0.00  0.01
13:35:50 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   44    0.00  0.01  0.93  0.00  0.00  0.01
13:35:50 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   44    0.00  0.01  0.93  0.00  0.00  0.01
13:35:51 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.88  0.00  0.00  0.01
13:35:51 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.93  0.00  0.00  0.01
13:35:51 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.93  0.00  0.00  0.01
13:35:52 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.87  0.00  0.00  0.01
13:35:52 B17F.1   B17F13400714  stand   91   NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
13:35:52 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
13:35:53 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.01  0.93  0.00  0.00  0.01
13:35:53 B17F.1   B17F13400714  stand   100  NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.01  0.93  0.00  0.00  0.01
13:35:53 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.01  0.80  0.00  0.00  0.01
13:35:54 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.01  0.73  0.00  0.00  0.01
13:35:54 B17F.1   B17F13400714  stand   68   NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.01  0.93  0.00  0.00  0.01
13:35:54 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.01  0.93  0.00  0.00  0.01
13:35:55 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.01  0.93  0.00  0.00  0.01
13:35:55 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.01  0.67  0.00  0.00  0.01
13:35:55 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.01  0.93  0.00  0.00  0.01
13:35:56 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.01  0.83  0.00  0.00  0.01
13:35:56 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.01  0.93  0.00  0.00  0.01
13:35:56 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.01  0.93  0.00  0.00  0.01
13:35:57 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.93  0.00  0.00  0.01
13:35:57 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.93  0.00  0.00  0.01
13:35:57 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.91  0.00  0.00  0.01
13:35:58 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.01  0.93  0.00  0.00  0.01
13:35:58 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.01  0.93  0.00  0.00  0.01
13:35:58 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.01  0.93  0.00  0.00  0.01
13:35:59 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.01  0.93  0.00  0.00  0.01
13:35:59 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.01  0.93  0.00  0.00  0.01
13:35:59 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.01  0.93  0.00  0.00  0.01
13:36:00 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.01  0.93  0.00  0.00  0.01
13:36:00 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.01  0.93  0.00  0.00  0.01
13:36:00 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.01  0.93  0.00  0.00  0.01
13:36:01 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.01  0.93  0.00  0.00  0.01
13:36:01 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.01  0.93  0.00  0.00  0.01
13:36:01 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.01  0.93  0.00  0.00  0.01
13:36:02 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   24    0.00  0.01  0.93  0.00  0.00  0.01
13:36:02 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   24    0.00  0.01  0.93  0.00  0.00  0.01
13:36:02 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   24    0.00  0.01  0.93  0.00  0.00  0.01
13:36:03 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   25    0.00  0.01  0.90  0.00  0.01  0.01
13:36:03 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   25    0.00  0.01  0.90  0.00  0.01  0.01
13:36:03 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   25    0.00  0.01  0.90  0.00  0.01  0.01
13:36:04 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
13:36:04 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
13:36:04 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
13:36:05 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   27    0.00  0.01  0.93  0.00  0.00  0.01
13:36:05 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   27    0.00  0.01  0.93  0.00  0.00  0.01
13:36:05 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   27    0.00  0.01  0.92  0.00  0.00  0.01
13:36:06 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   28    0.00  0.01  0.93  0.00  0.00  0.01
13:36:06 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   28    0.00  0.01  0.85  0.00  0.00  0.01
13:36:06 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   28    0.00  0.01  0.93  0.00  0.00  0.01
13:36:07 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   29    0.00  0.01  0.93  0.00  0.00  0.01
13:36:07 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   29    0.00  0.01  0.93  0.00  0.00  0.01
13:36:07 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   29    0.00  0.01  0.77  0.00  0.00  0.01
13:36:08 B17F.2   B17F23515081  stand   84   NoReport stand              trk  1.00 OpenFloor  3   30    0.00  0.01  0.71  0.00  0.00  0.01
13:36:08 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   30    0.00  0.01  0.93  0.00  0.00  0.01
13:36:08 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   30    0.00  0.01  0.93  0.00  0.00  0.01
13:36:09 B17F.2   B17F23515081  stand   85   NoReport stand              trk  1.00 OpenFloor  3   31    0.00  0.01  0.65  0.00  0.00  0.01
13:36:09 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   31    0.00  0.01  0.93  0.00  0.00  0.01
13:36:09 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   31    0.00  0.01  0.93  0.00  0.00  0.01
13:36:10 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   32    0.00  0.01  0.93  0.00  0.00  0.01
13:36:10 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   32    0.00  0.01  0.93  0.00  0.00  0.01
13:36:10 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   32    0.00  0.01  0.59  0.00  0.00  0.01
13:36:11 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   33    0.00  0.01  0.93  0.00  0.00  0.01
13:36:11 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   33    0.00  0.01  0.93  0.00  0.00  0.01
13:36:11 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   33    0.00  0.01  0.55  0.00  0.00  0.01
13:36:12 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.01  0.93  0.00  0.00  0.01
13:36:12 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.01  0.51  0.00  0.00  0.01
13:36:12 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.01  0.93  0.00  0.00  0.01
13:36:13 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   35    0.00  0.01  0.93  0.00  0.00  0.01
13:36:13 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   35    0.00  0.01  0.93  0.00  0.00  0.01
13:36:13 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   35    0.00  0.01  0.48  0.00  0.00  0.01
13:36:14 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   36    0.00  0.01  0.93  0.00  0.00  0.01
13:36:14 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   36    0.00  0.01  0.46  0.00  0.00  0.01
13:36:14 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   36    0.00  0.01  0.93  0.00  0.00  0.01
13:36:15 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   37    0.00  0.01  0.93  0.00  0.00  0.01
13:36:15 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   37    0.00  0.01  0.44  0.00  0.00  0.01
13:36:15 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   37    0.00  0.01  0.93  0.00  0.00  0.01
13:36:16 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   38    0.00  0.01  0.93  0.00  0.00  0.01
13:36:16 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   38    0.00  0.01  0.93  0.00  0.00  0.01
13:36:16 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   38    0.00  0.00  0.42  0.00  0.00  0.01
13:36:17 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   39    0.00  0.01  0.93  0.00  0.00  0.01
13:36:17 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   39    0.00  0.01  0.93  0.00  0.00  0.01
13:36:17 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   39    0.00  0.00  0.41  0.00  0.00  0.01
13:36:18 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   40    0.00  0.01  0.93  0.00  0.00  0.01
13:36:18 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   40    0.00  0.00  0.40  0.00  0.00  0.01
13:36:18 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   40    0.00  0.01  0.93  0.00  0.00  0.01
13:36:19 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   41    0.00  0.01  0.93  0.00  0.00  0.01
13:36:19 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   41    0.00  0.01  0.93  0.00  0.00  0.01
13:36:19 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   41    0.00  0.00  0.40  0.00  0.00  0.01
13:36:20 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   42    0.00  0.01  0.93  0.00  0.00  0.01
13:36:20 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   42    0.00  0.01  0.55  0.00  0.00  0.01
13:36:20 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   42    0.00  0.01  0.93  0.00  0.00  0.01
13:36:21 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   43    0.00  0.01  0.93  0.00  0.00  0.01
13:36:21 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   43    0.00  0.01  0.51  0.00  0.00  0.01
13:36:21 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   43    0.00  0.01  0.93  0.00  0.00  0.01
13:36:22 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   44    0.00  0.01  0.93  0.00  0.00  0.01
13:36:22 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   44    0.00  0.01  0.93  0.00  0.00  0.01
13:36:22 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   44    0.00  0.01  0.48  0.00  0.00  0.01
13:36:23 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.93  0.00  0.00  0.01
13:36:23 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.93  0.00  0.00  0.01
13:36:23 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.62  0.00  0.00  0.01
13:36:24 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
13:36:24 B17F.2   B17F23515081  stand   77   NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.81  0.00  0.00  0.01
13:36:24 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
13:36:25 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   47    0.00  0.01  0.93  0.00  0.00  0.01
13:36:25 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   47    0.00  0.01  0.93  0.00  0.00  0.01
13:36:25 B17F.2   B17F23515081  stand   77   NoReport stand              trk  1.00 OpenFloor  3   47    0.00  0.01  0.74  0.00  0.00  0.01
13:36:26 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   48    0.00  0.01  0.93  0.00  0.00  0.01
13:36:26 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   48    0.00  0.01  0.93  0.00  0.00  0.01
13:36:26 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   48    0.00  0.01  0.68  0.00  0.00  0.01
13:36:27 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   49    0.00  0.01  0.93  0.00  0.00  0.01
13:36:27 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   49    0.00  0.01  0.93  0.00  0.00  0.01
13:36:27 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   49    0.00  0.01  0.62  0.00  0.00  0.01
13:36:28 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   50    0.00  0.01  0.93  0.00  0.00  0.01
13:36:28 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   50    0.00  0.01  0.93  0.00  0.00  0.01
13:36:28 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   50    0.00  0.01  0.57  0.00  0.00  0.01
13:36:29 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   51    0.00  0.01  0.93  0.00  0.00  0.01
13:36:29 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   51    0.00  0.01  0.93  0.00  0.00  0.01
13:36:29 B17F.2   B17F23515081  stand   82   NoReport stand              trk  1.00 OpenFloor  3   51    0.00  0.01  0.53  0.00  0.00  0.01
13:36:30 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   52    0.00  0.01  0.93  0.00  0.00  0.01
13:36:30 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   52    0.00  0.01  0.93  0.00  0.00  0.01
13:36:30 B17F.2   B17F23515081  stand   89   NoReport stand              trk  1.00 OpenFloor  3   52    0.00  0.01  0.50  0.00  0.00  0.01
13:36:31 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   53    0.00  0.01  0.93  0.00  0.00  0.01
13:36:31 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   53    0.00  0.01  0.93  0.00  0.00  0.01
13:36:31 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   53    0.00  0.01  0.83  0.00  0.00  0.01
13:36:32 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   54    0.00  0.01  0.93  0.00  0.00  0.01
13:36:32 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   54    0.00  0.01  0.93  0.00  0.00  0.01
13:36:32 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   54    0.00  0.01  0.77  0.00  0.00  0.01
13:36:33 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   55    0.00  0.01  0.93  0.00  0.00  0.01
13:36:33 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   55    0.00  0.01  0.93  0.00  0.00  0.01
13:36:33 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   55    0.00  0.01  0.70  0.00  0.00  0.01
13:36:34 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   56    0.00  0.01  0.93  0.00  0.00  0.01
13:36:34 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   56    0.00  0.01  0.64  0.00  0.00  0.01
13:36:34 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   56    0.00  0.01  0.93  0.00  0.00  0.01
13:36:35 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   57    0.00  0.01  0.93  0.00  0.00  0.01
13:36:35 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   57    0.00  0.01  0.93  0.00  0.00  0.01
13:36:35 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   57    0.00  0.01  0.59  0.00  0.00  0.01
13:36:36 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   58    0.00  0.01  0.93  0.00  0.00  0.01
13:36:36 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   58    0.00  0.01  0.93  0.00  0.00  0.01
13:36:36 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   58    0.00  0.01  0.55  0.00  0.00  0.01
13:36:37 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   59    0.00  0.01  0.93  0.00  0.00  0.01
13:36:37 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   59    0.00  0.01  0.93  0.00  0.00  0.01
13:36:37 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   59    0.00  0.01  0.51  0.00  0.00  0.01
13:36:38 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   60    0.00  0.01  0.93  0.00  0.00  0.01
13:36:38 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   60    0.00  0.01  0.93  0.00  0.00  0.01
13:36:38 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   60    0.00  0.01  0.48  0.00  0.00  0.01
13:36:39 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   61    0.00  0.01  0.93  0.00  0.00  0.01
13:36:39 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   61    0.00  0.01  0.93  0.00  0.00  0.01
13:36:39 B17F.2   B17F23515081  stand   80   NoReport stand              trk  1.00 OpenFloor  3   61    0.00  0.01  0.83  0.00  0.00  0.01
13:36:40 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   62    0.00  0.01  0.93  0.00  0.00  0.01
13:36:40 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   62    0.00  0.01  0.93  0.00  0.00  0.01
13:36:40 B17F.2   B17F23515081  stand   82   NoReport stand              trk  1.00 OpenFloor  3   62    0.00  0.01  0.76  0.00  0.00  0.01
13:36:41 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   63    0.00  0.01  0.93  0.00  0.00  0.01
13:36:41 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   63    0.00  0.01  0.93  0.00  0.00  0.01
13:36:41 B17F.2   B17F23515081  stand   73   NoReport stand              trk  1.00 OpenFloor  3   63    0.00  0.01  0.70  0.00  0.00  0.01
13:36:42 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   64    0.00  0.01  0.93  0.00  0.00  0.01
13:36:42 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   64    0.00  0.01  0.93  0.00  0.00  0.01
13:36:42 B17F.2   B17F23515081  stand   69   NoReport stand              trk  1.00 OpenFloor  3   64    0.00  0.01  0.84  0.00  0.00  0.01
13:36:43 B17F.1   B17F13400714  stand   57   NoReport stand              trk  1.00 OpenFloor  3   65    0.00  0.01  0.93  0.00  0.00  0.01
13:36:43 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   65    0.00  0.01  0.93  0.00  0.00  0.01
13:36:43 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   65    0.00  0.01  0.77  0.00  0.00  0.01
13:36:44 B17F.1   B17F13400714  stand   83   NoReport stand              trk  1.00 OpenFloor  3   66    0.00  0.01  0.93  0.00  0.00  0.01
13:36:44 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   66    0.00  0.01  0.93  0.00  0.00  0.01
13:36:44 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   66    0.00  0.01  0.90  0.00  0.00  0.01
13:36:45 B17F.1   B17F13400714  walk    102  NoReport walk               trk  1.00 OpenFloor  3   67    0.00  0.01  0.93  0.00  0.00  0.01
13:36:45 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   67    0.00  0.01  0.93  0.00  0.00  0.01
13:36:45 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   67    0.00  0.01  0.83  0.00  0.00  0.01
13:36:46 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   68    0.00  0.01  0.93  0.00  0.00  0.01
13:36:46 B17F.1   B17F13400714  walk    101  NoReport walk               trk  1.00 OpenFloor  3   68    0.00  0.01  0.93  0.00  0.00  0.01
13:36:46 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   68    0.00  0.01  0.76  0.00  0.00  0.01
13:36:47 B17F.1   B17F13400714  walk    91   NoReport walk               trk  1.00 OpenFloor  3   69    0.00  0.01  0.93  0.00  0.00  0.01
13:36:47 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   69    0.00  0.01  0.93  0.00  0.00  0.01
13:36:47 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   69    0.00  0.01  0.69  0.00  0.00  0.01
13:36:48 B17F.1   B17F13400714  walk    0    NoReport walk               trk  1.00 OpenFloor  3   70    0.00  0.01  0.93  0.00  0.00  0.01
13:36:48 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   70    0.00  0.01  0.63  0.00  0.00  0.01
13:36:48 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   70    0.00  0.01  0.93  0.00  0.00  0.01
13:36:49 B17F.1   B17F13400714  walk    0    NoReport walk               trk  1.00 OpenFloor  3   71    0.00  0.01  0.93  0.00  0.00  0.01
13:36:49 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   71    0.00  0.01  0.93  0.00  0.00  0.01
13:36:49 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   71    0.00  0.01  0.58  0.00  0.00  0.01
13:36:50 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   72    0.00  0.01  0.93  0.00  0.00  0.01
13:36:50 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   72    0.00  0.01  0.93  0.00  0.00  0.01
13:36:50 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   72    0.00  0.01  0.54  0.00  0.00  0.01
13:36:51 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   73    0.00  0.01  0.93  0.00  0.00  0.01
13:36:51 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   73    0.00  0.01  0.51  0.00  0.00  0.01
13:36:51 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   73    0.00  0.01  0.93  0.00  0.00  0.01
13:36:52 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   74    0.00  0.01  0.93  0.00  0.00  0.01
13:36:52 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   74    0.00  0.01  0.93  0.00  0.00  0.01
13:36:52 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   74    0.00  0.01  0.48  0.00  0.00  0.01
13:36:53 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   75    0.00  0.01  0.93  0.00  0.00  0.01
13:36:53 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   75    0.00  0.01  0.93  0.00  0.00  0.01
13:36:53 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   75    0.00  0.01  0.45  0.00  0.00  0.01
13:36:54 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   76    0.00  0.01  0.93  0.00  0.00  0.01
13:36:54 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   76    0.00  0.01  0.93  0.00  0.00  0.01
13:36:54 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   76    0.00  0.01  0.44  0.00  0.00  0.01
13:36:55 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   77    0.00  0.01  0.93  0.00  0.00  0.01
13:36:55 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   77    0.00  0.01  0.93  0.00  0.00  0.01
13:36:55 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   77    0.00  0.00  0.42  0.00  0.00  0.01
13:36:56 B17F.1   B17F13400714  stand   105  NoReport stand              trk  1.00 OpenFloor  3   78    0.00  0.01  0.93  0.00  0.00  0.01
13:36:56 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   78    0.00  0.01  0.93  0.00  0.00  0.01
13:36:56 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   78    0.00  0.00  0.41  0.00  0.00  0.01
13:36:57 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   79    0.00  0.01  0.93  0.00  0.00  0.01
13:36:57 B17F.1   B17F13400714  walk    112  NoReport walk               trk  1.00 OpenFloor  3   79    0.00  0.01  0.93  0.00  0.00  0.01
13:36:57 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   79    0.00  0.00  0.40  0.00  0.00  0.01
13:36:58 B17F.1   B17F13400714  walk    118  NoReport walk               trk  1.00 OpenFloor  3   80    0.00  0.01  0.93  0.00  0.00  0.01
13:36:58 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   80    0.00  0.01  0.93  0.00  0.00  0.01
13:36:58 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   80    0.00  0.01  0.80  0.01  0.00  0.01
13:36:59 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   81    0.00  0.01  0.73  0.00  0.00  0.01
13:36:59 B17F.1   B17F13400714  walk    117  NoReport walk               trk  1.00 OpenFloor  3   81    0.00  0.01  0.93  0.00  0.00  0.01
13:36:59 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   81    0.00  0.01  0.93  0.00  0.00  0.01
13:37:00 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   82    0.00  0.01  0.93  0.00  0.00  0.01
13:37:00 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   82    0.00  0.01  0.67  0.00  0.00  0.01
13:37:00 B17F.1   B17F13400714  walk    104  NoReport walk               trk  1.00 OpenFloor  3   82    0.00  0.01  0.93  0.00  0.00  0.01
13:37:01 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   83    0.00  0.01  0.93  0.00  0.00  0.01
13:37:01 B17F.1   B17F13400714  walk    95   NoReport walk               trk  1.00 OpenFloor  3   83    0.00  0.01  0.93  0.00  0.00  0.01
13:37:01 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   83    0.00  0.01  0.62  0.00  0.00  0.01
13:37:02 B17F.1   B17F13400714  walk    108  NoReport walk               trk  1.00 OpenFloor  3   84    0.00  0.01  0.93  0.00  0.00  0.01
13:37:02 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   84    0.00  0.01  0.93  0.00  0.00  0.01
13:37:02 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   84    0.00  0.01  0.57  0.00  0.00  0.01
13:37:03 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   85    0.00  0.01  0.49  0.00  0.00  0.01
13:37:03 B17F.1   B17F13400714  walk    134  NoReport walk               trk  1.00 OpenFloor  3   85    0.00  0.01  0.90  0.00  0.01  0.01
13:37:03 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   85    0.00  0.01  0.90  0.00  0.01  0.01
13:37:04 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   86    0.00  0.01  0.75  0.00  0.00  0.01
13:37:04 B17F.1   B17F13400714  walk    145  NoReport walk               trk  1.00 OpenFloor  3   86    0.00  0.01  0.93  0.00  0.00  0.01
13:37:04 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   86    0.00  0.01  0.93  0.00  0.00  0.01
13:37:05 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   87    0.00  0.01  0.93  0.00  0.00  0.01
13:37:05 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   87    0.00  0.01  0.93  0.00  0.00  0.01
13:37:05 B17F.2   B17F23515081  stand   80   NoReport stand              trk  1.00 OpenFloor  3   87    0.00  0.01  0.86  0.00  0.00  0.01
13:37:06 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   88    0.00  0.01  0.93  0.00  0.00  0.01
13:37:06 B17F.1   B17F13400714  stand   134  NoReport stand              trk  1.00 OpenFloor  3   88    0.00  0.01  0.93  0.00  0.00  0.01
13:37:06 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   88    0.00  0.01  0.79  0.00  0.00  0.01
13:37:07 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   89    0.00  0.01  0.93  0.00  0.00  0.01
13:37:07 B17F.2   B17F23515081  stand   77   NoReport stand              trk  1.00 OpenFloor  3   89    0.00  0.01  0.82  0.00  0.00  0.01
13:37:07 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   89    0.00  0.01  0.93  0.00  0.00  0.01
13:37:08 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   90    0.00  0.01  0.93  0.00  0.00  0.01
13:37:08 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   90    0.00  0.01  0.83  0.00  0.00  0.01
13:37:08 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   90    0.00  0.01  0.93  0.00  0.00  0.01
13:37:09 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   91    0.00  0.01  0.93  0.00  0.00  0.01
13:37:09 B17F.1   B17F13400714  stand   116  NoReport stand              trk  1.00 OpenFloor  3   91    0.00  0.01  0.93  0.00  0.00  0.01
13:37:09 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   91    0.00  0.01  0.76  0.00  0.00  0.01
13:37:10 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   92    0.00  0.01  0.93  0.00  0.00  0.01
13:37:10 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   92    0.00  0.01  0.93  0.00  0.00  0.01
13:37:10 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   92    0.00  0.01  0.70  0.00  0.00  0.01
13:37:11 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   93    0.00  0.01  0.93  0.00  0.00  0.01
13:37:11 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   93    0.00  0.01  0.64  0.00  0.00  0.01
13:37:11 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   93    0.00  0.01  0.93  0.00  0.00  0.01
13:37:12 B17F.2   B17F23515081  stand   81   NoReport stand              trk  1.00 OpenFloor  3   94    0.00  0.01  0.59  0.00  0.00  0.01
13:37:12 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   94    0.00  0.01  0.93  0.00  0.00  0.01
13:37:12 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   94    0.00  0.01  0.93  0.00  0.00  0.01
13:37:13 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   95    0.00  0.01  0.93  0.00  0.00  0.01
13:37:13 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   95    0.00  0.01  0.54  0.00  0.00  0.01
13:37:13 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   95    0.00  0.01  0.93  0.00  0.00  0.01
13:37:14 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   96    0.00  0.01  0.93  0.00  0.00  0.01
13:37:14 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   96    0.00  0.01  0.51  0.00  0.00  0.01
13:37:14 B17F.1   B17F13400714  stand   107  NoReport stand              trk  1.00 OpenFloor  3   96    0.00  0.01  0.93  0.00  0.00  0.01
13:37:15 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   97    0.00  0.01  0.93  0.00  0.00  0.01
13:37:15 B17F.1   B17F13400714  stand   142  NoReport stand              trk  1.00 OpenFloor  3   97    0.00  0.01  0.93  0.00  0.00  0.01
13:37:15 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   97    0.00  0.01  0.48  0.00  0.00  0.01
13:37:16 B17F.1   B17F13400714  stand   136  NoReport stand              trk  1.00 OpenFloor  3   98    0.00  0.01  0.93  0.00  0.00  0.01
13:37:16 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   98    0.00  0.01  0.93  0.00  0.00  0.01
13:37:16 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   98    0.00  0.01  0.46  0.00  0.00  0.01
13:37:17 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   99    0.00  0.01  0.93  0.00  0.00  0.01
13:37:17 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   99    0.00  0.01  0.60  0.00  0.00  0.01
13:37:17 B17F.1   B17F13400714  stand   106  NoReport stand              trk  1.00 OpenFloor  3   99    0.00  0.01  0.93  0.00  0.00  0.01
13:37:18 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   100   0.00  0.01  0.93  0.00  0.00  0.01
13:37:18 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   100   0.00  0.01  0.70  0.00  0.00  0.01
13:37:18 B17F.1   B17F13400714  walk    67   NoReport walk               trk  1.00 OpenFloor  3   100   0.00  0.01  0.93  0.00  0.00  0.01
13:37:19 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   101   0.00  0.01  0.93  0.00  0.00  0.01
13:37:19 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   101   0.00  0.01  0.64  0.00  0.00  0.01
13:37:19 B17F.1   B17F13400714  walk    88   NoReport walk               trk  1.00 OpenFloor  3   101   0.00  0.01  0.93  0.00  0.00  0.01
13:37:20 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   102   0.00  0.01  0.93  0.00  0.00  0.01
13:37:20 B17F.2   B17F23515081  stand   71   NoReport stand              trk  1.00 OpenFloor  3   102   0.00  0.01  0.82  0.00  0.00  0.01
13:37:20 B17F.1   B17F13400714  walk    0    NoReport walk               trk  1.00 OpenFloor  3   102   0.00  0.01  0.93  0.00  0.00  0.01
13:37:21 B17F.0   B17F03400714  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:37:21 B17F.2   B17F23515081  stand   82   NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.91  0.00  0.00  0.01
13:37:21 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 OpenFloor  3   0     0.00  0.02  0.85  0.00  0.00  0.02
13:37:22 B17F.E   -             -       0    NoReport np=2               room -    OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:37:22 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.02  0.84  0.00  0.00  0.02
13:37:22 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 OpenFloor  3   0     0.00  0.05  0.46  0.00  0.02  0.05
13:37:23 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.04  0.20  0.01  0.03  0.03
13:37:23 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.77  0.00  0.01  0.02
13:37:24 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.70  0.00  0.01  0.02
13:37:24 B17F.1   B17F13400714  sit     36   NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.02  0.09  0.01  0.02  0.01
13:37:25 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.64  0.00  0.01  0.02
13:37:25 B17F.1   B17F13400714  sit     87   NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.01  0.09  0.01  0.01  0.01
13:37:26 B17F.1   B17F13400714  sit     90   NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.01  0.17  0.01  0.01  0.01
13:37:26 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.59  0.00  0.01  0.02
13:37:27 B17F.1   B17F13400714  sit     67   NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.01  0.07  0.01  0.00  0.01
13:37:27 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.76  0.01  0.01  0.02
13:37:28 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 BlindOpen  2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:28 B17F.2   B17F23515081  walk    79   NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.82  0.01  0.01  0.02
13:37:29 B17F.2   B17F23515081  walk    83   NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.76  0.00  0.01  0.02
13:37:29 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.01  0.13  0.02  0.01  0.02
13:37:30 B17F.2   B17F23515081  walk    0    NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.70  0.00  0.01  0.02
13:37:30 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.45  0.03  0.01  0.02
13:37:30 B17F.2   B17F23515081  walk    84   NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.64  0.00  0.01  0.02
13:37:30 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.69  0.02  0.01  0.02
13:37:31 B17F.2   B17F23515081  stand   82   NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.70  0.01  0.01  0.02
13:37:31 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.79  0.01  0.01  0.02
13:37:32 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.80  0.01  0.01  0.02
13:37:32 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.83  0.01  0.01  0.02
13:37:33 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.74  0.00  0.01  0.02
13:37:33 B17F.1   B17F13400714  stand   85   NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.84  0.00  0.01  0.02
13:37:34 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.68  0.00  0.01  0.02
13:37:34 B17F.1   B17F13400714  walk    89   NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:37:35 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.63  0.00  0.01  0.02
13:37:35 B17F.1   B17F13400714  walk    116  NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:37:36 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.58  0.00  0.01  0.02
13:37:36 B17F.1   B17F13400714  walk    118  NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:37:37 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.01  0.54  0.01  0.01  0.02
13:37:37 B17F.1   B17F13400714  walk    112  NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:37:38 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.01  0.50  0.01  0.01  0.01
13:37:38 B17F.1   B17F13400714  walk    99   NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:37:39 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.01  0.47  0.01  0.01  0.01
13:37:39 B17F.1   B17F13400714  walk    67   NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:37:40 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.70  0.01  0.01  0.02
13:37:40 B17F.1   B17F13400714  walk    0    NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:37:41 B17F.1   B17F13400714  walk    0    NoReport walk               trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:37:41 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.80  0.01  0.01  0.02
13:37:42 B17F.1   B17F13400714  sit     79   NoReport sit                trk  1.00 Empty      2   0     0.00  0.05  0.48  0.01  0.02  0.05
13:37:42 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.83  0.00  0.01  0.02
13:37:43 B17F.1   B17F13400714  sit     91   NoReport sit                trk  1.00 Empty      2   0     0.00  0.02  0.53  0.01  0.02  0.02
13:37:43 B17F.2   B17F23515081  walk    0    NoReport walk               trk  1.00 Empty      2   0     0.00  0.02  0.82  0.00  0.01  0.02
13:37:44 B17F.2   B17F23515081  walk    0    NoReport walk               trk  1.00 Empty      2   0     0.00  0.02  0.76  0.00  0.01  0.02
13:37:44 B17F.1   B17F13400714  sit     82   NoReport sit                trk  1.00 Empty      2   0     0.00  0.02  0.52  0.01  0.01  0.01
13:37:45 B17F.1   B17F13400714  sit     68   NoReport sit                trk  1.00 Empty      2   0     0.00  0.02  0.19  0.01  0.01  0.02
13:37:45 B17F.2   B17F23515081  walk    0    NoReport walk               trk  1.00 Empty      2   0     0.00  0.02  0.70  0.00  0.01  0.02
13:37:46 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.74  0.00  0.01  0.02
13:37:46 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 Empty      2   0     0.00  0.01  0.08  0.01  0.01  0.01
13:37:47 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.81  0.00  0.01  0.02
13:37:47 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 Empty      2   0     0.00  0.01  0.05  0.01  0.01  0.01
13:37:48 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.75  0.00  0.01  0.02
13:37:48 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 Empty      2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:49 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.69  0.00  0.01  0.02
13:37:49 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 Empty      2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:50 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.63  0.00  0.01  0.02
13:37:50 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 Empty      2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:51 B17F.2   B17F23515081  stand   77   NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.70  0.01  0.01  0.02
13:37:51 B17F.1   B17F13400714  sit     0    NoReport sit                trk  1.00 Empty      2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:52 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.64  0.00  0.01  0.02
13:37:52 B17F.1   B17F13400714  sit     48   NoReport sit                trk  1.00 Empty      2   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:53 B17F.2   B17F23515081  stand   74   NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.59  0.00  0.01  0.02
13:37:53 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.01  0.13  0.02  0.01  0.02
13:37:54 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.67  0.01  0.01  0.02
13:37:54 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.45  0.03  0.01  0.02
13:37:55 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.68  0.02  0.01  0.02
13:37:55 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.72  0.01  0.01  0.02
13:37:56 B17F.2   B17F23515081  stand   89   NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.67  0.00  0.01  0.02
13:37:56 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.79  0.01  0.01  0.02
13:37:57 B17F.2   B17F23515081  stand   85   NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.62  0.00  0.01  0.02
13:37:57 B17F.1   B17F13400714  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.83  0.01  0.01  0.02
13:37:58 B17F.1   B17F13400714  stand   65   NoReport stand              trk  1.00 Empty      2   14    0.00  0.02  0.84  0.00  0.01  0.02
13:37:58 B17F.2   B17F23515081  stand   0    NoReport stand              trk  1.00 Empty      2   14    0.00  0.02  0.57  0.00  0.01  0.02
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
13:34:00.714 B17F.1    walk   255  -20    120    80    80        
13:34:00.714 B17F.0    sit    4    -60    0      0     80   126  
13:34:01.694 B17F.1    walk   255  -50    120    68    80   120  
13:34:01.694 B17F.0    sit    4    -60    0      0     80   120  
13:34:02.698 B17F.1    walk   255  -110   130    78    80   139  
13:34:02.698 B17F.0    sit    4    -60    0      0     80   139  
13:34:03.694 B17F.1    walk   255  -140   150    85    80   170  
13:34:03.694 B17F.0    sit    4    -60    0      0     80   170  
13:34:04.751 B17F.0    sit    4    -60    0      0     80   0    
13:34:04.751 B17F.1    walk   255  -160   140    99    80   172  
13:34:05.592 B17F.0    sit    4    -60    0      0     80   172  
13:34:05.592 B17F.1    walk   255  -160   140    88    80   172  
13:34:06.607 B17F.0    sit    4    -60    0      0     80   172  
13:34:06.607 B17F.1    stand  255  -160   160    0     80   188  
13:34:07.591 B17F.0    sit    4    -60    0      0     80   188  
13:34:07.591 B17F.1    stand  255  -160   150    0     80   180  
13:34:08.593 B17F.0    sit    4    -60    0      0     80   180  
13:34:08.593 B17F.1    stand  255  -170   150    0     80   186  
13:34:09.574 B17F.0    sit    4    -60    0      0     80   186  
13:34:09.574 B17F.1    stand  255  -170   120    106   80   162  
13:34:10.558 B17F.0    sit    4    -60    0      0     80   162  
13:34:10.558 B17F.1    stand  255  -160   130    98    80   164  
13:34:11.557 B17F.0    sit    4    -60    0      0     80   164  
13:34:11.557 B17F.1    stand  255  -170   90     96    80   142  
13:34:12.561 B17F.0    sit    4    -60    0      0     80   142  
13:34:12.561 B17F.1    stand  255  -140   80     83    80   113  
13:34:13.563 B17F.0    sit    4    -50    60     84    80   92   
13:34:13.563 B17F.1    stand  255  -140   70     0     80   90   
13:34:14.566 B17F.0    walk   4    -30    90     83    80   111  
13:34:14.566 B17F.1    stand  255  -140   60     0     80   114  
13:34:15.565 B17F.0    walk   4    0      140    80    80   161  
13:34:15.565 B17F.1    stand  255  -100   50     0     80   134  
13:34:16.565 B17F.0    walk   4    40     100    61    80   148  
13:34:16.565 B17F.1    stand  255  -60    40     0     80   116  
13:34:17.564 B17F.1    stand  255  -60    50     0     80   10   
13:34:17.564 B17F.0    walk   4    70     110    77    80   143  
13:34:18.573 B17F.1    stand  255  -80    30     76    80   170  
13:34:18.573 B17F.0    walk   4    40     110    0     80   144  
13:34:19.567 B17F.0    sit    4    50     100    93    80   14   
13:34:19.567 B17F.1    stand  255  -120   40     0     80   180  
13:34:20.569 B17F.0    walk   4    50     100    84    80   180  
13:34:20.569 B17F.1    stand  255  -120   40     0     80   180  
13:34:21.477 B17F.0    walk   4    50     90     91    80   177  
13:34:21.477 B17F.1    stand  255  -120   40     0     80   177  
13:34:22.470 B17F.0    stand  4    40     90     0     80   167  
13:34:22.470 B17F.1    stand  255  -110   40     0     80   158  
13:34:23.462 B17F.0    stand  4    40     90     89    80   158  
13:34:23.462 B17F.1    stand  255  -110   40     0     80   158  
13:34:24.468 B17F.0    stand  4    40     90     91    80   158  
13:34:24.468 B17F.1    stand  255  -110   40     0     80   158  
13:34:25.506 B17F.0    stand  4    40     90     0     80   158  
13:34:25.506 B17F.1    stand  255  -110   40     0     80   158  
13:34:26.507 B17F.0    stand  4    50     90     87    80   167  
13:34:26.507 B17F.1    stand  255  -110   40     0     80   167  
13:34:27.506 B17F.0    stand  4    50     80     78    80   164  
13:34:27.506 B17F.1    stand  255  -110   40     0     80   164  
13:34:28.407 B17F.0    stand  4    50     80     66    80   164  
13:34:28.407 B17F.1    stand  255  -110   40     0     80   164  
13:34:29.408 B17F.0    stand  4    50     100    0     80   170  
13:34:29.408 B17F.1    stand  255  -110   40     0     80   170  
13:34:30.411 B17F.0    stand  4    60     90     73    80   177  
13:34:30.411 B17F.1    stand  255  -110   40     0     80   177  
13:34:31.410 B17F.0    stand  4    60     100    85    80   180  
13:34:31.410 B17F.1    stand  255  -110   40     0     80   180  
13:34:32.409 B17F.0    stand  4    50     80     70    80   164  
13:34:32.409 B17F.1    stand  255  -110   40     0     80   164  
13:34:33.417 B17F.0    stand  4    50     80     77    80   164  
13:34:33.417 B17F.1    stand  255  -110   40     0     80   164  
13:34:34.417 B17F.0    stand  4    50     90     0     80   167  
13:34:34.417 B17F.1    stand  255  -110   40     0     80   167  
13:34:35.412 B17F.0    stand  4    50     90     0     80   167  
13:34:35.412 B17F.1    stand  255  -110   40     0     80   167  
13:34:36.416 B17F.0    stand  4    50     90     0     80   167  
13:34:36.416 B17F.1    stand  255  -110   40     0     80   167  
13:34:37.415 B17F.1    stand  255  -110   40     0     80   0    
13:34:37.415 B17F.0    stand  4    50     90     0     80   167  
13:34:38.417 B17F.1    stand  255  -110   40     0     80   167  
13:34:38.417 B17F.0    stand  4    50     90     0     80   167  
13:34:39.317 B17F.1    stand  255  -110   40     0     80   167  
13:34:39.317 B17F.0    stand  4    50     90     0     80   167  
13:34:40.329 B17F.1    stand  255  -110   40     0     80   167  
13:34:40.329 B17F.0    stand  4    50     90     0     80   167  
13:34:41.325 B17F.1    stand  255  -110   40     0     80   167  
13:34:41.325 B17F.0    stand  4    60     110    0     80   183  
13:34:42.322 B17F.0    stand  4    60     80     76    80   30   
13:34:42.322 B17F.1    stand  255  -110   40     0     80   174  
13:34:43.318 B17F.0    stand  4    50     90     85    80   167  
13:34:43.318 B17F.1    stand  255  -110   40     0     80   167  
13:34:44.323 B17F.0    stand  4    50     90     71    80   167  
13:34:44.323 B17F.1    stand  255  -110   40     0     80   167  
13:34:45.323 B17F.0    stand  4    30     100    77    80   152  
13:34:45.323 B17F.1    stand  255  -110   40     0     80   152  
13:34:46.324 B17F.0    stand  4    10     120    81    80   144  
13:34:46.324 B17F.1    stand  255  -110   40     0     80   144  
13:34:47.322 B17F.0    stand  4    -10    120    83    80   128  
13:34:47.322 B17F.1    stand  255  -120   30     0     80   142  
13:34:48.328 B17F.0    walk   4    -20    90     89    80   116  
13:34:48.328 B17F.1    stand  255  -110   80     0     80   90   
13:34:49.324 B17F.0    walk   4    -40    70     87    80   70   
13:34:49.324 B17F.1    stand  255  -110   90     0     80   72   
13:34:50.331 B17F.0    walk   4    -20    100    78    80   90   
13:34:50.331 B17F.1    stand  255  -110   80     0     80   92   
13:34:51.230 B17F.0    walk   4    20     110    62    80   133  
13:34:51.230 B17F.1    stand  255  -110   80     0     80   133  
13:34:52.228 B17F.0    walk   4    40     90     0     80   150  
13:34:52.228 B17F.1    stand  255  -90    20     0     80   147  
13:34:53.232 B17F.0    walk   4    50     90     0     80   156  
13:34:53.232 B17F.1    stand  255  -70    20     92    80   138  
13:34:54.226 B17F.0    sit    4    50     100    81    80   144  
13:34:54.226 B17F.1    stand  255  -70    20     82    80   144  
13:34:55.231 B17F.0    sit    4    60     100    0     80   152  
13:34:55.231 B17F.1    stand  255  -80    30     71    80   156  
13:34:56.229 B17F.0    sit    4    50     80     79    80   139  
13:34:56.229 B17F.1    stand  255  -80    20     0     80   143  
13:34:57.171 B17F.1    stand  255  -80    10     0     80   10   
13:34:57.171 B17F.0    sit    4    50     80     0     80   147  
13:34:58.172 B17F.1    stand  255  -80    10     0     80   147  
13:34:58.172 B17F.0    sit    4    50     80     77    80   147  
13:34:59.173 B17F.0    sit    4    50     80     0     80   0    
13:34:59.173 B17F.1    stand  255  -80    10     0     80   147  
13:35:00.177 B17F.0    sit    4    50     70     76    80   143  
13:35:00.177 B17F.1    stand  255  -80    10     0     80   143  
13:35:01.178 B17F.0    sit    4    60     60     78    80   148  
13:35:01.178 B17F.1    stand  255  -100   10     0     80   167  
13:35:02.179 B17F.0    sit    4    50     80     74    80   165  
13:35:02.179 B17F.1    stand  255  -110   20     0     80   170  
13:35:03.180 B17F.0    sit    4    50     80     69    80   170  
13:35:03.180 B17F.1    stand  255  -50    50     0     80   104  
13:35:04.240 B17F.1    stand  255  -40    50     0     80   10   
13:35:04.240 B17F.0    sit    4    60     80     79    80   104  
13:35:05.181 B17F.1    stand  255  -100   40     63    80   164  
13:35:05.181 B17F.0    walk   4    150    30     0     80   250  
13:35:06.192 B17F.1    walk   255  -150   50     35    80   300  
13:35:06.192 B17F.0    walk   4    200    40     0     80   350  
13:35:07.184 B17F.1    walk   255  -170   50     0     80   370  
13:35:07.184 B17F.0    walk   4    220    30     0     80   390  
13:35:08.182 B17F.1    stand  255  -170   50     0     80   390  
13:35:08.182 B17F.0    stand  4    210    30     0     80   380  
13:35:09.075 B17F.1    stand  255  -170   50     0     80   380  
13:35:09.075 B17F.0    stand  4    210    30     0     80   380  
13:35:10.079 B17F.1    stand  255  -180   50     0     80   390  
13:35:10.079 B17F.0    stand  4    210    30     0     80   390  
13:35:11.081 B17F.1    stand  255  -180   50     0     80   390  
13:35:11.081 B17F.0    stand  4    210    40     0     80   390  
13:35:12.080 B17F.1    stand  255  -180   50     0     80   390  
13:35:12.080 B17F.0    stand  4    210    40     0     80   390  
13:35:13.089 B17F.1    stand  255  -180   50     0     80   390  
13:35:13.089 B17F.0    stand  4    210    40     0     80   390  
13:35:14.134 B17F.1    stand  255  -180   50     0     80   390  
13:35:14.134 B17F.0    stand  4    210    40     0     80   390  
13:35:15.081 B17F.0    stand  4    210    40     0     80   0    
13:35:15.081 B17F.1    stand  255  -190   40     0     80   400  
13:35:15.081 B17F.2    stand  255  50     90     79    80   245  
13:35:16.049 B17F.1    stand  255  -200   20     0     80   259  
13:35:16.049 B17F.2    stand  255  50     90     0     80   259  
13:35:16.049 B17F.0    stand  4    210    40     0     80   167  
13:35:17.080 B17F.0    stand  4    210    40     0     80   0    
13:35:17.080 B17F.1    stand  255  -200   20     0     80   410  
13:35:17.080 B17F.2    stand  255  50     80     82    80   257  
13:35:18.050 B17F.1    stand  255  -200   30     0     80   254  
13:35:18.050 B17F.0    stand  4    210    40     0     80   410  
13:35:18.050 B17F.2    stand  255  50     80     0     80   164  
13:35:19.053 B17F.0    stand  4    210    40     0     80   164  
13:35:19.053 B17F.1    stand  255  -200   20     0     80   410  
13:35:19.053 B17F.2    stand  255  50     80     0     80   257  
13:35:20.061 B17F.2    stand  255  50     80     0     80   0    
13:35:20.061 B17F.1    stand  255  -200   20     0     80   257  
13:35:20.061 B17F.0    stand  4    210    40     0     80   410  
13:35:21.054 B17F.2    stand  255  50     80     0     80   164  
13:35:21.054 B17F.1    stand  255  -200   20     0     80   257  
13:35:21.054 B17F.0    stand  4    210    40     0     80   410  
13:35:22.064 B17F.2    stand  255  50     80     0     80   164  
13:35:22.064 B17F.1    stand  255  -200   20     0     80   257  
13:35:22.064 B17F.0    stand  4    210    40     0     80   410  
13:35:23.056 B17F.0    stand  4    210    40     0     80   0    
13:35:23.056 B17F.1    stand  255  -200   20     0     80   410  
13:35:23.056 B17F.2    stand  255  50     80     0     80   257  
13:35:23.957 B17F.2    stand  255  50     80     0     80   0    
13:35:23.957 B17F.1    stand  255  -200   20     0     80   257  
13:35:23.957 B17F.0    stand  4    210    40     0     80   410  
13:35:24.958 B17F.2    stand  255  50     80     0     80   164  
13:35:24.958 B17F.1    stand  255  -200   20     0     80   257  
13:35:24.958 B17F.0    stand  4    210    40     0     80   410  
13:35:25.958 B17F.2    stand  255  50     80     76    80   164  
13:35:25.958 B17F.1    stand  255  -200   20     0     80   257  
13:35:25.958 B17F.0    stand  4    210    40     0     80   410  
13:35:26.974 B17F.2    stand  255  50     80     0     80   164  
13:35:26.974 B17F.1    stand  255  -200   40     64    80   253  
13:35:26.974 B17F.0    stand  4    210    40     0     80   410  
13:35:27.959 B17F.2    stand  255  50     80     0     80   164  
13:35:27.959 B17F.1    stand  255  -190   40     0     80   243  
13:35:27.959 B17F.0    stand  4    210    40     0     80   400  
13:35:28.961 B17F.2    stand  255  50     90     0     80   167  
13:35:28.961 B17F.1    stand  255  -200   50     0     80   253  
13:35:28.961 B17F.0    stand  4    210    40     0     80   410  
13:35:29.960 B17F.0    stand  4    210    40     0     80   0    
13:35:29.960 B17F.1    stand  255  -210   50     0     80   420  
13:35:29.960 B17F.2    stand  255  50     90     0     80   263  
13:35:30.969 B17F.2    stand  255  50     90     81    80   0    
13:35:30.969 B17F.1    stand  255  -200   50     0     80   253  
13:35:30.969 B17F.0    stand  4    210    40     0     80   410  
13:35:31.965 B17F.2    stand  255  50     90     76    80   167  
13:35:31.965 B17F.1    stand  255  -200   40     0     80   254  
13:35:31.965 B17F.0    stand  4    210    40     0     80   410  
13:35:32.970 B17F.2    stand  255  50     80     79    80   164  
13:35:32.970 B17F.1    stand  255  -200   40     0     80   253  
13:35:32.970 B17F.0    stand  4    210    40     0     80   410  
13:35:33.966 B17F.2    stand  255  60     100    83    80   161  
13:35:33.966 B17F.1    stand  255  -200   40     0     80   266  
13:35:33.966 B17F.0    stand  4    210    40     0     80   410  
13:35:34.965 B17F.2    stand  255  50     80     0     80   164  
13:35:34.965 B17F.1    stand  255  -200   40     0     80   253  
13:35:34.965 B17F.0    stand  4    210    40     0     80   410  
13:35:35.860 B17F.2    stand  255  60     100    76    80   161  
13:35:35.860 B17F.1    stand  255  -200   40     0     80   266  
13:35:35.860 B17F.0    stand  4    210    40     0     80   410  
13:35:36.861 B17F.2    stand  255  90     130    0     80   150  
13:35:36.861 B17F.1    stand  255  -200   40     0     80   303  
13:35:36.861 B17F.0    stand  4    210    40     0     80   410  
13:35:37.879 B17F.2    stand  255  50     100    86    80   170  
13:35:37.879 B17F.1    stand  255  -200   40     0     80   257  
13:35:37.879 B17F.0    stand  4    210    40     0     80   410  
13:35:38.866 B17F.2    stand  255  50     80     78    80   164  
13:35:38.866 B17F.1    stand  255  -200   40     0     80   253  
13:35:38.866 B17F.0    stand  4    210    40     0     80   410  
13:35:39.862 B17F.2    stand  255  50     80     74    80   164  
13:35:39.862 B17F.1    stand  255  -200   40     0     80   253  
13:35:39.862 B17F.0    stand  4    210    40     0     80   410  
13:35:40.864 B17F.2    stand  255  50     70     0     80   162  
13:35:40.864 B17F.1    stand  255  -200   40     0     80   251  
13:35:40.864 B17F.0    stand  4    210    40     0     80   410  
13:35:41.880 B17F.0    stand  4    210    40     0     80   0    
13:35:41.880 B17F.1    stand  255  -200   40     0     80   410  
13:35:41.880 B17F.2    stand  255  50     90     0     80   254  
13:35:42.873 B17F.1    stand  255  -200   40     0     80   254  
13:35:42.873 B17F.2    stand  255  60     100    84    80   266  
13:35:42.873 B17F.0    stand  4    210    40     0     80   161  
13:35:43.869 B17F.2    stand  255  50     80     0     80   164  
13:35:43.869 B17F.1    stand  255  -200   40     0     80   253  
13:35:43.869 B17F.0    stand  4    210    40     0     80   410  
13:35:44.867 B17F.2    stand  255  70     110    0     80   156  
13:35:44.867 B17F.1    stand  255  -200   40     0     80   278  
13:35:44.867 B17F.0    stand  4    210    40     0     80   410  
13:35:45.779 B17F.0    stand  4    210    40     0     80   0    
13:35:45.779 B17F.1    stand  255  -200   40     0     80   410  
13:35:45.779 B17F.2    stand  255  70     110    0     80   278  
13:35:46.787 B17F.2    stand  255  70     110    0     80   0    
13:35:46.787 B17F.1    stand  255  -200   40     0     80   278  
13:35:46.787 B17F.0    stand  4    210    40     0     80   410  
13:35:47.782 B17F.2    stand  255  70     110    0     80   156  
13:35:47.782 B17F.1    stand  255  -200   40     0     80   278  
13:35:47.782 B17F.0    stand  4    210    40     0     80   410  
13:35:48.784 B17F.2    stand  255  70     110    0     80   156  
13:35:48.784 B17F.1    stand  255  -200   40     0     80   278  
13:35:48.784 B17F.0    stand  4    210    40     0     80   410  
13:35:49.783 B17F.2    stand  255  60     110    0     80   165  
13:35:49.783 B17F.1    stand  255  -200   40     0     80   269  
13:35:49.783 B17F.0    stand  4    210    40     0     80   410  
13:35:50.788 B17F.2    stand  255  50     90     82    80   167  
13:35:50.788 B17F.1    stand  255  -200   40     0     80   254  
13:35:50.788 B17F.0    stand  4    210    40     0     80   410  
13:35:51.786 B17F.2    stand  255  50     70     0     80   162  
13:35:51.786 B17F.1    stand  255  -200   40     0     80   251  
13:35:51.786 B17F.0    stand  4    210    40     0     80   410  
13:35:52.792 B17F.2    stand  255  50     80     0     80   164  
13:35:52.792 B17F.1    stand  255  -190   90     91    80   240  
13:35:52.792 B17F.0    stand  4    210    40     0     80   403  
13:35:53.806 B17F.0    stand  4    210    40     0     80   0    
13:35:53.806 B17F.1    stand  255  -160   120    100   80   378  
13:35:53.806 B17F.2    stand  255  60     80     0     80   223  
13:35:54.792 B17F.2    stand  255  60     90     0     80   10   
13:35:54.792 B17F.1    stand  255  -180   50     68    80   243  
13:35:54.792 B17F.0    stand  4    210    40     0     80   390  
13:35:55.790 B17F.0    stand  4    210    40     0     80   0    
13:35:55.790 B17F.2    stand  255  60     90     0     80   158  
13:35:55.790 B17F.1    stand  255  -180   50     0     80   243  
13:35:56.791 B17F.2    stand  255  50     80     0     80   231  
13:35:56.791 B17F.1    stand  255  -180   50     0     80   231  
13:35:56.791 B17F.0    stand  4    210    40     0     80   390  
13:35:57.689 B17F.0    stand  4    210    40     0     80   0    
13:35:57.689 B17F.1    stand  255  -180   50     0     80   390  
13:35:57.689 B17F.2    stand  255  50     70     0     80   230  
13:35:58.684 B17F.0    stand  4    210    40     0     80   162  
13:35:58.684 B17F.1    stand  255  -180   50     0     80   390  
13:35:58.684 B17F.2    stand  255  50     70     0     80   230  
13:35:59.698 B17F.2    stand  255  50     70     0     80   0    
13:35:59.698 B17F.1    stand  255  -180   50     0     80   230  
13:35:59.698 B17F.0    stand  4    210    40     0     80   390  
13:36:00.697 B17F.0    stand  4    210    40     0     80   0    
13:36:00.697 B17F.1    stand  255  -180   50     0     80   390  
13:36:00.697 B17F.2    stand  255  50     70     0     80   230  
13:36:01.657 B17F.2    stand  255  50     70     0     80   0    
13:36:01.657 B17F.1    stand  255  -180   50     0     80   230  
13:36:01.657 B17F.0    stand  4    210    40     0     80   390  
13:36:02.668 B17F.1    stand  255  -180   50     0     80   390  
13:36:02.668 B17F.2    stand  255  50     70     0     80   230  
13:36:02.668 B17F.0    stand  4    210    40     0     80   162  
13:36:03.713 B17F.0    stand  4    210    40     0     80   0    
13:36:03.713 B17F.1    stand  255  -180   50     0     80   390  
13:36:03.713 B17F.2    stand  255  50     70     0     80   230  
13:36:04.657 B17F.1    stand  255  -180   50     0     80   230  
13:36:04.657 B17F.2    stand  255  50     70     0     80   230  
13:36:04.657 B17F.0    stand  4    210    40     0     80   162  
13:36:05.657 B17F.1    stand  255  -180   50     0     80   390  
13:36:05.657 B17F.0    stand  4    210    40     0     80   390  
13:36:05.657 B17F.2    stand  255  50     80     0     80   164  
13:36:06.658 B17F.1    stand  255  -180   50     0     80   231  
13:36:06.658 B17F.2    stand  255  50     80     0     80   231  
13:36:06.658 B17F.0    stand  4    210    40     0     80   164  
13:36:07.668 B17F.0    stand  4    210    40     0     80   0    
13:36:07.668 B17F.1    stand  255  -180   50     0     80   390  
13:36:07.668 B17F.2    stand  255  50     80     0     80   231  
13:36:08.662 B17F.2    stand  255  50     80     84    80   0    
13:36:08.662 B17F.1    stand  255  -180   50     0     80   231  
13:36:08.662 B17F.0    stand  4    210    40     0     80   390  
13:36:09.662 B17F.2    stand  255  60     90     85    80   158  
13:36:09.662 B17F.1    stand  255  -180   50     0     80   243  
13:36:09.662 B17F.0    stand  4    210    40     0     80   390  
13:36:10.670 B17F.1    stand  255  -180   50     0     80   390  
13:36:10.670 B17F.0    stand  4    210    40     0     80   390  
13:36:10.670 B17F.2    stand  255  70     110    0     80   156  
13:36:11.662 B17F.1    stand  255  -180   50     0     80   257  
13:36:11.662 B17F.0    stand  4    210    40     0     80   390  
13:36:11.662 B17F.2    stand  255  70     110    0     80   156  
13:36:12.664 B17F.0    stand  4    210    40     0     80   156  
13:36:12.664 B17F.2    stand  255  60     100    0     80   161  
13:36:12.664 B17F.1    stand  255  -180   50     0     80   245  
13:36:13.561 B17F.1    stand  255  -180   50     0     80   0    
13:36:13.561 B17F.0    stand  4    210    40     0     80   390  
13:36:13.561 B17F.2    stand  255  60     100    0     80   161  
13:36:14.566 B17F.1    stand  255  -180   50     0     80   245  
13:36:14.566 B17F.2    stand  255  60     90     0     80   243  
13:36:14.566 B17F.0    stand  4    210    40     0     80   158  
13:36:15.563 B17F.1    stand  255  -180   50     0     80   390  
13:36:15.563 B17F.2    stand  255  60     90     0     80   243  
13:36:15.563 B17F.0    stand  4    210    40     0     80   158  
13:36:16.561 B17F.1    stand  255  -180   50     0     80   390  
13:36:16.561 B17F.0    stand  4    210    40     0     80   390  
13:36:16.561 B17F.2    stand  255  60     90     0     80   158  
13:36:17.579 B17F.1    stand  255  -180   50     0     80   243  
13:36:17.579 B17F.0    stand  4    210    40     0     80   390  
13:36:17.579 B17F.2    stand  255  60     90     0     80   158  
13:36:18.576 B17F.0    stand  4    210    40     0     80   158  
13:36:18.576 B17F.2    stand  255  60     90     0     80   158  
13:36:18.576 B17F.1    stand  255  -180   50     0     80   243  
13:36:19.583 B17F.1    stand  255  -180   50     0     80   0    
13:36:19.583 B17F.0    stand  4    210    40     0     80   390  
13:36:19.583 B17F.2    stand  255  50     80     0     80   164  
13:36:20.582 B17F.1    stand  255  -180   50     0     80   231  
13:36:20.582 B17F.2    stand  255  50     80     0     80   231  
13:36:20.582 B17F.0    stand  4    210    40     0     80   164  
13:36:21.585 B17F.0    stand  4    210    40     0     80   0    
13:36:21.585 B17F.2    stand  255  50     80     0     80   164  
13:36:21.585 B17F.1    stand  255  -180   50     0     80   231  
13:36:22.579 B17F.1    stand  255  -180   50     0     80   0    
13:36:22.579 B17F.0    stand  4    210    40     0     80   390  
13:36:22.579 B17F.2    stand  255  50     80     0     80   164  
13:36:23.480 B17F.1    stand  255  -180   50     0     80   231  
13:36:23.480 B17F.0    stand  4    210    40     0     80   390  
13:36:23.480 B17F.2    stand  255  50     70     0     80   162  
13:36:24.479 B17F.0    stand  4    210    40     0     80   162  
13:36:24.479 B17F.2    stand  255  50     80     77    80   164  
13:36:24.479 B17F.1    stand  255  -180   50     0     80   231  
13:36:25.485 B17F.1    stand  255  -180   50     0     80   0    
13:36:25.485 B17F.0    stand  4    210    40     0     80   390  
13:36:25.485 B17F.2    stand  255  60     90     77    80   158  
13:36:26.480 B17F.1    stand  255  -180   50     0     80   243  
13:36:26.480 B17F.0    stand  4    210    40     0     80   390  
13:36:26.480 B17F.2    stand  255  80     120    0     80   152  
13:36:27.482 B17F.1    stand  255  -180   50     0     80   269  
13:36:27.482 B17F.0    stand  4    210    40     0     80   390  
13:36:27.482 B17F.2    stand  255  70     110    0     80   156  
13:36:28.485 B17F.1    stand  255  -180   50     0     80   257  
13:36:28.485 B17F.0    stand  4    210    40     0     80   390  
13:36:28.485 B17F.2    stand  255  70     110    0     80   156  
13:36:29.489 B17F.1    stand  255  -180   50     0     80   257  
13:36:29.489 B17F.0    stand  4    210    40     0     80   390  
13:36:29.489 B17F.2    stand  255  60     110    82    80   165  
13:36:30.492 B17F.1    stand  255  -180   50     0     80   247  
13:36:30.492 B17F.0    stand  4    210    40     0     80   390  
13:36:30.492 B17F.2    stand  255  50     80     89    80   164  
13:36:31.484 B17F.1    stand  255  -180   50     0     80   231  
13:36:31.484 B17F.0    stand  4    210    40     0     80   390  
13:36:31.484 B17F.2    stand  255  50     80     0     80   164  
13:36:32.484 B17F.0    stand  4    210    40     0     80   164  
13:36:32.484 B17F.1    stand  255  -180   50     0     80   390  
13:36:32.484 B17F.2    stand  255  50     80     0     80   231  
13:36:33.485 B17F.1    stand  255  -180   50     0     80   231  
13:36:33.485 B17F.0    stand  4    210    40     0     80   390  
13:36:33.485 B17F.2    stand  255  50     80     0     80   164  
13:36:34.488 B17F.1    stand  255  -180   50     0     80   231  
13:36:34.488 B17F.2    stand  255  50     80     0     80   231  
13:36:34.488 B17F.0    stand  4    210    40     0     80   164  
13:36:35.381 B17F.1    stand  255  -180   50     0     80   390  
13:36:35.381 B17F.0    stand  4    210    40     0     80   390  
13:36:35.381 B17F.2    stand  255  70     100    0     80   152  
13:36:36.381 B17F.1    stand  255  -180   50     0     80   254  
13:36:36.381 B17F.0    stand  4    210    40     0     80   390  
13:36:36.381 B17F.2    stand  255  90     120    0     80   144  
13:36:37.390 B17F.1    stand  255  -180   50     0     80   278  
13:36:37.390 B17F.0    stand  4    210    40     0     80   390  
13:36:37.390 B17F.2    stand  255  90     110    0     80   138  
13:36:38.391 B17F.1    stand  255  -180   50     0     80   276  
13:36:38.391 B17F.0    stand  4    210    40     0     80   390  
13:36:38.391 B17F.2    stand  255  50     80     0     80   164  
13:36:39.392 B17F.1    stand  255  -180   50     0     80   231  
13:36:39.392 B17F.0    stand  4    210    40     0     80   390  
13:36:39.392 B17F.2    stand  255  50     80     80    80   164  
13:36:40.395 B17F.1    stand  255  -180   50     0     80   231  
13:36:40.395 B17F.0    stand  4    210    40     0     80   390  
13:36:40.395 B17F.2    stand  255  60     90     82    80   158  
13:36:41.387 B17F.1    stand  255  -180   50     0     80   243  
13:36:41.387 B17F.0    stand  4    210    40     0     80   390  
13:36:41.387 B17F.2    stand  255  60     80     73    80   155  
13:36:42.402 B17F.1    stand  255  -180   50     0     80   241  
13:36:42.402 B17F.0    stand  4    210    40     0     80   390  
13:36:42.402 B17F.2    stand  255  60     100    69    80   161  
13:36:43.388 B17F.1    stand  255  -180   40     57    80   247  
13:36:43.388 B17F.0    stand  4    210    40     0     80   390  
13:36:43.388 B17F.2    stand  255  50     80     0     80   164  
13:36:44.406 B17F.1    stand  255  -170   90     83    80   220  
13:36:44.406 B17F.0    stand  4    210    40     0     80   383  
13:36:44.406 B17F.2    stand  255  50     80     0     80   164  
13:36:45.392 B17F.1    walk   255  -140   140    102   80   199  
13:36:45.392 B17F.0    stand  4    210    40     0     80   364  
13:36:45.392 B17F.2    stand  255  50     80     0     80   164  
13:36:46.394 B17F.0    stand  4    210    40     0     80   164  
13:36:46.394 B17F.1    walk   255  -130   150    101   80   357  
13:36:46.394 B17F.2    stand  255  60     90     0     80   199  
13:36:47.288 B17F.1    walk   255  -160   70     91    80   220  
13:36:47.288 B17F.0    stand  4    210    40     0     80   371  
13:36:47.288 B17F.2    stand  255  60     100    0     80   161  
13:36:48.292 B17F.1    walk   255  -170   40     0     80   237  
13:36:48.292 B17F.2    stand  255  60     100    0     80   237  
13:36:48.292 B17F.0    stand  4    210    40     0     80   161  
13:36:49.301 B17F.1    walk   255  -170   10     0     80   381  
13:36:49.301 B17F.0    stand  4    210    40     0     80   381  
13:36:49.301 B17F.2    stand  255  60     100    0     80   161  
13:36:50.288 B17F.1    stand  255  -170   10     0     80   246  
13:36:50.288 B17F.0    stand  4    210    40     0     80   381  
13:36:50.288 B17F.2    stand  255  60     100    0     80   161  
13:36:51.289 B17F.0    stand  4    210    40     0     80   161  
13:36:51.289 B17F.2    stand  255  60     100    0     80   161  
13:36:51.289 B17F.1    stand  255  -180   20     0     80   252  
13:36:52.293 B17F.1    stand  255  -180   30     0     80   10   
13:36:52.293 B17F.0    stand  4    210    40     0     80   390  
13:36:52.293 B17F.2    stand  255  60     100    0     80   161  
13:36:53.301 B17F.1    stand  255  -180   40     0     80   247  
13:36:53.301 B17F.0    stand  4    210    40     0     80   390  
13:36:53.301 B17F.2    stand  255  60     100    0     80   161  
13:36:54.295 B17F.1    stand  255  -190   50     0     80   254  
13:36:54.295 B17F.0    stand  4    180    20     0     80   371  
13:36:54.295 B17F.2    stand  255  60     100    0     80   144  
13:36:55.295 B17F.1    stand  255  -210   40     0     80   276  
13:36:55.295 B17F.0    stand  4    160    10     0     80   371  
13:36:55.295 B17F.2    stand  255  60     110    0     80   141  
13:36:56.298 B17F.1    stand  255  -160   80     105   80   222  
13:36:56.298 B17F.0    stand  4    170    10     0     80   337  
13:36:56.298 B17F.2    stand  255  60     110    0     80   148  
13:36:57.300 B17F.0    stand  4    170    10     0     80   148  
13:36:57.300 B17F.1    walk   255  -120   140    112   80   317  
13:36:57.300 B17F.2    stand  255  60     80     0     80   189  
13:36:58.304 B17F.1    walk   255  -120   140    118   80   189  
13:36:58.304 B17F.0    stand  4    170    10     0     80   317  
13:36:58.304 B17F.2    stand  255  60     80     0     80   130  
13:36:59.203 B17F.2    stand  255  60     90     0     80   10   
13:36:59.203 B17F.1    walk   255  -120   140    117   80   186  
13:36:59.203 B17F.0    stand  4    170    10     0     80   317  
13:37:00.203 B17F.0    stand  4    170    10     0     80   0    
13:37:00.203 B17F.2    stand  255  60     90     0     80   136  
13:37:00.203 B17F.1    walk   255  -140   130    104   80   203  
13:37:01.199 B17F.0    stand  4    170    10     0     80   332  
13:37:01.199 B17F.1    walk   255  -160   130    95    80   351  
13:37:01.199 B17F.2    stand  255  60     90     0     80   223  
13:37:02.193 B17F.1    walk   255  -170   190    108   80   250  
13:37:02.193 B17F.0    stand  4    170    10     0     80   384  
13:37:02.193 B17F.2    stand  255  60     100    0     80   142  
13:37:03.253 B17F.2    stand  255  50     90     0     80   14   
13:37:03.253 B17F.1    walk   255  -200   240    134   80   291  
13:37:03.253 B17F.0    stand  4    170    10     0     80   435  
13:37:04.201 B17F.2    stand  255  50     80     0     80   138  
13:37:04.201 B17F.1    walk   255  -190   250    145   80   294  
13:37:04.201 B17F.0    stand  4    170    10     0     80   432  
13:37:05.197 B17F.0    stand  4    170    10     0     80   0    
13:37:05.197 B17F.1    stand  255  -190   200    0     80   407  
13:37:05.197 B17F.2    stand  255  60     90     80    80   273  
13:37:06.208 B17F.0    stand  4    170    10     0     80   136  
13:37:06.208 B17F.1    stand  255  -190   210    134   80   411  
13:37:06.208 B17F.2    stand  255  50     80     0     80   272  
13:37:07.212 B17F.0    stand  4    170    10     0     80   138  
13:37:07.212 B17F.2    stand  255  50     80     77    80   138  
13:37:07.212 B17F.1    stand  255  -180   220    0     80   269  
13:37:08.211 B17F.0    stand  4    170    10     0     80   408  
13:37:08.211 B17F.2    stand  255  50     80     0     80   138  
13:37:08.211 B17F.1    stand  255  -180   220    0     80   269  
13:37:09.110 B17F.0    stand  4    170    10     0     80   408  
13:37:09.110 B17F.1    stand  255  -180   220    116   80   408  
13:37:09.110 B17F.2    stand  255  50     80     0     80   269  
13:37:10.110 B17F.1    stand  255  -190   230    0     80   283  
13:37:10.110 B17F.0    stand  4    170    10     0     80   421  
13:37:10.110 B17F.2    stand  255  50     80     0     80   138  
13:37:11.123 B17F.1    stand  255  -190   230    0     80   283  
13:37:11.123 B17F.2    stand  255  50     80     0     80   283  
13:37:11.123 B17F.0    stand  4    170    10     0     80   138  
13:37:12.116 B17F.2    stand  255  50     90     81    80   144  
13:37:12.116 B17F.0    stand  4    170    10     0     80   144  
13:37:12.116 B17F.1    stand  255  -190   230    0     80   421  
13:37:13.119 B17F.0    stand  4    170    10     0     80   421  
13:37:13.119 B17F.2    stand  255  70     120    0     80   148  
13:37:13.119 B17F.1    stand  255  -190   230    0     80   282  
13:37:14.114 B17F.0    stand  4    170    10     0     80   421  
13:37:14.114 B17F.2    stand  255  70     120    0     80   148  
13:37:14.114 B17F.1    stand  255  -180   220    107   80   269  
13:37:15.119 B17F.0    stand  4    170    10     0     80   408  
13:37:15.119 B17F.1    stand  255  -190   200    142   80   407  
13:37:15.119 B17F.2    stand  255  60     100    0     80   269  
13:37:16.130 B17F.1    stand  255  -190   230    136   80   281  
13:37:16.130 B17F.0    stand  4    170    10     0     80   421  
13:37:16.130 B17F.2    stand  255  60     90     0     80   136  
13:37:17.120 B17F.0    stand  4    170    10     0     80   136  
13:37:17.120 B17F.2    stand  255  50     80     0     80   138  
13:37:17.120 B17F.1    stand  255  -160   140    106   80   218  
13:37:18.130 B17F.0    stand  4    170    10     0     80   354  
13:37:18.130 B17F.2    stand  255  50     80     0     80   138  
13:37:18.130 B17F.1    walk   255  -150   70     67    80   200  
13:37:19.121 B17F.0    stand  4    160    10     0     80   315  
13:37:19.121 B17F.2    stand  255  50     80     0     80   130  
13:37:19.121 B17F.1    walk   255  -130   20     88    80   189  
13:37:20.139 B17F.0    stand  4    150    10     0     80   280  
13:37:20.139 B17F.2    stand  255  70     70     71    80   100  
13:37:20.139 B17F.1    walk   255  -170   20     0     80   245  
13:37:21.041 B17F.0    stand  4    150    10     0     80   320  
13:37:21.041 B17F.2    stand  255  70     60     82    80   94   
13:37:21.041 B17F.1    sit    255  -130   10     0     80   206  
13:37:22.078 B17F.2    stand  255  50     80     0     80   193  
13:37:22.078 B17F.1    sit    255  -130   20     0     80   189  
13:37:23.035 B17F.1    sit    255  -130   20     0     80   0    
13:37:23.035 B17F.2    stand  255  60     80     0     80   199  
13:37:24.033 B17F.2    stand  255  60     110    0     80   30   
13:37:24.033 B17F.1    sit    255  -130   20     36    80   210  
13:37:25.029 B17F.2    stand  255  60     120    0     80   214  
13:37:25.029 B17F.1    sit    255  -90    30     87    80   174  
13:37:26.032 B17F.1    sit    255  -80    20     90    80   14   
13:37:26.032 B17F.2    stand  255  50     80     0     80   143  
13:37:27.050 B17F.1    sit    255  -110   30     67    80   167  
13:37:27.050 B17F.2    stand  255  60     70     0     80   174  
13:37:28.033 B17F.1    sit    255  -130   20     0     80   196  
13:37:28.033 B17F.2    walk   255  80     50     79    80   212  
13:37:29.094 B17F.2    walk   255  60     100    83    80   53   
13:37:29.094 B17F.1    stand  255  -170   10     0     80   246  
13:37:30.036 B17F.2    walk   255  50     90     0     80   234  
13:37:30.036 B17F.1    stand  255  -170   10     0     80   234  
13:37:30.933 B17F.2    walk   255  60     110    84    80   250  
13:37:30.933 B17F.1    stand  255  -170   10     0     80   250  
13:37:31.934 B17F.2    stand  255  50     80     82    80   230  
13:37:31.934 B17F.1    stand  255  -170   10     0     80   230  
13:37:32.935 B17F.2    stand  255  50     80     0     80   230  
13:37:32.935 B17F.1    stand  255  -160   30     0     80   215  
13:37:33.938 B17F.2    stand  255  50     80     0     80   215  
13:37:33.938 B17F.1    stand  255  -180   40     85    80   233  
13:37:34.937 B17F.2    stand  255  50     80     0     80   233  
13:37:34.937 B17F.1    walk   255  -140   100    89    80   191  
13:37:35.938 B17F.2    stand  255  50     80     0     80   191  
13:37:35.938 B17F.1    walk   255  -130   140    116   80   189  
13:37:36.938 B17F.2    stand  255  50     80     0     80   189  
13:37:36.938 B17F.1    walk   255  -120   140    118   80   180  
13:37:37.882 B17F.2    stand  255  50     80     0     80   180  
13:37:37.882 B17F.1    walk   255  -140   140    112   80   199  
13:37:38.878 B17F.2    stand  255  50     80     0     80   199  
13:37:38.878 B17F.1    walk   255  -130   120    99    80   184  
13:37:39.883 B17F.2    stand  255  50     80     0     80   184  
13:37:39.883 B17F.1    walk   255  -140   60     67    80   191  
13:37:40.880 B17F.2    stand  255  100    0      0     80   247  
13:37:40.880 B17F.1    walk   255  -110   10     0     80   210  
13:37:41.889 B17F.1    walk   255  -100   0      0     80   14   
13:37:41.889 B17F.2    stand  255  110    0      0     80   210  
13:37:42.882 B17F.1    sit    255  -110   20     79    80   220  
13:37:42.882 B17F.2    stand  255  110    0      0     80   220  
13:37:43.889 B17F.1    sit    255  -90    50     91    80   206  
13:37:43.889 B17F.2    walk   255  60     90     0     80   155  
13:37:44.885 B17F.2    walk   255  50     90     0     80   10   
13:37:44.885 B17F.1    sit    255  -70    40     82    80   130  
13:37:45.886 B17F.1    sit    255  -100   30     68    80   31   
13:37:45.886 B17F.2    walk   255  50     90     0     80   161  
13:37:46.886 B17F.2    stand  255  50     80     0     80   10   
13:37:46.886 B17F.1    sit    255  -120   20     0     80   180  
13:37:47.890 B17F.2    stand  255  50     80     0     80   180  
13:37:47.890 B17F.1    sit    255  -120   20     0     80   180  
13:37:48.893 B17F.2    stand  255  60     100    0     80   196  
13:37:48.893 B17F.1    sit    255  -130   20     0     80   206  
13:37:49.782 B17F.2    stand  255  50     90     0     80   193  
13:37:49.782 B17F.1    sit    255  -130   20     0     80   193  
13:37:50.785 B17F.2    stand  255  50     90     0     80   193  
13:37:50.785 B17F.1    sit    255  -130   20     0     80   193  
13:37:51.793 B17F.2    stand  255  50     80     77    80   189  
13:37:51.793 B17F.1    sit    255  -130   20     0     80   189  
13:37:52.786 B17F.2    stand  255  60     90     0     80   202  
13:37:52.786 B17F.1    sit    255  -140   20     48    80   211  
13:37:53.857 B17F.2    stand  255  60     90     74    80   211  
13:37:53.857 B17F.1    stand  255  -160   40     0     80   225  
13:37:54.736 B17F.2    stand  255  50     80     0     80   213  
13:37:54.736 B17F.1    stand  255  -170   40     0     80   223  
13:37:55.741 B17F.1    stand  255  -170   50     0     80   10   
13:37:55.741 B17F.2    stand  255  50     80     0     80   222  
13:37:56.741 B17F.2    stand  255  60     100    89    80   22   
13:37:56.741 B17F.1    stand  255  -170   50     0     80   235  
13:37:57.748 B17F.2    stand  255  60     90     85    80   233  
13:37:57.748 B17F.1    stand  255  -170   40     0     80   235  
13:37:58.742 B17F.1    stand  255  -160   70     65    80   31   
13:37:58.742 B17F.2    stand  255  60     100    0     80   222  

```

**汇总**: xray tick 614 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

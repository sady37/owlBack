# case-B17F-0629-13421358 — 每 tick belief 时间线 (room fd00:0:3:111:2:300, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
13:42:00 B17F.2   B17F24200863  stand   0    -        stand              trk  0.50 Empty      2   0     0.00  0.02  0.26  0.00  0.69  0.03
13:42:00 B17F.1   B17F14200863  stand   0    -        stand              trk  0.50 Empty      2   0     0.00  0.03  0.15  0.00  0.79  0.04
13:42:01 B17F.1   B17F14200863  stand   0    -        stand              trk  0.51 OpenFloor  2   0     0.00  0.03  0.40  0.00  0.53  0.01
13:42:01 B17F.2   B17F24200863  stand   0    -        stand              trk  0.51 OpenFloor  2   0     0.00  0.02  0.52  0.00  0.40  0.01
13:42:02 B17F.1   B17F14200863  stand   0    -        stand              trk  0.52 OpenFloor  2   1     0.00  0.03  0.63  0.00  0.26  0.01
13:42:02 B17F.2   B17F24200863  stand   0    -        stand              trk  0.52 OpenFloor  2   1     0.00  0.02  0.70  0.00  0.18  0.02
13:42:03 B17F.2   B17F24200863  stand   0    -        stand              trk  0.53 OpenFloor  2   2     0.00  0.02  0.79  0.00  0.07  0.02
13:42:03 B17F.1   B17F14200863  stand   0    -        stand              trk  0.53 OpenFloor  2   2     0.00  0.02  0.76  0.00  0.11  0.02
13:42:04 B17F.2   B17F24200863  stand   0    -        stand              trk  0.54 OpenFloor  2   3     0.00  0.02  0.83  0.00  0.03  0.02
13:42:04 B17F.1   B17F14200863  stand   0    -        stand              trk  0.54 OpenFloor  2   3     0.00  0.02  0.82  0.00  0.04  0.02
13:42:05 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   4     0.00  0.02  0.84  0.00  0.02  0.02
13:42:05 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   4     0.00  0.02  0.84  0.00  0.02  0.02
13:42:06 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   5     0.00  0.02  0.85  0.00  0.01  0.02
13:42:06 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   5     0.00  0.02  0.84  0.00  0.01  0.02
13:42:07 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   6     0.00  0.02  0.85  0.00  0.01  0.02
13:42:07 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   6     0.00  0.02  0.85  0.00  0.01  0.02
13:42:08 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
13:42:08 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
13:42:09 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   8     0.00  0.02  0.85  0.00  0.01  0.02
13:42:09 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   8     0.00  0.02  0.85  0.00  0.01  0.02
13:42:10 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   9     0.00  0.02  0.85  0.00  0.01  0.02
13:42:10 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   9     0.00  0.02  0.85  0.00  0.01  0.02
13:42:11 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   10    0.00  0.02  0.85  0.00  0.01  0.02
13:42:11 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   10    0.00  0.02  0.85  0.00  0.01  0.02
13:42:12 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   11    0.00  0.02  0.85  0.00  0.01  0.02
13:42:12 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   11    0.00  0.02  0.85  0.00  0.01  0.02
13:42:13 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   12    0.00  0.02  0.85  0.00  0.01  0.02
13:42:13 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   12    0.00  0.02  0.85  0.00  0.01  0.02
13:42:14 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
13:42:14 B17F.1   B17F14200863  stand   86   -        stand              trk  0.55 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
13:42:15 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:15 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:16 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:16 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:17 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:17 B17F.1   B17F14200863  stand   98   -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:18 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:18 B17F.1   B17F14200863  stand   78   -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:19 B17F.1   B17F14200863  walk    82   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:19 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:20 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:20 B17F.1   B17F14200863  walk    77   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:21 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:21 B17F.1   B17F14200863  walk    74   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:22 B17F.1   B17F14200863  walk    71   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.83  0.00  0.01  0.02
13:42:22 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:23 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:23 B17F.1   B17F14200863  walk    92   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.76  0.00  0.01  0.02
13:42:24 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:24 B17F.1   B17F14200863  walk    80   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.78  0.00  0.01  0.02
13:42:25 B17F.1   B17F14200863  walk    85   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.83  0.00  0.01  0.02
13:42:25 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:26 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:26 B17F.1   B17F14200863  walk    77   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
13:42:27 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:27 B17F.1   B17F14200863  walk    93   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:28 B17F.1   B17F14200863  walk    0    -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:28 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:29 B17F.1   B17F14200863  walk    0    -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:29 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:30 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:30 B17F.1   B17F14200863  walk    0    -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:31 B17F.1   B17F14200863  walk    89   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:31 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:32 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:32 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:33 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:33 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:34 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:34 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:35 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:35 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:36 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:36 B17F.1   B17F14200863  stand   76   -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:37 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:37 B17F.1   B17F14200863  stand   60   -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:38 B17F.1   B17F14200863  stand   99   -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:38 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:39 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:39 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:40 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:40 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:41 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:41 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:42 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:42 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:43 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:43 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:44 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:44 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:45 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:45 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:46 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:46 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:47 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:47 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:48 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:48 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:49 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:49 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:50 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:50 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:42:51 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
13:42:51 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
13:42:52 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
13:42:52 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
13:42:53 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
13:42:53 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
13:42:54 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
13:42:54 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
13:42:55 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
13:42:55 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
13:42:56 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
13:42:56 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
13:42:57 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:42:57 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:42:58 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
13:42:58 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
13:42:59 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
13:42:59 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
13:43:00 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   23    0.00  0.03  0.81  0.00  0.02  0.02
13:43:00 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   23    0.00  0.03  0.81  0.00  0.02  0.02
13:43:01 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   24    0.00  0.02  0.84  0.00  0.01  0.02
13:43:01 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   24    0.00  0.02  0.84  0.00  0.01  0.02
13:43:02 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   25    0.00  0.02  0.84  0.00  0.01  0.02
13:43:02 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   25    0.00  0.02  0.84  0.00  0.01  0.02
13:43:03 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:43:03 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:43:04 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
13:43:04 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
13:43:05 B17F.1   B17F14200863  stand   57   -        stand              trk  0.55 OpenFloor  2   28    0.00  0.02  0.85  0.00  0.01  0.02
13:43:05 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   28    0.00  0.02  0.85  0.00  0.01  0.02
13:43:06 B17F.2   B17F24200863  stand   83   -        stand              trk  0.55 OpenFloor  2   29    0.00  0.02  0.85  0.00  0.01  0.02
13:43:06 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   29    0.00  0.02  0.85  0.00  0.01  0.02
13:43:07 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   30    0.00  0.02  0.85  0.00  0.01  0.02
13:43:07 B17F.2   B17F24200863  stand   94   -        stand              trk  0.55 OpenFloor  2   30    0.00  0.02  0.85  0.00  0.01  0.02
13:43:08 B17F.2   B17F24200863  stand   97   -        stand              trk  0.55 OpenFloor  2   31    0.00  0.02  0.85  0.00  0.01  0.02
13:43:08 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   31    0.00  0.02  0.85  0.00  0.01  0.02
13:43:09 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:09 B17F.2   B17F24200863  walk    75   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:10 B17F.2   B17F24200863  walk    67   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.83  0.00  0.01  0.02
13:43:10 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:11 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:11 B17F.2   B17F24200863  walk    67   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.76  0.00  0.01  0.02
13:43:12 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:12 B17F.2   B17F24200863  walk    71   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.78  0.00  0.01  0.02
13:43:13 B17F.2   B17F24200863  walk    87   -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.72  0.00  0.01  0.02
13:43:13 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:14 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:14 B17F.2   B17F24200863  sit     78   -        sit                trk  0.55 OpenFloor  2   0     0.00  0.02  0.36  0.00  0.01  0.02
13:43:15 B17F.2   B17F24200863  sit     0    -        sit                trk  0.55 OpenFloor  2   0     0.00  0.02  0.13  0.01  0.01  0.02
13:43:15 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:16 B17F.1   B17F14200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:16 B17F.2   B17F24200863  sit     84   -        sit                trk  0.55 OpenFloor  2   0     0.00  0.01  0.06  0.01  0.00  0.01
13:43:17 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   14    0.00  0.01  0.38  0.02  0.01  0.02
13:43:17 B17F.1   B17F14200863  stand   60   -        stand              trk  0.55 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
13:43:18 B17F.1   B17F14200863  stand   69   -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:18 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.65  0.02  0.01  0.02
13:43:19 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.78  0.01  0.01  0.02
13:43:19 B17F.1   B17F14200863  stand   89   -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:20 B17F.1   B17F14200863  walk    127  -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:20 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.82  0.01  0.01  0.02
13:43:21 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
13:43:21 B17F.1   B17F14200863  walk    104  -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:22 B17F.1   B17F14200863  walk    127  -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:22 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:23 B17F.1   B17F14200863  walk    125  -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:23 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:24 B17F.1   B17F14200863  walk    113  -        walk               trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:24 B17F.2   B17F24200863  stand   0    -        stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:25 B17F.E   -             -       0    -        np=3               room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:43:25 B17F.2   B17F24200863  stand   0    -        stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.92  0.00  0.01  0.01
13:43:25 B17F.0   B17F04325242  stand   89   -        stand              trk  0.50 OpenFloor  3   0     0.00  0.02  0.26  0.00  0.69  0.03
13:43:25 B17F.1   B17F14200863  walk    0    -        walk               trk  1.00 OpenFloor  3   0     0.00  0.01  0.92  0.00  0.01  0.01
13:43:26 B17F.1   B17F14200863  walk    0    -        walk               trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:43:26 B17F.2   B17F24200863  stand   0    -        stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:43:26 B17F.0   B17F04325242  stand   82   -        stand              trk  0.51 OpenFloor  3   0     0.00  0.02  0.64  0.00  0.25  0.01
13:43:27 B17F.2   B17F24200863  stand   0    -        stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:43:27 B17F.1   B17F14200863  walk    111  -        walk               trk  1.00 OpenFloor  3   0     0.00  0.01  0.97  0.00  0.00  0.01
13:43:27 B17F.0   B17F04325242  stand   0    -        stand              trk  0.52 OpenFloor  3   0     0.00  0.01  0.75  0.00  0.05  0.01
13:43:28 B17F.1   B17F14200863  walk    0    -        walk               trk  1.00 OpenFloor  3   0     0.00  0.01  0.94  0.00  0.00  0.01
13:43:28 B17F.0   B17F04325242  stand   0    -        stand              trk  0.53 OpenFloor  3   0     0.00  0.01  0.71  0.00  0.01  0.01
13:43:28 B17F.2   B17F24200863  stand   0    -        stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:43:29 B17F.2   B17F24200863  stand   0    -        stand              trk  1.00 OpenFloor  3   11    0.00  0.01  0.93  0.00  0.00  0.01
13:43:29 B17F.0   B17F04325242  stand   0    -        stand              trk  0.54 OpenFloor  3   11    0.00  0.01  0.65  0.00  0.00  0.01
13:43:29 B17F.1   B17F14200863  walk    115  -        walk               trk  1.00 OpenFloor  3   11    0.00  0.01  0.93  0.00  0.00  0.01
13:43:30 B17F.2   B17F24200863  stand   0    -        stand              trk  0.22 OpenFloor  2   4     0.00  0.01  0.93  0.00  0.00  0.01
13:43:30 B17F.0   B17F04325242  stand   0    -        stand              trk  0.55 OpenFloor  2   4     0.00  0.01  0.60  0.00  0.00  0.01
13:43:30 B17F.1   B17F14200863  walk    102  -        walk               trk  1.00 OpenFloor  2   4     0.00  0.01  0.93  0.00  0.00  0.01
13:43:31 B17F.2   B17F24200863  stand   0    -        stand              trk  0.06 OpenFloor  2   5     0.00  0.01  0.93  0.00  0.00  0.01
13:43:31 B17F.0   B17F04325242  stand   89   -        stand              trk  1.00 OpenFloor  2   5     0.00  0.01  0.70  0.00  0.00  0.01
13:43:31 B17F.1   B17F14200863  walk    0    -        walk               trk  1.00 OpenFloor  2   5     0.00  0.01  0.93  0.00  0.00  0.01
13:43:32 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  2   6     0.00  0.01  0.93  0.00  0.00  0.01
13:43:32 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   6     0.00  0.01  0.64  0.00  0.00  0.01
13:43:32 B17F.1   B17F14200863  walk    0    -        walk               trk  1.00 OpenFloor  2   6     0.00  0.01  0.93  0.00  0.00  0.01
13:43:33 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  2   7     0.00  0.01  0.93  0.00  0.00  0.01
13:43:33 B17F.0   B17F04325242  stand   94   -        stand              trk  1.00 OpenFloor  2   7     0.00  0.01  0.59  0.00  0.00  0.01
13:43:33 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  2   7     0.00  0.01  0.93  0.00  0.00  0.01
13:43:34 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  2   8     0.00  0.01  0.93  0.00  0.00  0.01
13:43:34 B17F.0   B17F04325242  stand   98   -        stand              trk  1.00 OpenFloor  2   8     0.00  0.01  0.55  0.00  0.00  0.01
13:43:34 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  2   8     0.00  0.01  0.93  0.00  0.00  0.01
13:43:35 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   9     0.00  0.01  0.93  0.00  0.00  0.01
13:43:35 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   9     0.00  0.01  0.51  0.00  0.00  0.01
13:43:35 B17F.1   B17F14200863  stand   88   -        stand              trk  1.00 OpenFloor  2   9     0.00  0.01  0.93  0.00  0.00  0.01
13:43:36 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  2   10    0.00  0.01  0.93  0.00  0.00  0.01
13:43:36 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   10    0.00  0.01  0.48  0.00  0.00  0.01
13:43:36 B17F.1   B17F14200863  stand   99   -        stand              trk  1.00 OpenFloor  2   10    0.00  0.01  0.93  0.00  0.00  0.01
13:43:36 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   11    0.00  0.01  0.46  0.00  0.00  0.01
13:43:36 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   11    0.00  0.01  0.93  0.00  0.00  0.01
13:43:36 B17F.1   B17F14200863  stand   86   -        stand              trk  1.00 OpenFloor  2   11    0.00  0.01  0.93  0.00  0.00  0.01
13:43:37 B17F.1   B17F14200863  stand   80   -        stand              trk  1.00 OpenFloor  2   12    0.00  0.01  0.93  0.00  0.00  0.01
13:43:37 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   12    0.00  0.01  0.93  0.00  0.00  0.01
13:43:37 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   12    0.00  0.01  0.44  0.00  0.00  0.01
13:43:38 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  2   13    0.00  0.01  0.93  0.00  0.00  0.01
13:43:38 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   13    0.00  0.01  0.93  0.00  0.00  0.01
13:43:38 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   13    0.00  0.00  0.42  0.00  0.00  0.01
13:43:39 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   14    0.00  0.01  0.93  0.00  0.00  0.01
13:43:39 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  2   14    0.00  0.01  0.93  0.00  0.00  0.01
13:43:39 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   14    0.00  0.00  0.41  0.00  0.00  0.01
13:43:40 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  2   15    0.00  0.01  0.93  0.00  0.00  0.01
13:43:40 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   15    0.00  0.00  0.40  0.00  0.00  0.01
13:43:40 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   15    0.00  0.01  0.93  0.00  0.00  0.01
13:43:41 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  2   15    0.00  0.01  0.93  0.00  0.00  0.01
13:43:42 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   16    0.00  0.01  0.93  0.00  0.00  0.01
13:43:42 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   16    0.00  0.00  0.40  0.00  0.00  0.01
13:43:42 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  2   16    0.00  0.01  0.93  0.00  0.00  0.01
13:43:42 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   17    0.00  0.01  0.93  0.00  0.00  0.01
13:43:42 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   17    0.00  0.00  0.39  0.00  0.00  0.01
13:43:42 B17F.1   B17F14200863  stand   104  -        stand              trk  1.00 OpenFloor  2   17    0.00  0.01  0.93  0.00  0.00  0.01
13:43:43 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   18    0.00  0.01  0.93  0.00  0.00  0.01
13:43:43 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   18    0.00  0.00  0.39  0.00  0.00  0.01
13:43:43 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  2   18    0.00  0.01  0.93  0.00  0.00  0.01
13:43:44 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   19    0.00  0.01  0.93  0.00  0.00  0.01
13:43:44 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   19    0.00  0.00  0.38  0.00  0.00  0.01
13:43:44 B17F.1   B17F14200863  stand   88   -        stand              trk  1.00 OpenFloor  2   19    0.00  0.01  0.93  0.00  0.00  0.01
13:43:45 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   20    0.00  0.01  0.93  0.00  0.00  0.01
13:43:45 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   20    0.00  0.00  0.38  0.00  0.00  0.01
13:43:45 B17F.1   B17F14200863  stand   93   -        stand              trk  1.00 OpenFloor  2   20    0.00  0.01  0.93  0.00  0.00  0.01
13:43:46 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   21    0.00  0.01  0.86  0.00  0.00  0.01
13:43:46 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   21    0.00  0.00  0.38  0.00  0.00  0.01
13:43:46 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  2   21    0.00  0.01  0.93  0.00  0.00  0.01
13:43:47 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   22    0.00  0.01  0.92  0.00  0.00  0.01
13:43:47 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   22    0.00  0.00  0.38  0.00  0.00  0.01
13:43:47 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  2   22    0.00  0.01  0.93  0.00  0.00  0.01
13:43:48 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   23    0.00  0.01  0.93  0.00  0.00  0.01
13:43:48 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   23    0.00  0.00  0.38  0.00  0.00  0.01
13:43:48 B17F.1   B17F14200863  stand   68   -        stand              trk  1.00 OpenFloor  2   23    0.00  0.01  0.93  0.00  0.00  0.01
13:43:49 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   24    0.00  0.01  0.93  0.00  0.00  0.01
13:43:49 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   24    0.00  0.00  0.38  0.00  0.00  0.01
13:43:49 B17F.1   B17F14200863  walk    115  -        walk               trk  1.00 OpenFloor  2   24    0.00  0.01  0.93  0.00  0.00  0.01
13:43:50 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   25    0.00  0.00  0.37  0.00  0.00  0.01
13:43:50 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   25    0.00  0.01  0.93  0.00  0.00  0.01
13:43:50 B17F.1   B17F14200863  walk    0    -        walk               trk  1.00 OpenFloor  2   25    0.00  0.01  0.93  0.00  0.00  0.01
13:43:51 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  2   26    0.00  0.01  0.93  0.00  0.00  0.01
13:43:51 B17F.1   B17F14200863  walk    119  -        walk               trk  1.00 OpenFloor  2   26    0.00  0.00  0.98  0.00  0.00  0.00
13:43:51 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   26    0.00  0.00  0.37  0.00  0.00  0.01
13:43:52 B17F.1   B17F14200863  walk    115  -        walk               trk  1.00 OpenFloor  2   27    0.00  0.00  0.97  0.00  0.00  0.01
13:43:52 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   27    0.00  0.01  0.53  0.00  0.00  0.01
13:43:52 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   27    0.00  0.01  0.93  0.00  0.00  0.01
13:43:53 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   28    0.00  0.01  0.93  0.00  0.00  0.01
13:43:53 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   28    0.00  0.01  0.65  0.00  0.00  0.01
13:43:53 B17F.1   B17F14200863  walk    91   -        walk               trk  1.00 OpenFloor  2   28    0.00  0.01  0.97  0.00  0.00  0.01
13:43:54 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:43:54 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.60  0.00  0.00  0.01
13:43:54 B17F.1   B17F14200863  walk    78   -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
13:43:55 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:43:55 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.56  0.00  0.00  0.01
13:43:55 B17F.1   B17F14200863  walk    76   -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.86  0.00  0.00  0.01
13:43:56 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:43:56 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.52  0.00  0.00  0.01
13:43:56 B17F.1   B17F14200863  walk    66   -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.79  0.00  0.00  0.01
13:43:57 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:43:57 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.49  0.00  0.00  0.01
13:43:57 B17F.1   B17F14200863  walk    101  -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.72  0.00  0.00  0.01
13:43:58 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.46  0.00  0.00  0.01
13:43:58 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:43:58 B17F.1   B17F14200863  walk    0    -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.66  0.00  0.00  0.01
13:43:59 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:00 B17F.1   B17F14200863  sit     115  -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.15  0.00  0.00  0.00
13:44:00 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.90  0.00  0.01  0.01
13:44:00 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.42  0.00  0.00  0.01
13:44:00 B17F.1   B17F14200863  sit     99   -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.03  0.00  0.00  0.00
13:44:00 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:00 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.42  0.00  0.00  0.01
13:44:01 B17F.1   B17F14200863  sit     107  -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.02  0.00  0.00  0.00
13:44:01 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:01 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.00  0.41  0.00  0.00  0.01
13:44:02 B17F.1   B17F14200863  sit     86   -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.02  0.00  0.00  0.00
13:44:02 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.00  0.40  0.00  0.00  0.01
13:44:02 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:03 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.00  0.39  0.00  0.00  0.01
13:44:03 B17F.1   B17F14200863  sit     90   -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.02  0.00  0.00  0.00
13:44:03 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:04 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:04 B17F.1   B17F14200863  stand   99   -        stand              trk  1.00 OpenFloor  2   0     0.00  0.00  0.06  0.00  0.00  0.00
13:44:04 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.00  0.39  0.00  0.00  0.01
13:44:05 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   9     0.00  0.00  0.38  0.00  0.00  0.01
13:44:05 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   9     0.00  0.01  0.93  0.00  0.00  0.01
13:44:05 B17F.1   B17F14200863  stand   84   -        stand              trk  1.00 OpenFloor  2   9     0.00  0.00  0.15  0.00  0.00  0.00
13:44:06 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   10    0.00  0.00  0.38  0.00  0.00  0.01
13:44:06 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  2   10    0.00  0.01  0.93  0.00  0.00  0.01
13:44:06 B17F.1   B17F14200863  stand   79   -        stand              trk  1.00 OpenFloor  2   10    0.00  0.00  0.21  0.00  0.00  0.00
13:44:07 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   11    0.00  0.01  0.68  0.01  0.00  0.01
13:44:07 B17F.1   B17F14200863  stand   71   -        stand              trk  1.00 OpenFloor  2   11    0.00  0.00  0.25  0.00  0.00  0.01
13:44:07 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   11    0.00  0.01  0.93  0.00  0.00  0.01
13:44:08 B17F.1   B17F14200863  stand   73   -        stand              trk  1.00 OpenFloor  2   12    0.00  0.01  0.44  0.01  0.00  0.01
13:44:08 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   12    0.00  0.01  0.93  0.00  0.00  0.01
13:44:08 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   12    0.00  0.01  0.84  0.00  0.00  0.01
13:44:09 B17F.0   B17F04325242  walk    81   -        walk               trk  1.00 Sit        2   0     0.00  0.01  0.92  0.00  0.00  0.01
13:44:09 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        2   0     0.00  0.01  0.26  0.01  0.00  0.02
13:44:09 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:10 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:10 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.01  0.05  0.01  0.00  0.01
13:44:10 B17F.0   B17F04325242  walk    68   -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:11 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:11 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.01  0.00  0.00  0.00
13:44:11 B17F.0   B17F04325242  walk    0    -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:12 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.02  0.00  0.00  0.00
13:44:12 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:12 B17F.0   B17F04325242  walk    80   -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:13 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.02  0.01  0.00  0.00
13:44:13 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:13 B17F.0   B17F04325242  walk    66   -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:14 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.02  0.01  0.00  0.00
13:44:14 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:14 B17F.0   B17F04325242  sit     68   -        sit                trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.00  0.02
13:44:15 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.02  0.01  0.00  0.00
13:44:15 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:15 B17F.0   B17F04325242  walk    81   -        walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.63  0.00  0.01  0.03
13:44:16 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.02  0.01  0.00  0.00
13:44:16 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:16 B17F.0   B17F04325242  walk    118  -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.87  0.00  0.01  0.01
13:44:17 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.01  0.00  0.00  0.00
13:44:17 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:17 B17F.0   B17F04325242  walk    126  -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.92  0.00  0.00  0.01
13:44:18 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:18 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.01  0.00  0.00  0.00
13:44:18 B17F.0   B17F04325242  walk    115  -        walk               trk  1.00 OpenFloor  2   0     0.00  0.00  0.99  0.00  0.00  0.00
13:44:19 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:19 B17F.1   B17F14200863  sit     85   -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.01  0.00  0.00  0.00
13:44:19 B17F.0   B17F04325242  walk    0    -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
13:44:20 B17F.0   B17F04325242  walk    0    -        walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:44:20 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:20 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.01  0.00  0.00  0.00
13:44:21 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.00  0.00  0.00  0.00
13:44:21 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:21 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.00  0.98  0.00  0.00  0.00
13:44:22 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.00  0.00  0.00  0.00
13:44:22 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:22 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
13:44:23 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:23 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.00  0.00  0.00  0.00
13:44:23 B17F.0   B17F04325242  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:44:24 B17F.0   -             stand   0    -        stand              room -    OpenFloor  1   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:24 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:24 B17F.1   B17F14200863  sit     87   -        sit                trk  1.00 OpenFloor  1   0     0.00  0.00  0.00  0.00  0.00  0.00
13:44:24 B17F.E   -             -       0    -        ExitRoom(rdr)      room -    OpenFloor  1   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:25 B17F.E   -             -       0    -        np=2               room -    OpenFloor  1   0     0.00  0.01  0.93  0.00  0.00  0.01
13:44:25 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  1   0     0.00  0.02  0.87  0.00  0.00  0.02
13:44:25 B17F.1   B17F14200863  sit     72   -        sit                trk  1.00 OpenFloor  1   0     0.00  0.00  0.02  0.00  0.00  0.00
13:44:26 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   8     0.00  0.02  0.86  0.00  0.01  0.02
13:44:26 B17F.1   B17F14200863  sit     80   -        sit                trk  1.00 OpenFloor  1   8     0.00  0.00  0.02  0.00  0.00  0.00
13:44:27 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   9     0.00  0.02  0.85  0.00  0.01  0.02
13:44:27 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   9     0.00  0.00  0.02  0.00  0.00  0.00
13:44:28 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   10    0.00  0.02  0.85  0.00  0.01  0.02
13:44:28 B17F.1   B17F14200863  sit     90   -        sit                trk  1.00 OpenFloor  1   10    0.00  0.00  0.02  0.00  0.00  0.00
13:44:29 B17F.1   B17F14200863  sit     85   -        sit                trk  1.00 OpenFloor  1   11    0.00  0.00  0.03  0.00  0.00  0.00
13:44:29 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   11    0.00  0.02  0.85  0.00  0.01  0.02
13:44:30 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   12    0.00  0.02  0.85  0.00  0.01  0.02
13:44:30 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   12    0.00  0.00  0.04  0.00  0.00  0.00
13:44:31 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   13    0.00  0.02  0.85  0.00  0.01  0.02
13:44:31 B17F.1   B17F14200863  sit     90   -        sit                trk  1.00 OpenFloor  1   13    0.00  0.00  0.02  0.00  0.00  0.00
13:44:32 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   14    0.00  0.02  0.85  0.00  0.01  0.02
13:44:32 B17F.1   B17F14200863  sit     95   -        sit                trk  1.00 OpenFloor  1   14    0.00  0.00  0.03  0.00  0.00  0.00
13:44:33 B17F.1   B17F14200863  sit     94   -        sit                trk  1.00 OpenFloor  1   15    0.00  0.00  0.04  0.00  0.00  0.00
13:44:33 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   15    0.00  0.02  0.85  0.00  0.01  0.02
13:44:34 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   16    0.00  0.02  0.85  0.00  0.01  0.02
13:44:34 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   16    0.00  0.00  0.02  0.00  0.00  0.00
13:44:35 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
13:44:35 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   17    0.00  0.00  0.01  0.00  0.00  0.00
13:44:36 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
13:44:36 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   18    0.00  0.00  0.01  0.00  0.00  0.00
13:44:37 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   19    0.00  0.02  0.85  0.00  0.01  0.02
13:44:37 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   19    0.00  0.00  0.01  0.00  0.00  0.00
13:44:38 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   20    0.00  0.00  0.01  0.00  0.00  0.00
13:44:38 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  1   20    0.00  0.02  0.85  0.00  0.01  0.02
13:44:39 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   21    0.00  0.00  0.01  0.00  0.00  0.00
13:44:39 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   21    0.00  0.02  0.85  0.00  0.01  0.02
13:44:40 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   22    0.00  0.00  0.01  0.00  0.00  0.00
13:44:40 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   22    0.00  0.02  0.84  0.00  0.01  0.03
13:44:41 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   23    0.00  0.00  0.01  0.00  0.00  0.00
13:44:41 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   23    0.00  0.02  0.82  0.00  0.01  0.05
13:44:42 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   24    0.00  0.00  0.01  0.00  0.00  0.01
13:44:42 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   24    0.00  0.02  0.79  0.00  0.02  0.08
13:44:43 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   25    0.00  0.00  0.01  0.00  0.00  0.01
13:44:43 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   25    0.00  0.02  0.73  0.00  0.03  0.13
13:44:44 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   26    0.00  0.00  0.01  0.00  0.00  0.01
13:44:44 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   26    0.00  0.02  0.66  0.00  0.04  0.20
13:44:45 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   27    0.00  0.00  0.02  0.00  0.00  0.03
13:44:45 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   27    0.00  0.01  0.54  0.00  0.06  0.32
13:44:46 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Left       1   28    0.00  0.01  0.37  0.00  0.08  0.49
13:44:46 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Left       1   28    0.00  0.00  0.03  0.01  0.01  0.08
13:44:47 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   29    0.00  0.01  0.19  0.00  0.09  0.69
13:44:47 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   29    0.00  0.00  0.03  0.01  0.02  0.14
13:44:48 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   30    0.00  0.00  0.07  0.00  0.08  0.84
13:44:48 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   30    0.00  0.00  0.01  0.00  0.01  0.07
13:44:49 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   31    0.00  0.00  0.01  0.00  0.01  0.07
13:44:49 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   31    0.00  0.00  0.01  0.00  0.06  0.93
13:44:50 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   32    0.00  0.00  0.01  0.00  0.01  0.10
13:44:50 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   32    0.00  0.00  0.00  0.00  0.04  0.96
13:44:51 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   33    0.00  0.00  0.01  0.00  0.01  0.14
13:44:51 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   33    0.00  0.00  0.00  0.00  0.02  0.98
13:44:52 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   34    0.00  0.00  0.01  0.00  0.01  0.22
13:44:52 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   34    0.00  0.00  0.00  0.00  0.02  0.98
13:44:53 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   35    0.00  0.00  0.01  0.00  0.01  0.34
13:44:53 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   35    0.00  0.00  0.00  0.00  0.01  0.99
13:44:54 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Left       1   36    0.00  0.00  0.00  0.00  0.02  0.52
13:44:54 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Left       1   36    0.00  0.00  0.00  0.00  0.01  0.99
13:44:55 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Left       1   37    0.00  0.00  0.00  0.00  0.02  0.68
13:44:55 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Left       1   37    0.00  0.00  0.00  0.00  0.01  0.99
13:44:56 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Left       1   38    0.00  0.00  0.00  0.00  0.01  0.99
13:44:56 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Left       1   38    0.00  0.00  0.00  0.00  0.03  0.79
13:44:57 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Left       1   39    0.00  0.00  0.00  0.00  0.01  0.99
13:44:57 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Left       1   39    0.00  0.00  0.00  0.00  0.03  0.86
13:44:58 B17F.1   B17F14200863  sit     91   -        sit                trk  1.00 Left       1   40    0.00  0.00  0.01  0.00  0.20  0.42
13:44:58 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 Left       1   40    0.00  0.00  0.00  0.00  0.01  0.99
13:44:59 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   41    0.00  0.00  0.01  0.00  0.04  0.96
13:44:59 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   41    0.00  0.01  0.05  0.01  0.12  0.01
13:45:00 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   42    0.00  0.00  0.00  0.00  0.01  0.99
13:45:00 B17F.1   B17F14200863  sit     98   -        sit                trk  1.00 Sit        1   42    0.00  0.00  0.02  0.00  0.01  0.03
13:45:01 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   43    0.00  0.00  0.00  0.00  0.01  0.99
13:45:01 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   43    0.00  0.00  0.02  0.00  0.00  0.03
13:45:02 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   44    0.00  0.00  0.00  0.00  0.01  0.99
13:45:02 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   44    0.00  0.00  0.01  0.00  0.00  0.27
13:45:03 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   45    0.00  0.00  0.00  0.00  0.01  0.99
13:45:03 B17F.1   B17F14200863  sit     103  -        sit                trk  1.00 Sit        1   45    0.00  0.00  0.01  0.00  0.02  0.07
13:45:04 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   46    0.00  0.00  0.00  0.00  0.01  0.99
13:45:04 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   46    0.00  0.00  0.02  0.00  0.01  0.04
13:45:05 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   47    0.00  0.00  0.00  0.00  0.01  0.99
13:45:05 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   47    0.00  0.00  0.01  0.00  0.00  0.28
13:45:06 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 Sit        1   48    0.00  0.00  0.00  0.00  0.01  0.99
13:45:06 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   48    0.00  0.00  0.00  0.00  0.01  0.49
13:45:07 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 Left       1   49    0.00  0.00  0.00  0.00  0.01  0.99
13:45:07 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Left       1   49    0.00  0.00  0.00  0.00  0.02  0.66
13:45:08 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Left       1   50    0.00  0.00  0.00  0.00  0.01  0.99
13:45:08 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Left       1   50    0.00  0.00  0.00  0.00  0.03  0.77
13:45:09 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Left       1   51    0.00  0.00  0.00  0.00  0.01  0.99
13:45:09 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Left       1   51    0.00  0.00  0.00  0.00  0.03  0.85
13:45:10 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Sit        1   52    0.00  0.00  0.00  0.00  0.85  0.15
13:45:10 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   52    0.00  0.00  0.01  0.00  0.31  0.05
13:45:11 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 Sit        1   53    0.00  0.01  0.15  0.00  0.80  0.02
13:45:11 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Sit        1   53    0.00  0.00  0.01  0.00  0.04  0.00
13:45:12 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 Empty      1   54    0.00  0.02  0.40  0.00  0.53  0.01
13:45:12 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 Empty      1   54    0.00  0.00  0.01  0.00  0.00  0.00
13:45:13 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   55    0.00  0.02  0.64  0.00  0.26  0.01
13:45:13 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   55    0.00  0.00  0.01  0.00  0.00  0.00
13:45:14 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   56    0.00  0.02  0.77  0.00  0.11  0.02
13:45:14 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   56    0.00  0.00  0.01  0.00  0.00  0.00
13:45:15 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   57    0.00  0.02  0.82  0.00  0.04  0.02
13:45:15 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   57    0.00  0.00  0.01  0.00  0.00  0.00
13:45:16 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   58    0.00  0.00  0.01  0.00  0.00  0.00
13:45:16 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   58    0.00  0.02  0.84  0.00  0.02  0.02
13:45:17 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   59    0.00  0.02  0.84  0.00  0.01  0.02
13:45:17 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   59    0.00  0.00  0.01  0.00  0.00  0.00
13:45:18 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   60    0.00  0.02  0.85  0.00  0.01  0.02
13:45:18 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   60    0.00  0.00  0.01  0.00  0.00  0.00
13:45:19 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   61    0.00  0.02  0.85  0.00  0.01  0.02
13:45:19 B17F.1   B17F14200863  sit     85   -        sit                trk  1.00 OpenFloor  1   61    0.00  0.00  0.02  0.00  0.00  0.00
13:45:20 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   62    0.00  0.00  0.01  0.00  0.00  0.00
13:45:20 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   62    0.00  0.02  0.85  0.00  0.01  0.02
13:45:21 B17F.1   B17F14200863  sit     78   -        sit                trk  1.00 OpenFloor  1   63    0.00  0.00  0.01  0.00  0.00  0.00
13:45:21 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   63    0.00  0.02  0.85  0.00  0.01  0.02
13:45:22 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   64    0.00  0.02  0.85  0.00  0.01  0.02
13:45:22 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   64    0.00  0.00  0.01  0.00  0.00  0.00
13:45:23 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   65    0.00  0.02  0.85  0.00  0.01  0.02
13:45:23 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   65    0.00  0.00  0.02  0.00  0.00  0.00
13:45:24 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   66    0.00  0.02  0.85  0.00  0.01  0.02
13:45:24 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   66    0.00  0.00  0.02  0.01  0.00  0.00
13:45:25 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   67    0.00  0.02  0.85  0.00  0.01  0.02
13:45:25 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   67    0.00  0.00  0.01  0.00  0.00  0.00
13:45:26 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   68    0.00  0.02  0.85  0.00  0.01  0.02
13:45:26 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   68    0.00  0.00  0.01  0.00  0.00  0.00
13:45:27 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   69    0.00  0.02  0.85  0.00  0.01  0.02
13:45:27 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   69    0.00  0.00  0.01  0.00  0.00  0.00
13:45:28 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   70    0.00  0.02  0.85  0.00  0.01  0.02
13:45:28 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   70    0.00  0.00  0.01  0.00  0.00  0.00
13:45:29 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   71    0.00  0.00  0.01  0.00  0.00  0.00
13:45:29 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   71    0.00  0.02  0.85  0.00  0.01  0.02
13:45:30 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   72    0.00  0.02  0.85  0.00  0.01  0.02
13:45:30 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   72    0.00  0.00  0.01  0.00  0.00  0.00
13:45:31 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   73    0.00  0.00  0.01  0.00  0.00  0.00
13:45:31 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   73    0.00  0.02  0.85  0.00  0.01  0.02
13:45:32 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   74    0.00  0.02  0.85  0.00  0.01  0.02
13:45:32 B17F.1   B17F14200863  sit     95   -        sit                trk  1.00 OpenFloor  1   74    0.00  0.00  0.02  0.00  0.00  0.00
13:45:33 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   75    0.00  0.02  0.85  0.00  0.01  0.02
13:45:33 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   75    0.00  0.00  0.02  0.00  0.00  0.00
13:45:34 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   76    0.00  0.02  0.85  0.00  0.01  0.02
13:45:34 B17F.1   B17F14200863  sit     0    -        sit                trk  1.00 OpenFloor  1   76    0.00  0.00  0.01  0.00  0.00  0.00
13:45:35 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   77    0.00  0.02  0.85  0.00  0.01  0.02
13:45:35 B17F.1   B17F14200863  sit     91   -        sit                trk  1.00 OpenFloor  1   77    0.00  0.00  0.02  0.00  0.00  0.00
13:45:36 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   78    0.00  0.02  0.85  0.00  0.01  0.02
13:45:36 B17F.1   B17F14200863  sit     83   -        sit                trk  1.00 OpenFloor  1   78    0.00  0.00  0.03  0.00  0.00  0.00
13:45:37 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   79    0.00  0.02  0.85  0.00  0.01  0.02
13:45:37 B17F.1   B17F14200863  sit     106  -        sit                trk  1.00 OpenFloor  1   79    0.00  0.00  0.04  0.00  0.00  0.00
13:45:38 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   80    0.00  0.02  0.85  0.00  0.01  0.02
13:45:38 B17F.1   B17F14200863  sit     94   -        sit                trk  1.00 OpenFloor  1   80    0.00  0.00  0.04  0.00  0.00  0.00
13:45:39 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   81    0.00  0.02  0.85  0.00  0.01  0.02
13:45:39 B17F.1   B17F14200863  sit     79   -        sit                trk  1.00 OpenFloor  1   81    0.00  0.00  0.02  0.00  0.00  0.00
13:45:40 B17F.1   B17F14200863  sit     71   -        sit                trk  1.00 OpenFloor  1   82    0.00  0.00  0.01  0.00  0.00  0.00
13:45:40 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   82    0.00  0.02  0.85  0.00  0.01  0.02
13:45:41 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:41 B17F.1   B17F14200863  stand   82   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.03  0.00  0.00  0.00
13:45:42 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:42 B17F.1   B17F14200863  walk    55   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.01  0.22  0.02  0.00  0.01
13:45:43 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:43 B17F.1   B17F14200863  walk    63   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.01  0.54  0.02  0.01  0.02
13:45:44 B17F.1   B17F14200863  walk    95   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.73  0.01  0.01  0.02
13:45:44 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:45 B17F.1   B17F14200863  walk    84   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.81  0.01  0.01  0.02
13:45:45 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:46 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:46 B17F.1   B17F14200863  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
13:45:47 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:47 B17F.1   B17F14200863  walk    75   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:48 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:48 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.01  0.96  0.00  0.00  0.01
13:45:49 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:49 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.01  0.94  0.00  0.00  0.01
13:45:49 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:49 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.87  0.00  0.00  0.02
13:45:50 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.01  0.96  0.00  0.00  0.01
13:45:50 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:51 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:51 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.01  0.97  0.00  0.00  0.01
13:45:52 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:52 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.01  0.97  0.00  0.00  0.01
13:45:53 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:53 B17F.1   B17F14200863  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.01  0.97  0.00  0.00  0.01
13:45:53 B17F.E1  B17F14200863  -       0    -        ExitRoom(rdr)      trk  1.00 OpenFloor  1   0     0.00  0.01  0.97  0.00  0.00  0.01
13:45:54 B17F.E   -             -       0    -        np=1               room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:45:54 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  0   0     0.00  0.04  0.74  0.00  0.02  0.04
13:45:55 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.68  0.01  0.03  0.04
13:45:56 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.65  0.01  0.04  0.03
13:45:57 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.63  0.01  0.04  0.03
13:45:58 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.62  0.02  0.05  0.03
13:45:59 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.62  0.02  0.05  0.03
13:46:00 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:01 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:02 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:03 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:04 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:05 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:06 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:07 B17F.2   B17F24200863  stand   0    -        stand              trk  0.03 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:08 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:09 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.61  0.02  0.05  0.04
13:46:10 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.60  0.02  0.06  0.05
13:46:11 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.58  0.02  0.06  0.06
13:46:12 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.56  0.02  0.08  0.08
13:46:13 B17F.2   B17F24200863  stand   0    -        stand              trk  0.02 OpenFloor  0   0     0.01  0.05  0.54  0.02  0.09  0.10
13:46:14 B17F.E   -             -       0    -        np=0  ★0           room -    OpenFloor  0   0     0.01  0.05  0.54  0.02  0.09  0.10
13:46:14 B17F.88  -             88      -    -        no-target(88)      room -    OpenFloor  0   0     0.01  0.05  0.56  0.02  0.13  0.04
13:46:15 B17F.88  -             88      -    -        no-target(88)      room -    OpenFloor  0   0     0.01  0.07  0.41  0.03  0.15  0.04
13:46:16 B17F.88  -             88      -    -        no-target(88)      room -    OpenFloor  0   0     0.02  0.08  0.31  0.04  0.17  0.04
13:46:17 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.08  0.31  0.04  0.17  0.04
13:46:18 B17F.E   -             -       0    -        np=1               room -    OpenFloor  0   0     0.02  0.08  0.31  0.04  0.17  0.04
13:46:18 B17F.E   -             -       0    -        EnterRoom(rdr)     room -    OpenFloor  0   0     0.02  0.08  0.31  0.04  0.17  0.04
13:46:18 B17F.0   B17F04618762  stand   111  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.26  0.00  0.69  0.03
13:46:19 B17F.0   B17F04618762  stand   89   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.35  0.00  0.54  0.02
13:46:20 B17F.0   B17F04618762  stand   107  -        stand              trk  1.00 BlindOpen  1   0     0.00  0.04  0.42  0.01  0.41  0.02
13:46:21 B17F.0   B17F04618762  stand   99   -        stand              trk  1.00 BlindOpen  1   0     0.00  0.04  0.48  0.01  0.30  0.02
13:46:22 B17F.0   B17F04618762  stand   116  -        stand              trk  1.00 Empty      1   0     0.00  0.05  0.53  0.01  0.22  0.03
13:46:23 B17F.0   B17F04618762  stand   130  -        stand              trk  1.00 Empty      1   0     0.00  0.05  0.56  0.01  0.16  0.03
13:46:24 B17F.0   B17F04618762  walk    116  -        walk               trk  1.00 Empty      1   0     0.00  0.03  0.73  0.01  0.08  0.02
13:46:25 B17F.0   B17F04618762  walk    0    -        walk               trk  1.00 Empty      1   0     0.00  0.04  0.68  0.01  0.06  0.03
13:46:26 B17F.0   B17F04618762  stand   0    -        stand              trk  1.00 Empty      1   0     0.00  0.05  0.65  0.01  0.06  0.03
13:46:27 B17F.0   B17F04618762  stand   0    -        stand              trk  1.00 Empty      1   0     0.01  0.05  0.63  0.01  0.05  0.03
13:46:28 B17F.0   B17F04618762  stand   0    -        stand              trk  1.00 Empty      1   0     0.01  0.05  0.62  0.02  0.05  0.03
13:46:29 B17F.0   B17F04618762  stand   0    -        stand              trk  1.00 Empty      1   0     0.01  0.05  0.62  0.02  0.05  0.03
13:46:30 B17F.0   B17F04618762  stand   0    -        stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:46:30 B17F.E   -             -       0    -        ExitRoom(rdr)      room -    Empty      1   0     0.05  0.10  0.16  0.11  0.21  0.02
13:46:31 B17F.E   -             -       0    -        np=0  ★0           room -    Empty      1   0     0.05  0.10  0.16  0.11  0.21  0.02
13:46:31 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.06  0.10  0.16  0.11  0.21  0.02
13:46:32 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:33 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:40 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:46:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.06  0.10  0.16  0.12  0.21  0.02
13:47:12 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:44 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:47:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.21  0.02
13:48:15 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:47 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:48:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:19 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:51 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:49:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
13:50:23 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:54 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:50:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
13:51:26 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:58 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:51:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.20  0.02
13:52:29 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:52:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:53:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
13:53:01 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:33 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:53:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:05 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:36 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:54:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:55:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:55:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:55:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:55:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:55:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:55:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:55:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:55:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
13:55:08 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:40 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:55:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:12 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:43 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:56:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
13:57:15 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:47 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:57:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:19 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:58:50 B17F.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
13:42:00.863 B17F.2    stand  255  -50    20     0     80        
13:42:00.863 B17F.1    stand  255  -190   50     0     80   143  
13:42:01.850 B17F.1    stand  255  -190   50     0     80   0    
13:42:01.850 B17F.2    stand  255  -50    20     0     80   143  
13:42:02.813 B17F.1    stand  255  -190   50     0     80   143  
13:42:02.813 B17F.2    stand  255  -50    20     0     80   143  
13:42:03.816 B17F.2    stand  255  -50    20     0     80   0    
13:42:03.816 B17F.1    stand  255  -190   50     0     80   143  
13:42:04.813 B17F.2    stand  255  -50    20     0     80   143  
13:42:04.813 B17F.1    stand  255  -190   50     0     80   143  
13:42:05.816 B17F.1    stand  255  -190   50     0     80   0    
13:42:05.816 B17F.2    stand  255  -50    20     0     80   143  
13:42:06.714 B17F.2    stand  255  -50    20     0     80   0    
13:42:06.714 B17F.1    stand  255  -190   50     0     80   143  
13:42:07.708 B17F.2    stand  255  -50    20     0     80   143  
13:42:07.708 B17F.1    stand  255  -190   50     0     80   143  
13:42:08.712 B17F.1    stand  255  -190   50     0     80   0    
13:42:08.712 B17F.2    stand  255  -50    20     0     80   143  
13:42:09.718 B17F.2    stand  255  -50    20     0     80   0    
13:42:09.718 B17F.1    stand  255  -190   50     0     80   143  
13:42:10.712 B17F.2    stand  255  -50    20     0     80   143  
13:42:10.712 B17F.1    stand  255  -190   50     0     80   143  
13:42:11.712 B17F.1    stand  255  -190   50     0     80   0    
13:42:11.712 B17F.2    stand  255  -50    20     0     80   143  
13:42:12.713 B17F.2    stand  255  -50    20     0     80   0    
13:42:12.713 B17F.1    stand  255  -190   50     0     80   143  
13:42:13.717 B17F.2    stand  255  -50    20     0     80   143  
13:42:13.717 B17F.1    stand  255  -200   60     0     80   155  
13:42:14.720 B17F.2    stand  255  -50    20     0     80   155  
13:42:14.720 B17F.1    stand  255  -210   100    86    80   178  
13:42:15.721 B17F.2    stand  255  -50    20     0     80   178  
13:42:15.721 B17F.1    stand  255  -210   110    0     80   183  
13:42:16.719 B17F.1    stand  255  -210   100    0     80   10   
13:42:16.719 B17F.2    stand  255  -50    20     0     80   178  
13:42:17.722 B17F.2    stand  255  -50    20     0     80   0    
13:42:17.722 B17F.1    stand  255  -190   100    98    80   161  
13:42:18.639 B17F.2    stand  255  -50    20     0     80   161  
13:42:18.639 B17F.1    stand  255  -150   70     78    80   111  
13:42:19.614 B17F.1    walk   255  -140   80     82    80   14   
13:42:19.614 B17F.2    stand  255  -50    20     0     80   108  
13:42:20.625 B17F.2    stand  255  -50    20     0     80   0    
13:42:20.625 B17F.1    walk   255  -110   100    77    80   100  
13:42:21.617 B17F.2    stand  255  -50    20     0     80   100  
13:42:21.617 B17F.1    walk   255  -60    130    74    80   110  
13:42:22.657 B17F.1    walk   255  0      150    71    80   63   
13:42:22.657 B17F.2    stand  255  -50    20     0     80   139  
13:42:23.626 B17F.2    stand  255  -50    20     0     80   0    
13:42:23.626 B17F.1    walk   255  20     150    92    80   147  
13:42:24.622 B17F.2    stand  255  -50    20     0     80   147  
13:42:24.622 B17F.1    walk   255  0      130    80    80   120  
13:42:25.638 B17F.1    walk   255  -30    140    85    80   31   
13:42:25.638 B17F.2    stand  255  -60    30     0     80   114  
13:42:26.643 B17F.2    stand  255  -60    30     0     80   0    
13:42:26.643 B17F.1    walk   255  -90    110    77    80   85   
13:42:27.641 B17F.2    stand  255  -60    30     0     80   85   
13:42:27.641 B17F.1    walk   255  -170   120    93    80   142  
13:42:28.539 B17F.1    walk   255  -200   120    0     80   30   
13:42:28.539 B17F.2    stand  255  -60    30     0     80   166  
13:42:29.533 B17F.1    walk   255  -200   120    0     80   166  
13:42:29.533 B17F.2    stand  255  -60    30     0     80   166  
13:42:30.544 B17F.2    stand  255  -60    30     0     80   0    
13:42:30.544 B17F.1    walk   255  -200   130    0     80   172  
13:42:31.536 B17F.1    walk   255  -200   80     89    80   50   
13:42:31.536 B17F.2    stand  255  -60    30     0     80   148  
13:42:32.538 B17F.1    stand  255  -190   60     0     80   133  
13:42:32.538 B17F.2    stand  255  -60    30     0     80   133  
13:42:33.540 B17F.2    stand  255  -60    30     0     80   0    
13:42:33.540 B17F.1    stand  255  -190   60     0     80   133  
13:42:34.552 B17F.2    stand  255  -60    30     0     80   133  
13:42:34.552 B17F.1    stand  255  -190   60     0     80   133  
13:42:35.541 B17F.1    stand  255  -190   60     0     80   0    
13:42:35.541 B17F.2    stand  255  -60    30     0     80   133  
13:42:36.585 B17F.2    stand  255  -60    30     0     80   0    
13:42:36.585 B17F.1    stand  255  -150   40     76    80   90   
13:42:37.546 B17F.2    stand  255  -60    30     0     80   90   
13:42:37.546 B17F.1    stand  255  -130   70     60    80   80   
13:42:38.546 B17F.1    stand  255  -130   50     99    80   20   
13:42:38.546 B17F.2    stand  255  -60    30     0     80   72   
13:42:39.543 B17F.2    stand  255  -60    30     0     80   0    
13:42:39.543 B17F.1    stand  255  -130   50     0     80   72   
13:42:40.442 B17F.2    stand  255  -60    30     0     80   72   
13:42:40.442 B17F.1    stand  255  -130   60     0     80   76   
13:42:41.445 B17F.1    stand  255  -130   60     0     80   0    
13:42:41.445 B17F.2    stand  255  -60    30     0     80   76   
13:42:42.452 B17F.2    stand  255  -60    30     0     80   0    
13:42:42.452 B17F.1    stand  255  -130   50     0     80   72   
13:42:43.455 B17F.2    stand  255  -60    20     0     80   76   
13:42:43.455 B17F.1    stand  255  -130   50     0     80   76   
13:42:44.455 B17F.1    stand  255  -130   50     0     80   0    
13:42:44.455 B17F.2    stand  255  -50    20     0     80   85   
13:42:45.454 B17F.2    stand  255  -50    30     0     80   10   
13:42:45.454 B17F.1    stand  255  -130   50     0     80   82   
13:42:46.460 B17F.2    stand  255  -50    30     0     80   82   
13:42:46.460 B17F.1    stand  255  -130   50     0     80   82   
13:42:47.456 B17F.1    stand  255  -130   50     0     80   0    
13:42:47.456 B17F.2    stand  255  -50    30     0     80   82   
13:42:48.466 B17F.2    stand  255  -50    30     0     80   0    
13:42:48.466 B17F.1    stand  255  -130   50     0     80   82   
13:42:49.460 B17F.2    stand  255  -50    30     0     80   82   
13:42:49.460 B17F.1    stand  255  -130   50     0     80   82   
13:42:50.363 B17F.1    stand  255  -130   50     0     80   0    
13:42:50.363 B17F.2    stand  255  -50    30     0     80   82   
13:42:51.358 B17F.2    stand  255  -50    30     0     80   0    
13:42:51.358 B17F.1    stand  255  -130   50     0     80   82   
13:42:52.360 B17F.1    stand  255  -130   50     0     80   0    
13:42:52.360 B17F.2    stand  255  -50    30     0     80   82   
13:42:53.357 B17F.2    stand  255  -50    30     0     80   0    
13:42:53.357 B17F.1    stand  255  -130   50     0     80   82   
13:42:54.362 B17F.2    stand  255  -50    30     0     80   82   
13:42:54.362 B17F.1    stand  255  -130   50     0     80   82   
13:42:55.362 B17F.1    stand  255  -130   50     0     80   0    
13:42:55.362 B17F.2    stand  255  -50    30     0     80   82   
13:42:56.362 B17F.2    stand  255  -50    30     0     80   0    
13:42:56.362 B17F.1    stand  255  -130   50     0     80   82   
13:42:57.362 B17F.2    stand  255  -50    30     0     80   82   
13:42:57.362 B17F.1    stand  255  -130   50     0     80   82   
13:42:58.303 B17F.1    stand  255  -130   50     0     80   0    
13:42:58.303 B17F.2    stand  255  -50    30     0     80   82   
13:42:59.294 B17F.2    stand  255  -50    30     0     80   0    
13:42:59.294 B17F.1    stand  255  -130   50     0     80   82   
13:43:00.354 B17F.2    stand  255  -50    30     0     80   82   
13:43:00.354 B17F.1    stand  255  -130   50     0     80   82   
13:43:01.295 B17F.1    stand  255  -130   50     0     80   0    
13:43:01.295 B17F.2    stand  255  -50    30     0     80   82   
13:43:02.301 B17F.2    stand  255  -50    30     0     80   0    
13:43:02.301 B17F.1    stand  255  -130   50     0     80   82   
13:43:03.301 B17F.1    stand  255  -130   50     0     80   0    
13:43:03.301 B17F.2    stand  255  -50    30     0     80   82   
13:43:04.297 B17F.2    stand  255  -50    30     0     80   0    
13:43:04.297 B17F.1    stand  255  -130   50     0     80   82   
13:43:05.303 B17F.1    stand  255  -110   60     57    80   22   
13:43:05.303 B17F.2    stand  255  -60    30     0     80   58   
13:43:06.304 B17F.2    stand  255  -60    20     83    80   10   
13:43:06.304 B17F.1    stand  255  -110   60     0     80   64   
13:43:07.306 B17F.1    stand  255  -110   70     0     80   10   
13:43:07.306 B17F.2    stand  255  -50    40     94    80   67   
13:43:08.305 B17F.2    stand  255  -30    50     97    80   22   
13:43:08.305 B17F.1    stand  255  -110   80     0     80   85   
13:43:09.303 B17F.1    stand  255  -110   100    0     80   20   
13:43:09.303 B17F.2    walk   255  0      80     75    80   111  
13:43:10.201 B17F.2    walk   255  10     120    67    80   41   
13:43:10.201 B17F.1    stand  255  -130   90     0     80   143  
13:43:11.206 B17F.1    stand  255  -120   90     0     80   10   
13:43:11.206 B17F.2    walk   255  50     90     67    80   170  
13:43:12.207 B17F.1    stand  255  -120   80     0     80   170  
13:43:12.207 B17F.2    walk   255  50     100    71    80   171  
13:43:13.205 B17F.2    walk   255  50     90     87    80   10   
13:43:13.205 B17F.1    stand  255  -120   80     0     80   170  
13:43:14.175 B17F.1    stand  255  -120   80     0     80   0    
13:43:14.175 B17F.2    sit    255  50     80     78    80   170  
13:43:15.165 B17F.2    sit    255  50     80     0     80   0    
13:43:15.165 B17F.1    stand  255  -120   80     0     80   170  
13:43:16.173 B17F.1    stand  255  -120   80     0     80   0    
13:43:16.173 B17F.2    sit    255  60     90     84    80   180  
13:43:17.169 B17F.2    stand  255  150    50     0     80   98   
13:43:17.169 B17F.1    stand  255  -130   60     60    80   280  
13:43:18.179 B17F.1    stand  255  -170   30     69    80   50   
13:43:18.179 B17F.2    stand  255  180    40     0     80   350  
13:43:19.171 B17F.2    stand  255  180    50     0     80   10   
13:43:19.171 B17F.1    stand  255  -160   100    89    80   343  
13:43:20.174 B17F.1    walk   255  -130   150    127   80   58   
13:43:20.174 B17F.2    stand  255  170    50     0     80   316  
13:43:21.170 B17F.2    stand  255  170    50     0     80   0    
13:43:21.170 B17F.1    walk   255  -150   170    104   80   341  
13:43:22.175 B17F.1    walk   255  -170   220    127   80   53   
13:43:22.175 B17F.2    stand  255  170    50     0     80   380  
13:43:23.190 B17F.1    walk   255  -190   210    125   80   393  
13:43:23.190 B17F.2    stand  255  170    50     0     80   393  
13:43:24.178 B17F.1    walk   255  -170   170    113   80   360  
13:43:24.178 B17F.2    stand  255  170    50     0     80   360  
13:43:25.242 B17F.2    stand  255  170    50     0     80   0    
13:43:25.242 B17F.0    stand  255  60     100    89    80   120  
13:43:25.242 B17F.1    walk   255  -170   230    0     80   264  
13:43:26.070 B17F.1    walk   255  -150   270    0     80   44   
13:43:26.070 B17F.2    stand  255  170    50     0     80   388  
13:43:26.070 B17F.0    stand  255  60     110    82    80   125  
13:43:27.072 B17F.2    stand  255  170    50     0     80   125  
13:43:27.072 B17F.1    walk   255  -100   290    111   80   361  
13:43:27.072 B17F.0    stand  255  60     110    0     80   240  
13:43:28.078 B17F.1    walk   255  -130   290    0     80   261  
13:43:28.078 B17F.0    stand  255  60     110    0     80   261  
13:43:28.078 B17F.2    stand  255  170    50     0     80   125  
13:43:29.073 B17F.2    stand  255  170    50     0     80   0    
13:43:29.073 B17F.0    stand  255  60     110    0     80   125  
13:43:29.073 B17F.1    walk   255  -170   240    115   80   264  
13:43:30.088 B17F.2    stand  255  170    50     0     80   389  
13:43:30.088 B17F.0    stand  255  50     90     0     80   126  
13:43:30.088 B17F.1    walk   255  -180   160    102   80   240  
13:43:31.086 B17F.2    stand  255  170    50     0     80   366  
13:43:31.086 B17F.0    stand  255  60     90     89    80   117  
13:43:31.086 B17F.1    walk   255  -170   140    0     80   235  
13:43:32.085 B17F.2    stand  255  170    50     0     80   351  
13:43:32.085 B17F.0    stand  255  60     90     0     80   117  
13:43:32.085 B17F.1    walk   255  -160   140    0     80   225  
13:43:33.090 B17F.2    stand  255  170    50     0     80   342  
13:43:33.090 B17F.0    stand  255  60     90     94    80   117  
13:43:33.090 B17F.1    stand  255  -170   140    0     80   235  
13:43:34.084 B17F.2    stand  255  170    50     0     80   351  
13:43:34.084 B17F.0    stand  255  60     100    98    80   120  
13:43:34.084 B17F.1    stand  255  -170   140    0     80   233  
13:43:35.084 B17F.2    stand  255  170    50     0     80   351  
13:43:35.084 B17F.0    stand  255  60     90     0     80   117  
13:43:35.084 B17F.1    stand  255  -170   130    88    80   233  
13:43:36.115 B17F.2    stand  255  170    50     0     80   349  
13:43:36.115 B17F.0    stand  255  60     90     0     80   117  
13:43:36.115 B17F.1    stand  255  -180   140    99    80   245  
13:43:36.988 B17F.0    stand  255  60     90     0     80   245  
13:43:36.988 B17F.2    stand  255  170    50     0     80   117  
13:43:36.988 B17F.1    stand  255  -210   110    86    80   384  
13:43:37.985 B17F.1    stand  255  -210   120    80    80   10   
13:43:37.985 B17F.2    stand  255  170    50     0     80   386  
13:43:37.985 B17F.0    stand  255  60     90     0     80   117  
13:43:38.986 B17F.1    stand  255  -210   100    0     80   270  
13:43:38.986 B17F.2    stand  255  190    60     0     80   401  
13:43:38.986 B17F.0    stand  255  60     90     0     80   133  
13:43:39.984 B17F.2    stand  255  220    90     0     80   160  
13:43:39.984 B17F.1    stand  255  -210   120    0     80   431  
13:43:39.984 B17F.0    stand  255  60     90     0     80   271  
13:43:40.996 B17F.1    stand  255  -210   120    0     80   271  
13:43:40.996 B17F.0    stand  255  60     90     0     80   271  
13:43:40.996 B17F.2    stand  255  220    90     0     80   160  
13:43:42.000 B17F.2    stand  255  220    90     0     80   0    
13:43:42.000 B17F.0    stand  255  60     90     0     80   160  
13:43:42.000 B17F.1    stand  255  -210   120    0     80   271  
13:43:42.987 B17F.2    stand  255  210    90     0     80   421  
13:43:42.987 B17F.0    stand  255  60     90     0     80   150  
13:43:42.987 B17F.1    stand  255  -210   100    104   80   270  
13:43:43.994 B17F.2    stand  255  210    80     0     80   420  
13:43:43.994 B17F.0    stand  255  60     90     0     80   150  
13:43:43.994 B17F.1    stand  255  -210   120    0     80   271  
13:43:44.992 B17F.2    stand  255  210    80     0     80   421  
13:43:44.992 B17F.0    stand  255  60     90     0     80   150  
13:43:44.992 B17F.1    stand  255  -210   100    88    80   270  
13:43:45.991 B17F.2    stand  255  210    90     0     80   420  
13:43:45.991 B17F.0    stand  255  60     90     0     80   150  
13:43:45.991 B17F.1    stand  255  -200   60     93    80   261  
13:43:46.993 B17F.2    stand  255  190    110    0     80   393  
13:43:46.993 B17F.0    stand  255  60     90     0     80   131  
13:43:46.993 B17F.1    stand  255  -200   60     0     80   261  
13:43:47.995 B17F.2    stand  255  250    90     0     80   450  
13:43:47.995 B17F.0    stand  255  60     90     0     80   190  
13:43:47.995 B17F.1    stand  255  -190   60     0     80   251  
13:43:48.892 B17F.2    stand  255  260    100    0     80   451  
13:43:48.892 B17F.0    stand  255  60     90     0     80   200  
13:43:48.892 B17F.1    stand  255  -170   80     68    80   230  
13:43:49.886 B17F.2    stand  255  250    100    0     80   420  
13:43:49.886 B17F.0    stand  255  60     90     0     80   190  
13:43:49.886 B17F.1    walk   255  -160   170    115   80   234  
13:43:50.890 B17F.0    stand  255  60     110    0     80   228  
13:43:50.890 B17F.2    stand  255  250    100    0     80   190  
13:43:50.890 B17F.1    walk   255  -170   240    0     80   442  
13:43:51.913 B17F.2    stand  255  250    100    0     80   442  
13:43:51.913 B17F.1    walk   255  -170   310    119   80   469  
13:43:51.913 B17F.0    stand  255  60     90     0     80   318  
13:43:52.890 B17F.1    walk   255  -120   340    115   80   308  
13:43:52.890 B17F.0    stand  255  50     80     0     80   310  
13:43:52.890 B17F.2    stand  255  250    100    0     80   200  
13:43:53.897 B17F.2    stand  255  250    100    0     80   0    
13:43:53.897 B17F.0    stand  255  50     80     0     80   200  
13:43:53.897 B17F.1    walk   255  -60    300    91    80   245  
13:43:54.894 B17F.2    stand  255  250    100    0     80   368  
13:43:54.894 B17F.0    stand  255  50     80     0     80   200  
13:43:54.894 B17F.1    walk   255  0      250    78    80   177  
13:43:55.898 B17F.2    stand  255  250    100    0     80   291  
13:43:55.898 B17F.0    stand  255  60     100    0     80   190  
13:43:55.898 B17F.1    walk   255  20     200    76    80   107  
13:43:56.927 B17F.2    stand  255  250    100    0     80   250  
13:43:56.927 B17F.0    stand  255  60     100    0     80   190  
13:43:56.927 B17F.1    walk   255  30     190    66    80   94   
13:43:57.903 B17F.2    stand  255  250    100    0     80   237  
13:43:57.903 B17F.0    stand  255  60     100    0     80   190  
13:43:57.903 B17F.1    walk   255  40     190    101   80   92   
13:43:58.896 B17F.0    stand  255  60     100    0     80   92   
13:43:58.896 B17F.2    stand  255  250    100    0     80   190  
13:43:58.896 B17F.1    walk   255  50     170    0     80   211  
13:44:00.045 B17F.1    sit    255  50     160    115   80   10   
13:44:00.045 B17F.2    stand  255  250    100    0     80   208  
13:44:00.045 B17F.0    stand  255  60     100    0     80   190  
13:44:00.811 B17F.1    sit    255  50     160    99    80   60   
13:44:00.811 B17F.2    stand  255  250    100    0     80   208  
13:44:00.811 B17F.0    stand  255  60     100    0     80   190  
13:44:01.797 B17F.1    sit    255  50     170    107   80   70   
13:44:01.797 B17F.2    stand  255  250    100    0     80   211  
13:44:01.797 B17F.0    stand  255  60     100    0     80   190  
13:44:02.793 B17F.1    sit    255  50     180    86    80   80   
13:44:02.793 B17F.0    stand  255  60     100    0     80   80   
13:44:02.793 B17F.2    stand  255  250    100    0     80   190  
13:44:03.796 B17F.0    stand  255  60     100    0     80   190  
13:44:03.796 B17F.1    sit    255  50     190    90    80   90   
13:44:03.796 B17F.2    stand  255  250    100    0     80   219  
13:44:04.804 B17F.2    stand  255  250    100    0     80   0    
13:44:04.804 B17F.1    stand  255  50     180    99    80   215  
13:44:04.804 B17F.0    stand  255  60     100    0     80   80   
13:44:05.796 B17F.0    stand  255  60     100    0     80   0    
13:44:05.796 B17F.2    stand  255  250    100    0     80   190  
13:44:05.796 B17F.1    stand  255  40     180    84    80   224  
13:44:06.815 B17F.0    stand  255  60     100    0     80   82   
13:44:06.815 B17F.2    stand  255  250    100    0     80   190  
13:44:06.815 B17F.1    stand  255  40     180    79    80   224  
13:44:07.797 B17F.0    stand  255  50     80     0     80   100  
13:44:07.797 B17F.1    stand  255  30     190    71    80   111  
13:44:07.797 B17F.2    stand  255  250    100    0     80   237  
13:44:08.808 B17F.1    stand  255  20     170    73    80   240  
13:44:08.808 B17F.2    stand  255  250    100    0     80   240  
13:44:08.808 B17F.0    stand  255  40     100    0     80   210  
13:44:09.800 B17F.0    walk   255  0      80     81    80   44   
13:44:09.800 B17F.1    sit    255  10     150    0     80   70   
13:44:09.800 B17F.2    stand  255  250    100    0     80   245  
13:44:10.805 B17F.2    stand  255  250    100    0     80   0    
13:44:10.805 B17F.1    sit    255  10     140    0     80   243  
13:44:10.805 B17F.0    walk   255  -30    60     68    80   89   
13:44:11.701 B17F.2    stand  255  250    100    0     80   282  
13:44:11.701 B17F.1    sit    255  10     140    0     80   243  
13:44:11.701 B17F.0    walk   255  -70    40     0     80   128  
13:44:12.702 B17F.1    sit    255  10     140    0     80   128  
13:44:12.702 B17F.2    stand  255  250    100    0     80   243  
13:44:12.702 B17F.0    walk   255  -50    30     80    80   308  
13:44:13.709 B17F.1    sit    255  10     130    0     80   116  
13:44:13.709 B17F.2    stand  255  250    100    0     80   241  
13:44:13.709 B17F.0    walk   255  -90    30     66    80   347  
13:44:14.704 B17F.1    sit    255  10     130    0     80   141  
13:44:14.704 B17F.2    stand  255  180    30     0     80   197  
13:44:14.704 B17F.0    sit    255  -130   50     68    80   310  
13:44:15.709 B17F.1    sit    255  10     130    0     80   161  
13:44:15.709 B17F.2    stand  255  180    30     0     80   197  
13:44:15.709 B17F.0    walk   255  -160   100    81    80   347  
13:44:16.705 B17F.1    sit    255  10     130    0     80   172  
13:44:16.705 B17F.2    stand  255  180    30     0     80   197  
13:44:16.705 B17F.0    walk   255  -170   180    118   80   380  
13:44:17.716 B17F.1    sit    255  50     80     0     80   241  
13:44:17.716 B17F.2    stand  255  180    30     0     80   139  
13:44:17.716 B17F.0    walk   255  -180   280    126   80   438  
13:44:18.719 B17F.2    stand  255  180    30     0     80   438  
13:44:18.719 B17F.1    sit    255  50     80     0     80   139  
13:44:18.719 B17F.0    walk   255  -160   370    115   80   358  
13:44:19.715 B17F.2    stand  255  180    30     0     80   480  
13:44:19.715 B17F.1    sit    255  60     90     85    80   134  
13:44:19.715 B17F.0    walk   255  -160   410    0     80   388  
13:44:20.718 B17F.0    walk   255  -160   400    0     80   10   
13:44:20.718 B17F.2    stand  255  180    30     0     80   502  
13:44:20.718 B17F.1    sit    255  70     110    0     80   136  
13:44:21.731 B17F.1    sit    255  70     110    0     80   0    
13:44:21.731 B17F.2    stand  255  180    30     0     80   136  
13:44:21.731 B17F.0    stand  255  -160   400    0     80   502  
13:44:22.619 B17F.1    sit    255  70     110    0     80   370  
13:44:22.619 B17F.2    stand  255  180    30     0     80   136  
13:44:22.619 B17F.0    stand  255  -160   400    0     80   502  
13:44:23.613 B17F.2    stand  255  180    30     0     80   502  
13:44:23.613 B17F.1    sit    255  70     110    0     80   136  
13:44:23.613 B17F.0    stand  255  -160   390    0     80   362  
13:44:24.620 B17F.0    stand  2    -160   390    0     80   0    
13:44:24.620 B17F.2    stand  255  180    30     0     80   495  
13:44:24.620 B17F.1    sit    255  60     100    87    80   138  
13:44:25.669 B17F.2    stand  255  180    30     0     80   138  
13:44:25.669 B17F.1    sit    255  60     100    72    80   138  
13:44:26.628 B17F.2    stand  255  180    30     0     80   138  
13:44:26.628 B17F.1    sit    255  60     100    80    80   138  
13:44:27.627 B17F.2    stand  255  180    30     0     80   138  
13:44:27.627 B17F.1    sit    255  70     110    0     80   136  
13:44:28.628 B17F.2    stand  255  180    30     0     80   136  
13:44:28.628 B17F.1    sit    255  60     90     90    80   134  
13:44:29.639 B17F.1    sit    255  60     90     85    80   0    
13:44:29.639 B17F.2    stand  255  180    30     0     80   134  
13:44:30.644 B17F.2    stand  255  180    30     0     80   0    
13:44:30.644 B17F.1    sit    255  60     90     0     80   134  
13:44:31.634 B17F.2    stand  255  180    30     0     80   134  
13:44:31.634 B17F.1    sit    255  60     90     90    80   134  
13:44:32.630 B17F.2    stand  255  180    30     0     80   134  
13:44:32.630 B17F.1    sit    255  60     100    95    80   138  
13:44:33.531 B17F.1    sit    255  60     100    94    80   0    
13:44:33.531 B17F.2    stand  255  180    30     0     80   138  
13:44:34.533 B17F.2    stand  255  180    30     0     80   0    
13:44:34.533 B17F.1    sit    255  60     100    0     80   138  
13:44:35.540 B17F.2    stand  255  180    30     0     80   138  
13:44:35.540 B17F.1    sit    255  60     100    0     80   138  
13:44:36.536 B17F.2    stand  255  180    30     0     80   138  
13:44:36.536 B17F.1    sit    255  60     100    0     80   138  
13:44:37.537 B17F.2    stand  255  180    30     0     80   138  
13:44:37.537 B17F.1    sit    255  70     100    0     80   130  
13:44:38.588 B17F.1    sit    255  70     100    0     80   0    
13:44:38.588 B17F.2    stand  255  180    30     0     80   130  
13:44:39.538 B17F.1    sit    255  70     100    0     80   130  
13:44:39.538 B17F.2    stand  255  180    30     0     80   130  
13:44:40.545 B17F.1    sit    255  70     100    0     80   130  
13:44:40.545 B17F.2    stand  255  180    30     0     80   130  
13:44:41.540 B17F.1    sit    255  70     100    0     80   130  
13:44:41.540 B17F.2    stand  255  180    30     0     80   130  
13:44:42.550 B17F.1    sit    255  70     100    0     80   130  
13:44:42.550 B17F.2    stand  255  180    30     0     80   130  
13:44:43.545 B17F.1    sit    255  70     100    0     80   130  
13:44:43.545 B17F.2    stand  255  180    30     0     80   130  
13:44:44.455 B17F.1    sit    255  70     100    0     80   130  
13:44:44.455 B17F.2    stand  255  180    30     0     80   130  
13:44:45.441 B17F.1    sit    255  50     80     0     80   139  
13:44:45.441 B17F.2    stand  255  180    30     0     80   139  
13:44:46.444 B17F.2    stand  255  180    30     0     80   0    
13:44:46.444 B17F.1    sit    255  50     70     0     80   136  
13:44:47.456 B17F.2    stand  255  180    30     0     80   136  
13:44:47.456 B17F.1    sit    255  50     80     0     80   139  
13:44:48.442 B17F.2    stand  255  180    30     0     80   139  
13:44:48.442 B17F.1    sit    255  50     80     0     80   139  
13:44:49.444 B17F.1    sit    255  50     80     0     80   0    
13:44:49.444 B17F.2    stand  255  180    30     0     80   139  
13:44:50.385 B17F.1    sit    255  50     80     0     80   139  
13:44:50.385 B17F.2    stand  255  180    30     0     80   139  
13:44:51.388 B17F.1    sit    255  50     80     0     80   139  
13:44:51.388 B17F.2    stand  255  180    30     0     80   139  
13:44:52.388 B17F.1    sit    255  50     80     0     80   139  
13:44:52.388 B17F.2    stand  255  180    30     0     80   139  
13:44:53.395 B17F.1    sit    255  50     80     0     80   139  
13:44:53.395 B17F.2    stand  255  180    30     0     80   139  
13:44:54.392 B17F.1    sit    255  50     80     0     80   139  
13:44:54.392 B17F.2    stand  255  180    30     0     80   139  
13:44:55.398 B17F.1    sit    255  50     80     0     80   139  
13:44:55.398 B17F.2    stand  255  180    30     0     80   139  
13:44:56.394 B17F.2    stand  255  180    30     0     80   0    
13:44:56.394 B17F.1    sit    255  50     80     0     80   139  
13:44:57.399 B17F.2    stand  255  180    30     0     80   139  
13:44:57.399 B17F.1    sit    255  50     80     0     80   139  
13:44:58.413 B17F.1    sit    255  60     90     91    80   14   
13:44:58.413 B17F.2    stand  255  180    30     0     80   134  
13:44:59.636 B17F.2    stand  255  180    30     0     80   0    
13:44:59.636 B17F.1    sit    255  60     90     0     80   134  
13:45:00.401 B17F.2    stand  255  180    30     0     80   134  
13:45:00.401 B17F.1    sit    255  60     90     98    80   134  
13:45:01.397 B17F.2    stand  255  180    30     0     80   134  
13:45:01.397 B17F.1    sit    255  70     100    0     80   130  
13:45:02.290 B17F.2    stand  255  180    30     0     80   130  
13:45:02.290 B17F.1    sit    255  70     100    0     80   130  
13:45:03.296 B17F.2    stand  255  180    30     0     80   130  
13:45:03.296 B17F.1    sit    255  70     100    103   80   130  
13:45:04.292 B17F.2    stand  255  180    30     0     80   130  
13:45:04.292 B17F.1    sit    255  60     100    0     80   138  
13:45:05.295 B17F.2    stand  255  180    30     0     80   138  
13:45:05.295 B17F.1    sit    255  70     110    0     80   136  
13:45:06.359 B17F.2    stand  255  180    30     0     80   136  
13:45:06.359 B17F.1    sit    255  70     110    0     80   136  
13:45:07.359 B17F.2    stand  255  180    30     0     80   136  
13:45:07.359 B17F.1    sit    255  70     110    0     80   136  
13:45:08.244 B17F.2    stand  255  180    30     0     80   136  
13:45:08.244 B17F.1    sit    255  70     110    0     80   136  
13:45:09.249 B17F.2    stand  255  180    30     0     80   136  
13:45:09.249 B17F.1    sit    255  70     110    0     80   136  
13:45:10.250 B17F.2    stand  255  180    30     0     80   136  
13:45:10.250 B17F.1    sit    255  70     110    0     80   136  
13:45:11.265 B17F.2    stand  255  180    30     0     80   136  
13:45:11.265 B17F.1    sit    255  70     110    0     80   136  
13:45:12.248 B17F.2    stand  255  180    30     0     80   136  
13:45:12.248 B17F.1    sit    255  70     110    0     80   136  
13:45:13.248 B17F.2    stand  255  180    30     0     80   136  
13:45:13.248 B17F.1    sit    255  70     110    0     80   136  
13:45:14.248 B17F.2    stand  255  180    30     0     80   136  
13:45:14.248 B17F.1    sit    255  70     110    0     80   136  
13:45:15.249 B17F.2    stand  255  180    30     0     80   136  
13:45:15.249 B17F.1    sit    255  70     110    0     80   136  
13:45:16.252 B17F.1    sit    255  70     110    0     80   0    
13:45:16.252 B17F.2    stand  255  180    30     0     80   136  
13:45:17.253 B17F.2    stand  255  180    30     0     80   0    
13:45:17.253 B17F.1    sit    255  70     110    0     80   136  
13:45:18.253 B17F.2    stand  255  180    30     0     80   136  
13:45:18.253 B17F.1    sit    255  70     110    0     80   136  
13:45:19.257 B17F.2    stand  255  180    30     0     80   136  
13:45:19.257 B17F.1    sit    255  50     80     85    80   139  
13:45:20.147 B17F.1    sit    255  60     90     0     80   14   
13:45:20.147 B17F.2    stand  255  180    30     0     80   134  
13:45:21.159 B17F.1    sit    255  60     90     78    80   134  
13:45:21.159 B17F.2    stand  255  180    30     0     80   134  
13:45:22.165 B17F.2    stand  255  180    30     0     80   0    
13:45:22.165 B17F.1    sit    255  50     80     0     80   139  
13:45:23.165 B17F.2    stand  255  180    30     0     80   139  
13:45:23.165 B17F.1    sit    255  50     80     0     80   139  
13:45:24.174 B17F.2    stand  255  180    30     0     80   139  
13:45:24.174 B17F.1    sit    255  50     80     0     80   139  
13:45:25.168 B17F.2    stand  255  180    30     0     80   139  
13:45:25.168 B17F.1    sit    255  50     80     0     80   139  
13:45:26.178 B17F.2    stand  255  180    30     0     80   139  
13:45:26.178 B17F.1    sit    255  50     80     0     80   139  
13:45:27.173 B17F.2    stand  255  180    30     0     80   139  
13:45:27.173 B17F.1    sit    255  50     80     0     80   139  
13:45:28.172 B17F.2    stand  255  180    30     0     80   139  
13:45:28.172 B17F.1    sit    255  50     80     0     80   139  
13:45:29.177 B17F.1    sit    255  50     80     0     80   0    
13:45:29.177 B17F.2    stand  255  180    30     0     80   139  
13:45:30.068 B17F.2    stand  255  180    30     0     80   0    
13:45:30.068 B17F.1    sit    255  50     80     0     80   139  
13:45:31.070 B17F.1    sit    255  50     80     0     80   0    
13:45:31.070 B17F.2    stand  255  180    30     0     80   139  
13:45:32.069 B17F.2    stand  255  180    30     0     80   0    
13:45:32.069 B17F.1    sit    255  60     100    95    80   138  
13:45:33.071 B17F.2    stand  255  180    30     0     80   138  
13:45:33.071 B17F.1    sit    255  60     90     0     80   134  
13:45:34.078 B17F.2    stand  255  180    30     0     80   134  
13:45:34.078 B17F.1    sit    255  60     90     0     80   134  
13:45:35.074 B17F.2    stand  255  180    30     0     80   134  
13:45:35.074 B17F.1    sit    255  60     90     91    80   134  
13:45:36.076 B17F.2    stand  255  180    30     0     80   134  
13:45:36.076 B17F.1    sit    255  70     100    83    80   130  
13:45:37.074 B17F.2    stand  255  180    30     0     80   130  
13:45:37.074 B17F.1    sit    255  60     100    106   80   138  
13:45:38.010 B17F.2    stand  255  180    30     0     80   138  
13:45:38.010 B17F.1    sit    255  50     90     94    80   143  
13:45:39.011 B17F.2    stand  255  180    30     0     80   143  
13:45:39.011 B17F.1    sit    255  50     100    79    80   147  
13:45:40.013 B17F.1    sit    255  40     110    71    80   14   
13:45:40.013 B17F.2    stand  255  180    30     0     80   161  
13:45:41.009 B17F.2    stand  255  180    30     0     80   0    
13:45:41.009 B17F.1    stand  255  20     150    82    80   200  
13:45:42.014 B17F.2    stand  255  180    30     0     80   200  
13:45:42.014 B17F.1    walk   255  0      210    55    80   254  
13:45:43.017 B17F.2    stand  255  180    30     0     80   254  
13:45:43.017 B17F.1    walk   255  -50    290    63    80   347  
13:45:44.012 B17F.1    walk   255  -90    350    95    80   72   
13:45:44.012 B17F.2    stand  255  180    30     0     80   418  
13:45:45.050 B17F.1    walk   255  -100   410    84    80   472  
13:45:45.050 B17F.2    stand  255  180    30     0     80   472  
13:45:46.014 B17F.2    stand  255  180    30     0     80   0    
13:45:46.014 B17F.1    walk   255  -100   410    0     80   472  
13:45:47.022 B17F.2    stand  255  180    30     0     80   472  
13:45:47.022 B17F.1    walk   255  -60    400    75    80   441  
13:45:48.020 B17F.2    stand  255  180    30     0     80   441  
13:45:48.020 B17F.1    stand  255  -20    380    0     80   403  
13:45:49.016 B17F.2    stand  255  180    30     0     80   403  
13:45:49.016 B17F.1    stand  255  -40    370    0     80   404  
13:45:49.908 B17F.2    stand  255  180    30     0     80   404  
13:45:49.908 B17F.1    stand  255  -50    370    0     80   410  
13:45:50.909 B17F.1    stand  255  -50    370    0     80   0    
13:45:50.909 B17F.2    stand  255  180    30     0     80   410  
13:45:51.909 B17F.2    stand  255  180    30     0     80   0    
13:45:51.909 B17F.1    stand  255  -50    370    0     80   410  
13:45:52.914 B17F.2    stand  255  180    30     0     80   410  
13:45:52.914 B17F.1    stand  255  -50    370    0     80   410  
13:45:53.888 B17F.2    stand  255  180    30     0     80   410  
13:45:53.888 B17F.1    stand  2    -50    370    0     80   410  
13:45:54.937 B17F.2    stand  255  180    30     0     80   410  
13:45:55.887 B17F.2    stand  255  180    30     0     80   0    
13:45:56.900 B17F.2    stand  255  180    30     0     80   0    
13:45:57.889 B17F.2    stand  255  180    30     0     80   0    
13:45:58.941 B17F.2    stand  255  180    30     0     80   0    
13:45:59.899 B17F.2    stand  255  180    30     0     80   0    
13:46:00.898 B17F.2    stand  255  180    30     0     80   0    
13:46:01.897 B17F.2    stand  255  180    30     0     80   0    
13:46:02.898 B17F.2    stand  255  180    30     0     80   0    
13:46:03.898 B17F.2    stand  255  180    30     0     80   0    
13:46:04.787 B17F.2    stand  255  180    30     0     80   0    
13:46:05.792 B17F.2    stand  255  180    30     0     80   0    
13:46:06.789 B17F.2    stand  255  180    30     0     80   0    
13:46:07.804 B17F.2    stand  255  180    30     0     80   0    
13:46:08.789 B17F.2    stand  255  180    30     0     80   0    
13:46:09.799 B17F.2    stand  255  180    30     0     80   0    
13:46:10.796 B17F.2    stand  255  180    30     0     80   0    
13:46:11.797 B17F.2    stand  255  180    30     0     80   0    
13:46:12.799 B17F.2    stand  255  180    30     0     80   0    
13:46:13.796 B17F.2    stand  255  180    30     0     80   0    
13:46:14.860 B17F.88   88     -    -      -      -     -    -    
13:46:15.701 B17F.88   88     -    -      -      -     -    -    
13:46:16.700 B17F.88   88     -    -      -      -     -    -    
13:46:18.762 B17F.0    stand  2    -90    390    111   80   450  
13:46:19.716 B17F.0    stand  2    -90    340    89    80   50   
13:46:20.714 B17F.0    stand  2    -80    320    107   80   22   
13:46:21.729 B17F.0    stand  2    -70    320    99    80   10   
13:46:22.715 B17F.0    stand  2    -110   320    116   80   40   
13:46:23.732 B17F.0    stand  2    -150   320    130   80   40   
13:46:24.716 B17F.0    walk   2    -210   320    116   80   60   
13:46:25.728 B17F.0    walk   2    -210   330    0     80   10   
13:46:26.615 B17F.0    stand  2    -210   330    0     80   0    
13:46:27.614 B17F.0    stand  2    -210   330    0     80   0    
13:46:28.614 B17F.0    stand  2    -210   330    0     80   0    
13:46:29.614 B17F.0    stand  2    -210   330    0     80   0    
13:46:30.645 B17F.0    stand  4    -210   330    0     80   0    
13:46:31.669 B17F.88   88     -    -      -      -     -    -    
13:46:32.644 B17F.88   88     -    -      -      -     -    -    
13:46:33.629 B17F.88   88     -    -      -      -     -    -    
13:46:40.504 B17F.88   88     -    -      -      -     -    -    
13:47:12.672 B17F.88   88     -    -      -      -     -    -    
13:47:44.057 B17F.88   88     -    -      -      -     -    -    
13:48:15.944 B17F.88   88     -    -      -      -     -    -    
13:48:47.533 B17F.88   88     -    -      -      -     -    -    
13:49:19.537 B17F.88   88     -    -      -      -     -    -    
13:49:51.025 B17F.88   88     -    -      -      -     -    -    
13:50:23.125 B17F.88   88     -    -      -      -     -    -    
13:50:54.457 B17F.88   88     -    -      -      -     -    -    
13:51:26.409 B17F.88   88     -    -      -      -     -    -    
13:51:58.359 B17F.88   88     -    -      -      -     -    -    
13:52:29.747 B17F.88   88     -    -      -      -     -    -    
13:53:01.641 B17F.88   88     -    -      -      -     -    -    
13:53:33.169 B17F.88   88     -    -      -      -     -    -    
13:54:05.237 B17F.88   88     -    -      -      -     -    -    
13:54:36.666 B17F.88   88     -    -      -      -     -    -    
13:55:08.820 B17F.88   88     -    -      -      -     -    -    
13:55:40.222 B17F.88   88     -    -      -      -     -    -    
13:56:12.106 B17F.88   88     -    -      -      -     -    -    
13:56:43.712 B17F.88   88     -    -      -      -     -    -    
13:57:15.697 B17F.88   88     -    -      -      -     -    -    
13:57:47.185 B17F.88   88     -    -      -      -     -    -    
13:58:19.398 B17F.88   88     -    -      -      -     -    -    
13:58:50.621 B17F.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 613 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

## 完整原始记录（按时间排序，data_value 全文不删字段）
```
time     ms             device.tid   event          x      y      z     原始记录
13:42:00 1782762120827  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762120827, "track_count": 2, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 60}
13:42:00 1782762120827  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:42:00 1782762120863  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:00 1782762120863  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:01 1782762121850  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:01 1782762121850  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:02 1782762122813  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:02 1782762122813  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:03 1782762123816  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:03 1782762123816  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:04 1782762124813  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:04 1782762124813  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:05 1782762125816  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:05 1782762125816  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:06 1782762126714  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:06 1782762126714  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:07 1782762127708  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:07 1782762127708  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:08 1782762128712  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:08 1782762128712  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:09 1782762129718  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:09 1782762129718  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:10 1782762130712  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:10 1782762130712  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:11 1782762131712  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:11 1782762131712  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:12 1782762132713  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:12 1782762132713  B17F.1       track          -190   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:13 1782762133717  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:13 1782762133717  B17F.1       track          -200   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -200, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:14 1782762134720  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:14 1782762134720  B17F.1       track          -210   100    86    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 100, "position_z": 86, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:15 1782762135721  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:15 1782762135721  B17F.1       track          -210   110    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:16 1782762136719  B17F.1       track          -210   100    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:16 1782762136719  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:17 1782762137722  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:17 1782762137722  B17F.1       track          -190   100    98    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 100, "position_z": 98, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:18 1782762138639  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:18 1782762138639  B17F.1       track          -150   70     78    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -150, "position_y": 70, "position_z": 78, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:19 1782762139614  B17F.1       track          -140   80     82    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -140, "position_y": 80, "position_z": 82, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:19 1782762139614  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:20 1782762140625  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:20 1782762140625  B17F.1       track          -110   100    77    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -110, "position_y": 100, "position_z": 77, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:21 1782762141617  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:21 1782762141617  B17F.1       track          -60    130    74    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -60, "position_y": 130, "position_z": 74, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:22 1782762142657  B17F.1       track          0      150    71    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 0, "position_y": 150, "position_z": 71, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:22 1782762142657  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:23 1782762143626  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:23 1782762143626  B17F.1       track          20     150    92    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 20, "position_y": 150, "position_z": 92, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:24 1782762144622  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:24 1782762144622  B17F.1       track          0      130    80    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 0, "position_y": 130, "position_z": 80, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:25 1782762145638  B17F.1       track          -30    140    85    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -30, "position_y": 140, "position_z": 85, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:25 1782762145638  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:26 1782762146643  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:26 1782762146643  B17F.1       track          -90    110    77    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -90, "position_y": 110, "position_z": 77, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:27 1782762147641  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:27 1782762147641  B17F.1       track          -170   120    93    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 120, "position_z": 93, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:28 1782762148539  B17F.1       track          -200   120    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -200, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:28 1782762148539  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:29 1782762149533  B17F.1       track          -200   120    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -200, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:29 1782762149533  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:30 1782762150544  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:30 1782762150544  B17F.1       track          -200   130    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -200, "position_y": 130, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:31 1782762151536  B17F.1       track          -200   80     89    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -200, "position_y": 80, "position_z": 89, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:31 1782762151536  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:32 1782762152538  B17F.1       track          -190   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:32 1782762152538  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:33 1782762153540  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:33 1782762153540  B17F.1       track          -190   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:34 1782762154552  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:34 1782762154552  B17F.1       track          -190   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:35 1782762155541  B17F.1       track          -190   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:35 1782762155541  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:36 1782762156585  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:36 1782762156585  B17F.1       track          -150   40     76    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -150, "position_y": 40, "position_z": 76, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:37 1782762157546  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:37 1782762157546  B17F.1       track          -130   70     60    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 70, "position_z": 60, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:38 1782762158546  B17F.1       track          -130   50     99    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 99, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:38 1782762158546  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:39 1782762159543  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:39 1782762159543  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:40 1782762160442  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:40 1782762160442  B17F.1       track          -130   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:41 1782762161445  B17F.1       track          -130   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:41 1782762161445  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:42 1782762162452  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:42 1782762162452  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:43 1782762163455  B17F.2       track          -60    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:43 1782762163455  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:44 1782762164455  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:44 1782762164455  B17F.2       track          -50    20     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:45 1782762165454  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:45 1782762165454  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:46 1782762166460  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:46 1782762166460  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:47 1782762167456  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:47 1782762167456  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:48 1782762168466  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:48 1782762168466  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:49 1782762169460  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:49 1782762169460  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:50 1782762170363  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:50 1782762170363  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:51 1782762171358  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:51 1782762171358  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:52 1782762172360  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:52 1782762172360  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:53 1782762173357  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:53 1782762173357  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:54 1782762174362  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:54 1782762174362  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:55 1782762175362  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:55 1782762175362  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:56 1782762176362  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:56 1782762176362  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:57 1782762177362  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:57 1782762177362  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:58 1782762178303  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:58 1782762178303  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:59 1782762179294  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:42:59 1782762179294  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:00 1782762180317  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762180317, "track_count": 2, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 60}
13:43:00 1782762180317  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:43:00 1782762180354  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:00 1782762180354  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:01 1782762181295  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:01 1782762181295  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:02 1782762182301  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:02 1782762182301  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:03 1782762183301  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:03 1782762183301  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:04 1782762184297  B17F.2       track          -50    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:04 1782762184297  B17F.1       track          -130   50     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:05 1782762185303  B17F.1       track          -110   60     57    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -110, "position_y": 60, "position_z": 57, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:05 1782762185303  B17F.2       track          -60    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:06 1782762186304  B17F.2       track          -60    20     83    {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -60, "position_y": 20, "position_z": 83, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:06 1782762186304  B17F.1       track          -110   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -110, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:07 1782762187306  B17F.1       track          -110   70     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -110, "position_y": 70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:07 1782762187306  B17F.2       track          -50    40     94    {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -50, "position_y": 40, "position_z": 94, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:08 1782762188305  B17F.2       track          -30    50     97    {"pose": 4, "area_id": 255, "track_id": 2, "position_x": -30, "position_y": 50, "position_z": 97, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:08 1782762188305  B17F.1       track          -110   80     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -110, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:09 1782762189303  B17F.1       track          -110   100    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -110, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:09 1782762189303  B17F.2       track          0      80     75    {"pose": 1, "area_id": 255, "track_id": 2, "position_x": 0, "position_y": 80, "position_z": 75, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:10 1782762190201  B17F.2       track          10     120    67    {"pose": 1, "area_id": 255, "track_id": 2, "position_x": 10, "position_y": 120, "position_z": 67, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:10 1782762190201  B17F.1       track          -130   90     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 90, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:11 1782762191206  B17F.1       track          -120   90     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -120, "position_y": 90, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:11 1782762191206  B17F.2       track          50     90     67    {"pose": 1, "area_id": 255, "track_id": 2, "position_x": 50, "position_y": 90, "position_z": 67, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:12 1782762192207  B17F.1       track          -120   80     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -120, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:12 1782762192207  B17F.2       track          50     100    71    {"pose": 1, "area_id": 255, "track_id": 2, "position_x": 50, "position_y": 100, "position_z": 71, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:13 1782762193205  B17F.2       track          50     90     87    {"pose": 1, "area_id": 255, "track_id": 2, "position_x": 50, "position_y": 90, "position_z": 87, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:13 1782762193205  B17F.1       track          -120   80     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -120, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:14 1782762194175  B17F.1       track          -120   80     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -120, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:14 1782762194175  B17F.2       track          50     80     78    {"pose": 3, "area_id": 255, "track_id": 2, "position_x": 50, "position_y": 80, "position_z": 78, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:15 1782762195165  B17F.2       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 2, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:15 1782762195165  B17F.1       track          -120   80     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -120, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:16 1782762196173  B17F.1       track          -120   80     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -120, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:16 1782762196173  B17F.2       track          60     90     84    {"pose": 3, "area_id": 255, "track_id": 2, "position_x": 60, "position_y": 90, "position_z": 84, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:17 1782762197169  B17F.2       track          150    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 150, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:17 1782762197169  B17F.1       track          -130   60     60    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 60, "position_z": 60, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:18 1782762198179  B17F.1       track          -170   30     69    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 30, "position_z": 69, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:18 1782762198179  B17F.2       track          180    40     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:19 1782762199171  B17F.2       track          180    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:19 1782762199171  B17F.1       track          -160   100    89    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -160, "position_y": 100, "position_z": 89, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:20 1782762200174  B17F.1       track          -130   150    127   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 150, "position_z": 127, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:20 1782762200174  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:21 1782762201170  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:21 1782762201170  B17F.1       track          -150   170    104   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -150, "position_y": 170, "position_z": 104, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:22 1782762202175  B17F.1       track          -170   220    127   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 220, "position_z": 127, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:22 1782762202175  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:23 1782762203190  B17F.1       track          -190   210    125   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 210, "position_z": 125, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:23 1782762203190  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:24 1782762204178  B17F.1       track          -170   170    113   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 170, "position_z": 113, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:24 1782762204178  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:43:25 1782762205195  B17F.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782762205195, "event_status": "start", "number_people": 3, "respiratory_rate": -1}
13:43:25 1782762205242  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:25 1782762205242  B17F.0       track          60     100    89    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 89, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:25 1782762205242  B17F.1       track          -170   230    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 230, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:26 1782762206070  B17F.1       track          -150   270    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -150, "position_y": 270, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:26 1782762206070  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:26 1782762206070  B17F.0       track          60     110    82    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 110, "position_z": 82, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:27 1782762207072  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:27 1782762207072  B17F.1       track          -100   290    111   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -100, "position_y": 290, "position_z": 111, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:27 1782762207072  B17F.0       track          60     110    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 110, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:28 1782762208078  B17F.1       track          -130   290    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -130, "position_y": 290, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:28 1782762208078  B17F.0       track          60     110    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 110, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:28 1782762208078  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:29 1782762209073  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:29 1782762209073  B17F.0       track          60     110    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 110, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:29 1782762209073  B17F.1       track          -170   240    115   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 240, "position_z": 115, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:30 1782762210088  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:30 1782762210088  B17F.0       track          50     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 50, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:30 1782762210088  B17F.1       track          -180   160    102   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -180, "position_y": 160, "position_z": 102, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:31 1782762211086  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:31 1782762211086  B17F.0       track          60     90     89    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 89, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:31 1782762211086  B17F.1       track          -170   140    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 140, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:32 1782762212085  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:32 1782762212085  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:32 1782762212085  B17F.1       track          -160   140    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -160, "position_y": 140, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:33 1782762213090  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:33 1782762213090  B17F.0       track          60     90     94    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 94, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:33 1782762213090  B17F.1       track          -170   140    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 140, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:34 1782762214084  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:34 1782762214084  B17F.0       track          60     100    98    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 98, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:34 1782762214084  B17F.1       track          -170   140    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 140, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:35 1782762215084  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:35 1782762215084  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:35 1782762215084  B17F.1       track          -170   130    88    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 130, "position_z": 88, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:36 1782762216115  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:36 1782762216115  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:36 1782762216115  B17F.1       track          -180   140    99    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -180, "position_y": 140, "position_z": 99, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:36 1782762216988  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:36 1782762216988  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:36 1782762216988  B17F.1       track          -210   110    86    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 110, "position_z": 86, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:37 1782762217985  B17F.1       track          -210   120    80    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 120, "position_z": 80, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:37 1782762217985  B17F.2       track          170    50     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 170, "position_y": 50, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:37 1782762217985  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:38 1782762218986  B17F.1       track          -210   100    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:38 1782762218986  B17F.2       track          190    60     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 190, "position_y": 60, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:38 1782762218986  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:39 1782762219984  B17F.2       track          220    90     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 220, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:39 1782762219984  B17F.1       track          -210   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 120, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:39 1782762219984  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:40 1782762220996  B17F.1       track          -210   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 120, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:40 1782762220996  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:40 1782762220996  B17F.2       track          220    90     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 220, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:42 1782762222000  B17F.2       track          220    90     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 220, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:42 1782762222000  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:42 1782762222000  B17F.1       track          -210   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 120, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:42 1782762222987  B17F.2       track          210    90     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 210, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:42 1782762222987  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:42 1782762222987  B17F.1       track          -210   100    104   {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 100, "position_z": 104, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:43 1782762223994  B17F.2       track          210    80     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 210, "position_y": 80, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:43 1782762223994  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:43 1782762223994  B17F.1       track          -210   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 120, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:44 1782762224992  B17F.2       track          210    80     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 210, "position_y": 80, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:44 1782762224992  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:44 1782762224992  B17F.1       track          -210   100    88    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -210, "position_y": 100, "position_z": 88, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:45 1782762225991  B17F.2       track          210    90     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 210, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:45 1782762225991  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:45 1782762225991  B17F.1       track          -200   60     93    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -200, "position_y": 60, "position_z": 93, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:46 1782762226993  B17F.2       track          190    110    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 190, "position_y": 110, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:46 1782762226993  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:46 1782762226993  B17F.1       track          -200   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -200, "position_y": 60, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:47 1782762227995  B17F.2       track          250    90     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:47 1782762227995  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:47 1782762227995  B17F.1       track          -190   60     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -190, "position_y": 60, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:48 1782762228892  B17F.2       track          260    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 260, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:48 1782762228892  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:48 1782762228892  B17F.1       track          -170   80     68    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 80, "position_z": 68, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:49 1782762229886  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:49 1782762229886  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:49 1782762229886  B17F.1       track          -160   170    115   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -160, "position_y": 170, "position_z": 115, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:50 1782762230890  B17F.0       track          60     110    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 110, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:50 1782762230890  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:50 1782762230890  B17F.1       track          -170   240    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 240, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:51 1782762231913  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:51 1782762231913  B17F.1       track          -170   310    119   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -170, "position_y": 310, "position_z": 119, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:51 1782762231913  B17F.0       track          60     90     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:52 1782762232890  B17F.1       track          -120   340    115   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -120, "position_y": 340, "position_z": 115, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:52 1782762232890  B17F.0       track          50     80     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:52 1782762232890  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:53 1782762233897  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:53 1782762233897  B17F.0       track          50     80     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:53 1782762233897  B17F.1       track          -60    300    91    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -60, "position_y": 300, "position_z": 91, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:54 1782762234894  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:54 1782762234894  B17F.0       track          50     80     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:54 1782762234894  B17F.1       track          0      250    78    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 0, "position_y": 250, "position_z": 78, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:55 1782762235898  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:55 1782762235898  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:55 1782762235898  B17F.1       track          20     200    76    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 20, "position_y": 200, "position_z": 76, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:56 1782762236927  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:56 1782762236927  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:56 1782762236927  B17F.1       track          30     190    66    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 30, "position_y": 190, "position_z": 66, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:57 1782762237903  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:57 1782762237903  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:57 1782762237903  B17F.1       track          40     190    101   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 190, "position_z": 101, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:58 1782762238896  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:58 1782762238896  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:58 1782762238896  B17F.1       track          50     170    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 170, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:43:59 1782762239911  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762239911, "track_count": 3, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 60}
13:43:59 1782762239911  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:44:00 1782762240045  B17F.1       track          50     160    115   {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 160, "position_z": 115, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:00 1782762240045  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:00 1782762240045  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:00 1782762240811  B17F.1       track          50     160    99    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 160, "position_z": 99, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:00 1782762240811  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:00 1782762240811  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:01 1782762241797  B17F.1       track          50     170    107   {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 170, "position_z": 107, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:01 1782762241797  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:01 1782762241797  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:02 1782762242793  B17F.1       track          50     180    86    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 180, "position_z": 86, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:02 1782762242793  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:02 1782762242793  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:03 1782762243796  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:03 1782762243796  B17F.1       track          50     190    90    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 190, "position_z": 90, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:03 1782762243796  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:04 1782762244804  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:04 1782762244804  B17F.1       track          50     180    99    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 180, "position_z": 99, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:04 1782762244804  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:05 1782762245796  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:05 1782762245796  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:05 1782762245796  B17F.1       track          40     180    84    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 180, "position_z": 84, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:06 1782762246815  B17F.0       track          60     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:06 1782762246815  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:06 1782762246815  B17F.1       track          40     180    79    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 180, "position_z": 79, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:07 1782762247797  B17F.0       track          50     80     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:07 1782762247797  B17F.1       track          30     190    71    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 30, "position_y": 190, "position_z": 71, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:07 1782762247797  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:08 1782762248808  B17F.1       track          20     170    73    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 20, "position_y": 170, "position_z": 73, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:08 1782762248808  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:08 1782762248808  B17F.0       track          40     100    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:09 1782762249800  B17F.0       track          0      80     81    {"pose": 1, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": 80, "position_z": 81, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:09 1782762249800  B17F.1       track          10     150    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 10, "position_y": 150, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:09 1782762249800  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:10 1782762250805  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:10 1782762250805  B17F.1       track          10     140    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 10, "position_y": 140, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:10 1782762250805  B17F.0       track          -30    60     68    {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -30, "position_y": 60, "position_z": 68, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:11 1782762251701  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:11 1782762251701  B17F.1       track          10     140    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 10, "position_y": 140, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:11 1782762251701  B17F.0       track          -70    40     0     {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -70, "position_y": 40, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:12 1782762252702  B17F.1       track          10     140    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 10, "position_y": 140, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:12 1782762252702  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:12 1782762252702  B17F.0       track          -50    30     80    {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -50, "position_y": 30, "position_z": 80, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:13 1782762253709  B17F.1       track          10     130    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 10, "position_y": 130, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:13 1782762253709  B17F.2       track          250    100    0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 250, "position_y": 100, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:13 1782762253709  B17F.0       track          -90    30     66    {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -90, "position_y": 30, "position_z": 66, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:14 1782762254704  B17F.1       track          10     130    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 10, "position_y": 130, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:14 1782762254704  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:14 1782762254704  B17F.0       track          -130   50     68    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": -130, "position_y": 50, "position_z": 68, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:15 1782762255709  B17F.1       track          10     130    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 10, "position_y": 130, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:15 1782762255709  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:15 1782762255709  B17F.0       track          -160   100    81    {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 100, "position_z": 81, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:16 1782762256705  B17F.1       track          10     130    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 10, "position_y": 130, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:16 1782762256705  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:16 1782762256705  B17F.0       track          -170   180    118   {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 180, "position_z": 118, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:17 1782762257716  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:17 1782762257716  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:17 1782762257716  B17F.0       track          -180   280    126   {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 126, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:18 1782762258719  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:18 1782762258719  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:18 1782762258719  B17F.0       track          -160   370    115   {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 370, "position_z": 115, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:19 1782762259715  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:19 1782762259715  B17F.1       track          60     90     85    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 85, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:19 1782762259715  B17F.0       track          -160   410    0     {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 410, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:20 1782762260718  B17F.0       track          -160   400    0     {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 400, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:20 1782762260718  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:20 1782762260718  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:21 1782762261731  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:21 1782762261731  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:21 1782762261731  B17F.0       track          -160   400    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 400, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:22 1782762262619  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:22 1782762262619  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:22 1782762262619  B17F.0       track          -160   400    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 400, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:23 1782762263613  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:23 1782762263613  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:23 1782762263613  B17F.0       track          -160   390    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 390, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:24 1782762264620  B17F.0       track          -160   390    0     {"pose": 4, "event": 2, "area_id": 2, "track_id": 0, "position_x": -160, "position_y": 390, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:24 1782762264620  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:24 1782762264620  B17F.1       track          60     100    87    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 87, "track_count": 3, "remaining_time": 0, "track_confidence": 80}
13:44:24 1782762264658  B17F         ExitRoom       -      -      -     {"heart_rate": -1, "event_since": 1782762264658, "event_status": "start", "respiratory_rate": -1}
13:44:25 1782762265632  B17F.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782762265632, "event_status": "start", "number_people": 2, "respiratory_rate": -1}
13:44:25 1782762265669  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:25 1782762265669  B17F.1       track          60     100    72    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 72, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:26 1782762266628  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:26 1782762266628  B17F.1       track          60     100    80    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 80, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:27 1782762267627  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:27 1782762267627  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:28 1782762268628  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:28 1782762268628  B17F.1       track          60     90     90    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 90, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:29 1782762269639  B17F.1       track          60     90     85    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 85, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:29 1782762269639  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:30 1782762270644  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:30 1782762270644  B17F.1       track          60     90     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:31 1782762271634  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:31 1782762271634  B17F.1       track          60     90     90    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 90, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:32 1782762272630  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:32 1782762272630  B17F.1       track          60     100    95    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 95, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:33 1782762273531  B17F.1       track          60     100    94    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 94, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:33 1782762273531  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:34 1782762274533  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:34 1782762274533  B17F.1       track          60     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:35 1782762275540  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:35 1782762275540  B17F.1       track          60     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:36 1782762276536  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:36 1782762276536  B17F.1       track          60     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:37 1782762277537  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:37 1782762277537  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:38 1782762278588  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:38 1782762278588  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:39 1782762279538  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:39 1782762279538  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:40 1782762280545  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:40 1782762280545  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:41 1782762281540  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:41 1782762281540  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:42 1782762282550  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:42 1782762282550  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:43 1782762283545  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:43 1782762283545  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:44 1782762284455  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:44 1782762284455  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:45 1782762285441  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:45 1782762285441  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:46 1782762286444  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:46 1782762286444  B17F.1       track          50     70     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:47 1782762287456  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:47 1782762287456  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:48 1782762288442  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:48 1782762288442  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:49 1782762289444  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:49 1782762289444  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:50 1782762290385  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:50 1782762290385  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:51 1782762291388  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:51 1782762291388  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:52 1782762292388  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:52 1782762292388  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:53 1782762293395  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:53 1782762293395  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:54 1782762294392  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:54 1782762294392  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:55 1782762295398  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:55 1782762295398  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:56 1782762296394  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:56 1782762296394  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:57 1782762297399  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:57 1782762297399  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:58 1782762298413  B17F.1       track          60     90     91    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 91, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:58 1782762298413  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:59 1782762299408  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762299408, "track_count": 2, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 60}
13:44:59 1782762299408  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:44:59 1782762299636  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:44:59 1782762299636  B17F.1       track          60     90     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:00 1782762300401  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:00 1782762300401  B17F.1       track          60     90     98    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 98, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:01 1782762301397  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:01 1782762301397  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:02 1782762302290  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:02 1782762302290  B17F.1       track          70     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:03 1782762303296  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:03 1782762303296  B17F.1       track          70     100    103   {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 103, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:04 1782762304292  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:04 1782762304292  B17F.1       track          60     100    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:05 1782762305295  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:05 1782762305295  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:06 1782762306359  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:06 1782762306359  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:07 1782762307359  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:07 1782762307359  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:08 1782762308244  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:08 1782762308244  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:09 1782762309249  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:09 1782762309249  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:10 1782762310250  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:10 1782762310250  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:11 1782762311265  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:11 1782762311265  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:12 1782762312248  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:12 1782762312248  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:13 1782762313248  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:13 1782762313248  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:14 1782762314248  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:14 1782762314248  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:15 1782762315249  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:15 1782762315249  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:16 1782762316252  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:16 1782762316252  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:17 1782762317253  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:17 1782762317253  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:18 1782762318253  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:18 1782762318253  B17F.1       track          70     110    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 110, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:19 1782762319257  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:19 1782762319257  B17F.1       track          50     80     85    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 85, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:20 1782762320147  B17F.1       track          60     90     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:20 1782762320147  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:21 1782762321159  B17F.1       track          60     90     78    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 78, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:21 1782762321159  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:22 1782762322165  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:22 1782762322165  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:23 1782762323165  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:23 1782762323165  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:24 1782762324174  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:24 1782762324174  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:25 1782762325168  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:25 1782762325168  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:26 1782762326178  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:26 1782762326178  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:27 1782762327173  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:27 1782762327173  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:28 1782762328172  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:28 1782762328172  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:29 1782762329177  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:29 1782762329177  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:30 1782762330068  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:30 1782762330068  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:31 1782762331070  B17F.1       track          50     80     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:31 1782762331070  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:32 1782762332069  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:32 1782762332069  B17F.1       track          60     100    95    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 95, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:33 1782762333071  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:33 1782762333071  B17F.1       track          60     90     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:34 1782762334078  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:34 1782762334078  B17F.1       track          60     90     0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:35 1782762335074  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:35 1782762335074  B17F.1       track          60     90     91    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 90, "position_z": 91, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:36 1782762336076  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:36 1782762336076  B17F.1       track          70     100    83    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 70, "position_y": 100, "position_z": 83, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:37 1782762337074  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:37 1782762337074  B17F.1       track          60     100    106   {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": 100, "position_z": 106, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:38 1782762338010  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:38 1782762338010  B17F.1       track          50     90     94    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 90, "position_z": 94, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:39 1782762339011  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:39 1782762339011  B17F.1       track          50     100    79    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 100, "position_z": 79, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:40 1782762340013  B17F.1       track          40     110    71    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 110, "position_z": 71, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:40 1782762340013  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:41 1782762341009  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:41 1782762341009  B17F.1       track          20     150    82    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 20, "position_y": 150, "position_z": 82, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:42 1782762342014  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:42 1782762342014  B17F.1       track          0      210    55    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 0, "position_y": 210, "position_z": 55, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:43 1782762343017  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:43 1782762343017  B17F.1       track          -50    290    63    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -50, "position_y": 290, "position_z": 63, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:44 1782762344012  B17F.1       track          -90    350    95    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -90, "position_y": 350, "position_z": 95, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:44 1782762344012  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:45 1782762345050  B17F.1       track          -100   410    84    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -100, "position_y": 410, "position_z": 84, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:45 1782762345050  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:46 1782762346014  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:46 1782762346014  B17F.1       track          -100   410    0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -100, "position_y": 410, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:47 1782762347022  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:47 1782762347022  B17F.1       track          -60    400    75    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -60, "position_y": 400, "position_z": 75, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:48 1782762348020  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:48 1782762348020  B17F.1       track          -20    380    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -20, "position_y": 380, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:49 1782762349016  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:49 1782762349016  B17F.1       track          -40    370    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -40, "position_y": 370, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:49 1782762349908  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:49 1782762349908  B17F.1       track          -50    370    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -50, "position_y": 370, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:50 1782762350909  B17F.1       track          -50    370    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -50, "position_y": 370, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:50 1782762350909  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:51 1782762351909  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:51 1782762351909  B17F.1       track          -50    370    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -50, "position_y": 370, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:52 1782762352914  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:52 1782762352914  B17F.1       track          -50    370    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -50, "position_y": 370, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:53 1782762353888  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:53 1782762353888  B17F.1       track          -50    370    0     {"pose": 4, "event": 2, "area_id": 2, "track_id": 1, "position_x": -50, "position_y": 370, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
13:45:53 1782762353922  B17F.1       ExitRoom       -      -      -     {"track_id": 1, "heart_rate": -1, "event_since": 1782762353922, "event_status": "start", "respiratory_rate": -1}
13:45:54 1782762354892  B17F.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782762354892, "event_status": "start", "number_people": 1, "respiratory_rate": -1}
13:45:54 1782762354937  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:45:55 1782762355887  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:45:56 1782762356900  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:45:57 1782762357889  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:45:58 1782762358904  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:45:58 1782762358904  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762358904, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 4, "respiratory_rate": -1, "multi_person_duration": 56}
13:45:58 1782762358941  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:45:59 1782762359899  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:00 1782762360898  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:01 1782762361897  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:02 1782762362898  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:03 1782762363898  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:04 1782762364787  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:05 1782762365792  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:06 1782762366789  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:07 1782762367804  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:08 1782762368789  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:09 1782762369799  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:10 1782762370796  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:11 1782762371797  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:12 1782762372799  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:13 1782762373796  B17F.2       track          180    30     0     {"pose": 4, "area_id": 255, "track_id": 2, "position_x": 180, "position_y": 30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:14 1782762374820  B17F.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782762374820, "event_status": "start", "number_people": 0, "respiratory_rate": -1}
13:46:14 1782762374860  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:15 1782762375701  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:16 1782762376700  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:18 1782762378722  B17F.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782762378722, "event_status": "start", "number_people": 1, "respiratory_rate": -1}
13:46:18 1782762378760  B17F         EnterRoom      -      -      -     {"heart_rate": -1, "event_since": 1782762378760, "event_status": "start", "respiratory_rate": -1}
13:46:18 1782762378762  B17F.0       track          -90    390    111   {"pose": 4, "event": 1, "area_id": 2, "track_id": 0, "position_x": -90, "position_y": 390, "position_z": 111, "remaining_time": 0, "track_confidence": 80}
13:46:19 1782762379716  B17F.0       track          -90    340    89    {"pose": 4, "area_id": 2, "track_id": 0, "position_x": -90, "position_y": 340, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
13:46:20 1782762380714  B17F.0       track          -80    320    107   {"pose": 4, "area_id": 2, "track_id": 0, "position_x": -80, "position_y": 320, "position_z": 107, "remaining_time": 0, "track_confidence": 80}
13:46:21 1782762381729  B17F.0       track          -70    320    99    {"pose": 4, "area_id": 2, "track_id": 0, "position_x": -70, "position_y": 320, "position_z": 99, "remaining_time": 0, "track_confidence": 80}
13:46:22 1782762382715  B17F.0       track          -110   320    116   {"pose": 4, "area_id": 2, "track_id": 0, "position_x": -110, "position_y": 320, "position_z": 116, "remaining_time": 0, "track_confidence": 80}
13:46:23 1782762383732  B17F.0       track          -150   320    130   {"pose": 4, "area_id": 2, "track_id": 0, "position_x": -150, "position_y": 320, "position_z": 130, "remaining_time": 0, "track_confidence": 80}
13:46:24 1782762384716  B17F.0       track          -210   320    116   {"pose": 1, "area_id": 2, "track_id": 0, "position_x": -210, "position_y": 320, "position_z": 116, "remaining_time": 0, "track_confidence": 80}
13:46:25 1782762385728  B17F.0       track          -210   330    0     {"pose": 1, "area_id": 2, "track_id": 0, "position_x": -210, "position_y": 330, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:26 1782762386615  B17F.0       track          -210   330    0     {"pose": 4, "area_id": 2, "track_id": 0, "position_x": -210, "position_y": 330, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:27 1782762387614  B17F.0       track          -210   330    0     {"pose": 4, "area_id": 2, "track_id": 0, "position_x": -210, "position_y": 330, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:28 1782762388614  B17F.0       track          -210   330    0     {"pose": 4, "area_id": 2, "track_id": 0, "position_x": -210, "position_y": 330, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:29 1782762389614  B17F.0       track          -210   330    0     {"pose": 4, "area_id": 2, "track_id": 0, "position_x": -210, "position_y": 330, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:30 1782762390645  B17F.0       track          -210   330    0     {"pose": 4, "event": 2, "area_id": 4, "track_id": 0, "position_x": -210, "position_y": 330, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:30 1782762390889  B17F         ExitRoom       -      -      -     {"heart_rate": -1, "event_since": 1782762390889, "event_status": "start", "respiratory_rate": -1}
13:46:31 1782762391632  B17F.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782762391632, "event_status": "start", "number_people": 0, "respiratory_rate": -1}
13:46:31 1782762391669  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:32 1782762392644  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:33 1782762393629  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:46:40 1782762400504  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:47:12 1782762432262  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762432262, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 2, "stand_duration": 27, "respiratory_rate": -1, "multi_person_duration": 0}
13:47:12 1782762432262  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:47:12 1782762432672  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:47:44 1782762464057  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:48:15 1782762495748  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762495748, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:48:15 1782762495748  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:48:15 1782762495944  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:48:47 1782762527533  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:49:19 1782762559232  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762559232, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:49:19 1782762559232  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:49:19 1782762559537  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:49:51 1782762591025  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:50:22 1782762622726  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:50:22 1782762622726  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762622726, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:50:23 1782762623125  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:50:54 1782762654457  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:51:26 1782762686268  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:51:26 1782762686268  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762686268, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:51:26 1782762686409  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:51:57 1782762717957  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762717957, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:51:57 1782762717957  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:51:58 1782762718359  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:52:29 1782762749747  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:53:01 1782762781449  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762781449, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:53:01 1782762781449  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:53:01 1782762781641  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:53:33 1782762813169  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:54:05 1782762845028  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:54:05 1782762845028  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762845028, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:54:05 1782762845237  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:54:36 1782762876666  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:55:08 1782762908415  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762908415, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:55:08 1782762908415  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:55:08 1782762908820  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:55:40 1782762940222  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:56:11 1782762971903  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:56:11 1782762971903  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782762971903, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:56:12 1782762972106  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:56:43 1782763003712  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:57:15 1782763035395  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:57:15 1782763035395  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782763035395, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:57:15 1782763035697  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:57:47 1782763067185  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:58:18 1782763098881  B17F.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782763098881, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
13:58:18 1782763098881  B17F.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
13:58:19 1782763099398  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
13:58:50 1782763130621  B17F.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
```

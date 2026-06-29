# case-cd2b-0627-04270432 — 每 tick belief 时间线 (room fd00:0:3:112:3:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
04:27:00 CD2B.0   CD2B03146333  stand   119  NoReport stand              room -    OpenFloor  1   0     0.00  0.03  0.76  0.01  0.03  0.02
04:27:01 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.86  0.01  0.02  0.01
04:27:02 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:03 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.03  0.81  0.00  0.02  0.02
04:27:04 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
04:27:05 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
04:27:06 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.01  0.92  0.00  0.01  0.01
04:27:07 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.03  0.82  0.00  0.02  0.02
04:27:08 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
04:27:09 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.01  0.91  0.00  0.01  0.01
04:27:10 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.03  0.82  0.00  0.02  0.02
04:27:11 333B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   0     0.00  0.03  0.82  0.00  0.02  0.02
04:27:11 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.01  0.90  0.00  0.01  0.01
04:27:12 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.86  0.00  0.01  0.02
04:27:13 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:14 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:15 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:16 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:17 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.03  0.81  0.00  0.02  0.02
04:27:18 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
04:27:19 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.92  0.00  0.01  0.01
04:27:20 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.96  0.00  0.00  0.01
04:27:21 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
04:27:22 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.87  0.00  0.00  0.02
04:27:23 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.86  0.00  0.02  0.01
04:27:24 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:25 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:26 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.92  0.00  0.01  0.01
04:27:27 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.96  0.00  0.00  0.01
04:27:28 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
04:27:29 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.96  0.00  0.00  0.01
04:27:30 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:31 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
04:27:32 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:33 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:34 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
04:27:35 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:36 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:37 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:38 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:39 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:40 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:41 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
04:27:42 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
04:27:43 333B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
04:27:43 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
04:27:44 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.93  0.00  0.00  0.01
04:27:45 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.87  0.00  0.00  0.02
04:27:46 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
04:27:47 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.88  0.00  0.00  0.02
04:27:48 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.86  0.00  0.01  0.02
04:27:49 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:50 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:51 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:52 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:53 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.03  0.81  0.00  0.02  0.02
04:27:54 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
04:27:55 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
04:27:56 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:57 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:58 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:27:59 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:28:00 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:28:01 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.32  0.00  0.31  0.30
04:28:02 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.87  0.00  0.01  0.02
04:28:03 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:28:04 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
04:28:05 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.03  0.81  0.00  0.02  0.02
04:28:06 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.89  0.00  0.01  0.01
04:28:07 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.03  0.59  0.00  0.01  0.25
04:28:08 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.03  0.40  0.00  0.11  0.34
04:28:09 CD2B.0   CD2B03146333  stand   144  NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.58  0.00  0.08  0.24
04:28:10 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.73  0.00  0.05  0.15
04:28:11 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.84  0.00  0.03  0.09
04:28:12 CD2B.0   CD2B03146333  stand   139  NoReport stand              room -    OpenFloor  2   0     0.00  0.00  0.92  0.00  0.02  0.05
04:28:13 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  0.96  0.00  0.01  0.03
04:28:14 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  0.98  0.00  0.00  0.01
04:28:14 333B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  2   0     0.00  0.00  0.98  0.00  0.00  0.01
04:28:15 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  0.99  0.00  0.00  0.01
04:28:16 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  0.99  0.00  0.00  0.00
04:28:17 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:18 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:19 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:20 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:21 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:22 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:23 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:24 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:25 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:26 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:27 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:28 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:29 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:30 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:31 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:32 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:33 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.01  0.04  0.50  0.01  0.03  0.26
04:28:34 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.03  0.66  0.00  0.02  0.17
04:28:35 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.80  0.00  0.01  0.10
04:28:36 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.89  0.00  0.01  0.06
04:28:37 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  0.94  0.00  0.00  0.03
04:28:38 CD2B.E   -             -       0    NoReport np=2               room -    OpenFloor  2   0     0.00  0.00  0.94  0.00  0.00  0.03
04:28:38 CD2B.E1  CD2B12838096  -       0    NoReport EnterRoom(rdr)     trk  1.00 OpenFloor  2   0     0.00  0.00  0.94  0.00  0.00  0.03
04:28:38 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  0.98  0.00  0.00  0.01
04:28:38 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.00  0.99  0.00  0.00  0.00
04:28:39 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:39 CD2B.1   CD2B12838096  stand   86   NoReport stand              trk  0.44 OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:40 CD2B.1   CD2B12838096  walk    84   NoReport walk               trk  0.86 OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:40 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:41 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:41 CD2B.1   CD2B12838096  walk    83   NoReport walk               trk  0.83 OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:42 CD2B.1   CD2B12838096  walk    72   NoReport walk               trk  0.90 OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:42 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:43 CD2B.1   CD2B12838096  walk    88   NoReport walk               trk  0.92 OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:43 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.00  1.00  0.00  0.00  0.00
04:28:44 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.03  0.67  0.00  0.02  0.17
04:28:44 CD2B.1   CD2B12838096  walk    70   NoReport walk               trk  0.92 OpenFloor  2   0     0.00  0.03  0.67  0.00  0.02  0.17
04:28:45 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.83  0.00  0.04  0.03
04:28:45 CD2B.1   CD2B12838096  walk    83   NoReport walk               trk  0.92 OpenFloor  2   0     0.00  0.02  0.83  0.00  0.04  0.03
04:28:46 CD2B.1   CD2B12838096  walk    0    NoReport walk               trk  0.92 OpenFloor  2   0     0.00  0.01  0.95  0.00  0.01  0.01
04:28:46 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.95  0.00  0.01  0.01
04:28:46 333B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  2   0     0.00  0.01  0.95  0.00  0.01  0.01
04:28:47 CD2B.1   CD2B12838096  walk    71   NoReport walk               trk  0.92 OpenFloor  2   0     0.00  0.02  0.88  0.00  0.01  0.02
04:28:47 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.88  0.00  0.01  0.02
04:28:47 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.91  0.00  0.01  0.01
04:28:47 CD2B.1   CD2B12838096  walk    94   NoReport walk               trk  0.92 OpenFloor  2   0     0.00  0.02  0.46  0.00  0.13  0.29
04:28:48 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.51  0.00  0.14  0.26
04:28:48 CD2B.1   CD2B12838096  walk    80   NoReport walk               trk  0.92 OpenFloor  2   0     0.00  0.02  0.51  0.00  0.14  0.26
04:28:49 CD2B.1   CD2B12838096  walk    0    NoReport walk               trk  0.92 OpenFloor  2   0     0.00  0.01  0.70  0.00  0.09  0.16
04:28:49 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
04:28:50 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.65  0.00  0.07  0.21
04:28:50 CD2B.1   CD2B12838096  walk    0    NoReport walk               trk  0.92 OpenFloor  2   0     0.00  0.02  0.65  0.00  0.07  0.21
04:28:51 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  0.92 OpenFloor  2   0     0.00  0.02  0.61  0.00  0.08  0.22
04:28:51 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.61  0.00  0.08  0.22
04:28:52 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  0.92 OpenFloor  2   0     0.00  0.01  0.86  0.00  0.03  0.08
04:28:52 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.86  0.00  0.03  0.08
04:28:53 CD2B.E   -             -       0    NoReport np=3               room -    OpenFloor  2   0     0.00  0.01  0.86  0.00  0.03  0.08
04:28:53 CD2B.2   CD2B22853978  stand   66   NoReport stand              trk  0.50 OpenFloor  3   0     0.00  0.02  0.41  0.00  0.54  0.03
04:28:53 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  3   0     0.00  0.02  0.78  0.00  0.03  0.10
04:28:53 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.02  0.78  0.00  0.03  0.10
04:28:54 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  3   0     0.00  0.01  0.78  0.00  0.17  0.01
04:28:54 CD2B.2   CD2B22853978  stand   74   NoReport stand              trk  0.51 OpenFloor  3   0     0.00  0.01  0.78  0.00  0.17  0.01
04:28:54 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.00  0.97  0.00  0.01  0.02
04:28:55 CD2B.2   CD2B22853978  walk    70   NoReport walk               trk  1.00 OpenFloor  3   0     0.00  0.01  0.91  0.00  0.03  0.01
04:28:55 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  3   0     0.00  0.01  0.94  0.00  0.00  0.01
04:28:55 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.94  0.00  0.00  0.01
04:28:56 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
04:28:56 CD2B.2   CD2B22853978  walk    81   NoReport walk               trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.01  0.01
04:28:56 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
04:28:57 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  3   8     0.00  0.01  0.93  0.00  0.00  0.01
04:28:57 CD2B.2   CD2B22853978  walk    77   NoReport walk               trk  1.00 OpenFloor  3   8     0.00  0.00  0.98  0.00  0.00  0.00
04:28:57 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  3   8     0.00  0.01  0.93  0.00  0.00  0.01
04:28:58 333B.E   -             -       0    NoReport np=1               room -    OpenFloor  3   8     0.00  0.01  0.93  0.00  0.00  0.01
04:28:58 333B.E   -             -       0    NoReport EnterRoom(rdr)     room -    OpenFloor  3   8     0.00  0.01  0.93  0.00  0.00  0.01
04:28:58 333B.0   -             stand   101  NoReport stand              room -    OpenFloor  3   8     0.00  0.01  0.93  0.00  0.00  0.01
04:28:58 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  3   9     0.00  0.02  0.79  0.00  0.01  0.10
04:28:58 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.02  0.79  0.00  0.01  0.10
04:28:58 CD2B.2   CD2B22853978  walk    0    NoReport walk               trk  1.00 OpenFloor  3   9     0.00  0.01  0.89  0.00  0.01  0.06
04:28:59 333B.0   -             stand   83   NoReport stand              room -    OpenFloor  3   9     0.00  0.02  0.79  0.00  0.01  0.10
04:28:59 CD2B.E   -             -       0    NoReport np=2               room -    OpenFloor  3   9     0.00  0.02  0.79  0.00  0.01  0.10
04:28:59 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  2   11    0.00  0.02  0.82  0.00  0.04  0.03
04:28:59 CD2B.2   CD2B22853978  walk    0    NoReport walk               trk  1.00 OpenFloor  2   11    0.00  0.01  0.92  0.00  0.01  0.01
04:29:00 333B.0   -             walk    82   NoReport walk               room -    OpenFloor  2   11    0.00  0.02  0.82  0.00  0.04  0.03
04:29:00 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  2   11    0.00  0.02  0.72  0.00  0.02  0.16
04:29:00 CD2B.2   CD2B22853978  stand   0    NoReport stand              trk  1.00 OpenFloor  2   11    0.00  0.01  0.86  0.00  0.00  0.09
04:29:01 333B.0   -             walk    84   NoReport walk               room -    OpenFloor  2   11    0.00  0.02  0.72  0.00  0.02  0.16
04:29:01 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  2   12    0.00  0.01  0.91  0.00  0.01  0.05
04:29:01 CD2B.2   CD2B22853978  stand   0    NoReport stand              trk  1.00 OpenFloor  2   12    0.00  0.00  0.99  0.00  0.00  0.01
04:29:02 333B.0   -             walk    87   NoReport walk               room -    OpenFloor  2   12    0.00  0.01  0.91  0.00  0.01  0.05
04:29:02 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  2   13    0.00  0.02  0.87  0.00  0.02  0.02
04:29:02 CD2B.2   CD2B22853978  stand   0    NoReport stand              trk  1.00 OpenFloor  2   13    0.00  0.00  0.97  0.00  0.00  0.01
04:29:03 333B.0   -             walk    90   NoReport walk               room -    OpenFloor  2   13    0.00  0.02  0.87  0.00  0.02  0.02
04:29:03 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.01  0.96  0.00  0.00  0.01
04:29:03 CD2B.2   CD2B22853978  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.00  1.00  0.00  0.00  0.00
04:29:03 CD2B.E2  CD2B22853978  -       0    NoReport ExitRoom(rdr)      trk  1.00 OpenFloor  2   14    0.00  0.00  1.00  0.00  0.00  0.00
04:29:04 333B.0   -             walk    120  NoReport walk               room -    OpenFloor  2   14    0.00  0.01  0.96  0.00  0.00  0.01
04:29:04 CD2B.E   -             -       0    NoReport np=1               room -    OpenFloor  2   14    0.00  0.01  0.96  0.00  0.00  0.01
04:29:04 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.03  0.79  0.00  0.01  0.04
04:29:05 333B.0   -             walk    0    NoReport walk               room -    OpenFloor  2   15    0.00  0.03  0.79  0.00  0.01  0.04
04:29:05 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   16    0.00  0.02  0.88  0.00  0.00  0.02
04:29:06 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   16    0.00  0.02  0.88  0.00  0.00  0.02
04:29:06 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   17    0.00  0.01  0.94  0.00  0.00  0.01
04:29:07 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   17    0.00  0.01  0.94  0.00  0.00  0.01
04:29:07 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   18    0.00  0.01  0.97  0.00  0.00  0.01
04:29:08 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   18    0.00  0.01  0.97  0.00  0.00  0.01
04:29:08 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   19    0.00  0.00  0.98  0.00  0.00  0.00
04:29:09 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   19    0.00  0.00  0.98  0.00  0.00  0.00
04:29:09 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   20    0.00  0.00  0.99  0.00  0.00  0.00
04:29:09 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   20    0.00  0.00  0.99  0.00  0.00  0.00
04:29:10 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   21    0.00  0.00  1.00  0.00  0.00  0.00
04:29:10 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   21    0.00  0.00  1.00  0.00  0.00  0.00
04:29:11 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   22    0.00  0.00  1.00  0.00  0.00  0.00
04:29:11 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   22    0.00  0.00  1.00  0.00  0.00  0.00
04:29:12 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   23    0.00  0.00  1.00  0.00  0.00  0.00
04:29:12 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   23    0.00  0.00  1.00  0.00  0.00  0.00
04:29:13 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   24    0.00  0.00  1.00  0.00  0.00  0.00
04:29:13 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   24    0.00  0.00  1.00  0.00  0.00  0.00
04:29:14 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   25    0.00  0.00  1.00  0.00  0.00  0.00
04:29:14 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   25    0.00  0.00  1.00  0.00  0.00  0.00
04:29:15 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   26    0.00  0.00  1.00  0.00  0.00  0.00
04:29:15 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   26    0.00  0.00  1.00  0.00  0.00  0.00
04:29:16 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   27    0.00  0.00  1.00  0.00  0.00  0.00
04:29:16 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   27    0.00  0.00  1.00  0.00  0.00  0.00
04:29:17 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   28    0.00  0.00  1.00  0.00  0.00  0.00
04:29:17 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   28    0.00  0.00  1.00  0.00  0.00  0.00
04:29:18 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   29    0.00  0.00  1.00  0.00  0.00  0.00
04:29:18 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   29    0.00  0.00  1.00  0.00  0.00  0.00
04:29:19 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   30    0.00  0.00  1.00  0.00  0.00  0.00
04:29:19 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   30    0.00  0.00  1.00  0.00  0.00  0.00
04:29:20 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   31    0.00  0.00  1.00  0.00  0.00  0.00
04:29:20 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   31    0.00  0.00  1.00  0.00  0.00  0.00
04:29:21 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   32    0.00  0.00  1.00  0.00  0.00  0.00
04:29:21 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   32    0.00  0.00  1.00  0.00  0.00  0.00
04:29:22 CD2B.1   CD2B12838096  stand   0    NoReport stand              trk  1.00 OpenFloor  1   33    0.00  0.00  1.00  0.00  0.00  0.00
04:29:22 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   33    0.00  0.00  1.00  0.00  0.00  0.00
04:29:23 CD2B.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   33    0.00  0.00  1.00  0.00  0.00  0.00
04:29:23 CD2B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   34    0.00  0.00  1.00  0.00  0.00  0.00
04:29:23 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   34    0.00  0.00  1.00  0.00  0.00  0.00
04:29:24 CD2B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:24 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:25 CD2B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:25 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:26 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:27 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:28 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:29 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:30 333B.0   -             stand   57   NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:31 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:32 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:33 333B.0   -             stand   141  NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:34 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:35 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:36 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:37 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:38 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:39 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:40 333B.0   -             stand   118  NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:41 333B.0   -             stand   87   NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:42 333B.0   -             stand   96   NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:43 333B.0   -             stand   103  NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:44 333B.0   -             stand   103  NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:45 333B.0   -             stand   85   NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:46 333B.0   -             stand   82   NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:47 333B.0   -             stand   110  NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:48 333B.0   -             stand   113  NoReport stand              room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:49 333B.0   -             walk    77   NoReport walk               room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:29:50 333B.0   -             walk    59   NoReport walk               room -    Left       0   0     0.00  0.01  0.04  0.00  0.04  0.88
04:29:50 CD2B.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:29:51 333B.0   -             walk    79   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:29:52 333B.0   -             walk    93   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:29:53 333B.0   -             walk    79   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:29:54 333B.0   -             walk    83   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:29:55 333B.0   -             walk    95   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:29:56 333B.0   -             walk    71   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:29:57 333B.0   -             walk    92   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:29:58 333B.0   -             walk    82   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:29:59 333B.0   -             walk    83   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:30:00 333B.0   -             walk    77   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:30:01 333B.0   -             walk    90   NoReport walk               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:30:01 CD2B.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:30:01 CD2B.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
04:30:01 CD2B.0   CD2B03146333  stand   81   NoReport stand              room -    Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
04:30:02 CD2B.0   CD2B03146333  stand   74   NoReport stand              room -    Empty      1   0     0.00  0.03  0.35  0.00  0.54  0.02
04:30:02 333B.0   -             walk    0    NoReport walk               room -    Empty      1   0     0.00  0.03  0.35  0.00  0.54  0.02
04:30:03 CD2B.0   CD2B03146333  walk    67   NoReport walk               room -    OpenFloor  1   0     0.00  0.03  0.43  0.01  0.41  0.02
04:30:03 333B.0   -             walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.03  0.43  0.01  0.41  0.02
04:30:04 CD2B.0   CD2B03146333  walk    69   NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.49  0.01  0.30  0.02
04:30:04 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.04  0.49  0.01  0.30  0.02
04:30:05 CD2B.0   CD2B03146333  walk    83   NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.53  0.01  0.22  0.03
04:30:05 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.04  0.53  0.01  0.22  0.03
04:30:06 CD2B.0   CD2B03146333  walk    88   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.56  0.01  0.16  0.03
04:30:06 333B.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.56  0.01  0.16  0.03
04:30:06 333B.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  1   0     0.00  0.05  0.56  0.01  0.16  0.03
04:30:07 CD2B.0   CD2B03146333  walk    66   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.58  0.02  0.12  0.03
04:30:07 333B.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   0     0.00  0.05  0.58  0.02  0.12  0.03
04:30:07 333B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   0     0.00  0.05  0.58  0.02  0.12  0.03
04:30:08 CD2B.0   CD2B03146333  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.59  0.02  0.10  0.03
04:30:08 333B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   0     0.00  0.05  0.59  0.02  0.10  0.03
04:30:09 CD2B.0   CD2B03146333  walk    79   NoReport walk               room -    OpenFloor  1   0     0.00  0.02  0.85  0.01  0.04  0.01
04:30:09 333B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   0     0.00  0.02  0.85  0.01  0.04  0.01
04:30:10 CD2B.0   CD2B03146333  walk    79   NoReport walk               room -    OpenFloor  1   0     0.00  0.01  0.92  0.00  0.02  0.01
04:30:11 CD2B.0   CD2B03146333  walk    82   NoReport walk               room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.01  0.00
04:30:12 CD2B.0   CD2B03146333  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.00  0.99  0.00  0.00  0.00
04:30:12 333B.E   -             -       0    NoReport np=1               room -    OpenFloor  1   0     0.00  0.00  0.99  0.00  0.00  0.00
04:30:12 333B.E   -             -       0    NoReport EnterRoom(rdr)     room -    OpenFloor  1   0     0.00  0.00  0.99  0.00  0.00  0.00
04:30:12 333B.0   -             stand   113  NoReport stand              room -    OpenFloor  1   0     0.00  0.00  0.99  0.00  0.00  0.00
04:30:12 333B.0   -             stand   114  NoReport stand              room -    OpenFloor  1   0     0.00  0.00  0.99  0.00  0.00  0.00
04:30:13 CD2B.0   CD2B03146333  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.00  0.99  0.00  0.00  0.00
04:30:13 333B.0   -             stand   69   NoReport stand              room -    OpenFloor  1   0     0.00  0.00  0.99  0.00  0.00  0.00
04:30:14 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:14 333B.0   -             walk    85   NoReport walk               room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:15 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:15 333B.0   -             walk    97   NoReport walk               room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:16 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:16 333B.0   -             walk    90   NoReport walk               room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:17 CD2B.0   CD2B03146333  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:17 CD2B.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:17 333B.0   -             walk    100  NoReport walk               room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:18 CD2B.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:18 CD2B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:18 333B.0   -             walk    89   NoReport walk               room -    OpenFloor  1   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:19 CD2B.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:19 333B.0   -             walk    95   NoReport walk               room -    OpenFloor  0   0     0.00  0.00  1.00  0.00  0.00  0.00
04:30:20 CD2B.88  -             88      -    NoReport no-target(88)      room -    Left       0   0     0.00  0.00  0.22  0.00  0.00  0.78
04:30:20 333B.0   -             walk    93   NoReport walk               room -    Left       0   0     0.00  0.00  0.22  0.00  0.00  0.78
04:30:21 333B.0   -             walk    83   NoReport walk               room -    Left       0   0     0.00  0.00  0.22  0.00  0.00  0.78
04:30:22 CD2B.88  -             88      -    NoReport no-target(88)      room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:22 333B.0   -             walk    103  NoReport walk               room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:23 333B.0   -             walk    102  NoReport walk               room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:24 333B.0   -             walk    84   NoReport walk               room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:25 333B.0   -             stand   0    NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:26 333B.0   -             stand   0    NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:27 333B.0   -             stand   0    NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:28 333B.0   -             stand   79   NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:29 333B.0   -             stand   0    NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:30 333B.0   -             stand   102  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:31 333B.0   -             stand   107  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:32 333B.0   -             stand   90   NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:33 333B.0   -             stand   94   NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:34 333B.0   -             stand   0    NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:35 333B.0   -             stand   0    NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:36 333B.0   -             stand   90   NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:37 333B.0   -             stand   106  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:38 333B.0   -             stand   98   NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:39 333B.0   -             stand   102  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:40 333B.0   -             stand   102  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:41 333B.0   -             stand   96   NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:42 333B.0   -             stand   97   NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:43 333B.0   -             stand   98   NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:44 333B.0   -             stand   82   NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:45 333B.0   -             stand   108  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:46 333B.0   -             stand   135  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:47 333B.0   -             stand   114  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:48 333B.0   -             stand   0    NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:49 333B.0   -             stand   120  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:50 333B.0   -             stand   121  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:51 333B.0   -             stand   125  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:52 333B.0   -             stand   125  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:53 333B.0   -             stand   126  NoReport stand              room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  1.00
04:30:54 CD2B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:30:54 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:30:55 333B.0   -             stand   122  InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:30:56 333B.0   -             stand   117  InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:30:57 333B.0   -             stand   117  InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:30:58 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:30:59 333B.0   -             stand   95   InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:00 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:01 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:02 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:03 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:04 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:05 333B.0   -             stand   116  InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:06 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:07 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:08 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:09 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:10 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:11 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:12 333B.0   -             stand   108  InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:13 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:14 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:14 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:15 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:16 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:17 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:18 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:19 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:20 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:21 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:22 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:23 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:24 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:25 CD2B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:25 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:26 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:27 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:28 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:29 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:30 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:31 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:32 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:33 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:34 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:35 333B.0   -             stand   88   InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:36 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:37 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:38 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:39 333B.0   -             stand   92   InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:40 333B.0   -             stand   84   InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:41 333B.0   -             walk    106  InBed    walk               room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:42 333B.0   -             walk    102  InBed    walk               room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:43 333B.0   -             walk    89   InBed    walk               room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:44 333B.0   -             walk    85   InBed    walk               room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:45 333B.0   -             walk    0    InBed    walk               room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:46 CD2B.E   -             -       0    InBed    np=1               room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:46 CD2B.E   -             -       0    InBed    EnterRoom(rdr)     room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:46 CD2B.0   CD2B03146333  stand   72   InBed    stand              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:46 CD2B.0   CD2B03146333  walk    61   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:46 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:47 CD2B.0   CD2B03146333  walk    72   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:47 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:48 CD2B.0   CD2B03146333  walk    82   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:48 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:49 CD2B.0   CD2B03146333  walk    92   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:49 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:50 CD2B.0   CD2B03146333  walk    93   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:50 333B.0   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:50 333B.E   -             -       0    InBed    ExitRoom(rdr)      room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:51 CD2B.0   CD2B03146333  walk    112  InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:51 333B.E   -             -       0    InBed    np=0  ★0           room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:51 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:52 CD2B.0   CD2B03146333  walk    98   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:52 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:53 CD2B.0   CD2B03146333  walk    81   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:53 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:54 CD2B.0   CD2B03146333  walk    109  InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:55 CD2B.0   CD2B03146333  walk    89   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:56 CD2B.0   CD2B03146333  walk    82   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:56 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:57 CD2B.0   CD2B03146333  walk    79   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:58 CD2B.0   CD2B03146333  walk    67   InBed    walk               trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:31:59 CD2B.0   CD2B03146333  lying   71   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:00 CD2B.0   CD2B03146333  lying   67   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:00 CD2B.E   -             -       0    InBed    InBed(rdr)         room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:01 CD2B.0   CD2B03146333  lying   71   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:02 CD2B.0   CD2B03146333  lying   60   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:03 CD2B.0   CD2B03146333  lying   70   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:04 CD2B.0   CD2B03146333  lying   68   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:05 CD2B.0   CD2B03146333  lying   60   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:06 CD2B.0   CD2B03146333  lying   66   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:07 CD2B.0   CD2B03146333  lying   72   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:08 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:08 CD2B.0   CD2B03146333  lying   70   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:09 1641.E   -             -       0    InBed    InBed(pad)         room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:09 1641.E   -             -       0    InBed    InBed(pad)         room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:09 CD2B.0   CD2B03146333  lying   66   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:10 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=15 mv=1 turn=0 room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:10 CD2B.0   CD2B03146333  lying   60   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:11 CD2B.0   CD2B03146333  lying   63   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:12 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=16 mv=1 turn=0 room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:12 CD2B.0   CD2B03146333  lying   65   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:13 CD2B.0   CD2B03146333  lying   59   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:14 1641.0   -             pad     -    InBed    pad InBed HR=None RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:14 CD2B.0   CD2B03146333  lying   65   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:15 CD2B.0   CD2B03146333  lying   61   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:16 1641.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=1 turn=0 room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:16 CD2B.0   CD2B03146333  lying   59   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:17 CD2B.0   CD2B03146333  lying   59   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:18 1641.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=1 turn=0 room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:18 CD2B.0   CD2B03146333  lying   64   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:19 CD2B.0   CD2B03146333  lying   58   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:20 1641.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:20 CD2B.0   CD2B03146333  lying   53   InBed    lying              trk  1.00 Bed        1   0     0.00  1.00  0.00  0.00  0.00  0.00
04:32:21 CD2B.0   CD2B03146333  lying   51   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:22 1641.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:22 CD2B.0   CD2B03146333  lying   49   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:23 CD2B.0   CD2B03146333  lying   50   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:24 1641.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:24 CD2B.0   CD2B03146333  lying   54   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:25 CD2B.0   CD2B03146333  lying   54   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:26 CD2B.0   CD2B03146333  lying   51   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:27 1641.0   -             pad     -    InBed    pad InBed HR=None RR=14 mv=0 turn=1 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:27 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:28 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:28 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:29 1641.0   -             pad     -    InBed    pad InBed HR=None RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:29 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:30 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:31 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:31 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:32 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:33 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:33 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:34 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:35 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:35 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:36 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:37 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:37 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:38 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:39 1641.0   -             pad     -    InBed    pad InBed HR=67 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:39 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:40 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:41 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:41 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:42 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:43 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:43 CD2B.0   CD2B03146333  lying   56   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:44 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:45 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=14 mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:45 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:46 CD2B.0   CD2B03146333  lying   49   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:47 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:47 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:48 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:49 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:49 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:50 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:51 1641.0   -             pad     -    InBed    pad InBed HR=55 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:51 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:51 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:52 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:53 1641.0   -             pad     -    InBed    pad InBed HR=55 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:54 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:55 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=19 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:55 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:56 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:57 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:57 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:58 CD2B.0   CD2B03146333  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
04:32:59 1641.0   -             pad     -    InBed    pad InBed HR=57 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
04:27:11.082 333B.88   88     -    -      -      -     -    -    
04:27:43.028 333B.88   88     -    -      -      -     -    -    
04:28:14.402 333B.88   88     -    -      -      -     -    -    
04:28:46.312 333B.88   88     -    -      -      -     -    -    
04:28:58.864 333B.0    stand  1    -30    190    101   80        
04:28:59.044 333B.0    stand  1    -50    200    83    80   22   
04:29:00.045 333B.0    walk   1    -150   190    82    80   100  
04:29:01.046 333B.0    walk   1    -210   190    84    80   60   
04:29:02.048 333B.0    walk   1    -260   170    87    80   53   
04:29:03.052 333B.0    walk   1    -300   180    90    80   41   
04:29:04.048 333B.0    walk   1    -320   220    120   80   44   
04:29:05.052 333B.0    walk   1    -280   240    0     80   44   
04:29:06.064 333B.0    stand  1    -300   230    0     80   22   
04:29:07.060 333B.0    stand  1    -300   220    0     80   10   
04:29:08.055 333B.0    stand  1    -300   220    0     80   0    
04:29:09.056 333B.0    stand  1    -300   220    0     80   0    
04:29:09.953 333B.0    stand  1    -300   220    0     80   0    
04:29:10.956 333B.0    stand  1    -300   220    0     80   0    
04:29:11.954 333B.0    stand  1    -300   220    0     80   0    
04:29:12.962 333B.0    stand  1    -300   220    0     80   0    
04:29:13.958 333B.0    stand  1    -300   220    0     80   0    
04:29:14.959 333B.0    stand  1    -300   220    0     80   0    
04:29:15.968 333B.0    stand  1    -300   220    0     80   0    
04:29:16.966 333B.0    stand  1    -300   220    0     80   0    
04:29:17.960 333B.0    stand  1    -300   220    0     80   0    
04:29:18.964 333B.0    stand  1    -300   220    0     80   0    
04:29:19.968 333B.0    stand  1    -300   220    0     80   0    
04:29:20.966 333B.0    stand  1    -300   220    0     80   0    
04:29:21.858 333B.0    stand  1    -300   220    0     80   0    
04:29:22.858 333B.0    stand  1    -300   220    0     80   0    
04:29:23.860 333B.0    stand  1    -300   220    0     80   0    
04:29:24.860 333B.0    stand  1    -300   220    0     80   0    
04:29:25.862 333B.0    stand  1    -300   220    0     80   0    
04:29:26.864 333B.0    stand  1    -300   220    0     80   0    
04:29:27.867 333B.0    stand  1    -300   220    0     80   0    
04:29:28.866 333B.0    stand  1    -300   220    0     80   0    
04:29:29.866 333B.0    stand  1    -300   220    0     80   0    
04:29:30.870 333B.0    stand  1    -290   230    57    80   14   
04:29:31.870 333B.0    stand  1    -330   200    0     80   50   
04:29:32.869 333B.0    stand  1    -330   200    0     80   0    
04:29:33.762 333B.0    stand  1    -310   210    141   80   22   
04:29:34.764 333B.0    stand  1    -310   220    0     80   10   
04:29:35.764 333B.0    stand  1    -310   220    0     80   0    
04:29:36.768 333B.0    stand  1    -330   200    0     80   28   
04:29:37.770 333B.0    stand  1    -320   200    0     80   10   
04:29:38.784 333B.0    stand  1    -310   200    0     80   10   
04:29:39.834 333B.0    stand  1    -310   200    0     80   0    
04:29:40.792 333B.0    stand  1    -290   220    118   80   28   
04:29:41.788 333B.0    stand  1    -290   190    87    80   30   
04:29:42.790 333B.0    stand  1    -280   170    96    80   22   
04:29:43.692 333B.0    stand  1    -290   190    103   80   22   
04:29:44.688 333B.0    stand  1    -280   220    103   80   31   
04:29:45.687 333B.0    stand  1    -280   210    85    80   10   
04:29:46.692 333B.0    stand  1    -290   190    82    80   22   
04:29:47.688 333B.0    stand  1    -290   180    110   80   10   
04:29:48.689 333B.0    stand  1    -270   180    113   80   20   
04:29:49.692 333B.0    walk   1    -210   130    77    80   78   
04:29:50.696 333B.0    walk   1    -190   90     59    80   44   
04:29:51.692 333B.0    walk   1    -200   120    79    80   31   
04:29:52.694 333B.0    walk   1    -240   180    93    80   72   
04:29:53.705 333B.0    walk   1    -290   200    79    80   53   
04:29:54.596 333B.0    walk   1    -290   190    83    80   10   
04:29:55.608 333B.0    walk   1    -270   180    95    80   22   
04:29:56.598 333B.0    walk   1    -210   140    71    80   72   
04:29:57.600 333B.0    walk   1    -160   180    92    80   64   
04:29:58.602 333B.0    walk   1    -100   200    82    80   63   
04:29:59.600 333B.0    walk   1    -100   200    83    80   0    
04:30:00.624 333B.0    walk   1    -70    210    77    80   31   
04:30:01.604 333B.0    walk   1    0      160    90    80   86   
04:30:02.604 333B.0    walk   1    0      150    0     80   10   
04:30:03.605 333B.0    walk   1    0      150    0     80   0    
04:30:04.605 333B.0    stand  1    0      150    0     80   0    
04:30:05.507 333B.0    stand  1    0      150    0     80   0    
04:30:06.506 333B.0    stand  1    0      150    0     80   0    
04:30:07.568 333B.88   88     -    -      -      -     -    -    
04:30:08.519 333B.88   88     -    -      -      -     -    -    
04:30:09.522 333B.88   88     -    -      -      -     -    -    
04:30:12.464 333B.0    stand  1    -10    180    113   80   31   
04:30:12.502 333B.0    stand  1    -10    190    114   80   10   
04:30:13.469 333B.0    stand  1    -70    210    69    80   63   
04:30:14.468 333B.0    walk   1    -80    240    85    80   31   
04:30:15.470 333B.0    walk   1    -100   260    97    80   28   
04:30:16.472 333B.0    walk   1    -110   260    90    80   10   
04:30:17.468 333B.0    walk   1    -160   240    100   80   53   
04:30:18.472 333B.0    walk   1    -230   190    89    80   86   
04:30:19.474 333B.0    walk   1    -260   180    95    80   31   
04:30:20.474 333B.0    walk   1    -210   170    93    80   50   
04:30:21.474 333B.0    walk   1    -160   150    83    80   53   
04:30:22.376 333B.0    walk   1    -120   240    103   80   98   
04:30:23.376 333B.0    walk   1    -100   280    102   80   44   
04:30:24.378 333B.0    walk   1    -90    260    84    80   22   
04:30:25.376 333B.0    stand  1    -90    290    0     80   30   
04:30:26.418 333B.0    stand  1    -90    260    0     80   30   
04:30:27.420 333B.0    stand  1    -100   260    0     80   10   
04:30:28.422 333B.0    stand  1    -90    260    79    80   10   
04:30:29.316 333B.0    stand  1    -100   260    0     80   10   
04:30:30.314 333B.0    stand  1    -70    280    102   80   36   
04:30:31.318 333B.0    stand  1    -90    270    107   80   22   
04:30:32.316 333B.0    stand  1    -100   260    90    80   14   
04:30:33.322 333B.0    stand  1    -90    260    94    80   10   
04:30:34.320 333B.0    stand  1    -100   270    0     80   14   
04:30:35.324 333B.0    stand  1    -100   260    0     80   10   
04:30:36.323 333B.0    stand  1    -100   260    90    80   0    
04:30:37.328 333B.0    stand  1    -100   280    106   80   20   
04:30:38.324 333B.0    stand  1    -100   260    98    80   20   
04:30:39.373 333B.0    stand  1    -90    260    102   80   10   
04:30:40.324 333B.0    stand  1    -80    240    102   80   22   
04:30:41.221 333B.0    stand  1    -70    210    96    80   31   
04:30:42.228 333B.0    stand  1    -80    230    97    80   22   
04:30:43.225 333B.0    stand  1    -90    260    98    80   31   
04:30:44.226 333B.0    stand  1    -90    270    82    80   10   
04:30:45.234 333B.0    stand  1    -100   270    108   80   10   
04:30:46.236 333B.0    stand  1    -100   270    135   80   0    
04:30:47.232 333B.0    stand  1    -100   260    114   80   10   
04:30:48.234 333B.0    stand  1    -110   290    0     80   31   
04:30:49.232 333B.0    stand  1    -100   270    120   80   22   
04:30:50.234 333B.0    stand  1    -110   270    121   80   10   
04:30:51.238 333B.0    stand  1    -100   270    125   80   10   
04:30:52.130 333B.0    stand  1    -100   270    125   80   0    
04:30:53.132 333B.0    stand  1    -90    270    126   80   10   
04:30:54.135 333B.0    stand  1    -90    280    0     80   10   
04:30:55.146 333B.0    stand  1    -110   270    122   80   22   
04:30:56.136 333B.0    stand  1    -100   270    117   80   10   
04:30:57.136 333B.0    stand  1    -110   260    117   80   14   
04:30:58.153 333B.0    stand  1    -110   260    0     80   0    
04:30:59.138 333B.0    stand  1    -90    270    95    80   22   
04:31:00.140 333B.0    stand  1    -90    270    0     80   0    
04:31:01.142 333B.0    stand  1    -80    250    0     80   22   
04:31:02.141 333B.0    stand  1    -80    250    0     80   0    
04:31:03.141 333B.0    stand  1    -80    250    0     80   0    
04:31:04.048 333B.0    stand  1    -80    250    0     80   0    
04:31:05.036 333B.0    stand  1    -90    270    116   80   22   
04:31:06.037 333B.0    stand  1    -90    270    0     80   0    
04:31:07.037 333B.0    stand  1    -100   270    0     80   10   
04:31:08.044 333B.0    stand  1    -100   270    0     80   0    
04:31:09.040 333B.0    stand  1    -100   270    0     80   0    
04:31:10.050 333B.0    stand  1    -100   270    0     80   0    
04:31:11.060 333B.0    stand  1    -100   270    0     80   0    
04:31:12.044 333B.0    stand  1    -120   270    108   80   20   
04:31:13.044 333B.0    stand  1    -130   270    0     80   10   
04:31:14.045 333B.0    stand  1    -130   270    0     80   0    
04:31:14.945 333B.0    stand  1    -80    270    0     80   50   
04:31:15.956 333B.0    stand  1    -70    270    0     80   10   
04:31:16.956 333B.0    stand  1    -70    260    0     80   10   
04:31:17.952 333B.0    stand  1    -100   280    0     80   36   
04:31:18.949 333B.0    stand  1    -100   300    0     80   20   
04:31:19.954 333B.0    stand  1    -80    280    0     80   28   
04:31:20.954 333B.0    stand  1    -80    280    0     80   0    
04:31:21.954 333B.0    stand  1    -80    280    0     80   0    
04:31:22.954 333B.0    stand  1    -80    280    0     80   0    
04:31:23.956 333B.0    stand  1    -100   290    0     80   22   
04:31:24.956 333B.0    stand  1    -80    280    0     80   22   
04:31:25.968 333B.0    stand  1    -80    260    0     80   20   
04:31:26.849 333B.0    stand  1    -80    260    0     80   0    
04:31:27.853 333B.0    stand  1    -100   300    0     80   44   
04:31:28.854 333B.0    stand  1    -60    310    0     80   41   
04:31:29.853 333B.0    stand  1    -100   290    0     80   44   
04:31:30.857 333B.0    stand  1    -100   290    0     80   0    
04:31:31.857 333B.0    stand  1    -100   290    0     80   0    
04:31:32.857 333B.0    stand  1    -100   290    0     80   0    
04:31:33.859 333B.0    stand  1    -100   290    0     80   0    
04:31:34.865 333B.0    stand  1    -100   290    0     80   0    
04:31:35.862 333B.0    stand  1    -90    280    88    80   14   
04:31:36.866 333B.0    stand  1    -80    270    0     80   14   
04:31:37.764 333B.0    stand  1    -70    280    0     80   14   
04:31:38.766 333B.0    stand  1    0      280    0     80   70   
04:31:39.813 333B.0    stand  1    -80    260    92    80   82   
04:31:40.768 333B.0    stand  1    -70    270    84    80   14   
04:31:41.766 333B.0    walk   1    -100   270    106   80   30   
04:31:42.769 333B.0    walk   1    -80    260    102   80   22   
04:31:43.770 333B.0    walk   1    -40    230    89    80   50   
04:31:44.768 333B.0    walk   1    0      190    85    80   56   
04:31:45.776 333B.0    walk   1    0      190    0     80   0    
04:31:46.693 333B.0    stand  1    0      190    0     80   0    
04:31:47.696 333B.0    stand  1    0      190    0     80   0    
04:31:48.693 333B.0    stand  1    0      190    0     80   0    
04:31:49.696 333B.0    stand  1    0      190    0     80   0    
04:31:50.693 333B.0    stand  1    0      190    0     80   0    
04:31:51.752 333B.88   88     -    -      -      -     -    -    
04:31:52.704 333B.88   88     -    -      -      -     -    -    
04:31:53.718 333B.88   88     -    -      -      -     -    -    
04:31:56.708 333B.88   88     -    -      -      -     -    -    
04:32:28.352 333B.88   88     -    -      -      -     -    -    

04:27:00.837 CD2B.0    stand  255  -150   480    119   80        
04:27:01.828 CD2B.0    stand  255  -130   490    0     80   22   
04:27:02.828 CD2B.0    stand  255  -130   490    0     80   0    
04:27:03.831 CD2B.0    stand  255  -130   490    0     80   0    
04:27:04.837 CD2B.0    stand  255  -130   490    0     80   0    
04:27:05.834 CD2B.0    stand  255  -130   490    0     80   0    
04:27:06.834 CD2B.0    stand  255  -130   490    0     80   0    
04:27:07.836 CD2B.0    stand  255  -140   490    0     80   10   
04:27:08.838 CD2B.0    stand  255  -140   490    0     80   0    
04:27:09.836 CD2B.0    stand  255  -140   490    0     80   0    
04:27:10.838 CD2B.0    stand  255  -140   490    0     80   0    
04:27:11.728 CD2B.0    stand  255  -140   490    0     80   0    
04:27:12.731 CD2B.0    stand  255  -140   490    0     80   0    
04:27:13.730 CD2B.0    stand  255  -140   490    0     80   0    
04:27:14.735 CD2B.0    stand  255  -140   490    0     80   0    
04:27:15.734 CD2B.0    stand  255  -140   490    0     80   0    
04:27:16.736 CD2B.0    stand  255  -140   490    0     80   0    
04:27:17.754 CD2B.0    stand  255  -140   490    0     80   0    
04:27:18.758 CD2B.0    stand  255  -140   490    0     80   0    
04:27:19.760 CD2B.0    stand  255  -140   490    0     80   0    
04:27:20.670 CD2B.0    stand  255  -140   490    0     80   0    
04:27:21.659 CD2B.0    stand  255  -140   490    0     80   0    
04:27:22.660 CD2B.0    stand  255  -140   490    0     80   0    
04:27:23.662 CD2B.0    stand  255  -140   490    0     80   0    
04:27:24.667 CD2B.0    stand  255  -140   490    0     80   0    
04:27:25.662 CD2B.0    stand  255  -140   490    0     80   0    
04:27:26.662 CD2B.0    stand  255  -140   490    0     80   0    
04:27:27.662 CD2B.0    stand  255  -140   490    0     80   0    
04:27:28.664 CD2B.0    stand  255  -140   490    0     80   0    
04:27:29.678 CD2B.0    stand  255  -140   490    0     80   0    
04:27:30.672 CD2B.0    stand  255  -140   490    0     80   0    
04:27:31.670 CD2B.0    stand  255  -140   490    0     80   0    
04:27:32.560 CD2B.0    stand  255  -140   490    0     80   0    
04:27:33.571 CD2B.0    stand  255  -140   490    0     80   0    
04:27:34.568 CD2B.0    stand  255  -140   490    0     80   0    
04:27:35.570 CD2B.0    stand  255  -140   490    0     80   0    
04:27:36.583 CD2B.0    stand  255  -140   490    0     80   0    
04:27:37.576 CD2B.0    stand  255  -140   490    0     80   0    
04:27:38.580 CD2B.0    stand  255  -140   490    0     80   0    
04:27:39.580 CD2B.0    stand  255  -140   490    0     80   0    
04:27:40.578 CD2B.0    stand  255  -140   490    0     80   0    
04:27:41.630 CD2B.0    stand  255  -140   490    0     80   0    
04:27:42.580 CD2B.0    stand  255  -140   490    0     80   0    
04:27:43.472 CD2B.0    stand  255  -140   490    0     80   0    
04:27:44.478 CD2B.0    stand  255  -140   490    0     80   0    
04:27:45.474 CD2B.0    stand  255  -140   490    0     80   0    
04:27:46.482 CD2B.0    stand  255  -140   490    0     80   0    
04:27:47.483 CD2B.0    stand  255  -140   490    0     80   0    
04:27:48.478 CD2B.0    stand  255  -140   490    0     80   0    
04:27:49.426 CD2B.0    stand  255  -140   490    0     80   0    
04:27:50.424 CD2B.0    stand  255  -140   490    0     80   0    
04:27:51.426 CD2B.0    stand  255  -140   490    0     80   0    
04:27:52.426 CD2B.0    stand  255  -140   490    0     80   0    
04:27:53.444 CD2B.0    stand  255  -140   490    0     80   0    
04:27:54.442 CD2B.0    stand  255  -140   490    0     80   0    
04:27:55.432 CD2B.0    stand  255  -140   490    0     80   0    
04:27:56.432 CD2B.0    stand  255  -140   490    0     80   0    
04:27:57.433 CD2B.0    stand  255  -140   490    0     80   0    
04:27:58.440 CD2B.0    stand  255  -140   490    0     80   0    
04:27:59.446 CD2B.0    stand  255  -140   490    0     80   0    
04:28:00.450 CD2B.0    stand  255  -140   490    0     80   0    
04:28:01.328 CD2B.0    stand  255  -140   490    0     80   0    
04:28:02.333 CD2B.0    stand  255  -140   490    0     80   0    
04:28:03.338 CD2B.0    stand  255  -140   490    0     80   0    
04:28:04.342 CD2B.0    stand  255  -140   490    0     80   0    
04:28:05.389 CD2B.0    stand  255  -140   490    0     80   0    
04:28:06.288 CD2B.0    stand  255  -100   500    0     80   41   
04:28:07.290 CD2B.0    stand  255  -170   500    0     80   70   
04:28:08.296 CD2B.0    stand  255  -190   490    0     80   22   
04:28:09.292 CD2B.0    stand  255  -170   500    144   80   22   
04:28:10.292 CD2B.0    stand  255  -150   490    0     80   22   
04:28:11.292 CD2B.0    stand  255  -140   510    0     80   22   
04:28:12.298 CD2B.0    stand  255  -150   500    139   80   14   
04:28:13.302 CD2B.0    stand  255  -170   510    0     80   22   
04:28:14.296 CD2B.0    stand  255  -170   510    0     80   0    
04:28:15.300 CD2B.0    stand  255  -170   510    0     80   0    
04:28:16.298 CD2B.0    stand  255  -170   510    0     80   0    
04:28:17.305 CD2B.0    stand  255  -170   510    0     80   0    
04:28:18.196 CD2B.0    stand  255  -170   510    0     80   0    
04:28:19.193 CD2B.0    stand  255  -170   510    0     80   0    
04:28:20.200 CD2B.0    stand  255  -170   510    0     80   0    
04:28:21.213 CD2B.0    stand  255  -170   510    0     80   0    
04:28:22.204 CD2B.0    stand  255  -170   510    0     80   0    
04:28:23.205 CD2B.0    stand  255  -170   510    0     80   0    
04:28:24.206 CD2B.0    stand  255  -170   510    0     80   0    
04:28:25.208 CD2B.0    stand  255  -170   510    0     80   0    
04:28:26.208 CD2B.0    stand  255  -170   510    0     80   0    
04:28:27.205 CD2B.0    stand  255  -170   510    0     80   0    
04:28:28.208 CD2B.0    stand  255  -170   510    0     80   0    
04:28:29.105 CD2B.0    stand  255  -170   510    0     80   0    
04:28:30.108 CD2B.0    stand  255  -170   510    0     80   0    
04:28:31.112 CD2B.0    stand  255  -170   510    0     80   0    
04:28:32.109 CD2B.0    stand  255  -170   510    0     80   0    
04:28:33.113 CD2B.0    stand  255  -170   510    0     80   0    
04:28:34.111 CD2B.0    stand  255  -170   510    0     80   0    
04:28:35.113 CD2B.0    stand  255  -170   510    0     80   0    
04:28:36.112 CD2B.0    stand  255  -160   500    0     80   14   
04:28:37.048 CD2B.0    stand  255  -160   500    0     80   0    
04:28:38.096 CD2B.0    stand  255  -160   500    0     80   0    
04:28:38.096 CD2B.1    stand  5    -330   10     0     80   518  
04:28:39.056 CD2B.0    stand  255  -160   500    0     80   518  
04:28:39.056 CD2B.1    stand  5    -300   30     86    80   490  
04:28:40.060 CD2B.1    walk   5    -260   60     84    80   50   
04:28:40.060 CD2B.0    stand  255  -150   490    0     80   443  
04:28:41.101 CD2B.0    stand  255  -150   500    0     80   10   
04:28:41.101 CD2B.1    walk   5    -220   130    83    80   376  
04:28:42.056 CD2B.1    walk   5    -230   220    72    80   90   
04:28:42.056 CD2B.0    stand  255  -140   500    0     80   294  
04:28:43.056 CD2B.1    walk   5    -220   330    88    80   187  
04:28:43.056 CD2B.0    stand  255  -160   500    0     80   180  
04:28:44.060 CD2B.0    stand  255  -160   500    0     80   0    
04:28:44.060 CD2B.1    walk   5    -210   410    70    80   102  
04:28:45.060 CD2B.0    stand  255  -200   530    0     80   120  
04:28:45.060 CD2B.1    walk   5    -190   470    83    80   60   
04:28:46.060 CD2B.1    walk   5    -200   530    0     80   60   
04:28:46.060 CD2B.0    stand  255  -230   530    0     80   30   
04:28:47.060 CD2B.1    walk   5    -190   510    71    80   44   
04:28:47.060 CD2B.0    stand  255  -230   530    0     80   44   
04:28:47.960 CD2B.0    stand  255  -230   530    0     80   0    
04:28:47.960 CD2B.1    walk   5    -200   490    94    80   50   
04:28:48.957 CD2B.0    stand  255  -230   530    0     80   50   
04:28:48.957 CD2B.1    walk   5    -250   460    80    80   72   
04:28:49.960 CD2B.1    walk   5    -250   430    0     80   30   
04:28:49.960 CD2B.0    stand  255  -230   530    0     80   101  
04:28:50.957 CD2B.0    stand  255  -230   530    0     80   0    
04:28:50.957 CD2B.1    walk   5    -250   440    0     80   92   
04:28:51.958 CD2B.1    stand  5    -260   440    0     80   10   
04:28:51.958 CD2B.0    stand  255  -230   530    0     80   94   
04:28:52.916 CD2B.1    stand  5    -260   440    0     80   94   
04:28:52.916 CD2B.0    stand  255  -230   530    0     80   94   
04:28:53.978 CD2B.2    stand  255  -290   310    66    80   228  
04:28:53.978 CD2B.0    stand  255  -230   530    0     80   228  
04:28:53.978 CD2B.1    stand  5    -260   440    0     80   94   
04:28:54.926 CD2B.0    stand  255  -230   530    0     80   94   
04:28:54.926 CD2B.2    stand  255  -270   270    74    80   263  
04:28:54.926 CD2B.1    stand  5    -260   440    0     80   170  
04:28:55.928 CD2B.2    walk   255  -260   180    70    80   260  
04:28:55.928 CD2B.0    stand  255  -230   530    0     80   351  
04:28:55.928 CD2B.1    stand  5    -260   440    0     80   94   
04:28:56.928 CD2B.1    stand  5    -260   440    0     80   0    
04:28:56.928 CD2B.2    walk   255  -230   110    81    80   331  
04:28:56.928 CD2B.0    stand  255  -230   530    0     80   420  
04:28:57.932 CD2B.1    stand  5    -260   440    0     80   94   
04:28:57.932 CD2B.2    walk   255  -170   10     77    80   439  
04:28:57.932 CD2B.0    stand  255  -230   530    0     80   523  
04:28:58.930 CD2B.0    stand  255  -230   530    0     80   0    
04:28:58.930 CD2B.1    stand  5    -260   440    0     80   94   
04:28:58.930 CD2B.2    walk   255  -150   -20    0     80   472  
04:28:59.992 CD2B.1    stand  5    -260   440    0     80   472  
04:28:59.992 CD2B.2    walk   255  -150   -20    0     80   472  
04:29:00.943 CD2B.1    stand  5    -260   440    0     80   472  
04:29:00.943 CD2B.2    stand  255  -150   -10    0     80   463  
04:29:01.844 CD2B.1    stand  5    -260   440    0     80   463  
04:29:01.844 CD2B.2    stand  255  -150   -10    0     80   463  
04:29:02.842 CD2B.1    stand  5    -260   440    0     80   463  
04:29:02.842 CD2B.2    stand  255  -150   -10    0     80   463  
04:29:03.844 CD2B.1    stand  5    -260   440    0     80   463  
04:29:03.844 CD2B.2    stand  None -150   -10    0     80   463  
04:29:04.896 CD2B.1    stand  5    -260   440    0     80   463  
04:29:05.854 CD2B.1    stand  5    -260   440    0     80   0    
04:29:06.856 CD2B.1    stand  5    -260   440    0     80   0    
04:29:07.856 CD2B.1    stand  5    -260   440    0     80   0    
04:29:08.860 CD2B.1    stand  5    -260   440    0     80   0    
04:29:09.866 CD2B.1    stand  5    -260   440    0     80   0    
04:29:10.866 CD2B.1    stand  5    -260   440    0     80   0    
04:29:11.766 CD2B.1    stand  5    -260   440    0     80   0    
04:29:12.762 CD2B.1    stand  5    -260   440    0     80   0    
04:29:13.766 CD2B.1    stand  5    -260   440    0     80   0    
04:29:14.776 CD2B.1    stand  5    -260   440    0     80   0    
04:29:15.765 CD2B.1    stand  5    -260   440    0     80   0    
04:29:16.764 CD2B.1    stand  5    -260   440    0     80   0    
04:29:17.768 CD2B.1    stand  5    -260   440    0     80   0    
04:29:18.769 CD2B.1    stand  5    -260   440    0     80   0    
04:29:19.770 CD2B.1    stand  5    -260   440    0     80   0    
04:29:20.772 CD2B.1    stand  5    -260   440    0     80   0    
04:29:21.772 CD2B.1    stand  5    -260   440    0     80   0    
04:29:22.776 CD2B.1    stand  5    -260   440    0     80   0    
04:29:23.714 CD2B.88   88     -    -      -      -     -    -    
04:29:24.674 CD2B.88   88     -    -      -      -     -    -    
04:29:25.673 CD2B.88   88     -    -      -      -     -    -    
04:29:50.724 CD2B.88   88     -    -      -      -     -    -    
04:30:01.666 CD2B.0    stand  None -170   10     81    80   439  
04:30:02.424 CD2B.0    stand  None -210   60     74    80   64   
04:30:03.436 CD2B.0    walk   None -280   160    67    80   122  
04:30:04.428 CD2B.0    walk   None -350   220    69    80   92   
04:30:05.430 CD2B.0    walk   None -330   180    83    80   44   
04:30:06.324 CD2B.0    walk   None -280   100    88    80   94   
04:30:07.324 CD2B.0    walk   None -330   20     66    80   94   
04:30:08.326 CD2B.0    walk   None -350   -20    0     80   44   
04:30:09.330 CD2B.0    walk   None -340   0      79    80   22   
04:30:10.330 CD2B.0    walk   None -270   0      79    80   70   
04:30:11.329 CD2B.0    walk   None -170   0      82    80   100  
04:30:12.334 CD2B.0    walk   None -140   -20    0     80   36   
04:30:13.336 CD2B.0    walk   None -140   -20    0     80   0    
04:30:14.334 CD2B.0    stand  None -150   -20    0     80   10   
04:30:15.336 CD2B.0    stand  None -150   -10    0     80   10   
04:30:16.336 CD2B.0    stand  None -150   -10    0     80   0    
04:30:17.238 CD2B.0    stand  None -150   -10    0     80   0    
04:30:18.288 CD2B.88   88     -    -      -      -     -    -    
04:30:19.244 CD2B.88   88     -    -      -      -     -    -    
04:30:20.246 CD2B.88   88     -    -      -      -     -    -    
04:30:22.248 CD2B.88   88     -    -      -      -     -    -    
04:30:54.112 CD2B.88   88     -    -      -      -     -    -    
04:31:25.736 CD2B.88   88     -    -      -      -     -    -    
04:31:46.333 CD2B.0    stand  None -190   60     72    80   80   
04:31:46.515 CD2B.0    walk   None -210   120    61    80   63   
04:31:47.524 CD2B.0    walk   None -230   200    72    80   82   
04:31:48.517 CD2B.0    walk   None -230   260    82    80   60   
04:31:49.549 CD2B.0    walk   None -210   280    92    80   28   
04:31:50.552 CD2B.0    walk   None -180   300    93    80   36   
04:31:51.556 CD2B.0    walk   None -160   280    112   80   28   
04:31:52.594 CD2B.0    walk   None -170   290    98    80   14   
04:31:53.554 CD2B.0    walk   None -150   290    81    80   20   
04:31:54.456 CD2B.0    walk   None -120   290    109   80   30   
04:31:55.452 CD2B.0    walk   None -110   300    89    80   14   
04:31:56.454 CD2B.0    walk   None -120   290    82    80   14   
04:31:57.457 CD2B.0    walk   None -110   250    79    80   41   
04:31:58.457 CD2B.0    walk   None -100   200    67    80   50   
04:31:59.456 CD2B.0    lying  None -120   180    71    80   28   
04:32:00.461 CD2B.0    lying  1    -90    210    67    80   42   
04:32:01.460 CD2B.0    lying  1    -90    220    71    80   10   
04:32:02.460 CD2B.0    lying  1    -90    200    60    80   20   
04:32:03.460 CD2B.0    lying  1    -80    220    70    80   22   
04:32:04.460 CD2B.0    lying  1    -90    210    68    80   14   
04:32:05.365 CD2B.0    lying  1    -90    210    60    80   0    
04:32:06.374 CD2B.0    lying  1    -110   200    66    80   22   
04:32:07.366 CD2B.0    lying  1    -80    210    72    80   31   
04:32:08.368 CD2B.0    lying  1    -80    210    70    80   0    
04:32:09.369 CD2B.0    lying  1    -90    210    66    80   10   
04:32:10.397 CD2B.0    lying  1    -90    210    60    80   0    
04:32:11.373 CD2B.0    lying  1    -100   210    63    80   10   
04:32:12.378 CD2B.0    lying  1    -100   200    65    80   10   
04:32:13.376 CD2B.0    lying  1    -120   210    59    80   22   
04:32:14.373 CD2B.0    lying  1    -130   200    65    80   14   
04:32:15.384 CD2B.0    lying  1    -130   200    61    80   0    
04:32:16.376 CD2B.0    lying  1    -140   200    59    80   10   
04:32:17.269 CD2B.0    lying  1    -110   210    59    80   31   
04:32:18.269 CD2B.0    lying  1    -120   210    64    80   10   
04:32:19.270 CD2B.0    lying  1    -110   210    58    80   10   
04:32:20.274 CD2B.0    lying  1    -100   210    53    80   10   
04:32:21.236 CD2B.0    lying  1    -80    210    51    80   20   
04:32:22.238 CD2B.0    lying  1    -40    220    49    80   41   
04:32:23.237 CD2B.0    lying  1    -40    230    50    80   10   
04:32:24.238 CD2B.0    lying  1    -60    220    54    80   22   
04:32:25.241 CD2B.0    lying  1    -30    230    54    80   31   
04:32:26.240 CD2B.0    lying  1    -40    230    51    80   10   
04:32:27.246 CD2B.0    lying  1    -40    230    0     80   0    
04:32:28.242 CD2B.0    lying  1    -40    230    0     80   0    
04:32:29.244 CD2B.0    lying  1    -50    230    0     80   10   
04:32:30.248 CD2B.0    lying  1    -50    230    0     80   0    
04:32:31.254 CD2B.0    lying  1    -50    230    0     80   0    
04:32:32.146 CD2B.0    lying  1    -50    230    0     80   0    
04:32:33.147 CD2B.0    lying  1    -50    230    0     80   0    
04:32:34.152 CD2B.0    lying  1    -50    230    0     80   0    
04:32:35.148 CD2B.0    lying  1    -50    230    0     80   0    
04:32:36.152 CD2B.0    lying  1    -50    230    0     80   0    
04:32:37.188 CD2B.0    lying  1    -50    230    0     80   0    
04:32:38.184 CD2B.0    lying  1    -50    230    0     80   0    
04:32:39.236 CD2B.0    lying  1    -50    230    0     80   0    
04:32:40.084 CD2B.0    lying  1    -50    230    0     80   0    
04:32:41.084 CD2B.0    lying  1    -50    230    0     80   0    
04:32:42.082 CD2B.0    lying  1    -50    230    0     80   0    
04:32:43.084 CD2B.0    lying  1    -70    220    56    80   22   
04:32:44.084 CD2B.0    lying  1    -40    230    0     80   31   
04:32:45.084 CD2B.0    lying  1    -40    220    0     80   10   
04:32:46.101 CD2B.0    lying  1    -170   190    49    80   133  
04:32:47.090 CD2B.0    lying  1    -180   170    0     80   22   
04:32:48.089 CD2B.0    lying  1    -30    230    0     80   161  
04:32:49.089 CD2B.0    lying  1    -30    230    0     80   0    
04:32:50.096 CD2B.0    lying  1    -30    230    0     80   0    
04:32:51.096 CD2B.0    lying  1    -30    230    0     80   0    
04:32:51.985 CD2B.0    lying  1    -40    230    0     80   10   
04:32:52.996 CD2B.0    lying  1    -40    230    0     80   0    
04:32:54.004 CD2B.0    lying  1    -40    230    0     80   0    
04:32:55.006 CD2B.0    lying  1    -40    230    0     80   0    
04:32:56.004 CD2B.0    lying  1    -40    230    0     80   0    
04:32:57.010 CD2B.0    lying  1    -40    230    0     80   0    
04:32:58.005 CD2B.0    lying  1    -40    230    0     80   0    

```

**汇总**: xray tick 488 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

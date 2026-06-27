# case-1797-0627-13261332 — 每 tick belief 时间线 (room fd00:0:3:411:3:200, TZ Asia/Shanghai)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
13:26:00 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 Empty      2   0     0.00  0.02  0.26  0.00  0.69  0.03
13:26:00 1797.0   179703105005  stand   56   NoReport stand              room -    Empty      2   0     0.00  0.03  0.15  0.00  0.79  0.04
13:26:01 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.02  0.52  0.00  0.40  0.01
13:26:01 1797.0   179703105005  stand   59   NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.52  0.00  0.40  0.01
13:26:02 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   1     0.00  0.02  0.70  0.00  0.18  0.02
13:26:02 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  2   1     0.00  0.02  0.70  0.00  0.18  0.02
13:26:03 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   2     0.00  0.02  0.79  0.00  0.07  0.02
13:26:03 1797.0   179703105005  stand   52   NoReport stand              room -    OpenFloor  2   2     0.00  0.02  0.79  0.00  0.07  0.02
13:26:04 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   3     0.00  0.02  0.83  0.00  0.03  0.02
13:26:04 1797.0   179703105005  stand   56   NoReport stand              room -    OpenFloor  2   3     0.00  0.02  0.83  0.00  0.03  0.02
13:26:05 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   4     0.00  0.02  0.84  0.00  0.02  0.02
13:26:05 1797.0   179703105005  stand   72   NoReport stand              room -    OpenFloor  2   4     0.00  0.02  0.84  0.00  0.02  0.02
13:26:06 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   5     0.00  0.02  0.85  0.00  0.01  0.02
13:26:06 1797.0   179703105005  stand   71   NoReport stand              room -    OpenFloor  2   5     0.00  0.02  0.85  0.00  0.01  0.02
13:26:07 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
13:26:07 1797.0   179703105005  stand   84   NoReport stand              room -    OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
13:26:08 1797.0   179703105005  stand   46   NoReport stand              room -    OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
13:26:08 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
13:26:09 1797.0   179703105005  stand   65   NoReport stand              room -    OpenFloor  2   9     0.01  0.05  0.68  0.01  0.03  0.04
13:26:09 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   9     0.01  0.05  0.68  0.01  0.03  0.04
13:26:10 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   9     0.01  0.05  0.68  0.01  0.03  0.04
13:26:11 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   10    0.00  0.03  0.79  0.01  0.02  0.02
13:26:11 1797.0   179703105005  stand   66   NoReport stand              room -    OpenFloor  2   10    0.01  0.04  0.72  0.01  0.02  0.04
13:26:11 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   10    0.00  0.03  0.83  0.00  0.02  0.02
13:26:11 1797.0   179703105005  stand   64   NoReport stand              room -    OpenFloor  2   10    0.00  0.03  0.80  0.01  0.02  0.02
13:26:12 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   10    0.00  0.03  0.80  0.01  0.02  0.02
13:26:13 1797.0   179703105005  stand   80   NoReport stand              room -    OpenFloor  2   12    0.01  0.04  0.73  0.01  0.02  0.04
13:26:13 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   12    0.01  0.04  0.73  0.01  0.02  0.04
13:26:13 1797.0   179703105005  stand   48   NoReport stand              room -    OpenFloor  2   12    0.00  0.03  0.81  0.01  0.02  0.02
13:26:13 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   12    0.00  0.03  0.81  0.01  0.02  0.02
13:26:14 1797.0   179703105005  stand   42   NoReport stand              room -    OpenFloor  2   13    0.00  0.02  0.83  0.00  0.01  0.02
13:26:14 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   13    0.00  0.02  0.83  0.00  0.01  0.02
13:26:15 1797.0   179703105005  stand   43   NoReport stand              room -    OpenFloor  2   14    0.00  0.02  0.84  0.00  0.01  0.02
13:26:15 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   14    0.00  0.02  0.84  0.00  0.01  0.02
13:26:16 1797.0   179703105005  stand   43   NoReport stand              room -    OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
13:26:16 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
13:26:17 1797.0   179703105005  stand   52   NoReport stand              room -    OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
13:26:17 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
13:26:18 1797.0   179703105005  stand   54   NoReport stand              room -    OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
13:26:18 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
13:26:19 1797.0   179703105005  stand   56   NoReport stand              room -    OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
13:26:19 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
13:26:20 1797.0   179703105005  stand   61   NoReport stand              room -    OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
13:26:20 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
13:26:21 1797.0   179703105005  stand   65   NoReport stand              room -    OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:26:21 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:26:22 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:26:23 1797.0   179703105005  stand   61   NoReport stand              room -    OpenFloor  2   22    0.00  0.04  0.74  0.00  0.02  0.04
13:26:23 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   22    0.00  0.04  0.74  0.00  0.02  0.04
13:26:23 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   22    0.00  0.03  0.81  0.00  0.02  0.02
13:26:23 1797.0   179703105005  stand   65   NoReport stand              room -    OpenFloor  2   22    0.00  0.03  0.81  0.00  0.02  0.02
13:26:24 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   23    0.00  0.02  0.84  0.00  0.01  0.02
13:26:24 1797.0   179703105005  walk    76   NoReport walk               room -    OpenFloor  2   23    0.00  0.02  0.84  0.00  0.01  0.02
13:26:25 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   24    0.00  0.02  0.84  0.00  0.01  0.02
13:26:25 1797.0   179703105005  walk    89   NoReport walk               room -    OpenFloor  2   24    0.00  0.02  0.84  0.00  0.01  0.02
13:26:26 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
13:26:26 1797.0   179703105005  walk    100  NoReport walk               room -    OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
13:26:27 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:26:27 1797.0   179703105005  walk    84   NoReport walk               room -    OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:26:28 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:26:29 1797.0   179703105005  walk    76   NoReport walk               room -    OpenFloor  2   28    0.00  0.04  0.74  0.00  0.02  0.04
13:26:29 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   28    0.00  0.04  0.74  0.00  0.02  0.04
13:26:29 1797.0   179703105005  walk    105  NoReport walk               room -    OpenFloor  2   16    0.00  0.03  0.81  0.00  0.02  0.02
13:26:29 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   16    0.00  0.03  0.81  0.00  0.02  0.02
13:26:30 1797.0   179703105005  walk    67   NoReport walk               room -    OpenFloor  2   17    0.00  0.02  0.84  0.00  0.01  0.02
13:26:30 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   17    0.00  0.02  0.84  0.00  0.01  0.02
13:26:31 1797.0   179703105005  walk    92   NoReport walk               room -    OpenFloor  2   18    0.00  0.02  0.84  0.00  0.01  0.02
13:26:31 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   18    0.00  0.02  0.84  0.00  0.01  0.02
13:26:32 1797.0   179703105005  walk    82   NoReport walk               room -    OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
13:26:32 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
13:26:33 1797.0   179703105005  walk    71   NoReport walk               room -    OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:26:33 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:26:34 1797.0   179703105005  walk    74   NoReport walk               room -    OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
13:26:34 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
13:26:35 1797.0   179703105005  walk    77   NoReport walk               room -    OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
13:26:35 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
13:26:36 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
13:26:36 1797.0   179703105005  walk    80   NoReport walk               room -    OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
13:26:37 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
13:26:38 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   24    0.00  0.02  0.85  0.00  0.01  0.02
13:26:38 1797.0   179703105005  walk    51   NoReport walk               room -    OpenFloor  2   24    0.00  0.02  0.85  0.00  0.01  0.02
13:26:38 1797.0   179703105005  walk    81   NoReport walk               room -    OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
13:26:38 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
13:26:39 1797.0   179703105005  walk    67   NoReport walk               room -    OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:26:39 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:26:40 1797.0   179703105005  walk    35   NoReport walk               room -    OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
13:26:40 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
13:26:41 1797.0   179703105005  walk    54   NoReport walk               room -    OpenFloor  2   16    0.00  0.01  0.92  0.00  0.01  0.01
13:26:41 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   16    0.00  0.01  0.92  0.00  0.01  0.01
13:26:42 1797.0   179703105005  walk    78   NoReport walk               room -    OpenFloor  2   17    0.00  0.01  0.96  0.00  0.00  0.01
13:26:42 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   17    0.00  0.01  0.96  0.00  0.00  0.01
13:26:43 1797.0   179703105005  walk    68   NoReport walk               room -    OpenFloor  2   18    0.00  0.01  0.97  0.00  0.00  0.01
13:26:43 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   18    0.00  0.01  0.97  0.00  0.00  0.01
13:26:44 1797.0   179703105005  walk    62   NoReport walk               room -    OpenFloor  2   19    0.00  0.01  0.97  0.00  0.00  0.01
13:26:44 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   19    0.00  0.01  0.97  0.00  0.00  0.01
13:26:45 1797.0   179703105005  walk    56   NoReport walk               room -    OpenFloor  2   20    0.00  0.01  0.97  0.00  0.00  0.01
13:26:45 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   20    0.00  0.01  0.97  0.00  0.00  0.01
13:26:46 1797.0   179703105005  walk    60   NoReport walk               room -    OpenFloor  2   21    0.00  0.01  0.97  0.00  0.00  0.01
13:26:46 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   21    0.00  0.01  0.97  0.00  0.00  0.01
13:26:47 1797.0   179703105005  walk    65   NoReport walk               room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:47 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:48 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:48 1797.0   179703105005  walk    58   NoReport walk               room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:49 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:49 1797.0   179703105005  walk    45   NoReport walk               room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:50 1797.0   179703105005  walk    105  NoReport walk               room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:50 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:51 1797.0   179703105005  walk    88   NoReport walk               room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:51 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:52 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:52 1797.0   179703105005  walk    70   NoReport walk               room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:53 1797.0   179703105005  walk    65   NoReport walk               room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:53 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:54 1797.0   179703105005  walk    66   NoReport walk               room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:54 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:55 1797.0   179703105005  walk    69   NoReport walk               room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:55 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
13:26:56 1797.0   179703105005  walk    51   NoReport walk               room -    OpenFloor  2   14    0.00  0.01  0.97  0.00  0.00  0.01
13:26:56 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   14    0.00  0.01  0.97  0.00  0.00  0.01
13:26:57 1797.0   179703105005  walk    85   NoReport walk               room -    OpenFloor  2   15    0.00  0.01  0.97  0.00  0.00  0.01
13:26:57 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   15    0.00  0.01  0.97  0.00  0.00  0.01
13:26:58 1797.0   179703105005  walk    55   NoReport walk               room -    OpenFloor  2   16    0.00  0.01  0.97  0.00  0.00  0.01
13:26:58 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   16    0.00  0.01  0.97  0.00  0.00  0.01
13:26:59 1797.0   179703105005  walk    74   NoReport walk               room -    OpenFloor  2   17    0.00  0.01  0.97  0.00  0.00  0.01
13:26:59 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   17    0.00  0.01  0.97  0.00  0.00  0.01
13:27:00 1797.0   179703105005  walk    80   NoReport walk               room -    OpenFloor  2   18    0.00  0.01  0.97  0.00  0.00  0.01
13:27:00 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   18    0.00  0.01  0.97  0.00  0.00  0.01
13:27:01 1797.0   179703105005  walk    75   NoReport walk               room -    OpenFloor  2   19    0.00  0.01  0.97  0.00  0.00  0.01
13:27:01 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   19    0.00  0.01  0.97  0.00  0.00  0.01
13:27:02 1797.0   179703105005  walk    59   NoReport walk               room -    OpenFloor  2   20    0.00  0.01  0.97  0.00  0.00  0.01
13:27:02 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   20    0.00  0.01  0.97  0.00  0.00  0.01
13:27:03 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   21    0.00  0.01  0.97  0.00  0.00  0.01
13:27:03 1797.0   179703105005  walk    61   NoReport walk               room -    OpenFloor  2   21    0.00  0.01  0.97  0.00  0.00  0.01
13:27:04 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   22    0.00  0.01  0.97  0.00  0.00  0.01
13:27:04 1797.0   179703105005  walk    63   NoReport walk               room -    OpenFloor  2   22    0.00  0.01  0.97  0.00  0.00  0.01
13:27:05 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   23    0.00  0.01  0.97  0.00  0.00  0.01
13:27:05 1797.0   179703105005  walk    68   NoReport walk               room -    OpenFloor  2   23    0.00  0.01  0.97  0.00  0.00  0.01
13:27:06 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   24    0.00  0.01  0.97  0.00  0.00  0.01
13:27:06 1797.0   179703105005  walk    46   NoReport walk               room -    OpenFloor  2   24    0.00  0.01  0.97  0.00  0.00  0.01
13:27:07 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
13:27:07 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   25    0.00  0.01  0.97  0.00  0.00  0.01
13:27:08 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   26    0.00  0.01  0.97  0.00  0.00  0.01
13:27:08 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:27:09 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
13:27:09 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   27    0.00  0.01  0.97  0.00  0.00  0.01
13:27:10 1797.E1  179712600752  -       0    NoReport ExitRoom(rdr)      trk  0.96 OpenFloor  2   28    0.00  0.02  0.88  0.00  0.00  0.02
13:27:10 1797.1   179712600752  stand   0    NoReport stand              trk  0.96 OpenFloor  2   28    0.00  0.01  0.96  0.00  0.00  0.01
13:27:10 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  2   28    0.00  0.04  0.74  0.00  0.02  0.04
13:27:11 1797.E   -             -       0    NoReport np=1               room -    OpenFloor  2   28    0.00  0.04  0.74  0.00  0.02  0.04
13:27:11 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  2   30    0.01  0.05  0.68  0.01  0.03  0.04
13:27:12 1797.0   179703105005  stand   82   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.65  0.01  0.04  0.03
13:27:13 1797.0   179703105005  stand   79   NoReport stand              room -    OpenFloor  1   0     0.00  0.03  0.78  0.01  0.03  0.02
13:27:14 1797.0   179703105005  stand   76   NoReport stand              room -    OpenFloor  1   0     0.01  0.04  0.70  0.01  0.03  0.03
13:27:15 1797.0   179703105005  walk    99   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.66  0.01  0.04  0.03
13:27:16 1797.0   179703105005  walk    73   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.64  0.01  0.04  0.03
13:27:17 1797.0   179703105005  walk    143  NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.63  0.02  0.05  0.03
13:27:18 1797.0   179703105005  walk    69   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
13:27:19 1797.0   179703105005  walk    81   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
13:27:20 1797.0   179703105005  walk    78   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
13:27:21 1797.0   179703105005  walk    98   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
13:27:22 1797.0   179703105005  walk    132  NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
13:27:23 1797.0   179703105005  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
13:27:24 1797.0   179703105005  walk    85   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
13:27:25 1797.0   179703105005  walk    98   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
13:27:26 1797.0   179703105005  walk    126  NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
13:27:27 1797.0   179703105005  walk    95   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:28 1797.0   179703105005  walk    91   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:29 1797.0   179703105005  walk    120  NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:30 1797.0   179703105005  walk    49   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:31 1797.0   179703105005  walk    80   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:32 1797.0   179703105005  stand   128  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:33 1797.0   179703105005  stand   115  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:34 1797.0   179703105005  stand   114  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:35 1797.0   179703105005  stand   98   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:36 1797.0   179703105005  stand   71   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:37 1797.0   179703105005  stand   99   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:38 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:27:38 1797.0   179703105005  stand   27   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:27:39 1797.0   179703105005  stand   80   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:40 1797.0   179703105005  stand   86   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:41 1797.0   179703105005  stand   90   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:42 1797.0   179703105005  stand   117  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:43 1797.0   179703105005  stand   121  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:44 1797.0   179703105005  stand   96   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:45 1797.0   179703105005  stand   82   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:46 1797.0   179703105005  stand   78   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:47 1797.0   179703105005  stand   96   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:48 1797.0   179703105005  stand   94   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:49 1797.0   179703105005  stand   84   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:50 1797.0   179703105005  stand   99   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:27:51 1797.0   179703105005  stand   70   NoReport stand              room -    OpenFloor  1   28    0.00  0.05  0.61  0.02  0.05  0.03
13:27:52 1797.0   179703105005  stand   130  NoReport stand              room -    OpenFloor  1   29    0.00  0.05  0.61  0.02  0.05  0.03
13:27:53 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   29    0.00  0.05  0.61  0.02  0.05  0.03
13:27:54 1797.0   179703105005  stand   92   NoReport stand              room -    OpenFloor  1   31    0.00  0.05  0.61  0.02  0.05  0.03
13:27:55 1797.0   179703105005  stand   107  NoReport stand              room -    OpenFloor  1   31    0.00  0.05  0.61  0.02  0.05  0.03
13:27:55 1797.0   179703105005  stand   100  NoReport stand              room -    OpenFloor  1   32    0.00  0.05  0.61  0.02  0.05  0.03
13:27:56 1797.0   179703105005  stand   87   NoReport stand              room -    OpenFloor  1   33    0.00  0.05  0.61  0.02  0.05  0.03
13:27:57 1797.0   179703105005  stand   108  NoReport stand              room -    OpenFloor  1   34    0.00  0.05  0.61  0.02  0.05  0.03
13:27:58 1797.0   179703105005  stand   67   NoReport stand              room -    OpenFloor  1   35    0.00  0.05  0.61  0.02  0.05  0.03
13:27:59 1797.0   179703105005  stand   63   NoReport stand              room -    OpenFloor  1   36    0.01  0.05  0.61  0.02  0.05  0.03
13:28:00 1797.0   179703105005  stand   72   NoReport stand              room -    OpenFloor  1   37    0.01  0.05  0.61  0.02  0.05  0.03
13:28:01 1797.0   179703105005  stand   91   NoReport stand              room -    OpenFloor  1   38    0.00  0.05  0.61  0.02  0.05  0.03
13:28:02 1797.0   179703105005  stand   84   NoReport stand              room -    OpenFloor  1   39    0.00  0.05  0.61  0.02  0.05  0.03
13:28:03 1797.0   179703105005  stand   69   NoReport stand              room -    OpenFloor  1   40    0.01  0.05  0.61  0.02  0.05  0.03
13:28:04 1797.0   179703105005  stand   106  NoReport stand              room -    OpenFloor  1   41    0.00  0.05  0.61  0.02  0.05  0.03
13:28:05 1797.0   179703105005  stand   74   NoReport stand              room -    OpenFloor  1   42    0.01  0.05  0.61  0.02  0.05  0.03
13:28:06 1797.0   179703105005  stand   79   NoReport stand              room -    OpenFloor  1   43    0.01  0.05  0.61  0.02  0.05  0.03
13:28:07 1797.0   179703105005  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:08 1797.0   179703105005  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:09 1797.0   179703105005  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:10 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:11 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:12 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:13 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:14 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:15 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:16 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:17 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:18 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:19 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:20 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:21 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:21 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:22 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:23 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:24 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:25 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:26 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:27 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:28 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:29 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:29 1797.0   179703105005  stand   86   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:30 1797.0   179703105005  stand   85   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:31 1797.0   179703105005  stand   87   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:32 1797.0   179703105005  walk    98   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:33 1797.0   179703105005  walk    113  NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:34 1797.0   179703105005  walk    94   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:35 1797.0   179703105005  walk    97   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:36 1797.0   179703105005  walk    110  NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:37 1797.0   179703105005  walk    83   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:38 1797.0   179703105005  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:39 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:40 1797.0   179703105005  stand   139  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:41 1797.0   179703105005  stand   50   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:42 1797.0   179703105005  stand   91   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:43 1797.0   179703105005  stand   142  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:44 1797.0   179703105005  stand   145  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:45 1797.0   179703105005  stand   104  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:46 1797.0   179703105005  stand   127  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:47 1797.0   179703105005  stand   39   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:48 1797.0   179703105005  stand   119  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:49 1797.0   179703105005  stand   111  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:50 1797.0   179703105005  stand   66   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:51 1797.0   179703105005  stand   82   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:52 1797.0   179703105005  stand   76   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:28:53 1797.0   179703105005  stand   100  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:54 1797.0   179703105005  stand   115  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:55 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:56 1797.0   179703105005  stand   139  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:57 1797.0   179703105005  stand   131  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:58 1797.0   179703105005  stand   106  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:28:59 1797.0   179703105005  stand   80   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:00 1797.0   179703105005  stand   74   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:01 1797.0   179703105005  stand   55   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:29:02 1797.0   179703105005  stand   66   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:29:03 1797.0   179703105005  stand   78   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:29:04 1797.0   179703105005  walk    49   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:05 1797.0   179703105005  walk    61   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:06 1797.0   179703105005  walk    77   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:07 1797.0   179703105005  walk    77   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:08 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:09 1797.0   179703105005  walk    71   NoReport walk               room -    OpenFloor  1   0     0.00  0.07  0.51  0.03  0.08  0.03
13:29:09 1797.0   179703105005  walk    84   NoReport walk               room -    OpenFloor  1   0     0.00  0.06  0.55  0.03  0.07  0.03
13:29:10 1797.0   179703105005  walk    98   NoReport walk               room -    OpenFloor  1   0     0.00  0.06  0.58  0.03  0.07  0.03
13:29:11 1797.0   179703105005  walk    136  NoReport walk               room -    OpenFloor  1   0     0.00  0.06  0.59  0.02  0.06  0.03
13:29:12 1797.0   179703105005  walk    107  NoReport walk               room -    OpenFloor  1   0     0.00  0.06  0.60  0.02  0.06  0.03
13:29:13 1797.0   179703105005  walk    99   NoReport walk               room -    OpenFloor  1   0     0.00  0.06  0.61  0.02  0.06  0.03
13:29:14 1797.0   179703105005  walk    102  NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
13:29:15 1797.0   179703105005  walk    66   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:16 1797.0   179703105005  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:17 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:18 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:29:19 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:29:20 1797.0   179703105005  stand   129  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:21 1797.0   179703105005  stand   126  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:22 1797.0   179703105005  stand   51   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:29:23 1797.0   179703105005  stand   82   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:24 1797.0   179703105005  stand   59   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:29:25 1797.0   179703105005  stand   77   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:29:26 1797.0   179703105005  stand   87   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:27 1797.0   179703105005  stand   125  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:28 1797.0   179703105005  stand   135  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:29 1797.0   179703105005  stand   130  NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:30 1797.0   179703105005  stand   92   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:31 1797.0   179703105005  stand   69   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:32 1797.0   179703105005  stand   75   NoReport stand              room -    OpenFloor  1   27    0.01  0.05  0.61  0.02  0.05  0.03
13:29:33 1797.0   179703105005  walk    73   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:29:34 1797.0   179703105005  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.03  0.76  0.01  0.03  0.02
13:29:35 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.03  0.82  0.01  0.02  0.02
13:29:36 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
13:29:37 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
13:29:38 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:29:39 1797.0   179703105005  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:29:39 1797.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:29:40 1797.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
13:29:40 1797.88  -             88      -    NoReport no-target(88)      room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  0.99
13:29:41 1797.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:42 1797.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:53 1797.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:29:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:26 1797.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:57 1797.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:30:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:31:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:31:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:31:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:31:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:31:04 1797.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:31:05 1797.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:31:05 1797.0   179703105005  stand   94   NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
13:31:05 1797.0   179703105005  stand   60   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
13:31:06 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
13:31:07 1797.0   179703105005  stand   70   NoReport stand              trk  1.00 OpenFloor  1   2     0.00  0.04  0.54  0.00  0.28  0.03
13:31:07 1797.0   179703105005  walk    60   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.58  0.01  0.15  0.03
13:31:08 1797.0   179703105005  walk    49   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.59  0.01  0.11  0.03
13:31:09 1797.0   179703105005  walk    99   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.60  0.02  0.09  0.03
13:31:10 1797.0   179703105005  walk    77   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.08  0.03
13:31:11 1797.0   179703105005  walk    32   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.07  0.03
13:31:12 1797.0   179703105005  walk    108  NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
13:31:13 1797.0   179703105005  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
13:31:14 1797.0   179703105005  stand   83   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
13:31:15 1797.0   179703105005  stand   76   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.06  0.03
13:31:16 1797.0   179703105005  stand   65   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:31:17 1797.0   179703105005  stand   106  NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:18 1797.0   179703105005  stand   103  NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:19 1797.0   179703105005  stand   97   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:20 1797.0   179703105005  stand   111  NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:21 1797.0   179703105005  stand   74   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:22 1797.0   179703105005  stand   72   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:31:23 1797.0   179703105005  stand   63   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:31:24 1797.0   179703105005  stand   118  NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:25 1797.0   179703105005  stand   75   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:31:26 1797.0   179703105005  stand   77   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:31:27 1797.0   179703105005  stand   70   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:31:28 1797.0   179703105005  stand   77   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:31:29 1797.0   179703105005  stand   83   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:30 1797.0   179703105005  stand   80   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:31 1797.0   179703105005  stand   105  NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:32 1797.0   179703105005  stand   76   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:33 1797.0   179703105005  stand   93   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:34 1797.0   179703105005  walk    86   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:35 1797.0   179703105005  walk    86   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:36 1797.0   179703105005  walk    85   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:37 1797.0   179703105005  walk    96   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:38 1797.0   179703105005  walk    136  NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:39 1797.0   179703105005  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:40 1797.0   179703105005  walk    85   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:41 1797.0   179703105005  walk    91   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:41 1797.0   179703105005  walk    90   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:42 1797.0   179703105005  walk    79   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:43 1797.0   179703105005  walk    87   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:44 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:45 1797.0   179703105005  walk    81   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:45 1797.0   179703105005  walk    80   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:46 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:47 1797.0   179703105005  walk    79   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:47 1797.0   179703105005  walk    77   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:48 1797.0   179703105005  walk    86   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:49 1797.0   179703105005  walk    79   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:50 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:51 1797.0   179703105005  walk    84   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:51 1797.0   179703105005  walk    59   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:52 1797.0   179703105005  walk    101  NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:53 1797.0   179703105005  walk    86   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:54 1797.0   179703105005  walk    89   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:55 1797.0   179703105005  walk    71   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:56 1797.0   179703105005  walk    79   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:57 1797.0   179703105005  walk    85   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:58 1797.0   179703105005  walk    85   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:31:59 1797.0   179703105005  walk    102  NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:00 1797.0   179703105005  walk    78   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:01 1797.0   179703105005  walk    69   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:02 1797.0   179703105005  walk    90   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:03 1797.0   179703105005  walk    70   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:04 1797.0   179703105005  walk    68   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:05 1797.0   179703105005  walk    86   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:06 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:07 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:08 1797.0   179703105005  walk    80   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:08 1797.0   179703105005  walk    103  NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:08 1797.0   179703105005  walk    66   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:09 1797.0   179703105005  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:10 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:11 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:12 1797.0   179703105005  walk    63   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:12 1797.0   179703105005  walk    74   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:13 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:14 1797.0   179703105005  walk    70   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:14 1797.0   179703105005  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:14 1797.0   179703105005  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:32:15 1797.0   179703105005  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.76  0.01  0.03  0.02
13:32:16 1797.0   179703105005  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.82  0.01  0.02  0.02
13:32:17 1797.0   179703105005  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.04  0.72  0.01  0.02  0.04
13:32:18 1797.0   179703105005  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.67  0.01  0.03  0.04
13:32:19 1797.0   179703105005  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.65  0.01  0.04  0.03
13:32:20 1797.0   179703105005  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.63  0.01  0.05  0.03
13:32:21 1797.0   179703105005  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.77  0.01  0.03  0.02
13:32:21 1797.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  1   0     0.01  0.03  0.77  0.01  0.03  0.02
13:32:21 1797.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   0     0.01  0.03  0.77  0.01  0.03  0.02
13:32:21 1797.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   0     0.00  0.03  0.82  0.01  0.02  0.02
13:32:22 1797.88  -             88      -    NoReport no-target(88)      room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  0.99
13:32:23 -.-      -             -       -    NoReport (no frame, held)   room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  0.99
13:32:24 -.-      -             -       -    NoReport (no frame, held)   room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  0.99
13:32:25 1797.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:32:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:32:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:32:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:32:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:32:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:32:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
13:32:32 1797.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
13:26:00.752 1797.1    stand  None 10     110    0     80        
13:26:00.752 1797.0    stand  255  -30    -30    56    80   145  
13:26:01.749 1797.1    stand  None 10     110    0     80   145  
13:26:01.749 1797.0    stand  255  -20    -30    59    80   143  
13:26:02.749 1797.1    stand  None 10     110    0     80   143  
13:26:02.749 1797.0    stand  255  -30    -40    0     80   155  
13:26:03.682 1797.1    stand  None 10     110    0     80   155  
13:26:03.682 1797.0    stand  255  -30    -30    52    80   145  
13:26:04.683 1797.1    stand  None 10     110    0     80   145  
13:26:04.683 1797.0    stand  255  -20    -20    56    80   133  
13:26:05.684 1797.1    stand  None -20    110    0     80   130  
13:26:05.684 1797.0    stand  255  -30    0      72    80   110  
13:26:06.691 1797.1    stand  None -20    100    0     80   100  
13:26:06.691 1797.0    stand  255  -30    0      71    80   100  
13:26:07.835 1797.1    stand  None -20    90     0     80   90   
13:26:07.835 1797.0    stand  255  -40    -10    84    80   101  
13:26:08.741 1797.0    stand  255  -50    0      46    80   14   
13:26:08.741 1797.1    stand  None -30    100    0     80   101  
13:26:09.971 1797.0    stand  255  -40    -10    65    80   110  
13:26:09.971 1797.1    stand  None -50    100    0     80   110  
13:26:11.719 1797.1    stand  None -40    110    0     80   14   
13:26:11.719 1797.0    stand  255  -40    -10    66    80   120  
13:26:11.721 1797.1    stand  None -20    120    0     80   131  
13:26:11.721 1797.0    stand  255  -40    10     64    80   111  
13:26:13.245 1797.0    stand  255  -50    0      80    80   14   
13:26:13.245 1797.1    stand  None -30    90     0     80   92   
13:26:13.691 1797.0    stand  255  -50    -10    48    80   101  
13:26:13.691 1797.1    stand  None -70    110    0     80   121  
13:26:14.597 1797.0    stand  255  -50    -10    42    80   121  
13:26:14.597 1797.1    stand  None -80    120    0     80   133  
13:26:15.602 1797.0    stand  255  -50    -20    43    80   143  
13:26:15.602 1797.1    stand  None -30    100    0     80   121  
13:26:16.593 1797.0    stand  255  -40    -10    43    80   110  
13:26:16.593 1797.1    stand  None -70    100    0     80   114  
13:26:17.595 1797.0    stand  255  -40    -20    52    80   123  
13:26:17.595 1797.1    stand  None -80    100    0     80   126  
13:26:18.595 1797.0    stand  255  -40    -20    54    80   126  
13:26:18.595 1797.1    stand  None -50    90     0     80   110  
13:26:19.556 1797.0    stand  255  -30    -20    56    80   111  
13:26:19.556 1797.1    stand  None -50    90     0     80   111  
13:26:20.554 1797.0    stand  255  -30    -30    61    80   121  
13:26:20.554 1797.1    stand  None -40    110    0     80   140  
13:26:21.554 1797.0    stand  255  -40    -30    65    80   140  
13:26:21.554 1797.1    stand  None -40    110    0     80   140  
13:26:23.393 1797.0    stand  255  -30    -10    61    80   120  
13:26:23.393 1797.1    stand  None -50    110    0     80   121  
13:26:23.693 1797.1    stand  None -50    110    0     80   0    
13:26:23.693 1797.0    stand  255  -30    -10    65    80   121  
13:26:24.569 1797.1    stand  None -50    110    0     80   121  
13:26:24.569 1797.0    walk   255  -20    0      76    80   114  
13:26:25.568 1797.1    stand  None -30    120    0     80   120  
13:26:25.568 1797.0    walk   255  -20    10     89    80   110  
13:26:26.570 1797.1    stand  None -30    120    0     80   110  
13:26:26.570 1797.0    walk   255  -10    10     100   80   111  
13:26:27.567 1797.1    stand  None -30    120    0     80   111  
13:26:27.567 1797.0    walk   255  0      20     84    80   104  
13:26:29.250 1797.0    walk   255  0      20     76    80   0    
13:26:29.250 1797.1    stand  None -30    120    0     80   104  
13:26:29.600 1797.0    walk   255  0      40     105   80   85   
13:26:29.600 1797.1    stand  None -30    120    0     80   85   
13:26:30.581 1797.0    walk   255  0      20     67    80   104  
13:26:30.581 1797.1    stand  None -30    120    0     80   104  
13:26:31.455 1797.0    walk   255  0      20     92    80   104  
13:26:31.455 1797.1    stand  None -30    120    0     80   104  
13:26:32.468 1797.0    walk   255  0      0      82    80   123  
13:26:32.468 1797.1    stand  None -30    120    0     80   123  
13:26:33.463 1797.0    walk   255  0      10     71    80   114  
13:26:33.463 1797.1    stand  None -30    120    0     80   114  
13:26:34.461 1797.0    walk   255  0      10     74    80   114  
13:26:34.461 1797.1    stand  None -30    120    0     80   114  
13:26:35.473 1797.0    walk   255  0      10     77    80   114  
13:26:35.473 1797.1    stand  None -30    130    0     80   123  
13:26:36.485 1797.1    stand  None -20    120    0     80   14   
13:26:36.485 1797.0    walk   255  0      20     80    80   101  
13:26:38.233 1797.1    stand  None -20    120    0     80   101  
13:26:38.233 1797.0    walk   255  -20    50     51    80   70   
13:26:38.539 1797.0    walk   255  -20    40     81    80   10   
13:26:38.539 1797.1    stand  None -20    120    0     80   80   
13:26:39.485 1797.0    walk   255  -10    20     67    80   100  
13:26:39.485 1797.1    stand  None -20    120    0     80   100  
13:26:40.382 1797.0    walk   255  0      0      35    80   121  
13:26:40.382 1797.1    stand  None -60    120    0     80   134  
13:26:41.397 1797.0    walk   255  0      0      54    80   134  
13:26:41.397 1797.1    stand  None 20     120    0     80   121  
13:26:42.383 1797.0    walk   255  -10    0      78    80   123  
13:26:42.383 1797.1    stand  None 20     120    0     80   123  
13:26:43.385 1797.0    walk   255  -10    0      68    80   123  
13:26:43.385 1797.1    stand  None 20     120    0     80   123  
13:26:44.385 1797.0    walk   255  -20    0      62    80   126  
13:26:44.385 1797.1    stand  None 20     120    0     80   126  
13:26:45.386 1797.0    walk   255  -20    0      56    80   126  
13:26:45.386 1797.1    stand  None 20     120    0     80   126  
13:26:46.399 1797.0    walk   255  -10    50     60    80   76   
13:26:46.399 1797.1    stand  None 20     120    0     80   76   
13:26:47.388 1797.0    walk   255  -20    60     65    80   72   
13:26:47.388 1797.1    stand  None 20     120    0     80   72   
13:26:48.390 1797.1    stand  None 20     120    0     80   0    
13:26:48.390 1797.0    walk   255  -20    40     58    80   89   
13:26:49.399 1797.1    stand  None 20     120    0     80   89   
13:26:49.399 1797.0    walk   255  -30    40     45    80   94   
13:26:50.395 1797.0    walk   255  -20    50     105   80   14   
13:26:50.395 1797.1    stand  None 20     120    0     80   80   
13:26:51.392 1797.0    walk   255  -20    40     88    80   89   
13:26:51.392 1797.1    stand  None 20     120    0     80   89   
13:26:52.293 1797.1    stand  None 20     120    0     80   0    
13:26:52.293 1797.0    walk   255  -20    30     70    80   98   
13:26:53.305 1797.0    walk   255  -30    20     65    80   14   
13:26:53.305 1797.1    stand  None 20     120    0     80   111  
13:26:54.290 1797.0    walk   255  -40    10     66    80   125  
13:26:54.290 1797.1    stand  None 20     120    0     80   125  
13:26:55.299 1797.0    walk   255  -30    0      69    80   130  
13:26:55.299 1797.1    stand  None 20     120    0     80   130  
13:26:56.303 1797.0    walk   255  -30    0      51    80   130  
13:26:56.303 1797.1    stand  None 20     120    0     80   130  
13:26:57.344 1797.0    walk   255  -10    0      85    80   123  
13:26:57.344 1797.1    stand  None 20     120    0     80   123  
13:26:58.323 1797.0    walk   255  -10    0      55    80   123  
13:26:58.323 1797.1    stand  None 20     120    0     80   123  
13:26:59.297 1797.0    walk   255  0      -10    74    80   131  
13:26:59.297 1797.1    stand  None 20     120    0     80   131  
13:27:00.322 1797.0    walk   255  10     0      80    80   120  
13:27:00.322 1797.1    stand  None 20     120    0     80   120  
13:27:01.299 1797.0    walk   255  30     -10    75    80   130  
13:27:01.299 1797.1    stand  None 20     120    0     80   130  
13:27:02.303 1797.0    walk   255  30     0      59    80   120  
13:27:02.303 1797.1    stand  None 20     120    0     80   120  
13:27:03.304 1797.1    stand  None 20     120    0     80   0    
13:27:03.304 1797.0    walk   255  30     0      61    80   120  
13:27:04.203 1797.1    stand  None 20     120    0     80   120  
13:27:04.203 1797.0    walk   255  30     20     63    80   100  
13:27:05.197 1797.1    stand  None 20     120    0     80   100  
13:27:05.197 1797.0    walk   255  50     60     68    80   67   
13:27:06.194 1797.1    stand  None 20     120    0     80   67   
13:27:06.194 1797.0    walk   255  10     90     46    80   31   
13:27:07.202 1797.0    stand  255  -10    90     0     80   20   
13:27:07.202 1797.1    stand  None 20     120    0     80   42   
13:27:08.157 1797.1    stand  None 20     120    0     80   0    
13:27:08.157 1797.0    stand  255  -10    90     0     80   42   
13:27:09.157 1797.0    stand  255  0      100    0     80   14   
13:27:09.157 1797.1    stand  None 20     120    0     80   28   
13:27:10.389 1797.1    stand  None 20     120    0     80   0    
13:27:10.389 1797.0    stand  255  0      100    0     80   28   
13:27:11.411 1797.0    stand  255  0      100    0     80   0    
13:27:12.171 1797.0    stand  255  0      110    82    80   10   
13:27:13.174 1797.0    stand  255  30     100    79    80   31   
13:27:14.175 1797.0    stand  255  30     80     76    80   20   
13:27:15.173 1797.0    walk   255  50     50     99    80   36   
13:27:16.186 1797.0    walk   255  50     10     73    80   40   
13:27:17.175 1797.0    walk   255  70     0      143   80   22   
13:27:18.178 1797.0    walk   255  60     0      69    80   10   
13:27:19.072 1797.0    walk   255  70     0      81    80   10   
13:27:20.073 1797.0    walk   255  70     10     78    80   10   
13:27:21.075 1797.0    walk   255  70     30     98    80   20   
13:27:22.087 1797.0    walk   255  90     10     132   80   28   
13:27:23.076 1797.0    walk   255  30     50     0     80   72   
13:27:24.100 1797.0    walk   255  60     20     85    80   42   
13:27:25.100 1797.0    walk   255  60     20     98    80   0    
13:27:26.099 1797.0    walk   255  70     20     126   80   10   
13:27:27.106 1797.0    walk   255  70     20     95    80   0    
13:27:28.063 1797.0    walk   255  70     20     91    80   0    
13:27:29.003 1797.0    walk   255  80     20     120   80   10   
13:27:30.001 1797.0    walk   255  60     30     49    80   22   
13:27:31.002 1797.0    walk   255  70     20     80    80   14   
13:27:32.007 1797.0    stand  255  80     20     128   80   10   
13:27:33.004 1797.0    stand  255  70     30     115   80   14   
13:27:34.010 1797.0    stand  255  80     30     114   80   10   
13:27:35.006 1797.0    stand  255  70     20     98    80   14   
13:27:36.007 1797.0    stand  255  70     20     71    80   0    
13:27:37.009 1797.0    stand  255  70     20     99    80   0    
13:27:38.012 1797.0    stand  255  60     30     0     80   14   
13:27:38.909 1797.0    stand  255  60     30     27    80   0    
13:27:39.910 1797.0    stand  255  60     20     80    80   10   
13:27:40.910 1797.0    stand  255  60     40     86    80   20   
13:27:41.912 1797.0    stand  255  70     30     90    80   14   
13:27:42.918 1797.0    stand  255  70     30     117   80   0    
13:27:43.920 1797.0    stand  255  70     20     121   80   10   
13:27:44.915 1797.0    stand  255  70     30     96    80   10   
13:27:45.921 1797.0    stand  255  70     20     82    80   10   
13:27:46.917 1797.0    stand  255  60     30     78    80   14   
13:27:47.926 1797.0    stand  255  50     40     96    80   14   
13:27:48.919 1797.0    stand  255  40     40     94    80   10   
13:27:49.921 1797.0    stand  255  50     30     84    80   14   
13:27:50.816 1797.0    stand  255  40     40     99    80   14   
13:27:51.816 1797.0    stand  255  60     20     70    80   28   
13:27:52.817 1797.0    stand  255  80     20     130   80   20   
13:27:54.732 1797.0    stand  255  70     30     92    80   14   
13:27:55.035 1797.0    stand  255  70     20     107   80   10   
13:27:55.776 1797.0    stand  255  60     30     100   80   14   
13:27:56.782 1797.0    stand  255  70     30     87    80   10   
13:27:57.786 1797.0    stand  255  70     20     108   80   10   
13:27:58.778 1797.0    stand  255  60     20     67    80   10   
13:27:59.780 1797.0    stand  255  50     20     63    80   10   
13:28:00.806 1797.0    stand  255  40     30     72    80   14   
13:28:01.784 1797.0    stand  255  30     40     91    80   14   
13:28:02.785 1797.0    stand  255  40     30     84    80   14   
13:28:03.785 1797.0    stand  255  40     30     69    80   0    
13:28:04.794 1797.0    stand  255  40     40     106   80   10   
13:28:05.786 1797.0    stand  255  40     30     74    80   10   
13:28:06.791 1797.0    stand  255  40     50     79    80   20   
13:28:07.677 1797.0    walk   255  20     100    0     80   53   
13:28:08.685 1797.0    walk   255  10     110    0     80   14   
13:28:09.983 1797.0    walk   255  0      90     0     80   22   
13:28:10.680 1797.0    stand  255  0      70     0     80   20   
13:28:11.730 1797.0    stand  255  0      70     0     80   0    
13:28:12.731 1797.0    stand  255  0      70     0     80   0    
13:28:13.639 1797.0    stand  255  0      70     0     80   0    
13:28:14.630 1797.0    stand  255  0      80     0     80   10   
13:28:15.640 1797.0    stand  255  0      80     0     80   0    
13:28:16.632 1797.0    stand  255  0      80     0     80   0    
13:28:17.634 1797.0    stand  255  0      80     0     80   0    
13:28:18.634 1797.0    stand  255  0      80     0     80   0    
13:28:19.636 1797.0    stand  255  0      80     0     80   0    
13:28:21.249 1797.0    stand  255  0      80     0     80   0    
13:28:21.639 1797.0    stand  255  0      80     0     80   0    
13:28:22.643 1797.0    stand  255  0      80     0     80   0    
13:28:23.639 1797.0    stand  255  0      80     0     80   0    
13:28:24.644 1797.0    stand  255  0      80     0     80   0    
13:28:25.534 1797.0    stand  255  0      80     0     80   0    
13:28:26.547 1797.0    stand  255  0      80     0     80   0    
13:28:28.215 1797.0    stand  255  0      80     0     80   0    
13:28:29.748 1797.0    stand  255  0      80     0     80   0    
13:28:29.750 1797.0    stand  255  40     90     86    80   41   
13:28:30.550 1797.0    stand  255  50     50     85    80   41   
13:28:31.551 1797.0    stand  255  50     30     87    80   20   
13:28:32.552 1797.0    walk   255  60     30     98    80   10   
13:28:33.553 1797.0    walk   255  70     30     113   80   10   
13:28:34.553 1797.0    walk   255  80     20     94    80   14   
13:28:35.556 1797.0    walk   255  80     20     97    80   0    
13:28:36.447 1797.0    walk   255  80     20     110   80   0    
13:28:37.446 1797.0    walk   255  90     30     83    80   14   
13:28:38.450 1797.0    walk   255  60     20     0     80   31   
13:28:39.448 1797.0    stand  255  50     20     0     80   10   
13:28:40.452 1797.0    stand  255  80     20     139   80   30   
13:28:41.459 1797.0    stand  255  80     30     50    80   10   
13:28:42.452 1797.0    stand  255  80     20     91    80   10   
13:28:43.453 1797.0    stand  255  70     30     142   80   14   
13:28:44.381 1797.0    stand  255  80     20     145   80   14   
13:28:45.389 1797.0    stand  255  80     20     104   80   0    
13:28:46.384 1797.0    stand  255  70     20     127   80   10   
13:28:47.388 1797.0    stand  255  70     20     39    80   0    
13:28:48.385 1797.0    stand  255  80     20     119   80   10   
13:28:49.391 1797.0    stand  255  80     10     111   80   10   
13:28:50.388 1797.0    stand  255  70     20     66    80   14   
13:28:51.389 1797.0    stand  255  70     20     82    80   0    
13:28:52.389 1797.0    stand  255  70     20     76    80   0    
13:28:53.391 1797.0    stand  255  70     20     100   80   0    
13:28:54.398 1797.0    stand  255  80     20     115   80   10   
13:28:55.655 1797.0    stand  255  110    0      0     80   36   
13:28:56.287 1797.0    stand  255  90     10     139   80   22   
13:28:57.286 1797.0    stand  255  80     10     131   80   10   
13:28:58.293 1797.0    stand  255  90     10     106   80   10   
13:28:59.289 1797.0    stand  255  60     10     80    80   30   
13:29:00.255 1797.0    stand  255  60     0      74    80   10   
13:29:01.257 1797.0    stand  255  90     -40    55    80   50   
13:29:02.257 1797.0    stand  255  90     -40    66    80   0    
13:29:03.261 1797.0    stand  255  70     -10    78    80   36   
13:29:04.260 1797.0    walk   255  60     0      49    80   14   
13:29:05.287 1797.0    walk   255  60     0      61    80   0    
13:29:06.266 1797.0    walk   255  80     0      77    80   20   
13:29:07.263 1797.0    walk   255  80     0      77    80   0    
13:29:09.791 1797.0    walk   255  90     10     71    80   14   
13:29:09.792 1797.0    walk   255  80     10     84    80   10   
13:29:10.268 1797.0    walk   255  80     10     98    80   0    
13:29:11.267 1797.0    walk   255  80     10     136   80   0    
13:29:12.161 1797.0    walk   255  90     10     107   80   10   
13:29:13.167 1797.0    walk   255  90     10     99    80   0    
13:29:14.161 1797.0    walk   255  80     20     102   80   14   
13:29:15.165 1797.0    walk   255  80     20     66    80   0    
13:29:16.755 1797.0    walk   255  70     10     0     80   14   
13:29:17.231 1797.0    stand  255  70     20     0     80   10   
13:29:18.201 1797.0    stand  255  80     30     0     80   14   
13:29:19.176 1797.0    stand  255  70     30     0     80   10   
13:29:20.176 1797.0    stand  255  80     30     129   80   10   
13:29:21.175 1797.0    stand  255  80     30     126   80   0    
13:29:22.176 1797.0    stand  255  40     50     51    80   44   
13:29:23.075 1797.0    stand  255  60     20     82    80   36   
13:29:24.073 1797.0    stand  255  50     20     59    80   10   
13:29:25.077 1797.0    stand  255  60     20     77    80   10   
13:29:26.077 1797.0    stand  255  70     30     87    80   14   
13:29:27.079 1797.0    stand  255  70     30     125   80   0    
13:29:28.084 1797.0    stand  255  70     20     135   80   10   
13:29:29.095 1797.0    stand  255  70     30     130   80   10   
13:29:30.081 1797.0    stand  255  70     20     92    80   10   
13:29:31.085 1797.0    stand  255  60     20     69    80   10   
13:29:32.007 1797.0    stand  255  50     40     75    80   22   
13:29:33.004 1797.0    walk   255  40     90     73    80   50   
13:29:34.003 1797.0    walk   255  30     110    0     80   22   
13:29:35.005 1797.0    stand  255  30     120    0     80   10   
13:29:36.005 1797.0    stand  255  30     120    0     80   0    
13:29:37.007 1797.0    stand  255  20     120    0     80   10   
13:29:38.020 1797.0    stand  255  20     130    0     80   10   
13:29:39.008 1797.0    stand  None 20     130    0     80   0    
13:29:40.321 1797.88   88     -    -      -      -     -    -    
13:29:41.048 1797.88   88     -    -      -      -     -    -    
13:29:42.003 1797.88   88     -    -      -      -     -    -    
13:29:53.846 1797.88   88     -    -      -      -     -    -    
13:30:26.013 1797.88   88     -    -      -      -     -    -    
13:30:57.353 1797.88   88     -    -      -      -     -    -    
13:31:05.005 1797.0    stand  None 50     120    94    80   31   
13:31:05.330 1797.0    stand  None 40     90     60    80   31   
13:31:07.248 1797.0    stand  None 40     70     70    80   20   
13:31:07.559 1797.0    walk   None 40     50     60    80   20   
13:31:08.324 1797.0    walk   None 30     40     49    80   14   
13:31:09.222 1797.0    walk   None 20     0      99    80   41   
13:31:10.217 1797.0    walk   None 0      0      77    80   20   
13:31:11.236 1797.0    walk   None 0      -10    32    80   10   
13:31:12.229 1797.0    walk   None 0      -20    108   80   10   
13:31:13.221 1797.0    stand  None 0      -20    0     80   0    
13:31:14.222 1797.0    stand  None 0      -20    83    80   0    
13:31:15.224 1797.0    stand  None -10    -10    76    80   14   
13:31:16.231 1797.0    stand  None -20    -10    65    80   10   
13:31:17.224 1797.0    stand  None -40    -30    106   80   28   
13:31:18.234 1797.0    stand  None -40    -20    103   80   10   
13:31:19.243 1797.0    stand  None -40    -10    97    80   10   
13:31:20.227 1797.0    stand  None -40    -30    111   80   20   
13:31:21.146 1797.0    stand  None -30    -20    74    80   14   
13:31:22.123 1797.0    stand  None -20    -20    72    80   10   
13:31:23.130 1797.0    stand  None -10    -20    63    80   10   
13:31:24.125 1797.0    stand  None -10    -30    118   80   10   
13:31:25.173 1797.0    stand  None -20    -20    75    80   14   
13:31:26.132 1797.0    stand  None -20    -30    77    80   10   
13:31:27.129 1797.0    stand  None -30    -30    70    80   10   
13:31:28.137 1797.0    stand  None -40    -20    77    80   14   
13:31:29.133 1797.0    stand  None -40    0      83    80   20   
13:31:30.131 1797.0    stand  None -30    10     80    80   14   
13:31:31.139 1797.0    stand  None -50    30     105   80   28   
13:31:32.136 1797.0    stand  None -60    0      76    80   31   
13:31:33.764 1797.0    stand  None -30    30     93    80   42   
13:31:34.072 1797.0    walk   None 0      50     86    80   36   
13:31:35.032 1797.0    walk   None -20    40     86    80   22   
13:31:36.033 1797.0    walk   None -60    20     85    80   44   
13:31:37.042 1797.0    walk   None -60    30     96    80   10   
13:31:38.036 1797.0    walk   None -70    40     136   80   14   
13:31:39.036 1797.0    walk   None -80    70     0     80   31   
13:31:40.067 1797.0    walk   None -60    20     85    80   53   
13:31:41.021 1797.0    walk   None -70    30     91    80   14   
13:31:41.988 1797.0    walk   None -80    30     90    80   10   
13:31:42.968 1797.0    walk   None -90    30     79    80   10   
13:31:43.975 1797.0    walk   None -80    20     87    80   14   
13:31:45.147 1797.0    walk   None -80    20     81    80   0    
13:31:45.984 1797.0    walk   None -90    30     80    80   14   
13:31:47.021 1797.0    walk   None -90    30     79    80   0    
13:31:47.971 1797.0    walk   None -90    40     77    80   10   
13:31:48.977 1797.0    walk   None -90    40     86    80   0    
13:31:49.970 1797.0    walk   None -80    40     79    80   10   
13:31:51.016 1797.0    walk   None -80    30     84    80   10   
13:31:51.971 1797.0    walk   None -90    50     59    80   22   
13:31:52.873 1797.0    walk   None -70    20     101   80   36   
13:31:53.887 1797.0    walk   None -70    30     86    80   10   
13:31:54.867 1797.0    walk   None -90    60     89    80   36   
13:31:55.876 1797.0    walk   None -70    40     71    80   28   
13:31:56.889 1797.0    walk   None -70    20     79    80   20   
13:31:57.875 1797.0    walk   None -80    30     85    80   14   
13:31:58.878 1797.0    walk   None -80    30     85    80   0    
13:31:59.893 1797.0    walk   None -70    40     102   80   14   
13:32:00.883 1797.0    walk   None -70    40     78    80   0    
13:32:01.897 1797.0    walk   None -70    20     69    80   20   
13:32:02.881 1797.0    walk   None -60    0      90    80   22   
13:32:03.826 1797.0    walk   None -70    10     70    80   14   
13:32:04.786 1797.0    walk   None -60    0      68    80   14   
13:32:05.802 1797.0    walk   None -50    -20    86    80   22   
13:32:08.787 1797.0    walk   None -50    -20    80    80   0    
13:32:08.789 1797.0    walk   None -50    -20    103   80   0    
13:32:08.790 1797.0    walk   None -50    -20    66    80   0    
13:32:09.789 1797.0    walk   None -40    -20    0     80   10   
13:32:12.268 1797.0    walk   None -40    -10    63    80   10   
13:32:12.270 1797.0    walk   None -20    0      74    80   22   
13:32:14.218 1797.0    walk   None 0      30     70    80   36   
13:32:14.220 1797.0    walk   None 0      80     0     80   50   
13:32:14.689 1797.0    walk   None 10     90     0     80   14   
13:32:15.687 1797.0    stand  None 10     100    0     80   10   
13:32:16.691 1797.0    stand  None 10     110    0     80   10   
13:32:17.696 1797.0    stand  None 10     110    0     80   0    
13:32:18.690 1797.0    stand  None 10     110    0     80   0    
13:32:19.691 1797.0    stand  None 10     110    0     80   0    
13:32:20.691 1797.0    stand  None 10     110    0     80   0    
13:32:21.381 1797.0    stand  None 20     110    0     80   10   
13:32:21.793 1797.88   88     -    -      -      -     -    -    
13:32:22.660 1797.88   88     -    -      -      -     -    -    
13:32:25.274 1797.88   88     -    -      -      -     -    -    
13:32:32.549 1797.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 388 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

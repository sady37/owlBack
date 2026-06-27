# case-B17F-0626-13321344 — 每 tick belief 时间线 (room fd00:0:3:111:2:300, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
13:32:00 B17F.1   B17F13200777  walk    72   NoReport walk               trk  0.50 Empty      2   0     0.00  0.02  0.26  0.00  0.69  0.03
13:32:00 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.50 Empty      2   0     0.00  0.03  0.15  0.00  0.79  0.04
13:32:01 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.51 Empty      2   1     0.00  0.03  0.40  0.00  0.53  0.01
13:32:01 B17F.1   B17F13200777  walk    80   NoReport walk               trk  0.51 Empty      2   1     0.00  0.02  0.51  0.00  0.40  0.01
13:32:02 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.52 OpenFloor  2   2     0.00  0.03  0.63  0.00  0.26  0.01
13:32:02 B17F.1   B17F13200777  walk    89   NoReport walk               trk  0.74 OpenFloor  2   2     0.00  0.02  0.68  0.00  0.17  0.02
13:32:03 B17F.1   B17F13200777  walk    77   NoReport walk               trk  0.74 OpenFloor  2   3     0.00  0.02  0.71  0.00  0.06  0.02
13:32:03 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.53 OpenFloor  2   3     0.00  0.02  0.76  0.00  0.11  0.02
13:32:04 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   4     0.00  0.02  0.82  0.00  0.04  0.02
13:32:04 B17F.1   B17F13200777  walk    86   NoReport walk               trk  0.75 OpenFloor  2   4     0.00  0.02  0.68  0.00  0.03  0.02
13:32:05 B17F.1   B17F13200777  walk    92   NoReport walk               trk  0.75 OpenFloor  2   5     0.00  0.02  0.63  0.00  0.01  0.02
13:32:05 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   5     0.00  0.02  0.84  0.00  0.02  0.02
13:32:06 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   6     0.00  0.02  0.84  0.00  0.01  0.02
13:32:06 B17F.1   B17F13200777  walk    91   NoReport walk               trk  0.75 OpenFloor  2   6     0.00  0.02  0.70  0.01  0.01  0.02
13:32:07 B17F.1   B17F13200777  walk    86   NoReport walk               trk  0.75 OpenFloor  2   7     0.00  0.02  0.65  0.00  0.01  0.02
13:32:07 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
13:32:08 B17F.1   B17F13200777  walk    0    NoReport walk               trk  0.75 OpenFloor  2   8     0.00  0.02  0.60  0.00  0.01  0.02
13:32:08 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   8     0.00  0.02  0.85  0.00  0.01  0.02
13:32:09 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   9     0.00  0.02  0.85  0.00  0.01  0.02
13:32:09 B17F.1   B17F13200777  sit     76   NoReport sit                trk  0.75 OpenFloor  2   9     0.00  0.01  0.25  0.00  0.01  0.01
13:32:10 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   9     0.00  0.02  0.85  0.00  0.01  0.02
13:32:10 B17F.1   B17F13200777  sit     72   NoReport sit                trk  0.75 OpenFloor  2   9     0.00  0.00  0.03  0.00  0.00  0.00
13:32:11 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   10    0.00  0.02  0.85  0.00  0.01  0.02
13:32:11 B17F.1   B17F13200777  sit     65   NoReport sit                trk  0.75 OpenFloor  2   10    0.00  0.00  0.01  0.00  0.00  0.00
13:32:12 B17F.1   B17F13200777  sit     0    NoReport sit                trk  0.75 OpenFloor  2   11    0.00  0.00  0.02  0.00  0.00  0.00
13:32:12 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   11    0.00  0.02  0.85  0.00  0.01  0.02
13:32:13 B17F.1   B17F13200777  sit     0    NoReport sit                trk  0.75 OpenFloor  2   12    0.00  0.00  0.03  0.01  0.00  0.01
13:32:13 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   12    0.00  0.02  0.85  0.00  0.01  0.02
13:32:14 B17F.1   B17F13200777  sit     0    NoReport sit                trk  0.75 OpenFloor  2   13    0.00  0.00  0.04  0.01  0.00  0.01
13:32:14 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
13:32:15 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
13:32:15 B17F.1   B17F13200777  sit     72   NoReport sit                trk  0.75 OpenFloor  2   14    0.00  0.00  0.04  0.01  0.00  0.01
13:32:16 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
13:32:16 B17F.1   B17F13200777  sit     84   NoReport sit                trk  0.75 OpenFloor  2   15    0.00  0.00  0.02  0.00  0.00  0.00
13:32:17 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
13:32:17 B17F.1   B17F13200777  sit     76   NoReport sit                trk  0.75 OpenFloor  2   16    0.00  0.00  0.02  0.00  0.00  0.00
13:32:18 B17F.1   B17F13200777  sit     75   NoReport sit                trk  0.75 OpenFloor  2   17    0.00  0.00  0.01  0.00  0.00  0.00
13:32:18 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
13:32:19 B17F.1   B17F13200777  sit     65   NoReport sit                trk  0.75 OpenFloor  2   18    0.00  0.00  0.01  0.00  0.00  0.00
13:32:19 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
13:32:20 B17F.1   B17F13200777  sit     88   NoReport sit                trk  0.75 OpenFloor  2   19    0.00  0.00  0.03  0.00  0.00  0.00
13:32:20 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
13:32:21 B17F.1   B17F13200777  sit     80   NoReport sit                trk  0.75 OpenFloor  2   20    0.00  0.00  0.04  0.00  0.00  0.00
13:32:21 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:32:22 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
13:32:22 B17F.1   B17F13200777  sit     69   NoReport sit                trk  0.75 OpenFloor  2   21    0.00  0.00  0.02  0.00  0.00  0.00
13:32:23 B17F.1   B17F13200777  sit     67   NoReport sit                trk  0.75 OpenFloor  2   22    0.00  0.00  0.02  0.00  0.00  0.00
13:32:23 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
13:32:24 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
13:32:24 B17F.1   B17F13200777  sit     71   NoReport sit                trk  0.75 OpenFloor  2   23    0.00  0.00  0.02  0.01  0.00  0.00
13:32:25 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   24    0.00  0.02  0.85  0.00  0.01  0.02
13:32:25 B17F.1   B17F13200777  sit     78   NoReport sit                trk  0.75 OpenFloor  2   24    0.00  0.00  0.01  0.00  0.00  0.00
13:32:26 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
13:32:26 B17F.1   B17F13200777  sit     70   NoReport sit                trk  0.75 OpenFloor  2   25    0.00  0.00  0.01  0.00  0.00  0.00
13:32:27 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:32:27 B17F.1   B17F13200777  sit     84   NoReport sit                trk  0.75 OpenFloor  2   26    0.00  0.00  0.03  0.00  0.00  0.00
13:32:28 B17F.1   B17F13200777  sit     77   NoReport sit                trk  0.75 OpenFloor  2   27    0.00  0.00  0.01  0.00  0.00  0.00
13:32:28 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
13:32:29 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   28    0.00  0.02  0.85  0.00  0.01  0.02
13:32:29 B17F.1   B17F13200777  sit     73   NoReport sit                trk  0.75 OpenFloor  2   28    0.00  0.00  0.01  0.00  0.00  0.00
13:32:30 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   29    0.00  0.02  0.85  0.00  0.01  0.02
13:32:30 B17F.1   B17F13200777  sit     72   NoReport sit                trk  0.75 OpenFloor  2   29    0.00  0.00  0.01  0.00  0.00  0.00
13:32:31 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   30    0.00  0.02  0.85  0.00  0.01  0.02
13:32:31 B17F.1   B17F13200777  sit     70   NoReport sit                trk  0.75 OpenFloor  2   30    0.00  0.00  0.01  0.00  0.00  0.00
13:32:32 B17F.1   B17F13200777  sit     67   NoReport sit                trk  0.75 OpenFloor  2   31    0.00  0.00  0.01  0.00  0.00  0.00
13:32:32 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   31    0.00  0.02  0.85  0.00  0.01  0.02
13:32:33 B17F.1   B17F13200777  sit     67   NoReport sit                trk  0.75 OpenFloor  2   32    0.00  0.00  0.01  0.00  0.00  0.00
13:32:33 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   32    0.00  0.02  0.85  0.00  0.01  0.02
13:32:34 B17F.1   B17F13200777  sit     81   NoReport sit                trk  0.75 OpenFloor  2   33    0.00  0.00  0.03  0.00  0.00  0.00
13:32:34 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   33    0.00  0.02  0.85  0.00  0.01  0.02
13:32:35 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   34    0.00  0.02  0.85  0.00  0.01  0.02
13:32:35 B17F.1   B17F13200777  sit     74   NoReport sit                trk  0.75 OpenFloor  2   34    0.00  0.00  0.04  0.00  0.00  0.00
13:32:36 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   35    0.00  0.02  0.85  0.00  0.01  0.02
13:32:36 B17F.1   B17F13200777  walk    85   NoReport walk               trk  0.75 OpenFloor  2   35    0.00  0.01  0.13  0.02  0.00  0.02
13:32:37 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   36    0.00  0.02  0.85  0.00  0.01  0.02
13:32:37 B17F.1   B17F13200777  walk    68   NoReport walk               trk  0.75 OpenFloor  2   36    0.00  0.01  0.45  0.03  0.01  0.02
13:32:38 B17F.1   B17F13200777  walk    91   NoReport walk               trk  0.75 OpenFloor  2   37    0.00  0.02  0.69  0.02  0.01  0.02
13:32:38 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   37    0.00  0.02  0.85  0.00  0.01  0.02
13:32:39 B17F.1   B17F13200777  walk    103  NoReport walk               trk  0.75 OpenFloor  2   38    0.00  0.02  0.79  0.01  0.01  0.02
13:32:39 B17F.0   B17F03200777  stand   0    NoReport stand              trk  0.54 OpenFloor  2   38    0.00  0.02  0.85  0.00  0.01  0.02
13:32:40 B17F.E   -             -       0    NoReport np=3               room -    OpenFloor  2   38    0.00  0.02  0.85  0.00  0.01  0.02
13:32:40 B17F.1   B17F13200777  walk    0    NoReport walk               trk  1.00 OpenFloor  3   39    0.00  0.01  0.91  0.00  0.01  0.01
13:32:40 B17F.2   B17F23240568  stand   70   NoReport stand              trk  0.50 OpenFloor  3   39    0.00  0.02  0.26  0.00  0.69  0.03
13:32:40 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   39    0.00  0.01  0.92  0.00  0.01  0.01
13:32:41 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   40    0.00  0.01  0.93  0.00  0.00  0.01
13:32:41 B17F.1   B17F13200777  walk    114  NoReport walk               trk  1.00 OpenFloor  3   40    0.00  0.01  0.93  0.00  0.00  0.01
13:32:41 B17F.2   B17F23240568  stand   0    NoReport stand              trk  0.51 OpenFloor  3   40    0.00  0.02  0.64  0.00  0.25  0.01
13:32:42 B17F.1   B17F13200777  walk    108  NoReport walk               trk  1.00 OpenFloor  3   41    0.00  0.01  0.93  0.00  0.00  0.01
13:32:42 B17F.2   B17F23240568  stand   0    NoReport stand              trk  0.52 OpenFloor  3   41    0.00  0.01  0.74  0.00  0.05  0.01
13:32:42 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   41    0.00  0.01  0.93  0.00  0.00  0.01
13:32:43 B17F.1   B17F13200777  walk    0    NoReport walk               trk  1.00 OpenFloor  3   42    0.00  0.01  0.93  0.00  0.00  0.01
13:32:43 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   42    0.00  0.01  0.93  0.00  0.00  0.01
13:32:43 B17F.2   B17F23240568  stand   83   NoReport stand              trk  0.53 OpenFloor  3   42    0.00  0.01  0.71  0.00  0.01  0.01
13:32:44 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   43    0.00  0.01  0.93  0.00  0.00  0.01
13:32:44 B17F.2   B17F23240568  stand   0    NoReport stand              trk  0.54 OpenFloor  3   43    0.00  0.01  0.77  0.00  0.00  0.01
13:32:44 B17F.1   B17F13200777  walk    118  NoReport walk               trk  1.00 OpenFloor  3   43    0.00  0.01  0.93  0.00  0.00  0.01
13:32:45 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   44    0.00  0.01  0.93  0.00  0.00  0.01
13:32:45 B17F.2   B17F23240568  stand   74   NoReport stand              trk  0.55 OpenFloor  3   44    0.00  0.01  0.81  0.00  0.00  0.01
13:32:45 B17F.1   B17F13200777  walk    0    NoReport walk               trk  1.00 OpenFloor  3   44    0.00  0.01  0.93  0.00  0.00  0.01
13:32:46 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.74  0.00  0.00  0.01
13:32:46 B17F.1   B17F13200777  walk    106  NoReport walk               trk  1.00 OpenFloor  3   45    0.00  0.01  0.93  0.00  0.00  0.01
13:32:46 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.93  0.00  0.00  0.01
13:32:47 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
13:32:47 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.79  0.00  0.00  0.01
13:32:47 B17F.1   B17F13200777  walk    89   NoReport walk               trk  1.00 OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
13:32:48 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   47    0.00  0.01  0.93  0.00  0.00  0.01
13:32:48 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 OpenFloor  3   47    0.00  0.01  0.82  0.00  0.00  0.01
13:32:48 B17F.1   B17F13200777  walk    82   NoReport walk               trk  1.00 OpenFloor  3   47    0.00  0.01  0.93  0.00  0.00  0.01
13:32:49 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   8     0.00  0.01  0.93  0.00  0.00  0.01
13:32:49 B17F.1   B17F13200777  walk    0    NoReport walk               trk  1.00 OpenFloor  3   8     0.00  0.01  0.93  0.00  0.00  0.01
13:32:49 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 OpenFloor  3   8     0.00  0.01  0.75  0.00  0.00  0.01
13:32:50 B17F.1   B17F13200777  walk    0    NoReport walk               trk  1.00 OpenFloor  3   9     0.00  0.01  0.93  0.00  0.00  0.01
13:32:50 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.01  0.93  0.00  0.00  0.01
13:32:50 B17F.2   B17F23240568  stand   80   NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.01  0.68  0.00  0.00  0.01
13:32:51 B17F.2   B17F23240568  stand   55   NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.62  0.00  0.00  0.01
13:32:51 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:32:51 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   0     0.00  0.02  0.85  0.00  0.00  0.02
13:32:52 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 Sit        3   0     0.00  0.04  0.36  0.00  0.01  0.04
13:32:52 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:32:52 B17F.2   B17F23240568  walk    36   NoReport walk               trk  1.00 Sit        3   0     0.00  0.01  0.58  0.00  0.00  0.01
13:32:53 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 Sit        3   0     0.00  0.01  0.08  0.01  0.01  0.01
13:32:53 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:32:53 B17F.2   B17F23240568  walk    59   NoReport walk               trk  1.00 Sit        3   0     0.00  0.01  0.79  0.00  0.00  0.01
13:32:54 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   0     0.00  0.00  0.03  0.01  0.00  0.01
13:32:54 B17F.2   B17F23240568  sit     51   NoReport sit                trk  1.00 OpenFloor  3   0     0.00  0.01  0.20  0.00  0.01  0.01
13:32:54 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.90  0.00  0.01  0.01
13:32:54 B17F.E   -             -       0    NoReport Walking(rdr)       room -    OpenFloor  3   0     0.00  0.01  0.90  0.00  0.01  0.01
13:32:54 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:32:54 B17F.2   B17F23240568  sitgnd  56   NoReport sitgnd             trk  1.00 Sit        3   0     0.00  0.02  0.13  0.02  0.01  0.02
13:32:54 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 Sit        3   0     0.00  0.00  0.02  0.01  0.00  0.00
13:32:55 B17F.2   B17F23240568  sit     65   NoReport sit                trk  1.00 OpenFloor  3   0     0.00  0.00  0.00  0.00  0.00  0.00
13:32:55 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   0     0.00  0.00  0.02  0.01  0.00  0.00
13:32:55 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.93  0.00  0.00  0.01
13:32:56 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   7     0.00  0.00  0.02  0.01  0.00  0.00
13:32:56 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 OpenFloor  3   7     0.00  0.00  0.00  0.00  0.00  0.00
13:32:56 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   7     0.00  0.01  0.93  0.00  0.00  0.01
13:32:57 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   8     0.00  0.01  0.93  0.00  0.00  0.01
13:32:57 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 OpenFloor  3   8     0.00  0.00  0.00  0.00  0.00  0.00
13:32:57 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   8     0.00  0.00  0.02  0.01  0.00  0.00
13:32:58 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 OpenFloor  3   9     0.00  0.00  0.00  0.00  0.00  0.00
13:32:58 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   9     0.00  0.00  0.02  0.01  0.00  0.00
13:32:58 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.01  0.93  0.00  0.00  0.01
13:32:59 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   10    0.00  0.00  0.02  0.01  0.00  0.00
13:32:59 B17F.2   B17F23240568  sit     81   NoReport sit                trk  1.00 OpenFloor  3   10    0.00  0.00  0.01  0.00  0.00  0.00
13:32:59 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   10    0.00  0.01  0.93  0.00  0.00  0.01
13:33:00 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   11    0.00  0.01  0.93  0.00  0.00  0.01
13:33:00 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   11    0.00  0.00  0.02  0.01  0.00  0.00
13:33:00 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 OpenFloor  3   11    0.00  0.00  0.01  0.00  0.00  0.00
13:33:01 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   12    0.00  0.00  0.02  0.01  0.00  0.00
13:33:01 B17F.2   B17F23240568  sit     47   NoReport sit                trk  1.00 OpenFloor  3   12    0.00  0.00  0.00  0.00  0.00  0.00
13:33:01 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   12    0.00  0.01  0.93  0.00  0.00  0.01
13:33:02 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   13    0.00  0.00  0.02  0.01  0.00  0.00
13:33:02 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 OpenFloor  3   13    0.00  0.00  0.00  0.00  0.00  0.00
13:33:02 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   13    0.00  0.01  0.93  0.00  0.00  0.01
13:33:03 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 OpenFloor  3   14    0.00  0.00  0.00  0.00  0.00  0.00
13:33:03 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.01  0.93  0.00  0.00  0.01
13:33:03 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   14    0.00  0.00  0.02  0.01  0.00  0.00
13:33:04 B17F.2   B17F23240568  sit     59   NoReport sit                trk  1.00 OpenFloor  3   15    0.00  0.00  0.02  0.00  0.00  0.00
13:33:04 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   15    0.00  0.00  0.02  0.01  0.00  0.00
13:33:04 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.01  0.93  0.00  0.00  0.01
13:33:05 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.01  0.93  0.00  0.00  0.01
13:33:05 B17F.2   B17F23240568  sit     47   NoReport sit                trk  1.00 OpenFloor  3   16    0.00  0.00  0.02  0.01  0.00  0.00
13:33:05 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   16    0.00  0.00  0.02  0.01  0.00  0.00
13:33:06 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.01  0.93  0.00  0.00  0.01
13:33:06 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   17    0.00  0.00  0.02  0.01  0.00  0.00
13:33:06 B17F.2   B17F23240568  sit     81   NoReport sit                trk  1.00 OpenFloor  3   17    0.00  0.00  0.00  0.00  0.00  0.00
13:33:07 B17F.2   B17F23240568  sit     71   NoReport sit                trk  1.00 OpenFloor  3   18    0.00  0.00  0.00  0.00  0.00  0.00
13:33:07 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.01  0.93  0.00  0.00  0.01
13:33:07 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   18    0.00  0.00  0.02  0.01  0.00  0.00
13:33:08 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.93  0.00  0.00  0.01
13:33:08 B17F.2   B17F23240568  sit     80   NoReport sit                trk  1.00 OpenFloor  3   19    0.00  0.00  0.01  0.00  0.00  0.00
13:33:08 B17F.1   B17F13200777  sit     0    NoReport sit                trk  1.00 OpenFloor  3   19    0.00  0.00  0.02  0.01  0.00  0.00
13:33:09 B17F.E   -             -       0    NoReport np=2               room -    OpenFloor  3   19    0.00  0.01  0.93  0.00  0.00  0.01
13:33:09 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.02  0.87  0.00  0.00  0.02
13:33:09 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 OpenFloor  3   20    0.00  0.00  0.02  0.00  0.00  0.00
13:33:10 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        2   21    0.00  0.02  0.86  0.00  0.01  0.02
13:33:10 B17F.2   B17F23240568  sit     65   NoReport sit                trk  1.00 Sit        2   21    0.00  0.00  0.01  0.00  0.00  0.00
13:33:11 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        2   22    0.00  0.02  0.85  0.00  0.01  0.02
13:33:11 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Sit        2   22    0.00  0.00  0.02  0.00  0.00  0.00
13:33:12 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        2   23    0.00  0.02  0.85  0.00  0.01  0.02
13:33:12 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Sit        2   23    0.00  0.00  0.01  0.00  0.00  0.00
13:33:13 B17F.2   B17F23240568  sit     72   NoReport sit                trk  1.00 Sit        2   24    0.00  0.00  0.01  0.00  0.00  0.00
13:33:13 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        2   24    0.00  0.02  0.85  0.00  0.01  0.02
13:33:14 B17F.2   B17F23240568  sit     68   NoReport sit                trk  1.00 Sit        2   25    0.00  0.00  0.01  0.00  0.00  0.00
13:33:14 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        2   25    0.00  0.02  0.85  0.00  0.01  0.02
13:33:15 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Sit        2   26    0.00  0.00  0.01  0.00  0.00  0.00
13:33:15 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        2   26    0.00  0.02  0.85  0.00  0.01  0.02
13:33:16 B17F.2   B17F23240568  sit     64   NoReport sit                trk  1.00 Sit        2   27    0.00  0.00  0.01  0.00  0.00  0.00
13:33:16 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        2   27    0.00  0.02  0.85  0.00  0.01  0.02
13:33:17 B17F.2   B17F23240568  sit     47   NoReport sit                trk  1.00 Sit        2   28    0.00  0.00  0.01  0.00  0.00  0.00
13:33:17 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Sit        2   28    0.00  0.02  0.85  0.00  0.01  0.02
13:33:18 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   29    0.00  0.02  0.85  0.00  0.01  0.02
13:33:18 B17F.2   B17F23240568  sit     23   NoReport sit                trk  1.00 BlindOpen  2   29    0.00  0.00  0.01  0.00  0.00  0.00
13:33:19 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   30    0.00  0.02  0.85  0.00  0.01  0.02
13:33:19 B17F.2   B17F23240568  sit     25   NoReport sit                trk  1.00 BlindOpen  2   30    0.00  0.00  0.01  0.00  0.00  0.00
13:33:20 B17F.2   B17F23240568  sit     40   NoReport sit                trk  1.00 BlindOpen  2   31    0.00  0.00  0.01  0.00  0.00  0.00
13:33:20 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   31    0.00  0.02  0.85  0.00  0.01  0.02
13:33:21 B17F.2   B17F23240568  sit     10   NoReport sit                trk  1.00 BlindOpen  2   32    0.00  0.00  0.01  0.00  0.00  0.00
13:33:21 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   32    0.00  0.02  0.85  0.00  0.01  0.02
13:33:22 B17F.2   B17F23240568  sit     37   NoReport sit                trk  1.00 BlindOpen  2   33    0.00  0.00  0.02  0.00  0.00  0.00
13:33:22 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   33    0.00  0.02  0.85  0.00  0.01  0.02
13:33:23 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   34    0.00  0.02  0.85  0.00  0.01  0.02
13:33:23 B17F.2   B17F23240568  sit     46   NoReport sit                trk  1.00 BlindOpen  2   34    0.00  0.00  0.02  0.01  0.00  0.00
13:33:24 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 BlindOpen  2   35    0.00  0.00  0.01  0.00  0.00  0.00
13:33:24 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   35    0.00  0.02  0.85  0.00  0.01  0.02
13:33:25 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 BlindOpen  2   36    0.00  0.00  0.01  0.00  0.00  0.00
13:33:25 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   36    0.00  0.02  0.85  0.00  0.01  0.02
13:33:26 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 BlindOpen  2   37    0.00  0.00  0.01  0.00  0.00  0.00
13:33:26 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   37    0.00  0.02  0.85  0.00  0.01  0.02
13:33:27 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   38    0.00  0.02  0.85  0.00  0.01  0.02
13:33:27 B17F.2   B17F23240568  stand   67   NoReport stand              trk  1.00 BlindOpen  2   38    0.00  0.00  0.03  0.00  0.00  0.00
13:33:28 B17F.2   B17F23240568  stand   67   NoReport stand              trk  1.00 BlindOpen  2   39    0.00  0.01  0.22  0.02  0.00  0.01
13:33:28 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 BlindOpen  2   39    0.00  0.02  0.85  0.00  0.01  0.02
13:33:29 B17F.2   B17F23240568  stand   74   NoReport stand              trk  1.00 Empty      2   40    0.00  0.01  0.26  0.01  0.00  0.01
13:33:29 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Empty      2   40    0.00  0.02  0.85  0.00  0.01  0.02
13:33:30 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      2   41    0.00  0.01  0.29  0.01  0.00  0.01
13:33:30 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Empty      2   41    0.00  0.02  0.85  0.00  0.01  0.02
13:33:31 B17F.E   -             -       0    NoReport np=1               room -    Empty      2   41    0.06  0.11  0.16  0.13  0.19  0.02
13:33:31 B17F.2   B17F23240568  stand   79   NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.29  0.02  0.01  0.02
13:33:32 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.30  0.02  0.01  0.02
13:33:33 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.30  0.02  0.02  0.02
13:33:34 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.31  0.02  0.02  0.02
13:33:35 B17F.2   B17F23240568  stand   73   NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.32  0.02  0.02  0.02
13:33:36 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.32  0.02  0.02  0.02
13:33:37 B17F.2   B17F23240568  stand   84   NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.33  0.02  0.02  0.02
13:33:38 B17F.2   B17F23240568  stand   71   NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.33  0.02  0.02  0.02
13:33:39 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.33  0.02  0.02  0.02
13:33:40 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.33  0.02  0.02  0.02
13:33:40 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:33:41 B17F.2   B17F23240568  stand   82   NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:33:42 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:33:43 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:33:44 B17F.2   B17F23240568  stand   63   NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:33:45 B17F.2   B17F23240568  walk    34   NoReport walk               trk  1.00 Empty      1   0     0.00  0.04  0.45  0.03  0.03  0.03
13:33:46 B17F.2   B17F23240568  walk    15   NoReport walk               trk  1.00 Empty      1   0     0.00  0.04  0.52  0.03  0.04  0.03
13:33:47 B17F.2   B17F23240568  walk    32   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.56  0.03  0.05  0.03
13:33:48 B17F.2   B17F23240568  walk    64   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.53  0.02  0.05  0.03
13:33:49 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.04  0.50  0.02  0.04  0.03
13:33:50 B17F.2   B17F23240568  stand   66   NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.47  0.02  0.04  0.03
13:33:51 B17F.2   B17F23240568  stand   66   NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.45  0.02  0.04  0.03
13:33:52 B17F.2   B17F23240568  stand   67   NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.43  0.02  0.03  0.03
13:33:53 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.41  0.02  0.03  0.03
13:33:54 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.39  0.02  0.03  0.02
13:33:55 B17F.2   B17F23240568  stand   81   NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.38  0.02  0.03  0.02
13:33:56 B17F.2   B17F23240568  stand   51   NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.37  0.02  0.03  0.02
13:33:57 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.03  0.36  0.02  0.03  0.02
13:33:58 B17F.2   B17F23240568  stand   64   NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.36  0.02  0.03  0.02
13:33:59 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.36  0.02  0.03  0.02
13:34:00 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.35  0.02  0.03  0.02
13:34:01 B17F.2   B17F23240568  stand   64   NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.35  0.02  0.03  0.02
13:34:02 B17F.2   B17F23240568  stand   70   NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.35  0.02  0.03  0.02
13:34:03 B17F.2   B17F23240568  stand   74   NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.35  0.02  0.03  0.02
13:34:04 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.35  0.02  0.03  0.02
13:34:05 B17F.2   B17F23240568  stand   62   NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.45  0.03  0.03  0.03
13:34:06 B17F.2   B17F23240568  stand   62   NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.52  0.03  0.04  0.03
13:34:07 B17F.2   B17F23240568  stand   70   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.56  0.03  0.05  0.03
13:34:08 B17F.2   B17F23240568  stand   65   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.53  0.02  0.04  0.03
13:34:09 B17F.2   B17F23240568  stand   79   NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.50  0.02  0.04  0.03
13:34:10 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.47  0.02  0.04  0.03
13:34:11 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.45  0.02  0.04  0.03
13:34:12 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.43  0.02  0.03  0.03
13:34:13 B17F.2   B17F23240568  stand   55   NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.41  0.02  0.03  0.03
13:34:14 B17F.2   B17F23240568  stand   69   NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.40  0.02  0.03  0.02
13:34:15 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.39  0.02  0.03  0.02
13:34:16 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.48  0.03  0.04  0.03
13:34:17 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.45  0.02  0.04  0.03
13:34:18 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.03  0.19  0.02  0.03  0.02
13:34:19 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.02  0.08  0.02  0.02  0.01
13:34:20 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.05  0.01  0.01  0.01
13:34:21 B17F.2   B17F23240568  sit     82   NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.08  0.01  0.01  0.01
13:34:22 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.05  0.01  0.00  0.01
13:34:23 B17F.2   B17F23240568  sit     57   NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:24 B17F.2   B17F23240568  sit     59   NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:25 B17F.2   B17F23240568  sit     81   NoReport sit                trk  1.00 Empty      1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:34:26 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:27 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:28 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:29 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:30 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:31 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:34:32 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:34:33 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:34 B17F.2   B17F23240568  sit     71   NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:35 B17F.2   B17F23240568  sit     59   NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:36 B17F.2   B17F23240568  sit     41   NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:34:37 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:34:38 B17F.2   B17F23240568  sit     32   NoReport sit                trk  1.00 Empty      1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:39 B17F.2   B17F23240568  sit     67   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:40 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:41 B17F.2   B17F23240568  sit     76   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:42 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:43 B17F.2   B17F23240568  sit     78   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:44 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:45 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:46 B17F.2   B17F23240568  sit     72   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:47 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:48 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:49 B17F.2   B17F23240568  sit     69   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:50 B17F.2   B17F23240568  sit     65   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:51 B17F.2   B17F23240568  sit     70   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:52 B17F.2   B17F23240568  sit     62   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:53 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:54 B17F.2   B17F23240568  sit     69   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:55 B17F.2   B17F23240568  sit     68   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:56 B17F.2   B17F23240568  sit     65   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:57 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:58 B17F.2   B17F23240568  sit     72   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:34:59 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:00 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:35:01 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.01  0.02
13:35:02 B17F.2   B17F23240568  sit     84   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.02  0.01  0.01
13:35:03 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.01  0.01  0.01
13:35:04 B17F.2   B17F23240568  sit     66   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:05 B17F.2   B17F23240568  sit     83   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.01  0.00  0.01
13:35:06 B17F.2   B17F23240568  sit     66   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:07 B17F.2   B17F23240568  sit     65   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:08 B17F.2   B17F23240568  sit     84   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:35:09 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:10 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:11 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:12 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:13 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:14 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:15 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:16 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:17 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:18 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:19 B17F.2   B17F23240568  sit     80   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:35:20 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:21 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:22 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:23 B17F.2   B17F23240568  sit     70   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:24 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:35:25 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.01  0.02
13:35:26 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.02  0.02
13:35:27 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.02  0.01  0.01
13:35:28 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:35:29 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:30 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:31 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:32 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:33 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:34 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:35 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:36 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:37 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:38 B17F.2   B17F23240568  sit     70   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:39 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:40 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   28    0.00  0.01  0.04  0.01  0.00  0.01
13:35:41 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   29    0.00  0.01  0.04  0.01  0.00  0.01
13:35:42 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   30    0.00  0.01  0.04  0.01  0.00  0.01
13:35:43 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   31    0.00  0.01  0.04  0.01  0.00  0.01
13:35:44 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:45 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:46 B17F.2   B17F23240568  sit     69   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:46 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:47 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:35:48 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.01  0.02
13:35:49 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.02  0.01  0.01
13:35:50 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:35:51 B17F.2   B17F23240568  sit     84   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.01  0.00  0.01
13:35:52 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.01  0.00  0.01
13:35:53 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:54 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:55 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:56 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:57 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:58 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:35:59 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:00 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:01 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:02 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:03 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:36:04 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.01  0.02
13:36:05 B17F.2   B17F23240568  sit     53   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.02  0.01  0.01
13:36:06 B17F.2   B17F23240568  sit     48   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:36:07 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:08 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:36:09 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:36:10 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:11 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:12 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:13 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:14 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:15 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:16 B17F.2   B17F23240568  sit     16   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:17 B17F.2   B17F23240568  sit     17   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:18 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:19 B17F.2   B17F23240568  sit     17   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:20 B17F.2   B17F23240568  sit     19   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:21 B17F.2   B17F23240568  sit     55   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:22 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:23 B17F.2   B17F23240568  sit     19   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:24 B17F.2   B17F23240568  sit     50   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:25 B17F.2   B17F23240568  sit     25   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:36:26 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.01  0.02
13:36:27 B17F.2   B17F23240568  sit     39   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.02  0.02
13:36:28 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.02  0.01  0.01
13:36:29 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:36:30 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:31 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:32 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:33 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:34 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:35 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:36 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:37 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:38 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:39 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:40 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:41 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:42 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:43 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:44 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:45 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:46 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:47 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:48 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:49 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:50 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:51 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:36:52 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   27    0.00  0.01  0.04  0.01  0.00  0.01
13:36:53 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   28    0.00  0.01  0.04  0.01  0.00  0.01
13:36:54 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   29    0.00  0.01  0.04  0.01  0.00  0.01
13:36:55 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   30    0.00  0.01  0.04  0.01  0.00  0.01
13:36:56 B17F.2   B17F23240568  sit     71   NoReport sit                trk  1.00 Fallen     1   31    0.00  0.01  0.04  0.01  0.00  0.01
13:36:57 B17F.2   B17F23240568  sit     65   NoReport sit                trk  1.00 Fallen     1   32    0.00  0.01  0.04  0.01  0.00  0.01
13:36:58 B17F.2   B17F23240568  sit     65   NoReport sit                trk  1.00 Fallen     1   33    0.00  0.01  0.07  0.02  0.01  0.02
13:36:59 B17F.2   B17F23240568  sit     62   NoReport sit                trk  1.00 Fallen     1   34    0.00  0.01  0.08  0.03  0.01  0.02
13:37:00 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   35    0.00  0.01  0.05  0.02  0.01  0.01
13:37:01 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   36    0.00  0.01  0.04  0.01  0.01  0.01
13:37:02 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   37    0.00  0.01  0.04  0.01  0.00  0.01
13:37:03 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   38    0.00  0.01  0.04  0.01  0.00  0.01
13:37:04 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   39    0.00  0.01  0.04  0.01  0.00  0.01
13:37:05 B17F.2   B17F23240568  sit     69   NoReport sit                trk  1.00 Fallen     1   40    0.00  0.01  0.04  0.01  0.00  0.01
13:37:06 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:07 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:08 B17F.2   B17F23240568  sit     66   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:09 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:10 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:11 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:37:12 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.01  0.02
13:37:13 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.02  0.01  0.01
13:37:14 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:37:15 B17F.2   B17F23240568  sit     72   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:16 B17F.2   B17F23240568  sit     86   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:37:17 B17F.2   B17F23240568  sit     72   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:18 B17F.2   B17F23240568  sit     87   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.01  0.00  0.01
13:37:19 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:20 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:21 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:22 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:23 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:24 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:25 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:26 B17F.2   B17F23240568  sit     69   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:27 B17F.2   B17F23240568  sit     97   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:37:28 B17F.2   B17F23240568  sit     76   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:29 B17F.2   B17F23240568  sit     61   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:30 B17F.2   B17F23240568  sit     50   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:31 B17F.2   B17F23240568  sit     62   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:32 B17F.2   B17F23240568  sit     78   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:37:33 B17F.2   B17F23240568  walk    94   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.02  0.24  0.04  0.02  0.03
13:37:34 B17F.2   B17F23240568  walk    75   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.37  0.04  0.03  0.03
13:37:35 B17F.2   B17F23240568  walk    91   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.47  0.04  0.04  0.03
13:37:36 B17F.2   B17F23240568  walk    85   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.45  0.03  0.04  0.03
13:37:37 B17F.2   B17F23240568  sit     76   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.03  0.19  0.02  0.03  0.02
13:37:38 B17F.2   B17F23240568  sit     63   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.08  0.02  0.02  0.01
13:37:39 B17F.2   B17F23240568  sit     71   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.01  0.01  0.01
13:37:40 B17F.2   B17F23240568  sit     64   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:37:41 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:42 B17F.2   B17F23240568  sit     72   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:43 B17F.2   B17F23240568  sit     85   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:37:44 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:45 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:46 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:37:47 B17F.2   B17F23240568  stand   81   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.01  0.13  0.02  0.01  0.02
13:37:48 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.01  0.19  0.02  0.01  0.02
13:37:49 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.02  0.23  0.02  0.01  0.02
13:37:50 B17F.2   B17F23240568  stand   67   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.02  0.26  0.02  0.02  0.02
13:37:51 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.02  0.28  0.02  0.02  0.02
13:37:52 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.02  0.30  0.02  0.02  0.02
13:37:53 B17F.2   B17F23240568  stand   76   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.02  0.31  0.02  0.02  0.02
13:37:53 B17F.2   B17F23240568  walk    68   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.32  0.02  0.02  0.02
13:37:54 B17F.2   B17F23240568  walk    71   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.32  0.02  0.02  0.02
13:37:55 B17F.2   B17F23240568  walk    69   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.33  0.02  0.02  0.02
13:37:56 B17F.2   B17F23240568  walk    76   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.33  0.02  0.02  0.02
13:37:57 B17F.2   B17F23240568  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.33  0.02  0.02  0.02
13:37:58 B17F.2   B17F23240568  stand   77   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:37:59 B17F.2   B17F23240568  stand   84   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:38:00 B17F.2   B17F23240568  stand   71   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:38:01 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:38:02 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:38:03 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.34  0.02  0.02  0.02
13:38:04 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.03  0.34  0.02  0.02  0.02
13:38:05 B17F.2   B17F23240568  stand   65   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.03  0.34  0.02  0.02  0.02
13:38:06 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.04  0.45  0.03  0.03  0.03
13:38:07 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.04  0.52  0.03  0.04  0.03
13:38:08 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.04  0.49  0.02  0.04  0.03
13:38:09 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.04  0.46  0.02  0.04  0.03
13:38:10 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.04  0.44  0.02  0.04  0.03
13:38:11 B17F.2   B17F23240568  stand   64   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.04  0.42  0.02  0.03  0.03
13:38:12 B17F.2   B17F23240568  stand   76   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.03  0.41  0.02  0.03  0.02
13:38:13 B17F.2   B17F23240568  stand   73   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.03  0.39  0.02  0.03  0.02
13:38:14 B17F.2   B17F23240568  stand   81   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.38  0.02  0.03  0.02
13:38:15 B17F.2   B17F23240568  stand   80   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.38  0.02  0.03  0.02
13:38:16 B17F.2   B17F23240568  stand   70   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.37  0.02  0.03  0.02
13:38:17 B17F.2   B17F23240568  stand   71   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.36  0.02  0.03  0.02
13:38:18 B17F.2   B17F23240568  stand   86   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.36  0.02  0.03  0.02
13:38:19 B17F.2   B17F23240568  stand   75   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.36  0.02  0.03  0.02
13:38:20 B17F.2   B17F23240568  stand   82   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.35  0.02  0.03  0.02
13:38:21 B17F.2   B17F23240568  stand   91   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.35  0.02  0.03  0.02
13:38:22 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.13  0.02  0.02  0.02
13:38:23 B17F.2   B17F23240568  sit     82   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.12  0.01  0.01  0.01
13:38:24 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.06  0.01  0.01  0.01
13:38:25 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:26 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:27 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:28 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:29 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:30 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:31 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:32 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:33 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:34 B17F.2   B17F23240568  sit     83   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:38:35 B17F.2   B17F23240568  sit     76   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:36 B17F.2   B17F23240568  sit     85   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.01  0.00  0.01
13:38:37 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:38 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:39 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:38:40 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.01  0.02
13:38:41 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.02  0.02
13:38:42 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.02  0.01  0.01
13:38:43 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:38:44 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:45 B17F.2   B17F23240568  sit     90   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:38:46 B17F.2   B17F23240568  sit     82   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.01  0.00  0.01
13:38:47 B17F.2   B17F23240568  sit     85   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.01  0.00  0.01
13:38:48 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.02  0.01  0.02
13:38:49 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.01  0.01  0.01
13:38:50 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:51 B17F.2   B17F23240568  sit     84   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:38:52 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:53 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:54 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:55 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:56 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:57 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:58 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:38:59 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:00 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:01 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:02 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:03 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:04 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:05 B17F.2   B17F23240568  sit     78   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:06 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:07 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:08 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:09 B17F.2   B17F23240568  sit     94   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:39:10 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:11 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:12 B17F.2   B17F23240568  sit     90   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:39:13 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:14 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:15 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:16 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:17 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:18 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:19 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:20 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:21 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:22 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:23 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:24 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:25 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:26 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:27 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:28 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:29 B17F.2   B17F23240568  sit     70   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:30 B17F.2   B17F23240568  sit     78   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:31 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:32 B17F.2   B17F23240568  sit     71   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:33 B17F.2   B17F23240568  sit     84   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:39:34 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:35 B17F.2   B17F23240568  sit     87   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.01  0.00  0.01
13:39:36 B17F.2   B17F23240568  sit     84   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.01  0.00  0.01
13:39:37 B17F.2   B17F23240568  sit     76   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.01  0.00  0.01
13:39:38 B17F.2   B17F23240568  sit     69   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:39 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:40 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:41 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:42 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:43 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:44 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:45 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:46 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:47 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.07  0.02  0.01  0.02
13:39:48 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.03  0.01  0.02
13:39:49 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.02  0.01  0.01
13:39:50 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.01  0.01
13:39:51 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:52 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:53 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:54 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:55 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:56 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:39:57 B17F.2   B17F23240568  sit     83   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:39:58 B17F.2   B17F23240568  sit     80   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.01  0.00  0.01
13:39:58 B17F.2   B17F23240568  sit     76   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.01  0.00  0.01
13:39:59 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:00 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:01 B17F.2   B17F23240568  sit     83   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:40:02 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:03 B17F.2   B17F23240568  sit     72   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:04 B17F.2   B17F23240568  sit     80   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:40:05 B17F.2   B17F23240568  sit     90   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.01  0.00  0.01
13:40:06 B17F.2   B17F23240568  sit     72   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.01  0.00  0.01
13:40:07 B17F.2   B17F23240568  sit     78   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:08 B17F.2   B17F23240568  sit     71   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:09 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:10 B17F.2   B17F23240568  sit     84   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:40:11 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:12 B17F.2   B17F23240568  sit     78   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:13 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:14 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:15 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:16 B17F.2   B17F23240568  sit     93   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.00  0.07  0.01  0.00  0.01
13:40:17 B17F.2   B17F23240568  sit     80   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.08  0.01  0.00  0.01
13:40:18 B17F.2   B17F23240568  sit     91   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.01  0.00  0.01
13:40:19 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.05  0.01  0.00  0.01
13:40:20 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:21 B17F.2   B17F23240568  sit     67   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.04  0.01  0.00  0.01
13:40:22 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   28    0.00  0.01  0.04  0.01  0.00  0.01
13:40:23 B17F.2   B17F23240568  sit     69   NoReport sit                trk  1.00 Fallen     1   29    0.00  0.01  0.04  0.01  0.00  0.01
13:40:24 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 Fallen     1   30    0.00  0.01  0.04  0.01  0.00  0.01
13:40:25 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   31    0.00  0.01  0.04  0.01  0.00  0.01
13:40:26 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   32    0.00  0.01  0.04  0.01  0.00  0.01
13:40:27 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   33    0.00  0.01  0.04  0.01  0.00  0.01
13:40:28 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   34    0.00  0.01  0.04  0.01  0.00  0.01
13:40:29 B17F.2   B17F23240568  sit     86   NoReport sit                trk  1.00 Fallen     1   35    0.00  0.00  0.07  0.01  0.00  0.01
13:40:30 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 Fallen     1   36    0.00  0.01  0.04  0.01  0.00  0.01
13:40:31 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   37    0.00  0.01  0.04  0.01  0.00  0.01
13:40:32 B17F.2   B17F23240568  sit     78   NoReport sit                trk  1.00 Fallen     1   38    0.00  0.01  0.04  0.01  0.00  0.01
13:40:33 B17F.2   B17F23240568  sit     76   NoReport sit                trk  1.00 Fallen     1   39    0.00  0.01  0.04  0.01  0.00  0.01
13:40:34 B17F.2   B17F23240568  sit     78   NoReport sit                trk  1.00 Fallen     1   40    0.00  0.01  0.04  0.01  0.00  0.01
13:40:35 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   41    0.00  0.01  0.04  0.01  0.00  0.01
13:40:36 B17F.2   B17F23240568  sit     80   NoReport sit                trk  1.00 Fallen     1   42    0.00  0.00  0.07  0.01  0.00  0.01
13:40:37 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Fallen     1   43    0.00  0.01  0.04  0.01  0.00  0.01
13:40:38 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   44    0.00  0.01  0.04  0.01  0.00  0.01
13:40:39 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Fallen     1   45    0.00  0.01  0.04  0.01  0.00  0.01
13:40:40 B17F.2   B17F23240568  sit     78   NoReport sit                trk  1.00 Fallen     1   46    0.00  0.01  0.04  0.01  0.00  0.01
13:40:41 B17F.2   B17F23240568  sit     77   NoReport sit                trk  1.00 Fallen     1   47    0.00  0.01  0.04  0.01  0.00  0.01
13:40:42 B17F.2   B17F23240568  sit     94   NoReport sit                trk  1.00 Fallen     1   48    0.00  0.00  0.07  0.01  0.00  0.01
13:40:43 B17F.2   B17F23240568  sit     90   NoReport sit                trk  1.00 Fallen     1   49    0.00  0.01  0.08  0.01  0.00  0.01
13:40:44 B17F.2   B17F23240568  sit     80   NoReport sit                trk  1.00 Fallen     1   50    0.00  0.01  0.09  0.01  0.00  0.01
13:40:45 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   51    0.00  0.01  0.05  0.01  0.00  0.01
13:40:46 B17F.2   B17F23240568  sit     84   NoReport sit                trk  1.00 Fallen     1   52    0.00  0.01  0.08  0.01  0.00  0.01
13:40:47 B17F.2   B17F23240568  sit     86   NoReport sit                trk  1.00 Fallen     1   53    0.00  0.01  0.09  0.01  0.00  0.01
13:40:48 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   54    0.00  0.01  0.05  0.01  0.00  0.01
13:40:49 B17F.2   B17F23240568  sit     76   NoReport sit                trk  1.00 Fallen     1   55    0.00  0.01  0.04  0.01  0.00  0.01
13:40:50 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   56    0.00  0.01  0.04  0.01  0.00  0.01
13:40:51 B17F.2   B17F23240568  sit     79   NoReport sit                trk  1.00 Fallen     1   57    0.00  0.01  0.04  0.01  0.00  0.01
13:40:52 B17F.2   B17F23240568  sit     86   NoReport sit                trk  1.00 Fallen     1   58    0.00  0.00  0.07  0.01  0.00  0.01
13:40:53 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   59    0.00  0.01  0.04  0.01  0.00  0.01
13:40:54 B17F.2   B17F23240568  sit     70   NoReport sit                trk  1.00 Fallen     1   60    0.00  0.01  0.07  0.02  0.01  0.02
13:40:55 B17F.2   B17F23240568  sit     75   NoReport sit                trk  1.00 Fallen     1   61    0.00  0.01  0.08  0.03  0.01  0.02
13:40:56 B17F.2   B17F23240568  sit     90   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.16  0.03  0.01  0.02
13:40:57 B17F.2   B17F23240568  sit     83   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.21  0.03  0.02  0.02
13:40:58 B17F.2   B17F23240568  walk    86   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.35  0.04  0.02  0.03
13:40:59 B17F.2   B17F23240568  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.46  0.04  0.04  0.03
13:41:00 B17F.2   B17F23240568  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.52  0.03  0.04  0.03
13:41:01 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.05  0.56  0.03  0.05  0.03
13:41:02 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.06  0.36  0.03  0.06  0.04
13:41:03 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.06  0.24  0.04  0.07  0.03
13:41:04 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.05  0.17  0.04  0.07  0.02
13:41:05 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.04  0.13  0.04  0.06  0.02
13:41:06 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.03  0.11  0.04  0.05  0.02
13:41:07 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.03  0.10  0.04  0.04  0.02
13:41:08 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.03  0.10  0.04  0.04  0.02
13:41:09 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.04  0.03  0.02
13:41:10 B17F.2   B17F23240568  sit     39   NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.03  0.02
13:41:11 B17F.2   B17F23240568  sit     68   NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.03  0.02
13:41:12 B17F.2   B17F23240568  sit     83   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.17  0.03  0.02  0.02
13:41:13 B17F.2   B17F23240568  sit     83   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.22  0.03  0.02  0.02
13:41:14 B17F.2   B17F23240568  walk    78   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.36  0.04  0.03  0.03
13:41:15 B17F.2   B17F23240568  walk    83   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.36  0.03  0.03  0.02
13:41:16 B17F.2   B17F23240568  walk    82   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.35  0.03  0.03  0.02
13:41:17 B17F.2   B17F23240568  sit     73   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.04  0.21  0.03  0.03  0.03
13:41:18 B17F.2   B17F23240568  walk    88   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.35  0.04  0.04  0.03
13:41:19 B17F.2   B17F23240568  walk    62   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.45  0.04  0.05  0.03
13:41:20 B17F.2   B17F23240568  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.52  0.04  0.05  0.03
13:41:21 B17F.2   B17F23240568  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.56  0.03  0.05  0.03
13:41:22 B17F.2   B17F23240568  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.58  0.03  0.05  0.03
13:41:23 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.05  0.59  0.02  0.05  0.03
13:41:24 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.60  0.02  0.05  0.03
13:41:25 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:41:26 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:41:27 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:41:28 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:41:29 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:41:30 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:41:31 B17F.E   -             -       0    NoReport np=2               room -    Fallen     1   0     0.25  0.09  0.13  0.11  0.16  0.01
13:41:31 B17F.0   B17F03200777  stand   72   NoReport stand              trk  1.00 Fallen     2   0     0.16  0.06  0.33  0.07  0.10  0.01
13:41:31 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.01  0.03  0.76  0.01  0.03  0.02
13:41:32 B17F.0   B17F03200777  stand   51   NoReport stand              trk  1.00 Fallen     2   0     0.08  0.03  0.48  0.03  0.05  0.01
13:41:32 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.03  0.82  0.01  0.02  0.02
13:41:33 B17F.0   B17F03200777  stand   95   NoReport stand              trk  1.00 Fallen     2   0     0.01  0.03  0.71  0.02  0.03  0.02
13:41:33 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.84  0.00  0.01  0.02
13:41:34 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.84  0.00  0.01  0.02
13:41:34 B17F.0   B17F03200777  stand   88   NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.80  0.01  0.02  0.02
13:41:35 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:35 B17F.0   B17F03200777  stand   86   NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.84  0.01  0.01  0.02
13:41:36 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:36 B17F.0   B17F03200777  walk    91   NoReport walk               trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:37 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:37 B17F.0   B17F03200777  walk    71   NoReport walk               trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:38 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:38 B17F.0   B17F03200777  walk    77   NoReport walk               trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:39 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:39 B17F.0   B17F03200777  walk    94   NoReport walk               trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:40 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:40 B17F.0   B17F03200777  walk    107  NoReport walk               trk  1.00 Fallen     2   0     0.00  0.02  0.85  0.00  0.01  0.02
13:41:41 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   16    0.00  0.02  0.85  0.00  0.01  0.02
13:41:41 B17F.0   B17F03200777  walk    122  NoReport walk               trk  1.00 Fallen     2   16    0.00  0.02  0.85  0.00  0.01  0.02
13:41:42 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   17    0.00  0.02  0.85  0.00  0.01  0.02
13:41:42 B17F.0   B17F03200777  walk    0    NoReport walk               trk  1.00 Fallen     2   17    0.00  0.01  0.96  0.00  0.00  0.01
13:41:43 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   18    0.00  0.02  0.85  0.00  0.01  0.02
13:41:43 B17F.0   B17F03200777  walk    0    NoReport walk               trk  1.00 Fallen     2   18    0.00  0.01  0.94  0.00  0.00  0.01
13:41:44 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   19    0.00  0.02  0.85  0.00  0.01  0.02
13:41:44 B17F.0   B17F03200777  walk    0    NoReport walk               trk  1.00 Fallen     2   19    0.00  0.01  0.97  0.00  0.00  0.01
13:41:45 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   20    0.00  0.02  0.85  0.00  0.01  0.02
13:41:45 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Fallen     2   20    0.00  0.01  0.97  0.00  0.00  0.01
13:41:46 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   21    0.00  0.02  0.85  0.00  0.01  0.02
13:41:46 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Fallen     2   21    0.00  0.01  0.97  0.00  0.00  0.01
13:41:47 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   22    0.00  0.02  0.85  0.00  0.01  0.02
13:41:47 B17F.0   B17F03200777  stand   0    NoReport stand              trk  1.00 Fallen     2   22    0.00  0.01  0.97  0.00  0.00  0.01
13:41:47 B17F.E   -             -       0    NoReport ExitRoom(rdr)      room -    Fallen     2   22    0.25  0.09  0.13  0.11  0.16  0.01
13:41:48 B17F.E   -             -       0    NoReport np=1               room -    Fallen     2   22    0.25  0.09  0.13  0.11  0.16  0.01
13:41:48 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     2   23    0.00  0.04  0.74  0.00  0.02  0.04
13:41:49 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   24    0.01  0.05  0.68  0.01  0.03  0.04
13:41:50 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   25    0.01  0.05  0.65  0.01  0.04  0.03
13:41:51 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   26    0.01  0.05  0.62  0.02  0.05  0.03
13:41:52 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   27    0.01  0.05  0.62  0.02  0.05  0.03
13:41:53 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   28    0.01  0.05  0.62  0.02  0.05  0.03
13:41:54 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   29    0.01  0.05  0.61  0.02  0.05  0.03
13:41:54 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   30    0.01  0.05  0.61  0.02  0.05  0.03
13:41:55 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   31    0.01  0.05  0.61  0.02  0.05  0.03
13:41:56 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   32    0.01  0.05  0.61  0.02  0.05  0.03
13:41:57 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   33    0.01  0.05  0.61  0.02  0.05  0.03
13:41:58 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   34    0.01  0.05  0.61  0.02  0.05  0.03
13:41:59 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   35    0.01  0.05  0.61  0.02  0.05  0.03
13:42:00 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   36    0.01  0.05  0.61  0.02  0.05  0.03
13:42:01 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   37    0.01  0.05  0.61  0.02  0.05  0.03
13:42:02 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   38    0.01  0.05  0.61  0.02  0.05  0.03
13:42:03 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   39    0.01  0.05  0.61  0.02  0.05  0.03
13:42:04 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   40    0.01  0.05  0.61  0.02  0.05  0.03
13:42:05 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   41    0.01  0.05  0.61  0.02  0.05  0.03
13:42:06 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   42    0.01  0.05  0.61  0.02  0.05  0.03
13:42:07 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   43    0.01  0.05  0.61  0.02  0.05  0.03
13:42:08 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   44    0.01  0.05  0.61  0.02  0.05  0.03
13:42:09 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   45    0.01  0.05  0.61  0.02  0.05  0.03
13:42:10 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   46    0.01  0.05  0.61  0.02  0.05  0.03
13:42:11 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   47    0.01  0.05  0.61  0.02  0.05  0.03
13:42:12 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   48    0.01  0.05  0.61  0.02  0.05  0.03
13:42:13 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   49    0.01  0.05  0.61  0.02  0.05  0.03
13:42:14 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   50    0.01  0.05  0.61  0.02  0.05  0.03
13:42:15 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   51    0.01  0.05  0.61  0.02  0.05  0.03
13:42:16 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   52    0.01  0.05  0.61  0.02  0.05  0.03
13:42:17 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   53    0.01  0.05  0.61  0.02  0.05  0.03
13:42:18 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   54    0.01  0.05  0.61  0.02  0.05  0.03
13:42:19 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   55    0.01  0.05  0.61  0.02  0.05  0.03
13:42:20 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   56    0.01  0.05  0.61  0.02  0.05  0.03
13:42:21 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   57    0.01  0.05  0.61  0.02  0.05  0.03
13:42:22 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   58    0.01  0.05  0.61  0.02  0.05  0.03
13:42:23 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   59    0.01  0.05  0.61  0.02  0.05  0.03
13:42:24 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   60    0.01  0.05  0.61  0.02  0.05  0.03
13:42:25 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   61    0.01  0.05  0.61  0.02  0.05  0.03
13:42:26 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   62    0.01  0.05  0.61  0.02  0.05  0.03
13:42:27 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   63    0.01  0.05  0.61  0.02  0.05  0.03
13:42:28 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   64    0.01  0.05  0.61  0.02  0.05  0.03
13:42:29 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   65    0.01  0.05  0.61  0.02  0.05  0.03
13:42:30 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   66    0.01  0.05  0.61  0.02  0.05  0.03
13:42:31 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   67    0.01  0.05  0.61  0.02  0.05  0.03
13:42:32 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   68    0.01  0.05  0.61  0.02  0.05  0.03
13:42:33 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   69    0.01  0.05  0.61  0.02  0.05  0.03
13:42:34 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   70    0.01  0.05  0.61  0.02  0.05  0.03
13:42:35 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   71    0.01  0.05  0.61  0.02  0.05  0.03
13:42:36 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   72    0.01  0.05  0.61  0.02  0.05  0.03
13:42:37 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   73    0.01  0.05  0.61  0.02  0.05  0.03
13:42:38 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   74    0.01  0.05  0.61  0.02  0.05  0.03
13:42:39 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:40 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:41 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:42 B17F.2   B17F23240568  stand   40   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:43 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:44 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:45 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:46 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:47 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:48 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:49 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:50 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:51 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:52 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:53 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:54 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:55 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:56 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:57 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:58 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:42:59 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:00 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:01 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:02 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:03 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:04 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:05 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:06 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:07 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:08 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:09 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:10 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:11 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:12 B17F.2   B17F23240568  walk    28   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:43:13 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:14 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:15 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:16 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:17 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:18 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:19 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:20 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:21 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:22 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:23 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:24 B17F.2   B17F23240568  stand   55   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:25 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:26 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:27 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:28 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:29 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:30 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:31 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:32 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:33 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:34 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:43:35 B17F.2   B17F23240568  stand   80   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:43:36 B17F.2   B17F23240568  stand   87   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:43:37 B17F.2   B17F23240568  walk    77   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:43:38 B17F.2   B17F23240568  walk    77   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.58  0.02  0.05  0.03
13:43:39 B17F.2   B17F23240568  walk    90   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.55  0.02  0.05  0.03
13:43:40 B17F.2   B17F23240568  walk    125  NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.51  0.02  0.04  0.03
13:43:41 B17F.2   B17F23240568  walk    126  NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.49  0.02  0.04  0.03
13:43:42 B17F.2   B17F23240568  walk    109  NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.46  0.02  0.04  0.03
13:43:43 B17F.2   B17F23240568  walk    99   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.44  0.02  0.04  0.03
13:43:44 B17F.2   B17F23240568  walk    96   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.42  0.02  0.03  0.03
13:43:45 B17F.2   B17F23240568  sit     98   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.29  0.01  0.02  0.02
13:43:46 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.11  0.01  0.02  0.02
13:43:47 B17F.2   B17F23240568  sit     103  NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.10  0.01  0.01  0.01
13:43:48 B17F.2   B17F23240568  sit     93   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.10  0.01  0.01  0.01
13:43:49 B17F.2   B17F23240568  sit     104  NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.10  0.01  0.00  0.01
13:43:50 B17F.2   B17F23240568  sit     126  NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.01  0.00  0.01
13:43:51 B17F.2   B17F23240568  sit     97   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.01  0.00  0.01
13:43:52 B17F.2   B17F23240568  sit     103  NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.01  0.00  0.01
13:43:53 B17F.2   B17F23240568  sit     97   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.01  0.00  0.01
13:43:54 B17F.2   B17F23240568  sit     74   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.01  0.09  0.02  0.01  0.02
13:43:55 B17F.2   B17F23240568  walk    79   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.02  0.26  0.04  0.02  0.03
13:43:56 B17F.2   B17F23240568  walk    90   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.39  0.04  0.03  0.03
13:43:57 B17F.2   B17F23240568  walk    110  NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.48  0.04  0.04  0.03
13:43:58 B17F.2   B17F23240568  walk    82   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.53  0.04  0.05  0.03
13:43:58 B17F.2   B17F23240568  walk    67   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.57  0.03  0.05  0.03
13:43:59 B17F.2   B17F23240568  walk    77   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.59  0.03  0.05  0.03
13:44:00 B17F.2   B17F23240568  walk    84   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.55  0.02  0.05  0.03
13:44:01 B17F.2   B17F23240568  walk    68   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.52  0.02  0.05  0.03
13:44:02 B17F.2   B17F23240568  walk    87   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.50  0.02  0.04  0.03
13:44:03 B17F.2   B17F23240568  walk    85   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.47  0.02  0.04  0.03
13:44:04 B17F.2   B17F23240568  walk    86   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.53  0.02  0.04  0.03
13:44:05 B17F.2   B17F23240568  walk    69   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.57  0.02  0.05  0.03
13:44:06 B17F.2   B17F23240568  walk    63   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.59  0.02  0.05  0.03
13:44:07 B17F.2   B17F23240568  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.60  0.02  0.05  0.03
13:44:08 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
13:44:09 B17F.2   B17F23240568  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
13:44:10 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.07  0.40  0.03  0.07  0.04
13:44:11 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.07  0.27  0.03  0.09  0.03
13:44:12 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.06  0.19  0.04  0.09  0.03
13:44:13 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.05  0.15  0.04  0.08  0.02
13:44:14 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.04  0.12  0.04  0.06  0.02
13:44:15 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.03  0.11  0.04  0.05  0.02
13:44:16 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.03  0.10  0.04  0.04  0.02
13:44:17 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.10  0.04  0.03  0.02
13:44:18 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.04  0.03  0.02
13:44:19 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.03  0.02
13:44:20 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.03  0.02
13:44:21 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:22 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:23 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:24 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:25 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:26 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:27 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:28 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:29 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:30 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:31 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:32 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:33 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:34 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:35 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
13:44:36 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   28    0.01  0.02  0.09  0.03  0.02  0.02
13:44:37 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   29    0.01  0.02  0.09  0.03  0.02  0.02
13:44:38 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   30    0.01  0.02  0.09  0.03  0.02  0.02
13:44:39 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   31    0.01  0.02  0.09  0.03  0.02  0.02
13:44:40 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   32    0.01  0.02  0.09  0.03  0.02  0.02
13:44:41 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   33    0.01  0.02  0.09  0.03  0.02  0.02
13:44:42 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   34    0.01  0.02  0.09  0.03  0.02  0.02
13:44:43 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   35    0.01  0.02  0.09  0.03  0.02  0.02
13:44:44 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   36    0.01  0.02  0.09  0.03  0.02  0.02
13:44:45 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   37    0.01  0.02  0.09  0.03  0.02  0.02
13:44:46 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   38    0.01  0.02  0.09  0.03  0.02  0.02
13:44:47 B17F.2   B17F23240568  sit     0    NoReport sit                trk  1.00 Fallen     1   39    0.01  0.02  0.09  0.03  0.02  0.02
13:44:48 B17F.2   B17F23240568  sit     70   NoReport sit                trk  1.00 Fallen     1   40    0.01  0.02  0.09  0.03  0.02  0.02
13:44:49 B17F.2   B17F23240568  sit     84   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.16  0.03  0.02  0.02
13:44:50 B17F.2   B17F23240568  walk    69   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.32  0.04  0.03  0.03
13:44:51 B17F.2   B17F23240568  walk    85   NoReport walk               trk  1.00 Sit        1   0     0.00  0.03  0.33  0.03  0.03  0.02
13:44:52 B17F.2   B17F23240568  walk    86   NoReport walk               trk  1.00 Sit        1   0     0.00  0.03  0.33  0.03  0.03  0.02
13:44:53 B17F.2   B17F23240568  walk    80   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.44  0.03  0.03  0.03
13:44:54 B17F.2   B17F23240568  walk    77   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.51  0.03  0.04  0.03
13:44:55 B17F.2   B17F23240568  walk    62   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.56  0.03  0.05  0.03
13:44:56 B17F.2   B17F23240568  walk    75   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.58  0.03  0.05  0.03
13:44:57 B17F.2   B17F23240568  walk    120  NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.60  0.02  0.05  0.03
13:44:58 B17F.2   B17F23240568  walk    117  NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.60  0.02  0.05  0.03
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
13:32:00.777 B17F.1    walk   3    30     80     72    80        
13:32:00.777 B17F.0    stand  3    160    30     0     80   139  
13:32:01.780 B17F.0    stand  3    160    30     0     80   0    
13:32:01.780 B17F.1    walk   3    70     60     80    80   94   
13:32:02.782 B17F.0    stand  3    170    30     0     80   104  
13:32:02.782 B17F.1    walk   3    90     90     89    80   100  
13:32:03.778 B17F.1    walk   3    70     90     77    80   20   
13:32:03.778 B17F.0    stand  3    160    30     0     80   108  
13:32:04.779 B17F.0    stand  3    170    30     0     80   10   
13:32:04.779 B17F.1    walk   3    50     100    86    80   138  
13:32:05.781 B17F.1    walk   3    40     90     92    80   14   
13:32:05.781 B17F.0    stand  3    170    30     0     80   143  
13:32:06.781 B17F.0    stand  3    170    30     0     80   0    
13:32:06.781 B17F.1    walk   3    40     90     91    80   143  
13:32:07.782 B17F.1    walk   3    60     100    86    80   22   
13:32:07.782 B17F.0    stand  3    170    30     0     80   130  
13:32:08.782 B17F.1    walk   3    110    130    0     80   116  
13:32:08.782 B17F.0    stand  3    170    30     0     80   116  
13:32:09.796 B17F.0    stand  3    170    30     0     80   0    
13:32:09.796 B17F.1    sit    3    50     100    76    80   138  
13:32:10.688 B17F.0    stand  3    170    30     0     80   138  
13:32:10.688 B17F.1    sit    3    50     90     72    80   134  
13:32:11.681 B17F.0    stand  3    170    30     0     80   134  
13:32:11.681 B17F.1    sit    3    70     90     65    80   116  
13:32:12.682 B17F.1    sit    3    130    80     0     80   60   
13:32:12.682 B17F.0    stand  3    170    30     0     80   64   
13:32:13.685 B17F.1    sit    3    120    80     0     80   70   
13:32:13.685 B17F.0    stand  3    170    30     0     80   70   
13:32:14.683 B17F.1    sit    3    120    60     0     80   58   
13:32:14.683 B17F.0    stand  3    170    30     0     80   58   
13:32:15.686 B17F.0    stand  3    170    30     0     80   0    
13:32:15.686 B17F.1    sit    3    60     80     72    80   120  
13:32:16.688 B17F.0    stand  3    170    30     0     80   120  
13:32:16.688 B17F.1    sit    3    50     90     84    80   134  
13:32:17.687 B17F.0    stand  3    170    30     0     80   134  
13:32:17.687 B17F.1    sit    3    50     90     76    80   134  
13:32:18.690 B17F.1    sit    3    50     90     75    80   0    
13:32:18.690 B17F.0    stand  3    170    30     0     80   134  
13:32:19.694 B17F.1    sit    3    40     90     65    80   143  
13:32:19.694 B17F.0    stand  3    170    30     0     80   143  
13:32:20.692 B17F.1    sit    3    50     110    88    80   144  
13:32:20.692 B17F.0    stand  3    170    30     0     80   144  
13:32:21.591 B17F.1    sit    3    90     100    80    80   106  
13:32:21.591 B17F.0    stand  3    170    30     0     80   106  
13:32:22.593 B17F.0    stand  3    170    30     0     80   0    
13:32:22.593 B17F.1    sit    3    50     100    69    80   138  
13:32:23.593 B17F.1    sit    3    40     80     67    80   22   
13:32:23.593 B17F.0    stand  3    170    30     0     80   139  
13:32:24.597 B17F.0    stand  3    170    30     0     80   0    
13:32:24.597 B17F.1    sit    3    40     90     71    80   143  
13:32:25.596 B17F.0    stand  3    170    30     0     80   143  
13:32:25.596 B17F.1    sit    3    40     90     78    80   143  
13:32:26.600 B17F.0    stand  3    170    30     0     80   143  
13:32:26.600 B17F.1    sit    3    60     80     70    80   120  
13:32:27.598 B17F.0    stand  3    170    30     0     80   120  
13:32:27.598 B17F.1    sit    3    50     90     84    80   134  
13:32:28.600 B17F.1    sit    3    50     90     77    80   0    
13:32:28.600 B17F.0    stand  3    170    30     0     80   134  
13:32:29.606 B17F.0    stand  3    170    30     0     80   0    
13:32:29.606 B17F.1    sit    3    60     90     73    80   125  
13:32:30.601 B17F.0    stand  3    170    30     0     80   125  
13:32:30.601 B17F.1    sit    3    60     90     72    80   125  
13:32:31.607 B17F.0    stand  3    170    30     0     80   125  
13:32:31.607 B17F.1    sit    3    60     90     70    80   125  
13:32:32.603 B17F.1    sit    3    60     90     67    80   0    
13:32:32.603 B17F.0    stand  3    170    30     0     80   125  
13:32:33.509 B17F.1    sit    3    70     90     67    80   116  
13:32:33.509 B17F.0    stand  3    170    30     0     80   116  
13:32:34.501 B17F.1    sit    3    70     100    81    80   122  
13:32:34.501 B17F.0    stand  3    170    30     0     80   122  
13:32:35.504 B17F.0    stand  3    170    30     0     80   0    
13:32:35.504 B17F.1    sit    3    50     70     74    80   126  
13:32:36.506 B17F.0    stand  3    170    30     0     80   126  
13:32:36.506 B17F.1    walk   3    -10    100    85    80   193  
13:32:37.505 B17F.0    stand  3    170    30     0     80   193  
13:32:37.505 B17F.1    walk   3    -120   60     68    80   291  
13:32:38.504 B17F.1    walk   3    -140   110    91    80   53   
13:32:38.504 B17F.0    stand  3    170    30     0     80   320  
13:32:39.506 B17F.1    walk   3    -140   120    103   80   322  
13:32:39.506 B17F.0    stand  3    170    30     0     80   322  
13:32:40.568 B17F.1    walk   3    -140   130    0     80   325  
13:32:40.568 B17F.2    stand  255  60     80     70    80   206  
13:32:40.568 B17F.0    stand  3    170    30     0     80   120  
13:32:41.517 B17F.0    stand  3    170    30     0     80   0    
13:32:41.517 B17F.1    walk   3    -130   110    114   80   310  
13:32:41.517 B17F.2    stand  255  50     90     0     80   181  
13:32:42.517 B17F.1    walk   3    -150   120    108   80   202  
13:32:42.517 B17F.2    stand  255  40     90     0     80   192  
13:32:42.517 B17F.0    stand  3    170    30     0     80   143  
13:32:43.416 B17F.1    walk   3    -140   120    0     80   322  
13:32:43.416 B17F.0    stand  3    170    30     0     80   322  
13:32:43.416 B17F.2    stand  255  50     80     83    80   130  
13:32:44.417 B17F.0    stand  3    170    30     0     80   130  
13:32:44.417 B17F.2    stand  255  50     90     0     80   134  
13:32:44.417 B17F.1    walk   3    -130   120    118   80   182  
13:32:45.420 B17F.0    stand  3    170    30     0     80   313  
13:32:45.420 B17F.2    stand  255  40     80     74    80   139  
13:32:45.420 B17F.1    walk   3    -150   140    0     80   199  
13:32:46.420 B17F.2    stand  255  60     90     0     80   215  
13:32:46.420 B17F.1    walk   3    -170   140    106   80   235  
13:32:46.420 B17F.0    stand  3    170    30     0     80   357  
13:32:47.419 B17F.0    stand  3    170    30     0     80   0    
13:32:47.419 B17F.2    stand  255  50     80     0     80   130  
13:32:47.419 B17F.1    walk   3    -140   70     89    80   190  
13:32:48.421 B17F.0    stand  3    120    20     0     80   264  
13:32:48.421 B17F.2    stand  255  50     80     0     80   92   
13:32:48.421 B17F.1    walk   3    -90    40     82    80   145  
13:32:49.426 B17F.0    stand  3    100    10     0     80   192  
13:32:49.426 B17F.1    walk   3    -60    80     0     80   174  
13:32:49.426 B17F.2    stand  255  60     90     0     80   120  
13:32:50.398 B17F.1    walk   3    -50    110    0     80   111  
13:32:50.398 B17F.0    stand  3    100    10     0     80   180  
13:32:50.398 B17F.2    stand  255  90     120    80    80   110  
13:32:51.369 B17F.2    stand  255  110    130    55    80   22   
13:32:51.369 B17F.0    stand  3    100    10     0     80   120  
13:32:51.369 B17F.1    sit    3    -40    120    0     80   178  
13:32:52.376 B17F.1    sit    3    -40    130    0     80   10   
13:32:52.376 B17F.0    stand  3    100    10     0     80   184  
13:32:52.376 B17F.2    walk   255  130    90     36    80   85   
13:32:53.368 B17F.1    sit    3    -40    130    0     80   174  
13:32:53.368 B17F.0    stand  3    100    10     0     80   184  
13:32:53.368 B17F.2    walk   255  130    120    59    80   114  
13:32:54.418 B17F.1    sit    3    -40    130    0     80   170  
13:32:54.418 B17F.2    sit    255  130    80     51    80   177  
13:32:54.418 B17F.0    stand  3    100    10     0     80   76   
13:32:54.657 B17F.0    stand  3    100    10     0     80   0    
13:32:54.657 B17F.2    sitgnd 255  130    80     56    80   76   
13:32:54.657 B17F.1    sit    3    -30    130    0     80   167  
13:32:55.496 B17F.2    sit    255  130    90     65    80   164  
13:32:55.496 B17F.1    sit    3    -30    130    0     80   164  
13:32:55.496 B17F.0    stand  3    100    10     0     80   176  
13:32:56.313 B17F.1    sit    3    -30    140    0     80   183  
13:32:56.313 B17F.2    sit    255  120    80     0     80   161  
13:32:56.313 B17F.0    stand  3    100    10     0     80   72   
13:32:57.312 B17F.0    stand  3    100    10     0     80   0    
13:32:57.312 B17F.2    sit    255  120    80     0     80   72   
13:32:57.312 B17F.1    sit    3    -30    140    0     80   161  
13:32:58.313 B17F.2    sit    255  130    100    73    80   164  
13:32:58.313 B17F.1    sit    3    -30    140    0     80   164  
13:32:58.313 B17F.0    stand  3    100    10     0     80   183  
13:32:59.321 B17F.1    sit    3    -30    140    0     80   183  
13:32:59.321 B17F.2    sit    255  130    120    81    80   161  
13:32:59.321 B17F.0    stand  3    100    10     0     80   114  
13:33:00.317 B17F.0    stand  3    100    10     0     80   0    
13:33:00.317 B17F.1    sit    3    -30    140    0     80   183  
13:33:00.317 B17F.2    sit    255  130    130    0     80   160  
13:33:01.317 B17F.1    sit    3    -30    140    0     80   160  
13:33:01.317 B17F.2    sit    255  120    130    47    80   150  
13:33:01.317 B17F.0    stand  3    100    10     0     80   121  
13:33:02.318 B17F.1    sit    3    -30    140    0     80   183  
13:33:02.318 B17F.2    sit    255  140    110    0     80   172  
13:33:02.318 B17F.0    stand  3    100    10     0     80   107  
13:33:03.324 B17F.2    sit    255  130    120    0     80   114  
13:33:03.324 B17F.0    stand  3    100    10     0     80   114  
13:33:03.324 B17F.1    sit    3    -30    140    0     80   183  
13:33:04.325 B17F.2    sit    255  130    60     59    80   178  
13:33:04.325 B17F.1    sit    3    -30    140    0     80   178  
13:33:04.325 B17F.0    stand  3    100    10     0     80   183  
13:33:05.329 B17F.0    stand  3    100    10     0     80   0    
13:33:05.329 B17F.2    sit    255  140    70     47    80   72   
13:33:05.329 B17F.1    sit    3    -30    140    0     80   183  
13:33:06.333 B17F.0    stand  3    100    10     0     80   183  
13:33:06.333 B17F.1    sit    3    -30    140    0     80   183  
13:33:06.333 B17F.2    sit    255  130    90     81    80   167  
13:33:07.328 B17F.2    sit    255  90     100    71    80   41   
13:33:07.328 B17F.0    stand  3    100    10     0     80   90   
13:33:07.328 B17F.1    sit    3    -30    140    0     80   183  
13:33:08.217 B17F.0    stand  3    100    10     0     80   183  
13:33:08.217 B17F.2    sit    255  80     100    80    80   92   
13:33:08.217 B17F.1    sit    3    -30    140    0     80   117  
13:33:09.276 B17F.0    stand  3    100    10     0     80   183  
13:33:09.276 B17F.2    sit    255  90     90     0     80   80   
13:33:10.229 B17F.0    stand  3    100    10     0     80   80   
13:33:10.229 B17F.2    sit    255  60     80     65    80   80   
13:33:11.230 B17F.0    stand  3    100    10     0     80   80   
13:33:11.230 B17F.2    sit    255  100    100    0     80   90   
13:33:12.232 B17F.0    stand  3    100    10     0     80   90   
13:33:12.232 B17F.2    sit    255  70     90     0     80   85   
13:33:13.234 B17F.2    sit    255  80     90     72    80   10   
13:33:13.234 B17F.0    stand  3    100    10     0     80   82   
13:33:14.232 B17F.2    sit    255  100    110    68    80   100  
13:33:14.232 B17F.0    stand  3    100    10     0     80   100  
13:33:15.234 B17F.2    sit    255  90     110    0     80   100  
13:33:15.234 B17F.0    stand  3    100    10     0     80   100  
13:33:16.236 B17F.2    sit    255  70     90     64    80   85   
13:33:16.236 B17F.0    stand  3    100    10     0     80   85   
13:33:17.234 B17F.2    sit    255  70     90     47    80   85   
13:33:17.234 B17F.0    stand  3    100    10     0     80   85   
13:33:18.237 B17F.0    stand  3    100    10     0     80   0    
13:33:18.237 B17F.2    sit    255  70     90     23    80   85   
13:33:19.137 B17F.0    stand  3    100    10     0     80   85   
13:33:19.137 B17F.2    sit    255  70     90     25    80   85   
13:33:20.131 B17F.2    sit    255  70     90     40    80   0    
13:33:20.131 B17F.0    stand  3    100    10     0     80   85   
13:33:21.130 B17F.2    sit    255  60     90     10    80   89   
13:33:21.130 B17F.0    stand  3    100    10     0     80   89   
13:33:22.136 B17F.2    sit    255  60     80     37    80   80   
13:33:22.136 B17F.0    stand  3    100    10     0     80   80   
13:33:23.138 B17F.0    stand  3    100    10     0     80   0    
13:33:23.138 B17F.2    sit    255  80     100    46    80   92   
13:33:24.138 B17F.2    sit    255  90     100    0     80   10   
13:33:24.138 B17F.0    stand  3    100    10     0     80   90   
13:33:25.140 B17F.2    sit    255  80     100    0     80   92   
13:33:25.140 B17F.0    stand  3    100    10     0     80   92   
13:33:26.140 B17F.2    sit    255  120    110    0     80   101  
13:33:26.140 B17F.0    stand  3    100    10     0     80   101  
13:33:27.141 B17F.0    stand  3    100    10     0     80   0    
13:33:27.141 B17F.2    stand  255  90     100    67    80   90   
13:33:28.142 B17F.2    stand  255  50     80     67    80   44   
13:33:28.142 B17F.0    stand  3    100    10     0     80   86   
13:33:29.143 B17F.2    stand  255  100    110    74    80   100  
13:33:29.143 B17F.0    stand  3    100    10     0     80   100  
13:33:30.040 B17F.2    stand  255  140    120    0     80   117  
13:33:30.040 B17F.0    stand  3    100    10     0     80   117  
13:33:31.244 B17F.2    stand  255  130    120    79    80   114  
13:33:32.060 B17F.2    stand  255  130    120    0     80   0    
13:33:33.052 B17F.2    stand  255  130    120    0     80   0    
13:33:34.052 B17F.2    stand  255  130    100    0     80   20   
13:33:35.057 B17F.2    stand  255  80     100    73    80   50   
13:33:36.054 B17F.2    stand  255  50     80     0     80   36   
13:33:37.060 B17F.2    stand  255  130    100    84    80   82   
13:33:38.064 B17F.2    stand  255  110    110    71    80   22   
13:33:39.060 B17F.2    stand  255  120    110    0     80   10   
13:33:40.061 B17F.2    stand  255  130    120    0     80   14   
13:33:40.957 B17F.2    stand  255  130    120    0     80   0    
13:33:41.956 B17F.2    stand  255  130    110    82    80   10   
13:33:42.957 B17F.2    stand  255  130    120    0     80   10   
13:33:43.954 B17F.2    stand  255  140    130    0     80   14   
13:33:44.956 B17F.2    stand  255  50     80     63    80   102  
13:33:45.956 B17F.2    walk   255  50     80     34    80   0    
13:33:46.960 B17F.2    walk   255  50     80     15    80   0    
13:33:47.963 B17F.2    walk   255  50     80     32    80   0    
13:33:48.959 B17F.2    walk   255  50     80     64    80   0    
13:33:49.966 B17F.2    stand  255  60     80     0     80   10   
13:33:50.960 B17F.2    stand  255  50     80     66    80   10   
13:33:51.975 B17F.2    stand  255  50     80     66    80   0    
13:33:52.862 B17F.2    stand  255  50     80     67    80   0    
13:33:53.856 B17F.2    stand  255  60     80     0     80   10   
13:33:54.894 B17F.2    stand  255  60     80     0     80   0    
13:33:55.840 B17F.2    stand  255  100    100    81    80   44   
13:33:56.846 B17F.2    stand  255  120    100    51    80   20   
13:33:57.840 B17F.2    stand  255  120    80     0     80   20   
13:33:58.841 B17F.2    stand  255  60     90     64    80   60   
13:33:59.849 B17F.2    stand  255  60     90     0     80   0    
13:34:00.849 B17F.2    stand  255  70     90     0     80   10   
13:34:01.857 B17F.2    stand  255  50     80     64    80   22   
13:34:02.848 B17F.2    stand  255  90     100    70    80   44   
13:34:03.856 B17F.2    stand  255  90     100    74    80   0    
13:34:04.850 B17F.2    stand  255  50     80     0     80   44   
13:34:05.850 B17F.2    stand  255  50     80     62    80   0    
13:34:06.751 B17F.2    stand  255  50     80     62    80   0    
13:34:07.746 B17F.2    stand  255  50     80     70    80   0    
13:34:08.753 B17F.2    stand  255  60     90     65    80   14   
13:34:09.746 B17F.2    stand  255  70     110    79    80   22   
13:34:10.768 B17F.2    stand  255  50     90     0     80   28   
13:34:11.780 B17F.2    stand  255  50     90     0     80   0    
13:34:12.777 B17F.2    stand  255  50     90     0     80   0    
13:34:13.774 B17F.2    stand  255  60     90     55    80   10   
13:34:14.777 B17F.2    stand  255  100    90     69    80   40   
13:34:15.673 B17F.2    stand  255  50     80     0     80   50   
13:34:16.674 B17F.2    stand  255  50     70     0     80   10   
13:34:17.676 B17F.2    stand  255  50     90     0     80   20   
13:34:18.674 B17F.2    sit    255  50     90     0     80   0    
13:34:19.674 B17F.2    sit    255  50     90     0     80   0    
13:34:20.676 B17F.2    sit    255  50     90     0     80   0    
13:34:21.678 B17F.2    sit    255  140    100    82    80   90   
13:34:22.685 B17F.2    sit    255  100    90     74    80   41   
13:34:23.687 B17F.2    sit    255  120    90     57    80   20   
13:34:24.679 B17F.2    sit    255  130    110    59    80   22   
13:34:25.684 B17F.2    sit    255  120    100    81    80   14   
13:34:26.584 B17F.2    sit    255  140    120    0     80   28   
13:34:27.584 B17F.2    sit    255  140    120    0     80   0    
13:34:28.587 B17F.2    sit    255  130    120    0     80   10   
13:34:29.591 B17F.2    sit    255  90     100    73    80   44   
13:34:30.594 B17F.2    sit    255  50     80     0     80   44   
13:34:31.589 B17F.2    sit    255  50     80     0     80   0    
13:34:32.594 B17F.2    sit    255  130    120    0     80   89   
13:34:33.594 B17F.2    sit    255  130    130    0     80   10   
13:34:34.592 B17F.2    sit    255  130    120    71    80   10   
13:34:35.592 B17F.2    sit    255  80     90     59    80   58   
13:34:36.592 B17F.2    sit    255  50     80     41    80   31   
13:34:37.594 B17F.2    sit    255  50     90     0     80   10   
13:34:38.497 B17F.2    sit    255  50     80     32    80   10   
13:34:39.493 B17F.2    sit    255  70     90     67    80   22   
13:34:40.492 B17F.2    sit    255  70     90     74    80   0    
13:34:41.491 B17F.2    sit    255  60     100    76    80   14   
13:34:42.460 B17F.2    sit    255  70     90     77    80   14   
13:34:43.460 B17F.2    sit    255  100    120    78    80   42   
13:34:44.459 B17F.2    sit    255  60     90     74    80   50   
13:34:45.458 B17F.2    sit    255  70     90     0     80   10   
13:34:46.462 B17F.2    sit    255  50     90     72    80   20   
13:34:47.467 B17F.2    sit    255  90     100    0     80   41   
13:34:48.468 B17F.2    sit    255  50     80     0     80   44   
13:34:49.469 B17F.2    sit    255  60     90     69    80   14   
13:34:50.466 B17F.2    sit    255  50     80     65    80   14   
13:34:51.474 B17F.2    sit    255  80     100    70    80   36   
13:34:52.475 B17F.2    sit    255  70     90     62    80   14   
13:34:53.522 B17F.2    sit    255  80     90     77    80   10   
13:34:54.365 B17F.2    sit    255  90     90     69    80   10   
13:34:55.362 B17F.2    sit    255  70     90     68    80   20   
13:34:56.368 B17F.2    sit    255  90     100    65    80   22   
13:34:57.362 B17F.2    sit    255  100    80     0     80   22   
13:34:58.407 B17F.2    sit    255  80     90     72    80   22   
13:34:59.412 B17F.2    sit    255  50     80     0     80   31   
13:35:00.415 B17F.2    sit    255  70     80     0     80   20   
13:35:01.305 B17F.2    sit    255  50     80     0     80   20   
13:35:02.306 B17F.2    sit    255  100    100    84    80   53   
13:35:03.305 B17F.2    sit    255  50     80     0     80   53   
13:35:04.306 B17F.2    sit    255  50     90     66    80   10   
13:35:05.315 B17F.2    sit    255  60     80     83    80   14   
13:35:06.314 B17F.2    sit    255  60     100    66    80   20   
13:35:07.315 B17F.2    sit    255  70     100    65    80   10   
13:35:08.312 B17F.2    sit    255  60     100    84    80   10   
13:35:09.316 B17F.2    sit    255  60     100    0     80   0    
13:35:10.314 B17F.2    sit    255  130    120    0     80   72   
13:35:11.317 B17F.2    sit    255  80     110    0     80   50   
13:35:12.320 B17F.2    sit    255  80     110    0     80   0    
13:35:13.208 B17F.2    sit    255  80     110    0     80   0    
13:35:14.221 B17F.2    sit    255  70     110    0     80   10   
13:35:15.215 B17F.2    sit    255  80     120    0     80   14   
13:35:16.222 B17F.2    sit    255  110    110    0     80   31   
13:35:17.216 B17F.2    sit    255  110    110    0     80   0    
13:35:18.217 B17F.2    sit    255  110    110    0     80   0    
13:35:19.221 B17F.2    sit    255  90     110    80    80   20   
13:35:20.220 B17F.2    sit    255  60     90     0     80   36   
13:35:21.219 B17F.2    sit    255  80     110    0     80   28   
13:35:22.221 B17F.2    sit    255  70     110    0     80   10   
13:35:23.229 B17F.2    sit    255  60     80     70    80   31   
13:35:24.124 B17F.2    sit    255  60     80     0     80   0    
13:35:25.138 B17F.2    sit    255  60     80     0     80   0    
13:35:26.122 B17F.2    sit    255  60     80     0     80   0    
13:35:27.122 B17F.2    sit    255  60     80     0     80   0    
13:35:28.122 B17F.2    sit    255  60     80     0     80   0    
13:35:29.129 B17F.2    sit    255  60     80     0     80   0    
13:35:30.132 B17F.2    sit    255  60     80     0     80   0    
13:35:31.125 B17F.2    sit    255  60     80     0     80   0    
13:35:32.125 B17F.2    sit    255  60     80     0     80   0    
13:35:33.136 B17F.2    sit    255  60     80     0     80   0    
13:35:34.128 B17F.2    sit    255  60     80     0     80   0    
13:35:35.134 B17F.2    sit    255  60     80     0     80   0    
13:35:36.030 B17F.2    sit    255  60     80     0     80   0    
13:35:37.027 B17F.2    sit    255  60     90     75    80   10   
13:35:38.028 B17F.2    sit    255  60     80     70    80   10   
13:35:39.027 B17F.2    sit    255  60     100    0     80   20   
13:35:40.028 B17F.2    sit    255  70     80     0     80   22   
13:35:41.029 B17F.2    sit    255  70     80     0     80   0    
13:35:42.028 B17F.2    sit    255  100    90     0     80   31   
13:35:43.031 B17F.2    sit    255  60     80     0     80   41   
13:35:44.040 B17F.2    sit    255  50     80     0     80   10   
13:35:45.032 B17F.2    sit    255  130    110    0     80   85   
13:35:46.038 B17F.2    sit    255  90     100    69    80   41   
13:35:46.943 B17F.2    sit    255  60     80     0     80   36   
13:35:47.937 B17F.2    sit    255  50     80     0     80   10   
13:35:48.936 B17F.2    sit    255  50     80     0     80   0    
13:35:49.942 B17F.2    sit    255  90     110    0     80   50   
13:35:50.945 B17F.2    sit    255  120    120    0     80   31   
13:35:51.940 B17F.2    sit    255  80     110    84    80   41   
13:35:52.990 B17F.2    sit    255  100    100    75    80   22   
13:35:53.946 B17F.2    sit    255  70     110    0     80   31   
13:35:54.958 B17F.2    sit    255  100    100    77    80   31   
13:35:55.954 B17F.2    sit    255  120    120    0     80   28   
13:35:56.951 B17F.2    sit    255  130    100    0     80   22   
13:35:57.947 B17F.2    sit    255  60     90     75    80   70   
13:35:58.838 B17F.2    sit    255  120    120    0     80   67   
13:35:59.840 B17F.2    sit    255  110    110    0     80   14   
13:36:00.841 B17F.2    sit    255  50     110    0     80   60   
13:36:01.844 B17F.2    sit    255  50     100    0     80   10   
13:36:02.853 B17F.2    sit    255  50     80     0     80   20   
13:36:03.850 B17F.2    sit    255  50     80     0     80   0    
13:36:04.857 B17F.2    sit    255  50     80     0     80   0    
13:36:05.851 B17F.2    sit    255  100    100    53    80   53   
13:36:06.852 B17F.2    sit    255  90     100    48    80   10   
13:36:07.857 B17F.2    sit    255  50     80     0     80   44   
13:36:08.853 B17F.2    sit    255  60     80     0     80   10   
13:36:09.752 B17F.2    sit    255  100    90     0     80   41   
13:36:10.755 B17F.2    sit    255  120    120    0     80   36   
13:36:11.753 B17F.2    sit    255  130    110    0     80   14   
13:36:12.760 B17F.2    sit    255  120    110    0     80   10   
13:36:13.767 B17F.2    sit    255  120    110    0     80   0    
13:36:14.757 B17F.2    sit    255  50     90     0     80   72   
13:36:15.758 B17F.2    sit    255  50     100    0     80   10   
13:36:16.765 B17F.2    sit    255  60     100    16    80   10   
13:36:17.759 B17F.2    sit    255  60     100    17    80   0    
13:36:18.760 B17F.2    sit    255  60     100    0     80   0    
13:36:19.762 B17F.2    sit    255  60     100    17    80   0    
13:36:20.762 B17F.2    sit    255  60     100    19    80   0    
13:36:21.656 B17F.2    sit    255  60     100    55    80   0    
13:36:22.658 B17F.2    sit    255  130    110    0     80   70   
13:36:23.661 B17F.2    sit    255  130    100    19    80   10   
13:36:24.658 B17F.2    sit    255  70     80     50    80   63   
13:36:25.660 B17F.2    sit    255  50     80     25    80   20   
13:36:26.662 B17F.2    sit    255  50     80     0     80   0    
13:36:27.660 B17F.2    sit    255  50     80     39    80   0    
13:36:28.665 B17F.2    sit    255  50     80     0     80   0    
13:36:29.662 B17F.2    sit    255  50     80     0     80   0    
13:36:30.664 B17F.2    sit    255  50     80     0     80   0    
13:36:31.665 B17F.2    sit    255  50     80     0     80   0    
13:36:32.666 B17F.2    sit    255  50     80     0     80   0    
13:36:33.562 B17F.2    sit    255  50     80     0     80   0    
13:36:34.565 B17F.2    sit    255  50     80     0     80   0    
13:36:35.561 B17F.2    sit    255  50     80     0     80   0    
13:36:36.560 B17F.2    sit    255  50     80     0     80   0    
13:36:37.564 B17F.2    sit    255  50     80     0     80   0    
13:36:38.584 B17F.2    sit    255  50     80     0     80   0    
13:36:39.573 B17F.2    sit    255  50     80     0     80   0    
13:36:40.572 B17F.2    sit    255  50     80     0     80   0    
13:36:41.568 B17F.2    sit    255  50     80     0     80   0    
13:36:42.566 B17F.2    sit    255  50     80     0     80   0    
13:36:43.573 B17F.2    sit    255  50     80     0     80   0    
13:36:44.569 B17F.2    sit    255  50     80     0     80   0    
13:36:45.462 B17F.2    sit    255  50     80     0     80   0    
13:36:46.468 B17F.2    sit    255  50     80     0     80   0    
13:36:47.468 B17F.2    sit    255  50     80     0     80   0    
13:36:48.469 B17F.2    sit    255  50     80     0     80   0    
13:36:49.478 B17F.2    sit    255  50     80     0     80   0    
13:36:50.490 B17F.2    sit    255  50     80     0     80   0    
13:36:51.485 B17F.2    sit    255  50     80     0     80   0    
13:36:52.539 B17F.2    sit    255  50     80     0     80   0    
13:36:53.484 B17F.2    sit    255  50     80     0     80   0    
13:36:54.492 B17F.2    sit    255  60     90     75    80   14   
13:36:55.393 B17F.2    sit    255  60     100    0     80   10   
13:36:56.388 B17F.2    sit    255  60     100    71    80   0    
13:36:57.390 B17F.2    sit    255  50     80     65    80   22   
13:36:58.387 B17F.2    sit    255  50     80     65    80   0    
13:36:59.388 B17F.2    sit    255  50     80     62    80   0    
13:37:00.390 B17F.2    sit    255  50     90     0     80   10   
13:37:01.393 B17F.2    sit    255  50     90     0     80   0    
13:37:02.392 B17F.2    sit    255  50     80     0     80   10   
13:37:03.394 B17F.2    sit    255  50     90     0     80   10   
13:37:04.395 B17F.2    sit    255  50     80     0     80   10   
13:37:05.405 B17F.2    sit    255  70     90     69    80   22   
13:37:06.402 B17F.2    sit    255  130    120    0     80   67   
13:37:07.288 B17F.2    sit    255  100    100    79    80   36   
13:37:08.294 B17F.2    sit    255  50     80     66    80   53   
13:37:09.289 B17F.2    sit    255  60     90     73    80   14   
13:37:10.292 B17F.2    sit    255  50     80     0     80   14   
13:37:11.293 B17F.2    sit    255  50     80     0     80   0    
13:37:12.297 B17F.2    sit    255  50     80     0     80   0    
13:37:13.296 B17F.2    sit    255  70     90     0     80   22   
13:37:14.297 B17F.2    sit    255  60     90     77    80   10   
13:37:15.297 B17F.2    sit    255  50     80     72    80   14   
13:37:16.298 B17F.2    sit    255  60     90     86    80   14   
13:37:17.300 B17F.2    sit    255  70     90     72    80   10   
13:37:18.298 B17F.2    sit    255  80     110    87    80   22   
13:37:19.192 B17F.2    sit    255  60     90     74    80   28   
13:37:20.202 B17F.2    sit    255  90     110    0     80   36   
13:37:21.195 B17F.2    sit    255  120    100    0     80   31   
13:37:22.200 B17F.2    sit    255  140    100    0     80   20   
13:37:23.197 B17F.2    sit    255  130    110    0     80   14   
13:37:24.196 B17F.2    sit    255  130    110    0     80   0    
13:37:25.201 B17F.2    sit    255  130    110    0     80   0    
13:37:26.196 B17F.2    sit    255  70     90     69    80   63   
13:37:27.210 B17F.2    sit    255  50     100    97    80   22   
13:37:28.205 B17F.2    sit    255  60     90     76    80   14   
13:37:29.206 B17F.2    sit    255  110    90     61    80   50   
13:37:30.202 B17F.2    sit    255  140    90     50    80   30   
13:37:31.103 B17F.2    sit    255  120    80     62    80   22   
13:37:32.103 B17F.2    sit    255  90     80     78    80   30   
13:37:33.101 B17F.2    walk   255  30     70     94    80   60   
13:37:34.102 B17F.2    walk   255  -10    70     75    80   40   
13:37:35.108 B17F.2    walk   255  -20    100    91    80   31   
13:37:36.105 B17F.2    walk   255  70     80     85    80   92   
13:37:37.104 B17F.2    sit    255  90     90     76    80   22   
13:37:38.106 B17F.2    sit    255  90     100    63    80   10   
13:37:39.119 B17F.2    sit    255  90     90     71    80   10   
13:37:40.120 B17F.2    sit    255  60     100    64    80   31   
13:37:41.122 B17F.2    sit    255  60     90     73    80   10   
13:37:42.010 B17F.2    sit    255  60     90     72    80   0    
13:37:43.009 B17F.2    sit    255  60     90     85    80   0    
13:37:44.012 B17F.2    sit    255  60     90     0     80   0    
13:37:45.024 B17F.2    sit    255  60     90     0     80   0    
13:37:46.012 B17F.2    sit    255  140    130    0     80   89   
13:37:47.013 B17F.2    stand  255  120    90     81    80   44   
13:37:48.028 B17F.2    stand  255  90     90     0     80   30   
13:37:49.021 B17F.2    stand  255  90     90     0     80   0    
13:37:50.024 B17F.2    stand  255  150    130    67    80   72   
13:37:51.022 B17F.2    stand  255  140    110    0     80   22   
13:37:52.072 B17F.2    stand  255  140    130    0     80   20   
13:37:53.024 B17F.2    stand  255  90     120    76    80   50   
13:37:53.940 B17F.2    walk   255  30     90     68    80   67   
13:37:54.924 B17F.2    walk   255  50     100    71    80   22   
13:37:55.928 B17F.2    walk   255  100    100    69    80   50   
13:37:56.929 B17F.2    walk   255  120    100    76    80   20   
13:37:57.929 B17F.2    walk   255  120    90     0     80   10   
13:37:58.943 B17F.2    stand  255  90     90     77    80   30   
13:37:59.932 B17F.2    stand  255  130    100    84    80   41   
13:38:00.930 B17F.2    stand  255  80     100    71    80   50   
13:38:01.942 B17F.2    stand  255  130    120    0     80   53   
13:38:02.932 B17F.2    stand  255  70     120    0     80   60   
13:38:03.833 B17F.2    stand  255  80     110    0     80   14   
13:38:04.833 B17F.2    stand  255  80     110    0     80   0    
13:38:05.834 B17F.2    stand  255  50     80     65    80   42   
13:38:06.840 B17F.2    stand  255  50     80     0     80   0    
13:38:07.860 B17F.2    stand  255  60     80     0     80   10   
13:38:08.838 B17F.2    stand  255  130    110    0     80   76   
13:38:09.843 B17F.2    stand  255  130    110    0     80   0    
13:38:10.776 B17F.2    stand  255  130    110    0     80   0    
13:38:11.784 B17F.2    stand  255  50     80     64    80   85   
13:38:12.785 B17F.2    stand  255  60     90     76    80   14   
13:38:13.787 B17F.2    stand  255  70     100    73    80   14   
13:38:14.785 B17F.2    stand  255  70     100    81    80   0    
13:38:15.781 B17F.2    stand  255  60     100    80    80   10   
13:38:16.788 B17F.2    stand  255  50     90     70    80   14   
13:38:17.785 B17F.2    stand  255  60     90     71    80   10   
13:38:18.785 B17F.2    stand  255  60     90     86    80   0    
13:38:19.787 B17F.2    stand  255  70     100    75    80   14   
13:38:20.786 B17F.2    stand  255  70     90     82    80   10   
13:38:21.788 B17F.2    stand  255  60     100    91    80   14   
13:38:22.680 B17F.2    sit    255  80     110    77    80   22   
13:38:23.686 B17F.2    sit    255  100    110    82    80   20   
13:38:24.687 B17F.2    sit    255  130    110    0     80   30   
13:38:25.685 B17F.2    sit    255  100    130    77    80   36   
13:38:26.748 B17F.2    sit    255  90     120    75    80   14   
13:38:27.640 B17F.2    sit    255  70     120    0     80   20   
13:38:28.642 B17F.2    sit    255  70     110    0     80   10   
13:38:29.648 B17F.2    sit    255  70     110    0     80   0    
13:38:30.645 B17F.2    sit    255  70     110    0     80   0    
13:38:31.645 B17F.2    sit    255  120    130    0     80   53   
13:38:32.644 B17F.2    sit    255  120    130    0     80   0    
13:38:33.648 B17F.2    sit    255  120    130    0     80   0    
13:38:34.650 B17F.2    sit    255  90     110    83    80   36   
13:38:35.646 B17F.2    sit    255  110    130    76    80   28   
13:38:36.650 B17F.2    sit    255  100    100    85    80   31   
13:38:37.649 B17F.2    sit    255  60     100    0     80   40   
13:38:38.652 B17F.2    sit    255  50     80     0     80   22   
13:38:39.545 B17F.2    sit    255  60     80     0     80   10   
13:38:40.545 B17F.2    sit    255  60     80     0     80   0    
13:38:41.553 B17F.2    sit    255  60     80     0     80   0    
13:38:42.580 B17F.2    sit    255  60     80     0     80   0    
13:38:43.560 B17F.2    sit    255  60     80     0     80   0    
13:38:44.562 B17F.2    sit    255  60     80     0     80   0    
13:38:45.561 B17F.2    sit    255  60     90     90    80   10   
13:38:46.568 B17F.2    sit    255  50     80     82    80   14   
13:38:47.565 B17F.2    sit    255  60     80     85    80   10   
13:38:48.565 B17F.2    sit    255  40     80     0     80   20   
13:38:49.468 B17F.2    sit    255  40     80     0     80   0    
13:38:50.466 B17F.2    sit    255  50     90     0     80   14   
13:38:51.518 B17F.2    sit    255  60     100    84    80   14   
13:38:52.468 B17F.2    sit    255  70     100    79    80   10   
13:38:53.468 B17F.2    sit    255  60     100    0     80   10   
13:38:54.474 B17F.2    sit    255  60     90     0     80   10   
13:38:55.480 B17F.2    sit    255  140    120    0     80   85   
13:38:56.476 B17F.2    sit    255  140    120    0     80   0    
13:38:57.474 B17F.2    sit    255  100    130    0     80   41   
13:38:58.392 B17F.2    sit    255  70     130    0     80   30   
13:38:59.403 B17F.2    sit    255  70     130    0     80   0    
13:39:00.395 B17F.2    sit    255  70     130    0     80   0    
13:39:01.396 B17F.2    sit    255  70     130    0     80   0    
13:39:02.397 B17F.2    sit    255  70     130    0     80   0    
13:39:03.397 B17F.2    sit    255  70     130    0     80   0    
13:39:04.398 B17F.2    sit    255  70     130    0     80   0    
13:39:05.398 B17F.2    sit    255  80     100    78    80   31   
13:39:06.398 B17F.2    sit    255  60     80     0     80   28   
13:39:07.401 B17F.2    sit    255  80     90     0     80   22   
13:39:08.402 B17F.2    sit    255  130    110    0     80   53   
13:39:09.407 B17F.2    sit    255  130    110    94    80   0    
13:39:10.298 B17F.2    sit    255  120    120    0     80   14   
13:39:11.300 B17F.2    sit    255  110    130    0     80   14   
13:39:12.303 B17F.2    sit    255  90     100    90    80   36   
13:39:13.298 B17F.2    sit    255  80     80     0     80   22   
13:39:14.269 B17F.2    sit    255  80     90     0     80   10   
13:39:15.276 B17F.2    sit    255  60     90     0     80   20   
13:39:16.271 B17F.2    sit    255  60     90     0     80   0    
13:39:17.270 B17F.2    sit    255  70     90     0     80   10   
13:39:18.268 B17F.2    sit    255  70     90     0     80   0    
13:39:19.268 B17F.2    sit    255  70     90     0     80   0    
13:39:20.277 B17F.2    sit    255  70     90     0     80   0    
13:39:21.272 B17F.2    sit    255  70     90     0     80   0    
13:39:22.276 B17F.2    sit    255  70     90     0     80   0    
13:39:23.274 B17F.2    sit    255  70     90     0     80   0    
13:39:24.282 B17F.2    sit    255  70     90     0     80   0    
13:39:25.277 B17F.2    sit    255  70     90     0     80   0    
13:39:26.170 B17F.2    sit    255  70     90     0     80   0    
13:39:27.168 B17F.2    sit    255  70     90     0     80   0    
13:39:28.172 B17F.2    sit    255  70     90     0     80   0    
13:39:29.172 B17F.2    sit    255  70     90     70    80   0    
13:39:30.176 B17F.2    sit    255  70     100    78    80   10   
13:39:31.196 B17F.2    sit    255  70     90     75    80   10   
13:39:32.198 B17F.2    sit    255  60     80     71    80   14   
13:39:33.195 B17F.2    sit    255  60     100    84    80   20   
13:39:34.196 B17F.2    sit    255  60     80     0     80   20   
13:39:35.202 B17F.2    sit    255  60     100    87    80   20   
13:39:36.094 B17F.2    sit    255  120    120    84    80   63   
13:39:37.105 B17F.2    sit    255  80     110    76    80   41   
13:39:38.092 B17F.2    sit    255  100    110    69    80   20   
13:39:39.093 B17F.2    sit    255  80     100    79    80   22   
13:39:40.097 B17F.2    sit    255  60     90     74    80   22   
13:39:41.093 B17F.2    sit    255  60     90     0     80   0    
13:39:42.096 B17F.2    sit    255  60     90     0     80   0    
13:39:43.096 B17F.2    sit    255  60     90     0     80   0    
13:39:44.098 B17F.2    sit    255  60     90     0     80   0    
13:39:45.098 B17F.2    sit    255  60     90     0     80   0    
13:39:46.099 B17F.2    sit    255  50     80     0     80   14   
13:39:47.000 B17F.2    sit    255  60     80     0     80   10   
13:39:48.002 B17F.2    sit    255  60     80     0     80   0    
13:39:49.005 B17F.2    sit    255  60     80     0     80   0    
13:39:50.004 B17F.2    sit    255  60     80     0     80   0    
13:39:51.073 B17F.2    sit    255  60     80     0     80   0    
13:39:52.004 B17F.2    sit    255  60     80     0     80   0    
13:39:53.008 B17F.2    sit    255  130    110    0     80   76   
13:39:54.013 B17F.2    sit    255  90     110    0     80   40   
13:39:55.011 B17F.2    sit    255  70     100    0     80   22   
13:39:56.015 B17F.2    sit    255  60     90     73    80   14   
13:39:57.010 B17F.2    sit    255  70     100    83    80   14   
13:39:58.010 B17F.2    sit    255  60     90     80    80   14   
13:39:58.910 B17F.2    sit    255  60     100    76    80   10   
13:39:59.913 B17F.2    sit    255  70     90     73    80   14   
13:40:00.949 B17F.2    sit    255  50     80     77    80   22   
13:40:01.908 B17F.2    sit    255  70     90     83    80   22   
13:40:02.910 B17F.2    sit    255  50     80     77    80   22   
13:40:03.922 B17F.2    sit    255  50     90     72    80   10   
13:40:04.919 B17F.2    sit    255  50     90     80    80   0    
13:40:05.913 B17F.2    sit    255  60     100    90    80   14   
13:40:06.916 B17F.2    sit    255  50     90     72    80   14   
13:40:07.917 B17F.2    sit    255  60     90     78    80   10   
13:40:08.925 B17F.2    sit    255  80     100    71    80   22   
13:40:09.918 B17F.2    sit    255  70     90     74    80   14   
13:40:10.809 B17F.2    sit    255  70     100    84    80   10   
13:40:11.809 B17F.2    sit    255  70     100    75    80   0    
13:40:12.812 B17F.2    sit    255  70     100    78    80   0    
13:40:13.817 B17F.2    sit    255  60     100    74    80   10   
13:40:14.818 B17F.2    sit    255  60     90     75    80   10   
13:40:15.816 B17F.2    sit    255  50     90     77    80   10   
13:40:16.818 B17F.2    sit    255  80     100    93    80   31   
13:40:17.818 B17F.2    sit    255  70     100    80    80   10   
13:40:18.834 B17F.2    sit    255  70     100    91    80   0    
13:40:19.826 B17F.2    sit    255  50     90     79    80   22   
13:40:20.828 B17F.2    sit    255  60     90     73    80   10   
13:40:21.729 B17F.2    sit    255  60     90     67    80   0    
13:40:22.728 B17F.2    sit    255  70     100    79    80   14   
13:40:23.732 B17F.2    sit    255  70     90     69    80   10   
13:40:24.729 B17F.2    sit    255  60     90     73    80   10   
13:40:25.728 B17F.2    sit    255  60     100    77    80   10   
13:40:26.726 B17F.2    sit    255  70     90     79    80   14   
13:40:27.728 B17F.2    sit    255  70     100    75    80   10   
13:40:28.730 B17F.2    sit    255  70     90     75    80   10   
13:40:29.731 B17F.2    sit    255  60     90     86    80   10   
13:40:30.748 B17F.2    sit    255  60     90     73    80   0    
13:40:31.732 B17F.2    sit    255  60     90     75    80   0    
13:40:32.742 B17F.2    sit    255  60     100    78    80   10   
13:40:33.628 B17F.2    sit    255  70     100    76    80   10   
13:40:34.636 B17F.2    sit    255  70     100    78    80   0    
13:40:35.640 B17F.2    sit    255  60     90     79    80   14   
13:40:36.639 B17F.2    sit    255  70     100    80    80   14   
13:40:37.638 B17F.2    sit    255  70     90     74    80   10   
13:40:38.642 B17F.2    sit    255  80     100    0     80   14   
13:40:39.652 B17F.2    sit    255  80     110    74    80   10   
13:40:40.648 B17F.2    sit    255  80     90     78    80   20   
13:40:41.648 B17F.2    sit    255  70     100    77    80   14   
13:40:42.648 B17F.2    sit    255  70     110    94    80   10   
13:40:43.644 B17F.2    sit    255  70     100    90    80   10   
13:40:44.538 B17F.2    sit    255  70     100    80    80   0    
13:40:45.538 B17F.2    sit    255  80     110    79    80   14   
13:40:46.538 B17F.2    sit    255  70     110    84    80   10   
13:40:47.550 B17F.2    sit    255  70     110    86    80   0    
13:40:48.548 B17F.2    sit    255  70     100    79    80   10   
13:40:49.541 B17F.2    sit    255  60     90     76    80   14   
13:40:50.592 B17F.2    sit    255  60     90     79    80   0    
13:40:51.552 B17F.2    sit    255  80     100    79    80   22   
13:40:52.550 B17F.2    sit    255  70     100    86    80   10   
13:40:53.550 B17F.2    sit    255  60     100    75    80   10   
13:40:54.550 B17F.2    sit    255  80     70     70    80   36   
13:40:55.549 B17F.2    sit    255  90     70     75    80   10   
13:40:56.451 B17F.2    sit    255  80     50     90    80   22   
13:40:57.448 B17F.2    sit    255  40     40     83    80   41   
13:40:58.449 B17F.2    walk   255  -70    50     86    80   110  
13:40:59.449 B17F.2    walk   255  -120   30     0     80   53   
13:41:00.446 B17F.2    walk   255  -100   20     0     80   22   
13:41:01.449 B17F.2    stand  255  -100   20     0     80   0    
13:41:02.450 B17F.2    sit    255  -100   10     0     80   10   
13:41:03.450 B17F.2    sit    255  -90    10     0     80   10   
13:41:04.450 B17F.2    sit    255  -90    10     0     80   0    
13:41:05.454 B17F.2    sit    255  -90    10     0     80   0    
13:41:06.461 B17F.2    sit    255  -90    10     0     80   0    
13:41:07.351 B17F.2    sit    255  -90    10     0     80   0    
13:41:08.357 B17F.2    sit    255  -90    10     0     80   0    
13:41:09.356 B17F.2    sit    255  -90    10     0     80   0    
13:41:10.355 B17F.2    sit    255  -90    10     39    80   0    
13:41:11.361 B17F.2    sit    255  -100   20     68    80   14   
13:41:12.360 B17F.2    sit    255  -60    30     83    80   41   
13:41:13.361 B17F.2    sit    255  -10    80     83    80   70   
13:41:14.365 B17F.2    walk   255  70     60     78    80   82   
13:41:15.389 B17F.2    walk   255  100    90     83    80   42   
13:41:16.361 B17F.2    walk   255  110    80     82    80   14   
13:41:17.366 B17F.2    sit    255  70     60     73    80   44   
13:41:18.364 B17F.2    walk   255  -50    70     88    80   120  
13:41:19.264 B17F.2    walk   255  -100   40     62    80   58   
13:41:20.261 B17F.2    walk   255  -150   80     0     80   64   
13:41:21.262 B17F.2    walk   255  -160   80     0     80   10   
13:41:22.273 B17F.2    walk   255  -160   80     0     80   0    
13:41:23.276 B17F.2    stand  255  -160   80     0     80   0    
13:41:24.282 B17F.2    stand  255  -160   70     0     80   10   
13:41:25.280 B17F.2    stand  255  -160   70     0     80   0    
13:41:26.280 B17F.2    stand  255  -160   70     0     80   0    
13:41:27.287 B17F.2    stand  255  -160   70     0     80   0    
13:41:28.292 B17F.2    stand  255  -160   70     0     80   0    
13:41:29.184 B17F.2    stand  255  -160   70     0     80   0    
13:41:30.178 B17F.2    stand  255  -180   40     0     80   36   
13:41:31.238 B17F.0    stand  255  140    100    72    80   325  
13:41:31.238 B17F.2    stand  255  -180   30     0     80   327  
13:41:32.193 B17F.0    stand  255  140    90     51    80   325  
13:41:32.193 B17F.2    stand  255  -180   30     0     80   325  
13:41:33.191 B17F.0    stand  255  110    50     95    80   290  
13:41:33.191 B17F.2    stand  255  -180   30     0     80   290  
13:41:34.193 B17F.2    stand  255  -180   30     0     80   0    
13:41:34.193 B17F.0    stand  255  100    60     88    80   281  
13:41:35.193 B17F.2    stand  255  -180   20     0     80   282  
13:41:35.193 B17F.0    stand  255  50     60     86    80   233  
13:41:36.195 B17F.2    stand  255  -180   20     0     80   233  
13:41:36.195 B17F.0    walk   255  -50    70     91    80   139  
13:41:37.193 B17F.2    stand  255  -180   20     0     80   139  
13:41:37.193 B17F.0    walk   255  -80    50     71    80   104  
13:41:38.198 B17F.2    stand  255  -180   20     0     80   104  
13:41:38.198 B17F.0    walk   255  -130   70     77    80   70   
13:41:39.104 B17F.2    stand  255  -180   20     0     80   70   
13:41:39.104 B17F.0    walk   255  -150   110    94    80   94   
13:41:40.100 B17F.2    stand  255  -180   20     0     80   94   
13:41:40.100 B17F.0    walk   255  -160   200    107   80   181  
13:41:41.100 B17F.2    stand  255  -180   20     0     80   181  
13:41:41.100 B17F.0    walk   255  -190   290    122   80   270  
13:41:42.103 B17F.2    stand  255  -180   20     0     80   270  
13:41:42.103 B17F.0    walk   255  -180   370    0     80   350  
13:41:43.108 B17F.2    stand  255  -180   20     0     80   350  
13:41:43.108 B17F.0    walk   255  -180   370    0     80   350  
13:41:44.108 B17F.2    stand  255  -180   20     0     80   350  
13:41:44.108 B17F.0    walk   255  -180   360    0     80   340  
13:41:45.104 B17F.2    stand  255  -180   20     0     80   340  
13:41:45.104 B17F.0    stand  255  -180   360    0     80   340  
13:41:46.110 B17F.2    stand  255  -180   20     0     80   340  
13:41:46.110 B17F.0    stand  255  -180   360    0     80   340  
13:41:47.108 B17F.2    stand  255  -180   20     0     80   340  
13:41:47.108 B17F.0    stand  3    -180   360    0     80   340  
13:41:48.165 B17F.2    stand  255  -180   20     0     80   340  
13:41:49.024 B17F.2    stand  255  -180   20     0     80   0    
13:41:50.021 B17F.2    stand  255  -180   20     0     80   0    
13:41:51.075 B17F.2    stand  255  -180   20     0     80   0    
13:41:52.018 B17F.2    stand  255  -180   20     0     80   0    
13:41:53.030 B17F.2    stand  255  -180   20     0     80   0    
13:41:54.029 B17F.2    stand  255  -180   20     0     80   0    
13:41:54.973 B17F.2    stand  255  -180   20     0     80   0    
13:41:55.975 B17F.2    stand  255  -180   20     0     80   0    
13:41:56.972 B17F.2    stand  255  -180   20     0     80   0    
13:41:57.972 B17F.2    stand  255  -180   20     0     80   0    
13:41:58.973 B17F.2    stand  255  -180   20     0     80   0    
13:41:59.978 B17F.2    stand  255  -180   20     0     80   0    
13:42:00.977 B17F.2    stand  255  -180   20     0     80   0    
13:42:01.980 B17F.2    stand  255  -180   20     0     80   0    
13:42:02.976 B17F.2    stand  255  -180   20     0     80   0    
13:42:03.981 B17F.2    stand  255  -180   20     0     80   0    
13:42:04.978 B17F.2    stand  255  -180   20     0     80   0    
13:42:05.980 B17F.2    stand  255  -180   20     0     80   0    
13:42:06.872 B17F.2    stand  255  -180   20     0     80   0    
13:42:07.872 B17F.2    stand  255  -180   20     0     80   0    
13:42:08.874 B17F.2    stand  255  -180   20     0     80   0    
13:42:09.877 B17F.2    stand  255  -180   20     0     80   0    
13:42:10.909 B17F.2    stand  255  -180   20     0     80   0    
13:42:11.906 B17F.2    stand  255  -180   20     0     80   0    
13:42:12.910 B17F.2    stand  255  -180   20     0     80   0    
13:42:13.912 B17F.2    stand  255  -150   70     0     80   58   
13:42:14.908 B17F.2    stand  255  -150   70     0     80   0    
13:42:15.801 B17F.2    stand  255  -150   70     0     80   0    
13:42:16.801 B17F.2    stand  255  -150   70     0     80   0    
13:42:17.804 B17F.2    stand  255  -150   70     0     80   0    
13:42:18.805 B17F.2    stand  255  -150   70     0     80   0    
13:42:19.806 B17F.2    stand  255  -150   70     0     80   0    
13:42:20.806 B17F.2    stand  255  -150   70     0     80   0    
13:42:21.806 B17F.2    stand  255  -150   70     0     80   0    
13:42:22.815 B17F.2    stand  255  -150   70     0     80   0    
13:42:23.808 B17F.2    stand  255  -150   70     0     80   0    
13:42:24.817 B17F.2    stand  255  -150   70     0     80   0    
13:42:25.810 B17F.2    stand  255  -150   70     0     80   0    
13:42:26.814 B17F.2    stand  255  -150   70     0     80   0    
13:42:27.705 B17F.2    stand  255  -150   70     0     80   0    
13:42:28.707 B17F.2    stand  255  -150   70     0     80   0    
13:42:29.710 B17F.2    stand  255  -150   70     0     80   0    
13:42:30.710 B17F.2    stand  255  -150   70     0     80   0    
13:42:31.708 B17F.2    stand  255  -150   70     0     80   0    
13:42:32.708 B17F.2    stand  255  -150   70     0     80   0    
13:42:33.710 B17F.2    stand  255  -150   70     0     80   0    
13:42:34.720 B17F.2    stand  255  -150   70     0     80   0    
13:42:35.711 B17F.2    stand  255  -150   70     0     80   0    
13:42:36.713 B17F.2    stand  255  -150   70     0     80   0    
13:42:37.721 B17F.2    stand  255  -130   40     0     80   36   
13:42:38.714 B17F.2    stand  255  -130   30     0     80   10   
13:42:39.614 B17F.2    stand  255  -80    70     0     80   64   
13:42:40.622 B17F.2    stand  255  -70    70     0     80   10   
13:42:41.611 B17F.2    stand  255  -70    70     0     80   0    
13:42:42.613 B17F.2    stand  255  -90    60     40    80   22   
13:42:43.619 B17F.2    stand  255  -120   20     0     80   50   
13:42:44.613 B17F.2    stand  255  -70    30     0     80   50   
13:42:45.616 B17F.2    stand  255  -80    10     0     80   22   
13:42:46.617 B17F.2    stand  255  -90    10     0     80   10   
13:42:47.622 B17F.2    stand  255  -90    10     0     80   0    
13:42:48.627 B17F.2    stand  255  -90    10     0     80   0    
13:42:49.670 B17F.2    stand  255  -90    10     0     80   0    
13:42:50.621 B17F.2    stand  255  -90    10     0     80   0    
13:42:51.516 B17F.2    stand  255  -70    20     0     80   22   
13:42:52.517 B17F.2    stand  255  -60    20     0     80   10   
13:42:53.521 B17F.2    stand  255  -70    20     0     80   10   
13:42:54.514 B17F.2    stand  255  -70    20     0     80   0    
13:42:55.517 B17F.2    stand  255  -70    20     0     80   0    
13:42:56.517 B17F.2    stand  255  -60    20     0     80   10   
13:42:57.517 B17F.2    stand  255  -60    10     0     80   10   
13:42:58.538 B17F.2    stand  255  -60    10     0     80   0    
13:42:59.541 B17F.2    stand  255  -60    10     0     80   0    
13:43:00.539 B17F.2    stand  255  -60    10     0     80   0    
13:43:01.432 B17F.2    stand  255  -40    40     0     80   36   
13:43:02.477 B17F.2    stand  255  -40    40     0     80   0    
13:43:03.433 B17F.2    stand  255  -40    40     0     80   0    
13:43:04.441 B17F.2    stand  255  -40    40     0     80   0    
13:43:05.444 B17F.2    stand  255  -40    40     0     80   0    
13:43:06.444 B17F.2    stand  255  -40    40     0     80   0    
13:43:07.440 B17F.2    stand  255  -40    40     0     80   0    
13:43:08.440 B17F.2    stand  255  -40    40     0     80   0    
13:43:09.439 B17F.2    stand  255  -40    40     0     80   0    
13:43:10.441 B17F.2    stand  255  -160   50     0     80   120  
13:43:11.444 B17F.2    stand  255  -150   10     0     80   41   
13:43:12.442 B17F.2    walk   255  -140   20     28    80   14   
13:43:13.341 B17F.2    stand  255  -120   20     0     80   20   
13:43:14.349 B17F.2    stand  255  -120   20     0     80   0    
13:43:15.357 B17F.2    stand  255  -110   20     0     80   10   
13:43:16.357 B17F.2    stand  255  -110   20     0     80   0    
13:43:17.362 B17F.2    stand  255  -110   20     0     80   0    
13:43:18.363 B17F.2    stand  255  -110   20     0     80   0    
13:43:19.366 B17F.2    stand  255  -120   10     0     80   14   
13:43:20.371 B17F.2    stand  255  -120   10     0     80   0    
13:43:21.362 B17F.2    stand  255  -120   10     0     80   0    
13:43:22.372 B17F.2    stand  255  -100   20     0     80   22   
13:43:23.260 B17F.2    stand  255  -50    40     0     80   53   
13:43:24.264 B17F.2    stand  255  -70    30     55    80   22   
13:43:25.262 B17F.2    stand  255  -30    10     0     80   44   
13:43:26.265 B17F.2    stand  255  -20    0      0     80   14   
13:43:27.265 B17F.2    stand  255  -20    0      0     80   0    
13:43:28.264 B17F.2    stand  255  -70    0      0     80   50   
13:43:29.264 B17F.2    stand  255  -70    0      0     80   0    
13:43:30.264 B17F.2    stand  255  -70    0      0     80   0    
13:43:31.193 B17F.2    stand  255  -70    0      0     80   0    
13:43:32.196 B17F.2    stand  255  -70    0      0     80   0    
13:43:33.193 B17F.2    stand  255  -70    0      0     80   0    
13:43:34.195 B17F.2    stand  255  -80    10     0     80   14   
13:43:35.200 B17F.2    stand  255  -90    30     80    80   22   
13:43:36.197 B17F.2    stand  255  -40    90     87    80   78   
13:43:37.198 B17F.2    walk   255  0      110    77    80   44   
13:43:38.201 B17F.2    walk   255  20     140    77    80   36   
13:43:39.206 B17F.2    walk   255  30     150    90    80   14   
13:43:40.203 B17F.2    walk   255  50     150    125   80   20   
13:43:41.209 B17F.2    walk   255  40     140    126   80   14   
13:43:42.207 B17F.2    walk   255  50     150    109   80   14   
13:43:43.105 B17F.2    walk   255  60     160    99    80   14   
13:43:44.100 B17F.2    walk   255  60     150    96    80   10   
13:43:45.105 B17F.2    sit    255  60     140    98    80   10   
13:43:46.101 B17F.2    sit    255  40     120    0     80   28   
13:43:47.076 B17F.2    sit    255  60     140    103   80   28   
13:43:48.069 B17F.2    sit    255  60     140    93    80   0    
13:43:49.116 B17F.2    sit    255  40     150    104   80   22   
13:43:50.068 B17F.2    sit    255  40     140    126   80   10   
13:43:51.071 B17F.2    sit    255  40     160    97    80   20   
13:43:52.074 B17F.2    sit    255  60     150    103   80   22   
13:43:53.072 B17F.2    sit    255  40     140    97    80   22   
13:43:54.088 B17F.2    sit    255  0      120    74    80   44   
13:43:55.088 B17F.2    walk   255  -20    100    79    80   28   
13:43:56.076 B17F.2    walk   255  -50    50     90    80   58   
13:43:57.080 B17F.2    walk   255  -40    10     110   80   41   
13:43:58.078 B17F.2    walk   255  -30    50     82    80   41   
13:43:58.976 B17F.2    walk   255  -10    70     67    80   28   
13:43:59.973 B17F.2    walk   255  0      130    77    80   60   
13:44:00.989 B17F.2    walk   255  20     150    84    80   28   
13:44:01.999 B17F.2    walk   255  40     150    68    80   20   
13:44:02.990 B17F.2    walk   255  20     130    87    80   28   
13:44:03.981 B17F.2    walk   255  10     120    85    80   14   
13:44:04.982 B17F.2    walk   255  -20    100    86    80   36   
13:44:05.988 B17F.2    walk   255  -60    50     69    80   64   
13:44:06.984 B17F.2    walk   255  -100   60     63    80   41   
13:44:07.994 B17F.2    walk   255  -70    20     0     80   50   
13:44:08.988 B17F.2    stand  255  -90    0      0     80   28   
13:44:09.882 B17F.2    stand  255  -100   0      0     80   10   
13:44:10.881 B17F.2    sit    255  -100   0      0     80   0    
13:44:11.889 B17F.2    sit    255  -100   0      0     80   0    
13:44:12.884 B17F.2    sit    255  -100   0      0     80   0    
13:44:13.885 B17F.2    sit    255  -100   0      0     80   0    
13:44:14.885 B17F.2    sit    255  -100   0      0     80   0    
13:44:15.893 B17F.2    sit    255  -100   0      0     80   0    
13:44:16.895 B17F.2    sit    255  -100   0      0     80   0    
13:44:17.888 B17F.2    sit    255  -100   0      0     80   0    
13:44:18.892 B17F.2    sit    255  -100   0      0     80   0    
13:44:19.892 B17F.2    sit    255  -100   0      0     80   0    
13:44:20.892 B17F.2    sit    255  -100   0      0     80   0    
13:44:21.789 B17F.2    sit    255  -100   0      0     80   0    
13:44:22.788 B17F.2    sit    255  -100   0      0     80   0    
13:44:23.785 B17F.2    sit    255  -100   0      0     80   0    
13:44:24.795 B17F.2    sit    255  -100   0      0     80   0    
13:44:25.788 B17F.2    sit    255  -100   0      0     80   0    
13:44:26.796 B17F.2    sit    255  -100   0      0     80   0    
13:44:27.806 B17F.2    sit    255  -100   0      0     80   0    
13:44:28.795 B17F.2    sit    255  -100   0      0     80   0    
13:44:29.800 B17F.2    sit    255  -100   0      0     80   0    
13:44:30.797 B17F.2    sit    255  -100   0      0     80   0    
13:44:31.794 B17F.2    sit    255  -100   0      0     80   0    
13:44:32.796 B17F.2    sit    255  -100   0      0     80   0    
13:44:33.691 B17F.2    sit    255  -100   0      0     80   0    
13:44:34.690 B17F.2    sit    255  -100   10     0     80   10   
13:44:35.690 B17F.2    sit    255  -100   10     0     80   0    
13:44:36.697 B17F.2    sit    255  -100   10     0     80   0    
13:44:37.694 B17F.2    sit    255  -80    10     0     80   20   
13:44:38.692 B17F.2    sit    255  -60    10     0     80   20   
13:44:39.702 B17F.2    sit    255  -60    10     0     80   0    
13:44:40.697 B17F.2    sit    255  -60    10     0     80   0    
13:44:41.702 B17F.2    sit    255  -70    10     0     80   10   
13:44:42.701 B17F.2    sit    255  -110   0      0     80   41   
13:44:43.707 B17F.2    sit    255  -110   0      0     80   0    
13:44:44.701 B17F.2    sit    255  -110   0      0     80   0    
13:44:45.596 B17F.2    sit    255  -110   0      0     80   0    
13:44:46.597 B17F.2    sit    255  -110   0      0     80   0    
13:44:47.593 B17F.2    sit    255  -110   10     0     80   10   
13:44:48.656 B17F.2    sit    255  -80    30     70    80   36   
13:44:49.600 B17F.2    sit    255  -30    100    84    80   86   
13:44:50.619 B17F.2    walk   255  10     140    69    80   56   
13:44:51.615 B17F.2    walk   255  30     160    85    80   28   
13:44:52.615 B17F.2    walk   255  30     150    86    80   10   
13:44:53.623 B17F.2    walk   255  0      130    80    80   36   
13:44:54.624 B17F.2    walk   255  -50    70     77    80   78   
13:44:55.522 B17F.2    walk   255  -90    50     62    80   44   
13:44:56.524 B17F.2    walk   255  -140   80     75    80   58   
13:44:57.518 B17F.2    walk   255  -180   160    120   80   89   
13:44:58.521 B17F.2    walk   255  -200   270    117   80   111  

```

**汇总**: xray tick 938 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

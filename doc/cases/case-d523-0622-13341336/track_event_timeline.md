# case-d523-0622-13341336 — 卧室(09E7+D523 双雷达同房) 每 tick belief 时间线

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
13:34:00 D523.0   D52303400920  stand   0    NoReport stand              trk  0.50 Empty      1   0     0.00  0.03  0.08  0.00  0.85  0.04
13:34:01 D523.0   D52303400920  stand   0    NoReport stand              trk  0.51 Empty      1   0     0.00  0.04  0.10  0.00  0.82  0.01
13:34:02 D523.0   D52303400920  stand   0    NoReport stand              trk  0.52 Empty      1   1     0.00  0.04  0.11  0.01  0.76  0.01
13:34:03 D523.0   D52303400920  stand   0    NoReport stand              trk  0.53 Empty      1   2     0.00  0.05  0.12  0.01  0.71  0.01
13:34:04 D523.0   D52303400920  stand   0    NoReport stand              trk  0.54 Empty      1   3     0.00  0.05  0.13  0.02  0.66  0.01
13:34:05 D523.0   D52303400920  stand   0    NoReport stand              trk  0.55 Empty      1   4     0.00  0.06  0.14  0.03  0.62  0.01
13:34:06 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   5     0.01  0.06  0.14  0.04  0.58  0.01
13:34:07 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   6     0.01  0.07  0.14  0.04  0.54  0.01
13:34:08 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   7     0.01  0.07  0.15  0.05  0.51  0.02
13:34:09 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   8     0.01  0.07  0.15  0.06  0.49  0.02
13:34:10 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   9     0.01  0.08  0.15  0.06  0.46  0.02
13:34:11 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   10    0.01  0.08  0.15  0.07  0.44  0.02
13:34:12 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   11    0.02  0.08  0.15  0.07  0.42  0.02
13:34:13 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   11    0.02  0.08  0.15  0.07  0.42  0.02
13:34:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      1   11    0.02  0.08  0.15  0.07  0.42  0.02
13:34:15 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   14    0.02  0.08  0.15  0.08  0.40  0.02
13:34:15 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   14    0.02  0.09  0.15  0.08  0.38  0.02
13:34:15 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   14    0.02  0.09  0.15  0.09  0.36  0.02
13:34:16 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   15    0.03  0.09  0.15  0.09  0.35  0.02
13:34:17 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   16    0.03  0.09  0.15  0.10  0.34  0.02
13:34:18 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   17    0.03  0.09  0.15  0.10  0.33  0.02
13:34:19 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   18    0.03  0.10  0.15  0.10  0.31  0.02
13:34:20 D523.0   D52303400920  stand   113  NoReport stand              trk  1.00 Empty      1   19    0.03  0.10  0.16  0.10  0.30  0.02
13:34:21 D523.0   D52303400920  stand   105  NoReport stand              trk  1.00 Empty      1   0     0.04  0.10  0.16  0.11  0.30  0.02
13:34:22 D523.0   D52303400920  stand   129  NoReport stand              trk  1.00 Empty      1   0     0.04  0.10  0.15  0.11  0.29  0.02
13:34:23 D523.0   D52303400920  stand   108  NoReport stand              trk  1.00 Empty      1   0     0.04  0.10  0.15  0.11  0.28  0.02
13:34:24 D523.0   D52303400920  stand   117  NoReport stand              trk  1.00 Empty      1   0     0.04  0.10  0.15  0.11  0.27  0.02
13:34:25 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.05  0.10  0.15  0.12  0.27  0.02
13:34:26 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.05  0.10  0.15  0.12  0.26  0.02
13:34:27 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.05  0.10  0.15  0.12  0.25  0.02
13:34:28 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.05  0.10  0.15  0.12  0.25  0.02
13:34:29 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.06  0.10  0.15  0.12  0.25  0.02
13:34:30 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.06  0.10  0.15  0.12  0.24  0.02
13:34:31 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.06  0.10  0.15  0.12  0.24  0.02
13:34:32 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.06  0.10  0.15  0.12  0.23  0.02
13:34:33 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.07  0.10  0.15  0.12  0.23  0.02
13:34:34 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.07  0.10  0.15  0.12  0.23  0.02
13:34:35 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.07  0.10  0.15  0.13  0.22  0.02
13:34:36 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.07  0.10  0.15  0.13  0.22  0.02
13:34:37 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.07  0.10  0.15  0.13  0.22  0.02
13:34:38 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.08  0.11  0.15  0.13  0.22  0.02
13:34:39 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.08  0.11  0.15  0.13  0.21  0.02
13:34:40 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.08  0.11  0.15  0.13  0.21  0.02
13:34:41 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.09  0.11  0.15  0.13  0.21  0.02
13:34:42 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.09  0.11  0.15  0.13  0.21  0.02
13:34:43 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.09  0.10  0.15  0.13  0.21  0.02
13:34:44 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.09  0.10  0.15  0.13  0.20  0.02
13:34:45 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   0     0.09  0.10  0.15  0.13  0.20  0.02
13:34:45 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.09  0.10  0.15  0.13  0.20  0.02
13:34:46 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.10  0.10  0.15  0.13  0.20  0.02
13:34:47 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.10  0.10  0.15  0.13  0.20  0.02
13:34:48 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.10  0.10  0.15  0.13  0.20  0.02
13:34:49 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.10  0.10  0.15  0.13  0.20  0.02
13:34:50 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.10  0.10  0.15  0.13  0.20  0.02
13:34:51 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.10  0.10  0.15  0.13  0.20  0.02
13:34:52 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   27    0.11  0.10  0.15  0.13  0.20  0.02
13:34:53 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   28    0.11  0.10  0.15  0.13  0.20  0.02
13:34:54 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   29    0.11  0.10  0.15  0.13  0.19  0.02
13:34:55 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   30    0.11  0.10  0.15  0.13  0.19  0.02
13:34:56 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   31    0.11  0.10  0.15  0.13  0.19  0.02
13:34:57 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   32    0.12  0.10  0.15  0.13  0.19  0.02
13:34:58 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   33    0.12  0.10  0.15  0.13  0.19  0.02
13:34:59 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   34    0.12  0.10  0.15  0.13  0.19  0.02
13:35:00 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   35    0.12  0.10  0.15  0.13  0.19  0.02
13:35:01 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   36    0.12  0.10  0.15  0.13  0.19  0.02
13:35:02 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   37    0.12  0.10  0.14  0.13  0.19  0.02
13:35:03 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   38    0.13  0.10  0.14  0.13  0.19  0.02
13:35:04 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   39    0.13  0.10  0.14  0.13  0.19  0.02
13:35:05 D523.0   D52303400920  stand   0    NoReport stand              trk  1.00 Empty      1   40    0.13  0.10  0.14  0.13  0.19  0.02
13:35:06 D523.E   -             -       0    NoReport np=0  ★0           room -    Empty      1   40    0.13  0.10  0.14  0.13  0.19  0.02
13:35:06 D523.88  -             88      -    NoReport no-target(88)      room -    Empty      1   41    0.13  0.10  0.14  0.13  0.19  0.02
13:35:07 D523.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:35:08 D523.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:35:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
13:35:10 D523.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:17 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:42 D523.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:49 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:35:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.19  0.02
13:36:13 D523.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:17 1903.0   -             pad     -    NoReport pad InBed HR=75 RR=None mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:18 1903.0   -             pad     -    NoReport pad InBed HR=75 RR=None mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:19 1903.E   -             -       0    NoReport InBed(pad)         room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:19 1903.E   -             -       0    NoReport InBed(pad)         room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:20 1903.0   -             pad     -    NoReport pad InBed HR=75 RR=15 mv=0 turn=1 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:20 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:22 1903.0   -             pad     -    NoReport pad InBed HR=78 RR=14 mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:24 1903.0   -             pad     -    NoReport pad InBed HR=68 RR=None mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:26 1903.0   -             pad     -    NoReport pad InBed HR=68 RR=None mv=1 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:28 1903.0   -             pad     -    NoReport pad InBed HR=68 RR=None mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:30 1903.0   -             pad     -    NoReport pad InBed HR=None RR=14 mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:32 1903.0   -             pad     -    NoReport pad InBed HR=None RR=None mv=1 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:34 1903.0   -             pad     -    NoReport pad InBed HR=60 RR=None mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:36 1903.0   -             pad     -    NoReport pad InBed HR=None RR=14 mv=0 turn=1 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:37 1903.0   -             pad     -    NoReport pad InBed HR=56 RR=14 mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:38 1903.0   -             pad     -    NoReport pad InBed HR=56 RR=14 mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:41 1903.0   -             pad     -    NoReport pad InBed HR=52 RR=14 mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:43 1903.0   -             pad     -    NoReport pad InBed HR=51 RR=14 mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:45 1903.0   -             pad     -    NoReport pad InBed HR=52 RR=12 mv=0 turn=0 room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
13:36:45 D523.88  -             88      -    NoReport no-target(88)      room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:46 -.-      -             -       -    NoReport (no frame, held)   room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:47 1903.0   -             pad     -    NoReport pad InBed HR=51 RR=12 mv=0 turn=0 room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:48 -.-      -             -       -    NoReport (no frame, held)   room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:49 1903.0   -             pad     -    NoReport pad InBed HR=54 RR=14 mv=0 turn=0 room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:50 -.-      -             -       -    NoReport (no frame, held)   room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:51 1903.0   -             pad     -    NoReport pad InBed HR=54 RR=None mv=0 turn=0 room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:52 D5F7.E   -             -       0    NoReport np=0  ★0           room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:52 D5F7.88  -             88      -    NoReport no-target(88)      room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:53 1903.0   -             pad     -    NoReport pad InBed HR=54 RR=14 mv=1 turn=0 room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:54 -.-      -             -       -    NoReport (no frame, held)   room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:55 1903.0   -             pad     -    NoReport pad InBed HR=None RR=None mv=0 turn=0 room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:56 -.-      -             -       -    NoReport (no frame, held)   room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:57 1903.0   -             pad     -    NoReport pad InBed HR=None RR=None mv=1 turn=0 room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:58 1903.0   -             pad     -    NoReport pad InBed HR=None RR=None mv=0 turn=0 room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
13:36:59 1903.0   -             pad     -    NoReport pad InBed HR=None RR=None mv=1 turn=0 room -    Left       0   0     0.10  0.07  0.10  0.09  0.21  0.23
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
13:34:00.920 D523.0    stand  1    -260   490    0     80        
13:34:01.817 D523.0    stand  1    -260   490    0     80   0    
13:34:02.810 D523.0    stand  1    -260   490    0     80   0    
13:34:03.809 D523.0    stand  1    -250   500    0     80   14   
13:34:04.816 D523.0    stand  1    -250   510    0     80   10   
13:34:05.869 D523.0    stand  1    -270   490    0     80   28   
13:34:06.857 D523.0    stand  1    -270   490    0     80   0    
13:34:07.871 D523.0    stand  1    -270   490    0     80   0    
13:34:08.765 D523.0    stand  1    -270   490    0     80   0    
13:34:09.769 D523.0    stand  1    -270   490    0     80   0    
13:34:10.761 D523.0    stand  1    -270   490    0     80   0    
13:34:11.785 D523.0    stand  1    -270   490    0     80   0    
13:34:12.805 D523.0    stand  1    -270   490    0     80   0    
13:34:15.154 D523.0    stand  1    -260   490    0     80   10   
13:34:15.156 D523.0    stand  1    -260   490    0     80   0    
13:34:15.759 D523.0    stand  1    -260   490    0     80   0    
13:34:16.760 D523.0    stand  1    -260   490    0     80   0    
13:34:17.761 D523.0    stand  1    -260   490    0     80   0    
13:34:18.767 D523.0    stand  1    -260   490    0     80   0    
13:34:19.765 D523.0    stand  1    -260   490    0     80   0    
13:34:20.660 D523.0    stand  1    -270   480    113   80   14   
13:34:21.658 D523.0    stand  1    -290   440    105   80   44   
13:34:22.663 D523.0    stand  1    -290   450    129   80   10   
13:34:23.661 D523.0    stand  1    -290   450    108   80   0    
13:34:24.662 D523.0    stand  1    -260   410    117   80   50   
13:34:25.662 D523.0    stand  1    -210   370    0     80   64   
13:34:26.663 D523.0    stand  1    -210   370    0     80   0    
13:34:27.667 D523.0    stand  1    -210   370    0     80   0    
13:34:28.666 D523.0    stand  1    -210   380    0     80   10   
13:34:29.678 D523.0    stand  1    -210   380    0     80   0    
13:34:30.668 D523.0    stand  1    -210   380    0     80   0    
13:34:31.568 D523.0    stand  1    -210   380    0     80   0    
13:34:32.568 D523.0    stand  1    -210   380    0     80   0    
13:34:33.570 D523.0    stand  1    -210   380    0     80   0    
13:34:34.572 D523.0    stand  1    -210   380    0     80   0    
13:34:35.571 D523.0    stand  1    -210   380    0     80   0    
13:34:36.575 D523.0    stand  1    -210   380    0     80   0    
13:34:37.520 D523.0    stand  1    -210   380    0     80   0    
13:34:38.525 D523.0    stand  1    -210   380    0     80   0    
13:34:39.526 D523.0    stand  1    -210   380    0     80   0    
13:34:40.573 D523.0    stand  1    -210   380    0     80   0    
13:34:41.527 D523.0    stand  1    -210   380    0     80   0    
13:34:42.534 D523.0    stand  1    -210   380    0     80   0    
13:34:43.525 D523.0    stand  1    -210   380    0     80   0    
13:34:44.528 D523.0    stand  1    -210   380    0     80   0    
13:34:45.536 D523.0    stand  1    -210   380    0     80   0    
13:34:46.529 D523.0    stand  1    -210   380    0     80   0    
13:34:47.535 D523.0    stand  1    -210   380    0     80   0    
13:34:48.533 D523.0    stand  1    -210   380    0     80   0    
13:34:49.429 D523.0    stand  1    -210   380    0     80   0    
13:34:50.424 D523.0    stand  1    -210   380    0     80   0    
13:34:51.425 D523.0    stand  1    -210   380    0     80   0    
13:34:52.429 D523.0    stand  1    -210   380    0     80   0    
13:34:53.444 D523.0    stand  1    -210   380    0     80   0    
13:34:54.482 D523.0    stand  1    -210   380    0     80   0    
13:34:55.376 D523.0    stand  1    -210   380    0     80   0    
13:34:56.376 D523.0    stand  1    -210   380    0     80   0    
13:34:57.379 D523.0    stand  1    -210   380    0     80   0    
13:34:58.396 D523.0    stand  1    -210   380    0     80   0    
13:34:59.386 D523.0    stand  1    -210   380    0     80   0    
13:35:00.382 D523.0    stand  1    -210   380    0     80   0    
13:35:01.382 D523.0    stand  1    -210   380    0     80   0    
13:35:02.388 D523.0    stand  1    -210   380    0     80   0    
13:35:03.386 D523.0    stand  1    -210   380    0     80   0    
13:35:04.384 D523.0    stand  1    -210   380    0     80   0    
13:35:05.387 D523.0    stand  1    -210   380    0     80   0    
13:35:06.445 D523.88   88     -    -      -      -     -    -    
13:35:07.278 D523.88   88     -    -      -      -     -    -    
13:35:08.280 D523.88   88     -    -      -      -     -    -    
13:35:10.298 D523.88   88     -    -      -      -     -    -    
13:35:42.096 D523.88   88     -    -      -      -     -    -    
13:36:13.809 D523.88   88     -    -      -      -     -    -    
13:36:45.633 D523.88   88     -    -      -      -     -    -    

13:34:13.665 D5F7.88   88     -    -      -      -     -    -    
13:34:45.534 D5F7.88   88     -    -      -      -     -    -    
13:35:17.048 D5F7.88   88     -    -      -      -     -    -    
13:35:49.106 D5F7.88   88     -    -      -      -     -    -    
13:36:20.540 D5F7.88   88     -    -      -      -     -    -    
13:36:52.722 D5F7.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 76 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

# case-b197-0630-08020806 — 每 tick belief 时间线 (room fd00:0:3:111:1:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
08:02:06 B197.E   -             -       0    NoReport np=1               room -    Empty      1   0     0.00  0.03  0.15  0.00  0.79  0.04
08:02:06 B197.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      1   0     0.00  0.03  0.15  0.00  0.79  0.04
08:02:06 B197.0   B19700206195  stand   121  NoReport stand              trk  0.92 Empty      1   0     0.00  0.03  0.15  0.00  0.79  0.04
08:02:06 B197.0   B19700206195  stand   0    NoReport stand              trk  0.92 Empty      1   0     0.00  0.03  0.25  0.00  0.66  0.01
08:02:07 B197.0   B19700206195  stand   0    NoReport stand              trk  0.92 Empty      1   1     0.00  0.04  0.34  0.01  0.51  0.02
08:02:08 B197.0   B19700206195  stand   119  NoReport stand              trk  0.92 OpenFloor  1   2     0.00  0.04  0.42  0.01  0.38  0.02
08:02:09 B197.0   B19700206195  stand   0    NoReport stand              trk  0.92 OpenFloor  1   3     0.00  0.05  0.48  0.01  0.28  0.02
08:02:10 B197.0   B19700206195  stand   0    NoReport stand              trk  0.92 OpenFloor  1   4     0.00  0.05  0.53  0.01  0.21  0.03
08:02:11 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   5     0.01  0.05  0.58  0.02  0.12  0.03
08:02:12 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   6     0.01  0.05  0.59  0.02  0.09  0.03
08:02:13 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   7     0.01  0.05  0.60  0.02  0.08  0.03
08:02:14 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   8     0.01  0.05  0.60  0.02  0.07  0.03
08:02:15 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   9     0.01  0.05  0.60  0.02  0.06  0.03
08:02:16 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   10    0.01  0.05  0.61  0.02  0.06  0.03
08:02:17 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   11    0.01  0.05  0.61  0.02  0.06  0.03
08:02:18 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   12    0.01  0.05  0.61  0.02  0.06  0.03
08:02:19 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   13    0.01  0.05  0.61  0.02  0.05  0.03
08:02:20 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   14    0.01  0.05  0.61  0.02  0.05  0.03
08:02:21 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   15    0.01  0.05  0.61  0.02  0.05  0.03
08:02:22 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   16    0.01  0.05  0.61  0.02  0.05  0.03
08:02:23 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   17    0.01  0.05  0.61  0.02  0.05  0.03
08:02:24 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   18    0.01  0.05  0.61  0.02  0.05  0.03
08:02:25 B197.0   B19700206195  stand   0    NoReport stand              trk  1.00 OpenFloor  1   19    0.01  0.05  0.61  0.02  0.05  0.03
08:02:26 B197.0   B19700206195  stand   49   NoReport stand              trk  1.00 OpenFloor  1   20    0.01  0.05  0.61  0.02  0.05  0.03
08:02:26 B197.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  1   20    0.01  0.05  0.61  0.02  0.05  0.03
08:02:27 B197.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   20    0.01  0.05  0.61  0.02  0.05  0.03
08:02:27 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:28 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:29 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:57 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:29 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:03:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:00 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:19 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:04:20 B197.0   B19700420496  stand   99   NoReport stand              trk  0.50 Empty      1   0     0.00  0.03  0.15  0.00  0.79  0.04
08:04:20 B197.0   B19700420496  stand   0    NoReport stand              trk  0.50 Empty      1   0     0.00  0.03  0.25  0.00  0.66  0.01
08:04:21 B197.0   B19700420496  stand   78   NoReport stand              trk  0.89 Empty      1   1     0.00  0.04  0.34  0.01  0.51  0.02
08:04:22 B197.0   B19700420496  stand   0    NoReport stand              trk  0.89 OpenFloor  1   2     0.00  0.04  0.42  0.01  0.38  0.02
08:04:23 B197.0   B19700420496  stand   0    NoReport stand              trk  0.90 OpenFloor  1   0     0.00  0.05  0.48  0.01  0.28  0.02
08:04:24 B197.0   B19700420496  stand   0    NoReport stand              trk  0.90 OpenFloor  1   0     0.01  0.05  0.53  0.01  0.21  0.03
08:04:25 B197.0   B19700420496  stand   103  NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.56  0.02  0.15  0.03
08:04:26 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.58  0.02  0.12  0.03
08:04:27 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.59  0.02  0.10  0.03
08:04:28 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.60  0.02  0.08  0.03
08:04:29 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.60  0.02  0.07  0.03
08:04:30 B197.0   B19700420496  stand   88   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
08:04:31 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.06  0.03
08:04:32 B197.0   B19700420496  stand   71   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.06  0.03
08:04:33 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.06  0.03
08:04:34 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:35 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:36 B197.0   B19700420496  stand   61   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:37 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:38 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:39 B197.0   B19700420496  stand   76   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:40 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:41 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:42 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:43 B197.0   B19700420496  stand   78   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:44 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:45 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:46 B197.0   B19700420496  stand   80   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
08:04:47 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:48 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:49 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:50 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:51 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:52 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:53 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:54 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:55 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:56 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:57 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:58 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:04:59 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:00 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:01 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:02 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:03 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:04 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:05 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:06 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:07 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:08 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:09 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:10 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:11 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:12 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:13 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:14 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:15 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:16 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:17 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:18 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:19 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   27    0.01  0.05  0.61  0.02  0.05  0.03
08:05:20 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   28    0.01  0.05  0.61  0.02  0.05  0.03
08:05:21 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   29    0.01  0.05  0.61  0.02  0.05  0.03
08:05:22 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   30    0.01  0.05  0.61  0.02  0.05  0.03
08:05:23 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   31    0.01  0.05  0.61  0.02  0.05  0.03
08:05:24 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   32    0.01  0.05  0.61  0.02  0.05  0.03
08:05:25 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   33    0.01  0.05  0.61  0.02  0.05  0.03
08:05:26 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   34    0.01  0.05  0.61  0.02  0.05  0.03
08:05:27 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   35    0.01  0.05  0.61  0.02  0.05  0.03
08:05:28 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   36    0.01  0.05  0.61  0.02  0.05  0.03
08:05:29 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   37    0.01  0.05  0.61  0.02  0.05  0.03
08:05:30 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   38    0.01  0.05  0.61  0.02  0.05  0.03
08:05:31 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   39    0.01  0.05  0.61  0.02  0.05  0.03
08:05:32 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   40    0.01  0.05  0.61  0.02  0.05  0.03
08:05:33 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   41    0.01  0.05  0.61  0.02  0.05  0.03
08:05:34 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   42    0.01  0.05  0.61  0.02  0.05  0.03
08:05:35 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   43    0.01  0.05  0.61  0.02  0.05  0.03
08:05:36 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:37 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:38 B197.0   B19700420496  stand   60   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:39 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:40 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:41 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:42 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:43 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:44 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:45 B197.0   B19700420496  stand   81   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
08:05:46 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:47 B197.0   B19700420496  stand   98   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
08:05:48 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:49 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:50 B197.0   B19700420496  stand   80   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
08:05:51 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:52 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:53 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:54 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:55 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:56 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:57 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:58 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:05:59 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:00 B197.0   B19700420496  stand   76   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:01 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:02 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:03 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:03 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:04 B197.0   B19700420496  stand   79   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:05 B197.0   B19700420496  stand   97   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
08:06:06 B197.0   B19700420496  stand   109  NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
08:06:07 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:08 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:09 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:10 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:11 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:12 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:13 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:14 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:15 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:16 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:17 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:18 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:19 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:20 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:21 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:22 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:23 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:24 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:25 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:26 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:27 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:28 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:29 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:30 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:31 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:32 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:33 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:34 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:35 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:36 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:37 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:38 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:39 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:40 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:41 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
08:06:42 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   28    0.01  0.05  0.61  0.02  0.05  0.03
08:06:43 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   29    0.01  0.05  0.61  0.02  0.05  0.03
08:06:44 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   30    0.01  0.05  0.61  0.02  0.05  0.03
08:06:45 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   31    0.01  0.05  0.61  0.02  0.05  0.03
08:06:46 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   32    0.01  0.05  0.61  0.02  0.05  0.03
08:06:47 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   33    0.01  0.05  0.61  0.02  0.05  0.03
08:06:48 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   34    0.01  0.05  0.61  0.02  0.05  0.03
08:06:49 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   35    0.01  0.05  0.61  0.02  0.05  0.03
08:06:50 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   36    0.01  0.05  0.61  0.02  0.05  0.03
08:06:51 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   37    0.01  0.05  0.61  0.02  0.05  0.03
08:06:52 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   38    0.01  0.05  0.61  0.02  0.05  0.03
08:06:53 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   39    0.01  0.05  0.61  0.02  0.05  0.03
08:06:54 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   40    0.01  0.05  0.61  0.02  0.05  0.03
08:06:55 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   41    0.01  0.05  0.61  0.02  0.05  0.03
08:06:56 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   42    0.01  0.05  0.61  0.02  0.05  0.03
08:06:57 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   43    0.01  0.05  0.61  0.02  0.05  0.03
08:06:58 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   44    0.01  0.05  0.61  0.02  0.05  0.03
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
08:02:06.195 B197.0    stand  0    -490   170    121   80        
08:02:06.827 B197.0    stand  0    -490   170    0     80   0    
08:02:07.834 B197.0    stand  0    -490   160    0     80   10   
08:02:08.830 B197.0    stand  0    -500   160    119   80   10   
08:02:09.837 B197.0    stand  0    -500   180    0     80   20   
08:02:10.836 B197.0    stand  0    -500   170    0     80   10   
08:02:11.887 B197.0    stand  0    -500   170    0     80   0    
08:02:12.832 B197.0    stand  0    -500   170    0     80   0    
08:02:13.837 B197.0    stand  0    -500   170    0     80   0    
08:02:14.836 B197.0    stand  0    -500   170    0     80   0    
08:02:15.840 B197.0    stand  0    -500   170    0     80   0    
08:02:16.836 B197.0    stand  0    -500   170    0     80   0    
08:02:17.839 B197.0    stand  0    -500   170    0     80   0    
08:02:18.732 B197.0    stand  0    -500   170    0     80   0    
08:02:19.732 B197.0    stand  0    -500   170    0     80   0    
08:02:20.736 B197.0    stand  0    -500   170    0     80   0    
08:02:21.740 B197.0    stand  0    -500   170    0     80   0    
08:02:22.744 B197.0    stand  0    -500   170    0     80   0    
08:02:23.737 B197.0    stand  0    -500   170    0     80   0    
08:02:24.738 B197.0    stand  0    -500   170    0     80   0    
08:02:25.739 B197.0    stand  0    -500   170    0     80   0    
08:02:26.746 B197.0    stand  0    -490   140    49    80   31   
08:02:27.800 B197.88   88     -    -      -      -     -    -    
08:02:28.751 B197.88   88     -    -      -      -     -    -    
08:02:29.650 B197.88   88     -    -      -      -     -    -    
08:02:57.495 B197.88   88     -    -      -      -     -    -    
08:03:29.294 B197.88   88     -    -      -      -     -    -    
08:04:00.967 B197.88   88     -    -      -      -     -    -    
08:04:20.496 B197.0    stand  255  -150   500    99    80   495  
08:04:20.751 B197.0    stand  255  -120   510    0     80   31   
08:04:21.751 B197.0    stand  255  -100   510    78    80   20   
08:04:22.749 B197.0    stand  255  -130   500    0     80   31   
08:04:23.754 B197.0    stand  255  -170   500    0     80   40   
08:04:24.756 B197.0    stand  255  -180   520    0     80   22   
08:04:25.754 B197.0    stand  255  -170   510    103   80   14   
08:04:26.754 B197.0    stand  255  -120   510    0     80   50   
08:04:27.773 B197.0    stand  255  -120   510    0     80   0    
08:04:28.772 B197.0    stand  255  -120   510    0     80   0    
08:04:29.777 B197.0    stand  255  -120   510    0     80   0    
08:04:30.680 B197.0    stand  255  -100   510    88    80   20   
08:04:31.677 B197.0    stand  255  -120   510    0     80   20   
08:04:32.679 B197.0    stand  255  -120   510    71    80   0    
08:04:33.678 B197.0    stand  255  -100   500    0     80   22   
08:04:34.679 B197.0    stand  255  -90    510    0     80   14   
08:04:35.679 B197.0    stand  255  -90    510    0     80   0    
08:04:36.680 B197.0    stand  255  -90    510    61    80   0    
08:04:37.686 B197.0    stand  255  -90    510    0     80   0    
08:04:38.682 B197.0    stand  255  -90    510    0     80   0    
08:04:39.688 B197.0    stand  255  -120   510    76    80   30   
08:04:40.683 B197.0    stand  255  -120   510    0     80   0    
08:04:41.582 B197.0    stand  255  -120   510    0     80   0    
08:04:42.583 B197.0    stand  255  -120   510    0     80   0    
08:04:43.588 B197.0    stand  255  -120   510    78    80   0    
08:04:44.593 B197.0    stand  255  -120   510    0     80   0    
08:04:45.587 B197.0    stand  255  -120   510    0     80   0    
08:04:46.593 B197.0    stand  255  -120   510    80    80   0    
08:04:47.586 B197.0    stand  255  -120   510    0     80   0    
08:04:48.592 B197.0    stand  255  -120   510    0     80   0    
08:04:49.593 B197.0    stand  255  -120   510    0     80   0    
08:04:50.600 B197.0    stand  255  -120   510    0     80   0    
08:04:51.592 B197.0    stand  255  -110   450    0     80   60   
08:04:52.593 B197.0    stand  255  -110   450    0     80   0    
08:04:53.483 B197.0    stand  255  -110   460    0     80   10   
08:04:54.485 B197.0    stand  255  -110   460    0     80   0    
08:04:55.488 B197.0    stand  255  -110   460    0     80   0    
08:04:56.486 B197.0    stand  255  -110   460    0     80   0    
08:04:57.489 B197.0    stand  255  -110   460    0     80   0    
08:04:58.503 B197.0    stand  255  -110   450    0     80   10   
08:04:59.439 B197.0    stand  255  -110   450    0     80   0    
08:05:00.442 B197.0    stand  255  -110   450    0     80   0    
08:05:01.441 B197.0    stand  255  -110   450    0     80   0    
08:05:02.439 B197.0    stand  255  -110   450    0     80   0    
08:05:03.440 B197.0    stand  255  -110   450    0     80   0    
08:05:04.443 B197.0    stand  255  -110   450    0     80   0    
08:05:05.441 B197.0    stand  255  -110   450    0     80   0    
08:05:06.447 B197.0    stand  255  -110   450    0     80   0    
08:05:07.443 B197.0    stand  255  -110   450    0     80   0    
08:05:08.451 B197.0    stand  255  -110   450    0     80   0    
08:05:09.451 B197.0    stand  255  -110   450    0     80   0    
08:05:10.503 B197.0    stand  255  -110   430    0     80   20   
08:05:11.340 B197.0    stand  255  -110   430    0     80   0    
08:05:12.341 B197.0    stand  255  -110   430    0     80   0    
08:05:13.357 B197.0    stand  255  -110   430    0     80   0    
08:05:14.347 B197.0    stand  255  -110   430    0     80   0    
08:05:15.355 B197.0    stand  255  -110   430    0     80   0    
08:05:16.397 B197.0    stand  255  -110   430    0     80   0    
08:05:17.304 B197.0    stand  255  -110   430    0     80   0    
08:05:18.302 B197.0    stand  255  -110   430    0     80   0    
08:05:19.293 B197.0    stand  255  -110   430    0     80   0    
08:05:20.300 B197.0    stand  255  -110   430    0     80   0    
08:05:21.297 B197.0    stand  255  -110   430    0     80   0    
08:05:22.296 B197.0    stand  255  -110   430    0     80   0    
08:05:23.302 B197.0    stand  255  -110   430    0     80   0    
08:05:24.304 B197.0    stand  255  -110   430    0     80   0    
08:05:25.303 B197.0    stand  255  -110   430    0     80   0    
08:05:26.308 B197.0    stand  255  -110   430    0     80   0    
08:05:27.306 B197.0    stand  255  -110   430    0     80   0    
08:05:28.306 B197.0    stand  255  -110   430    0     80   0    
08:05:29.196 B197.0    stand  255  -110   430    0     80   0    
08:05:30.197 B197.0    stand  255  -110   430    0     80   0    
08:05:31.200 B197.0    stand  255  -110   430    0     80   0    
08:05:32.210 B197.0    stand  255  -110   430    0     80   0    
08:05:33.210 B197.0    stand  255  -110   430    0     80   0    
08:05:34.211 B197.0    stand  255  -110   430    0     80   0    
08:05:35.220 B197.0    stand  255  -110   430    0     80   0    
08:05:36.217 B197.0    stand  255  -170   470    0     80   72   
08:05:37.213 B197.0    stand  255  -170   470    0     80   0    
08:05:38.220 B197.0    stand  255  -170   490    60    80   20   
08:05:39.215 B197.0    stand  255  -160   510    0     80   22   
08:05:40.110 B197.0    stand  255  -160   510    0     80   0    
08:05:41.110 B197.0    stand  255  -160   510    0     80   0    
08:05:42.111 B197.0    stand  255  -160   500    0     80   10   
08:05:43.115 B197.0    stand  255  -160   500    0     80   0    
08:05:44.112 B197.0    stand  255  -160   500    0     80   0    
08:05:45.121 B197.0    stand  255  -160   500    81    80   0    
08:05:46.123 B197.0    stand  255  -160   510    0     80   10   
08:05:47.118 B197.0    stand  255  -130   510    98    80   30   
08:05:48.118 B197.0    stand  255  -110   500    0     80   22   
08:05:49.119 B197.0    stand  255  -110   500    0     80   0    
08:05:50.129 B197.0    stand  255  -150   500    80    80   40   
08:05:51.120 B197.0    stand  255  -160   500    0     80   10   
08:05:52.017 B197.0    stand  255  -130   500    0     80   30   
08:05:53.017 B197.0    stand  255  -120   500    0     80   10   
08:05:54.019 B197.0    stand  255  -100   520    0     80   28   
08:05:55.021 B197.0    stand  255  -100   470    0     80   50   
08:05:56.016 B197.0    stand  255  -150   460    0     80   50   
08:05:57.018 B197.0    stand  255  -150   470    0     80   10   
08:05:58.030 B197.0    stand  255  -150   470    0     80   0    
08:05:59.019 B197.0    stand  255  -140   470    0     80   10   
08:06:00.020 B197.0    stand  255  -100   510    76    80   56   
08:06:01.025 B197.0    stand  255  -100   510    0     80   0    
08:06:02.023 B197.0    stand  255  -120   520    0     80   22   
08:06:03.024 B197.0    stand  255  -110   490    0     80   31   
08:06:03.924 B197.0    stand  255  -110   480    0     80   10   
08:06:04.925 B197.0    stand  255  -150   490    79    80   41   
08:06:05.919 B197.0    stand  255  -100   490    97    80   50   
08:06:06.922 B197.0    stand  255  -130   510    109   80   36   
08:06:07.928 B197.0    stand  255  -110   510    0     80   20   
08:06:08.923 B197.0    stand  255  -110   510    0     80   0    
08:06:09.972 B197.0    stand  255  -110   510    0     80   0    
08:06:10.932 B197.0    stand  255  -110   510    0     80   0    
08:06:11.926 B197.0    stand  255  -110   510    0     80   0    
08:06:12.946 B197.0    stand  255  -110   510    0     80   0    
08:06:13.928 B197.0    stand  255  -110   430    0     80   80   
08:06:14.927 B197.0    stand  255  -100   480    0     80   50   
08:06:15.824 B197.0    stand  255  -100   480    0     80   0    
08:06:16.824 B197.0    stand  255  -100   480    0     80   0    
08:06:17.828 B197.0    stand  255  -100   480    0     80   0    
08:06:18.827 B197.0    stand  255  -100   480    0     80   0    
08:06:19.841 B197.0    stand  255  -100   480    0     80   0    
08:06:20.848 B197.0    stand  255  -100   480    0     80   0    
08:06:21.844 B197.0    stand  255  -110   460    0     80   22   
08:06:22.846 B197.0    stand  255  -110   460    0     80   0    
08:06:23.848 B197.0    stand  255  -110   460    0     80   0    
08:06:24.847 B197.0    stand  255  -110   460    0     80   0    
08:06:25.744 B197.0    stand  255  -110   460    0     80   0    
08:06:26.749 B197.0    stand  255  -110   460    0     80   0    
08:06:27.743 B197.0    stand  255  -110   480    0     80   20   
08:06:28.743 B197.0    stand  255  -110   480    0     80   0    
08:06:29.746 B197.0    stand  255  -110   480    0     80   0    
08:06:30.754 B197.0    stand  255  -110   480    0     80   0    
08:06:31.748 B197.0    stand  255  -110   480    0     80   0    
08:06:32.749 B197.0    stand  255  -110   480    0     80   0    
08:06:33.748 B197.0    stand  255  -110   480    0     80   0    
08:06:34.754 B197.0    stand  255  -110   480    0     80   0    
08:06:35.759 B197.0    stand  255  -110   480    0     80   0    
08:06:36.757 B197.0    stand  255  -110   480    0     80   0    
08:06:37.644 B197.0    stand  255  -110   480    0     80   0    
08:06:38.649 B197.0    stand  255  -110   480    0     80   0    
08:06:39.647 B197.0    stand  255  -110   480    0     80   0    
08:06:40.663 B197.0    stand  255  -110   480    0     80   0    
08:06:41.650 B197.0    stand  255  -110   480    0     80   0    
08:06:42.652 B197.0    stand  255  -110   480    0     80   0    
08:06:43.657 B197.0    stand  255  -110   480    0     80   0    
08:06:44.670 B197.0    stand  255  -110   480    0     80   0    
08:06:45.653 B197.0    stand  255  -110   480    0     80   0    
08:06:46.654 B197.0    stand  255  -110   480    0     80   0    
08:06:47.659 B197.0    stand  255  -110   480    0     80   0    
08:06:48.665 B197.0    stand  255  -110   480    0     80   0    
08:06:49.561 B197.0    stand  255  -110   480    0     80   0    
08:06:50.551 B197.0    stand  255  -110   480    0     80   0    
08:06:51.560 B197.0    stand  255  -110   480    0     80   0    
08:06:52.553 B197.0    stand  255  -110   480    0     80   0    
08:06:53.557 B197.0    stand  255  -110   480    0     80   0    
08:06:54.556 B197.0    stand  255  -110   480    0     80   0    
08:06:55.566 B197.0    stand  255  -110   480    0     80   0    
08:06:56.560 B197.0    stand  255  -110   480    0     80   0    
08:06:57.559 B197.0    stand  255  -110   480    0     80   0    
08:06:58.560 B197.0    stand  255  -110   480    0     80   0    

```

**汇总**: xray tick 194 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

## 完整原始记录（按时间排序，data_value 全文不删字段）
```
time     ms             device.tid   event          x      y      z     原始记录
08:02:06 1782828126156  B197.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782828126156, "event_status": "start", "number_people": 1, "respiratory_rate": -1}
08:02:06 1782828126192  B197         EnterRoom      -      -      -     {"heart_rate": -1, "event_since": 1782828126192, "event_status": "start", "respiratory_rate": -1}
08:02:06 1782828126195  B197.0       track          -490   170    121   {"pose": 4, "event": 1, "area_id": 0, "track_id": 0, "position_x": -490, "position_y": 170, "position_z": 121, "remaining_time": 0, "track_confidence": 80}
08:02:06 1782828126827  B197.0       track          -490   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -490, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:07 1782828127834  B197.0       track          -490   160    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -490, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:08 1782828128830  B197.0       track          -500   160    119   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 160, "position_z": 119, "remaining_time": 0, "track_confidence": 80}
08:02:09 1782828129837  B197.0       track          -500   180    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 180, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:10 1782828130836  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:11 1782828131848  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828131848, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 4, "respiratory_rate": -1, "multi_person_duration": 0}
08:02:11 1782828131848  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:02:11 1782828131887  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:12 1782828132832  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:13 1782828133837  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:14 1782828134836  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:15 1782828135840  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:16 1782828136836  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:17 1782828137839  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:18 1782828138732  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:19 1782828139732  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:20 1782828140736  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:21 1782828141740  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:22 1782828142744  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:23 1782828143737  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:24 1782828144738  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:25 1782828145739  B197.0       track          -500   170    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:26 1782828146746  B197.0       track          -490   140    49    {"pose": 4, "event": 2, "area_id": 0, "track_id": 0, "position_x": -490, "position_y": 140, "position_z": 49, "remaining_time": 0, "track_confidence": 80}
08:02:26 1782828146785  B197         ExitRoom       -      -      -     {"heart_rate": -1, "event_since": 1782828146785, "event_status": "start", "respiratory_rate": -1}
08:02:27 1782828147758  B197.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782828147758, "event_status": "start", "number_people": 0, "respiratory_rate": -1}
08:02:27 1782828147800  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:28 1782828148751  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:29 1782828149650  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:02:57 1782828177495  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:03:29 1782828209175  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828209175, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 16, "respiratory_rate": -1, "multi_person_duration": 0}
08:03:29 1782828209175  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:03:29 1782828209294  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:00 1782828240967  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:19 1782828259864  B197.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782828259864, "event_status": "start", "number_people": 1, "respiratory_rate": -1}
08:04:20 1782828260493  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828260493, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:04:20 1782828260493  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:04:20 1782828260496  B197.0       track          -150   500    99    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 500, "position_z": 99, "remaining_time": 0, "track_confidence": 80}
08:04:20 1782828260751  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:21 1782828261751  B197.0       track          -100   510    78    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 510, "position_z": 78, "remaining_time": 0, "track_confidence": 80}
08:04:22 1782828262749  B197.0       track          -130   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -130, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:23 1782828263754  B197.0       track          -170   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:24 1782828264756  B197.0       track          -180   520    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -180, "position_y": 520, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:25 1782828265754  B197.0       track          -170   510    103   {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 510, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
08:04:26 1782828266754  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:27 1782828267773  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:28 1782828268772  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:29 1782828269777  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:30 1782828270680  B197.0       track          -100   510    88    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 510, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
08:04:31 1782828271677  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:32 1782828272679  B197.0       track          -120   510    71    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 71, "remaining_time": 0, "track_confidence": 80}
08:04:33 1782828273678  B197.0       track          -100   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:34 1782828274679  B197.0       track          -90    510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -90, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:35 1782828275679  B197.0       track          -90    510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -90, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:36 1782828276680  B197.0       track          -90    510    61    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -90, "position_y": 510, "position_z": 61, "remaining_time": 0, "track_confidence": 80}
08:04:37 1782828277686  B197.0       track          -90    510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -90, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:38 1782828278682  B197.0       track          -90    510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -90, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:39 1782828279688  B197.0       track          -120   510    76    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 76, "remaining_time": 0, "track_confidence": 80}
08:04:40 1782828280683  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:41 1782828281582  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:42 1782828282583  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:43 1782828283588  B197.0       track          -120   510    78    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 78, "remaining_time": 0, "track_confidence": 80}
08:04:44 1782828284593  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:45 1782828285587  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:46 1782828286593  B197.0       track          -120   510    80    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 80, "remaining_time": 0, "track_confidence": 80}
08:04:47 1782828287586  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:48 1782828288592  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:49 1782828289593  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:50 1782828290600  B197.0       track          -120   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:51 1782828291592  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:52 1782828292593  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:53 1782828293483  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:54 1782828294485  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:55 1782828295488  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:56 1782828296486  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:57 1782828297489  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:58 1782828298503  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:04:59 1782828299439  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:00 1782828300442  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:01 1782828301441  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:02 1782828302439  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:03 1782828303440  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:04 1782828304443  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:05 1782828305441  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:06 1782828306447  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:07 1782828307443  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:08 1782828308451  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:09 1782828309451  B197.0       track          -110   450    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 450, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:10 1782828310466  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828310466, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 50, "respiratory_rate": -1, "multi_person_duration": 0}
08:05:10 1782828310466  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:05:10 1782828310503  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:11 1782828311340  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:12 1782828312341  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:13 1782828313357  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:14 1782828314347  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:15 1782828315355  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:16 1782828316397  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:17 1782828317304  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:18 1782828318302  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:19 1782828319293  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:20 1782828320300  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:21 1782828321297  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:22 1782828322296  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:23 1782828323302  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:24 1782828324304  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:25 1782828325303  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:26 1782828326308  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:27 1782828327306  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:28 1782828328306  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:29 1782828329196  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:30 1782828330197  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:31 1782828331200  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:32 1782828332210  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:33 1782828333210  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:34 1782828334211  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:35 1782828335220  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:36 1782828336217  B197.0       track          -170   470    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 470, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:37 1782828337213  B197.0       track          -170   470    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 470, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:38 1782828338220  B197.0       track          -170   490    60    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 490, "position_z": 60, "remaining_time": 0, "track_confidence": 80}
08:05:39 1782828339215  B197.0       track          -160   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:40 1782828340110  B197.0       track          -160   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:41 1782828341110  B197.0       track          -160   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:42 1782828342111  B197.0       track          -160   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:43 1782828343115  B197.0       track          -160   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:44 1782828344112  B197.0       track          -160   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:45 1782828345121  B197.0       track          -160   500    81    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 500, "position_z": 81, "remaining_time": 0, "track_confidence": 80}
08:05:46 1782828346123  B197.0       track          -160   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:47 1782828347118  B197.0       track          -130   510    98    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -130, "position_y": 510, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
08:05:48 1782828348118  B197.0       track          -110   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:49 1782828349119  B197.0       track          -110   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:50 1782828350129  B197.0       track          -150   500    80    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 500, "position_z": 80, "remaining_time": 0, "track_confidence": 80}
08:05:51 1782828351120  B197.0       track          -160   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:52 1782828352017  B197.0       track          -130   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -130, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:53 1782828353017  B197.0       track          -120   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:54 1782828354019  B197.0       track          -100   520    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 520, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:55 1782828355021  B197.0       track          -100   470    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 470, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:56 1782828356016  B197.0       track          -150   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:57 1782828357018  B197.0       track          -150   470    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 470, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:58 1782828358030  B197.0       track          -150   470    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 470, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:05:59 1782828359019  B197.0       track          -140   470    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -140, "position_y": 470, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:00 1782828360020  B197.0       track          -100   510    76    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 510, "position_z": 76, "remaining_time": 0, "track_confidence": 80}
08:06:01 1782828361025  B197.0       track          -100   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:02 1782828362023  B197.0       track          -120   520    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -120, "position_y": 520, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:03 1782828363024  B197.0       track          -110   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:03 1782828363924  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:04 1782828364925  B197.0       track          -150   490    79    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 490, "position_z": 79, "remaining_time": 0, "track_confidence": 80}
08:06:05 1782828365919  B197.0       track          -100   490    97    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 490, "position_z": 97, "remaining_time": 0, "track_confidence": 80}
08:06:06 1782828366922  B197.0       track          -130   510    109   {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -130, "position_y": 510, "position_z": 109, "remaining_time": 0, "track_confidence": 80}
08:06:07 1782828367928  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:08 1782828368923  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:09 1782828369936  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828369936, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "respiratory_rate": -1, "multi_person_duration": 0}
08:06:09 1782828369936  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:06:09 1782828369972  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:10 1782828370932  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:11 1782828371926  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:12 1782828372946  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:13 1782828373928  B197.0       track          -110   430    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 430, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:14 1782828374927  B197.0       track          -100   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:15 1782828375824  B197.0       track          -100   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:16 1782828376824  B197.0       track          -100   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:17 1782828377828  B197.0       track          -100   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:18 1782828378827  B197.0       track          -100   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:19 1782828379841  B197.0       track          -100   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:20 1782828380848  B197.0       track          -100   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:21 1782828381844  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:22 1782828382846  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:23 1782828383848  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:24 1782828384847  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:25 1782828385744  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:26 1782828386749  B197.0       track          -110   460    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 460, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:27 1782828387743  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:28 1782828388743  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:29 1782828389746  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:30 1782828390754  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:31 1782828391748  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:32 1782828392749  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:33 1782828393748  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:34 1782828394754  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:35 1782828395759  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:36 1782828396757  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:37 1782828397644  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:38 1782828398649  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:39 1782828399647  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:40 1782828400663  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:41 1782828401650  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:42 1782828402652  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:43 1782828403657  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:44 1782828404670  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:45 1782828405653  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:46 1782828406654  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:47 1782828407659  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:48 1782828408665  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:49 1782828409561  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:50 1782828410551  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:51 1782828411560  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:52 1782828412553  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:53 1782828413557  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:54 1782828414556  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:55 1782828415566  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:56 1782828416560  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:57 1782828417559  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:06:58 1782828418560  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
```

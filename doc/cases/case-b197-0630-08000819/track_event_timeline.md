# case-b197-0630-08000819 — 每 tick belief 时间线 (room fd00:0:3:111:1:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
08:00:18 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:28 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:28 B197.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:28 B197.0   B19700028703  stand   83   NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
08:00:29 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
08:00:30 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   1     0.00  0.02  0.70  0.00  0.18  0.02
08:00:31 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   2     0.00  0.02  0.79  0.00  0.07  0.02
08:00:32 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   3     0.00  0.02  0.83  0.00  0.03  0.02
08:00:33 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   4     0.00  0.02  0.84  0.00  0.02  0.02
08:00:34 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   5     0.00  0.02  0.85  0.00  0.01  0.02
08:00:35 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   6     0.00  0.02  0.85  0.00  0.01  0.02
08:00:36 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   7     0.00  0.02  0.85  0.00  0.01  0.02
08:00:37 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   8     0.00  0.02  0.85  0.00  0.01  0.02
08:00:38 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   9     0.00  0.02  0.85  0.00  0.01  0.02
08:00:39 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   10    0.00  0.02  0.85  0.00  0.01  0.02
08:00:39 B197.0   B19700028703  stand   0    NoReport stand              trk  1.00 OpenFloor  1   11    0.00  0.02  0.85  0.00  0.01  0.02
08:00:39 B197.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  1   11    0.00  0.02  0.85  0.00  0.01  0.02
08:00:40 B197.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   11    0.00  0.02  0.85  0.00  0.01  0.02
08:00:40 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:41 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:42 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:50 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:00:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:22 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:53 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:01:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:06 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
08:02:06 B197.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
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
08:06:59 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   45    0.01  0.05  0.61  0.02  0.05  0.03
08:07:00 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   46    0.01  0.05  0.61  0.02  0.05  0.03
08:07:01 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   47    0.01  0.05  0.61  0.02  0.05  0.03
08:07:02 B197.0   B19700420496  stand   68   NoReport stand              trk  1.00 OpenFloor  1   48    0.01  0.05  0.61  0.02  0.05  0.03
08:07:03 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   49    0.01  0.05  0.61  0.02  0.05  0.03
08:07:04 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   50    0.01  0.05  0.61  0.02  0.05  0.03
08:07:05 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   51    0.01  0.05  0.61  0.02  0.05  0.03
08:07:06 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   52    0.01  0.05  0.61  0.02  0.05  0.03
08:07:07 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   53    0.01  0.05  0.61  0.02  0.05  0.03
08:07:08 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   54    0.01  0.05  0.61  0.02  0.05  0.03
08:07:09 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   55    0.01  0.05  0.61  0.02  0.05  0.03
08:07:10 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   56    0.01  0.05  0.61  0.02  0.05  0.03
08:07:11 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   57    0.01  0.05  0.61  0.02  0.05  0.03
08:07:12 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   58    0.01  0.05  0.61  0.02  0.05  0.03
08:07:13 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   59    0.01  0.05  0.61  0.02  0.05  0.03
08:07:14 B197.0   B19700420496  stand   72   NoReport stand              trk  1.00 OpenFloor  1   60    0.01  0.05  0.61  0.02  0.05  0.03
08:07:15 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   61    0.01  0.05  0.61  0.02  0.05  0.03
08:07:16 B197.0   B19700420496  stand   85   NoReport stand              trk  1.00 OpenFloor  1   62    0.00  0.05  0.61  0.02  0.05  0.03
08:07:17 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   63    0.01  0.05  0.61  0.02  0.05  0.03
08:07:18 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   64    0.01  0.05  0.61  0.02  0.05  0.03
08:07:19 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   65    0.01  0.05  0.61  0.02  0.05  0.03
08:07:20 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   66    0.01  0.05  0.61  0.02  0.05  0.03
08:07:21 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   67    0.01  0.05  0.61  0.02  0.05  0.03
08:07:22 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   68    0.01  0.05  0.61  0.02  0.05  0.03
08:07:23 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   69    0.01  0.05  0.61  0.02  0.05  0.03
08:07:24 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   70    0.01  0.05  0.61  0.02  0.05  0.03
08:07:25 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   71    0.01  0.05  0.61  0.02  0.05  0.03
08:07:26 B197.0   B19700420496  stand   77   NoReport stand              trk  1.00 OpenFloor  1   72    0.01  0.05  0.61  0.02  0.05  0.03
08:07:27 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   73    0.01  0.05  0.61  0.02  0.05  0.03
08:07:28 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   74    0.01  0.05  0.61  0.02  0.05  0.03
08:07:29 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   75    0.01  0.05  0.61  0.02  0.05  0.03
08:07:30 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   76    0.01  0.05  0.61  0.02  0.05  0.03
08:07:31 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   77    0.01  0.05  0.61  0.02  0.05  0.03
08:07:32 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   78    0.01  0.05  0.61  0.02  0.05  0.03
08:07:33 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   79    0.01  0.05  0.61  0.02  0.05  0.03
08:07:34 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   80    0.01  0.05  0.61  0.02  0.05  0.03
08:07:35 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   81    0.01  0.05  0.61  0.02  0.05  0.03
08:07:36 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   82    0.01  0.05  0.61  0.02  0.05  0.03
08:07:37 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   83    0.01  0.05  0.61  0.02  0.05  0.03
08:07:38 B197.0   B19700420496  stand   89   NoReport stand              trk  1.00 OpenFloor  1   84    0.00  0.05  0.61  0.02  0.05  0.03
08:07:39 B197.0   B19700420496  stand   0    NoReport stand              trk  1.00 OpenFloor  1   85    0.01  0.05  0.61  0.02  0.05  0.03
08:07:40 B197.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   85    0.01  0.05  0.61  0.02  0.05  0.03
08:07:40 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   86    0.01  0.05  0.61  0.02  0.05  0.03
08:07:41 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.01  0.08  0.44  0.03  0.08  0.05
08:07:42 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
08:07:43 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:44 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:45 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:46 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:47 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:48 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:49 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:50 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:51 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:52 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:53 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:54 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:55 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:56 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:57 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:58 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:07:59 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:00 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:01 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:02 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:03 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:04 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:05 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:06 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:07 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:08 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:09 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:10 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:11 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:12 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:13 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
08:08:14 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
08:08:46 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:08:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
08:09:18 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
08:09:50 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:09:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:09:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:09:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:09:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:09:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:09:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:09:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:09:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:09:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
08:10:21 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.11  0.19  0.02
08:10:53 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:10:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:10:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:10:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:10:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:10:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:10:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
08:11:25 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.20  0.02
08:11:57 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:11:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:11:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:29 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:12:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:00 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
08:13:32 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:13:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:04 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
08:14:36 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:14:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:15:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:15:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:15:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:15:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:15:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:15:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:15:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
08:15:07 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:39 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:15:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
08:16:11 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.20  0.02
08:16:42 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:16:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:14 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
08:17:46 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:17:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
08:18:18 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.03  0.08  0.00  0.85  0.04
08:18:49 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:18:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.04  0.10  0.00  0.82  0.01
08:19:21 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.05  0.12  0.01  0.71  0.01
08:19:53 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.05  0.13  0.02  0.66  0.01
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
08:00:18.938 B197.88   88     -    -      -      -     -    -    
08:00:28.703 B197.0    stand  0    -500   110    83    80        
08:00:29.650 B197.0    stand  0    -500   100    0     80   10   
08:00:30.646 B197.0    stand  0    -500   100    0     80   0    
08:00:31.650 B197.0    stand  0    -500   100    0     80   0    
08:00:32.656 B197.0    stand  0    -500   100    0     80   0    
08:00:33.650 B197.0    stand  0    -500   100    0     80   0    
08:00:34.654 B197.0    stand  0    -500   100    0     80   0    
08:00:35.653 B197.0    stand  0    -500   100    0     80   0    
08:00:36.653 B197.0    stand  0    -500   100    0     80   0    
08:00:37.626 B197.0    stand  0    -500   100    0     80   0    
08:00:38.551 B197.0    stand  0    -500   100    0     80   0    
08:00:39.550 B197.0    stand  0    -500   100    0     80   0    
08:00:39.756 B197.0    stand  0    -500   100    0     80   0    
08:00:40.605 B197.88   88     -    -      -      -     -    -    
08:00:41.569 B197.88   88     -    -      -      -     -    -    
08:00:42.574 B197.88   88     -    -      -      -     -    -    
08:00:50.456 B197.88   88     -    -      -      -     -    -    
08:01:22.429 B197.88   88     -    -      -      -     -    -    
08:01:53.935 B197.88   88     -    -      -      -     -    -    
08:02:06.195 B197.0    stand  0    -490   170    121   80   70   
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
08:06:59.560 B197.0    stand  255  -110   480    0     80   0    
08:07:00.565 B197.0    stand  255  -110   480    0     80   0    
08:07:01.457 B197.0    stand  255  -110   480    0     80   0    
08:07:02.456 B197.0    stand  255  -130   480    68    80   20   
08:07:03.456 B197.0    stand  255  -160   490    0     80   31   
08:07:04.456 B197.0    stand  255  -160   490    0     80   0    
08:07:05.457 B197.0    stand  255  -160   490    0     80   0    
08:07:06.458 B197.0    stand  255  -160   490    0     80   0    
08:07:07.468 B197.0    stand  255  -160   490    0     80   0    
08:07:08.476 B197.0    stand  255  -160   490    0     80   0    
08:07:09.532 B197.0    stand  255  -160   490    0     80   0    
08:07:10.481 B197.0    stand  255  -160   490    0     80   0    
08:07:11.488 B197.0    stand  255  -160   490    0     80   0    
08:07:12.370 B197.0    stand  255  -160   490    0     80   0    
08:07:13.365 B197.0    stand  255  -160   490    0     80   0    
08:07:14.368 B197.0    stand  255  -150   490    72    80   10   
08:07:15.368 B197.0    stand  255  -150   500    0     80   10   
08:07:16.373 B197.0    stand  255  -150   500    85    80   0    
08:07:17.370 B197.0    stand  255  -160   500    0     80   10   
08:07:18.379 B197.0    stand  255  -150   480    0     80   22   
08:07:19.394 B197.0    stand  255  -150   490    0     80   10   
08:07:20.376 B197.0    stand  255  -150   480    0     80   10   
08:07:21.376 B197.0    stand  255  -150   480    0     80   0    
08:07:22.374 B197.0    stand  255  -150   490    0     80   10   
08:07:23.375 B197.0    stand  255  -150   490    0     80   0    
08:07:24.271 B197.0    stand  255  -150   490    0     80   0    
08:07:25.276 B197.0    stand  255  -150   490    0     80   0    
08:07:26.271 B197.0    stand  255  -140   490    77    80   10   
08:07:27.281 B197.0    stand  255  -110   510    0     80   36   
08:07:28.273 B197.0    stand  255  -110   510    0     80   0    
08:07:29.274 B197.0    stand  255  -110   510    0     80   0    
08:07:30.280 B197.0    stand  255  -110   510    0     80   0    
08:07:31.276 B197.0    stand  255  -110   510    0     80   0    
08:07:32.275 B197.0    stand  255  -110   510    0     80   0    
08:07:33.284 B197.0    stand  255  -110   510    0     80   0    
08:07:34.285 B197.0    stand  255  -110   510    0     80   0    
08:07:35.281 B197.0    stand  255  -110   510    0     80   0    
08:07:36.173 B197.0    stand  255  -110   510    0     80   0    
08:07:37.182 B197.0    stand  255  -110   510    0     80   0    
08:07:38.181 B197.0    stand  255  -110   510    89    80   0    
08:07:39.177 B197.0    stand  255  -110   520    0     80   10   
08:07:40.234 B197.88   88     -    -      -      -     -    -    
08:07:41.190 B197.88   88     -    -      -      -     -    -    
08:07:42.200 B197.88   88     -    -      -      -     -    -    
08:07:43.200 B197.88   88     -    -      -      -     -    -    
08:08:14.994 B197.88   88     -    -      -      -     -    -    
08:08:46.608 B197.88   88     -    -      -      -     -    -    
08:09:18.588 B197.88   88     -    -      -      -     -    -    
08:09:50.101 B197.88   88     -    -      -      -     -    -    
08:10:21.976 B197.88   88     -    -      -      -     -    -    
08:10:53.638 B197.88   88     -    -      -      -     -    -    
08:11:25.674 B197.88   88     -    -      -      -     -    -    
08:11:57.126 B197.88   88     -    -      -      -     -    -    
08:12:29.049 B197.88   88     -    -      -      -     -    -    
08:13:00.557 B197.88   88     -    -      -      -     -    -    
08:13:32.644 B197.88   88     -    -      -      -     -    -    
08:14:04.052 B197.88   88     -    -      -      -     -    -    
08:14:36.236 B197.88   88     -    -      -      -     -    -    
08:15:07.673 B197.88   88     -    -      -      -     -    -    
08:15:39.279 B197.88   88     -    -      -      -     -    -    
08:16:11.265 B197.88   88     -    -      -      -     -    -    
08:16:42.787 B197.88   88     -    -      -      -     -    -    
08:17:14.956 B197.88   88     -    -      -      -     -    -    
08:17:46.256 B197.88   88     -    -      -      -     -    -    
08:18:18.342 B197.88   88     -    -      -      -     -    -    
08:18:49.801 B197.88   88     -    -      -      -     -    -    
08:19:21.939 B197.88   88     -    -      -      -     -    -    
08:19:53.293 B197.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 297 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

## 完整原始记录（按时间排序，data_value 全文不删字段）
```
time     ms             device.tid   event          x      y      z     原始记录
08:00:18 1782828018712  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828018712, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:00:18 1782828018712  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:00:18 1782828018938  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:28 1782828028650  B197.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782828028650, "event_status": "start", "number_people": 1, "respiratory_rate": -1}
08:00:28 1782828028700  B197         EnterRoom      -      -      -     {"heart_rate": -1, "event_since": 1782828028700, "event_status": "start", "respiratory_rate": -1}
08:00:28 1782828028703  B197.0       track          -500   110    83    {"pose": 4, "event": 1, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 110, "position_z": 83, "remaining_time": 0, "track_confidence": 80}
08:00:29 1782828029650  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:30 1782828030646  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:31 1782828031650  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:32 1782828032656  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:33 1782828033650  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:34 1782828034654  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:35 1782828035653  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:36 1782828036653  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:37 1782828037626  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:38 1782828038551  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:39 1782828039550  B197.0       track          -500   100    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:39 1782828039756  B197.0       track          -500   100    0     {"pose": 4, "event": 2, "area_id": 0, "track_id": 0, "position_x": -500, "position_y": 100, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:39 1782828039797  B197         ExitRoom       -      -      -     {"heart_rate": -1, "event_since": 1782828039797, "event_status": "start", "respiratory_rate": -1}
08:00:40 1782828040569  B197.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782828040569, "event_status": "start", "number_people": 0, "respiratory_rate": -1}
08:00:40 1782828040605  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:41 1782828041569  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:42 1782828042574  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:00:50 1782828050456  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:01:22 1782828082305  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:01:22 1782828082305  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828082305, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 12, "respiratory_rate": -1, "multi_person_duration": 0}
08:01:22 1782828082429  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:01:53 1782828113935  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
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
08:06:09 1782828369936  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:06:09 1782828369936  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828369936, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "respiratory_rate": -1, "multi_person_duration": 0}
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
08:06:59 1782828419560  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:00 1782828420565  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:01 1782828421457  B197.0       track          -110   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:02 1782828422456  B197.0       track          -130   480    68    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -130, "position_y": 480, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
08:07:03 1782828423456  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:04 1782828424456  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:05 1782828425457  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:06 1782828426458  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:07 1782828427468  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:08 1782828428476  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:09 1782828429494  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828429494, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "respiratory_rate": -1, "multi_person_duration": 0}
08:07:09 1782828429494  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:07:09 1782828429532  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:10 1782828430481  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:11 1782828431488  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:12 1782828432370  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:13 1782828433365  B197.0       track          -160   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:14 1782828434368  B197.0       track          -150   490    72    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 490, "position_z": 72, "remaining_time": 0, "track_confidence": 80}
08:07:15 1782828435368  B197.0       track          -150   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:16 1782828436373  B197.0       track          -150   500    85    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 500, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
08:07:17 1782828437370  B197.0       track          -160   500    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -160, "position_y": 500, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:18 1782828438379  B197.0       track          -150   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:19 1782828439394  B197.0       track          -150   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:20 1782828440376  B197.0       track          -150   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:21 1782828441376  B197.0       track          -150   480    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 480, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:22 1782828442374  B197.0       track          -150   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:23 1782828443375  B197.0       track          -150   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:24 1782828444271  B197.0       track          -150   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:25 1782828445276  B197.0       track          -150   490    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -150, "position_y": 490, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:26 1782828446271  B197.0       track          -140   490    77    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -140, "position_y": 490, "position_z": 77, "remaining_time": 0, "track_confidence": 80}
08:07:27 1782828447281  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:28 1782828448273  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:29 1782828449274  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:30 1782828450280  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:31 1782828451276  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:32 1782828452275  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:33 1782828453284  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:34 1782828454285  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:35 1782828455281  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:36 1782828456173  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:37 1782828457182  B197.0       track          -110   510    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:38 1782828458181  B197.0       track          -110   510    89    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 510, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
08:07:39 1782828459177  B197.0       track          -110   520    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 520, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:40 1782828460196  B197.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1782828460196, "event_status": "start", "number_people": 0, "respiratory_rate": -1}
08:07:40 1782828460234  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:41 1782828461190  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:42 1782828462200  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:07:43 1782828463200  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:08:14 1782828494873  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828494873, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 32, "respiratory_rate": -1, "multi_person_duration": 0}
08:08:14 1782828494873  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:08:14 1782828494994  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:08:46 1782828526608  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:09:18 1782828558472  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828558472, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:09:18 1782828558472  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:09:18 1782828558588  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:09:50 1782828590101  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:10:21 1782828621852  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:10:21 1782828621852  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828621852, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:10:21 1782828621976  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:10:53 1782828653638  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:11:25 1782828685340  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828685340, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:11:25 1782828685340  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:11:25 1782828685674  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:11:57 1782828717126  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:12:28 1782828748916  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828748916, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:12:28 1782828748916  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:12:29 1782828749049  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:13:00 1782828780557  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:13:32 1782828812376  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:13:32 1782828812376  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828812376, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:13:32 1782828812644  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:14:04 1782828844052  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:14:35 1782828875872  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828875872, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:14:35 1782828875872  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:14:36 1782828876236  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:15:07 1782828907556  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828907556, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:15:07 1782828907556  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:15:07 1782828907673  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:15:39 1782828939279  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:16:11 1782828971100  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782828971100, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:16:11 1782828971100  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:16:11 1782828971265  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:16:42 1782829002787  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:17:14 1782829034634  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782829034634, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:17:14 1782829034634  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:17:14 1782829034956  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:17:46 1782829066256  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:18:18 1782829098013  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782829098013, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:18:18 1782829098013  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:18:18 1782829098342  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:18:49 1782829129801  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:19:21 1782829161542  B197.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782829161542, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
08:19:21 1782829161542  B197.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
08:19:21 1782829161939  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
08:19:53 1782829193293  B197.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
```

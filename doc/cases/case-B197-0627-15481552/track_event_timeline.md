# case-B197-0627-15481552 — 每 tick belief 时间线 (room fd00:0:3:111:1:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
15:48:09 B197.E   -             -       0    NoReport np=0  ★0           room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:09 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:41 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:48:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:12 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:44 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:51 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:49:51 B197.0   B19705009530  stand   79   NoReport stand              room -    Empty      1   0     0.00  0.03  0.15  0.00  0.79  0.04
15:49:52 B197.0   B19705009530  walk    113  NoReport walk               room -    Empty      1   0     0.00  0.03  0.40  0.00  0.53  0.01
15:49:53 B197.0   B19705009530  walk    68   NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.46  0.00  0.38  0.02
15:49:54 B197.0   B19705009530  walk    75   NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.51  0.01  0.28  0.03
15:49:55 B197.0   B19705009530  walk    90   NoReport walk               room -    OpenFloor  1   0     0.00  0.03  0.70  0.01  0.13  0.02
15:49:56 B197.0   B19705009530  walk    100  NoReport walk               room -    OpenFloor  1   0     0.00  0.03  0.80  0.01  0.05  0.02
15:49:57 B197.0   B19705009530  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.71  0.01  0.04  0.04
15:49:58 B197.0   B19705009530  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.03  0.80  0.01  0.03  0.02
15:49:59 B197.0   B19705009530  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.83  0.00  0.02  0.02
15:50:00 B197.0   B19705009530  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
15:50:01 B197.0   B19705009530  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:01 B197.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:02 B197.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:02 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:03 B197.88  -             88      -    NoReport no-target(88)      room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  0.99
15:50:04 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:50:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:50:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:50:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:50:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:50:09 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:50:09 B197.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
15:50:09 B197.0   B19705009530  stand   105  NoReport stand              trk  0.88 Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
15:50:10 B197.0   B19705009530  walk    93   NoReport walk               trk  0.96 Empty      1   0     0.00  0.03  0.35  0.00  0.54  0.02
15:50:11 B197.0   B19705009530  walk    108  NoReport walk               trk  0.99 OpenFloor  1   0     0.00  0.04  0.42  0.01  0.41  0.02
15:50:12 B197.0   B19705009530  walk    91   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.03  0.65  0.01  0.20  0.02
15:50:13 B197.0   B19705009530  walk    90   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.03  0.77  0.00  0.08  0.02
15:50:14 B197.0   B19705009530  walk    85   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.03  0.80  0.01  0.03  0.02
15:50:15 B197.0   B19705009530  walk    85   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.71  0.01  0.03  0.04
15:50:16 B197.0   B19705009530  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.67  0.01  0.04  0.04
15:50:17 B197.0   B19705009530  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.65  0.01  0.05  0.03
15:50:18 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.77  0.01  0.03  0.02
15:50:19 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.82  0.01  0.02  0.02
15:50:20 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
15:50:21 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:22 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:23 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:24 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:25 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:26 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:27 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:28 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:29 B197.0   B19705009530  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:30 B197.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:30 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
15:50:31 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.01  0.06  0.58  0.01  0.03  0.06
15:50:32 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:33 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:34 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:35 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:36 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:37 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:38 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:39 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:40 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:41 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:42 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:43 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:44 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:45 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:46 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:47 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
15:50:48 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:49 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:50 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:51 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:52 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:53 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:54 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:55 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:56 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:57 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:58 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:50:59 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:00 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:01 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:02 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:03 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:04 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:05 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:06 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:07 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:08 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:09 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:10 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:11 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:12 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:13 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:14 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:15 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:16 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:17 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:18 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:19 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
15:51:20 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:21 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:22 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:23 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:24 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:25 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:26 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:27 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:28 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:29 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:30 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:31 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:32 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:33 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:34 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:35 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:36 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:37 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:38 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:39 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:40 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:41 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:42 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:43 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:44 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:45 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:46 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:47 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:48 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:49 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:50 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
15:51:51 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:51:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:51:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:51:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:51:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:51:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:51:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:51:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:51:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
15:52:23 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
15:52:55 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
15:48:09.494 B197.88   88     -    -      -      -     -    -    
15:48:41.405 B197.88   88     -    -      -      -     -    -    
15:49:12.884 B197.88   88     -    -      -      -     -    -    
15:49:44.694 B197.88   88     -    -      -      -     -    -    
15:49:51.630 B197.0    stand  255  -380   50     79    80        
15:49:52.577 B197.0    walk   255  -410   140    113   80   94   
15:49:53.496 B197.0    walk   255  -410   240    68    80   100  
15:49:54.502 B197.0    walk   255  -460   360    75    80   130  
15:49:55.498 B197.0    walk   255  -470   470    90    80   110  
15:49:56.504 B197.0    walk   255  -490   570    100   80   101  
15:49:57.504 B197.0    walk   255  -510   580    0     80   22   
15:49:58.510 B197.0    walk   255  -510   570    0     80   10   
15:49:59.503 B197.0    stand  255  -510   560    0     80   10   
15:50:00.502 B197.0    stand  255  -500   560    0     80   10   
15:50:01.510 B197.0    stand  3    -500   560    0     80   0    
15:50:02.565 B197.88   88     -    -      -      -     -    -    
15:50:03.416 B197.88   88     -    -      -      -     -    -    
15:50:04.415 B197.88   88     -    -      -      -     -    -    
15:50:09.530 B197.0    stand  3    -430   520    105   80   80   
15:50:10.360 B197.0    walk   3    -430   440    93    80   80   
15:50:11.364 B197.0    walk   3    -450   340    108   80   101  
15:50:12.366 B197.0    walk   3    -460   250    91    80   90   
15:50:13.362 B197.0    walk   3    -420   160    90    80   98   
15:50:14.433 B197.0    walk   3    -390   50     85    80   114  
15:50:15.366 B197.0    walk   3    -360   0      85    80   58   
15:50:16.370 B197.0    walk   3    -360   0      0     80   0    
15:50:17.373 B197.0    walk   3    -360   0      0     80   0    
15:50:18.368 B197.0    stand  3    -360   0      0     80   0    
15:50:19.373 B197.0    stand  3    -360   0      0     80   0    
15:50:20.370 B197.0    stand  3    -360   10     0     80   10   
15:50:21.269 B197.0    stand  3    -360   10     0     80   0    
15:50:22.272 B197.0    stand  3    -360   10     0     80   0    
15:50:23.280 B197.0    stand  3    -360   10     0     80   0    
15:50:24.283 B197.0    stand  3    -360   10     0     80   0    
15:50:25.283 B197.0    stand  3    -360   10     0     80   0    
15:50:26.276 B197.0    stand  3    -360   10     0     80   0    
15:50:27.281 B197.0    stand  3    -360   10     0     80   0    
15:50:28.277 B197.0    stand  3    -360   10     0     80   0    
15:50:29.278 B197.0    stand  3    -360   10     0     80   0    
15:50:30.333 B197.88   88     -    -      -      -     -    -    
15:50:31.291 B197.88   88     -    -      -      -     -    -    
15:50:32.192 B197.88   88     -    -      -      -     -    -    
15:50:48.080 B197.88   88     -    -      -      -     -    -    
15:51:20.238 B197.88   88     -    -      -      -     -    -    
15:51:51.610 B197.88   88     -    -      -      -     -    -    
15:52:23.516 B197.88   88     -    -      -      -     -    -    
15:52:55.030 B197.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 52 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

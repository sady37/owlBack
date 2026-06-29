# case-B197-0627-09280932 — 每 tick belief 时间线 (room fd00:0:3:111:1:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
09:28:17 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:48 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:28:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:20 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:43 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:43 B197.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:29:43 B197.0   B19703009313  stand   0    NoReport stand              room -    Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
09:29:43 B197.0   B19703009313  stand   123  NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
09:29:44 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.70  0.00  0.18  0.02
09:29:45 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.79  0.00  0.07  0.02
09:29:46 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.83  0.00  0.03  0.02
09:29:47 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.02  0.02
09:29:48 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:29:49 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:29:50 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:29:51 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:29:52 B197.0   B19703009313  stand   101  NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:29:53 B197.0   B19703009313  stand   82   NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:29:54 B197.0   B19703009313  walk    85   NoReport walk               room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:29:55 B197.0   B19703009313  walk    83   NoReport walk               room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:29:56 B197.0   B19703009313  walk    89   NoReport walk               room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:29:57 B197.0   B19703009313  walk    97   NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.74  0.00  0.02  0.04
09:29:58 B197.0   B19703009313  walk    107  NoReport walk               room -    OpenFloor  1   0     0.00  0.03  0.81  0.00  0.02  0.02
09:29:59 B197.0   B19703009313  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.72  0.01  0.02  0.04
09:30:00 B197.0   B19703009313  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.03  0.80  0.01  0.02  0.02
09:30:01 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.83  0.00  0.01  0.02
09:30:02 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
09:30:03 B197.0   B19703009313  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:03 B197.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:04 B197.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:04 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:05 B197.88  -             88      -    NoReport no-target(88)      room -    Left       0   0     0.00  0.00  0.00  0.00  0.00  0.99
09:30:06 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:30:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:30:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:30:09 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:30:09 B197.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
09:30:09 B197.0   B19703009313  stand   70   NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
09:30:10 B197.0   B19703009313  stand   95   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
09:30:11 B197.0   B19703009313  walk    106  NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.70  0.00  0.18  0.02
09:30:12 B197.0   B19703009313  walk    91   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.80  0.00  0.07  0.02
09:30:13 B197.0   B19703009313  walk    91   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.83  0.00  0.03  0.02
09:30:14 B197.0   B19703009313  walk    112  NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.02  0.02
09:30:15 B197.0   B19703009313  walk    79   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:16 B197.0   B19703009313  walk    100  NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:17 B197.0   B19703009313  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:18 B197.0   B19703009313  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.03  0.81  0.00  0.02  0.02
09:30:19 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
09:30:20 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
09:30:21 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:22 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:23 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:24 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:25 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:26 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:27 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:28 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:29 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:29 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:30 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:31 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:32 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:33 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:34 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:35 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:36 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:37 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:38 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:39 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:40 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:41 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:42 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
09:30:43 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   27    0.00  0.02  0.85  0.00  0.01  0.02
09:30:44 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   28    0.00  0.02  0.85  0.00  0.01  0.02
09:30:45 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
09:30:46 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
09:30:47 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
09:30:48 B197.0   B19703009313  stand   0    NoReport stand              trk  1.00 OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
09:30:49 B197.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
09:30:49 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
09:30:50 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.01  0.06  0.58  0.01  0.03  0.06
09:30:51 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
09:30:52 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
09:30:53 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
09:30:54 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.01  0.08  0.42  0.02  0.07  0.05
09:30:55 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:30:56 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:30:57 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:30:58 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:30:59 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:00 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:01 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:02 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:03 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:04 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:05 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:06 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:07 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:08 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:09 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:10 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:11 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:12 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:13 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:14 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:15 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:16 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:17 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:18 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:19 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:20 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:21 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:22 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:23 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:24 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:25 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:26 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.33  0.03  0.11  0.04
09:31:27 B197.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:28 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:29 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:30 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:31 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:32 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:33 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:34 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:35 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:36 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:37 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:38 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:39 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:40 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:41 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:42 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:43 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:44 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:45 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:46 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:47 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:48 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:49 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:50 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:51 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:52 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:53 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:54 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:55 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:56 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:57 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:58 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.23  0.06  0.15  0.03
09:31:59 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
09:32:31 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
09:28:17.032 B197.88   88     -    -      -      -     -    -    
09:28:48.987 B197.88   88     -    -      -      -     -    -    
09:29:20.640 B197.88   88     -    -      -      -     -    -    
09:29:43.362 B197.0    stand  1    -430   10     0     80        
09:29:43.428 B197.0    stand  1    -430   10     123   80   0    
09:29:44.324 B197.0    stand  1    -400   10     0     80   30   
09:29:45.331 B197.0    stand  1    -360   20     0     80   41   
09:29:46.327 B197.0    stand  1    -350   20     0     80   10   
09:29:47.330 B197.0    stand  1    -360   20     0     80   10   
09:29:48.330 B197.0    stand  1    -360   20     0     80   0    
09:29:49.333 B197.0    stand  1    -360   20     0     80   0    
09:29:50.333 B197.0    stand  1    -360   20     0     80   0    
09:29:51.333 B197.0    stand  1    -360   20     0     80   0    
09:29:52.340 B197.0    stand  1    -350   20     101   80   10   
09:29:53.335 B197.0    stand  1    -380   50     82    80   42   
09:29:54.335 B197.0    walk   1    -430   130    85    80   94   
09:29:55.336 B197.0    walk   1    -470   250    83    80   126  
09:29:56.256 B197.0    walk   1    -490   360    89    80   111  
09:29:57.232 B197.0    walk   1    -480   470    97    80   110  
09:29:58.234 B197.0    walk   1    -490   570    107   80   100  
09:29:59.235 B197.0    walk   1    -510   570    0     80   20   
09:30:00.236 B197.0    walk   1    -510   560    0     80   10   
09:30:01.236 B197.0    stand  1    -510   560    0     80   0    
09:30:02.245 B197.0    stand  1    -510   550    0     80   10   
09:30:03.240 B197.0    stand  3    -500   550    0     80   10   
09:30:04.312 B197.88   88     -    -      -      -     -    -    
09:30:05.255 B197.88   88     -    -      -      -     -    -    
09:30:06.150 B197.88   88     -    -      -      -     -    -    
09:30:09.313 B197.0    stand  3    -460   490    70    80   72   
09:30:10.121 B197.0    stand  3    -470   420    95    80   70   
09:30:11.124 B197.0    walk   3    -480   310    106   80   110  
09:30:12.121 B197.0    walk   3    -470   220    91    80   90   
09:30:13.121 B197.0    walk   3    -430   150    91    80   80   
09:30:14.125 B197.0    walk   3    -420   40     112   80   110  
09:30:15.124 B197.0    walk   3    -380   20     79    80   44   
09:30:16.129 B197.0    walk   3    -320   10     100   80   60   
09:30:17.130 B197.0    walk   3    -310   0      0     80   14   
09:30:18.176 B197.0    walk   3    -310   10     0     80   10   
09:30:19.128 B197.0    stand  3    -320   10     0     80   10   
09:30:20.134 B197.0    stand  3    -320   10     0     80   0    
09:30:21.028 B197.0    stand  3    -320   10     0     80   0    
09:30:22.032 B197.0    stand  3    -320   10     0     80   0    
09:30:23.033 B197.0    stand  3    -320   10     0     80   0    
09:30:24.031 B197.0    stand  3    -320   20     0     80   10   
09:30:25.063 B197.0    stand  3    -320   20     0     80   0    
09:30:26.058 B197.0    stand  3    -320   20     0     80   0    
09:30:27.068 B197.0    stand  3    -320   20     0     80   0    
09:30:28.061 B197.0    stand  3    -320   20     0     80   0    
09:30:29.064 B197.0    stand  3    -320   20     0     80   0    
09:30:29.976 B197.0    stand  3    -320   20     0     80   0    
09:30:30.968 B197.0    stand  3    -320   20     0     80   0    
09:30:31.967 B197.0    stand  3    -320   20     0     80   0    
09:30:32.961 B197.0    stand  3    -320   20     0     80   0    
09:30:33.960 B197.0    stand  3    -320   20     0     80   0    
09:30:34.961 B197.0    stand  3    -320   20     0     80   0    
09:30:35.968 B197.0    stand  3    -320   20     0     80   0    
09:30:36.999 B197.0    stand  3    -320   20     0     80   0    
09:30:37.964 B197.0    stand  3    -320   20     0     80   0    
09:30:38.965 B197.0    stand  3    -320   20     0     80   0    
09:30:39.966 B197.0    stand  3    -320   20     0     80   0    
09:30:40.969 B197.0    stand  3    -320   20     0     80   0    
09:30:41.865 B197.0    stand  3    -320   20     0     80   0    
09:30:42.889 B197.0    stand  3    -320   20     0     80   0    
09:30:43.866 B197.0    stand  3    -320   20     0     80   0    
09:30:44.864 B197.0    stand  3    -320   20     0     80   0    
09:30:45.865 B197.0    stand  3    -320   20     0     80   0    
09:30:46.866 B197.0    stand  3    -320   20     0     80   0    
09:30:47.868 B197.0    stand  3    -320   20     0     80   0    
09:30:48.868 B197.0    stand  3    -320   20     0     80   0    
09:30:49.921 B197.88   88     -    -      -      -     -    -    
09:30:50.880 B197.88   88     -    -      -      -     -    -    
09:30:51.878 B197.88   88     -    -      -      -     -    -    
09:30:55.746 B197.88   88     -    -      -      -     -    -    
09:31:27.804 B197.88   88     -    -      -      -     -    -    
09:31:59.237 B197.88   88     -    -      -      -     -    -    
09:32:31.092 B197.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 81 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

# case-ffdb-0701-12531255 — 每 tick belief 时间线 (room fd00:0:3:411:2:100, TZ Asia/Shanghai)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
12:53:00 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.50 Empty      1   0     0.00  0.02  0.15  0.00  0.79  0.04
12:53:01 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.51 Empty      1   0     0.00  0.02  0.25  0.00  0.67  0.01
12:53:02 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.52 Empty      1   1     0.00  0.02  0.35  0.00  0.53  0.02
12:53:03 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.53 OpenFloor  1   2     0.00  0.03  0.43  0.01  0.40  0.02
12:53:04 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.54 OpenFloor  1   3     0.00  0.03  0.49  0.01  0.29  0.02
12:53:05 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.55 OpenFloor  1   4     0.00  0.03  0.54  0.01  0.21  0.03
12:53:06 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   5     0.00  0.03  0.57  0.01  0.16  0.03
12:53:07 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   6     0.00  0.03  0.59  0.01  0.12  0.03
12:53:08 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   7     0.00  0.03  0.60  0.01  0.10  0.03
12:53:08 B267.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   7     0.00  0.03  0.60  0.01  0.10  0.03
12:53:09 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   8     0.00  0.03  0.61  0.02  0.08  0.03
12:53:10 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   9     0.00  0.03  0.62  0.02  0.07  0.03
12:53:11 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   10    0.01  0.03  0.62  0.02  0.07  0.03
12:53:12 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   11    0.01  0.03  0.62  0.02  0.06  0.03
12:53:13 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   12    0.01  0.03  0.63  0.02  0.06  0.03
12:53:14 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   13    0.01  0.03  0.63  0.02  0.06  0.03
12:53:15 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   14    0.01  0.03  0.63  0.02  0.06  0.03
12:53:16 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   15    0.01  0.03  0.63  0.02  0.06  0.03
12:53:17 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   16    0.01  0.03  0.63  0.02  0.06  0.03
12:53:17 D747.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   16    0.01  0.03  0.63  0.02  0.06  0.03
12:53:18 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   17    0.01  0.03  0.63  0.02  0.06  0.03
12:53:19 FFDB.0   FFDB05300569  stand   88   -        stand              trk  1.00 OpenFloor  1   18    0.00  0.04  0.62  0.02  0.05  0.03
12:53:20 FFDB.0   FFDB05300569  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
12:53:21 FFDB.0   FFDB05300569  walk    76   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
12:53:22 FFDB.0   FFDB05300569  walk    84   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
12:53:23 FFDB.0   FFDB05300569  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.58  0.02  0.05  0.03
12:53:24 FFDB.0   FFDB05300569  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.55  0.02  0.05  0.03
12:53:25 FFDB.0   FFDB05300569  walk    86   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.58  0.02  0.05  0.03
12:53:26 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.59  0.02  0.05  0.03
12:53:27 FFDB.0   FFDB05300569  stand   59   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.60  0.02  0.05  0.03
12:53:28 FFDB.0   FFDB05300569  walk    47   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.62  0.02  0.05  0.03
12:53:29 FFDB.0   FFDB05300569  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.62  0.02  0.05  0.03
12:53:30 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.62  0.02  0.05  0.03
12:53:31 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:53:32 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:53:33 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:53:34 FFDB.E0  FFDB05300569  -       0    -        Walking(rdr)       trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:53:34 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 OpenFloor  1   0     0.03  0.07  0.43  0.03  0.07  0.05
12:53:34 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 OpenFloor  1   0     0.10  0.08  0.31  0.04  0.10  0.04
12:53:35 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.26  0.07  0.21  0.04  0.10  0.02
12:53:36 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.52  0.05  0.12  0.03  0.07  0.01
12:53:37 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.76  0.02  0.05  0.02  0.04  0.01
12:53:38 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.90  0.01  0.02  0.01  0.02  0.00
12:53:39 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.96  0.00  0.01  0.00  0.01  0.00
12:53:40 B267.88  -             88      -    -        no-target(88)      room -    Fallen     1   0     0.96  0.00  0.01  0.00  0.01  0.00
12:53:40 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.98  0.00  0.01  0.00  0.00  0.00
12:53:41 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.99  0.00  0.00  0.00  0.00  0.00
12:53:42 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.99  0.00  0.00  0.00  0.00  0.00
12:53:43 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.99  0.00  0.00  0.00  0.00  0.00
12:53:44 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.99  0.00  0.00  0.00  0.00  0.00
12:53:45 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.99  0.00  0.00  0.00  0.00  0.00
12:53:46 FFDB.0   FFDB05300569  susfall 0    -        susfall            trk  1.00 Fallen     1   0     0.99  0.00  0.00  0.00  0.00  0.00
12:53:47 FFDB.0   FFDB05300569  susfall 25   -        susfall            trk  1.00 Fallen     1   0     0.99  0.00  0.00  0.00  0.00  0.00
12:53:47 FFDB.0   FFDB05300569  susfall 33   -        susfall            trk  1.00 Fallen     1   0     0.99  0.00  0.00  0.00  0.00  0.00
12:53:48 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 Fallen     1   0     0.97  0.00  0.02  0.00  0.00  0.00
12:53:49 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 Fallen     1   0     0.93  0.00  0.05  0.00  0.00  0.00
12:53:49 D747.88  -             88      -    -        no-target(88)      room -    Fallen     1   0     0.93  0.00  0.05  0.00  0.00  0.00
12:53:50 FFDB.0   FFDB05300569  stand   55   -        stand              trk  1.00 Fallen     1   0     0.88  0.00  0.08  0.00  0.00  0.00
12:53:51 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 Fallen     1   0     0.82  0.01  0.12  0.00  0.00  0.01
12:53:52 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 Fallen     1   0     0.73  0.01  0.16  0.00  0.01  0.01
12:53:53 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 Fallen     1   0     0.62  0.02  0.21  0.00  0.01  0.01
12:53:54 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 Fallen     1   0     0.50  0.02  0.26  0.01  0.02  0.01
12:53:55 FFDB.0   FFDB05300569  stand   49   -        stand              trk  1.00 Fallen     1   0     0.41  0.03  0.33  0.01  0.02  0.02
12:53:56 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.31  0.03  0.37  0.01  0.03  0.02
12:53:57 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.24  0.04  0.44  0.02  0.03  0.02
12:53:58 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.18  0.04  0.49  0.02  0.04  0.03
12:53:59 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.09  0.04  0.51  0.02  0.04  0.03
12:54:00 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.06  0.04  0.50  0.02  0.04  0.03
12:54:01 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.04  0.04  0.48  0.02  0.04  0.03
12:54:02 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.03  0.04  0.46  0.02  0.04  0.03
12:54:03 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.02  0.04  0.44  0.02  0.04  0.03
12:54:04 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.02  0.04  0.43  0.02  0.03  0.03
12:54:04 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.41  0.02  0.03  0.03
12:54:05 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.40  0.02  0.03  0.02
12:54:06 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   0     0.01  0.03  0.40  0.02  0.03  0.02
12:54:07 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 Sit        1   0     0.01  0.03  0.39  0.02  0.03  0.02
12:54:07 FFDB.0   FFDB05300569  stand   48   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.48  0.03  0.04  0.03
12:54:08 FFDB.0   FFDB05300569  stand   59   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.02  0.55  0.03  0.04  0.03
12:54:09 FFDB.0   FFDB05300569  stand   85   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.59  0.03  0.05  0.03
12:54:10 FFDB.0   FFDB05300569  stand   75   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.60  0.02  0.05  0.03
12:54:11 FFDB.0   FFDB05300569  stand   67   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.60  0.02  0.05  0.03
12:54:12 B267.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   0     0.01  0.04  0.60  0.02  0.05  0.03
12:54:12 FFDB.0   FFDB05300569  stand   77   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.62  0.02  0.05  0.03
12:54:13 FFDB.0   FFDB05300569  stand   64   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.02  0.63  0.02  0.06  0.03
12:54:14 FFDB.0   FFDB05300569  stand   67   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.02  0.63  0.02  0.06  0.03
12:54:15 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   0     0.01  0.02  0.63  0.02  0.06  0.03
12:54:16 FFDB.0   FFDB05300569  walk    52   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.64  0.02  0.06  0.03
12:54:16 FFDB.0   FFDB05300569  walk    50   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.64  0.02  0.06  0.03
12:54:17 FFDB.0   FFDB05300569  walk    64   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.64  0.01  0.06  0.03
12:54:18 FFDB.0   FFDB05300569  walk    57   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.64  0.01  0.06  0.03
12:54:19 FFDB.0   FFDB05300569  walk    65   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.64  0.01  0.06  0.03
12:54:20 D747.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   0     0.00  0.02  0.64  0.01  0.06  0.03
12:54:20 FFDB.0   FFDB05300569  walk    65   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.03  0.64  0.01  0.06  0.03
12:54:21 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.64  0.01  0.06  0.03
12:54:22 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.63  0.01  0.06  0.03
12:54:23 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.01  0.06  0.03
12:54:24 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.01  0.06  0.03
12:54:25 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.01  0.06  0.03
12:54:26 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.01  0.06  0.03
12:54:27 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.01  0.06  0.03
12:54:28 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.01  0.06  0.03
12:54:29 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.01  0.06  0.03
12:54:30 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.01  0.06  0.03
12:54:31 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.01  0.06  0.03
12:54:32 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.02  0.06  0.03
12:54:33 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.02  0.06  0.03
12:54:34 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.02  0.06  0.03
12:54:35 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.02  0.06  0.03
12:54:36 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.02  0.06  0.03
12:54:37 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.62  0.01  0.05  0.03
12:54:38 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.62  0.02  0.05  0.03
12:54:39 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.62  0.02  0.05  0.03
12:54:40 FFDB.0   FFDB05300569  stand   25   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:41 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:42 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.02  0.05  0.03
12:54:43 B267.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   0     0.01  0.03  0.63  0.02  0.05  0.03
12:54:43 FFDB.0   FFDB05300569  stand   12   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.02  0.05  0.03
12:54:44 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.63  0.02  0.06  0.03
12:54:45 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.62  0.02  0.05  0.03
12:54:46 FFDB.0   FFDB05300569  stand   37   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.62  0.02  0.05  0.03
12:54:47 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:48 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:49 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:50 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:51 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:52 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:52 D747.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:53 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:54 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:55 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:56 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:57 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:58 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:54:59 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.07  0.50  0.03  0.08  0.03
12:54:59 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.55  0.03  0.07  0.03
12:55:00 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.57  0.03  0.07  0.03
12:55:01 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.59  0.02  0.06  0.03
12:55:02 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.60  0.02  0.06  0.03
12:55:03 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.60  0.02  0.06  0.03
12:55:04 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:05 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:06 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:07 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:08 FFDB.0   FFDB05300569  walk    23   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
12:55:09 FFDB.0   FFDB05300569  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
12:55:10 FFDB.0   FFDB05300569  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
12:55:11 FFDB.0   FFDB05300569  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
12:55:12 FFDB.0   FFDB05300569  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
12:55:13 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
12:55:14 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:15 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:15 B267.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:16 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:17 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:18 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:19 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:20 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:21 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:22 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:23 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:24 D747.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:24 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:25 FFDB.E   -             -       0    -        np=2               room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
12:55:25 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.01  0.03  0.76  0.01  0.03  0.02
12:55:25 FFDB.1   FFDB15525594  stand   56   -        stand              trk  0.50 OpenFloor  2   0     0.00  0.02  0.15  0.00  0.79  0.04
12:55:26 FFDB.1   FFDB15525594  stand   69   -        stand              trk  0.51 OpenFloor  2   0     0.00  0.01  0.40  0.00  0.54  0.01
12:55:26 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.82  0.01  0.02  0.02
12:55:27 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
12:55:27 FFDB.1   FFDB15525594  stand   57   -        stand              trk  0.52 OpenFloor  2   0     0.00  0.01  0.64  0.00  0.27  0.01
12:55:28 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
12:55:28 FFDB.1   FFDB15525594  stand   50   -        stand              trk  0.53 OpenFloor  2   0     0.00  0.01  0.77  0.00  0.11  0.02
12:55:29 FFDB.1   FFDB15525594  stand   40   -        stand              trk  0.54 OpenFloor  2   0     0.00  0.01  0.83  0.00  0.05  0.02
12:55:29 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:30 FFDB.0   FFDB05300569  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:30 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  2   0     0.00  0.01  0.83  0.00  0.02  0.02
12:55:31 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.22 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:31 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.85  0.00  0.01  0.02
12:55:32 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.07 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:32 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:33 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.03 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:33 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:34 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:34 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.03 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:35 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.03 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:35 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:36 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.03 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:36 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:37 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:37 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:38 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:38 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:39 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:39 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:40 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:40 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:41 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:41 FFDB.1   FFDB15525594  stand   38   -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:42 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:42 FFDB.1   FFDB15525594  stand   57   -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:43 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:43 FFDB.1   FFDB15525594  stand   72   -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:44 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:44 FFDB.1   FFDB15525594  stand   54   -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:45 FFDB.1   FFDB15525594  stand   71   -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:45 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:46 FFDB.1   FFDB15525594  walk    43   -        walk               trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:46 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:46 B267.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:47 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:47 FFDB.1   FFDB15525594  walk    60   -        walk               trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:48 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:48 FFDB.1   FFDB15525594  walk    67   -        walk               trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:49 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:49 FFDB.1   FFDB15525594  walk    67   -        walk               trk  0.90 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:50 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:50 FFDB.1   FFDB15525594  walk    70   -        walk               trk  0.90 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:51 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:51 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:52 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:52 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:53 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:53 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:54 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:54 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:55 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:55 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:56 FFDB.1   FFDB15525594  stand   102  -        stand              trk  0.90 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:56 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:56 D747.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:57 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.01  0.86  0.00  0.01  0.02
12:55:57 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
12:55:58 FFDB.0   FFDB05300569  stand   0    -        stand              trk  0.02 OpenFloor  1   0     0.00  0.03  0.81  0.00  0.02  0.02
12:55:58 FFDB.1   FFDB15525594  stand   0    -        stand              trk  0.90 OpenFloor  1   0     0.00  0.03  0.70  0.01  0.03  0.04
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
12:53:08.627 B267.88   88     -    -      -      -     -    -    
12:53:40.017 B267.88   88     -    -      -      -     -    -    
12:54:12.084 B267.88   88     -    -      -      -     -    -    
12:54:43.505 B267.88   88     -    -      -      -     -    -    
12:55:15.572 B267.88   88     -    -      -      -     -    -    
12:55:46.995 B267.88   88     -    -      -      -     -    -    

12:53:17.490 D747.88   88     -    -      -      -     -    -    
12:53:49.338 D747.88   88     -    -      -      -     -    -    
12:54:20.714 D747.88   88     -    -      -      -     -    -    
12:54:52.724 D747.88   88     -    -      -      -     -    -    
12:55:24.264 D747.88   88     -    -      -      -     -    -    
12:55:56.212 D747.88   88     -    -      -      -     -    -    

12:53:00.569 FFDB.0    stand  6    -350   260    0     80        
12:53:01.457 FFDB.0    stand  6    -350   260    0     80   0    
12:53:02.460 FFDB.0    stand  6    -350   260    0     80   0    
12:53:03.459 FFDB.0    stand  6    -350   260    0     80   0    
12:53:04.463 FFDB.0    stand  6    -350   260    0     80   0    
12:53:05.463 FFDB.0    stand  6    -350   260    0     80   0    
12:53:06.470 FFDB.0    stand  6    -350   260    0     80   0    
12:53:07.466 FFDB.0    stand  6    -350   260    0     80   0    
12:53:08.494 FFDB.0    stand  6    -350   260    0     80   0    
12:53:09.505 FFDB.0    stand  6    -350   260    0     80   0    
12:53:10.468 FFDB.0    stand  6    -350   260    0     80   0    
12:53:11.472 FFDB.0    stand  6    -350   260    0     80   0    
12:53:12.472 FFDB.0    stand  6    -350   260    0     80   0    
12:53:13.362 FFDB.0    stand  6    -350   260    0     80   0    
12:53:14.372 FFDB.0    stand  6    -350   260    0     80   0    
12:53:15.366 FFDB.0    stand  6    -350   260    0     80   0    
12:53:16.364 FFDB.0    stand  6    -350   260    0     80   0    
12:53:17.365 FFDB.0    stand  6    -350   260    0     80   0    
12:53:18.368 FFDB.0    stand  6    -350   260    0     80   0    
12:53:19.372 FFDB.0    stand  6    -400   290    88    80   58   
12:53:20.383 FFDB.0    walk   6    -470   340    0     80   86   
12:53:21.384 FFDB.0    walk   6    -440   320    76    80   36   
12:53:22.386 FFDB.0    walk   6    -380   280    84    80   72   
12:53:23.281 FFDB.0    walk   6    -420   250    0     80   50   
12:53:24.283 FFDB.0    walk   6    -390   290    0     80   50   
12:53:25.283 FFDB.0    walk   6    -390   300    86    80   10   
12:53:26.284 FFDB.0    stand  6    -400   320    0     80   22   
12:53:27.285 FFDB.0    stand  6    -380   280    59    80   44   
12:53:28.286 FFDB.0    walk   6    -340   250    47    80   50   
12:53:29.289 FFDB.0    walk   6    -380   270    0     80   44   
12:53:30.290 FFDB.0    stand  6    -370   250    0     80   22   
12:53:31.292 FFDB.0    stand  6    -370   250    0     80   0    
12:53:32.291 FFDB.0    stand  6    -370   250    0     80   0    
12:53:34.002 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:34.289 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:35.194 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:36.197 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:37.211 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:38.209 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:39.208 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:40.224 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:41.211 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:42.210 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:43.212 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:44.216 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:45.118 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:46.107 FFDB.0    susfall 6    -370   250    0     80   0    
12:53:47.109 FFDB.0    susfall 6    -350   260    25    80   22   
12:53:47.312 FFDB.0    susfall 6    -340   260    33    80   10   
12:53:48.315 FFDB.0    stand  6    -340   250    0     80   10   
12:53:49.143 FFDB.0    stand  6    -340   250    0     80   0    
12:53:50.139 FFDB.0    stand  6    -380   270    55    80   44   
12:53:51.142 FFDB.0    stand  6    -420   250    0     80   44   
12:53:52.141 FFDB.0    stand  6    -420   250    0     80   0    
12:53:53.046 FFDB.0    stand  6    -420   250    0     80   0    
12:53:54.046 FFDB.0    stand  6    -420   250    0     80   0    
12:53:55.056 FFDB.0    stand  6    -380   270    49    80   44   
12:53:56.048 FFDB.0    stand  6    -400   280    0     80   22   
12:53:57.051 FFDB.0    stand  6    -400   280    0     80   0    
12:53:58.049 FFDB.0    stand  6    -400   280    0     80   0    
12:53:59.272 FFDB.0    stand  6    -400   280    0     80   0    
12:54:00.048 FFDB.0    stand  6    -400   270    0     80   10   
12:54:01.051 FFDB.0    stand  6    -400   270    0     80   0    
12:54:02.057 FFDB.0    stand  6    -400   270    0     80   0    
12:54:03.061 FFDB.0    stand  6    -400   270    0     80   0    
12:54:04.052 FFDB.0    stand  6    -400   270    0     80   0    
12:54:04.946 FFDB.0    stand  6    -400   270    0     80   0    
12:54:05.949 FFDB.0    stand  6    -400   270    0     80   0    
12:54:07.051 FFDB.0    stand  6    -400   270    0     80   0    
12:54:07.949 FFDB.0    stand  6    -340   290    48    80   63   
12:54:08.918 FFDB.0    stand  6    -320   270    59    80   28   
12:54:09.917 FFDB.0    stand  6    -330   280    85    80   14   
12:54:10.920 FFDB.0    stand  6    -360   290    75    80   31   
12:54:11.919 FFDB.0    stand  6    -360   240    67    80   50   
12:54:12.923 FFDB.0    stand  6    -350   210    77    80   31   
12:54:13.927 FFDB.0    stand  6    -330   210    64    80   20   
12:54:14.921 FFDB.0    stand  6    -310   210    67    80   20   
12:54:16.473 FFDB.0    walk   6    -270   160    52    80   64   
12:54:16.924 FFDB.0    walk   6    -260   130    50    80   31   
12:54:17.930 FFDB.0    walk   6    -270   130    64    80   10   
12:54:18.930 FFDB.0    walk   6    -280   110    57    80   22   
12:54:19.930 FFDB.0    walk   6    -290   120    65    80   14   
12:54:20.820 FFDB.0    walk   6    -290   100    65    80   20   
12:54:21.821 FFDB.0    stand  6    -310   120    0     80   28   
12:54:22.821 FFDB.0    stand  6    -310   120    0     80   0    
12:54:23.822 FFDB.0    stand  6    -310   120    0     80   0    
12:54:24.847 FFDB.0    stand  6    -310   120    0     80   0    
12:54:25.836 FFDB.0    stand  6    -310   120    0     80   0    
12:54:26.838 FFDB.0    stand  6    -310   120    0     80   0    
12:54:27.841 FFDB.0    stand  6    -310   120    0     80   0    
12:54:28.842 FFDB.0    stand  6    -310   120    0     80   0    
12:54:29.841 FFDB.0    stand  6    -310   120    0     80   0    
12:54:30.741 FFDB.0    stand  6    -310   120    0     80   0    
12:54:31.741 FFDB.0    stand  6    -310   120    0     80   0    
12:54:32.740 FFDB.0    stand  6    -310   120    0     80   0    
12:54:33.745 FFDB.0    stand  6    -310   120    0     80   0    
12:54:34.742 FFDB.0    stand  6    -310   120    0     80   0    
12:54:35.748 FFDB.0    stand  6    -310   120    0     80   0    
12:54:36.747 FFDB.0    stand  6    -250   90     0     80   67   
12:54:37.748 FFDB.0    stand  6    -230   80     0     80   22   
12:54:38.747 FFDB.0    stand  6    -230   80     0     80   0    
12:54:39.746 FFDB.0    stand  6    -240   80     0     80   10   
12:54:40.665 FFDB.0    stand  6    -240   60     25    80   20   
12:54:41.662 FFDB.0    stand  6    -230   80     0     80   22   
12:54:42.663 FFDB.0    stand  6    -230   90     0     80   10   
12:54:43.665 FFDB.0    stand  6    -220   80     12    80   14   
12:54:44.662 FFDB.0    stand  6    -220   80     0     80   0    
12:54:45.665 FFDB.0    stand  6    -180   20     0     80   72   
12:54:46.668 FFDB.0    stand  6    -210   40     37    80   36   
12:54:47.666 FFDB.0    stand  6    -220   50     0     80   14   
12:54:48.667 FFDB.0    stand  6    -190   10     0     80   50   
12:54:49.673 FFDB.0    stand  6    -170   -10    0     80   28   
12:54:50.673 FFDB.0    stand  6    -160   -10    0     80   10   
12:54:51.673 FFDB.0    stand  6    -170   -10    0     80   10   
12:54:52.561 FFDB.0    stand  6    -170   -10    0     80   0    
12:54:53.565 FFDB.0    stand  6    -170   -10    0     80   0    
12:54:54.564 FFDB.0    stand  6    -170   -10    0     80   0    
12:54:55.568 FFDB.0    stand  6    -170   -10    0     80   0    
12:54:56.534 FFDB.0    stand  6    -170   -10    0     80   0    
12:54:57.537 FFDB.0    stand  6    -170   -10    0     80   0    
12:54:59.485 FFDB.0    stand  6    -170   -10    0     80   0    
12:54:59.689 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:00.534 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:01.538 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:02.536 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:03.539 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:04.540 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:05.547 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:06.549 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:07.542 FFDB.0    stand  6    -260   70     0     80   120  
12:55:08.435 FFDB.0    walk   6    -230   60     23    80   31   
12:55:09.517 FFDB.0    walk   6    -170   0      0     80   84   
12:55:10.437 FFDB.0    walk   6    -160   -10    0     80   14   
12:55:11.438 FFDB.0    walk   6    -170   -10    0     80   10   
12:55:12.469 FFDB.0    walk   6    -170   -10    0     80   0    
12:55:13.478 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:14.474 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:15.475 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:16.378 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:17.371 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:18.372 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:19.376 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:20.386 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:21.375 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:22.378 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:23.382 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:24.383 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:25.594 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:25.594 FFDB.1    stand  255  -260   100    56    80   142  
12:55:26.389 FFDB.1    stand  255  -250   90     69    80   14   
12:55:26.389 FFDB.0    stand  6    -170   -10    0     80   128  
12:55:27.294 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:27.294 FFDB.1    stand  255  -270   100    57    80   148  
12:55:28.295 FFDB.0    stand  6    -170   -10    0     80   148  
12:55:28.295 FFDB.1    stand  255  -280   110    50    80   162  
12:55:29.291 FFDB.1    stand  255  -310   120    40    80   31   
12:55:29.291 FFDB.0    stand  6    -170   -10    0     80   191  
12:55:30.292 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:30.292 FFDB.1    stand  255  -330   120    0     80   206  
12:55:31.294 FFDB.0    stand  6    -170   -10    0     80   206  
12:55:31.294 FFDB.1    stand  255  -330   120    0     80   206  
12:55:32.293 FFDB.0    stand  6    -170   -10    0     80   206  
12:55:32.293 FFDB.1    stand  255  -330   120    0     80   206  
12:55:33.296 FFDB.0    stand  6    -170   -10    0     80   206  
12:55:33.296 FFDB.1    stand  255  -330   120    0     80   206  
12:55:34.296 FFDB.1    stand  255  -330   120    0     80   0    
12:55:34.296 FFDB.0    stand  6    -180   0      0     80   192  
12:55:35.298 FFDB.0    stand  6    -200   10     0     80   22   
12:55:35.298 FFDB.1    stand  255  -330   120    0     80   170  
12:55:36.305 FFDB.0    stand  6    -200   10     0     80   170  
12:55:36.305 FFDB.1    stand  255  -330   120    0     80   170  
12:55:37.298 FFDB.0    stand  6    -170   -10    0     80   206  
12:55:37.298 FFDB.1    stand  255  -330   120    0     80   206  
12:55:38.194 FFDB.0    stand  6    -170   -10    0     80   206  
12:55:38.194 FFDB.1    stand  255  -330   120    0     80   206  
12:55:39.195 FFDB.0    stand  6    -170   -10    0     80   206  
12:55:39.195 FFDB.1    stand  255  -330   120    0     80   206  
12:55:40.198 FFDB.0    stand  6    -170   -10    0     80   206  
12:55:40.198 FFDB.1    stand  255  -330   120    0     80   206  
12:55:41.200 FFDB.0    stand  6    -170   -10    0     80   206  
12:55:41.200 FFDB.1    stand  255  -260   90     38    80   134  
12:55:42.201 FFDB.0    stand  6    -170   -10    0     80   134  
12:55:42.201 FFDB.1    stand  255  -260   110    57    80   150  
12:55:43.206 FFDB.0    stand  6    -170   -10    0     80   150  
12:55:43.206 FFDB.1    stand  255  -260   120    72    80   158  
12:55:44.200 FFDB.0    stand  6    -170   -10    0     80   158  
12:55:44.200 FFDB.1    stand  255  -280   150    54    80   194  
12:55:45.151 FFDB.1    stand  255  -310   190    71    80   50   
12:55:45.151 FFDB.0    stand  6    -170   -10    0     80   244  
12:55:46.145 FFDB.1    walk   255  -300   220    43    80   264  
12:55:46.145 FFDB.0    stand  6    -170   -10    0     80   264  
12:55:47.146 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:47.146 FFDB.1    walk   255  -310   270    60    80   313  
12:55:48.147 FFDB.0    stand  6    -170   -10    0     80   313  
12:55:48.147 FFDB.1    walk   255  -360   270    67    80   338  
12:55:49.148 FFDB.0    stand  6    -170   -10    0     80   338  
12:55:49.148 FFDB.1    walk   255  -370   290    67    80   360  
12:55:50.149 FFDB.0    stand  6    -170   -10    0     80   360  
12:55:50.149 FFDB.1    walk   255  -350   290    70    80   349  
12:55:51.150 FFDB.1    stand  255  -370   300    0     80   22   
12:55:51.150 FFDB.0    stand  6    -170   -10    0     80   368  
12:55:52.160 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:52.160 FFDB.1    stand  255  -370   300    0     80   368  
12:55:53.158 FFDB.0    stand  6    -170   -10    0     80   368  
12:55:53.158 FFDB.1    stand  255  -370   290    0     80   360  
12:55:54.162 FFDB.0    stand  6    -170   -10    0     80   360  
12:55:54.162 FFDB.1    stand  255  -370   290    0     80   360  
12:55:55.157 FFDB.1    stand  255  -370   290    0     80   0    
12:55:55.157 FFDB.0    stand  6    -170   -10    0     80   360  
12:55:56.056 FFDB.1    stand  255  -360   300    102   80   363  
12:55:56.056 FFDB.0    stand  6    -170   -10    0     80   363  
12:55:57.054 FFDB.1    stand  255  -340   290    0     80   344  
12:55:57.054 FFDB.0    stand  6    -170   -10    0     80   344  
12:55:58.278 FFDB.0    stand  6    -170   -10    0     80   0    
12:55:58.278 FFDB.1    stand  255  -340   290    0     80   344  

```

**汇总**: xray tick 218 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

## 完整原始记录（按时间排序，data_value 全文不删字段）
```
time     ms             device.tid   event          x      y      z     原始记录
12:53:00 1782881580569  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:01 1782881581457  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:02 1782881582460  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:03 1782881583459  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:04 1782881584463  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:05 1782881585463  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:06 1782881586470  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:07 1782881587466  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:08 1782881588288  B267.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881588288, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 0}
12:53:08 1782881588288  B267.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:53:08 1782881588494  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:08 1782881588627  B267.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:09 1782881589505  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:10 1782881590468  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:11 1782881591472  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:12 1782881592472  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:13 1782881593362  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:14 1782881594372  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:15 1782881595366  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:16 1782881596364  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:17 1782881597228  D747.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:53:17 1782881597228  D747.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881597228, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 0}
12:53:17 1782881597365  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:17 1782881597490  D747.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:18 1782881598368  FFDB.0       track          -350   260    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:19 1782881599372  FFDB.0       track          -400   290    88    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 290, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
12:53:20 1782881600383  FFDB.0       track          -470   340    0     {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -470, "position_y": 340, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:21 1782881601384  FFDB.0       track          -440   320    76    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -440, "position_y": 320, "position_z": 76, "remaining_time": 0, "track_confidence": 80}
12:53:22 1782881602386  FFDB.0       track          -380   280    84    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -380, "position_y": 280, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
12:53:23 1782881603281  FFDB.0       track          -420   250    0     {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -420, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:24 1782881604283  FFDB.0       track          -390   290    0     {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -390, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:25 1782881605283  FFDB.0       track          -390   300    86    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -390, "position_y": 300, "position_z": 86, "remaining_time": 0, "track_confidence": 80}
12:53:26 1782881606284  FFDB.0       track          -400   320    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 320, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:27 1782881607285  FFDB.0       track          -380   280    59    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -380, "position_y": 280, "position_z": 59, "remaining_time": 0, "track_confidence": 80}
12:53:28 1782881608286  FFDB.0       track          -340   250    47    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -340, "position_y": 250, "position_z": 47, "remaining_time": 0, "track_confidence": 80}
12:53:29 1782881609289  FFDB.0       track          -380   270    0     {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -380, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:30 1782881610290  FFDB.0       track          -370   250    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:31 1782881611292  FFDB.0       track          -370   250    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:32 1782881612291  FFDB.0       track          -370   250    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:34 1782881614000  FFDB.0       Walking        -      -      -     {"pose": 1, "track_id": 0, "event_type": 2, "event_since": 1782881614000, "event_status": "start"}
12:53:34 1782881614002  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:34 1782881614289  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:35 1782881615194  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:36 1782881616197  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:37 1782881617211  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:38 1782881618209  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:39 1782881619208  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:40 1782881620017  B267.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:40 1782881620224  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:41 1782881621211  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:42 1782881622210  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:43 1782881623212  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:44 1782881624216  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:45 1782881625118  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:46 1782881626107  FFDB.0       track          -370   250    0     {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -370, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:47 1782881627109  FFDB.0       track          -350   260    25    {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 260, "position_z": 25, "remaining_time": 0, "track_confidence": 80}
12:53:47 1782881627312  FFDB.0       track          -340   260    33    {"pose": 2, "area_id": 6, "track_id": 0, "position_x": -340, "position_y": 260, "position_z": 33, "remaining_time": 0, "track_confidence": 80}
12:53:48 1782881628132  FFDB.0       Initialization -      -      -     {"pose": 0, "track_id": 0, "event_type": 2, "event_since": 1782881628132, "event_status": "start"}
12:53:48 1782881628315  FFDB.0       track          -340   250    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -340, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:49 1782881629031  D747.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:53:49 1782881629031  D747.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881629031, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 0}
12:53:49 1782881629143  FFDB.0       track          -340   250    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -340, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:49 1782881629338  D747.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:50 1782881630139  FFDB.0       track          -380   270    55    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -380, "position_y": 270, "position_z": 55, "remaining_time": 0, "track_confidence": 80}
12:53:51 1782881631142  FFDB.0       track          -420   250    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -420, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:52 1782881632141  FFDB.0       track          -420   250    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -420, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:53 1782881633046  FFDB.0       track          -420   250    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -420, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:54 1782881634046  FFDB.0       track          -420   250    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -420, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:55 1782881635056  FFDB.0       track          -380   270    49    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -380, "position_y": 270, "position_z": 49, "remaining_time": 0, "track_confidence": 80}
12:53:56 1782881636048  FFDB.0       track          -400   280    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:57 1782881637051  FFDB.0       track          -400   280    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:58 1782881638049  FFDB.0       track          -400   280    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:53:59 1782881639065  FFDB.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881639065, "event_status": "instant", "lie_duration": 15, "walk_distance": 3, "walk_duration": 8, "stand_duration": 37, "multi_person_duration": 0}
12:53:59 1782881639065  FFDB.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:53:59 1782881639272  FFDB.0       track          -400   280    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:00 1782881640048  FFDB.0       track          -400   270    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:01 1782881641051  FFDB.0       track          -400   270    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:02 1782881642057  FFDB.0       track          -400   270    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:03 1782881643061  FFDB.0       track          -400   270    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:04 1782881644052  FFDB.0       track          -400   270    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:04 1782881644946  FFDB.0       track          -400   270    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:05 1782881645949  FFDB.0       track          -400   270    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:07 1782881647051  FFDB.0       track          -400   270    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -400, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:07 1782881647949  FFDB.0       track          -340   290    48    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -340, "position_y": 290, "position_z": 48, "remaining_time": 0, "track_confidence": 80}
12:54:08 1782881648918  FFDB.0       track          -320   270    59    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -320, "position_y": 270, "position_z": 59, "remaining_time": 0, "track_confidence": 80}
12:54:09 1782881649917  FFDB.0       track          -330   280    85    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -330, "position_y": 280, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
12:54:10 1782881650920  FFDB.0       track          -360   290    75    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -360, "position_y": 290, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
12:54:11 1782881651798  B267.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881651798, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 0}
12:54:11 1782881651798  B267.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:54:11 1782881651919  FFDB.0       track          -360   240    67    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -360, "position_y": 240, "position_z": 67, "remaining_time": 0, "track_confidence": 80}
12:54:12 1782881652084  B267.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:12 1782881652923  FFDB.0       track          -350   210    77    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -350, "position_y": 210, "position_z": 77, "remaining_time": 0, "track_confidence": 80}
12:54:13 1782881653927  FFDB.0       track          -330   210    64    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -330, "position_y": 210, "position_z": 64, "remaining_time": 0, "track_confidence": 80}
12:54:14 1782881654921  FFDB.0       track          -310   210    67    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 210, "position_z": 67, "remaining_time": 0, "track_confidence": 80}
12:54:16 1782881656473  FFDB.0       track          -270   160    52    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -270, "position_y": 160, "position_z": 52, "remaining_time": 0, "track_confidence": 80}
12:54:16 1782881656924  FFDB.0       track          -260   130    50    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -260, "position_y": 130, "position_z": 50, "remaining_time": 0, "track_confidence": 80}
12:54:17 1782881657930  FFDB.0       track          -270   130    64    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -270, "position_y": 130, "position_z": 64, "remaining_time": 0, "track_confidence": 80}
12:54:18 1782881658930  FFDB.0       track          -280   110    57    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -280, "position_y": 110, "position_z": 57, "remaining_time": 0, "track_confidence": 80}
12:54:19 1782881659930  FFDB.0       track          -290   120    65    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -290, "position_y": 120, "position_z": 65, "remaining_time": 0, "track_confidence": 80}
12:54:20 1782881660714  D747.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:20 1782881660820  FFDB.0       track          -290   100    65    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -290, "position_y": 100, "position_z": 65, "remaining_time": 0, "track_confidence": 80}
12:54:21 1782881661821  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:22 1782881662821  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:23 1782881663822  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:24 1782881664847  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:25 1782881665836  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:26 1782881666838  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:27 1782881667841  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:28 1782881668842  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:29 1782881669841  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:30 1782881670741  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:31 1782881671741  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:32 1782881672740  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:33 1782881673745  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:34 1782881674742  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:35 1782881675748  FFDB.0       track          -310   120    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -310, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:36 1782881676747  FFDB.0       track          -250   90     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -250, "position_y": 90, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:37 1782881677748  FFDB.0       track          -230   80     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -230, "position_y": 80, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:38 1782881678747  FFDB.0       track          -230   80     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -230, "position_y": 80, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:39 1782881679746  FFDB.0       track          -240   80     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -240, "position_y": 80, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:40 1782881680665  FFDB.0       track          -240   60     25    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -240, "position_y": 60, "position_z": 25, "remaining_time": 0, "track_confidence": 80}
12:54:41 1782881681662  FFDB.0       track          -230   80     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -230, "position_y": 80, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:42 1782881682663  FFDB.0       track          -230   90     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -230, "position_y": 90, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:43 1782881683505  B267.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:43 1782881683665  FFDB.0       track          -220   80     12    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -220, "position_y": 80, "position_z": 12, "remaining_time": 0, "track_confidence": 80}
12:54:44 1782881684662  FFDB.0       track          -220   80     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -220, "position_y": 80, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:45 1782881685665  FFDB.0       track          -180   20     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -180, "position_y": 20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:46 1782881686668  FFDB.0       track          -210   40     37    {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -210, "position_y": 40, "position_z": 37, "remaining_time": 0, "track_confidence": 80}
12:54:47 1782881687666  FFDB.0       track          -220   50     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -220, "position_y": 50, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:48 1782881688667  FFDB.0       track          -190   10     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -190, "position_y": 10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:49 1782881689673  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:50 1782881690673  FFDB.0       track          -160   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:51 1782881691673  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:52 1782881692460  D747.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881692460, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 0}
12:54:52 1782881692460  D747.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:54:52 1782881692561  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:52 1782881692724  D747.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:53 1782881693565  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:54 1782881694564  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:55 1782881695568  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:56 1782881696534  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:57 1782881697537  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:59 1782881699482  FFDB.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881699482, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 6, "stand_duration": 54, "multi_person_duration": 0}
12:54:59 1782881699482  FFDB.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:54:59 1782881699485  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:54:59 1782881699689  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:00 1782881700534  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:01 1782881701538  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:02 1782881702536  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:03 1782881703539  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:04 1782881704540  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:05 1782881705547  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:06 1782881706549  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:07 1782881707542  FFDB.0       track          -260   70     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -260, "position_y": 70, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:08 1782881708435  FFDB.0       track          -230   60     23    {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -230, "position_y": 60, "position_z": 23, "remaining_time": 0, "track_confidence": 80}
12:55:09 1782881709517  FFDB.0       track          -170   0      0     {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:10 1782881710437  FFDB.0       track          -160   -10    0     {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:11 1782881711438  FFDB.0       track          -170   -10    0     {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:12 1782881712469  FFDB.0       track          -170   -10    0     {"pose": 1, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:13 1782881713478  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:14 1782881714474  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:15 1782881715320  B267.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881715320, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 0}
12:55:15 1782881715320  B267.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:55:15 1782881715475  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:15 1782881715572  B267.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:16 1782881716378  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:17 1782881717371  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:18 1782881718372  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:19 1782881719376  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:20 1782881720386  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:21 1782881721375  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:22 1782881722378  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:23 1782881723382  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:24 1782881724264  D747.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:24 1782881724383  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:25 1782881725397  FFDB         number_people  -      -      -     {"event_type": 3, "event_since": 1782881725397, "event_status": "start", "number_people": 2}
12:55:25 1782881725594  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:25 1782881725594  FFDB.1       track          -260   100    56    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -260, "position_y": 100, "position_z": 56, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:26 1782881726389  FFDB.1       track          -250   90     69    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -250, "position_y": 90, "position_z": 69, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:26 1782881726389  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:27 1782881727294  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:27 1782881727294  FFDB.1       track          -270   100    57    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -270, "position_y": 100, "position_z": 57, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:28 1782881728295  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:28 1782881728295  FFDB.1       track          -280   110    50    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -280, "position_y": 110, "position_z": 50, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:29 1782881729291  FFDB.1       track          -310   120    40    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -310, "position_y": 120, "position_z": 40, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:29 1782881729291  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:30 1782881730292  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:30 1782881730292  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:31 1782881731294  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:31 1782881731294  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:32 1782881732293  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:32 1782881732293  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:33 1782881733296  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:33 1782881733296  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:34 1782881734296  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:34 1782881734296  FFDB.0       track          -180   0      0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -180, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:35 1782881735298  FFDB.0       track          -200   10     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -200, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:35 1782881735298  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:36 1782881736305  FFDB.0       track          -200   10     0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -200, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:36 1782881736305  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:37 1782881737298  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:37 1782881737298  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:38 1782881738194  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:38 1782881738194  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:39 1782881739195  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:39 1782881739195  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:40 1782881740198  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:40 1782881740198  FFDB.1       track          -330   120    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -330, "position_y": 120, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:41 1782881741200  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:41 1782881741200  FFDB.1       track          -260   90     38    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -260, "position_y": 90, "position_z": 38, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:42 1782881742201  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:42 1782881742201  FFDB.1       track          -260   110    57    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -260, "position_y": 110, "position_z": 57, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:43 1782881743206  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:43 1782881743206  FFDB.1       track          -260   120    72    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -260, "position_y": 120, "position_z": 72, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:44 1782881744200  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:44 1782881744200  FFDB.1       track          -280   150    54    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -280, "position_y": 150, "position_z": 54, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:45 1782881745151  FFDB.1       track          -310   190    71    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -310, "position_y": 190, "position_z": 71, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:45 1782881745151  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:46 1782881746145  FFDB.1       track          -300   220    43    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -300, "position_y": 220, "position_z": 43, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:46 1782881746145  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:46 1782881746995  B267.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:47 1782881747146  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:47 1782881747146  FFDB.1       track          -310   270    60    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -310, "position_y": 270, "position_z": 60, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:48 1782881748147  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:48 1782881748147  FFDB.1       track          -360   270    67    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -360, "position_y": 270, "position_z": 67, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:49 1782881749148  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:49 1782881749148  FFDB.1       track          -370   290    67    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -370, "position_y": 290, "position_z": 67, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:50 1782881750149  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:50 1782881750149  FFDB.1       track          -350   290    70    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": -350, "position_y": 290, "position_z": 70, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:51 1782881751150  FFDB.1       track          -370   300    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -370, "position_y": 300, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:51 1782881751150  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:52 1782881752160  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:52 1782881752160  FFDB.1       track          -370   300    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -370, "position_y": 300, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:53 1782881753158  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:53 1782881753158  FFDB.1       track          -370   290    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -370, "position_y": 290, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:54 1782881754162  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:54 1782881754162  FFDB.1       track          -370   290    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -370, "position_y": 290, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:55 1782881755157  FFDB.1       track          -370   290    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -370, "position_y": 290, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:55 1782881755157  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:55 1782881755952  D747.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881755952, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 0}
12:55:55 1782881755952  D747.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:55:56 1782881756056  FFDB.1       track          -360   300    102   {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -360, "position_y": 300, "position_z": 102, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:56 1782881756056  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:56 1782881756212  D747.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
12:55:57 1782881757054  FFDB.1       track          -340   290    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -340, "position_y": 290, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:57 1782881757054  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:58 1782881758065  FFDB.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
12:55:58 1782881758065  FFDB.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782881758065, "event_status": "instant", "lie_duration": 0, "walk_distance": 1, "walk_duration": 5, "stand_duration": 23, "multi_person_duration": 32}
12:55:58 1782881758278  FFDB.0       track          -170   -10    0     {"pose": 4, "area_id": 6, "track_id": 0, "position_x": -170, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
12:55:58 1782881758278  FFDB.1       track          -340   290    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": -340, "position_y": 290, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
```

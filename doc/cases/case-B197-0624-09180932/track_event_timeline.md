# case-B197-0624-09180932 — 每 tick belief 时间线 (room fd00:0:3:111:1:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
09:18:00 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
09:18:01 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
09:18:02 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   1     0.00  0.02  0.70  0.00  0.18  0.02
09:18:03 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   2     0.00  0.02  0.79  0.00  0.07  0.02
09:18:04 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   3     0.00  0.02  0.83  0.00  0.03  0.02
09:18:05 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   4     0.00  0.02  0.84  0.00  0.02  0.02
09:18:06 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   5     0.00  0.02  0.85  0.00  0.01  0.02
09:18:07 B197.0   B19702835843  stand   89   NoReport stand              room -    OpenFloor  1   6     0.00  0.01  0.86  0.00  0.01  0.02
09:18:08 B197.0   B19702835843  walk    78   NoReport walk               room -    OpenFloor  1   0     0.00  0.02  0.75  0.00  0.02  0.04
09:18:09 B197.0   B19702835843  walk    100  NoReport walk               room -    OpenFloor  1   0     0.00  0.02  0.70  0.01  0.03  0.04
09:18:10 B197.0   B19702835843  walk    71   NoReport walk               room -    OpenFloor  1   0     0.00  0.02  0.67  0.01  0.04  0.04
09:18:11 B197.0   B19702835843  walk    83   NoReport walk               room -    OpenFloor  1   0     0.00  0.02  0.66  0.01  0.05  0.03
09:18:12 B197.0   B19702835843  walk    83   NoReport walk               room -    OpenFloor  1   0     0.00  0.02  0.57  0.01  0.05  0.03
09:18:13 B197.0   B19702835843  walk    64   NoReport walk               room -    OpenFloor  1   0     0.00  0.01  0.54  0.01  0.04  0.03
09:18:14 B197.0   B19702835843  walk    76   NoReport walk               room -    OpenFloor  1   0     0.00  0.01  0.50  0.01  0.04  0.03
09:18:15 B197.0   B19702835843  walk    96   NoReport walk               room -    OpenFloor  1   0     0.00  0.01  0.48  0.02  0.04  0.03
09:18:16 B197.0   B19702835843  walk    100  NoReport walk               room -    OpenFloor  1   0     0.00  0.01  0.45  0.02  0.04  0.03
09:18:17 B197.0   B19702835843  walk    87   NoReport walk               room -    OpenFloor  1   0     0.00  0.01  0.43  0.02  0.03  0.03
09:18:18 B197.0   B19702835843  walk    68   NoReport walk               room -    OpenFloor  1   0     0.00  0.01  0.42  0.02  0.03  0.03
09:18:19 B197.0   B19702835843  walk    95   NoReport walk               room -    Sit        1   0     0.00  0.01  0.40  0.02  0.03  0.03
09:18:20 B197.0   B19702835843  walk    90   NoReport walk               room -    Sit        1   0     0.00  0.01  0.39  0.02  0.03  0.02
09:18:21 B197.0   B19702835843  walk    107  NoReport walk               room -    Sit        1   0     0.00  0.01  0.38  0.02  0.03  0.02
09:18:22 B197.0   B19702835843  walk    82   NoReport walk               room -    OpenFloor  1   0     0.00  0.01  0.48  0.02  0.04  0.03
09:18:23 B197.0   B19702835843  walk    80   NoReport walk               room -    OpenFloor  1   0     0.00  0.03  0.54  0.02  0.04  0.03
09:18:24 B197.0   B19702835843  walk    98   NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.57  0.02  0.05  0.03
09:18:25 B197.0   B19702835843  walk    64   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.59  0.02  0.05  0.03
09:18:26 B197.0   B19702835843  walk    70   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.60  0.02  0.05  0.03
09:18:27 B197.0   B19702835843  walk    52   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
09:18:28 B197.0   B19702835843  walk    76   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
09:18:29 B197.0   B19702835843  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
09:18:30 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
09:18:31 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:32 B197.0   B19702835843  stand   90   NoReport stand              room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
09:18:33 B197.0   B19702835843  stand   68   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:34 B197.0   B19702835843  stand   52   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:35 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:36 B197.0   B19702835843  stand   65   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:37 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:38 B197.0   B19702835843  stand   77   NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:39 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:40 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:41 B197.E   -             -       0    NoReport np=2               room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:18:41 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.01  0.03  0.76  0.01  0.03  0.02
09:18:41 B197.1   B19711841555  stand   54   NoReport stand              trk  0.50 OpenFloor  2   0     0.00  0.03  0.15  0.00  0.79  0.04
09:18:42 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.03  0.82  0.01  0.02  0.02
09:18:42 B197.1   B19711841555  stand   54   NoReport stand              trk  0.51 OpenFloor  2   0     0.00  0.03  0.40  0.00  0.53  0.01
09:18:43 B197.1   B19711841555  stand   19   NoReport stand              trk  0.78 OpenFloor  2   0     0.00  0.03  0.63  0.00  0.26  0.01
09:18:43 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
09:18:44 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
09:18:44 B197.1   B19711841555  stand   0    NoReport stand              trk  0.92 OpenFloor  2   0     0.00  0.02  0.76  0.00  0.11  0.02
09:18:45 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:45 B197.1   B19711841555  stand   37   NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.82  0.00  0.04  0.02
09:18:46 B197.1   B19711841555  stand   85   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.02  0.02
09:18:46 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:47 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:47 B197.1   B19711841555  stand   66   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:48 B197.1   B19711841555  stand   88   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:48 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:49 B197.1   B19711841555  stand   71   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:49 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:50 B197.1   B19711841555  stand   67   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:50 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:51 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:51 B197.1   B19711841555  stand   74   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:52 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:52 B197.1   B19711841555  stand   59   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:53 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:53 B197.1   B19711841555  stand   0    NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:54 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:54 B197.1   B19711841555  stand   0    NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:55 B197.1   B19711841555  stand   0    NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:55 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:56 B197.1   B19711841555  stand   39   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:56 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:57 B197.1   B19711841555  stand   0    NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:57 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:58 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:58 B197.1   B19711841555  stand   78   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:59 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:18:59 B197.1   B19711841555  stand   99   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:00 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:00 B197.1   B19711841555  stand   106  NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:01 B197.1   B19711841555  stand   0    NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:01 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:02 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
09:19:02 B197.1   B19711841555  stand   78   NoReport stand              trk  0.53 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
09:19:03 B197.1   B19711841555  stand   98   NoReport stand              trk  0.53 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
09:19:03 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
09:19:04 B197.1   B19711841555  stand   0    NoReport stand              trk  0.53 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
09:19:04 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
09:19:05 B197.1   B19711841555  stand   67   NoReport stand              trk  0.53 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
09:19:05 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
09:19:06 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
09:19:06 B197.1   B19711841555  stand   95   NoReport stand              trk  0.53 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
09:19:07 B197.1   B19711841555  stand   82   NoReport stand              trk  0.53 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
09:19:07 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
09:19:08 B197.1   B19711841555  stand   63   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:08 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:09 B197.1   B19711841555  stand   36   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:09 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:10 B197.1   B19711841555  stand   81   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:10 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:11 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:11 B197.1   B19711841555  stand   91   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:12 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:12 B197.1   B19711841555  stand   83   NoReport stand              trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:13 B197.1   B19711841555  walk    59   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:13 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:14 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:14 B197.1   B19711841555  walk    19   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:15 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:15 B197.1   B19711841555  walk    40   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:16 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:16 B197.1   B19711841555  walk    89   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:17 B197.1   B19711841555  walk    71   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:17 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:18 B197.1   B19711841555  walk    74   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:18 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:19 B197.1   B19711841555  walk    92   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:19 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:20 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:20 B197.1   B19711841555  walk    96   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:21 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:21 B197.1   B19711841555  walk    100  NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:22 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:22 B197.1   B19711841555  walk    110  NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:23 B197.1   B19711841555  walk    90   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:23 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:24 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:24 B197.1   B19711841555  walk    101  NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:25 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:25 B197.1   B19711841555  walk    100  NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:26 B197.1   B19711841555  walk    71   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:26 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:27 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:27 B197.1   B19711841555  walk    112  NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:28 B197.1   B19711841555  walk    81   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:28 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:29 B197.1   B19711841555  walk    83   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:29 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:30 B197.1   B19711841555  walk    77   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:30 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:31 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:31 B197.1   B19711841555  walk    44   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:32 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:32 B197.1   B19711841555  walk    99   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:33 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:33 B197.1   B19711841555  walk    90   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:34 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:34 B197.1   B19711841555  walk    106  NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:35 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:35 B197.1   B19711841555  walk    107  NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:36 B197.1   B19711841555  walk    80   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:36 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:37 B197.1   B19711841555  walk    57   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:37 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:38 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:38 B197.1   B19711841555  walk    48   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:38 B197.E   -             -       0    NoReport Walking(rdr)       room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:38 B197.1   B19711841555  sitgnd  74   NoReport sitgnd             trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:38 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:39 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:39 B197.1   B19711841555  walk    94   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:40 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:40 B197.1   B19711841555  walk    78   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:41 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:41 B197.1   B19711841555  walk    73   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:42 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:42 B197.1   B19711841555  walk    84   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:43 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:43 B197.1   B19711841555  walk    73   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:43 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:43 B197.1   B19711841555  walk    98   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:44 B197.1   B19711841555  walk    98   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:44 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:45 B197.E   -             -       0    NoReport np=3               room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:46 B197.1   B19711841555  walk    81   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:46 B197.2   B19721946030  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:46 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:46 B197.2   B19721946030  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:46 B197.1   B19711841555  walk    77   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:46 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:47 B197.2   B19721946030  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:47 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:47 B197.1   B19711841555  walk    65   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:48 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:48 B197.1   B19711841555  walk    75   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:48 B197.2   B19721946030  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:49 B197.2   B19721946030  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:49 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:49 B197.1   B19711841555  walk    80   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:50 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:50 B197.2   B19721946030  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:50 B197.1   B19711841555  walk    93   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:51 B197.1   B19711841555  walk    93   NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:51 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:51 B197.2   B19721946030  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:52 B197.2   B19721946030  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:52 B197.1   B19711841555  walk    102  NoReport walk               trk  0.53 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:52 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
09:19:53 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:53 B197.1   B19711841555  walk    0    NoReport walk               trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:53 B197.2   B19721946030  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:54 B197.2   B19721946030  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:54 B197.1   B19711841555  walk    0    NoReport walk               trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:54 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:55 B197.2   B19721946030  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:55 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:55 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:56 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:56 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:56 B197.2   B19721946030  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:57 B197.2   B19721946030  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:57 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:57 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:58 B197.2   B19721946030  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:58 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:58 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:59 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:59 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:19:59 B197.2   B19721946030  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:00 B197.0   B19702835843  stand   0    NoReport stand              room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:00 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:00 B197.2   B19721946030  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:01 B197.E   -             -       0    NoReport np=2               room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:01 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:01 B197.2   B19721946030  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:02 B197.E   -             -       0    NoReport np=1               room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:02 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:03 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:04 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:05 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:06 B197.1   B19711841555  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:07 B197.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  3   9     0.00  0.03  0.78  0.00  0.01  0.04
09:20:07 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  1   14    0.03  0.10  0.21  0.07  0.16  0.03
09:20:08 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
09:20:09 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.19  0.09  0.18  0.02
09:20:29 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:20:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:21:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
09:21:01 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.17  0.10  0.19  0.02
09:21:33 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:21:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:04 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
09:22:36 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:22:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:23:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:23:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:23:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:23:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:23:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:23:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:23:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:23:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.11  0.20  0.02
09:23:08 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:40 B197.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:23:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
09:24:12 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:43 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:24:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
09:25:15 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:47 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:25:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
09:26:19 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:50 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:26:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
09:27:22 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:42 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:42 B197.0   B19702835843  walk    92   NoReport walk               room -    Empty      1   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:43 B197.0   B19702835843  walk    110  NoReport walk               room -    Empty      1   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:44 B197.0   B19702835843  walk    90   NoReport walk               room -    Empty      1   0     0.09  0.10  0.15  0.13  0.19  0.02
09:27:45 B197.0   B19702835843  walk    62   NoReport walk               room -    Empty      1   0     0.10  0.10  0.15  0.13  0.19  0.02
09:27:46 B197.0   B19702835843  walk    71   NoReport walk               room -    Empty      1   0     0.10  0.10  0.15  0.13  0.19  0.02
09:27:47 B197.0   B19702835843  walk    118  NoReport walk               room -    Empty      1   0     0.10  0.10  0.15  0.13  0.19  0.02
09:27:48 B197.0   B19702835843  walk    89   NoReport walk               room -    Empty      1   0     0.10  0.10  0.15  0.13  0.19  0.02
09:27:49 B197.0   B19702835843  walk    0    NoReport walk               room -    Empty      1   0     0.10  0.10  0.15  0.13  0.19  0.02
09:27:50 B197.0   B19702835843  walk    0    NoReport walk               room -    Empty      1   0     0.11  0.10  0.15  0.13  0.19  0.02
09:27:51 B197.0   B19702835843  walk    0    NoReport walk               room -    Empty      1   0     0.11  0.10  0.15  0.13  0.19  0.02
09:27:52 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.11  0.10  0.15  0.13  0.19  0.02
09:27:53 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.11  0.10  0.15  0.13  0.19  0.02
09:27:54 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.11  0.10  0.15  0.13  0.19  0.02
09:27:55 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.11  0.10  0.15  0.13  0.19  0.02
09:27:56 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.12  0.10  0.15  0.13  0.19  0.02
09:27:57 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.12  0.10  0.15  0.13  0.19  0.02
09:27:58 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.12  0.10  0.15  0.13  0.19  0.02
09:27:59 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.12  0.10  0.15  0.13  0.19  0.02
09:28:00 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.12  0.10  0.15  0.13  0.19  0.02
09:28:01 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.13  0.10  0.15  0.13  0.19  0.02
09:28:01 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:02 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:03 B197.0   B19702835843  stand   0    NoReport stand              room -    Empty      1   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:03 B197.E   -             -       0    NoReport ExitRoom(rdr)      room -    Empty      1   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:03 B197.E   -             -       0    NoReport np=0  ★0           room -    Empty      1   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:04 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      1   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:04 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:05 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.13  0.10  0.14  0.12  0.19  0.02
09:28:26 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:35 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:35 B197.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.14  0.10  0.14  0.12  0.18  0.02
09:28:35 B197.0   B19702835843  stand   108  NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
09:28:36 B197.0   B19702835843  stand   95   NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.52  0.00  0.40  0.01
09:28:37 B197.0   B19702835843  walk    95   NoReport walk               trk  1.00 Empty      1   0     0.00  0.04  0.54  0.00  0.28  0.03
09:28:38 B197.0   B19702835843  walk    79   NoReport walk               trk  1.00 Empty      1   0     0.00  0.04  0.57  0.01  0.20  0.03
09:28:39 B197.0   B19702835843  walk    83   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.58  0.01  0.15  0.03
09:28:40 B197.0   B19702835843  walk    92   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.59  0.01  0.11  0.03
09:28:41 B197.0   B19702835843  walk    91   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.60  0.02  0.09  0.03
09:28:42 B197.0   B19702835843  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.08  0.03
09:28:43 B197.0   B19702835843  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.07  0.03
09:28:44 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.06  0.03
09:28:45 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.06  0.03
09:28:46 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.06  0.03
09:28:47 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.06  0.03
09:28:48 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:28:49 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:28:50 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:28:51 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:28:52 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:28:53 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:28:54 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:28:55 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:28:56 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:28:57 B197.E   -             -       0    NoReport np=0  ★0           room -    Empty      1   0     0.17  0.10  0.14  0.12  0.18  0.02
09:28:57 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      1   0     0.17  0.10  0.14  0.12  0.18  0.02
09:28:58 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:28:59 B197.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:15 B197.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.17  0.10  0.14  0.12  0.18  0.02
09:29:15 B197.0   B19702835843  stand   103  NoReport stand              trk  1.00 Empty      1   39    0.01  0.08  0.39  0.05  0.12  0.02
09:29:16 B197.0   B19702835843  stand   89   NoReport stand              trk  1.00 Empty      1   0     0.01  0.07  0.47  0.05  0.10  0.03
09:29:17 B197.0   B19702835843  walk    87   NoReport walk               trk  1.00 Empty      1   0     0.00  0.06  0.52  0.04  0.09  0.03
09:29:18 B197.0   B19702835843  walk    89   NoReport walk               trk  1.00 Empty      1   0     0.00  0.06  0.56  0.03  0.07  0.03
09:29:19 B197.0   B19702835843  walk    89   NoReport walk               trk  1.00 Empty      1   0     0.00  0.06  0.58  0.03  0.07  0.03
09:29:20 B197.0   B19702835843  walk    74   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.06  0.59  0.03  0.06  0.03
09:29:21 B197.0   B19702835843  walk    42   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.06  0.60  0.02  0.06  0.03
09:29:22 B197.0   B19702835843  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.06  0.61  0.02  0.06  0.03
09:29:23 B197.0   B19702835843  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.06  0.03
09:29:24 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
09:29:25 B197.0   B19702835843  stand   34   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:26 B197.0   B19702835843  stand   33   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:27 B197.0   B19702835843  stand   42   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:28 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:29 B197.0   B19702835843  stand   41   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:30 B197.0   B19702835843  stand   31   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:31 B197.0   B19702835843  stand   46   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:32 B197.0   B19702835843  stand   41   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:33 B197.0   B19702835843  stand   39   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:34 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:35 B197.0   B19702835843  stand   44   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:36 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:37 B197.0   B19702835843  stand   63   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:38 B197.0   B19702835843  stand   44   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:39 B197.0   B19702835843  stand   36   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:40 B197.0   B19702835843  stand   36   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:41 B197.0   B19702835843  stand   32   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:42 B197.0   B19702835843  stand   35   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:43 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:44 B197.0   B19702835843  stand   26   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:45 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:46 B197.0   B19702835843  stand   40   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:47 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:48 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:49 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:50 B197.0   B19702835843  stand   24   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:51 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:52 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   28    0.01  0.05  0.61  0.02  0.05  0.03
09:29:53 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   29    0.01  0.05  0.61  0.02  0.05  0.03
09:29:54 B197.0   B19702835843  stand   24   NoReport stand              trk  1.00 Fallen     1   30    0.01  0.05  0.61  0.02  0.05  0.03
09:29:55 B197.0   B19702835843  stand   27   NoReport stand              trk  1.00 Fallen     1   31    0.01  0.05  0.61  0.02  0.05  0.03
09:29:56 B197.0   B19702835843  stand   23   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:57 B197.0   B19702835843  stand   20   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:58 B197.0   B19702835843  stand   33   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:29:59 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:00 B197.0   B19702835843  stand   16   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:01 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:02 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:03 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:04 B197.0   B19702835843  stand   24   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:05 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:06 B197.0   B19702835843  stand   21   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:06 B197.0   B19702835843  stand   21   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:07 -.-      -             -       -    NoReport (no frame, held)   room -    Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
09:30:08 B197.0   B19702835843  stand   32   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:08 B197.0   B19702835843  stand   23   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:09 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:10 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:11 -.-      -             -       -    NoReport (no frame, held)   room -    Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
09:30:12 B197.0   B19702835843  stand   16   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:12 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:13 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:14 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:15 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:16 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:17 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:18 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:19 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:20 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:30:21 B197.0   B19702835843  stand   27   NoReport stand              trk  1.00 Fallen     1   27    0.01  0.05  0.61  0.02  0.05  0.03
09:30:22 B197.0   B19702835843  stand   32   NoReport stand              trk  1.00 Fallen     1   28    0.01  0.05  0.61  0.02  0.05  0.03
09:30:23 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   29    0.01  0.05  0.61  0.02  0.05  0.03
09:30:24 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   30    0.01  0.05  0.61  0.02  0.05  0.03
09:30:25 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   31    0.01  0.05  0.61  0.02  0.05  0.03
09:30:26 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   32    0.01  0.05  0.61  0.02  0.05  0.03
09:30:27 B197.0   B19702835843  stand   21   NoReport stand              trk  1.00 Fallen     1   33    0.01  0.05  0.61  0.02  0.05  0.03
09:30:28 B197.0   B19702835843  stand   11   NoReport stand              trk  1.00 Fallen     1   34    0.01  0.05  0.61  0.02  0.05  0.03
09:30:29 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   35    0.01  0.05  0.61  0.02  0.05  0.03
09:30:30 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   36    0.01  0.05  0.61  0.02  0.05  0.03
09:30:31 B197.0   B19702835843  stand   20   NoReport stand              trk  1.00 Fallen     1   37    0.01  0.05  0.61  0.02  0.05  0.03
09:30:32 B197.0   B19702835843  stand   21   NoReport stand              trk  1.00 Fallen     1   38    0.01  0.05  0.61  0.02  0.05  0.03
09:30:33 B197.0   B19702835843  stand   34   NoReport stand              trk  1.00 Fallen     1   39    0.01  0.05  0.61  0.02  0.05  0.03
09:30:34 B197.0   B19702835843  stand   15   NoReport stand              trk  1.00 Fallen     1   40    0.01  0.05  0.61  0.02  0.05  0.03
09:30:35 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   41    0.01  0.05  0.61  0.02  0.05  0.03
09:30:36 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   42    0.01  0.05  0.61  0.02  0.05  0.03
09:30:37 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   43    0.01  0.05  0.61  0.02  0.05  0.03
09:30:38 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   44    0.01  0.05  0.61  0.02  0.05  0.03
09:30:39 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   45    0.01  0.05  0.61  0.02  0.05  0.03
09:30:40 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   46    0.01  0.05  0.61  0.02  0.05  0.03
09:30:41 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   47    0.01  0.05  0.61  0.02  0.05  0.03
09:30:42 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   48    0.01  0.05  0.61  0.02  0.05  0.03
09:30:43 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   49    0.01  0.05  0.61  0.02  0.05  0.03
09:30:44 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   50    0.01  0.05  0.61  0.02  0.05  0.03
09:30:45 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   51    0.01  0.05  0.61  0.02  0.05  0.03
09:30:46 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   52    0.01  0.05  0.61  0.02  0.05  0.03
09:30:47 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   53    0.01  0.05  0.61  0.02  0.05  0.03
09:30:48 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   54    0.01  0.05  0.61  0.02  0.05  0.03
09:30:49 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   55    0.01  0.05  0.61  0.02  0.05  0.03
09:30:50 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   56    0.01  0.05  0.61  0.02  0.05  0.03
09:30:51 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   57    0.01  0.05  0.61  0.02  0.05  0.03
09:30:52 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   58    0.01  0.05  0.61  0.02  0.05  0.03
09:30:53 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   59    0.01  0.05  0.61  0.02  0.05  0.03
09:30:54 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   60    0.01  0.05  0.61  0.02  0.05  0.03
09:30:55 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   61    0.01  0.05  0.61  0.02  0.05  0.03
09:30:56 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   62    0.01  0.05  0.61  0.02  0.05  0.03
09:30:57 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   63    0.01  0.05  0.61  0.02  0.05  0.03
09:30:58 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   64    0.01  0.05  0.61  0.02  0.05  0.03
09:30:59 B197.0   B19702835843  stand   27   NoReport stand              trk  1.00 Fallen     1   65    0.01  0.05  0.61  0.02  0.05  0.03
09:31:00 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   66    0.01  0.05  0.61  0.02  0.05  0.03
09:31:01 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   67    0.01  0.05  0.61  0.02  0.05  0.03
09:31:02 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   68    0.01  0.05  0.61  0.02  0.05  0.03
09:31:03 B197.0   B19702835843  stand   36   NoReport stand              trk  1.00 Fallen     1   69    0.01  0.05  0.61  0.02  0.05  0.03
09:31:04 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   70    0.01  0.05  0.61  0.02  0.05  0.03
09:31:05 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   71    0.01  0.05  0.61  0.02  0.05  0.03
09:31:06 B197.0   B19702835843  stand   22   NoReport stand              trk  1.00 Fallen     1   72    0.01  0.05  0.61  0.02  0.05  0.03
09:31:07 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   73    0.01  0.05  0.61  0.02  0.05  0.03
09:31:08 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   74    0.01  0.05  0.61  0.02  0.05  0.03
09:31:09 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   75    0.01  0.05  0.61  0.02  0.05  0.03
09:31:10 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   76    0.01  0.05  0.61  0.02  0.05  0.03
09:31:11 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   77    0.01  0.05  0.61  0.02  0.05  0.03
09:31:12 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   78    0.01  0.05  0.61  0.02  0.05  0.03
09:31:13 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   79    0.01  0.05  0.61  0.02  0.05  0.03
09:31:14 B197.0   B19702835843  stand   23   NoReport stand              trk  1.00 Fallen     1   80    0.01  0.05  0.61  0.02  0.05  0.03
09:31:15 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   81    0.01  0.05  0.61  0.02  0.05  0.03
09:31:16 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   82    0.01  0.05  0.61  0.02  0.05  0.03
09:31:17 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   83    0.01  0.05  0.61  0.02  0.05  0.03
09:31:18 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   84    0.01  0.05  0.61  0.02  0.05  0.03
09:31:19 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   85    0.01  0.05  0.61  0.02  0.05  0.03
09:31:20 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   86    0.01  0.05  0.61  0.02  0.05  0.03
09:31:21 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   87    0.01  0.05  0.61  0.02  0.05  0.03
09:31:22 B197.0   B19702835843  stand   25   NoReport stand              trk  1.00 Fallen     1   88    0.01  0.05  0.61  0.02  0.05  0.03
09:31:23 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   89    0.01  0.05  0.61  0.02  0.05  0.03
09:31:24 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   90    0.01  0.05  0.61  0.02  0.05  0.03
09:31:25 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   91    0.01  0.05  0.61  0.02  0.05  0.03
09:31:26 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   92    0.01  0.05  0.61  0.02  0.05  0.03
09:31:27 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   93    0.01  0.05  0.61  0.02  0.05  0.03
09:31:28 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   94    0.01  0.05  0.61  0.02  0.05  0.03
09:31:29 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   95    0.01  0.05  0.61  0.02  0.05  0.03
09:31:30 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   96    0.01  0.05  0.61  0.02  0.05  0.03
09:31:31 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:32 B197.0   B19702835843  stand   15   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:33 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:34 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:35 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:36 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:37 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:38 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:39 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:40 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   28    0.01  0.05  0.61  0.02  0.05  0.03
09:31:41 B197.0   B19702835843  stand   31   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:42 B197.0   B19702835843  stand   25   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:43 B197.0   B19702835843  stand   9    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:44 B197.0   B19702835843  stand   14   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:45 B197.0   B19702835843  stand   9    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:46 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   28    0.01  0.05  0.61  0.02  0.05  0.03
09:31:47 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   29    0.01  0.05  0.61  0.02  0.05  0.03
09:31:48 B197.0   B19702835843  stand   17   NoReport stand              trk  1.00 Fallen     1   30    0.01  0.05  0.61  0.02  0.05  0.03
09:31:49 B197.0   B19702835843  stand   17   NoReport stand              trk  1.00 Fallen     1   31    0.01  0.05  0.61  0.02  0.05  0.03
09:31:50 B197.0   B19702835843  stand   22   NoReport stand              trk  1.00 Fallen     1   32    0.01  0.05  0.61  0.02  0.05  0.03
09:31:51 B197.0   B19702835843  stand   20   NoReport stand              trk  1.00 Fallen     1   33    0.01  0.05  0.61  0.02  0.05  0.03
09:31:52 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:53 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:54 B197.0   B19702835843  stand   19   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:55 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:56 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:57 B197.0   B19702835843  stand   30   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:58 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:31:59 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:00 B197.0   B19702835843  stand   12   NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:01 B197.0   B19702835843  stand   14   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:02 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:03 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:04 B197.0   B19702835843  stand   17   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:05 B197.0   B19702835843  stand   5    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:06 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:06 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:07 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:08 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:09 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:10 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:11 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:12 B197.0   B19702835843  stand   24   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:13 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:14 B197.0   B19702835843  stand   26   NoReport stand              trk  1.00 Empty      1   27    0.01  0.05  0.61  0.02  0.05  0.03
09:32:15 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   28    0.01  0.05  0.61  0.02  0.05  0.03
09:32:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      1   28    0.05  0.10  0.15  0.12  0.27  0.02
09:32:17 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   29    0.01  0.05  0.61  0.02  0.05  0.03
09:32:17 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   30    0.01  0.05  0.61  0.02  0.05  0.03
09:32:18 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   31    0.01  0.05  0.61  0.02  0.05  0.03
09:32:19 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   32    0.01  0.05  0.61  0.02  0.05  0.03
09:32:20 B197.0   B19702835843  stand   46   NoReport stand              trk  1.00 Empty      1   33    0.01  0.05  0.61  0.02  0.05  0.03
09:32:21 B197.0   B19702835843  stand   22   NoReport stand              trk  1.00 Empty      1   34    0.01  0.05  0.61  0.02  0.05  0.03
09:32:22 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   35    0.01  0.05  0.61  0.02  0.05  0.03
09:32:23 B197.0   B19702835843  stand   26   NoReport stand              trk  1.00 Empty      1   36    0.01  0.05  0.61  0.02  0.05  0.03
09:32:24 B197.0   B19702835843  stand   12   NoReport stand              trk  1.00 Empty      1   37    0.01  0.05  0.61  0.02  0.05  0.03
09:32:25 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   38    0.01  0.05  0.61  0.02  0.05  0.03
09:32:26 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   39    0.01  0.05  0.61  0.02  0.05  0.03
09:32:27 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:28 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:29 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:30 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:31 B197.0   B19702835843  stand   21   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:32 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:33 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:34 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:35 B197.0   B19702835843  stand   7    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:36 B197.0   B19702835843  stand   47   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:37 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:38 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:39 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:40 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:41 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   28    0.01  0.05  0.61  0.02  0.05  0.03
09:32:42 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   29    0.01  0.05  0.61  0.02  0.05  0.03
09:32:43 B197.0   B19702835843  stand   24   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:44 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:45 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:46 B197.0   B19702835843  stand   26   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:47 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:48 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:49 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:50 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:51 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:52 B197.0   B19702835843  stand   18   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:53 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:54 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:55 B197.0   B19702835843  stand   31   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:56 B197.0   B19702835843  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:57 B197.0   B19702835843  stand   25   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
09:32:58 B197.0   B19702835843  stand   24   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
09:18:00.851 B197.0    stand  255  -480   40     0     80        
09:18:01.841 B197.0    stand  255  -480   40     0     80   0    
09:18:02.838 B197.0    stand  255  -480   40     0     80   0    
09:18:03.837 B197.0    stand  255  -480   40     0     80   0    
09:18:04.840 B197.0    stand  255  -480   40     0     80   0    
09:18:05.841 B197.0    stand  255  -480   40     0     80   0    
09:18:06.847 B197.0    stand  255  -480   40     0     80   0    
09:18:07.848 B197.0    stand  255  -460   60     89    80   28   
09:18:08.737 B197.0    walk   255  -440   150    78    80   92   
09:18:09.739 B197.0    walk   255  -410   180    100   80   42   
09:18:10.742 B197.0    walk   255  -350   170    71    80   60   
09:18:11.740 B197.0    walk   255  -380   170    83    80   30   
09:18:12.796 B197.0    walk   255  -390   150    83    80   22   
09:18:13.742 B197.0    walk   255  -390   140    64    80   10   
09:18:14.742 B197.0    walk   255  -400   140    76    80   10   
09:18:15.742 B197.0    walk   255  -410   140    96    80   10   
09:18:16.744 B197.0    walk   255  -390   120    100   80   28   
09:18:17.750 B197.0    walk   255  -380   100    87    80   22   
09:18:18.745 B197.0    walk   255  -380   110    68    80   10   
09:18:19.748 B197.0    walk   255  -390   130    95    80   22   
09:18:20.644 B197.0    walk   255  -390   120    90    80   10   
09:18:21.644 B197.0    walk   255  -400   120    107   80   10   
09:18:22.644 B197.0    walk   255  -420   200    82    80   82   
09:18:23.651 B197.0    walk   255  -470   270    80    80   86   
09:18:24.645 B197.0    walk   255  -440   310    98    80   50   
09:18:25.649 B197.0    walk   255  -340   310    64    80   100  
09:18:26.688 B197.0    walk   255  -240   300    70    80   100  
09:18:27.586 B197.0    walk   255  -210   320    52    80   36   
09:18:28.594 B197.0    walk   255  -190   350    76    80   36   
09:18:29.598 B197.0    walk   255  -190   360    0     80   10   
09:18:30.591 B197.0    stand  255  -190   350    0     80   10   
09:18:31.599 B197.0    stand  255  -190   350    0     80   0    
09:18:32.599 B197.0    stand  255  -200   350    90    80   10   
09:18:33.595 B197.0    stand  255  -190   360    68    80   14   
09:18:34.618 B197.0    stand  255  -210   380    52    80   28   
09:18:35.607 B197.0    stand  255  -200   380    0     80   10   
09:18:36.598 B197.0    stand  255  -200   390    65    80   10   
09:18:37.603 B197.0    stand  255  -190   370    0     80   22   
09:18:38.498 B197.0    stand  255  -210   410    77    80   44   
09:18:39.497 B197.0    stand  255  -230   490    0     80   82   
09:18:40.500 B197.0    stand  255  -230   490    0     80   0    
09:18:41.555 B197.0    stand  255  -220   490    0     80   10   
09:18:41.555 B197.1    stand  255  -220   340    54    80   150  
09:18:42.515 B197.0    stand  255  -220   490    0     80   150  
09:18:42.515 B197.1    stand  255  -260   330    54    80   164  
09:18:43.521 B197.1    stand  255  -290   330    19    80   30   
09:18:43.521 B197.0    stand  255  -220   480    0     80   165  
09:18:44.518 B197.0    stand  255  -220   480    0     80   0    
09:18:44.518 B197.1    stand  255  -300   330    0     80   170  
09:18:45.521 B197.0    stand  255  -220   480    0     80   170  
09:18:45.521 B197.1    stand  255  -270   320    37    80   167  
09:18:46.527 B197.1    stand  255  -250   320    85    80   20   
09:18:46.527 B197.0    stand  255  -220   480    0     80   162  
09:18:47.522 B197.0    stand  255  -220   480    0     80   0    
09:18:47.522 B197.1    stand  255  -210   320    66    80   160  
09:18:48.433 B197.1    stand  255  -200   340    88    80   22   
09:18:48.433 B197.0    stand  255  -210   480    0     80   140  
09:18:49.420 B197.1    stand  255  -190   360    71    80   121  
09:18:49.420 B197.0    stand  255  -160   480    0     80   123  
09:18:50.424 B197.1    stand  255  -190   360    67    80   123  
09:18:50.424 B197.0    stand  255  -180   480    0     80   120  
09:18:51.428 B197.0    stand  255  -160   480    0     80   20   
09:18:51.428 B197.1    stand  255  -170   360    74    80   120  
09:18:52.424 B197.0    stand  255  -180   480    0     80   120  
09:18:52.424 B197.1    stand  255  -180   360    59    80   120  
09:18:53.425 B197.0    stand  255  -200   470    0     80   111  
09:18:53.425 B197.1    stand  255  -180   360    0     80   111  
09:18:54.425 B197.0    stand  255  -200   470    0     80   111  
09:18:54.425 B197.1    stand  255  -170   380    0     80   94   
09:18:55.432 B197.1    stand  255  -170   380    0     80   0    
09:18:55.432 B197.0    stand  255  -190   470    0     80   92   
09:18:56.427 B197.1    stand  255  -170   380    39    80   92   
09:18:56.427 B197.0    stand  255  -190   470    0     80   92   
09:18:57.430 B197.1    stand  255  -170   350    0     80   121  
09:18:57.430 B197.0    stand  255  -160   480    0     80   130  
09:18:58.428 B197.0    stand  255  -170   490    0     80   14   
09:18:58.428 B197.1    stand  255  -200   340    78    80   152  
09:18:59.329 B197.0    stand  255  -170   490    0     80   152  
09:18:59.329 B197.1    stand  255  -190   350    99    80   141  
09:19:00.332 B197.0    stand  255  -170   490    0     80   141  
09:19:00.332 B197.1    stand  255  -180   360    106   80   130  
09:19:01.332 B197.1    stand  255  -180   350    0     80   10   
09:19:01.332 B197.0    stand  255  -170   490    0     80   140  
09:19:02.333 B197.0    stand  255  -170   490    0     80   0    
09:19:02.333 B197.1    stand  255  -190   350    78    80   141  
09:19:03.335 B197.1    stand  255  -210   350    98    80   20   
09:19:03.335 B197.0    stand  255  -170   490    0     80   145  
09:19:04.340 B197.1    stand  255  -200   350    0     80   143  
09:19:04.340 B197.0    stand  255  -180   490    0     80   141  
09:19:05.337 B197.1    stand  255  -190   350    67    80   140  
09:19:05.337 B197.0    stand  255  -190   490    0     80   140  
09:19:06.336 B197.0    stand  255  -190   480    0     80   10   
09:19:06.336 B197.1    stand  255  -190   350    95    80   130  
09:19:07.337 B197.1    stand  255  -200   340    82    80   14   
09:19:07.337 B197.0    stand  255  -210   470    0     80   130  
09:19:08.338 B197.1    stand  255  -240   340    63    80   133  
09:19:08.338 B197.0    stand  255  -230   490    0     80   150  
09:19:09.353 B197.1    stand  255  -270   360    36    80   136  
09:19:09.353 B197.0    stand  255  -250   470    0     80   111  
09:19:10.345 B197.1    stand  255  -260   340    81    80   130  
09:19:10.345 B197.0    stand  255  -260   480    0     80   140  
09:19:11.241 B197.0    stand  255  -390   520    0     80   136  
09:19:11.241 B197.1    stand  255  -230   330    91    80   248  
09:19:12.237 B197.0    stand  255  -400   480    0     80   226  
09:19:12.237 B197.1    stand  255  -220   290    83    80   261  
09:19:13.289 B197.1    walk   255  -190   260    59    80   42   
09:19:13.289 B197.0    stand  255  -400   460    0     80   290  
09:19:14.241 B197.0    stand  255  -400   480    0     80   20   
09:19:14.241 B197.1    walk   255  -170   250    19    80   325  
09:19:15.205 B197.0    stand  255  -390   480    0     80   318  
09:19:15.205 B197.1    walk   255  -190   250    40    80   304  
09:19:16.206 B197.0    stand  255  -390   450    0     80   282  
09:19:16.206 B197.1    walk   255  -210   270    89    80   254  
09:19:17.222 B197.1    walk   255  -280   300    71    80   76   
09:19:17.222 B197.0    stand  255  -380   470    0     80   197  
09:19:18.207 B197.1    walk   255  -340   340    74    80   136  
09:19:18.207 B197.0    stand  255  -340   490    0     80   150  
09:19:19.210 B197.1    walk   255  -400   370    92    80   134  
09:19:19.210 B197.0    stand  255  -380   480    0     80   111  
09:19:20.212 B197.0    stand  255  -380   500    0     80   20   
09:19:20.212 B197.1    walk   255  -450   400    96    80   122  
09:19:21.213 B197.0    stand  255  -430   500    0     80   101  
09:19:21.213 B197.1    walk   255  -470   390    100   80   117  
09:19:22.217 B197.0    stand  255  -430   500    0     80   117  
09:19:22.217 B197.1    walk   255  -470   350    110   80   155  
09:19:23.213 B197.1    walk   255  -460   310    90    80   41   
09:19:23.213 B197.0    stand  255  -430   500    0     80   192  
09:19:24.213 B197.0    stand  255  -430   500    0     80   0    
09:19:24.213 B197.1    walk   255  -470   280    101   80   223  
09:19:25.216 B197.0    stand  255  -430   500    0     80   223  
09:19:25.216 B197.1    walk   255  -460   210    100   80   291  
09:19:26.114 B197.1    walk   255  -410   120    71    80   102  
09:19:26.114 B197.0    stand  255  -430   500    0     80   380  
09:19:27.118 B197.0    stand  255  -430   500    0     80   0    
09:19:27.118 B197.1    walk   255  -430   60     112   80   440  
09:19:28.117 B197.1    walk   255  -430   90     81    80   30   
09:19:28.117 B197.0    stand  255  -430   500    0     80   410  
09:19:29.124 B197.1    walk   255  -430   170    83    80   330  
09:19:29.124 B197.0    stand  255  -430   500    0     80   330  
09:19:30.119 B197.1    walk   255  -470   210    77    80   292  
09:19:30.119 B197.0    stand  255  -430   500    0     80   292  
09:19:31.132 B197.0    stand  255  -430   500    0     80   0    
09:19:31.132 B197.1    walk   255  -450   250    44    80   250  
09:19:32.134 B197.0    stand  255  -370   550    0     80   310  
09:19:32.134 B197.1    walk   255  -490   250    99    80   323  
09:19:33.133 B197.0    stand  255  -370   550    0     80   323  
09:19:33.133 B197.1    walk   255  -480   310    90    80   264  
09:19:34.146 B197.0    stand  255  -370   550    0     80   264  
09:19:34.146 B197.1    walk   255  -440   330    106   80   230  
09:19:35.136 B197.0    stand  255  -370   550    0     80   230  
09:19:35.136 B197.1    walk   255  -340   340    107   80   212  
09:19:36.042 B197.1    walk   255  -230   300    80    80   117  
09:19:36.042 B197.0    stand  255  -370   490    0     80   236  
09:19:37.041 B197.1    walk   255  -180   250    57    80   306  
09:19:37.041 B197.0    stand  255  -410   450    0     80   304  
09:19:38.039 B197.0    stand  255  -410   450    0     80   0    
09:19:38.039 B197.1    walk   255  -190   250    48    80   297  
09:19:38.585 B197.1    sitgnd 255  -200   250    74    80   10   
09:19:38.585 B197.0    stand  255  -420   440    0     80   290  
09:19:39.129 B197.0    stand  255  -390   470    0     80   42   
09:19:39.129 B197.1    walk   255  -230   260    94    80   264  
09:19:40.082 B197.0    stand  255  -400   480    0     80   278  
09:19:40.082 B197.1    walk   255  -270   300    78    80   222  
09:19:41.080 B197.0    stand  255  -380   480    0     80   210  
09:19:41.080 B197.1    walk   255  -330   310    73    80   177  
09:19:42.084 B197.0    stand  255  -360   480    0     80   172  
09:19:42.084 B197.1    walk   255  -380   330    84    80   151  
09:19:43.082 B197.0    stand  255  -370   460    0     80   130  
09:19:43.082 B197.1    walk   255  -360   330    73    80   130  
09:19:43.976 B197.0    stand  255  -370   500    0     80   170  
09:19:43.976 B197.1    walk   255  -300   330    98    80   183  
09:19:44.980 B197.1    walk   255  -220   320    98    80   80   
09:19:44.980 B197.0    stand  255  -400   490    0     80   247  
09:19:46.030 B197.1    walk   255  -150   320    81    80   302  
09:19:46.030 B197.2    stand  255  -150   470    0     80   150  
09:19:46.030 B197.0    stand  255  -410   510    0     80   263  
09:19:46.986 B197.2    stand  255  -100   480    0     80   311  
09:19:46.986 B197.1    walk   255  -130   320    77    80   162  
09:19:46.986 B197.0    stand  255  -410   530    0     80   350  
09:19:47.984 B197.2    stand  255  -190   480    0     80   225  
09:19:47.984 B197.0    stand  255  -400   510    0     80   212  
09:19:47.984 B197.1    walk   255  -200   310    65    80   282  
09:19:48.987 B197.0    stand  255  -380   500    0     80   261  
09:19:48.987 B197.1    walk   255  -270   320    75    80   210  
09:19:48.987 B197.2    stand  255  -270   470    0     80   150  
09:19:49.988 B197.2    stand  255  -360   480    0     80   90   
09:19:49.988 B197.0    stand  255  -370   510    0     80   31   
09:19:49.988 B197.1    walk   255  -370   350    80    80   160  
09:19:50.993 B197.0    stand  255  -370   510    0     80   160  
09:19:50.993 B197.2    stand  255  -390   490    0     80   28   
09:19:50.993 B197.1    walk   255  -430   400    93    80   98   
09:19:51.993 B197.1    walk   255  -470   460    93    80   72   
09:19:51.993 B197.0    stand  255  -370   510    0     80   111  
09:19:51.993 B197.2    stand  255  -430   480    0     80   67   
09:19:52.990 B197.2    stand  255  -420   480    0     80   10   
09:19:52.990 B197.1    walk   255  -480   520    102   80   72   
09:19:52.990 B197.0    stand  255  -370   510    0     80   110  
09:19:53.890 B197.0    stand  255  -370   510    0     80   0    
09:19:53.890 B197.1    walk   255  -530   530    0     80   161  
09:19:53.890 B197.2    stand  255  -420   480    0     80   120  
09:19:54.890 B197.2    stand  255  -420   480    0     80   0    
09:19:54.890 B197.1    walk   255  -530   540    0     80   125  
09:19:54.890 B197.0    stand  255  -370   510    0     80   162  
09:19:55.892 B197.2    stand  255  -420   480    0     80   58   
09:19:55.892 B197.1    stand  255  -530   540    0     80   125  
09:19:55.892 B197.0    stand  255  -370   510    0     80   162  
09:19:56.894 B197.0    stand  255  -370   510    0     80   0    
09:19:56.894 B197.1    stand  255  -530   540    0     80   162  
09:19:56.894 B197.2    stand  255  -420   480    0     80   125  
09:19:57.904 B197.2    stand  255  -410   480    0     80   10   
09:19:57.904 B197.0    stand  255  -370   510    0     80   50   
09:19:57.904 B197.1    stand  255  -530   540    0     80   162  
09:19:58.895 B197.2    stand  255  -410   480    0     80   134  
09:19:58.895 B197.1    stand  255  -530   540    0     80   134  
09:19:58.895 B197.0    stand  255  -370   510    0     80   162  
09:19:59.902 B197.0    stand  255  -370   510    0     80   0    
09:19:59.902 B197.1    stand  255  -530   540    0     80   162  
09:19:59.902 B197.2    stand  255  -410   480    0     80   134  
09:20:00.897 B197.0    stand  255  -370   510    0     80   50   
09:20:00.897 B197.1    stand  255  -530   540    0     80   162  
09:20:00.897 B197.2    stand  255  -410   480    0     80   134  
09:20:01.953 B197.1    stand  255  -530   540    0     80   134  
09:20:01.953 B197.2    stand  255  -410   480    0     80   134  
09:20:02.877 B197.1    stand  255  -530   540    0     80   134  
09:20:03.838 B197.1    stand  255  -530   540    0     80   0    
09:20:04.837 B197.1    stand  255  -530   540    0     80   0    
09:20:05.840 B197.1    stand  255  -530   540    0     80   0    
09:20:06.848 B197.1    stand  255  -530   540    0     80   0    
09:20:07.901 B197.88   88     -    -      -      -     -    -    
09:20:08.857 B197.88   88     -    -      -      -     -    -    
09:20:09.855 B197.88   88     -    -      -      -     -    -    
09:20:29.701 B197.88   88     -    -      -      -     -    -    
09:21:01.344 B197.88   88     -    -      -      -     -    -    
09:21:33.222 B197.88   88     -    -      -      -     -    -    
09:22:04.891 B197.88   88     -    -      -      -     -    -    
09:22:36.808 B197.88   88     -    -      -      -     -    -    
09:23:08.397 B197.88   88     -    -      -      -     -    -    
09:23:40.400 B197.88   88     -    -      -      -     -    -    
09:24:12.042 B197.88   88     -    -      -      -     -    -    
09:24:43.609 B197.88   88     -    -      -      -     -    -    
09:25:15.645 B197.88   88     -    -      -      -     -    -    
09:25:47.044 B197.88   88     -    -      -      -     -    -    
09:26:19.226 B197.88   88     -    -      -      -     -    -    
09:26:50.528 B197.88   88     -    -      -      -     -    -    
09:27:22.506 B197.88   88     -    -      -      -     -    -    
09:27:42.477 B197.0    walk   255  -470   430    92    80   125  
09:27:43.137 B197.0    walk   255  -450   310    110   80   121  
09:27:44.138 B197.0    walk   255  -460   210    90    80   100  
09:27:45.138 B197.0    walk   255  -470   140    62    80   70   
09:27:46.142 B197.0    walk   255  -520   90     71    80   70   
09:27:47.144 B197.0    walk   255  -530   70     118   80   22   
09:27:48.143 B197.0    walk   255  -530   70     89    80   0    
09:27:49.148 B197.0    walk   255  -540   190    0     80   120  
09:27:50.150 B197.0    walk   255  -530   200    0     80   14   
09:27:51.047 B197.0    walk   255  -530   200    0     80   0    
09:27:52.139 B197.0    stand  255  -530   200    0     80   0    
09:27:53.050 B197.0    stand  255  -530   200    0     80   0    
09:27:54.051 B197.0    stand  255  -530   200    0     80   0    
09:27:55.053 B197.0    stand  255  -530   200    0     80   0    
09:27:56.052 B197.0    stand  255  -530   200    0     80   0    
09:27:57.054 B197.0    stand  255  -530   200    0     80   0    
09:27:58.057 B197.0    stand  255  -530   200    0     80   0    
09:27:59.056 B197.0    stand  255  -530   200    0     80   0    
09:28:00.058 B197.0    stand  255  -530   200    0     80   0    
09:28:01.057 B197.0    stand  255  -530   200    0     80   0    
09:28:01.956 B197.0    stand  255  -530   200    0     80   0    
09:28:02.967 B197.0    stand  255  -530   200    0     80   0    
09:28:03.156 B197.0    stand  3    -530   200    0     80   0    
09:28:04.015 B197.88   88     -    -      -      -     -    -    
09:28:04.973 B197.88   88     -    -      -      -     -    -    
09:28:05.976 B197.88   88     -    -      -      -     -    -    
09:28:26.101 B197.88   88     -    -      -      -     -    -    
09:28:35.843 B197.0    stand  3    -530   80     108   80   120  
09:28:36.677 B197.0    stand  3    -510   120    95    80   44   
09:28:37.678 B197.0    walk   3    -500   190    95    80   70   
09:28:38.677 B197.0    walk   3    -490   280    79    80   90   
09:28:39.682 B197.0    walk   3    -480   370    83    80   90   
09:28:40.689 B197.0    walk   3    -470   470    92    80   100  
09:28:41.680 B197.0    walk   3    -480   510    91    80   41   
09:28:42.690 B197.0    walk   3    -480   560    0     80   50   
09:28:43.687 B197.0    walk   3    -480   550    0     80   10   
09:28:44.684 B197.0    stand  3    -480   550    0     80   0    
09:28:45.686 B197.0    stand  3    -480   550    0     80   0    
09:28:46.687 B197.0    stand  3    -480   540    0     80   10   
09:28:47.686 B197.0    stand  3    -480   540    0     80   0    
09:28:48.577 B197.0    stand  3    -480   540    0     80   0    
09:28:49.579 B197.0    stand  3    -480   540    0     80   0    
09:28:50.581 B197.0    stand  3    -480   540    0     80   0    
09:28:51.586 B197.0    stand  3    -480   540    0     80   0    
09:28:52.582 B197.0    stand  3    -480   540    0     80   0    
09:28:53.584 B197.0    stand  3    -480   540    0     80   0    
09:28:54.584 B197.0    stand  3    -480   540    0     80   0    
09:28:55.590 B197.0    stand  3    -480   540    0     80   0    
09:28:56.590 B197.0    stand  3    -480   540    0     80   0    
09:28:57.652 B197.88   88     -    -      -      -     -    -    
09:28:58.609 B197.88   88     -    -      -      -     -    -    
09:28:59.489 B197.88   88     -    -      -      -     -    -    
09:29:15.566 B197.0    stand  255  -470   500    103   80   41   
09:29:16.389 B197.0    stand  255  -460   450    89    80   50   
09:29:17.386 B197.0    walk   255  -460   390    87    80   60   
09:29:18.394 B197.0    walk   255  -400   340    89    80   78   
09:29:19.390 B197.0    walk   255  -320   300    89    80   89   
09:29:20.388 B197.0    walk   255  -230   320    74    80   92   
09:29:21.392 B197.0    walk   255  -190   330    42    80   41   
09:29:22.391 B197.0    walk   255  -190   320    0     80   10   
09:29:23.392 B197.0    walk   255  -220   320    0     80   30   
09:29:24.292 B197.0    stand  255  -210   350    0     80   31   
09:29:25.293 B197.0    stand  255  -210   340    34    80   10   
09:29:26.293 B197.0    stand  255  -200   320    33    80   22   
09:29:27.298 B197.0    stand  255  -210   330    42    80   14   
09:29:28.296 B197.0    stand  255  -220   330    0     80   10   
09:29:29.296 B197.0    stand  255  -230   320    41    80   14   
09:29:30.301 B197.0    stand  255  -220   310    31    80   14   
09:29:31.298 B197.0    stand  255  -210   360    46    80   50   
09:29:32.300 B197.0    stand  255  -220   350    41    80   14   
09:29:33.302 B197.0    stand  255  -210   340    39    80   14   
09:29:34.302 B197.0    stand  255  -220   330    0     80   14   
09:29:35.305 B197.0    stand  255  -210   340    44    80   14   
09:29:36.194 B197.0    stand  255  -210   340    0     80   0    
09:29:37.194 B197.0    stand  255  -210   350    63    80   10   
09:29:38.201 B197.0    stand  255  -200   350    44    80   10   
09:29:39.198 B197.0    stand  255  -210   350    36    80   10   
09:29:40.198 B197.0    stand  255  -220   320    36    80   31   
09:29:41.199 B197.0    stand  255  -220   320    32    80   0    
09:29:42.202 B197.0    stand  255  -220   310    35    80   10   
09:29:43.229 B197.0    stand  255  -250   330    0     80   36   
09:29:44.205 B197.0    stand  255  -250   320    26    80   10   
09:29:45.206 B197.0    stand  255  -220   320    0     80   30   
09:29:46.205 B197.0    stand  255  -230   330    40    80   14   
09:29:47.209 B197.0    stand  255  -220   320    0     80   14   
09:29:48.104 B197.0    stand  255  -220   320    0     80   0    
09:29:49.100 B197.0    stand  255  -220   330    0     80   10   
09:29:50.108 B197.0    stand  255  -210   310    24    80   22   
09:29:51.105 B197.0    stand  255  -220   330    0     80   22   
09:29:52.110 B197.0    stand  255  -220   330    0     80   0    
09:29:53.108 B197.0    stand  255  -220   320    0     80   10   
09:29:54.109 B197.0    stand  255  -250   320    24    80   30   
09:29:55.044 B197.0    stand  255  -250   340    27    80   20   
09:29:56.046 B197.0    stand  255  -260   320    23    80   22   
09:29:57.053 B197.0    stand  255  -250   310    20    80   14   
09:29:58.054 B197.0    stand  255  -270   350    33    80   44   
09:29:59.053 B197.0    stand  255  -250   330    0     80   28   
09:30:00.053 B197.0    stand  255  -290   340    16    80   41   
09:30:01.052 B197.0    stand  255  -250   330    0     80   41   
09:30:02.053 B197.0    stand  255  -250   330    0     80   0    
09:30:03.061 B197.0    stand  255  -240   340    0     80   14   
09:30:04.055 B197.0    stand  255  -250   330    24    80   14   
09:30:05.060 B197.0    stand  255  -260   350    0     80   22   
09:30:06.056 B197.0    stand  255  -250   330    21    80   22   
09:30:06.951 B197.0    stand  255  -250   320    21    80   10   
09:30:08.002 B197.0    stand  255  -250   330    32    80   10   
09:30:08.957 B197.0    stand  255  -260   320    23    80   14   
09:30:09.952 B197.0    stand  255  -240   300    0     80   28   
09:30:10.969 B197.0    stand  255  -250   330    0     80   31   
09:30:12.008 B197.0    stand  255  -260   310    16    80   22   
09:30:12.902 B197.0    stand  255  -240   320    0     80   22   
09:30:13.900 B197.0    stand  255  -240   320    0     80   0    
09:30:14.902 B197.0    stand  255  -240   320    0     80   0    
09:30:15.904 B197.0    stand  255  -240   320    0     80   0    
09:30:16.909 B197.0    stand  255  -240   320    0     80   0    
09:30:17.910 B197.0    stand  255  -240   320    0     80   0    
09:30:18.910 B197.0    stand  255  -240   320    0     80   0    
09:30:19.914 B197.0    stand  255  -250   330    0     80   14   
09:30:20.908 B197.0    stand  255  -250   320    0     80   10   
09:30:21.910 B197.0    stand  255  -250   340    27    80   20   
09:30:22.912 B197.0    stand  255  -240   330    32    80   14   
09:30:23.912 B197.0    stand  255  -250   340    0     80   14   
09:30:24.802 B197.0    stand  255  -250   330    0     80   10   
09:30:25.821 B197.0    stand  255  -250   330    0     80   0    
09:30:26.813 B197.0    stand  255  -250   330    0     80   0    
09:30:27.826 B197.0    stand  255  -260   320    21    80   14   
09:30:28.816 B197.0    stand  255  -250   310    11    80   14   
09:30:29.817 B197.0    stand  255  -260   330    0     80   22   
09:30:30.820 B197.0    stand  255  -260   330    0     80   0    
09:30:31.819 B197.0    stand  255  -270   330    20    80   10   
09:30:32.824 B197.0    stand  255  -240   300    21    80   42   
09:30:33.822 B197.0    stand  255  -250   310    34    80   14   
09:30:34.837 B197.0    stand  255  -260   320    15    80   14   
09:30:35.718 B197.0    stand  255  -280   340    0     80   28   
09:30:36.717 B197.0    stand  255  -280   340    0     80   0    
09:30:37.717 B197.0    stand  255  -280   340    0     80   0    
09:30:38.721 B197.0    stand  255  -280   340    0     80   0    
09:30:39.720 B197.0    stand  255  -280   330    0     80   10   
09:30:40.728 B197.0    stand  255  -280   330    0     80   0    
09:30:41.724 B197.0    stand  255  -280   330    0     80   0    
09:30:42.721 B197.0    stand  255  -280   330    0     80   0    
09:30:43.723 B197.0    stand  255  -280   330    0     80   0    
09:30:44.730 B197.0    stand  255  -280   330    0     80   0    
09:30:45.725 B197.0    stand  255  -250   320    0     80   31   
09:30:46.725 B197.0    stand  255  -250   330    0     80   10   
09:30:47.618 B197.0    stand  255  -250   330    0     80   0    
09:30:48.622 B197.0    stand  255  -250   330    0     80   0    
09:30:49.619 B197.0    stand  255  -260   340    0     80   14   
09:30:50.625 B197.0    stand  255  -250   320    0     80   22   
09:30:51.624 B197.0    stand  255  -250   330    0     80   10   
09:30:52.624 B197.0    stand  255  -250   320    0     80   10   
09:30:53.628 B197.0    stand  255  -250   320    0     80   0    
09:30:54.630 B197.0    stand  255  -270   330    0     80   22   
09:30:55.625 B197.0    stand  255  -250   320    0     80   22   
09:30:56.629 B197.0    stand  255  -250   320    0     80   0    
09:30:57.629 B197.0    stand  255  -260   340    0     80   22   
09:30:58.631 B197.0    stand  255  -240   320    0     80   28   
09:30:59.525 B197.0    stand  255  -260   320    27    80   20   
09:31:00.534 B197.0    stand  255  -240   310    0     80   22   
09:31:01.529 B197.0    stand  255  -270   330    0     80   36   
09:31:02.536 B197.0    stand  255  -280   350    0     80   22   
09:31:03.528 B197.0    stand  255  -250   330    36    80   36   
09:31:04.534 B197.0    stand  255  -260   340    0     80   14   
09:31:05.532 B197.0    stand  255  -260   340    0     80   0    
09:31:06.584 B197.0    stand  255  -270   340    22    80   10   
09:31:07.538 B197.0    stand  255  -240   310    0     80   42   
09:31:08.539 B197.0    stand  255  -240   310    0     80   0    
09:31:09.536 B197.0    stand  255  -230   310    0     80   10   
09:31:10.541 B197.0    stand  255  -230   310    0     80   0    
09:31:11.426 B197.0    stand  255  -240   310    0     80   10   
09:31:12.428 B197.0    stand  255  -250   310    0     80   10   
09:31:13.437 B197.0    stand  255  -250   330    0     80   20   
09:31:14.431 B197.0    stand  255  -260   340    23    80   14   
09:31:15.453 B197.0    stand  255  -270   320    0     80   22   
09:31:16.450 B197.0    stand  255  -270   320    0     80   0    
09:31:17.456 B197.0    stand  255  -240   340    0     80   36   
09:31:18.454 B197.0    stand  255  -250   310    0     80   31   
09:31:19.480 B197.0    stand  255  -240   330    0     80   22   
09:31:20.464 B197.0    stand  255  -250   320    0     80   14   
09:31:21.348 B197.0    stand  255  -240   300    0     80   22   
09:31:22.348 B197.0    stand  255  -260   320    25    80   28   
09:31:23.351 B197.0    stand  255  -260   330    0     80   10   
09:31:24.350 B197.0    stand  255  -250   330    0     80   10   
09:31:25.350 B197.0    stand  255  -250   330    0     80   0    
09:31:26.356 B197.0    stand  255  -250   330    0     80   0    
09:31:27.361 B197.0    stand  255  -270   330    0     80   20   
09:31:28.354 B197.0    stand  255  -250   320    0     80   22   
09:31:29.374 B197.0    stand  255  -250   320    0     80   0    
09:31:30.358 B197.0    stand  255  -270   320    0     80   20   
09:31:31.357 B197.0    stand  255  -290   330    0     80   22   
09:31:32.362 B197.0    stand  255  -250   310    15    80   44   
09:31:33.266 B197.0    stand  255  -250   330    0     80   20   
09:31:34.255 B197.0    stand  255  -250   330    0     80   0    
09:31:35.253 B197.0    stand  255  -240   340    0     80   14   
09:31:36.257 B197.0    stand  255  -240   340    0     80   0    
09:31:37.258 B197.0    stand  255  -280   350    0     80   41   
09:31:38.256 B197.0    stand  255  -250   340    0     80   31   
09:31:39.257 B197.0    stand  255  -240   320    0     80   22   
09:31:40.264 B197.0    stand  255  -240   320    0     80   0    
09:31:41.261 B197.0    stand  255  -270   330    31    80   31   
09:31:42.260 B197.0    stand  255  -270   330    25    80   0    
09:31:43.262 B197.0    stand  255  -250   320    9     80   22   
09:31:44.261 B197.0    stand  255  -260   340    14    80   22   
09:31:45.159 B197.0    stand  255  -270   330    9     80   14   
09:31:46.164 B197.0    stand  255  -280   350    0     80   22   
09:31:47.165 B197.0    stand  255  -240   340    0     80   41   
09:31:48.161 B197.0    stand  255  -240   320    17    80   20   
09:31:49.173 B197.0    stand  255  -250   330    17    80   14   
09:31:50.165 B197.0    stand  255  -250   340    22    80   10   
09:31:51.162 B197.0    stand  255  -240   330    20    80   14   
09:31:52.168 B197.0    stand  255  -220   330    0     80   20   
09:31:53.167 B197.0    stand  255  -250   340    0     80   31   
09:31:54.166 B197.0    stand  255  -260   330    19    80   14   
09:31:55.170 B197.0    stand  255  -250   330    0     80   10   
09:31:56.170 B197.0    stand  255  -240   340    0     80   14   
09:31:57.061 B197.0    stand  255  -240   320    30    80   20   
09:31:58.066 B197.0    stand  255  -250   330    0     80   14   
09:31:59.064 B197.0    stand  255  -240   300    0     80   31   
09:32:00.066 B197.0    stand  255  -250   310    12    80   14   
09:32:01.065 B197.0    stand  255  -240   310    14    80   10   
09:32:02.076 B197.0    stand  255  -260   320    0     80   22   
09:32:03.071 B197.0    stand  255  -240   330    0     80   22   
09:32:04.090 B197.0    stand  255  -250   340    17    80   14   
09:32:05.084 B197.0    stand  255  -270   340    5     80   20   
09:32:06.132 B197.0    stand  255  -270   330    0     80   10   
09:32:06.992 B197.0    stand  255  -240   330    0     80   30   
09:32:07.983 B197.0    stand  255  -240   330    0     80   0    
09:32:08.982 B197.0    stand  255  -250   320    0     80   14   
09:32:09.983 B197.0    stand  255  -240   330    0     80   14   
09:32:10.985 B197.0    stand  255  -230   320    0     80   14   
09:32:11.988 B197.0    stand  255  -230   320    0     80   0    
09:32:12.988 B197.0    stand  255  -240   330    24    80   14   
09:32:13.987 B197.0    stand  255  -260   330    0     80   20   
09:32:14.988 B197.0    stand  255  -250   340    26    80   14   
09:32:15.997 B197.0    stand  255  -250   330    0     80   10   
09:32:17.008 B197.0    stand  255  -250   330    0     80   0    
09:32:17.992 B197.0    stand  255  -250   330    0     80   0    
09:32:18.887 B197.0    stand  255  -250   330    0     80   0    
09:32:19.894 B197.0    stand  255  -250   330    0     80   0    
09:32:20.894 B197.0    stand  255  -250   330    46    80   0    
09:32:21.897 B197.0    stand  255  -240   330    22    80   10   
09:32:22.896 B197.0    stand  255  -240   330    0     80   0    
09:32:23.897 B197.0    stand  255  -240   320    26    80   10   
09:32:24.916 B197.0    stand  255  -250   310    12    80   14   
09:32:25.901 B197.0    stand  255  -250   340    0     80   30   
09:32:26.936 B197.0    stand  255  -240   330    0     80   14   
09:32:27.905 B197.0    stand  255  -290   330    0     80   50   
09:32:28.902 B197.0    stand  255  -270   330    0     80   20   
09:32:29.796 B197.0    stand  255  -250   340    0     80   22   
09:32:30.798 B197.0    stand  255  -260   340    0     80   10   
09:32:31.803 B197.0    stand  255  -270   340    21    80   10   
09:32:32.798 B197.0    stand  255  -290   340    0     80   20   
09:32:33.802 B197.0    stand  255  -290   350    0     80   10   
09:32:34.800 B197.0    stand  255  -240   320    0     80   58   
09:32:35.803 B197.0    stand  255  -270   350    7     80   42   
09:32:36.808 B197.0    stand  255  -250   330    47    80   28   
09:32:37.812 B197.0    stand  255  -260   340    0     80   14   
09:32:38.822 B197.0    stand  255  -250   330    0     80   14   
09:32:39.806 B197.0    stand  255  -250   320    0     80   10   
09:32:40.822 B197.0    stand  255  -240   310    0     80   14   
09:32:41.705 B197.0    stand  255  -250   330    0     80   22   
09:32:42.700 B197.0    stand  255  -240   330    0     80   10   
09:32:43.709 B197.0    stand  255  -220   330    24    80   20   
09:32:44.702 B197.0    stand  255  -240   320    0     80   22   
09:32:45.706 B197.0    stand  255  -240   330    0     80   10   
09:32:46.706 B197.0    stand  255  -260   340    26    80   22   
09:32:47.708 B197.0    stand  255  -240   310    0     80   36   
09:32:48.707 B197.0    stand  255  -260   340    0     80   36   
09:32:49.708 B197.0    stand  255  -280   350    0     80   22   
09:32:50.713 B197.0    stand  255  -280   350    0     80   0    
09:32:51.717 B197.0    stand  255  -280   350    0     80   0    
09:32:52.618 B197.0    stand  255  -240   300    18    80   64   
09:32:53.618 B197.0    stand  255  -240   310    0     80   10   
09:32:54.616 B197.0    stand  255  -240   300    0     80   10   
09:32:55.617 B197.0    stand  255  -250   330    31    80   31   
09:32:56.620 B197.0    stand  255  -230   300    0     80   36   
09:32:57.618 B197.0    stand  255  -240   310    25    80   14   
09:32:58.628 B197.0    stand  255  -250   310    24    80   10   

```

**汇总**: xray tick 409 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

# case-0777-0624-12001201 — 每 tick belief 时间线 (room fd00:0:3:311:2:100, TZ America/Los_Angeles)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
12:00:00 0777.0   077700000139  FALL    0    NoReport FALL               trk  0.50 Empty      1   0     0.00  0.02  0.08  0.00  0.86  0.04
12:00:01 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    Empty      1   0     0.00  0.02  0.08  0.00  0.86  0.04
12:00:01 0777.0   077700000139  FALL    0    NoReport FALL               trk  0.51 Empty      1   1     0.00  0.02  0.10  0.00  0.83  0.01
12:00:02 0777.0   077700000139  FALL    32   NoReport FALL               trk  0.52 Empty      1   2     0.03  0.02  0.11  0.01  0.76  0.01
12:00:03 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    Empty      1   2     0.03  0.02  0.11  0.01  0.76  0.01
12:00:03 0777.0   077700000139  FALL    25   NoReport FALL               trk  0.53 Empty      1   2     0.20  0.02  0.10  0.01  0.58  0.01
12:00:04 0777.0   077700000139  FALL    16   NoReport FALL               trk  0.54 Fallen     1   3     0.67  0.01  0.04  0.00  0.22  0.00
12:00:05 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    Fallen     1   3     0.67  0.01  0.04  0.00  0.22  0.00
12:00:05 0777.0   077700000139  FALL    39   NoReport FALL               trk  0.55 Fallen     1   4     0.94  0.00  0.01  0.00  0.04  0.00
12:00:06 0777.0   077700000139  FALL    35   NoReport FALL               trk  1.00 Fallen     1   5     0.99  0.00  0.00  0.00  0.00  0.00
12:00:07 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    Fallen     1   5     0.99  0.00  0.00  0.00  0.00  0.00
12:00:07 0777.0   077700000139  FALL    42   NoReport FALL               trk  1.00 Fallen     1   6     1.00  0.00  0.00  0.00  0.00  0.00
12:00:07 0777.E   -             -       0    NoReport Fall(rdr)          room -    Fallen     1   6     1.00  0.00  0.00  0.00  0.00  0.00
12:00:07 0777.0   077700000139  susfall 70   NoReport susfall            trk  1.00 Fallen     1   0     1.00  0.00  0.00  0.00  0.00  0.00
12:00:08 0777.0   077700000139  stand   62   NoReport stand              trk  1.00 Fallen     1   0     0.98  0.00  0.02  0.00  0.00  0.00
12:00:09 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    Fallen     1   0     0.98  0.00  0.02  0.00  0.00  0.00
12:00:09 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.95  0.00  0.04  0.00  0.00  0.00
12:00:10 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.91  0.00  0.07  0.00  0.00  0.00
12:00:11 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    Fallen     1   0     0.91  0.00  0.07  0.00  0.00  0.00
12:00:11 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.86  0.00  0.10  0.00  0.00  0.00
12:00:11 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.79  0.00  0.15  0.00  0.01  0.01
12:00:12 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.71  0.00  0.20  0.00  0.01  0.01
12:00:13 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    Fallen     1   0     0.71  0.00  0.20  0.00  0.01  0.01
12:00:13 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.62  0.00  0.26  0.00  0.02  0.01
12:00:14 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.52  0.01  0.32  0.00  0.02  0.02
12:00:15 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    Fallen     1   0     0.52  0.01  0.32  0.00  0.02  0.02
12:00:15 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.42  0.01  0.38  0.01  0.03  0.02
12:00:16 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.33  0.01  0.44  0.01  0.03  0.02
12:00:17 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   0     0.33  0.01  0.44  0.01  0.03  0.02
12:00:17 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.25  0.01  0.49  0.01  0.04  0.03
12:00:18 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.19  0.01  0.53  0.01  0.04  0.03
12:00:19 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   0     0.19  0.01  0.53  0.01  0.04  0.03
12:00:19 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.14  0.01  0.56  0.01  0.05  0.03
12:00:20 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.14  0.01  0.56  0.01  0.05  0.03
12:00:21 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   0     0.14  0.01  0.56  0.01  0.05  0.03
12:00:21 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.07  0.01  0.60  0.01  0.05  0.03
12:00:21 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.05  0.01  0.61  0.01  0.05  0.03
12:00:22 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.04  0.01  0.62  0.01  0.05  0.03
12:00:23 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   0     0.04  0.01  0.62  0.01  0.05  0.03
12:00:23 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.03  0.01  0.63  0.01  0.05  0.03
12:00:24 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.02  0.02  0.63  0.01  0.06  0.03
12:00:25 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   0     0.02  0.02  0.63  0.01  0.06  0.03
12:00:25 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.02  0.02  0.63  0.01  0.06  0.03
12:00:26 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.02  0.02  0.63  0.01  0.06  0.03
12:00:27 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   0     0.02  0.02  0.63  0.01  0.06  0.03
12:00:27 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.02  0.64  0.01  0.06  0.03
12:00:28 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.02  0.64  0.01  0.06  0.03
12:00:29 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   0     0.01  0.02  0.64  0.01  0.06  0.03
12:00:29 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.02  0.64  0.01  0.06  0.03
12:00:30 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   25    0.01  0.02  0.64  0.01  0.06  0.03
12:00:31 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   25    0.01  0.02  0.64  0.01  0.06  0.03
12:00:31 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   26    0.01  0.02  0.64  0.01  0.06  0.03
12:00:32 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   27    0.01  0.02  0.64  0.01  0.06  0.03
12:00:33 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   27    0.01  0.02  0.64  0.01  0.06  0.03
12:00:33 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   28    0.01  0.02  0.64  0.01  0.06  0.03
12:00:34 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   29    0.01  0.02  0.64  0.01  0.06  0.03
12:00:35 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   29    0.01  0.02  0.64  0.01  0.06  0.03
12:00:35 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   30    0.01  0.02  0.64  0.01  0.06  0.03
12:00:36 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   31    0.01  0.02  0.64  0.01  0.06  0.03
12:00:37 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   31    0.01  0.02  0.64  0.01  0.06  0.03
12:00:37 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   32    0.01  0.02  0.64  0.01  0.06  0.03
12:00:38 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   33    0.01  0.02  0.64  0.01  0.06  0.03
12:00:39 0777.0   077700000139  stand   29   NoReport stand              trk  1.00 OpenFloor  1   34    0.01  0.02  0.64  0.01  0.06  0.03
12:00:40 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   34    0.01  0.02  0.64  0.01  0.06  0.03
12:00:40 0777.0   077700000139  stand   40   NoReport stand              trk  1.00 OpenFloor  1   35    0.01  0.02  0.64  0.01  0.06  0.03
12:00:41 0777.0   077700000139  stand   68   NoReport stand              trk  1.00 OpenFloor  1   36    0.01  0.02  0.64  0.01  0.06  0.03
12:00:42 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   36    0.01  0.02  0.64  0.01  0.06  0.03
12:00:42 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   37    0.01  0.02  0.64  0.01  0.06  0.03
12:00:43 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   38    0.01  0.02  0.64  0.01  0.06  0.03
12:00:44 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   38    0.01  0.02  0.64  0.01  0.06  0.03
12:00:44 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   39    0.01  0.02  0.64  0.01  0.06  0.03
12:00:45 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   40    0.01  0.02  0.63  0.01  0.06  0.03
12:00:46 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   40    0.01  0.02  0.63  0.01  0.06  0.03
12:00:46 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   41    0.01  0.02  0.63  0.01  0.06  0.03
12:00:47 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   42    0.01  0.02  0.63  0.01  0.06  0.03
12:00:48 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   42    0.01  0.02  0.63  0.01  0.06  0.03
12:00:48 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   43    0.01  0.02  0.63  0.01  0.06  0.03
12:00:49 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   44    0.01  0.02  0.63  0.01  0.06  0.03
12:00:50 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   44    0.01  0.02  0.63  0.01  0.06  0.03
12:00:50 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   45    0.01  0.02  0.63  0.01  0.06  0.03
12:00:51 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   46    0.01  0.02  0.63  0.01  0.06  0.03
12:00:52 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   46    0.01  0.02  0.63  0.01  0.06  0.03
12:00:52 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   47    0.01  0.02  0.63  0.01  0.06  0.03
12:00:53 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   48    0.01  0.02  0.63  0.01  0.06  0.03
12:00:54 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   48    0.01  0.02  0.63  0.01  0.06  0.03
12:00:54 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   49    0.01  0.02  0.63  0.01  0.06  0.03
12:00:55 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   50    0.01  0.02  0.63  0.01  0.06  0.03
12:00:56 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   50    0.01  0.02  0.63  0.01  0.06  0.03
12:00:56 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   51    0.01  0.02  0.63  0.01  0.06  0.03
12:00:57 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   52    0.01  0.02  0.63  0.01  0.06  0.03
12:00:58 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   52    0.01  0.02  0.63  0.01  0.06  0.03
12:00:58 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   53    0.01  0.02  0.63  0.01  0.06  0.03
12:00:59 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   54    0.01  0.03  0.63  0.01  0.06  0.03
12:01:00 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   54    0.01  0.03  0.63  0.01  0.06  0.03
12:01:00 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   55    0.01  0.03  0.63  0.01  0.06  0.03
12:01:01 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   56    0.01  0.03  0.63  0.01  0.06  0.03
12:01:02 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   56    0.01  0.03  0.63  0.01  0.06  0.03
12:01:02 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   57    0.01  0.03  0.63  0.01  0.06  0.03
12:01:03 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   58    0.01  0.03  0.63  0.01  0.06  0.03
12:01:04 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   58    0.01  0.03  0.63  0.01  0.06  0.03
12:01:04 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   59    0.01  0.03  0.63  0.01  0.06  0.03
12:01:05 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   60    0.01  0.03  0.63  0.01  0.06  0.03
12:01:06 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   60    0.01  0.03  0.63  0.01  0.06  0.03
12:01:06 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   61    0.01  0.03  0.63  0.01  0.06  0.03
12:01:07 0777.0   077700000139  stand   28   NoReport stand              trk  1.00 OpenFloor  1   62    0.01  0.03  0.63  0.01  0.06  0.03
12:01:08 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   62    0.01  0.03  0.63  0.01  0.06  0.03
12:01:08 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   63    0.01  0.03  0.63  0.01  0.06  0.03
12:01:09 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   64    0.01  0.03  0.63  0.01  0.06  0.03
12:01:10 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   64    0.01  0.03  0.63  0.01  0.06  0.03
12:01:10 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   65    0.01  0.03  0.63  0.02  0.06  0.03
12:01:11 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   66    0.01  0.03  0.63  0.02  0.06  0.03
12:01:12 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   66    0.01  0.03  0.63  0.02  0.06  0.03
12:01:12 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   67    0.01  0.03  0.63  0.02  0.06  0.03
12:01:13 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   68    0.01  0.03  0.63  0.02  0.06  0.03
12:01:14 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   68    0.01  0.03  0.63  0.02  0.06  0.03
12:01:14 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   69    0.01  0.03  0.63  0.02  0.06  0.03
12:01:15 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   70    0.01  0.03  0.63  0.02  0.06  0.03
12:01:16 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   70    0.01  0.03  0.63  0.02  0.06  0.03
12:01:16 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   71    0.01  0.03  0.63  0.02  0.06  0.03
12:01:17 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   72    0.01  0.03  0.63  0.02  0.06  0.03
12:01:18 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   72    0.01  0.03  0.63  0.02  0.06  0.03
12:01:18 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   73    0.01  0.03  0.63  0.02  0.06  0.03
12:01:19 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   74    0.01  0.03  0.63  0.02  0.06  0.03
12:01:20 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   74    0.01  0.03  0.63  0.02  0.06  0.03
12:01:20 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   75    0.01  0.03  0.63  0.02  0.06  0.03
12:01:21 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   76    0.01  0.03  0.63  0.02  0.06  0.03
12:01:22 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   76    0.01  0.03  0.63  0.02  0.06  0.03
12:01:22 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   77    0.01  0.03  0.63  0.02  0.06  0.03
12:01:23 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   78    0.01  0.03  0.63  0.02  0.06  0.03
12:01:24 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   78    0.01  0.03  0.63  0.02  0.06  0.03
12:01:24 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   79    0.01  0.03  0.63  0.02  0.06  0.03
12:01:25 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   80    0.01  0.03  0.63  0.02  0.06  0.03
12:01:26 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   80    0.01  0.03  0.63  0.02  0.06  0.03
12:01:26 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   81    0.01  0.03  0.63  0.02  0.06  0.03
12:01:27 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   82    0.01  0.03  0.63  0.02  0.06  0.03
12:01:28 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   83    0.01  0.03  0.63  0.02  0.06  0.03
12:01:29 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   83    0.01  0.03  0.63  0.02  0.06  0.03
12:01:29 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   84    0.01  0.03  0.63  0.02  0.06  0.03
12:01:30 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   85    0.01  0.03  0.63  0.02  0.06  0.03
12:01:31 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   85    0.01  0.03  0.63  0.02  0.06  0.03
12:01:31 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   86    0.01  0.03  0.63  0.02  0.06  0.03
12:01:32 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   87    0.01  0.03  0.63  0.02  0.06  0.03
12:01:33 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   87    0.01  0.03  0.63  0.02  0.06  0.03
12:01:33 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   88    0.01  0.03  0.63  0.02  0.06  0.03
12:01:34 0777.0   077700000139  stand   46   NoReport stand              trk  1.00 OpenFloor  1   89    0.01  0.03  0.63  0.02  0.06  0.03
12:01:35 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   89    0.01  0.03  0.63  0.02  0.06  0.03
12:01:35 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   90    0.01  0.03  0.63  0.02  0.06  0.03
12:01:36 0777.0   077700000139  stand   30   NoReport stand              trk  1.00 OpenFloor  1   91    0.01  0.03  0.63  0.02  0.06  0.03
12:01:37 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   91    0.01  0.03  0.63  0.02  0.06  0.03
12:01:37 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   92    0.01  0.03  0.63  0.02  0.06  0.03
12:01:38 0777.0   077700000139  stand   24   NoReport stand              trk  1.00 OpenFloor  1   93    0.00  0.02  0.60  0.01  0.05  0.03
12:01:39 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   93    0.00  0.02  0.60  0.01  0.05  0.03
12:01:39 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   94    0.00  0.02  0.56  0.01  0.05  0.03
12:01:40 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   95    0.00  0.02  0.53  0.01  0.04  0.03
12:01:41 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   95    0.00  0.02  0.53  0.01  0.04  0.03
12:01:41 0777.0   077700000139  stand   39   NoReport stand              trk  1.00 OpenFloor  1   96    0.00  0.02  0.57  0.02  0.05  0.03
12:01:42 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   97    0.00  0.03  0.60  0.02  0.05  0.03
12:01:43 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   97    0.00  0.03  0.60  0.02  0.05  0.03
12:01:43 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   98    0.01  0.03  0.61  0.02  0.05  0.03
12:01:44 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   99    0.01  0.03  0.62  0.02  0.05  0.03
12:01:45 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   99    0.01  0.03  0.62  0.02  0.05  0.03
12:01:45 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   100   0.01  0.03  0.62  0.02  0.05  0.03
12:01:46 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   101   0.01  0.03  0.63  0.02  0.05  0.03
12:01:47 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   101   0.01  0.03  0.63  0.02  0.05  0.03
12:01:47 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   102   0.01  0.03  0.63  0.02  0.05  0.03
12:01:48 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   103   0.01  0.03  0.63  0.02  0.05  0.03
12:01:49 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   103   0.01  0.03  0.63  0.02  0.05  0.03
12:01:49 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   104   0.01  0.03  0.63  0.02  0.06  0.03
12:01:50 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   105   0.01  0.03  0.63  0.02  0.06  0.03
12:01:51 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   105   0.01  0.03  0.63  0.02  0.06  0.03
12:01:51 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   106   0.01  0.03  0.63  0.02  0.06  0.03
12:01:52 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   107   0.01  0.03  0.63  0.02  0.06  0.03
12:01:53 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   107   0.01  0.03  0.63  0.02  0.06  0.03
12:01:53 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   108   0.01  0.03  0.63  0.02  0.06  0.03
12:01:54 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   109   0.01  0.03  0.63  0.02  0.06  0.03
12:01:55 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   109   0.01  0.03  0.63  0.02  0.06  0.03
12:01:55 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   110   0.01  0.03  0.63  0.02  0.06  0.03
12:01:56 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   111   0.01  0.03  0.63  0.02  0.06  0.03
12:01:57 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   111   0.01  0.03  0.63  0.02  0.06  0.03
12:01:57 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   112   0.01  0.03  0.63  0.02  0.06  0.03
12:01:58 0777.0   077700000139  stand   0    NoReport stand              trk  1.00 OpenFloor  1   113   0.01  0.03  0.63  0.02  0.06  0.03
12:01:59 0866.0   -             pad     -    NoReport pad LeftBed HR=None RR=None mv=0 turn=0 room -    OpenFloor  1   113   0.01  0.03  0.63  0.02  0.06  0.03
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
12:00:00.139 0777.0    FALL   1    -180   170    0     80        
12:00:01.141 0777.0    FALL   1    -180   170    0     80   0    
12:00:02.145 0777.0    FALL   1    -170   180    32    80   14   
12:00:03.034 0777.0    FALL   1    -170   190    25    80   10   
12:00:04.035 0777.0    FALL   1    -140   180    16    80   31   
12:00:05.038 0777.0    FALL   1    -160   180    39    80   20   
12:00:06.039 0777.0    FALL   1    -180   180    35    80   20   
12:00:07.039 0777.0    FALL   1    -190   170    42    80   14   
12:00:07.492 0777.0    susfall 1    -200   180    70    80   14   
12:00:08.064 0777.0    stand  1    -190   190    62    80   14   
12:00:09.060 0777.0    stand  1    -190   190    0     80   0    
12:00:10.062 0777.0    stand  1    -180   180    0     80   14   
12:00:11.071 0777.0    stand  1    -180   180    0     80   0    
12:00:11.964 0777.0    stand  1    -180   180    0     80   0    
12:00:12.964 0777.0    stand  1    -190   190    0     80   14   
12:00:13.965 0777.0    stand  1    -200   200    0     80   14   
12:00:14.965 0777.0    stand  1    -190   200    0     80   10   
12:00:15.968 0777.0    stand  1    -190   190    0     80   10   
12:00:16.966 0777.0    stand  1    -190   190    0     80   0    
12:00:17.966 0777.0    stand  1    -190   190    0     80   0    
12:00:18.974 0777.0    stand  1    -190   190    0     80   0    
12:00:19.984 0777.0    stand  1    -190   190    0     80   0    
12:00:21.038 0777.0    stand  1    -190   190    0     80   0    
12:00:21.880 0777.0    stand  1    -190   190    0     80   0    
12:00:22.881 0777.0    stand  1    -190   190    0     80   0    
12:00:23.883 0777.0    stand  1    -190   190    0     80   0    
12:00:24.886 0777.0    stand  1    -190   190    0     80   0    
12:00:25.886 0777.0    stand  1    -190   190    0     80   0    
12:00:26.895 0777.0    stand  1    -190   190    0     80   0    
12:00:27.902 0777.0    stand  1    -190   190    0     80   0    
12:00:28.887 0777.0    stand  1    -190   190    0     80   0    
12:00:29.905 0777.0    stand  1    -190   190    0     80   0    
12:00:30.902 0777.0    stand  1    -190   190    0     80   0    
12:00:31.892 0777.0    stand  1    -190   190    0     80   0    
12:00:32.889 0777.0    stand  1    -190   190    0     80   0    
12:00:33.784 0777.0    stand  1    -190   190    0     80   0    
12:00:34.785 0777.0    stand  1    -190   190    0     80   0    
12:00:35.785 0777.0    stand  1    -190   190    0     80   0    
12:00:36.786 0777.0    stand  1    -190   190    0     80   0    
12:00:37.790 0777.0    stand  1    -190   190    0     80   0    
12:00:38.791 0777.0    stand  1    -190   190    0     80   0    
12:00:39.793 0777.0    stand  1    -170   190    29    80   20   
12:00:40.791 0777.0    stand  1    -180   190    40    80   10   
12:00:41.794 0777.0    stand  1    -190   190    68    80   10   
12:00:42.796 0777.0    stand  1    -200   200    0     80   14   
12:00:43.799 0777.0    stand  1    -200   200    0     80   0    
12:00:44.797 0777.0    stand  1    -200   200    0     80   0    
12:00:45.689 0777.0    stand  1    -200   200    0     80   0    
12:00:46.692 0777.0    stand  1    -200   200    0     80   0    
12:00:47.696 0777.0    stand  1    -200   200    0     80   0    
12:00:48.697 0777.0    stand  1    -200   200    0     80   0    
12:00:49.693 0777.0    stand  1    -200   200    0     80   0    
12:00:50.692 0777.0    stand  1    -200   200    0     80   0    
12:00:51.643 0777.0    stand  1    -200   200    0     80   0    
12:00:52.640 0777.0    stand  1    -200   200    0     80   0    
12:00:53.641 0777.0    stand  1    -200   200    0     80   0    
12:00:54.645 0777.0    stand  1    -200   200    0     80   0    
12:00:55.653 0777.0    stand  1    -200   200    0     80   0    
12:00:56.648 0777.0    stand  1    -200   200    0     80   0    
12:00:57.648 0777.0    stand  1    -200   200    0     80   0    
12:00:58.651 0777.0    stand  1    -200   200    0     80   0    
12:00:59.648 0777.0    stand  1    -200   200    0     80   0    
12:01:00.648 0777.0    stand  1    -200   200    0     80   0    
12:01:01.652 0777.0    stand  1    -200   200    0     80   0    
12:01:02.650 0777.0    stand  1    -200   200    0     80   0    
12:01:03.545 0777.0    stand  1    -200   200    0     80   0    
12:01:04.547 0777.0    stand  1    -200   200    0     80   0    
12:01:05.548 0777.0    stand  1    -200   200    0     80   0    
12:01:06.555 0777.0    stand  1    -200   200    0     80   0    
12:01:07.591 0777.0    stand  1    -180   190    28    80   22   
12:01:08.600 0777.0    stand  1    -170   180    0     80   14   
12:01:09.596 0777.0    stand  1    -170   180    0     80   0    
12:01:10.490 0777.0    stand  1    -170   180    0     80   0    
12:01:11.490 0777.0    stand  1    -170   180    0     80   0    
12:01:12.506 0777.0    stand  1    -170   180    0     80   0    
12:01:13.493 0777.0    stand  1    -170   180    0     80   0    
12:01:14.493 0777.0    stand  1    -170   180    0     80   0    
12:01:15.494 0777.0    stand  1    -170   180    0     80   0    
12:01:16.496 0777.0    stand  1    -170   180    0     80   0    
12:01:17.497 0777.0    stand  1    -170   180    0     80   0    
12:01:18.500 0777.0    stand  1    -170   180    0     80   0    
12:01:19.498 0777.0    stand  1    -170   180    0     80   0    
12:01:20.535 0777.0    stand  1    -170   180    0     80   0    
12:01:21.500 0777.0    stand  1    -170   180    0     80   0    
12:01:22.395 0777.0    stand  1    -170   180    0     80   0    
12:01:23.403 0777.0    stand  1    -170   180    0     80   0    
12:01:24.404 0777.0    stand  1    -170   180    0     80   0    
12:01:25.403 0777.0    stand  1    -170   180    0     80   0    
12:01:26.407 0777.0    stand  1    -170   180    0     80   0    
12:01:27.408 0777.0    stand  1    -170   180    0     80   0    
12:01:28.406 0777.0    stand  1    -170   180    0     80   0    
12:01:29.473 0777.0    stand  1    -170   180    0     80   0    
12:01:30.414 0777.0    stand  1    -170   180    0     80   0    
12:01:31.410 0777.0    stand  1    -170   180    0     80   0    
12:01:32.416 0777.0    stand  1    -170   180    0     80   0    
12:01:33.306 0777.0    stand  1    -170   180    0     80   0    
12:01:34.304 0777.0    stand  1    -170   190    46    80   10   
12:01:35.307 0777.0    stand  1    -180   200    0     80   14   
12:01:36.308 0777.0    stand  1    -170   180    30    80   22   
12:01:37.316 0777.0    stand  1    -160   170    0     80   14   
12:01:38.309 0777.0    stand  1    -150   170    24    80   10   
12:01:39.313 0777.0    stand  1    -150   170    0     80   0    
12:01:40.315 0777.0    stand  1    -150   170    0     80   0    
12:01:41.312 0777.0    stand  1    -170   180    39    80   22   
12:01:42.316 0777.0    stand  1    -170   180    0     80   0    
12:01:43.315 0777.0    stand  1    -170   180    0     80   0    
12:01:44.322 0777.0    stand  1    -170   180    0     80   0    
12:01:45.209 0777.0    stand  1    -170   180    0     80   0    
12:01:46.209 0777.0    stand  1    -170   180    0     80   0    
12:01:47.209 0777.0    stand  1    -170   180    0     80   0    
12:01:48.209 0777.0    stand  1    -170   180    0     80   0    
12:01:49.213 0777.0    stand  1    -170   180    0     80   0    
12:01:50.217 0777.0    stand  1    -170   180    0     80   0    
12:01:51.223 0777.0    stand  1    -170   180    0     80   0    
12:01:52.215 0777.0    stand  1    -170   180    0     80   0    
12:01:53.223 0777.0    stand  1    -170   180    0     80   0    
12:01:54.222 0777.0    stand  1    -170   180    0     80   0    
12:01:55.228 0777.0    stand  1    -170   180    0     80   0    
12:01:56.123 0777.0    stand  1    -170   180    0     80   0    
12:01:57.131 0777.0    stand  1    -170   180    0     80   0    
12:01:58.123 0777.0    stand  1    -170   180    0     80   0    

```

**汇总**: xray tick 123 | fire 0 | Fall 事件 1 (12:00:07) | 结论 = FN(看到 Fall 但 fire=0)

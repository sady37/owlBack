# case-d523-0629-17021705 — 每 tick belief 时间线 (room fd00:0:3:111:3:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
17:02:00 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   0     0.00  0.03  0.08  0.00  0.85  0.04
17:02:00 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   0     0.00  0.03  0.08  0.00  0.85  0.04
17:02:01 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   1     0.00  0.04  0.10  0.00  0.82  0.01
17:02:02 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   1     0.00  0.04  0.11  0.01  0.76  0.01
17:02:03 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   3     0.00  0.05  0.12  0.01  0.71  0.01
17:02:04 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   4     0.00  0.05  0.13  0.02  0.66  0.01
17:02:05 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   5     0.00  0.06  0.14  0.03  0.62  0.01
17:02:06 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   6     0.01  0.06  0.14  0.04  0.58  0.01
17:02:07 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   7     0.01  0.07  0.14  0.04  0.54  0.01
17:02:08 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   7     0.01  0.07  0.15  0.05  0.51  0.02
17:02:09 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   8     0.01  0.07  0.15  0.06  0.49  0.02
17:02:09 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   9     0.01  0.08  0.15  0.06  0.46  0.02
17:02:10 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   9     0.01  0.08  0.15  0.07  0.44  0.02
17:02:11 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   10    0.02  0.08  0.15  0.07  0.42  0.02
17:02:12 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   11    0.02  0.08  0.15  0.08  0.40  0.02
17:02:13 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   12    0.02  0.09  0.15  0.08  0.38  0.02
17:02:14 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   13    0.02  0.09  0.15  0.09  0.36  0.02
17:02:15 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   14    0.03  0.09  0.15  0.09  0.35  0.02
17:02:16 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   15    0.03  0.09  0.15  0.10  0.34  0.02
17:02:17 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   16    0.03  0.09  0.15  0.10  0.33  0.02
17:02:18 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   17    0.03  0.10  0.15  0.10  0.31  0.02
17:02:19 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   18    0.03  0.10  0.16  0.10  0.30  0.02
17:02:20 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   19    0.04  0.10  0.16  0.11  0.30  0.02
17:02:21 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   20    0.04  0.10  0.15  0.11  0.29  0.02
17:02:22 D523.0   D52300400881  stand   85   NoReport stand              room -    Empty      1   21    0.04  0.10  0.15  0.11  0.28  0.02
17:02:23 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   22    0.04  0.10  0.15  0.11  0.27  0.02
17:02:24 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   23    0.05  0.10  0.15  0.12  0.27  0.02
17:02:25 D523.0   D52300400881  stand   94   NoReport stand              room -    Empty      1   24    0.05  0.10  0.15  0.12  0.26  0.02
17:02:26 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   25    0.05  0.10  0.15  0.12  0.25  0.02
17:02:27 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   26    0.05  0.10  0.15  0.12  0.25  0.02
17:02:28 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   27    0.06  0.10  0.15  0.12  0.25  0.02
17:02:29 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   28    0.06  0.10  0.15  0.12  0.24  0.02
17:02:30 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   29    0.06  0.10  0.15  0.12  0.24  0.02
17:02:31 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   30    0.06  0.10  0.15  0.12  0.23  0.02
17:02:31 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   30    0.06  0.10  0.15  0.12  0.23  0.02
17:02:32 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   31    0.07  0.10  0.15  0.12  0.23  0.02
17:02:33 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   32    0.07  0.10  0.15  0.12  0.23  0.02
17:02:34 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   33    0.07  0.10  0.15  0.13  0.22  0.02
17:02:35 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   34    0.07  0.10  0.15  0.13  0.22  0.02
17:02:36 D523.0   D52300400881  stand   87   NoReport stand              room -    Empty      1   35    0.07  0.10  0.15  0.13  0.22  0.02
17:02:37 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   36    0.08  0.11  0.15  0.13  0.22  0.02
17:02:38 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   37    0.08  0.11  0.15  0.13  0.21  0.02
17:02:39 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   38    0.08  0.11  0.15  0.13  0.21  0.02
17:02:40 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   39    0.08  0.11  0.15  0.13  0.21  0.02
17:02:41 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   40    0.09  0.11  0.15  0.13  0.21  0.02
17:02:41 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   41    0.09  0.10  0.15  0.13  0.21  0.02
17:02:42 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   41    0.09  0.10  0.15  0.13  0.20  0.02
17:02:43 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   42    0.09  0.10  0.15  0.13  0.20  0.02
17:02:43 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   43    0.10  0.10  0.15  0.13  0.20  0.02
17:02:44 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   44    0.10  0.10  0.15  0.13  0.20  0.02
17:02:45 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   45    0.10  0.10  0.15  0.13  0.20  0.02
17:02:46 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   46    0.10  0.10  0.15  0.13  0.20  0.02
17:02:47 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   47    0.10  0.10  0.15  0.13  0.20  0.02
17:02:48 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   48    0.11  0.10  0.15  0.13  0.20  0.02
17:02:49 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   49    0.11  0.10  0.15  0.13  0.20  0.02
17:02:50 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   50    0.11  0.10  0.15  0.13  0.19  0.02
17:02:51 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   51    0.11  0.10  0.15  0.13  0.19  0.02
17:02:52 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   52    0.11  0.10  0.15  0.13  0.19  0.02
17:02:53 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   53    0.12  0.10  0.15  0.13  0.19  0.02
17:02:54 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   54    0.12  0.10  0.15  0.13  0.19  0.02
17:02:55 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   55    0.12  0.10  0.15  0.13  0.19  0.02
17:02:56 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   56    0.12  0.10  0.15  0.13  0.19  0.02
17:02:57 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   57    0.12  0.10  0.15  0.13  0.19  0.02
17:02:58 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   58    0.12  0.10  0.14  0.13  0.19  0.02
17:02:59 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   59    0.13  0.10  0.14  0.13  0.19  0.02
17:03:00 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   60    0.13  0.10  0.14  0.13  0.19  0.02
17:03:01 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   61    0.13  0.10  0.14  0.13  0.19  0.02
17:03:02 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   62    0.13  0.10  0.14  0.13  0.19  0.02
17:03:03 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   62    0.13  0.10  0.14  0.13  0.19  0.02
17:03:03 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   63    0.13  0.10  0.14  0.12  0.19  0.02
17:03:04 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   64    0.13  0.10  0.14  0.12  0.19  0.02
17:03:05 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   65    0.14  0.10  0.14  0.12  0.19  0.02
17:03:06 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   66    0.14  0.10  0.14  0.12  0.19  0.02
17:03:07 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   67    0.14  0.10  0.14  0.12  0.19  0.02
17:03:08 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   68    0.14  0.10  0.14  0.12  0.18  0.02
17:03:09 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   69    0.14  0.10  0.14  0.12  0.18  0.02
17:03:10 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   70    0.14  0.10  0.14  0.12  0.18  0.02
17:03:11 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   71    0.14  0.10  0.14  0.12  0.18  0.02
17:03:12 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   72    0.14  0.10  0.14  0.12  0.18  0.02
17:03:13 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   72    0.15  0.10  0.14  0.12  0.18  0.02
17:03:13 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   73    0.15  0.10  0.14  0.12  0.18  0.02
17:03:14 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   74    0.15  0.10  0.14  0.12  0.18  0.02
17:03:15 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   75    0.15  0.10  0.14  0.12  0.18  0.02
17:03:16 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   76    0.15  0.10  0.14  0.12  0.18  0.02
17:03:17 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   77    0.15  0.10  0.14  0.12  0.18  0.02
17:03:18 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   78    0.15  0.10  0.14  0.12  0.18  0.02
17:03:19 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   79    0.16  0.10  0.14  0.12  0.18  0.02
17:03:20 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   80    0.16  0.10  0.14  0.12  0.18  0.02
17:03:21 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   81    0.16  0.10  0.14  0.12  0.18  0.02
17:03:22 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   82    0.16  0.10  0.14  0.12  0.18  0.02
17:03:23 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   83    0.16  0.10  0.14  0.12  0.18  0.02
17:03:24 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   84    0.16  0.10  0.14  0.12  0.18  0.02
17:03:25 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      1   85    0.16  0.10  0.14  0.12  0.18  0.02
17:03:26 D523.0   D52300400881  stand   71   NoReport stand              room -    Empty      1   86    0.16  0.10  0.14  0.12  0.18  0.02
17:03:27 09E7.E   -             -       0    NoReport np=1               room -    Empty      1   86    0.16  0.10  0.14  0.12  0.18  0.02
17:03:27 09E7.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      1   87    0.16  0.10  0.14  0.12  0.18  0.02
17:03:27 09E7.0   09E700431918  stand   61   NoReport stand              room -    Empty      2   87    0.17  0.10  0.14  0.12  0.18  0.02
17:03:27 D523.0   D52300400881  stand   102  NoReport stand              room -    Empty      2   87    0.17  0.10  0.14  0.12  0.18  0.02
17:03:28 09E7.0   09E700431918  stand   0    NoReport stand              room -    Empty      2   87    0.17  0.10  0.14  0.12  0.18  0.02
17:03:28 D523.0   D52300400881  stand   103  NoReport stand              room -    Empty      2   88    0.17  0.10  0.14  0.12  0.18  0.02
17:03:29 09E7.0   09E700431918  stand   0    NoReport stand              room -    Empty      2   88    0.17  0.10  0.14  0.12  0.18  0.02
17:03:29 D523.E   -             -       0    NoReport np=2               room -    Empty      2   88    0.17  0.10  0.14  0.12  0.18  0.02
17:03:29 D523.1   D52310329613  stand   100  NoReport stand              trk  0.50 Empty      3   0     0.00  0.03  0.08  0.00  0.85  0.04
17:03:29 D523.0   D52300400881  walk    0    NoReport walk               room -    Empty      3   0     0.17  0.10  0.14  0.12  0.18  0.02
17:03:30 09E7.0   09E700431918  stand   0    NoReport stand              room -    Empty      3   0     0.17  0.10  0.14  0.12  0.18  0.02
17:03:30 D523.1   D52310329613  walk    106  NoReport walk               trk  0.75 Empty      3   0     0.00  0.04  0.11  0.01  0.76  0.01
17:03:30 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      3   0     0.17  0.10  0.14  0.12  0.18  0.02
17:03:31 09E7.0   09E700431918  stand   0    NoReport stand              room -    Empty      3   0     0.17  0.10  0.14  0.12  0.18  0.02
17:03:31 D523.0   D52300400881  stand   0    NoReport stand              room -    Empty      3   0     0.17  0.10  0.14  0.12  0.18  0.02
17:03:31 D523.1   D52310329613  walk    100  NoReport walk               trk  0.77 Empty      3   0     0.00  0.05  0.13  0.02  0.66  0.01
17:03:32 09E7.0   09E700431918  stand   0    NoReport stand              room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.18  0.02
17:03:32 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.18  0.02
17:03:32 D523.1   D52310329613  walk    117  NoReport walk               trk  0.81 Fallen     3   0     0.01  0.06  0.14  0.04  0.58  0.01
17:03:33 09E7.0   09E700431918  stand   0    NoReport stand              room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:33 D523.1   D52310329613  walk    0    NoReport walk               trk  0.92 Fallen     3   0     0.01  0.07  0.15  0.05  0.51  0.02
17:03:33 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:34 09E7.0   09E700431918  stand   0    NoReport stand              room -    Fallen     2   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:34 D523.1   D52310329613  walk    0    NoReport walk               trk  0.75 Fallen     3   0     0.01  0.08  0.15  0.06  0.46  0.02
17:03:34 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:35 09E7.0   09E700431918  stand   0    NoReport stand              room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:35 09E7.E   -             -       0    NoReport ExitRoom(rdr)      room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:35 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:35 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:35 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   0     0.02  0.08  0.15  0.07  0.42  0.02
17:03:36 09E7.E   -             -       0    NoReport np=0  ★0           room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:36 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     3   0     0.18  0.10  0.14  0.12  0.17  0.02
17:03:36 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   8     0.18  0.10  0.14  0.12  0.17  0.02
17:03:36 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   8     0.02  0.09  0.15  0.08  0.38  0.02
17:03:37 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     2   9     0.18  0.10  0.14  0.12  0.17  0.02
17:03:37 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   9     0.03  0.09  0.15  0.09  0.35  0.02
17:03:37 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   9     0.19  0.09  0.14  0.12  0.17  0.02
17:03:38 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     2   10    0.19  0.09  0.14  0.12  0.17  0.02
17:03:38 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   10    0.03  0.09  0.15  0.10  0.33  0.02
17:03:38 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   10    0.19  0.09  0.14  0.12  0.17  0.02
17:03:39 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   11    0.19  0.09  0.14  0.12  0.17  0.02
17:03:39 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   11    0.03  0.10  0.15  0.10  0.31  0.02
17:03:40 D523.1   D52310329613  stand   116  NoReport stand              trk  0.75 Fallen     2   12    0.03  0.10  0.16  0.10  0.30  0.02
17:03:40 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   12    0.19  0.09  0.14  0.12  0.17  0.02
17:03:41 D523.1   D52310329613  walk    109  NoReport walk               trk  0.75 Fallen     2   13    0.04  0.10  0.16  0.11  0.30  0.02
17:03:41 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   13    0.19  0.09  0.14  0.12  0.17  0.02
17:03:42 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   14    0.19  0.09  0.14  0.12  0.17  0.02
17:03:42 D523.1   D52310329613  walk    120  NoReport walk               trk  0.75 Fallen     2   14    0.04  0.10  0.15  0.11  0.29  0.02
17:03:43 D523.1   D52310329613  walk    105  NoReport walk               trk  0.75 Fallen     2   15    0.04  0.10  0.15  0.11  0.28  0.02
17:03:43 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   15    0.19  0.09  0.14  0.12  0.17  0.02
17:03:44 D523.1   D52310329613  walk    82   NoReport walk               trk  0.75 Fallen     2   16    0.05  0.10  0.15  0.12  0.27  0.02
17:03:44 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   16    0.19  0.09  0.14  0.12  0.17  0.02
17:03:44 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     2   17    0.19  0.09  0.13  0.12  0.17  0.02
17:03:45 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   17    0.19  0.09  0.13  0.11  0.17  0.02
17:03:45 D523.1   D52310329613  walk    107  NoReport walk               trk  0.75 Fallen     2   17    0.05  0.10  0.15  0.12  0.25  0.02
17:03:46 D523.1   D52310329613  walk    0    NoReport walk               trk  0.75 Fallen     2   18    0.05  0.10  0.15  0.12  0.25  0.02
17:03:46 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   18    0.19  0.09  0.13  0.11  0.17  0.02
17:03:47 D523.1   D52310329613  walk    0    NoReport walk               trk  0.75 Fallen     2   19    0.06  0.10  0.15  0.12  0.25  0.02
17:03:47 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   19    0.20  0.09  0.13  0.11  0.17  0.02
17:03:48 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   20    0.20  0.09  0.13  0.11  0.17  0.02
17:03:48 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   20    0.06  0.10  0.15  0.12  0.24  0.02
17:03:49 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   21    0.06  0.10  0.15  0.12  0.24  0.02
17:03:49 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   21    0.20  0.09  0.13  0.11  0.17  0.02
17:03:50 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   22    0.06  0.10  0.15  0.12  0.23  0.02
17:03:50 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   22    0.20  0.09  0.13  0.11  0.17  0.02
17:03:51 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   23    0.20  0.09  0.13  0.11  0.17  0.02
17:03:51 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   23    0.07  0.10  0.15  0.12  0.23  0.02
17:03:52 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   24    0.07  0.10  0.15  0.12  0.23  0.02
17:03:52 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   24    0.20  0.09  0.13  0.11  0.17  0.02
17:03:53 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   25    0.07  0.10  0.15  0.13  0.22  0.02
17:03:53 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   25    0.20  0.09  0.13  0.11  0.17  0.02
17:03:54 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   26    0.07  0.10  0.15  0.13  0.22  0.02
17:03:54 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   26    0.20  0.09  0.13  0.11  0.17  0.02
17:03:55 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   27    0.07  0.10  0.15  0.13  0.22  0.02
17:03:55 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   27    0.20  0.09  0.13  0.11  0.17  0.02
17:03:56 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   28    0.20  0.09  0.13  0.11  0.17  0.02
17:03:56 D523.1   D52310329613  stand   118  NoReport stand              trk  0.75 Fallen     2   28    0.08  0.11  0.15  0.13  0.22  0.02
17:03:57 D523.1   D52310329613  walk    116  NoReport walk               trk  0.75 Fallen     2   29    0.08  0.11  0.15  0.13  0.21  0.02
17:03:57 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   29    0.20  0.09  0.13  0.11  0.17  0.02
17:03:58 D523.1   D52310329613  walk    101  NoReport walk               trk  0.75 Fallen     2   30    0.08  0.11  0.15  0.13  0.21  0.02
17:03:58 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   30    0.20  0.09  0.13  0.11  0.17  0.02
17:03:59 D523.0   D52300400881  stand   0    NoReport stand              room -    Fallen     2   0     0.20  0.09  0.13  0.11  0.17  0.02
17:03:59 D523.1   D52310329613  walk    73   NoReport walk               trk  0.75 Fallen     2   0     0.08  0.11  0.15  0.13  0.21  0.02
17:04:00 D523.E   -             -       0    NoReport np=1               room -    Fallen     2   0     0.20  0.09  0.13  0.11  0.17  0.02
17:04:00 D523.1   D52310329613  walk    122  NoReport walk               trk  0.75 Fallen     2   0     0.09  0.11  0.15  0.13  0.21  0.02
17:04:00 D523.E   -             -       0    NoReport np=2               room -    Fallen     2   0     0.20  0.09  0.13  0.11  0.17  0.02
17:04:00 D523.E   -             -       0    NoReport EnterRoom(rdr)     room -    Fallen     2   0     0.20  0.09  0.13  0.11  0.17  0.02
17:04:00 D523.0   D52300400881  stand   0    NoReport stand              trk  1.00 Fallen     2   0     0.00  0.03  0.08  0.00  0.85  0.04
17:04:00 D523.1   D52310329613  walk    109  NoReport walk               trk  0.75 Fallen     2   0     0.09  0.11  0.15  0.13  0.21  0.02
17:04:01 D523.1   D52310329613  walk    0    NoReport walk               trk  0.75 Fallen     2   0     0.09  0.10  0.15  0.13  0.21  0.02
17:04:01 D523.0   D52300400881  stand   103  NoReport stand              trk  1.00 Fallen     2   0     0.00  0.04  0.10  0.00  0.82  0.01
17:04:02 D523.1   D52310329613  walk    0    NoReport walk               trk  0.75 Fallen     2   1     0.09  0.10  0.15  0.13  0.20  0.02
17:04:02 D523.0   D52300400881  stand   121  NoReport stand              trk  0.58 Fallen     2   1     0.00  0.04  0.11  0.01  0.76  0.01
17:04:03 D523.0   D52300400881  walk    105  NoReport walk               trk  0.84 Fallen     2   0     0.00  0.05  0.12  0.01  0.71  0.01
17:04:03 D523.1   D52310329613  walk    0    NoReport walk               trk  0.75 Fallen     2   0     0.09  0.10  0.15  0.13  0.20  0.02
17:04:04 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.10  0.10  0.15  0.13  0.20  0.02
17:04:04 D523.0   D52300400881  walk    99   NoReport walk               trk  0.71 Fallen     2   0     0.00  0.05  0.13  0.02  0.66  0.01
17:04:05 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.10  0.10  0.15  0.13  0.20  0.02
17:04:05 D523.0   D52300400881  walk    116  NoReport walk               trk  0.68 Fallen     2   0     0.00  0.06  0.14  0.03  0.62  0.01
17:04:06 D523.0   D52300400881  walk    97   NoReport walk               trk  0.68 Fallen     2   0     0.01  0.06  0.14  0.04  0.58  0.01
17:04:06 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.10  0.10  0.15  0.13  0.20  0.02
17:04:06 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     2   0     0.21  0.09  0.13  0.11  0.17  0.02
17:04:07 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.10  0.10  0.15  0.13  0.20  0.02
17:04:07 D523.0   D52300400881  walk    126  NoReport walk               trk  0.68 Fallen     2   0     0.01  0.07  0.14  0.04  0.54  0.01
17:04:08 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.10  0.10  0.15  0.13  0.20  0.02
17:04:08 D523.0   D52300400881  stand   0    NoReport stand              trk  0.68 Fallen     2   0     0.01  0.07  0.15  0.05  0.51  0.02
17:04:09 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.10  0.10  0.15  0.13  0.20  0.02
17:04:09 D523.0   D52300400881  stand   0    NoReport stand              trk  0.68 Fallen     2   0     0.01  0.07  0.15  0.06  0.49  0.02
17:04:10 D523.0   D52300400881  stand   121  NoReport stand              trk  0.68 Fallen     2   0     0.01  0.08  0.15  0.06  0.46  0.02
17:04:10 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.11  0.10  0.15  0.13  0.20  0.02
17:04:11 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.11  0.10  0.15  0.13  0.20  0.02
17:04:11 D523.0   D52300400881  stand   111  NoReport stand              trk  0.68 Fallen     2   0     0.01  0.08  0.15  0.07  0.44  0.02
17:04:12 D523.0   D52300400881  stand   85   NoReport stand              trk  0.68 Fallen     2   0     0.02  0.08  0.15  0.07  0.42  0.02
17:04:12 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.11  0.10  0.15  0.13  0.19  0.02
17:04:13 D523.1   D52310329613  stand   79   NoReport stand              trk  0.75 Fallen     2   0     0.11  0.10  0.15  0.13  0.19  0.02
17:04:13 D523.0   D52300400881  walk    0    NoReport walk               trk  0.68 Fallen     2   0     0.02  0.08  0.15  0.08  0.40  0.02
17:04:14 D523.1   D52310329613  stand   92   NoReport stand              trk  0.75 Fallen     2   0     0.11  0.10  0.15  0.13  0.19  0.02
17:04:14 D523.0   D52300400881  walk    0    NoReport walk               trk  0.68 Fallen     2   0     0.02  0.09  0.15  0.08  0.38  0.02
17:04:15 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.12  0.10  0.15  0.13  0.19  0.02
17:04:15 D523.0   D52300400881  walk    83   NoReport walk               trk  0.68 Fallen     2   0     0.02  0.09  0.15  0.09  0.36  0.02
17:04:16 D523.0   D52300400881  walk    88   NoReport walk               trk  0.68 Fallen     2   0     0.03  0.09  0.15  0.09  0.35  0.02
17:04:16 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.12  0.10  0.15  0.13  0.19  0.02
17:04:16 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     2   0     0.21  0.09  0.13  0.11  0.17  0.02
17:04:17 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.12  0.10  0.15  0.13  0.19  0.02
17:04:17 D523.0   D52300400881  walk    117  NoReport walk               trk  0.68 Fallen     2   0     0.03  0.09  0.15  0.10  0.33  0.02
17:04:18 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.12  0.10  0.15  0.13  0.19  0.02
17:04:18 D523.0   D52300400881  walk    106  NoReport walk               trk  0.68 Fallen     2   0     0.03  0.10  0.15  0.10  0.31  0.02
17:04:19 D523.1   D52310329613  stand   78   NoReport stand              trk  0.75 Fallen     2   0     0.12  0.10  0.14  0.13  0.19  0.02
17:04:19 D523.0   D52300400881  walk    0    NoReport walk               trk  0.68 Fallen     2   0     0.03  0.10  0.16  0.10  0.30  0.02
17:04:20 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.13  0.10  0.14  0.13  0.19  0.02
17:04:20 D523.0   D52300400881  walk    0    NoReport walk               trk  0.68 Fallen     2   0     0.04  0.10  0.16  0.11  0.30  0.02
17:04:21 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.13  0.10  0.14  0.13  0.19  0.02
17:04:21 D523.0   D52300400881  walk    107  NoReport walk               trk  0.68 Fallen     2   0     0.04  0.10  0.15  0.11  0.29  0.02
17:04:22 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.13  0.10  0.14  0.13  0.19  0.02
17:04:22 D523.0   D52300400881  walk    100  NoReport walk               trk  0.68 Fallen     2   0     0.04  0.10  0.15  0.11  0.28  0.02
17:04:23 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.13  0.10  0.14  0.13  0.19  0.02
17:04:23 D523.0   D52300400881  walk    104  NoReport walk               trk  0.68 Fallen     2   0     0.04  0.10  0.15  0.11  0.27  0.02
17:04:24 D523.0   D52300400881  walk    96   NoReport walk               trk  0.68 Fallen     2   0     0.05  0.10  0.15  0.12  0.27  0.02
17:04:24 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.13  0.10  0.14  0.12  0.19  0.02
17:04:25 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.13  0.10  0.14  0.12  0.19  0.02
17:04:25 D523.0   D52300400881  walk    100  NoReport walk               trk  0.68 Fallen     2   0     0.05  0.10  0.15  0.12  0.26  0.02
17:04:26 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.14  0.10  0.14  0.12  0.19  0.02
17:04:26 D523.0   D52300400881  walk    100  NoReport walk               trk  0.68 Fallen     2   0     0.05  0.10  0.15  0.12  0.25  0.02
17:04:27 D523.0   D52300400881  walk    81   NoReport walk               trk  0.68 Fallen     2   0     0.05  0.10  0.15  0.12  0.25  0.02
17:04:27 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.14  0.10  0.14  0.12  0.19  0.02
17:04:28 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.14  0.10  0.14  0.12  0.19  0.02
17:04:28 D523.0   D52300400881  walk    88   NoReport walk               trk  0.68 Fallen     2   0     0.06  0.10  0.15  0.12  0.25  0.02
17:04:29 D523.0   D52300400881  walk    105  NoReport walk               trk  0.68 Fallen     2   0     0.06  0.10  0.15  0.12  0.24  0.02
17:04:29 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.14  0.10  0.14  0.12  0.18  0.02
17:04:30 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.14  0.10  0.14  0.12  0.18  0.02
17:04:30 D523.0   D52300400881  walk    128  NoReport walk               trk  0.68 Fallen     2   0     0.06  0.10  0.15  0.12  0.24  0.02
17:04:31 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     2   0     0.14  0.10  0.14  0.12  0.18  0.02
17:04:31 D523.0   D52300400881  walk    108  NoReport walk               trk  0.68 Fallen     2   0     0.06  0.10  0.15  0.12  0.23  0.02
17:04:31 09E7.E   -             -       0    NoReport np=1               room -    Fallen     2   0     0.22  0.09  0.13  0.11  0.16  0.02
17:04:31 09E7.E   -             -       0    NoReport EnterRoom(rdr)     room -    Fallen     2   0     0.22  0.09  0.13  0.11  0.16  0.02
17:04:31 09E7.0   09E700431918  stand   56   NoReport stand              trk  0.96 Fallen     3   0     0.00  0.03  0.15  0.00  0.79  0.04
17:04:32 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   0     0.15  0.10  0.14  0.12  0.18  0.02
17:04:32 D523.0   D52300400881  walk    113  NoReport walk               trk  0.68 Fallen     3   0     0.07  0.10  0.15  0.13  0.22  0.02
17:04:32 09E7.0   09E700431918  stand   0    NoReport stand              trk  0.96 Fallen     3   0     0.00  0.04  0.46  0.00  0.38  0.02
17:04:33 D523.0   D52300400881  walk    0    NoReport walk               trk  0.68 Fallen     3   0     0.07  0.10  0.15  0.13  0.22  0.02
17:04:33 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   0     0.15  0.10  0.14  0.12  0.18  0.02
17:04:33 09E7.0   09E700431918  stand   0    NoReport stand              trk  0.96 Fallen     3   0     0.00  0.04  0.64  0.01  0.12  0.03
17:04:34 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   0     0.15  0.10  0.14  0.12  0.18  0.02
17:04:34 D523.0   D52300400881  walk    119  NoReport walk               trk  0.68 Fallen     3   0     0.08  0.11  0.15  0.13  0.21  0.02
17:04:34 09E7.0   09E700431918  stand   65   NoReport stand              trk  0.96 Fallen     3   0     0.01  0.04  0.70  0.01  0.05  0.03
17:04:35 D523.0   D52300400881  walk    120  NoReport walk               trk  0.68 Fallen     3   0     0.08  0.11  0.15  0.13  0.21  0.02
17:04:35 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   0     0.15  0.10  0.14  0.12  0.18  0.02
17:04:35 09E7.0   09E700431918  stand   0    NoReport stand              trk  0.96 Fallen     3   10    0.01  0.04  0.71  0.01  0.03  0.04
17:04:36 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   10    0.16  0.10  0.14  0.12  0.18  0.02
17:04:36 D523.0   D52300400881  walk    120  NoReport walk               trk  0.68 Fallen     3   10    0.09  0.11  0.15  0.13  0.21  0.02
17:04:36 09E7.0   09E700431918  stand   0    NoReport stand              trk  0.96 Fallen     3   11    0.01  0.04  0.71  0.01  0.02  0.04
17:04:37 D523.0   D52300400881  walk    78   NoReport walk               trk  0.68 Fallen     3   11    0.09  0.10  0.15  0.13  0.20  0.02
17:04:37 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   11    0.16  0.10  0.14  0.12  0.18  0.02
17:04:37 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   12    0.01  0.04  0.71  0.01  0.02  0.04
17:04:38 D523.0   D52300400881  walk    79   NoReport walk               trk  0.68 Fallen     3   12    0.10  0.10  0.15  0.13  0.20  0.02
17:04:38 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   12    0.16  0.10  0.14  0.12  0.18  0.02
17:04:38 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   13    0.01  0.04  0.71  0.01  0.02  0.04
17:04:38 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   13    0.16  0.10  0.14  0.12  0.18  0.02
17:04:38 D523.0   D52300400881  walk    116  NoReport walk               trk  0.68 Fallen     3   13    0.10  0.10  0.15  0.13  0.20  0.02
17:04:39 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     3   13    0.23  0.09  0.13  0.11  0.16  0.02
17:04:39 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   14    0.01  0.04  0.71  0.01  0.02  0.04
17:04:39 D523.0   D52300400881  walk    111  NoReport walk               trk  0.68 Fallen     3   14    0.10  0.10  0.15  0.13  0.20  0.02
17:04:39 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   14    0.17  0.10  0.14  0.12  0.18  0.02
17:04:40 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   15    0.01  0.04  0.71  0.01  0.02  0.04
17:04:40 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   15    0.17  0.10  0.14  0.12  0.18  0.02
17:04:40 D523.0   D52300400881  walk    101  NoReport walk               trk  0.68 Fallen     3   15    0.11  0.10  0.15  0.13  0.20  0.02
17:04:41 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   16    0.01  0.04  0.71  0.01  0.02  0.04
17:04:41 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   16    0.17  0.10  0.14  0.12  0.18  0.02
17:04:41 D523.0   D52300400881  walk    92   NoReport walk               trk  0.68 Fallen     3   16    0.11  0.10  0.15  0.13  0.19  0.02
17:04:42 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   17    0.01  0.04  0.71  0.01  0.02  0.04
17:04:43 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   17    0.17  0.10  0.14  0.12  0.18  0.02
17:04:43 D523.0   D52300400881  walk    90   NoReport walk               trk  0.68 Fallen     3   17    0.11  0.10  0.15  0.13  0.19  0.02
17:04:43 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   18    0.01  0.04  0.71  0.01  0.02  0.04
17:04:44 D523.0   D52300400881  walk    83   NoReport walk               trk  0.68 Fallen     3   18    0.12  0.10  0.15  0.13  0.19  0.02
17:04:44 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   18    0.17  0.10  0.14  0.12  0.18  0.02
17:04:44 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   19    0.01  0.04  0.70  0.01  0.03  0.04
17:04:45 D523.0   D52300400881  walk    114  NoReport walk               trk  0.68 Fallen     3   19    0.12  0.10  0.15  0.13  0.19  0.02
17:04:45 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   19    0.18  0.10  0.14  0.12  0.18  0.02
17:04:45 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   20    0.01  0.04  0.71  0.01  0.02  0.04
17:04:46 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   20    0.18  0.10  0.14  0.12  0.17  0.02
17:04:46 D523.0   D52300400881  walk    0    NoReport walk               trk  0.68 Fallen     3   20    0.13  0.10  0.14  0.13  0.19  0.02
17:04:46 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   21    0.01  0.04  0.71  0.01  0.02  0.04
17:04:47 D523.0   D52300400881  stand   0    NoReport stand              trk  0.68 Fallen     3   21    0.13  0.10  0.14  0.13  0.19  0.02
17:04:47 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   21    0.18  0.10  0.14  0.12  0.17  0.02
17:04:47 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   22    0.01  0.04  0.71  0.01  0.02  0.04
17:04:47 D523.0   D52300400881  stand   78   NoReport stand              trk  0.68 Fallen     3   22    0.13  0.10  0.14  0.12  0.19  0.02
17:04:47 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   22    0.18  0.10  0.14  0.12  0.17  0.02
17:04:48 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   23    0.01  0.04  0.71  0.01  0.02  0.04
17:04:48 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   23    0.18  0.10  0.14  0.12  0.17  0.02
17:04:48 D523.0   D52300400881  stand   96   NoReport stand              trk  0.68 Fallen     3   23    0.14  0.10  0.14  0.12  0.19  0.02
17:04:49 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   24    0.01  0.04  0.71  0.01  0.02  0.04
17:04:49 D523.0   D52300400881  walk    116  NoReport walk               trk  0.68 Fallen     3   24    0.14  0.10  0.14  0.12  0.19  0.02
17:04:49 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   24    0.19  0.09  0.14  0.12  0.17  0.02
17:04:50 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   25    0.01  0.04  0.71  0.01  0.02  0.04
17:04:50 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   25    0.19  0.09  0.14  0.12  0.17  0.02
17:04:50 D523.0   D52300400881  walk    0    NoReport walk               trk  0.68 Fallen     3   25    0.14  0.10  0.14  0.12  0.18  0.02
17:04:51 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   26    0.01  0.04  0.71  0.01  0.02  0.04
17:04:51 D523.0   D52300400881  walk    0    NoReport walk               trk  0.68 Fallen     3   26    0.14  0.10  0.14  0.12  0.18  0.02
17:04:51 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   26    0.19  0.09  0.14  0.12  0.17  0.02
17:04:52 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   27    0.01  0.04  0.71  0.01  0.02  0.04
17:04:52 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   27    0.19  0.09  0.14  0.12  0.17  0.02
17:04:52 D523.0   D52300400881  stand   0    NoReport stand              trk  0.68 Fallen     3   27    0.15  0.10  0.14  0.12  0.18  0.02
17:04:53 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   28    0.00  0.02  0.83  0.00  0.01  0.02
17:04:53 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   28    0.19  0.09  0.14  0.12  0.17  0.02
17:04:53 D523.0   D52300400881  stand   0    NoReport stand              trk  0.68 Fallen     3   28    0.15  0.10  0.14  0.12  0.18  0.02
17:04:54 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   29    0.00  0.02  0.88  0.00  0.00  0.02
17:04:54 D523.0   D52300400881  stand   0    NoReport stand              trk  0.68 Fallen     3   29    0.15  0.10  0.14  0.12  0.18  0.02
17:04:54 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   29    0.19  0.09  0.13  0.12  0.17  0.02
17:04:55 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   30    0.00  0.02  0.88  0.00  0.00  0.02
17:04:55 D523.0   D52300400881  stand   0    NoReport stand              trk  0.68 Fallen     3   30    0.15  0.10  0.14  0.12  0.18  0.02
17:04:55 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   30    0.19  0.09  0.13  0.11  0.17  0.02
17:04:55 D523.E   -             -       0    NoReport ExitRoom(rdr)      room -    Fallen     3   30    0.23  0.09  0.13  0.11  0.16  0.02
17:04:56 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     3   31    0.00  0.02  0.88  0.00  0.00  0.02
17:04:56 D523.E   -             -       0    NoReport np=1               room -    Fallen     3   31    0.23  0.09  0.13  0.11  0.16  0.02
17:04:56 D523.1   D52310329613  stand   0    NoReport stand              trk  0.75 Fallen     3   31    0.20  0.09  0.13  0.11  0.17  0.02
17:04:57 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   32    0.00  0.02  0.85  0.00  0.01  0.02
17:04:57 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   32    0.20  0.09  0.13  0.11  0.17  0.02
17:04:58 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   33    0.00  0.02  0.85  0.00  0.01  0.02
17:04:58 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   33    0.20  0.09  0.13  0.11  0.17  0.02
17:04:59 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   34    0.00  0.02  0.85  0.00  0.01  0.02
17:04:59 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   34    0.20  0.09  0.13  0.11  0.17  0.02
17:05:00 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   35    0.00  0.02  0.85  0.00  0.01  0.02
17:05:00 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   35    0.20  0.09  0.13  0.11  0.17  0.02
17:05:01 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   36    0.00  0.04  0.74  0.00  0.02  0.04
17:05:01 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   36    0.20  0.09  0.13  0.11  0.17  0.02
17:05:02 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   37    0.01  0.05  0.65  0.01  0.04  0.03
17:05:02 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   37    0.20  0.09  0.13  0.11  0.17  0.02
17:05:03 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   38    0.01  0.05  0.62  0.02  0.05  0.03
17:05:03 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   38    0.21  0.09  0.13  0.11  0.17  0.02
17:05:04 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   39    0.01  0.05  0.62  0.02  0.05  0.03
17:05:04 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   39    0.21  0.09  0.13  0.11  0.17  0.02
17:05:05 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   40    0.01  0.05  0.61  0.02  0.05  0.03
17:05:05 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   40    0.21  0.09  0.13  0.11  0.17  0.02
17:05:06 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   41    0.01  0.05  0.61  0.02  0.05  0.03
17:05:06 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   41    0.21  0.09  0.13  0.11  0.17  0.02
17:05:07 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   42    0.01  0.05  0.61  0.02  0.05  0.03
17:05:07 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   42    0.21  0.09  0.13  0.11  0.17  0.02
17:05:08 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   43    0.01  0.05  0.61  0.02  0.05  0.03
17:05:08 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   43    0.21  0.09  0.13  0.11  0.17  0.02
17:05:09 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   44    0.01  0.05  0.61  0.02  0.05  0.03
17:05:09 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   44    0.21  0.09  0.13  0.11  0.17  0.02
17:05:10 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   45    0.01  0.05  0.61  0.02  0.05  0.03
17:05:10 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     2   45    0.24  0.09  0.13  0.11  0.16  0.02
17:05:10 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   45    0.21  0.09  0.13  0.11  0.17  0.02
17:05:11 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   46    0.01  0.05  0.61  0.02  0.05  0.03
17:05:11 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   46    0.21  0.09  0.13  0.11  0.17  0.02
17:05:12 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   47    0.01  0.05  0.61  0.02  0.05  0.03
17:05:12 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   47    0.21  0.09  0.13  0.11  0.17  0.02
17:05:13 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   48    0.01  0.05  0.61  0.02  0.05  0.03
17:05:13 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   48    0.21  0.09  0.13  0.11  0.18  0.03
17:05:14 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   49    0.01  0.05  0.61  0.02  0.05  0.03
17:05:14 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   49    0.20  0.09  0.13  0.11  0.19  0.03
17:05:15 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   50    0.01  0.05  0.61  0.02  0.05  0.03
17:05:15 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   50    0.20  0.08  0.12  0.10  0.21  0.04
17:05:16 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   51    0.01  0.05  0.61  0.02  0.05  0.03
17:05:16 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   51    0.19  0.08  0.12  0.10  0.23  0.05
17:05:17 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   52    0.01  0.05  0.61  0.02  0.05  0.03
17:05:17 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   52    0.17  0.08  0.12  0.09  0.26  0.06
17:05:18 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   53    0.01  0.05  0.61  0.02  0.05  0.03
17:05:18 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   53    0.16  0.07  0.11  0.08  0.29  0.07
17:05:19 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   54    0.01  0.05  0.61  0.02  0.05  0.03
17:05:19 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   54    0.13  0.06  0.11  0.07  0.35  0.09
17:05:20 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   55    0.01  0.05  0.61  0.02  0.05  0.03
17:05:20 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   55    0.11  0.06  0.10  0.06  0.39  0.12
17:05:21 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   56    0.01  0.05  0.61  0.02  0.05  0.03
17:05:21 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   56    0.09  0.05  0.09  0.05  0.43  0.15
17:05:22 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   57    0.01  0.05  0.61  0.02  0.05  0.03
17:05:22 D523.1   D52310329613  stand   0    NoReport stand              trk  1.00 Fallen     2   57    0.06  0.04  0.08  0.04  0.46  0.20
17:05:23 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     2   58    0.01  0.05  0.61  0.02  0.05  0.03
17:05:23 D523.E   -             -       0    NoReport np=0  ★0           room -    Fallen     2   58    0.24  0.09  0.13  0.11  0.16  0.02
17:05:23 D523.88  -             88      -    NoReport no-target(88)      room -    Fallen     2   58    0.24  0.09  0.13  0.11  0.16  0.02
17:05:24 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   52    0.01  0.05  0.61  0.02  0.05  0.03
17:05:24 D523.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   52    0.24  0.09  0.13  0.11  0.16  0.02
17:05:25 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   53    0.01  0.05  0.61  0.02  0.05  0.03
17:05:25 D523.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   53    0.24  0.09  0.13  0.11  0.16  0.02
17:05:26 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   54    0.01  0.05  0.61  0.02  0.05  0.03
17:05:27 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   55    0.01  0.05  0.61  0.02  0.05  0.03
17:05:28 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   56    0.01  0.05  0.61  0.02  0.05  0.03
17:05:29 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   57    0.01  0.05  0.61  0.02  0.05  0.03
17:05:30 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   58    0.01  0.05  0.61  0.02  0.05  0.03
17:05:31 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   59    0.01  0.05  0.61  0.02  0.05  0.03
17:05:32 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   60    0.01  0.05  0.61  0.02  0.05  0.03
17:05:33 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   61    0.01  0.05  0.61  0.02  0.05  0.03
17:05:33 D523.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   61    0.24  0.09  0.13  0.11  0.16  0.01
17:05:34 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   62    0.01  0.05  0.61  0.02  0.05  0.03
17:05:35 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   63    0.01  0.05  0.61  0.02  0.05  0.03
17:05:36 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   64    0.01  0.05  0.61  0.02  0.05  0.03
17:05:37 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   65    0.01  0.05  0.61  0.02  0.05  0.03
17:05:38 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   66    0.01  0.05  0.61  0.02  0.05  0.03
17:05:39 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   67    0.01  0.05  0.61  0.02  0.05  0.03
17:05:40 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   68    0.01  0.05  0.61  0.02  0.05  0.03
17:05:41 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   69    0.01  0.05  0.61  0.02  0.05  0.03
17:05:42 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   70    0.01  0.05  0.61  0.02  0.05  0.03
17:05:42 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   70    0.24  0.09  0.13  0.11  0.16  0.01
17:05:43 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   71    0.01  0.05  0.61  0.02  0.05  0.03
17:05:44 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   72    0.01  0.05  0.61  0.02  0.05  0.03
17:05:45 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   73    0.01  0.05  0.61  0.02  0.05  0.03
17:05:45 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   74    0.01  0.05  0.61  0.02  0.05  0.03
17:05:46 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   75    0.01  0.05  0.61  0.02  0.05  0.03
17:05:47 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   76    0.01  0.05  0.61  0.02  0.05  0.03
17:05:48 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   77    0.01  0.05  0.61  0.02  0.05  0.03
17:05:49 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   78    0.01  0.05  0.61  0.02  0.05  0.03
17:05:50 -.-      -             -       -    NoReport (no frame, held)   room -    Fallen     1   78    0.24  0.09  0.13  0.11  0.16  0.01
17:05:51 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   79    0.01  0.05  0.61  0.02  0.05  0.03
17:05:52 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   80    0.01  0.05  0.61  0.02  0.05  0.03
17:05:53 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   81    0.01  0.05  0.61  0.02  0.05  0.03
17:05:54 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   82    0.01  0.05  0.61  0.02  0.05  0.03
17:05:54 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   82    0.01  0.05  0.61  0.02  0.05  0.03
17:05:55 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   83    0.01  0.05  0.61  0.02  0.05  0.03
17:05:56 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   84    0.01  0.05  0.61  0.02  0.05  0.03
17:05:57 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   85    0.01  0.05  0.61  0.02  0.05  0.03
17:05:58 09E7.0   09E700431918  stand   0    NoReport stand              trk  1.00 Fallen     1   86    0.01  0.05  0.61  0.02  0.05  0.03
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
17:02:09.714 09E7.88   88     -    -      -      -     -    -    
17:02:41.767 09E7.88   88     -    -      -      -     -    -    
17:03:13.201 09E7.88   88     -    -      -      -     -    -    
17:03:27.552 09E7.0    stand  0    260    -150   61    80        
17:03:28.091 09E7.0    stand  0    280    -110   0     80   44   
17:03:29.090 09E7.0    stand  0    300    -20    0     80   92   
17:03:30.088 09E7.0    stand  0    310    100    0     80   120  
17:03:31.085 09E7.0    stand  0    310    110    0     80   10   
17:03:32.089 09E7.0    stand  0    310    110    0     80   0    
17:03:33.088 09E7.0    stand  0    310    100    0     80   10   
17:03:34.089 09E7.0    stand  0    310    100    0     80   0    
17:03:35.096 09E7.0    stand  0    310    100    0     80   0    
17:03:36.157 09E7.88   88     -    -      -      -     -    -    
17:03:37.101 09E7.88   88     -    -      -      -     -    -    
17:03:38.108 09E7.88   88     -    -      -      -     -    -    
17:03:44.948 09E7.88   88     -    -      -      -     -    -    
17:04:16.756 09E7.88   88     -    -      -      -     -    -    
17:04:31.918 09E7.0    stand  0    260    -150   56    80   254  
17:04:32.630 09E7.0    stand  0    260    -130   0     80   20   
17:04:33.635 09E7.0    stand  0    260    -130   0     80   0    
17:04:34.633 09E7.0    stand  0    260    -130   65    80   0    
17:04:35.637 09E7.0    stand  0    250    -160   0     80   31   
17:04:36.534 09E7.0    stand  0    240    -160   0     80   10   
17:04:37.539 09E7.0    stand  0    240    -160   0     80   0    
17:04:38.538 09E7.0    stand  0    240    -160   0     80   0    
17:04:39.538 09E7.0    stand  0    240    -160   0     80   0    
17:04:40.539 09E7.0    stand  0    240    -160   0     80   0    
17:04:41.538 09E7.0    stand  0    240    -160   0     80   0    
17:04:42.540 09E7.0    stand  0    240    -160   0     80   0    
17:04:43.558 09E7.0    stand  0    240    -160   0     80   0    
17:04:44.539 09E7.0    stand  0    240    -160   0     80   0    
17:04:45.565 09E7.0    stand  0    240    -160   0     80   0    
17:04:46.458 09E7.0    stand  0    240    -160   0     80   0    
17:04:47.452 09E7.0    stand  0    240    -160   0     80   0    
17:04:48.453 09E7.0    stand  0    240    -160   0     80   0    
17:04:49.457 09E7.0    stand  0    240    -160   0     80   0    
17:04:50.464 09E7.0    stand  0    240    -160   0     80   0    
17:04:51.456 09E7.0    stand  0    240    -160   0     80   0    
17:04:52.540 09E7.0    stand  0    240    -160   0     80   0    
17:04:53.457 09E7.0    stand  0    270    -160   0     80   30   
17:04:54.460 09E7.0    stand  0    280    -150   0     80   14   
17:04:55.461 09E7.0    stand  0    280    -150   0     80   0    
17:04:56.462 09E7.0    stand  0    280    -150   0     80   0    
17:04:57.465 09E7.0    stand  0    280    -150   0     80   0    
17:04:58.356 09E7.0    stand  0    280    -150   0     80   0    
17:04:59.362 09E7.0    stand  0    280    -150   0     80   0    
17:05:00.360 09E7.0    stand  0    270    -150   0     80   10   
17:05:01.372 09E7.0    stand  0    260    -130   0     80   22   
17:05:02.376 09E7.0    stand  0    260    -130   0     80   0    
17:05:03.384 09E7.0    stand  0    260    -130   0     80   0    
17:05:04.382 09E7.0    stand  0    260    -130   0     80   0    
17:05:05.381 09E7.0    stand  0    260    -130   0     80   0    
17:05:06.381 09E7.0    stand  0    260    -130   0     80   0    
17:05:07.380 09E7.0    stand  0    260    -130   0     80   0    
17:05:08.278 09E7.0    stand  0    260    -130   0     80   0    
17:05:09.276 09E7.0    stand  0    260    -130   0     80   0    
17:05:10.291 09E7.0    stand  0    260    -130   0     80   0    
17:05:11.278 09E7.0    stand  0    260    -130   0     80   0    
17:05:12.281 09E7.0    stand  0    260    -130   0     80   0    
17:05:13.278 09E7.0    stand  0    260    -130   0     80   0    
17:05:14.281 09E7.0    stand  0    260    -130   0     80   0    
17:05:15.281 09E7.0    stand  0    260    -130   0     80   0    
17:05:16.288 09E7.0    stand  0    260    -130   0     80   0    
17:05:17.290 09E7.0    stand  0    260    -130   0     80   0    
17:05:18.198 09E7.0    stand  0    260    -130   0     80   0    
17:05:19.256 09E7.0    stand  0    260    -130   0     80   0    
17:05:20.198 09E7.0    stand  0    260    -130   0     80   0    
17:05:21.202 09E7.0    stand  0    260    -130   0     80   0    
17:05:22.204 09E7.0    stand  0    260    -130   0     80   0    
17:05:23.202 09E7.0    stand  0    260    -130   0     80   0    
17:05:24.201 09E7.0    stand  0    260    -130   0     80   0    
17:05:25.203 09E7.0    stand  0    260    -130   0     80   0    
17:05:26.204 09E7.0    stand  0    260    -130   0     80   0    
17:05:27.210 09E7.0    stand  0    260    -130   0     80   0    
17:05:28.205 09E7.0    stand  0    260    -130   0     80   0    
17:05:29.208 09E7.0    stand  0    260    -130   0     80   0    
17:05:30.099 09E7.0    stand  0    260    -130   0     80   0    
17:05:31.102 09E7.0    stand  0    260    -130   0     80   0    
17:05:32.102 09E7.0    stand  0    260    -130   0     80   0    
17:05:33.102 09E7.0    stand  0    260    -130   0     80   0    
17:05:34.066 09E7.0    stand  0    260    -130   0     80   0    
17:05:35.071 09E7.0    stand  0    260    -130   0     80   0    
17:05:36.071 09E7.0    stand  0    260    -130   0     80   0    
17:05:37.077 09E7.0    stand  0    260    -130   0     80   0    
17:05:38.075 09E7.0    stand  0    260    -130   0     80   0    
17:05:39.072 09E7.0    stand  0    260    -130   0     80   0    
17:05:40.081 09E7.0    stand  0    260    -130   0     80   0    
17:05:41.074 09E7.0    stand  0    260    -130   0     80   0    
17:05:42.078 09E7.0    stand  0    260    -130   0     80   0    
17:05:43.076 09E7.0    stand  0    260    -130   0     80   0    
17:05:44.083 09E7.0    stand  0    260    -130   0     80   0    
17:05:45.080 09E7.0    stand  0    260    -130   0     80   0    
17:05:45.971 09E7.0    stand  0    260    -130   0     80   0    
17:05:46.973 09E7.0    stand  0    260    -130   0     80   0    
17:05:47.974 09E7.0    stand  0    260    -130   0     80   0    
17:05:48.974 09E7.0    stand  0    260    -130   0     80   0    
17:05:49.999 09E7.0    stand  0    260    -130   0     80   0    
17:05:51.003 09E7.0    stand  0    260    -130   0     80   0    
17:05:52.006 09E7.0    stand  0    260    -130   0     80   0    
17:05:53.004 09E7.0    stand  0    260    -130   0     80   0    
17:05:54.008 09E7.0    stand  0    260    -130   0     80   0    
17:05:54.899 09E7.0    stand  0    260    -130   0     80   0    
17:05:55.900 09E7.0    stand  0    260    -130   0     80   0    
17:05:56.909 09E7.0    stand  0    260    -130   0     80   0    
17:05:57.901 09E7.0    stand  0    260    -130   0     80   0    
17:05:58.909 09E7.0    stand  0    260    -130   0     80   0    

17:02:00.296 D523.0    stand  255  -260   500    0     80        
17:02:01.302 D523.0    stand  255  -270   500    0     80   10   
17:02:02.288 D523.0    stand  255  -270   500    0     80   0    
17:02:03.301 D523.0    stand  255  -270   500    0     80   0    
17:02:04.297 D523.0    stand  255  -270   500    0     80   0    
17:02:05.297 D523.0    stand  255  -270   500    0     80   0    
17:02:06.312 D523.0    stand  255  -270   500    0     80   0    
17:02:07.301 D523.0    stand  255  -270   500    0     80   0    
17:02:08.194 D523.0    stand  255  -270   500    0     80   0    
17:02:09.197 D523.0    stand  255  -270   500    0     80   0    
17:02:10.204 D523.0    stand  255  -270   500    0     80   0    
17:02:11.200 D523.0    stand  255  -270   500    0     80   0    
17:02:12.199 D523.0    stand  255  -270   500    0     80   0    
17:02:13.204 D523.0    stand  255  -270   500    0     80   0    
17:02:14.201 D523.0    stand  255  -270   500    0     80   0    
17:02:15.201 D523.0    stand  255  -270   500    0     80   0    
17:02:16.205 D523.0    stand  255  -270   500    0     80   0    
17:02:17.202 D523.0    stand  255  -270   500    0     80   0    
17:02:18.204 D523.0    stand  255  -270   500    0     80   0    
17:02:19.216 D523.0    stand  255  -270   500    0     80   0    
17:02:20.106 D523.0    stand  255  -270   500    0     80   0    
17:02:21.101 D523.0    stand  255  -270   500    0     80   0    
17:02:22.100 D523.0    stand  255  -280   500    85    80   10   
17:02:23.108 D523.0    stand  255  -290   480    0     80   22   
17:02:24.108 D523.0    stand  255  -290   480    0     80   0    
17:02:25.109 D523.0    stand  255  -280   500    94    80   22   
17:02:26.106 D523.0    stand  255  -280   500    0     80   0    
17:02:27.109 D523.0    stand  255  -280   500    0     80   0    
17:02:28.109 D523.0    stand  255  -280   500    0     80   0    
17:02:29.108 D523.0    stand  255  -280   500    0     80   0    
17:02:30.108 D523.0    stand  255  -280   500    0     80   0    
17:02:31.112 D523.0    stand  255  -280   500    0     80   0    
17:02:32.005 D523.0    stand  255  -280   500    0     80   0    
17:02:33.008 D523.0    stand  255  -280   500    0     80   0    
17:02:34.008 D523.0    stand  255  -280   500    0     80   0    
17:02:35.010 D523.0    stand  255  -280   500    0     80   0    
17:02:36.008 D523.0    stand  255  -280   500    87    80   0    
17:02:37.008 D523.0    stand  255  -270   500    0     80   10   
17:02:38.019 D523.0    stand  255  -270   500    0     80   0    
17:02:39.012 D523.0    stand  255  -270   500    0     80   0    
17:02:40.014 D523.0    stand  255  -270   500    0     80   0    
17:02:41.014 D523.0    stand  255  -270   500    0     80   0    
17:02:42.013 D523.0    stand  255  -270   510    0     80   10   
17:02:43.021 D523.0    stand  255  -270   510    0     80   0    
17:02:43.910 D523.0    stand  255  -270   510    0     80   0    
17:02:44.918 D523.0    stand  255  -280   500    0     80   14   
17:02:45.958 D523.0    stand  255  -280   500    0     80   0    
17:02:46.908 D523.0    stand  255  -280   500    0     80   0    
17:02:47.912 D523.0    stand  255  -280   500    0     80   0    
17:02:48.910 D523.0    stand  255  -270   510    0     80   14   
17:02:49.912 D523.0    stand  255  -270   510    0     80   0    
17:02:50.934 D523.0    stand  255  -270   510    0     80   0    
17:02:51.932 D523.0    stand  255  -270   510    0     80   0    
17:02:52.936 D523.0    stand  255  -270   510    0     80   0    
17:02:53.826 D523.0    stand  255  -270   510    0     80   0    
17:02:54.828 D523.0    stand  255  -270   510    0     80   0    
17:02:55.833 D523.0    stand  255  -270   510    0     80   0    
17:02:56.835 D523.0    stand  255  -270   510    0     80   0    
17:02:57.832 D523.0    stand  255  -270   510    0     80   0    
17:02:58.830 D523.0    stand  255  -270   510    0     80   0    
17:02:59.831 D523.0    stand  255  -270   510    0     80   0    
17:03:00.834 D523.0    stand  255  -270   510    0     80   0    
17:03:01.840 D523.0    stand  255  -270   510    0     80   0    
17:03:02.836 D523.0    stand  255  -270   510    0     80   0    
17:03:03.837 D523.0    stand  255  -270   510    0     80   0    
17:03:04.840 D523.0    stand  255  -270   510    0     80   0    
17:03:05.729 D523.0    stand  255  -270   510    0     80   0    
17:03:06.740 D523.0    stand  255  -270   510    0     80   0    
17:03:07.742 D523.0    stand  255  -270   510    0     80   0    
17:03:08.743 D523.0    stand  255  -270   510    0     80   0    
17:03:09.743 D523.0    stand  255  -270   510    0     80   0    
17:03:10.746 D523.0    stand  255  -270   510    0     80   0    
17:03:11.749 D523.0    stand  255  -270   510    0     80   0    
17:03:12.753 D523.0    stand  255  -270   510    0     80   0    
17:03:13.756 D523.0    stand  255  -270   500    0     80   10   
17:03:14.748 D523.0    stand  255  -270   510    0     80   10   
17:03:15.758 D523.0    stand  255  -260   510    0     80   10   
17:03:16.648 D523.0    stand  255  -260   510    0     80   0    
17:03:17.645 D523.0    stand  255  -270   510    0     80   10   
17:03:18.646 D523.0    stand  255  -270   510    0     80   0    
17:03:19.647 D523.0    stand  255  -270   510    0     80   0    
17:03:20.656 D523.0    stand  255  -270   500    0     80   10   
17:03:21.648 D523.0    stand  255  -290   490    0     80   22   
17:03:22.664 D523.0    stand  255  -290   490    0     80   0    
17:03:23.651 D523.0    stand  255  -290   490    0     80   0    
17:03:24.652 D523.0    stand  255  -290   490    0     80   0    
17:03:25.657 D523.0    stand  255  -290   490    0     80   0    
17:03:26.657 D523.0    stand  255  -260   480    71    80   31   
17:03:27.658 D523.0    stand  255  -290   490    102   80   31   
17:03:28.559 D523.0    stand  255  -310   460    103   80   36   
17:03:29.613 D523.1    stand  255  -300   310    100   80   150  
17:03:29.613 D523.0    walk   255  -320   450    0     80   141  
17:03:30.592 D523.1    walk   255  -350   190    106   80   261  
17:03:30.592 D523.0    stand  255  -320   440    0     80   251  
17:03:31.564 D523.0    stand  255  -320   450    0     80   10   
17:03:31.564 D523.1    walk   255  -400   100    100   80   359  
17:03:32.561 D523.0    stand  255  -320   450    0     80   359  
17:03:32.561 D523.1    walk   255  -490   90     117   80   398  
17:03:33.567 D523.1    walk   255  -520   110    0     80   36   
17:03:33.567 D523.0    stand  255  -310   450    0     80   399  
17:03:34.564 D523.1    walk   255  -520   120    0     80   391  
17:03:34.564 D523.0    stand  255  -310   450    0     80   391  
17:03:35.568 D523.0    stand  255  -310   450    0     80   0    
17:03:35.568 D523.1    stand  255  -520   120    0     80   391  
17:03:36.569 D523.0    stand  255  -310   450    0     80   391  
17:03:36.569 D523.1    stand  255  -510   120    0     80   385  
17:03:37.575 D523.1    stand  255  -510   120    0     80   0    
17:03:37.575 D523.0    stand  255  -310   450    0     80   385  
17:03:38.578 D523.1    stand  255  -510   120    0     80   385  
17:03:38.578 D523.0    stand  255  -310   450    0     80   385  
17:03:39.460 D523.0    stand  255  -310   450    0     80   0    
17:03:39.460 D523.1    stand  255  -510   120    0     80   385  
17:03:40.465 D523.1    stand  255  -440   120    116   80   70   
17:03:40.465 D523.0    stand  255  -310   450    0     80   354  
17:03:41.463 D523.1    walk   255  -360   150    109   80   304  
17:03:41.463 D523.0    stand  255  -310   450    0     80   304  
17:03:42.476 D523.0    stand  255  -310   450    0     80   0    
17:03:42.476 D523.1    walk   255  -350   210    120   80   243  
17:03:43.464 D523.1    walk   255  -330   170    105   80   44   
17:03:43.464 D523.0    stand  255  -310   450    0     80   280  
17:03:44.528 D523.1    walk   255  -410   100    82    80   364  
17:03:44.528 D523.0    stand  255  -310   450    0     80   364  
17:03:45.468 D523.0    stand  255  -310   450    0     80   0    
17:03:45.468 D523.1    walk   255  -510   130    107   80   377  
17:03:46.467 D523.1    walk   255  -510   140    0     80   10   
17:03:46.467 D523.0    stand  255  -310   450    0     80   368  
17:03:47.493 D523.1    walk   255  -500   140    0     80   363  
17:03:47.493 D523.0    stand  255  -310   450    0     80   363  
17:03:48.470 D523.0    stand  255  -310   450    0     80   0    
17:03:48.470 D523.1    stand  255  -500   140    0     80   363  
17:03:49.479 D523.1    stand  255  -500   140    0     80   0    
17:03:49.479 D523.0    stand  255  -310   450    0     80   363  
17:03:50.478 D523.1    stand  255  -500   140    0     80   363  
17:03:50.478 D523.0    stand  255  -310   450    0     80   363  
17:03:51.363 D523.0    stand  255  -310   450    0     80   0    
17:03:51.363 D523.1    stand  255  -500   140    0     80   363  
17:03:52.364 D523.1    stand  255  -500   140    0     80   0    
17:03:52.364 D523.0    stand  255  -310   450    0     80   363  
17:03:53.365 D523.1    stand  255  -500   140    0     80   363  
17:03:53.365 D523.0    stand  255  -310   450    0     80   363  
17:03:54.370 D523.1    stand  255  -500   140    0     80   363  
17:03:54.370 D523.0    stand  255  -310   450    0     80   363  
17:03:55.375 D523.1    stand  255  -500   140    0     80   363  
17:03:55.375 D523.0    stand  255  -310   450    0     80   363  
17:03:56.379 D523.0    stand  255  -310   450    0     80   0    
17:03:56.379 D523.1    stand  255  -480   90     118   80   398  
17:03:57.380 D523.1    walk   255  -380   130    116   80   107  
17:03:57.380 D523.0    stand  255  -310   450    0     80   327  
17:03:58.379 D523.1    walk   255  -330   220    101   80   230  
17:03:58.379 D523.0    stand  255  -310   450    0     80   230  
17:03:59.380 D523.0    stand  255  -310   450    0     80   0    
17:03:59.380 D523.1    walk   255  -280   330    73    80   123  
17:04:00.442 D523.1    walk   255  -290   420    122   80   90   
17:04:00.881 D523.0    stand  0    -470   80     0     80   384  
17:04:00.881 D523.1    walk   255  -300   460    109   80   416  
17:04:01.293 D523.1    walk   255  -290   480    0     80   22   
17:04:01.293 D523.0    stand  0    -450   100    103   80   412  
17:04:02.294 D523.1    walk   255  -300   490    0     80   417  
17:04:02.294 D523.0    stand  0    -420   100    121   80   408  
17:04:03.288 D523.0    walk   0    -380   130    105   80   50   
17:04:03.288 D523.1    walk   255  -270   490    0     80   376  
17:04:04.289 D523.1    stand  255  -280   480    0     80   14   
17:04:04.289 D523.0    walk   0    -330   150    99    80   333  
17:04:05.320 D523.1    stand  255  -280   480    0     80   333  
17:04:05.320 D523.0    walk   0    -350   190    116   80   298  
17:04:06.300 D523.0    walk   0    -340   220    97    80   31   
17:04:06.300 D523.1    stand  255  -270   480    0     80   269  
17:04:07.293 D523.1    stand  255  -270   500    0     80   20   
17:04:07.293 D523.0    walk   0    -320   220    126   80   284  
17:04:08.298 D523.1    stand  255  -250   510    0     80   298  
17:04:08.298 D523.0    stand  0    -330   210    0     80   310  
17:04:09.296 D523.1    stand  255  -260   490    0     80   288  
17:04:09.296 D523.0    stand  0    -330   210    0     80   288  
17:04:10.306 D523.0    stand  0    -320   210    121   80   10   
17:04:10.306 D523.1    stand  255  -270   480    0     80   274  
17:04:11.208 D523.1    stand  255  -280   470    0     80   14   
17:04:11.208 D523.0    stand  0    -360   240    111   80   243  
17:04:12.219 D523.0    stand  0    -320   290    85    80   64   
17:04:12.219 D523.1    stand  255  -300   440    0     80   151  
17:04:13.211 D523.1    stand  255  -290   460    79    80   22   
17:04:13.211 D523.0    walk   0    -300   320    0     80   140  
17:04:14.210 D523.1    stand  255  -290   480    92    80   160  
17:04:14.210 D523.0    walk   0    -300   340    0     80   140  
17:04:15.210 D523.1    stand  255  -270   490    0     80   152  
17:04:15.210 D523.0    walk   0    -300   350    83    80   143  
17:04:16.211 D523.0    walk   0    -320   330    88    80   28   
17:04:16.211 D523.1    stand  255  -290   480    0     80   152  
17:04:17.212 D523.1    stand  255  -290   470    0     80   10   
17:04:17.212 D523.0    walk   0    -300   350    117   80   120  
17:04:18.226 D523.1    stand  255  -280   480    0     80   131  
17:04:18.226 D523.0    walk   0    -280   340    106   80   140  
17:04:19.215 D523.1    stand  255  -280   490    78    80   150  
17:04:19.215 D523.0    walk   0    -280   380    0     80   110  
17:04:20.218 D523.1    stand  255  -300   490    0     80   111  
17:04:20.218 D523.0    walk   0    -280   370    0     80   121  
17:04:21.216 D523.1    stand  255  -290   490    0     80   120  
17:04:21.216 D523.0    walk   0    -290   360    107   80   130  
17:04:22.224 D523.1    stand  255  -300   480    0     80   120  
17:04:22.224 D523.0    walk   0    -300   350    100   80   130  
17:04:23.120 D523.1    stand  255  -290   450    0     80   100  
17:04:23.120 D523.0    walk   0    -280   310    104   80   140  
17:04:24.110 D523.0    walk   0    -290   260    96    80   50   
17:04:24.110 D523.1    stand  255  -290   440    0     80   180  
17:04:25.110 D523.1    stand  255  -310   350    0     80   92   
17:04:25.110 D523.0    walk   0    -280   220    100   80   133  
17:04:26.121 D523.1    stand  255  -320   350    0     80   136  
17:04:26.121 D523.0    walk   0    -270   190    100   80   167  
17:04:27.081 D523.0    walk   0    -250   160    81    80   36   
17:04:27.081 D523.1    stand  255  -310   360    0     80   208  
17:04:28.080 D523.1    stand  255  -310   360    0     80   0    
17:04:28.080 D523.0    walk   0    -240   110    88    80   259  
17:04:29.080 D523.0    walk   0    -230   100    105   80   14   
17:04:29.080 D523.1    stand  255  -310   360    0     80   272  
17:04:30.083 D523.1    stand  255  -310   360    0     80   0    
17:04:30.083 D523.0    walk   0    -240   100    128   80   269  
17:04:31.081 D523.1    stand  255  -310   360    0     80   269  
17:04:31.081 D523.0    walk   0    -240   100    108   80   269  
17:04:32.096 D523.1    stand  255  -310   360    0     80   269  
17:04:32.096 D523.0    walk   0    -240   110    113   80   259  
17:04:33.083 D523.0    walk   0    -230   100    0     80   14   
17:04:33.083 D523.1    stand  255  -310   360    0     80   272  
17:04:34.092 D523.1    stand  255  -310   360    0     80   0    
17:04:34.092 D523.0    walk   0    -230   110    119   80   262  
17:04:35.090 D523.0    walk   0    -230   110    120   80   0    
17:04:35.090 D523.1    stand  255  -310   360    0     80   262  
17:04:36.089 D523.1    stand  255  -310   360    0     80   0    
17:04:36.089 D523.0    walk   0    -230   110    120   80   262  
17:04:37.092 D523.0    walk   0    -270   120    78    80   41   
17:04:37.092 D523.1    stand  255  -310   360    0     80   243  
17:04:38.090 D523.0    walk   0    -280   110    79    80   251  
17:04:38.090 D523.1    stand  255  -310   360    0     80   251  
17:04:38.980 D523.1    stand  255  -310   360    0     80   0    
17:04:38.980 D523.0    walk   0    -290   110    116   80   250  
17:04:39.985 D523.0    walk   0    -290   110    111   80   0    
17:04:39.985 D523.1    stand  255  -310   360    0     80   250  
17:04:40.993 D523.1    stand  255  -310   360    0     80   0    
17:04:40.993 D523.0    walk   0    -290   110    101   80   250  
17:04:41.987 D523.1    stand  255  -310   360    0     80   250  
17:04:41.987 D523.0    walk   0    -300   100    92    80   260  
17:04:43.011 D523.1    stand  255  -310   360    0     80   260  
17:04:43.011 D523.0    walk   0    -340   100    90    80   261  
17:04:44.058 D523.0    walk   0    -370   90     83    80   31   
17:04:44.058 D523.1    stand  255  -310   360    0     80   276  
17:04:45.017 D523.0    walk   0    -380   90     114   80   278  
17:04:45.017 D523.1    stand  255  -310   360    0     80   278  
17:04:46.018 D523.1    stand  255  -310   360    0     80   0    
17:04:46.018 D523.0    walk   0    -380   90     0     80   278  
17:04:47.013 D523.0    stand  0    -390   90     0     80   10   
17:04:47.013 D523.1    stand  255  -310   360    0     80   281  
17:04:47.908 D523.0    stand  0    -390   90     78    80   281  
17:04:47.908 D523.1    stand  255  -310   360    0     80   281  
17:04:48.910 D523.1    stand  255  -310   360    0     80   0    
17:04:48.910 D523.0    stand  0    -420   100    96    80   282  
17:04:49.912 D523.0    walk   0    -480   90     116   80   60   
17:04:49.912 D523.1    stand  255  -310   360    0     80   319  
17:04:50.916 D523.1    stand  255  -310   360    0     80   0    
17:04:50.916 D523.0    walk   0    -520   60     0     80   366  
17:04:51.953 D523.0    walk   0    -520   60     0     80   0    
17:04:51.953 D523.1    stand  255  -310   360    0     80   366  
17:04:52.929 D523.1    stand  255  -310   360    0     80   0    
17:04:52.929 D523.0    stand  0    -510   60     0     80   360  
17:04:53.914 D523.1    stand  255  -310   360    0     80   360  
17:04:53.914 D523.0    stand  0    -510   60     0     80   360  
17:04:54.917 D523.0    stand  0    -510   60     0     80   0    
17:04:54.917 D523.1    stand  255  -310   360    0     80   360  
17:04:55.916 D523.0    stand  2    -510   60     0     80   360  
17:04:55.916 D523.1    stand  255  -310   360    0     80   360  
17:04:56.973 D523.1    stand  255  -310   360    0     80   0    
17:04:57.841 D523.1    stand  255  -310   360    0     80   0    
17:04:58.836 D523.1    stand  255  -310   360    0     80   0    
17:04:59.828 D523.1    stand  255  -310   360    0     80   0    
17:05:00.831 D523.1    stand  255  -310   360    0     80   0    
17:05:01.835 D523.1    stand  255  -310   360    0     80   0    
17:05:02.832 D523.1    stand  255  -310   360    0     80   0    
17:05:03.832 D523.1    stand  255  -310   360    0     80   0    
17:05:04.836 D523.1    stand  255  -310   360    0     80   0    
17:05:05.836 D523.1    stand  255  -310   360    0     80   0    
17:05:06.855 D523.1    stand  255  -310   360    0     80   0    
17:05:07.835 D523.1    stand  255  -310   360    0     80   0    
17:05:08.839 D523.1    stand  255  -310   360    0     80   0    
17:05:09.732 D523.1    stand  255  -310   360    0     80   0    
17:05:10.736 D523.1    stand  255  -310   360    0     80   0    
17:05:11.736 D523.1    stand  255  -310   360    0     80   0    
17:05:12.743 D523.1    stand  255  -310   360    0     80   0    
17:05:13.739 D523.1    stand  255  -310   360    0     80   0    
17:05:14.692 D523.1    stand  255  -310   360    0     80   0    
17:05:15.691 D523.1    stand  255  -310   360    0     80   0    
17:05:16.692 D523.1    stand  255  -310   360    0     80   0    
17:05:17.701 D523.1    stand  255  -310   360    0     80   0    
17:05:18.706 D523.1    stand  255  -310   360    0     80   0    
17:05:19.697 D523.1    stand  255  -310   360    0     80   0    
17:05:20.705 D523.1    stand  255  -310   360    0     80   0    
17:05:21.700 D523.1    stand  255  -310   360    0     80   0    
17:05:22.707 D523.1    stand  255  -310   360    0     80   0    
17:05:23.766 D523.88   88     -    -      -      -     -    -    
17:05:24.610 D523.88   88     -    -      -      -     -    -    
17:05:25.610 D523.88   88     -    -      -      -     -    -    
17:05:33.535 D523.88   88     -    -      -      -     -    -    

17:02:00.297 D5F7.88   88     -    -      -      -     -    -    
17:02:31.932 D5F7.88   88     -    -      -      -     -    -    
17:03:03.473 D5F7.88   88     -    -      -      -     -    -    
17:03:35.530 D5F7.88   88     -    -      -      -     -    -    
17:04:06.941 D5F7.88   88     -    -      -      -     -    -    
17:04:39.114 D5F7.88   88     -    -      -      -     -    -    
17:05:10.439 D5F7.88   88     -    -      -      -     -    -    
17:05:42.501 D5F7.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 411 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

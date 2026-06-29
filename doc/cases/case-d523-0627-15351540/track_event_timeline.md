# case-d523-0627-15351540 — 每 tick belief 时间线 (room fd00:0:3:111:3:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
15:35:00 D523.0   D52303500626  stand   0    NoReport stand              trk  0.50 Empty      1   0     0.00  0.03  0.08  0.00  0.85  0.04
15:35:01 D523.0   D52303500626  stand   0    NoReport stand              trk  0.51 Empty      1   0     0.00  0.04  0.10  0.00  0.82  0.01
15:35:02 D523.0   D52303500626  stand   0    NoReport stand              trk  0.52 Empty      1   2     0.00  0.04  0.11  0.01  0.76  0.01
15:35:03 D523.0   D52303500626  stand   0    NoReport stand              trk  0.53 Empty      1   3     0.00  0.05  0.12  0.01  0.71  0.01
15:35:04 D523.0   D52303500626  stand   0    NoReport stand              trk  0.54 Empty      1   4     0.00  0.05  0.13  0.02  0.66  0.01
15:35:05 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   5     0.00  0.06  0.14  0.03  0.62  0.01
15:35:06 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   6     0.01  0.06  0.14  0.04  0.58  0.01
15:35:07 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   7     0.01  0.07  0.14  0.04  0.54  0.01
15:35:08 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   7     0.01  0.07  0.15  0.05  0.51  0.02
15:35:08 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   7     0.01  0.07  0.15  0.06  0.49  0.02
15:35:09 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   8     0.01  0.08  0.15  0.06  0.46  0.02
15:35:10 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   9     0.01  0.08  0.15  0.07  0.44  0.02
15:35:11 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   10    0.02  0.08  0.15  0.07  0.42  0.02
15:35:12 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   11    0.02  0.08  0.15  0.08  0.40  0.02
15:35:13 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   12    0.02  0.09  0.15  0.08  0.38  0.02
15:35:14 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   13    0.02  0.09  0.15  0.09  0.36  0.02
15:35:15 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   14    0.03  0.09  0.15  0.09  0.35  0.02
15:35:16 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   15    0.03  0.09  0.15  0.10  0.34  0.02
15:35:17 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   16    0.03  0.09  0.15  0.10  0.33  0.02
15:35:18 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   17    0.03  0.10  0.15  0.10  0.31  0.02
15:35:19 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   18    0.03  0.10  0.16  0.10  0.30  0.02
15:35:20 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   19    0.04  0.10  0.16  0.11  0.30  0.02
15:35:21 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   20    0.04  0.10  0.15  0.11  0.29  0.02
15:35:22 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   21    0.04  0.10  0.15  0.11  0.28  0.02
15:35:23 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   22    0.04  0.10  0.15  0.11  0.27  0.02
15:35:24 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   23    0.05  0.10  0.15  0.12  0.27  0.02
15:35:25 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   24    0.05  0.10  0.15  0.12  0.26  0.02
15:35:26 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   25    0.05  0.10  0.15  0.12  0.25  0.02
15:35:27 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   26    0.05  0.10  0.15  0.12  0.25  0.02
15:35:28 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   27    0.06  0.10  0.15  0.12  0.25  0.02
15:35:29 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   28    0.06  0.10  0.15  0.12  0.24  0.02
15:35:30 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   28    0.06  0.10  0.15  0.12  0.24  0.02
15:35:30 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   29    0.06  0.10  0.15  0.12  0.24  0.02
15:35:31 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   30    0.06  0.10  0.15  0.12  0.23  0.02
15:35:32 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   31    0.07  0.10  0.15  0.12  0.23  0.02
15:35:33 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   32    0.07  0.10  0.15  0.12  0.23  0.02
15:35:34 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   33    0.07  0.10  0.15  0.13  0.22  0.02
15:35:35 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   34    0.07  0.10  0.15  0.13  0.22  0.02
15:35:36 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   35    0.07  0.10  0.15  0.13  0.22  0.02
15:35:37 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   36    0.08  0.11  0.15  0.13  0.22  0.02
15:35:38 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   37    0.08  0.11  0.15  0.13  0.21  0.02
15:35:39 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   38    0.08  0.11  0.15  0.13  0.21  0.02
15:35:40 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   39    0.09  0.11  0.15  0.13  0.21  0.02
15:35:40 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   39    0.09  0.10  0.15  0.13  0.21  0.02
15:35:41 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   40    0.09  0.10  0.15  0.13  0.20  0.02
15:35:42 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   41    0.09  0.10  0.15  0.13  0.20  0.02
15:35:43 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   42    0.10  0.10  0.15  0.13  0.20  0.02
15:35:44 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   43    0.10  0.10  0.15  0.13  0.20  0.02
15:35:45 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   44    0.10  0.10  0.15  0.13  0.20  0.02
15:35:46 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   45    0.10  0.10  0.15  0.13  0.20  0.02
15:35:47 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   46    0.10  0.10  0.15  0.13  0.20  0.02
15:35:48 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   47    0.10  0.10  0.15  0.13  0.20  0.02
15:35:49 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   48    0.11  0.10  0.15  0.13  0.20  0.02
15:35:50 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   49    0.11  0.10  0.15  0.13  0.20  0.02
15:35:51 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   50    0.11  0.10  0.15  0.13  0.19  0.02
15:35:52 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   51    0.11  0.10  0.15  0.13  0.19  0.02
15:35:53 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   52    0.11  0.10  0.15  0.13  0.19  0.02
15:35:54 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   53    0.12  0.10  0.15  0.13  0.19  0.02
15:35:55 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   54    0.12  0.10  0.15  0.13  0.19  0.02
15:35:56 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   55    0.12  0.10  0.15  0.13  0.19  0.02
15:35:57 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   56    0.12  0.10  0.15  0.13  0.19  0.02
15:35:58 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   57    0.12  0.10  0.15  0.13  0.19  0.02
15:35:59 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   58    0.12  0.10  0.14  0.13  0.19  0.02
15:36:00 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   59    0.13  0.10  0.14  0.13  0.19  0.02
15:36:01 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   60    0.13  0.10  0.14  0.13  0.19  0.02
15:36:01 D5F7.E   -             -       0    NoReport np=0  ★0           room -    Empty      1   60    0.13  0.10  0.14  0.13  0.19  0.02
15:36:02 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   60    0.13  0.10  0.14  0.13  0.19  0.02
15:36:02 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   61    0.13  0.10  0.14  0.13  0.19  0.02
15:36:03 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   62    0.13  0.10  0.14  0.13  0.19  0.02
15:36:04 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   63    0.13  0.10  0.14  0.12  0.19  0.02
15:36:05 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   64    0.13  0.10  0.14  0.12  0.19  0.02
15:36:06 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   65    0.14  0.10  0.14  0.12  0.19  0.02
15:36:07 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   66    0.14  0.10  0.14  0.12  0.19  0.02
15:36:08 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   67    0.14  0.10  0.14  0.12  0.19  0.02
15:36:09 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   68    0.14  0.10  0.14  0.12  0.18  0.02
15:36:10 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   69    0.14  0.10  0.14  0.12  0.18  0.02
15:36:11 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   70    0.14  0.10  0.14  0.12  0.18  0.02
15:36:11 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   71    0.14  0.10  0.14  0.12  0.18  0.02
15:36:12 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   71    0.14  0.10  0.14  0.12  0.18  0.02
15:36:13 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   72    0.15  0.10  0.14  0.12  0.18  0.02
15:36:14 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   73    0.15  0.10  0.14  0.12  0.18  0.02
15:36:15 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   74    0.15  0.10  0.14  0.12  0.18  0.02
15:36:16 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   75    0.15  0.10  0.14  0.12  0.18  0.02
15:36:17 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   76    0.15  0.10  0.14  0.12  0.18  0.02
15:36:17 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   77    0.15  0.10  0.14  0.12  0.18  0.02
15:36:18 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   78    0.15  0.10  0.14  0.12  0.18  0.02
15:36:19 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   79    0.16  0.10  0.14  0.12  0.18  0.02
15:36:20 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   80    0.16  0.10  0.14  0.12  0.18  0.02
15:36:21 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   81    0.16  0.10  0.14  0.12  0.18  0.02
15:36:22 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   82    0.16  0.10  0.14  0.12  0.18  0.02
15:36:23 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   83    0.16  0.10  0.14  0.12  0.18  0.02
15:36:24 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   84    0.16  0.10  0.14  0.12  0.18  0.02
15:36:25 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   85    0.16  0.10  0.14  0.12  0.18  0.02
15:36:26 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   86    0.16  0.10  0.14  0.12  0.18  0.02
15:36:27 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   87    0.16  0.10  0.14  0.12  0.18  0.02
15:36:28 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   88    0.17  0.10  0.14  0.12  0.18  0.02
15:36:29 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   89    0.17  0.10  0.14  0.12  0.18  0.02
15:36:30 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   90    0.17  0.10  0.14  0.12  0.18  0.02
15:36:31 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   91    0.17  0.10  0.14  0.12  0.18  0.02
15:36:32 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   92    0.17  0.10  0.14  0.12  0.18  0.02
15:36:33 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   93    0.17  0.10  0.14  0.12  0.18  0.02
15:36:33 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      1   93    0.17  0.10  0.14  0.12  0.18  0.02
15:36:34 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   94    0.17  0.10  0.14  0.12  0.18  0.02
15:36:35 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   95    0.17  0.10  0.14  0.12  0.18  0.02
15:36:36 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   96    0.17  0.10  0.14  0.12  0.18  0.02
15:36:37 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Empty      1   97    0.17  0.10  0.14  0.12  0.18  0.02
15:36:38 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   98    0.18  0.10  0.14  0.12  0.18  0.02
15:36:39 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   99    0.18  0.10  0.14  0.12  0.17  0.02
15:36:40 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   100   0.18  0.10  0.14  0.12  0.17  0.02
15:36:41 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   101   0.18  0.10  0.14  0.12  0.17  0.02
15:36:42 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   102   0.18  0.10  0.14  0.12  0.17  0.02
15:36:43 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   103   0.18  0.10  0.14  0.12  0.17  0.02
15:36:43 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   103   0.18  0.10  0.14  0.12  0.17  0.02
15:36:44 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   104   0.18  0.10  0.14  0.12  0.17  0.02
15:36:45 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   105   0.18  0.10  0.14  0.12  0.17  0.02
15:36:46 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   106   0.19  0.09  0.14  0.12  0.17  0.02
15:36:47 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   107   0.19  0.09  0.14  0.12  0.17  0.02
15:36:48 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   108   0.19  0.09  0.14  0.12  0.17  0.02
15:36:49 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   109   0.19  0.09  0.14  0.12  0.17  0.02
15:36:50 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   110   0.19  0.09  0.14  0.12  0.17  0.02
15:36:51 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   111   0.19  0.09  0.14  0.12  0.17  0.02
15:36:52 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   112   0.19  0.09  0.14  0.12  0.17  0.02
15:36:53 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   113   0.19  0.09  0.14  0.12  0.17  0.02
15:36:54 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   114   0.19  0.09  0.14  0.12  0.17  0.02
15:36:55 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   115   0.19  0.09  0.14  0.12  0.17  0.02
15:36:56 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   116   0.19  0.09  0.13  0.12  0.17  0.02
15:36:57 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   117   0.19  0.09  0.13  0.11  0.17  0.02
15:36:58 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   118   0.19  0.09  0.13  0.11  0.17  0.02
15:36:59 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   119   0.20  0.09  0.13  0.11  0.17  0.02
15:37:00 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   120   0.20  0.09  0.13  0.11  0.17  0.02
15:37:01 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   121   0.20  0.09  0.13  0.11  0.17  0.02
15:37:02 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   121   0.20  0.09  0.13  0.11  0.17  0.02
15:37:03 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   122   0.20  0.09  0.13  0.11  0.17  0.02
15:37:04 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   123   0.20  0.09  0.13  0.11  0.17  0.02
15:37:05 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   123   0.20  0.09  0.13  0.11  0.17  0.02
15:37:05 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   124   0.20  0.09  0.13  0.11  0.17  0.02
15:37:06 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   125   0.20  0.09  0.13  0.11  0.17  0.02
15:37:07 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   126   0.20  0.09  0.13  0.11  0.17  0.02
15:37:08 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   127   0.20  0.09  0.13  0.11  0.17  0.02
15:37:09 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   128   0.20  0.09  0.13  0.11  0.17  0.02
15:37:10 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   129   0.20  0.09  0.13  0.11  0.17  0.02
15:37:11 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   130   0.20  0.09  0.13  0.11  0.17  0.02
15:37:12 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   131   0.20  0.09  0.13  0.11  0.17  0.02
15:37:13 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   132   0.20  0.09  0.13  0.11  0.17  0.02
15:37:14 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   133   0.21  0.09  0.13  0.11  0.17  0.02
15:37:15 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   134   0.21  0.09  0.13  0.11  0.17  0.02
15:37:15 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   134   0.21  0.09  0.13  0.11  0.17  0.02
15:37:16 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   135   0.21  0.09  0.13  0.11  0.17  0.02
15:37:17 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   136   0.21  0.09  0.13  0.11  0.17  0.02
15:37:18 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   137   0.21  0.09  0.13  0.11  0.17  0.02
15:37:19 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   138   0.21  0.09  0.13  0.11  0.17  0.02
15:37:20 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   139   0.21  0.09  0.13  0.11  0.17  0.02
15:37:21 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   140   0.21  0.09  0.13  0.11  0.17  0.02
15:37:22 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   141   0.21  0.09  0.13  0.11  0.17  0.02
15:37:23 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   142   0.21  0.09  0.13  0.11  0.17  0.02
15:37:24 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
15:37:25 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
15:37:26 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
15:37:27 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
15:37:28 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
15:37:29 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
15:37:30 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
15:37:31 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.21  0.09  0.13  0.11  0.17  0.02
15:37:32 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:33 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:34 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:35 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:36 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:37 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:37 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:38 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:39 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:40 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:41 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.17  0.02
15:37:42 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:43 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:44 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:45 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:46 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:47 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:47 D523.0   D52303500626  stand   66   NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:48 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:49 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:50 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.22  0.09  0.13  0.11  0.16  0.02
15:37:51 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   26    0.22  0.09  0.13  0.11  0.16  0.02
15:37:52 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   27    0.22  0.09  0.13  0.11  0.16  0.02
15:37:53 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   28    0.22  0.09  0.13  0.11  0.16  0.02
15:37:54 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   29    0.22  0.09  0.13  0.11  0.16  0.02
15:37:55 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   30    0.22  0.09  0.13  0.11  0.16  0.02
15:37:56 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   31    0.22  0.09  0.13  0.11  0.16  0.02
15:37:57 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   32    0.23  0.09  0.13  0.11  0.16  0.02
15:37:58 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   33    0.23  0.09  0.13  0.11  0.16  0.02
15:37:59 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   34    0.23  0.09  0.13  0.11  0.16  0.02
15:38:00 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   35    0.23  0.09  0.13  0.11  0.16  0.02
15:38:01 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   36    0.23  0.09  0.13  0.11  0.16  0.02
15:38:02 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   37    0.23  0.09  0.13  0.11  0.16  0.02
15:38:03 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   38    0.23  0.09  0.13  0.11  0.16  0.02
15:38:04 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   39    0.23  0.09  0.13  0.11  0.16  0.02
15:38:05 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   40    0.23  0.09  0.13  0.11  0.16  0.02
15:38:06 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   41    0.23  0.09  0.13  0.11  0.16  0.02
15:38:07 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   42    0.23  0.09  0.13  0.11  0.16  0.02
15:38:08 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   43    0.23  0.09  0.13  0.11  0.16  0.02
15:38:08 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   43    0.23  0.09  0.13  0.11  0.16  0.02
15:38:09 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   44    0.23  0.09  0.13  0.11  0.16  0.02
15:38:10 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   45    0.23  0.09  0.13  0.11  0.16  0.02
15:38:11 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   46    0.23  0.09  0.13  0.11  0.16  0.02
15:38:12 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   47    0.23  0.09  0.13  0.11  0.16  0.02
15:38:13 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   48    0.23  0.09  0.13  0.11  0.16  0.02
15:38:14 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   49    0.23  0.09  0.13  0.11  0.16  0.02
15:38:15 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   50    0.23  0.09  0.13  0.11  0.16  0.02
15:38:16 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   51    0.23  0.09  0.13  0.11  0.16  0.02
15:38:17 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   52    0.23  0.09  0.13  0.11  0.16  0.02
15:38:18 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   53    0.23  0.09  0.13  0.11  0.16  0.02
15:38:19 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   54    0.23  0.09  0.13  0.11  0.16  0.02
15:38:19 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   54    0.23  0.09  0.13  0.11  0.16  0.02
15:38:20 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   55    0.23  0.09  0.13  0.11  0.16  0.02
15:38:21 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   56    0.23  0.09  0.13  0.11  0.16  0.02
15:38:21 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   57    0.23  0.09  0.13  0.11  0.16  0.02
15:38:22 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   58    0.23  0.09  0.13  0.11  0.16  0.02
15:38:23 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   59    0.23  0.09  0.13  0.11  0.16  0.02
15:38:24 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   60    0.23  0.09  0.13  0.11  0.16  0.02
15:38:25 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   61    0.23  0.09  0.13  0.11  0.16  0.02
15:38:26 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   62    0.23  0.09  0.13  0.11  0.16  0.02
15:38:27 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   63    0.23  0.09  0.13  0.11  0.16  0.02
15:38:28 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   64    0.23  0.09  0.13  0.11  0.16  0.02
15:38:29 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   65    0.23  0.09  0.13  0.11  0.16  0.02
15:38:30 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   66    0.23  0.09  0.13  0.11  0.16  0.02
15:38:31 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   67    0.23  0.09  0.13  0.11  0.16  0.02
15:38:32 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   68    0.23  0.09  0.13  0.11  0.16  0.02
15:38:33 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   69    0.23  0.09  0.13  0.11  0.16  0.02
15:38:34 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   70    0.23  0.09  0.13  0.11  0.16  0.02
15:38:35 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   71    0.23  0.09  0.13  0.11  0.16  0.02
15:38:36 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   72    0.24  0.09  0.13  0.11  0.16  0.02
15:38:37 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   73    0.24  0.09  0.13  0.11  0.16  0.02
15:38:38 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   74    0.24  0.09  0.13  0.11  0.16  0.02
15:38:39 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   75    0.24  0.09  0.13  0.11  0.16  0.02
15:38:40 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   75    0.24  0.09  0.13  0.11  0.16  0.02
15:38:40 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   76    0.24  0.09  0.13  0.11  0.16  0.02
15:38:41 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   77    0.24  0.09  0.13  0.11  0.16  0.02
15:38:42 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   78    0.24  0.09  0.13  0.11  0.16  0.02
15:38:43 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   79    0.24  0.09  0.13  0.11  0.16  0.02
15:38:44 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   80    0.24  0.09  0.13  0.11  0.16  0.02
15:38:45 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   81    0.24  0.09  0.13  0.11  0.16  0.02
15:38:46 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   82    0.24  0.09  0.13  0.11  0.16  0.02
15:38:47 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   83    0.24  0.09  0.13  0.11  0.16  0.02
15:38:48 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   84    0.24  0.09  0.13  0.11  0.16  0.02
15:38:49 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   85    0.24  0.09  0.13  0.11  0.16  0.02
15:38:50 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   85    0.24  0.09  0.13  0.11  0.16  0.02
15:38:50 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   86    0.24  0.09  0.13  0.11  0.16  0.02
15:38:51 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   87    0.24  0.09  0.13  0.11  0.16  0.02
15:38:52 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   88    0.24  0.09  0.13  0.11  0.16  0.02
15:38:53 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   89    0.24  0.09  0.13  0.11  0.16  0.02
15:38:54 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   90    0.24  0.09  0.13  0.11  0.16  0.02
15:38:55 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   91    0.24  0.09  0.13  0.11  0.16  0.02
15:38:56 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   92    0.24  0.09  0.13  0.11  0.16  0.02
15:38:57 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   93    0.24  0.09  0.13  0.11  0.16  0.02
15:38:58 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   94    0.24  0.09  0.13  0.11  0.16  0.02
15:38:59 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   95    0.24  0.09  0.13  0.11  0.16  0.02
15:39:00 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   96    0.24  0.09  0.13  0.11  0.16  0.02
15:39:01 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   97    0.24  0.09  0.13  0.11  0.16  0.02
15:39:02 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   98    0.24  0.09  0.13  0.11  0.16  0.02
15:39:03 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   99    0.24  0.09  0.13  0.11  0.16  0.02
15:39:04 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   100   0.24  0.09  0.13  0.11  0.16  0.02
15:39:05 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   101   0.24  0.09  0.13  0.11  0.16  0.02
15:39:06 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   102   0.24  0.09  0.13  0.11  0.16  0.02
15:39:07 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   103   0.24  0.09  0.13  0.11  0.16  0.02
15:39:08 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   104   0.24  0.09  0.13  0.11  0.16  0.02
15:39:09 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   105   0.24  0.09  0.13  0.11  0.16  0.02
15:39:10 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   106   0.24  0.09  0.13  0.11  0.16  0.02
15:39:11 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   107   0.24  0.09  0.13  0.11  0.16  0.02
15:39:12 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   107   0.24  0.09  0.13  0.11  0.16  0.02
15:39:12 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   108   0.24  0.09  0.13  0.11  0.16  0.02
15:39:13 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   109   0.24  0.09  0.13  0.11  0.16  0.02
15:39:14 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   110   0.24  0.09  0.13  0.11  0.16  0.02
15:39:15 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   111   0.24  0.09  0.13  0.11  0.16  0.02
15:39:16 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   112   0.24  0.09  0.13  0.11  0.16  0.02
15:39:17 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   113   0.24  0.09  0.13  0.11  0.16  0.02
15:39:18 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   114   0.24  0.09  0.13  0.11  0.16  0.02
15:39:19 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   114   0.24  0.09  0.13  0.11  0.16  0.02
15:39:20 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   115   0.24  0.09  0.13  0.11  0.16  0.02
15:39:21 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   116   0.24  0.09  0.13  0.11  0.16  0.02
15:39:22 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   117   0.24  0.09  0.13  0.11  0.16  0.02
15:39:22 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   118   0.24  0.09  0.13  0.11  0.16  0.02
15:39:23 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   118   0.24  0.09  0.13  0.11  0.16  0.02
15:39:24 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   119   0.24  0.09  0.13  0.11  0.16  0.02
15:39:25 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   120   0.24  0.09  0.13  0.11  0.16  0.02
15:39:26 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   121   0.24  0.09  0.13  0.11  0.16  0.02
15:39:27 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   122   0.24  0.09  0.13  0.11  0.16  0.02
15:39:28 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   123   0.24  0.09  0.13  0.11  0.16  0.02
15:39:29 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   125   0.24  0.09  0.13  0.11  0.16  0.02
15:39:30 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   125   0.24  0.09  0.13  0.11  0.16  0.01
15:39:31 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   126   0.24  0.09  0.13  0.11  0.16  0.01
15:39:32 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   127   0.24  0.09  0.13  0.11  0.16  0.01
15:39:33 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   128   0.24  0.09  0.13  0.11  0.16  0.01
15:39:34 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   129   0.24  0.09  0.13  0.11  0.16  0.01
15:39:35 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   130   0.24  0.09  0.13  0.11  0.16  0.01
15:39:36 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   131   0.24  0.09  0.13  0.11  0.16  0.01
15:39:37 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   132   0.24  0.09  0.13  0.11  0.16  0.01
15:39:38 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   133   0.24  0.09  0.13  0.11  0.16  0.01
15:39:39 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   134   0.24  0.09  0.13  0.11  0.16  0.01
15:39:40 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   135   0.24  0.09  0.13  0.11  0.16  0.01
15:39:41 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   136   0.24  0.09  0.13  0.11  0.16  0.01
15:39:42 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   137   0.24  0.09  0.13  0.11  0.16  0.01
15:39:43 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   138   0.24  0.09  0.13  0.11  0.16  0.01
15:39:44 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   139   0.24  0.09  0.13  0.11  0.16  0.01
15:39:44 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   139   0.24  0.09  0.13  0.11  0.16  0.01
15:39:45 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   140   0.24  0.09  0.13  0.11  0.16  0.01
15:39:46 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   141   0.24  0.09  0.13  0.11  0.16  0.01
15:39:47 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   142   0.24  0.09  0.13  0.11  0.16  0.01
15:39:48 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   143   0.24  0.09  0.13  0.11  0.16  0.01
15:39:49 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   144   0.24  0.09  0.13  0.11  0.16  0.01
15:39:50 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   145   0.24  0.09  0.13  0.11  0.16  0.01
15:39:51 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   146   0.24  0.09  0.13  0.11  0.16  0.01
15:39:52 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   147   0.24  0.09  0.13  0.11  0.16  0.01
15:39:53 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   148   0.24  0.09  0.13  0.11  0.16  0.01
15:39:53 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   149   0.24  0.09  0.13  0.11  0.16  0.01
15:39:54 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   149   0.24  0.09  0.13  0.11  0.16  0.01
15:39:55 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   150   0.24  0.09  0.13  0.11  0.16  0.01
15:39:56 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   151   0.24  0.09  0.13  0.11  0.16  0.01
15:39:57 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   152   0.24  0.09  0.13  0.11  0.16  0.01
15:39:58 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   153   0.24  0.09  0.13  0.11  0.16  0.01
15:39:59 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   154   0.24  0.09  0.13  0.11  0.16  0.01
15:40:00 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   155   0.24  0.09  0.13  0.11  0.16  0.01
15:40:01 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   156   0.24  0.09  0.13  0.11  0.16  0.01
15:40:02 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   157   0.24  0.09  0.13  0.11  0.16  0.01
15:40:03 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   158   0.24  0.09  0.13  0.11  0.16  0.01
15:40:04 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   159   0.24  0.09  0.13  0.11  0.16  0.01
15:40:05 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   160   0.24  0.09  0.13  0.11  0.16  0.01
15:40:06 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   161   0.24  0.09  0.13  0.11  0.16  0.01
15:40:07 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   162   0.24  0.09  0.13  0.11  0.16  0.01
15:40:08 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   163   0.24  0.09  0.13  0.11  0.16  0.01
15:40:09 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   164   0.24  0.09  0.13  0.11  0.16  0.01
15:40:10 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   165   0.24  0.09  0.13  0.11  0.16  0.01
15:40:11 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   166   0.25  0.09  0.13  0.11  0.16  0.01
15:40:12 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   167   0.25  0.09  0.13  0.11  0.16  0.01
15:40:13 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   168   0.25  0.09  0.13  0.11  0.16  0.01
15:40:14 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   169   0.25  0.09  0.13  0.11  0.16  0.01
15:40:15 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   170   0.25  0.09  0.13  0.11  0.16  0.01
15:40:15 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   170   0.25  0.09  0.13  0.11  0.16  0.01
15:40:16 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   171   0.25  0.09  0.13  0.11  0.16  0.01
15:40:17 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   172   0.25  0.09  0.13  0.11  0.16  0.01
15:40:18 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   173   0.25  0.09  0.13  0.11  0.16  0.01
15:40:19 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   174   0.25  0.09  0.13  0.11  0.16  0.01
15:40:20 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   175   0.25  0.09  0.13  0.11  0.16  0.01
15:40:21 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   176   0.25  0.09  0.13  0.11  0.16  0.01
15:40:22 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   177   0.25  0.09  0.13  0.11  0.16  0.01
15:40:23 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   178   0.25  0.09  0.13  0.11  0.16  0.01
15:40:24 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   179   0.25  0.09  0.13  0.11  0.16  0.01
15:40:25 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   180   0.25  0.09  0.13  0.11  0.16  0.01
15:40:25 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   181   0.25  0.09  0.13  0.11  0.16  0.01
15:40:26 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   181   0.25  0.09  0.13  0.11  0.16  0.01
15:40:27 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   182   0.25  0.09  0.13  0.11  0.16  0.01
15:40:27 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   183   0.25  0.09  0.13  0.11  0.16  0.01
15:40:28 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   184   0.25  0.09  0.13  0.11  0.16  0.01
15:40:29 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   185   0.25  0.09  0.13  0.11  0.16  0.01
15:40:30 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   186   0.25  0.09  0.13  0.11  0.16  0.01
15:40:31 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   187   0.25  0.09  0.13  0.11  0.16  0.01
15:40:32 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   188   0.25  0.09  0.13  0.11  0.16  0.01
15:40:33 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   189   0.25  0.09  0.13  0.11  0.16  0.01
15:40:34 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   190   0.25  0.09  0.13  0.11  0.16  0.01
15:40:35 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   191   0.25  0.09  0.13  0.11  0.16  0.01
15:40:36 -.-      -             -       -    NoReport (no frame, held)   room -    Fallen     1   191   0.25  0.09  0.13  0.11  0.16  0.01
15:40:37 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   192   0.25  0.09  0.13  0.11  0.16  0.01
15:40:37 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   193   0.25  0.09  0.13  0.11  0.16  0.01
15:40:38 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   194   0.25  0.09  0.13  0.11  0.16  0.01
15:40:39 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   195   0.25  0.09  0.13  0.11  0.16  0.01
15:40:40 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   196   0.25  0.09  0.13  0.11  0.16  0.01
15:40:41 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   197   0.25  0.09  0.13  0.11  0.16  0.01
15:40:42 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   198   0.25  0.09  0.13  0.11  0.16  0.01
15:40:43 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   199   0.25  0.09  0.13  0.11  0.16  0.01
15:40:44 D523.0   D52303500626  stand   105  NoReport stand              trk  1.00 Fallen     1   200   0.25  0.09  0.13  0.11  0.16  0.01
15:40:45 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   201   0.25  0.09  0.13  0.11  0.16  0.01
15:40:46 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   202   0.25  0.09  0.13  0.11  0.16  0.01
15:40:47 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   203   0.25  0.09  0.13  0.11  0.16  0.01
15:40:48 D5F7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   203   0.25  0.09  0.13  0.11  0.16  0.01
15:40:48 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   204   0.25  0.09  0.13  0.11  0.16  0.01
15:40:49 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   205   0.25  0.09  0.13  0.11  0.16  0.01
15:40:50 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   206   0.25  0.09  0.13  0.11  0.16  0.01
15:40:51 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   207   0.25  0.09  0.13  0.11  0.16  0.01
15:40:52 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   208   0.25  0.09  0.13  0.11  0.16  0.01
15:40:53 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   209   0.25  0.09  0.13  0.11  0.16  0.01
15:40:54 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   210   0.25  0.09  0.13  0.11  0.16  0.01
15:40:55 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   211   0.25  0.09  0.13  0.11  0.16  0.01
15:40:56 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   212   0.25  0.09  0.13  0.11  0.16  0.01
15:40:57 09E7.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   212   0.25  0.09  0.13  0.11  0.16  0.01
15:40:57 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   213   0.25  0.09  0.13  0.11  0.16  0.01
15:40:58 D523.0   D52303500626  stand   0    NoReport stand              trk  1.00 Fallen     1   214   0.25  0.09  0.13  0.11  0.16  0.01
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
15:35:08.279 09E7.88   88     -    -      -      -     -    -    
15:35:40.196 09E7.88   88     -    -      -      -     -    -    
15:36:11.775 09E7.88   88     -    -      -      -     -    -    
15:36:43.780 09E7.88   88     -    -      -      -     -    -    
15:37:15.422 09E7.88   88     -    -      -      -     -    -    
15:37:47.006 09E7.88   88     -    -      -      -     -    -    
15:38:19.019 09E7.88   88     -    -      -      -     -    -    
15:38:50.432 09E7.88   88     -    -      -      -     -    -    
15:39:22.610 09E7.88   88     -    -      -      -     -    -    
15:39:53.917 09E7.88   88     -    -      -      -     -    -    
15:40:25.890 09E7.88   88     -    -      -      -     -    -    
15:40:57.470 09E7.88   88     -    -      -      -     -    -    

15:35:00.626 D523.0    stand  2    -300   500    0     80        
15:35:01.624 D523.0    stand  2    -300   500    0     80   0    
15:35:02.627 D523.0    stand  2    -300   500    0     80   0    
15:35:03.633 D523.0    stand  2    -300   500    0     80   0    
15:35:04.633 D523.0    stand  2    -300   500    0     80   0    
15:35:05.632 D523.0    stand  2    -300   500    0     80   0    
15:35:06.628 D523.0    stand  2    -300   500    0     80   0    
15:35:07.634 D523.0    stand  2    -300   500    0     80   0    
15:35:08.540 D523.0    stand  2    -300   500    0     80   0    
15:35:09.532 D523.0    stand  2    -290   500    0     80   10   
15:35:10.530 D523.0    stand  2    -290   500    0     80   0    
15:35:11.527 D523.0    stand  2    -290   500    0     80   0    
15:35:12.537 D523.0    stand  2    -290   500    0     80   0    
15:35:13.529 D523.0    stand  2    -290   500    0     80   0    
15:35:14.536 D523.0    stand  2    -290   500    0     80   0    
15:35:15.529 D523.0    stand  2    -290   500    0     80   0    
15:35:16.534 D523.0    stand  2    -290   500    0     80   0    
15:35:17.532 D523.0    stand  2    -290   500    0     80   0    
15:35:18.536 D523.0    stand  2    -290   500    0     80   0    
15:35:19.538 D523.0    stand  2    -290   500    0     80   0    
15:35:20.424 D523.0    stand  2    -290   500    0     80   0    
15:35:21.426 D523.0    stand  2    -290   500    0     80   0    
15:35:22.428 D523.0    stand  2    -290   500    0     80   0    
15:35:23.433 D523.0    stand  2    -290   500    0     80   0    
15:35:24.430 D523.0    stand  2    -290   500    0     80   0    
15:35:25.436 D523.0    stand  2    -290   500    0     80   0    
15:35:26.432 D523.0    stand  2    -290   500    0     80   0    
15:35:27.433 D523.0    stand  2    -290   500    0     80   0    
15:35:28.437 D523.0    stand  2    -290   500    0     80   0    
15:35:29.436 D523.0    stand  2    -290   500    0     80   0    
15:35:30.441 D523.0    stand  2    -290   500    0     80   0    
15:35:31.448 D523.0    stand  2    -290   500    0     80   0    
15:35:32.341 D523.0    stand  2    -290   500    0     80   0    
15:35:33.331 D523.0    stand  2    -290   500    0     80   0    
15:35:34.333 D523.0    stand  2    -290   500    0     80   0    
15:35:35.332 D523.0    stand  2    -290   500    0     80   0    
15:35:36.333 D523.0    stand  2    -290   500    0     80   0    
15:35:37.356 D523.0    stand  2    -290   500    0     80   0    
15:35:38.353 D523.0    stand  2    -290   500    0     80   0    
15:35:39.406 D523.0    stand  2    -290   500    0     80   0    
15:35:40.356 D523.0    stand  2    -300   500    0     80   10   
15:35:41.353 D523.0    stand  2    -340   500    0     80   40   
15:35:42.251 D523.0    stand  2    -340   500    0     80   0    
15:35:43.252 D523.0    stand  2    -340   500    0     80   0    
15:35:44.252 D523.0    stand  2    -340   500    0     80   0    
15:35:45.254 D523.0    stand  2    -340   500    0     80   0    
15:35:46.252 D523.0    stand  2    -330   500    0     80   10   
15:35:47.254 D523.0    stand  2    -330   500    0     80   0    
15:35:48.257 D523.0    stand  2    -330   500    0     80   0    
15:35:49.264 D523.0    stand  2    -330   500    0     80   0    
15:35:50.258 D523.0    stand  2    -330   500    0     80   0    
15:35:51.258 D523.0    stand  2    -330   500    0     80   0    
15:35:52.262 D523.0    stand  2    -330   500    0     80   0    
15:35:53.260 D523.0    stand  2    -330   500    0     80   0    
15:35:54.154 D523.0    stand  2    -330   500    0     80   0    
15:35:55.158 D523.0    stand  2    -330   500    0     80   0    
15:35:56.156 D523.0    stand  2    -330   500    0     80   0    
15:35:57.162 D523.0    stand  2    -330   500    0     80   0    
15:35:58.158 D523.0    stand  2    -330   500    0     80   0    
15:35:59.158 D523.0    stand  2    -330   500    0     80   0    
15:36:00.160 D523.0    stand  2    -330   500    0     80   0    
15:36:01.161 D523.0    stand  2    -330   500    0     80   0    
15:36:02.171 D523.0    stand  2    -330   500    0     80   0    
15:36:03.168 D523.0    stand  2    -330   500    0     80   0    
15:36:04.165 D523.0    stand  2    -330   500    0     80   0    
15:36:05.164 D523.0    stand  2    -330   500    0     80   0    
15:36:06.057 D523.0    stand  2    -330   500    0     80   0    
15:36:07.076 D523.0    stand  2    -330   500    0     80   0    
15:36:08.101 D523.0    stand  2    -330   500    0     80   0    
15:36:09.064 D523.0    stand  2    -330   500    0     80   0    
15:36:10.060 D523.0    stand  2    -330   500    0     80   0    
15:36:11.064 D523.0    stand  2    -330   500    0     80   0    
15:36:12.064 D523.0    stand  2    -330   500    0     80   0    
15:36:13.065 D523.0    stand  2    -330   500    0     80   0    
15:36:14.065 D523.0    stand  2    -330   500    0     80   0    
15:36:15.066 D523.0    stand  2    -290   510    0     80   41   
15:36:16.066 D523.0    stand  2    -290   510    0     80   0    
15:36:17.069 D523.0    stand  2    -310   490    0     80   28   
15:36:17.964 D523.0    stand  2    -310   490    0     80   0    
15:36:18.964 D523.0    stand  2    -310   490    0     80   0    
15:36:19.964 D523.0    stand  2    -310   490    0     80   0    
15:36:20.964 D523.0    stand  2    -310   490    0     80   0    
15:36:21.967 D523.0    stand  2    -310   490    0     80   0    
15:36:22.968 D523.0    stand  2    -310   490    0     80   0    
15:36:23.978 D523.0    stand  2    -310   490    0     80   0    
15:36:24.987 D523.0    stand  2    -310   490    0     80   0    
15:36:25.988 D523.0    stand  2    -310   490    0     80   0    
15:36:26.989 D523.0    stand  2    -310   490    0     80   0    
15:36:27.882 D523.0    stand  2    -310   490    0     80   0    
15:36:28.892 D523.0    stand  2    -310   490    0     80   0    
15:36:29.886 D523.0    stand  2    -310   490    0     80   0    
15:36:30.884 D523.0    stand  2    -310   490    0     80   0    
15:36:31.886 D523.0    stand  2    -310   490    0     80   0    
15:36:32.893 D523.0    stand  2    -310   490    0     80   0    
15:36:33.899 D523.0    stand  2    -310   490    0     80   0    
15:36:34.893 D523.0    stand  2    -310   490    0     80   0    
15:36:35.893 D523.0    stand  2    -310   490    0     80   0    
15:36:36.895 D523.0    stand  2    -310   490    0     80   0    
15:36:37.902 D523.0    stand  2    -310   490    0     80   0    
15:36:38.941 D523.0    stand  2    -310   490    0     80   0    
15:36:39.788 D523.0    stand  2    -310   490    0     80   0    
15:36:40.793 D523.0    stand  2    -310   490    0     80   0    
15:36:41.797 D523.0    stand  2    -310   490    0     80   0    
15:36:42.796 D523.0    stand  2    -310   490    0     80   0    
15:36:43.804 D523.0    stand  2    -310   490    0     80   0    
15:36:44.802 D523.0    stand  2    -310   490    0     80   0    
15:36:45.799 D523.0    stand  2    -310   490    0     80   0    
15:36:46.800 D523.0    stand  2    -330   510    0     80   28   
15:36:47.802 D523.0    stand  2    -330   510    0     80   0    
15:36:48.802 D523.0    stand  2    -330   510    0     80   0    
15:36:49.808 D523.0    stand  2    -330   510    0     80   0    
15:36:50.707 D523.0    stand  2    -330   510    0     80   0    
15:36:51.700 D523.0    stand  2    -330   510    0     80   0    
15:36:52.700 D523.0    stand  2    -330   510    0     80   0    
15:36:53.702 D523.0    stand  2    -330   510    0     80   0    
15:36:54.702 D523.0    stand  2    -330   510    0     80   0    
15:36:55.702 D523.0    stand  2    -330   510    0     80   0    
15:36:56.709 D523.0    stand  2    -330   510    0     80   0    
15:36:57.705 D523.0    stand  2    -330   510    0     80   0    
15:36:58.712 D523.0    stand  2    -330   510    0     80   0    
15:36:59.711 D523.0    stand  2    -330   510    0     80   0    
15:37:00.708 D523.0    stand  2    -330   510    0     80   0    
15:37:01.709 D523.0    stand  2    -330   510    0     80   0    
15:37:02.606 D523.0    stand  2    -330   510    0     80   0    
15:37:03.604 D523.0    stand  2    -330   510    0     80   0    
15:37:04.608 D523.0    stand  2    -330   510    0     80   0    
15:37:05.606 D523.0    stand  2    -330   510    0     80   0    
15:37:06.610 D523.0    stand  2    -330   510    0     80   0    
15:37:07.606 D523.0    stand  2    -330   510    0     80   0    
15:37:08.608 D523.0    stand  2    -330   510    0     80   0    
15:37:09.609 D523.0    stand  2    -330   510    0     80   0    
15:37:10.610 D523.0    stand  2    -330   510    0     80   0    
15:37:11.610 D523.0    stand  2    -330   510    0     80   0    
15:37:12.618 D523.0    stand  2    -330   510    0     80   0    
15:37:13.514 D523.0    stand  2    -330   510    0     80   0    
15:37:14.523 D523.0    stand  2    -330   510    0     80   0    
15:37:15.517 D523.0    stand  2    -330   510    0     80   0    
15:37:16.522 D523.0    stand  2    -330   510    0     80   0    
15:37:17.524 D523.0    stand  2    -330   510    0     80   0    
15:37:18.546 D523.0    stand  2    -330   510    0     80   0    
15:37:19.520 D523.0    stand  2    -330   510    0     80   0    
15:37:20.525 D523.0    stand  2    -330   510    0     80   0    
15:37:21.522 D523.0    stand  2    -330   510    0     80   0    
15:37:22.528 D523.0    stand  2    -330   510    0     80   0    
15:37:23.525 D523.0    stand  2    -330   500    0     80   10   
15:37:24.528 D523.0    stand  2    -260   520    0     80   72   
15:37:25.417 D523.0    stand  2    -260   520    0     80   0    
15:37:26.418 D523.0    stand  2    -260   520    0     80   0    
15:37:27.420 D523.0    stand  2    -260   520    0     80   0    
15:37:28.420 D523.0    stand  2    -270   520    0     80   10   
15:37:29.428 D523.0    stand  2    -270   520    0     80   0    
15:37:30.430 D523.0    stand  2    -270   520    0     80   0    
15:37:31.432 D523.0    stand  2    -270   520    0     80   0    
15:37:32.434 D523.0    stand  2    -270   520    0     80   0    
15:37:33.436 D523.0    stand  2    -270   520    0     80   0    
15:37:34.437 D523.0    stand  2    -270   520    0     80   0    
15:37:35.438 D523.0    stand  2    -270   520    0     80   0    
15:37:36.335 D523.0    stand  2    -270   520    0     80   0    
15:37:37.338 D523.0    stand  2    -270   520    0     80   0    
15:37:38.382 D523.0    stand  2    -270   520    0     80   0    
15:37:39.338 D523.0    stand  2    -270   520    0     80   0    
15:37:40.336 D523.0    stand  2    -270   520    0     80   0    
15:37:41.339 D523.0    stand  2    -270   520    0     80   0    
15:37:42.353 D523.0    stand  2    -270   520    0     80   0    
15:37:43.336 D523.0    stand  2    -270   520    0     80   0    
15:37:44.341 D523.0    stand  2    -270   520    0     80   0    
15:37:45.342 D523.0    stand  2    -270   510    0     80   10   
15:37:46.348 D523.0    stand  2    -310   490    0     80   44   
15:37:47.344 D523.0    stand  2    -310   490    66    80   0    
15:37:48.237 D523.0    stand  2    -280   480    0     80   31   
15:37:49.234 D523.0    stand  2    -270   480    0     80   10   
15:37:50.240 D523.0    stand  2    -270   480    0     80   0    
15:37:51.238 D523.0    stand  2    -270   480    0     80   0    
15:37:52.238 D523.0    stand  2    -270   480    0     80   0    
15:37:53.238 D523.0    stand  2    -270   480    0     80   0    
15:37:54.244 D523.0    stand  2    -270   480    0     80   0    
15:37:55.243 D523.0    stand  2    -270   480    0     80   0    
15:37:56.241 D523.0    stand  2    -270   480    0     80   0    
15:37:57.242 D523.0    stand  2    -270   480    0     80   0    
15:37:58.252 D523.0    stand  2    -270   480    0     80   0    
15:37:59.249 D523.0    stand  2    -270   480    0     80   0    
15:38:00.141 D523.0    stand  2    -270   480    0     80   0    
15:38:01.144 D523.0    stand  2    -270   480    0     80   0    
15:38:02.140 D523.0    stand  2    -270   480    0     80   0    
15:38:03.142 D523.0    stand  2    -270   480    0     80   0    
15:38:04.142 D523.0    stand  2    -270   480    0     80   0    
15:38:05.152 D523.0    stand  2    -270   480    0     80   0    
15:38:06.144 D523.0    stand  2    -270   480    0     80   0    
15:38:07.152 D523.0    stand  2    -270   480    0     80   0    
15:38:08.147 D523.0    stand  2    -270   480    0     80   0    
15:38:09.148 D523.0    stand  2    -270   490    0     80   10   
15:38:10.150 D523.0    stand  2    -270   490    0     80   0    
15:38:11.154 D523.0    stand  2    -270   490    0     80   0    
15:38:12.041 D523.0    stand  2    -270   490    0     80   0    
15:38:13.042 D523.0    stand  2    -270   490    0     80   0    
15:38:14.044 D523.0    stand  2    -270   490    0     80   0    
15:38:15.044 D523.0    stand  2    -270   490    0     80   0    
15:38:16.056 D523.0    stand  2    -270   490    0     80   0    
15:38:17.068 D523.0    stand  2    -270   490    0     80   0    
15:38:18.065 D523.0    stand  2    -270   490    0     80   0    
15:38:19.074 D523.0    stand  2    -270   490    0     80   0    
15:38:20.068 D523.0    stand  2    -280   490    0     80   10   
15:38:21.073 D523.0    stand  2    -290   500    0     80   14   
15:38:21.967 D523.0    stand  2    -300   500    0     80   10   
15:38:22.962 D523.0    stand  2    -300   500    0     80   0    
15:38:23.965 D523.0    stand  2    -300   500    0     80   0    
15:38:24.965 D523.0    stand  2    -300   500    0     80   0    
15:38:25.968 D523.0    stand  2    -300   500    0     80   0    
15:38:26.968 D523.0    stand  2    -300   500    0     80   0    
15:38:27.969 D523.0    stand  2    -300   500    0     80   0    
15:38:28.969 D523.0    stand  2    -300   500    0     80   0    
15:38:29.971 D523.0    stand  2    -300   500    0     80   0    
15:38:30.971 D523.0    stand  2    -300   500    0     80   0    
15:38:31.977 D523.0    stand  2    -300   500    0     80   0    
15:38:32.977 D523.0    stand  2    -300   500    0     80   0    
15:38:33.868 D523.0    stand  2    -300   500    0     80   0    
15:38:34.870 D523.0    stand  2    -300   500    0     80   0    
15:38:35.879 D523.0    stand  2    -300   500    0     80   0    
15:38:36.870 D523.0    stand  2    -300   500    0     80   0    
15:38:37.922 D523.0    stand  2    -300   500    0     80   0    
15:38:38.869 D523.0    stand  2    -300   500    0     80   0    
15:38:39.872 D523.0    stand  2    -300   500    0     80   0    
15:38:40.873 D523.0    stand  2    -300   500    0     80   0    
15:38:41.876 D523.0    stand  2    -300   500    0     80   0    
15:38:42.874 D523.0    stand  2    -300   500    0     80   0    
15:38:43.876 D523.0    stand  2    -300   500    0     80   0    
15:38:44.880 D523.0    stand  2    -300   500    0     80   0    
15:38:45.769 D523.0    stand  2    -300   500    0     80   0    
15:38:46.771 D523.0    stand  2    -300   500    0     80   0    
15:38:47.771 D523.0    stand  2    -300   500    0     80   0    
15:38:48.780 D523.0    stand  2    -300   500    0     80   0    
15:38:49.773 D523.0    stand  2    -300   500    0     80   0    
15:38:50.778 D523.0    stand  2    -300   500    0     80   0    
15:38:51.780 D523.0    stand  2    -300   500    0     80   0    
15:38:52.777 D523.0    stand  2    -300   500    0     80   0    
15:38:53.777 D523.0    stand  2    -300   500    0     80   0    
15:38:54.781 D523.0    stand  2    -300   500    0     80   0    
15:38:55.782 D523.0    stand  2    -300   500    0     80   0    
15:38:56.784 D523.0    stand  2    -300   500    0     80   0    
15:38:57.676 D523.0    stand  2    -300   500    0     80   0    
15:38:58.674 D523.0    stand  2    -300   500    0     80   0    
15:38:59.676 D523.0    stand  2    -300   500    0     80   0    
15:39:00.677 D523.0    stand  2    -300   500    0     80   0    
15:39:01.679 D523.0    stand  2    -300   500    0     80   0    
15:39:02.680 D523.0    stand  2    -300   500    0     80   0    
15:39:03.684 D523.0    stand  2    -300   490    0     80   10   
15:39:04.682 D523.0    stand  2    -280   490    0     80   20   
15:39:05.713 D523.0    stand  2    -290   500    0     80   14   
15:39:06.700 D523.0    stand  2    -290   510    0     80   10   
15:39:07.594 D523.0    stand  2    -300   500    0     80   14   
15:39:08.600 D523.0    stand  2    -320   500    0     80   20   
15:39:09.611 D523.0    stand  2    -330   500    0     80   10   
15:39:10.597 D523.0    stand  2    -330   500    0     80   0    
15:39:11.606 D523.0    stand  2    -330   500    0     80   0    
15:39:12.602 D523.0    stand  2    -330   500    0     80   0    
15:39:13.606 D523.0    stand  2    -330   500    0     80   0    
15:39:14.609 D523.0    stand  2    -330   500    0     80   0    
15:39:15.604 D523.0    stand  2    -330   500    0     80   0    
15:39:16.606 D523.0    stand  2    -330   500    0     80   0    
15:39:17.607 D523.0    stand  2    -330   500    0     80   0    
15:39:18.611 D523.0    stand  2    -330   500    0     80   0    
15:39:19.497 D523.0    stand  2    -330   500    0     80   0    
15:39:20.503 D523.0    stand  2    -330   500    0     80   0    
15:39:21.510 D523.0    stand  2    -330   500    0     80   0    
15:39:22.510 D523.0    stand  2    -330   500    0     80   0    
15:39:23.514 D523.0    stand  2    -310   510    0     80   22   
15:39:24.513 D523.0    stand  2    -290   510    0     80   20   
15:39:25.515 D523.0    stand  2    -290   510    0     80   0    
15:39:26.517 D523.0    stand  2    -290   510    0     80   0    
15:39:27.520 D523.0    stand  2    -290   510    0     80   0    
15:39:28.525 D523.0    stand  2    -290   510    0     80   0    
15:39:29.528 D523.0    stand  2    -290   510    0     80   0    
15:39:30.410 D523.0    stand  2    -290   510    0     80   0    
15:39:31.412 D523.0    stand  2    -290   510    0     80   0    
15:39:32.412 D523.0    stand  2    -290   510    0     80   0    
15:39:33.422 D523.0    stand  2    -290   510    0     80   0    
15:39:34.416 D523.0    stand  2    -290   510    0     80   0    
15:39:35.422 D523.0    stand  2    -290   510    0     80   0    
15:39:36.416 D523.0    stand  2    -290   510    0     80   0    
15:39:37.472 D523.0    stand  2    -290   510    0     80   0    
15:39:38.424 D523.0    stand  2    -290   510    0     80   0    
15:39:39.423 D523.0    stand  2    -290   510    0     80   0    
15:39:40.421 D523.0    stand  2    -290   510    0     80   0    
15:39:41.436 D523.0    stand  2    -290   510    0     80   0    
15:39:42.326 D523.0    stand  2    -290   510    0     80   0    
15:39:43.316 D523.0    stand  2    -290   510    0     80   0    
15:39:44.316 D523.0    stand  2    -290   510    0     80   0    
15:39:45.316 D523.0    stand  2    -290   510    0     80   0    
15:39:46.323 D523.0    stand  2    -290   510    0     80   0    
15:39:47.320 D523.0    stand  2    -290   510    0     80   0    
15:39:48.324 D523.0    stand  2    -290   510    0     80   0    
15:39:49.326 D523.0    stand  2    -290   510    0     80   0    
15:39:50.322 D523.0    stand  2    -290   510    0     80   0    
15:39:51.324 D523.0    stand  2    -290   510    0     80   0    
15:39:52.326 D523.0    stand  2    -290   510    0     80   0    
15:39:53.226 D523.0    stand  2    -290   510    0     80   0    
15:39:54.228 D523.0    stand  2    -290   510    0     80   0    
15:39:55.230 D523.0    stand  2    -290   510    0     80   0    
15:39:56.232 D523.0    stand  2    -290   510    0     80   0    
15:39:57.230 D523.0    stand  2    -290   510    0     80   0    
15:39:58.232 D523.0    stand  2    -290   510    0     80   0    
15:39:59.233 D523.0    stand  2    -300   510    0     80   10   
15:40:00.237 D523.0    stand  2    -290   480    0     80   31   
15:40:01.236 D523.0    stand  2    -290   480    0     80   0    
15:40:02.234 D523.0    stand  2    -290   480    0     80   0    
15:40:03.240 D523.0    stand  2    -280   490    0     80   14   
15:40:04.240 D523.0    stand  2    -280   490    0     80   0    
15:40:05.136 D523.0    stand  2    -300   470    0     80   28   
15:40:06.135 D523.0    stand  2    -280   480    0     80   22   
15:40:07.137 D523.0    stand  2    -290   470    0     80   14   
15:40:08.134 D523.0    stand  2    -290   470    0     80   0    
15:40:09.144 D523.0    stand  2    -290   470    0     80   0    
15:40:10.150 D523.0    stand  2    -290   470    0     80   0    
15:40:11.147 D523.0    stand  2    -290   470    0     80   0    
15:40:12.148 D523.0    stand  2    -290   470    0     80   0    
15:40:13.152 D523.0    stand  2    -290   470    0     80   0    
15:40:14.150 D523.0    stand  2    -290   470    0     80   0    
15:40:15.155 D523.0    stand  2    -290   470    0     80   0    
15:40:16.048 D523.0    stand  2    -290   470    0     80   0    
15:40:17.045 D523.0    stand  2    -290   470    0     80   0    
15:40:18.045 D523.0    stand  2    -290   470    0     80   0    
15:40:19.052 D523.0    stand  2    -290   470    0     80   0    
15:40:20.063 D523.0    stand  2    -290   470    0     80   0    
15:40:21.050 D523.0    stand  2    -290   470    0     80   0    
15:40:22.048 D523.0    stand  2    -290   470    0     80   0    
15:40:23.050 D523.0    stand  2    -290   470    0     80   0    
15:40:24.052 D523.0    stand  2    -290   470    0     80   0    
15:40:25.053 D523.0    stand  2    -290   470    0     80   0    
15:40:26.053 D523.0    stand  2    -290   470    0     80   0    
15:40:27.057 D523.0    stand  2    -290   470    0     80   0    
15:40:27.949 D523.0    stand  2    -290   470    0     80   0    
15:40:28.950 D523.0    stand  2    -290   470    0     80   0    
15:40:29.948 D523.0    stand  2    -290   470    0     80   0    
15:40:30.952 D523.0    stand  2    -290   470    0     80   0    
15:40:31.950 D523.0    stand  2    -290   470    0     80   0    
15:40:32.954 D523.0    stand  2    -290   470    0     80   0    
15:40:33.960 D523.0    stand  2    -290   470    0     80   0    
15:40:34.960 D523.0    stand  2    -290   470    0     80   0    
15:40:35.954 D523.0    stand  2    -290   470    0     80   0    
15:40:37.008 D523.0    stand  2    -290   470    0     80   0    
15:40:37.963 D523.0    stand  2    -290   470    0     80   0    
15:40:38.957 D523.0    stand  2    -290   470    0     80   0    
15:40:39.851 D523.0    stand  2    -290   470    0     80   0    
15:40:40.857 D523.0    stand  2    -290   470    0     80   0    
15:40:41.869 D523.0    stand  2    -290   470    0     80   0    
15:40:42.854 D523.0    stand  2    -290   470    0     80   0    
15:40:43.862 D523.0    stand  2    -290   470    0     80   0    
15:40:44.856 D523.0    stand  2    -280   480    105   80   14   
15:40:45.858 D523.0    stand  2    -280   470    0     80   10   
15:40:46.858 D523.0    stand  2    -280   480    0     80   10   
15:40:47.859 D523.0    stand  2    -280   480    0     80   0    
15:40:48.860 D523.0    stand  2    -280   480    0     80   0    
15:40:49.868 D523.0    stand  2    -280   480    0     80   0    
15:40:50.864 D523.0    stand  2    -280   480    0     80   0    
15:40:51.754 D523.0    stand  2    -280   480    0     80   0    
15:40:52.755 D523.0    stand  2    -280   480    0     80   0    
15:40:53.759 D523.0    stand  2    -280   480    0     80   0    
15:40:54.762 D523.0    stand  2    -280   480    0     80   0    
15:40:55.756 D523.0    stand  2    -280   480    0     80   0    
15:40:56.758 D523.0    stand  2    -280   480    0     80   0    
15:40:57.778 D523.0    stand  2    -280   480    0     80   0    
15:40:58.784 D523.0    stand  2    -280   480    0     80   0    

15:35:30.264 D5F7.88   88     -    -      -      -     -    -    
15:36:02.004 D5F7.88   88     -    -      -      -     -    -    
15:36:33.981 D5F7.88   88     -    -      -      -     -    -    
15:37:05.410 D5F7.88   88     -    -      -      -     -    -    
15:37:37.241 D5F7.88   88     -    -      -      -     -    -    
15:38:08.912 D5F7.88   88     -    -      -      -     -    -    
15:38:40.826 D5F7.88   88     -    -      -      -     -    -    
15:39:12.321 D5F7.88   88     -    -      -      -     -    -    
15:39:44.416 D5F7.88   88     -    -      -      -     -    -    
15:40:15.792 D5F7.88   88     -    -      -      -     -    -    
15:40:48.007 D5F7.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 386 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

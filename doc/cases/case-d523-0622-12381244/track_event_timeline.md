# case-d523-0622-12381244 — 卧室(09E7+D523 双雷达同房) 每 tick belief 时间线

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
12:38:00 D523.0   D52303800989  stand   122  NoReport stand              trk  0.50 Empty      1   0     0.00  0.03  0.08  0.00  0.85  0.04
12:38:01 D523.0   D52303800989  walk    104  NoReport walk               trk  0.89 Empty      1   0     0.00  0.04  0.10  0.00  0.82  0.01
12:38:02 D523.0   D52303800989  walk    87   NoReport walk               trk  0.98 Empty      1   0     0.00  0.04  0.11  0.01  0.76  0.01
12:38:03 D523.0   D52303800989  walk    115  NoReport walk               trk  0.99 Empty      1   0     0.00  0.05  0.12  0.01  0.71  0.01
12:38:04 D523.0   D52303800989  walk    95   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.13  0.02  0.66  0.01
12:38:05 D523.0   D52303800989  walk    97   NoReport walk               trk  1.00 Empty      1   0     0.00  0.06  0.14  0.03  0.62  0.01
12:38:06 D523.0   D52303800989  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.01  0.06  0.14  0.04  0.58  0.01
12:38:07 D523.0   D52303800989  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.01  0.07  0.15  0.05  0.51  0.02
12:38:08 D523.0   D52303800989  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.01  0.07  0.15  0.06  0.49  0.02
12:38:09 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.08  0.15  0.06  0.46  0.02
12:38:10 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.08  0.15  0.07  0.44  0.02
12:38:11 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.02  0.08  0.15  0.07  0.42  0.02
12:38:12 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.02  0.08  0.15  0.08  0.40  0.02
12:38:13 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.02  0.09  0.15  0.08  0.38  0.02
12:38:14 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.02  0.09  0.15  0.09  0.36  0.02
12:38:15 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.03  0.09  0.15  0.09  0.35  0.02
12:38:16 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.03  0.09  0.15  0.10  0.34  0.02
12:38:17 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.03  0.09  0.15  0.10  0.33  0.02
12:38:18 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.03  0.10  0.15  0.10  0.31  0.02
12:38:19 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.03  0.10  0.16  0.10  0.30  0.02
12:38:20 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.04  0.10  0.16  0.11  0.30  0.02
12:38:21 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.04  0.10  0.15  0.11  0.29  0.02
12:38:22 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.04  0.10  0.15  0.11  0.28  0.02
12:38:23 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.04  0.10  0.15  0.11  0.27  0.02
12:38:24 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.05  0.10  0.15  0.12  0.27  0.02
12:38:25 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.05  0.10  0.15  0.12  0.26  0.02
12:38:26 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.05  0.10  0.15  0.12  0.25  0.02
12:38:27 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.05  0.10  0.15  0.12  0.25  0.02
12:38:28 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.06  0.10  0.15  0.12  0.25  0.02
12:38:29 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.06  0.10  0.15  0.12  0.24  0.02
12:38:30 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.06  0.10  0.15  0.12  0.24  0.02
12:38:31 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.06  0.10  0.15  0.12  0.23  0.02
12:38:32 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.07  0.10  0.15  0.12  0.23  0.02
12:38:33 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.07  0.10  0.15  0.12  0.23  0.02
12:38:34 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   27    0.07  0.10  0.15  0.13  0.22  0.02
12:38:35 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   28    0.07  0.10  0.15  0.13  0.22  0.02
12:38:36 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   29    0.07  0.10  0.15  0.13  0.22  0.02
12:38:37 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   30    0.08  0.11  0.15  0.13  0.22  0.02
12:38:38 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   31    0.08  0.11  0.15  0.13  0.21  0.02
12:38:39 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   32    0.08  0.11  0.15  0.13  0.21  0.02
12:38:40 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   33    0.08  0.11  0.15  0.13  0.21  0.02
12:38:41 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   34    0.09  0.11  0.15  0.13  0.21  0.02
12:38:42 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   35    0.09  0.11  0.15  0.13  0.21  0.02
12:38:43 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   36    0.09  0.10  0.15  0.13  0.21  0.02
12:38:44 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   37    0.09  0.10  0.15  0.13  0.20  0.02
12:38:45 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   38    0.09  0.10  0.15  0.13  0.20  0.02
12:38:46 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   39    0.10  0.10  0.15  0.13  0.20  0.02
12:38:47 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   40    0.10  0.10  0.15  0.13  0.20  0.02
12:38:48 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   41    0.10  0.10  0.15  0.13  0.20  0.02
12:38:49 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   42    0.10  0.10  0.15  0.13  0.20  0.02
12:38:50 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   43    0.10  0.10  0.15  0.13  0.20  0.02
12:38:51 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   44    0.10  0.10  0.15  0.13  0.20  0.02
12:38:52 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   45    0.11  0.10  0.15  0.13  0.20  0.02
12:38:53 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   46    0.11  0.10  0.15  0.13  0.20  0.02
12:38:54 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   47    0.11  0.10  0.15  0.13  0.19  0.02
12:38:55 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   48    0.11  0.10  0.15  0.13  0.19  0.02
12:38:56 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   49    0.11  0.10  0.15  0.13  0.19  0.02
12:38:57 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   50    0.12  0.10  0.15  0.13  0.19  0.02
12:38:58 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   51    0.12  0.10  0.15  0.13  0.19  0.02
12:38:59 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   52    0.12  0.10  0.15  0.13  0.19  0.02
12:39:00 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   53    0.12  0.10  0.15  0.13  0.19  0.02
12:39:01 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   54    0.12  0.10  0.15  0.13  0.19  0.02
12:39:02 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   55    0.12  0.10  0.14  0.13  0.19  0.02
12:39:03 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   56    0.13  0.10  0.14  0.13  0.19  0.02
12:39:04 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   57    0.13  0.10  0.14  0.13  0.19  0.02
12:39:05 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   58    0.13  0.10  0.14  0.13  0.19  0.02
12:39:06 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   59    0.13  0.10  0.14  0.13  0.19  0.02
12:39:07 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   60    0.13  0.10  0.14  0.12  0.19  0.02
12:39:08 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   61    0.14  0.10  0.14  0.12  0.19  0.02
12:39:09 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   62    0.14  0.10  0.14  0.12  0.19  0.02
12:39:10 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   63    0.14  0.10  0.14  0.12  0.19  0.02
12:39:11 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   64    0.14  0.10  0.14  0.12  0.18  0.02
12:39:12 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   65    0.14  0.10  0.14  0.12  0.18  0.02
12:39:13 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   66    0.14  0.10  0.14  0.12  0.18  0.02
12:39:14 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   67    0.14  0.10  0.14  0.12  0.18  0.02
12:39:15 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   68    0.14  0.10  0.14  0.12  0.18  0.02
12:39:16 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   69    0.15  0.10  0.14  0.12  0.18  0.02
12:39:17 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   70    0.15  0.10  0.14  0.12  0.18  0.02
12:39:18 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   71    0.15  0.10  0.14  0.12  0.18  0.02
12:39:19 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   72    0.15  0.10  0.14  0.12  0.18  0.02
12:39:20 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   73    0.15  0.10  0.14  0.12  0.18  0.02
12:39:21 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   74    0.15  0.10  0.14  0.12  0.18  0.02
12:39:22 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   75    0.15  0.10  0.14  0.12  0.18  0.02
12:39:23 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   76    0.16  0.10  0.14  0.12  0.18  0.02
12:39:24 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   77    0.16  0.10  0.14  0.12  0.18  0.02
12:39:25 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   78    0.16  0.10  0.14  0.12  0.18  0.02
12:39:26 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   79    0.16  0.10  0.14  0.12  0.18  0.02
12:39:27 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   80    0.16  0.10  0.14  0.12  0.18  0.02
12:39:28 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   81    0.16  0.10  0.14  0.12  0.18  0.02
12:39:29 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   82    0.16  0.10  0.14  0.12  0.18  0.02
12:39:30 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   83    0.16  0.10  0.14  0.12  0.18  0.02
12:39:31 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   84    0.16  0.10  0.14  0.12  0.18  0.02
12:39:32 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   85    0.17  0.10  0.14  0.12  0.18  0.02
12:39:33 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   86    0.17  0.10  0.14  0.12  0.18  0.02
12:39:34 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   87    0.17  0.10  0.14  0.12  0.18  0.02
12:39:35 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   88    0.17  0.10  0.14  0.12  0.18  0.02
12:39:36 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   89    0.17  0.10  0.14  0.12  0.18  0.02
12:39:37 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   90    0.17  0.10  0.14  0.12  0.18  0.02
12:39:38 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   91    0.17  0.10  0.14  0.12  0.18  0.02
12:39:39 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   92    0.17  0.10  0.14  0.12  0.18  0.02
12:39:40 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   93    0.17  0.10  0.14  0.12  0.18  0.02
12:39:41 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Empty      1   94    0.17  0.10  0.14  0.12  0.18  0.02
12:39:42 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   95    0.18  0.10  0.14  0.12  0.18  0.02
12:39:43 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   96    0.18  0.10  0.14  0.12  0.18  0.02
12:39:44 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   97    0.18  0.10  0.14  0.12  0.17  0.02
12:39:45 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   98    0.18  0.10  0.14  0.12  0.17  0.02
12:39:46 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   99    0.18  0.10  0.14  0.12  0.17  0.02
12:39:47 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   100   0.18  0.10  0.14  0.12  0.17  0.02
12:39:48 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   101   0.18  0.10  0.14  0.12  0.17  0.02
12:39:49 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   102   0.18  0.10  0.14  0.12  0.17  0.02
12:39:50 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   103   0.18  0.10  0.14  0.12  0.17  0.02
12:39:51 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   104   0.18  0.10  0.14  0.12  0.17  0.02
12:39:52 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   105   0.18  0.10  0.14  0.12  0.17  0.02
12:39:53 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   106   0.19  0.09  0.14  0.12  0.17  0.02
12:39:54 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   107   0.19  0.09  0.14  0.12  0.17  0.02
12:39:55 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   108   0.19  0.09  0.14  0.12  0.17  0.02
12:39:55 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   109   0.19  0.09  0.14  0.12  0.17  0.02
12:39:56 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   110   0.19  0.09  0.14  0.12  0.17  0.02
12:39:57 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   111   0.19  0.09  0.14  0.12  0.17  0.02
12:39:58 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   112   0.19  0.09  0.14  0.12  0.17  0.02
12:39:59 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   113   0.19  0.09  0.14  0.12  0.17  0.02
12:40:00 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   114   0.19  0.09  0.14  0.12  0.17  0.02
12:40:01 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   115   0.19  0.09  0.14  0.12  0.17  0.02
12:40:02 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   116   0.19  0.09  0.13  0.12  0.17  0.02
12:40:03 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   117   0.19  0.09  0.13  0.11  0.17  0.02
12:40:04 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   118   0.19  0.09  0.13  0.11  0.17  0.02
12:40:05 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   119   0.20  0.09  0.13  0.11  0.17  0.02
12:40:07 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   120   0.20  0.09  0.13  0.11  0.17  0.02
12:40:07 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   120   0.20  0.09  0.13  0.11  0.17  0.02
12:40:08 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   121   0.20  0.09  0.13  0.11  0.17  0.02
12:40:09 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   122   0.20  0.09  0.13  0.11  0.17  0.02
12:40:10 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   123   0.20  0.09  0.13  0.11  0.17  0.02
12:40:11 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   124   0.20  0.09  0.13  0.11  0.17  0.02
12:40:12 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   125   0.20  0.09  0.13  0.11  0.17  0.02
12:40:13 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   126   0.20  0.09  0.13  0.11  0.17  0.02
12:40:14 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   127   0.20  0.09  0.13  0.11  0.17  0.02
12:40:15 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   128   0.20  0.09  0.13  0.11  0.17  0.02
12:40:16 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   129   0.20  0.09  0.13  0.11  0.17  0.02
12:40:17 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   130   0.20  0.09  0.13  0.11  0.17  0.02
12:40:18 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   131   0.20  0.09  0.13  0.11  0.17  0.02
12:40:19 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   132   0.21  0.09  0.13  0.11  0.17  0.02
12:40:20 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   133   0.21  0.09  0.13  0.11  0.17  0.02
12:40:21 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   134   0.21  0.09  0.13  0.11  0.17  0.02
12:40:22 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   135   0.21  0.09  0.13  0.11  0.17  0.02
12:40:23 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   136   0.21  0.09  0.13  0.11  0.17  0.02
12:40:24 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   137   0.21  0.09  0.13  0.11  0.17  0.02
12:40:25 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   138   0.21  0.09  0.13  0.11  0.17  0.02
12:40:26 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   139   0.21  0.09  0.13  0.11  0.17  0.02
12:40:27 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   140   0.21  0.09  0.13  0.11  0.17  0.02
12:40:28 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   141   0.21  0.09  0.13  0.11  0.17  0.02
12:40:29 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   142   0.21  0.09  0.13  0.11  0.17  0.02
12:40:30 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   143   0.21  0.09  0.13  0.11  0.17  0.02
12:40:31 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   144   0.21  0.09  0.13  0.11  0.17  0.02
12:40:32 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   145   0.21  0.09  0.13  0.11  0.17  0.02
12:40:33 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   146   0.21  0.09  0.13  0.11  0.17  0.02
12:40:34 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   147   0.21  0.09  0.13  0.11  0.17  0.02
12:40:35 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   148   0.21  0.09  0.13  0.11  0.17  0.02
12:40:36 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   149   0.21  0.09  0.13  0.11  0.17  0.02
12:40:37 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   150   0.21  0.09  0.13  0.11  0.17  0.02
12:40:38 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   151   0.21  0.09  0.13  0.11  0.17  0.02
12:40:39 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   152   0.22  0.09  0.13  0.11  0.17  0.02
12:40:40 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   153   0.22  0.09  0.13  0.11  0.17  0.02
12:40:41 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   154   0.22  0.09  0.13  0.11  0.17  0.02
12:40:42 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   155   0.22  0.09  0.13  0.11  0.17  0.02
12:40:43 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   156   0.22  0.09  0.13  0.11  0.17  0.02
12:40:44 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   157   0.22  0.09  0.13  0.11  0.17  0.02
12:40:45 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   158   0.22  0.09  0.13  0.11  0.17  0.02
12:40:46 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   159   0.22  0.09  0.13  0.11  0.17  0.02
12:40:47 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   160   0.22  0.09  0.13  0.11  0.17  0.02
12:40:48 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   161   0.22  0.09  0.13  0.11  0.17  0.02
12:40:49 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   162   0.22  0.09  0.13  0.11  0.17  0.02
12:40:50 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   163   0.22  0.09  0.13  0.11  0.16  0.02
12:40:51 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   164   0.22  0.09  0.13  0.11  0.16  0.02
12:40:52 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   165   0.22  0.09  0.13  0.11  0.16  0.02
12:40:53 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   166   0.22  0.09  0.13  0.11  0.16  0.02
12:40:54 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   167   0.22  0.09  0.13  0.11  0.16  0.02
12:40:55 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   168   0.22  0.09  0.13  0.11  0.16  0.02
12:40:56 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   169   0.22  0.09  0.13  0.11  0.16  0.02
12:40:57 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   170   0.22  0.09  0.13  0.11  0.16  0.02
12:40:58 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   171   0.22  0.09  0.13  0.11  0.16  0.02
12:40:59 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   172   0.22  0.09  0.13  0.11  0.16  0.02
12:41:00 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   173   0.22  0.09  0.13  0.11  0.16  0.02
12:41:01 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   174   0.22  0.09  0.13  0.11  0.16  0.02
12:41:02 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   175   0.22  0.09  0.13  0.11  0.16  0.02
12:41:03 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   176   0.22  0.09  0.13  0.11  0.16  0.02
12:41:04 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   177   0.22  0.09  0.13  0.11  0.16  0.02
12:41:05 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   178   0.22  0.09  0.13  0.11  0.16  0.02
12:41:06 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   179   0.23  0.09  0.13  0.11  0.16  0.02
12:41:07 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   180   0.23  0.09  0.13  0.11  0.16  0.02
12:41:08 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   181   0.23  0.09  0.13  0.11  0.16  0.02
12:41:09 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   182   0.23  0.09  0.13  0.11  0.16  0.02
12:41:10 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   183   0.23  0.09  0.13  0.11  0.16  0.02
12:41:11 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   184   0.23  0.09  0.13  0.11  0.16  0.02
12:41:12 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   185   0.23  0.09  0.13  0.11  0.16  0.02
12:41:13 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   186   0.23  0.09  0.13  0.11  0.16  0.02
12:41:14 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   187   0.23  0.09  0.13  0.11  0.16  0.02
12:41:15 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   188   0.23  0.09  0.13  0.11  0.16  0.02
12:41:16 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   189   0.23  0.09  0.13  0.11  0.16  0.02
12:41:17 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   190   0.23  0.09  0.13  0.11  0.16  0.02
12:41:18 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   191   0.23  0.09  0.13  0.11  0.16  0.02
12:41:19 D523.0   D52303800989  stand   0    NoReport stand              trk  1.00 Fallen     1   192   0.23  0.09  0.13  0.11  0.16  0.02
12:43:47 0978.E   -             -       0    InBed    InBed(pad)         room -    Fallen     0   0     0.23  0.09  0.13  0.11  0.16  0.02
12:43:47 0978.E   -             -       0    InBed    InBed(pad)         room -    Fallen     0   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:19 0978.E   -             -       0    LeftBed  LeftBed(pad)       room -    Fallen     0   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:19 0978.E   -             -       0    LeftBed  LeftBed(pad)       room -    Fallen     0   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:28 D523.0   D52303800989  walk    64   LeftBed  walk               trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:29 D523.0   D52303800989  walk    84   LeftBed  walk               trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:30 D523.0   D52303800989  walk    0    LeftBed  walk               trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:31 D523.0   D52303800989  walk    0    LeftBed  walk               trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:32 D523.0   D52303800989  walk    0    LeftBed  walk               trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:33 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:34 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:35 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:36 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:37 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:38 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.23  0.09  0.13  0.11  0.16  0.02
12:44:39 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:40 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:41 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:42 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:43 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:44 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:45 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:46 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:47 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:48 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:49 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:50 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:51 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:52 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:53 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:54 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:55 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:56 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:57 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
12:44:58 D523.0   D52303800989  stand   0    LeftBed  stand              trk  1.00 Fallen     1   0     0.24  0.09  0.13  0.11  0.16  0.02
```

**汇总**: xray tick 254 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

---

## 机制分析（手动，2026-06-22；按规则 #3 查机制非 fire）

**实况（ground truth）**：9e7 掉线，同房仅 D523+sleepad；人走进 D523 视野、**摔倒**、D523 帧**冻结**、sleepad **InBed**。

### 发现 1（真 bug，已修）：skip-room 太粗，一台离线拖垮整房

bootstrap declare-area 原按**房**跳：room 100 的 9e7 掉线 → declare-area 超时 → **整个 room 100 被跳过** → 在线的 D523（241 帧真数据）+ sleepad **全不监控**，摔倒分析根本没发生（修前 D523 在 xray 出现 0 帧）。
**修复**：`registerAllRooms` 改 **per-radar 容错**——某台 declare-area 拉不到只丢**那台床区贡献**，不 `continue` 整房；房只要还有任一在线雷达/sleepad 就照常注册。修后 D523 进管线 **254 帧**。

### 发现 2：D523 固件把摔冻成站立（冻结假人，上游）

D523 的 241 帧 pose = **{4 stand:219, 1 walk:13, None:9}，零 pose=5（fall）**。人摔在视野里、固件冻在 pose=4 站立 → SFall 钉死 0.18~0.23、永不到 0.85。**冻结轨识别**待补：Δ=0 持续 + conf 不降 = 无效 pose。

### 发现 3：MM 里 D523 对床三通道全断（坏配置）

| 矩阵项 | 值 | 因 |
|---|---|---|
| Matrix_D523–Bed（covers） | **0** | D523 canvas 没画床 → `RadarBedReachCount=0` |
| Matrix_D523–sleepad（samebed） | **0** | covers⊗onbed = 0⊗1 = 0（pad_samebed=0） |
| D523 固件 bedAreaIDs | **`[]`** | declare-area 空 → N(MN) 永不命中 |

根因 = 配置坏的 D523（[[install_model_dual_source_firmware_drift]] / [[mm_per_device_covers_ownership]]）。物理上 D523 看得见床，但 canvas 没画 → MM 如实反映"无关系"。

### 发现 4（slice-2 吸纳机制生效）：sleepad 撑住占用

D523 丢/冻摔倒的人后（present=false 仍 top=Fallen 0.23，无 exit→不抹不压=留疑），**sleepad InBed → 因 samebed=0 未被吸纳 → uncovered+1 → real_people 0→1**：雷达丢了摔倒的人、**sleepad 接住"房里有人"，引擎不瞎** = 正是缺陷②修复。

### 发现 5（Q4）：InBed 不停 stillbox；豁免在 floor 层、D523 无床几何用不上

D523 lost 轨 still_sec 一路涨 194→**341(InBed)**→373，**InBed 不停计数**（lost 续算 lostStillCarryMs），只在 12:44:28 重抓移动轨才 reset 0。
机制：InBed 豁免在 **floor 开火层**（`contactInBed`=近床∧InBed→不发），**不在 stillbox 计数层**。而 **D523 covers=0/无床几何 → NearBedMask false → `contactInBed` 对 D523 不生效** → 若 still 涨到 720s，floor 会**误火**（sleepad 说在床、D523 用不上这豁免）。= 发现 3 床三通道全断的连锁后果。

### 残留 / 下一步

- **冻结轨识别**：Δ=0 持续 N 秒 + conf 不降 → 判无效 pose/lost，别当活"站立"占着还不报。
- **slice-3 learned 层**：D523 几何 covers=0（坏配置），但 30s 同向 Beta-Bernoulli 能从共现事件把 samebed(D523↔sleepad) 学回来 → 坏配置靠 learned 补（强动机）。
- **跨雷达对账**（case 11571200 wart）：同房一台新鲜观测应能解析/压另一台的 stale 盲区 faller。

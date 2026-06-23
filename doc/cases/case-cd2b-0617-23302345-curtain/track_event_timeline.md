# case-cd2b-0617-23302345-curtain — 卧室(09E7+D523 双雷达同房) 每 tick belief 时间线

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
23:30:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  0.50 Empty      1   0     0.00  0.05  0.14  0.00  0.77  0.04
23:30:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  0.51 Empty      1   1     0.00  0.10  0.23  0.00  0.61  0.01
23:30:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  0.52 Empty      1   2     0.00  0.15  0.31  0.01  0.45  0.01
23:30:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  0.53 OpenFloor  1   3     0.00  0.20  0.35  0.02  0.30  0.02
23:30:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  0.54 OpenFloor  1   4     0.00  0.26  0.38  0.03  0.20  0.02
23:30:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  0.55 OpenFloor  1   4     0.00  0.30  0.39  0.03  0.13  0.02
23:30:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 OpenFloor  1   5     0.00  0.34  0.38  0.04  0.09  0.02
23:30:06 -.-      -             -       -    InBed    (no frame, held)   room -    OpenFloor  1   5     0.00  0.34  0.38  0.04  0.09  0.02
23:30:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 OpenFloor  1   6     0.00  0.37  0.38  0.05  0.06  0.02
23:30:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   7     0.00  0.40  0.36  0.05  0.05  0.02
23:30:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   8     0.00  0.42  0.35  0.05  0.04  0.02
23:30:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   9     0.00  0.44  0.34  0.05  0.03  0.02
23:30:10 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   9     0.00  0.44  0.34  0.05  0.03  0.02
23:30:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   10    0.00  0.46  0.33  0.06  0.03  0.02
23:30:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   11    0.00  0.47  0.32  0.06  0.02  0.02
23:30:12 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   11    0.00  0.47  0.32  0.06  0.02  0.02
23:30:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   12    0.00  0.48  0.32  0.06  0.02  0.02
23:30:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   13    0.00  0.49  0.31  0.06  0.02  0.02
23:30:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   14    0.00  0.50  0.31  0.06  0.02  0.02
23:30:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   15    0.00  0.50  0.30  0.06  0.02  0.02
23:30:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   16    0.00  0.51  0.30  0.06  0.02  0.02
23:30:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   17    0.00  0.51  0.29  0.06  0.02  0.02
23:30:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   18    0.00  0.52  0.29  0.07  0.02  0.02
23:30:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   19    0.00  0.52  0.29  0.07  0.02  0.02
23:30:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   20    0.00  0.52  0.29  0.07  0.02  0.02
23:30:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   21    0.00  0.52  0.29  0.07  0.02  0.02
23:30:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   22    0.00  0.52  0.29  0.07  0.02  0.02
23:30:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   23    0.00  0.53  0.29  0.07  0.02  0.02
23:30:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   24    0.00  0.53  0.28  0.07  0.02  0.02
23:30:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   25    0.00  0.53  0.28  0.07  0.02  0.02
23:30:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   26    0.00  0.53  0.28  0.07  0.02  0.02
23:30:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   27    0.00  0.53  0.28  0.07  0.02  0.02
23:30:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   28    0.00  0.53  0.28  0.07  0.02  0.02
23:30:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   29    0.00  0.53  0.28  0.07  0.02  0.02
23:30:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   30    0.00  0.53  0.28  0.07  0.02  0.02
23:30:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   31    0.00  0.53  0.28  0.07  0.02  0.02
23:30:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   32    0.00  0.53  0.28  0.07  0.02  0.02
23:30:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   33    0.00  0.53  0.28  0.07  0.02  0.02
23:30:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   34    0.00  0.53  0.28  0.07  0.02  0.02
23:30:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   35    0.00  0.53  0.28  0.07  0.02  0.02
23:30:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   36    0.00  0.53  0.28  0.07  0.02  0.02
23:30:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   37    0.00  0.53  0.28  0.07  0.02  0.02
23:30:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   38    0.00  0.53  0.28  0.07  0.02  0.02
23:30:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   39    0.00  0.53  0.28  0.07  0.02  0.02
23:30:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   40    0.00  0.53  0.28  0.07  0.02  0.02
23:30:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   41    0.00  0.53  0.28  0.07  0.02  0.02
23:30:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   42    0.00  0.53  0.28  0.07  0.02  0.02
23:30:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   43    0.00  0.53  0.28  0.07  0.02  0.02
23:30:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   44    0.00  0.53  0.28  0.07  0.02  0.02
23:30:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   45    0.00  0.53  0.28  0.07  0.02  0.02
23:30:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   46    0.00  0.53  0.28  0.07  0.02  0.02
23:30:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   47    0.00  0.53  0.28  0.07  0.02  0.02
23:30:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   48    0.00  0.53  0.28  0.07  0.02  0.02
23:30:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   49    0.00  0.53  0.28  0.07  0.02  0.02
23:30:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   50    0.00  0.53  0.28  0.07  0.02  0.02
23:30:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   51    0.00  0.53  0.28  0.07  0.02  0.02
23:30:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   52    0.00  0.53  0.28  0.07  0.02  0.02
23:30:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   53    0.00  0.53  0.28  0.07  0.02  0.02
23:30:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   54    0.00  0.53  0.28  0.07  0.02  0.02
23:30:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   55    0.00  0.53  0.28  0.07  0.02  0.02
23:30:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   56    0.00  0.53  0.28  0.07  0.02  0.02
23:30:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   57    0.00  0.53  0.28  0.07  0.02  0.02
23:30:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   58    0.00  0.53  0.28  0.07  0.02  0.02
23:30:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   59    0.00  0.53  0.28  0.07  0.02  0.02
23:31:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   60    0.00  0.53  0.28  0.07  0.02  0.02
23:31:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   61    0.00  0.53  0.28  0.07  0.02  0.02
23:31:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   62    0.00  0.53  0.28  0.07  0.02  0.02
23:31:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   63    0.00  0.53  0.28  0.07  0.02  0.02
23:31:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   64    0.00  0.53  0.28  0.07  0.02  0.02
23:31:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   65    0.00  0.53  0.28  0.07  0.02  0.02
23:31:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   66    0.00  0.53  0.28  0.07  0.02  0.02
23:31:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   67    0.00  0.53  0.28  0.07  0.02  0.02
23:31:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   68    0.00  0.53  0.28  0.07  0.02  0.02
23:31:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:10 CD2B.0   CD2B03000092  stand   64   InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:11 CD2B.0   CD2B03000092  stand   8    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:13 CD2B.0   CD2B03000092  stand   52   InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:16 CD2B.0   CD2B03000092  stand   82   InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.53  0.28  0.07  0.02  0.02
23:31:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   27    0.00  0.53  0.28  0.07  0.02  0.02
23:31:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   28    0.00  0.53  0.28  0.07  0.02  0.02
23:31:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   29    0.00  0.53  0.28  0.07  0.02  0.02
23:31:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   30    0.00  0.53  0.28  0.07  0.02  0.02
23:31:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   31    0.00  0.53  0.28  0.07  0.02  0.02
23:31:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   32    0.00  0.53  0.28  0.07  0.02  0.02
23:31:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   33    0.00  0.53  0.28  0.07  0.02  0.02
23:31:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   34    0.00  0.53  0.28  0.07  0.02  0.02
23:31:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   35    0.00  0.53  0.28  0.07  0.02  0.02
23:31:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   36    0.00  0.53  0.28  0.07  0.02  0.02
23:31:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   37    0.00  0.53  0.28  0.07  0.02  0.02
23:31:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   38    0.00  0.53  0.28  0.07  0.02  0.02
23:31:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   39    0.00  0.53  0.28  0.07  0.02  0.02
23:31:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   40    0.00  0.53  0.28  0.07  0.02  0.02
23:31:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   41    0.00  0.53  0.28  0.07  0.02  0.02
23:31:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   43    0.00  0.53  0.28  0.07  0.02  0.02
23:31:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   43    0.00  0.53  0.28  0.07  0.02  0.02
23:31:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   44    0.00  0.53  0.28  0.07  0.02  0.02
23:31:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   45    0.00  0.53  0.28  0.07  0.02  0.02
23:31:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   46    0.00  0.53  0.28  0.07  0.02  0.02
23:31:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   47    0.00  0.53  0.28  0.07  0.02  0.02
23:31:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   48    0.00  0.53  0.28  0.07  0.02  0.02
23:31:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   49    0.00  0.53  0.28  0.07  0.02  0.02
23:31:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   50    0.00  0.53  0.28  0.07  0.02  0.02
23:31:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   51    0.00  0.53  0.28  0.07  0.02  0.02
23:31:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   52    0.00  0.53  0.28  0.07  0.02  0.02
23:31:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   53    0.00  0.53  0.28  0.07  0.02  0.02
23:31:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   54    0.00  0.53  0.28  0.07  0.02  0.02
23:31:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   55    0.00  0.53  0.28  0.07  0.02  0.02
23:32:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   56    0.00  0.53  0.28  0.07  0.02  0.02
23:32:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   57    0.00  0.53  0.28  0.07  0.02  0.02
23:32:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   58    0.00  0.53  0.28  0.07  0.02  0.02
23:32:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   59    0.00  0.53  0.28  0.07  0.02  0.02
23:32:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   60    0.00  0.53  0.28  0.07  0.02  0.02
23:32:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   61    0.00  0.53  0.28  0.07  0.02  0.02
23:32:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   62    0.00  0.53  0.28  0.07  0.02  0.02
23:32:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   63    0.00  0.53  0.28  0.07  0.02  0.02
23:32:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   64    0.00  0.53  0.28  0.07  0.02  0.02
23:32:08 CD2B.0   CD2B03000092  stand   55   InBed    stand              trk  1.00 Bed        1   65    0.00  0.53  0.28  0.07  0.02  0.02
23:32:09 CD2B.0   CD2B03000092  stand   26   InBed    stand              trk  1.00 Bed        1   66    0.00  0.53  0.28  0.07  0.02  0.02
23:32:10 CD2B.0   CD2B03000092  stand   15   InBed    stand              trk  1.00 Bed        1   67    0.00  0.53  0.28  0.07  0.02  0.02
23:32:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   68    0.00  0.53  0.28  0.07  0.02  0.02
23:32:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   69    0.00  0.53  0.28  0.07  0.02  0.02
23:32:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   70    0.00  0.53  0.28  0.07  0.02  0.02
23:32:14 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   70    0.00  0.53  0.28  0.07  0.02  0.02
23:32:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   71    0.00  0.53  0.28  0.07  0.02  0.02
23:32:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   72    0.00  0.53  0.28  0.07  0.02  0.02
23:32:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   73    0.00  0.53  0.28  0.07  0.02  0.02
23:32:17 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   73    0.00  0.53  0.28  0.07  0.02  0.02
23:32:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   74    0.00  0.53  0.28  0.07  0.02  0.02
23:32:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   75    0.00  0.53  0.28  0.07  0.02  0.02
23:32:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   76    0.00  0.53  0.28  0.07  0.02  0.02
23:32:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   77    0.00  0.53  0.28  0.07  0.02  0.02
23:32:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   78    0.00  0.53  0.28  0.07  0.02  0.02
23:32:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   79    0.00  0.53  0.28  0.07  0.02  0.02
23:32:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   80    0.00  0.53  0.28  0.07  0.02  0.02
23:32:24 CD2B.0   CD2B03000092  stand   40   InBed    stand              trk  1.00 Bed        1   81    0.00  0.53  0.28  0.07  0.02  0.02
23:32:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   82    0.00  0.53  0.28  0.07  0.02  0.02
23:32:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   83    0.00  0.53  0.28  0.07  0.02  0.02
23:32:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   84    0.00  0.53  0.28  0.07  0.02  0.02
23:32:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   85    0.00  0.53  0.28  0.07  0.02  0.02
23:32:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   86    0.00  0.53  0.28  0.07  0.02  0.02
23:32:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   87    0.00  0.53  0.28  0.07  0.02  0.02
23:32:31 CD2B.E   -             -       0    InBed    np=1               room -    Bed        1   87    0.00  0.53  0.28  0.07  0.02  0.02
23:32:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   88    0.00  0.53  0.28  0.07  0.02  0.02
23:32:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   89    0.00  0.53  0.28  0.07  0.02  0.02
23:32:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   90    0.00  0.53  0.28  0.07  0.02  0.02
23:32:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   91    0.00  0.53  0.28  0.07  0.02  0.02
23:32:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   92    0.00  0.53  0.28  0.07  0.02  0.02
23:32:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   93    0.00  0.53  0.28  0.07  0.02  0.02
23:32:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   94    0.00  0.53  0.28  0.07  0.02  0.02
23:32:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   95    0.00  0.53  0.28  0.07  0.02  0.02
23:32:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   96    0.00  0.53  0.28  0.07  0.02  0.02
23:32:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   97    0.00  0.53  0.28  0.07  0.02  0.02
23:32:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   98    0.00  0.53  0.28  0.07  0.02  0.02
23:32:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   99    0.00  0.53  0.28  0.07  0.02  0.02
23:32:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   100   0.00  0.53  0.28  0.07  0.02  0.02
23:32:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   101   0.00  0.53  0.28  0.07  0.02  0.02
23:32:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   102   0.00  0.53  0.28  0.07  0.02  0.02
23:32:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   103   0.00  0.53  0.28  0.07  0.02  0.02
23:32:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   104   0.00  0.53  0.28  0.07  0.02  0.02
23:32:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   105   0.00  0.53  0.28  0.07  0.02  0.02
23:32:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   106   0.00  0.53  0.28  0.07  0.02  0.02
23:32:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   107   0.00  0.53  0.28  0.07  0.02  0.02
23:32:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   108   0.00  0.53  0.28  0.07  0.02  0.02
23:32:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   109   0.00  0.53  0.28  0.07  0.02  0.02
23:32:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   110   0.00  0.53  0.28  0.07  0.02  0.02
23:32:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   111   0.00  0.53  0.28  0.07  0.02  0.02
23:32:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   112   0.00  0.53  0.28  0.07  0.02  0.02
23:32:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   113   0.00  0.53  0.28  0.07  0.02  0.02
23:32:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   114   0.00  0.53  0.28  0.07  0.02  0.02
23:32:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   115   0.00  0.53  0.28  0.07  0.02  0.02
23:32:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   116   0.00  0.53  0.28  0.07  0.02  0.02
23:33:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   117   0.00  0.53  0.28  0.07  0.02  0.02
23:33:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   118   0.00  0.53  0.28  0.07  0.02  0.02
23:33:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   119   0.00  0.53  0.28  0.07  0.02  0.02
23:33:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   120   0.00  0.53  0.28  0.07  0.02  0.02
23:33:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   121   0.00  0.53  0.28  0.07  0.02  0.02
23:33:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   122   0.00  0.53  0.28  0.07  0.02  0.02
23:33:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   123   0.00  0.53  0.28  0.07  0.02  0.02
23:33:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   123   0.00  0.53  0.28  0.07  0.02  0.02
23:33:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   124   0.00  0.53  0.28  0.07  0.02  0.02
23:33:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   125   0.00  0.53  0.28  0.07  0.02  0.02
23:33:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   126   0.00  0.53  0.28  0.07  0.02  0.02
23:33:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   127   0.00  0.53  0.28  0.07  0.02  0.02
23:33:12 CD2B.0   CD2B03000092  stand   14   InBed    stand              trk  1.00 Bed        1   129   0.00  0.53  0.28  0.07  0.02  0.02
23:33:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   130   0.00  0.53  0.28  0.07  0.02  0.02
23:33:14 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   130   0.00  0.53  0.28  0.07  0.02  0.02
23:33:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   131   0.00  0.53  0.28  0.07  0.02  0.02
23:33:15 CD2B.0   CD2B03000092  stand   67   InBed    stand              trk  1.00 Bed        1   132   0.00  0.53  0.28  0.07  0.02  0.02
23:33:16 CD2B.0   CD2B03000092  stand   9    InBed    stand              trk  1.00 Bed        1   133   0.00  0.53  0.28  0.07  0.02  0.02
23:33:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   133   0.00  0.53  0.28  0.07  0.02  0.02
23:33:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   134   0.00  0.53  0.28  0.07  0.02  0.02
23:33:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   135   0.00  0.53  0.28  0.07  0.02  0.02
23:33:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   136   0.00  0.53  0.28  0.07  0.02  0.02
23:33:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   137   0.00  0.53  0.28  0.07  0.02  0.02
23:33:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   138   0.00  0.53  0.28  0.07  0.02  0.02
23:33:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   139   0.00  0.53  0.28  0.07  0.02  0.02
23:33:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   140   0.00  0.53  0.28  0.07  0.02  0.02
23:33:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   141   0.00  0.53  0.28  0.07  0.02  0.02
23:33:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   142   0.00  0.53  0.28  0.07  0.02  0.02
23:33:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   143   0.00  0.53  0.28  0.07  0.02  0.02
23:33:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   144   0.00  0.53  0.28  0.07  0.02  0.02
23:33:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:33 CD2B.0   CD2B03000092  stand   84   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:39 CD2B.0   CD2B03000092  stand   38   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:33:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:14 CD2B.0   CD2B03000092  stand   47   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:22 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:23 CD2B.0   CD2B03000092  stand   102  InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:23 CD2B.0   CD2B03000092  stand   7    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:29 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:34:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:25 CD2B.0   CD2B03000092  stand   2    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:35:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:19 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:36:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:11 CD2B.0   CD2B03000092  stand   68   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:12 CD2B.0   CD2B03000092  stand   50   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:13 CD2B.0   CD2B03000092  walk    50   InBed    walk               trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:14 CD2B.0   CD2B03000092  walk    56   InBed    walk               trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:15 CD2B.0   CD2B03000092  walk    51   InBed    walk               trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:16 CD2B.0   CD2B03000092  walk    15   InBed    walk               trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:17 CD2B.0   CD2B03000092  walk    0    InBed    walk               trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:25 CD2B.0   CD2B03000092  stand   54   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:27 CD2B.0   CD2B03000092  stand   75   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:48 CD2B.0   CD2B03000092  stand   2    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:49 CD2B.0   CD2B03000092  stand   23   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:37:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:28 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:31 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:33 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:38:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:49 -.-      -             -       -    InBed    (no frame, held)   room -    Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:39:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:34 CD2B.0   CD2B03000092  stand   6    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:40:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:41:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:42:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:43:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:11 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:27 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:44:59 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:00 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:01 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:02 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:03 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:04 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:05 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:06 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:07 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:08 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:09 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:10 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:11 CD2B.0   CD2B03000092  stand   10   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:12 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:13 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:14 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:15 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:16 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:17 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:17 CD2B.E   -             -       0    InBed    LeftBed(rdr)       room -    Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:18 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:19 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:20 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:21 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:22 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:23 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:24 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:25 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:26 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:27 CD2B.0   CD2B03000092  stand   14   InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:28 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:29 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:30 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:31 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:32 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:33 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:34 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:35 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:36 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:37 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:38 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:39 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:40 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:41 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:42 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:43 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:44 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:45 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:46 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:47 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:48 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:49 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:50 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:51 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:52 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:53 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:54 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:55 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:56 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:57 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
23:45:58 CD2B.0   CD2B03000092  stand   0    InBed    stand              trk  1.00 Bed        1   145   0.00  0.53  0.28  0.07  0.02  0.02
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
23:30:00.092 CD2B.0    stand  1    -40    160    0     80        
23:30:01.118 CD2B.0    stand  1    -40    160    0     80   0    
23:30:02.093 CD2B.0    stand  1    -40    160    0     80   0    
23:30:03.097 CD2B.0    stand  1    -40    160    0     80   0    
23:30:04.095 CD2B.0    stand  1    -40    160    0     80   0    
23:30:04.991 CD2B.0    stand  1    -40    160    0     80   0    
23:30:05.989 CD2B.0    stand  1    -40    160    0     80   0    
23:30:07.000 CD2B.0    stand  1    -40    160    0     80   0    
23:30:07.993 CD2B.0    stand  1    -40    160    0     80   0    
23:30:08.998 CD2B.0    stand  1    -40    160    0     80   0    
23:30:09.995 CD2B.0    stand  1    -40    160    0     80   0    
23:30:11.001 CD2B.0    stand  1    -40    160    0     80   0    
23:30:11.997 CD2B.0    stand  1    -40    160    0     80   0    
23:30:13.001 CD2B.0    stand  1    -40    160    0     80   0    
23:30:14.008 CD2B.0    stand  1    -40    160    0     80   0    
23:30:14.998 CD2B.0    stand  1    -40    160    0     80   0    
23:30:15.909 CD2B.0    stand  1    -40    160    0     80   0    
23:30:16.900 CD2B.0    stand  1    -40    160    0     80   0    
23:30:17.903 CD2B.0    stand  1    -40    160    0     80   0    
23:30:18.904 CD2B.0    stand  1    -40    160    0     80   0    
23:30:19.907 CD2B.0    stand  1    -40    160    0     80   0    
23:30:20.910 CD2B.0    stand  1    -40    160    0     80   0    
23:30:21.919 CD2B.0    stand  1    -40    160    0     80   0    
23:30:22.907 CD2B.0    stand  1    -40    160    0     80   0    
23:30:23.910 CD2B.0    stand  1    -40    160    0     80   0    
23:30:24.929 CD2B.0    stand  1    -40    160    0     80   0    
23:30:25.910 CD2B.0    stand  1    -40    160    0     80   0    
23:30:26.918 CD2B.0    stand  1    -40    160    0     80   0    
23:30:27.810 CD2B.0    stand  1    -40    160    0     80   0    
23:30:28.818 CD2B.0    stand  1    -40    160    0     80   0    
23:30:29.810 CD2B.0    stand  1    -40    160    0     80   0    
23:30:30.825 CD2B.0    stand  1    -40    160    0     80   0    
23:30:31.882 CD2B.0    stand  1    -40    160    0     80   0    
23:30:32.829 CD2B.0    stand  1    -40    160    0     80   0    
23:30:33.833 CD2B.0    stand  1    -40    160    0     80   0    
23:30:34.835 CD2B.0    stand  1    -40    160    0     80   0    
23:30:35.833 CD2B.0    stand  1    -40    160    0     80   0    
23:30:36.753 CD2B.0    stand  1    -40    160    0     80   0    
23:30:37.732 CD2B.0    stand  1    -40    160    0     80   0    
23:30:38.779 CD2B.0    stand  1    -40    160    0     80   0    
23:30:39.736 CD2B.0    stand  1    -40    160    0     80   0    
23:30:40.807 CD2B.0    stand  1    -40    160    0     80   0    
23:30:41.738 CD2B.0    stand  1    -40    160    0     80   0    
23:30:42.736 CD2B.0    stand  1    -40    160    0     80   0    
23:30:43.737 CD2B.0    stand  1    -40    160    0     80   0    
23:30:44.752 CD2B.0    stand  1    -40    160    0     80   0    
23:30:45.741 CD2B.0    stand  1    -40    160    0     80   0    
23:30:46.785 CD2B.0    stand  1    -40    160    0     80   0    
23:30:47.743 CD2B.0    stand  1    -40    160    0     80   0    
23:30:48.735 CD2B.0    stand  1    -40    160    0     80   0    
23:30:49.640 CD2B.0    stand  1    -40    160    0     80   0    
23:30:50.640 CD2B.0    stand  1    -40    160    0     80   0    
23:30:51.644 CD2B.0    stand  1    -40    160    0     80   0    
23:30:52.668 CD2B.0    stand  1    -40    160    0     80   0    
23:30:53.640 CD2B.0    stand  1    -40    160    0     80   0    
23:30:54.660 CD2B.0    stand  1    -40    160    0     80   0    
23:30:55.644 CD2B.0    stand  1    -40    160    0     80   0    
23:30:56.650 CD2B.0    stand  1    -40    160    0     80   0    
23:30:57.657 CD2B.0    stand  1    -40    160    0     80   0    
23:30:58.647 CD2B.0    stand  1    -40    160    0     80   0    
23:30:59.662 CD2B.0    stand  1    -40    160    0     80   0    
23:31:00.539 CD2B.0    stand  1    -40    160    0     80   0    
23:31:01.544 CD2B.0    stand  1    -40    160    0     80   0    
23:31:02.575 CD2B.0    stand  1    -40    160    0     80   0    
23:31:03.548 CD2B.0    stand  1    -30    130    0     80   31   
23:31:04.548 CD2B.0    stand  1    -30    120    0     80   10   
23:31:05.549 CD2B.0    stand  1    -30    120    0     80   0    
23:31:06.548 CD2B.0    stand  1    -30    120    0     80   0    
23:31:07.546 CD2B.0    stand  1    -30    120    0     80   0    
23:31:08.589 CD2B.0    stand  1    -20    120    0     80   10   
23:31:09.549 CD2B.0    stand  1    0      90     0     80   36   
23:31:10.554 CD2B.0    stand  1    0      90     64    80   0    
23:31:11.552 CD2B.0    stand  1    0      110    8     80   20   
23:31:12.449 CD2B.0    stand  1    0      110    0     80   0    
23:31:13.445 CD2B.0    stand  1    0      110    52    80   0    
23:31:14.446 CD2B.0    stand  1    -10    120    0     80   14   
23:31:15.449 CD2B.0    stand  1    -10    120    0     80   0    
23:31:16.448 CD2B.0    stand  1    0      110    82    80   14   
23:31:17.449 CD2B.0    stand  1    0      100    0     80   10   
23:31:18.451 CD2B.0    stand  1    0      100    0     80   0    
23:31:19.456 CD2B.0    stand  1    0      100    0     80   0    
23:31:20.465 CD2B.0    stand  1    0      100    0     80   0    
23:31:21.468 CD2B.0    stand  1    0      100    0     80   0    
23:31:22.367 CD2B.0    stand  1    0      100    0     80   0    
23:31:23.366 CD2B.0    stand  1    0      100    0     80   0    
23:31:24.368 CD2B.0    stand  1    0      100    0     80   0    
23:31:25.370 CD2B.0    stand  1    0      100    0     80   0    
23:31:26.369 CD2B.0    stand  1    0      100    0     80   0    
23:31:27.368 CD2B.0    stand  1    0      100    0     80   0    
23:31:28.370 CD2B.0    stand  1    0      100    0     80   0    
23:31:29.376 CD2B.0    stand  1    0      100    0     80   0    
23:31:30.373 CD2B.0    stand  1    0      100    0     80   0    
23:31:31.421 CD2B.0    stand  1    0      100    0     80   0    
23:31:32.376 CD2B.0    stand  1    0      100    0     80   0    
23:31:33.380 CD2B.0    stand  1    0      100    0     80   0    
23:31:34.270 CD2B.0    stand  1    0      100    0     80   0    
23:31:35.291 CD2B.0    stand  1    0      100    0     80   0    
23:31:36.272 CD2B.0    stand  1    0      100    0     80   0    
23:31:37.274 CD2B.0    stand  1    0      100    0     80   0    
23:31:38.279 CD2B.0    stand  1    0      100    0     80   0    
23:31:39.312 CD2B.0    stand  1    0      100    0     80   0    
23:31:40.284 CD2B.0    stand  1    0      100    0     80   0    
23:31:41.281 CD2B.0    stand  1    0      100    0     80   0    
23:31:42.280 CD2B.0    stand  1    0      100    0     80   0    
23:31:43.280 CD2B.0    stand  1    0      100    0     80   0    
23:31:44.286 CD2B.0    stand  1    0      100    0     80   0    
23:31:45.180 CD2B.0    stand  1    0      100    0     80   0    
23:31:46.601 CD2B.0    stand  1    0      100    0     80   0    
23:31:47.181 CD2B.0    stand  1    0      100    0     80   0    
23:31:48.184 CD2B.0    stand  1    0      100    0     80   0    
23:31:49.193 CD2B.0    stand  1    0      100    0     80   0    
23:31:50.203 CD2B.0    stand  1    0      100    0     80   0    
23:31:51.187 CD2B.0    stand  1    0      100    0     80   0    
23:31:52.126 CD2B.0    stand  1    0      100    0     80   0    
23:31:53.125 CD2B.0    stand  1    0      100    0     80   0    
23:31:54.125 CD2B.0    stand  1    0      100    0     80   0    
23:31:55.148 CD2B.0    stand  1    0      100    0     80   0    
23:31:56.128 CD2B.0    stand  1    0      100    0     80   0    
23:31:57.134 CD2B.0    stand  1    0      120    0     80   20   
23:31:58.132 CD2B.0    stand  1    0      120    0     80   0    
23:31:59.168 CD2B.0    stand  1    0      120    0     80   0    
23:32:00.132 CD2B.0    stand  1    0      120    0     80   0    
23:32:01.133 CD2B.0    stand  1    0      120    0     80   0    
23:32:02.146 CD2B.0    stand  1    0      120    0     80   0    
23:32:03.140 CD2B.0    stand  1    0      120    0     80   0    
23:32:04.060 CD2B.0    stand  1    0      120    0     80   0    
23:32:05.029 CD2B.0    stand  1    0      120    0     80   0    
23:32:06.029 CD2B.0    stand  1    0      120    0     80   0    
23:32:07.053 CD2B.0    stand  1    0      120    0     80   0    
23:32:08.096 CD2B.0    stand  1    0      120    0     80   0    
23:32:08.990 CD2B.0    stand  1    0      100    55    80   20   
23:32:09.991 CD2B.0    stand  1    -20    130    26    80   36   
23:32:10.992 CD2B.0    stand  1    0      120    15    80   22   
23:32:11.994 CD2B.0    stand  1    10     100    0     80   22   
23:32:12.993 CD2B.0    stand  1    10     90     0     80   10   
23:32:13.996 CD2B.0    stand  1    0      110    0     80   22   
23:32:15.012 CD2B.0    stand  1    0      120    0     80   10   
23:32:16.001 CD2B.0    stand  1    0      120    0     80   0    
23:32:16.996 CD2B.0    stand  1    0      120    0     80   0    
23:32:18.006 CD2B.0    stand  1    0      120    0     80   0    
23:32:19.000 CD2B.0    stand  1    0      120    0     80   0    
23:32:20.000 CD2B.0    stand  1    0      120    0     80   0    
23:32:20.900 CD2B.0    stand  1    0      120    0     80   0    
23:32:21.910 CD2B.0    stand  1    0      120    0     80   0    
23:32:22.893 CD2B.0    stand  1    0      120    0     80   0    
23:32:23.908 CD2B.0    stand  1    0      120    0     80   0    
23:32:24.912 CD2B.0    stand  1    0      130    40    80   10   
23:32:25.912 CD2B.0    stand  1    0      100    0     80   30   
23:32:26.912 CD2B.0    stand  1    0      100    0     80   0    
23:32:27.912 CD2B.0    stand  1    0      100    0     80   0    
23:32:28.917 CD2B.0    stand  1    0      100    0     80   0    
23:32:29.913 CD2B.0    stand  1    -10    100    0     80   10   
23:32:30.820 CD2B.0    stand  1    -10    100    0     80   0    
23:32:31.879 CD2B.0    stand  1    -10    100    0     80   0    
23:32:32.828 CD2B.0    stand  1    -10    100    0     80   0    
23:32:33.825 CD2B.0    stand  1    -10    100    0     80   0    
23:32:34.831 CD2B.0    stand  1    -10    100    0     80   0    
23:32:35.826 CD2B.0    stand  1    -10    100    0     80   0    
23:32:36.828 CD2B.0    stand  1    -10    100    0     80   0    
23:32:37.835 CD2B.0    stand  1    -10    100    0     80   0    
23:32:38.830 CD2B.0    stand  1    -10    100    0     80   0    
23:32:39.744 CD2B.0    stand  1    -10    100    0     80   0    
23:32:40.802 CD2B.0    stand  1    -10    100    0     80   0    
23:32:41.745 CD2B.0    stand  1    -10    100    0     80   0    
23:32:42.749 CD2B.0    stand  1    -10    100    0     80   0    
23:32:43.757 CD2B.0    stand  1    -10    100    0     80   0    
23:32:44.749 CD2B.0    stand  1    -10    100    0     80   0    
23:32:45.746 CD2B.0    stand  1    -10    100    0     80   0    
23:32:46.750 CD2B.0    stand  1    -10    100    0     80   0    
23:32:47.750 CD2B.0    stand  1    -10    100    0     80   0    
23:32:48.750 CD2B.0    stand  1    -10    100    0     80   0    
23:32:49.756 CD2B.0    stand  1    -10    100    0     80   0    
23:32:50.800 CD2B.0    stand  1    -10    100    0     80   0    
23:32:51.651 CD2B.0    stand  1    -10    100    0     80   0    
23:32:52.646 CD2B.0    stand  1    -10    100    0     80   0    
23:32:53.645 CD2B.0    stand  1    -10    100    0     80   0    
23:32:54.647 CD2B.0    stand  1    -10    100    0     80   0    
23:32:55.729 CD2B.0    stand  1    -10    100    0     80   0    
23:32:56.615 CD2B.0    stand  1    -10    100    0     80   0    
23:32:57.672 CD2B.0    stand  1    -10    100    0     80   0    
23:32:58.621 CD2B.0    stand  1    -10    100    0     80   0    
23:32:59.661 CD2B.0    stand  1    -10    100    0     80   0    
23:33:00.625 CD2B.0    stand  1    -10    100    0     80   0    
23:33:01.635 CD2B.0    stand  1    -10    100    0     80   0    
23:33:02.626 CD2B.0    stand  1    -10    100    0     80   0    
23:33:03.627 CD2B.0    stand  1    -10    100    0     80   0    
23:33:04.622 CD2B.0    stand  1    -10    100    0     80   0    
23:33:05.655 CD2B.0    stand  1    -10    100    0     80   0    
23:33:06.631 CD2B.0    stand  1    -10    100    0     80   0    
23:33:07.518 CD2B.0    stand  1    -10    100    0     80   0    
23:33:08.519 CD2B.0    stand  1    -10    100    0     80   0    
23:33:09.522 CD2B.0    stand  1    -10    100    0     80   0    
23:33:10.519 CD2B.0    stand  1    -10    100    0     80   0    
23:33:11.526 CD2B.0    stand  1    -10    100    0     80   0    
23:33:12.548 CD2B.0    stand  1    -10    100    14    80   0    
23:33:13.549 CD2B.0    stand  1    -10    100    0     80   0    
23:33:15.080 CD2B.0    stand  1    -10    100    0     80   0    
23:33:15.559 CD2B.0    stand  1    0      100    67    80   10   
23:33:16.559 CD2B.0    stand  1    -10    100    9     80   10   
23:33:17.440 CD2B.0    stand  1    -10    120    0     80   20   
23:33:18.437 CD2B.0    stand  1    -10    130    0     80   10   
23:33:19.441 CD2B.0    stand  1    -10    130    0     80   0    
23:33:20.441 CD2B.0    stand  1    -10    130    0     80   0    
23:33:21.441 CD2B.0    stand  1    -10    130    0     80   0    
23:33:22.444 CD2B.0    stand  1    -10    130    0     80   0    
23:33:23.444 CD2B.0    stand  1    -10    130    0     80   0    
23:33:24.449 CD2B.0    stand  1    -10    130    0     80   0    
23:33:25.453 CD2B.0    stand  1    -10    130    0     80   0    
23:33:26.450 CD2B.0    stand  1    -10    130    0     80   0    
23:33:27.448 CD2B.0    stand  1    -10    130    0     80   0    
23:33:28.354 CD2B.0    stand  1    -10    130    0     80   0    
23:33:29.353 CD2B.0    stand  1    -10    130    0     80   0    
23:33:30.402 CD2B.0    stand  1    -10    130    0     80   0    
23:33:31.363 CD2B.0    stand  1    -10    130    0     80   0    
23:33:32.354 CD2B.0    stand  1    -10    120    0     80   10   
23:33:33.368 CD2B.0    stand  1    0      110    84    80   14   
23:33:34.379 CD2B.0    stand  1    0      130    0     80   20   
23:33:35.356 CD2B.0    stand  1    0      130    0     80   0    
23:33:36.360 CD2B.0    stand  1    0      130    0     80   0    
23:33:37.361 CD2B.0    stand  1    0      130    0     80   0    
23:33:38.383 CD2B.0    stand  1    0      130    0     80   0    
23:33:39.363 CD2B.0    stand  1    0      130    38    80   0    
23:33:40.301 CD2B.0    stand  1    -10    130    0     80   10   
23:33:41.261 CD2B.0    stand  1    -10    130    0     80   0    
23:33:42.259 CD2B.0    stand  1    -20    130    0     80   10   
23:33:43.259 CD2B.0    stand  1    -20    130    0     80   0    
23:33:44.256 CD2B.0    stand  1    -20    130    0     80   0    
23:33:45.259 CD2B.0    stand  1    -20    130    0     80   0    
23:33:46.266 CD2B.0    stand  1    -20    130    0     80   0    
23:33:47.264 CD2B.0    stand  1    -20    130    0     80   0    
23:33:48.265 CD2B.0    stand  1    -20    130    0     80   0    
23:33:49.280 CD2B.0    stand  1    -20    130    0     80   0    
23:33:50.261 CD2B.0    stand  1    -20    130    0     80   0    
23:33:51.266 CD2B.0    stand  1    -20    130    0     80   0    
23:33:52.161 CD2B.0    stand  1    -20    130    0     80   0    
23:33:53.158 CD2B.0    stand  1    -20    130    0     80   0    
23:33:54.159 CD2B.0    stand  1    -20    130    0     80   0    
23:33:55.166 CD2B.0    stand  1    -20    130    0     80   0    
23:33:56.162 CD2B.0    stand  1    -20    130    0     80   0    
23:33:57.167 CD2B.0    stand  1    -20    130    0     80   0    
23:33:58.164 CD2B.0    stand  1    -20    130    0     80   0    
23:33:59.168 CD2B.0    stand  1    -20    130    0     80   0    
23:34:00.176 CD2B.0    stand  1    -20    130    0     80   0    
23:34:01.177 CD2B.0    stand  1    -20    130    0     80   0    
23:34:02.176 CD2B.0    stand  1    -20    130    0     80   0    
23:34:03.067 CD2B.0    stand  1    -20    130    0     80   0    
23:34:04.069 CD2B.0    stand  1    -20    130    0     80   0    
23:34:05.072 CD2B.0    stand  1    -20    130    0     80   0    
23:34:06.072 CD2B.0    stand  1    -20    130    0     80   0    
23:34:07.081 CD2B.0    stand  1    -20    130    0     80   0    
23:34:08.075 CD2B.0    stand  1    -20    130    0     80   0    
23:34:09.078 CD2B.0    stand  1    -20    130    0     80   0    
23:34:10.075 CD2B.0    stand  1    -20    130    0     80   0    
23:34:11.081 CD2B.0    stand  1    -20    130    0     80   0    
23:34:12.084 CD2B.0    stand  1    -20    130    0     80   0    
23:34:13.078 CD2B.0    stand  1    -20    130    0     80   0    
23:34:14.080 CD2B.0    stand  1    -20    130    0     80   0    
23:34:14.976 CD2B.0    stand  1    -20    130    47    80   0    
23:34:15.986 CD2B.0    stand  1    -20    130    0     80   0    
23:34:16.989 CD2B.0    stand  1    10     110    0     80   36   
23:34:17.986 CD2B.0    stand  1    10     110    0     80   0    
23:34:18.988 CD2B.0    stand  1    10     110    0     80   0    
23:34:19.992 CD2B.0    stand  1    10     110    0     80   0    
23:34:20.992 CD2B.0    stand  1    10     110    0     80   0    
23:34:21.997 CD2B.0    stand  1    10     110    0     80   0    
23:34:23.001 CD2B.0    stand  1    0      110    102   80   10   
23:34:23.993 CD2B.0    stand  1    0      110    7     80   0    
23:34:24.894 CD2B.0    stand  1    0      110    0     80   0    
23:34:25.894 CD2B.0    stand  1    0      110    0     80   0    
23:34:26.899 CD2B.0    stand  1    0      110    0     80   0    
23:34:27.915 CD2B.0    stand  1    0      110    0     80   0    
23:34:28.902 CD2B.0    stand  1    0      110    0     80   0    
23:34:30.078 CD2B.0    stand  1    0      110    0     80   0    
23:34:30.900 CD2B.0    stand  1    0      110    0     80   0    
23:34:31.841 CD2B.0    stand  1    0      110    0     80   0    
23:34:32.843 CD2B.0    stand  1    0      110    0     80   0    
23:34:33.838 CD2B.0    stand  1    0      110    0     80   0    
23:34:34.841 CD2B.0    stand  1    0      110    0     80   0    
23:34:35.842 CD2B.0    stand  1    0      110    0     80   0    
23:34:36.843 CD2B.0    stand  1    0      110    0     80   0    
23:34:37.863 CD2B.0    stand  1    0      110    0     80   0    
23:34:38.844 CD2B.0    stand  1    0      110    0     80   0    
23:34:39.858 CD2B.0    stand  1    0      110    0     80   0    
23:34:40.846 CD2B.0    stand  1    0      110    0     80   0    
23:34:41.861 CD2B.0    stand  1    0      110    0     80   0    
23:34:42.849 CD2B.0    stand  1    0      110    0     80   0    
23:34:43.752 CD2B.0    stand  1    0      110    0     80   0    
23:34:44.746 CD2B.0    stand  1    0      110    0     80   0    
23:34:45.762 CD2B.0    stand  1    0      110    0     80   0    
23:34:46.744 CD2B.0    stand  1    0      110    0     80   0    
23:34:47.813 CD2B.0    stand  1    0      110    0     80   0    
23:34:48.704 CD2B.0    stand  1    0      110    0     80   0    
23:34:49.703 CD2B.0    stand  1    0      110    0     80   0    
23:34:50.706 CD2B.0    stand  1    0      110    0     80   0    
23:34:51.708 CD2B.0    stand  1    0      110    0     80   0    
23:34:52.706 CD2B.0    stand  1    0      110    0     80   0    
23:34:53.719 CD2B.0    stand  1    0      110    0     80   0    
23:34:54.709 CD2B.0    stand  1    0      110    0     80   0    
23:34:55.731 CD2B.0    stand  1    0      110    0     80   0    
23:34:56.712 CD2B.0    stand  1    0      110    0     80   0    
23:34:57.716 CD2B.0    stand  1    0      110    0     80   0    
23:34:58.711 CD2B.0    stand  1    0      110    0     80   0    
23:34:59.724 CD2B.0    stand  1    0      110    0     80   0    
23:35:00.624 CD2B.0    stand  1    0      110    0     80   0    
23:35:01.606 CD2B.0    stand  1    0      110    0     80   0    
23:35:02.606 CD2B.0    stand  1    0      110    0     80   0    
23:35:03.636 CD2B.0    stand  1    0      110    0     80   0    
23:35:04.625 CD2B.0    stand  1    0      110    0     80   0    
23:35:05.642 CD2B.0    stand  1    0      110    0     80   0    
23:35:06.629 CD2B.0    stand  1    0      110    0     80   0    
23:35:07.629 CD2B.0    stand  1    0      110    0     80   0    
23:35:08.633 CD2B.0    stand  1    0      110    0     80   0    
23:35:09.637 CD2B.0    stand  1    0      110    0     80   0    
23:35:10.547 CD2B.0    stand  1    0      110    0     80   0    
23:35:11.528 CD2B.0    stand  1    0      110    0     80   0    
23:35:12.545 CD2B.0    stand  1    0      110    0     80   0    
23:35:13.527 CD2B.0    stand  1    0      110    0     80   0    
23:35:14.534 CD2B.0    stand  1    0      110    0     80   0    
23:35:15.531 CD2B.0    stand  1    0      110    0     80   0    
23:35:16.537 CD2B.0    stand  1    0      110    0     80   0    
23:35:17.533 CD2B.0    stand  1    0      110    0     80   0    
23:35:18.533 CD2B.0    stand  1    0      110    0     80   0    
23:35:19.534 CD2B.0    stand  1    0      110    0     80   0    
23:35:20.454 CD2B.0    stand  1    0      110    0     80   0    
23:35:21.446 CD2B.0    stand  1    0      110    0     80   0    
23:35:22.452 CD2B.0    stand  1    0      110    0     80   0    
23:35:23.529 CD2B.0    stand  1    0      110    0     80   0    
23:35:24.449 CD2B.0    stand  1    0      110    0     80   0    
23:35:25.451 CD2B.0    stand  1    0      110    2     80   0    
23:35:26.450 CD2B.0    stand  1    0      110    0     80   0    
23:35:27.458 CD2B.0    stand  1    0      110    0     80   0    
23:35:28.459 CD2B.0    stand  1    0      110    0     80   0    
23:35:29.507 CD2B.0    stand  1    0      110    0     80   0    
23:35:30.457 CD2B.0    stand  1    0      110    0     80   0    
23:35:31.474 CD2B.0    stand  1    0      110    0     80   0    
23:35:32.348 CD2B.0    stand  1    0      110    0     80   0    
23:35:33.350 CD2B.0    stand  1    0      110    0     80   0    
23:35:34.351 CD2B.0    stand  1    0      110    0     80   0    
23:35:35.355 CD2B.0    stand  1    0      110    0     80   0    
23:35:36.321 CD2B.0    stand  1    0      110    0     80   0    
23:35:37.320 CD2B.0    stand  1    0      110    0     80   0    
23:35:38.322 CD2B.0    stand  1    0      110    0     80   0    
23:35:39.354 CD2B.0    stand  1    0      110    0     80   0    
23:35:40.321 CD2B.0    stand  1    0      110    0     80   0    
23:35:41.338 CD2B.0    stand  1    0      110    0     80   0    
23:35:42.324 CD2B.0    stand  1    0      110    0     80   0    
23:35:43.325 CD2B.0    stand  1    0      110    0     80   0    
23:35:44.326 CD2B.0    stand  1    0      110    0     80   0    
23:35:45.330 CD2B.0    stand  1    0      110    0     80   0    
23:35:46.329 CD2B.0    stand  1    0      110    0     80   0    
23:35:47.335 CD2B.0    stand  1    0      110    0     80   0    
23:35:48.222 CD2B.0    stand  1    0      110    0     80   0    
23:35:49.220 CD2B.0    stand  1    0      110    0     80   0    
23:35:50.230 CD2B.0    stand  1    0      110    0     80   0    
23:35:51.224 CD2B.0    stand  1    0      110    0     80   0    
23:35:52.267 CD2B.0    stand  1    0      110    0     80   0    
23:35:53.262 CD2B.0    stand  1    0      110    0     80   0    
23:35:54.261 CD2B.0    stand  1    0      110    0     80   0    
23:35:55.271 CD2B.0    stand  1    0      110    0     80   0    
23:35:56.162 CD2B.0    stand  1    0      110    0     80   0    
23:35:57.160 CD2B.0    stand  1    0      110    0     80   0    
23:35:58.165 CD2B.0    stand  1    0      110    0     80   0    
23:35:59.176 CD2B.0    stand  1    0      110    0     80   0    
23:36:00.175 CD2B.0    stand  1    0      110    0     80   0    
23:36:01.162 CD2B.0    stand  1    0      110    0     80   0    
23:36:02.164 CD2B.0    stand  1    0      110    0     80   0    
23:36:03.164 CD2B.0    stand  1    0      110    0     80   0    
23:36:04.185 CD2B.0    stand  1    0      110    0     80   0    
23:36:05.249 CD2B.0    stand  1    0      110    0     80   0    
23:36:06.166 CD2B.0    stand  1    0      110    0     80   0    
23:36:07.169 CD2B.0    stand  1    0      110    0     80   0    
23:36:08.065 CD2B.0    stand  1    0      110    0     80   0    
23:36:09.066 CD2B.0    stand  1    0      110    0     80   0    
23:36:10.138 CD2B.0    stand  1    0      110    0     80   0    
23:36:11.080 CD2B.0    stand  1    0      110    0     80   0    
23:36:12.084 CD2B.0    stand  1    0      110    0     80   0    
23:36:13.073 CD2B.0    stand  1    0      110    0     80   0    
23:36:14.072 CD2B.0    stand  1    0      110    0     80   0    
23:36:15.073 CD2B.0    stand  1    0      110    0     80   0    
23:36:16.087 CD2B.0    stand  1    0      110    0     80   0    
23:36:17.079 CD2B.0    stand  1    0      110    0     80   0    
23:36:18.120 CD2B.0    stand  1    0      110    0     80   0    
23:36:18.977 CD2B.0    stand  1    0      110    0     80   0    
23:36:20.008 CD2B.0    stand  1    0      110    0     80   0    
23:36:20.975 CD2B.0    stand  1    0      110    0     80   0    
23:36:21.982 CD2B.0    stand  1    0      110    0     80   0    
23:36:22.976 CD2B.0    stand  1    0      110    0     80   0    
23:36:23.937 CD2B.0    stand  1    0      110    0     80   0    
23:36:24.937 CD2B.0    stand  1    0      110    0     80   0    
23:36:25.933 CD2B.0    stand  1    0      110    0     80   0    
23:36:26.942 CD2B.0    stand  1    0      110    0     80   0    
23:36:27.938 CD2B.0    stand  1    0      110    0     80   0    
23:36:28.990 CD2B.0    stand  1    0      110    0     80   0    
23:36:29.949 CD2B.0    stand  1    0      110    0     80   0    
23:36:30.939 CD2B.0    stand  1    0      110    0     80   0    
23:36:31.946 CD2B.0    stand  1    0      110    0     80   0    
23:36:32.951 CD2B.0    stand  1    0      110    0     80   0    
23:36:33.941 CD2B.0    stand  1    0      110    0     80   0    
23:36:34.957 CD2B.0    stand  1    0      110    0     80   0    
23:36:35.839 CD2B.0    stand  1    0      110    0     80   0    
23:36:36.837 CD2B.0    stand  1    0      110    0     80   0    
23:36:37.849 CD2B.0    stand  1    0      110    0     80   0    
23:36:38.841 CD2B.0    stand  1    0      110    0     80   0    
23:36:39.893 CD2B.0    stand  1    0      110    0     80   0    
23:36:40.893 CD2B.0    stand  1    0      110    0     80   0    
23:36:41.856 CD2B.0    stand  1    0      110    0     80   0    
23:36:42.792 CD2B.0    stand  1    0      110    0     80   0    
23:36:43.790 CD2B.0    stand  1    0      110    0     80   0    
23:36:44.792 CD2B.0    stand  1    0      110    0     80   0    
23:36:45.792 CD2B.0    stand  1    0      110    0     80   0    
23:36:46.822 CD2B.0    stand  1    0      110    0     80   0    
23:36:47.798 CD2B.0    stand  1    0      110    0     80   0    
23:36:48.796 CD2B.0    stand  1    0      110    0     80   0    
23:36:49.806 CD2B.0    stand  1    0      110    0     80   0    
23:36:50.816 CD2B.0    stand  1    0      110    0     80   0    
23:36:51.799 CD2B.0    stand  1    0      110    0     80   0    
23:36:52.802 CD2B.0    stand  1    0      110    0     80   0    
23:36:53.700 CD2B.0    stand  1    0      110    0     80   0    
23:36:54.707 CD2B.0    stand  1    0      110    0     80   0    
23:36:55.704 CD2B.0    stand  1    0      110    0     80   0    
23:36:56.719 CD2B.0    stand  1    0      110    0     80   0    
23:36:57.709 CD2B.0    stand  1    0      110    0     80   0    
23:36:58.710 CD2B.0    stand  1    0      110    0     80   0    
23:36:59.716 CD2B.0    stand  1    0      110    0     80   0    
23:37:00.769 CD2B.0    stand  1    0      110    0     80   0    
23:37:01.713 CD2B.0    stand  1    0      110    0     80   0    
23:37:02.719 CD2B.0    stand  1    0      110    0     80   0    
23:37:03.613 CD2B.0    stand  1    0      110    0     80   0    
23:37:04.618 CD2B.0    stand  1    0      110    0     80   0    
23:37:05.613 CD2B.0    stand  1    0      110    0     80   0    
23:37:06.616 CD2B.0    stand  1    0      110    0     80   0    
23:37:07.620 CD2B.0    stand  1    0      120    0     80   10   
23:37:08.635 CD2B.0    stand  1    0      120    0     80   0    
23:37:09.620 CD2B.0    stand  1    0      120    0     80   0    
23:37:10.631 CD2B.0    stand  1    0      110    0     80   10   
23:37:11.622 CD2B.0    stand  1    0      120    68    80   10   
23:37:12.550 CD2B.0    stand  1    -10    130    50    80   14   
23:37:13.546 CD2B.0    walk   1    -20    140    50    80   14   
23:37:14.586 CD2B.0    walk   1    0      100    56    80   44   
23:37:15.558 CD2B.0    walk   1    0      100    51    80   0    
23:37:16.553 CD2B.0    walk   1    10     110    15    80   14   
23:37:17.550 CD2B.0    walk   1    10     120    0     80   10   
23:37:18.551 CD2B.0    stand  1    10     120    0     80   0    
23:37:19.553 CD2B.0    stand  1    10     120    0     80   0    
23:37:20.574 CD2B.0    stand  1    10     120    0     80   0    
23:37:21.553 CD2B.0    stand  1    10     120    0     80   0    
23:37:22.553 CD2B.0    stand  1    10     120    0     80   0    
23:37:23.464 CD2B.0    stand  1    10     120    0     80   0    
23:37:24.467 CD2B.0    stand  1    10     120    0     80   0    
23:37:25.459 CD2B.0    stand  1    0      110    54    80   14   
23:37:26.462 CD2B.0    stand  1    0      100    0     80   10   
23:37:27.462 CD2B.0    stand  1    0      130    75    80   30   
23:37:28.457 CD2B.0    stand  1    0      140    0     80   10   
23:37:29.845 CD2B.0    stand  1    0      140    0     80   0    
23:37:30.426 CD2B.0    stand  1    0      140    0     80   0    
23:37:31.418 CD2B.0    stand  1    0      140    0     80   0    
23:37:32.418 CD2B.0    stand  1    0      140    0     80   0    
23:37:33.419 CD2B.0    stand  1    0      140    0     80   0    
23:37:34.424 CD2B.0    stand  1    0      140    0     80   0    
23:37:35.422 CD2B.0    stand  1    0      140    0     80   0    
23:37:36.424 CD2B.0    stand  1    0      140    0     80   0    
23:37:37.422 CD2B.0    stand  1    0      140    0     80   0    
23:37:38.429 CD2B.0    stand  1    0      140    0     80   0    
23:37:39.333 CD2B.0    stand  1    0      140    0     80   0    
23:37:40.331 CD2B.0    stand  1    0      140    0     80   0    
23:37:41.342 CD2B.0    stand  1    0      140    0     80   0    
23:37:42.334 CD2B.0    stand  1    0      140    0     80   0    
23:37:43.334 CD2B.0    stand  1    0      140    0     80   0    
23:37:44.334 CD2B.0    stand  1    0      140    0     80   0    
23:37:45.343 CD2B.0    stand  1    0      140    0     80   0    
23:37:46.339 CD2B.0    stand  1    0      140    0     80   0    
23:37:47.336 CD2B.0    stand  1    0      130    0     80   10   
23:37:48.359 CD2B.0    stand  1    0      120    2     80   10   
23:37:49.345 CD2B.0    stand  1    -10    120    23    80   10   
23:37:50.241 CD2B.0    stand  1    0      110    0     80   14   
23:37:51.246 CD2B.0    stand  1    0      110    0     80   0    
23:37:52.245 CD2B.0    stand  1    0      110    0     80   0    
23:37:53.244 CD2B.0    stand  1    0      110    0     80   0    
23:37:54.245 CD2B.0    stand  1    0      110    0     80   0    
23:37:55.241 CD2B.0    stand  1    0      110    0     80   0    
23:37:56.246 CD2B.0    stand  1    0      110    0     80   0    
23:37:57.244 CD2B.0    stand  1    0      110    0     80   0    
23:37:58.270 CD2B.0    stand  1    0      110    0     80   0    
23:37:59.247 CD2B.0    stand  1    0      110    0     80   0    
23:38:00.157 CD2B.0    stand  1    0      110    0     80   0    
23:38:01.160 CD2B.0    stand  1    0      110    0     80   0    
23:38:02.160 CD2B.0    stand  1    0      110    0     80   0    
23:38:03.162 CD2B.0    stand  1    0      110    0     80   0    
23:38:04.166 CD2B.0    stand  1    0      110    0     80   0    
23:38:05.173 CD2B.0    stand  1    0      110    0     80   0    
23:38:06.164 CD2B.0    stand  1    0      110    0     80   0    
23:38:07.168 CD2B.0    stand  1    0      110    0     80   0    
23:38:08.165 CD2B.0    stand  1    0      110    0     80   0    
23:38:09.178 CD2B.0    stand  1    0      110    0     80   0    
23:38:10.173 CD2B.0    stand  1    0      110    0     80   0    
23:38:11.179 CD2B.0    stand  1    0      110    0     80   0    
23:38:12.082 CD2B.0    stand  1    0      110    0     80   0    
23:38:13.060 CD2B.0    stand  1    0      110    0     80   0    
23:38:14.061 CD2B.0    stand  1    0      110    0     80   0    
23:38:15.073 CD2B.0    stand  1    0      110    0     80   0    
23:38:16.030 CD2B.0    stand  1    0      110    0     80   0    
23:38:17.044 CD2B.0    stand  1    0      110    0     80   0    
23:38:18.036 CD2B.0    stand  1    0      110    0     80   0    
23:38:19.034 CD2B.0    stand  1    0      110    0     80   0    
23:38:20.044 CD2B.0    stand  1    0      110    0     80   0    
23:38:21.051 CD2B.0    stand  1    0      110    0     80   0    
23:38:22.036 CD2B.0    stand  1    0      110    0     80   0    
23:38:23.037 CD2B.0    stand  1    0      110    0     80   0    
23:38:24.036 CD2B.0    stand  1    0      110    0     80   0    
23:38:25.041 CD2B.0    stand  1    0      110    0     80   0    
23:38:26.053 CD2B.0    stand  1    0      110    0     80   0    
23:38:27.046 CD2B.0    stand  1    0      110    0     80   0    
23:38:27.940 CD2B.0    stand  1    0      110    0     80   0    
23:38:29.000 CD2B.0    stand  1    0      110    0     80   0    
23:38:29.943 CD2B.0    stand  1    0      110    0     80   0    
23:38:30.936 CD2B.0    stand  1    0      110    0     80   0    
23:38:32.002 CD2B.0    stand  1    0      110    0     80   0    
23:38:32.969 CD2B.0    stand  1    0      110    0     80   0    
23:38:34.015 CD2B.0    stand  1    0      110    0     80   0    
23:38:34.971 CD2B.0    stand  1    0      110    0     80   0    
23:38:35.905 CD2B.0    stand  1    0      110    0     80   0    
23:38:36.877 CD2B.0    stand  1    0      110    0     80   0    
23:38:37.885 CD2B.0    stand  1    0      110    0     80   0    
23:38:38.883 CD2B.0    stand  1    0      110    0     80   0    
23:38:39.884 CD2B.0    stand  1    0      110    0     80   0    
23:38:40.875 CD2B.0    stand  1    0      110    0     80   0    
23:38:41.883 CD2B.0    stand  1    0      110    0     80   0    
23:38:42.879 CD2B.0    stand  1    0      110    0     80   0    
23:38:43.918 CD2B.0    stand  1    0      110    0     80   0    
23:38:44.890 CD2B.0    stand  1    0      110    0     80   0    
23:38:45.903 CD2B.0    stand  1    0      110    0     80   0    
23:38:46.881 CD2B.0    stand  1    0      110    0     80   0    
23:38:47.781 CD2B.0    stand  1    0      110    0     80   0    
23:38:48.794 CD2B.0    stand  1    0      110    0     80   0    
23:38:49.813 CD2B.0    stand  1    0      110    0     80   0    
23:38:50.806 CD2B.0    stand  1    0      110    0     80   0    
23:38:51.800 CD2B.0    stand  1    0      110    0     80   0    
23:38:52.791 CD2B.0    stand  1    0      110    0     80   0    
23:38:53.792 CD2B.0    stand  1    0      110    0     80   0    
23:38:54.798 CD2B.0    stand  1    0      110    0     80   0    
23:38:55.799 CD2B.0    stand  1    0      110    0     80   0    
23:38:56.797 CD2B.0    stand  1    0      110    0     80   0    
23:38:57.693 CD2B.0    stand  1    0      110    0     80   0    
23:38:58.692 CD2B.0    stand  1    0      110    0     80   0    
23:38:59.698 CD2B.0    stand  1    0      110    0     80   0    
23:39:00.697 CD2B.0    stand  1    0      110    0     80   0    
23:39:01.696 CD2B.0    stand  1    0      110    0     80   0    
23:39:02.701 CD2B.0    stand  1    0      110    0     80   0    
23:39:03.699 CD2B.0    stand  1    0      110    0     80   0    
23:39:04.642 CD2B.0    stand  1    0      110    0     80   0    
23:39:05.638 CD2B.0    stand  1    0      110    0     80   0    
23:39:06.638 CD2B.0    stand  1    0      110    0     80   0    
23:39:07.646 CD2B.0    stand  1    0      110    0     80   0    
23:39:08.639 CD2B.0    stand  1    0      110    0     80   0    
23:39:09.647 CD2B.0    stand  1    0      110    0     80   0    
23:39:10.669 CD2B.0    stand  1    0      110    0     80   0    
23:39:11.653 CD2B.0    stand  1    0      110    0     80   0    
23:39:12.646 CD2B.0    stand  1    0      110    0     80   0    
23:39:13.645 CD2B.0    stand  1    0      110    0     80   0    
23:39:14.657 CD2B.0    stand  1    0      110    0     80   0    
23:39:15.649 CD2B.0    stand  1    0      110    0     80   0    
23:39:16.542 CD2B.0    stand  1    0      110    0     80   0    
23:39:17.543 CD2B.0    stand  1    0      110    0     80   0    
23:39:18.557 CD2B.0    stand  1    0      110    0     80   0    
23:39:19.548 CD2B.0    stand  1    0      110    0     80   0    
23:39:20.615 CD2B.0    stand  1    0      110    0     80   0    
23:39:21.503 CD2B.0    stand  1    0      110    0     80   0    
23:39:22.501 CD2B.0    stand  1    0      110    0     80   0    
23:39:23.509 CD2B.0    stand  1    0      110    0     80   0    
23:39:24.509 CD2B.0    stand  1    0      110    0     80   0    
23:39:25.520 CD2B.0    stand  1    0      110    0     80   0    
23:39:26.504 CD2B.0    stand  1    0      110    0     80   0    
23:39:27.610 CD2B.0    stand  1    0      110    0     80   0    
23:39:28.515 CD2B.0    stand  1    0      110    0     80   0    
23:39:29.521 CD2B.0    stand  1    0      110    0     80   0    
23:39:30.514 CD2B.0    stand  1    0      110    0     80   0    
23:39:31.524 CD2B.0    stand  1    0      110    0     80   0    
23:39:32.515 CD2B.0    stand  1    0      110    0     80   0    
23:39:33.405 CD2B.0    stand  1    0      110    0     80   0    
23:39:34.405 CD2B.0    stand  1    0      110    0     80   0    
23:39:35.417 CD2B.0    stand  1    0      110    0     80   0    
23:39:36.419 CD2B.0    stand  1    0      110    0     80   0    
23:39:37.416 CD2B.0    stand  1    0      110    0     80   0    
23:39:38.425 CD2B.0    stand  1    0      110    0     80   0    
23:39:39.421 CD2B.0    stand  1    0      110    0     80   0    
23:39:40.418 CD2B.0    stand  1    0      110    0     80   0    
23:39:41.420 CD2B.0    stand  1    0      110    0     80   0    
23:39:42.423 CD2B.0    stand  1    0      110    0     80   0    
23:39:43.474 CD2B.0    stand  1    0      110    0     80   0    
23:39:44.349 CD2B.0    stand  1    0      110    0     80   0    
23:39:45.323 CD2B.0    stand  1    0      110    0     80   0    
23:39:46.318 CD2B.0    stand  1    0      110    0     80   0    
23:39:47.324 CD2B.0    stand  1    0      110    0     80   0    
23:39:48.343 CD2B.0    stand  1    0      110    0     80   0    
23:39:50.102 CD2B.0    stand  1    0      110    0     80   0    
23:39:50.326 CD2B.0    stand  1    0      110    0     80   0    
23:39:51.326 CD2B.0    stand  1    0      110    0     80   0    
23:39:52.326 CD2B.0    stand  1    0      110    0     80   0    
23:39:53.327 CD2B.0    stand  1    0      110    0     80   0    
23:39:54.329 CD2B.0    stand  1    0      110    0     80   0    
23:39:55.330 CD2B.0    stand  1    0      110    0     80   0    
23:39:56.234 CD2B.0    stand  1    0      110    0     80   0    
23:39:57.224 CD2B.0    stand  1    0      110    0     80   0    
23:39:58.226 CD2B.0    stand  1    0      110    0     80   0    
23:39:59.225 CD2B.0    stand  1    0      110    0     80   0    
23:40:00.229 CD2B.0    stand  1    0      110    0     80   0    
23:40:01.231 CD2B.0    stand  1    0      110    0     80   0    
23:40:02.231 CD2B.0    stand  1    0      110    0     80   0    
23:40:03.233 CD2B.0    stand  1    0      110    0     80   0    
23:40:04.241 CD2B.0    stand  1    0      110    0     80   0    
23:40:05.233 CD2B.0    stand  1    0      110    0     80   0    
23:40:06.249 CD2B.0    stand  1    0      110    0     80   0    
23:40:07.234 CD2B.0    stand  1    0      110    0     80   0    
23:40:08.154 CD2B.0    stand  1    0      110    0     80   0    
23:40:09.125 CD2B.0    stand  1    0      110    0     80   0    
23:40:10.127 CD2B.0    stand  1    0      110    0     80   0    
23:40:11.129 CD2B.0    stand  1    0      110    0     80   0    
23:40:12.178 CD2B.0    stand  1    0      110    0     80   0    
23:40:13.137 CD2B.0    stand  1    0      110    0     80   0    
23:40:14.136 CD2B.0    stand  1    0      110    0     80   0    
23:40:15.137 CD2B.0    stand  1    0      110    0     80   0    
23:40:16.140 CD2B.0    stand  1    0      110    0     80   0    
23:40:17.146 CD2B.0    stand  1    0      110    0     80   0    
23:40:18.138 CD2B.0    stand  1    0      110    0     80   0    
23:40:19.135 CD2B.0    stand  1    0      110    0     80   0    
23:40:20.030 CD2B.0    stand  1    0      110    0     80   0    
23:40:21.028 CD2B.0    stand  1    0      110    0     80   0    
23:40:22.077 CD2B.0    stand  1    0      110    0     80   0    
23:40:23.030 CD2B.0    stand  1    0      110    0     80   0    
23:40:24.096 CD2B.0    stand  1    0      110    0     80   0    
23:40:25.054 CD2B.0    stand  1    0      110    0     80   0    
23:40:26.065 CD2B.0    stand  1    0      110    0     80   0    
23:40:27.105 CD2B.0    stand  1    0      110    0     80   0    
23:40:28.066 CD2B.0    stand  1    0      110    0     80   0    
23:40:29.052 CD2B.0    stand  1    0      110    0     80   0    
23:40:29.978 CD2B.0    stand  1    0      110    0     80   0    
23:40:30.953 CD2B.0    stand  1    0      110    0     80   0    
23:40:31.982 CD2B.0    stand  1    0      110    0     80   0    
23:40:32.957 CD2B.0    stand  1    0      110    0     80   0    
23:40:33.961 CD2B.0    stand  1    0      110    0     80   0    
23:40:34.954 CD2B.0    stand  1    0      110    6     80   0    
23:40:35.956 CD2B.0    stand  1    -10    110    0     80   10   
23:40:36.956 CD2B.0    stand  1    -10    110    0     80   0    
23:40:37.958 CD2B.0    stand  1    -10    110    0     80   0    
23:40:38.957 CD2B.0    stand  1    -10    110    0     80   0    
23:40:39.965 CD2B.0    stand  1    -10    110    0     80   0    
23:40:40.972 CD2B.0    stand  1    -10    110    0     80   0    
23:40:41.853 CD2B.0    stand  1    -10    110    0     80   0    
23:40:42.853 CD2B.0    stand  1    -10    110    0     80   0    
23:40:43.854 CD2B.0    stand  1    -10    110    0     80   0    
23:40:44.863 CD2B.0    stand  1    -10    110    0     80   0    
23:40:45.862 CD2B.0    stand  1    -10    110    0     80   0    
23:40:46.859 CD2B.0    stand  1    -10    110    0     80   0    
23:40:47.860 CD2B.0    stand  1    -10    110    0     80   0    
23:40:48.861 CD2B.0    stand  1    -10    110    0     80   0    
23:40:49.862 CD2B.0    stand  1    -10    110    0     80   0    
23:40:50.866 CD2B.0    stand  1    -10    110    0     80   0    
23:40:51.862 CD2B.0    stand  1    -10    110    0     80   0    
23:40:52.864 CD2B.0    stand  1    -10    110    0     80   0    
23:40:53.766 CD2B.0    stand  1    -10    110    0     80   0    
23:40:54.756 CD2B.0    stand  1    -10    110    0     80   0    
23:40:55.758 CD2B.0    stand  1    -10    110    0     80   0    
23:40:56.764 CD2B.0    stand  1    -10    110    0     80   0    
23:40:57.764 CD2B.0    stand  1    -10    110    0     80   0    
23:40:58.765 CD2B.0    stand  1    -10    110    0     80   0    
23:40:59.762 CD2B.0    stand  1    -10    110    0     80   0    
23:41:00.766 CD2B.0    stand  1    -10    110    0     80   0    
23:41:01.777 CD2B.0    stand  1    -10    110    0     80   0    
23:41:02.774 CD2B.0    stand  1    -10    110    0     80   0    
23:41:03.768 CD2B.0    stand  1    -10    110    0     80   0    
23:41:04.768 CD2B.0    stand  1    -10    110    0     80   0    
23:41:05.667 CD2B.0    stand  1    -10    110    0     80   0    
23:41:06.668 CD2B.0    stand  1    -10    110    0     80   0    
23:41:07.663 CD2B.0    stand  1    -10    110    0     80   0    
23:41:08.671 CD2B.0    stand  1    -10    110    0     80   0    
23:41:09.665 CD2B.0    stand  1    -10    110    0     80   0    
23:41:10.666 CD2B.0    stand  1    -10    110    0     80   0    
23:41:11.666 CD2B.0    stand  1    -10    110    0     80   0    
23:41:12.682 CD2B.0    stand  1    -10    110    0     80   0    
23:41:13.706 CD2B.0    stand  1    -10    110    0     80   0    
23:41:14.689 CD2B.0    stand  1    -10    110    0     80   0    
23:41:15.581 CD2B.0    stand  1    -10    110    0     80   0    
23:41:16.583 CD2B.0    stand  1    -10    110    0     80   0    
23:41:17.582 CD2B.0    stand  1    -10    110    0     80   0    
23:41:18.583 CD2B.0    stand  1    -10    110    0     80   0    
23:41:19.631 CD2B.0    stand  1    -10    110    0     80   0    
23:41:20.625 CD2B.0    stand  1    -10    110    0     80   0    
23:41:21.587 CD2B.0    stand  1    -10    110    0     80   0    
23:41:22.597 CD2B.0    stand  1    -10    110    0     80   0    
23:41:23.591 CD2B.0    stand  1    -10    110    0     80   0    
23:41:24.594 CD2B.0    stand  1    -10    110    0     80   0    
23:41:25.589 CD2B.0    stand  1    -10    110    0     80   0    
23:41:26.652 CD2B.0    stand  1    -10    110    0     80   0    
23:41:27.487 CD2B.0    stand  1    -10    110    0     80   0    
23:41:28.493 CD2B.0    stand  1    -10    110    0     80   0    
23:41:29.504 CD2B.0    stand  1    -10    110    0     80   0    
23:41:30.496 CD2B.0    stand  1    -10    110    0     80   0    
23:41:31.497 CD2B.0    stand  1    -10    110    0     80   0    
23:41:32.498 CD2B.0    stand  1    -10    110    0     80   0    
23:41:33.499 CD2B.0    stand  1    -10    110    0     80   0    
23:41:34.501 CD2B.0    stand  1    -10    110    0     80   0    
23:41:35.500 CD2B.0    stand  1    -10    110    0     80   0    
23:41:36.502 CD2B.0    stand  1    -10    110    0     80   0    
23:41:37.516 CD2B.0    stand  1    -10    110    0     80   0    
23:41:38.397 CD2B.0    stand  1    -10    110    0     80   0    
23:41:39.398 CD2B.0    stand  1    -10    110    0     80   0    
23:41:40.398 CD2B.0    stand  1    -10    110    0     80   0    
23:41:41.404 CD2B.0    stand  1    -10    110    0     80   0    
23:41:42.402 CD2B.0    stand  1    -10    110    0     80   0    
23:41:43.403 CD2B.0    stand  1    -10    110    0     80   0    
23:41:44.406 CD2B.0    stand  1    -10    110    0     80   0    
23:41:45.406 CD2B.0    stand  1    -10    110    0     80   0    
23:41:46.404 CD2B.0    stand  1    -10    110    0     80   0    
23:41:47.409 CD2B.0    stand  1    -10    110    0     80   0    
23:41:48.408 CD2B.0    stand  1    -10    110    0     80   0    
23:41:49.414 CD2B.0    stand  1    -10    110    0     80   0    
23:41:50.300 CD2B.0    stand  1    -10    110    0     80   0    
23:41:51.301 CD2B.0    stand  1    -10    110    0     80   0    
23:41:52.302 CD2B.0    stand  1    -10    110    0     80   0    
23:41:53.308 CD2B.0    stand  1    -10    110    0     80   0    
23:41:54.305 CD2B.0    stand  1    -10    110    0     80   0    
23:41:55.311 CD2B.0    stand  1    -10    110    0     80   0    
23:41:56.310 CD2B.0    stand  1    -10    110    0     80   0    
23:41:57.317 CD2B.0    stand  1    -10    110    0     80   0    
23:41:58.308 CD2B.0    stand  1    -10    110    0     80   0    
23:41:59.312 CD2B.0    stand  1    -10    110    0     80   0    
23:42:00.348 CD2B.0    stand  1    -10    110    0     80   0    
23:42:01.212 CD2B.0    stand  1    -10    110    0     80   0    
23:42:02.215 CD2B.0    stand  1    -10    110    0     80   0    
23:42:03.217 CD2B.0    stand  1    -10    110    0     80   0    
23:42:04.221 CD2B.0    stand  1    -10    110    0     80   0    
23:42:05.218 CD2B.0    stand  1    -10    110    0     80   0    
23:42:06.221 CD2B.0    stand  1    -10    110    0     80   0    
23:42:07.219 CD2B.0    stand  1    -10    110    0     80   0    
23:42:08.219 CD2B.0    stand  1    -10    110    0     80   0    
23:42:09.220 CD2B.0    stand  1    -10    110    0     80   0    
23:42:10.222 CD2B.0    stand  1    -10    110    0     80   0    
23:42:11.222 CD2B.0    stand  1    -10    110    0     80   0    
23:42:12.227 CD2B.0    stand  1    -10    110    0     80   0    
23:42:13.117 CD2B.0    stand  1    -10    110    0     80   0    
23:42:14.118 CD2B.0    stand  1    -10    110    0     80   0    
23:42:15.134 CD2B.0    stand  1    -10    110    0     80   0    
23:42:16.126 CD2B.0    stand  1    -10    110    0     80   0    
23:42:17.128 CD2B.0    stand  1    -10    110    0     80   0    
23:42:18.130 CD2B.0    stand  1    -10    110    0     80   0    
23:42:19.128 CD2B.0    stand  1    -10    110    0     80   0    
23:42:20.129 CD2B.0    stand  1    -10    110    0     80   0    
23:42:21.176 CD2B.0    stand  1    -10    110    0     80   0    
23:42:22.161 CD2B.0    stand  1    -10    110    0     80   0    
23:42:23.148 CD2B.0    stand  1    -10    110    0     80   0    
23:42:24.034 CD2B.0    stand  1    -10    110    0     80   0    
23:42:25.030 CD2B.0    stand  1    -10    110    0     80   0    
23:42:26.083 CD2B.0    stand  1    -10    110    0     80   0    
23:42:27.033 CD2B.0    stand  1    -10    110    0     80   0    
23:42:28.035 CD2B.0    stand  1    -10    110    0     80   0    
23:42:29.034 CD2B.0    stand  1    -10    110    0     80   0    
23:42:30.040 CD2B.0    stand  1    -10    110    0     80   0    
23:42:31.037 CD2B.0    stand  1    -10    110    0     80   0    
23:42:32.042 CD2B.0    stand  1    -10    110    0     80   0    
23:42:33.038 CD2B.0    stand  1    -10    110    0     80   0    
23:42:34.038 CD2B.0    stand  1    -10    110    0     80   0    
23:42:35.045 CD2B.0    stand  1    -10    110    0     80   0    
23:42:35.939 CD2B.0    stand  1    -10    110    0     80   0    
23:42:36.933 CD2B.0    stand  1    -10    110    0     80   0    
23:42:37.937 CD2B.0    stand  1    -10    110    0     80   0    
23:42:38.937 CD2B.0    stand  1    -10    110    0     80   0    
23:42:39.936 CD2B.0    stand  1    -10    110    0     80   0    
23:42:40.937 CD2B.0    stand  1    -10    110    0     80   0    
23:42:41.939 CD2B.0    stand  1    -10    110    0     80   0    
23:42:42.939 CD2B.0    stand  1    -10    110    0     80   0    
23:42:43.952 CD2B.0    stand  1    -10    110    0     80   0    
23:42:44.941 CD2B.0    stand  1    -10    110    0     80   0    
23:42:45.948 CD2B.0    stand  1    -10    110    0     80   0    
23:42:46.943 CD2B.0    stand  1    -10    110    0     80   0    
23:42:47.850 CD2B.0    stand  1    -10    110    0     80   0    
23:42:48.842 CD2B.0    stand  1    -10    110    0     80   0    
23:42:49.839 CD2B.0    stand  1    -10    110    0     80   0    
23:42:50.841 CD2B.0    stand  1    -10    110    0     80   0    
23:42:51.841 CD2B.0    stand  1    -10    110    0     80   0    
23:42:52.854 CD2B.0    stand  1    -10    110    0     80   0    
23:42:53.843 CD2B.0    stand  1    -10    110    0     80   0    
23:42:54.844 CD2B.0    stand  1    -10    110    0     80   0    
23:42:55.846 CD2B.0    stand  1    -10    110    0     80   0    
23:42:56.876 CD2B.0    stand  1    -10    110    0     80   0    
23:42:57.846 CD2B.0    stand  1    -10    110    0     80   0    
23:42:58.849 CD2B.0    stand  1    -10    110    0     80   0    
23:42:59.739 CD2B.0    stand  1    -10    110    0     80   0    
23:43:00.749 CD2B.0    stand  1    -10    110    0     80   0    
23:43:01.743 CD2B.0    stand  1    -10    110    0     80   0    
23:43:02.747 CD2B.0    stand  1    -10    110    0     80   0    
23:43:03.745 CD2B.0    stand  1    -10    110    0     80   0    
23:43:04.760 CD2B.0    stand  1    -10    110    0     80   0    
23:43:05.764 CD2B.0    stand  1    -10    110    0     80   0    
23:43:06.762 CD2B.0    stand  1    -10    110    0     80   0    
23:43:07.762 CD2B.0    stand  1    -10    110    0     80   0    
23:43:08.762 CD2B.0    stand  1    -10    110    0     80   0    
23:43:09.660 CD2B.0    stand  1    -10    110    0     80   0    
23:43:10.662 CD2B.0    stand  1    -10    110    0     80   0    
23:43:11.662 CD2B.0    stand  1    -10    110    0     80   0    
23:43:12.672 CD2B.0    stand  1    -10    110    0     80   0    
23:43:13.669 CD2B.0    stand  1    -10    110    0     80   0    
23:43:14.673 CD2B.0    stand  1    -10    110    0     80   0    
23:43:15.669 CD2B.0    stand  1    -10    110    0     80   0    
23:43:16.668 CD2B.0    stand  1    -10    110    0     80   0    
23:43:17.670 CD2B.0    stand  1    -10    110    0     80   0    
23:43:18.675 CD2B.0    stand  1    -10    110    0     80   0    
23:43:19.672 CD2B.0    stand  1    -10    110    0     80   0    
23:43:20.572 CD2B.0    stand  1    -10    110    0     80   0    
23:43:21.573 CD2B.0    stand  1    -10    110    0     80   0    
23:43:22.574 CD2B.0    stand  1    -10    110    0     80   0    
23:43:23.580 CD2B.0    stand  1    -10    110    0     80   0    
23:43:24.578 CD2B.0    stand  1    -10    110    0     80   0    
23:43:25.628 CD2B.0    stand  1    -10    110    0     80   0    
23:43:26.578 CD2B.0    stand  1    -10    110    0     80   0    
23:43:27.581 CD2B.0    stand  1    -10    110    0     80   0    
23:43:28.581 CD2B.0    stand  1    -10    110    0     80   0    
23:43:29.581 CD2B.0    stand  1    -10    110    0     80   0    
23:43:30.586 CD2B.0    stand  1    -10    110    0     80   0    
23:43:31.589 CD2B.0    stand  1    -10    110    0     80   0    
23:43:32.478 CD2B.0    stand  1    -10    110    0     80   0    
23:43:33.489 CD2B.0    stand  1    -10    110    0     80   0    
23:43:34.478 CD2B.0    stand  1    -10    110    0     80   0    
23:43:35.488 CD2B.0    stand  1    -10    110    0     80   0    
23:43:36.484 CD2B.0    stand  1    -10    110    0     80   0    
23:43:37.485 CD2B.0    stand  1    -10    110    0     80   0    
23:43:38.487 CD2B.0    stand  1    -10    110    0     80   0    
23:43:39.482 CD2B.0    stand  1    -10    110    0     80   0    
23:43:40.503 CD2B.0    stand  1    -10    110    0     80   0    
23:43:41.484 CD2B.0    stand  1    -10    110    0     80   0    
23:43:42.490 CD2B.0    stand  1    -10    110    0     80   0    
23:43:43.487 CD2B.0    stand  1    -10    110    0     80   0    
23:43:44.386 CD2B.0    stand  1    -10    110    0     80   0    
23:43:45.397 CD2B.0    stand  1    -10    110    0     80   0    
23:43:46.383 CD2B.0    stand  1    -10    110    0     80   0    
23:43:47.383 CD2B.0    stand  1    -10    110    0     80   0    
23:43:48.384 CD2B.0    stand  1    -10    110    0     80   0    
23:43:49.393 CD2B.0    stand  1    -10    110    0     80   0    
23:43:50.386 CD2B.0    stand  1    -10    110    0     80   0    
23:43:51.387 CD2B.0    stand  1    -10    110    0     80   0    
23:43:52.396 CD2B.0    stand  1    -10    110    0     80   0    
23:43:53.397 CD2B.0    stand  1    -10    110    0     80   0    
23:43:54.405 CD2B.0    stand  1    -10    110    0     80   0    
23:43:55.293 CD2B.0    stand  1    -10    110    0     80   0    
23:43:56.294 CD2B.0    stand  1    -10    110    0     80   0    
23:43:57.297 CD2B.0    stand  1    -10    110    0     80   0    
23:43:58.300 CD2B.0    stand  1    -10    110    0     80   0    
23:43:59.300 CD2B.0    stand  1    -10    110    0     80   0    
23:44:00.302 CD2B.0    stand  1    -10    110    0     80   0    
23:44:01.298 CD2B.0    stand  1    -10    110    0     80   0    
23:44:02.299 CD2B.0    stand  1    -10    110    0     80   0    
23:44:03.316 CD2B.0    stand  1    -10    110    0     80   0    
23:44:04.303 CD2B.0    stand  1    -10    110    0     80   0    
23:44:05.308 CD2B.0    stand  1    -10    110    0     80   0    
23:44:06.305 CD2B.0    stand  1    -10    110    0     80   0    
23:44:07.209 CD2B.0    stand  1    -10    110    0     80   0    
23:44:08.206 CD2B.0    stand  1    -10    110    0     80   0    
23:44:09.206 CD2B.0    stand  1    -10    110    0     80   0    
23:44:10.206 CD2B.0    stand  1    -10    110    0     80   0    
23:44:11.217 CD2B.0    stand  1    -10    110    0     80   0    
23:44:12.219 CD2B.0    stand  1    -10    110    0     80   0    
23:44:13.212 CD2B.0    stand  1    -10    110    0     80   0    
23:44:14.211 CD2B.0    stand  1    -10    110    0     80   0    
23:44:15.213 CD2B.0    stand  1    -10    110    0     80   0    
23:44:16.214 CD2B.0    stand  1    -10    110    0     80   0    
23:44:17.216 CD2B.0    stand  1    -10    110    0     80   0    
23:44:18.111 CD2B.0    stand  1    -10    110    0     80   0    
23:44:19.111 CD2B.0    stand  1    -10    110    0     80   0    
23:44:20.116 CD2B.0    stand  1    -10    110    0     80   0    
23:44:21.113 CD2B.0    stand  1    -10    110    0     80   0    
23:44:22.122 CD2B.0    stand  1    -10    110    0     80   0    
23:44:23.119 CD2B.0    stand  1    -10    110    0     80   0    
23:44:24.134 CD2B.0    stand  1    -10    110    0     80   0    
23:44:25.176 CD2B.0    stand  1    -10    110    0     80   0    
23:44:26.128 CD2B.0    stand  1    -10    110    0     80   0    
23:44:27.127 CD2B.0    stand  1    -10    110    0     80   0    
23:44:28.123 CD2B.0    stand  1    -10    110    0     80   0    
23:44:29.129 CD2B.0    stand  1    -10    110    0     80   0    
23:44:30.016 CD2B.0    stand  1    -10    110    0     80   0    
23:44:31.023 CD2B.0    stand  1    -10    110    0     80   0    
23:44:32.029 CD2B.0    stand  1    -10    110    0     80   0    
23:44:33.053 CD2B.0    stand  1    -10    110    0     80   0    
23:44:34.021 CD2B.0    stand  1    -10    110    0     80   0    
23:44:35.020 CD2B.0    stand  1    -10    110    0     80   0    
23:44:36.030 CD2B.0    stand  1    -10    110    0     80   0    
23:44:37.046 CD2B.0    stand  1    -10    110    0     80   0    
23:44:38.025 CD2B.0    stand  1    -10    110    0     80   0    
23:44:39.021 CD2B.0    stand  1    -10    110    0     80   0    
23:44:40.028 CD2B.0    stand  1    -10    110    0     80   0    
23:44:40.926 CD2B.0    stand  1    -10    110    0     80   0    
23:44:41.926 CD2B.0    stand  1    -10    110    0     80   0    
23:44:42.925 CD2B.0    stand  1    -10    110    0     80   0    
23:44:43.929 CD2B.0    stand  1    -10    110    0     80   0    
23:44:44.927 CD2B.0    stand  1    -10    110    0     80   0    
23:44:45.929 CD2B.0    stand  1    -10    110    0     80   0    
23:44:46.933 CD2B.0    stand  1    -10    110    0     80   0    
23:44:47.937 CD2B.0    stand  1    -10    110    0     80   0    
23:44:48.933 CD2B.0    stand  1    -10    110    0     80   0    
23:44:49.933 CD2B.0    stand  1    -10    110    0     80   0    
23:44:50.934 CD2B.0    stand  1    -10    110    0     80   0    
23:44:51.935 CD2B.0    stand  1    -10    110    0     80   0    
23:44:52.829 CD2B.0    stand  1    -10    110    0     80   0    
23:44:53.836 CD2B.0    stand  1    -10    110    0     80   0    
23:44:54.829 CD2B.0    stand  1    -10    110    0     80   0    
23:44:55.846 CD2B.0    stand  1    -10    110    0     80   0    
23:44:56.843 CD2B.0    stand  1    -10    110    0     80   0    
23:44:57.844 CD2B.0    stand  1    -10    110    0     80   0    
23:44:58.841 CD2B.0    stand  1    -10    110    0     80   0    
23:44:59.846 CD2B.0    stand  1    -10    110    0     80   0    
23:45:00.844 CD2B.0    stand  1    -10    110    0     80   0    
23:45:01.852 CD2B.0    stand  1    -10    110    0     80   0    
23:45:02.845 CD2B.0    stand  1    -10    110    0     80   0    
23:45:03.747 CD2B.0    stand  1    -10    110    0     80   0    
23:45:04.750 CD2B.0    stand  1    -10    110    0     80   0    
23:45:05.741 CD2B.0    stand  1    -10    110    0     80   0    
23:45:06.742 CD2B.0    stand  1    -10    110    0     80   0    
23:45:07.744 CD2B.0    stand  1    0      120    0     80   14   
23:45:08.746 CD2B.0    stand  1    0      120    0     80   0    
23:45:09.755 CD2B.0    stand  1    0      120    0     80   0    
23:45:10.748 CD2B.0    stand  1    0      120    0     80   0    
23:45:11.754 CD2B.0    stand  1    10     80     10    80   41   
23:45:12.757 CD2B.0    stand  1    10     80     0     80   0    
23:45:13.764 CD2B.0    stand  1    10     80     0     80   0    
23:45:14.752 CD2B.0    stand  1    10     80     0     80   0    
23:45:15.652 CD2B.0    stand  1    10     70     0     80   10   
23:45:16.649 CD2B.0    stand  1    10     70     0     80   0    
23:45:17.668 CD2B.0    stand  1    10     70     0     80   0    
23:45:18.648 CD2B.0    stand  255  10     70     0     80   0    
23:45:19.652 CD2B.0    stand  255  10     70     0     80   0    
23:45:20.668 CD2B.0    stand  255  10     70     0     80   0    
23:45:21.658 CD2B.0    stand  255  10     70     0     80   0    
23:45:22.673 CD2B.0    stand  255  10     70     0     80   0    
23:45:23.659 CD2B.0    stand  255  10     70     0     80   0    
23:45:24.705 CD2B.0    stand  255  10     70     0     80   0    
23:45:25.657 CD2B.0    stand  255  10     70     0     80   0    
23:45:26.663 CD2B.0    stand  255  10     70     0     80   0    
23:45:27.551 CD2B.0    stand  255  0      90     14    80   22   
23:45:28.553 CD2B.0    stand  255  0      100    0     80   10   
23:45:29.553 CD2B.0    stand  255  0      100    0     80   0    
23:45:30.569 CD2B.0    stand  255  0      100    0     80   0    
23:45:31.555 CD2B.0    stand  255  0      100    0     80   0    
23:45:32.558 CD2B.0    stand  255  0      100    0     80   0    
23:45:33.606 CD2B.0    stand  255  0      100    0     80   0    
23:45:34.561 CD2B.0    stand  255  0      100    0     80   0    
23:45:35.561 CD2B.0    stand  255  0      100    0     80   0    
23:45:36.568 CD2B.0    stand  255  0      100    0     80   0    
23:45:37.561 CD2B.0    stand  255  0      100    0     80   0    
23:45:38.473 CD2B.0    stand  255  0      100    0     80   0    
23:45:39.467 CD2B.0    stand  255  0      100    0     80   0    
23:45:40.462 CD2B.0    stand  255  0      100    0     80   0    
23:45:41.465 CD2B.0    stand  255  0      100    0     80   0    
23:45:42.464 CD2B.0    stand  255  0      100    0     80   0    
23:45:43.470 CD2B.0    stand  255  0      100    0     80   0    
23:45:44.484 CD2B.0    stand  255  0      100    0     80   0    
23:45:45.474 CD2B.0    stand  255  0      100    0     80   0    
23:45:46.474 CD2B.0    stand  255  0      100    0     80   0    
23:45:47.477 CD2B.0    stand  255  0      100    0     80   0    
23:45:48.477 CD2B.0    stand  255  0      100    0     80   0    
23:45:49.375 CD2B.0    stand  255  0      100    0     80   0    
23:45:50.377 CD2B.0    stand  255  0      100    0     80   0    
23:45:51.375 CD2B.0    stand  255  0      100    0     80   0    
23:45:52.375 CD2B.0    stand  255  0      100    0     80   0    
23:45:53.377 CD2B.0    stand  255  0      100    0     80   0    
23:45:54.377 CD2B.0    stand  255  0      100    0     80   0    
23:45:55.379 CD2B.0    stand  255  0      100    0     80   0    
23:45:56.380 CD2B.0    stand  255  0      100    0     80   0    
23:45:57.387 CD2B.0    stand  255  0      100    0     80   0    
23:45:58.385 CD2B.0    stand  255  0      100    0     80   0    

```

**汇总**: xray tick 215 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

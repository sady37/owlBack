# case-cd2b-0616-06400701 — 每 tick belief 时间线 (room fd00:0:3:112:3:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
06:40:00 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=16 mv=0 turn=0 room -    Empty      1   0     0.00  0.05  0.14  0.00  0.77  0.04
06:40:00 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  0.50 Empty      1   0     0.00  0.05  0.14  0.00  0.77  0.04
06:40:01 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=16 mv=0 turn=0 room -    Empty      1   0     0.00  0.05  0.14  0.00  0.77  0.04
06:40:01 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  0.51 Empty      1   1     0.00  0.35  0.17  0.00  0.44  0.01
06:40:02 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Empty      1   1     0.00  0.35  0.17  0.00  0.44  0.01
06:40:02 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  0.52 Bed        1   2     0.00  0.77  0.08  0.01  0.11  0.00
06:40:03 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=16 mv=0 turn=0 room -    Bed        1   2     0.00  0.77  0.08  0.01  0.11  0.00
06:40:03 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  0.53 Bed        1   3     0.00  0.93  0.04  0.01  0.02  0.00
06:40:04 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=17 mv=0 turn=0 room -    Bed        1   3     0.00  0.93  0.04  0.01  0.02  0.00
06:40:04 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  0.54 Bed        1   4     0.00  0.96  0.02  0.01  0.00  0.00
06:40:05 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   4     0.00  0.96  0.02  0.01  0.00  0.00
06:40:05 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   5     0.00  0.96  0.02  0.01  0.00  0.00
06:40:06 0865.0   -             pad     -    InBed    pad InBed HR=83 RR=17 mv=0 turn=0 room -    Bed        1   5     0.00  0.96  0.02  0.01  0.00  0.00
06:40:06 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   6     0.00  0.96  0.02  0.01  0.00  0.00
06:40:07 1641.0   -             pad     -    InBed    pad InBed HR=59 RR=15 mv=0 turn=0 room -    Bed        1   6     0.00  0.96  0.02  0.01  0.00  0.00
06:40:07 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   6     0.00  0.96  0.02  0.01  0.00  0.00
06:40:08 0865.0   -             pad     -    InBed    pad InBed HR=84 RR=17 mv=0 turn=0 room -    Bed        1   6     0.00  0.96  0.02  0.01  0.00  0.00
06:40:08 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   7     0.00  0.96  0.02  0.01  0.00  0.00
06:40:09 1641.0   -             pad     -    InBed    pad InBed HR=57 RR=15 mv=0 turn=0 room -    Bed        1   7     0.00  0.96  0.02  0.01  0.00  0.00
06:40:09 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   8     0.00  0.96  0.02  0.01  0.00  0.00
06:40:10 0865.0   -             pad     -    InBed    pad InBed HR=83 RR=17 mv=0 turn=0 room -    Bed        1   8     0.00  0.96  0.02  0.01  0.00  0.00
06:40:10 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   9     0.00  0.96  0.02  0.01  0.00  0.00
06:40:10 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   9     0.00  0.96  0.02  0.01  0.00  0.00
06:40:11 1641.0   -             pad     -    InBed    pad InBed HR=57 RR=15 mv=0 turn=0 room -    Bed        1   9     0.00  0.96  0.02  0.01  0.00  0.00
06:40:11 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   10    0.00  0.96  0.02  0.01  0.00  0.00
06:40:12 0865.0   -             pad     -    InBed    pad InBed HR=82 RR=17 mv=0 turn=0 room -    Bed        1   10    0.00  0.96  0.02  0.01  0.00  0.00
06:40:12 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   11    0.00  0.96  0.02  0.01  0.00  0.00
06:40:13 1641.0   -             pad     -    InBed    pad InBed HR=57 RR=15 mv=0 turn=0 room -    Bed        1   11    0.00  0.96  0.02  0.01  0.00  0.00
06:40:13 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   12    0.00  0.96  0.02  0.01  0.00  0.00
06:40:14 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=17 mv=0 turn=0 room -    Bed        1   12    0.00  0.96  0.02  0.01  0.00  0.00
06:40:14 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   13    0.00  0.96  0.02  0.01  0.00  0.00
06:40:15 1641.0   -             pad     -    InBed    pad InBed HR=57 RR=15 mv=0 turn=0 room -    Bed        1   13    0.00  0.96  0.02  0.01  0.00  0.00
06:40:15 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   14    0.00  0.96  0.02  0.01  0.00  0.00
06:40:16 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=17 mv=0 turn=0 room -    Bed        1   14    0.00  0.96  0.02  0.01  0.00  0.00
06:40:16 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   15    0.00  0.96  0.02  0.01  0.00  0.00
06:40:17 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=15 mv=0 turn=0 room -    Bed        1   15    0.00  0.96  0.02  0.01  0.00  0.00
06:40:17 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   16    0.00  0.96  0.02  0.01  0.00  0.00
06:40:18 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=17 mv=0 turn=0 room -    Bed        1   16    0.00  0.96  0.02  0.01  0.00  0.00
06:40:18 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   17    0.00  0.96  0.02  0.01  0.00  0.00
06:40:19 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   18    0.00  0.96  0.02  0.01  0.00  0.00
06:40:20 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        1   18    0.00  0.96  0.02  0.01  0.00  0.00
06:40:20 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=15 mv=0 turn=0 room -    Bed        1   18    0.00  0.96  0.02  0.01  0.00  0.00
06:40:20 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   19    0.00  0.96  0.02  0.01  0.00  0.00
06:40:21 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   20    0.00  0.96  0.02  0.01  0.00  0.00
06:40:22 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        1   20    0.00  0.96  0.02  0.01  0.00  0.00
06:40:22 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   20    0.00  0.96  0.02  0.01  0.00  0.00
06:40:22 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   21    0.00  0.96  0.02  0.01  0.00  0.00
06:40:23 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   22    0.00  0.96  0.02  0.01  0.00  0.00
06:40:24 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        1   22    0.00  0.96  0.02  0.01  0.00  0.00
06:40:24 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   22    0.00  0.96  0.02  0.01  0.00  0.00
06:40:24 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   23    0.00  0.96  0.02  0.01  0.00  0.00
06:40:25 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   24    0.00  0.96  0.02  0.01  0.00  0.00
06:40:26 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        1   24    0.00  0.96  0.02  0.01  0.00  0.00
06:40:26 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   24    0.00  0.96  0.02  0.01  0.00  0.00
06:40:26 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   25    0.00  0.96  0.02  0.01  0.00  0.00
06:40:27 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   26    0.00  0.96  0.02  0.01  0.00  0.00
06:40:28 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=17 mv=0 turn=0 room -    Bed        1   26    0.00  0.96  0.02  0.01  0.00  0.00
06:40:28 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=None mv=0 turn=0 room -    Bed        1   26    0.00  0.96  0.02  0.01  0.00  0.00
06:40:28 CD2B.0   CD2B04000458  lying   62   InBed    lying              trk  1.00 Bed        1   27    0.00  0.99  0.00  0.00  0.00  0.00
06:40:29 CD2B.0   CD2B04000458  lying   65   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:30 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:30 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:30 CD2B.0   CD2B04000458  lying   70   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:31 CD2B.0   CD2B04000458  lying   70   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:32 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:32 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:32 CD2B.0   CD2B04000458  lying   74   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:33 CD2B.0   CD2B04000458  lying   75   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:34 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:34 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:34 CD2B.0   CD2B04000458  lying   73   InBed    lying              trk  1.00 Bed        1   0     0.00  0.96  0.01  0.02  0.00  0.00
06:40:35 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:36 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:36 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:36 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:38 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:38 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:40 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:40 1641.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=0 turn=1 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:40:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:42 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:42 1641.0   -             pad     -    InBed    pad InBed HR=None RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:42 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:44 0865.0   -             pad     -    InBed    pad InBed HR=82 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:44 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:45 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:46 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:46 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:46 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:47 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:48 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:48 1641.0   -             pad     -    InBed    pad InBed HR=70 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:50 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:50 1641.0   -             pad     -    InBed    pad InBed HR=70 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:40:51 CD2B.0   CD2B04000458  stand   74   InBed    stand              trk  1.00 Bed        1   0     0.00  0.97  0.02  0.01  0.00  0.00
06:40:52 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.97  0.02  0.01  0.00  0.00
06:40:52 1641.0   -             pad     -    InBed    pad InBed HR=71 RR=13 mv=1 turn=0 room -    Bed        1   0     0.00  0.97  0.02  0.01  0.00  0.00
06:40:52 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:52 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:53 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:54 1641.0   -             pad     -    InBed    pad InBed HR=71 RR=13 mv=1 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:54 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:54 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:55 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:56 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:56 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:57 1641.0   -             pad     -    InBed    pad InBed HR=71 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:57 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:58 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:59 1641.0   -             pad     -    InBed    pad InBed HR=71 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:40:59 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:00 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:00 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:01 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:01 1641.0   -             pad     -    InBed    pad InBed HR=71 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:02 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:02 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:03 1641.0   -             pad     -    InBed    pad InBed HR=71 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:03 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:04 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:04 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:04 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:05 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:05 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:06 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:06 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:07 1641.0   -             pad     -    InBed    pad InBed HR=59 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:07 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:08 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:08 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:09 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:09 CD2B.0   CD2B04000458  walk    74   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:10 0865.0   -             pad     -    InBed    pad InBed HR=72 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:41:10 CD2B.0   CD2B04000458  lying   76   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:11 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:11 CD2B.0   CD2B04000458  lying   61   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:12 0865.0   -             pad     -    InBed    pad InBed HR=72 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:12 CD2B.E   -             -       0    InBed    np=1               room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:13 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:13 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:14 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:14 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:15 1641.0   -             pad     -    InBed    pad InBed HR=None RR=12 mv=0 turn=1 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:16 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:17 1641.0   -             pad     -    InBed    pad InBed HR=None RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:18 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:19 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:20 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:21 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:22 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:23 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:24 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:25 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:26 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:26 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:27 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:28 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:28 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:29 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:29 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:30 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:30 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:31 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:31 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:32 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:32 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:33 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:33 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:34 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:34 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:35 1641.0   -             pad     -    InBed    pad InBed HR=57 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:35 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:41:36 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   26    0.00  0.99  0.00  0.00  0.00  0.00
06:41:37 1641.0   -             pad     -    InBed    pad InBed HR=57 RR=11 mv=0 turn=0 room -    Bed        1   26    0.00  0.99  0.00  0.00  0.00  0.00
06:41:37 0865.0   -             pad     -    InBed    pad InBed HR=83 RR=15 mv=0 turn=0 room -    Bed        1   26    0.00  0.99  0.00  0.00  0.00  0.00
06:41:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   27    0.00  0.99  0.00  0.00  0.00  0.00
06:41:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   28    0.00  0.99  0.00  0.00  0.00  0.00
06:41:39 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=11 mv=0 turn=0 room -    Bed        1   28    0.00  0.99  0.00  0.00  0.00  0.00
06:41:39 0865.0   -             pad     -    InBed    pad InBed HR=84 RR=15 mv=0 turn=0 room -    Bed        1   28    0.00  0.99  0.00  0.00  0.00  0.00
06:41:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:41:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   30    0.00  0.99  0.00  0.00  0.00  0.00
06:41:41 1641.0   -             pad     -    InBed    pad InBed HR=59 RR=11 mv=0 turn=0 room -    Bed        1   30    0.00  0.99  0.00  0.00  0.00  0.00
06:41:41 0865.0   -             pad     -    InBed    pad InBed HR=83 RR=16 mv=0 turn=0 room -    Bed        1   30    0.00  0.99  0.00  0.00  0.00  0.00
06:41:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   31    0.00  0.99  0.00  0.00  0.00  0.00
06:41:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   32    0.00  0.99  0.00  0.00  0.00  0.00
06:41:43 1641.0   -             pad     -    InBed    pad InBed HR=59 RR=11 mv=0 turn=0 room -    Bed        1   32    0.00  0.99  0.00  0.00  0.00  0.00
06:41:43 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Bed        1   32    0.00  0.99  0.00  0.00  0.00  0.00
06:41:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   33    0.00  0.99  0.00  0.00  0.00  0.00
06:41:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   34    0.00  0.99  0.00  0.00  0.00  0.00
06:41:45 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   34    0.00  0.99  0.00  0.00  0.00  0.00
06:41:45 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=15 mv=0 turn=0 room -    Bed        1   34    0.00  0.99  0.00  0.00  0.00  0.00
06:41:45 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   35    0.00  0.99  0.00  0.00  0.00  0.00
06:41:46 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   35    0.00  0.99  0.00  0.00  0.00  0.00
06:41:46 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   36    0.00  0.99  0.00  0.00  0.00  0.00
06:41:47 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   36    0.00  0.99  0.00  0.00  0.00  0.00
06:41:47 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=15 mv=0 turn=0 room -    Bed        1   36    0.00  0.99  0.00  0.00  0.00  0.00
06:41:47 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   37    0.00  0.99  0.00  0.00  0.00  0.00
06:41:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
06:41:49 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
06:41:49 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=16 mv=0 turn=0 room -    Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
06:41:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   39    0.00  0.99  0.00  0.00  0.00  0.00
06:41:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   40    0.00  0.99  0.00  0.00  0.00  0.00
06:41:51 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   40    0.00  0.99  0.00  0.00  0.00  0.00
06:41:51 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=16 mv=0 turn=0 room -    Bed        1   40    0.00  0.99  0.00  0.00  0.00  0.00
06:41:51 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   41    0.00  0.99  0.00  0.00  0.00  0.00
06:41:52 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   42    0.00  0.99  0.00  0.00  0.00  0.00
06:41:53 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   42    0.00  0.99  0.00  0.00  0.00  0.00
06:41:53 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=16 mv=0 turn=0 room -    Bed        1   42    0.00  0.99  0.00  0.00  0.00  0.00
06:41:53 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   43    0.00  0.99  0.00  0.00  0.00  0.00
06:41:54 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   44    0.00  0.99  0.00  0.00  0.00  0.00
06:41:55 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   44    0.00  0.99  0.00  0.00  0.00  0.00
06:41:55 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=17 mv=0 turn=0 room -    Bed        1   44    0.00  0.99  0.00  0.00  0.00  0.00
06:41:55 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   45    0.00  0.99  0.00  0.00  0.00  0.00
06:41:56 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   46    0.00  0.99  0.00  0.00  0.00  0.00
06:41:57 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   46    0.00  0.99  0.00  0.00  0.00  0.00
06:41:57 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=18 mv=0 turn=0 room -    Bed        1   46    0.00  0.99  0.00  0.00  0.00  0.00
06:41:57 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   47    0.00  0.99  0.00  0.00  0.00  0.00
06:41:58 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   48    0.00  0.99  0.00  0.00  0.00  0.00
06:41:59 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   48    0.00  0.99  0.00  0.00  0.00  0.00
06:41:59 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=18 mv=0 turn=0 room -    Bed        1   48    0.00  0.99  0.00  0.00  0.00  0.00
06:41:59 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   49    0.00  0.99  0.00  0.00  0.00  0.00
06:42:00 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   50    0.00  0.99  0.00  0.00  0.00  0.00
06:42:01 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   50    0.00  0.99  0.00  0.00  0.00  0.00
06:42:01 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=18 mv=0 turn=0 room -    Bed        1   50    0.00  0.99  0.00  0.00  0.00  0.00
06:42:01 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   51    0.00  0.99  0.00  0.00  0.00  0.00
06:42:02 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   52    0.00  0.99  0.00  0.00  0.00  0.00
06:42:03 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   52    0.00  0.99  0.00  0.00  0.00  0.00
06:42:03 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=18 mv=0 turn=0 room -    Bed        1   52    0.00  0.99  0.00  0.00  0.00  0.00
06:42:03 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   53    0.00  0.99  0.00  0.00  0.00  0.00
06:42:04 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   54    0.00  0.99  0.00  0.00  0.00  0.00
06:42:05 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   54    0.00  0.99  0.00  0.00  0.00  0.00
06:42:05 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=18 mv=0 turn=0 room -    Bed        1   54    0.00  0.99  0.00  0.00  0.00  0.00
06:42:05 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   55    0.00  0.99  0.00  0.00  0.00  0.00
06:42:06 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   56    0.00  0.99  0.00  0.00  0.00  0.00
06:42:07 1641.0   -             pad     -    InBed    pad InBed HR=59 RR=11 mv=0 turn=0 room -    Bed        1   56    0.00  0.99  0.00  0.00  0.00  0.00
06:42:07 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=18 mv=0 turn=0 room -    Bed        1   56    0.00  0.99  0.00  0.00  0.00  0.00
06:42:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   57    0.00  0.99  0.00  0.00  0.00  0.00
06:42:08 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   58    0.00  0.99  0.00  0.00  0.00  0.00
06:42:09 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=17 mv=0 turn=0 room -    Bed        1   58    0.00  0.99  0.00  0.00  0.00  0.00
06:42:09 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=11 mv=0 turn=0 room -    Bed        1   58    0.00  0.99  0.00  0.00  0.00  0.00
06:42:09 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   59    0.00  0.99  0.00  0.00  0.00  0.00
06:42:10 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   60    0.00  0.99  0.00  0.00  0.00  0.00
06:42:11 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   60    0.00  0.99  0.00  0.00  0.00  0.00
06:42:11 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=17 mv=0 turn=0 room -    Bed        1   60    0.00  0.99  0.00  0.00  0.00  0.00
06:42:11 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   61    0.00  0.99  0.00  0.00  0.00  0.00
06:42:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   62    0.00  0.99  0.00  0.00  0.00  0.00
06:42:13 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   62    0.00  0.99  0.00  0.00  0.00  0.00
06:42:13 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=17 mv=0 turn=0 room -    Bed        1   62    0.00  0.99  0.00  0.00  0.00  0.00
06:42:13 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   63    0.00  0.99  0.00  0.00  0.00  0.00
06:42:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   64    0.00  0.99  0.00  0.00  0.00  0.00
06:42:15 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=17 mv=0 turn=0 room -    Bed        1   64    0.00  0.99  0.00  0.00  0.00  0.00
06:42:15 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   64    0.00  0.99  0.00  0.00  0.00  0.00
06:42:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   65    0.00  0.99  0.00  0.00  0.00  0.00
06:42:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   66    0.00  0.99  0.00  0.00  0.00  0.00
06:42:17 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=17 mv=0 turn=0 room -    Bed        1   66    0.00  0.99  0.00  0.00  0.00  0.00
06:42:17 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=12 mv=0 turn=0 room -    Bed        1   66    0.00  0.99  0.00  0.00  0.00  0.00
06:42:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   67    0.00  0.99  0.00  0.00  0.00  0.00
06:42:18 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   67    0.00  0.99  0.00  0.00  0.00  0.00
06:42:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   68    0.00  0.99  0.00  0.00  0.00  0.00
06:42:19 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=17 mv=0 turn=0 room -    Bed        1   68    0.00  0.99  0.00  0.00  0.00  0.00
06:42:19 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=12 mv=0 turn=0 room -    Bed        1   68    0.00  0.99  0.00  0.00  0.00  0.00
06:42:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   69    0.00  0.99  0.00  0.00  0.00  0.00
06:42:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   70    0.00  0.99  0.00  0.00  0.00  0.00
06:42:21 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=17 mv=0 turn=0 room -    Bed        1   70    0.00  0.99  0.00  0.00  0.00  0.00
06:42:21 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   70    0.00  0.99  0.00  0.00  0.00  0.00
06:42:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   71    0.00  0.99  0.00  0.00  0.00  0.00
06:42:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   72    0.00  0.99  0.00  0.00  0.00  0.00
06:42:23 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   72    0.00  0.99  0.00  0.00  0.00  0.00
06:42:23 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   72    0.00  0.99  0.00  0.00  0.00  0.00
06:42:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   73    0.00  0.99  0.00  0.00  0.00  0.00
06:42:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   74    0.00  0.99  0.00  0.00  0.00  0.00
06:42:25 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=16 mv=0 turn=0 room -    Bed        1   74    0.00  0.99  0.00  0.00  0.00  0.00
06:42:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   75    0.00  0.99  0.00  0.00  0.00  0.00
06:42:26 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   75    0.00  0.99  0.00  0.00  0.00  0.00
06:42:26 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   76    0.00  0.99  0.00  0.00  0.00  0.00
06:42:27 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=16 mv=0 turn=0 room -    Bed        1   76    0.00  0.99  0.00  0.00  0.00  0.00
06:42:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   77    0.00  0.99  0.00  0.00  0.00  0.00
06:42:28 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   77    0.00  0.99  0.00  0.00  0.00  0.00
06:42:28 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   78    0.00  0.99  0.00  0.00  0.00  0.00
06:42:29 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=16 mv=0 turn=0 room -    Bed        1   78    0.00  0.99  0.00  0.00  0.00  0.00
06:42:29 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   79    0.00  0.99  0.00  0.00  0.00  0.00
06:42:30 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   79    0.00  0.99  0.00  0.00  0.00  0.00
06:42:30 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   80    0.00  0.99  0.00  0.00  0.00  0.00
06:42:31 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=16 mv=0 turn=0 room -    Bed        1   80    0.00  0.99  0.00  0.00  0.00  0.00
06:42:31 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   81    0.00  0.99  0.00  0.00  0.00  0.00
06:42:32 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   81    0.00  0.99  0.00  0.00  0.00  0.00
06:42:32 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   82    0.00  0.99  0.00  0.00  0.00  0.00
06:42:33 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=16 mv=0 turn=0 room -    Bed        1   82    0.00  0.99  0.00  0.00  0.00  0.00
06:42:33 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   83    0.00  0.99  0.00  0.00  0.00  0.00
06:42:34 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   83    0.00  0.99  0.00  0.00  0.00  0.00
06:42:34 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   84    0.00  0.99  0.00  0.00  0.00  0.00
06:42:35 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   84    0.00  0.99  0.00  0.00  0.00  0.00
06:42:35 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   85    0.00  0.99  0.00  0.00  0.00  0.00
06:42:36 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   85    0.00  0.99  0.00  0.00  0.00  0.00
06:42:36 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   86    0.00  0.99  0.00  0.00  0.00  0.00
06:42:37 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   86    0.00  0.99  0.00  0.00  0.00  0.00
06:42:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   87    0.00  0.99  0.00  0.00  0.00  0.00
06:42:38 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   87    0.00  0.99  0.00  0.00  0.00  0.00
06:42:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   88    0.00  0.99  0.00  0.00  0.00  0.00
06:42:39 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   88    0.00  0.99  0.00  0.00  0.00  0.00
06:42:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   89    0.00  0.99  0.00  0.00  0.00  0.00
06:42:40 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=12 mv=0 turn=0 room -    Bed        1   89    0.00  0.99  0.00  0.00  0.00  0.00
06:42:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   90    0.00  0.99  0.00  0.00  0.00  0.00
06:42:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   91    0.00  0.99  0.00  0.00  0.00  0.00
06:42:42 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   91    0.00  0.99  0.00  0.00  0.00  0.00
06:42:42 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=12 mv=0 turn=0 room -    Bed        1   91    0.00  0.99  0.00  0.00  0.00  0.00
06:42:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   92    0.00  0.99  0.00  0.00  0.00  0.00
06:42:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   93    0.00  0.99  0.00  0.00  0.00  0.00
06:42:44 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=12 mv=0 turn=0 room -    Bed        1   93    0.00  0.99  0.00  0.00  0.00  0.00
06:42:44 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=17 mv=0 turn=0 room -    Bed        1   93    0.00  0.99  0.00  0.00  0.00  0.00
06:42:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   94    0.00  0.99  0.00  0.00  0.00  0.00
06:42:45 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   95    0.00  0.99  0.00  0.00  0.00  0.00
06:42:46 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=12 mv=0 turn=0 room -    Bed        1   95    0.00  0.99  0.00  0.00  0.00  0.00
06:42:46 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=17 mv=0 turn=0 room -    Bed        1   95    0.00  0.99  0.00  0.00  0.00  0.00
06:42:46 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   96    0.00  0.99  0.00  0.00  0.00  0.00
06:42:47 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   97    0.00  0.99  0.00  0.00  0.00  0.00
06:42:48 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=18 mv=0 turn=0 room -    Bed        1   97    0.00  0.99  0.00  0.00  0.00  0.00
06:42:48 1641.0   -             pad     -    InBed    pad InBed HR=67 RR=12 mv=0 turn=0 room -    Bed        1   97    0.00  0.99  0.00  0.00  0.00  0.00
06:42:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   98    0.00  0.99  0.00  0.00  0.00  0.00
06:42:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   99    0.00  0.99  0.00  0.00  0.00  0.00
06:42:49 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   99    0.00  0.99  0.00  0.00  0.00  0.00
06:42:50 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=18 mv=0 turn=0 room -    Bed        1   99    0.00  0.99  0.00  0.00  0.00  0.00
06:42:50 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=12 mv=0 turn=0 room -    Bed        1   99    0.00  0.99  0.00  0.00  0.00  0.00
06:42:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   100   0.00  0.99  0.00  0.00  0.00  0.00
06:42:51 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   101   0.00  0.99  0.00  0.00  0.00  0.00
06:42:52 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=12 mv=0 turn=0 room -    Bed        1   101   0.00  0.99  0.00  0.00  0.00  0.00
06:42:52 0865.0   -             pad     -    InBed    pad InBed HR=82 RR=18 mv=0 turn=0 room -    Bed        1   101   0.00  0.99  0.00  0.00  0.00  0.00
06:42:52 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   102   0.00  0.99  0.00  0.00  0.00  0.00
06:42:53 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   103   0.00  0.99  0.00  0.00  0.00  0.00
06:42:54 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=18 mv=0 turn=0 room -    Bed        1   103   0.00  0.99  0.00  0.00  0.00  0.00
06:42:54 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   103   0.00  0.99  0.00  0.00  0.00  0.00
06:42:54 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   104   0.00  0.99  0.00  0.00  0.00  0.00
06:42:55 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   105   0.00  0.99  0.00  0.00  0.00  0.00
06:42:56 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=19 mv=0 turn=0 room -    Bed        1   105   0.00  0.99  0.00  0.00  0.00  0.00
06:42:56 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=12 mv=0 turn=0 room -    Bed        1   105   0.00  0.99  0.00  0.00  0.00  0.00
06:42:56 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   106   0.00  0.99  0.00  0.00  0.00  0.00
06:42:57 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   107   0.00  0.99  0.00  0.00  0.00  0.00
06:42:58 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=19 mv=0 turn=0 room -    Bed        1   107   0.00  0.99  0.00  0.00  0.00  0.00
06:42:58 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=13 mv=0 turn=0 room -    Bed        1   107   0.00  0.99  0.00  0.00  0.00  0.00
06:42:58 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   108   0.00  0.99  0.00  0.00  0.00  0.00
06:42:59 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   109   0.00  0.99  0.00  0.00  0.00  0.00
06:42:59 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   110   0.00  0.99  0.00  0.00  0.00  0.00
06:43:00 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=19 mv=0 turn=0 room -    Bed        1   110   0.00  0.99  0.00  0.00  0.00  0.00
06:43:00 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   110   0.00  0.99  0.00  0.00  0.00  0.00
06:43:00 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   111   0.00  0.99  0.00  0.00  0.00  0.00
06:43:01 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   112   0.00  0.99  0.00  0.00  0.00  0.00
06:43:02 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=19 mv=0 turn=0 room -    Bed        1   112   0.00  0.99  0.00  0.00  0.00  0.00
06:43:02 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   112   0.00  0.99  0.00  0.00  0.00  0.00
06:43:02 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   113   0.00  0.99  0.00  0.00  0.00  0.00
06:43:03 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   114   0.00  0.99  0.00  0.00  0.00  0.00
06:43:04 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   114   0.00  0.99  0.00  0.00  0.00  0.00
06:43:04 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=19 mv=0 turn=0 room -    Bed        1   114   0.00  0.99  0.00  0.00  0.00  0.00
06:43:04 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   115   0.00  0.99  0.00  0.00  0.00  0.00
06:43:05 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   116   0.00  0.99  0.00  0.00  0.00  0.00
06:43:06 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   116   0.00  0.99  0.00  0.00  0.00  0.00
06:43:06 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=19 mv=0 turn=0 room -    Bed        1   116   0.00  0.99  0.00  0.00  0.00  0.00
06:43:06 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   117   0.00  0.99  0.00  0.00  0.00  0.00
06:43:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   118   0.00  0.99  0.00  0.00  0.00  0.00
06:43:08 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   118   0.00  0.99  0.00  0.00  0.00  0.00
06:43:08 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=19 mv=0 turn=0 room -    Bed        1   118   0.00  0.99  0.00  0.00  0.00  0.00
06:43:08 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   119   0.00  0.99  0.00  0.00  0.00  0.00
06:43:09 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   119   0.00  0.99  0.00  0.00  0.00  0.00
06:43:10 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   119   0.00  0.99  0.00  0.00  0.00  0.00
06:43:10 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=19 mv=0 turn=0 room -    Bed        1   119   0.00  0.99  0.00  0.00  0.00  0.00
06:43:10 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   120   0.00  0.99  0.00  0.00  0.00  0.00
06:43:11 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   122   0.00  0.99  0.00  0.00  0.00  0.00
06:43:12 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   122   0.00  0.99  0.00  0.00  0.00  0.00
06:43:12 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=19 mv=0 turn=0 room -    Bed        1   122   0.00  0.99  0.00  0.00  0.00  0.00
06:43:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   122   0.00  0.99  0.00  0.00  0.00  0.00
06:43:13 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   123   0.00  0.99  0.00  0.00  0.00  0.00
06:43:14 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   123   0.00  0.99  0.00  0.00  0.00  0.00
06:43:14 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=19 mv=0 turn=0 room -    Bed        1   123   0.00  0.99  0.00  0.00  0.00  0.00
06:43:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   124   0.00  0.99  0.00  0.00  0.00  0.00
06:43:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   126   0.00  0.99  0.00  0.00  0.00  0.00
06:43:16 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=18 mv=0 turn=0 room -    Bed        1   126   0.00  0.99  0.00  0.00  0.00  0.00
06:43:16 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   126   0.00  0.99  0.00  0.00  0.00  0.00
06:43:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   126   0.00  0.99  0.00  0.00  0.00  0.00
06:43:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   128   0.00  0.99  0.00  0.00  0.00  0.00
06:43:18 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=18 mv=0 turn=0 room -    Bed        1   128   0.00  0.99  0.00  0.00  0.00  0.00
06:43:18 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   128   0.00  0.99  0.00  0.00  0.00  0.00
06:43:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   128   0.00  0.99  0.00  0.00  0.00  0.00
06:43:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   129   0.00  0.99  0.00  0.00  0.00  0.00
06:43:20 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   129   0.00  0.99  0.00  0.00  0.00  0.00
06:43:20 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   129   0.00  0.99  0.00  0.00  0.00  0.00
06:43:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   131   0.00  0.99  0.00  0.00  0.00  0.00
06:43:21 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   131   0.00  0.99  0.00  0.00  0.00  0.00
06:43:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   131   0.00  0.99  0.00  0.00  0.00  0.00
06:43:22 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   131   0.00  0.99  0.00  0.00  0.00  0.00
06:43:22 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   131   0.00  0.99  0.00  0.00  0.00  0.00
06:43:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   132   0.00  0.99  0.00  0.00  0.00  0.00
06:43:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   133   0.00  0.99  0.00  0.00  0.00  0.00
06:43:24 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   133   0.00  0.99  0.00  0.00  0.00  0.00
06:43:24 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=17 mv=0 turn=0 room -    Bed        1   133   0.00  0.99  0.00  0.00  0.00  0.00
06:43:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   134   0.00  0.99  0.00  0.00  0.00  0.00
06:43:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   135   0.00  0.99  0.00  0.00  0.00  0.00
06:43:26 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=17 mv=0 turn=0 room -    Bed        1   135   0.00  0.99  0.00  0.00  0.00  0.00
06:43:26 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   135   0.00  0.99  0.00  0.00  0.00  0.00
06:43:26 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   136   0.00  0.99  0.00  0.00  0.00  0.00
06:43:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   137   0.00  0.99  0.00  0.00  0.00  0.00
06:43:28 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   137   0.00  0.99  0.00  0.00  0.00  0.00
06:43:28 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=17 mv=0 turn=0 room -    Bed        1   137   0.00  0.99  0.00  0.00  0.00  0.00
06:43:28 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   138   0.00  0.99  0.00  0.00  0.00  0.00
06:43:29 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   139   0.00  0.99  0.00  0.00  0.00  0.00
06:43:30 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   139   0.00  0.99  0.00  0.00  0.00  0.00
06:43:30 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=17 mv=0 turn=0 room -    Bed        1   139   0.00  0.99  0.00  0.00  0.00  0.00
06:43:30 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   140   0.00  0.99  0.00  0.00  0.00  0.00
06:43:31 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   141   0.00  0.99  0.00  0.00  0.00  0.00
06:43:32 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=17 mv=0 turn=0 room -    Bed        1   141   0.00  0.99  0.00  0.00  0.00  0.00
06:43:32 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   141   0.00  0.99  0.00  0.00  0.00  0.00
06:43:32 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   142   0.00  0.99  0.00  0.00  0.00  0.00
06:43:33 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   143   0.00  0.99  0.00  0.00  0.00  0.00
06:43:34 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   143   0.00  0.99  0.00  0.00  0.00  0.00
06:43:34 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=17 mv=0 turn=0 room -    Bed        1   143   0.00  0.99  0.00  0.00  0.00  0.00
06:43:34 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   144   0.00  0.99  0.00  0.00  0.00  0.00
06:43:35 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   145   0.00  0.99  0.00  0.00  0.00  0.00
06:43:36 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   145   0.00  0.99  0.00  0.00  0.00  0.00
06:43:36 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=17 mv=0 turn=0 room -    Bed        1   145   0.00  0.99  0.00  0.00  0.00  0.00
06:43:36 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   146   0.00  0.99  0.00  0.00  0.00  0.00
06:43:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   147   0.00  0.99  0.00  0.00  0.00  0.00
06:43:38 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   147   0.00  0.99  0.00  0.00  0.00  0.00
06:43:38 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=16 mv=0 turn=0 room -    Bed        1   147   0.00  0.99  0.00  0.00  0.00  0.00
06:43:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   148   0.00  0.99  0.00  0.00  0.00  0.00
06:43:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   149   0.00  0.99  0.00  0.00  0.00  0.00
06:43:40 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=16 mv=0 turn=0 room -    Bed        1   149   0.00  0.99  0.00  0.00  0.00  0.00
06:43:40 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   149   0.00  0.99  0.00  0.00  0.00  0.00
06:43:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   150   0.00  0.99  0.00  0.00  0.00  0.00
06:43:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   151   0.00  0.99  0.00  0.00  0.00  0.00
06:43:42 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=16 mv=0 turn=0 room -    Bed        1   151   0.00  0.99  0.00  0.00  0.00  0.00
06:43:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   152   0.00  0.99  0.00  0.00  0.00  0.00
06:43:43 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   152   0.00  0.99  0.00  0.00  0.00  0.00
06:43:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   153   0.00  0.99  0.00  0.00  0.00  0.00
06:43:44 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   153   0.00  0.99  0.00  0.00  0.00  0.00
06:43:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   154   0.00  0.99  0.00  0.00  0.00  0.00
06:43:45 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   154   0.00  0.99  0.00  0.00  0.00  0.00
06:43:45 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   155   0.00  0.99  0.00  0.00  0.00  0.00
06:43:46 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Bed        1   155   0.00  0.99  0.00  0.00  0.00  0.00
06:43:46 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   156   0.00  0.99  0.00  0.00  0.00  0.00
06:43:47 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   156   0.00  0.99  0.00  0.00  0.00  0.00
06:43:47 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   157   0.00  0.99  0.00  0.00  0.00  0.00
06:43:48 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=16 mv=0 turn=0 room -    Bed        1   157   0.00  0.99  0.00  0.00  0.00  0.00
06:43:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   158   0.00  0.99  0.00  0.00  0.00  0.00
06:43:49 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=12 mv=0 turn=0 room -    Bed        1   158   0.00  0.99  0.00  0.00  0.00  0.00
06:43:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   159   0.00  0.99  0.00  0.00  0.00  0.00
06:43:50 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=16 mv=0 turn=0 room -    Bed        1   159   0.00  0.99  0.00  0.00  0.00  0.00
06:43:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   160   0.00  0.99  0.00  0.00  0.00  0.00
06:43:51 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=12 mv=0 turn=0 room -    Bed        1   160   0.00  0.99  0.00  0.00  0.00  0.00
06:43:51 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   161   0.00  0.99  0.00  0.00  0.00  0.00
06:43:52 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Bed        1   161   0.00  0.99  0.00  0.00  0.00  0.00
06:43:52 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   162   0.00  0.99  0.00  0.00  0.00  0.00
06:43:53 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   162   0.00  0.99  0.00  0.00  0.00  0.00
06:43:53 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   162   0.00  0.99  0.00  0.00  0.00  0.00
06:43:53 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   163   0.00  0.99  0.00  0.00  0.00  0.00
06:43:54 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=16 mv=0 turn=0 room -    Bed        1   163   0.00  0.99  0.00  0.00  0.00  0.00
06:43:54 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   164   0.00  0.99  0.00  0.00  0.00  0.00
06:43:55 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   164   0.00  0.99  0.00  0.00  0.00  0.00
06:43:55 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   165   0.00  0.99  0.00  0.00  0.00  0.00
06:43:56 0865.0   -             pad     -    InBed    pad InBed HR=83 RR=16 mv=0 turn=0 room -    Bed        1   165   0.00  0.99  0.00  0.00  0.00  0.00
06:43:56 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   166   0.00  0.99  0.00  0.00  0.00  0.00
06:43:57 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   166   0.00  0.99  0.00  0.00  0.00  0.00
06:43:57 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   167   0.00  0.99  0.00  0.00  0.00  0.00
06:43:58 0865.0   -             pad     -    InBed    pad InBed HR=83 RR=15 mv=0 turn=0 room -    Bed        1   167   0.00  0.99  0.00  0.00  0.00  0.00
06:43:58 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   168   0.00  0.99  0.00  0.00  0.00  0.00
06:43:59 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   168   0.00  0.99  0.00  0.00  0.00  0.00
06:43:59 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   169   0.00  0.99  0.00  0.00  0.00  0.00
06:44:00 0865.0   -             pad     -    InBed    pad InBed HR=84 RR=15 mv=0 turn=0 room -    Bed        1   169   0.00  0.99  0.00  0.00  0.00  0.00
06:44:00 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   170   0.00  0.99  0.00  0.00  0.00  0.00
06:44:01 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   170   0.00  0.99  0.00  0.00  0.00  0.00
06:44:01 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   171   0.00  0.99  0.00  0.00  0.00  0.00
06:44:02 0865.0   -             pad     -    InBed    pad InBed HR=82 RR=16 mv=0 turn=0 room -    Bed        1   171   0.00  0.99  0.00  0.00  0.00  0.00
06:44:02 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   172   0.00  0.99  0.00  0.00  0.00  0.00
06:44:03 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   172   0.00  0.99  0.00  0.00  0.00  0.00
06:44:03 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   173   0.00  0.99  0.00  0.00  0.00  0.00
06:44:04 0865.0   -             pad     -    InBed    pad InBed HR=82 RR=16 mv=0 turn=0 room -    Bed        1   173   0.00  0.99  0.00  0.00  0.00  0.00
06:44:04 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   174   0.00  0.99  0.00  0.00  0.00  0.00
06:44:05 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=10 mv=0 turn=0 room -    Bed        1   174   0.00  0.99  0.00  0.00  0.00  0.00
06:44:05 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   175   0.00  0.99  0.00  0.00  0.00  0.00
06:44:06 0865.0   -             pad     -    InBed    pad InBed HR=82 RR=16 mv=0 turn=0 room -    Bed        1   175   0.00  0.99  0.00  0.00  0.00  0.00
06:44:06 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   176   0.00  0.99  0.00  0.00  0.00  0.00
06:44:07 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   176   0.00  0.99  0.00  0.00  0.00  0.00
06:44:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   177   0.00  0.99  0.00  0.00  0.00  0.00
06:44:08 0865.0   -             pad     -    InBed    pad InBed HR=83 RR=16 mv=0 turn=0 room -    Bed        1   177   0.00  0.99  0.00  0.00  0.00  0.00
06:44:08 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   178   0.00  0.99  0.00  0.00  0.00  0.00
06:44:09 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=10 mv=0 turn=0 room -    Bed        1   178   0.00  0.99  0.00  0.00  0.00  0.00
06:44:09 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   179   0.00  0.99  0.00  0.00  0.00  0.00
06:44:10 0865.0   -             pad     -    InBed    pad InBed HR=82 RR=16 mv=0 turn=0 room -    Bed        1   179   0.00  0.99  0.00  0.00  0.00  0.00
06:44:10 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   180   0.00  0.99  0.00  0.00  0.00  0.00
06:44:11 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=10 mv=0 turn=0 room -    Bed        1   180   0.00  0.99  0.00  0.00  0.00  0.00
06:44:11 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   181   0.00  0.99  0.00  0.00  0.00  0.00
06:44:12 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Bed        1   181   0.00  0.99  0.00  0.00  0.00  0.00
06:44:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   182   0.00  0.99  0.00  0.00  0.00  0.00
06:44:13 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   182   0.00  0.99  0.00  0.00  0.00  0.00
06:44:13 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   183   0.00  0.99  0.00  0.00  0.00  0.00
06:44:14 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   183   0.00  0.99  0.00  0.00  0.00  0.00
06:44:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   184   0.00  0.99  0.00  0.00  0.00  0.00
06:44:15 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=10 mv=0 turn=0 room -    Bed        1   184   0.00  0.99  0.00  0.00  0.00  0.00
06:44:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   185   0.00  0.99  0.00  0.00  0.00  0.00
06:44:16 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=16 mv=0 turn=0 room -    Bed        1   185   0.00  0.99  0.00  0.00  0.00  0.00
06:44:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   186   0.00  0.99  0.00  0.00  0.00  0.00
06:44:17 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   186   0.00  0.99  0.00  0.00  0.00  0.00
06:44:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   187   0.00  0.99  0.00  0.00  0.00  0.00
06:44:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   188   0.00  0.99  0.00  0.00  0.00  0.00
06:44:19 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=16 mv=0 turn=0 room -    Bed        1   188   0.00  0.99  0.00  0.00  0.00  0.00
06:44:19 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   188   0.00  0.99  0.00  0.00  0.00  0.00
06:44:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   189   0.00  0.99  0.00  0.00  0.00  0.00
06:44:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   190   0.00  0.99  0.00  0.00  0.00  0.00
06:44:21 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=16 mv=0 turn=0 room -    Bed        1   190   0.00  0.99  0.00  0.00  0.00  0.00
06:44:21 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=10 mv=0 turn=0 room -    Bed        1   190   0.00  0.99  0.00  0.00  0.00  0.00
06:44:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   191   0.00  0.99  0.00  0.00  0.00  0.00
06:44:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   192   0.00  0.99  0.00  0.00  0.00  0.00
06:44:23 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=17 mv=0 turn=0 room -    Bed        1   192   0.00  0.99  0.00  0.00  0.00  0.00
06:44:23 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=10 mv=0 turn=0 room -    Bed        1   192   0.00  0.99  0.00  0.00  0.00  0.00
06:44:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   193   0.00  0.99  0.00  0.00  0.00  0.00
06:44:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   194   0.00  0.99  0.00  0.00  0.00  0.00
06:44:25 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=17 mv=0 turn=0 room -    Bed        1   194   0.00  0.99  0.00  0.00  0.00  0.00
06:44:25 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=10 mv=0 turn=0 room -    Bed        1   194   0.00  0.99  0.00  0.00  0.00  0.00
06:44:25 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   194   0.00  0.99  0.00  0.00  0.00  0.00
06:44:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   195   0.00  0.99  0.00  0.00  0.00  0.00
06:44:26 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   196   0.00  0.99  0.00  0.00  0.00  0.00
06:44:27 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   196   0.00  0.99  0.00  0.00  0.00  0.00
06:44:27 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=10 mv=0 turn=0 room -    Bed        1   196   0.00  0.99  0.00  0.00  0.00  0.00
06:44:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   197   0.00  0.99  0.00  0.00  0.00  0.00
06:44:28 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   198   0.00  0.99  0.00  0.00  0.00  0.00
06:44:29 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=17 mv=0 turn=0 room -    Bed        1   198   0.00  0.99  0.00  0.00  0.00  0.00
06:44:29 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=10 mv=0 turn=0 room -    Bed        1   198   0.00  0.99  0.00  0.00  0.00  0.00
06:44:29 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   199   0.00  0.99  0.00  0.00  0.00  0.00
06:44:30 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   200   0.00  0.99  0.00  0.00  0.00  0.00
06:44:31 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   200   0.00  0.99  0.00  0.00  0.00  0.00
06:44:31 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=10 mv=0 turn=0 room -    Bed        1   200   0.00  0.99  0.00  0.00  0.00  0.00
06:44:31 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   201   0.00  0.99  0.00  0.00  0.00  0.00
06:44:32 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   202   0.00  0.99  0.00  0.00  0.00  0.00
06:44:33 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   202   0.00  0.99  0.00  0.00  0.00  0.00
06:44:33 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=10 mv=0 turn=0 room -    Bed        1   202   0.00  0.99  0.00  0.00  0.00  0.00
06:44:33 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   203   0.00  0.99  0.00  0.00  0.00  0.00
06:44:34 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   204   0.00  0.99  0.00  0.00  0.00  0.00
06:44:35 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=17 mv=0 turn=0 room -    Bed        1   204   0.00  0.99  0.00  0.00  0.00  0.00
06:44:35 1641.0   -             pad     -    InBed    pad InBed HR=70 RR=11 mv=0 turn=0 room -    Bed        1   204   0.00  0.99  0.00  0.00  0.00  0.00
06:44:35 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   205   0.00  0.99  0.00  0.00  0.00  0.00
06:44:36 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   206   0.00  0.99  0.00  0.00  0.00  0.00
06:44:37 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=17 mv=0 turn=0 room -    Bed        1   206   0.00  0.99  0.00  0.00  0.00  0.00
06:44:37 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=10 mv=0 turn=0 room -    Bed        1   206   0.00  0.99  0.00  0.00  0.00  0.00
06:44:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   207   0.00  0.99  0.00  0.00  0.00  0.00
06:44:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   208   0.00  0.99  0.00  0.00  0.00  0.00
06:44:39 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=17 mv=0 turn=0 room -    Bed        1   208   0.00  0.99  0.00  0.00  0.00  0.00
06:44:39 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=10 mv=0 turn=0 room -    Bed        1   208   0.00  0.99  0.00  0.00  0.00  0.00
06:44:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   209   0.00  0.99  0.00  0.00  0.00  0.00
06:44:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   210   0.00  0.99  0.00  0.00  0.00  0.00
06:44:41 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        1   210   0.00  0.99  0.00  0.00  0.00  0.00
06:44:41 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=11 mv=0 turn=0 room -    Bed        1   210   0.00  0.99  0.00  0.00  0.00  0.00
06:44:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   211   0.00  0.99  0.00  0.00  0.00  0.00
06:44:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   212   0.00  0.99  0.00  0.00  0.00  0.00
06:44:43 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=17 mv=0 turn=0 room -    Bed        1   212   0.00  0.99  0.00  0.00  0.00  0.00
06:44:43 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=10 mv=0 turn=0 room -    Bed        1   212   0.00  0.99  0.00  0.00  0.00  0.00
06:44:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   213   0.00  0.99  0.00  0.00  0.00  0.00
06:44:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   214   0.00  0.99  0.00  0.00  0.00  0.00
06:44:45 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=17 mv=0 turn=0 room -    Bed        1   214   0.00  0.99  0.00  0.00  0.00  0.00
06:44:45 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=10 mv=0 turn=0 room -    Bed        1   214   0.00  0.99  0.00  0.00  0.00  0.00
06:44:45 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   215   0.00  0.99  0.00  0.00  0.00  0.00
06:44:46 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   216   0.00  0.99  0.00  0.00  0.00  0.00
06:44:47 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   216   0.00  0.99  0.00  0.00  0.00  0.00
06:44:47 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=17 mv=0 turn=0 room -    Bed        1   216   0.00  0.99  0.00  0.00  0.00  0.00
06:44:47 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   217   0.00  0.99  0.00  0.00  0.00  0.00
06:44:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   218   0.00  0.99  0.00  0.00  0.00  0.00
06:44:49 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   218   0.00  0.99  0.00  0.00  0.00  0.00
06:44:49 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   218   0.00  0.99  0.00  0.00  0.00  0.00
06:44:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   219   0.00  0.99  0.00  0.00  0.00  0.00
06:44:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   220   0.00  0.99  0.00  0.00  0.00  0.00
06:44:51 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   220   0.00  0.99  0.00  0.00  0.00  0.00
06:44:51 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=16 mv=0 turn=0 room -    Bed        1   220   0.00  0.99  0.00  0.00  0.00  0.00
06:44:51 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   221   0.00  0.99  0.00  0.00  0.00  0.00
06:44:52 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   222   0.00  0.99  0.00  0.00  0.00  0.00
06:44:53 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Bed        1   222   0.00  0.99  0.00  0.00  0.00  0.00
06:44:53 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   222   0.00  0.99  0.00  0.00  0.00  0.00
06:44:53 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   223   0.00  0.99  0.00  0.00  0.00  0.00
06:44:54 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   224   0.00  0.99  0.00  0.00  0.00  0.00
06:44:55 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=12 mv=0 turn=0 room -    Bed        1   224   0.00  0.99  0.00  0.00  0.00  0.00
06:44:55 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=16 mv=0 turn=0 room -    Bed        1   224   0.00  0.99  0.00  0.00  0.00  0.00
06:44:55 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   225   0.00  0.99  0.00  0.00  0.00  0.00
06:44:56 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   226   0.00  0.99  0.00  0.00  0.00  0.00
06:44:56 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   226   0.00  0.99  0.00  0.00  0.00  0.00
06:44:57 0865.0   -             pad     -    InBed    pad InBed HR=82 RR=16 mv=0 turn=0 room -    Bed        1   226   0.00  0.99  0.00  0.00  0.00  0.00
06:44:57 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   226   0.00  0.99  0.00  0.00  0.00  0.00
06:44:57 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   227   0.00  0.99  0.00  0.00  0.00  0.00
06:44:58 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   228   0.00  0.99  0.00  0.00  0.00  0.00
06:44:59 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=16 mv=0 turn=0 room -    Bed        1   228   0.00  0.99  0.00  0.00  0.00  0.00
06:44:59 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   229   0.00  0.99  0.00  0.00  0.00  0.00
06:45:00 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   229   0.00  0.99  0.00  0.00  0.00  0.00
06:45:00 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   230   0.00  0.99  0.00  0.00  0.00  0.00
06:45:01 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=16 mv=0 turn=0 room -    Bed        1   230   0.00  0.99  0.00  0.00  0.00  0.00
06:45:01 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   231   0.00  0.99  0.00  0.00  0.00  0.00
06:45:02 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=12 mv=0 turn=0 room -    Bed        1   231   0.00  0.99  0.00  0.00  0.00  0.00
06:45:02 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   232   0.00  0.99  0.00  0.00  0.00  0.00
06:45:03 0865.0   -             pad     -    InBed    pad InBed HR=79 RR=16 mv=0 turn=0 room -    Bed        1   232   0.00  0.99  0.00  0.00  0.00  0.00
06:45:03 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   233   0.00  0.99  0.00  0.00  0.00  0.00
06:45:03 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   234   0.00  0.99  0.00  0.00  0.00  0.00
06:45:04 1641.0   -             pad     -    InBed    pad InBed HR=59 RR=11 mv=0 turn=0 room -    Bed        1   234   0.00  0.99  0.00  0.00  0.00  0.00
06:45:04 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   235   0.00  0.99  0.00  0.00  0.00  0.00
06:45:05 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Bed        1   235   0.00  0.99  0.00  0.00  0.00  0.00
06:45:05 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   236   0.00  0.99  0.00  0.00  0.00  0.00
06:45:06 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=11 mv=0 turn=0 room -    Bed        1   236   0.00  0.99  0.00  0.00  0.00  0.00
06:45:06 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   237   0.00  0.99  0.00  0.00  0.00  0.00
06:45:07 0865.0   -             pad     -    InBed    pad InBed HR=83 RR=16 mv=0 turn=0 room -    Bed        1   237   0.00  0.99  0.00  0.00  0.00  0.00
06:45:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   238   0.00  0.99  0.00  0.00  0.00  0.00
06:45:08 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=12 mv=0 turn=0 room -    Bed        1   238   0.00  0.99  0.00  0.00  0.00  0.00
06:45:08 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   239   0.00  0.99  0.00  0.00  0.00  0.00
06:45:09 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=16 mv=0 turn=0 room -    Bed        1   239   0.00  0.99  0.00  0.00  0.00  0.00
06:45:09 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   240   0.00  0.99  0.00  0.00  0.00  0.00
06:45:10 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   240   0.00  0.99  0.00  0.00  0.00  0.00
06:45:10 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   241   0.00  0.99  0.00  0.00  0.00  0.00
06:45:11 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=17 mv=0 turn=0 room -    Bed        1   241   0.00  0.99  0.00  0.00  0.00  0.00
06:45:11 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   242   0.00  0.99  0.00  0.00  0.00  0.00
06:45:12 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   242   0.00  0.99  0.00  0.00  0.00  0.00
06:45:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   243   0.00  0.99  0.00  0.00  0.00  0.00
06:45:13 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=17 mv=0 turn=0 room -    Bed        1   243   0.00  0.99  0.00  0.00  0.00  0.00
06:45:13 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   244   0.00  0.99  0.00  0.00  0.00  0.00
06:45:14 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   244   0.00  0.99  0.00  0.00  0.00  0.00
06:45:15 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=18 mv=0 turn=0 room -    Bed        1   245   0.00  0.99  0.00  0.00  0.00  0.00
06:45:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   245   0.00  0.99  0.00  0.00  0.00  0.00
06:45:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   245   0.00  0.99  0.00  0.00  0.00  0.00
06:45:16 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   245   0.00  0.99  0.00  0.00  0.00  0.00
06:45:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   246   0.00  0.99  0.00  0.00  0.00  0.00
06:45:17 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=19 mv=0 turn=0 room -    Bed        1   246   0.00  0.99  0.00  0.00  0.00  0.00
06:45:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   247   0.00  0.99  0.00  0.00  0.00  0.00
06:45:18 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   247   0.00  0.99  0.00  0.00  0.00  0.00
06:45:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   248   0.00  0.99  0.00  0.00  0.00  0.00
06:45:19 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=19 mv=0 turn=0 room -    Bed        1   248   0.00  0.99  0.00  0.00  0.00  0.00
06:45:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   249   0.00  0.99  0.00  0.00  0.00  0.00
06:45:20 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   249   0.00  0.99  0.00  0.00  0.00  0.00
06:45:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   250   0.00  0.99  0.00  0.00  0.00  0.00
06:45:21 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=18 mv=0 turn=0 room -    Bed        1   250   0.00  0.99  0.00  0.00  0.00  0.00
06:45:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   251   0.00  0.99  0.00  0.00  0.00  0.00
06:45:22 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   251   0.00  0.99  0.00  0.00  0.00  0.00
06:45:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   252   0.00  0.99  0.00  0.00  0.00  0.00
06:45:23 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=18 mv=0 turn=0 room -    Bed        1   252   0.00  0.99  0.00  0.00  0.00  0.00
06:45:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   253   0.00  0.99  0.00  0.00  0.00  0.00
06:45:24 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   253   0.00  0.99  0.00  0.00  0.00  0.00
06:45:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   254   0.00  0.99  0.00  0.00  0.00  0.00
06:45:25 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=18 mv=0 turn=0 room -    Bed        1   254   0.00  0.99  0.00  0.00  0.00  0.00
06:45:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   255   0.00  0.99  0.00  0.00  0.00  0.00
06:45:26 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   255   0.00  0.99  0.00  0.00  0.00  0.00
06:45:26 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   256   0.00  0.99  0.00  0.00  0.00  0.00
06:45:27 0865.0   -             pad     -    InBed    pad InBed HR=81 RR=18 mv=0 turn=0 room -    Bed        1   256   0.00  0.99  0.00  0.00  0.00  0.00
06:45:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   257   0.00  0.99  0.00  0.00  0.00  0.00
06:45:28 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   257   0.00  0.99  0.00  0.00  0.00  0.00
06:45:28 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   257   0.00  0.99  0.00  0.00  0.00  0.00
06:45:28 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   258   0.00  0.99  0.00  0.00  0.00  0.00
06:45:29 0865.0   -             pad     -    InBed    pad InBed HR=83 RR=17 mv=0 turn=0 room -    Bed        1   258   0.00  0.99  0.00  0.00  0.00  0.00
06:45:29 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   259   0.00  0.99  0.00  0.00  0.00  0.00
06:45:30 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   259   0.00  0.99  0.00  0.00  0.00  0.00
06:45:30 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   260   0.00  0.99  0.00  0.00  0.00  0.00
06:45:31 0865.0   -             pad     -    InBed    pad InBed HR=84 RR=17 mv=0 turn=0 room -    Bed        1   260   0.00  0.99  0.00  0.00  0.00  0.00
06:45:31 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   261   0.00  0.99  0.00  0.00  0.00  0.00
06:45:32 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   261   0.00  0.99  0.00  0.00  0.00  0.00
06:45:32 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   262   0.00  0.99  0.00  0.00  0.00  0.00
06:45:33 0865.0   -             pad     -    InBed    pad InBed HR=84 RR=17 mv=0 turn=0 room -    Bed        1   262   0.00  0.99  0.00  0.00  0.00  0.00
06:45:33 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   263   0.00  0.99  0.00  0.00  0.00  0.00
06:45:34 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   263   0.00  0.99  0.00  0.00  0.00  0.00
06:45:34 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   264   0.00  0.99  0.00  0.00  0.00  0.00
06:45:35 0865.0   -             pad     -    InBed    pad InBed HR=82 RR=17 mv=0 turn=0 room -    Bed        1   264   0.00  0.99  0.00  0.00  0.00  0.00
06:45:35 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   265   0.00  0.99  0.00  0.00  0.00  0.00
06:45:36 1641.0   -             pad     -    InBed    pad InBed HR=57 RR=11 mv=0 turn=0 room -    Bed        1   265   0.00  0.99  0.00  0.00  0.00  0.00
06:45:36 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   266   0.00  0.99  0.00  0.00  0.00  0.00
06:45:37 0865.0   -             pad     -    InBed    pad InBed HR=80 RR=16 mv=0 turn=0 room -    Bed        1   266   0.00  0.99  0.00  0.00  0.00  0.00
06:45:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   267   0.00  0.99  0.00  0.00  0.00  0.00
06:45:38 1641.0   -             pad     -    InBed    pad InBed HR=57 RR=11 mv=0 turn=0 room -    Bed        1   267   0.00  0.99  0.00  0.00  0.00  0.00
06:45:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   268   0.00  0.99  0.00  0.00  0.00  0.00
06:45:39 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   268   0.00  0.99  0.00  0.00  0.00  0.00
06:45:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   269   0.00  0.99  0.00  0.00  0.00  0.00
06:45:40 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=12 mv=0 turn=0 room -    Bed        1   269   0.00  0.99  0.00  0.00  0.00  0.00
06:45:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   270   0.00  0.99  0.00  0.00  0.00  0.00
06:45:41 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   270   0.00  0.99  0.00  0.00  0.00  0.00
06:45:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   271   0.00  0.99  0.00  0.00  0.00  0.00
06:45:42 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   271   0.00  0.99  0.00  0.00  0.00  0.00
06:45:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   272   0.00  0.99  0.00  0.00  0.00  0.00
06:45:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   273   0.00  0.99  0.00  0.00  0.00  0.00
06:45:44 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=12 mv=0 turn=0 room -    Bed        1   273   0.00  0.99  0.00  0.00  0.00  0.00
06:45:44 0865.0   -             pad     -    InBed    pad InBed HR=77 RR=16 mv=0 turn=0 room -    Bed        1   273   0.00  0.99  0.00  0.00  0.00  0.00
06:45:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   274   0.00  0.99  0.00  0.00  0.00  0.00
06:45:45 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   275   0.00  0.99  0.00  0.00  0.00  0.00
06:45:46 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   275   0.00  0.99  0.00  0.00  0.00  0.00
06:45:46 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=16 mv=0 turn=0 room -    Bed        1   275   0.00  0.99  0.00  0.00  0.00  0.00
06:45:46 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   276   0.00  0.99  0.00  0.00  0.00  0.00
06:45:47 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   277   0.00  0.99  0.00  0.00  0.00  0.00
06:45:48 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   277   0.00  0.99  0.00  0.00  0.00  0.00
06:45:48 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=16 mv=0 turn=0 room -    Bed        1   277   0.00  0.99  0.00  0.00  0.00  0.00
06:45:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   278   0.00  0.99  0.00  0.00  0.00  0.00
06:45:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   279   0.00  0.99  0.00  0.00  0.00  0.00
06:45:50 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=12 mv=0 turn=0 room -    Bed        1   279   0.00  0.99  0.00  0.00  0.00  0.00
06:45:50 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   279   0.00  0.99  0.00  0.00  0.00  0.00
06:45:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   280   0.00  0.99  0.00  0.00  0.00  0.00
06:45:51 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   281   0.00  0.99  0.00  0.00  0.00  0.00
06:45:52 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=12 mv=0 turn=0 room -    Bed        1   281   0.00  0.99  0.00  0.00  0.00  0.00
06:45:52 0865.0   -             pad     -    InBed    pad InBed HR=78 RR=16 mv=0 turn=0 room -    Bed        1   281   0.00  0.99  0.00  0.00  0.00  0.00
06:45:52 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   282   0.00  0.99  0.00  0.00  0.00  0.00
06:45:53 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   283   0.00  0.99  0.00  0.00  0.00  0.00
06:45:54 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=12 mv=0 turn=0 room -    Bed        1   283   0.00  0.99  0.00  0.00  0.00  0.00
06:45:54 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=16 mv=0 turn=0 room -    Bed        1   283   0.00  0.99  0.00  0.00  0.00  0.00
06:45:54 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   284   0.00  0.99  0.00  0.00  0.00  0.00
06:45:55 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   285   0.00  0.99  0.00  0.00  0.00  0.00
06:45:56 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   285   0.00  0.99  0.00  0.00  0.00  0.00
06:45:56 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=16 mv=0 turn=0 room -    Bed        1   285   0.00  0.99  0.00  0.00  0.00  0.00
06:45:56 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   286   0.00  0.99  0.00  0.00  0.00  0.00
06:45:57 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   287   0.00  0.99  0.00  0.00  0.00  0.00
06:45:58 0865.0   -             pad     -    InBed    pad InBed HR=71 RR=16 mv=0 turn=0 room -    Bed        1   287   0.00  0.99  0.00  0.00  0.00  0.00
06:45:58 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   287   0.00  0.99  0.00  0.00  0.00  0.00
06:45:58 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   288   0.00  0.99  0.00  0.00  0.00  0.00
06:45:59 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   289   0.00  0.99  0.00  0.00  0.00  0.00
06:46:00 0865.0   -             pad     -    InBed    pad InBed HR=72 RR=16 mv=0 turn=0 room -    Bed        1   289   0.00  0.99  0.00  0.00  0.00  0.00
06:46:00 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   289   0.00  0.99  0.00  0.00  0.00  0.00
06:46:00 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   289   0.00  0.99  0.00  0.00  0.00  0.00
06:46:00 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   290   0.00  0.99  0.00  0.00  0.00  0.00
06:46:01 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   291   0.00  0.99  0.00  0.00  0.00  0.00
06:46:02 0865.0   -             pad     -    InBed    pad InBed HR=72 RR=15 mv=0 turn=0 room -    Bed        1   291   0.00  0.99  0.00  0.00  0.00  0.00
06:46:02 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=13 mv=0 turn=0 room -    Bed        1   291   0.00  0.99  0.00  0.00  0.00  0.00
06:46:02 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   292   0.00  0.99  0.00  0.00  0.00  0.00
06:46:03 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   293   0.00  0.99  0.00  0.00  0.00  0.00
06:46:04 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=14 mv=0 turn=0 room -    Bed        1   293   0.00  0.99  0.00  0.00  0.00  0.00
06:46:04 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   293   0.00  0.99  0.00  0.00  0.00  0.00
06:46:04 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   294   0.00  0.99  0.00  0.00  0.00  0.00
06:46:05 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   295   0.00  0.99  0.00  0.00  0.00  0.00
06:46:06 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=15 mv=0 turn=0 room -    Bed        1   295   0.00  0.99  0.00  0.00  0.00  0.00
06:46:06 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   295   0.00  0.99  0.00  0.00  0.00  0.00
06:46:06 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   296   0.00  0.99  0.00  0.00  0.00  0.00
06:46:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   297   0.00  0.99  0.00  0.00  0.00  0.00
06:46:08 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=15 mv=0 turn=0 room -    Bed        1   297   0.00  0.99  0.00  0.00  0.00  0.00
06:46:08 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   297   0.00  0.99  0.00  0.00  0.00  0.00
06:46:08 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   298   0.00  0.99  0.00  0.00  0.00  0.00
06:46:09 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   299   0.00  0.99  0.00  0.00  0.00  0.00
06:46:10 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=15 mv=0 turn=0 room -    Bed        1   299   0.00  0.99  0.00  0.00  0.00  0.00
06:46:10 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=12 mv=0 turn=0 room -    Bed        1   299   0.00  0.99  0.00  0.00  0.00  0.00
06:46:10 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   300   0.00  0.99  0.00  0.00  0.00  0.00
06:46:11 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   301   0.00  0.99  0.00  0.00  0.00  0.00
06:46:12 0865.0   -             pad     -    InBed    pad InBed HR=75 RR=15 mv=0 turn=0 room -    Bed        1   301   0.00  0.99  0.00  0.00  0.00  0.00
06:46:12 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=13 mv=0 turn=0 room -    Bed        1   301   0.00  0.99  0.00  0.00  0.00  0.00
06:46:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   302   0.00  0.99  0.00  0.00  0.00  0.00
06:46:13 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   303   0.00  0.99  0.00  0.00  0.00  0.00
06:46:14 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   303   0.00  0.99  0.00  0.00  0.00  0.00
06:46:14 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=15 mv=0 turn=0 room -    Bed        1   303   0.00  0.99  0.00  0.00  0.00  0.00
06:46:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   304   0.00  0.99  0.00  0.00  0.00  0.00
06:46:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   305   0.00  0.99  0.00  0.00  0.00  0.00
06:46:16 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=16 mv=0 turn=0 room -    Bed        1   305   0.00  0.99  0.00  0.00  0.00  0.00
06:46:16 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   305   0.00  0.99  0.00  0.00  0.00  0.00
06:46:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   306   0.00  0.99  0.00  0.00  0.00  0.00
06:46:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   307   0.00  0.99  0.00  0.00  0.00  0.00
06:46:18 0865.0   -             pad     -    InBed    pad InBed HR=68 RR=16 mv=0 turn=0 room -    Bed        1   307   0.00  0.99  0.00  0.00  0.00  0.00
06:46:18 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=12 mv=0 turn=0 room -    Bed        1   307   0.00  0.99  0.00  0.00  0.00  0.00
06:46:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   308   0.00  0.99  0.00  0.00  0.00  0.00
06:46:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   309   0.00  0.99  0.00  0.00  0.00  0.00
06:46:20 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   309   0.00  0.99  0.00  0.00  0.00  0.00
06:46:20 0865.0   -             pad     -    InBed    pad InBed HR=69 RR=15 mv=0 turn=0 room -    Bed        1   309   0.00  0.99  0.00  0.00  0.00  0.00
06:46:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   310   0.00  0.99  0.00  0.00  0.00  0.00
06:46:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   311   0.00  0.99  0.00  0.00  0.00  0.00
06:46:22 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=15 mv=0 turn=0 room -    Bed        1   311   0.00  0.99  0.00  0.00  0.00  0.00
06:46:22 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   311   0.00  0.99  0.00  0.00  0.00  0.00
06:46:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   312   0.00  0.99  0.00  0.00  0.00  0.00
06:46:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   313   0.00  0.99  0.00  0.00  0.00  0.00
06:46:24 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=15 mv=0 turn=0 room -    Bed        1   313   0.00  0.99  0.00  0.00  0.00  0.00
06:46:24 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   313   0.00  0.99  0.00  0.00  0.00  0.00
06:46:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   314   0.00  0.99  0.00  0.00  0.00  0.00
06:46:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   315   0.00  0.99  0.00  0.00  0.00  0.00
06:46:26 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=15 mv=0 turn=0 room -    Bed        1   315   0.00  0.99  0.00  0.00  0.00  0.00
06:46:26 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   315   0.00  0.99  0.00  0.00  0.00  0.00
06:46:26 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   316   0.00  0.99  0.00  0.00  0.00  0.00
06:46:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   317   0.00  0.99  0.00  0.00  0.00  0.00
06:46:28 0865.0   -             pad     -    InBed    pad InBed HR=71 RR=17 mv=0 turn=0 room -    Bed        1   317   0.00  0.99  0.00  0.00  0.00  0.00
06:46:28 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=12 mv=0 turn=0 room -    Bed        1   317   0.00  0.99  0.00  0.00  0.00  0.00
06:46:28 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   318   0.00  0.99  0.00  0.00  0.00  0.00
06:46:29 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   319   0.00  0.99  0.00  0.00  0.00  0.00
06:46:30 0865.0   -             pad     -    InBed    pad InBed HR=72 RR=17 mv=0 turn=0 room -    Bed        1   319   0.00  0.99  0.00  0.00  0.00  0.00
06:46:30 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   319   0.00  0.99  0.00  0.00  0.00  0.00
06:46:30 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   320   0.00  0.99  0.00  0.00  0.00  0.00
06:46:31 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   321   0.00  0.99  0.00  0.00  0.00  0.00
06:46:32 0865.0   -             pad     -    InBed    pad InBed HR=72 RR=17 mv=0 turn=0 room -    Bed        1   321   0.00  0.99  0.00  0.00  0.00  0.00
06:46:32 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=12 mv=0 turn=0 room -    Bed        1   321   0.00  0.99  0.00  0.00  0.00  0.00
06:46:32 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   321   0.00  0.99  0.00  0.00  0.00  0.00
06:46:32 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   322   0.00  0.99  0.00  0.00  0.00  0.00
06:46:33 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   323   0.00  0.99  0.00  0.00  0.00  0.00
06:46:34 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=16 mv=0 turn=0 room -    Bed        1   323   0.00  0.99  0.00  0.00  0.00  0.00
06:46:34 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   323   0.00  0.99  0.00  0.00  0.00  0.00
06:46:34 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   324   0.00  0.99  0.00  0.00  0.00  0.00
06:46:35 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   325   0.00  0.99  0.00  0.00  0.00  0.00
06:46:36 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=16 mv=0 turn=0 room -    Bed        1   325   0.00  0.99  0.00  0.00  0.00  0.00
06:46:36 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=12 mv=0 turn=0 room -    Bed        1   325   0.00  0.99  0.00  0.00  0.00  0.00
06:46:36 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   326   0.00  0.99  0.00  0.00  0.00  0.00
06:46:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   327   0.00  0.99  0.00  0.00  0.00  0.00
06:46:38 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=16 mv=0 turn=0 room -    Bed        1   327   0.00  0.99  0.00  0.00  0.00  0.00
06:46:38 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=12 mv=0 turn=0 room -    Bed        1   327   0.00  0.99  0.00  0.00  0.00  0.00
06:46:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   328   0.00  0.99  0.00  0.00  0.00  0.00
06:46:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   329   0.00  0.99  0.00  0.00  0.00  0.00
06:46:40 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=17 mv=0 turn=0 room -    Bed        1   329   0.00  0.99  0.00  0.00  0.00  0.00
06:46:40 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   329   0.00  0.99  0.00  0.00  0.00  0.00
06:46:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   330   0.00  0.99  0.00  0.00  0.00  0.00
06:46:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   331   0.00  0.99  0.00  0.00  0.00  0.00
06:46:42 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=17 mv=0 turn=0 room -    Bed        1   331   0.00  0.99  0.00  0.00  0.00  0.00
06:46:42 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   331   0.00  0.99  0.00  0.00  0.00  0.00
06:46:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   332   0.00  0.99  0.00  0.00  0.00  0.00
06:46:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   333   0.00  0.99  0.00  0.00  0.00  0.00
06:46:44 0865.0   -             pad     -    InBed    pad InBed HR=68 RR=18 mv=0 turn=0 room -    Bed        1   333   0.00  0.99  0.00  0.00  0.00  0.00
06:46:44 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   333   0.00  0.99  0.00  0.00  0.00  0.00
06:46:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   334   0.00  0.99  0.00  0.00  0.00  0.00
06:46:45 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   335   0.00  0.99  0.00  0.00  0.00  0.00
06:46:46 0865.0   -             pad     -    InBed    pad InBed HR=68 RR=18 mv=0 turn=0 room -    Bed        1   335   0.00  0.99  0.00  0.00  0.00  0.00
06:46:46 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   335   0.00  0.99  0.00  0.00  0.00  0.00
06:46:46 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   336   0.00  0.99  0.00  0.00  0.00  0.00
06:46:47 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   337   0.00  0.99  0.00  0.00  0.00  0.00
06:46:48 0865.0   -             pad     -    InBed    pad InBed HR=68 RR=18 mv=0 turn=0 room -    Bed        1   337   0.00  0.99  0.00  0.00  0.00  0.00
06:46:48 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=11 mv=0 turn=0 room -    Bed        1   337   0.00  0.99  0.00  0.00  0.00  0.00
06:46:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   338   0.00  0.99  0.00  0.00  0.00  0.00
06:46:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   339   0.00  0.99  0.00  0.00  0.00  0.00
06:46:50 0865.0   -             pad     -    InBed    pad InBed HR=69 RR=18 mv=0 turn=0 room -    Bed        1   339   0.00  0.99  0.00  0.00  0.00  0.00
06:46:50 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=11 mv=0 turn=0 room -    Bed        1   339   0.00  0.99  0.00  0.00  0.00  0.00
06:46:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   340   0.00  0.99  0.00  0.00  0.00  0.00
06:46:51 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   341   0.00  0.99  0.00  0.00  0.00  0.00
06:46:52 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=19 mv=0 turn=0 room -    Bed        1   341   0.00  0.99  0.00  0.00  0.00  0.00
06:46:52 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=11 mv=0 turn=0 room -    Bed        1   341   0.00  0.99  0.00  0.00  0.00  0.00
06:46:52 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   342   0.00  0.99  0.00  0.00  0.00  0.00
06:46:53 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   343   0.00  0.99  0.00  0.00  0.00  0.00
06:46:54 0865.0   -             pad     -    InBed    pad InBed HR=71 RR=20 mv=0 turn=0 room -    Bed        1   343   0.00  0.99  0.00  0.00  0.00  0.00
06:46:54 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=11 mv=0 turn=0 room -    Bed        1   343   0.00  0.99  0.00  0.00  0.00  0.00
06:46:54 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   344   0.00  0.99  0.00  0.00  0.00  0.00
06:46:55 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   345   0.00  0.99  0.00  0.00  0.00  0.00
06:46:56 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=11 mv=0 turn=0 room -    Bed        1   345   0.00  0.99  0.00  0.00  0.00  0.00
06:46:56 0865.0   -             pad     -    InBed    pad InBed HR=70 RR=20 mv=0 turn=0 room -    Bed        1   345   0.00  0.99  0.00  0.00  0.00  0.00
06:46:56 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   346   0.00  0.99  0.00  0.00  0.00  0.00
06:46:57 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   347   0.00  0.99  0.00  0.00  0.00  0.00
06:46:58 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=0 turn=0 room -    Bed        1   347   0.00  0.99  0.00  0.00  0.00  0.00
06:46:58 0865.0   -             pad     -    InBed    pad InBed HR=69 RR=21 mv=0 turn=0 room -    Bed        1   347   0.00  0.99  0.00  0.00  0.00  0.00
06:46:58 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   348   0.00  0.99  0.00  0.00  0.00  0.00
06:46:59 CD2B.0   CD2B04000458  lying   75   InBed    lying              trk  1.00 Bed        1   349   0.00  0.99  0.00  0.00  0.00  0.00
06:47:00 0865.0   -             pad     -    InBed    pad InBed HR=74 RR=21 mv=0 turn=0 room -    Bed        1   349   0.00  0.99  0.00  0.00  0.00  0.00
06:47:00 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=None mv=1 turn=0 room -    Bed        1   349   0.00  0.99  0.00  0.00  0.00  0.00
06:47:00 CD2B.0   CD2B04000458  lying   79   InBed    lying              trk  1.00 Bed        1   350   0.00  0.99  0.00  0.00  0.00  0.00
06:47:01 CD2B.0   CD2B04000458  lying   70   InBed    lying              trk  1.00 Bed        1   351   0.00  0.99  0.00  0.00  0.00  0.00
06:47:02 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=None mv=0 turn=0 room -    Bed        1   351   0.00  0.99  0.00  0.00  0.00  0.00
06:47:02 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=19 mv=0 turn=0 room -    Bed        1   351   0.00  0.99  0.00  0.00  0.00  0.00
06:47:02 CD2B.0   CD2B04000458  lying   70   InBed    lying              trk  1.00 Bed        1   352   0.00  0.99  0.00  0.00  0.00  0.00
06:47:03 CD2B.0   CD2B04000458  lying   71   InBed    lying              trk  1.00 Bed        1   353   0.00  0.99  0.00  0.00  0.00  0.00
06:47:03 333B.E   -             -       0    InBed    np=0  ★0           room -    Bed        1   353   0.00  0.99  0.00  0.00  0.00  0.00
06:47:03 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   353   0.00  0.99  0.00  0.00  0.00  0.00
06:47:04 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=20 mv=0 turn=0 room -    Bed        1   353   0.00  0.99  0.00  0.00  0.00  0.00
06:47:04 CD2B.0   CD2B04000458  lying   77   InBed    lying              trk  1.00 Bed        1   354   0.00  0.99  0.00  0.00  0.00  0.00
06:47:05 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   354   0.00  0.99  0.00  0.00  0.00  0.00
06:47:05 CD2B.0   CD2B04000458  lying   67   InBed    lying              trk  1.00 Bed        1   0     0.00  0.96  0.01  0.02  0.00  0.00
06:47:06 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.01  0.02  0.00  0.00
06:47:06 CD2B.0   CD2B04000458  lying   67   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:07 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:08 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:08 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:09 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=1 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:09 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:10 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:10 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:11 1641.0   -             pad     -    InBed    pad InBed HR=None RR=11 mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:11 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:12 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:13 1641.0   -             pad     -    InBed    pad InBed HR=None RR=11 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:14 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=11 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:15 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=11 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:15 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:17 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=11 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:17 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:19 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=11 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:19 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:21 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=11 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:21 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:23 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:23 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:25 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:25 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:26 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:27 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:27 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:28 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:29 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:29 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:29 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:30 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:31 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:31 1641.0   -             pad     -    InBed    pad InBed HR=67 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:31 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:32 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   27    0.00  0.99  0.00  0.00  0.00  0.00
06:47:33 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   27    0.00  0.99  0.00  0.00  0.00  0.00
06:47:33 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=18 mv=0 turn=0 room -    Bed        1   27    0.00  0.99  0.00  0.00  0.00  0.00
06:47:33 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   28    0.00  0.99  0.00  0.00  0.00  0.00
06:47:34 CD2B.0   CD2B04000458  lying   70   InBed    lying              trk  1.00 Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:47:35 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:47:35 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=None mv=0 turn=0 room -    Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:47:35 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:47:35 CD2B.0   CD2B04000458  lying   64   InBed    lying              trk  1.00 Bed        1   30    0.00  0.99  0.00  0.00  0.00  0.00
06:47:36 CD2B.0   CD2B04000458  lying   66   InBed    lying              trk  1.00 Bed        1   31    0.00  0.99  0.00  0.00  0.00  0.00
06:47:37 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=1 turn=0 room -    Bed        1   31    0.00  0.99  0.00  0.00  0.00  0.00
06:47:37 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   31    0.00  0.99  0.00  0.00  0.00  0.00
06:47:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:39 0865.0   -             pad     -    InBed    pad InBed HR=None RR=19 mv=0 turn=1 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:39 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.96  0.01  0.02  0.00  0.00
06:47:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:41 0865.0   -             pad     -    InBed    pad InBed HR=None RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:41 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=1 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:43 1641.0   -             pad     -    InBed    pad InBed HR=None RR=20 mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:47:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:44 0865.0   -             pad     -    InBed    pad InBed HR=None RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:45 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:45 1641.0   -             pad     -    InBed    pad InBed HR=None RR=20 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:45 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:46 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:46 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:47 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=20 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:47 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:48 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:49 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=20 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:50 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:51 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:51 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:52 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:52 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:53 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:53 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:54 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:54 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:55 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:55 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:56 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:56 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:57 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:57 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:58 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:58 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:59 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:47:59 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:00 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:00 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:01 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:01 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:02 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:02 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:03 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:03 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:04 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:04 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:05 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:05 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   28    0.00  0.99  0.00  0.00  0.00  0.00
06:48:06 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   28    0.00  0.99  0.00  0.00  0.00  0.00
06:48:06 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:48:07 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=16 mv=0 turn=0 room -    Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:48:07 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:48:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   30    0.00  0.99  0.00  0.00  0.00  0.00
06:48:08 0865.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=1 room -    Bed        1   30    0.00  0.99  0.00  0.00  0.00  0.00
06:48:08 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   31    0.00  0.99  0.00  0.00  0.00  0.00
06:48:09 1641.0   -             pad     -    InBed    pad InBed HR=59 RR=16 mv=0 turn=0 room -    Bed        1   31    0.00  0.99  0.00  0.00  0.00  0.00
06:48:09 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   32    0.00  0.99  0.00  0.00  0.00  0.00
06:48:10 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   32    0.00  0.99  0.00  0.00  0.00  0.00
06:48:10 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   33    0.00  0.99  0.00  0.00  0.00  0.00
06:48:11 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=16 mv=0 turn=0 room -    Bed        1   33    0.00  0.99  0.00  0.00  0.00  0.00
06:48:11 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   34    0.00  0.99  0.00  0.00  0.00  0.00
06:48:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   35    0.00  0.99  0.00  0.00  0.00  0.00
06:48:13 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=None mv=0 turn=0 room -    Bed        1   35    0.00  0.99  0.00  0.00  0.00  0.00
06:48:13 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=16 mv=0 turn=0 room -    Bed        1   35    0.00  0.99  0.00  0.00  0.00  0.00
06:48:13 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   36    0.00  0.99  0.00  0.00  0.00  0.00
06:48:14 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=None mv=1 turn=0 room -    Bed        1   36    0.00  0.99  0.00  0.00  0.00  0.00
06:48:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   37    0.00  0.99  0.00  0.00  0.00  0.00
06:48:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
06:48:16 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
06:48:16 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=19 mv=0 turn=0 room -    Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
06:48:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   39    0.00  0.99  0.00  0.00  0.00  0.00
06:48:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   40    0.00  0.99  0.00  0.00  0.00  0.00
06:48:18 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   40    0.00  0.99  0.00  0.00  0.00  0.00
06:48:18 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=18 mv=0 turn=0 room -    Bed        1   40    0.00  0.99  0.00  0.00  0.00  0.00
06:48:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   41    0.00  0.99  0.00  0.00  0.00  0.00
06:48:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   42    0.00  0.99  0.00  0.00  0.00  0.00
06:48:20 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   42    0.00  0.99  0.00  0.00  0.00  0.00
06:48:20 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=18 mv=0 turn=0 room -    Bed        1   42    0.00  0.99  0.00  0.00  0.00  0.00
06:48:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   43    0.00  0.99  0.00  0.00  0.00  0.00
06:48:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   44    0.00  0.99  0.00  0.00  0.00  0.00
06:48:22 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   44    0.00  0.99  0.00  0.00  0.00  0.00
06:48:22 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=18 mv=0 turn=0 room -    Bed        1   44    0.00  0.99  0.00  0.00  0.00  0.00
06:48:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   45    0.00  0.99  0.00  0.00  0.00  0.00
06:48:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   46    0.00  0.99  0.00  0.00  0.00  0.00
06:48:24 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   46    0.00  0.99  0.00  0.00  0.00  0.00
06:48:24 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=16 mv=0 turn=0 room -    Bed        1   46    0.00  0.99  0.00  0.00  0.00  0.00
06:48:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   47    0.00  0.99  0.00  0.00  0.00  0.00
06:48:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   48    0.00  0.99  0.00  0.00  0.00  0.00
06:48:26 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=16 mv=0 turn=0 room -    Bed        1   48    0.00  0.99  0.00  0.00  0.00  0.00
06:48:26 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   48    0.00  0.99  0.00  0.00  0.00  0.00
06:48:26 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   49    0.00  0.99  0.00  0.00  0.00  0.00
06:48:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   50    0.00  0.99  0.00  0.00  0.00  0.00
06:48:28 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   50    0.00  0.99  0.00  0.00  0.00  0.00
06:48:28 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=15 mv=0 turn=0 room -    Bed        1   50    0.00  0.99  0.00  0.00  0.00  0.00
06:48:28 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   51    0.00  0.99  0.00  0.00  0.00  0.00
06:48:29 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   52    0.00  0.99  0.00  0.00  0.00  0.00
06:48:30 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   52    0.00  0.99  0.00  0.00  0.00  0.00
06:48:30 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=15 mv=0 turn=0 room -    Bed        1   52    0.00  0.99  0.00  0.00  0.00  0.00
06:48:30 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   53    0.00  0.99  0.00  0.00  0.00  0.00
06:48:31 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   54    0.00  0.99  0.00  0.00  0.00  0.00
06:48:32 0865.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   54    0.00  0.99  0.00  0.00  0.00  0.00
06:48:32 1641.0   -             pad     -    InBed    pad InBed HR=67 RR=15 mv=0 turn=0 room -    Bed        1   54    0.00  0.99  0.00  0.00  0.00  0.00
06:48:32 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   55    0.00  0.99  0.00  0.00  0.00  0.00
06:48:33 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   56    0.00  0.99  0.00  0.00  0.00  0.00
06:48:34 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=14 mv=0 turn=0 room -    Bed        1   56    0.00  0.99  0.00  0.00  0.00  0.00
06:48:34 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   57    0.00  0.99  0.00  0.00  0.00  0.00
06:48:35 0865.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=0 turn=0 room -    Bed        1   57    0.00  0.99  0.00  0.00  0.00  0.00
06:48:35 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   58    0.00  0.99  0.00  0.00  0.00  0.00
06:48:36 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=15 mv=0 turn=0 room -    Bed        1   58    0.00  0.99  0.00  0.00  0.00  0.00
06:48:36 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   59    0.00  0.99  0.00  0.00  0.00  0.00
06:48:37 0865.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=0 turn=0 room -    Bed        1   59    0.00  0.99  0.00  0.00  0.00  0.00
06:48:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   60    0.00  0.99  0.00  0.00  0.00  0.00
06:48:38 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=14 mv=0 turn=0 room -    Bed        1   60    0.00  0.99  0.00  0.00  0.00  0.00
06:48:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   61    0.00  0.99  0.00  0.00  0.00  0.00
06:48:39 0865.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=0 turn=0 room -    Bed        1   61    0.00  0.99  0.00  0.00  0.00  0.00
06:48:39 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   61    0.00  0.99  0.00  0.00  0.00  0.00
06:48:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   62    0.00  0.99  0.00  0.00  0.00  0.00
06:48:40 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=16 mv=0 turn=0 room -    Bed        1   62    0.00  0.99  0.00  0.00  0.00  0.00
06:48:40 0865.0   -             pad     -    InBed    pad InBed HR=76 RR=15 mv=0 turn=0 room -    Bed        1   62    0.00  0.99  0.00  0.00  0.00  0.00
06:48:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   63    0.00  0.99  0.00  0.00  0.00  0.00
06:48:41 0865.0   -             pad     -    InBed    pad InBed HR=73 RR=15 mv=0 turn=0 room -    Bed        1   63    0.00  0.99  0.00  0.00  0.00  0.00
06:48:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   64    0.00  0.99  0.00  0.00  0.00  0.00
06:48:42 0865.0   -             pad     -    InBed    pad LeftBed HR=None RR=None mv=0 turn=0 room -    Bed        1   64    0.00  0.99  0.00  0.00  0.00  0.00
06:48:42 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=16 mv=0 turn=0 room -    Bed        1   64    0.00  0.99  0.00  0.00  0.00  0.00
06:48:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   65    0.00  0.99  0.00  0.00  0.00  0.00
06:48:42 333B.E   -             -       0    InBed    np=1               room -    Bed        1   65    0.00  0.99  0.00  0.00  0.00  0.00
06:48:42 333B.E   -             -       0    InBed    EnterRoom(rdr)     room -    Bed        1   65    0.00  0.99  0.00  0.00  0.00  0.00
06:48:42 333B.0   CD2B04000458  stand   88   InBed    stand              trk  1.00 Bed        1   65    0.00  0.99  0.00  0.00  0.00  0.00
06:48:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   66    0.00  0.99  0.00  0.00  0.00  0.00
06:48:43 333B.0   CD2B04000458  stand   85   InBed    stand              trk  1.00 Bed        1   66    0.00  0.99  0.00  0.00  0.00  0.00
06:48:44 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=None mv=0 turn=0 room -    Bed        1   66    0.00  0.99  0.00  0.00  0.00  0.00
06:48:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   67    0.00  0.99  0.00  0.00  0.00  0.00
06:48:44 0865.E   -             -       0    InBed    LeftBed(pad)       room -    Bed        1   67    0.00  0.99  0.00  0.00  0.00  0.00
06:48:44 0865.E   -             -       0    InBed    LeftBed(pad)       room -    Bed        1   67    0.00  0.99  0.00  0.00  0.00  0.00
06:48:44 333B.0   CD2B04000458  walk    85   InBed    walk               trk  1.00 Bed        1   67    0.00  0.99  0.00  0.00  0.00  0.00
06:48:45 CD2B.0   CD2B04000458  lying   68   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:45 333B.0   CD2B04000458  walk    68   InBed    walk               trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:46 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:46 CD2B.0   CD2B04000458  lying   68   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:46 333B.0   CD2B04000458  walk    80   InBed    walk               trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:48:47 CD2B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.97  0.02  0.01  0.00  0.00
06:48:47 333B.0   CD2B04000458  walk    80   InBed    walk               trk  1.00 Bed        1   0     0.00  0.97  0.02  0.01  0.00  0.00
06:48:48 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.97  0.02  0.01  0.00  0.00
06:48:48 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:48 333B.0   CD2B04000458  walk    98   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:49 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:49 333B.0   CD2B04000458  walk    80   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:50 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=15 mv=1 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:50 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:50 333B.0   CD2B04000458  walk    89   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:51 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:51 333B.0   CD2B04000458  walk    85   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:52 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:52 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:52 333B.0   CD2B04000458  walk    91   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:53 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:53 333B.0   CD2B04000458  walk    120  InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:54 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:54 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:54 333B.0   CD2B04000458  walk    112  InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:55 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:55 333B.0   CD2B04000458  walk    109  InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:56 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=16 mv=1 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:56 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:56 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:57 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:57 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:58 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:58 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:58 333B.0   CD2B04000458  stand   92   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:59 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:48:59 333B.0   CD2B04000458  stand   96   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:00 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:00 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:00 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:01 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:01 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:02 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:02 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:02 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:03 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:03 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:04 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:04 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:04 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:05 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:05 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:06 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:06 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:06 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:07 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:07 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:08 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:08 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:08 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:09 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:09 333B.0   CD2B04000458  stand   110  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:10 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=19 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:10 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:10 333B.0   CD2B04000458  stand   105  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:11 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:11 333B.0   CD2B04000458  stand   95   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:12 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=19 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:12 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:12 333B.0   CD2B04000458  stand   121  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:13 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:13 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:14 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=19 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:14 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:14 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:14 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   27    0.00  0.96  0.02  0.01  0.00  0.00
06:49:15 333B.0   CD2B04000458  stand   112  InBed    stand              trk  1.00 Bed        1   27    0.00  0.96  0.02  0.01  0.00  0.00
06:49:15 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   28    0.00  0.96  0.02  0.01  0.00  0.00
06:49:16 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=16 mv=0 turn=0 room -    Bed        1   28    0.00  0.96  0.02  0.01  0.00  0.00
06:49:16 333B.0   CD2B04000458  stand   119  InBed    stand              trk  1.00 Bed        1   28    0.00  0.96  0.02  0.01  0.00  0.00
06:49:16 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   29    0.00  0.96  0.02  0.01  0.00  0.00
06:49:17 333B.0   CD2B04000458  stand   112  InBed    stand              trk  1.00 Bed        1   29    0.00  0.96  0.02  0.01  0.00  0.00
06:49:17 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   30    0.00  0.96  0.02  0.01  0.00  0.00
06:49:18 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=17 mv=0 turn=0 room -    Bed        1   30    0.00  0.96  0.02  0.01  0.00  0.00
06:49:18 333B.0   CD2B04000458  stand   106  InBed    stand              trk  1.00 Bed        1   30    0.00  0.96  0.02  0.01  0.00  0.00
06:49:18 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   31    0.00  0.96  0.02  0.01  0.00  0.00
06:49:19 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   31    0.00  0.96  0.02  0.01  0.00  0.00
06:49:19 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   32    0.00  0.96  0.02  0.01  0.00  0.00
06:49:20 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=17 mv=0 turn=0 room -    Bed        1   32    0.00  0.96  0.02  0.01  0.00  0.00
06:49:20 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   32    0.00  0.96  0.02  0.01  0.00  0.00
06:49:20 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   33    0.00  0.96  0.02  0.01  0.00  0.00
06:49:21 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   33    0.00  0.96  0.02  0.01  0.00  0.00
06:49:21 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   34    0.00  0.96  0.02  0.01  0.00  0.00
06:49:22 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   34    0.00  0.96  0.02  0.01  0.00  0.00
06:49:22 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   35    0.00  0.96  0.02  0.01  0.00  0.00
06:49:23 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=16 mv=0 turn=0 room -    Bed        1   35    0.00  0.96  0.02  0.01  0.00  0.00
06:49:23 333B.0   CD2B04000458  stand   104  InBed    stand              trk  1.00 Bed        1   35    0.00  0.96  0.02  0.01  0.00  0.00
06:49:23 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   36    0.00  0.96  0.02  0.01  0.00  0.00
06:49:24 333B.0   CD2B04000458  stand   101  InBed    stand              trk  1.00 Bed        1   36    0.00  0.96  0.02  0.01  0.00  0.00
06:49:24 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   37    0.00  0.96  0.02  0.01  0.00  0.00
06:49:25 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=16 mv=0 turn=0 room -    Bed        1   37    0.00  0.96  0.02  0.01  0.00  0.00
06:49:25 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   37    0.00  0.96  0.02  0.01  0.00  0.00
06:49:25 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   38    0.00  0.96  0.02  0.01  0.00  0.00
06:49:26 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   38    0.00  0.96  0.02  0.01  0.00  0.00
06:49:26 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   39    0.00  0.96  0.02  0.01  0.00  0.00
06:49:27 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=15 mv=0 turn=0 room -    Bed        1   39    0.00  0.96  0.02  0.01  0.00  0.00
06:49:27 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   39    0.00  0.96  0.02  0.01  0.00  0.00
06:49:27 CD2B.0   CD2B04000458  stand   67   InBed    stand              trk  1.00 Bed        1   40    0.00  0.96  0.02  0.01  0.00  0.00
06:49:28 333B.0   CD2B04000458  stand   95   InBed    stand              trk  1.00 Bed        1   40    0.00  0.96  0.02  0.01  0.00  0.00
06:49:28 CD2B.0   CD2B04000458  walk    61   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:29 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:29 333B.0   CD2B04000458  stand   90   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:29 CD2B.0   CD2B04000458  walk    66   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:30 333B.0   CD2B04000458  stand   104  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:49:30 CD2B.0   CD2B04000458  lying   64   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:31 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:31 333B.0   CD2B04000458  stand   90   InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:31 CD2B.0   CD2B04000458  lying   63   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:32 333B.0   CD2B04000458  stand   81   InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:32 CD2B.0   CD2B04000458  lying   62   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:33 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:33 333B.0   CD2B04000458  walk    71   InBed    walk               trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:33 CD2B.0   CD2B04000458  lying   68   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:34 333B.0   CD2B04000458  walk    63   InBed    walk               trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:34 CD2B.0   CD2B04000458  lying   70   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:35 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:35 333B.0   CD2B04000458  walk    89   InBed    walk               trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:35 CD2B.0   CD2B04000458  lying   68   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:36 333B.0   CD2B04000458  walk    84   InBed    walk               trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:36 CD2B.0   CD2B04000458  lying   75   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:37 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:37 333B.0   CD2B04000458  walk    135  InBed    walk               trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:37 CD2B.0   CD2B04000458  lying   68   InBed    lying              trk  1.00 Bed        1   0     0.00  0.96  0.01  0.02  0.00  0.00
06:49:38 333B.0   CD2B04000458  walk    129  InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.01  0.02  0.00  0.00
06:49:38 CD2B.0   CD2B04000458  lying   74   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:39 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:39 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:39 CD2B.0   CD2B04000458  lying   67   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:40 333B.0   CD2B04000458  stand   136  InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:40 CD2B.0   CD2B04000458  lying   66   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:41 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:41 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:41 CD2B.0   CD2B04000458  lying   71   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:42 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:43 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:43 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:43 CD2B.0   CD2B04000458  lying   60   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:44 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:44 CD2B.0   CD2B04000458  lying   65   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:45 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:45 333B.0   CD2B04000458  stand   146  InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:45 CD2B.0   CD2B04000458  lying   66   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:46 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:46 CD2B.0   CD2B04000458  lying   65   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:47 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:47 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:47 CD2B.0   CD2B04000458  lying   66   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:48 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:49 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:49 333B.0   CD2B04000458  stand   140  InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:50 333B.0   CD2B04000458  stand   102  InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:51 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:51 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:51 CD2B.0   CD2B04000458  lying   59   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:52 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:52 CD2B.0   CD2B04000458  lying   61   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:53 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:53 333B.0   CD2B04000458  stand   126  InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:53 CD2B.0   CD2B04000458  lying   63   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:54 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:54 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:55 1641.0   -             pad     -    InBed    pad InBed HR=None RR=15 mv=0 turn=1 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:55 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:49:55 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:56 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:56 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:57 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:57 333B.0   CD2B04000458  stand   117  InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:57 CD2B.0   CD2B04000458  lying   67   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:58 333B.0   CD2B04000458  stand   114  InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:58 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:59 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:59 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:49:59 CD2B.0   CD2B04000458  lying   58   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:00 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:00 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:00 CD2B.0   CD2B04000458  lying   64   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:01 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:01 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:01 CD2B.0   CD2B04000458  lying   60   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:02 333B.0   CD2B04000458  stand   111  InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:02 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.97  0.02  0.01  0.00  0.00
06:50:03 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.97  0.02  0.01  0.00  0.00
06:50:03 333B.0   CD2B04000458  stand   85   InBed    stand              trk  1.00 Bed        1   0     0.00  0.97  0.02  0.01  0.00  0.00
06:50:03 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:04 333B.0   CD2B04000458  stand   93   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:04 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:05 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:05 333B.0   CD2B04000458  stand   66   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:05 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:06 333B.0   CD2B04000458  stand   98   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:06 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:07 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:07 333B.0   CD2B04000458  walk    79   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:07 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:08 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:08 333B.0   CD2B04000458  walk    82   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:08 CD2B.0   CD2B04000458  walk    61   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:09 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:09 333B.0   CD2B04000458  walk    74   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:09 CD2B.0   CD2B04000458  walk    81   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:10 333B.0   CD2B04000458  walk    78   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:10 CD2B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:11 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:11 333B.0   CD2B04000458  walk    81   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:11 CD2B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:12 333B.0   CD2B04000458  walk    112  InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:12 CD2B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:13 333B.0   CD2B04000458  walk    111  InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:13 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:14 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:14 333B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:14 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.83  0.09  0.06  0.00  0.01
06:50:15 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.83  0.09  0.06  0.00  0.01
06:50:15 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.75  0.13  0.07  0.00  0.01
06:50:16 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=1 room -    Bed        1   0     0.00  0.75  0.13  0.07  0.00  0.01
06:50:16 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.75  0.13  0.07  0.00  0.01
06:50:16 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.92  0.04  0.02  0.00  0.00
06:50:17 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.92  0.04  0.02  0.00  0.00
06:50:17 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.01  0.00  0.00
06:50:18 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.01  0.00  0.00
06:50:18 333B.0   CD2B04000458  stand   134  InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.01  0.00  0.00
06:50:18 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:19 333B.0   CD2B04000458  stand   138  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:19 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:20 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:20 333B.0   CD2B04000458  stand   102  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:20 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:21 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:21 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:22 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:22 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:22 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:23 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:23 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:24 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:24 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:24 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:25 333B.0   CD2B04000458  stand   96   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:25 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:26 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:26 333B.0   CD2B04000458  stand   113  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:26 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:27 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:27 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:28 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:28 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:28 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:29 333B.0   CD2B04000458  stand   100  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:29 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:30 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:30 333B.0   CD2B04000458  stand   111  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:30 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:30 333B.0   CD2B04000458  stand   119  InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:31 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:31 333B.0   CD2B04000458  stand   92   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:32 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:32 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:32 333B.0   CD2B04000458  stand   82   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:33 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:33 333B.0   CD2B04000458  walk    91   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:34 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=16 mv=1 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:34 CD2B.0   CD2B04000458  stand   73   InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:34 333B.0   CD2B04000458  walk    108  InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:35 CD2B.0   CD2B04000458  lying   66   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:35 333B.0   CD2B04000458  walk    107  InBed    walk               trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:36 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:36 CD2B.0   CD2B04000458  stand   65   InBed    stand              trk  1.00 Bed        1   0     0.00  0.85  0.08  0.06  0.00  0.01
06:50:36 333B.0   CD2B04000458  walk    81   InBed    walk               trk  1.00 Bed        1   0     0.00  0.85  0.08  0.06  0.00  0.01
06:50:37 CD2B.0   CD2B04000458  stand   57   InBed    stand              trk  1.00 Bed        1   0     0.00  0.77  0.13  0.07  0.00  0.01
06:50:37 333B.0   CD2B04000458  walk    103  InBed    walk               trk  1.00 Bed        1   0     0.00  0.77  0.13  0.07  0.00  0.01
06:50:38 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.77  0.13  0.07  0.00  0.01
06:50:38 CD2B.0   CD2B04000458  stand   58   InBed    stand              trk  1.00 Bed        1   0     0.00  0.71  0.16  0.08  0.01  0.01
06:50:38 333B.0   CD2B04000458  walk    102  InBed    walk               trk  1.00 Bed        1   0     0.00  0.71  0.16  0.08  0.01  0.01
06:50:39 CD2B.0   CD2B04000458  lying   47   InBed    lying              trk  1.00 Bed        1   0     0.00  0.89  0.04  0.04  0.00  0.01
06:50:39 333B.0   CD2B04000458  walk    71   InBed    walk               trk  1.00 Bed        1   0     0.00  0.89  0.04  0.04  0.00  0.01
06:50:40 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.89  0.04  0.04  0.00  0.01
06:50:40 CD2B.0   CD2B04000458  lying   48   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
06:50:40 333B.0   CD2B04000458  walk    82   InBed    walk               trk  1.00 Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
06:50:41 CD2B.0   CD2B04000458  lying   55   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:41 333B.0   CD2B04000458  walk    89   InBed    walk               trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:42 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:42 CD2B.0   CD2B04000458  lying   62   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:42 333B.0   CD2B04000458  walk    119  InBed    walk               trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:43 CD2B.0   CD2B04000458  stand   48   InBed    stand              trk  1.00 Bed        1   0     0.00  0.83  0.09  0.06  0.00  0.01
06:50:43 333B.0   CD2B04000458  walk    107  InBed    walk               trk  1.00 Bed        1   0     0.00  0.83  0.09  0.06  0.00  0.01
06:50:44 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.83  0.09  0.06  0.00  0.01
06:50:44 CD2B.0   CD2B04000458  stand   63   InBed    stand              trk  1.00 Bed        1   0     0.00  0.75  0.13  0.08  0.00  0.01
06:50:44 333B.0   CD2B04000458  walk    100  InBed    walk               trk  1.00 Bed        1   0     0.00  0.75  0.13  0.08  0.00  0.01
06:50:45 CD2B.0   CD2B04000458  stand   73   InBed    stand              trk  1.00 Bed        1   0     0.00  0.70  0.16  0.08  0.01  0.01
06:50:45 333B.0   CD2B04000458  walk    97   InBed    walk               trk  1.00 Bed        1   0     0.00  0.70  0.16  0.08  0.01  0.01
06:50:46 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=1 room -    Bed        1   0     0.00  0.70  0.16  0.08  0.01  0.01
06:50:46 CD2B.0   CD2B04000458  stand   71   InBed    stand              trk  1.00 Bed        1   0     0.00  0.91  0.05  0.02  0.00  0.00
06:50:46 333B.0   CD2B04000458  walk    102  InBed    walk               trk  1.00 Bed        1   0     0.00  0.91  0.05  0.02  0.00  0.00
06:50:47 CD2B.0   CD2B04000458  walk    77   InBed    walk               trk  1.00 Bed        1   0     0.00  0.95  0.03  0.01  0.00  0.00
06:50:47 333B.0   CD2B04000458  walk    116  InBed    walk               trk  1.00 Bed        1   0     0.00  0.95  0.03  0.01  0.00  0.00
06:50:48 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.03  0.01  0.00  0.00
06:50:48 CD2B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:48 333B.0   CD2B04000458  walk    83   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:50:49 CD2B.0   CD2B04000458  lying   66   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:49 333B.0   CD2B04000458  walk    76   InBed    walk               trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:50 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:50:50 CD2B.0   CD2B04000458  lying   68   InBed    lying              trk  1.00 Bed        1   0     0.00  0.96  0.01  0.02  0.00  0.00
06:50:50 333B.0   CD2B04000458  walk    84   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.01  0.02  0.00  0.00
06:50:51 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.01  0.02  0.00  0.00
06:50:51 CD2B.0   CD2B04000458  lying   82   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:51 333B.0   CD2B04000458  walk    101  InBed    walk               trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:52 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:52 CD2B.0   CD2B04000458  lying   75   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:52 333B.0   CD2B04000458  walk    75   InBed    walk               trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:53 CD2B.0   CD2B04000458  lying   81   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:53 333B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:54 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:54 CD2B.0   CD2B04000458  lying   71   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:54 333B.0   CD2B04000458  walk    114  InBed    walk               trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:55 CD2B.0   CD2B04000458  lying   85   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:55 333B.0   CD2B04000458  walk    85   InBed    walk               trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:56 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:56 CD2B.0   CD2B04000458  lying   84   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:50:56 333B.0   CD2B04000458  walk    82   InBed    walk               trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:50:57 CD2B.0   CD2B04000458  lying   75   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:57 333B.0   CD2B04000458  walk    88   InBed    walk               trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:58 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:58 CD2B.0   CD2B04000458  lying   77   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:58 333B.0   CD2B04000458  walk    67   InBed    walk               trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:50:59 CD2B.0   CD2B04000458  lying   90   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:50:59 333B.0   CD2B04000458  walk    93   InBed    walk               trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:51:00 CD2B.0   CD2B04000458  lying   74   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:00 333B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:01 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:01 CD2B.0   CD2B04000458  lying   89   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:51:01 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:51:02 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:02 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:03 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:03 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:03 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:04 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:04 333B.E   -             -       0    InBed    np=2               room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:04 333B.E1  -             -       0    InBed    EnterRoom(rdr)     room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:04 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:04 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:04 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:04 333B.1   -             stand   79   InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:05 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:51:05 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:05 333B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:05 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:05 333B.E   -             -       0    InBed    ExitRoom(rdr)      room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:06 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:06 333B.E   -             -       0    InBed    np=1               room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:06 333B.1   -             stand   61   InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:07 1641.0   -             pad     -    InBed    pad InBed HR=None RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:07 333B.1   -             stand   68   InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:08 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:08 333B.1   -             stand   78   InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:09 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=16 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:09 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:09 333B.1   -             stand   82   InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:10 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:10 333B.1   -             stand   76   InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:11 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:11 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:11 333B.1   -             stand   68   InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:12 333B.1   -             stand   99   InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:13 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:13 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:13 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:14 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:15 1641.0   -             pad     -    InBed    pad InBed HR=67 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:15 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:16 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:17 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:17 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:18 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:19 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:19 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:20 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:21 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:21 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:22 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:23 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:23 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:24 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:25 1641.0   -             pad     -    InBed    pad InBed HR=70 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:25 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:26 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:26 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:27 1641.0   -             pad     -    InBed    pad InBed HR=70 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:27 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:51:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   27    0.00  0.99  0.00  0.00  0.00  0.00
06:51:28 333B.1   -             stand   84   InBed    stand              room -    Bed        1   27    0.00  0.99  0.00  0.00  0.00  0.00
06:51:28 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   28    0.00  0.99  0.00  0.00  0.00  0.00
06:51:29 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=14 mv=0 turn=0 room -    Bed        1   28    0.00  0.99  0.00  0.00  0.00  0.00
06:51:29 333B.1   -             stand   46   InBed    stand              room -    Bed        1   28    0.00  0.99  0.00  0.00  0.00  0.00
06:51:29 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:51:30 333B.1   -             stand   0    InBed    stand              room -    Bed        1   29    0.00  0.99  0.00  0.00  0.00  0.00
06:51:30 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   30    0.00  0.99  0.00  0.00  0.00  0.00
06:51:31 1641.0   -             pad     -    InBed    pad InBed HR=67 RR=15 mv=0 turn=0 room -    Bed        1   30    0.00  0.99  0.00  0.00  0.00  0.00
06:51:31 333B.1   -             stand   0    InBed    stand              room -    Bed        1   30    0.00  0.99  0.00  0.00  0.00  0.00
06:51:31 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   31    0.00  0.99  0.00  0.00  0.00  0.00
06:51:32 333B.1   -             stand   0    InBed    stand              room -    Bed        1   31    0.00  0.99  0.00  0.00  0.00  0.00
06:51:32 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   32    0.00  0.99  0.00  0.00  0.00  0.00
06:51:33 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=15 mv=0 turn=0 room -    Bed        1   32    0.00  0.99  0.00  0.00  0.00  0.00
06:51:33 333B.1   -             stand   0    InBed    stand              room -    Bed        1   32    0.00  0.99  0.00  0.00  0.00  0.00
06:51:33 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   33    0.00  0.99  0.00  0.00  0.00  0.00
06:51:34 333B.1   -             stand   0    InBed    stand              room -    Bed        1   33    0.00  0.99  0.00  0.00  0.00  0.00
06:51:34 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   34    0.00  0.99  0.00  0.00  0.00  0.00
06:51:35 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=16 mv=0 turn=0 room -    Bed        1   34    0.00  0.99  0.00  0.00  0.00  0.00
06:51:35 333B.1   -             stand   0    InBed    stand              room -    Bed        1   34    0.00  0.99  0.00  0.00  0.00  0.00
06:51:35 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   35    0.00  0.99  0.00  0.00  0.00  0.00
06:51:36 333B.1   -             stand   0    InBed    stand              room -    Bed        1   35    0.00  0.99  0.00  0.00  0.00  0.00
06:51:36 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   36    0.00  0.99  0.00  0.00  0.00  0.00
06:51:37 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=16 mv=0 turn=0 room -    Bed        1   36    0.00  0.99  0.00  0.00  0.00  0.00
06:51:37 333B.1   -             stand   51   InBed    stand              room -    Bed        1   36    0.00  0.99  0.00  0.00  0.00  0.00
06:51:37 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   37    0.00  0.99  0.00  0.00  0.00  0.00
06:51:38 333B.1   -             stand   0    InBed    stand              room -    Bed        1   37    0.00  0.99  0.00  0.00  0.00  0.00
06:51:38 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
06:51:39 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=16 mv=0 turn=0 room -    Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
06:51:39 333B.1   -             stand   0    InBed    stand              room -    Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
06:51:39 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   39    0.00  0.99  0.00  0.00  0.00  0.00
06:51:40 333B.1   -             stand   0    InBed    stand              room -    Bed        1   39    0.00  0.99  0.00  0.00  0.00  0.00
06:51:40 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   40    0.00  0.99  0.00  0.00  0.00  0.00
06:51:41 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   40    0.00  0.99  0.00  0.00  0.00  0.00
06:51:41 333B.1   -             stand   57   InBed    stand              room -    Bed        1   40    0.00  0.99  0.00  0.00  0.00  0.00
06:51:41 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   41    0.00  0.99  0.00  0.00  0.00  0.00
06:51:42 333B.1   -             stand   77   InBed    stand              room -    Bed        1   41    0.00  0.99  0.00  0.00  0.00  0.00
06:51:42 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   42    0.00  0.99  0.00  0.00  0.00  0.00
06:51:43 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   42    0.00  0.99  0.00  0.00  0.00  0.00
06:51:43 333B.1   -             stand   0    InBed    stand              room -    Bed        1   42    0.00  0.99  0.00  0.00  0.00  0.00
06:51:43 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   43    0.00  0.99  0.00  0.00  0.00  0.00
06:51:44 333B.1   -             stand   0    InBed    stand              room -    Bed        1   43    0.00  0.99  0.00  0.00  0.00  0.00
06:51:44 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   44    0.00  0.99  0.00  0.00  0.00  0.00
06:51:45 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=14 mv=0 turn=0 room -    Bed        1   44    0.00  0.99  0.00  0.00  0.00  0.00
06:51:45 333B.1   -             stand   0    InBed    stand              room -    Bed        1   44    0.00  0.99  0.00  0.00  0.00  0.00
06:51:45 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   45    0.00  0.99  0.00  0.00  0.00  0.00
06:51:46 333B.1   -             stand   0    InBed    stand              room -    Bed        1   45    0.00  0.99  0.00  0.00  0.00  0.00
06:51:46 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   46    0.00  0.99  0.00  0.00  0.00  0.00
06:51:47 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   46    0.00  0.99  0.00  0.00  0.00  0.00
06:51:47 333B.1   -             stand   0    InBed    stand              room -    Bed        1   46    0.00  0.99  0.00  0.00  0.00  0.00
06:51:47 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   47    0.00  0.99  0.00  0.00  0.00  0.00
06:51:48 333B.1   -             stand   0    InBed    stand              room -    Bed        1   47    0.00  0.99  0.00  0.00  0.00  0.00
06:51:48 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   48    0.00  0.99  0.00  0.00  0.00  0.00
06:51:49 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   48    0.00  0.99  0.00  0.00  0.00  0.00
06:51:49 333B.1   -             stand   0    InBed    stand              room -    Bed        1   48    0.00  0.99  0.00  0.00  0.00  0.00
06:51:49 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   49    0.00  0.99  0.00  0.00  0.00  0.00
06:51:50 333B.1   -             stand   0    InBed    stand              room -    Bed        1   49    0.00  0.99  0.00  0.00  0.00  0.00
06:51:50 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   50    0.00  0.99  0.00  0.00  0.00  0.00
06:51:51 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=16 mv=0 turn=0 room -    Bed        1   50    0.00  0.99  0.00  0.00  0.00  0.00
06:51:51 333B.1   -             stand   72   InBed    stand              room -    Bed        1   50    0.00  0.99  0.00  0.00  0.00  0.00
06:51:51 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   51    0.00  0.99  0.00  0.00  0.00  0.00
06:51:52 333B.1   -             stand   75   InBed    stand              room -    Bed        1   51    0.00  0.99  0.00  0.00  0.00  0.00
06:51:52 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   52    0.00  0.99  0.00  0.00  0.00  0.00
06:51:53 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=16 mv=0 turn=0 room -    Bed        1   52    0.00  0.99  0.00  0.00  0.00  0.00
06:51:53 333B.1   -             stand   0    InBed    stand              room -    Bed        1   52    0.00  0.99  0.00  0.00  0.00  0.00
06:51:53 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   53    0.00  0.99  0.00  0.00  0.00  0.00
06:51:54 333B.1   -             stand   0    InBed    stand              room -    Bed        1   53    0.00  0.99  0.00  0.00  0.00  0.00
06:51:54 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   54    0.00  0.99  0.00  0.00  0.00  0.00
06:51:55 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=16 mv=0 turn=0 room -    Bed        1   54    0.00  0.99  0.00  0.00  0.00  0.00
06:51:55 333B.1   -             stand   46   InBed    stand              room -    Bed        1   54    0.00  0.99  0.00  0.00  0.00  0.00
06:51:55 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   55    0.00  0.99  0.00  0.00  0.00  0.00
06:51:56 333B.1   -             stand   66   InBed    stand              room -    Bed        1   55    0.00  0.99  0.00  0.00  0.00  0.00
06:51:56 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   56    0.00  0.99  0.00  0.00  0.00  0.00
06:51:57 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   56    0.00  0.99  0.00  0.00  0.00  0.00
06:51:57 333B.1   -             stand   70   InBed    stand              room -    Bed        1   56    0.00  0.99  0.00  0.00  0.00  0.00
06:51:57 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   57    0.00  0.99  0.00  0.00  0.00  0.00
06:51:58 333B.1   -             stand   85   InBed    stand              room -    Bed        1   57    0.00  0.99  0.00  0.00  0.00  0.00
06:51:58 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   58    0.00  0.99  0.00  0.00  0.00  0.00
06:51:59 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=15 mv=0 turn=0 room -    Bed        1   58    0.00  0.99  0.00  0.00  0.00  0.00
06:51:59 333B.1   -             stand   85   InBed    stand              room -    Bed        1   58    0.00  0.99  0.00  0.00  0.00  0.00
06:51:59 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   59    0.00  0.99  0.00  0.00  0.00  0.00
06:52:00 333B.1   -             stand   0    InBed    stand              room -    Bed        1   59    0.00  0.99  0.00  0.00  0.00  0.00
06:52:00 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   60    0.00  0.99  0.00  0.00  0.00  0.00
06:52:01 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=14 mv=0 turn=0 room -    Bed        1   60    0.00  0.99  0.00  0.00  0.00  0.00
06:52:01 333B.1   -             stand   0    InBed    stand              room -    Bed        1   60    0.00  0.99  0.00  0.00  0.00  0.00
06:52:01 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   61    0.00  0.99  0.00  0.00  0.00  0.00
06:52:02 333B.1   -             stand   0    InBed    stand              room -    Bed        1   61    0.00  0.99  0.00  0.00  0.00  0.00
06:52:02 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   62    0.00  0.99  0.00  0.00  0.00  0.00
06:52:03 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=15 mv=0 turn=0 room -    Bed        1   62    0.00  0.99  0.00  0.00  0.00  0.00
06:52:03 333B.1   -             stand   0    InBed    stand              room -    Bed        1   62    0.00  0.99  0.00  0.00  0.00  0.00
06:52:03 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   63    0.00  0.99  0.00  0.00  0.00  0.00
06:52:04 333B.1   -             stand   103  InBed    stand              room -    Bed        1   63    0.00  0.99  0.00  0.00  0.00  0.00
06:52:04 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   64    0.00  0.99  0.00  0.00  0.00  0.00
06:52:05 1641.0   -             pad     -    InBed    pad InBed HR=60 RR=15 mv=0 turn=0 room -    Bed        1   64    0.00  0.99  0.00  0.00  0.00  0.00
06:52:05 333B.1   -             stand   0    InBed    stand              room -    Bed        1   64    0.00  0.99  0.00  0.00  0.00  0.00
06:52:05 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   65    0.00  0.99  0.00  0.00  0.00  0.00
06:52:06 333B.1   -             stand   0    InBed    stand              room -    Bed        1   65    0.00  0.99  0.00  0.00  0.00  0.00
06:52:06 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   66    0.00  0.99  0.00  0.00  0.00  0.00
06:52:07 1641.0   -             pad     -    InBed    pad InBed HR=62 RR=15 mv=0 turn=0 room -    Bed        1   66    0.00  0.99  0.00  0.00  0.00  0.00
06:52:07 333B.1   -             stand   0    InBed    stand              room -    Bed        1   66    0.00  0.99  0.00  0.00  0.00  0.00
06:52:07 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   67    0.00  0.99  0.00  0.00  0.00  0.00
06:52:08 333B.1   -             stand   0    InBed    stand              room -    Bed        1   67    0.00  0.99  0.00  0.00  0.00  0.00
06:52:08 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   68    0.00  0.99  0.00  0.00  0.00  0.00
06:52:09 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=16 mv=0 turn=0 room -    Bed        1   68    0.00  0.99  0.00  0.00  0.00  0.00
06:52:09 333B.1   -             stand   0    InBed    stand              room -    Bed        1   68    0.00  0.99  0.00  0.00  0.00  0.00
06:52:09 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   69    0.00  0.99  0.00  0.00  0.00  0.00
06:52:10 333B.1   -             stand   0    InBed    stand              room -    Bed        1   69    0.00  0.99  0.00  0.00  0.00  0.00
06:52:10 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   70    0.00  0.99  0.00  0.00  0.00  0.00
06:52:11 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=16 mv=0 turn=0 room -    Bed        1   70    0.00  0.99  0.00  0.00  0.00  0.00
06:52:11 333B.1   -             stand   0    InBed    stand              room -    Bed        1   70    0.00  0.99  0.00  0.00  0.00  0.00
06:52:11 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   71    0.00  0.99  0.00  0.00  0.00  0.00
06:52:12 333B.1   -             stand   0    InBed    stand              room -    Bed        1   71    0.00  0.99  0.00  0.00  0.00  0.00
06:52:12 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   72    0.00  0.99  0.00  0.00  0.00  0.00
06:52:13 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=16 mv=0 turn=0 room -    Bed        1   72    0.00  0.99  0.00  0.00  0.00  0.00
06:52:13 333B.1   -             stand   0    InBed    stand              room -    Bed        1   72    0.00  0.99  0.00  0.00  0.00  0.00
06:52:13 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   73    0.00  0.99  0.00  0.00  0.00  0.00
06:52:14 333B.1   -             stand   0    InBed    stand              room -    Bed        1   73    0.00  0.99  0.00  0.00  0.00  0.00
06:52:14 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   74    0.00  0.99  0.00  0.00  0.00  0.00
06:52:15 1641.0   -             pad     -    InBed    pad InBed HR=61 RR=16 mv=0 turn=0 room -    Bed        1   74    0.00  0.99  0.00  0.00  0.00  0.00
06:52:15 333B.1   -             stand   0    InBed    stand              room -    Bed        1   74    0.00  0.99  0.00  0.00  0.00  0.00
06:52:15 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   75    0.00  0.99  0.00  0.00  0.00  0.00
06:52:16 333B.1   -             stand   95   InBed    stand              room -    Bed        1   75    0.00  0.99  0.00  0.00  0.00  0.00
06:52:16 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   76    0.00  0.99  0.00  0.00  0.00  0.00
06:52:17 1641.0   -             pad     -    InBed    pad InBed HR=59 RR=16 mv=0 turn=0 room -    Bed        1   76    0.00  0.99  0.00  0.00  0.00  0.00
06:52:17 333B.1   -             stand   0    InBed    stand              room -    Bed        1   76    0.00  0.99  0.00  0.00  0.00  0.00
06:52:17 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   77    0.00  0.99  0.00  0.00  0.00  0.00
06:52:18 333B.1   -             stand   0    InBed    stand              room -    Bed        1   77    0.00  0.99  0.00  0.00  0.00  0.00
06:52:18 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   78    0.00  0.99  0.00  0.00  0.00  0.00
06:52:19 333B.1   -             stand   98   InBed    stand              room -    Bed        1   78    0.00  0.99  0.00  0.00  0.00  0.00
06:52:19 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   79    0.00  0.99  0.00  0.00  0.00  0.00
06:52:20 1641.0   -             pad     -    InBed    pad InBed HR=58 RR=17 mv=0 turn=0 room -    Bed        1   79    0.00  0.99  0.00  0.00  0.00  0.00
06:52:20 333B.1   -             stand   0    InBed    stand              room -    Bed        1   79    0.00  0.99  0.00  0.00  0.00  0.00
06:52:20 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   80    0.00  0.99  0.00  0.00  0.00  0.00
06:52:21 333B.1   -             stand   84   InBed    stand              room -    Bed        1   80    0.00  0.99  0.00  0.00  0.00  0.00
06:52:21 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   81    0.00  0.99  0.00  0.00  0.00  0.00
06:52:22 1641.0   -             pad     -    InBed    pad InBed HR=59 RR=18 mv=0 turn=0 room -    Bed        1   81    0.00  0.99  0.00  0.00  0.00  0.00
06:52:22 333B.1   -             stand   93   InBed    stand              room -    Bed        1   81    0.00  0.99  0.00  0.00  0.00  0.00
06:52:22 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   82    0.00  0.99  0.00  0.00  0.00  0.00
06:52:23 333B.1   -             stand   0    InBed    stand              room -    Bed        1   82    0.00  0.99  0.00  0.00  0.00  0.00
06:52:23 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   83    0.00  0.99  0.00  0.00  0.00  0.00
06:52:24 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=18 mv=0 turn=0 room -    Bed        1   83    0.00  0.99  0.00  0.00  0.00  0.00
06:52:24 333B.1   -             stand   80   InBed    stand              room -    Bed        1   83    0.00  0.99  0.00  0.00  0.00  0.00
06:52:24 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   84    0.00  0.99  0.00  0.00  0.00  0.00
06:52:25 333B.1   -             stand   90   InBed    stand              room -    Bed        1   84    0.00  0.99  0.00  0.00  0.00  0.00
06:52:25 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   85    0.00  0.99  0.00  0.00  0.00  0.00
06:52:26 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=17 mv=0 turn=0 room -    Bed        1   85    0.00  0.99  0.00  0.00  0.00  0.00
06:52:26 333B.1   -             stand   69   InBed    stand              room -    Bed        1   85    0.00  0.99  0.00  0.00  0.00  0.00
06:52:26 CD2B.0   CD2B04000458  lying   97   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:27 333B.1   -             stand   87   InBed    stand              room -    Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:27 CD2B.0   CD2B04000458  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:52:28 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:52:28 333B.1   -             stand   97   InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:52:28 CD2B.0   CD2B04000458  lying   76   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:52:29 333B.1   -             stand   114  InBed    stand              room -    Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
06:52:29 CD2B.0   CD2B04000458  lying   88   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:30 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:30 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:30 CD2B.0   CD2B04000458  lying   88   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:31 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:31 CD2B.0   CD2B04000458  lying   97   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:32 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:32 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.99  0.01  0.00  0.00  0.00
06:52:32 CD2B.0   CD2B04000458  lying   98   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:52:33 333B.1   -             stand   113  InBed    stand              room -    Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:52:33 CD2B.0   CD2B04000458  lying   87   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:34 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:34 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:34 CD2B.0   CD2B04000458  lying   108  InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:35 333B.1   -             stand   77   InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:35 CD2B.0   CD2B04000458  lying   94   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:36 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:36 333B.1   -             stand   72   InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:36 CD2B.0   CD2B04000458  lying   89   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:36 333B.1   -             stand   110  InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:37 CD2B.0   CD2B04000458  lying   82   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:37 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:38 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:38 CD2B.0   CD2B04000458  lying   84   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:38 333B.1   -             stand   82   InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:39 CD2B.0   CD2B04000458  lying   78   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:52:39 333B.1   -             stand   80   InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:52:40 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:52:40 CD2B.0   CD2B04000458  lying   78   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:52:40 333B.1   -             stand   103  InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:52:41 CD2B.0   CD2B04000458  lying   82   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:52:41 333B.1   -             stand   108  InBed    stand              room -    Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:52:42 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:52:42 CD2B.0   CD2B04000458  lying   90   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:42 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:43 CD2B.0   CD2B04000458  lying   84   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:43 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:44 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:44 CD2B.0   CD2B04000458  lying   85   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:44 333B.1   -             stand   110  InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:45 CD2B.0   CD2B04000458  lying   85   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:45 333B.1   -             stand   116  InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:46 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:46 CD2B.0   CD2B04000458  lying   86   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:46 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:47 CD2B.0   CD2B04000458  lying   89   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:47 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:48 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:48 CD2B.0   CD2B04000458  lying   98   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:48 333B.1   -             stand   121  InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:49 CD2B.0   CD2B04000458  lying   89   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:49 333B.1   -             stand   95   InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:50 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:50 CD2B.0   CD2B04000458  lying   95   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:50 333B.1   -             stand   90   InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:51 CD2B.0   CD2B04000458  lying   91   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:51 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:52 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:52 CD2B.0   CD2B04000458  lying   88   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:52 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:53 CD2B.0   CD2B04000458  lying   81   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:53 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:54 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:54 CD2B.0   CD2B04000458  lying   85   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:54 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:55 CD2B.0   CD2B04000458  lying   75   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:52:55 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:52:56 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:52:56 CD2B.0   CD2B04000458  lying   86   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:52:56 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:52:57 CD2B.0   CD2B04000458  lying   91   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:57 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:58 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:58 CD2B.0   CD2B04000458  lying   92   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:58 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:59 CD2B.0   CD2B04000458  lying   91   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:52:59 333B.1   -             stand   101  InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:53:00 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:53:00 CD2B.0   CD2B04000458  lying   76   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:00 333B.1   -             stand   59   InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:01 CD2B.0   CD2B04000458  lying   69   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:01 333B.1   -             stand   103  InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:02 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:02 CD2B.0   CD2B04000458  lying   72   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:02 333B.1   -             stand   107  InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:03 CD2B.0   CD2B04000458  lying   66   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:03 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:04 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:04 CD2B.0   CD2B04000458  lying   85   InBed    lying              trk  1.00 Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:53:04 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.94  0.03  0.02  0.00  0.00
06:53:05 CD2B.0   CD2B04000458  lying   82   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:53:05 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:53:06 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:53:06 CD2B.0   CD2B04000458  lying   88   InBed    lying              trk  1.00 Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:53:06 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.93  0.04  0.02  0.00  0.00
06:53:07 CD2B.0   CD2B04000458  lying   72   InBed    lying              trk  1.00 Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:07 333B.1   -             stand   106  InBed    stand              room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:08 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.95  0.02  0.02  0.00  0.00
06:53:08 CD2B.0   CD2B04000458  walk    46   InBed    walk               trk  1.00 Bed        1   0     0.00  0.83  0.09  0.06  0.00  0.01
06:53:08 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.83  0.09  0.06  0.00  0.01
06:53:09 CD2B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.75  0.13  0.08  0.00  0.01
06:53:09 333B.1   -             stand   104  InBed    stand              room -    Bed        1   0     0.00  0.75  0.13  0.08  0.00  0.01
06:53:10 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.75  0.13  0.08  0.00  0.01
06:53:10 CD2B.0   CD2B04000458  walk    35   InBed    walk               trk  1.00 Bed        1   0     0.00  0.70  0.17  0.08  0.01  0.01
06:53:10 333B.1   -             stand   118  InBed    stand              room -    Bed        1   0     0.00  0.70  0.17  0.08  0.01  0.01
06:53:11 CD2B.0   CD2B04000458  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.64  0.21  0.08  0.01  0.01
06:53:11 333B.1   -             stand   115  InBed    stand              room -    Bed        1   0     0.00  0.64  0.21  0.08  0.01  0.01
06:53:12 1641.0   -             pad     -    InBed    pad InBed HR=None RR=None mv=1 turn=0 room -    Bed        1   0     0.00  0.64  0.21  0.08  0.01  0.01
06:53:12 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.61  0.22  0.08  0.01  0.01
06:53:12 333B.1   -             stand   108  InBed    stand              room -    Bed        1   0     0.00  0.61  0.22  0.08  0.01  0.01
06:53:13 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.60  0.23  0.08  0.01  0.01
06:53:13 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.60  0.23  0.08  0.01  0.01
06:53:14 1641.0   -             pad     -    InBed    pad InBed HR=None RR=17 mv=0 turn=1 room -    Bed        1   0     0.00  0.60  0.23  0.08  0.01  0.01
06:53:14 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.88  0.07  0.02  0.00  0.00
06:53:14 333B.1   -             stand   108  InBed    stand              room -    Bed        1   0     0.00  0.88  0.07  0.02  0.00  0.00
06:53:15 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.03  0.01  0.00  0.00
06:53:15 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.95  0.03  0.01  0.00  0.00
06:53:16 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:16 333B.1   -             stand   0    InBed    stand              room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:17 1641.0   -             pad     -    InBed    pad InBed HR=None RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:17 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:17 333B.1   -             stand   82   InBed    stand              room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:18 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:18 333B.1   -             stand   108  InBed    stand              room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:19 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:19 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:19 333B.1   -             stand   84   InBed    stand              room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:20 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:20 333B.1   -             stand   77   InBed    stand              room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:21 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:21 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:21 333B.1   -             walk    82   InBed    walk               room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:22 1641.0   -             pad     -    InBed    pad LeftBed HR=None RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
06:53:22 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.83  0.09  0.06  0.00  0.01
06:53:22 333B.1   -             walk    73   InBed    walk               room -    Bed        1   0     0.00  0.83  0.09  0.06  0.00  0.01
06:53:23 CD2B.0   CD2B04000458  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.75  0.13  0.07  0.00  0.01
06:53:23 333B.1   -             walk    0    InBed    walk               room -    Bed        1   0     0.00  0.75  0.13  0.07  0.00  0.01
06:53:23 1641.E   -             -       0    LeftBed  LeftBed(pad)       room -    OpenFloor  1   0     0.00  0.00  0.97  0.02  0.00  0.00
06:53:23 1641.E   -             -       0    LeftBed  LeftBed(pad)       room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:24 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:24 333B.1   -             walk    0    LeftBed  walk               room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:25 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:25 333B.1   -             stand   0    LeftBed  stand              room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:25 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:26 333B.1   -             stand   0    LeftBed  stand              room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:26 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:27 333B.1   -             stand   0    LeftBed  stand              room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:27 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:28 333B.1   -             stand   0    LeftBed  stand              room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:28 333B.E1  -             -       0    LeftBed  ExitRoom(rdr)      room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:28 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:29 333B.E   -             -       0    LeftBed  np=0  ★0           room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:29 333B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:29 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:30 333B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:30 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:31 333B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:31 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:32 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:33 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:34 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   0     0.00  0.00  0.98  0.00  0.00  0.00
06:53:35 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   27    0.00  0.00  0.98  0.00  0.00  0.00
06:53:36 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   28    0.00  0.00  0.98  0.00  0.00  0.00
06:53:37 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   29    0.00  0.00  0.98  0.00  0.00  0.00
06:53:38 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   30    0.00  0.00  0.98  0.00  0.00  0.00
06:53:39 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   31    0.00  0.00  0.98  0.00  0.00  0.00
06:53:40 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   32    0.00  0.00  0.98  0.00  0.00  0.00
06:53:41 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   33    0.00  0.00  0.98  0.00  0.00  0.00
06:53:42 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   34    0.00  0.00  0.98  0.00  0.00  0.00
06:53:43 CD2B.0   CD2B04000458  stand   0    LeftBed  stand              trk  1.00 OpenFloor  1   35    0.00  0.00  0.98  0.00  0.00  0.00
06:53:44 CD2B.E   -             -       0    LeftBed  np=0  ★0           room -    OpenFloor  1   35    0.00  0.00  0.98  0.00  0.00  0.00
06:53:44 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  1   36    0.00  0.00  0.98  0.00  0.00  0.00
06:53:45 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  0   0     0.01  0.00  0.70  0.00  0.00  0.07
06:53:46 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:47 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:48 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:49 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:50 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:51 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:52 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:53 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:54 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:55 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:56 333B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  0   0     0.01  0.00  0.51  0.01  0.06  0.06
06:53:56 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:53:57 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:53:58 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:53:59 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:00 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:01 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:02 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:03 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:04 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:05 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:06 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:07 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:08 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:09 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:10 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:11 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:12 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:13 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:14 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:15 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:16 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:17 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:18 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:19 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:20 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:21 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:22 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:23 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:24 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:25 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:26 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:27 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:28 333B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  0   0     0.02  0.00  0.39  0.02  0.11  0.05
06:54:28 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:29 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:30 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:31 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:32 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:33 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:34 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:35 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:36 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:37 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:38 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:39 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:40 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:41 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:42 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:43 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:44 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:45 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:46 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:47 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:48 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:49 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:50 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:51 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:52 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:53 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:54 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:55 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:56 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:57 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:58 -.-      -             -       -    LeftBed  (no frame, held)   room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:54:59 333B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  0   0     0.02  0.00  0.27  0.03  0.17  0.03
06:55:00 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:01 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:02 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:03 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:04 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:05 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:06 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:07 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:08 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:09 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:10 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:11 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:12 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:13 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:14 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:15 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:16 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:17 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:18 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:19 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:20 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:21 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:22 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:23 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:24 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:25 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:26 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:27 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:28 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:29 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:30 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:31 333B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.03  0.00  0.24  0.04  0.19  0.03
06:55:32 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:33 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:34 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:35 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:36 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:37 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:38 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:39 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:40 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:41 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:42 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:43 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:44 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:45 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:46 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:47 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:48 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:49 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:50 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:51 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:52 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:53 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:54 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:55 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:56 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:57 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:58 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:55:59 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:56:00 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:56:01 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:56:02 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:56:03 333B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.04  0.00  0.21  0.05  0.21  0.02
06:56:03 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:04 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:05 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:06 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:07 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:08 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:09 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:10 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:11 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:12 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:13 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:14 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:15 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:16 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:17 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:18 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:19 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:20 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:21 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:22 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:23 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:24 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:25 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:26 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:27 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:28 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:29 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:30 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:31 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:32 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:33 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:34 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:35 333B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.04  0.00  0.20  0.06  0.22  0.02
06:56:35 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:36 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:37 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:38 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:39 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:40 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:41 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:42 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:43 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:44 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:45 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:46 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:47 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:48 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:49 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:50 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:51 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:52 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:53 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:54 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:55 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:56 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:57 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:58 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:56:59 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:00 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:01 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:02 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:03 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:04 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:05 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:06 333B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:07 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:08 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:09 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:10 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:11 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:12 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:13 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:14 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:15 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:16 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:17 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:18 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:19 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:20 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:21 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:22 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:23 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:24 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:25 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:26 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:27 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:28 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:29 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:30 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:31 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:32 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:33 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:34 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:35 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:36 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:37 -.-      -             -       -    LeftBed  (no frame, held)   room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:38 333B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.05  0.00  0.19  0.06  0.23  0.02
06:57:39 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:39 333B.E   -             -       0    LeftBed  np=1               room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:39 333B.E   -             -       0    LeftBed  EnterRoom(rdr)     room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:39 333B.0   -             stand   88   LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:40 333B.0   -             walk    89   LeftBed  walk               room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:41 333B.0   -             walk    91   LeftBed  walk               room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:42 333B.0   -             walk    81   LeftBed  walk               room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:43 333B.0   -             walk    88   LeftBed  walk               room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:44 333B.0   -             walk    98   LeftBed  walk               room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:45 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:46 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:47 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:48 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:49 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:50 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:51 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:52 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:53 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:54 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:55 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:56 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:57 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:58 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:57:59 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:00 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:01 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:02 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:03 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:04 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:05 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:06 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:07 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:08 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:09 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:10 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.06  0.00  0.18  0.07  0.24  0.02
06:58:11 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:11 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:12 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:13 333B.0   -             stand   105  LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:14 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:15 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:16 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:17 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:18 333B.0   -             stand   106  LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:19 333B.0   -             stand   101  LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:20 333B.0   -             walk    104  LeftBed  walk               room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:21 333B.0   -             walk    91   LeftBed  walk               room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:22 333B.0   -             walk    103  LeftBed  walk               room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:23 333B.0   -             walk    94   LeftBed  walk               room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:24 333B.0   -             walk    94   LeftBed  walk               room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:25 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:26 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:27 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:28 333B.0   -             stand   92   LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:29 333B.0   -             stand   87   LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:30 333B.0   -             stand   90   LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:31 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:32 333B.0   -             stand   123  LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:33 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:34 333B.0   -             stand   113  LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:35 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:36 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:37 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:38 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:39 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:40 333B.0   -             stand   110  LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:41 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:42 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:42 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:42 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:43 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:44 333B.0   -             stand   114  LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:45 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:46 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:47 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:48 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:49 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:50 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:51 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:52 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:53 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:54 333B.0   -             stand   59   LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:55 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:56 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:57 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:58 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:58:59 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:00 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:01 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:02 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:03 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:04 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:05 333B.0   -             stand   66   LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:06 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:07 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:08 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:09 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:10 333B.0   -             stand   103  LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:11 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:12 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:13 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.07  0.00  0.18  0.07  0.24  0.02
06:59:14 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:14 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:15 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:16 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:17 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:18 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:19 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:20 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:21 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:22 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:23 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:24 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:25 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:26 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:27 333B.0   -             stand   91   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:28 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:29 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:30 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:31 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:32 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:33 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:34 333B.0   -             stand   90   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:35 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:36 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:37 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:38 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:39 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:40 333B.0   -             stand   103  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:41 333B.0   -             stand   86   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:42 333B.0   -             stand   90   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:43 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:44 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:45 333B.0   -             stand   82   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:45 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:46 333B.0   -             stand   112  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:47 333B.0   -             stand   83   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:48 333B.0   -             stand   95   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:49 333B.0   -             stand   110  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:50 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:51 333B.0   -             stand   94   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:52 333B.0   -             stand   85   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:53 333B.0   -             stand   97   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:54 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:55 333B.0   -             stand   105  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:56 333B.0   -             stand   105  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:57 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:58 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
06:59:59 333B.0   -             stand   106  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:00 333B.0   -             stand   114  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:01 333B.0   -             stand   122  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:02 333B.0   -             stand   87   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:03 333B.0   -             stand   99   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:04 333B.0   -             stand   101  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:05 333B.0   -             stand   111  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:06 333B.0   -             stand   98   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:07 333B.0   -             stand   95   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:08 333B.0   -             stand   99   LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:09 333B.0   -             stand   115  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:10 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:11 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:12 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:13 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:14 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:15 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:16 333B.0   -             stand   116  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:17 333B.0   -             stand   105  LeftBed  stand              room -    BlindOpen  0   0     0.08  0.00  0.18  0.07  0.25  0.02
07:00:17 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:18 333B.0   -             stand   113  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:19 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:20 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:21 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:22 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:23 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:24 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:25 333B.0   -             stand   110  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:26 333B.0   -             stand   85   LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:27 333B.0   -             stand   84   LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:28 333B.0   -             stand   115  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:29 333B.0   -             stand   129  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:30 333B.0   -             stand   115  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:31 333B.0   -             stand   120  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:32 333B.0   -             stand   141  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:33 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:34 333B.0   -             stand   138  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:35 333B.0   -             stand   153  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:36 333B.0   -             stand   122  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:37 333B.0   -             stand   93   LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:38 333B.0   -             stand   111  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:39 333B.0   -             stand   93   LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:40 333B.0   -             stand   105  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:41 333B.0   -             stand   124  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:42 333B.0   -             stand   126  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:43 333B.0   -             stand   96   LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:44 333B.0   -             stand   106  LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:45 333B.0   -             stand   96   LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:46 333B.0   -             stand   98   LeftBed  stand              room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:47 333B.0   -             walk    107  LeftBed  walk               room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:48 333B.0   -             walk    95   LeftBed  walk               room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:49 333B.0   -             walk    120  LeftBed  walk               room -    BlindOpen  0   0     0.09  0.00  0.17  0.07  0.25  0.02
07:00:49 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:50 333B.0   -             walk    99   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:51 333B.0   -             walk    91   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:51 333B.0   -             walk    73   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:52 333B.0   -             walk    76   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:53 333B.0   -             walk    83   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:54 333B.0   -             walk    88   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:55 333B.0   -             walk    102  LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:56 333B.0   -             walk    85   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:57 333B.0   -             walk    84   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:58 333B.0   -             walk    97   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:00:59 333B.0   -             walk    90   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:00 333B.0   -             walk    105  LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:01 333B.0   -             walk    94   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:02 333B.0   -             walk    94   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:03 333B.0   -             walk    103  LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:04 333B.0   -             walk    120  LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:05 333B.0   -             walk    122  LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:06 333B.0   -             walk    84   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:07 333B.0   -             walk    105  LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:08 333B.0   -             walk    118  LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:09 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:10 333B.0   -             stand   115  LeftBed  stand              room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:11 333B.0   -             stand   84   LeftBed  stand              room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:12 333B.0   -             stand   103  LeftBed  stand              room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:13 333B.0   -             walk    103  LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:14 333B.0   -             walk    104  LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:15 333B.0   -             walk    80   LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:16 333B.0   -             walk    0    LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:17 333B.0   -             walk    0    LeftBed  walk               room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:18 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:19 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:20 333B.0   -             stand   0    LeftBed  stand              room -    BlindOpen  0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:21 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:21 333B.0   -             stand   0    LeftBed  stand              room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:21 333B.E   -             -       0    LeftBed  ExitRoom(rdr)      room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:22 333B.E   -             -       0    LeftBed  np=0  ★0           room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:22 333B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:23 333B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:24 333B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:25 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:26 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:27 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:28 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:29 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:30 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:31 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:32 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:33 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:34 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:35 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:36 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:37 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:38 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:39 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:40 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:41 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:42 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:43 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:44 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:45 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:46 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:47 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:48 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:49 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:50 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:51 -.-      -             -       -    LeftBed  (no frame, held)   room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:52 333B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      0   0     0.10  0.00  0.17  0.07  0.25  0.02
07:01:52 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      0   0     0.11  0.00  0.17  0.07  0.25  0.02
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
06:40:10.995 333B.88   88     -    -      -      -     -    -    
06:40:42.913 333B.88   88     -    -      -      -     -    -    
06:41:14.432 333B.88   88     -    -      -      -     -    -    
06:41:46.292 333B.88   88     -    -      -      -     -    -    
06:42:18.150 333B.88   88     -    -      -      -     -    -    
06:42:49.660 333B.88   88     -    -      -      -     -    -    
06:43:21.741 333B.88   88     -    -      -      -     -    -    
06:43:53.146 333B.88   88     -    -      -      -     -    -    
06:44:25.330 333B.88   88     -    -      -      -     -    -    
06:44:56.633 333B.88   88     -    -      -      -     -    -    
06:45:28.618 333B.88   88     -    -      -      -     -    -    
06:46:00.181 333B.88   88     -    -      -      -     -    -    
06:46:32.200 333B.88   88     -    -      -      -     -    -    
06:47:03.842 333B.88   88     -    -      -      -     -    -    
06:47:35.526 333B.88   88     -    -      -      -     -    -    
06:48:07.102 333B.88   88     -    -      -      -     -    -    
06:48:39.177 333B.88   88     -    -      -      -     -    -    
06:48:42.982 333B.0    stand  1    -20    190    88    80        
06:48:43.806 333B.0    stand  1    -70    200    85    80   50   
06:48:44.808 333B.0    walk   1    -130   140    85    80   84   
06:48:45.808 333B.0    walk   1    -120   170    68    80   31   
06:48:46.808 333B.0    walk   1    -130   250    80    80   80   
06:48:47.812 333B.0    walk   1    -140   260    80    80   14   
06:48:48.812 333B.0    walk   1    -200   230    98    80   67   
06:48:49.817 333B.0    walk   1    -200   160    80    80   70   
06:48:50.813 333B.0    walk   1    -210   120    89    80   41   
06:48:51.819 333B.0    walk   1    -190   180    85    80   63   
06:48:52.822 333B.0    walk   1    -190   260    91    80   80   
06:48:53.821 333B.0    walk   1    -180   280    120   80   22   
06:48:54.717 333B.0    walk   1    -190   270    112   80   14   
06:48:55.731 333B.0    walk   1    -180   270    109   80   10   
06:48:56.720 333B.0    stand  1    -190   260    0     80   14   
06:48:57.720 333B.0    stand  1    -190   270    0     80   10   
06:48:58.734 333B.0    stand  1    -180   280    92    80   14   
06:48:59.722 333B.0    stand  1    -140   290    96    80   41   
06:49:00.733 333B.0    stand  1    -130   290    0     80   10   
06:49:01.726 333B.0    stand  1    -140   280    0     80   14   
06:49:02.724 333B.0    stand  1    -120   300    0     80   28   
06:49:03.728 333B.0    stand  1    -100   310    0     80   22   
06:49:04.726 333B.0    stand  1    -110   310    0     80   10   
06:49:05.628 333B.0    stand  1    -110   310    0     80   0    
06:49:06.627 333B.0    stand  1    -110   310    0     80   0    
06:49:07.628 333B.0    stand  1    -150   280    0     80   50   
06:49:08.630 333B.0    stand  1    -140   290    0     80   14   
06:49:09.632 333B.0    stand  1    -140   290    110   80   0    
06:49:10.636 333B.0    stand  1    -150   280    105   80   14   
06:49:11.632 333B.0    stand  1    -190   270    95    80   41   
06:49:12.682 333B.0    stand  1    -220   270    121   80   30   
06:49:13.635 333B.0    stand  1    -190   260    0     80   31   
06:49:14.568 333B.0    stand  1    -180   260    0     80   10   
06:49:15.556 333B.0    stand  1    -180   260    112   80   0    
06:49:16.558 333B.0    stand  1    -180   280    119   80   20   
06:49:17.560 333B.0    stand  1    -180   270    112   80   10   
06:49:18.560 333B.0    stand  1    -180   270    106   80   0    
06:49:19.564 333B.0    stand  1    -180   280    0     80   10   
06:49:20.566 333B.0    stand  1    -180   280    0     80   0    
06:49:21.564 333B.0    stand  1    -180   280    0     80   0    
06:49:22.567 333B.0    stand  1    -180   280    0     80   0    
06:49:23.568 333B.0    stand  1    -210   270    104   80   31   
06:49:24.568 333B.0    stand  1    -180   270    101   80   30   
06:49:25.569 333B.0    stand  1    -190   260    0     80   14   
06:49:26.461 333B.0    stand  1    -190   260    0     80   0    
06:49:27.466 333B.0    stand  1    -190   260    0     80   0    
06:49:28.460 333B.0    stand  1    -190   260    95    80   0    
06:49:29.465 333B.0    stand  1    -180   270    90    80   14   
06:49:30.506 333B.0    stand  1    -170   270    104   80   10   
06:49:31.509 333B.0    stand  1    -170   250    90    80   20   
06:49:32.509 333B.0    stand  1    -190   180    81    80   72   
06:49:33.422 333B.0    walk   1    -200   120    71    80   60   
06:49:34.418 333B.0    walk   1    -190   110    63    80   14   
06:49:35.460 333B.0    walk   1    -180   190    89    80   80   
06:49:36.409 333B.0    walk   1    -180   270    84    80   80   
06:49:37.412 333B.0    walk   1    -180   280    135   80   10   
06:49:38.414 333B.0    walk   1    -190   260    129   80   22   
06:49:39.413 333B.0    stand  1    -200   260    0     80   10   
06:49:40.414 333B.0    stand  1    -180   270    136   80   22   
06:49:41.414 333B.0    stand  1    -170   280    0     80   14   
06:49:42.417 333B.0    stand  1    -180   280    0     80   10   
06:49:43.425 333B.0    stand  1    -200   250    0     80   36   
06:49:44.326 333B.0    stand  1    -200   240    0     80   10   
06:49:45.318 333B.0    stand  1    -170   280    146   80   50   
06:49:46.318 333B.0    stand  1    -180   280    0     80   10   
06:49:47.325 333B.0    stand  1    -180   290    0     80   10   
06:49:48.328 333B.0    stand  1    -180   280    0     80   10   
06:49:49.329 333B.0    stand  1    -180   280    140   80   0    
06:49:50.331 333B.0    stand  1    -190   270    102   80   14   
06:49:51.331 333B.0    stand  1    -180   270    0     80   10   
06:49:52.332 333B.0    stand  1    -190   280    0     80   14   
06:49:53.340 333B.0    stand  1    -170   280    126   80   20   
06:49:54.331 333B.0    stand  1    -190   280    0     80   20   
06:49:55.228 333B.0    stand  1    -190   280    0     80   0    
06:49:56.232 333B.0    stand  1    -180   290    0     80   14   
06:49:57.232 333B.0    stand  1    -180   270    117   80   20   
06:49:58.242 333B.0    stand  1    -200   270    114   80   20   
06:49:59.232 333B.0    stand  1    -190   290    0     80   22   
06:50:00.235 333B.0    stand  1    -160   280    0     80   31   
06:50:01.236 333B.0    stand  1    -170   270    0     80   14   
06:50:02.240 333B.0    stand  1    -180   280    111   80   14   
06:50:03.170 333B.0    stand  1    -160   260    85    80   28   
06:50:04.168 333B.0    stand  1    -120   270    93    80   41   
06:50:05.181 333B.0    stand  1    -100   270    66    80   20   
06:50:06.185 333B.0    stand  1    -120   260    98    80   22   
06:50:07.172 333B.0    walk   1    -150   180    79    80   85   
06:50:08.181 333B.0    walk   1    -190   120    82    80   72   
06:50:09.182 333B.0    walk   1    -170   140    74    80   28   
06:50:10.177 333B.0    walk   1    -110   220    78    80   100  
06:50:11.177 333B.0    walk   1    -90    270    81    80   53   
06:50:12.232 333B.0    walk   1    -80    280    112   80   14   
06:50:13.183 333B.0    walk   1    -90    290    111   80   14   
06:50:14.081 333B.0    walk   1    -90    280    0     80   10   
06:50:15.088 333B.0    stand  1    -100   280    0     80   10   
06:50:16.076 333B.0    stand  1    -100   280    0     80   0    
06:50:17.077 333B.0    stand  1    -110   270    0     80   14   
06:50:18.078 333B.0    stand  1    -100   280    134   80   14   
06:50:19.040 333B.0    stand  1    -100   280    138   80   0    
06:50:20.037 333B.0    stand  1    -90    270    102   80   14   
06:50:21.041 333B.0    stand  1    -100   280    0     80   14   
06:50:22.057 333B.0    stand  1    -90    270    0     80   14   
06:50:23.044 333B.0    stand  1    -90    260    0     80   10   
06:50:24.041 333B.0    stand  1    -90    270    0     80   10   
06:50:25.040 333B.0    stand  1    -80    270    96    80   10   
06:50:26.043 333B.0    stand  1    -110   270    113   80   30   
06:50:27.044 333B.0    stand  1    -100   270    0     80   10   
06:50:28.052 333B.0    stand  1    -80    290    0     80   28   
06:50:29.052 333B.0    stand  1    -100   270    100   80   28   
06:50:30.049 333B.0    stand  1    -100   280    111   80   10   
06:50:30.941 333B.0    stand  1    -100   270    119   80   10   
06:50:31.941 333B.0    stand  1    -90    270    92    80   10   
06:50:32.945 333B.0    stand  1    -80    210    82    80   60   
06:50:33.944 333B.0    walk   1    -30    150    91    80   78   
06:50:34.952 333B.0    walk   1    -10    100    108   80   53   
06:50:35.946 333B.0    walk   1    -20    120    107   80   22   
06:50:36.949 333B.0    walk   1    -10    120    81    80   10   
06:50:37.949 333B.0    walk   1    -10    120    103   80   0    
06:50:38.948 333B.0    walk   1    -10    130    102   80   10   
06:50:39.952 333B.0    walk   1    -40    180    71    80   58   
06:50:40.957 333B.0    walk   1    -80    230    82    80   64   
06:50:41.956 333B.0    walk   1    -90    280    89    80   50   
06:50:42.845 333B.0    walk   1    -80    290    119   80   14   
06:50:43.856 333B.0    walk   1    -90    280    107   80   14   
06:50:44.851 333B.0    walk   1    -80    280    100   80   10   
06:50:45.856 333B.0    walk   1    -70    280    97    80   10   
06:50:46.850 333B.0    walk   1    -70    280    102   80   0    
06:50:47.855 333B.0    walk   1    -100   280    116   80   30   
06:50:48.856 333B.0    walk   1    -80    280    83    80   20   
06:50:49.862 333B.0    walk   1    -130   250    76    80   58   
06:50:50.787 333B.0    walk   1    -160   280    84    80   42   
06:50:51.785 333B.0    walk   1    -130   290    101   80   31   
06:50:52.786 333B.0    walk   1    -120   280    75    80   14   
06:50:53.791 333B.0    walk   1    -100   280    0     80   20   
06:50:54.789 333B.0    walk   1    -180   280    114   80   80   
06:50:55.800 333B.0    walk   1    -210   260    85    80   36   
06:50:56.795 333B.0    walk   1    -150   270    82    80   60   
06:50:57.792 333B.0    walk   1    -140   270    88    80   10   
06:50:58.791 333B.0    walk   1    -60    230    67    80   89   
06:50:59.795 333B.0    walk   1    0      190    93    80   72   
06:51:00.800 333B.0    walk   1    0      230    0     80   40   
06:51:01.693 333B.0    stand  1    0      230    0     80   0    
06:51:02.699 333B.0    stand  1    0      230    0     80   0    
06:51:03.693 333B.0    stand  1    0      230    0     80   0    
06:51:04.154 333B.0    stand  1    0      230    0     80   0    
06:51:04.154 333B.1    stand  2    -220   270    0     80   223  
06:51:04.709 333B.0    stand  1    0      230    0     80   223  
06:51:04.709 333B.1    stand  2    -200   240    79    80   200  
06:51:05.708 333B.0    stand  1    0      230    0     80   200  
06:51:05.708 333B.1    stand  2    -190   230    0     80   190  
06:51:06.715 333B.1    stand  2    -200   250    61    80   22   
06:51:07.664 333B.1    stand  2    -210   250    68    80   10   
06:51:08.671 333B.1    stand  2    -200   250    78    80   10   
06:51:09.674 333B.1    stand  2    -200   240    82    80   10   
06:51:10.690 333B.1    stand  2    -210   240    76    80   10   
06:51:11.718 333B.1    stand  2    -210   240    68    80   0    
06:51:12.668 333B.1    stand  2    -210   250    99    80   10   
06:51:13.676 333B.1    stand  2    -210   230    0     80   20   
06:51:14.684 333B.1    stand  2    -210   230    0     80   0    
06:51:15.669 333B.1    stand  2    -210   230    0     80   0    
06:51:16.673 333B.1    stand  2    -210   230    0     80   0    
06:51:17.564 333B.1    stand  2    -210   230    0     80   0    
06:51:18.576 333B.1    stand  2    -210   230    0     80   0    
06:51:19.568 333B.1    stand  2    -210   230    0     80   0    
06:51:20.566 333B.1    stand  2    -210   230    0     80   0    
06:51:21.568 333B.1    stand  2    -210   230    0     80   0    
06:51:22.575 333B.1    stand  2    -210   230    0     80   0    
06:51:23.574 333B.1    stand  2    -210   230    0     80   0    
06:51:24.576 333B.1    stand  2    -210   230    0     80   0    
06:51:25.577 333B.1    stand  2    -210   230    0     80   0    
06:51:26.576 333B.1    stand  2    -200   230    0     80   10   
06:51:27.582 333B.1    stand  2    -220   230    0     80   20   
06:51:28.483 333B.1    stand  2    -210   250    84    80   22   
06:51:29.478 333B.1    stand  2    -220   230    46    80   22   
06:51:30.477 333B.1    stand  2    -210   250    0     80   22   
06:51:31.488 333B.1    stand  2    -210   250    0     80   0    
06:51:32.480 333B.1    stand  2    -210   250    0     80   0    
06:51:33.485 333B.1    stand  2    -210   250    0     80   0    
06:51:34.482 333B.1    stand  2    -210   240    0     80   10   
06:51:35.487 333B.1    stand  2    -210   240    0     80   0    
06:51:36.492 333B.1    stand  2    -210   230    0     80   10   
06:51:37.484 333B.1    stand  2    -210   250    51    80   20   
06:51:38.504 333B.1    stand  2    -210   230    0     80   20   
06:51:39.494 333B.1    stand  2    -200   230    0     80   10   
06:51:40.383 333B.1    stand  2    -200   220    0     80   10   
06:51:41.382 333B.1    stand  2    -200   230    57    80   10   
06:51:42.398 333B.1    stand  2    -210   240    77    80   14   
06:51:43.384 333B.1    stand  2    -200   240    0     80   10   
06:51:44.415 333B.1    stand  2    -210   240    0     80   10   
06:51:45.388 333B.1    stand  2    -210   240    0     80   0    
06:51:46.387 333B.1    stand  2    -210   240    0     80   0    
06:51:47.388 333B.1    stand  2    -210   240    0     80   0    
06:51:48.389 333B.1    stand  2    -210   240    0     80   0    
06:51:49.395 333B.1    stand  2    -210   240    0     80   0    
06:51:50.392 333B.1    stand  2    -210   240    0     80   0    
06:51:51.392 333B.1    stand  2    -210   230    72    80   10   
06:51:52.285 333B.1    stand  2    -200   230    75    80   10   
06:51:53.305 333B.1    stand  2    -200   230    0     80   0    
06:51:54.294 333B.1    stand  2    -210   230    0     80   10   
06:51:55.298 333B.1    stand  2    -200   240    46    80   14   
06:51:56.289 333B.1    stand  2    -200   240    66    80   0    
06:51:57.296 333B.1    stand  2    -210   240    70    80   10   
06:51:58.293 333B.1    stand  2    -220   240    85    80   10   
06:51:59.292 333B.1    stand  2    -210   230    85    80   14   
06:52:00.292 333B.1    stand  2    -210   250    0     80   20   
06:52:01.294 333B.1    stand  2    -210   230    0     80   20   
06:52:02.303 333B.1    stand  2    -210   230    0     80   0    
06:52:03.206 333B.1    stand  2    -210   230    0     80   0    
06:52:04.196 333B.1    stand  2    -210   240    103   80   10   
06:52:05.208 333B.1    stand  2    -210   240    0     80   0    
06:52:06.200 333B.1    stand  2    -210   240    0     80   0    
06:52:07.200 333B.1    stand  2    -210   240    0     80   0    
06:52:08.200 333B.1    stand  2    -210   240    0     80   0    
06:52:09.204 333B.1    stand  2    -220   230    0     80   14   
06:52:10.201 333B.1    stand  2    -220   230    0     80   0    
06:52:11.253 333B.1    stand  2    -220   230    0     80   0    
06:52:12.206 333B.1    stand  2    -200   230    0     80   20   
06:52:13.216 333B.1    stand  2    -190   240    0     80   14   
06:52:14.109 333B.1    stand  2    -190   240    0     80   0    
06:52:15.110 333B.1    stand  2    -200   240    0     80   10   
06:52:16.108 333B.1    stand  2    -210   240    95    80   10   
06:52:17.113 333B.1    stand  2    -210   230    0     80   10   
06:52:18.113 333B.1    stand  2    -210   230    0     80   0    
06:52:19.113 333B.1    stand  2    -210   240    98    80   10   
06:52:20.114 333B.1    stand  2    -210   230    0     80   10   
06:52:21.114 333B.1    stand  2    -220   250    84    80   22   
06:52:22.116 333B.1    stand  2    -200   270    93    80   28   
06:52:23.116 333B.1    stand  2    -220   260    0     80   22   
06:52:24.116 333B.1    stand  2    -210   260    80    80   10   
06:52:25.117 333B.1    stand  2    -190   270    90    80   22   
06:52:26.015 333B.1    stand  2    -200   250    69    80   22   
06:52:27.020 333B.1    stand  2    -190   250    87    80   10   
06:52:28.021 333B.1    stand  2    -190   240    97    80   10   
06:52:29.025 333B.1    stand  2    -190   240    114   80   0    
06:52:30.020 333B.1    stand  2    -190   240    0     80   0    
06:52:31.021 333B.1    stand  2    -200   240    0     80   10   
06:52:32.028 333B.1    stand  2    -210   230    0     80   14   
06:52:33.021 333B.1    stand  2    -190   250    113   80   28   
06:52:34.024 333B.1    stand  2    -220   240    0     80   31   
06:52:35.024 333B.1    stand  2    -200   260    77    80   28   
06:52:36.024 333B.1    stand  2    -200   270    72    80   10   
06:52:36.928 333B.1    stand  2    -210   280    110   80   14   
06:52:37.938 333B.1    stand  2    -210   250    0     80   30   
06:52:38.929 333B.1    stand  2    -210   270    82    80   20   
06:52:39.944 333B.1    stand  2    -200   270    80    80   10   
06:52:40.944 333B.1    stand  2    -210   280    103   80   14   
06:52:41.940 333B.1    stand  2    -220   270    108   80   14   
06:52:42.934 333B.1    stand  2    -200   270    0     80   20   
06:52:43.936 333B.1    stand  2    -190   250    0     80   22   
06:52:44.939 333B.1    stand  2    -180   270    110   80   22   
06:52:45.938 333B.1    stand  2    -210   280    116   80   31   
06:52:46.984 333B.1    stand  2    -190   290    0     80   22   
06:52:47.936 333B.1    stand  2    -190   280    0     80   10   
06:52:48.832 333B.1    stand  2    -210   280    121   80   20   
06:52:49.832 333B.1    stand  2    -220   280    95    80   10   
06:52:50.836 333B.1    stand  2    -170   270    90    80   50   
06:52:51.830 333B.1    stand  2    -170   280    0     80   10   
06:52:52.840 333B.1    stand  2    -160   250    0     80   31   
06:52:53.834 333B.1    stand  2    -150   260    0     80   14   
06:52:54.836 333B.1    stand  2    -150   270    0     80   10   
06:52:55.840 333B.1    stand  2    -180   290    0     80   36   
06:52:56.838 333B.1    stand  2    -200   280    0     80   22   
06:52:57.841 333B.1    stand  2    -160   270    0     80   41   
06:52:58.840 333B.1    stand  2    -170   300    0     80   31   
06:52:59.848 333B.1    stand  2    -180   270    101   80   31   
06:53:00.733 333B.1    stand  2    -200   240    59    80   36   
06:53:01.734 333B.1    stand  2    -200   250    103   80   10   
06:53:02.735 333B.1    stand  2    -190   240    107   80   14   
06:53:03.740 333B.1    stand  2    -200   240    0     80   10   
06:53:04.737 333B.1    stand  2    -200   240    0     80   0    
06:53:05.744 333B.1    stand  2    -200   240    0     80   0    
06:53:06.743 333B.1    stand  2    -200   240    0     80   0    
06:53:07.740 333B.1    stand  2    -200   260    106   80   20   
06:53:08.745 333B.1    stand  2    -200   240    0     80   20   
06:53:09.744 333B.1    stand  2    -200   240    104   80   0    
06:53:10.808 333B.1    stand  2    -200   230    118   80   10   
06:53:11.744 333B.1    stand  2    -210   250    115   80   22   
06:53:12.646 333B.1    stand  2    -200   260    108   80   14   
06:53:13.637 333B.1    stand  2    -200   240    0     80   20   
06:53:14.645 333B.1    stand  2    -200   230    108   80   10   
06:53:15.648 333B.1    stand  2    -200   270    0     80   40   
06:53:16.648 333B.1    stand  2    -200   270    0     80   0    
06:53:17.648 333B.1    stand  2    -210   260    82    80   14   
06:53:18.651 333B.1    stand  2    -220   280    108   80   22   
06:53:19.658 333B.1    stand  2    -220   270    84    80   10   
06:53:20.678 333B.1    stand  2    -200   270    77    80   20   
06:53:21.656 333B.1    walk   2    -110   220    82    80   102  
06:53:22.653 333B.1    walk   2    0      190    73    80   114  
06:53:23.546 333B.1    walk   2    0      190    0     80   0    
06:53:24.548 333B.1    walk   2    0      190    0     80   0    
06:53:25.553 333B.1    stand  2    0      190    0     80   0    
06:53:26.552 333B.1    stand  2    0      190    0     80   0    
06:53:27.556 333B.1    stand  2    0      190    0     80   0    
06:53:28.552 333B.1    stand  1    0      190    0     80   0    
06:53:29.604 333B.88   88     -    -      -      -     -    -    
06:53:30.578 333B.88   88     -    -      -      -     -    -    
06:53:31.564 333B.88   88     -    -      -      -     -    -    
06:53:56.366 333B.88   88     -    -      -      -     -    -    
06:54:28.364 333B.88   88     -    -      -      -     -    -    
06:54:59.768 333B.88   88     -    -      -      -     -    -    
06:55:31.961 333B.88   88     -    -      -      -     -    -    
06:56:03.262 333B.88   88     -    -      -      -     -    -    
06:56:35.269 333B.88   88     -    -      -      -     -    -    
06:57:06.752 333B.88   88     -    -      -      -     -    -    
06:57:38.837 333B.88   88     -    -      -      -     -    -    
06:57:39.756 333B.0    stand  1    -70    140    88    80   86   
06:57:40.502 333B.0    walk   1    -140   140    89    80   70   
06:57:41.514 333B.0    walk   1    -220   170    91    80   85   
06:57:42.504 333B.0    walk   1    -270   190    81    80   53   
06:57:43.506 333B.0    walk   1    -280   220    88    80   31   
06:57:44.510 333B.0    walk   1    -300   240    98    80   28   
06:57:45.508 333B.0    stand  1    -300   250    0     80   10   
06:57:46.510 333B.0    stand  1    -300   250    0     80   0    
06:57:47.512 333B.0    stand  1    -290   250    0     80   10   
06:57:48.410 333B.0    stand  1    -290   250    0     80   0    
06:57:49.416 333B.0    stand  1    -290   250    0     80   0    
06:57:50.422 333B.0    stand  1    -290   250    0     80   0    
06:57:51.414 333B.0    stand  1    -290   250    0     80   0    
06:57:52.414 333B.0    stand  1    -290   250    0     80   0    
06:57:53.416 333B.0    stand  1    -290   250    0     80   0    
06:57:54.424 333B.0    stand  1    -290   250    0     80   0    
06:57:55.417 333B.0    stand  1    -290   250    0     80   0    
06:57:56.420 333B.0    stand  1    -290   250    0     80   0    
06:57:57.421 333B.0    stand  1    -290   250    0     80   0    
06:57:58.421 333B.0    stand  1    -290   250    0     80   0    
06:57:59.421 333B.0    stand  1    -290   250    0     80   0    
06:58:00.314 333B.0    stand  1    -290   250    0     80   0    
06:58:01.316 333B.0    stand  1    -290   250    0     80   0    
06:58:02.318 333B.0    stand  1    -290   250    0     80   0    
06:58:03.323 333B.0    stand  1    -290   250    0     80   0    
06:58:04.321 333B.0    stand  1    -290   250    0     80   0    
06:58:05.323 333B.0    stand  1    -290   250    0     80   0    
06:58:06.334 333B.0    stand  1    -290   250    0     80   0    
06:58:07.322 333B.0    stand  1    -290   250    0     80   0    
06:58:08.388 333B.0    stand  1    -290   250    0     80   0    
06:58:09.324 333B.0    stand  1    -290   250    0     80   0    
06:58:10.326 333B.0    stand  1    -290   250    0     80   0    
06:58:11.330 333B.0    stand  1    -290   250    0     80   0    
06:58:12.218 333B.0    stand  1    -290   250    0     80   0    
06:58:13.225 333B.0    stand  1    -310   230    105   80   28   
06:58:14.220 333B.0    stand  1    -330   230    0     80   20   
06:58:15.226 333B.0    stand  1    -350   220    0     80   22   
06:58:16.229 333B.0    stand  1    -320   230    0     80   31   
06:58:17.224 333B.0    stand  1    -290   240    0     80   31   
06:58:18.228 333B.0    stand  1    -290   210    106   80   30   
06:58:19.258 333B.0    stand  1    -240   190    101   80   53   
06:58:20.260 333B.0    walk   1    -170   190    104   80   70   
06:58:21.149 333B.0    walk   1    -120   250    91    80   78   
06:58:22.152 333B.0    walk   1    -100   280    103   80   36   
06:58:23.157 333B.0    walk   1    -90    270    94    80   14   
06:58:24.153 333B.0    walk   1    -70    280    94    80   22   
06:58:25.156 333B.0    stand  1    -80    260    0     80   22   
06:58:26.156 333B.0    stand  1    -60    270    0     80   22   
06:58:27.157 333B.0    stand  1    -100   280    0     80   41   
06:58:28.160 333B.0    stand  1    -100   280    92    80   0    
06:58:29.166 333B.0    stand  1    -100   260    87    80   20   
06:58:30.170 333B.0    stand  1    -80    290    90    80   36   
06:58:31.160 333B.0    stand  1    -100   290    0     80   20   
06:58:32.059 333B.0    stand  1    -70    280    123   80   31   
06:58:33.062 333B.0    stand  1    -90    270    0     80   22   
06:58:34.062 333B.0    stand  1    -60    300    113   80   42   
06:58:35.068 333B.0    stand  1    -60    300    0     80   0    
06:58:36.068 333B.0    stand  1    -80    300    0     80   20   
06:58:37.072 333B.0    stand  1    -90    290    0     80   14   
06:58:38.068 333B.0    stand  1    -90    260    0     80   30   
06:58:39.072 333B.0    stand  1    -100   280    0     80   22   
06:58:40.069 333B.0    stand  1    -90    270    110   80   14   
06:58:41.069 333B.0    stand  1    -90    270    0     80   0    
06:58:42.072 333B.0    stand  1    -70    280    0     80   22   
06:58:42.978 333B.0    stand  1    -80    270    0     80   14   
06:58:43.978 333B.0    stand  1    -80    270    0     80   0    
06:58:44.972 333B.0    stand  1    -100   270    114   80   20   
06:58:45.976 333B.0    stand  1    -110   290    0     80   22   
06:58:46.985 333B.0    stand  1    -100   280    0     80   14   
06:58:47.974 333B.0    stand  1    -90    280    0     80   10   
06:58:48.977 333B.0    stand  1    -100   290    0     80   14   
06:58:49.976 333B.0    stand  1    -100   290    0     80   0    
06:58:50.906 333B.0    stand  1    -100   280    0     80   10   
06:58:51.909 333B.0    stand  1    -100   270    0     80   10   
06:58:52.909 333B.0    stand  1    -110   290    0     80   22   
06:58:53.912 333B.0    stand  1    -80    260    0     80   42   
06:58:54.916 333B.0    stand  1    -80    270    59    80   10   
06:58:55.914 333B.0    stand  1    -70    250    0     80   22   
06:58:56.919 333B.0    stand  1    -90    270    0     80   28   
06:58:57.916 333B.0    stand  1    -100   300    0     80   31   
06:58:58.916 333B.0    stand  1    -100   300    0     80   0    
06:58:59.917 333B.0    stand  1    -100   300    0     80   0    
06:59:00.918 333B.0    stand  1    -100   300    0     80   0    
06:59:01.934 333B.0    stand  1    -100   300    0     80   0    
06:59:02.818 333B.0    stand  1    -100   300    0     80   0    
06:59:03.814 333B.0    stand  1    -100   290    0     80   10   
06:59:04.812 333B.0    stand  1    -90    270    0     80   22   
06:59:05.826 333B.0    stand  1    -80    260    66    80   14   
06:59:06.782 333B.0    stand  1    -100   260    0     80   20   
06:59:07.781 333B.0    stand  1    -90    280    0     80   22   
06:59:08.837 333B.0    stand  1    -90    270    0     80   10   
06:59:09.782 333B.0    stand  1    -100   280    0     80   14   
06:59:10.786 333B.0    stand  1    -80    290    103   80   22   
06:59:11.785 333B.0    stand  1    -100   260    0     80   36   
06:59:12.792 333B.0    stand  1    -90    280    0     80   22   
06:59:13.788 333B.0    stand  1    -70    280    0     80   20   
06:59:14.788 333B.0    stand  1    -70    280    0     80   0    
06:59:15.790 333B.0    stand  1    -90    270    0     80   22   
06:59:16.794 333B.0    stand  1    -110   270    0     80   20   
06:59:17.792 333B.0    stand  1    -110   290    0     80   20   
06:59:18.689 333B.0    stand  1    -100   280    0     80   14   
06:59:19.693 333B.0    stand  1    -90    270    0     80   14   
06:59:20.685 333B.0    stand  1    -90    270    0     80   0    
06:59:21.694 333B.0    stand  1    -90    270    0     80   0    
06:59:22.688 333B.0    stand  1    -110   280    0     80   22   
06:59:23.708 333B.0    stand  1    -120   310    0     80   31   
06:59:24.704 333B.0    stand  1    -110   300    0     80   14   
06:59:25.709 333B.0    stand  1    -90    280    0     80   28   
06:59:26.708 333B.0    stand  1    -100   290    0     80   14   
06:59:27.710 333B.0    stand  1    -80    260    91    80   36   
06:59:28.608 333B.0    stand  1    -100   300    0     80   44   
06:59:29.608 333B.0    stand  1    -90    310    0     80   14   
06:59:30.616 333B.0    stand  1    -80    300    0     80   14   
06:59:31.612 333B.0    stand  1    -80    290    0     80   10   
06:59:32.618 333B.0    stand  1    -90    270    0     80   22   
06:59:33.608 333B.0    stand  1    -90    260    0     80   10   
06:59:34.608 333B.0    stand  1    -100   260    90    80   10   
06:59:35.610 333B.0    stand  1    -100   260    0     80   0    
06:59:36.612 333B.0    stand  1    -90    260    0     80   10   
06:59:37.613 333B.0    stand  1    -90    260    0     80   0    
06:59:38.613 333B.0    stand  1    -90    260    0     80   0    
06:59:39.517 333B.0    stand  1    -50    300    0     80   56   
06:59:40.516 333B.0    stand  1    -70    280    103   80   28   
06:59:41.537 333B.0    stand  1    -60    290    86    80   14   
06:59:42.520 333B.0    stand  1    -60    290    90    80   0    
06:59:43.528 333B.0    stand  1    -60    260    0     80   30   
06:59:44.520 333B.0    stand  1    -50    260    0     80   10   
06:59:45.520 333B.0    stand  1    -90    270    82    80   41   
06:59:46.523 333B.0    stand  1    -90    270    112   80   0    
06:59:47.522 333B.0    stand  1    -100   270    83    80   10   
06:59:48.528 333B.0    stand  1    -110   260    95    80   14   
06:59:49.526 333B.0    stand  1    -110   260    110   80   0    
06:59:50.526 333B.0    stand  1    -100   240    0     80   22   
06:59:51.420 333B.0    stand  1    -100   240    94    80   0    
06:59:52.422 333B.0    stand  1    -100   240    85    80   0    
06:59:53.424 333B.0    stand  1    -100   250    97    80   10   
06:59:54.428 333B.0    stand  1    -110   280    0     80   31   
06:59:55.390 333B.0    stand  1    -110   270    105   80   10   
06:59:56.398 333B.0    stand  1    -110   260    105   80   10   
06:59:57.398 333B.0    stand  1    -100   250    0     80   14   
06:59:58.404 333B.0    stand  1    -100   240    0     80   10   
06:59:59.395 333B.0    stand  1    -110   260    106   80   22   
07:00:00.393 333B.0    stand  1    -120   260    114   80   10   
07:00:01.394 333B.0    stand  1    -110   260    122   80   10   
07:00:02.399 333B.0    stand  1    -110   260    87    80   0    
07:00:03.397 333B.0    stand  1    -110   260    99    80   0    
07:00:04.403 333B.0    stand  1    -100   270    101   80   14   
07:00:05.405 333B.0    stand  1    -120   250    111   80   28   
07:00:06.400 333B.0    stand  1    -110   250    98    80   10   
07:00:07.292 333B.0    stand  1    -100   250    95    80   10   
07:00:08.356 333B.0    stand  1    -110   260    99    80   14   
07:00:09.300 333B.0    stand  1    -110   250    115   80   10   
07:00:10.308 333B.0    stand  1    -110   260    0     80   10   
07:00:11.324 333B.0    stand  1    -100   250    0     80   14   
07:00:12.328 333B.0    stand  1    -120   250    0     80   20   
07:00:13.327 333B.0    stand  1    -110   250    0     80   10   
07:00:14.338 333B.0    stand  1    -100   260    0     80   14   
07:00:15.335 333B.0    stand  1    -110   270    0     80   14   
07:00:16.221 333B.0    stand  1    -100   270    116   80   10   
07:00:17.224 333B.0    stand  1    -100   250    105   80   20   
07:00:18.227 333B.0    stand  1    -110   250    113   80   10   
07:00:19.222 333B.0    stand  1    -120   270    0     80   22   
07:00:20.248 333B.0    stand  1    -90    250    0     80   36   
07:00:21.231 333B.0    stand  1    -90    250    0     80   0    
07:00:22.238 333B.0    stand  1    -100   240    0     80   14   
07:00:23.233 333B.0    stand  1    -110   250    0     80   14   
07:00:24.231 333B.0    stand  1    -110   260    0     80   10   
07:00:25.236 333B.0    stand  1    -110   250    110   80   10   
07:00:26.229 333B.0    stand  1    -110   260    85    80   10   
07:00:27.238 333B.0    stand  1    -100   240    84    80   22   
07:00:28.128 333B.0    stand  1    -90    240    115   80   10   
07:00:29.127 333B.0    stand  1    -80    240    129   80   10   
07:00:30.129 333B.0    stand  1    -90    230    115   80   14   
07:00:31.127 333B.0    stand  1    -90    220    120   80   10   
07:00:32.134 333B.0    stand  1    -90    220    141   80   0    
07:00:33.137 333B.0    stand  1    -80    230    0     80   14   
07:00:34.132 333B.0    stand  1    -90    230    138   80   10   
07:00:35.137 333B.0    stand  1    -90    250    153   80   20   
07:00:36.636 333B.0    stand  1    -90    250    122   80   0    
07:00:37.135 333B.0    stand  1    -100   240    93    80   14   
07:00:38.137 333B.0    stand  1    -100   220    111   80   20   
07:00:39.137 333B.0    stand  1    -110   260    93    80   41   
07:00:40.037 333B.0    stand  1    -100   260    105   80   10   
07:00:41.040 333B.0    stand  1    -100   260    124   80   0    
07:00:42.032 333B.0    stand  1    -110   260    126   80   10   
07:00:43.037 333B.0    stand  1    -100   250    96    80   14   
07:00:44.032 333B.0    stand  1    -110   250    106   80   10   
07:00:45.040 333B.0    stand  1    -110   240    96    80   10   
07:00:46.041 333B.0    stand  1    -100   180    98    80   60   
07:00:47.036 333B.0    walk   1    -80    120    107   80   63   
07:00:48.039 333B.0    walk   1    -60    80     95    80   44   
07:00:49.041 333B.0    walk   1    -30    20     120   80   67   
07:00:50.049 333B.0    walk   1    -30    30     99    80   10   
07:00:51.041 333B.0    walk   1    -70    110    91    80   89   
07:00:51.934 333B.0    walk   1    -90    190    73    80   82   
07:00:52.938 333B.0    walk   1    -110   260    76    80   72   
07:00:53.936 333B.0    walk   1    -110   260    83    80   0    
07:00:54.948 333B.0    walk   1    -130   250    88    80   22   
07:00:55.938 333B.0    walk   1    -140   260    102   80   14   
07:00:56.948 333B.0    walk   1    -190   200    85    80   78   
07:00:57.942 333B.0    walk   1    -260   180    84    80   72   
07:00:58.953 333B.0    walk   1    -290   210    97    80   42   
07:00:59.965 333B.0    walk   1    -280   210    90    80   10   
07:01:00.977 333B.0    walk   1    -290   200    105   80   14   
07:01:01.857 333B.0    walk   1    -260   190    94    80   31   
07:01:02.857 333B.0    walk   1    -200   230    94    80   72   
07:01:03.864 333B.0    walk   1    -140   260    103   80   67   
07:01:04.868 333B.0    walk   1    -90    240    120   80   53   
07:01:05.858 333B.0    walk   1    -80    260    122   80   22   
07:01:06.908 333B.0    walk   1    -90    260    84    80   10   
07:01:07.860 333B.0    walk   1    -90    280    105   80   20   
07:01:08.862 333B.0    walk   1    -90    270    118   80   10   
07:01:09.864 333B.0    stand  1    -80    290    0     80   22   
07:01:10.866 333B.0    stand  1    -100   280    115   80   22   
07:01:11.868 333B.0    stand  1    -100   270    84    80   10   
07:01:12.769 333B.0    stand  1    -100   270    103   80   0    
07:01:13.774 333B.0    walk   1    -110   230    103   80   41   
07:01:14.768 333B.0    walk   1    -70    210    104   80   44   
07:01:15.776 333B.0    walk   1    -20    200    80    80   50   
07:01:16.780 333B.0    walk   1    0      190    0     80   22   
07:01:17.784 333B.0    walk   1    0      190    0     80   0    
07:01:18.780 333B.0    stand  1    0      190    0     80   0    
07:01:19.784 333B.0    stand  1    0      200    0     80   10   
07:01:20.796 333B.0    stand  1    0      200    0     80   0    
07:01:21.788 333B.0    stand  1    0      200    0     80   0    
07:01:22.736 333B.88   88     -    -      -      -     -    -    
07:01:23.692 333B.88   88     -    -      -      -     -    -    
07:01:24.692 333B.88   88     -    -      -      -     -    -    
07:01:52.440 333B.88   88     -    -      -      -     -    -    

06:40:00.458 CD2B.0    stand  1    -200   160    0     80        
06:40:01.460 CD2B.0    stand  1    -200   160    0     80   0    
06:40:02.463 CD2B.0    stand  1    -200   160    0     80   0    
06:40:03.466 CD2B.0    stand  1    -200   160    0     80   0    
06:40:04.468 CD2B.0    stand  1    -200   160    0     80   0    
06:40:05.481 CD2B.0    stand  1    -200   160    0     80   0    
06:40:06.475 CD2B.0    stand  1    -200   160    0     80   0    
06:40:07.360 CD2B.0    stand  1    -200   160    0     80   0    
06:40:08.363 CD2B.0    stand  1    -200   160    0     80   0    
06:40:09.366 CD2B.0    stand  1    -200   160    0     80   0    
06:40:10.367 CD2B.0    stand  1    -200   160    0     80   0    
06:40:11.329 CD2B.0    stand  1    -200   160    0     80   0    
06:40:12.339 CD2B.0    stand  1    -200   160    0     80   0    
06:40:13.334 CD2B.0    stand  1    -200   160    0     80   0    
06:40:14.334 CD2B.0    stand  1    -200   160    0     80   0    
06:40:15.340 CD2B.0    stand  1    -200   160    0     80   0    
06:40:16.334 CD2B.0    stand  1    -200   160    0     80   0    
06:40:17.391 CD2B.0    stand  1    -200   160    0     80   0    
06:40:18.335 CD2B.0    stand  1    -200   160    0     80   0    
06:40:19.335 CD2B.0    stand  1    -200   160    0     80   0    
06:40:20.340 CD2B.0    stand  1    -200   160    0     80   0    
06:40:21.337 CD2B.0    stand  1    -200   160    0     80   0    
06:40:22.338 CD2B.0    stand  1    -200   160    0     80   0    
06:40:23.231 CD2B.0    stand  1    -200   160    0     80   0    
06:40:24.233 CD2B.0    stand  1    -200   160    0     80   0    
06:40:25.238 CD2B.0    stand  1    -200   160    0     80   0    
06:40:26.242 CD2B.0    stand  1    -200   160    0     80   0    
06:40:27.284 CD2B.0    stand  1    -200   160    0     80   0    
06:40:28.286 CD2B.0    lying  1    -170   170    62    80   31   
06:40:29.282 CD2B.0    lying  1    -140   190    65    80   36   
06:40:30.183 CD2B.0    lying  1    -140   190    70    80   0    
06:40:31.180 CD2B.0    lying  1    -110   200    70    80   31   
06:40:32.184 CD2B.0    lying  1    -120   200    74    80   10   
06:40:33.189 CD2B.0    lying  1    -110   200    75    80   10   
06:40:34.183 CD2B.0    lying  1    -90    200    73    80   20   
06:40:35.183 CD2B.0    lying  1    -70    210    0     80   22   
06:40:36.947 CD2B.0    lying  1    -80    200    0     80   14   
06:40:37.195 CD2B.0    lying  1    -80    190    0     80   10   
06:40:38.190 CD2B.0    lying  1    -80    200    0     80   10   
06:40:39.186 CD2B.0    lying  1    -80    200    0     80   0    
06:40:40.187 CD2B.0    lying  1    -80    200    0     80   0    
06:40:41.089 CD2B.0    lying  1    -80    200    0     80   0    
06:40:42.093 CD2B.0    lying  1    -80    200    0     80   0    
06:40:43.089 CD2B.0    lying  1    -80    200    0     80   0    
06:40:44.098 CD2B.0    lying  1    -80    200    0     80   0    
06:40:45.092 CD2B.0    lying  1    -80    200    0     80   0    
06:40:46.093 CD2B.0    lying  1    -80    200    0     80   0    
06:40:47.095 CD2B.0    lying  1    -80    200    0     80   0    
06:40:48.097 CD2B.0    lying  1    -80    200    0     80   0    
06:40:49.096 CD2B.0    lying  1    -80    200    0     80   0    
06:40:50.097 CD2B.0    lying  1    -80    200    0     80   0    
06:40:51.104 CD2B.0    stand  1    -220   160    74    80   145  
06:40:52.101 CD2B.0    stand  1    -210   170    0     80   14   
06:40:52.993 CD2B.0    stand  1    -220   150    0     80   22   
06:40:53.992 CD2B.0    stand  1    -220   150    0     80   0    
06:40:54.999 CD2B.0    stand  1    -220   160    0     80   10   
06:40:55.996 CD2B.0    stand  1    -220   160    0     80   0    
06:40:56.996 CD2B.0    stand  1    -220   160    0     80   0    
06:40:57.997 CD2B.0    stand  1    -220   160    0     80   0    
06:40:59.008 CD2B.0    stand  1    -220   160    0     80   0    
06:41:00.012 CD2B.0    stand  1    -220   160    0     80   0    
06:41:01.000 CD2B.0    stand  1    -220   160    0     80   0    
06:41:02.000 CD2B.0    stand  1    -220   160    0     80   0    
06:41:03.007 CD2B.0    stand  1    -220   160    0     80   0    
06:41:04.009 CD2B.0    stand  1    -220   160    0     80   0    
06:41:04.903 CD2B.0    stand  1    -220   160    0     80   0    
06:41:05.903 CD2B.0    stand  1    -220   160    0     80   0    
06:41:06.898 CD2B.0    stand  1    -220   160    0     80   0    
06:41:07.898 CD2B.0    stand  1    -220   160    0     80   0    
06:41:08.900 CD2B.0    stand  1    -210   160    0     80   10   
06:41:09.900 CD2B.0    walk   1    -110   190    74    80   104  
06:41:10.909 CD2B.0    lying  1    -100   200    76    80   14   
06:41:11.908 CD2B.0    lying  1    -80    200    61    80   20   
06:41:12.958 CD2B.0    lying  1    -100   190    0     80   22   
06:41:13.912 CD2B.0    lying  1    -100   190    0     80   0    
06:41:14.920 CD2B.0    lying  1    -100   190    0     80   0    
06:41:15.806 CD2B.0    lying  1    -100   190    0     80   0    
06:41:16.808 CD2B.0    lying  1    -100   190    0     80   0    
06:41:17.876 CD2B.0    lying  1    -100   190    0     80   0    
06:41:18.809 CD2B.0    lying  1    -110   190    0     80   10   
06:41:19.817 CD2B.0    lying  1    -110   190    0     80   0    
06:41:20.818 CD2B.0    lying  1    -110   190    0     80   0    
06:41:21.819 CD2B.0    lying  1    -110   190    0     80   0    
06:41:22.814 CD2B.0    lying  1    -110   190    0     80   0    
06:41:23.814 CD2B.0    lying  1    -110   190    0     80   0    
06:41:24.817 CD2B.0    lying  1    -110   190    0     80   0    
06:41:25.817 CD2B.0    lying  1    -110   190    0     80   0    
06:41:26.818 CD2B.0    lying  1    -110   190    0     80   0    
06:41:27.720 CD2B.0    lying  1    -110   190    0     80   0    
06:41:28.720 CD2B.0    lying  1    -110   190    0     80   0    
06:41:29.714 CD2B.0    lying  1    -110   190    0     80   0    
06:41:30.714 CD2B.0    lying  1    -110   190    0     80   0    
06:41:31.722 CD2B.0    lying  1    -110   190    0     80   0    
06:41:32.720 CD2B.0    lying  1    -110   190    0     80   0    
06:41:33.728 CD2B.0    lying  1    -110   190    0     80   0    
06:41:34.727 CD2B.0    lying  1    -110   190    0     80   0    
06:41:35.725 CD2B.0    lying  1    -110   190    0     80   0    
06:41:36.732 CD2B.0    lying  1    -110   190    0     80   0    
06:41:37.724 CD2B.0    lying  1    -110   190    0     80   0    
06:41:38.621 CD2B.0    lying  1    -110   190    0     80   0    
06:41:39.626 CD2B.0    lying  1    -110   190    0     80   0    
06:41:40.624 CD2B.0    lying  1    -110   190    0     80   0    
06:41:41.626 CD2B.0    lying  1    -110   190    0     80   0    
06:41:42.632 CD2B.0    lying  1    -110   190    0     80   0    
06:41:43.628 CD2B.0    lying  1    -110   190    0     80   0    
06:41:44.636 CD2B.0    lying  1    -110   190    0     80   0    
06:41:45.632 CD2B.0    lying  1    -110   190    0     80   0    
06:41:46.631 CD2B.0    lying  1    -110   190    0     80   0    
06:41:47.551 CD2B.0    lying  1    -110   190    0     80   0    
06:41:48.554 CD2B.0    lying  1    -110   190    0     80   0    
06:41:49.553 CD2B.0    lying  1    -110   190    0     80   0    
06:41:50.555 CD2B.0    lying  1    -110   190    0     80   0    
06:41:51.555 CD2B.0    lying  1    -110   190    0     80   0    
06:41:52.563 CD2B.0    lying  1    -110   190    0     80   0    
06:41:53.557 CD2B.0    lying  1    -110   190    0     80   0    
06:41:54.569 CD2B.0    lying  1    -110   190    0     80   0    
06:41:55.561 CD2B.0    lying  1    -110   190    0     80   0    
06:41:56.559 CD2B.0    lying  1    -110   190    0     80   0    
06:41:57.566 CD2B.0    lying  1    -110   190    0     80   0    
06:41:58.586 CD2B.0    lying  1    -110   190    0     80   0    
06:41:59.456 CD2B.0    lying  1    -110   190    0     80   0    
06:42:00.459 CD2B.0    lying  1    -110   190    0     80   0    
06:42:01.471 CD2B.0    lying  1    -110   190    0     80   0    
06:42:02.457 CD2B.0    lying  1    -110   190    0     80   0    
06:42:03.428 CD2B.0    lying  1    -110   190    0     80   0    
06:42:04.425 CD2B.0    lying  1    -110   190    0     80   0    
06:42:05.436 CD2B.0    lying  1    -110   190    0     80   0    
06:42:06.430 CD2B.0    lying  1    -110   190    0     80   0    
06:42:07.434 CD2B.0    lying  1    -110   190    0     80   0    
06:42:08.434 CD2B.0    lying  1    -110   190    0     80   0    
06:42:09.429 CD2B.0    lying  1    -110   190    0     80   0    
06:42:10.430 CD2B.0    lying  1    -110   190    0     80   0    
06:42:11.437 CD2B.0    lying  1    -110   190    0     80   0    
06:42:12.433 CD2B.0    lying  1    -110   190    0     80   0    
06:42:13.434 CD2B.0    lying  1    -110   190    0     80   0    
06:42:14.438 CD2B.0    lying  1    -110   190    0     80   0    
06:42:15.340 CD2B.0    lying  1    -110   190    0     80   0    
06:42:16.329 CD2B.0    lying  1    -110   190    0     80   0    
06:42:17.386 CD2B.0    lying  1    -110   190    0     80   0    
06:42:18.330 CD2B.0    lying  1    -110   190    0     80   0    
06:42:19.352 CD2B.0    lying  1    -110   190    0     80   0    
06:42:20.352 CD2B.0    lying  1    -110   190    0     80   0    
06:42:21.364 CD2B.0    lying  1    -110   190    0     80   0    
06:42:22.369 CD2B.0    lying  1    -110   190    0     80   0    
06:42:23.361 CD2B.0    lying  1    -110   190    0     80   0    
06:42:24.254 CD2B.0    lying  1    -110   190    0     80   0    
06:42:25.265 CD2B.0    lying  1    -110   190    0     80   0    
06:42:26.256 CD2B.0    lying  1    -110   190    0     80   0    
06:42:27.257 CD2B.0    lying  1    -110   190    0     80   0    
06:42:28.264 CD2B.0    lying  1    -110   190    0     80   0    
06:42:29.264 CD2B.0    lying  1    -110   190    0     80   0    
06:42:30.268 CD2B.0    lying  1    -110   190    0     80   0    
06:42:31.270 CD2B.0    lying  1    -110   190    0     80   0    
06:42:32.264 CD2B.0    lying  1    -110   190    0     80   0    
06:42:33.272 CD2B.0    lying  1    -110   190    0     80   0    
06:42:34.264 CD2B.0    lying  1    -110   190    0     80   0    
06:42:35.276 CD2B.0    lying  1    -110   190    0     80   0    
06:42:36.166 CD2B.0    lying  1    -110   190    0     80   0    
06:42:37.164 CD2B.0    lying  1    -110   190    0     80   0    
06:42:38.164 CD2B.0    lying  1    -110   190    0     80   0    
06:42:39.162 CD2B.0    lying  1    -110   190    0     80   0    
06:42:40.164 CD2B.0    lying  1    -110   190    0     80   0    
06:42:41.165 CD2B.0    lying  1    -110   190    0     80   0    
06:42:42.165 CD2B.0    lying  1    -110   190    0     80   0    
06:42:43.166 CD2B.0    lying  1    -110   190    0     80   0    
06:42:44.173 CD2B.0    lying  1    -110   190    0     80   0    
06:42:45.178 CD2B.0    lying  1    -110   190    0     80   0    
06:42:46.173 CD2B.0    lying  1    -110   190    0     80   0    
06:42:47.183 CD2B.0    lying  1    -110   190    0     80   0    
06:42:48.068 CD2B.0    lying  1    -110   190    0     80   0    
06:42:49.064 CD2B.0    lying  1    -110   190    0     80   0    
06:42:50.068 CD2B.0    lying  1    -110   190    0     80   0    
06:42:51.072 CD2B.0    lying  1    -110   190    0     80   0    
06:42:52.067 CD2B.0    lying  1    -110   190    0     80   0    
06:42:53.068 CD2B.0    lying  1    -110   190    0     80   0    
06:42:54.076 CD2B.0    lying  1    -110   190    0     80   0    
06:42:55.071 CD2B.0    lying  1    -110   190    0     80   0    
06:42:56.077 CD2B.0    lying  1    -110   190    0     80   0    
06:42:57.076 CD2B.0    lying  1    -110   190    0     80   0    
06:42:58.086 CD2B.0    lying  1    -110   190    0     80   0    
06:42:59.076 CD2B.0    lying  1    -110   190    0     80   0    
06:42:59.969 CD2B.0    lying  1    -110   190    0     80   0    
06:43:00.972 CD2B.0    lying  1    -110   190    0     80   0    
06:43:01.976 CD2B.0    lying  1    -110   190    0     80   0    
06:43:02.970 CD2B.0    lying  1    -110   190    0     80   0    
06:43:03.974 CD2B.0    lying  1    -110   190    0     80   0    
06:43:04.972 CD2B.0    lying  1    -110   190    0     80   0    
06:43:05.974 CD2B.0    lying  1    -110   190    0     80   0    
06:43:06.988 CD2B.0    lying  1    -110   190    0     80   0    
06:43:07.988 CD2B.0    lying  1    -110   190    0     80   0    
06:43:08.989 CD2B.0    lying  1    -110   190    0     80   0    
06:43:09.886 CD2B.0    lying  1    -110   190    0     80   0    
06:43:10.888 CD2B.0    lying  1    -110   190    0     80   0    
06:43:11.905 CD2B.0    lying  1    -110   190    0     80   0    
06:43:12.889 CD2B.0    lying  1    -110   190    0     80   0    
06:43:13.894 CD2B.0    lying  1    -110   190    0     80   0    
06:43:14.892 CD2B.0    lying  1    -110   190    0     80   0    
06:43:15.944 CD2B.0    lying  1    -110   190    0     80   0    
06:43:16.893 CD2B.0    lying  1    -110   190    0     80   0    
06:43:17.901 CD2B.0    lying  1    -110   190    0     80   0    
06:43:18.895 CD2B.0    lying  1    -110   190    0     80   0    
06:43:19.896 CD2B.0    lying  1    -110   190    0     80   0    
06:43:20.910 CD2B.0    lying  1    -110   190    0     80   0    
06:43:21.792 CD2B.0    lying  1    -110   190    0     80   0    
06:43:22.797 CD2B.0    lying  1    -110   190    0     80   0    
06:43:23.801 CD2B.0    lying  1    -110   190    0     80   0    
06:43:24.808 CD2B.0    lying  1    -110   190    0     80   0    
06:43:25.804 CD2B.0    lying  1    -110   190    0     80   0    
06:43:26.813 CD2B.0    lying  1    -110   190    0     80   0    
06:43:27.810 CD2B.0    lying  1    -110   190    0     80   0    
06:43:28.804 CD2B.0    lying  1    -110   190    0     80   0    
06:43:29.806 CD2B.0    lying  1    -110   190    0     80   0    
06:43:30.806 CD2B.0    lying  1    -110   190    0     80   0    
06:43:31.809 CD2B.0    lying  1    -110   190    0     80   0    
06:43:32.711 CD2B.0    lying  1    -110   190    0     80   0    
06:43:33.711 CD2B.0    lying  1    -110   190    0     80   0    
06:43:34.708 CD2B.0    lying  1    -110   190    0     80   0    
06:43:35.706 CD2B.0    lying  1    -110   190    0     80   0    
06:43:36.707 CD2B.0    lying  1    -110   190    0     80   0    
06:43:37.718 CD2B.0    lying  1    -110   190    0     80   0    
06:43:38.709 CD2B.0    lying  1    -110   190    0     80   0    
06:43:39.657 CD2B.0    lying  1    -110   190    0     80   0    
06:43:40.649 CD2B.0    lying  1    -110   190    0     80   0    
06:43:41.650 CD2B.0    lying  1    -110   190    0     80   0    
06:43:42.650 CD2B.0    lying  1    -110   190    0     80   0    
06:43:43.656 CD2B.0    lying  1    -110   190    0     80   0    
06:43:44.653 CD2B.0    lying  1    -110   190    0     80   0    
06:43:45.652 CD2B.0    lying  1    -110   190    0     80   0    
06:43:46.654 CD2B.0    lying  1    -110   190    0     80   0    
06:43:47.660 CD2B.0    lying  1    -110   190    0     80   0    
06:43:48.661 CD2B.0    lying  1    -110   190    0     80   0    
06:43:49.661 CD2B.0    lying  1    -110   190    0     80   0    
06:43:50.689 CD2B.0    lying  1    -110   190    0     80   0    
06:43:51.558 CD2B.0    lying  1    -110   190    0     80   0    
06:43:52.557 CD2B.0    lying  1    -110   190    0     80   0    
06:43:53.556 CD2B.0    lying  1    -110   190    0     80   0    
06:43:54.555 CD2B.0    lying  1    -110   190    0     80   0    
06:43:55.522 CD2B.0    lying  1    -110   190    0     80   0    
06:43:56.521 CD2B.0    lying  1    -110   190    0     80   0    
06:43:57.526 CD2B.0    lying  1    -110   190    0     80   0    
06:43:58.523 CD2B.0    lying  1    -110   190    0     80   0    
06:43:59.525 CD2B.0    lying  1    -110   190    0     80   0    
06:44:00.523 CD2B.0    lying  1    -110   190    0     80   0    
06:44:01.529 CD2B.0    lying  1    -110   190    0     80   0    
06:44:02.532 CD2B.0    lying  1    -110   190    0     80   0    
06:44:03.532 CD2B.0    lying  1    -110   190    0     80   0    
06:44:04.528 CD2B.0    lying  1    -110   190    0     80   0    
06:44:05.532 CD2B.0    lying  1    -110   190    0     80   0    
06:44:06.529 CD2B.0    lying  1    -110   190    0     80   0    
06:44:07.424 CD2B.0    lying  1    -110   190    0     80   0    
06:44:08.448 CD2B.0    lying  1    -110   190    0     80   0    
06:44:09.426 CD2B.0    lying  1    -110   190    0     80   0    
06:44:10.427 CD2B.0    lying  1    -110   190    0     80   0    
06:44:11.438 CD2B.0    lying  1    -110   190    0     80   0    
06:44:12.432 CD2B.0    lying  1    -110   190    0     80   0    
06:44:13.436 CD2B.0    lying  1    -110   190    0     80   0    
06:44:14.433 CD2B.0    lying  1    -110   190    0     80   0    
06:44:15.493 CD2B.0    lying  1    -110   190    0     80   0    
06:44:16.442 CD2B.0    lying  1    -110   190    0     80   0    
06:44:17.440 CD2B.0    lying  1    -110   190    0     80   0    
06:44:18.336 CD2B.0    lying  1    -110   190    0     80   0    
06:44:19.337 CD2B.0    lying  1    -110   190    0     80   0    
06:44:20.337 CD2B.0    lying  1    -110   190    0     80   0    
06:44:21.338 CD2B.0    lying  1    -110   190    0     80   0    
06:44:22.340 CD2B.0    lying  1    -110   190    0     80   0    
06:44:23.342 CD2B.0    lying  1    -110   190    0     80   0    
06:44:24.343 CD2B.0    lying  1    -110   190    0     80   0    
06:44:25.358 CD2B.0    lying  1    -110   190    0     80   0    
06:44:26.352 CD2B.0    lying  1    -110   190    0     80   0    
06:44:27.272 CD2B.0    lying  1    -110   190    0     80   0    
06:44:28.265 CD2B.0    lying  1    -110   190    0     80   0    
06:44:29.269 CD2B.0    lying  1    -110   190    0     80   0    
06:44:30.285 CD2B.0    lying  1    -110   190    0     80   0    
06:44:31.268 CD2B.0    lying  1    -110   190    0     80   0    
06:44:32.272 CD2B.0    lying  1    -110   190    0     80   0    
06:44:33.269 CD2B.0    lying  1    -110   190    0     80   0    
06:44:34.285 CD2B.0    lying  1    -110   190    0     80   0    
06:44:35.308 CD2B.0    lying  1    -110   190    0     80   0    
06:44:36.272 CD2B.0    lying  1    -110   190    0     80   0    
06:44:37.273 CD2B.0    lying  1    -110   190    0     80   0    
06:44:38.274 CD2B.0    lying  1    -110   190    0     80   0    
06:44:39.170 CD2B.0    lying  1    -110   190    0     80   0    
06:44:40.172 CD2B.0    lying  1    -110   190    0     80   0    
06:44:41.168 CD2B.0    lying  1    -110   190    0     80   0    
06:44:42.174 CD2B.0    lying  1    -110   190    0     80   0    
06:44:43.145 CD2B.0    lying  1    -110   190    0     80   0    
06:44:44.137 CD2B.0    lying  1    -110   190    0     80   0    
06:44:45.146 CD2B.0    lying  1    -110   190    0     80   0    
06:44:46.138 CD2B.0    lying  1    -110   190    0     80   0    
06:44:47.145 CD2B.0    lying  1    -110   190    0     80   0    
06:44:48.141 CD2B.0    lying  1    -110   190    0     80   0    
06:44:49.141 CD2B.0    lying  1    -110   190    0     80   0    
06:44:50.146 CD2B.0    lying  1    -110   190    0     80   0    
06:44:51.149 CD2B.0    lying  1    -110   190    0     80   0    
06:44:52.153 CD2B.0    lying  1    -110   190    0     80   0    
06:44:53.145 CD2B.0    lying  1    -110   190    0     80   0    
06:44:54.145 CD2B.0    lying  1    -110   190    0     80   0    
06:44:55.049 CD2B.0    lying  1    -110   190    0     80   0    
06:44:56.045 CD2B.0    lying  1    -110   190    0     80   0    
06:44:57.040 CD2B.0    lying  1    -110   190    0     80   0    
06:44:58.053 CD2B.0    lying  1    -110   190    0     80   0    
06:44:59.041 CD2B.0    lying  1    -110   190    0     80   0    
06:45:00.066 CD2B.0    lying  1    -110   190    0     80   0    
06:45:01.068 CD2B.0    lying  1    -110   190    0     80   0    
06:45:02.067 CD2B.0    lying  1    -110   190    0     80   0    
06:45:03.082 CD2B.0    lying  1    -110   190    0     80   0    
06:45:03.971 CD2B.0    lying  1    -110   190    0     80   0    
06:45:04.974 CD2B.0    lying  1    -110   190    0     80   0    
06:45:05.969 CD2B.0    lying  1    -110   190    0     80   0    
06:45:06.977 CD2B.0    lying  1    -110   190    0     80   0    
06:45:07.972 CD2B.0    lying  1    -110   190    0     80   0    
06:45:08.971 CD2B.0    lying  1    -110   190    0     80   0    
06:45:09.976 CD2B.0    lying  1    -110   190    0     80   0    
06:45:10.980 CD2B.0    lying  1    -110   190    0     80   0    
06:45:11.979 CD2B.0    lying  1    -110   190    0     80   0    
06:45:12.982 CD2B.0    lying  1    -110   190    0     80   0    
06:45:13.976 CD2B.0    lying  1    -110   190    0     80   0    
06:45:15.032 CD2B.0    lying  1    -110   190    0     80   0    
06:45:15.876 CD2B.0    lying  1    -110   190    0     80   0    
06:45:16.879 CD2B.0    lying  1    -110   190    0     80   0    
06:45:17.878 CD2B.0    lying  1    -110   190    0     80   0    
06:45:18.881 CD2B.0    lying  1    -110   190    0     80   0    
06:45:19.893 CD2B.0    lying  1    -110   190    0     80   0    
06:45:20.885 CD2B.0    lying  1    -110   190    0     80   0    
06:45:21.881 CD2B.0    lying  1    -110   190    0     80   0    
06:45:22.893 CD2B.0    lying  1    -110   190    0     80   0    
06:45:23.888 CD2B.0    lying  1    -110   190    0     80   0    
06:45:24.888 CD2B.0    lying  1    -110   190    0     80   0    
06:45:25.886 CD2B.0    lying  1    -110   190    0     80   0    
06:45:26.793 CD2B.0    lying  1    -110   190    0     80   0    
06:45:27.789 CD2B.0    lying  1    -110   190    0     80   0    
06:45:28.786 CD2B.0    lying  1    -110   190    0     80   0    
06:45:29.791 CD2B.0    lying  1    -110   190    0     80   0    
06:45:30.785 CD2B.0    lying  1    -110   190    0     80   0    
06:45:31.741 CD2B.0    lying  1    -110   190    0     80   0    
06:45:32.744 CD2B.0    lying  1    -110   190    0     80   0    
06:45:33.744 CD2B.0    lying  1    -110   190    0     80   0    
06:45:34.745 CD2B.0    lying  1    -110   190    0     80   0    
06:45:35.746 CD2B.0    lying  1    -110   190    0     80   0    
06:45:36.757 CD2B.0    lying  1    -110   190    0     80   0    
06:45:37.748 CD2B.0    lying  1    -110   190    0     80   0    
06:45:38.749 CD2B.0    lying  1    -110   190    0     80   0    
06:45:39.752 CD2B.0    lying  1    -110   190    0     80   0    
06:45:40.760 CD2B.0    lying  1    -110   190    0     80   0    
06:45:41.756 CD2B.0    lying  1    -110   190    0     80   0    
06:45:42.757 CD2B.0    lying  1    -110   190    0     80   0    
06:45:43.650 CD2B.0    lying  1    -110   190    0     80   0    
06:45:44.648 CD2B.0    lying  1    -110   190    0     80   0    
06:45:45.653 CD2B.0    lying  1    -110   190    0     80   0    
06:45:46.652 CD2B.0    lying  1    -110   190    0     80   0    
06:45:47.698 CD2B.0    lying  1    -110   190    0     80   0    
06:45:48.701 CD2B.0    lying  1    -110   190    0     80   0    
06:45:49.601 CD2B.0    lying  1    -110   190    0     80   0    
06:45:50.598 CD2B.0    lying  1    -110   190    0     80   0    
06:45:51.605 CD2B.0    lying  1    -110   190    0     80   0    
06:45:52.601 CD2B.0    lying  1    -110   190    0     80   0    
06:45:53.602 CD2B.0    lying  1    -110   190    0     80   0    
06:45:54.607 CD2B.0    lying  1    -110   190    0     80   0    
06:45:55.606 CD2B.0    lying  1    -110   190    0     80   0    
06:45:56.606 CD2B.0    lying  1    -110   190    0     80   0    
06:45:57.607 CD2B.0    lying  1    -110   190    0     80   0    
06:45:58.606 CD2B.0    lying  1    -110   190    0     80   0    
06:45:59.612 CD2B.0    lying  1    -110   190    0     80   0    
06:46:00.616 CD2B.0    lying  1    -110   190    0     80   0    
06:46:01.505 CD2B.0    lying  1    -110   190    0     80   0    
06:46:02.512 CD2B.0    lying  1    -110   190    0     80   0    
06:46:03.509 CD2B.0    lying  1    -110   190    0     80   0    
06:46:04.509 CD2B.0    lying  1    -110   190    0     80   0    
06:46:05.514 CD2B.0    lying  1    -110   190    0     80   0    
06:46:06.517 CD2B.0    lying  1    -110   190    0     80   0    
06:46:07.517 CD2B.0    lying  1    -110   190    0     80   0    
06:46:08.513 CD2B.0    lying  1    -110   190    0     80   0    
06:46:09.521 CD2B.0    lying  1    -110   190    0     80   0    
06:46:10.515 CD2B.0    lying  1    -110   190    0     80   0    
06:46:11.520 CD2B.0    lying  1    -110   190    0     80   0    
06:46:12.420 CD2B.0    lying  1    -110   190    0     80   0    
06:46:13.417 CD2B.0    lying  1    -110   190    0     80   0    
06:46:14.488 CD2B.0    lying  1    -110   190    0     80   0    
06:46:15.425 CD2B.0    lying  1    -110   190    0     80   0    
06:46:16.424 CD2B.0    lying  1    -110   190    0     80   0    
06:46:17.419 CD2B.0    lying  1    -110   190    0     80   0    
06:46:18.424 CD2B.0    lying  1    -110   190    0     80   0    
06:46:19.358 CD2B.0    lying  1    -110   190    0     80   0    
06:46:20.360 CD2B.0    lying  1    -110   190    0     80   0    
06:46:21.364 CD2B.0    lying  1    -110   190    0     80   0    
06:46:22.362 CD2B.0    lying  1    -110   190    0     80   0    
06:46:23.366 CD2B.0    lying  1    -110   190    0     80   0    
06:46:24.374 CD2B.0    lying  1    -110   190    0     80   0    
06:46:25.365 CD2B.0    lying  1    -110   190    0     80   0    
06:46:26.369 CD2B.0    lying  1    -110   190    0     80   0    
06:46:27.366 CD2B.0    lying  1    -110   190    0     80   0    
06:46:28.368 CD2B.0    lying  1    -110   190    0     80   0    
06:46:29.375 CD2B.0    lying  1    -110   190    0     80   0    
06:46:30.377 CD2B.0    lying  1    -110   190    0     80   0    
06:46:31.262 CD2B.0    lying  1    -110   190    0     80   0    
06:46:32.263 CD2B.0    lying  1    -110   190    0     80   0    
06:46:33.264 CD2B.0    lying  1    -110   190    0     80   0    
06:46:34.268 CD2B.0    lying  1    -110   190    0     80   0    
06:46:35.244 CD2B.0    lying  1    -110   190    0     80   0    
06:46:36.236 CD2B.0    lying  1    -110   190    0     80   0    
06:46:37.238 CD2B.0    lying  1    -110   190    0     80   0    
06:46:38.245 CD2B.0    lying  1    -110   190    0     80   0    
06:46:39.234 CD2B.0    lying  1    -110   190    0     80   0    
06:46:40.238 CD2B.0    lying  1    -110   190    0     80   0    
06:46:41.236 CD2B.0    lying  1    -110   190    0     80   0    
06:46:42.238 CD2B.0    lying  1    -110   190    0     80   0    
06:46:43.240 CD2B.0    lying  1    -110   190    0     80   0    
06:46:44.240 CD2B.0    lying  1    -110   190    0     80   0    
06:46:45.240 CD2B.0    lying  1    -110   190    0     80   0    
06:46:46.248 CD2B.0    lying  1    -110   190    0     80   0    
06:46:47.145 CD2B.0    lying  1    -110   190    0     80   0    
06:46:48.136 CD2B.0    lying  1    -110   190    0     80   0    
06:46:49.140 CD2B.0    lying  1    -110   190    0     80   0    
06:46:50.141 CD2B.0    lying  1    -110   190    0     80   0    
06:46:51.138 CD2B.0    lying  1    -110   190    0     80   0    
06:46:52.157 CD2B.0    lying  1    -110   190    0     80   0    
06:46:53.162 CD2B.0    lying  1    -110   190    0     80   0    
06:46:54.158 CD2B.0    lying  1    -110   190    0     80   0    
06:46:55.168 CD2B.0    lying  1    -110   190    0     80   0    
06:46:56.161 CD2B.0    lying  1    -110   190    0     80   0    
06:46:57.062 CD2B.0    lying  1    -110   190    0     80   0    
06:46:58.056 CD2B.0    lying  1    -100   190    0     80   10   
06:46:59.056 CD2B.0    lying  1    -120   210    75    80   28   
06:47:00.058 CD2B.0    lying  1    -120   190    79    80   20   
06:47:01.063 CD2B.0    lying  1    -120   190    70    80   0    
06:47:02.061 CD2B.0    lying  1    -110   200    70    80   14   
06:47:03.064 CD2B.0    lying  1    -110   200    71    80   0    
06:47:04.067 CD2B.0    lying  1    -80    200    77    80   30   
06:47:05.062 CD2B.0    lying  1    -150   180    67    80   72   
06:47:06.064 CD2B.0    lying  1    -140   190    67    80   14   
06:47:07.076 CD2B.0    lying  1    -110   200    0     80   31   
06:47:07.966 CD2B.0    lying  1    -110   200    0     80   0    
06:47:08.967 CD2B.0    lying  1    -110   200    0     80   0    
06:47:09.969 CD2B.0    lying  1    -110   200    0     80   0    
06:47:10.968 CD2B.0    lying  1    -110   200    0     80   0    
06:47:11.981 CD2B.0    lying  1    -110   200    0     80   0    
06:47:12.974 CD2B.0    lying  1    -110   200    0     80   0    
06:47:14.024 CD2B.0    lying  1    -110   200    0     80   0    
06:47:14.972 CD2B.0    lying  1    -110   200    0     80   0    
06:47:15.976 CD2B.0    lying  1    -110   200    0     80   0    
06:47:16.976 CD2B.0    lying  1    -110   200    0     80   0    
06:47:17.977 CD2B.0    lying  1    -110   200    0     80   0    
06:47:18.977 CD2B.0    lying  1    -110   200    0     80   0    
06:47:19.870 CD2B.0    lying  1    -110   200    0     80   0    
06:47:20.870 CD2B.0    lying  1    -110   200    0     80   0    
06:47:21.876 CD2B.0    lying  1    -110   200    0     80   0    
06:47:22.872 CD2B.0    lying  1    -110   200    0     80   0    
06:47:23.837 CD2B.0    lying  1    -110   200    0     80   0    
06:47:24.839 CD2B.0    lying  1    -110   200    0     80   0    
06:47:25.841 CD2B.0    lying  1    -110   200    0     80   0    
06:47:26.841 CD2B.0    lying  1    -110   200    0     80   0    
06:47:27.848 CD2B.0    lying  1    -110   200    0     80   0    
06:47:28.848 CD2B.0    lying  1    -110   200    0     80   0    
06:47:29.853 CD2B.0    lying  1    -110   200    0     80   0    
06:47:30.848 CD2B.0    lying  1    -110   200    0     80   0    
06:47:31.848 CD2B.0    lying  1    -110   200    0     80   0    
06:47:32.846 CD2B.0    lying  1    -110   200    0     80   0    
06:47:33.848 CD2B.0    lying  1    -110   200    0     80   0    
06:47:34.854 CD2B.0    lying  1    -130   190    70    80   22   
06:47:35.746 CD2B.0    lying  1    -150   180    64    80   22   
06:47:36.749 CD2B.0    lying  1    -130   190    66    80   22   
06:47:37.756 CD2B.0    lying  1    -80    210    0     80   53   
06:47:38.768 CD2B.0    lying  1    -80    210    0     80   0    
06:47:39.785 CD2B.0    lying  1    -80    210    0     80   0    
06:47:40.797 CD2B.0    lying  1    -80    210    0     80   0    
06:47:41.782 CD2B.0    lying  1    -80    210    0     80   0    
06:47:42.780 CD2B.0    lying  1    -90    210    0     80   10   
06:47:43.680 CD2B.0    lying  1    -90    210    0     80   0    
06:47:44.680 CD2B.0    lying  1    -90    210    0     80   0    
06:47:45.686 CD2B.0    lying  1    -90    210    0     80   0    
06:47:46.684 CD2B.0    lying  1    -90    210    0     80   0    
06:47:47.689 CD2B.0    lying  1    -90    210    0     80   0    
06:47:48.684 CD2B.0    lying  1    -90    210    0     80   0    
06:47:49.686 CD2B.0    lying  1    -90    210    0     80   0    
06:47:50.685 CD2B.0    lying  1    -90    210    0     80   0    
06:47:51.692 CD2B.0    lying  1    -90    210    0     80   0    
06:47:52.688 CD2B.0    lying  1    -90    210    0     80   0    
06:47:53.690 CD2B.0    lying  1    -90    210    0     80   0    
06:47:54.700 CD2B.0    lying  1    -90    210    0     80   0    
06:47:55.597 CD2B.0    lying  1    -90    210    0     80   0    
06:47:56.600 CD2B.0    lying  1    -90    210    0     80   0    
06:47:57.592 CD2B.0    lying  1    -90    210    0     80   0    
06:47:58.594 CD2B.0    lying  1    -90    210    0     80   0    
06:47:59.593 CD2B.0    lying  1    -90    210    0     80   0    
06:48:00.594 CD2B.0    lying  1    -90    210    0     80   0    
06:48:01.596 CD2B.0    lying  1    -90    210    0     80   0    
06:48:02.596 CD2B.0    lying  1    -90    210    0     80   0    
06:48:03.604 CD2B.0    lying  1    -90    210    0     80   0    
06:48:04.605 CD2B.0    lying  1    -90    210    0     80   0    
06:48:05.609 CD2B.0    lying  1    -90    210    0     80   0    
06:48:06.495 CD2B.0    lying  1    -90    210    0     80   0    
06:48:07.496 CD2B.0    lying  1    -90    210    0     80   0    
06:48:08.500 CD2B.0    lying  1    -90    210    0     80   0    
06:48:09.504 CD2B.0    lying  1    -90    210    0     80   0    
06:48:10.498 CD2B.0    lying  1    -90    210    0     80   0    
06:48:11.508 CD2B.0    lying  1    -90    210    0     80   0    
06:48:12.506 CD2B.0    lying  1    -90    210    0     80   0    
06:48:13.550 CD2B.0    lying  1    -90    210    0     80   0    
06:48:14.505 CD2B.0    lying  1    -90    210    0     80   0    
06:48:15.502 CD2B.0    lying  1    -90    210    0     80   0    
06:48:16.506 CD2B.0    lying  1    -90    210    0     80   0    
06:48:17.507 CD2B.0    lying  1    -90    210    0     80   0    
06:48:18.398 CD2B.0    lying  1    -90    210    0     80   0    
06:48:19.402 CD2B.0    lying  1    -90    210    0     80   0    
06:48:20.400 CD2B.0    lying  1    -90    210    0     80   0    
06:48:21.401 CD2B.0    lying  1    -90    210    0     80   0    
06:48:22.403 CD2B.0    lying  1    -90    210    0     80   0    
06:48:23.410 CD2B.0    lying  1    -90    210    0     80   0    
06:48:24.404 CD2B.0    lying  1    -90    210    0     80   0    
06:48:25.413 CD2B.0    lying  1    -90    210    0     80   0    
06:48:26.410 CD2B.0    lying  1    -90    210    0     80   0    
06:48:27.422 CD2B.0    lying  1    -90    210    0     80   0    
06:48:28.412 CD2B.0    lying  1    -90    210    0     80   0    
06:48:29.316 CD2B.0    lying  1    -90    210    0     80   0    
06:48:30.312 CD2B.0    lying  1    -90    210    0     80   0    
06:48:31.312 CD2B.0    lying  1    -90    210    0     80   0    
06:48:32.315 CD2B.0    lying  1    -90    210    0     80   0    
06:48:33.314 CD2B.0    lying  1    -90    210    0     80   0    
06:48:34.321 CD2B.0    lying  1    -90    210    0     80   0    
06:48:35.334 CD2B.0    lying  1    -90    210    0     80   0    
06:48:36.331 CD2B.0    lying  1    -90    210    0     80   0    
06:48:37.322 CD2B.0    lying  1    -90    210    0     80   0    
06:48:38.321 CD2B.0    lying  1    -90    210    0     80   0    
06:48:39.321 CD2B.0    lying  1    -90    210    0     80   0    
06:48:40.328 CD2B.0    lying  1    -90    210    0     80   0    
06:48:41.220 CD2B.0    lying  1    -90    210    0     80   0    
06:48:42.221 CD2B.0    lying  1    -90    210    0     80   0    
06:48:43.237 CD2B.0    lying  1    -90    210    0     80   0    
06:48:44.229 CD2B.0    lying  1    -120   190    0     80   36   
06:48:45.238 CD2B.0    lying  1    -180   170    68    80   63   
06:48:46.233 CD2B.0    lying  1    -160   170    68    80   20   
06:48:47.239 CD2B.0    walk   1    -230   140    0     80   76   
06:48:48.233 CD2B.0    stand  1    -230   140    0     80   0    
06:48:49.235 CD2B.0    stand  1    -220   140    0     80   10   
06:48:50.240 CD2B.0    stand  1    -220   140    0     80   0    
06:48:51.134 CD2B.0    stand  1    -220   140    0     80   0    
06:48:52.137 CD2B.0    stand  1    -220   150    0     80   10   
06:48:53.137 CD2B.0    stand  1    -220   150    0     80   0    
06:48:54.137 CD2B.0    stand  1    -210   160    0     80   14   
06:48:55.140 CD2B.0    stand  1    -220   150    0     80   14   
06:48:56.140 CD2B.0    stand  1    -220   150    0     80   0    
06:48:57.140 CD2B.0    stand  1    -220   150    0     80   0    
06:48:58.144 CD2B.0    stand  1    -220   150    0     80   0    
06:48:59.146 CD2B.0    stand  1    -220   150    0     80   0    
06:49:00.145 CD2B.0    stand  1    -220   150    0     80   0    
06:49:01.146 CD2B.0    stand  1    -220   150    0     80   0    
06:49:02.148 CD2B.0    stand  1    -220   150    0     80   0    
06:49:03.040 CD2B.0    stand  1    -220   150    0     80   0    
06:49:04.040 CD2B.0    stand  1    -220   150    0     80   0    
06:49:05.043 CD2B.0    stand  1    -220   150    0     80   0    
06:49:06.054 CD2B.0    stand  1    -220   150    0     80   0    
06:49:07.056 CD2B.0    stand  1    -220   150    0     80   0    
06:49:08.045 CD2B.0    stand  1    -220   150    0     80   0    
06:49:09.045 CD2B.0    stand  1    -220   150    0     80   0    
06:49:10.045 CD2B.0    stand  1    -220   150    0     80   0    
06:49:11.052 CD2B.0    stand  1    -220   150    0     80   0    
06:49:12.053 CD2B.0    stand  1    -220   150    0     80   0    
06:49:13.096 CD2B.0    stand  1    -220   150    0     80   0    
06:49:14.052 CD2B.0    stand  1    -220   150    0     80   0    
06:49:14.948 CD2B.0    stand  1    -220   150    0     80   0    
06:49:15.944 CD2B.0    stand  1    -220   150    0     80   0    
06:49:16.952 CD2B.0    stand  1    -220   150    0     80   0    
06:49:17.944 CD2B.0    stand  1    -220   150    0     80   0    
06:49:18.947 CD2B.0    stand  1    -220   150    0     80   0    
06:49:19.954 CD2B.0    stand  1    -220   150    0     80   0    
06:49:20.958 CD2B.0    stand  1    -220   150    0     80   0    
06:49:21.956 CD2B.0    stand  1    -220   150    0     80   0    
06:49:22.956 CD2B.0    stand  1    -220   150    0     80   0    
06:49:23.950 CD2B.0    stand  1    -220   150    0     80   0    
06:49:24.953 CD2B.0    stand  1    -220   150    0     80   0    
06:49:25.954 CD2B.0    stand  1    -220   150    0     80   0    
06:49:26.853 CD2B.0    stand  1    -220   150    0     80   0    
06:49:27.850 CD2B.0    stand  1    -220   130    67    80   20   
06:49:28.852 CD2B.0    walk   1    -150   90     61    80   80   
06:49:29.852 CD2B.0    walk   1    -100   140    66    80   70   
06:49:30.852 CD2B.0    lying  1    -110   140    64    80   10   
06:49:31.872 CD2B.0    lying  1    -100   140    63    80   10   
06:49:32.860 CD2B.0    lying  1    -100   150    62    80   10   
06:49:33.863 CD2B.0    lying  1    -110   190    68    80   41   
06:49:34.862 CD2B.0    lying  1    -120   200    70    80   14   
06:49:35.864 CD2B.0    lying  1    -120   180    68    80   20   
06:49:36.872 CD2B.0    lying  1    -130   190    75    80   14   
06:49:37.769 CD2B.0    lying  1    -120   180    68    80   14   
06:49:38.764 CD2B.0    lying  1    -120   180    74    80   0    
06:49:39.765 CD2B.0    lying  1    -120   170    67    80   10   
06:49:40.782 CD2B.0    lying  1    -120   180    66    80   10   
06:49:41.766 CD2B.0    lying  1    -110   160    71    80   22   
06:49:42.769 CD2B.0    lying  1    -110   180    0     80   20   
06:49:43.768 CD2B.0    lying  1    -90    160    60    80   28   
06:49:44.776 CD2B.0    lying  1    -120   150    65    80   31   
06:49:45.770 CD2B.0    lying  1    -100   140    66    80   22   
06:49:46.771 CD2B.0    lying  1    -100   150    65    80   10   
06:49:47.682 CD2B.0    lying  1    -80    190    66    80   44   
06:49:48.684 CD2B.0    lying  1    -80    170    0     80   20   
06:49:49.687 CD2B.0    lying  1    -90    170    0     80   10   
06:49:50.684 CD2B.0    lying  1    -120   160    0     80   31   
06:49:51.684 CD2B.0    lying  1    -110   140    59    80   22   
06:49:52.688 CD2B.0    lying  1    -120   140    61    80   10   
06:49:53.692 CD2B.0    lying  1    -90    140    63    80   30   
06:49:54.688 CD2B.0    lying  1    -90    130    0     80   10   
06:49:55.689 CD2B.0    lying  1    -90    130    0     80   0    
06:49:56.702 CD2B.0    lying  1    -90    130    0     80   0    
06:49:57.693 CD2B.0    lying  1    -100   140    67    80   14   
06:49:58.598 CD2B.0    lying  1    -90    160    0     80   22   
06:49:59.594 CD2B.0    lying  1    -130   150    58    80   41   
06:50:00.596 CD2B.0    lying  1    -130   140    64    80   10   
06:50:01.599 CD2B.0    lying  1    -100   150    60    80   31   
06:50:02.595 CD2B.0    stand  1    -100   120    0     80   30   
06:50:03.558 CD2B.0    stand  1    -110   120    0     80   10   
06:50:04.554 CD2B.0    stand  1    -110   120    0     80   0    
06:50:05.560 CD2B.0    stand  1    -110   120    0     80   0    
06:50:06.557 CD2B.0    stand  1    -110   120    0     80   0    
06:50:07.556 CD2B.0    stand  1    -170   140    0     80   63   
06:50:08.556 CD2B.0    walk   1    -230   160    61    80   63   
06:50:09.556 CD2B.0    walk   1    -210   180    81    80   28   
06:50:10.566 CD2B.0    walk   1    -170   170    0     80   41   
06:50:11.997 CD2B.0    walk   1    -120   190    0     80   53   
06:50:12.773 CD2B.0    walk   1    -90    200    0     80   31   
06:50:13.560 CD2B.0    stand  1    -120   170    0     80   42   
06:50:14.571 CD2B.0    stand  1    -220   170    0     80   100  
06:50:15.456 CD2B.0    stand  1    -230   170    0     80   10   
06:50:16.457 CD2B.0    stand  1    -230   170    0     80   0    
06:50:17.456 CD2B.0    stand  1    -220   170    0     80   10   
06:50:18.462 CD2B.0    stand  1    -220   170    0     80   0    
06:50:19.495 CD2B.0    stand  1    -220   170    0     80   0    
06:50:20.494 CD2B.0    stand  1    -220   170    0     80   0    
06:50:21.504 CD2B.0    stand  1    -220   170    0     80   0    
06:50:22.498 CD2B.0    stand  1    -220   170    0     80   0    
06:50:23.393 CD2B.0    stand  1    -220   170    0     80   0    
06:50:24.394 CD2B.0    stand  1    -220   170    0     80   0    
06:50:25.398 CD2B.0    stand  1    -220   170    0     80   0    
06:50:26.397 CD2B.0    stand  1    -220   170    0     80   0    
06:50:27.394 CD2B.0    stand  1    -220   170    0     80   0    
06:50:28.396 CD2B.0    stand  1    -220   170    0     80   0    
06:50:29.404 CD2B.0    stand  1    -220   170    0     80   0    
06:50:30.408 CD2B.0    stand  1    -60    140    0     80   162  
06:50:31.408 CD2B.0    stand  1    -50    140    0     80   10   
06:50:32.400 CD2B.0    stand  1    -60    140    0     80   10   
06:50:33.404 CD2B.0    stand  1    -60    140    0     80   0    
06:50:34.401 CD2B.0    stand  1    -60    150    73    80   10   
06:50:35.320 CD2B.0    lying  1    -90    140    66    80   31   
06:50:36.312 CD2B.0    stand  1    -130   130    65    80   41   
06:50:37.326 CD2B.0    stand  1    -140   130    57    80   10   
06:50:38.312 CD2B.0    stand  1    -100   120    58    80   41   
06:50:39.313 CD2B.0    lying  1    -90    130    47    80   14   
06:50:40.320 CD2B.0    lying  1    -120   140    48    80   31   
06:50:41.318 CD2B.0    lying  1    -80    140    55    80   40   
06:50:42.320 CD2B.0    lying  1    -70    140    62    80   10   
06:50:43.320 CD2B.0    stand  1    -80    120    48    80   22   
06:50:44.320 CD2B.0    stand  1    -80    120    63    80   0    
06:50:45.225 CD2B.0    stand  1    -70    110    73    80   14   
06:50:46.218 CD2B.0    stand  1    -110   90     71    80   44   
06:50:47.220 CD2B.0    walk   1    -110   0      77    80   90   
06:50:48.225 CD2B.0    walk   1    -110   -10    0     80   10   
06:50:49.225 CD2B.0    lying  1    -120   190    66    80   200  
06:50:50.224 CD2B.0    lying  1    -120   180    68    80   10   
06:50:51.222 CD2B.0    lying  1    -130   170    82    80   14   
06:50:52.161 CD2B.0    lying  1    -120   180    75    80   14   
06:50:53.173 CD2B.0    lying  1    -120   170    81    80   10   
06:50:54.162 CD2B.0    lying  1    -160   160    71    80   41   
06:50:55.171 CD2B.0    lying  1    -120   170    85    80   41   
06:50:56.165 CD2B.0    lying  1    -110   180    84    80   14   
06:50:57.169 CD2B.0    lying  1    -120   190    75    80   14   
06:50:58.170 CD2B.0    lying  1    -90    200    77    80   31   
06:50:59.168 CD2B.0    lying  1    -60    210    90    80   31   
06:51:00.176 CD2B.0    lying  1    -110   180    74    80   58   
06:51:01.174 CD2B.0    lying  1    -90    200    89    80   28   
06:51:02.174 CD2B.0    lying  1    -130   170    0     80   50   
06:51:03.073 CD2B.0    lying  1    -80    210    0     80   64   
06:51:04.081 CD2B.0    lying  1    -130   160    0     80   70   
06:51:05.074 CD2B.0    lying  1    -130   170    0     80   10   
06:51:06.077 CD2B.0    lying  1    -130   170    0     80   0    
06:51:07.077 CD2B.0    lying  1    -130   170    0     80   0    
06:51:08.128 CD2B.0    lying  1    -130   170    0     80   0    
06:51:09.024 CD2B.0    lying  1    -130   170    0     80   0    
06:51:10.041 CD2B.0    lying  1    -130   170    0     80   0    
06:51:11.029 CD2B.0    lying  1    -130   170    0     80   0    
06:51:12.091 CD2B.0    lying  1    -130   170    0     80   0    
06:51:13.030 CD2B.0    lying  1    -130   170    0     80   0    
06:51:14.032 CD2B.0    lying  1    -130   170    0     80   0    
06:51:15.029 CD2B.0    lying  1    -130   170    0     80   0    
06:51:16.041 CD2B.0    lying  1    -130   170    0     80   0    
06:51:17.037 CD2B.0    lying  1    -130   170    0     80   0    
06:51:18.032 CD2B.0    lying  1    -130   170    0     80   0    
06:51:19.032 CD2B.0    lying  1    -130   170    0     80   0    
06:51:20.036 CD2B.0    lying  1    -130   170    0     80   0    
06:51:20.954 CD2B.0    lying  1    -130   170    0     80   0    
06:51:21.934 CD2B.0    lying  1    -130   170    0     80   0    
06:51:22.936 CD2B.0    lying  1    -130   170    0     80   0    
06:51:23.942 CD2B.0    lying  1    -130   170    0     80   0    
06:51:24.937 CD2B.0    lying  1    -130   170    0     80   0    
06:51:25.950 CD2B.0    lying  1    -130   170    0     80   0    
06:51:26.942 CD2B.0    lying  1    -130   170    0     80   0    
06:51:27.942 CD2B.0    lying  1    -130   170    0     80   0    
06:51:28.946 CD2B.0    lying  1    -130   170    0     80   0    
06:51:29.943 CD2B.0    lying  1    -130   170    0     80   0    
06:51:30.945 CD2B.0    lying  1    -130   170    0     80   0    
06:51:31.843 CD2B.0    lying  1    -130   170    0     80   0    
06:51:32.851 CD2B.0    lying  1    -130   170    0     80   0    
06:51:33.840 CD2B.0    lying  1    -130   170    0     80   0    
06:51:34.841 CD2B.0    lying  1    -130   170    0     80   0    
06:51:35.842 CD2B.0    lying  1    -130   170    0     80   0    
06:51:36.845 CD2B.0    lying  1    -130   170    0     80   0    
06:51:37.848 CD2B.0    lying  1    -130   170    0     80   0    
06:51:38.847 CD2B.0    lying  1    -130   170    0     80   0    
06:51:39.849 CD2B.0    lying  1    -130   170    0     80   0    
06:51:40.871 CD2B.0    lying  1    -130   170    0     80   0    
06:51:41.850 CD2B.0    lying  1    -130   170    0     80   0    
06:51:42.859 CD2B.0    lying  1    -130   170    0     80   0    
06:51:43.745 CD2B.0    lying  1    -130   170    0     80   0    
06:51:44.746 CD2B.0    lying  1    -130   170    0     80   0    
06:51:45.744 CD2B.0    lying  1    -130   170    0     80   0    
06:51:46.748 CD2B.0    lying  1    -130   170    0     80   0    
06:51:47.754 CD2B.0    lying  1    -130   170    0     80   0    
06:51:48.747 CD2B.0    lying  1    -130   170    0     80   0    
06:51:49.749 CD2B.0    lying  1    -130   170    0     80   0    
06:51:50.752 CD2B.0    lying  1    -130   170    0     80   0    
06:51:51.752 CD2B.0    lying  1    -130   170    0     80   0    
06:51:52.753 CD2B.0    lying  1    -130   170    0     80   0    
06:51:53.760 CD2B.0    lying  1    -130   170    0     80   0    
06:51:54.765 CD2B.0    lying  1    -130   170    0     80   0    
06:51:55.650 CD2B.0    lying  1    -130   170    0     80   0    
06:51:56.657 CD2B.0    lying  1    -130   170    0     80   0    
06:51:57.649 CD2B.0    lying  1    -130   170    0     80   0    
06:51:58.653 CD2B.0    lying  1    -130   170    0     80   0    
06:51:59.651 CD2B.0    lying  1    -130   170    0     80   0    
06:52:00.680 CD2B.0    lying  1    -130   170    0     80   0    
06:52:01.657 CD2B.0    lying  1    -130   170    0     80   0    
06:52:02.653 CD2B.0    lying  1    -130   170    0     80   0    
06:52:03.660 CD2B.0    lying  1    -130   170    0     80   0    
06:52:04.656 CD2B.0    lying  1    -130   170    0     80   0    
06:52:05.660 CD2B.0    lying  1    -130   170    0     80   0    
06:52:06.661 CD2B.0    lying  1    -130   170    0     80   0    
06:52:07.560 CD2B.0    lying  1    -130   170    0     80   0    
06:52:08.556 CD2B.0    lying  1    -130   170    0     80   0    
06:52:09.560 CD2B.0    lying  1    -130   170    0     80   0    
06:52:10.554 CD2B.0    lying  1    -130   170    0     80   0    
06:52:11.627 CD2B.0    lying  1    -130   170    0     80   0    
06:52:12.573 CD2B.0    lying  1    -130   170    0     80   0    
06:52:13.573 CD2B.0    lying  1    -130   170    0     80   0    
06:52:14.585 CD2B.0    lying  1    -130   170    0     80   0    
06:52:15.574 CD2B.0    lying  1    -130   170    0     80   0    
06:52:16.581 CD2B.0    lying  1    -130   170    0     80   0    
06:52:17.472 CD2B.0    lying  1    -130   170    0     80   0    
06:52:18.474 CD2B.0    lying  1    -130   170    0     80   0    
06:52:19.474 CD2B.0    lying  1    -130   170    0     80   0    
06:52:20.492 CD2B.0    lying  1    -130   170    0     80   0    
06:52:21.476 CD2B.0    lying  1    -130   170    0     80   0    
06:52:22.501 CD2B.0    lying  1    -130   170    0     80   0    
06:52:23.476 CD2B.0    lying  1    -130   170    0     80   0    
06:52:24.485 CD2B.0    lying  1    -130   170    0     80   0    
06:52:25.481 CD2B.0    lying  1    -130   170    0     80   0    
06:52:26.489 CD2B.0    lying  1    -60    220    97    80   86   
06:52:27.488 CD2B.0    lying  1    -20    250    0     80   50   
06:52:28.485 CD2B.0    lying  1    -110   200    76    80   102  
06:52:29.377 CD2B.0    lying  1    -120   220    88    80   22   
06:52:30.380 CD2B.0    lying  1    -110   230    88    80   14   
06:52:31.380 CD2B.0    lying  1    -110   210    97    80   20   
06:52:32.385 CD2B.0    lying  1    -120   200    98    80   14   
06:52:33.382 CD2B.0    lying  1    -120   200    87    80   0    
06:52:34.381 CD2B.0    lying  1    -130   210    108   80   14   
06:52:35.384 CD2B.0    lying  1    -120   220    94    80   14   
06:52:36.388 CD2B.0    lying  1    -130   220    89    80   10   
06:52:37.384 CD2B.0    lying  1    -120   200    82    80   22   
06:52:38.393 CD2B.0    lying  1    -130   210    84    80   14   
06:52:39.388 CD2B.0    lying  1    -190   180    78    80   67   
06:52:40.394 CD2B.0    lying  1    -190   180    78    80   0    
06:52:41.284 CD2B.0    lying  1    -180   180    82    80   10   
06:52:42.288 CD2B.0    lying  1    -130   200    90    80   53   
06:52:43.288 CD2B.0    lying  1    -100   200    84    80   30   
06:52:44.289 CD2B.0    lying  1    -100   190    85    80   10   
06:52:45.287 CD2B.0    lying  1    -120   190    85    80   20   
06:52:46.296 CD2B.0    lying  1    -130   200    86    80   14   
06:52:47.289 CD2B.0    lying  1    -100   200    89    80   30   
06:52:48.292 CD2B.0    lying  1    -110   200    98    80   10   
06:52:49.292 CD2B.0    lying  1    -110   200    89    80   0    
06:52:50.300 CD2B.0    lying  1    -120   200    95    80   10   
06:52:51.292 CD2B.0    lying  1    -120   200    91    80   0    
06:52:52.300 CD2B.0    lying  1    -120   200    88    80   0    
06:52:53.188 CD2B.0    lying  1    -120   190    81    80   10   
06:52:54.186 CD2B.0    lying  1    -100   200    85    80   22   
06:52:55.189 CD2B.0    lying  1    -80    210    75    80   22   
06:52:56.192 CD2B.0    lying  1    -100   210    86    80   20   
06:52:57.192 CD2B.0    lying  1    -90    210    91    80   10   
06:52:58.194 CD2B.0    lying  1    -110   210    92    80   20   
06:52:59.192 CD2B.0    lying  1    -100   200    91    80   14   
06:53:00.230 CD2B.0    lying  1    -100   210    76    80   10   
06:53:01.208 CD2B.0    lying  1    -80    200    69    80   22   
06:53:02.205 CD2B.0    lying  1    -70    200    72    80   10   
06:53:03.111 CD2B.0    lying  1    -100   190    66    80   31   
06:53:04.106 CD2B.0    lying  1    -100   220    85    80   30   
06:53:05.114 CD2B.0    lying  1    -110   210    82    80   14   
06:53:06.116 CD2B.0    lying  1    -130   230    88    80   28   
06:53:07.114 CD2B.0    lying  1    -140   220    72    80   14   
06:53:08.124 CD2B.0    walk   1    -220   290    46    80   106  
06:53:09.130 CD2B.0    walk   1    -210   310    0     80   22   
06:53:10.112 CD2B.0    walk   1    -230   290    35    80   28   
06:53:11.172 CD2B.0    walk   1    -220   290    0     80   10   
06:53:12.116 CD2B.0    stand  1    -200   280    0     80   22   
06:53:13.118 CD2B.0    stand  1    -200   280    0     80   0    
06:53:14.014 CD2B.0    stand  1    -200   280    0     80   0    
06:53:15.018 CD2B.0    stand  1    -200   280    0     80   0    
06:53:16.024 CD2B.0    stand  1    -200   280    0     80   0    
06:53:17.022 CD2B.0    stand  1    -200   280    0     80   0    
06:53:18.018 CD2B.0    stand  1    -200   280    0     80   0    
06:53:19.020 CD2B.0    stand  1    -200   280    0     80   0    
06:53:20.024 CD2B.0    stand  1    -200   280    0     80   0    
06:53:21.024 CD2B.0    stand  1    -200   280    0     80   0    
06:53:22.028 CD2B.0    stand  1    -200   280    0     80   0    
06:53:23.024 CD2B.0    stand  1    -200   280    0     80   0    
06:53:24.024 CD2B.0    stand  1    -200   280    0     80   0    
06:53:25.028 CD2B.0    stand  1    -200   280    0     80   0    
06:53:25.920 CD2B.0    stand  1    -200   280    0     80   0    
06:53:26.929 CD2B.0    stand  1    -200   280    0     80   0    
06:53:27.921 CD2B.0    stand  1    -200   280    0     80   0    
06:53:28.952 CD2B.0    stand  1    -200   280    0     80   0    
06:53:29.924 CD2B.0    stand  1    -200   280    0     80   0    
06:53:30.924 CD2B.0    stand  1    -200   280    0     80   0    
06:53:31.928 CD2B.0    stand  1    -200   280    0     80   0    
06:53:32.926 CD2B.0    stand  1    -200   280    0     80   0    
06:53:33.930 CD2B.0    stand  1    -200   280    0     80   0    
06:53:34.930 CD2B.0    stand  1    -200   280    0     80   0    
06:53:35.929 CD2B.0    stand  1    -200   280    0     80   0    
06:53:36.932 CD2B.0    stand  1    -200   280    0     80   0    
06:53:37.821 CD2B.0    stand  1    -200   280    0     80   0    
06:53:38.824 CD2B.0    stand  1    -200   280    0     80   0    
06:53:39.828 CD2B.0    stand  1    -200   280    0     80   0    
06:53:40.824 CD2B.0    stand  1    -200   280    0     80   0    
06:53:41.826 CD2B.0    stand  1    -200   280    0     80   0    
06:53:42.828 CD2B.0    stand  1    -200   280    0     80   0    
06:53:43.846 CD2B.0    stand  1    -200   280    0     80   0    
06:53:44.888 CD2B.88   88     -    -      -      -     -    -    
06:53:45.840 CD2B.88   88     -    -      -      -     -    -    
06:53:46.840 CD2B.88   88     -    -      -      -     -    -    
06:53:56.670 CD2B.88   88     -    -      -      -     -    -    
06:54:28.673 CD2B.88   88     -    -      -      -     -    -    
06:55:00.168 CD2B.88   88     -    -      -      -     -    -    
06:55:32.265 CD2B.88   88     -    -      -      -     -    -    
06:56:03.646 CD2B.88   88     -    -      -      -     -    -    
06:56:35.572 CD2B.88   88     -    -      -      -     -    -    
06:57:07.196 CD2B.88   88     -    -      -      -     -    -    
06:57:39.140 CD2B.88   88     -    -      -      -     -    -    
06:58:11.088 CD2B.88   88     -    -      -      -     -    -    
06:58:42.378 CD2B.88   88     -    -      -      -     -    -    
06:59:14.371 CD2B.88   88     -    -      -      -     -    -    
06:59:45.917 CD2B.88   88     -    -      -      -     -    -    
07:00:17.696 CD2B.88   88     -    -      -      -     -    -    
07:00:49.402 CD2B.88   88     -    -      -      -     -    -    
07:01:21.248 CD2B.88   88     -    -      -      -     -    -    
07:01:52.888 CD2B.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 878 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

## 完整原始记录（按时间排序，data_value 全文不删字段）
```
time     ms             device.tid   event          x      y      z     原始记录
06:40:00 1781613600000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:40:00 1781613600458  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:01 1781613601000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:40:01 1781613601460  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:02 1781613602000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:40:02 1781613602463  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:03 1781613603000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:40:03 1781613603466  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:04 1781613604000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:04 1781613604468  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:05 1781613605000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:05 1781613605481  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:06 1781613606000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 83, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:06 1781613606475  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:07 1781613607000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 59, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:07 1781613607360  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:08 1781613608000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 84, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:08 1781613608363  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:09 1781613609000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 57, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:09 1781613609366  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:10 1781613610000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 83, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:10 1781613610367  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:10 1781613610995  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:11 1781613611000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 57, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:11 1781613611329  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:12 1781613612000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 82, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:12 1781613612339  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:13 1781613613000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 57, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:13 1781613613334  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:14 1781613614000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:14 1781613614334  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:15 1781613615000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 57, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:15 1781613615340  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:16 1781613616000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:16 1781613616334  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:17 1781613617000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:17 1781613617349  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613617349, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "respiratory_rate": -1, "multi_person_duration": 0}
06:40:17 1781613617349  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:40:17 1781613617391  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:18 1781613618000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:18 1781613618335  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:19 1781613619335  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:20 1781613620000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:20 1781613620000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:20 1781613620340  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:21 1781613621337  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:22 1781613622000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:22 1781613622000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:22 1781613622338  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:23 1781613623231  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:24 1781613624000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:24 1781613624000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:24 1781613624233  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:25 1781613625238  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:26 1781613626000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:26 1781613626000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:26 1781613626242  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:27 1781613627284  CD2B.0       track          -200   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:28 1781613628000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:28 1781613628000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:40:28 1781613628286  CD2B.0       track          -170   170    62    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 170, "position_z": 62, "remaining_time": 0, "track_confidence": 80}
06:40:29 1781613629282  CD2B.0       track          -140   190    65    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 190, "position_z": 65, "remaining_time": 0, "track_confidence": 80}
06:40:30 1781613630000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:30 1781613630000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:40:30 1781613630183  CD2B.0       track          -140   190    70    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 190, "position_z": 70, "remaining_time": 0, "track_confidence": 80}
06:40:31 1781613631180  CD2B.0       track          -110   200    70    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 70, "remaining_time": 0, "track_confidence": 80}
06:40:32 1781613632000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:32 1781613632000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:40:32 1781613632184  CD2B.0       track          -120   200    74    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 200, "position_z": 74, "remaining_time": 0, "track_confidence": 80}
06:40:33 1781613633189  CD2B.0       track          -110   200    75    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
06:40:34 1781613634000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:40:34 1781613634000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:40:34 1781613634183  CD2B.0       track          -90    200    73    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 200, "position_z": 73, "remaining_time": 0, "track_confidence": 80}
06:40:35 1781613635183  CD2B.0       track          -70    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:36 1781613636000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:40:36 1781613636000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:40:36 1781613636947  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:37 1781613637195  CD2B.0       track          -80    190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:38 1781613638000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:40:38 1781613638000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:40:38 1781613638190  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:39 1781613639186  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:40 1781613640000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:40 1781613640000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:40:40 1781613640187  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:41 1781613641089  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:42 1781613642000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:42 1781613642000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:40:42 1781613642093  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:42 1781613642695  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:40:42 1781613642695  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613642695, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:40:42 1781613642913  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:43 1781613643089  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:44 1781613644000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 82, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:40:44 1781613644000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:40:44 1781613644098  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:45 1781613645092  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:46 1781613646000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:40:46 1781613646000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:40:46 1781613646093  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:47 1781613647095  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:48 1781613648000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:40:48 1781613648000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 70, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:40:48 1781613648097  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:49 1781613649096  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:50 1781613650000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:40:50 1781613650000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 70, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:40:50 1781613650097  CD2B.0       track          -80    200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:51 1781613651104  CD2B.0       track          -220   160    74    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 74, "remaining_time": 0, "track_confidence": 80}
06:40:52 1781613652000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:40:52 1781613652000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 71, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:40:52 1781613652101  CD2B.0       track          -210   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -210, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:52 1781613652993  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:53 1781613653992  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:54 1781613654000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 71, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:40:54 1781613654000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:40:54 1781613654999  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:55 1781613655996  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:56 1781613656000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:40:56 1781613656996  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:57 1781613657000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 71, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:40:57 1781613657997  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:40:58 1781613658000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:40:59 1781613659000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 71, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:40:59 1781613659008  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:00 1781613660000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:41:00 1781613660012  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:01 1781613661000  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:01 1781613661000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 71, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:41:02 1781613662000  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:02 1781613662000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:41:03 1781613663000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 71, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:41:03 1781613663007  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:04 1781613664000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:41:04 1781613664009  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:04 1781613664903  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:05 1781613665000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:41:05 1781613665903  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:06 1781613666000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:06 1781613666898  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:07 1781613667000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 59, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:41:07 1781613667898  CD2B.0       track          -220   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:08 1781613668000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:08 1781613668900  CD2B.0       track          -210   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -210, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:09 1781613669000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:41:09 1781613669900  CD2B.0       track          -110   190    74    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 74, "remaining_time": 0, "track_confidence": 80}
06:41:10 1781613670000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 72, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:10 1781613670909  CD2B.0       track          -100   200    76    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 200, "position_z": 76, "remaining_time": 0, "track_confidence": 80}
06:41:11 1781613671000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:41:11 1781613671908  CD2B.0       track          -80    200    61    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 61, "remaining_time": 0, "track_confidence": 80}
06:41:12 1781613672000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 72, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:12 1781613672920  CD2B.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1781613672920, "event_status": "start", "number_people": 1, "respiratory_rate": -1}
06:41:12 1781613672936  1641         sleep-stage    -      -      -     {"heart_rate": -1, "event_since": 1781613671000, "sleep_stage": 1, "event_status": "instant", "respiratory_rate": -1}
06:41:12 1781613672958  CD2B.0       track          -100   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:13 1781613673000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:41:13 1781613673912  CD2B.0       track          -100   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:14 1781613674000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:14 1781613674432  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:14 1781613674920  CD2B.0       track          -100   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:15 1781613675000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:41:15 1781613675806  CD2B.0       track          -100   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:16 1781613676000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:16 1781613676808  CD2B.0       track          -100   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:17 1781613677000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:41:17 1781613677837  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:41:17 1781613677837  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613677837, "track_count": 1, "event_status": "instant", "lie_duration": 28, "walk_distance": 1, "walk_duration": 1, "stand_duration": 31, "respiratory_rate": -1, "multi_person_duration": 0}
06:41:17 1781613677876  CD2B.0       track          -100   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:18 1781613678000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:18 1781613678809  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:19 1781613679000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:41:19 1781613679817  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:20 1781613680000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:20 1781613680818  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:21 1781613681000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:41:21 1781613681819  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:22 1781613682000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:22 1781613682814  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:23 1781613683000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:41:23 1781613683814  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:24 1781613684000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:24 1781613684817  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:25 1781613685000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:41:25 1781613685817  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:26 1781613686000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:41:26 1781613686818  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:27 1781613687000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:41:27 1781613687720  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:28 1781613688000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:41:28 1781613688720  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:29 1781613689000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:41:29 1781613689714  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:30 1781613690000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:41:30 1781613690714  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:31 1781613691000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:41:31 1781613691722  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:32 1781613692000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:41:32 1781613692720  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:33 1781613693000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:41:33 1781613693728  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:34 1781613694000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:41:34 1781613694727  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:35 1781613695000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 57, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:41:35 1781613695725  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:36 1781613696732  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:37 1781613697000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 57, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:37 1781613697000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 83, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:41:37 1781613697724  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:38 1781613698621  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:39 1781613699000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:39 1781613699000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 84, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:41:39 1781613699626  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:40 1781613700624  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:41 1781613701000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 59, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:41 1781613701000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 83, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:41:41 1781613701626  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:42 1781613702632  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:43 1781613703000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 59, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:43 1781613703000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:41:43 1781613703628  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:44 1781613704636  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:45 1781613705000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:45 1781613705000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:41:45 1781613705632  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:46 1781613706254  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613706254, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:41:46 1781613706254  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:41:46 1781613706292  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:46 1781613706631  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:47 1781613707000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:47 1781613707000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:41:47 1781613707551  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:48 1781613708554  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:49 1781613709000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:49 1781613709000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:41:49 1781613709553  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:50 1781613710555  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:51 1781613711000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:51 1781613711000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:41:51 1781613711555  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:52 1781613712563  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:53 1781613713000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:53 1781613713000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:41:53 1781613713557  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:54 1781613714569  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:55 1781613715000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:55 1781613715000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:41:55 1781613715561  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:56 1781613716559  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:57 1781613717000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:57 1781613717000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:41:57 1781613717566  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:58 1781613718586  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:41:59 1781613719000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:41:59 1781613719000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:41:59 1781613719456  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:00 1781613720459  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:01 1781613721000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:01 1781613721000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:42:01 1781613721471  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:02 1781613722457  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:03 1781613723000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:03 1781613723000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:42:03 1781613723428  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:04 1781613724425  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:05 1781613725000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:05 1781613725000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:42:05 1781613725436  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:06 1781613726430  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:07 1781613727000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 59, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:07 1781613727000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:42:07 1781613727434  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:08 1781613728434  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:09 1781613729000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:09 1781613729000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:09 1781613729429  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:10 1781613730430  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:11 1781613731000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:11 1781613731000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:11 1781613731437  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:12 1781613732433  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:13 1781613733000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:13 1781613733000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:13 1781613733142  1641         sleep-stage    -      -      -     {"heart_rate": -1, "event_since": 1781613731000, "sleep_stage": 2, "event_status": "instant", "respiratory_rate": -1}
06:42:13 1781613733434  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:14 1781613734438  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:15 1781613735000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:15 1781613735000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:15 1781613735340  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:16 1781613736329  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:17 1781613737000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:17 1781613737000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:17 1781613737345  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:42:17 1781613737345  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613737345, "track_count": 1, "event_status": "instant", "lie_duration": 60, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:42:17 1781613737386  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:17 1781613737929  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613737929, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:42:17 1781613737929  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:42:18 1781613738150  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:18 1781613738330  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:19 1781613739000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:19 1781613739000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:19 1781613739352  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:20 1781613740352  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:21 1781613741000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:21 1781613741000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:21 1781613741364  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:22 1781613742369  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:23 1781613743000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:42:23 1781613743000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:23 1781613743361  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:24 1781613744254  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:25 1781613745000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:42:25 1781613745265  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:26 1781613746000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:26 1781613746256  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:27 1781613747000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:42:27 1781613747257  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:28 1781613748000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:28 1781613748264  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:29 1781613749000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:42:29 1781613749264  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:30 1781613750000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:30 1781613750268  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:31 1781613751000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:42:31 1781613751270  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:32 1781613752000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:32 1781613752264  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:33 1781613753000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:42:33 1781613753272  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:34 1781613754000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:34 1781613754264  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:35 1781613755000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:42:35 1781613755276  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:36 1781613756000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:36 1781613756166  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:37 1781613757000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:37 1781613757164  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:38 1781613758000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:42:38 1781613758164  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:39 1781613759000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:39 1781613759162  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:40 1781613760000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:40 1781613760164  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:41 1781613761165  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:42 1781613762000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:42 1781613762000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:42 1781613762165  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:43 1781613763166  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:44 1781613764000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:44 1781613764000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:44 1781613764173  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:45 1781613765178  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:46 1781613766000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:46 1781613766000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:42:46 1781613766173  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:47 1781613767183  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:48 1781613768000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:42:48 1781613768000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 67, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:48 1781613768068  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:49 1781613769064  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:49 1781613769660  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:50 1781613770000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:42:50 1781613770000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:50 1781613770068  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:51 1781613771072  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:52 1781613772000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:52 1781613772000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 82, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:42:52 1781613772067  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:53 1781613773068  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:54 1781613774000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:42:54 1781613774000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:54 1781613774076  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:55 1781613775071  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:56 1781613776000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:42:56 1781613776000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:42:56 1781613776077  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:57 1781613777076  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:58 1781613778000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:42:58 1781613778000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:42:58 1781613778086  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:59 1781613779076  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:42:59 1781613779969  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:00 1781613780000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:43:00 1781613780000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:00 1781613780972  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:01 1781613781976  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:02 1781613782000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:43:02 1781613782000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:02 1781613782970  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:03 1781613783421  0865         sleep-stage    -      -      -     {"heart_rate": -1, "event_since": 1781613782000, "sleep_stage": 2, "event_status": "instant", "respiratory_rate": -1}
06:43:03 1781613783974  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:04 1781613784000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:04 1781613784000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:43:04 1781613784972  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:05 1781613785974  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:06 1781613786000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:06 1781613786000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:43:06 1781613786988  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:07 1781613787988  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:08 1781613788000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:08 1781613788000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:43:08 1781613788989  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:09 1781613789886  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:10 1781613790000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:10 1781613790000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:43:10 1781613790888  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:11 1781613791905  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:12 1781613792000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:12 1781613792000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:43:12 1781613792889  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:13 1781613793894  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:14 1781613794000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:14 1781613794000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:43:14 1781613794892  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:15 1781613795908  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613795908, "track_count": 1, "event_status": "instant", "lie_duration": 60, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:43:15 1781613795908  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:43:15 1781613795944  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:16 1781613796000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:43:16 1781613796000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:16 1781613796893  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:17 1781613797901  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:18 1781613798000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:43:18 1781613798000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:18 1781613798895  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:19 1781613799896  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:20 1781613800000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:20 1781613800000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:43:20 1781613800910  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:21 1781613801477  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613801477, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:43:21 1781613801477  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:43:21 1781613801741  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:21 1781613801792  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:22 1781613802000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:22 1781613802000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:43:22 1781613802797  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:23 1781613803801  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:24 1781613804000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:24 1781613804000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:43:24 1781613804808  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:25 1781613805804  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:26 1781613806000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:43:26 1781613806000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:26 1781613806813  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:27 1781613807810  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:28 1781613808000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:28 1781613808000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:43:28 1781613808804  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:29 1781613809806  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:30 1781613810000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:30 1781613810000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:43:30 1781613810806  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:31 1781613811809  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:32 1781613812000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:43:32 1781613812000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:32 1781613812711  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:33 1781613813711  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:34 1781613814000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:34 1781613814000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:43:34 1781613814708  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:35 1781613815706  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:36 1781613816000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:36 1781613816000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:43:36 1781613816707  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:37 1781613817718  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:38 1781613818000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:38 1781613818000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:38 1781613818709  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:39 1781613819657  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:40 1781613820000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:40 1781613820000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:40 1781613820649  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:41 1781613821650  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:42 1781613822000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:42 1781613822650  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:43 1781613823000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:43 1781613823656  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:44 1781613824000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:44 1781613824653  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:45 1781613825000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:45 1781613825652  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:46 1781613826000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:46 1781613826654  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:47 1781613827000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:47 1781613827660  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:48 1781613828000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:48 1781613828661  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:49 1781613829000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:49 1781613829661  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:50 1781613830000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:50 1781613830689  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:51 1781613831000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:51 1781613831558  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:52 1781613832000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:52 1781613832557  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:53 1781613833000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:53 1781613833146  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:53 1781613833556  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:54 1781613834000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:54 1781613834555  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:55 1781613835000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:55 1781613835522  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:56 1781613836000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 83, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:43:56 1781613836521  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:57 1781613837000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:43:57 1781613837526  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:58 1781613838000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 83, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:43:58 1781613838523  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:43:59 1781613839000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:43:59 1781613839525  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:00 1781613840000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 84, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:44:00 1781613840523  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:01 1781613841000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:01 1781613841529  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:02 1781613842000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 82, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:02 1781613842532  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:03 1781613843000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:03 1781613843532  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:04 1781613844000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 82, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:04 1781613844528  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:05 1781613845000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:05 1781613845532  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:06 1781613846000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 82, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:06 1781613846529  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:07 1781613847000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:07 1781613847424  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:08 1781613848000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 83, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:08 1781613848448  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:09 1781613849000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:09 1781613849426  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:10 1781613850000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 82, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:10 1781613850427  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:11 1781613851000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:11 1781613851438  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:12 1781613852000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:12 1781613852432  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:13 1781613853000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:13 1781613853436  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:14 1781613854000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:14 1781613854433  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:15 1781613855000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:15 1781613855448  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:44:15 1781613855448  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613855448, "track_count": 1, "event_status": "instant", "lie_duration": 60, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:44:15 1781613855493  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:16 1781613856000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:16 1781613856442  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:17 1781613857000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:17 1781613857440  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:18 1781613858336  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:19 1781613859000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:19 1781613859000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:19 1781613859337  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:20 1781613860337  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:21 1781613861000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:21 1781613861000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:21 1781613861338  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:22 1781613862340  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:23 1781613863000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:23 1781613863000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:23 1781613863342  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:24 1781613864343  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:25 1781613865000  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:44:25 1781613865000  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613865000, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:44:25 1781613865000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:25 1781613865000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:25 1781613865330  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:25 1781613865358  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:26 1781613866352  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:27 1781613867000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:27 1781613867000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:27 1781613867272  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:28 1781613868265  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:29 1781613869000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:29 1781613869000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:29 1781613869269  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:30 1781613870285  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:31 1781613871000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:31 1781613871000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:31 1781613871268  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:32 1781613872272  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:33 1781613873000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:33 1781613873000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:33 1781613873269  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:34 1781613874285  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:35 1781613875000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:35 1781613875000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 70, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:35 1781613875308  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:36 1781613876272  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:37 1781613877000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:37 1781613877000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:37 1781613877273  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:38 1781613878274  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:39 1781613879000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:39 1781613879000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:39 1781613879170  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:40 1781613880172  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:41 1781613881000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:41 1781613881000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:41 1781613881168  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:42 1781613882174  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:43 1781613883000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:43 1781613883000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:43 1781613883145  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:44 1781613884137  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:45 1781613885000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:45 1781613885000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 10, "track_confidence": 90, "vital_confidence": 0}
06:44:45 1781613885146  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:46 1781613886138  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:47 1781613887000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:47 1781613887000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:44:47 1781613887145  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:48 1781613888141  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:49 1781613889000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:49 1781613889000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:49 1781613889141  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:50 1781613890146  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:51 1781613891000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:51 1781613891000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:51 1781613891149  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:52 1781613892153  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:53 1781613893000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:53 1781613893000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:53 1781613893145  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:54 1781613894145  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:55 1781613895000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:44:55 1781613895000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:55 1781613895049  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:56 1781613896045  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:56 1781613896633  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:57 1781613897000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 82, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:57 1781613897000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:44:57 1781613897040  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:58 1781613898053  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:44:59 1781613899000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:44:59 1781613899041  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:00 1781613900000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:00 1781613900066  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:01 1781613901000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:01 1781613901068  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:02 1781613902000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:02 1781613902067  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:03 1781613903000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 79, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:03 1781613903082  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:03 1781613903971  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:04 1781613904000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 59, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:04 1781613904974  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:05 1781613905000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:05 1781613905969  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:06 1781613906000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:06 1781613906977  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:07 1781613907000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 83, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:07 1781613907972  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:08 1781613908000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:08 1781613908971  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:09 1781613909000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:09 1781613909976  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:10 1781613910000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:10 1781613910980  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:11 1781613911000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:45:11 1781613911979  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:12 1781613912000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:12 1781613912982  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:13 1781613913000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:45:13 1781613913976  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:14 1781613914000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:14 1781613914991  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613914991, "track_count": 1, "event_status": "instant", "lie_duration": 60, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:45:14 1781613914991  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:45:15 1781613915000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:45:15 1781613915032  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:15 1781613915876  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:16 1781613916000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:16 1781613916879  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:17 1781613917000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:45:17 1781613917878  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:18 1781613918000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:18 1781613918881  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:19 1781613919000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:45:19 1781613919893  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:20 1781613920000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:20 1781613920885  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:21 1781613921000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:45:21 1781613921881  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:22 1781613922000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:22 1781613922893  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:23 1781613923000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:45:23 1781613923888  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:24 1781613924000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:24 1781613924888  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:25 1781613925000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:45:25 1781613925886  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:26 1781613926000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:26 1781613926793  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:27 1781613927000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 81, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:45:27 1781613927789  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:28 1781613928000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:28 1781613928393  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613928393, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:45:28 1781613928393  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:45:28 1781613928618  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:28 1781613928786  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:29 1781613929000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 83, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:45:29 1781613929791  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:30 1781613930000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:30 1781613930785  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:31 1781613931000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 84, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:45:31 1781613931741  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:32 1781613932000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:32 1781613932744  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:33 1781613933000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 84, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:45:33 1781613933744  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:34 1781613934000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:34 1781613934745  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:35 1781613935000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 82, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:45:35 1781613935746  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:36 1781613936000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 57, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:36 1781613936757  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:37 1781613937000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 80, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:37 1781613937748  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:38 1781613938000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 57, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:45:38 1781613938749  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:39 1781613939000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:39 1781613939752  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:40 1781613940000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:40 1781613940760  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:41 1781613941000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:41 1781613941756  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:42 1781613942000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:42 1781613942757  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:43 1781613943650  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:44 1781613944000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:44 1781613944000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 77, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:44 1781613944648  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:45 1781613945653  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:46 1781613946000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:46 1781613946000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:46 1781613946652  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:47 1781613947698  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:48 1781613948000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:48 1781613948000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:48 1781613948701  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:49 1781613949601  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:50 1781613950000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:50 1781613950000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:50 1781613950598  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:51 1781613951605  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:52 1781613952000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:52 1781613952000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 78, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:52 1781613952601  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:53 1781613953602  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:54 1781613954000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:54 1781613954000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:54 1781613954607  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:55 1781613955606  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:56 1781613956000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:56 1781613956000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:56 1781613956606  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:57 1781613957607  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:58 1781613958000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 71, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:45:58 1781613958000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:45:58 1781613958606  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:45:59 1781613959612  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:00 1781613960000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 72, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:46:00 1781613960000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:46:00 1781613960181  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:00 1781613960616  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:01 1781613961505  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:02 1781613962000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 72, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:02 1781613962000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:46:02 1781613962512  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:03 1781613963509  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:04 1781613964000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:46:04 1781613964000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:04 1781613964509  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:05 1781613965514  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:06 1781613966000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:06 1781613966000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:06 1781613966517  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:07 1781613967517  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:08 1781613968000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:08 1781613968000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:08 1781613968513  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:09 1781613969521  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:10 1781613970000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:10 1781613970000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:46:10 1781613970515  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:11 1781613971520  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:12 1781613972000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 75, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:12 1781613972000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:46:12 1781613972420  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:13 1781613973417  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:14 1781613974000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:14 1781613974000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:14 1781613974442  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:46:14 1781613974442  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613974442, "track_count": 1, "event_status": "instant", "lie_duration": 60, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:46:14 1781613974488  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:15 1781613975425  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:16 1781613976000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:46:16 1781613976000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:16 1781613976424  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:17 1781613977419  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:18 1781613978000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:46:18 1781613978000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:46:18 1781613978424  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:19 1781613979358  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:20 1781613980000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:20 1781613980000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:20 1781613980360  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:21 1781613981364  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:22 1781613982000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:22 1781613982000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:22 1781613982362  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:23 1781613983366  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:24 1781613984000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:24 1781613984000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:24 1781613984374  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:25 1781613985365  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:26 1781613986000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:46:26 1781613986000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:26 1781613986369  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:27 1781613987366  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:28 1781613988000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 71, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:46:28 1781613988000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:46:28 1781613988368  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:29 1781613989375  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:30 1781613990000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 72, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:46:30 1781613990000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:30 1781613990377  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:31 1781613991262  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:31 1781613991882  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:46:31 1781613991882  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781613991882, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:46:32 1781613992000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 72, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:46:32 1781613992000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:46:32 1781613992200  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:32 1781613992263  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:33 1781613993264  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:34 1781613994000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:46:34 1781613994000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:34 1781613994268  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:35 1781613995244  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:36 1781613996000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:46:36 1781613996000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:46:36 1781613996236  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:37 1781613997238  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:38 1781613998000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:46:38 1781613998000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:46:38 1781613998245  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:39 1781613999234  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:40 1781614000000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:46:40 1781614000000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:40 1781614000238  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:41 1781614001236  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:42 1781614002000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:46:42 1781614002000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:42 1781614002238  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:43 1781614003240  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:44 1781614004000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:46:44 1781614004000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:44 1781614004240  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:45 1781614005240  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:46 1781614006000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:46:46 1781614006000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:46 1781614006248  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:47 1781614007145  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:48 1781614008000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:46:48 1781614008000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:48 1781614008136  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:49 1781614009140  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:50 1781614010000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:46:50 1781614010000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:50 1781614010141  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:51 1781614011138  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:52 1781614012000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:46:52 1781614012000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:52 1781614012157  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:53 1781614013162  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:54 1781614014000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 71, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 20, "track_confidence": 90, "vital_confidence": 0}
06:46:54 1781614014000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:54 1781614014158  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:55 1781614015168  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:56 1781614016000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:46:56 1781614016000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 70, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 20, "track_confidence": 90, "vital_confidence": 0}
06:46:56 1781614016161  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:57 1781614017062  CD2B.0       track          -110   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:58 1781614018000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:46:58 1781614018000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 21, "track_confidence": 90, "vital_confidence": 0}
06:46:58 1781614018056  CD2B.0       track          -100   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:46:59 1781614019056  CD2B.0       track          -120   210    75    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 210, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
06:47:00 1781614020000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 74, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 21, "track_confidence": 90, "vital_confidence": 0}
06:47:00 1781614020000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:00 1781614020058  CD2B.0       track          -120   190    79    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 190, "position_z": 79, "remaining_time": 0, "track_confidence": 80}
06:47:01 1781614021063  CD2B.0       track          -120   190    70    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 190, "position_z": 70, "remaining_time": 0, "track_confidence": 80}
06:47:02 1781614022000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:02 1781614022000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:47:02 1781614022061  CD2B.0       track          -110   200    70    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 70, "remaining_time": 0, "track_confidence": 80}
06:47:03 1781614023064  CD2B.0       track          -110   200    71    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 71, "remaining_time": 0, "track_confidence": 80}
06:47:03 1781614023696  333B.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1781614023696, "event_status": "start", "number_people": 0, "respiratory_rate": -1}
06:47:03 1781614023842  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:04 1781614024000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 20, "track_confidence": 90, "vital_confidence": 0}
06:47:04 1781614024067  CD2B.0       track          -80    200    77    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 77, "remaining_time": 0, "track_confidence": 80}
06:47:05 1781614025000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:05 1781614025062  CD2B.0       track          -150   180    67    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -150, "position_y": 180, "position_z": 67, "remaining_time": 0, "track_confidence": 80}
06:47:06 1781614026000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:47:06 1781614026064  CD2B.0       track          -140   190    67    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 190, "position_z": 67, "remaining_time": 0, "track_confidence": 80}
06:47:07 1781614027000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:07 1781614027076  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:07 1781614027966  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:08 1781614028000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:08 1781614028967  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:09 1781614029000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:09 1781614029969  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:10 1781614030000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:10 1781614030968  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:11 1781614031000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:47:11 1781614031981  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:12 1781614032000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:12 1781614032974  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:13 1781614033000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:47:13 1781614033984  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:47:13 1781614033984  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614033984, "track_count": 1, "event_status": "instant", "lie_duration": 60, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:47:14 1781614034000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:47:14 1781614034024  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:14 1781614034972  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:15 1781614035000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:47:15 1781614035000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:15 1781614035976  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:16 1781614036976  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:17 1781614037000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:47:17 1781614037000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:17 1781614037977  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:18 1781614038977  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:19 1781614039000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:47:19 1781614039000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:19 1781614039870  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:20 1781614040870  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:21 1781614041000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 11, "track_confidence": 90, "vital_confidence": 0}
06:47:21 1781614041000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:21 1781614041876  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:22 1781614042872  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:23 1781614043000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:23 1781614043000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 12, "track_confidence": 90, "vital_confidence": 0}
06:47:23 1781614043837  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:24 1781614044839  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:25 1781614045000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:25 1781614045000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:47:25 1781614045841  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:26 1781614046841  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:27 1781614047000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:27 1781614047000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:47:27 1781614047848  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:28 1781614048848  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:29 1781614049000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:29 1781614049000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:47:29 1781614049853  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:30 1781614050848  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:31 1781614051000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:31 1781614051000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 67, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:47:31 1781614051848  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:32 1781614052846  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:33 1781614053000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:33 1781614053000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:47:33 1781614053848  CD2B.0       track          -110   200    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:34 1781614054854  CD2B.0       track          -130   190    70    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 190, "position_z": 70, "remaining_time": 0, "track_confidence": 80}
06:47:35 1781614055000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:35 1781614055000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:35 1781614055449  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:47:35 1781614055449  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614055449, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:47:35 1781614055526  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:35 1781614055746  CD2B.0       track          -150   180    64    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -150, "position_y": 180, "position_z": 64, "remaining_time": 0, "track_confidence": 80}
06:47:36 1781614056749  CD2B.0       track          -130   190    66    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 190, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:47:37 1781614057000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:37 1781614057000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:37 1781614057756  CD2B.0       track          -80    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:38 1781614058768  CD2B.0       track          -80    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:39 1781614059000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:47:39 1781614059000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:39 1781614059785  CD2B.0       track          -80    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:40 1781614060797  CD2B.0       track          -80    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:41 1781614061000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:47:41 1781614061000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:41 1781614061782  CD2B.0       track          -80    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:42 1781614062780  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:42 1781614062907  1641         sleep-stage    -      -      -     {"heart_rate": -1, "event_since": 1781614061000, "sleep_stage": 1, "event_status": "instant", "respiratory_rate": -1}
06:47:43 1781614063000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 20, "track_confidence": 90, "vital_confidence": 0}
06:47:43 1781614063680  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:44 1781614064000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:47:44 1781614064680  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:45 1781614065000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:47:45 1781614065000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 20, "track_confidence": 90, "vital_confidence": 0}
06:47:45 1781614065686  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:46 1781614066000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:47:46 1781614066684  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:47 1781614067000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 20, "track_confidence": 90, "vital_confidence": 0}
06:47:47 1781614067689  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:48 1781614068000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:47:48 1781614068684  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:49 1781614069000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 20, "track_confidence": 90, "vital_confidence": 0}
06:47:49 1781614069686  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:50 1781614070000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:47:50 1781614070685  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:51 1781614071000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:47:51 1781614071692  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:52 1781614072000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:52 1781614072688  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:53 1781614073000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:47:53 1781614073690  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:54 1781614074000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:54 1781614074700  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:55 1781614075000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:47:55 1781614075597  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:56 1781614076000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:56 1781614076600  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:57 1781614077000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:47:57 1781614077592  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:58 1781614078000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:47:58 1781614078594  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:47:59 1781614079000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:47:59 1781614079593  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:00 1781614080000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:00 1781614080594  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:01 1781614081000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:01 1781614081596  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:02 1781614082000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:02 1781614082596  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:03 1781614083000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:03 1781614083604  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:04 1781614084000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:04 1781614084605  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:05 1781614085000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:05 1781614085609  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:06 1781614086000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:06 1781614086495  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:07 1781614087000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:07 1781614087102  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:07 1781614087496  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:08 1781614088000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:08 1781614088500  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:09 1781614089000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 59, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:09 1781614089504  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:10 1781614090000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:10 1781614090498  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:11 1781614091000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:11 1781614091508  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:12 1781614092506  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:13 1781614093000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:13 1781614093000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:13 1781614093514  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:48:13 1781614093514  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614093514, "track_count": 1, "event_status": "instant", "lie_duration": 60, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:48:13 1781614093550  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:14 1781614094000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:14 1781614094505  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:15 1781614095502  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:16 1781614096000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:16 1781614096000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:48:16 1781614096506  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:17 1781614097507  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:18 1781614098000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:18 1781614098000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:48:18 1781614098398  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:19 1781614099402  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:20 1781614100000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:20 1781614100000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:48:20 1781614100400  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:21 1781614101401  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:22 1781614102000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:22 1781614102000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:48:22 1781614102403  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:23 1781614103410  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:24 1781614104000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:24 1781614104000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:24 1781614104404  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:25 1781614105413  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:26 1781614106000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:26 1781614106000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:26 1781614106410  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:27 1781614107422  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:28 1781614108000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:28 1781614108000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:28 1781614108412  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:29 1781614109316  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:30 1781614110000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:30 1781614110000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:30 1781614110312  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:31 1781614111312  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:32 1781614112000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:32 1781614112000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 67, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:32 1781614112315  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:33 1781614113314  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:34 1781614114000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:48:34 1781614114321  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:35 1781614115000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:35 1781614115334  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:36 1781614116000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:36 1781614116331  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:37 1781614117000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:37 1781614117322  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:38 1781614118000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:48:38 1781614118321  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:38 1781614118916  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:48:38 1781614118916  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614118916, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:48:39 1781614119000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:39 1781614119177  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:39 1781614119321  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:40 1781614120000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:40 1781614120000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 76, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:40 1781614120328  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:41 1781614121000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 73, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:41 1781614121220  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:42 1781614122000  0865.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 1, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:42 1781614122000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:42 1781614122221  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:42 1781614122943  333B.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1781614122943, "event_status": "start", "number_people": 1, "respiratory_rate": -1}
06:48:42 1781614122980  333B         EnterRoom      -      -      -     {"heart_rate": -1, "event_since": 1781614122980, "event_status": "start", "respiratory_rate": -1}
06:48:42 1781614122982  333B.0       track          -20    190    88    {"pose": 4, "event": 1, "area_id": 1, "track_id": 0, "position_x": -20, "position_y": 190, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
06:48:43 1781614123237  CD2B.0       track          -90    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:43 1781614123806  333B.0       track          -70    200    85    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 200, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:48:44 1781614124000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:44 1781614124229  CD2B.0       track          -120   190    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:44 1781614124237  0865         sleep-stage    -      -      -     {"heart_rate": -1, "event_since": 1781614122000, "sleep_stage": 1, "event_status": "instant", "respiratory_rate": -1}
06:48:44 1781614124237  0865         LeftBed        -      -      -     {"bed_status": 1, "heart_rate": -1, "event_since": 1781614122000, "event_status": "instant", "respiratory_rate": -1}
06:48:44 1781614124238  0865         LeftBed        -      -      -     {"bed_status": 1, "heart_rate": -1, "event_since": 1781614122000, "event_status": "start", "respiratory_rate": -1}
06:48:44 1781614124808  333B.0       track          -130   140    85    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 140, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:48:45 1781614125238  CD2B.0       track          -180   170    68    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 170, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
06:48:45 1781614125808  333B.0       track          -120   170    68    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 170, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
06:48:46 1781614126000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:46 1781614126233  CD2B.0       track          -160   170    68    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -160, "position_y": 170, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
06:48:46 1781614126808  333B.0       track          -130   250    80    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 250, "position_z": 80, "remaining_time": 0, "track_confidence": 80}
06:48:47 1781614127239  CD2B.0       track          -230   140    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -230, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:47 1781614127812  333B.0       track          -140   260    80    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 260, "position_z": 80, "remaining_time": 0, "track_confidence": 80}
06:48:48 1781614128000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:48 1781614128233  CD2B.0       track          -230   140    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -230, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:48 1781614128812  333B.0       track          -200   230    98    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 230, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
06:48:49 1781614129235  CD2B.0       track          -220   140    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:49 1781614129817  333B.0       track          -200   160    80    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 160, "position_z": 80, "remaining_time": 0, "track_confidence": 80}
06:48:50 1781614130000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:48:50 1781614130240  CD2B.0       track          -220   140    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:50 1781614130813  333B.0       track          -210   120    89    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -210, "position_y": 120, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
06:48:51 1781614131134  CD2B.0       track          -220   140    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:51 1781614131819  333B.0       track          -190   180    85    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 180, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:48:52 1781614132000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:52 1781614132137  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:52 1781614132822  333B.0       track          -190   260    91    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 260, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
06:48:53 1781614133137  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:53 1781614133821  333B.0       track          -180   280    120   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 120, "remaining_time": 0, "track_confidence": 80}
06:48:54 1781614134000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:48:54 1781614134137  CD2B.0       track          -210   160    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -210, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:54 1781614134717  333B.0       track          -190   270    112   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 270, "position_z": 112, "remaining_time": 0, "track_confidence": 80}
06:48:55 1781614135140  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:55 1781614135731  333B.0       track          -180   270    109   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 270, "position_z": 109, "remaining_time": 0, "track_confidence": 80}
06:48:56 1781614136000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:56 1781614136140  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:56 1781614136720  333B.0       track          -190   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:57 1781614137140  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:57 1781614137720  333B.0       track          -190   270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:58 1781614138000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:48:58 1781614138144  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:58 1781614138734  333B.0       track          -180   280    92    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 92, "remaining_time": 0, "track_confidence": 80}
06:48:59 1781614139146  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:48:59 1781614139722  333B.0       track          -140   290    96    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 290, "position_z": 96, "remaining_time": 0, "track_confidence": 80}
06:49:00 1781614140000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:49:00 1781614140145  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:00 1781614140733  333B.0       track          -130   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:01 1781614141146  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:01 1781614141726  333B.0       track          -140   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:02 1781614142000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:49:02 1781614142148  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:02 1781614142724  333B.0       track          -120   300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:03 1781614143040  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:03 1781614143728  333B.0       track          -100   310    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 310, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:04 1781614144000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:49:04 1781614144040  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:04 1781614144726  333B.0       track          -110   310    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 310, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:05 1781614145043  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:05 1781614145628  333B.0       track          -110   310    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 310, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:06 1781614146000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:49:06 1781614146054  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:06 1781614146627  333B.0       track          -110   310    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 310, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:07 1781614147056  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:07 1781614147628  333B.0       track          -150   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -150, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:08 1781614148000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:49:08 1781614148045  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:08 1781614148630  333B.0       track          -140   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:09 1781614149045  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:09 1781614149632  333B.0       track          -140   290    110   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 290, "position_z": 110, "remaining_time": 0, "track_confidence": 80}
06:49:10 1781614150000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:49:10 1781614150045  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:10 1781614150636  333B.0       track          -150   280    105   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -150, "position_y": 280, "position_z": 105, "remaining_time": 0, "track_confidence": 80}
06:49:11 1781614151052  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:11 1781614151632  333B.0       track          -190   270    95    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 270, "position_z": 95, "remaining_time": 0, "track_confidence": 80}
06:49:12 1781614152000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:49:12 1781614152053  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:12 1781614152645  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614152645, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 4, "walk_duration": 12, "stand_duration": 17, "respiratory_rate": -1, "multi_person_duration": 0}
06:49:12 1781614152645  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:49:12 1781614152682  333B.0       track          -220   270    121   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 270, "position_z": 121, "remaining_time": 0, "track_confidence": 80}
06:49:13 1781614153060  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614153060, "track_count": 1, "event_status": "instant", "lie_duration": 35, "walk_distance": 0, "walk_duration": 1, "stand_duration": 24, "respiratory_rate": -1, "multi_person_duration": 0}
06:49:13 1781614153060  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:49:13 1781614153096  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:13 1781614153635  333B.0       track          -190   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:14 1781614154000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 19, "track_confidence": 90, "vital_confidence": 0}
06:49:14 1781614154052  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:14 1781614154568  333B.0       track          -180   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:14 1781614154948  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:15 1781614155556  333B.0       track          -180   260    112   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 260, "position_z": 112, "remaining_time": 0, "track_confidence": 80}
06:49:15 1781614155944  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:16 1781614156000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:49:16 1781614156558  333B.0       track          -180   280    119   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 119, "remaining_time": 0, "track_confidence": 80}
06:49:16 1781614156952  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:17 1781614157560  333B.0       track          -180   270    112   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 270, "position_z": 112, "remaining_time": 0, "track_confidence": 80}
06:49:17 1781614157944  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:18 1781614158000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:49:18 1781614158560  333B.0       track          -180   270    106   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 270, "position_z": 106, "remaining_time": 0, "track_confidence": 80}
06:49:18 1781614158947  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:19 1781614159564  333B.0       track          -180   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:19 1781614159954  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:20 1781614160000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:49:20 1781614160566  333B.0       track          -180   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:20 1781614160958  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:21 1781614161564  333B.0       track          -180   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:21 1781614161956  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:22 1781614162567  333B.0       track          -180   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:22 1781614162956  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:23 1781614163000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:49:23 1781614163568  333B.0       track          -210   270    104   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -210, "position_y": 270, "position_z": 104, "remaining_time": 0, "track_confidence": 80}
06:49:23 1781614163950  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:24 1781614164568  333B.0       track          -180   270    101   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 270, "position_z": 101, "remaining_time": 0, "track_confidence": 80}
06:49:24 1781614164953  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:25 1781614165000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:49:25 1781614165569  333B.0       track          -190   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:25 1781614165954  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:26 1781614166461  333B.0       track          -190   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:26 1781614166853  CD2B.0       track          -220   150    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 150, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:27 1781614167000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:49:27 1781614167466  333B.0       track          -190   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:27 1781614167850  CD2B.0       track          -220   130    67    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 130, "position_z": 67, "remaining_time": 0, "track_confidence": 80}
06:49:28 1781614168460  333B.0       track          -190   260    95    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 260, "position_z": 95, "remaining_time": 0, "track_confidence": 80}
06:49:28 1781614168852  CD2B.0       track          -150   90     61    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -150, "position_y": 90, "position_z": 61, "remaining_time": 0, "track_confidence": 80}
06:49:29 1781614169000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:49:29 1781614169465  333B.0       track          -180   270    90    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 270, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
06:49:29 1781614169852  CD2B.0       track          -100   140    66    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 140, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:49:30 1781614170506  333B.0       track          -170   270    104   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 270, "position_z": 104, "remaining_time": 0, "track_confidence": 80}
06:49:30 1781614170852  CD2B.0       track          -110   140    64    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 140, "position_z": 64, "remaining_time": 0, "track_confidence": 80}
06:49:31 1781614171000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:31 1781614171509  333B.0       track          -170   250    90    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 250, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
06:49:31 1781614171872  CD2B.0       track          -100   140    63    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 140, "position_z": 63, "remaining_time": 0, "track_confidence": 80}
06:49:32 1781614172509  333B.0       track          -190   180    81    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 180, "position_z": 81, "remaining_time": 0, "track_confidence": 80}
06:49:32 1781614172860  CD2B.0       track          -100   150    62    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 150, "position_z": 62, "remaining_time": 0, "track_confidence": 80}
06:49:33 1781614173000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:33 1781614173422  333B.0       track          -200   120    71    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 120, "position_z": 71, "remaining_time": 0, "track_confidence": 80}
06:49:33 1781614173863  CD2B.0       track          -110   190    68    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 190, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
06:49:34 1781614174418  333B.0       track          -190   110    63    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 110, "position_z": 63, "remaining_time": 0, "track_confidence": 80}
06:49:34 1781614174862  CD2B.0       track          -120   200    70    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 200, "position_z": 70, "remaining_time": 0, "track_confidence": 80}
06:49:35 1781614175000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:35 1781614175460  333B.0       track          -180   190    89    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 190, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
06:49:35 1781614175864  CD2B.0       track          -120   180    68    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 180, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
06:49:36 1781614176409  333B.0       track          -180   270    84    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 270, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
06:49:36 1781614176872  CD2B.0       track          -130   190    75    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 190, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
06:49:37 1781614177000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:37 1781614177412  333B.0       track          -180   280    135   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 135, "remaining_time": 0, "track_confidence": 80}
06:49:37 1781614177769  CD2B.0       track          -120   180    68    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 180, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
06:49:38 1781614178414  333B.0       track          -190   260    129   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 260, "position_z": 129, "remaining_time": 0, "track_confidence": 80}
06:49:38 1781614178764  CD2B.0       track          -120   180    74    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 180, "position_z": 74, "remaining_time": 0, "track_confidence": 80}
06:49:39 1781614179000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:39 1781614179413  333B.0       track          -200   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:39 1781614179765  CD2B.0       track          -120   170    67    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 170, "position_z": 67, "remaining_time": 0, "track_confidence": 80}
06:49:40 1781614180414  333B.0       track          -180   270    136   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 270, "position_z": 136, "remaining_time": 0, "track_confidence": 80}
06:49:40 1781614180782  CD2B.0       track          -120   180    66    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 180, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:49:41 1781614181000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:41 1781614181414  333B.0       track          -170   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:41 1781614181766  CD2B.0       track          -110   160    71    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 160, "position_z": 71, "remaining_time": 0, "track_confidence": 80}
06:49:42 1781614182417  333B.0       track          -180   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:42 1781614182769  CD2B.0       track          -110   180    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 180, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:43 1781614183000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:43 1781614183425  333B.0       track          -200   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:43 1781614183768  CD2B.0       track          -90    160    60    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 160, "position_z": 60, "remaining_time": 0, "track_confidence": 80}
06:49:44 1781614184326  333B.0       track          -200   240    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:44 1781614184776  CD2B.0       track          -120   150    65    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 150, "position_z": 65, "remaining_time": 0, "track_confidence": 80}
06:49:45 1781614185000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:45 1781614185318  333B.0       track          -170   280    146   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 280, "position_z": 146, "remaining_time": 0, "track_confidence": 80}
06:49:45 1781614185770  CD2B.0       track          -100   140    66    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 140, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:49:46 1781614186318  333B.0       track          -180   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:46 1781614186771  CD2B.0       track          -100   150    65    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 150, "position_z": 65, "remaining_time": 0, "track_confidence": 80}
06:49:47 1781614187000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:47 1781614187325  333B.0       track          -180   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:47 1781614187682  CD2B.0       track          -80    190    66    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 190, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:49:48 1781614188328  333B.0       track          -180   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:48 1781614188684  CD2B.0       track          -80    170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:49 1781614189000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:49 1781614189329  333B.0       track          -180   280    140   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 140, "remaining_time": 0, "track_confidence": 80}
06:49:49 1781614189687  CD2B.0       track          -90    170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:50 1781614190331  333B.0       track          -190   270    102   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 270, "position_z": 102, "remaining_time": 0, "track_confidence": 80}
06:49:50 1781614190684  CD2B.0       track          -120   160    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:51 1781614191000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:51 1781614191331  333B.0       track          -180   270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:51 1781614191684  CD2B.0       track          -110   140    59    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 140, "position_z": 59, "remaining_time": 0, "track_confidence": 80}
06:49:52 1781614192332  333B.0       track          -190   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:52 1781614192688  CD2B.0       track          -120   140    61    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 140, "position_z": 61, "remaining_time": 0, "track_confidence": 80}
06:49:53 1781614193000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:49:53 1781614193340  333B.0       track          -170   280    126   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 280, "position_z": 126, "remaining_time": 0, "track_confidence": 80}
06:49:53 1781614193692  CD2B.0       track          -90    140    63    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 140, "position_z": 63, "remaining_time": 0, "track_confidence": 80}
06:49:54 1781614194331  333B.0       track          -190   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:54 1781614194688  CD2B.0       track          -90    130    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 130, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:55 1781614195000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:49:55 1781614195228  333B.0       track          -190   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:55 1781614195689  CD2B.0       track          -90    130    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 130, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:56 1781614196232  333B.0       track          -180   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:56 1781614196702  CD2B.0       track          -90    130    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 130, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:57 1781614197000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:49:57 1781614197232  333B.0       track          -180   270    117   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 270, "position_z": 117, "remaining_time": 0, "track_confidence": 80}
06:49:57 1781614197693  CD2B.0       track          -100   140    67    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 140, "position_z": 67, "remaining_time": 0, "track_confidence": 80}
06:49:58 1781614198242  333B.0       track          -200   270    114   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 270, "position_z": 114, "remaining_time": 0, "track_confidence": 80}
06:49:58 1781614198598  CD2B.0       track          -90    160    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:59 1781614199000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:49:59 1781614199232  333B.0       track          -190   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:49:59 1781614199594  CD2B.0       track          -130   150    58    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 150, "position_z": 58, "remaining_time": 0, "track_confidence": 80}
06:50:00 1781614200000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:00 1781614200235  333B.0       track          -160   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -160, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:00 1781614200596  CD2B.0       track          -130   140    64    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 140, "position_z": 64, "remaining_time": 0, "track_confidence": 80}
06:50:01 1781614201000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:01 1781614201236  333B.0       track          -170   270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:01 1781614201599  CD2B.0       track          -100   150    60    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 150, "position_z": 60, "remaining_time": 0, "track_confidence": 80}
06:50:02 1781614202240  333B.0       track          -180   280    111   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 111, "remaining_time": 0, "track_confidence": 80}
06:50:02 1781614202595  CD2B.0       track          -100   120    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:03 1781614203000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:03 1781614203170  333B.0       track          -160   260    85    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -160, "position_y": 260, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:50:03 1781614203558  CD2B.0       track          -110   120    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:04 1781614204168  333B.0       track          -120   270    93    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 270, "position_z": 93, "remaining_time": 0, "track_confidence": 80}
06:50:04 1781614204554  CD2B.0       track          -110   120    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:05 1781614205000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:05 1781614205181  333B.0       track          -100   270    66    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:50:05 1781614205560  CD2B.0       track          -110   120    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:06 1781614206185  333B.0       track          -120   260    98    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 260, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
06:50:06 1781614206557  CD2B.0       track          -110   120    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 120, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:07 1781614207000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:07 1781614207172  333B.0       track          -150   180    79    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -150, "position_y": 180, "position_z": 79, "remaining_time": 0, "track_confidence": 80}
06:50:07 1781614207556  CD2B.0       track          -170   140    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:08 1781614208000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:08 1781614208181  333B.0       track          -190   120    82    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 120, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:50:08 1781614208556  CD2B.0       track          -230   160    61    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -230, "position_y": 160, "position_z": 61, "remaining_time": 0, "track_confidence": 80}
06:50:09 1781614209000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:09 1781614209182  333B.0       track          -170   140    74    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 140, "position_z": 74, "remaining_time": 0, "track_confidence": 80}
06:50:09 1781614209556  CD2B.0       track          -210   180    81    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -210, "position_y": 180, "position_z": 81, "remaining_time": 0, "track_confidence": 80}
06:50:10 1781614210177  333B.0       track          -110   220    78    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 220, "position_z": 78, "remaining_time": 0, "track_confidence": 80}
06:50:10 1781614210566  CD2B.0       track          -170   170    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:11 1781614211000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:11 1781614211177  333B.0       track          -90    270    81    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 81, "remaining_time": 0, "track_confidence": 80}
06:50:11 1781614211997  CD2B.0       track          -120   190    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:12 1781614212194  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614212194, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 4, "walk_duration": 10, "stand_duration": 50, "respiratory_rate": -1, "multi_person_duration": 0}
06:50:12 1781614212194  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:50:12 1781614212232  333B.0       track          -80    280    112   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 280, "position_z": 112, "remaining_time": 0, "track_confidence": 80}
06:50:12 1781614212575  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614212575, "track_count": 1, "event_status": "instant", "lie_duration": 32, "walk_distance": 2, "walk_duration": 5, "stand_duration": 23, "respiratory_rate": -1, "multi_person_duration": 0}
06:50:12 1781614212575  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:50:12 1781614212773  CD2B.0       track          -90    200    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:13 1781614213183  333B.0       track          -90    290    111   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 290, "position_z": 111, "remaining_time": 0, "track_confidence": 80}
06:50:13 1781614213560  CD2B.0       track          -120   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:14 1781614214000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:14 1781614214081  333B.0       track          -90    280    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:14 1781614214571  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:15 1781614215088  333B.0       track          -100   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:15 1781614215456  CD2B.0       track          -230   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -230, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:16 1781614216000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:16 1781614216076  333B.0       track          -100   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:16 1781614216457  CD2B.0       track          -230   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -230, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:17 1781614217077  333B.0       track          -110   270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:17 1781614217456  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:18 1781614218000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:18 1781614218078  333B.0       track          -100   280    134   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 134, "remaining_time": 0, "track_confidence": 80}
06:50:18 1781614218462  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:19 1781614219040  333B.0       track          -100   280    138   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 138, "remaining_time": 0, "track_confidence": 80}
06:50:19 1781614219495  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:20 1781614220000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:20 1781614220037  333B.0       track          -90    270    102   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 102, "remaining_time": 0, "track_confidence": 80}
06:50:20 1781614220494  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:21 1781614221041  333B.0       track          -100   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:21 1781614221504  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:22 1781614222000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:22 1781614222057  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:22 1781614222498  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:23 1781614223044  333B.0       track          -90    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:23 1781614223393  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:24 1781614224000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:24 1781614224041  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:24 1781614224394  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:25 1781614225040  333B.0       track          -80    270    96    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 270, "position_z": 96, "remaining_time": 0, "track_confidence": 80}
06:50:25 1781614225398  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:26 1781614226000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:50:26 1781614226043  333B.0       track          -110   270    113   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 270, "position_z": 113, "remaining_time": 0, "track_confidence": 80}
06:50:26 1781614226397  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:27 1781614227044  333B.0       track          -100   270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:27 1781614227394  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:28 1781614228000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:50:28 1781614228052  333B.0       track          -80    290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:28 1781614228396  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:29 1781614229052  333B.0       track          -100   270    100   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 100, "remaining_time": 0, "track_confidence": 80}
06:50:29 1781614229404  CD2B.0       track          -220   170    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:30 1781614230000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:50:30 1781614230049  333B.0       track          -100   280    111   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 111, "remaining_time": 0, "track_confidence": 80}
06:50:30 1781614230408  CD2B.0       track          -60    140    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:30 1781614230941  333B.0       track          -100   270    119   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 119, "remaining_time": 0, "track_confidence": 80}
06:50:31 1781614231408  CD2B.0       track          -50    140    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -50, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:31 1781614231941  333B.0       track          -90    270    92    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 92, "remaining_time": 0, "track_confidence": 80}
06:50:32 1781614232000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:32 1781614232400  CD2B.0       track          -60    140    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:32 1781614232945  333B.0       track          -80    210    82    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 210, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:50:33 1781614233404  CD2B.0       track          -60    140    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 140, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:33 1781614233944  333B.0       track          -30    150    91    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -30, "position_y": 150, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
06:50:34 1781614234000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:34 1781614234401  CD2B.0       track          -60    150    73    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 150, "position_z": 73, "remaining_time": 0, "track_confidence": 80}
06:50:34 1781614234952  333B.0       track          -10    100    108   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -10, "position_y": 100, "position_z": 108, "remaining_time": 0, "track_confidence": 80}
06:50:35 1781614235320  CD2B.0       track          -90    140    66    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 140, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:50:35 1781614235946  333B.0       track          -20    120    107   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -20, "position_y": 120, "position_z": 107, "remaining_time": 0, "track_confidence": 80}
06:50:36 1781614236000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:36 1781614236312  CD2B.0       track          -130   130    65    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 130, "position_z": 65, "remaining_time": 0, "track_confidence": 80}
06:50:36 1781614236949  333B.0       track          -10    120    81    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -10, "position_y": 120, "position_z": 81, "remaining_time": 0, "track_confidence": 80}
06:50:37 1781614237326  CD2B.0       track          -140   130    57    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 130, "position_z": 57, "remaining_time": 0, "track_confidence": 80}
06:50:37 1781614237949  333B.0       track          -10    120    103   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -10, "position_y": 120, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
06:50:38 1781614238000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:38 1781614238312  CD2B.0       track          -100   120    58    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 120, "position_z": 58, "remaining_time": 0, "track_confidence": 80}
06:50:38 1781614238948  333B.0       track          -10    130    102   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -10, "position_y": 130, "position_z": 102, "remaining_time": 0, "track_confidence": 80}
06:50:39 1781614239313  CD2B.0       track          -90    130    47    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 130, "position_z": 47, "remaining_time": 0, "track_confidence": 80}
06:50:39 1781614239952  333B.0       track          -40    180    71    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -40, "position_y": 180, "position_z": 71, "remaining_time": 0, "track_confidence": 80}
06:50:40 1781614240000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:40 1781614240320  CD2B.0       track          -120   140    48    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 140, "position_z": 48, "remaining_time": 0, "track_confidence": 80}
06:50:40 1781614240957  333B.0       track          -80    230    82    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 230, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:50:41 1781614241318  CD2B.0       track          -80    140    55    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 140, "position_z": 55, "remaining_time": 0, "track_confidence": 80}
06:50:41 1781614241956  333B.0       track          -90    280    89    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 280, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
06:50:42 1781614242000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:42 1781614242320  CD2B.0       track          -70    140    62    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 140, "position_z": 62, "remaining_time": 0, "track_confidence": 80}
06:50:42 1781614242845  333B.0       track          -80    290    119   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 290, "position_z": 119, "remaining_time": 0, "track_confidence": 80}
06:50:43 1781614243320  CD2B.0       track          -80    120    48    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 120, "position_z": 48, "remaining_time": 0, "track_confidence": 80}
06:50:43 1781614243856  333B.0       track          -90    280    107   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 280, "position_z": 107, "remaining_time": 0, "track_confidence": 80}
06:50:44 1781614244000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:44 1781614244320  CD2B.0       track          -80    120    63    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 120, "position_z": 63, "remaining_time": 0, "track_confidence": 80}
06:50:44 1781614244851  333B.0       track          -80    280    100   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 280, "position_z": 100, "remaining_time": 0, "track_confidence": 80}
06:50:45 1781614245225  CD2B.0       track          -70    110    73    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 110, "position_z": 73, "remaining_time": 0, "track_confidence": 80}
06:50:45 1781614245856  333B.0       track          -70    280    97    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 280, "position_z": 97, "remaining_time": 0, "track_confidence": 80}
06:50:46 1781614246000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:46 1781614246218  CD2B.0       track          -110   90     71    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 90, "position_z": 71, "remaining_time": 0, "track_confidence": 80}
06:50:46 1781614246850  333B.0       track          -70    280    102   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 280, "position_z": 102, "remaining_time": 0, "track_confidence": 80}
06:50:47 1781614247220  CD2B.0       track          -110   0      77    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 0, "position_z": 77, "remaining_time": 0, "track_confidence": 80}
06:50:47 1781614247855  333B.0       track          -100   280    116   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 116, "remaining_time": 0, "track_confidence": 80}
06:50:48 1781614248000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:50:48 1781614248225  CD2B.0       track          -110   -10    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:48 1781614248856  333B.0       track          -80    280    83    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 280, "position_z": 83, "remaining_time": 0, "track_confidence": 80}
06:50:49 1781614249225  CD2B.0       track          -120   190    66    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 190, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:50:49 1781614249862  333B.0       track          -130   250    76    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 250, "position_z": 76, "remaining_time": 0, "track_confidence": 80}
06:50:50 1781614250000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:50 1781614250224  CD2B.0       track          -120   180    68    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 180, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
06:50:50 1781614250787  333B.0       track          -160   280    84    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -160, "position_y": 280, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
06:50:51 1781614251000  1641.0       track          -      -      -     {"pose": 9, "track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:51 1781614251222  CD2B.0       track          -130   170    82    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:50:51 1781614251785  333B.0       track          -130   290    101   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 290, "position_z": 101, "remaining_time": 0, "track_confidence": 80}
06:50:52 1781614252000  1641.0       track          -      -      -     {"pose": 9, "track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:52 1781614252161  CD2B.0       track          -120   180    75    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 180, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
06:50:52 1781614252786  333B.0       track          -120   280    75    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 280, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
06:50:53 1781614253173  CD2B.0       track          -120   170    81    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 170, "position_z": 81, "remaining_time": 0, "track_confidence": 80}
06:50:53 1781614253791  333B.0       track          -100   280    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:50:54 1781614254000  1641.0       track          -      -      -     {"pose": 9, "track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:54 1781614254162  CD2B.0       track          -160   160    71    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -160, "position_y": 160, "position_z": 71, "remaining_time": 0, "track_confidence": 80}
06:50:54 1781614254789  333B.0       track          -180   280    114   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 280, "position_z": 114, "remaining_time": 0, "track_confidence": 80}
06:50:55 1781614255171  CD2B.0       track          -120   170    85    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 170, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:50:55 1781614255800  333B.0       track          -210   260    85    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -210, "position_y": 260, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:50:56 1781614256000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:56 1781614256165  CD2B.0       track          -110   180    84    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 180, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
06:50:56 1781614256795  333B.0       track          -150   270    82    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -150, "position_y": 270, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:50:57 1781614257169  CD2B.0       track          -120   190    75    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 190, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
06:50:57 1781614257792  333B.0       track          -140   270    88    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 270, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
06:50:58 1781614258000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:50:58 1781614258170  CD2B.0       track          -90    200    77    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 200, "position_z": 77, "remaining_time": 0, "track_confidence": 80}
06:50:58 1781614258791  333B.0       track          -60    230    67    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 230, "position_z": 67, "remaining_time": 0, "track_confidence": 80}
06:50:59 1781614259168  CD2B.0       track          -60    210    90    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 210, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
06:50:59 1781614259795  333B.0       track          0      190    93    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 190, "position_z": 93, "remaining_time": 0, "track_confidence": 80}
06:51:00 1781614260176  CD2B.0       track          -110   180    74    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 180, "position_z": 74, "remaining_time": 0, "track_confidence": 80}
06:51:00 1781614260800  333B.0       track          0      230    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:01 1781614261000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:51:01 1781614261174  CD2B.0       track          -90    200    89    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 200, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
06:51:01 1781614261693  333B.0       track          0      230    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:02 1781614262174  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:02 1781614262699  333B.0       track          0      230    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:03 1781614263000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:51:03 1781614263073  CD2B.0       track          -80    210    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 210, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:03 1781614263693  333B.0       track          0      230    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:04 1781614264081  CD2B.0       track          -130   160    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 160, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:04 1781614264114  333B.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1781614264114, "event_status": "start", "number_people": 2, "respiratory_rate": -1}
06:51:04 1781614264152  333B.1       EnterRoom      -      -      -     {"track_id": 1, "heart_rate": -1, "event_since": 1781614264152, "event_status": "start", "respiratory_rate": -1}
06:51:04 1781614264154  333B.0       track          0      230    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 230, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
06:51:04 1781614264154  333B.1       track          -220   270    0     {"pose": 4, "event": 1, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 270, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
06:51:04 1781614264709  333B.0       track          0      230    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 230, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
06:51:04 1781614264709  333B.1       track          -200   240    79    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 79, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
06:51:05 1781614265000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:51:05 1781614265074  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:05 1781614265708  333B.0       track          0      230    0     {"pose": 4, "event": 2, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 230, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
06:51:05 1781614265708  333B.1       track          -190   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 230, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
06:51:05 1781614265748  333B         ExitRoom       -      -      -     {"heart_rate": -1, "event_since": 1781614265748, "event_status": "start", "respiratory_rate": -1}
06:51:06 1781614266077  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:06 1781614266678  333B.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1781614266678, "event_status": "start", "number_people": 1, "respiratory_rate": -1}
06:51:06 1781614266715  333B.1       track          -200   250    61    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 250, "position_z": 61, "remaining_time": 0, "track_confidence": 80}
06:51:07 1781614267000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:51:07 1781614267077  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:07 1781614267664  333B.1       track          -210   250    68    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
06:51:08 1781614268128  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:08 1781614268671  333B.1       track          -200   250    78    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 250, "position_z": 78, "remaining_time": 0, "track_confidence": 80}
06:51:09 1781614269000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:51:09 1781614269024  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:09 1781614269674  333B.1       track          -200   240    82    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:51:10 1781614270041  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:10 1781614270690  333B.1       track          -210   240    76    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 76, "remaining_time": 0, "track_confidence": 80}
06:51:11 1781614271000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:51:11 1781614271029  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:11 1781614271680  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614271680, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 4, "walk_duration": 32, "stand_duration": 27, "respiratory_rate": -1, "multi_person_duration": 1}
06:51:11 1781614271680  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:51:11 1781614271718  333B.1       track          -210   240    68    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
06:51:12 1781614272051  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614272051, "track_count": 1, "event_status": "instant", "lie_duration": 27, "walk_distance": 1, "walk_duration": 4, "stand_duration": 29, "respiratory_rate": -1, "multi_person_duration": 0}
06:51:12 1781614272051  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:51:12 1781614272091  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:12 1781614272668  333B.1       track          -210   250    99    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 99, "remaining_time": 0, "track_confidence": 80}
06:51:13 1781614273000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 13, "track_confidence": 90, "vital_confidence": 0}
06:51:13 1781614273030  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:13 1781614273676  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:14 1781614274032  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:14 1781614274684  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:15 1781614275000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 67, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:51:15 1781614275029  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:15 1781614275669  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:16 1781614276041  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:16 1781614276673  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:17 1781614277000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:51:17 1781614277037  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:17 1781614277564  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:18 1781614278032  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:18 1781614278576  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:19 1781614279000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:19 1781614279032  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:19 1781614279568  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:20 1781614280036  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:20 1781614280566  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:20 1781614280954  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:21 1781614281000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:21 1781614281568  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:21 1781614281934  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:22 1781614282575  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:22 1781614282936  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:23 1781614283000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:23 1781614283574  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:23 1781614283942  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:24 1781614284576  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:24 1781614284937  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:25 1781614285000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 70, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:51:25 1781614285577  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:25 1781614285950  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:26 1781614286576  333B.1       track          -200   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:26 1781614286942  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:27 1781614287000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 70, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:51:27 1781614287582  333B.1       track          -220   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:27 1781614287942  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:28 1781614288483  333B.1       track          -210   250    84    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
06:51:28 1781614288946  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:29 1781614289000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 69, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:51:29 1781614289478  333B.1       track          -220   230    46    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 230, "position_z": 46, "remaining_time": 0, "track_confidence": 80}
06:51:29 1781614289943  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:30 1781614290477  333B.1       track          -210   250    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:30 1781614290945  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:31 1781614291000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 67, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:31 1781614291488  333B.1       track          -210   250    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:31 1781614291843  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:32 1781614292480  333B.1       track          -210   250    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:32 1781614292851  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:33 1781614293000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 68, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:33 1781614293485  333B.1       track          -210   250    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:33 1781614293840  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:34 1781614294482  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:34 1781614294841  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:35 1781614295000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 66, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:51:35 1781614295487  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:35 1781614295842  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:36 1781614296492  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:36 1781614296845  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:37 1781614297000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:51:37 1781614297484  333B.1       track          -210   250    51    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 51, "remaining_time": 0, "track_confidence": 80}
06:51:37 1781614297848  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:38 1781614298504  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:38 1781614298847  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:39 1781614299000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:51:39 1781614299494  333B.1       track          -200   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:39 1781614299849  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:40 1781614300383  333B.1       track          -200   220    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 220, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:40 1781614300871  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:41 1781614301000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:41 1781614301382  333B.1       track          -200   230    57    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 230, "position_z": 57, "remaining_time": 0, "track_confidence": 80}
06:51:41 1781614301850  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:42 1781614302398  333B.1       track          -210   240    77    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 77, "remaining_time": 0, "track_confidence": 80}
06:51:42 1781614302859  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:43 1781614303000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:43 1781614303384  333B.1       track          -200   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:43 1781614303745  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:44 1781614304415  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:44 1781614304746  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:45 1781614305000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:51:45 1781614305388  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:45 1781614305744  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:46 1781614306387  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:46 1781614306748  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:47 1781614307000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:47 1781614307388  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:47 1781614307754  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:48 1781614308389  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:48 1781614308747  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:49 1781614309000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:49 1781614309395  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:49 1781614309749  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:50 1781614310392  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:50 1781614310752  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:51 1781614311000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:51:51 1781614311392  333B.1       track          -210   230    72    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 72, "remaining_time": 0, "track_confidence": 80}
06:51:51 1781614311752  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:52 1781614312285  333B.1       track          -200   230    75    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 230, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
06:51:52 1781614312753  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:53 1781614313000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:51:53 1781614313305  333B.1       track          -200   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:53 1781614313760  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:54 1781614314294  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:54 1781614314765  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:55 1781614315000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:51:55 1781614315298  333B.1       track          -200   240    46    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 46, "remaining_time": 0, "track_confidence": 80}
06:51:55 1781614315650  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:56 1781614316289  333B.1       track          -200   240    66    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:51:56 1781614316657  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:57 1781614317000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:57 1781614317296  333B.1       track          -210   240    70    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 70, "remaining_time": 0, "track_confidence": 80}
06:51:57 1781614317649  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:58 1781614318293  333B.1       track          -220   240    85    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 240, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:51:58 1781614318653  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:51:59 1781614319000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:51:59 1781614319292  333B.1       track          -210   230    85    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:51:59 1781614319651  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:00 1781614320292  333B.1       track          -210   250    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:00 1781614320680  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:01 1781614321000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 14, "track_confidence": 90, "vital_confidence": 0}
06:52:01 1781614321294  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:01 1781614321657  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:02 1781614322303  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:02 1781614322653  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:03 1781614323000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:52:03 1781614323206  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:03 1781614323660  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:04 1781614324196  333B.1       track          -210   240    103   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
06:52:04 1781614324656  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:05 1781614325000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 60, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:52:05 1781614325208  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:05 1781614325660  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:06 1781614326200  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:06 1781614326661  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:07 1781614327000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 62, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 15, "track_confidence": 90, "vital_confidence": 0}
06:52:07 1781614327200  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:07 1781614327560  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:08 1781614328200  333B.1       track          -210   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:08 1781614328556  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:09 1781614329000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 65, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:52:09 1781614329204  333B.1       track          -220   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:09 1781614329560  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:10 1781614330201  333B.1       track          -220   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:10 1781614330554  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:11 1781614331000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:52:11 1781614331218  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:52:11 1781614331218  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614331218, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "respiratory_rate": -1, "multi_person_duration": 0}
06:52:11 1781614331253  333B.1       track          -220   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:11 1781614331591  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:52:11 1781614331591  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614331591, "track_count": 1, "event_status": "instant", "lie_duration": 60, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:52:11 1781614331627  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:12 1781614332206  333B.1       track          -200   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:12 1781614332573  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:13 1781614333000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:52:13 1781614333216  333B.1       track          -190   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:13 1781614333573  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:14 1781614334109  333B.1       track          -190   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:14 1781614334585  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:15 1781614335000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 61, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:52:15 1781614335110  333B.1       track          -200   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:15 1781614335574  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:16 1781614336108  333B.1       track          -210   240    95    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 95, "remaining_time": 0, "track_confidence": 80}
06:52:16 1781614336581  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:17 1781614337000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 59, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 16, "track_confidence": 90, "vital_confidence": 0}
06:52:17 1781614337113  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:17 1781614337472  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:18 1781614338113  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:18 1781614338474  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:19 1781614339113  333B.1       track          -210   240    98    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 240, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
06:52:19 1781614339474  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:20 1781614340000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 58, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:52:20 1781614340114  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:20 1781614340492  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:21 1781614341114  333B.1       track          -220   250    84    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 250, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
06:52:21 1781614341476  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:22 1781614342000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 59, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:52:22 1781614342116  333B.1       track          -200   270    93    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 270, "position_z": 93, "remaining_time": 0, "track_confidence": 80}
06:52:22 1781614342501  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:23 1781614343116  333B.1       track          -220   260    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:23 1781614343476  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:24 1781614344000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 63, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 18, "track_confidence": 90, "vital_confidence": 0}
06:52:24 1781614344116  333B.1       track          -210   260    80    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 260, "position_z": 80, "remaining_time": 0, "track_confidence": 80}
06:52:24 1781614344485  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:25 1781614345117  333B.1       track          -190   270    90    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 270, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
06:52:25 1781614345481  CD2B.0       track          -130   170    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 170, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:26 1781614346000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:52:26 1781614346015  333B.1       track          -200   250    69    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 250, "position_z": 69, "remaining_time": 0, "track_confidence": 80}
06:52:26 1781614346489  CD2B.0       track          -60    220    97    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 220, "position_z": 97, "remaining_time": 0, "track_confidence": 80}
06:52:27 1781614347020  333B.1       track          -190   250    87    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 250, "position_z": 87, "remaining_time": 0, "track_confidence": 80}
06:52:27 1781614347488  CD2B.0       track          -20    250    0     {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -20, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:28 1781614348000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:28 1781614348021  333B.1       track          -190   240    97    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 240, "position_z": 97, "remaining_time": 0, "track_confidence": 80}
06:52:28 1781614348485  CD2B.0       track          -110   200    76    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 76, "remaining_time": 0, "track_confidence": 80}
06:52:29 1781614349025  333B.1       track          -190   240    114   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 240, "position_z": 114, "remaining_time": 0, "track_confidence": 80}
06:52:29 1781614349377  CD2B.0       track          -120   220    88    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 220, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
06:52:30 1781614350000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:30 1781614350020  333B.1       track          -190   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:30 1781614350380  CD2B.0       track          -110   230    88    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 230, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
06:52:31 1781614351021  333B.1       track          -200   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:31 1781614351380  CD2B.0       track          -110   210    97    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 210, "position_z": 97, "remaining_time": 0, "track_confidence": 80}
06:52:32 1781614352000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:32 1781614352028  333B.1       track          -210   230    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:32 1781614352385  CD2B.0       track          -120   200    98    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 200, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
06:52:33 1781614353021  333B.1       track          -190   250    113   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 250, "position_z": 113, "remaining_time": 0, "track_confidence": 80}
06:52:33 1781614353382  CD2B.0       track          -120   200    87    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 200, "position_z": 87, "remaining_time": 0, "track_confidence": 80}
06:52:34 1781614354000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:34 1781614354024  333B.1       track          -220   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:34 1781614354381  CD2B.0       track          -130   210    108   {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 210, "position_z": 108, "remaining_time": 0, "track_confidence": 80}
06:52:35 1781614355024  333B.1       track          -200   260    77    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 260, "position_z": 77, "remaining_time": 0, "track_confidence": 80}
06:52:35 1781614355384  CD2B.0       track          -120   220    94    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 220, "position_z": 94, "remaining_time": 0, "track_confidence": 80}
06:52:36 1781614356000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:36 1781614356024  333B.1       track          -200   270    72    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 270, "position_z": 72, "remaining_time": 0, "track_confidence": 80}
06:52:36 1781614356388  CD2B.0       track          -130   220    89    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 220, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
06:52:36 1781614356928  333B.1       track          -210   280    110   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 280, "position_z": 110, "remaining_time": 0, "track_confidence": 80}
06:52:37 1781614357384  CD2B.0       track          -120   200    82    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 200, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:52:37 1781614357938  333B.1       track          -210   250    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:38 1781614358000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:38 1781614358393  CD2B.0       track          -130   210    84    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 210, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
06:52:38 1781614358929  333B.1       track          -210   270    82    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 270, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:52:39 1781614359388  CD2B.0       track          -190   180    78    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 180, "position_z": 78, "remaining_time": 0, "track_confidence": 80}
06:52:39 1781614359944  333B.1       track          -200   270    80    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 270, "position_z": 80, "remaining_time": 0, "track_confidence": 80}
06:52:40 1781614360000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:40 1781614360394  CD2B.0       track          -190   180    78    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 180, "position_z": 78, "remaining_time": 0, "track_confidence": 80}
06:52:40 1781614360944  333B.1       track          -210   280    103   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 280, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
06:52:41 1781614361284  CD2B.0       track          -180   180    82    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -180, "position_y": 180, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:52:41 1781614361940  333B.1       track          -220   270    108   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 270, "position_z": 108, "remaining_time": 0, "track_confidence": 80}
06:52:42 1781614362000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:42 1781614362288  CD2B.0       track          -130   200    90    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 200, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
06:52:42 1781614362934  333B.1       track          -200   270    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:43 1781614363288  CD2B.0       track          -100   200    84    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 200, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
06:52:43 1781614363936  333B.1       track          -190   250    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:44 1781614364000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:44 1781614364289  CD2B.0       track          -100   190    85    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 190, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:52:44 1781614364939  333B.1       track          -180   270    110   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -180, "position_y": 270, "position_z": 110, "remaining_time": 0, "track_confidence": 80}
06:52:45 1781614365287  CD2B.0       track          -120   190    85    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 190, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:52:45 1781614365938  333B.1       track          -210   280    116   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 280, "position_z": 116, "remaining_time": 0, "track_confidence": 80}
06:52:46 1781614366000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:46 1781614366296  CD2B.0       track          -130   200    86    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 200, "position_z": 86, "remaining_time": 0, "track_confidence": 80}
06:52:46 1781614366984  333B.1       track          -190   290    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:47 1781614367289  CD2B.0       track          -100   200    89    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 200, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
06:52:47 1781614367936  333B.1       track          -190   280    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:48 1781614368000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:48 1781614368292  CD2B.0       track          -110   200    98    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
06:52:48 1781614368832  333B.1       track          -210   280    121   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 280, "position_z": 121, "remaining_time": 0, "track_confidence": 80}
06:52:49 1781614369292  CD2B.0       track          -110   200    89    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 200, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
06:52:49 1781614369832  333B.1       track          -220   280    95    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 280, "position_z": 95, "remaining_time": 0, "track_confidence": 80}
06:52:50 1781614370000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:50 1781614370300  CD2B.0       track          -120   200    95    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 200, "position_z": 95, "remaining_time": 0, "track_confidence": 80}
06:52:50 1781614370836  333B.1       track          -170   270    90    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -170, "position_y": 270, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
06:52:51 1781614371292  CD2B.0       track          -120   200    91    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 200, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
06:52:51 1781614371830  333B.1       track          -170   280    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -170, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:52 1781614372000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:52 1781614372300  CD2B.0       track          -120   200    88    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 200, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
06:52:52 1781614372840  333B.1       track          -160   250    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -160, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:53 1781614373188  CD2B.0       track          -120   190    81    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 190, "position_z": 81, "remaining_time": 0, "track_confidence": 80}
06:52:53 1781614373834  333B.1       track          -150   260    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -150, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:54 1781614374000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:54 1781614374186  CD2B.0       track          -100   200    85    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 200, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:52:54 1781614374836  333B.1       track          -150   270    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -150, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:55 1781614375189  CD2B.0       track          -80    210    75    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 210, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
06:52:55 1781614375840  333B.1       track          -180   290    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -180, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:56 1781614376000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:56 1781614376192  CD2B.0       track          -100   210    86    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 210, "position_z": 86, "remaining_time": 0, "track_confidence": 80}
06:52:56 1781614376838  333B.1       track          -200   280    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:57 1781614377192  CD2B.0       track          -90    210    91    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 210, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
06:52:57 1781614377841  333B.1       track          -160   270    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -160, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:58 1781614378000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:52:58 1781614378194  CD2B.0       track          -110   210    92    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 210, "position_z": 92, "remaining_time": 0, "track_confidence": 80}
06:52:58 1781614378840  333B.1       track          -170   300    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -170, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:52:59 1781614379192  CD2B.0       track          -100   200    91    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 200, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
06:52:59 1781614379848  333B.1       track          -180   270    101   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -180, "position_y": 270, "position_z": 101, "remaining_time": 0, "track_confidence": 80}
06:53:00 1781614380000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:53:00 1781614380230  CD2B.0       track          -100   210    76    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 210, "position_z": 76, "remaining_time": 0, "track_confidence": 80}
06:53:00 1781614380733  333B.1       track          -200   240    59    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 59, "remaining_time": 0, "track_confidence": 80}
06:53:01 1781614381208  CD2B.0       track          -80    200    69    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 200, "position_z": 69, "remaining_time": 0, "track_confidence": 80}
06:53:01 1781614381734  333B.1       track          -200   250    103   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 250, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
06:53:02 1781614382000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:53:02 1781614382205  CD2B.0       track          -70    200    72    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 200, "position_z": 72, "remaining_time": 0, "track_confidence": 80}
06:53:02 1781614382735  333B.1       track          -190   240    107   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -190, "position_y": 240, "position_z": 107, "remaining_time": 0, "track_confidence": 80}
06:53:03 1781614383111  CD2B.0       track          -100   190    66    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 190, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:53:03 1781614383740  333B.1       track          -200   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:04 1781614384000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:53:04 1781614384106  CD2B.0       track          -100   220    85    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 220, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:53:04 1781614384737  333B.1       track          -200   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:05 1781614385114  CD2B.0       track          -110   210    82    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 210, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:53:05 1781614385744  333B.1       track          -200   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:06 1781614386000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:53:06 1781614386116  CD2B.0       track          -130   230    88    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 230, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
06:53:06 1781614386743  333B.1       track          -200   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:07 1781614387114  CD2B.0       track          -140   220    72    {"pose": 6, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 220, "position_z": 72, "remaining_time": 0, "track_confidence": 80}
06:53:07 1781614387740  333B.1       track          -200   260    106   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 260, "position_z": 106, "remaining_time": 0, "track_confidence": 80}
06:53:08 1781614388000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:53:08 1781614388124  CD2B.0       track          -220   290    46    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 290, "position_z": 46, "remaining_time": 0, "track_confidence": 80}
06:53:08 1781614388745  333B.1       track          -200   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:09 1781614389130  CD2B.0       track          -210   310    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -210, "position_y": 310, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:09 1781614389744  333B.1       track          -200   240    104   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 104, "remaining_time": 0, "track_confidence": 80}
06:53:10 1781614390000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:53:10 1781614390112  CD2B.0       track          -230   290    35    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -230, "position_y": 290, "position_z": 35, "remaining_time": 0, "track_confidence": 80}
06:53:10 1781614390762  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614390762, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "respiratory_rate": -1, "multi_person_duration": 0}
06:53:10 1781614390762  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:53:10 1781614390808  333B.1       track          -200   230    118   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 230, "position_z": 118, "remaining_time": 0, "track_confidence": 80}
06:53:11 1781614391128  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:53:11 1781614391128  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614391128, "track_count": 1, "event_status": "instant", "lie_duration": 58, "walk_distance": 0, "walk_duration": 2, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:53:11 1781614391172  CD2B.0       track          -220   290    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:11 1781614391744  333B.1       track          -210   250    115   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 250, "position_z": 115, "remaining_time": 0, "track_confidence": 80}
06:53:12 1781614392000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 1, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:53:12 1781614392116  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:12 1781614392646  333B.1       track          -200   260    108   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 260, "position_z": 108, "remaining_time": 0, "track_confidence": 80}
06:53:13 1781614393118  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:13 1781614393637  333B.1       track          -200   240    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:14 1781614394000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 1, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:53:14 1781614394014  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:14 1781614394645  333B.1       track          -200   230    108   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 230, "position_z": 108, "remaining_time": 0, "track_confidence": 80}
06:53:15 1781614395018  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:15 1781614395648  333B.1       track          -200   270    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:16 1781614396024  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:16 1781614396648  333B.1       track          -200   270    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:17 1781614397000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:53:17 1781614397022  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:17 1781614397648  333B.1       track          -210   260    82    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -210, "position_y": 260, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:53:18 1781614398018  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:18 1781614398651  333B.1       track          -220   280    108   {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 280, "position_z": 108, "remaining_time": 0, "track_confidence": 80}
06:53:19 1781614399000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:53:19 1781614399020  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:19 1781614399658  333B.1       track          -220   270    84    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -220, "position_y": 270, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
06:53:20 1781614400024  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:20 1781614400678  333B.1       track          -200   270    77    {"pose": 4, "area_id": 2, "track_id": 1, "position_x": -200, "position_y": 270, "position_z": 77, "remaining_time": 0, "track_confidence": 80}
06:53:21 1781614401000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 0, "heart_rate": 64, "init_status": 0, "pose_confidence": 90, "respiratory_rate": 17, "track_confidence": 90, "vital_confidence": 0}
06:53:21 1781614401024  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:21 1781614401656  333B.1       track          -110   220    82    {"pose": 1, "area_id": 2, "track_id": 1, "position_x": -110, "position_y": 220, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:53:22 1781614402000  1641.0       track          -      -      -     {"track_id": 0, "body_move": 0, "turn_over": 0, "bed_status": 1, "init_status": 0, "pose_confidence": 90, "track_confidence": 90, "vital_confidence": 0}
06:53:22 1781614402028  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:22 1781614402653  333B.1       track          0      190    73    {"pose": 1, "area_id": 2, "track_id": 1, "position_x": 0, "position_y": 190, "position_z": 73, "remaining_time": 0, "track_confidence": 80}
06:53:23 1781614403024  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:23 1781614403546  333B.1       track          0      190    0     {"pose": 1, "area_id": 2, "track_id": 1, "position_x": 0, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:23 1781614403775  1641         sleep-stage    -      -      -     {"heart_rate": -1, "event_since": 1781614402000, "sleep_stage": 1, "event_status": "instant", "respiratory_rate": -1}
06:53:23 1781614403775  1641         LeftBed        -      -      -     {"bed_status": 1, "heart_rate": -1, "event_since": 1781614402000, "event_status": "instant", "respiratory_rate": -1}
06:53:23 1781614403777  1641         LeftBed        -      -      -     {"bed_status": 1, "heart_rate": -1, "event_since": 1781614402000, "event_status": "start", "respiratory_rate": -1}
06:53:24 1781614404024  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:24 1781614404548  333B.1       track          0      190    0     {"pose": 1, "area_id": 2, "track_id": 1, "position_x": 0, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:25 1781614405028  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:25 1781614405553  333B.1       track          0      190    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": 0, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:25 1781614405920  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:26 1781614406552  333B.1       track          0      190    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": 0, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:26 1781614406929  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:27 1781614407556  333B.1       track          0      190    0     {"pose": 4, "area_id": 2, "track_id": 1, "position_x": 0, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:27 1781614407921  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:28 1781614408552  333B.1       track          0      190    0     {"pose": 4, "event": 2, "area_id": 1, "track_id": 1, "position_x": 0, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:28 1781614408596  333B.1       ExitRoom       -      -      -     {"track_id": 1, "heart_rate": -1, "event_since": 1781614408596, "event_status": "start", "respiratory_rate": -1}
06:53:28 1781614408952  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:29 1781614409568  333B.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1781614409568, "event_status": "start", "number_people": 0, "respiratory_rate": -1}
06:53:29 1781614409604  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:29 1781614409924  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:30 1781614410578  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:30 1781614410924  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:31 1781614411564  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:31 1781614411928  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:32 1781614412926  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:33 1781614413930  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:34 1781614414930  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:35 1781614415929  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:36 1781614416932  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:37 1781614417821  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:38 1781614418824  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:39 1781614419828  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:40 1781614420824  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:41 1781614421826  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:42 1781614422828  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:43 1781614423846  CD2B.0       track          -200   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:44 1781614424846  CD2B.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1781614424846, "event_status": "start", "number_people": 0, "respiratory_rate": -1}
06:53:44 1781614424888  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:45 1781614425840  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:46 1781614426840  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:56 1781614436366  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:53:56 1781614436670  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:54:28 1781614468040  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614468040, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 2, "walk_duration": 4, "stand_duration": 15, "respiratory_rate": -1, "multi_person_duration": 0}
06:54:28 1781614468040  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:54:28 1781614468364  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:54:28 1781614468485  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614468485, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 2, "stand_duration": 33, "respiratory_rate": -1, "multi_person_duration": 0}
06:54:28 1781614468485  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:54:28 1781614468673  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:54:59 1781614499768  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:55:00 1781614500168  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:55:31 1781614531584  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:55:31 1781614531584  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614531584, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:55:31 1781614531961  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:55:31 1781614531973  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614531973, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:55:31 1781614531973  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:55:32 1781614532265  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:56:03 1781614563262  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:56:03 1781614563646  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:56:35 1781614595076  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614595076, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:56:35 1781614595076  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:56:35 1781614595269  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:56:35 1781614595420  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:56:35 1781614595420  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614595420, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:56:35 1781614595572  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:06 1781614626752  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:07 1781614627196  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:38 1781614658504  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:57:38 1781614658504  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614658504, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:57:38 1781614658837  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:38 1781614658890  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:57:38 1781614658890  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614658890, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:57:39 1781614659140  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:39 1781614659709  333B.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1781614659709, "event_status": "start", "number_people": 1, "respiratory_rate": -1}
06:57:39 1781614659754  333B         EnterRoom      -      -      -     {"heart_rate": -1, "event_since": 1781614659754, "event_status": "start", "respiratory_rate": -1}
06:57:39 1781614659756  333B.0       track          -70    140    88    {"pose": 4, "event": 1, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 140, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
06:57:40 1781614660502  333B.0       track          -140   140    89    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 140, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
06:57:41 1781614661514  333B.0       track          -220   170    91    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -220, "position_y": 170, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
06:57:42 1781614662504  333B.0       track          -270   190    81    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -270, "position_y": 190, "position_z": 81, "remaining_time": 0, "track_confidence": 80}
06:57:43 1781614663506  333B.0       track          -280   220    88    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -280, "position_y": 220, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
06:57:44 1781614664510  333B.0       track          -300   240    98    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -300, "position_y": 240, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
06:57:45 1781614665508  333B.0       track          -300   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -300, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:46 1781614666510  333B.0       track          -300   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -300, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:47 1781614667512  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:48 1781614668410  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:49 1781614669416  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:50 1781614670422  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:51 1781614671414  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:52 1781614672414  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:53 1781614673416  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:54 1781614674424  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:55 1781614675417  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:56 1781614676420  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:57 1781614677421  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:58 1781614678421  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:57:59 1781614679421  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:00 1781614680314  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:01 1781614681316  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:02 1781614682318  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:03 1781614683323  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:04 1781614684321  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:05 1781614685323  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:06 1781614686334  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:07 1781614687322  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:08 1781614688350  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:58:08 1781614688350  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614688350, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 2, "walk_duration": 5, "stand_duration": 22, "respiratory_rate": -1, "multi_person_duration": 0}
06:58:08 1781614688388  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:09 1781614689324  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:10 1781614690326  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:10 1781614690741  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:58:10 1781614690741  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614690741, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:58:11 1781614691088  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:11 1781614691330  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:12 1781614692218  333B.0       track          -290   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:13 1781614693225  333B.0       track          -310   230    105   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -310, "position_y": 230, "position_z": 105, "remaining_time": 0, "track_confidence": 80}
06:58:14 1781614694220  333B.0       track          -330   230    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -330, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:15 1781614695226  333B.0       track          -350   220    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -350, "position_y": 220, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:16 1781614696229  333B.0       track          -320   230    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -320, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:17 1781614697224  333B.0       track          -290   240    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:18 1781614698228  333B.0       track          -290   210    106   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 210, "position_z": 106, "remaining_time": 0, "track_confidence": 80}
06:58:19 1781614699258  333B.0       track          -240   190    101   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -240, "position_y": 190, "position_z": 101, "remaining_time": 0, "track_confidence": 80}
06:58:20 1781614700260  333B.0       track          -170   190    104   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -170, "position_y": 190, "position_z": 104, "remaining_time": 0, "track_confidence": 80}
06:58:21 1781614701149  333B.0       track          -120   250    91    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 250, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
06:58:22 1781614702152  333B.0       track          -100   280    103   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
06:58:23 1781614703157  333B.0       track          -90    270    94    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 94, "remaining_time": 0, "track_confidence": 80}
06:58:24 1781614704153  333B.0       track          -70    280    94    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 280, "position_z": 94, "remaining_time": 0, "track_confidence": 80}
06:58:25 1781614705156  333B.0       track          -80    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:26 1781614706156  333B.0       track          -60    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:27 1781614707157  333B.0       track          -100   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:28 1781614708160  333B.0       track          -100   280    92    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 92, "remaining_time": 0, "track_confidence": 80}
06:58:29 1781614709166  333B.0       track          -100   260    87    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 260, "position_z": 87, "remaining_time": 0, "track_confidence": 80}
06:58:30 1781614710170  333B.0       track          -80    290    90    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 290, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
06:58:31 1781614711160  333B.0       track          -100   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:32 1781614712059  333B.0       track          -70    280    123   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 280, "position_z": 123, "remaining_time": 0, "track_confidence": 80}
06:58:33 1781614713062  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:34 1781614714062  333B.0       track          -60    300    113   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 300, "position_z": 113, "remaining_time": 0, "track_confidence": 80}
06:58:35 1781614715068  333B.0       track          -60    300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:36 1781614716068  333B.0       track          -80    300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:37 1781614717072  333B.0       track          -90    290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:38 1781614718068  333B.0       track          -90    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:39 1781614719072  333B.0       track          -100   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:40 1781614720069  333B.0       track          -90    270    110   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 110, "remaining_time": 0, "track_confidence": 80}
06:58:41 1781614721069  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:42 1781614722072  333B.0       track          -70    280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:42 1781614722378  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:42 1781614722978  333B.0       track          -80    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:43 1781614723978  333B.0       track          -80    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:44 1781614724972  333B.0       track          -100   270    114   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 114, "remaining_time": 0, "track_confidence": 80}
06:58:45 1781614725976  333B.0       track          -110   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:46 1781614726985  333B.0       track          -100   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:47 1781614727974  333B.0       track          -90    280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:48 1781614728977  333B.0       track          -100   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:49 1781614729976  333B.0       track          -100   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:50 1781614730906  333B.0       track          -100   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:51 1781614731909  333B.0       track          -100   270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:52 1781614732909  333B.0       track          -110   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:53 1781614733912  333B.0       track          -80    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:54 1781614734916  333B.0       track          -80    270    59    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 270, "position_z": 59, "remaining_time": 0, "track_confidence": 80}
06:58:55 1781614735914  333B.0       track          -70    250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:56 1781614736919  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:57 1781614737916  333B.0       track          -100   300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:58 1781614738916  333B.0       track          -100   300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:58:59 1781614739917  333B.0       track          -100   300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:00 1781614740918  333B.0       track          -100   300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:01 1781614741934  333B.0       track          -100   300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:02 1781614742818  333B.0       track          -100   300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:03 1781614743814  333B.0       track          -100   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:04 1781614744812  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:05 1781614745826  333B.0       track          -80    260    66    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 260, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
06:59:06 1781614746782  333B.0       track          -100   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:07 1781614747781  333B.0       track          -90    280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:08 1781614748799  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:59:08 1781614748799  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614748799, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 1, "walk_duration": 5, "stand_duration": 55, "respiratory_rate": -1, "multi_person_duration": 0}
06:59:08 1781614748837  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:09 1781614749782  333B.0       track          -100   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:10 1781614750786  333B.0       track          -80    290    103   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 290, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
06:59:11 1781614751785  333B.0       track          -100   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:12 1781614752792  333B.0       track          -90    280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:13 1781614753788  333B.0       track          -70    280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:14 1781614754121  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614754121, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
06:59:14 1781614754121  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
06:59:14 1781614754371  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:14 1781614754788  333B.0       track          -70    280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:15 1781614755790  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:16 1781614756794  333B.0       track          -110   270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:17 1781614757792  333B.0       track          -110   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:18 1781614758689  333B.0       track          -100   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:19 1781614759693  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:20 1781614760685  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:21 1781614761694  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:22 1781614762688  333B.0       track          -110   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:23 1781614763708  333B.0       track          -120   310    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 310, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:24 1781614764704  333B.0       track          -110   300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:25 1781614765709  333B.0       track          -90    280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:26 1781614766708  333B.0       track          -100   290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:27 1781614767710  333B.0       track          -80    260    91    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 260, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
06:59:28 1781614768608  333B.0       track          -100   300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:29 1781614769608  333B.0       track          -90    310    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 310, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:30 1781614770616  333B.0       track          -80    300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:31 1781614771612  333B.0       track          -80    290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:32 1781614772618  333B.0       track          -90    270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:33 1781614773608  333B.0       track          -90    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:34 1781614774608  333B.0       track          -100   260    90    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 260, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
06:59:35 1781614775610  333B.0       track          -100   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:36 1781614776612  333B.0       track          -90    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:37 1781614777613  333B.0       track          -90    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:38 1781614778613  333B.0       track          -90    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:39 1781614779517  333B.0       track          -50    300    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -50, "position_y": 300, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:40 1781614780516  333B.0       track          -70    280    103   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 280, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
06:59:41 1781614781537  333B.0       track          -60    290    86    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 290, "position_z": 86, "remaining_time": 0, "track_confidence": 80}
06:59:42 1781614782520  333B.0       track          -60    290    90    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 290, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
06:59:43 1781614783528  333B.0       track          -60    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:44 1781614784520  333B.0       track          -50    260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -50, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:45 1781614785520  333B.0       track          -90    270    82    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
06:59:45 1781614785917  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:46 1781614786523  333B.0       track          -90    270    112   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 112, "remaining_time": 0, "track_confidence": 80}
06:59:47 1781614787522  333B.0       track          -100   270    83    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 83, "remaining_time": 0, "track_confidence": 80}
06:59:48 1781614788528  333B.0       track          -110   260    95    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 95, "remaining_time": 0, "track_confidence": 80}
06:59:49 1781614789526  333B.0       track          -110   260    110   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 110, "remaining_time": 0, "track_confidence": 80}
06:59:50 1781614790526  333B.0       track          -100   240    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:51 1781614791420  333B.0       track          -100   240    94    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 240, "position_z": 94, "remaining_time": 0, "track_confidence": 80}
06:59:52 1781614792422  333B.0       track          -100   240    85    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 240, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
06:59:53 1781614793424  333B.0       track          -100   250    97    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 250, "position_z": 97, "remaining_time": 0, "track_confidence": 80}
06:59:54 1781614794428  333B.0       track          -110   280    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 280, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:55 1781614795390  333B.0       track          -110   270    105   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 270, "position_z": 105, "remaining_time": 0, "track_confidence": 80}
06:59:56 1781614796398  333B.0       track          -110   260    105   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 105, "remaining_time": 0, "track_confidence": 80}
06:59:57 1781614797398  333B.0       track          -100   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:58 1781614798404  333B.0       track          -100   240    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
06:59:59 1781614799395  333B.0       track          -110   260    106   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 106, "remaining_time": 0, "track_confidence": 80}
07:00:00 1781614800393  333B.0       track          -120   260    114   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 260, "position_z": 114, "remaining_time": 0, "track_confidence": 80}
07:00:01 1781614801394  333B.0       track          -110   260    122   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 122, "remaining_time": 0, "track_confidence": 80}
07:00:02 1781614802399  333B.0       track          -110   260    87    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 87, "remaining_time": 0, "track_confidence": 80}
07:00:03 1781614803397  333B.0       track          -110   260    99    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 99, "remaining_time": 0, "track_confidence": 80}
07:00:04 1781614804403  333B.0       track          -100   270    101   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 101, "remaining_time": 0, "track_confidence": 80}
07:00:05 1781614805405  333B.0       track          -120   250    111   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 250, "position_z": 111, "remaining_time": 0, "track_confidence": 80}
07:00:06 1781614806400  333B.0       track          -110   250    98    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 250, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
07:00:07 1781614807292  333B.0       track          -100   250    95    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 250, "position_z": 95, "remaining_time": 0, "track_confidence": 80}
07:00:08 1781614808308  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614808308, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "respiratory_rate": -1, "multi_person_duration": 0}
07:00:08 1781614808308  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
07:00:08 1781614808356  333B.0       track          -110   260    99    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 99, "remaining_time": 0, "track_confidence": 80}
07:00:09 1781614809300  333B.0       track          -110   250    115   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 250, "position_z": 115, "remaining_time": 0, "track_confidence": 80}
07:00:10 1781614810308  333B.0       track          -110   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:11 1781614811324  333B.0       track          -100   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:12 1781614812328  333B.0       track          -120   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:13 1781614813327  333B.0       track          -110   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:14 1781614814338  333B.0       track          -100   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:15 1781614815335  333B.0       track          -110   270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:16 1781614816221  333B.0       track          -100   270    116   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 116, "remaining_time": 0, "track_confidence": 80}
07:00:17 1781614817224  333B.0       track          -100   250    105   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 250, "position_z": 105, "remaining_time": 0, "track_confidence": 80}
07:00:17 1781614817618  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614817618, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
07:00:17 1781614817618  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
07:00:17 1781614817696  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:18 1781614818227  333B.0       track          -110   250    113   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 250, "position_z": 113, "remaining_time": 0, "track_confidence": 80}
07:00:19 1781614819222  333B.0       track          -120   270    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -120, "position_y": 270, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:20 1781614820248  333B.0       track          -90    250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:21 1781614821231  333B.0       track          -90    250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:22 1781614822238  333B.0       track          -100   240    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 240, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:23 1781614823233  333B.0       track          -110   250    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 250, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:24 1781614824231  333B.0       track          -110   260    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:25 1781614825236  333B.0       track          -110   250    110   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 250, "position_z": 110, "remaining_time": 0, "track_confidence": 80}
07:00:26 1781614826229  333B.0       track          -110   260    85    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
07:00:27 1781614827238  333B.0       track          -100   240    84    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 240, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
07:00:28 1781614828128  333B.0       track          -90    240    115   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 240, "position_z": 115, "remaining_time": 0, "track_confidence": 80}
07:00:29 1781614829127  333B.0       track          -80    240    129   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 240, "position_z": 129, "remaining_time": 0, "track_confidence": 80}
07:00:30 1781614830129  333B.0       track          -90    230    115   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 230, "position_z": 115, "remaining_time": 0, "track_confidence": 80}
07:00:31 1781614831127  333B.0       track          -90    220    120   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 220, "position_z": 120, "remaining_time": 0, "track_confidence": 80}
07:00:32 1781614832134  333B.0       track          -90    220    141   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 220, "position_z": 141, "remaining_time": 0, "track_confidence": 80}
07:00:33 1781614833137  333B.0       track          -80    230    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 230, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:34 1781614834132  333B.0       track          -90    230    138   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 230, "position_z": 138, "remaining_time": 0, "track_confidence": 80}
07:00:35 1781614835137  333B.0       track          -90    250    153   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 250, "position_z": 153, "remaining_time": 0, "track_confidence": 80}
07:00:36 1781614836636  333B.0       track          -90    250    122   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 250, "position_z": 122, "remaining_time": 0, "track_confidence": 80}
07:00:37 1781614837135  333B.0       track          -100   240    93    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 240, "position_z": 93, "remaining_time": 0, "track_confidence": 80}
07:00:38 1781614838137  333B.0       track          -100   220    111   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 220, "position_z": 111, "remaining_time": 0, "track_confidence": 80}
07:00:39 1781614839137  333B.0       track          -110   260    93    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 93, "remaining_time": 0, "track_confidence": 80}
07:00:40 1781614840037  333B.0       track          -100   260    105   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 260, "position_z": 105, "remaining_time": 0, "track_confidence": 80}
07:00:41 1781614841040  333B.0       track          -100   260    124   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 260, "position_z": 124, "remaining_time": 0, "track_confidence": 80}
07:00:42 1781614842032  333B.0       track          -110   260    126   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 126, "remaining_time": 0, "track_confidence": 80}
07:00:43 1781614843037  333B.0       track          -100   250    96    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 250, "position_z": 96, "remaining_time": 0, "track_confidence": 80}
07:00:44 1781614844032  333B.0       track          -110   250    106   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 250, "position_z": 106, "remaining_time": 0, "track_confidence": 80}
07:00:45 1781614845040  333B.0       track          -110   240    96    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 240, "position_z": 96, "remaining_time": 0, "track_confidence": 80}
07:00:46 1781614846041  333B.0       track          -100   180    98    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 180, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
07:00:47 1781614847036  333B.0       track          -80    120    107   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 120, "position_z": 107, "remaining_time": 0, "track_confidence": 80}
07:00:48 1781614848039  333B.0       track          -60    80     95    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -60, "position_y": 80, "position_z": 95, "remaining_time": 0, "track_confidence": 80}
07:00:49 1781614849041  333B.0       track          -30    20     120   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -30, "position_y": 20, "position_z": 120, "remaining_time": 0, "track_confidence": 80}
07:00:49 1781614849402  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:00:50 1781614850049  333B.0       track          -30    30     99    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -30, "position_y": 30, "position_z": 99, "remaining_time": 0, "track_confidence": 80}
07:00:51 1781614851041  333B.0       track          -70    110    91    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 110, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
07:00:51 1781614851934  333B.0       track          -90    190    73    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 190, "position_z": 73, "remaining_time": 0, "track_confidence": 80}
07:00:52 1781614852938  333B.0       track          -110   260    76    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 76, "remaining_time": 0, "track_confidence": 80}
07:00:53 1781614853936  333B.0       track          -110   260    83    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 260, "position_z": 83, "remaining_time": 0, "track_confidence": 80}
07:00:54 1781614854948  333B.0       track          -130   250    88    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -130, "position_y": 250, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
07:00:55 1781614855938  333B.0       track          -140   260    102   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 260, "position_z": 102, "remaining_time": 0, "track_confidence": 80}
07:00:56 1781614856948  333B.0       track          -190   200    85    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -190, "position_y": 200, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
07:00:57 1781614857942  333B.0       track          -260   180    84    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -260, "position_y": 180, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
07:00:58 1781614858953  333B.0       track          -290   210    97    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 210, "position_z": 97, "remaining_time": 0, "track_confidence": 80}
07:00:59 1781614859965  333B.0       track          -280   210    90    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -280, "position_y": 210, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
07:01:00 1781614860977  333B.0       track          -290   200    105   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -290, "position_y": 200, "position_z": 105, "remaining_time": 0, "track_confidence": 80}
07:01:01 1781614861857  333B.0       track          -260   190    94    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -260, "position_y": 190, "position_z": 94, "remaining_time": 0, "track_confidence": 80}
07:01:02 1781614862857  333B.0       track          -200   230    94    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -200, "position_y": 230, "position_z": 94, "remaining_time": 0, "track_confidence": 80}
07:01:03 1781614863864  333B.0       track          -140   260    103   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -140, "position_y": 260, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
07:01:04 1781614864868  333B.0       track          -90    240    120   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 240, "position_z": 120, "remaining_time": 0, "track_confidence": 80}
07:01:05 1781614865858  333B.0       track          -80    260    122   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 260, "position_z": 122, "remaining_time": 0, "track_confidence": 80}
07:01:06 1781614866872  333B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
07:01:06 1781614866872  333B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614866872, "track_count": 1, "event_status": "instant", "lie_duration": 0, "walk_distance": 6, "walk_duration": 19, "stand_duration": 41, "respiratory_rate": -1, "multi_person_duration": 0}
07:01:06 1781614866908  333B.0       track          -90    260    84    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 260, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
07:01:07 1781614867860  333B.0       track          -90    280    105   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 280, "position_z": 105, "remaining_time": 0, "track_confidence": 80}
07:01:08 1781614868862  333B.0       track          -90    270    118   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -90, "position_y": 270, "position_z": 118, "remaining_time": 0, "track_confidence": 80}
07:01:09 1781614869864  333B.0       track          -80    290    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -80, "position_y": 290, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:10 1781614870866  333B.0       track          -100   280    115   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 280, "position_z": 115, "remaining_time": 0, "track_confidence": 80}
07:01:11 1781614871868  333B.0       track          -100   270    84    {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 84, "remaining_time": 0, "track_confidence": 80}
07:01:12 1781614872769  333B.0       track          -100   270    103   {"pose": 4, "area_id": 1, "track_id": 0, "position_x": -100, "position_y": 270, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
07:01:13 1781614873774  333B.0       track          -110   230    103   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -110, "position_y": 230, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
07:01:14 1781614874768  333B.0       track          -70    210    104   {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -70, "position_y": 210, "position_z": 104, "remaining_time": 0, "track_confidence": 80}
07:01:15 1781614875776  333B.0       track          -20    200    80    {"pose": 1, "area_id": 1, "track_id": 0, "position_x": -20, "position_y": 200, "position_z": 80, "remaining_time": 0, "track_confidence": 80}
07:01:16 1781614876780  333B.0       track          0      190    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:17 1781614877784  333B.0       track          0      190    0     {"pose": 1, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:18 1781614878780  333B.0       track          0      190    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 190, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:19 1781614879784  333B.0       track          0      200    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:20 1781614880796  333B.0       track          0      200    0     {"pose": 4, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:21 1781614881099  CD2B.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1781614881099, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
07:01:21 1781614881099  CD2B.11      heart          -      -      -     {"track_id": 11, "track_confidence": 80}
07:01:21 1781614881248  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:21 1781614881788  333B.0       track          0      200    0     {"pose": 4, "event": 2, "area_id": 1, "track_id": 0, "position_x": 0, "position_y": 200, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:21 1781614881826  333B         ExitRoom       -      -      -     {"heart_rate": -1, "event_since": 1781614881826, "event_status": "start", "respiratory_rate": -1}
07:01:22 1781614882697  333B.10      number_people  -      -      -     {"track_id": 10, "heart_rate": -1, "event_since": 1781614882697, "event_status": "start", "number_people": 0, "respiratory_rate": -1}
07:01:22 1781614882736  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:23 1781614883692  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:24 1781614884692  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:52 1781614912440  333B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
07:01:52 1781614912888  CD2B.88      track          0      0      0     {"track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
```

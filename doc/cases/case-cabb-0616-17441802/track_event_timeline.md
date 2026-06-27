# case-cabb-0616-17441802 — 每 tick belief 时间线 (room fd00:0:3:411:1:200, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
03:44:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:03 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:34 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:37 CABB.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:37 CABB.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
03:44:37 CABB.0   CABB00248549  stand   95   NoReport stand              room -    Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
03:44:37 CABB.0   CABB00248549  stand   88   NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
03:44:38 CABB.0   CABB00248549  walk    78   NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.54  0.00  0.28  0.03
03:44:39 CABB.0   CABB00248549  walk    112  NoReport walk               room -    OpenFloor  1   0     0.00  0.04  0.57  0.01  0.20  0.03
03:44:40 CABB.0   CABB00248549  walk    114  NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.58  0.01  0.15  0.03
03:44:41 CABB.0   CABB00248549  walk    90   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.59  0.01  0.11  0.03
03:44:42 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   0     0.00  0.05  0.59  0.01  0.11  0.03
03:44:43 CABB.0   CABB00248549  walk    96   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.60  0.02  0.09  0.03
03:44:43 CABB.0   CABB00248549  walk    93   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.08  0.03
03:44:44 CABB.0   CABB00248549  walk    113  NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
03:44:45 CABB.0   CABB00248549  walk    91   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
03:44:46 CABB.0   CABB00248549  walk    90   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
03:44:47 CABB.0   CABB00248549  walk    56   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
03:44:48 CABB.0   CABB00248549  walk    97   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
03:44:49 CABB.0   CABB00248549  walk    70   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
03:44:50 CABB.0   CABB00248549  walk    70   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
03:44:51 CABB.0   CABB00248549  walk    78   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
03:44:52 CABB.0   CABB00248549  walk    92   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
03:44:53 CABB.0   CABB00248549  walk    47   NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
03:44:54 CABB.0   CABB00248549  walk    0    NoReport walk               room -    OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
03:44:55 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  1   0     0.01  0.07  0.41  0.03  0.07  0.04
03:44:56 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  1   0     0.01  0.07  0.28  0.03  0.09  0.03
03:44:57 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.06  0.20  0.04  0.09  0.03
03:44:58 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.05  0.15  0.04  0.08  0.02
03:44:59 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.04  0.12  0.04  0.07  0.02
03:45:00 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.03  0.11  0.04  0.05  0.02
03:45:01 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.03  0.10  0.04  0.04  0.02
03:45:02 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.10  0.04  0.04  0.02
03:45:03 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.04  0.03  0.02
03:45:04 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.03  0.02
03:45:05 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.03  0.02
03:45:06 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:07 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:08 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:09 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:10 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:11 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:12 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:13 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:14 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:15 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:16 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:17 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:18 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
03:45:19 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   28    0.01  0.02  0.09  0.03  0.02  0.02
03:45:20 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   29    0.01  0.02  0.09  0.03  0.02  0.02
03:45:21 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   30    0.01  0.02  0.09  0.03  0.02  0.02
03:45:22 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   31    0.01  0.02  0.09  0.03  0.02  0.02
03:45:23 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   32    0.01  0.02  0.09  0.03  0.02  0.02
03:45:24 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   33    0.01  0.02  0.09  0.03  0.02  0.02
03:45:25 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   34    0.01  0.02  0.09  0.03  0.02  0.02
03:45:26 -.-      -             -       -    NoReport (no frame, held)   room -    Sit        1   34    0.01  0.02  0.09  0.03  0.02  0.02
03:45:27 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        1   36    0.01  0.02  0.09  0.03  0.02  0.02
03:45:27 CABB.E   -             -       0    NoReport np=2               room -    Sit        1   36    0.01  0.02  0.09  0.03  0.02  0.02
03:45:27 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.50 Sit        2   36    0.00  0.02  0.26  0.00  0.69  0.03
03:45:27 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        2   36    0.00  0.01  0.05  0.02  0.01  0.01
03:45:28 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.51 Sit        2   37    0.00  0.02  0.52  0.00  0.40  0.01
03:45:28 CABB.0   CABB00248549  sit     0    NoReport sit                room -    Sit        2   37    0.00  0.01  0.04  0.01  0.01  0.01
03:45:29 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   38    0.00  0.02  0.70  0.00  0.18  0.02
03:45:29 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.52 OpenFloor  2   38    0.00  0.02  0.70  0.00  0.18  0.02
03:45:30 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.53 OpenFloor  2   39    0.00  0.02  0.79  0.00  0.07  0.02
03:45:30 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   39    0.00  0.02  0.79  0.00  0.07  0.02
03:45:31 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   40    0.00  0.02  0.83  0.00  0.03  0.02
03:45:31 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.54 OpenFloor  2   40    0.00  0.02  0.83  0.00  0.03  0.02
03:45:32 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   41    0.00  0.02  0.84  0.00  0.02  0.02
03:45:32 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   41    0.00  0.02  0.84  0.00  0.02  0.02
03:45:33 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   42    0.00  0.02  0.85  0.00  0.01  0.02
03:45:33 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   42    0.00  0.02  0.85  0.00  0.01  0.02
03:45:34 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   43    0.00  0.02  0.85  0.00  0.01  0.02
03:45:34 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   43    0.00  0.02  0.85  0.00  0.01  0.02
03:45:35 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   44    0.00  0.02  0.85  0.00  0.01  0.02
03:45:35 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   44    0.00  0.02  0.85  0.00  0.01  0.02
03:45:36 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   45    0.00  0.02  0.85  0.00  0.01  0.02
03:45:36 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   45    0.00  0.02  0.85  0.00  0.01  0.02
03:45:37 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   46    0.00  0.02  0.85  0.00  0.01  0.02
03:45:37 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   46    0.00  0.02  0.85  0.00  0.01  0.02
03:45:38 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   47    0.00  0.02  0.85  0.00  0.01  0.02
03:45:38 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   47    0.00  0.02  0.85  0.00  0.01  0.02
03:45:39 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   48    0.00  0.02  0.85  0.00  0.01  0.02
03:45:39 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   48    0.00  0.02  0.85  0.00  0.01  0.02
03:45:40 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   49    0.00  0.02  0.85  0.00  0.01  0.02
03:45:40 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   49    0.00  0.02  0.85  0.00  0.01  0.02
03:45:41 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   50    0.00  0.02  0.85  0.00  0.01  0.02
03:45:41 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   50    0.00  0.02  0.85  0.00  0.01  0.02
03:45:42 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   51    0.00  0.02  0.85  0.00  0.01  0.02
03:45:42 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   51    0.00  0.02  0.85  0.00  0.01  0.02
03:45:43 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   52    0.00  0.02  0.85  0.00  0.01  0.02
03:45:43 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   52    0.00  0.02  0.85  0.00  0.01  0.02
03:45:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   53    0.00  0.02  0.85  0.00  0.01  0.02
03:45:44 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   53    0.00  0.02  0.85  0.00  0.01  0.02
03:45:45 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   54    0.00  0.03  0.81  0.00  0.02  0.02
03:45:45 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   54    0.00  0.03  0.81  0.00  0.02  0.02
03:45:46 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   55    0.00  0.02  0.84  0.00  0.01  0.02
03:45:46 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   55    0.00  0.02  0.84  0.00  0.01  0.02
03:45:47 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   56    0.00  0.02  0.84  0.00  0.01  0.02
03:45:47 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   56    0.00  0.02  0.84  0.00  0.01  0.02
03:45:48 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   57    0.00  0.02  0.85  0.00  0.01  0.02
03:45:48 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   57    0.00  0.02  0.85  0.00  0.01  0.02
03:45:49 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   58    0.00  0.02  0.85  0.00  0.01  0.02
03:45:49 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   58    0.00  0.02  0.85  0.00  0.01  0.02
03:45:50 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   59    0.00  0.02  0.85  0.00  0.01  0.02
03:45:50 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   59    0.00  0.02  0.85  0.00  0.01  0.02
03:45:51 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   60    0.00  0.02  0.85  0.00  0.01  0.02
03:45:51 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   60    0.00  0.02  0.85  0.00  0.01  0.02
03:45:52 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   61    0.00  0.02  0.85  0.00  0.01  0.02
03:45:52 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   61    0.00  0.02  0.85  0.00  0.01  0.02
03:45:53 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   62    0.00  0.02  0.85  0.00  0.01  0.02
03:45:53 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   62    0.00  0.02  0.85  0.00  0.01  0.02
03:45:54 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   63    0.00  0.02  0.85  0.00  0.01  0.02
03:45:54 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   63    0.00  0.02  0.85  0.00  0.01  0.02
03:45:55 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   64    0.00  0.02  0.85  0.00  0.01  0.02
03:45:55 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   64    0.00  0.02  0.85  0.00  0.01  0.02
03:45:55 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   65    0.00  0.02  0.85  0.00  0.01  0.02
03:45:55 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   65    0.00  0.02  0.85  0.00  0.01  0.02
03:45:56 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   66    0.00  0.02  0.85  0.00  0.01  0.02
03:45:56 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   66    0.00  0.02  0.85  0.00  0.01  0.02
03:45:57 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   67    0.00  0.02  0.85  0.00  0.01  0.02
03:45:57 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   67    0.00  0.02  0.85  0.00  0.01  0.02
03:45:58 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   67    0.00  0.02  0.85  0.00  0.01  0.02
03:45:59 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   68    0.00  0.02  0.85  0.00  0.01  0.02
03:45:59 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   68    0.00  0.02  0.85  0.00  0.01  0.02
03:45:59 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   69    0.00  0.02  0.85  0.00  0.01  0.02
03:45:59 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   69    0.00  0.02  0.85  0.00  0.01  0.02
03:46:00 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   70    0.00  0.02  0.85  0.00  0.01  0.02
03:46:00 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   70    0.00  0.02  0.85  0.00  0.01  0.02
03:46:01 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   71    0.00  0.02  0.85  0.00  0.01  0.02
03:46:01 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   71    0.00  0.02  0.85  0.00  0.01  0.02
03:46:02 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   72    0.00  0.02  0.85  0.00  0.01  0.02
03:46:02 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   72    0.00  0.02  0.85  0.00  0.01  0.02
03:46:03 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   72    0.00  0.02  0.85  0.00  0.01  0.02
03:46:04 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   73    0.00  0.04  0.74  0.00  0.02  0.04
03:46:04 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   73    0.00  0.04  0.74  0.00  0.02  0.04
03:46:04 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   74    0.00  0.03  0.81  0.00  0.02  0.02
03:46:04 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   74    0.00  0.03  0.81  0.00  0.02  0.02
03:46:05 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   75    0.00  0.02  0.84  0.00  0.01  0.02
03:46:05 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   75    0.00  0.02  0.84  0.00  0.01  0.02
03:46:06 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   76    0.00  0.02  0.84  0.00  0.01  0.02
03:46:06 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   76    0.00  0.02  0.84  0.00  0.01  0.02
03:46:07 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   77    0.00  0.02  0.85  0.00  0.01  0.02
03:46:07 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   77    0.00  0.02  0.85  0.00  0.01  0.02
03:46:08 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   78    0.00  0.02  0.85  0.00  0.01  0.02
03:46:08 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   78    0.00  0.02  0.85  0.00  0.01  0.02
03:46:09 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   79    0.00  0.02  0.85  0.00  0.01  0.02
03:46:09 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   79    0.00  0.02  0.85  0.00  0.01  0.02
03:46:10 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   80    0.00  0.02  0.85  0.00  0.01  0.02
03:46:10 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   80    0.00  0.02  0.85  0.00  0.01  0.02
03:46:11 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   81    0.00  0.02  0.85  0.00  0.01  0.02
03:46:11 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   81    0.00  0.02  0.85  0.00  0.01  0.02
03:46:12 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   82    0.00  0.02  0.85  0.00  0.01  0.02
03:46:12 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   82    0.00  0.02  0.85  0.00  0.01  0.02
03:46:13 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   83    0.00  0.02  0.85  0.00  0.01  0.02
03:46:13 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   83    0.00  0.02  0.85  0.00  0.01  0.02
03:46:14 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   84    0.00  0.02  0.85  0.00  0.01  0.02
03:46:14 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   84    0.00  0.02  0.85  0.00  0.01  0.02
03:46:15 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   85    0.00  0.02  0.85  0.00  0.01  0.02
03:46:15 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   85    0.00  0.02  0.85  0.00  0.01  0.02
03:46:16 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   86    0.00  0.02  0.85  0.00  0.01  0.02
03:46:16 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   86    0.00  0.02  0.85  0.00  0.01  0.02
03:46:17 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   87    0.00  0.02  0.85  0.00  0.01  0.02
03:46:17 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   87    0.00  0.02  0.85  0.00  0.01  0.02
03:46:18 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   88    0.00  0.02  0.85  0.00  0.01  0.02
03:46:18 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   88    0.00  0.02  0.85  0.00  0.01  0.02
03:46:19 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   89    0.00  0.02  0.85  0.00  0.01  0.02
03:46:19 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   89    0.00  0.02  0.85  0.00  0.01  0.02
03:46:20 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   90    0.00  0.02  0.85  0.00  0.01  0.02
03:46:20 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   90    0.00  0.02  0.85  0.00  0.01  0.02
03:46:21 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   91    0.00  0.02  0.85  0.00  0.01  0.02
03:46:21 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   91    0.00  0.02  0.85  0.00  0.01  0.02
03:46:22 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   92    0.00  0.02  0.85  0.00  0.01  0.02
03:46:22 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   92    0.00  0.02  0.85  0.00  0.01  0.02
03:46:23 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   93    0.00  0.02  0.85  0.00  0.01  0.02
03:46:23 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   93    0.00  0.02  0.85  0.00  0.01  0.02
03:46:24 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   94    0.00  0.02  0.85  0.00  0.01  0.02
03:46:24 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   94    0.00  0.02  0.85  0.00  0.01  0.02
03:46:25 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   95    0.00  0.02  0.85  0.00  0.01  0.02
03:46:25 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   95    0.00  0.02  0.85  0.00  0.01  0.02
03:46:26 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   96    0.00  0.02  0.85  0.00  0.01  0.02
03:46:26 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   96    0.00  0.02  0.85  0.00  0.01  0.02
03:46:27 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   97    0.00  0.02  0.85  0.00  0.01  0.02
03:46:27 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   97    0.00  0.02  0.85  0.00  0.01  0.02
03:46:28 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   98    0.00  0.02  0.85  0.00  0.01  0.02
03:46:28 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   98    0.00  0.02  0.85  0.00  0.01  0.02
03:46:29 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   99    0.00  0.02  0.85  0.00  0.01  0.02
03:46:29 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   99    0.00  0.02  0.85  0.00  0.01  0.02
03:46:30 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   100   0.00  0.02  0.85  0.00  0.01  0.02
03:46:30 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   100   0.00  0.02  0.85  0.00  0.01  0.02
03:46:31 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   101   0.00  0.02  0.85  0.00  0.01  0.02
03:46:31 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   101   0.00  0.02  0.85  0.00  0.01  0.02
03:46:32 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   102   0.00  0.02  0.85  0.00  0.01  0.02
03:46:32 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   102   0.00  0.02  0.85  0.00  0.01  0.02
03:46:33 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   103   0.00  0.02  0.85  0.00  0.01  0.02
03:46:33 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   103   0.00  0.02  0.85  0.00  0.01  0.02
03:46:34 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   104   0.00  0.02  0.85  0.00  0.01  0.02
03:46:34 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   104   0.00  0.02  0.85  0.00  0.01  0.02
03:46:35 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   105   0.00  0.02  0.85  0.00  0.01  0.02
03:46:35 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   105   0.00  0.02  0.85  0.00  0.01  0.02
03:46:36 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   106   0.00  0.02  0.85  0.00  0.01  0.02
03:46:36 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   106   0.00  0.02  0.85  0.00  0.01  0.02
03:46:37 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   107   0.00  0.02  0.85  0.00  0.01  0.02
03:46:37 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   107   0.00  0.02  0.85  0.00  0.01  0.02
03:46:38 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   108   0.00  0.02  0.85  0.00  0.01  0.02
03:46:38 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   108   0.00  0.02  0.85  0.00  0.01  0.02
03:46:39 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   109   0.00  0.02  0.85  0.00  0.01  0.02
03:46:39 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   109   0.00  0.02  0.85  0.00  0.01  0.02
03:46:40 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   110   0.00  0.02  0.85  0.00  0.01  0.02
03:46:40 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   110   0.00  0.02  0.85  0.00  0.01  0.02
03:46:41 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   111   0.00  0.02  0.85  0.00  0.01  0.02
03:46:41 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   111   0.00  0.02  0.85  0.00  0.01  0.02
03:46:42 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   112   0.00  0.02  0.85  0.00  0.01  0.02
03:46:42 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   112   0.00  0.02  0.85  0.00  0.01  0.02
03:46:43 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   112   0.00  0.02  0.85  0.00  0.01  0.02
03:46:43 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   112   0.00  0.02  0.85  0.00  0.01  0.02
03:46:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   114   0.00  0.03  0.81  0.00  0.02  0.02
03:46:44 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   114   0.00  0.03  0.81  0.00  0.02  0.02
03:46:45 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   115   0.00  0.02  0.84  0.00  0.01  0.02
03:46:45 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   115   0.00  0.02  0.84  0.00  0.01  0.02
03:46:46 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   116   0.00  0.02  0.84  0.00  0.01  0.02
03:46:46 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   116   0.00  0.02  0.84  0.00  0.01  0.02
03:46:47 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   117   0.00  0.02  0.85  0.00  0.01  0.02
03:46:47 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   117   0.00  0.02  0.85  0.00  0.01  0.02
03:46:48 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   118   0.00  0.02  0.85  0.00  0.01  0.02
03:46:48 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   118   0.00  0.02  0.85  0.00  0.01  0.02
03:46:49 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   119   0.00  0.02  0.85  0.00  0.01  0.02
03:46:49 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   119   0.00  0.02  0.85  0.00  0.01  0.02
03:46:50 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   120   0.00  0.02  0.85  0.00  0.01  0.02
03:46:50 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   120   0.00  0.02  0.85  0.00  0.01  0.02
03:46:51 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   121   0.00  0.02  0.85  0.00  0.01  0.02
03:46:51 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   121   0.00  0.02  0.85  0.00  0.01  0.02
03:46:52 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   122   0.00  0.02  0.85  0.00  0.01  0.02
03:46:52 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   122   0.00  0.02  0.85  0.00  0.01  0.02
03:46:53 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   123   0.00  0.02  0.85  0.00  0.01  0.02
03:46:53 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   123   0.00  0.02  0.85  0.00  0.01  0.02
03:46:54 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   123   0.00  0.02  0.85  0.00  0.01  0.02
03:46:54 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   123   0.00  0.02  0.85  0.00  0.01  0.02
03:46:55 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   124   0.00  0.02  0.85  0.00  0.01  0.02
03:46:55 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   124   0.00  0.02  0.85  0.00  0.01  0.02
03:46:56 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   125   0.00  0.02  0.85  0.00  0.01  0.02
03:46:56 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   125   0.00  0.02  0.85  0.00  0.01  0.02
03:46:57 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   126   0.00  0.02  0.85  0.00  0.01  0.02
03:46:57 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   126   0.00  0.02  0.85  0.00  0.01  0.02
03:46:58 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   127   0.00  0.02  0.85  0.00  0.01  0.02
03:46:58 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   127   0.00  0.02  0.85  0.00  0.01  0.02
03:46:59 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   129   0.00  0.02  0.85  0.00  0.01  0.02
03:46:59 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   129   0.00  0.02  0.85  0.00  0.01  0.02
03:47:00 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   129   0.00  0.02  0.85  0.00  0.01  0.02
03:47:00 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   129   0.00  0.02  0.85  0.00  0.01  0.02
03:47:01 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   130   0.00  0.02  0.85  0.00  0.01  0.02
03:47:01 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   130   0.00  0.02  0.85  0.00  0.01  0.02
03:47:02 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   131   0.00  0.02  0.85  0.00  0.01  0.02
03:47:02 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   131   0.00  0.02  0.85  0.00  0.01  0.02
03:47:03 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   132   0.00  0.02  0.85  0.00  0.01  0.02
03:47:03 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   132   0.00  0.02  0.85  0.00  0.01  0.02
03:47:04 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   133   0.00  0.02  0.85  0.00  0.01  0.02
03:47:04 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   133   0.00  0.02  0.85  0.00  0.01  0.02
03:47:05 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   134   0.00  0.02  0.85  0.00  0.01  0.02
03:47:05 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   134   0.00  0.02  0.85  0.00  0.01  0.02
03:47:06 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   134   0.00  0.02  0.85  0.00  0.01  0.02
03:47:07 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   136   0.00  0.02  0.85  0.00  0.01  0.02
03:47:07 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   136   0.00  0.02  0.85  0.00  0.01  0.02
03:47:07 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   136   0.00  0.02  0.85  0.00  0.01  0.02
03:47:07 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   136   0.00  0.02  0.85  0.00  0.01  0.02
03:47:08 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   137   0.00  0.02  0.85  0.00  0.01  0.02
03:47:08 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   137   0.00  0.02  0.85  0.00  0.01  0.02
03:47:09 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   138   0.00  0.02  0.85  0.00  0.01  0.02
03:47:09 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   138   0.00  0.02  0.85  0.00  0.01  0.02
03:47:10 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   139   0.00  0.02  0.85  0.00  0.01  0.02
03:47:10 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   139   0.00  0.02  0.85  0.00  0.01  0.02
03:47:11 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   140   0.00  0.02  0.85  0.00  0.01  0.02
03:47:11 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   140   0.00  0.02  0.85  0.00  0.01  0.02
03:47:12 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   141   0.00  0.02  0.85  0.00  0.01  0.02
03:47:12 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   141   0.00  0.02  0.85  0.00  0.01  0.02
03:47:13 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   142   0.00  0.02  0.85  0.00  0.01  0.02
03:47:13 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   142   0.00  0.02  0.85  0.00  0.01  0.02
03:47:14 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   143   0.00  0.02  0.85  0.00  0.01  0.02
03:47:14 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   143   0.00  0.02  0.85  0.00  0.01  0.02
03:47:15 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   144   0.00  0.02  0.85  0.00  0.01  0.02
03:47:15 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   144   0.00  0.02  0.85  0.00  0.01  0.02
03:47:16 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   145   0.00  0.02  0.85  0.00  0.01  0.02
03:47:16 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   145   0.00  0.02  0.85  0.00  0.01  0.02
03:47:17 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   146   0.00  0.02  0.85  0.00  0.01  0.02
03:47:17 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   146   0.00  0.02  0.85  0.00  0.01  0.02
03:47:18 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   147   0.00  0.02  0.85  0.00  0.01  0.02
03:47:18 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   147   0.00  0.02  0.85  0.00  0.01  0.02
03:47:19 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   148   0.00  0.02  0.85  0.00  0.01  0.02
03:47:19 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   148   0.00  0.02  0.85  0.00  0.01  0.02
03:47:20 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   150   0.00  0.02  0.85  0.00  0.01  0.02
03:47:20 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   150   0.00  0.02  0.85  0.00  0.01  0.02
03:47:21 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   150   0.00  0.02  0.85  0.00  0.01  0.02
03:47:21 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   150   0.00  0.02  0.85  0.00  0.01  0.02
03:47:22 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   151   0.00  0.02  0.85  0.00  0.01  0.02
03:47:22 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   151   0.00  0.02  0.85  0.00  0.01  0.02
03:47:23 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   152   0.00  0.02  0.85  0.00  0.01  0.02
03:47:23 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   152   0.00  0.02  0.85  0.00  0.01  0.02
03:47:24 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   154   0.00  0.02  0.85  0.00  0.01  0.02
03:47:24 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   154   0.00  0.02  0.85  0.00  0.01  0.02
03:47:25 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   154   0.00  0.02  0.85  0.00  0.01  0.02
03:47:25 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   154   0.00  0.02  0.85  0.00  0.01  0.02
03:47:26 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   155   0.00  0.02  0.85  0.00  0.01  0.02
03:47:26 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   155   0.00  0.02  0.85  0.00  0.01  0.02
03:47:27 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   156   0.00  0.02  0.85  0.00  0.01  0.02
03:47:27 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   156   0.00  0.02  0.85  0.00  0.01  0.02
03:47:27 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   156   0.00  0.02  0.85  0.00  0.01  0.02
03:47:27 CABB.0   CABB00248549  sit     0    NoReport sit                room -    OpenFloor  2   156   0.00  0.02  0.85  0.00  0.01  0.02
03:47:27 CABB.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  2   156   0.00  0.02  0.85  0.00  0.01  0.02
03:47:28 CABB.E   -             -       0    NoReport np=1               room -    OpenFloor  2   156   0.00  0.02  0.85  0.00  0.01  0.02
03:47:28 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  2   157   0.00  0.04  0.74  0.00  0.02  0.04
03:47:29 CABB.1   CABB14527315  stand   0    NoReport stand              trk  0.55 OpenFloor  1   121   0.01  0.05  0.68  0.01  0.03  0.04
03:47:30 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   122   0.01  0.05  0.65  0.01  0.04  0.03
03:47:31 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   123   0.01  0.05  0.63  0.01  0.04  0.03
03:47:32 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   124   0.01  0.05  0.62  0.02  0.05  0.03
03:47:33 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   125   0.01  0.05  0.62  0.02  0.05  0.03
03:47:34 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   125   0.01  0.05  0.62  0.02  0.05  0.03
03:47:35 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   127   0.01  0.05  0.62  0.02  0.05  0.03
03:47:35 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   128   0.01  0.05  0.61  0.02  0.05  0.03
03:47:36 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   128   0.01  0.05  0.61  0.02  0.05  0.03
03:47:37 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   130   0.01  0.05  0.61  0.02  0.05  0.03
03:47:37 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   130   0.01  0.05  0.61  0.02  0.05  0.03
03:47:38 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   130   0.01  0.05  0.61  0.02  0.05  0.03
03:47:39 CABB.1   CABB14527315  stand   101  NoReport stand              trk  1.00 OpenFloor  1   131   0.00  0.05  0.61  0.02  0.05  0.03
03:47:40 CABB.1   CABB14527315  stand   104  NoReport stand              trk  1.00 OpenFloor  1   132   0.00  0.05  0.61  0.02  0.05  0.03
03:47:41 CABB.1   CABB14527315  stand   111  NoReport stand              trk  1.00 OpenFloor  1   133   0.00  0.05  0.61  0.02  0.05  0.03
03:47:42 CABB.1   CABB14527315  stand   129  NoReport stand              trk  1.00 OpenFloor  1   134   0.00  0.05  0.61  0.02  0.05  0.03
03:47:43 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   136   0.00  0.05  0.61  0.02  0.05  0.03
03:47:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   136   0.01  0.05  0.61  0.02  0.05  0.03
03:47:45 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   137   0.01  0.05  0.61  0.02  0.05  0.03
03:47:46 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   137   0.01  0.05  0.61  0.02  0.05  0.03
03:47:47 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   140   0.01  0.05  0.61  0.02  0.05  0.03
03:47:47 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   140   0.01  0.05  0.61  0.02  0.05  0.03
03:47:48 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   140   0.01  0.05  0.61  0.02  0.05  0.03
03:47:49 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   141   0.01  0.05  0.61  0.02  0.05  0.03
03:47:50 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   142   0.01  0.05  0.61  0.02  0.05  0.03
03:47:51 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   143   0.01  0.05  0.61  0.02  0.05  0.03
03:47:52 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   144   0.01  0.05  0.61  0.02  0.05  0.03
03:47:53 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   145   0.01  0.05  0.61  0.02  0.05  0.03
03:47:54 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   146   0.01  0.05  0.61  0.02  0.05  0.03
03:47:55 CABB.1   CABB14527315  stand   124  NoReport stand              trk  1.00 OpenFloor  1   147   0.00  0.05  0.61  0.02  0.05  0.03
03:47:56 CABB.1   CABB14527315  stand   105  NoReport stand              trk  1.00 OpenFloor  1   148   0.00  0.05  0.61  0.02  0.05  0.03
03:47:57 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   149   0.01  0.05  0.61  0.02  0.05  0.03
03:47:58 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   150   0.01  0.05  0.61  0.02  0.05  0.03
03:47:59 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   151   0.01  0.05  0.61  0.02  0.05  0.03
03:48:00 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   152   0.01  0.05  0.61  0.02  0.05  0.03
03:48:01 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   152   0.01  0.05  0.61  0.02  0.05  0.03
03:48:02 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   154   0.01  0.05  0.61  0.02  0.05  0.03
03:48:02 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   154   0.01  0.05  0.61  0.02  0.05  0.03
03:48:02 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   155   0.01  0.05  0.61  0.02  0.05  0.03
03:48:03 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   156   0.01  0.05  0.61  0.02  0.05  0.03
03:48:04 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   157   0.01  0.05  0.61  0.02  0.05  0.03
03:48:05 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   158   0.01  0.05  0.61  0.02  0.05  0.03
03:48:06 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   159   0.01  0.05  0.61  0.02  0.05  0.03
03:48:07 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   160   0.01  0.05  0.61  0.02  0.05  0.03
03:48:08 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   161   0.01  0.05  0.61  0.02  0.05  0.03
03:48:09 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   162   0.01  0.05  0.61  0.02  0.05  0.03
03:48:10 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   163   0.01  0.05  0.61  0.02  0.05  0.03
03:48:11 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   164   0.01  0.05  0.61  0.02  0.05  0.03
03:48:12 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   165   0.01  0.05  0.61  0.02  0.05  0.03
03:48:13 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   166   0.01  0.05  0.61  0.02  0.05  0.03
03:48:14 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   167   0.01  0.05  0.61  0.02  0.05  0.03
03:48:15 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   168   0.01  0.05  0.61  0.02  0.05  0.03
03:48:16 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   169   0.01  0.05  0.61  0.02  0.05  0.03
03:48:17 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   170   0.01  0.05  0.61  0.02  0.05  0.03
03:48:18 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   171   0.01  0.05  0.61  0.02  0.05  0.03
03:48:19 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   172   0.01  0.05  0.61  0.02  0.05  0.03
03:48:20 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   173   0.01  0.05  0.61  0.02  0.05  0.03
03:48:21 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   174   0.01  0.05  0.61  0.02  0.05  0.03
03:48:22 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   175   0.01  0.05  0.61  0.02  0.05  0.03
03:48:23 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   176   0.01  0.05  0.61  0.02  0.05  0.03
03:48:24 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   177   0.01  0.05  0.61  0.02  0.05  0.03
03:48:25 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   178   0.01  0.05  0.61  0.02  0.05  0.03
03:48:26 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   178   0.01  0.05  0.61  0.02  0.05  0.03
03:48:27 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   179   0.01  0.05  0.61  0.02  0.05  0.03
03:48:27 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   180   0.01  0.05  0.61  0.02  0.05  0.03
03:48:28 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   181   0.01  0.05  0.61  0.02  0.05  0.03
03:48:29 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   182   0.01  0.05  0.61  0.02  0.05  0.03
03:48:30 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   183   0.01  0.05  0.61  0.02  0.05  0.03
03:48:31 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   184   0.01  0.05  0.61  0.02  0.05  0.03
03:48:32 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   185   0.01  0.05  0.61  0.02  0.05  0.03
03:48:33 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   185   0.01  0.05  0.61  0.02  0.05  0.03
03:48:34 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   187   0.01  0.05  0.61  0.02  0.05  0.03
03:48:34 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   187   0.01  0.05  0.61  0.02  0.05  0.03
03:48:35 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   188   0.01  0.05  0.61  0.02  0.05  0.03
03:48:36 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   189   0.01  0.05  0.61  0.02  0.05  0.03
03:48:37 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   189   0.01  0.05  0.61  0.02  0.05  0.03
03:48:38 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   189   0.01  0.05  0.61  0.02  0.05  0.03
03:48:39 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   189   0.01  0.05  0.61  0.02  0.05  0.03
03:48:40 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   189   0.01  0.05  0.61  0.02  0.05  0.03
03:48:41 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   189   0.01  0.05  0.61  0.02  0.05  0.03
03:48:42 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   189   0.01  0.05  0.61  0.02  0.05  0.03
03:48:43 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   189   0.01  0.05  0.61  0.02  0.05  0.03
03:48:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   196   0.01  0.05  0.61  0.02  0.05  0.03
03:48:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   196   0.01  0.05  0.61  0.02  0.05  0.03
03:48:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   196   0.01  0.05  0.61  0.02  0.05  0.03
03:48:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   196   0.01  0.05  0.61  0.02  0.05  0.03
03:48:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   196   0.01  0.05  0.61  0.02  0.05  0.03
03:48:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   196   0.01  0.05  0.61  0.02  0.05  0.03
03:48:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   196   0.01  0.05  0.61  0.02  0.05  0.03
03:48:44 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   197   0.01  0.05  0.61  0.02  0.05  0.03
03:48:45 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   198   0.01  0.05  0.61  0.02  0.05  0.03
03:48:46 CABB.1   CABB14527315  stand   115  NoReport stand              trk  1.00 OpenFloor  1   199   0.00  0.05  0.61  0.02  0.05  0.03
03:48:47 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   199   0.00  0.05  0.61  0.02  0.05  0.03
03:48:48 CABB.1   CABB14527315  stand   137  NoReport stand              trk  1.00 OpenFloor  1   201   0.00  0.05  0.61  0.02  0.05  0.03
03:48:48 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   201   0.01  0.05  0.61  0.02  0.05  0.03
03:48:49 CABB.1   CABB14527315  stand   110  NoReport stand              trk  1.00 OpenFloor  1   202   0.00  0.05  0.61  0.02  0.05  0.03
03:48:50 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   203   0.01  0.05  0.61  0.02  0.05  0.03
03:48:51 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   204   0.01  0.05  0.61  0.02  0.05  0.03
03:48:52 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   205   0.01  0.05  0.61  0.02  0.05  0.03
03:48:53 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   206   0.01  0.05  0.61  0.02  0.05  0.03
03:48:54 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   206   0.01  0.05  0.61  0.02  0.05  0.03
03:48:55 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   208   0.01  0.05  0.61  0.02  0.05  0.03
03:48:55 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   208   0.01  0.05  0.61  0.02  0.05  0.03
03:48:56 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   209   0.01  0.05  0.61  0.02  0.05  0.03
03:48:57 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   210   0.01  0.05  0.61  0.02  0.05  0.03
03:48:58 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   211   0.01  0.05  0.61  0.02  0.05  0.03
03:48:59 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   212   0.01  0.05  0.61  0.02  0.05  0.03
03:49:00 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   213   0.01  0.05  0.61  0.02  0.05  0.03
03:49:01 CABB.1   CABB14527315  stand   89   NoReport stand              trk  1.00 OpenFloor  1   214   0.00  0.05  0.61  0.02  0.05  0.03
03:49:02 CABB.1   CABB14527315  stand   118  NoReport stand              trk  1.00 OpenFloor  1   215   0.00  0.05  0.61  0.02  0.05  0.03
03:49:03 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   216   0.01  0.05  0.61  0.02  0.05  0.03
03:49:04 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   217   0.01  0.05  0.61  0.02  0.05  0.03
03:49:05 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   217   0.01  0.05  0.61  0.02  0.05  0.03
03:49:06 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   219   0.01  0.05  0.61  0.02  0.05  0.03
03:49:06 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   219   0.01  0.05  0.61  0.02  0.05  0.03
03:49:07 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   219   0.01  0.05  0.61  0.02  0.05  0.03
03:49:08 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   221   0.01  0.05  0.61  0.02  0.05  0.03
03:49:08 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   221   0.01  0.05  0.61  0.02  0.05  0.03
03:49:09 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   222   0.01  0.05  0.61  0.02  0.05  0.03
03:49:10 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   223   0.01  0.05  0.61  0.02  0.05  0.03
03:49:11 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   224   0.01  0.05  0.61  0.02  0.05  0.03
03:49:12 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   225   0.01  0.05  0.61  0.02  0.05  0.03
03:49:13 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   226   0.01  0.05  0.61  0.02  0.05  0.03
03:49:14 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   227   0.01  0.05  0.61  0.02  0.05  0.03
03:49:15 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   228   0.01  0.05  0.61  0.02  0.05  0.03
03:49:16 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   229   0.01  0.05  0.61  0.02  0.05  0.03
03:49:17 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   230   0.01  0.05  0.61  0.02  0.05  0.03
03:49:18 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   231   0.01  0.05  0.61  0.02  0.05  0.03
03:49:19 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   232   0.01  0.05  0.61  0.02  0.05  0.03
03:49:20 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   233   0.01  0.05  0.61  0.02  0.05  0.03
03:49:21 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   233   0.01  0.05  0.61  0.02  0.05  0.03
03:49:22 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   234   0.01  0.05  0.61  0.02  0.05  0.03
03:49:22 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   235   0.01  0.05  0.61  0.02  0.05  0.03
03:49:23 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   236   0.01  0.05  0.61  0.02  0.05  0.03
03:49:24 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   236   0.01  0.05  0.61  0.02  0.05  0.03
03:49:25 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   238   0.01  0.05  0.61  0.02  0.05  0.03
03:49:25 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   238   0.01  0.05  0.61  0.02  0.05  0.03
03:49:26 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   238   0.01  0.05  0.61  0.02  0.05  0.03
03:49:27 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   239   0.01  0.05  0.61  0.02  0.05  0.03
03:49:27 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   240   0.01  0.05  0.61  0.02  0.05  0.03
03:49:28 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   240   0.01  0.05  0.61  0.02  0.05  0.03
03:49:29 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   241   0.01  0.05  0.61  0.02  0.05  0.03
03:49:30 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   242   0.01  0.05  0.61  0.02  0.05  0.03
03:49:31 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  1   242   0.01  0.05  0.61  0.02  0.05  0.03
03:49:32 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   244   0.01  0.05  0.61  0.02  0.05  0.03
03:49:32 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   245   0.01  0.05  0.61  0.02  0.05  0.03
03:49:33 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   245   0.01  0.05  0.61  0.02  0.05  0.03
03:49:34 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   247   0.01  0.05  0.61  0.02  0.05  0.03
03:49:35 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   247   0.01  0.05  0.61  0.02  0.05  0.03
03:49:36 CABB.1   CABB14527315  stand   0    NoReport stand              trk  1.00 OpenFloor  1   248   0.01  0.05  0.61  0.02  0.05  0.03
03:49:37 CABB.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   248   0.01  0.05  0.61  0.02  0.05  0.03
03:49:37 CABB.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.01  0.08  0.44  0.03  0.08  0.05
03:49:38 CABB.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
03:49:39 CABB.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:40 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:41 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:42 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:43 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:44 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:45 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:46 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:47 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:48 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:49 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:50 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:51 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:52 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.13  0.03
03:49:53 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  0   0     0.02  0.10  0.24  0.06  0.15  0.03
03:49:54 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:49:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:49:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:49:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:49:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:49:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
03:50:23 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
03:50:55 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:50:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:50:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:50:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:50:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.09  0.18  0.02
03:51:27 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.04  0.10  0.18  0.10  0.19  0.02
03:51:59 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:30 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:52:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:53:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:53:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.17  0.11  0.19  0.02
03:53:02 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.05  0.10  0.16  0.11  0.19  0.02
03:53:34 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:37 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:53:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:06 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:09 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:10 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:11 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:12 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:13 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:14 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:15 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:16 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:17 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:18 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:19 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:20 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:21 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:22 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:23 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:24 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:25 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:26 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:27 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:28 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:29 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:30 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:31 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:32 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:33 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:34 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:35 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:36 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:37 CABB.88  -             88      -    NoReport no-target(88)      room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:38 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:39 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:40 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:41 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:42 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:43 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:44 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:45 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:46 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:47 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:48 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:49 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:50 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:51 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:52 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:53 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:54 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:55 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:56 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:57 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:58 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:54:59 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:00 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:01 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:02 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:03 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:04 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:05 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:06 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:07 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:08 -.-      -             -       -    NoReport (no frame, held)   room -    BlindOpen  0   0     0.06  0.10  0.16  0.12  0.20  0.02
03:55:09 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:42 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:55:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.07  0.10  0.16  0.12  0.20  0.02
03:56:13 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.16  0.12  0.20  0.02
03:56:45 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:56:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:16 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.08  0.10  0.15  0.12  0.20  0.02
03:57:48 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:57:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:19 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:51 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:58:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.09  0.10  0.15  0.13  0.19  0.02
03:59:23 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:57 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
03:59:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:26 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.10  0.10  0.15  0.13  0.19  0.02
04:00:59 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:02 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:30 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:33 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:48 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:49 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:50 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:51 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:52 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:53 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:54 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:55 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:56 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:57 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:58 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:01:59 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:00 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:01 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:02 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:03 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:04 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:05 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:06 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:07 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:08 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:09 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:10 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:11 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:12 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:13 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:14 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:15 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:16 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:17 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:18 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:19 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:20 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:21 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:22 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:23 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:24 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:25 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:26 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:27 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:28 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:29 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:31 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:32 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:33 CABB.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:34 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:35 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:36 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:37 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:38 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:39 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:40 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:41 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:42 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:43 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:44 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:45 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:46 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:47 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:48 CABB.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.11  0.10  0.15  0.13  0.19  0.02
04:02:48 CABB.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
04:02:48 CABB.0   CABB00248549  stand   101  NoReport stand              trk  0.92 Empty      1   0     0.00  0.03  0.15  0.00  0.79  0.04
04:02:48 CABB.0   CABB00248549  stand   84   NoReport stand              trk  0.92 Empty      1   0     0.00  0.03  0.25  0.00  0.66  0.01
04:02:49 CABB.0   CABB00248549  stand   71   NoReport stand              trk  0.92 Empty      1   1     0.00  0.04  0.34  0.01  0.51  0.02
04:02:50 CABB.0   CABB00248549  stand   79   NoReport stand              trk  0.98 Empty      1   2     0.00  0.04  0.42  0.01  0.38  0.02
04:02:51 CABB.0   CABB00248549  walk    70   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.48  0.01  0.28  0.02
04:02:52 CABB.0   CABB00248549  walk    70   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.53  0.01  0.21  0.03
04:02:53 CABB.0   CABB00248549  walk    70   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.56  0.02  0.15  0.03
04:02:54 CABB.0   CABB00248549  walk    68   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.58  0.02  0.12  0.03
04:02:55 CABB.0   CABB00248549  walk    62   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.59  0.02  0.10  0.03
04:02:56 CABB.0   CABB00248549  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.60  0.02  0.08  0.03
04:02:57 CABB.0   CABB00248549  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.07  0.40  0.02  0.09  0.04
04:02:58 CABB.0   CABB00248549  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.07  0.27  0.03  0.11  0.03
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
03:44:03.255 CABB.88   88     -    -      -      -     -    -    
03:44:34.595 CABB.88   88     -    -      -      -     -    -    
03:44:37.650 CABB.0    stand  1    80     10     95    80        
03:44:37.652 CABB.0    stand  1    60     10     88    80   20   
03:44:38.578 CABB.0    walk   1    0      40     78    80   67   
03:44:39.589 CABB.0    walk   1    -10    40     112   80   10   
03:44:40.602 CABB.0    walk   1    -20    40     114   80   10   
03:44:41.598 CABB.0    walk   1    -30    50     90    80   14   
03:44:43.637 CABB.0    walk   1    -20    50     96    80   10   
03:44:43.639 CABB.0    walk   1    0      50     93    80   20   
03:44:44.947 CABB.0    walk   1    -20    50     113   80   20   
03:44:45.627 CABB.0    walk   1    -20    50     91    80   0    
03:44:46.627 CABB.0    walk   1    -30    60     90    80   14   
03:44:47.652 CABB.0    walk   1    -40    60     56    80   10   
03:44:48.545 CABB.0    walk   1    -30    50     97    80   14   
03:44:49.556 CABB.0    walk   1    -40    60     70    80   14   
03:44:50.559 CABB.0    walk   1    0      40     70    80   44   
03:44:51.558 CABB.0    walk   1    30     10     78    80   42   
03:44:52.558 CABB.0    walk   1    50     -10    92    80   28   
03:44:53.569 CABB.0    walk   1    40     0      47    80   14   
03:44:54.569 CABB.0    walk   1    40     10     0     80   10   
03:44:55.558 CABB.0    sit    1    40     10     0     80   0    
03:44:56.544 CABB.0    sit    1    40     10     0     80   0    
03:44:57.530 CABB.0    sit    1    40     10     0     80   0    
03:44:58.520 CABB.0    sit    1    40     10     0     80   0    
03:44:59.395 CABB.0    sit    1    40     10     0     80   0    
03:45:00.396 CABB.0    sit    1    40     10     0     80   0    
03:45:01.397 CABB.0    sit    1    40     10     0     80   0    
03:45:02.398 CABB.0    sit    1    40     10     0     80   0    
03:45:03.964 CABB.0    sit    1    40     10     0     80   0    
03:45:04.400 CABB.0    sit    1    40     10     0     80   0    
03:45:05.406 CABB.0    sit    1    40     10     0     80   0    
03:45:06.402 CABB.0    sit    1    40     10     0     80   0    
03:45:07.403 CABB.0    sit    1    40     10     0     80   0    
03:45:08.404 CABB.0    sit    1    40     10     0     80   0    
03:45:09.408 CABB.0    sit    1    40     10     0     80   0    
03:45:10.408 CABB.0    sit    1    40     10     0     80   0    
03:45:11.300 CABB.0    sit    1    40     10     0     80   0    
03:45:12.305 CABB.0    sit    1    40     10     0     80   0    
03:45:13.301 CABB.0    sit    1    40     10     0     80   0    
03:45:14.306 CABB.0    sit    1    40     10     0     80   0    
03:45:15.305 CABB.0    sit    1    40     10     0     80   0    
03:45:16.304 CABB.0    sit    1    40     10     0     80   0    
03:45:17.312 CABB.0    sit    1    40     10     0     80   0    
03:45:18.310 CABB.0    sit    1    40     10     0     80   0    
03:45:19.308 CABB.0    sit    1    40     10     0     80   0    
03:45:20.311 CABB.0    sit    1    40     10     0     80   0    
03:45:21.312 CABB.0    sit    1    40     10     0     80   0    
03:45:22.310 CABB.0    sit    1    40     10     0     80   0    
03:45:23.204 CABB.0    sit    1    40     10     0     80   0    
03:45:24.212 CABB.0    sit    1    40     10     0     80   0    
03:45:25.206 CABB.0    sit    1    40     10     0     80   0    
03:45:27.004 CABB.0    sit    1    40     10     0     80   0    
03:45:27.315 CABB.1    stand  255  -110   70     0     80   161  
03:45:27.315 CABB.0    sit    1    40     10     0     80   161  
03:45:28.220 CABB.1    stand  255  -110   70     0     80   161  
03:45:28.220 CABB.0    sit    1    40     10     0     80   161  
03:45:29.221 CABB.0    sit    1    40     10     0     80   0    
03:45:29.221 CABB.1    stand  255  -110   70     0     80   161  
03:45:30.220 CABB.1    stand  255  -110   70     0     80   0    
03:45:30.220 CABB.0    sit    1    40     10     0     80   161  
03:45:31.220 CABB.0    sit    1    40     10     0     80   0    
03:45:31.220 CABB.1    stand  255  -110   70     0     80   161  
03:45:32.221 CABB.1    stand  255  -110   70     0     80   0    
03:45:32.221 CABB.0    sit    1    40     10     0     80   161  
03:45:33.229 CABB.0    sit    1    40     10     0     80   0    
03:45:33.229 CABB.1    stand  255  -110   60     0     80   158  
03:45:34.118 CABB.1    stand  255  -110   60     0     80   0    
03:45:34.118 CABB.0    sit    1    40     10     0     80   158  
03:45:35.116 CABB.0    sit    1    40     10     0     80   0    
03:45:35.116 CABB.1    stand  255  -110   60     0     80   158  
03:45:36.117 CABB.1    stand  255  -110   60     0     80   0    
03:45:36.117 CABB.0    sit    1    40     10     0     80   158  
03:45:37.122 CABB.0    sit    1    40     10     0     80   0    
03:45:37.122 CABB.1    stand  255  -110   60     0     80   158  
03:45:38.120 CABB.1    stand  255  -110   60     0     80   0    
03:45:38.120 CABB.0    sit    1    40     10     0     80   158  
03:45:39.122 CABB.1    stand  255  -110   60     0     80   158  
03:45:39.122 CABB.0    sit    1    40     10     0     80   158  
03:45:40.123 CABB.0    sit    1    40     10     0     80   0    
03:45:40.123 CABB.1    stand  255  -110   60     0     80   158  
03:45:41.144 CABB.0    sit    1    40     10     0     80   158  
03:45:41.144 CABB.1    stand  255  -110   60     0     80   158  
03:45:42.150 CABB.1    stand  255  -110   60     0     80   0    
03:45:42.150 CABB.0    sit    1    40     10     0     80   158  
03:45:43.145 CABB.0    sit    1    40     10     0     80   0    
03:45:43.145 CABB.1    stand  255  -110   60     0     80   158  
03:45:44.047 CABB.1    stand  255  -110   60     0     80   0    
03:45:44.047 CABB.0    sit    1    40     10     0     80   158  
03:45:45.345 CABB.1    stand  255  -110   60     0     80   158  
03:45:45.345 CABB.0    sit    1    40     10     0     80   158  
03:45:46.053 CABB.1    stand  255  -110   60     0     80   158  
03:45:46.053 CABB.0    sit    1    40     10     0     80   158  
03:45:47.047 CABB.1    stand  255  -110   60     0     80   158  
03:45:47.047 CABB.0    sit    1    40     10     0     80   158  
03:45:48.039 CABB.0    sit    1    40     10     0     80   0    
03:45:48.039 CABB.1    stand  255  -110   60     0     80   158  
03:45:49.045 CABB.1    stand  255  -110   60     0     80   0    
03:45:49.045 CABB.0    sit    1    40     10     0     80   158  
03:45:50.044 CABB.0    sit    1    40     10     0     80   0    
03:45:50.044 CABB.1    stand  255  -110   60     0     80   158  
03:45:51.046 CABB.1    stand  255  -110   60     0     80   0    
03:45:51.046 CABB.0    sit    1    40     10     0     80   158  
03:45:52.054 CABB.0    sit    1    40     10     0     80   0    
03:45:52.054 CABB.1    stand  255  -110   60     0     80   158  
03:45:53.056 CABB.0    sit    1    40     10     0     80   158  
03:45:53.056 CABB.1    stand  255  -110   60     0     80   158  
03:45:54.052 CABB.1    stand  255  -110   60     0     80   0    
03:45:54.052 CABB.0    sit    1    40     10     0     80   158  
03:45:55.050 CABB.1    stand  255  -110   60     0     80   158  
03:45:55.050 CABB.0    sit    1    40     10     0     80   158  
03:45:55.940 CABB.1    stand  255  -110   60     0     80   158  
03:45:55.940 CABB.0    sit    1    40     10     0     80   158  
03:45:56.942 CABB.1    stand  255  -90    60     0     80   139  
03:45:56.942 CABB.0    sit    1    40     10     0     80   139  
03:45:57.961 CABB.0    sit    1    40     10     0     80   0    
03:45:57.961 CABB.1    stand  255  -90    60     0     80   139  
03:45:59.065 CABB.1    stand  255  -90    60     0     80   0    
03:45:59.065 CABB.0    sit    1    40     10     0     80   139  
03:45:59.945 CABB.0    sit    1    40     10     0     80   0    
03:45:59.945 CABB.1    stand  255  -90    60     0     80   139  
03:46:00.945 CABB.1    stand  255  -90    60     0     80   0    
03:46:00.945 CABB.0    sit    1    40     10     0     80   139  
03:46:01.950 CABB.0    sit    1    40     10     0     80   0    
03:46:01.950 CABB.1    stand  255  -90    60     0     80   139  
03:46:02.952 CABB.1    stand  255  -90    60     0     80   0    
03:46:02.952 CABB.0    sit    1    40     10     0     80   139  
03:46:04.484 CABB.0    sit    1    40     10     0     80   0    
03:46:04.484 CABB.1    stand  255  -90    60     0     80   139  
03:46:04.950 CABB.1    stand  255  -90    60     0     80   0    
03:46:04.950 CABB.0    sit    1    40     10     0     80   139  
03:46:05.959 CABB.0    sit    1    40     10     0     80   0    
03:46:05.959 CABB.1    stand  255  -90    60     0     80   139  
03:46:06.959 CABB.1    stand  255  -90    60     0     80   0    
03:46:06.959 CABB.0    sit    1    40     10     0     80   139  
03:46:07.856 CABB.1    stand  255  -90    60     0     80   139  
03:46:07.856 CABB.0    sit    1    40     10     0     80   139  
03:46:08.844 CABB.1    stand  255  -90    60     0     80   139  
03:46:08.844 CABB.0    sit    1    40     10     0     80   139  
03:46:09.846 CABB.1    stand  255  -90    60     0     80   139  
03:46:09.846 CABB.0    sit    1    40     10     0     80   139  
03:46:10.846 CABB.1    stand  255  -90    60     0     80   139  
03:46:10.846 CABB.0    sit    1    40     10     0     80   139  
03:46:11.850 CABB.0    sit    1    40     10     0     80   0    
03:46:11.850 CABB.1    stand  255  -90    60     0     80   139  
03:46:12.804 CABB.1    stand  255  -90    60     0     80   0    
03:46:12.804 CABB.0    sit    1    40     10     0     80   139  
03:46:13.809 CABB.0    sit    1    40     10     0     80   0    
03:46:13.809 CABB.1    stand  255  -90    60     0     80   139  
03:46:14.811 CABB.1    stand  255  -90    60     0     80   0    
03:46:14.811 CABB.0    sit    1    40     10     0     80   139  
03:46:15.806 CABB.0    sit    1    40     10     0     80   0    
03:46:15.806 CABB.1    stand  255  -90    60     0     80   139  
03:46:16.808 CABB.1    stand  255  -90    60     0     80   0    
03:46:16.808 CABB.0    sit    1    40     10     0     80   139  
03:46:17.812 CABB.0    sit    1    40     10     0     80   0    
03:46:17.812 CABB.1    stand  255  -90    60     0     80   139  
03:46:18.809 CABB.1    stand  255  -90    60     0     80   0    
03:46:18.809 CABB.0    sit    1    40     10     0     80   139  
03:46:19.812 CABB.0    sit    1    40     10     0     80   0    
03:46:19.812 CABB.1    stand  255  -90    60     0     80   139  
03:46:20.812 CABB.1    stand  255  -90    60     0     80   0    
03:46:20.812 CABB.0    sit    1    40     10     0     80   139  
03:46:21.812 CABB.0    sit    1    40     10     0     80   0    
03:46:21.812 CABB.1    stand  255  -90    60     0     80   139  
03:46:22.813 CABB.0    sit    1    40     10     0     80   139  
03:46:22.813 CABB.1    stand  255  -90    60     0     80   139  
03:46:23.820 CABB.1    stand  255  -90    60     0     80   0    
03:46:23.820 CABB.0    sit    1    40     10     0     80   139  
03:46:24.707 CABB.1    stand  255  -90    60     0     80   139  
03:46:24.707 CABB.0    sit    1    40     10     0     80   139  
03:46:25.722 CABB.0    sit    1    40     10     0     80   0    
03:46:25.722 CABB.1    stand  255  -90    60     0     80   139  
03:46:26.711 CABB.1    stand  255  -90    60     0     80   0    
03:46:26.711 CABB.0    sit    1    40     10     0     80   139  
03:46:27.712 CABB.0    sit    1    40     10     0     80   0    
03:46:27.712 CABB.1    stand  255  -90    60     0     80   139  
03:46:28.717 CABB.1    stand  255  -90    60     0     80   0    
03:46:28.717 CABB.0    sit    1    40     10     0     80   139  
03:46:29.762 CABB.0    sit    1    40     10     0     80   0    
03:46:29.762 CABB.1    stand  255  -90    60     0     80   139  
03:46:30.759 CABB.1    stand  255  -90    60     0     80   0    
03:46:30.759 CABB.0    sit    1    40     10     0     80   139  
03:46:31.657 CABB.0    sit    1    40     10     0     80   0    
03:46:31.657 CABB.1    stand  255  -90    60     0     80   139  
03:46:32.654 CABB.1    stand  255  -90    60     0     80   0    
03:46:32.654 CABB.0    sit    1    40     10     0     80   139  
03:46:33.655 CABB.0    sit    1    40     10     0     80   0    
03:46:33.655 CABB.1    stand  255  -90    60     0     80   139  
03:46:34.656 CABB.1    stand  255  -90    60     0     80   0    
03:46:34.656 CABB.0    sit    1    40     10     0     80   139  
03:46:35.660 CABB.0    sit    1    40     10     0     80   0    
03:46:35.660 CABB.1    stand  255  -90    60     0     80   139  
03:46:36.657 CABB.1    stand  255  -90    60     0     80   0    
03:46:36.657 CABB.0    sit    1    40     10     0     80   139  
03:46:37.660 CABB.1    stand  255  -90    60     0     80   139  
03:46:37.660 CABB.0    sit    1    40     10     0     80   139  
03:46:38.659 CABB.1    stand  255  -90    60     0     80   139  
03:46:38.659 CABB.0    sit    1    40     10     0     80   139  
03:46:39.662 CABB.1    stand  255  -90    60     0     80   139  
03:46:39.662 CABB.0    sit    1    40     10     0     80   139  
03:46:40.667 CABB.1    stand  255  -90    60     0     80   139  
03:46:40.667 CABB.0    sit    1    40     10     0     80   139  
03:46:41.669 CABB.0    sit    1    40     10     0     80   0    
03:46:41.669 CABB.1    stand  255  -90    60     0     80   139  
03:46:42.668 CABB.1    stand  255  -90    60     0     80   0    
03:46:42.668 CABB.0    sit    1    40     10     0     80   139  
03:46:43.556 CABB.0    sit    1    40     10     0     80   0    
03:46:43.556 CABB.1    stand  255  -90    60     0     80   139  
03:46:44.829 CABB.1    stand  255  -90    60     0     80   0    
03:46:44.829 CABB.0    sit    1    40     10     0     80   139  
03:46:45.568 CABB.1    stand  255  -90    60     0     80   139  
03:46:45.568 CABB.0    sit    1    40     10     0     80   139  
03:46:46.570 CABB.1    stand  255  -90    60     0     80   139  
03:46:46.570 CABB.0    sit    1    40     10     0     80   139  
03:46:47.571 CABB.0    sit    1    40     10     0     80   0    
03:46:47.571 CABB.1    stand  255  -90    60     0     80   139  
03:46:48.571 CABB.1    stand  255  -90    60     0     80   0    
03:46:48.571 CABB.0    sit    1    40     10     0     80   139  
03:46:49.576 CABB.1    stand  255  -90    60     0     80   139  
03:46:49.576 CABB.0    sit    1    40     10     0     80   139  
03:46:50.573 CABB.1    stand  255  -90    60     0     80   139  
03:46:50.573 CABB.0    sit    1    40     10     0     80   139  
03:46:51.574 CABB.1    stand  255  -90    60     0     80   139  
03:46:51.574 CABB.0    sit    1    40     10     0     80   139  
03:46:52.577 CABB.0    sit    1    40     10     0     80   0    
03:46:52.577 CABB.1    stand  255  -90    60     0     80   139  
03:46:53.579 CABB.0    sit    1    40     10     0     80   139  
03:46:53.579 CABB.1    stand  255  -90    60     0     80   139  
03:46:54.470 CABB.1    stand  255  -90    60     0     80   0    
03:46:54.470 CABB.0    sit    1    40     10     0     80   139  
03:46:55.469 CABB.0    sit    1    40     10     0     80   0    
03:46:55.469 CABB.1    stand  255  -90    60     0     80   139  
03:46:56.473 CABB.1    stand  255  -90    60     0     80   0    
03:46:56.473 CABB.0    sit    1    40     10     0     80   139  
03:46:57.473 CABB.0    sit    1    40     10     0     80   0    
03:46:57.473 CABB.1    stand  255  -90    60     0     80   139  
03:46:58.472 CABB.0    sit    1    40     10     0     80   139  
03:46:58.472 CABB.1    stand  255  -90    60     0     80   139  
03:46:59.576 CABB.0    sit    1    40     10     0     80   139  
03:46:59.576 CABB.1    stand  255  -90    60     0     80   139  
03:47:00.477 CABB.0    sit    1    40     10     0     80   139  
03:47:00.477 CABB.1    stand  255  -90    60     0     80   139  
03:47:01.479 CABB.0    sit    1    40     10     0     80   139  
03:47:01.479 CABB.1    stand  255  -90    60     0     80   139  
03:47:02.479 CABB.0    sit    1    40     10     0     80   139  
03:47:02.479 CABB.1    stand  255  -90    60     0     80   139  
03:47:03.481 CABB.1    stand  255  -90    60     0     80   0    
03:47:03.481 CABB.0    sit    1    40     10     0     80   139  
03:47:04.481 CABB.1    stand  255  -90    60     0     80   139  
03:47:04.481 CABB.0    sit    1    40     10     0     80   139  
03:47:05.482 CABB.1    stand  255  -90    60     0     80   139  
03:47:05.482 CABB.0    sit    1    40     10     0     80   139  
03:47:07.461 CABB.1    stand  255  -90    60     0     80   139  
03:47:07.461 CABB.0    sit    1    40     10     0     80   139  
03:47:07.463 CABB.1    stand  255  -90    60     0     80   139  
03:47:07.463 CABB.0    sit    1    40     10     0     80   139  
03:47:08.375 CABB.1    stand  255  -90    60     0     80   139  
03:47:08.375 CABB.0    sit    1    40     10     0     80   139  
03:47:09.376 CABB.0    sit    1    40     10     0     80   0    
03:47:09.376 CABB.1    stand  255  -90    60     0     80   139  
03:47:10.376 CABB.1    stand  255  -90    60     0     80   0    
03:47:10.376 CABB.0    sit    1    40     10     0     80   139  
03:47:11.377 CABB.1    stand  255  -90    60     0     80   139  
03:47:11.377 CABB.0    sit    1    40     10     0     80   139  
03:47:12.378 CABB.0    sit    1    40     10     0     80   0    
03:47:12.378 CABB.1    stand  255  -90    60     0     80   139  
03:47:13.385 CABB.1    stand  255  -90    60     0     80   0    
03:47:13.385 CABB.0    sit    1    40     10     0     80   139  
03:47:14.380 CABB.1    stand  255  -90    60     0     80   139  
03:47:14.380 CABB.0    sit    1    40     10     0     80   139  
03:47:15.382 CABB.1    stand  255  -90    60     0     80   139  
03:47:15.382 CABB.0    sit    1    40     10     0     80   139  
03:47:16.383 CABB.1    stand  255  -90    60     0     80   139  
03:47:16.383 CABB.0    sit    1    40     10     0     80   139  
03:47:17.284 CABB.1    stand  255  -90    60     0     80   139  
03:47:17.284 CABB.0    sit    1    40     10     0     80   139  
03:47:18.288 CABB.1    stand  255  -90    60     0     80   139  
03:47:18.288 CABB.0    sit    1    40     10     0     80   139  
03:47:19.286 CABB.1    stand  255  -90    60     0     80   139  
03:47:19.286 CABB.0    sit    1    40     10     0     80   139  
03:47:20.977 CABB.1    stand  255  -90    60     0     80   139  
03:47:20.977 CABB.0    sit    1    40     10     0     80   139  
03:47:21.288 CABB.1    stand  255  -90    60     0     80   139  
03:47:21.288 CABB.0    sit    1    40     10     0     80   139  
03:47:22.291 CABB.1    stand  255  -90    60     0     80   139  
03:47:22.291 CABB.0    sit    1    40     10     0     80   139  
03:47:23.291 CABB.1    stand  255  -90    60     0     80   139  
03:47:23.291 CABB.0    sit    1    40     10     0     80   139  
03:47:24.971 CABB.1    stand  255  -90    60     0     80   139  
03:47:24.971 CABB.0    sit    1    40     10     0     80   139  
03:47:25.292 CABB.1    stand  255  -90    60     0     80   139  
03:47:25.292 CABB.0    sit    1    40     10     0     80   139  
03:47:26.293 CABB.1    stand  255  -90    60     0     80   139  
03:47:26.293 CABB.0    sit    1    40     10     0     80   139  
03:47:27.296 CABB.0    sit    1    40     10     0     80   0    
03:47:27.296 CABB.1    stand  255  -90    60     0     80   139  
03:47:27.534 CABB.1    stand  255  -90    60     0     80   0    
03:47:27.534 CABB.0    sit    1    40     10     0     80   139  
03:47:28.453 CABB.1    stand  255  -90    60     0     80   139  
03:47:29.208 CABB.1    stand  255  -90    60     0     80   0    
03:47:30.207 CABB.1    stand  255  -90    60     0     80   0    
03:47:31.207 CABB.1    stand  255  -90    60     0     80   0    
03:47:32.212 CABB.1    stand  255  -90    60     0     80   0    
03:47:33.209 CABB.1    stand  255  -90    60     0     80   0    
03:47:35.007 CABB.1    stand  255  -90    60     0     80   0    
03:47:35.315 CABB.1    stand  255  -90    60     0     80   0    
03:47:37.976 CABB.1    stand  255  -90    60     0     80   0    
03:47:37.978 CABB.1    stand  255  -90    60     0     80   0    
03:47:38.284 CABB.1    stand  255  -90    60     0     80   0    
03:47:39.115 CABB.1    stand  255  -90    50     101   80   10   
03:47:40.114 CABB.1    stand  255  -80    40     104   80   14   
03:47:41.119 CABB.1    stand  255  -90    40     111   80   10   
03:47:42.118 CABB.1    stand  255  -80    40     129   80   10   
03:47:43.404 CABB.1    stand  255  -80    30     0     80   10   
03:47:44.119 CABB.1    stand  255  -80    30     0     80   0    
03:47:45.120 CABB.1    stand  255  -80    30     0     80   0    
03:47:47.501 CABB.1    stand  255  -80    30     0     80   0    
03:47:47.503 CABB.1    stand  255  -80    30     0     80   0    
03:47:48.123 CABB.1    stand  255  -80    30     0     80   0    
03:47:49.029 CABB.1    stand  255  -80    30     0     80   0    
03:47:50.029 CABB.1    stand  255  -80    30     0     80   0    
03:47:51.033 CABB.1    stand  255  -80    30     0     80   0    
03:47:52.041 CABB.1    stand  255  -80    30     0     80   0    
03:47:53.032 CABB.1    stand  255  -80    30     0     80   0    
03:47:54.036 CABB.1    stand  255  -80    30     0     80   0    
03:47:55.034 CABB.1    stand  255  -80    40     124   80   10   
03:47:56.038 CABB.1    stand  255  -70    30     105   80   14   
03:47:57.035 CABB.1    stand  255  -70    30     0     80   0    
03:47:58.037 CABB.1    stand  255  -70    30     0     80   0    
03:47:59.038 CABB.1    stand  255  -70    30     0     80   0    
03:48:00.198 CABB.1    stand  255  -70    30     0     80   0    
03:48:02.041 CABB.1    stand  255  -70    30     0     80   0    
03:48:02.043 CABB.1    stand  255  -70    30     0     80   0    
03:48:02.934 CABB.1    stand  255  -70    30     0     80   0    
03:48:03.935 CABB.1    stand  255  -70    30     0     80   0    
03:48:04.900 CABB.1    stand  255  -70    30     0     80   0    
03:48:05.901 CABB.1    stand  255  -70    30     0     80   0    
03:48:06.903 CABB.1    stand  255  -70    30     0     80   0    
03:48:07.914 CABB.1    stand  255  -70    30     0     80   0    
03:48:08.904 CABB.1    stand  255  -70    30     0     80   0    
03:48:09.907 CABB.1    stand  255  -70    30     0     80   0    
03:48:10.906 CABB.1    stand  255  -70    30     0     80   0    
03:48:11.908 CABB.1    stand  255  -70    30     0     80   0    
03:48:12.908 CABB.1    stand  255  -70    30     0     80   0    
03:48:13.909 CABB.1    stand  255  -70    30     0     80   0    
03:48:14.910 CABB.1    stand  255  -70    30     0     80   0    
03:48:15.912 CABB.1    stand  255  -70    30     0     80   0    
03:48:16.804 CABB.1    stand  255  -70    30     0     80   0    
03:48:17.805 CABB.1    stand  255  -70    30     0     80   0    
03:48:18.806 CABB.1    stand  255  -70    30     0     80   0    
03:48:19.807 CABB.1    stand  255  -70    30     0     80   0    
03:48:20.817 CABB.1    stand  255  -70    30     0     80   0    
03:48:21.829 CABB.1    stand  255  -70    30     0     80   0    
03:48:22.830 CABB.1    stand  255  -70    30     0     80   0    
03:48:23.834 CABB.1    stand  255  -70    30     0     80   0    
03:48:24.833 CABB.1    stand  255  -70    30     0     80   0    
03:48:25.732 CABB.1    stand  255  -70    30     0     80   0    
03:48:27.027 CABB.1    stand  255  -70    30     0     80   0    
03:48:27.736 CABB.1    stand  255  -70    30     0     80   0    
03:48:28.739 CABB.1    stand  255  -70    30     0     80   0    
03:48:29.737 CABB.1    stand  255  -70    30     0     80   0    
03:48:30.738 CABB.1    stand  255  -70    30     0     80   0    
03:48:31.738 CABB.1    stand  255  -70    30     0     80   0    
03:48:32.739 CABB.1    stand  255  -70    30     0     80   0    
03:48:34.505 CABB.1    stand  255  -70    30     0     80   0    
03:48:34.810 CABB.1    stand  255  -70    30     0     80   0    
03:48:35.743 CABB.1    stand  255  -70    30     0     80   0    
03:48:36.743 CABB.1    stand  255  -70    30     0     80   0    
03:48:44.026 CABB.1    stand  255  -70    30     0     80   0    
03:48:44.028 CABB.1    stand  255  -70    30     0     80   0    
03:48:44.029 CABB.1    stand  255  -70    30     0     80   0    
03:48:44.031 CABB.1    stand  255  -70    30     0     80   0    
03:48:44.032 CABB.1    stand  255  -70    30     0     80   0    
03:48:44.036 CABB.1    stand  255  -70    30     0     80   0    
03:48:44.038 CABB.1    stand  255  -70    30     0     80   0    
03:48:44.651 CABB.1    stand  255  -70    30     0     80   0    
03:48:45.649 CABB.1    stand  255  -70    30     0     80   0    
03:48:46.652 CABB.1    stand  255  -70    30     115   80   0    
03:48:48.532 CABB.1    stand  255  -80    40     137   80   14   
03:48:48.838 CABB.1    stand  255  -80    30     0     80   10   
03:48:49.549 CABB.1    stand  255  -80    40     110   80   10   
03:48:50.556 CABB.1    stand  255  -80    30     0     80   10   
03:48:51.552 CABB.1    stand  255  -80    30     0     80   0    
03:48:52.561 CABB.1    stand  255  -80    30     0     80   0    
03:48:53.510 CABB.1    stand  255  -80    30     0     80   0    
03:48:55.502 CABB.1    stand  255  -80    30     0     80   0    
03:48:55.802 CABB.1    stand  255  -80    30     0     80   0    
03:48:56.511 CABB.1    stand  255  -80    30     0     80   0    
03:48:57.511 CABB.1    stand  255  -80    30     0     80   0    
03:48:58.982 CABB.1    stand  255  -80    30     0     80   0    
03:48:59.513 CABB.1    stand  255  -80    30     0     80   0    
03:49:00.515 CABB.1    stand  255  -80    30     0     80   0    
03:49:01.740 CABB.1    stand  255  -70    30     89    80   10   
03:49:02.520 CABB.1    stand  255  -90    40     118   80   22   
03:49:03.518 CABB.1    stand  255  -100   30     0     80   14   
03:49:04.519 CABB.1    stand  255  -100   30     0     80   0    
03:49:06.457 CABB.1    stand  255  -100   30     0     80   0    
03:49:06.459 CABB.1    stand  255  -100   30     0     80   0    
03:49:08.501 CABB.1    stand  255  -100   30     0     80   0    
03:49:08.502 CABB.1    stand  255  -100   30     0     80   0    
03:49:09.460 CABB.1    stand  255  -100   30     0     80   0    
03:49:10.462 CABB.1    stand  255  -100   30     0     80   0    
03:49:11.463 CABB.1    stand  255  -100   30     0     80   0    
03:49:12.356 CABB.1    stand  255  -100   30     0     80   0    
03:49:13.357 CABB.1    stand  255  -100   30     0     80   0    
03:49:14.358 CABB.1    stand  255  -100   30     0     80   0    
03:49:15.360 CABB.1    stand  255  -100   30     0     80   0    
03:49:16.364 CABB.1    stand  255  -100   30     0     80   0    
03:49:17.362 CABB.1    stand  255  -100   30     0     80   0    
03:49:18.368 CABB.1    stand  255  -100   30     0     80   0    
03:49:19.365 CABB.1    stand  255  -100   30     0     80   0    
03:49:20.364 CABB.1    stand  255  -100   30     0     80   0    
03:49:22.027 CABB.1    stand  255  -100   30     0     80   0    
03:49:22.368 CABB.1    stand  255  -100   30     0     80   0    
03:49:23.368 CABB.1    stand  255  -100   30     0     80   0    
03:49:25.500 CABB.1    stand  255  -100   30     0     80   0    
03:49:25.502 CABB.1    stand  255  -100   30     0     80   0    
03:49:27.047 CABB.1    stand  255  -100   30     0     80   0    
03:49:27.351 CABB.1    stand  255  -100   30     0     80   0    
03:49:28.278 CABB.1    stand  255  -100   30     0     80   0    
03:49:29.276 CABB.1    stand  255  -100   30     0     80   0    
03:49:30.287 CABB.1    stand  255  -100   30     0     80   0    
03:49:32.056 CABB.1    stand  255  -100   30     0     80   0    
03:49:32.370 CABB.1    stand  255  -100   30     0     80   0    
03:49:33.306 CABB.1    stand  255  -100   30     0     80   0    
03:49:34.329 CABB.1    stand  255  -100   30     0     80   0    
03:49:35.205 CABB.1    stand  255  -100   30     0     80   0    
03:49:36.227 CABB.1    stand  255  -100   30     0     80   0    
03:49:37.534 CABB.88   88     -    -      -      -     -    -    
03:49:38.225 CABB.88   88     -    -      -      -     -    -    
03:49:39.229 CABB.88   88     -    -      -      -     -    -    
03:49:54.172 CABB.88   88     -    -      -      -     -    -    
03:50:23.867 CABB.88   88     -    -      -      -     -    -    
03:50:55.978 CABB.88   88     -    -      -      -     -    -    
03:51:27.348 CABB.88   88     -    -      -      -     -    -    
03:51:59.308 CABB.88   88     -    -      -      -     -    -    
03:52:30.920 CABB.88   88     -    -      -      -     -    -    
03:53:02.998 CABB.88   88     -    -      -      -     -    -    
03:53:34.417 CABB.88   88     -    -      -      -     -    -    
03:54:06.499 CABB.88   88     -    -      -      -     -    -    
03:54:37.737 CABB.88   88     -    -      -      -     -    -    
03:55:09.883 CABB.88   88     -    -      -      -     -    -    
03:55:42.552 CABB.88   88     -    -      -      -     -    -    
03:56:13.090 CABB.88   88     -    -      -      -     -    -    
03:56:45.017 CABB.88   88     -    -      -      -     -    -    
03:57:16.556 CABB.88   88     -    -      -      -     -    -    
03:57:48.607 CABB.88   88     -    -      -      -     -    -    
03:58:19.942 CABB.88   88     -    -      -      -     -    -    
03:58:51.994 CABB.88   88     -    -      -      -     -    -    
03:59:23.487 CABB.88   88     -    -      -      -     -    -    
03:59:57.533 CABB.88   88     -    -      -      -     -    -    
04:00:26.979 CABB.88   88     -    -      -      -     -    -    
04:00:59.075 CABB.88   88     -    -      -      -     -    -    
04:01:30.409 CABB.88   88     -    -      -      -     -    -    
04:02:02.564 CABB.88   88     -    -      -      -     -    -    
04:02:33.900 CABB.88   88     -    -      -      -     -    -    
04:02:48.549 CABB.0    stand  1    50     30     101   80   150  
04:02:48.850 CABB.0    stand  1    50     20     84    80   10   
04:02:49.797 CABB.0    stand  1    30     0      71    80   28   
04:02:50.785 CABB.0    stand  1    0      0      79    80   30   
04:02:51.787 CABB.0    walk   1    -10    0      70    80   10   
04:02:52.951 CABB.0    walk   1    0      0      70    80   10   
04:02:53.791 CABB.0    walk   1    0      0      70    80   0    
04:02:54.788 CABB.0    walk   1    0      0      68    80   0    
04:02:55.789 CABB.0    walk   1    0      -10    62    80   10   
04:02:56.790 CABB.0    walk   1    0      -10    0     80   0    
04:02:57.793 CABB.0    sit    1    0      -10    0     80   0    
04:02:58.791 CABB.0    sit    1    0      0      0     80   10   

```

**汇总**: xray tick 489 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

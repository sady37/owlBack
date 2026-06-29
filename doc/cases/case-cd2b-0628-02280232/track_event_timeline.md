# case-cd2b-0628-02280232 — 每 tick belief 时间线 (room fd00:0:3:112:3:100, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
02:28:00 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.50 Empty      2   0     0.00  0.09  0.24  0.00  0.64  0.03
02:28:00 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.50 Empty      2   0     0.00  0.03  0.15  0.00  0.79  0.04
02:28:00 333B.0   -             stand   0    InBed    stand              room -    Empty      2   0     0.00  0.03  0.15  0.00  0.79  0.04
02:28:01 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.51 Empty      2   1     0.00  0.19  0.43  0.00  0.33  0.01
02:28:01 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.51 Empty      2   1     0.00  0.03  0.40  0.00  0.53  0.01
02:28:01 333B.0   -             stand   0    InBed    stand              room -    Empty      2   1     0.00  0.03  0.40  0.00  0.53  0.01
02:28:02 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.52 OpenFloor  2   2     0.00  0.29  0.51  0.01  0.12  0.01
02:28:02 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.52 OpenFloor  2   2     0.00  0.03  0.63  0.00  0.26  0.01
02:28:02 333B.0   -             stand   77   InBed    stand              room -    OpenFloor  2   2     0.00  0.03  0.63  0.00  0.26  0.01
02:28:03 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.53 OpenFloor  2   3     0.00  0.03  0.76  0.01  0.08  0.02
02:28:03 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.53 OpenFloor  2   3     0.00  0.39  0.48  0.02  0.03  0.01
02:28:03 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   3     0.00  0.03  0.76  0.01  0.08  0.02
02:28:04 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   4     0.00  0.44  0.46  0.02  0.01  0.01
02:28:04 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   4     0.00  0.02  0.82  0.00  0.04  0.02
02:28:04 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   4     0.00  0.02  0.82  0.00  0.04  0.02
02:28:05 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   5     0.00  0.02  0.84  0.00  0.02  0.02
02:28:05 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   5     0.00  0.48  0.44  0.02  0.01  0.01
02:28:05 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   5     0.00  0.02  0.84  0.00  0.02  0.02
02:28:06 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   6     0.00  0.51  0.41  0.02  0.01  0.01
02:28:06 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   6     0.00  0.02  0.85  0.00  0.01  0.02
02:28:06 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   6     0.00  0.02  0.85  0.00  0.01  0.02
02:28:07 CD2B.1   CD2B12800032  stand   59   InBed    stand              trk  0.54 OpenFloor  2   7     0.00  0.53  0.39  0.02  0.00  0.01
02:28:07 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
02:28:07 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
02:28:07 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
02:28:07 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   7     0.00  0.55  0.38  0.02  0.00  0.01
02:28:08 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   7     0.00  0.02  0.85  0.00  0.01  0.02
02:28:08 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   8     0.00  0.02  0.85  0.00  0.01  0.02
02:28:08 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   8     0.00  0.56  0.36  0.02  0.00  0.01
02:28:09 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   8     0.00  0.02  0.85  0.00  0.01  0.02
02:28:09 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   9     0.00  0.58  0.35  0.02  0.00  0.01
02:28:09 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   9     0.00  0.02  0.85  0.00  0.01  0.02
02:28:10 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   9     0.00  0.02  0.85  0.00  0.01  0.02
02:28:10 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   10    0.00  0.02  0.85  0.00  0.01  0.02
02:28:10 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   10    0.00  0.59  0.34  0.02  0.00  0.01
02:28:11 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   10    0.00  0.02  0.85  0.00  0.01  0.02
02:28:11 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   11    0.00  0.02  0.85  0.00  0.01  0.02
02:28:11 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   11    0.00  0.60  0.33  0.02  0.00  0.01
02:28:12 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   11    0.00  0.02  0.85  0.00  0.01  0.02
02:28:12 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   12    0.00  0.60  0.33  0.02  0.00  0.01
02:28:12 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   12    0.00  0.02  0.85  0.00  0.01  0.02
02:28:13 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   12    0.00  0.02  0.85  0.00  0.01  0.02
02:28:13 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
02:28:13 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   13    0.00  0.61  0.32  0.02  0.00  0.01
02:28:14 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
02:28:14 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   14    0.00  0.61  0.32  0.03  0.00  0.01
02:28:14 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
02:28:15 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
02:28:15 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
02:28:15 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   15    0.00  0.62  0.32  0.03  0.00  0.01
02:28:16 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
02:28:16 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
02:28:16 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   16    0.00  0.62  0.31  0.03  0.00  0.01
02:28:17 333B.0   -             stand   0    InBed    stand              room -    OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
02:28:17 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
02:28:17 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   17    0.00  0.62  0.31  0.03  0.00  0.01
02:28:18 333B.0   -             stand   104  InBed    stand              room -    OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
02:28:18 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
02:28:18 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   18    0.00  0.62  0.31  0.03  0.00  0.01
02:28:19 333B.0   -             walk    86   InBed    walk               room -    OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
02:28:19 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
02:28:19 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   19    0.00  0.62  0.31  0.03  0.00  0.01
02:28:20 333B.0   -             walk    87   InBed    walk               room -    OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
02:28:20 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
02:28:20 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   20    0.00  0.63  0.31  0.03  0.00  0.01
02:28:21 333B.0   -             walk    97   InBed    walk               room -    OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
02:28:21 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
02:28:21 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  0.54 OpenFloor  2   21    0.00  0.63  0.31  0.03  0.00  0.01
02:28:22 333B.0   -             walk    0    InBed    walk               room -    OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
02:28:22 CD2B.E   -             -       0    InBed    np=3               room -    OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
02:28:22 CD2B.E2  CD2B22822912  -       0    InBed    EnterRoom(rdr)     room -    OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
02:28:22 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   22    0.00  0.01  0.92  0.00  0.01  0.01
02:28:22 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   22    0.00  0.65  0.32  0.01  0.00  0.00
02:28:22 CD2B.2   CD2B22822912  stand   70   InBed    stand              trk  1.00 OpenFloor  3   22    0.00  0.02  0.26  0.00  0.69  0.03
02:28:23 333B.0   -             walk    0    InBed    walk               room -    OpenFloor  3   22    0.00  0.01  0.92  0.00  0.01  0.01
02:28:23 CD2B.2   CD2B22822912  stand   98   InBed    stand              trk  1.00 OpenFloor  3   23    0.00  0.02  0.68  0.00  0.27  0.01
02:28:23 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   23    0.00  0.65  0.32  0.01  0.00  0.00
02:28:23 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   23    0.00  0.01  0.93  0.00  0.00  0.01
02:28:24 333B.0   -             sit     0    InBed    sit                room -    OpenFloor  3   23    0.00  0.01  0.93  0.00  0.00  0.01
02:28:24 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   24    0.00  0.66  0.31  0.01  0.00  0.00
02:28:24 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   24    0.00  0.01  0.93  0.00  0.00  0.01
02:28:24 CD2B.2   CD2B22822912  walk    81   InBed    walk               trk  1.00 OpenFloor  3   24    0.00  0.01  0.88  0.00  0.06  0.01
02:28:25 333B.0   -             sit     0    InBed    sit                room -    OpenFloor  3   24    0.00  0.01  0.93  0.00  0.00  0.01
02:28:25 CD2B.2   CD2B22822912  walk    54   InBed    walk               trk  1.00 OpenFloor  3   25    0.00  0.01  0.92  0.00  0.01  0.01
02:28:25 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   25    0.00  0.01  0.93  0.00  0.00  0.01
02:28:25 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   25    0.00  0.66  0.31  0.01  0.00  0.00
02:28:26 333B.0   -             sit     0    InBed    sit                room -    OpenFloor  3   25    0.00  0.01  0.93  0.00  0.00  0.01
02:28:26 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   26    0.00  0.66  0.31  0.01  0.00  0.00
02:28:26 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
02:28:26 CD2B.2   CD2B22822912  walk    67   InBed    walk               trk  1.00 OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
02:28:27 333B.0   -             sit     0    InBed    sit                room -    OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
02:28:27 333B.E   -             -       0    InBed    ExitRoom(rdr)      room -    OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
02:28:27 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   27    0.00  0.01  0.93  0.00  0.00  0.01
02:28:27 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   27    0.00  0.66  0.31  0.01  0.00  0.00
02:28:27 CD2B.2   CD2B22822912  walk    60   InBed    walk               trk  1.00 OpenFloor  3   27    0.00  0.01  0.93  0.00  0.00  0.01
02:28:28 333B.E   -             -       0    InBed    np=0  ★0           room -    OpenFloor  3   27    0.00  0.01  0.93  0.00  0.00  0.01
02:28:28 333B.88  -             88      -    InBed    no-target(88)      room -    OpenFloor  3   27    0.00  0.01  0.93  0.00  0.00  0.01
02:28:28 CD2B.2   CD2B22822912  walk    67   InBed    walk               trk  1.00 OpenFloor  3   28    0.00  0.01  0.93  0.00  0.00  0.01
02:28:28 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   28    0.00  0.01  0.93  0.00  0.00  0.01
02:28:28 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   28    0.00  0.66  0.31  0.01  0.00  0.00
02:28:29 333B.88  -             88      -    InBed    no-target(88)      room -    OpenFloor  3   28    0.00  0.01  0.93  0.00  0.00  0.01
02:28:29 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   29    0.00  0.01  0.93  0.00  0.00  0.01
02:28:29 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   29    0.00  0.66  0.31  0.01  0.00  0.00
02:28:29 CD2B.2   CD2B22822912  walk    88   InBed    walk               trk  1.00 OpenFloor  3   29    0.00  0.01  0.93  0.00  0.00  0.01
02:28:30 333B.88  -             88      -    InBed    no-target(88)      room -    OpenFloor  3   29    0.00  0.01  0.93  0.00  0.00  0.01
02:28:30 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   30    0.00  0.01  0.93  0.00  0.00  0.01
02:28:30 CD2B.2   CD2B22822912  walk    71   InBed    walk               trk  1.00 OpenFloor  3   30    0.00  0.01  0.93  0.00  0.00  0.01
02:28:30 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   30    0.00  0.66  0.31  0.01  0.00  0.00
02:28:31 CD2B.2   CD2B22822912  walk    67   InBed    walk               trk  1.00 OpenFloor  3   31    0.00  0.01  0.93  0.00  0.00  0.01
02:28:31 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   31    0.00  0.01  0.93  0.00  0.00  0.01
02:28:31 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   31    0.00  0.66  0.31  0.01  0.00  0.00
02:28:32 CD2B.2   CD2B22822912  walk    71   InBed    walk               trk  1.00 OpenFloor  3   32    0.00  0.01  0.93  0.00  0.00  0.01
02:28:32 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   32    0.00  0.01  0.93  0.00  0.00  0.01
02:28:32 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   32    0.00  0.66  0.31  0.01  0.00  0.00
02:28:33 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   33    0.00  0.01  0.93  0.00  0.00  0.01
02:28:33 CD2B.2   CD2B22822912  walk    84   InBed    walk               trk  1.00 OpenFloor  3   33    0.00  0.01  0.93  0.00  0.00  0.01
02:28:33 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   33    0.00  0.66  0.31  0.01  0.00  0.00
02:28:34 CD2B.2   CD2B22822912  walk    86   InBed    walk               trk  1.00 OpenFloor  3   34    0.00  0.01  0.93  0.00  0.00  0.01
02:28:34 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   34    0.00  0.66  0.31  0.01  0.00  0.00
02:28:34 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   34    0.00  0.01  0.93  0.00  0.00  0.01
02:28:35 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   35    0.00  0.66  0.31  0.01  0.00  0.00
02:28:35 CD2B.2   CD2B22822912  walk    81   InBed    walk               trk  1.00 OpenFloor  3   35    0.00  0.01  0.93  0.00  0.00  0.01
02:28:35 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   35    0.00  0.01  0.93  0.00  0.00  0.01
02:28:36 CD2B.2   CD2B22822912  walk    74   InBed    walk               trk  1.00 OpenFloor  3   36    0.00  0.01  0.93  0.00  0.00  0.01
02:28:36 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   36    0.00  0.01  0.93  0.00  0.00  0.01
02:28:36 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   36    0.00  0.66  0.31  0.01  0.00  0.00
02:28:37 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   37    0.00  0.66  0.31  0.01  0.00  0.00
02:28:37 CD2B.2   CD2B22822912  walk    71   InBed    walk               trk  1.00 OpenFloor  3   37    0.00  0.01  0.93  0.00  0.00  0.01
02:28:37 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   37    0.00  0.01  0.93  0.00  0.00  0.01
02:28:38 CD2B.2   CD2B22822912  walk    74   InBed    walk               trk  1.00 OpenFloor  3   38    0.00  0.01  0.93  0.00  0.00  0.01
02:28:38 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   38    0.00  0.66  0.31  0.01  0.00  0.00
02:28:38 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   38    0.00  0.01  0.93  0.00  0.00  0.01
02:28:39 CD2B.2   CD2B22822912  walk    72   InBed    walk               trk  1.00 OpenFloor  3   39    0.00  0.01  0.93  0.00  0.00  0.01
02:28:39 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   39    0.00  0.66  0.31  0.01  0.00  0.00
02:28:39 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   39    0.00  0.01  0.93  0.00  0.00  0.01
02:28:40 CD2B.2   CD2B22822912  walk    76   InBed    walk               trk  1.00 OpenFloor  3   40    0.00  0.01  0.93  0.00  0.00  0.01
02:28:40 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   40    0.00  0.01  0.93  0.00  0.00  0.01
02:28:40 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   40    0.00  0.66  0.31  0.01  0.00  0.00
02:28:41 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   41    0.00  0.01  0.93  0.00  0.00  0.01
02:28:41 CD2B.2   CD2B22822912  walk    70   InBed    walk               trk  1.00 OpenFloor  3   41    0.00  0.01  0.93  0.00  0.00  0.01
02:28:41 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   41    0.00  0.66  0.31  0.01  0.00  0.00
02:28:42 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   42    0.00  0.01  0.93  0.00  0.00  0.01
02:28:42 CD2B.2   CD2B22822912  walk    70   InBed    walk               trk  1.00 OpenFloor  3   42    0.00  0.01  0.93  0.00  0.00  0.01
02:28:42 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   42    0.00  0.66  0.31  0.01  0.00  0.00
02:28:43 CD2B.2   CD2B22822912  walk    80   InBed    walk               trk  1.00 OpenFloor  3   43    0.00  0.01  0.93  0.00  0.00  0.01
02:28:43 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   43    0.00  0.66  0.31  0.01  0.00  0.00
02:28:43 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   43    0.00  0.01  0.93  0.00  0.00  0.01
02:28:44 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   44    0.00  0.66  0.31  0.01  0.00  0.00
02:28:44 CD2B.2   CD2B22822912  walk    76   InBed    walk               trk  1.00 OpenFloor  3   44    0.00  0.01  0.93  0.00  0.00  0.01
02:28:44 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   44    0.00  0.01  0.93  0.00  0.00  0.01
02:28:45 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.93  0.00  0.00  0.01
02:28:45 CD2B.2   CD2B22822912  walk    61   InBed    walk               trk  1.00 OpenFloor  3   45    0.00  0.01  0.93  0.00  0.00  0.01
02:28:45 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   45    0.00  0.66  0.31  0.01  0.00  0.00
02:28:46 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
02:28:46 CD2B.2   CD2B22822912  walk    73   InBed    walk               trk  1.00 OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
02:28:46 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   46    0.00  0.66  0.31  0.01  0.00  0.00
02:28:47 333B.88  -             88      -    InBed    no-target(88)      room -    OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
02:28:47 CD2B.2   CD2B22822912  walk    72   InBed    walk               trk  1.00 OpenFloor  3   47    0.00  0.01  0.93  0.00  0.00  0.01
02:28:47 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   47    0.00  0.66  0.31  0.01  0.00  0.00
02:28:47 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   47    0.00  0.01  0.93  0.00  0.00  0.01
02:28:48 CD2B.2   CD2B22822912  walk    69   InBed    walk               trk  1.00 OpenFloor  3   48    0.00  0.01  0.93  0.00  0.00  0.01
02:28:48 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   48    0.00  0.66  0.31  0.01  0.00  0.00
02:28:48 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   48    0.00  0.01  0.93  0.00  0.00  0.01
02:28:49 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   49    0.00  0.66  0.31  0.01  0.00  0.00
02:28:49 CD2B.2   CD2B22822912  walk    75   InBed    walk               trk  1.00 OpenFloor  3   49    0.00  0.01  0.93  0.00  0.00  0.01
02:28:49 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   49    0.00  0.01  0.93  0.00  0.00  0.01
02:28:50 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   50    0.00  0.66  0.31  0.01  0.00  0.00
02:28:50 CD2B.2   CD2B22822912  walk    76   InBed    walk               trk  1.00 OpenFloor  3   50    0.00  0.01  0.93  0.00  0.00  0.01
02:28:50 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 OpenFloor  3   50    0.00  0.01  0.93  0.00  0.00  0.01
02:28:51 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=17 mv=0 turn=0 room -    OpenFloor  3   50    0.00  0.01  0.93  0.00  0.00  0.01
02:28:51 CD2B.2   CD2B22822912  sit     102  InBed    sit                trk  1.00 Bed        3   51    0.00  0.55  0.39  0.00  0.00  0.00
02:28:51 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   51    0.00  1.00  0.00  0.00  0.00  0.00
02:28:51 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   51    0.00  0.57  0.41  0.00  0.00  0.00
02:28:52 1641.E   -             -       0    InBed    InBed(pad)         room -    Bed        3   52    0.00  0.96  0.04  0.00  0.00  0.00
02:28:52 1641.E   -             -       0    InBed    InBed(pad)         room -    Bed        3   52    0.00  0.96  0.04  0.00  0.00  0.00
02:28:52 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   52    0.00  1.00  0.00  0.00  0.00  0.00
02:28:52 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   52    0.00  0.99  0.01  0.00  0.00  0.00
02:28:52 CD2B.2   CD2B22822912  sit     74   InBed    sit                trk  1.00 Bed        3   52    0.00  0.99  0.00  0.00  0.00  0.00
02:28:53 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=17 mv=0 turn=0 room -    Bed        3   52    0.00  0.99  0.01  0.00  0.00  0.00
02:28:53 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   53    0.00  0.99  0.01  0.00  0.00  0.00
02:28:53 CD2B.2   CD2B22822912  sit     72   InBed    sit                trk  1.00 Bed        3   53    0.00  1.00  0.00  0.00  0.00  0.00
02:28:53 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   53    0.00  1.00  0.00  0.00  0.00  0.00
02:28:54 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   54    0.00  1.00  0.00  0.00  0.00  0.00
02:28:54 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   54    0.00  0.99  0.01  0.00  0.00  0.00
02:28:54 CD2B.2   CD2B22822912  sit     84   InBed    sit                trk  1.00 Bed        3   54    0.00  1.00  0.00  0.00  0.00  0.00
02:28:55 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=17 mv=0 turn=0 room -    Bed        3   54    0.00  0.99  0.01  0.00  0.00  0.00
02:28:55 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   55    0.00  1.00  0.00  0.00  0.00  0.00
02:28:55 CD2B.2   CD2B22822912  sit     72   InBed    sit                trk  1.00 Bed        3   55    0.00  1.00  0.00  0.00  0.00  0.00
02:28:55 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   55    0.00  0.99  0.01  0.00  0.00  0.00
02:28:56 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   56    0.00  1.00  0.00  0.00  0.00  0.00
02:28:56 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   56    0.00  0.99  0.01  0.00  0.00  0.00
02:28:56 CD2B.2   CD2B22822912  sit     75   InBed    sit                trk  1.00 Bed        3   56    0.00  1.00  0.00  0.00  0.00  0.00
02:28:57 1641.0   -             pad     -    InBed    pad InBed HR=76 RR=17 mv=0 turn=0 room -    Bed        3   56    0.00  0.99  0.01  0.00  0.00  0.00
02:28:57 CD2B.2   CD2B22822912  sit     79   InBed    sit                trk  1.00 Bed        3   57    0.00  1.00  0.00  0.00  0.00  0.00
02:28:57 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   57    0.00  0.99  0.01  0.00  0.00  0.00
02:28:57 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   57    0.00  1.00  0.00  0.00  0.00  0.00
02:28:58 CD2B.2   CD2B22822912  sit     74   InBed    sit                trk  1.00 Bed        3   58    0.00  1.00  0.00  0.00  0.00  0.00
02:28:58 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   58    0.00  0.99  0.01  0.00  0.00  0.00
02:28:58 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   58    0.00  1.00  0.00  0.00  0.00  0.00
02:28:59 1641.0   -             pad     -    InBed    pad InBed HR=74 RR=17 mv=0 turn=0 room -    Bed        3   58    0.00  0.99  0.01  0.00  0.00  0.00
02:28:59 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   59    0.00  1.00  0.00  0.00  0.00  0.00
02:28:59 CD2B.2   CD2B22822912  sit     70   InBed    sit                trk  1.00 Bed        3   59    0.00  1.00  0.00  0.00  0.00  0.00
02:28:59 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   59    0.00  0.99  0.01  0.00  0.00  0.00
02:29:00 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   60    0.00  0.99  0.01  0.00  0.00  0.00
02:29:00 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   60    0.00  1.00  0.00  0.00  0.00  0.00
02:29:00 CD2B.2   CD2B22822912  sit     75   InBed    sit                trk  1.00 Bed        3   60    0.00  1.00  0.00  0.00  0.00  0.00
02:29:01 1641.0   -             pad     -    InBed    pad InBed HR=73 RR=15 mv=0 turn=0 room -    Bed        3   60    0.00  0.99  0.01  0.00  0.00  0.00
02:29:01 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   61    0.00  0.99  0.01  0.00  0.00  0.00
02:29:01 CD2B.2   CD2B22822912  sit     70   InBed    sit                trk  1.00 Bed        3   61    0.00  1.00  0.00  0.00  0.00  0.00
02:29:01 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   61    0.00  1.00  0.00  0.00  0.00  0.00
02:29:02 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   62    0.00  0.99  0.01  0.00  0.00  0.00
02:29:02 CD2B.2   CD2B22822912  sit     63   InBed    sit                trk  1.00 Bed        3   62    0.00  1.00  0.00  0.00  0.00  0.00
02:29:02 CD2B.1   CD2B12800032  walk    0    InBed    walk               trk  1.00 Bed        3   62    0.00  1.00  0.00  0.00  0.00  0.00
02:29:03 1641.0   -             pad     -    InBed    pad InBed HR=73 RR=None mv=0 turn=0 room -    Bed        3   62    0.00  0.99  0.01  0.00  0.00  0.00
02:29:03 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   63    0.00  0.99  0.01  0.00  0.00  0.00
02:29:03 CD2B.2   CD2B22822912  sit     71   InBed    sit                trk  1.00 Bed        3   63    0.00  1.00  0.00  0.00  0.00  0.00
02:29:03 CD2B.1   CD2B12800032  walk    0    InBed    walk               trk  1.00 Bed        3   63    0.00  1.00  0.00  0.00  0.00  0.00
02:29:04 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   64    0.00  0.99  0.01  0.00  0.00  0.00
02:29:04 CD2B.2   CD2B22822912  sit     74   InBed    sit                trk  1.00 Bed        3   64    0.00  1.00  0.00  0.00  0.00  0.00
02:29:04 CD2B.1   CD2B12800032  walk    0    InBed    walk               trk  1.00 Bed        3   64    0.00  1.00  0.00  0.00  0.00  0.00
02:29:05 1641.0   -             pad     -    InBed    pad InBed HR=67 RR=13 mv=1 turn=0 room -    Bed        3   64    0.00  0.99  0.01  0.00  0.00  0.00
02:29:05 CD2B.2   CD2B22822912  sit     69   InBed    sit                trk  1.00 Bed        3   65    0.00  1.00  0.00  0.00  0.00  0.00
02:29:05 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   65    0.00  0.99  0.01  0.00  0.00  0.00
02:29:05 CD2B.1   CD2B12800032  walk    0    InBed    walk               trk  1.00 Bed        3   65    0.00  1.00  0.00  0.00  0.00  0.00
02:29:06 CD2B.1   CD2B12800032  walk    0    InBed    walk               trk  1.00 Bed        3   66    0.00  1.00  0.00  0.00  0.00  0.00
02:29:06 CD2B.2   CD2B22822912  sit     68   InBed    sit                trk  1.00 Bed        3   66    0.00  1.00  0.00  0.00  0.00  0.00
02:29:06 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   66    0.00  0.99  0.01  0.00  0.00  0.00
02:29:07 1641.0   -             pad     -    InBed    pad InBed HR=67 RR=13 mv=0 turn=0 room -    Bed        3   66    0.00  0.99  0.01  0.00  0.00  0.00
02:29:07 CD2B.2   CD2B22822912  sit     92   InBed    sit                trk  1.00 Bed        3   67    0.00  0.99  0.01  0.00  0.00  0.00
02:29:07 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   67    0.00  1.00  0.00  0.00  0.00  0.00
02:29:07 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   67    0.00  0.99  0.01  0.00  0.00  0.00
02:29:08 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   68    0.00  1.00  0.00  0.00  0.00  0.00
02:29:08 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   68    0.00  0.99  0.01  0.00  0.00  0.00
02:29:08 CD2B.2   CD2B22822912  sit     72   InBed    sit                trk  1.00 Bed        3   68    0.00  0.99  0.01  0.00  0.00  0.00
02:29:09 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=17 mv=0 turn=0 room -    Bed        3   68    0.00  0.99  0.01  0.00  0.00  0.00
02:29:09 333B.E   -             -       0    InBed    np=1               room -    Bed        3   68    0.00  0.99  0.01  0.00  0.00  0.00
02:29:09 333B.E   -             -       0    InBed    EnterRoom(rdr)     room -    Bed        3   68    0.00  0.99  0.01  0.00  0.00  0.00
02:29:09 333B.0   -             stand   100  InBed    stand              room -    Bed        3   68    0.00  0.99  0.01  0.00  0.00  0.00
02:29:09 CD2B.2   CD2B22822912  sit     75   InBed    sit                trk  1.00 Bed        3   69    0.00  1.00  0.00  0.00  0.00  0.00
02:29:09 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   69    0.00  1.00  0.00  0.00  0.00  0.00
02:29:09 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   69    0.00  0.99  0.01  0.00  0.00  0.00
02:29:10 333B.0   -             stand   100  InBed    stand              room -    Bed        3   69    0.00  0.99  0.01  0.00  0.00  0.00
02:29:10 CD2B.2   CD2B22822912  sit     91   InBed    sit                trk  1.00 Bed        3   70    0.00  0.99  0.01  0.00  0.00  0.00
02:29:10 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   70    0.00  0.99  0.01  0.00  0.00  0.00
02:29:10 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   70    0.00  1.00  0.00  0.00  0.00  0.00
02:29:11 1641.0   -             pad     -    InBed    pad InBed HR=65 RR=21 mv=0 turn=0 room -    Bed        3   70    0.00  0.99  0.01  0.00  0.00  0.00
02:29:11 333B.0   -             stand   119  InBed    stand              room -    Bed        3   70    0.00  0.99  0.01  0.00  0.00  0.00
02:29:11 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   71    0.00  1.00  0.00  0.00  0.00  0.00
02:29:11 CD2B.2   CD2B22822912  sit     83   InBed    sit                trk  1.00 Bed        3   71    0.00  0.99  0.01  0.00  0.00  0.00
02:29:11 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   71    0.00  0.99  0.01  0.00  0.00  0.00
02:29:12 333B.0   -             walk    102  InBed    walk               room -    Bed        3   71    0.00  0.99  0.01  0.00  0.00  0.00
02:29:12 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   72    0.00  0.99  0.01  0.00  0.00  0.00
02:29:12 CD2B.2   CD2B22822912  sit     82   InBed    sit                trk  1.00 Bed        3   72    0.00  0.99  0.01  0.00  0.00  0.00
02:29:12 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   72    0.00  1.00  0.00  0.00  0.00  0.00
02:29:13 1641.0   -             pad     -    InBed    pad InBed HR=69 RR=21 mv=0 turn=0 room -    Bed        3   72    0.00  0.99  0.01  0.00  0.00  0.00
02:29:13 333B.0   -             walk    110  InBed    walk               room -    Bed        3   72    0.00  0.99  0.01  0.00  0.00  0.00
02:29:13 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   73    0.00  1.00  0.00  0.00  0.00  0.00
02:29:13 CD2B.2   CD2B22822912  sit     71   InBed    sit                trk  1.00 Bed        3   73    0.00  1.00  0.00  0.00  0.00  0.00
02:29:13 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   73    0.00  0.99  0.01  0.00  0.00  0.00
02:29:14 333B.0   -             walk    111  InBed    walk               room -    Bed        3   73    0.00  1.00  0.00  0.00  0.00  0.00
02:29:14 CD2B.2   CD2B22822912  sit     78   InBed    sit                trk  1.00 Bed        3   74    0.00  1.00  0.00  0.00  0.00  0.00
02:29:14 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   74    0.00  0.99  0.01  0.00  0.00  0.00
02:29:14 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   74    0.00  1.00  0.00  0.00  0.00  0.00
02:29:15 1641.0   -             pad     -    InBed    pad LeftBed HR=None RR=None mv=0 turn=0 room -    Bed        3   74    0.00  0.99  0.01  0.00  0.00  0.00
02:29:15 333B.0   -             walk    112  InBed    walk               room -    Bed        3   74    0.00  0.99  0.01  0.00  0.00  0.00
02:29:15 CD2B.1   CD2B12800032  stand   0    InBed    stand              trk  1.00 Bed        3   75    0.00  0.90  0.08  0.01  0.00  0.00
02:29:15 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        3   75    0.00  0.53  0.39  0.07  0.00  0.01
02:29:15 CD2B.2   CD2B22822912  sit     66   InBed    sit                trk  1.00 Bed        3   75    0.00  0.74  0.07  0.09  0.00  0.01
02:29:16 333B.0   -             walk    93   InBed    walk               room -    Bed        3   75    0.00  0.74  0.07  0.09  0.00  0.01
02:29:16 1641.E   -             -       0    LeftBed  LeftBed(pad)       room -    OpenFloor  3   76    0.00  0.00  1.00  0.00  0.00  0.00
02:29:16 1641.E   -             -       0    LeftBed  LeftBed(pad)       room -    OpenFloor  3   76    0.00  0.00  1.00  0.00  0.00  0.00
02:29:16 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   76    0.00  0.00  1.00  0.00  0.00  0.00
02:29:16 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   76    0.00  0.00  1.00  0.00  0.00  0.00
02:29:16 CD2B.2   CD2B22822912  sit     72   LeftBed  sit                trk  1.00 OpenFloor  3   76    0.00  0.00  1.00  0.00  0.00  0.00
02:29:17 333B.0   -             walk    100  LeftBed  walk               room -    OpenFloor  3   76    0.00  0.00  1.00  0.00  0.00  0.00
02:29:17 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   77    0.00  0.00  1.00  0.00  0.00  0.00
02:29:17 CD2B.2   CD2B22822912  sit     72   LeftBed  sit                trk  1.00 OpenFloor  3   77    0.00  0.00  1.00  0.00  0.00  0.00
02:29:17 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   77    0.00  0.00  1.00  0.00  0.00  0.00
02:29:18 333B.0   -             walk    82   LeftBed  walk               room -    OpenFloor  3   77    0.00  0.00  1.00  0.00  0.00  0.00
02:29:18 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   78    0.00  0.00  1.00  0.00  0.00  0.00
02:29:18 CD2B.2   CD2B22822912  sit     79   LeftBed  sit                trk  1.00 OpenFloor  3   78    0.00  0.00  1.00  0.00  0.00  0.00
02:29:18 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   78    0.00  0.00  1.00  0.00  0.00  0.00
02:29:19 333B.0   -             walk    103  LeftBed  walk               room -    OpenFloor  3   78    0.00  0.00  1.00  0.00  0.00  0.00
02:29:19 CD2B.2   CD2B22822912  sit     58   LeftBed  sit                trk  1.00 OpenFloor  3   79    0.00  0.00  1.00  0.00  0.00  0.00
02:29:19 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   79    0.00  0.00  1.00  0.00  0.00  0.00
02:29:19 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   79    0.00  0.00  1.00  0.00  0.00  0.00
02:29:20 333B.0   -             walk    126  LeftBed  walk               room -    OpenFloor  3   79    0.00  0.00  1.00  0.00  0.00  0.00
02:29:20 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   80    0.00  0.00  1.00  0.00  0.00  0.00
02:29:20 CD2B.2   CD2B22822912  sit     77   LeftBed  sit                trk  1.00 OpenFloor  3   80    0.00  0.00  1.00  0.00  0.00  0.00
02:29:20 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   80    0.00  0.00  1.00  0.00  0.00  0.00
02:29:21 333B.0   -             walk    122  LeftBed  walk               room -    OpenFloor  3   80    0.00  0.00  1.00  0.00  0.00  0.00
02:29:21 CD2B.2   CD2B22822912  sit     101  LeftBed  sit                trk  1.00 OpenFloor  3   81    0.00  0.00  0.99  0.00  0.00  0.00
02:29:21 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   81    0.00  0.00  1.00  0.00  0.00  0.00
02:29:21 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   81    0.00  0.00  1.00  0.00  0.00  0.00
02:29:22 333B.0   -             walk    136  LeftBed  walk               room -    OpenFloor  3   81    0.00  0.00  0.99  0.00  0.00  0.00
02:29:22 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   82    0.00  0.00  1.00  0.00  0.00  0.00
02:29:22 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   82    0.00  0.00  1.00  0.00  0.00  0.00
02:29:22 CD2B.2   CD2B22822912  sit     87   LeftBed  sit                trk  1.00 OpenFloor  3   82    0.00  0.00  0.99  0.00  0.00  0.00
02:29:23 333B.0   -             walk    0    LeftBed  walk               room -    OpenFloor  3   82    0.00  0.00  0.99  0.00  0.00  0.00
02:29:23 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   83    0.00  0.00  1.00  0.00  0.00  0.00
02:29:23 CD2B.2   CD2B22822912  sit     84   LeftBed  sit                trk  1.00 OpenFloor  3   83    0.00  0.00  1.00  0.00  0.00  0.00
02:29:23 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   83    0.00  0.00  1.00  0.00  0.00  0.00
02:29:24 333B.0   -             walk    0    LeftBed  walk               room -    OpenFloor  3   83    0.00  0.00  1.00  0.00  0.00  0.00
02:29:24 CD2B.2   CD2B22822912  sit     86   LeftBed  sit                trk  1.00 OpenFloor  3   84    0.00  0.00  1.00  0.00  0.00  0.00
02:29:24 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   84    0.00  0.00  1.00  0.00  0.00  0.00
02:29:24 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   84    0.00  0.00  1.00  0.00  0.00  0.00
02:29:25 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   84    0.00  0.00  1.00  0.00  0.00  0.00
02:29:25 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   85    0.00  0.00  1.00  0.00  0.00  0.00
02:29:25 CD2B.2   CD2B22822912  sit     83   LeftBed  sit                trk  1.00 OpenFloor  3   85    0.00  0.00  1.00  0.00  0.00  0.00
02:29:25 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   85    0.00  0.00  1.00  0.00  0.00  0.00
02:29:26 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   85    0.00  0.00  1.00  0.00  0.00  0.00
02:29:26 CD2B.2   CD2B22822912  sit     100  LeftBed  sit                trk  1.00 OpenFloor  3   86    0.00  0.00  1.00  0.00  0.00  0.00
02:29:26 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   86    0.00  0.00  1.00  0.00  0.00  0.00
02:29:26 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   86    0.00  0.00  1.00  0.00  0.00  0.00
02:29:27 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   86    0.00  0.00  1.00  0.00  0.00  0.00
02:29:27 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   87    0.00  0.00  1.00  0.00  0.00  0.00
02:29:27 CD2B.2   CD2B22822912  sit     89   LeftBed  sit                trk  1.00 OpenFloor  3   87    0.00  0.00  1.00  0.00  0.00  0.00
02:29:27 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   87    0.00  0.00  1.00  0.00  0.00  0.00
02:29:28 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   87    0.00  0.00  1.00  0.00  0.00  0.00
02:29:28 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   88    0.00  0.00  1.00  0.00  0.00  0.00
02:29:28 CD2B.2   CD2B22822912  sit     69   LeftBed  sit                trk  1.00 OpenFloor  3   88    0.00  0.00  1.00  0.00  0.00  0.00
02:29:28 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   88    0.00  0.00  1.00  0.00  0.00  0.00
02:29:29 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   88    0.00  0.00  1.00  0.00  0.00  0.00
02:29:29 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   89    0.00  0.00  1.00  0.00  0.00  0.00
02:29:29 CD2B.2   CD2B22822912  sit     55   LeftBed  sit                trk  1.00 OpenFloor  3   89    0.00  0.00  1.00  0.00  0.00  0.00
02:29:29 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   89    0.00  0.00  1.00  0.00  0.00  0.00
02:29:30 333B.0   -             stand   88   LeftBed  stand              room -    OpenFloor  3   89    0.00  0.00  1.00  0.00  0.00  0.00
02:29:30 CD2B.2   CD2B22822912  sit     50   LeftBed  sit                trk  1.00 OpenFloor  3   90    0.00  0.00  1.00  0.00  0.00  0.00
02:29:30 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   90    0.00  0.00  1.00  0.00  0.00  0.00
02:29:30 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   90    0.00  0.00  1.00  0.00  0.00  0.00
02:29:31 333B.0   -             stand   106  LeftBed  stand              room -    OpenFloor  3   90    0.00  0.00  1.00  0.00  0.00  0.00
02:29:31 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   91    0.00  0.00  1.00  0.00  0.00  0.00
02:29:31 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   91    0.00  0.00  1.00  0.00  0.00  0.00
02:29:31 CD2B.2   CD2B22822912  lying   50   LeftBed  lying              trk  1.00 OpenFloor  3   91    0.00  0.00  1.00  0.00  0.00  0.00
02:29:32 333B.0   -             stand   101  LeftBed  stand              room -    OpenFloor  3   91    0.00  0.00  1.00  0.00  0.00  0.00
02:29:32 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   92    0.00  0.00  1.00  0.00  0.00  0.00
02:29:32 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   92    0.00  0.00  1.00  0.00  0.00  0.00
02:29:32 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   92    0.00  0.00  1.00  0.00  0.00  0.00
02:29:33 333B.0   -             stand   108  LeftBed  stand              room -    OpenFloor  3   92    0.00  0.00  1.00  0.00  0.00  0.00
02:29:33 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   93    0.00  0.00  1.00  0.00  0.00  0.00
02:29:33 CD2B.2   CD2B22822912  lying   56   LeftBed  lying              trk  1.00 OpenFloor  3   93    0.00  0.00  1.00  0.00  0.00  0.00
02:29:33 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   93    0.00  0.00  1.00  0.00  0.00  0.00
02:29:34 333B.0   -             stand   98   LeftBed  stand              room -    OpenFloor  3   93    0.00  0.00  1.00  0.00  0.00  0.00
02:29:34 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   94    0.00  0.00  1.00  0.00  0.00  0.00
02:29:34 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   94    0.00  0.00  1.00  0.00  0.00  0.00
02:29:34 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   94    0.00  0.00  1.00  0.00  0.00  0.00
02:29:35 333B.0   -             stand   67   LeftBed  stand              room -    OpenFloor  3   94    0.00  0.00  1.00  0.00  0.00  0.00
02:29:35 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   95    0.00  0.00  1.00  0.00  0.00  0.00
02:29:35 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   95    0.00  0.00  1.00  0.00  0.00  0.00
02:29:35 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   95    0.00  0.00  1.00  0.00  0.00  0.00
02:29:36 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   95    0.00  0.00  1.00  0.00  0.00  0.00
02:29:36 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   96    0.00  0.00  1.00  0.00  0.00  0.00
02:29:36 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   96    0.00  0.00  1.00  0.00  0.00  0.00
02:29:36 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   96    0.00  0.00  1.00  0.00  0.00  0.00
02:29:37 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   96    0.00  0.00  1.00  0.00  0.00  0.00
02:29:37 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   97    0.00  0.00  1.00  0.00  0.00  0.00
02:29:37 CD2B.2   CD2B22822912  lying   49   LeftBed  lying              trk  1.00 OpenFloor  3   97    0.00  0.00  1.00  0.00  0.00  0.00
02:29:37 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   97    0.00  0.00  1.00  0.00  0.00  0.00
02:29:38 333B.0   -             stand   43   LeftBed  stand              room -    OpenFloor  3   97    0.00  0.00  1.00  0.00  0.00  0.00
02:29:38 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   98    0.00  0.00  1.00  0.00  0.00  0.00
02:29:38 CD2B.2   CD2B22822912  lying   42   LeftBed  lying              trk  1.00 OpenFloor  3   98    0.00  0.00  1.00  0.00  0.00  0.00
02:29:38 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   98    0.00  0.00  1.00  0.00  0.00  0.00
02:29:39 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   98    0.00  0.00  1.00  0.00  0.00  0.00
02:29:39 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   99    0.00  0.00  1.00  0.00  0.00  0.00
02:29:39 CD2B.2   CD2B22822912  lying   48   LeftBed  lying              trk  1.00 OpenFloor  3   99    0.00  0.00  1.00  0.00  0.00  0.00
02:29:39 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   99    0.00  0.00  1.00  0.00  0.00  0.00
02:29:40 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   99    0.00  0.00  1.00  0.00  0.00  0.00
02:29:40 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   100   0.00  0.00  1.00  0.00  0.00  0.00
02:29:40 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   100   0.00  0.00  1.00  0.00  0.00  0.00
02:29:40 CD2B.2   CD2B22822912  lying   48   LeftBed  lying              trk  1.00 OpenFloor  3   100   0.00  0.00  1.00  0.00  0.00  0.00
02:29:41 333B.0   -             stand   95   LeftBed  stand              room -    OpenFloor  3   100   0.00  0.00  1.00  0.00  0.00  0.00
02:29:41 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   101   0.00  0.00  1.00  0.00  0.00  0.00
02:29:41 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   101   0.00  0.00  1.00  0.00  0.00  0.00
02:29:41 CD2B.2   CD2B22822912  lying   49   LeftBed  lying              trk  1.00 OpenFloor  3   101   0.00  0.00  1.00  0.00  0.00  0.00
02:29:42 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   101   0.00  0.00  1.00  0.00  0.00  0.00
02:29:42 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   102   0.00  0.00  1.00  0.00  0.00  0.00
02:29:42 CD2B.2   CD2B22822912  lying   55   LeftBed  lying              trk  1.00 OpenFloor  3   102   0.00  0.00  1.00  0.00  0.00  0.00
02:29:42 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   102   0.00  0.00  1.00  0.00  0.00  0.00
02:29:42 333B.0   -             stand   154  LeftBed  stand              room -    OpenFloor  3   102   0.00  0.00  1.00  0.00  0.00  0.00
02:29:43 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   103   0.00  0.00  1.00  0.00  0.00  0.00
02:29:43 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   103   0.00  0.00  1.00  0.00  0.00  0.00
02:29:43 CD2B.2   CD2B22822912  lying   59   LeftBed  lying              trk  1.00 OpenFloor  3   103   0.00  0.00  1.00  0.00  0.00  0.00
02:29:43 333B.0   -             stand   107  LeftBed  stand              room -    OpenFloor  3   103   0.00  0.00  1.00  0.00  0.00  0.00
02:29:44 CD2B.2   CD2B22822912  lying   62   LeftBed  lying              trk  1.00 OpenFloor  3   104   0.00  0.00  1.00  0.00  0.00  0.00
02:29:44 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   104   0.00  0.00  1.00  0.00  0.00  0.00
02:29:44 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   104   0.00  0.00  1.00  0.00  0.00  0.00
02:29:44 333B.0   -             stand   131  LeftBed  stand              room -    OpenFloor  3   104   0.00  0.00  1.00  0.00  0.00  0.00
02:29:45 CD2B.2   CD2B22822912  lying   52   LeftBed  lying              trk  1.00 OpenFloor  3   105   0.00  0.00  1.00  0.00  0.00  0.00
02:29:45 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   105   0.00  0.00  1.00  0.00  0.00  0.00
02:29:45 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   105   0.00  0.00  1.00  0.00  0.00  0.00
02:29:45 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   105   0.00  0.00  1.00  0.00  0.00  0.00
02:29:46 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   106   0.00  0.00  1.00  0.00  0.00  0.00
02:29:46 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   106   0.00  0.00  1.00  0.00  0.00  0.00
02:29:46 CD2B.2   CD2B22822912  lying   56   LeftBed  lying              trk  1.00 OpenFloor  3   106   0.00  0.00  1.00  0.00  0.00  0.00
02:29:46 333B.0   -             stand   161  LeftBed  stand              room -    OpenFloor  3   106   0.00  0.00  1.00  0.00  0.00  0.00
02:29:47 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   107   0.00  0.00  1.00  0.00  0.00  0.00
02:29:47 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   107   0.00  0.00  1.00  0.00  0.00  0.00
02:29:47 CD2B.2   CD2B22822912  lying   54   LeftBed  lying              trk  1.00 OpenFloor  3   107   0.00  0.00  1.00  0.00  0.00  0.00
02:29:47 333B.0   -             stand   92   LeftBed  stand              room -    OpenFloor  3   107   0.00  0.00  1.00  0.00  0.00  0.00
02:29:48 CD2B.2   CD2B22822912  lying   54   LeftBed  lying              trk  1.00 OpenFloor  3   108   0.00  0.00  1.00  0.00  0.00  0.00
02:29:48 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   108   0.00  0.00  1.00  0.00  0.00  0.00
02:29:48 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   108   0.00  0.00  1.00  0.00  0.00  0.00
02:29:48 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   108   0.00  0.00  1.00  0.00  0.00  0.00
02:29:49 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   109   0.00  0.00  1.00  0.00  0.00  0.00
02:29:49 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   109   0.00  0.00  1.00  0.00  0.00  0.00
02:29:49 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   109   0.00  0.00  1.00  0.00  0.00  0.00
02:29:49 333B.0   -             stand   29   LeftBed  stand              room -    OpenFloor  3   109   0.00  0.00  1.00  0.00  0.00  0.00
02:29:50 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   110   0.00  0.00  1.00  0.00  0.00  0.00
02:29:50 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   110   0.00  0.00  1.00  0.00  0.00  0.00
02:29:50 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   110   0.00  0.00  1.00  0.00  0.00  0.00
02:29:50 333B.0   -             stand   127  LeftBed  stand              room -    OpenFloor  3   110   0.00  0.00  1.00  0.00  0.00  0.00
02:29:51 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   111   0.00  0.00  1.00  0.00  0.00  0.00
02:29:51 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   111   0.00  0.00  1.00  0.00  0.00  0.00
02:29:51 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   111   0.00  0.00  1.00  0.00  0.00  0.00
02:29:51 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   111   0.00  0.00  1.00  0.00  0.00  0.00
02:29:52 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   112   0.00  0.00  1.00  0.00  0.00  0.00
02:29:52 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   112   0.00  0.00  1.00  0.00  0.00  0.00
02:29:52 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   112   0.00  0.00  1.00  0.00  0.00  0.00
02:29:52 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   112   0.00  0.00  1.00  0.00  0.00  0.00
02:29:53 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   113   0.00  0.00  1.00  0.00  0.00  0.00
02:29:53 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   113   0.00  0.00  1.00  0.00  0.00  0.00
02:29:53 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   113   0.00  0.00  1.00  0.00  0.00  0.00
02:29:53 333B.0   -             stand   143  LeftBed  stand              room -    OpenFloor  3   113   0.00  0.00  1.00  0.00  0.00  0.00
02:29:54 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   114   0.00  0.00  1.00  0.00  0.00  0.00
02:29:54 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   114   0.00  0.00  1.00  0.00  0.00  0.00
02:29:54 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   114   0.00  0.00  1.00  0.00  0.00  0.00
02:29:54 333B.0   -             walk    114  LeftBed  walk               room -    OpenFloor  3   114   0.00  0.00  1.00  0.00  0.00  0.00
02:29:55 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   115   0.00  0.00  1.00  0.00  0.00  0.00
02:29:55 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   115   0.00  0.00  1.00  0.00  0.00  0.00
02:29:55 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   115   0.00  0.00  1.00  0.00  0.00  0.00
02:29:55 333B.0   -             walk    102  LeftBed  walk               room -    OpenFloor  3   115   0.00  0.00  1.00  0.00  0.00  0.00
02:29:56 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   116   0.00  0.00  1.00  0.00  0.00  0.00
02:29:56 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   116   0.00  0.00  1.00  0.00  0.00  0.00
02:29:56 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   116   0.00  0.00  1.00  0.00  0.00  0.00
02:29:56 333B.0   -             walk    143  LeftBed  walk               room -    OpenFloor  3   116   0.00  0.00  1.00  0.00  0.00  0.00
02:29:57 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   117   0.00  0.00  1.00  0.00  0.00  0.00
02:29:57 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   117   0.00  0.00  1.00  0.00  0.00  0.00
02:29:57 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   117   0.00  0.00  1.00  0.00  0.00  0.00
02:29:57 333B.0   -             walk    127  LeftBed  walk               room -    OpenFloor  3   117   0.00  0.00  1.00  0.00  0.00  0.00
02:29:58 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   118   0.00  0.00  1.00  0.00  0.00  0.00
02:29:58 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   118   0.00  0.00  1.00  0.00  0.00  0.00
02:29:58 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   118   0.00  0.00  1.00  0.00  0.00  0.00
02:29:58 333B.0   -             walk    122  LeftBed  walk               room -    OpenFloor  3   118   0.00  0.00  1.00  0.00  0.00  0.00
02:29:59 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   119   0.00  0.00  1.00  0.00  0.00  0.00
02:29:59 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   119   0.00  0.00  1.00  0.00  0.00  0.00
02:29:59 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   119   0.00  0.00  1.00  0.00  0.00  0.00
02:29:59 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   119   0.00  0.00  1.00  0.00  0.00  0.00
02:30:00 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   120   0.00  0.00  1.00  0.00  0.00  0.00
02:30:00 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   120   0.00  0.00  1.00  0.00  0.00  0.00
02:30:00 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   120   0.00  0.00  1.00  0.00  0.00  0.00
02:30:00 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   120   0.00  0.00  1.00  0.00  0.00  0.00
02:30:01 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   121   0.00  0.00  1.00  0.00  0.00  0.00
02:30:01 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   121   0.00  0.00  1.00  0.00  0.00  0.00
02:30:01 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   121   0.00  0.00  1.00  0.00  0.00  0.00
02:30:01 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   121   0.00  0.00  1.00  0.00  0.00  0.00
02:30:02 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   122   0.00  0.00  1.00  0.00  0.00  0.00
02:30:02 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   122   0.00  0.00  1.00  0.00  0.00  0.00
02:30:02 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   122   0.00  0.00  1.00  0.00  0.00  0.00
02:30:02 333B.0   -             stand   109  LeftBed  stand              room -    OpenFloor  3   122   0.00  0.00  1.00  0.00  0.00  0.00
02:30:03 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   122   0.00  0.00  1.00  0.00  0.00  0.00
02:30:03 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   122   0.00  0.00  1.00  0.00  0.00  0.00
02:30:03 CD2B.2   CD2B22822912  stand   56   LeftBed  stand              trk  1.00 OpenFloor  3   122   0.00  0.00  1.00  0.00  0.00  0.00
02:30:03 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   122   0.00  0.00  1.00  0.00  0.00  0.00
02:30:04 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   123   0.00  0.00  1.00  0.00  0.00  0.00
02:30:04 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   123   0.00  0.00  1.00  0.00  0.00  0.00
02:30:04 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   123   0.00  0.00  1.00  0.00  0.00  0.00
02:30:04 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   123   0.00  0.00  1.00  0.00  0.00  0.00
02:30:05 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   124   0.00  0.00  1.00  0.00  0.00  0.00
02:30:05 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   124   0.00  0.00  1.00  0.00  0.00  0.00
02:30:05 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   124   0.00  0.00  1.00  0.00  0.00  0.00
02:30:05 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   124   0.00  0.00  1.00  0.00  0.00  0.00
02:30:06 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   125   0.00  0.00  1.00  0.00  0.00  0.00
02:30:06 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   125   0.00  0.00  1.00  0.00  0.00  0.00
02:30:06 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   125   0.00  0.00  1.00  0.00  0.00  0.00
02:30:06 333B.0   -             stand   128  LeftBed  stand              room -    OpenFloor  3   125   0.00  0.00  1.00  0.00  0.00  0.00
02:30:07 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   126   0.00  0.00  1.00  0.00  0.00  0.00
02:30:07 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   126   0.00  0.00  1.00  0.00  0.00  0.00
02:30:07 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   126   0.00  0.00  1.00  0.00  0.00  0.00
02:30:07 333B.0   -             stand   125  LeftBed  stand              room -    OpenFloor  3   126   0.00  0.00  1.00  0.00  0.00  0.00
02:30:08 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   128   0.00  0.00  1.00  0.00  0.00  0.00
02:30:08 CD2B.2   CD2B22822912  stand   66   LeftBed  stand              trk  1.00 OpenFloor  3   128   0.00  0.00  1.00  0.00  0.00  0.00
02:30:08 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   128   0.00  0.00  1.00  0.00  0.00  0.00
02:30:08 333B.0   -             stand   102  LeftBed  stand              room -    OpenFloor  3   128   0.00  0.00  1.00  0.00  0.00  0.00
02:30:09 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   129   0.00  0.00  1.00  0.00  0.00  0.00
02:30:09 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   129   0.00  0.00  1.00  0.00  0.00  0.00
02:30:09 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   129   0.00  0.00  1.00  0.00  0.00  0.00
02:30:09 333B.0   -             stand   92   LeftBed  stand              room -    OpenFloor  3   129   0.00  0.00  1.00  0.00  0.00  0.00
02:30:10 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   130   0.00  0.00  1.00  0.00  0.00  0.00
02:30:10 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   130   0.00  0.00  1.00  0.00  0.00  0.00
02:30:10 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   130   0.00  0.00  1.00  0.00  0.00  0.00
02:30:10 333B.0   -             walk    93   LeftBed  walk               room -    OpenFloor  3   130   0.00  0.00  1.00  0.00  0.00  0.00
02:30:10 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   130   0.00  0.00  1.00  0.00  0.00  0.00
02:30:10 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   130   0.00  0.00  1.00  0.00  0.00  0.00
02:30:10 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   130   0.00  0.00  1.00  0.00  0.00  0.00
02:30:11 333B.0   -             walk    93   LeftBed  walk               room -    OpenFloor  3   130   0.00  0.00  1.00  0.00  0.00  0.00
02:30:11 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   131   0.00  0.00  1.00  0.00  0.00  0.00
02:30:11 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   131   0.00  0.00  1.00  0.00  0.00  0.00
02:30:11 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   131   0.00  0.00  1.00  0.00  0.00  0.00
02:30:12 333B.0   -             walk    135  LeftBed  walk               room -    OpenFloor  3   131   0.00  0.00  1.00  0.00  0.00  0.00
02:30:12 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   132   0.00  0.00  1.00  0.00  0.00  0.00
02:30:12 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   132   0.00  0.00  1.00  0.00  0.00  0.00
02:30:12 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   132   0.00  0.00  1.00  0.00  0.00  0.00
02:30:13 333B.0   -             walk    97   LeftBed  walk               room -    OpenFloor  3   132   0.00  0.00  1.00  0.00  0.00  0.00
02:30:13 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   133   0.00  0.00  1.00  0.00  0.00  0.00
02:30:13 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   133   0.00  0.00  1.00  0.00  0.00  0.00
02:30:13 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   133   0.00  0.00  1.00  0.00  0.00  0.00
02:30:14 333B.0   -             walk    93   LeftBed  walk               room -    OpenFloor  3   133   0.00  0.00  1.00  0.00  0.00  0.00
02:30:14 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   134   0.00  0.00  1.00  0.00  0.00  0.00
02:30:14 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   134   0.00  0.00  1.00  0.00  0.00  0.00
02:30:14 CD2B.2   CD2B22822912  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   134   0.00  0.00  1.00  0.00  0.00  0.00
02:30:15 333B.0   -             walk    104  LeftBed  walk               room -    OpenFloor  3   134   0.00  0.00  1.00  0.00  0.00  0.00
02:30:15 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   135   0.00  0.00  1.00  0.00  0.00  0.00
02:30:15 CD2B.2   CD2B22822912  stand   61   LeftBed  stand              trk  1.00 OpenFloor  3   135   0.00  0.00  1.00  0.00  0.00  0.00
02:30:15 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   135   0.00  0.00  1.00  0.00  0.00  0.00
02:30:16 333B.0   -             walk    86   LeftBed  walk               room -    OpenFloor  3   135   0.00  0.00  1.00  0.00  0.00  0.00
02:30:16 CD2B.2   CD2B22822912  lying   51   LeftBed  lying              trk  1.00 OpenFloor  3   136   0.00  0.00  0.99  0.00  0.00  0.00
02:30:16 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   136   0.00  0.00  1.00  0.00  0.00  0.00
02:30:16 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   136   0.00  0.00  1.00  0.00  0.00  0.00
02:30:17 333B.0   -             walk    92   LeftBed  walk               room -    OpenFloor  3   136   0.00  0.00  0.99  0.00  0.00  0.00
02:30:17 CD2B.2   CD2B22822912  lying   48   LeftBed  lying              trk  1.00 OpenFloor  3   137   0.00  0.00  0.99  0.00  0.00  0.00
02:30:17 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   137   0.00  0.00  1.00  0.00  0.00  0.00
02:30:17 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   137   0.00  0.00  1.00  0.00  0.00  0.00
02:30:18 333B.0   -             walk    85   LeftBed  walk               room -    OpenFloor  3   137   0.00  0.00  0.99  0.00  0.00  0.00
02:30:18 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   138   0.00  0.00  1.00  0.00  0.00  0.00
02:30:18 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   138   0.00  0.00  1.00  0.00  0.00  0.00
02:30:18 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   138   0.00  0.00  0.99  0.00  0.00  0.00
02:30:19 333B.0   -             walk    99   LeftBed  walk               room -    OpenFloor  3   138   0.00  0.00  0.99  0.00  0.00  0.00
02:30:19 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   139   0.00  0.00  1.00  0.00  0.00  0.00
02:30:19 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   139   0.00  0.00  1.00  0.00  0.00  0.00
02:30:19 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   139   0.00  0.00  0.99  0.00  0.00  0.00
02:30:20 333B.0   -             walk    88   LeftBed  walk               room -    OpenFloor  3   139   0.00  0.00  0.99  0.00  0.00  0.00
02:30:20 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   140   0.00  0.00  1.00  0.00  0.00  0.00
02:30:20 CD2B.2   CD2B22822912  lying   52   LeftBed  lying              trk  1.00 OpenFloor  3   140   0.00  0.00  0.99  0.00  0.00  0.00
02:30:20 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   140   0.00  0.00  1.00  0.00  0.00  0.00
02:30:21 333B.0   -             walk    125  LeftBed  walk               room -    OpenFloor  3   140   0.00  0.00  0.99  0.00  0.00  0.00
02:30:21 CD2B.2   CD2B22822912  lying   46   LeftBed  lying              trk  1.00 OpenFloor  3   141   0.00  0.00  0.99  0.00  0.00  0.00
02:30:21 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   141   0.00  0.00  1.00  0.00  0.00  0.00
02:30:21 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   141   0.00  0.00  1.00  0.00  0.00  0.00
02:30:22 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   141   0.00  0.00  0.99  0.00  0.00  0.00
02:30:22 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   142   0.00  0.00  1.00  0.00  0.00  0.00
02:30:22 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   142   0.00  0.00  1.00  0.00  0.00  0.00
02:30:22 CD2B.2   CD2B22822912  lying   54   LeftBed  lying              trk  1.00 OpenFloor  3   142   0.01  0.01  0.98  0.00  0.00  0.00
02:30:23 333B.0   -             stand   142  LeftBed  stand              room -    OpenFloor  3   142   0.01  0.01  0.98  0.00  0.00  0.00
02:30:23 CD2B.2   CD2B22822912  lying   55   LeftBed  lying              trk  1.00 OpenFloor  3   143   0.01  0.01  0.98  0.00  0.00  0.00
02:30:23 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   143   0.00  0.00  1.00  0.00  0.00  0.00
02:30:23 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   143   0.00  0.00  1.00  0.00  0.00  0.00
02:30:24 333B.0   -             stand   104  LeftBed  stand              room -    OpenFloor  3   143   0.01  0.01  0.98  0.00  0.00  0.00
02:30:24 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   144   0.00  0.00  1.00  0.00  0.00  0.00
02:30:24 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   144   0.00  0.00  1.00  0.00  0.00  0.00
02:30:24 CD2B.2   CD2B22822912  lying   55   LeftBed  lying              trk  1.00 OpenFloor  3   144   0.01  0.01  0.98  0.00  0.00  0.00
02:30:25 333B.0   -             stand   82   LeftBed  stand              room -    OpenFloor  3   144   0.01  0.01  0.98  0.00  0.00  0.00
02:30:25 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   145   0.01  0.01  0.98  0.00  0.00  0.00
02:30:25 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   145   0.00  0.00  1.00  0.00  0.00  0.00
02:30:25 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   145   0.00  0.00  1.00  0.00  0.00  0.00
02:30:26 333B.0   -             stand   122  LeftBed  stand              room -    OpenFloor  3   145   0.01  0.01  0.98  0.00  0.00  0.00
02:30:26 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   146   0.00  0.00  1.00  0.00  0.00  0.00
02:30:26 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   146   0.01  0.01  0.97  0.00  0.00  0.00
02:30:26 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   146   0.00  0.00  1.00  0.00  0.00  0.00
02:30:27 333B.0   -             stand   87   LeftBed  stand              room -    OpenFloor  3   146   0.01  0.01  0.97  0.00  0.00  0.00
02:30:27 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   147   0.00  0.00  0.99  0.00  0.00  0.00
02:30:27 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  3   147   0.00  0.00  1.00  0.00  0.00  0.00
02:30:27 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  3   147   0.01  0.01  0.97  0.00  0.00  0.00
02:30:28 CD2B.E   -             -       0    LeftBed  np=4               room -    OpenFloor  3   147   0.01  0.01  0.97  0.00  0.00  0.00
02:30:28 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  3   147   0.01  0.01  0.97  0.00  0.00  0.00
02:30:28 CD2B.E3  CD2B33028652  -       0    LeftBed  EnterRoom(rdr)     room -    OpenFloor  3   147   0.01  0.01  0.97  0.00  0.00  0.00
02:30:28 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   148   0.00  0.00  1.00  0.00  0.00  0.00
02:30:28 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  4   148   0.01  0.01  0.98  0.00  0.00  0.00
02:30:28 CD2B.3   CD2B33028652  stand   99   LeftBed  stand              trk  1.00 OpenFloor  4   148   0.00  0.01  0.59  0.00  0.39  0.02
02:30:28 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   148   0.00  0.00  1.00  0.00  0.00  0.00
02:30:28 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   148   0.00  0.00  1.00  0.00  0.00  0.00
02:30:28 CD2B.3   CD2B33028652  stand   99   LeftBed  stand              trk  1.00 OpenFloor  4   148   0.00  0.00  0.99  0.00  0.00  0.00
02:30:28 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  4   148   0.02  0.01  0.97  0.00  0.00  0.00
02:30:28 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   148   0.00  0.00  1.00  0.00  0.00  0.00
02:30:29 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  4   148   0.02  0.01  0.97  0.00  0.00  0.00
02:30:29 CD2B.3   CD2B33028652  stand   102  LeftBed  stand              trk  1.00 OpenFloor  4   149   0.00  0.00  1.00  0.00  0.00  0.00
02:30:29 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   149   0.00  0.00  1.00  0.00  0.00  0.00
02:30:29 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   149   0.00  0.00  1.00  0.00  0.00  0.00
02:30:29 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  4   149   0.02  0.01  0.97  0.00  0.00  0.00
02:30:30 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  4   149   0.02  0.01  0.97  0.00  0.00  0.00
02:30:30 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   150   0.00  0.00  1.00  0.00  0.00  0.00
02:30:30 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   150   0.00  0.00  1.00  0.00  0.00  0.00
02:30:30 CD2B.3   CD2B33028652  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   150   0.00  0.00  0.97  0.00  0.00  0.01
02:30:30 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  4   150   0.03  0.01  0.95  0.00  0.00  0.00
02:30:31 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  4   150   0.03  0.01  0.95  0.00  0.00  0.00
02:30:31 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   151   0.00  0.00  1.00  0.00  0.00  0.00
02:30:31 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   151   0.00  0.00  1.00  0.00  0.00  0.00
02:30:31 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  4   151   0.05  0.01  0.94  0.00  0.00  0.00
02:30:31 CD2B.3   CD2B33028652  stand   101  LeftBed  stand              trk  1.00 OpenFloor  4   151   0.00  0.00  0.99  0.00  0.00  0.00
02:30:32 333B.0   -             stand   0    LeftBed  stand              room -    OpenFloor  4   151   0.05  0.01  0.94  0.00  0.00  0.00
02:30:32 333B.E   -             -       0    LeftBed  ExitRoom(rdr)      room -    OpenFloor  4   151   0.05  0.01  0.94  0.00  0.00  0.00
02:30:32 CD2B.3   CD2B33028652  stand   84   LeftBed  stand              trk  1.00 OpenFloor  4   152   0.00  0.00  1.00  0.00  0.00  0.00
02:30:32 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  4   152   0.07  0.02  0.91  0.00  0.00  0.00
02:30:32 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   152   0.00  0.00  1.00  0.00  0.00  0.00
02:30:32 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   152   0.00  0.00  1.00  0.00  0.00  0.00
02:30:33 333B.E   -             -       0    LeftBed  np=0  ★0           room -    OpenFloor  4   152   0.07  0.02  0.91  0.00  0.00  0.00
02:30:33 333B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  4   152   0.07  0.02  0.91  0.00  0.00  0.00
02:30:33 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   153   0.00  0.00  1.00  0.00  0.00  0.00
02:30:33 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   153   0.00  0.00  1.00  0.00  0.00  0.00
02:30:33 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  4   153   0.12  0.02  0.85  0.00  0.00  0.00
02:30:33 CD2B.3   CD2B33028652  stand   68   LeftBed  stand              trk  1.00 OpenFloor  4   153   0.00  0.00  0.99  0.00  0.00  0.00
02:30:34 333B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  4   153   0.12  0.02  0.85  0.00  0.00  0.00
02:30:34 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   154   0.00  0.00  1.00  0.00  0.00  0.00
02:30:34 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  4   154   0.21  0.02  0.76  0.00  0.00  0.00
02:30:34 CD2B.3   CD2B33028652  stand   100  LeftBed  stand              trk  1.00 OpenFloor  4   154   0.00  0.00  1.00  0.00  0.00  0.00
02:30:34 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   154   0.00  0.00  1.00  0.00  0.00  0.00
02:30:35 333B.88  -             88      -    LeftBed  no-target(88)      room -    OpenFloor  4   154   0.21  0.02  0.76  0.00  0.00  0.00
02:30:35 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   155   0.00  0.00  1.00  0.00  0.00  0.00
02:30:35 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 OpenFloor  4   155   0.00  0.00  1.00  0.00  0.00  0.00
02:30:35 CD2B.3   CD2B33028652  stand   67   LeftBed  stand              trk  1.00 OpenFloor  4   155   0.00  0.00  0.99  0.00  0.00  0.00
02:30:35 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 OpenFloor  4   155   0.35  0.03  0.62  0.00  0.00  0.00
02:30:36 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 Fallen     4   156   0.56  0.02  0.41  0.00  0.00  0.00
02:30:36 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   156   0.00  0.00  1.00  0.00  0.00  0.00
02:30:36 CD2B.3   CD2B33028652  walk    74   LeftBed  walk               trk  1.00 Fallen     4   156   0.00  0.00  0.99  0.00  0.00  0.00
02:30:36 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   156   0.00  0.00  1.00  0.00  0.00  0.00
02:30:37 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 Fallen     4   157   0.77  0.02  0.21  0.00  0.00  0.00
02:30:37 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   157   0.00  0.00  1.00  0.00  0.00  0.00
02:30:37 CD2B.3   CD2B33028652  walk    75   LeftBed  walk               trk  1.00 Fallen     4   157   0.00  0.00  1.00  0.00  0.00  0.00
02:30:37 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   157   0.00  0.00  1.00  0.00  0.00  0.00
02:30:38 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 Fallen     4   158   0.90  0.01  0.09  0.00  0.00  0.00
02:30:38 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   158   0.00  0.00  1.00  0.00  0.00  0.00
02:30:38 CD2B.3   CD2B33028652  walk    76   LeftBed  walk               trk  1.00 Fallen     4   158   0.00  0.00  0.99  0.00  0.00  0.00
02:30:38 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   158   0.00  0.00  1.00  0.00  0.00  0.00
02:30:39 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 Fallen     4   159   0.96  0.00  0.03  0.00  0.00  0.00
02:30:39 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   159   0.00  0.00  1.00  0.00  0.00  0.00
02:30:39 CD2B.3   CD2B33028652  walk    61   LeftBed  walk               trk  1.00 Fallen     4   159   0.00  0.00  0.99  0.00  0.00  0.00
02:30:39 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   159   0.00  0.00  1.00  0.00  0.00  0.00
02:30:40 CD2B.1   CD2B12800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   160   0.00  0.00  1.00  0.00  0.00  0.00
02:30:40 CD2B.0   CD2B02800032  stand   0    LeftBed  stand              trk  1.00 Fallen     4   160   0.00  0.00  1.00  0.00  0.00  0.00
02:30:40 CD2B.3   CD2B33028652  walk    83   LeftBed  walk               trk  1.00 Fallen     4   160   0.00  0.00  0.97  0.00  0.00  0.01
02:30:40 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 Fallen     4   160   0.99  0.00  0.01  0.00  0.00  0.00
02:30:41 CD2B.E   -             -       0    LeftBed  np=2               room -    Fallen     4   160   0.99  0.00  0.01  0.00  0.00  0.00
02:30:41 CD2B.3   CD2B33028652  walk    75   LeftBed  walk               trk  1.00 Fallen     4   161   0.00  0.00  0.96  0.00  0.00  0.01
02:30:41 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 Fallen     4   161   0.99  0.00  0.01  0.00  0.00  0.00
02:30:42 CD2B.3   CD2B33028652  walk    0    LeftBed  walk               trk  1.00 Fallen     2   21    0.00  0.00  0.95  0.00  0.00  0.01
02:30:42 CD2B.2   CD2B22822912  lying   0    LeftBed  lying              trk  1.00 Fallen     2   21    0.99  0.00  0.01  0.00  0.00  0.00
02:30:43 CD2B.E   -             -       0    LeftBed  np=1               room -    Fallen     2   21    0.99  0.00  0.01  0.00  0.00  0.00
02:30:43 CD2B.3   CD2B33028652  walk    59   LeftBed  walk               trk  1.00 Fallen     2   22    0.00  0.02  0.86  0.00  0.00  0.03
02:30:44 CD2B.3   CD2B33028652  walk    73   LeftBed  walk               trk  1.00 Fallen     1   0     0.00  0.02  0.82  0.00  0.01  0.03
02:30:45 CD2B.3   CD2B33028652  walk    81   LeftBed  walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.81  0.00  0.01  0.03
02:30:46 CD2B.3   CD2B33028652  walk    86   LeftBed  walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.81  0.00  0.02  0.03
02:30:47 CD2B.3   CD2B33028652  walk    81   LeftBed  walk               trk  1.00 BlindOpen  1   0     0.00  0.02  0.80  0.00  0.02  0.03
02:30:48 CD2B.3   CD2B33028652  walk    72   LeftBed  walk               trk  1.00 BlindOpen  1   0     0.00  0.03  0.72  0.01  0.02  0.04
02:30:49 CD2B.3   CD2B33028652  walk    78   LeftBed  walk               trk  1.00 BlindOpen  1   0     0.00  0.02  0.77  0.01  0.03  0.03
02:30:50 CD2B.3   CD2B33028652  walk    91   LeftBed  walk               trk  1.00 BlindOpen  1   0     0.00  0.02  0.79  0.01  0.02  0.03
02:30:51 CD2B.3   CD2B33028652  walk    0    LeftBed  walk               trk  1.00 BlindOpen  1   0     0.00  0.02  0.79  0.01  0.02  0.03
02:30:52 CD2B.3   CD2B33028652  walk    0    LeftBed  walk               trk  1.00 BlindOpen  1   0     0.00  0.02  0.80  0.00  0.02  0.03
02:30:53 CD2B.3   CD2B33028652  walk    99   LeftBed  walk               trk  1.00 BlindOpen  1   0     0.00  0.02  0.80  0.00  0.02  0.03
02:30:54 333B.88  -             88      -    LeftBed  no-target(88)      room -    BlindOpen  1   0     0.05  0.04  0.18  0.08  0.21  0.02
02:30:54 CD2B.3   CD2B33028652  walk    101  LeftBed  walk               trk  1.00 BlindOpen  1   0     0.00  0.02  0.79  0.00  0.02  0.03
02:30:55 CD2B.3   CD2B33028652  walk    93   LeftBed  walk               trk  1.00 BlindOpen  1   0     0.00  0.02  0.79  0.00  0.02  0.03
02:30:56 CD2B.3   CD2B33028652  stand   0    LeftBed  stand              trk  1.00 BlindOpen  1   0     0.00  0.02  0.79  0.00  0.02  0.03
02:30:57 CD2B.3   CD2B33028652  stand   90   LeftBed  stand              trk  1.00 BlindOpen  1   0     0.00  0.02  0.79  0.00  0.02  0.03
02:30:58 CD2B.3   CD2B33028652  stand   111  LeftBed  stand              trk  1.00 BlindOpen  1   0     0.00  0.03  0.72  0.01  0.03  0.04
02:30:59 CD2B.3   CD2B33028652  stand   84   LeftBed  stand              trk  1.00 BlindOpen  1   0     0.00  0.03  0.68  0.01  0.04  0.04
02:31:00 CD2B.3   CD2B33028652  stand   61   LeftBed  stand              trk  1.00 BlindOpen  1   0     0.00  0.04  0.65  0.01  0.04  0.03
02:31:01 CD2B.3   CD2B33028652  stand   80   LeftBed  stand              trk  1.00 BlindOpen  1   0     0.00  0.03  0.72  0.01  0.04  0.03
02:31:02 CD2B.3   CD2B33028652  stand   0    LeftBed  stand              trk  1.00 BlindOpen  1   0     0.00  0.03  0.76  0.01  0.03  0.03
02:31:03 CD2B.3   CD2B33028652  stand   74   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.02  0.77  0.01  0.02  0.03
02:31:04 CD2B.3   CD2B33028652  stand   79   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.02  0.78  0.01  0.02  0.03
02:31:05 CD2B.3   CD2B33028652  stand   60   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.02  0.78  0.01  0.02  0.03
02:31:06 CD2B.3   CD2B33028652  stand   79   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.02  0.78  0.01  0.02  0.03
02:31:07 CD2B.3   CD2B33028652  stand   81   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.02  0.78  0.01  0.02  0.03
02:31:08 CD2B.3   CD2B33028652  stand   85   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.78  0.01  0.02  0.03
02:31:09 CD2B.3   CD2B33028652  stand   77   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.78  0.01  0.02  0.03
02:31:10 CD2B.3   CD2B33028652  stand   0    LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.78  0.01  0.02  0.03
02:31:11 CD2B.3   CD2B33028652  stand   63   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.77  0.01  0.02  0.03
02:31:12 CD2B.3   CD2B33028652  stand   84   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.77  0.01  0.02  0.03
02:31:13 CD2B.3   CD2B33028652  stand   99   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.77  0.01  0.02  0.03
02:31:14 CD2B.3   CD2B33028652  stand   98   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.77  0.01  0.02  0.03
02:31:15 CD2B.3   CD2B33028652  stand   87   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.77  0.01  0.02  0.03
02:31:16 CD2B.3   CD2B33028652  stand   96   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.77  0.01  0.02  0.03
02:31:17 CD2B.3   CD2B33028652  stand   88   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.77  0.01  0.02  0.03
02:31:18 CD2B.3   CD2B33028652  stand   66   LeftBed  stand              trk  1.00 Empty      1   0     0.00  0.03  0.77  0.01  0.02  0.03
02:31:19 CD2B.3   CD2B33028652  walk    65   LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.03  0.70  0.01  0.03  0.04
02:31:20 CD2B.3   CD2B33028652  walk    73   LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.02  0.80  0.01  0.02  0.02
02:31:21 CD2B.3   CD2B33028652  walk    64   LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.02  0.84  0.00  0.01  0.02
02:31:22 CD2B.3   CD2B33028652  walk    91   LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.02  0.85  0.00  0.01  0.02
02:31:23 333B.E   -             -       0    LeftBed  np=1               room -    Empty      1   0     0.14  0.05  0.15  0.09  0.22  0.02
02:31:23 CD2B.3   CD2B33028652  walk    0    LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.02  0.85  0.00  0.01  0.02
02:31:23 333B.E   -             -       0    LeftBed  EnterRoom(rdr)     room -    Empty      1   0     0.15  0.05  0.15  0.09  0.21  0.02
02:31:23 333B.0   -             stand   100  LeftBed  stand              room -    Empty      1   0     0.15  0.05  0.15  0.09  0.21  0.02
02:31:24 333B.0   -             walk    106  LeftBed  walk               room -    Empty      1   0     0.15  0.05  0.15  0.09  0.21  0.02
02:31:24 CD2B.3   CD2B33028652  sit     0    LeftBed  sit                trk  1.00 Empty      1   0     0.00  0.03  0.71  0.00  0.02  0.03
02:31:25 333B.0   -             walk    74   LeftBed  walk               room -    Empty      1   0     0.15  0.05  0.15  0.09  0.21  0.02
02:31:25 CD2B.3   CD2B33028652  sit     0    LeftBed  sit                trk  1.00 Empty      1   0     0.01  0.03  0.62  0.01  0.03  0.03
02:31:26 333B.0   -             walk    0    LeftBed  walk               room -    Empty      1   0     0.13  0.04  0.13  0.08  0.19  0.13
02:31:26 CD2B.3   CD2B33028652  sit     0    LeftBed  sit                trk  1.00 Left       1   0     0.01  0.03  0.56  0.01  0.03  0.03
02:31:27 333B.0   -             walk    0    LeftBed  walk               room -    Left       1   0     0.11  0.03  0.10  0.07  0.23  0.23
02:31:27 CD2B.3   CD2B33028652  sit     0    LeftBed  sit                trk  1.00 Empty      1   0     0.01  0.03  0.52  0.01  0.04  0.03
02:31:27 CD2B.E3  CD2B33028652  -       0    LeftBed  ExitRoom(rdr)      trk  1.00 Empty      1   0     0.01  0.03  0.52  0.01  0.04  0.03
02:31:28 333B.0   -             walk    0    LeftBed  walk               room -    Empty      1   0     0.10  0.05  0.16  0.10  0.24  0.02
02:31:28 CD2B.E   -             -       0    LeftBed  np=0  ★0           room -    Empty      1   0     0.10  0.05  0.16  0.10  0.24  0.02
02:31:28 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      1   0     0.11  0.05  0.16  0.10  0.24  0.02
02:31:29 333B.0   -             walk    116  LeftBed  walk               room -    Empty      1   0     0.11  0.05  0.16  0.10  0.24  0.02
02:31:29 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      0   0     0.11  0.05  0.16  0.10  0.24  0.02
02:31:30 333B.0   -             walk    108  LeftBed  walk               room -    Empty      0   0     0.11  0.05  0.16  0.10  0.24  0.02
02:31:30 CD2B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      0   0     0.11  0.05  0.15  0.10  0.23  0.02
02:31:31 333B.0   -             walk    108  LeftBed  walk               room -    Empty      0   0     0.11  0.05  0.15  0.10  0.23  0.02
02:31:32 333B.0   -             walk    85   LeftBed  walk               room -    Empty      0   0     0.11  0.05  0.15  0.10  0.23  0.02
02:31:33 333B.0   -             walk    117  LeftBed  walk               room -    Empty      0   0     0.11  0.05  0.15  0.10  0.23  0.02
02:31:34 333B.0   -             walk    0    LeftBed  walk               room -    Empty      0   0     0.11  0.05  0.15  0.10  0.23  0.02
02:31:34 CD2B.E   -             -       0    LeftBed  np=1               room -    Empty      0   0     0.11  0.05  0.15  0.10  0.23  0.02
02:31:34 CD2B.E   -             -       0    LeftBed  EnterRoom(rdr)     room -    Empty      0   0     0.11  0.05  0.15  0.10  0.23  0.02
02:31:34 CD2B.0   CD2B02800032  stand   70   LeftBed  stand              trk  1.00 Empty      1   0     0.03  0.02  0.29  0.02  0.55  0.01
02:31:35 333B.0   -             walk    0    LeftBed  walk               room -    Empty      1   0     0.12  0.05  0.15  0.10  0.23  0.02
02:31:35 CD2B.0   CD2B02800032  walk    67   LeftBed  walk               trk  1.00 Empty      1   0     0.01  0.02  0.45  0.02  0.37  0.02
02:31:36 333B.0   -             stand   0    LeftBed  stand              room -    Empty      1   0     0.12  0.05  0.15  0.10  0.23  0.02
02:31:36 CD2B.0   CD2B02800032  walk    82   LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.03  0.58  0.01  0.23  0.02
02:31:37 333B.0   -             stand   0    LeftBed  stand              room -    Empty      1   0     0.12  0.05  0.15  0.10  0.23  0.02
02:31:37 CD2B.0   CD2B02800032  walk    88   LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.03  0.66  0.01  0.14  0.02
02:31:38 333B.0   -             stand   0    LeftBed  stand              room -    Empty      1   0     0.12  0.05  0.15  0.10  0.23  0.02
02:31:38 CD2B.0   CD2B02800032  walk    91   LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.03  0.70  0.01  0.08  0.03
02:31:39 333B.0   -             stand   0    LeftBed  stand              room -    Empty      1   0     0.13  0.05  0.15  0.10  0.23  0.02
02:31:39 333B.E   -             -       0    LeftBed  ExitRoom(rdr)      room -    Empty      1   0     0.13  0.05  0.15  0.10  0.23  0.02
02:31:39 CD2B.0   CD2B02800032  walk    99   LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.03  0.73  0.01  0.05  0.03
02:31:40 333B.E   -             -       0    LeftBed  np=0  ★0           room -    Empty      1   0     0.13  0.05  0.15  0.10  0.22  0.02
02:31:40 333B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      1   0     0.13  0.05  0.15  0.10  0.22  0.02
02:31:40 CD2B.0   CD2B02800032  walk    64   LeftBed  walk               trk  1.00 Empty      1   0     0.00  0.02  0.78  0.01  0.03  0.02
02:31:41 333B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      1   0     0.13  0.05  0.15  0.10  0.22  0.02
02:31:41 CD2B.0   CD2B02800032  lying   53   LeftBed  lying              trk  1.00 Empty      1   0     0.01  0.11  0.62  0.01  0.03  0.04
02:31:42 333B.88  -             88      -    LeftBed  no-target(88)      room -    Empty      1   0     0.13  0.05  0.15  0.10  0.22  0.02
02:31:42 CD2B.0   CD2B02800032  lying   53   LeftBed  lying              trk  1.00 Empty      1   0     0.03  0.20  0.49  0.01  0.04  0.03
02:31:43 CD2B.0   CD2B02800032  walk    0    LeftBed  walk               trk  1.00 Empty      1   0     0.01  0.09  0.67  0.02  0.04  0.02
02:31:44 CD2B.0   CD2B02800032  sit     0    LeftBed  sit                trk  1.00 Empty      1   0     0.01  0.07  0.57  0.02  0.04  0.04
02:31:45 CD2B.0   CD2B02800032  sit     0    LeftBed  sit                trk  1.00 Empty      1   0     0.01  0.06  0.50  0.02  0.05  0.03
02:31:46 CD2B.0   CD2B02800032  sit     0    LeftBed  sit                trk  1.00 Empty      1   0     0.01  0.05  0.44  0.03  0.05  0.03
02:31:47 CD2B.0   CD2B02800032  sit     0    LeftBed  sit                trk  1.00 Empty      1   0     0.01  0.04  0.39  0.03  0.05  0.03
02:31:48 CD2B.0   CD2B02800032  sit     0    LeftBed  sit                trk  1.00 Empty      1   0     0.01  0.03  0.35  0.03  0.05  0.03
02:31:49 CD2B.0   CD2B02800032  sit     0    LeftBed  sit                trk  1.00 Empty      1   0     0.01  0.03  0.31  0.03  0.04  0.02
02:31:50 1641.0   -             pad     -    LeftBed  pad InBed HR=68 RR=21 mv=0 turn=0 room -    Empty      1   0     0.16  0.06  0.15  0.10  0.21  0.02
02:31:50 CD2B.0   CD2B02800032  sit     0    InBed    sit                trk  1.00 Bed        1   0     0.01  0.15  0.17  0.03  0.04  0.02
02:31:51 CD2B.0   CD2B02800032  sit     0    InBed    sit                trk  1.00 Bed        1   0     0.01  0.35  0.10  0.03  0.03  0.02
02:31:51 1641.E   -             -       0    InBed    InBed(pad)         room -    Bed        1   0     0.01  0.96  0.01  0.01  0.01  0.00
02:31:51 1641.E   -             -       0    InBed    InBed(pad)         room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:31:52 1641.0   -             pad     -    InBed    pad InBed HR=63 RR=20 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:31:52 CD2B.0   CD2B02800032  sit     0    InBed    sit                trk  1.00 Bed        1   0     0.00  0.91  0.02  0.03  0.00  0.00
02:31:53 CD2B.0   CD2B02800032  sit     0    InBed    sit                trk  1.00 Bed        1   0     0.00  0.92  0.02  0.03  0.00  0.00
02:31:54 1641.0   -             pad     -    InBed    pad InBed HR=66 RR=20 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:31:54 CD2B.0   CD2B02800032  sit     0    InBed    sit                trk  1.00 Bed        1   0     0.00  0.93  0.02  0.03  0.00  0.00
02:31:55 CD2B.0   CD2B02800032  sit     0    InBed    sit                trk  1.00 Bed        1   0     0.00  0.93  0.02  0.03  0.00  0.00
02:31:56 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=20 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:31:56 CD2B.0   CD2B02800032  sit     0    InBed    sit                trk  1.00 Bed        1   0     0.00  0.93  0.02  0.03  0.00  0.00
02:31:57 CD2B.0   CD2B02800032  sit     0    InBed    sit                trk  1.00 Bed        1   0     0.00  0.93  0.02  0.03  0.00  0.00
02:31:57 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:31:58 CD2B.0   CD2B02800032  sit     0    InBed    sit                trk  1.00 Bed        1   0     0.00  0.93  0.02  0.03  0.00  0.00
02:31:59 1641.0   -             pad     -    InBed    pad InBed HR=53 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:31:59 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.92  0.04  0.03  0.00  0.00
02:32:00 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.91  0.05  0.03  0.00  0.00
02:32:01 1641.0   -             pad     -    InBed    pad InBed HR=52 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:01 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.91  0.05  0.03  0.00  0.00
02:32:02 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.95  0.03  0.02  0.00  0.00
02:32:02 CD2B.E   -             -       0    InBed    InBed(rdr)         room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:03 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=17 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:03 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:04 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:05 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:05 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:06 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:07 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=18 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:07 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:08 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:09 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=15 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:09 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:10 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:11 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:11 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:12 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:13 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:13 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:14 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:15 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:15 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:15 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:16 CD2B.0   CD2B02800032  walk    49   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:17 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:17 CD2B.0   CD2B02800032  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:18 CD2B.0   CD2B02800032  walk    44   InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:19 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=None mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:19 CD2B.0   CD2B02800032  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:20 CD2B.0   CD2B02800032  walk    0    InBed    walk               trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:21 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=12 mv=1 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:21 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:22 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:23 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:23 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:24 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:25 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:25 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:26 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:27 1641.0   -             pad     -    InBed    pad InBed HR=56 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:27 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:28 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:29 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=11 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:29 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:30 333B.88  -             88      -    InBed    no-target(88)      room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:30 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:31 1641.0   -             pad     -    InBed    pad InBed HR=68 RR=11 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:31 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:32 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:33 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:33 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:34 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:35 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:35 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:36 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:37 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:38 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:38 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:39 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:40 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:40 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:41 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:42 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:42 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:43 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:44 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=13 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:44 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:45 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   0     0.00  0.96  0.02  0.01  0.00  0.00
02:32:46 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:46 CD2B.0   CD2B02800032  stand   65   InBed    stand              trk  1.00 Bed        1   28    0.00  0.96  0.02  0.01  0.00  0.00
02:32:47 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   29    0.00  0.96  0.02  0.01  0.00  0.00
02:32:48 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   29    0.00  0.94  0.02  0.03  0.00  0.00
02:32:48 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   30    0.00  0.96  0.02  0.01  0.00  0.00
02:32:49 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   31    0.00  0.96  0.02  0.01  0.00  0.00
02:32:50 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=13 mv=0 turn=0 room -    Bed        1   31    0.00  0.94  0.02  0.03  0.00  0.00
02:32:50 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   32    0.00  0.96  0.02  0.01  0.00  0.00
02:32:51 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   33    0.00  0.96  0.02  0.01  0.00  0.00
02:32:52 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   33    0.00  0.94  0.02  0.03  0.00  0.00
02:32:52 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   34    0.00  0.96  0.02  0.01  0.00  0.00
02:32:53 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   35    0.00  0.96  0.02  0.01  0.00  0.00
02:32:54 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=12 mv=0 turn=0 room -    Bed        1   35    0.00  0.94  0.02  0.03  0.00  0.00
02:32:54 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   36    0.00  0.96  0.02  0.01  0.00  0.00
02:32:55 CD2B.0   CD2B02800032  stand   0    InBed    stand              trk  1.00 Bed        1   37    0.00  0.96  0.02  0.01  0.00  0.00
02:32:56 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=13 mv=0 turn=0 room -    Bed        1   37    0.00  0.94  0.02  0.03  0.00  0.00
02:32:56 CD2B.0   CD2B02800032  lying   45   InBed    lying              trk  1.00 Bed        1   38    0.00  0.99  0.00  0.00  0.00  0.00
02:32:57 CD2B.0   CD2B02800032  lying   44   InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
02:32:58 1641.0   -             pad     -    InBed    pad InBed HR=64 RR=14 mv=0 turn=0 room -    Bed        1   0     0.00  0.94  0.02  0.03  0.00  0.00
02:32:58 CD2B.0   CD2B02800032  lying   0    InBed    lying              trk  1.00 Bed        1   0     0.00  0.99  0.00  0.00  0.00  0.00
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
02:28:00.871 333B.0    stand  1    -270   100    0     80        
02:28:01.921 333B.0    stand  1    -270   60     0     80   40   
02:28:02.772 333B.0    stand  1    -280   110    77    80   50   
02:28:03.773 333B.0    stand  1    -320   100    0     80   41   
02:28:04.772 333B.0    stand  1    -330   110    0     80   14   
02:28:05.770 333B.0    stand  1    -320   110    0     80   10   
02:28:06.772 333B.0    stand  1    -320   110    0     80   0    
02:28:07.773 333B.0    stand  1    -320   110    0     80   0    
02:28:08.804 333B.0    stand  1    -320   110    0     80   0    
02:28:09.802 333B.0    stand  1    -320   110    0     80   0    
02:28:10.705 333B.0    stand  1    -320   110    0     80   0    
02:28:11.704 333B.0    stand  1    -320   110    0     80   0    
02:28:12.706 333B.0    stand  1    -320   110    0     80   0    
02:28:13.713 333B.0    stand  1    -320   110    0     80   0    
02:28:14.705 333B.0    stand  1    -320   110    0     80   0    
02:28:15.706 333B.0    stand  1    -320   110    0     80   0    
02:28:16.720 333B.0    stand  1    -320   110    0     80   0    
02:28:17.713 333B.0    stand  1    -310   110    0     80   10   
02:28:18.710 333B.0    stand  1    -280   100    104   80   31   
02:28:19.710 333B.0    walk   1    -220   100    86    80   60   
02:28:20.721 333B.0    walk   1    -120   150    87    80   111  
02:28:21.718 333B.0    walk   1    -30    180    97    80   94   
02:28:22.609 333B.0    walk   1    0      170    0     80   31   
02:28:23.606 333B.0    walk   1    0      170    0     80   0    
02:28:24.614 333B.0    sit    1    0      170    0     80   0    
02:28:25.612 333B.0    sit    1    0      170    0     80   0    
02:28:26.616 333B.0    sit    1    0      170    0     80   0    
02:28:27.619 333B.0    sit    1    0      170    0     80   0    
02:28:28.673 333B.88   88     -    -      -      -     -    -    
02:28:29.624 333B.88   88     -    -      -      -     -    -    
02:28:30.625 333B.88   88     -    -      -      -     -    -    
02:28:47.428 333B.88   88     -    -      -      -     -    -    
02:29:09.454 333B.0    stand  1    -40    160    100   80   41   
02:29:10.225 333B.0    stand  1    -50    170    100   80   14   
02:29:11.228 333B.0    stand  1    -90    160    119   80   41   
02:29:12.242 333B.0    walk   1    -150   150    102   80   60   
02:29:13.244 333B.0    walk   1    -210   110    110   80   72   
02:29:14.248 333B.0    walk   1    -240   90     111   80   36   
02:29:15.250 333B.0    walk   1    -230   100    112   80   14   
02:29:16.246 333B.0    walk   1    -240   90     93    80   14   
02:29:17.248 333B.0    walk   1    -270   90     100   80   30   
02:29:18.248 333B.0    walk   1    -310   90     82    80   40   
02:29:19.249 333B.0    walk   1    -310   80     103   80   10   
02:29:20.144 333B.0    walk   1    -280   50     126   80   42   
02:29:21.144 333B.0    walk   1    -250   40     122   80   31   
02:29:22.157 333B.0    walk   1    -250   40     136   80   0    
02:29:23.147 333B.0    walk   1    -270   10     0     80   36   
02:29:24.153 333B.0    walk   1    -310   0      0     80   41   
02:29:25.149 333B.0    stand  1    -310   0      0     80   0    
02:29:26.152 333B.0    stand  1    -310   0      0     80   0    
02:29:27.151 333B.0    stand  1    -300   0      0     80   10   
02:29:28.155 333B.0    stand  1    -280   0      0     80   20   
02:29:29.162 333B.0    stand  1    -290   10     0     80   14   
02:29:30.158 333B.0    stand  1    -270   30     88    80   28   
02:29:31.061 333B.0    stand  1    -250   30     106   80   20   
02:29:32.057 333B.0    stand  1    -260   30     101   80   10   
02:29:33.056 333B.0    stand  1    -290   30     108   80   30   
02:29:34.058 333B.0    stand  1    -250   30     98    80   40   
02:29:35.061 333B.0    stand  1    -280   10     67    80   36   
02:29:36.063 333B.0    stand  1    -320   30     0     80   44   
02:29:37.061 333B.0    stand  1    -290   20     0     80   31   
02:29:38.062 333B.0    stand  1    -310   20     43    80   20   
02:29:39.070 333B.0    stand  1    -310   0      0     80   20   
02:29:40.065 333B.0    stand  1    -310   20     0     80   20   
02:29:41.065 333B.0    stand  1    -250   10     95    80   60   
02:29:42.072 333B.0    stand  1    -240   0      0     80   14   
02:29:42.961 333B.0    stand  1    -260   10     154   80   22   
02:29:43.966 333B.0    stand  1    -240   0      107   80   22   
02:29:44.964 333B.0    stand  1    -240   10     131   80   10   
02:29:45.964 333B.0    stand  1    -300   0      0     80   60   
02:29:46.966 333B.0    stand  1    -250   0      161   80   50   
02:29:47.966 333B.0    stand  1    -260   10     92    80   14   
02:29:48.965 333B.0    stand  1    -310   40     0     80   58   
02:29:49.968 333B.0    stand  1    -310   0      29    80   40   
02:29:50.968 333B.0    stand  1    -310   10     127   80   10   
02:29:51.969 333B.0    stand  1    -270   0      0     80   41   
02:29:52.972 333B.0    stand  1    -250   0      0     80   20   
02:29:53.978 333B.0    stand  1    -250   0      143   80   0    
02:29:54.865 333B.0    walk   1    -200   0      114   80   50   
02:29:55.864 333B.0    walk   1    -240   20     102   80   44   
02:29:56.868 333B.0    walk   1    -260   0      143   80   28   
02:29:57.871 333B.0    walk   1    -240   10     127   80   22   
02:29:58.876 333B.0    walk   1    -260   30     122   80   28   
02:29:59.871 333B.0    stand  1    -240   0      0     80   36   
02:30:00.925 333B.0    stand  1    -240   0      0     80   0    
02:30:01.880 333B.0    stand  1    -250   -10    0     80   14   
02:30:02.880 333B.0    stand  1    -240   0      109   80   14   
02:30:03.879 333B.0    stand  1    -230   10     0     80   14   
02:30:04.883 333B.0    stand  1    -270   20     0     80   41   
02:30:05.774 333B.0    stand  1    -260   0      0     80   22   
02:30:06.780 333B.0    stand  1    -230   0      128   80   30   
02:30:07.776 333B.0    stand  1    -230   0      125   80   0    
02:30:08.784 333B.0    stand  1    -250   50     102   80   53   
02:30:09.784 333B.0    stand  1    -270   60     92    80   22   
02:30:10.777 333B.0    walk   1    -250   100    93    80   44   
02:30:11.779 333B.0    walk   1    -250   110    93    80   10   
02:30:12.790 333B.0    walk   1    -240   90     135   80   22   
02:30:13.783 333B.0    walk   1    -240   100    97    80   10   
02:30:14.784 333B.0    walk   1    -220   110    93    80   22   
02:30:15.784 333B.0    walk   1    -160   130    104   80   63   
02:30:16.785 333B.0    walk   1    -90    180    86    80   86   
02:30:17.681 333B.0    walk   1    -50    200    92    80   44   
02:30:18.684 333B.0    walk   1    -10    200    85    80   40   
02:30:19.684 333B.0    walk   1    -10    190    99    80   10   
02:30:20.705 333B.0    walk   1    -50    190    88    80   40   
02:30:21.692 333B.0    walk   1    -60    180    125   80   14   
02:30:22.688 333B.0    stand  1    -10    170    0     80   50   
02:30:23.694 333B.0    stand  1    -70    190    142   80   63   
02:30:24.690 333B.0    stand  1    -60    200    104   80   14   
02:30:25.689 333B.0    stand  1    -20    200    82    80   40   
02:30:26.690 333B.0    stand  1    -10    200    122   80   10   
02:30:27.692 333B.0    stand  1    0      190    87    80   14   
02:30:28.596 333B.0    stand  1    0      190    0     80   0    
02:30:29.594 333B.0    stand  1    0      190    0     80   0    
02:30:30.598 333B.0    stand  1    0      190    0     80   0    
02:30:31.592 333B.0    stand  1    0      190    0     80   0    
02:30:32.594 333B.0    stand  1    0      190    0     80   0    
02:30:33.649 333B.88   88     -    -      -      -     -    -    
02:30:34.604 333B.88   88     -    -      -      -     -    -    
02:30:35.608 333B.88   88     -    -      -      -     -    -    
02:30:54.380 333B.88   88     -    -      -      -     -    -    
02:31:23.669 333B.0    stand  1    -50    150    100   80   64   
02:31:24.149 333B.0    walk   1    -150   150    106   80   100  
02:31:25.148 333B.0    walk   1    -240   90     74    80   108  
02:31:26.148 333B.0    walk   1    -340   120    0     80   104  
02:31:27.156 333B.0    walk   1    -350   130    0     80   14   
02:31:28.153 333B.0    walk   1    -340   130    0     80   10   
02:31:29.152 333B.0    walk   1    -340   130    116   80   0    
02:31:30.152 333B.0    walk   1    -280   110    108   80   63   
02:31:31.165 333B.0    walk   1    -190   110    108   80   90   
02:31:32.160 333B.0    walk   1    -90    140    85    80   104  
02:31:33.156 333B.0    walk   1    0      190    117   80   102  
02:31:34.163 333B.0    walk   1    0      170    0     80   20   
02:31:35.054 333B.0    walk   1    0      170    0     80   0    
02:31:36.140 333B.0    stand  1    0      170    0     80   0    
02:31:37.040 333B.0    stand  1    0      170    0     80   0    
02:31:38.040 333B.0    stand  1    -10    170    0     80   10   
02:31:39.041 333B.0    stand  1    -10    170    0     80   0    
02:31:40.091 333B.88   88     -    -      -      -     -    -    
02:31:41.070 333B.88   88     -    -      -      -     -    -    
02:31:42.052 333B.88   88     -    -      -      -     -    -    
02:31:57.956 333B.88   88     -    -      -      -     -    -    
02:32:30.021 333B.88   88     -    -      -      -     -    -    

02:28:00.032 CD2B.1    stand  1    -60    150    0     80        
02:28:00.032 CD2B.0    stand  255  -170   320    0     80   202  
02:28:01.040 CD2B.1    stand  1    -50    150    0     80   208  
02:28:01.040 CD2B.0    stand  255  -170   320    0     80   208  
02:28:02.036 CD2B.1    stand  1    -50    150    0     80   208  
02:28:02.036 CD2B.0    stand  255  -170   320    0     80   208  
02:28:03.085 CD2B.0    stand  255  -170   320    0     80   0    
02:28:03.085 CD2B.1    stand  1    -60    150    0     80   202  
02:28:04.037 CD2B.1    stand  1    -60    150    0     80   0    
02:28:04.037 CD2B.0    stand  255  -170   320    0     80   202  
02:28:05.042 CD2B.0    stand  255  -170   320    0     80   0    
02:28:05.042 CD2B.1    stand  1    -50    150    0     80   208  
02:28:06.042 CD2B.1    stand  1    -50    150    0     80   0    
02:28:06.042 CD2B.0    stand  255  -170   320    0     80   208  
02:28:07.046 CD2B.1    stand  1    -50    150    59    80   208  
02:28:07.046 CD2B.0    stand  255  -170   320    0     80   208  
02:28:07.934 CD2B.0    stand  255  -170   320    0     80   0    
02:28:07.934 CD2B.1    stand  1    -50    150    0     80   208  
02:28:08.969 CD2B.0    stand  255  -170   320    0     80   208  
02:28:08.969 CD2B.1    stand  1    -50    150    0     80   208  
02:28:09.942 CD2B.1    stand  1    -40    150    0     80   10   
02:28:09.942 CD2B.0    stand  255  -170   320    0     80   214  
02:28:10.946 CD2B.0    stand  255  -170   320    0     80   0    
02:28:10.946 CD2B.1    stand  1    -40    150    0     80   214  
02:28:11.940 CD2B.0    stand  255  -170   320    0     80   214  
02:28:11.940 CD2B.1    stand  1    -20    150    0     80   226  
02:28:12.940 CD2B.1    stand  1    -10    150    0     80   10   
02:28:12.940 CD2B.0    stand  255  -170   320    0     80   233  
02:28:13.941 CD2B.0    stand  255  -170   320    0     80   0    
02:28:13.941 CD2B.1    stand  1    -10    150    0     80   233  
02:28:14.942 CD2B.1    stand  1    -10    150    0     80   0    
02:28:14.942 CD2B.0    stand  255  -170   320    0     80   233  
02:28:15.969 CD2B.0    stand  255  -170   320    0     80   0    
02:28:15.969 CD2B.1    stand  1    -10    150    0     80   233  
02:28:16.960 CD2B.0    stand  255  -170   320    0     80   233  
02:28:16.960 CD2B.1    stand  1    -10    150    0     80   233  
02:28:17.854 CD2B.0    stand  255  -170   320    0     80   233  
02:28:17.854 CD2B.1    stand  1    -10    150    0     80   233  
02:28:18.854 CD2B.0    stand  255  -170   320    0     80   233  
02:28:18.854 CD2B.1    stand  1    -10    150    0     80   233  
02:28:19.854 CD2B.0    stand  255  -170   320    0     80   233  
02:28:19.854 CD2B.1    stand  1    -10    150    0     80   233  
02:28:20.861 CD2B.0    stand  255  -170   320    0     80   233  
02:28:20.861 CD2B.1    stand  1    -10    150    0     80   233  
02:28:21.864 CD2B.0    stand  255  -170   320    0     80   233  
02:28:21.864 CD2B.1    stand  1    -10    150    0     80   233  
02:28:22.912 CD2B.0    stand  255  -170   320    0     80   233  
02:28:22.912 CD2B.1    stand  1    -10    150    0     80   233  
02:28:22.912 CD2B.2    stand  None -130   0      70    80   192  
02:28:23.869 CD2B.2    stand  None -140   50     98    80   50   
02:28:23.869 CD2B.1    stand  1    -10    150    0     80   164  
02:28:23.869 CD2B.0    stand  255  -170   320    0     80   233  
02:28:24.876 CD2B.1    stand  1    -50    200    0     80   169  
02:28:24.876 CD2B.0    stand  255  -170   320    0     80   169  
02:28:24.876 CD2B.2    walk   None -140   100    81    80   222  
02:28:25.881 CD2B.2    walk   None -140   100    54    80   0    
02:28:25.881 CD2B.0    stand  255  -170   320    0     80   222  
02:28:25.881 CD2B.1    stand  1    -90    220    0     80   128  
02:28:26.874 CD2B.1    stand  1    -90    220    0     80   0    
02:28:26.874 CD2B.0    stand  255  -170   320    0     80   128  
02:28:26.874 CD2B.2    walk   None -140   100    67    80   222  
02:28:27.878 CD2B.0    stand  255  -170   320    0     80   222  
02:28:27.878 CD2B.1    stand  1    -90    220    0     80   128  
02:28:27.878 CD2B.2    walk   None -140   90     60    80   139  
02:28:28.772 CD2B.2    walk   None -120   80     67    80   22   
02:28:28.772 CD2B.0    stand  255  -170   320    0     80   245  
02:28:28.772 CD2B.1    stand  1    -10    180    0     80   212  
02:28:29.772 CD2B.0    stand  255  -170   320    0     80   212  
02:28:29.772 CD2B.1    stand  1    0      180    0     80   220  
02:28:29.772 CD2B.2    walk   None -80    90     88    80   120  
02:28:30.774 CD2B.0    stand  255  -170   320    0     80   246  
02:28:30.774 CD2B.2    walk   None -90    110    71    80   224  
02:28:30.774 CD2B.1    stand  1    0      190    0     80   120  
02:28:31.781 CD2B.2    walk   None -90    130    67    80   108  
02:28:31.781 CD2B.0    stand  255  -170   320    0     80   206  
02:28:31.781 CD2B.1    stand  1    -10    250    0     80   174  
02:28:32.777 CD2B.2    walk   None -80    140    71    80   130  
02:28:32.777 CD2B.0    stand  255  -170   320    0     80   201  
02:28:32.777 CD2B.1    stand  1    0      260    0     80   180  
02:28:33.790 CD2B.0    stand  255  -170   320    0     80   180  
02:28:33.790 CD2B.2    walk   None -80    120    84    80   219  
02:28:33.790 CD2B.1    stand  1    10     220    0     80   134  
02:28:34.780 CD2B.2    walk   None -60    80     86    80   156  
02:28:34.780 CD2B.1    stand  1    10     220    0     80   156  
02:28:34.780 CD2B.0    stand  255  -170   320    0     80   205  
02:28:35.780 CD2B.1    stand  1    0      220    0     80   197  
02:28:35.780 CD2B.2    walk   None -60    90     81    80   143  
02:28:35.780 CD2B.0    stand  255  -170   320    0     80   254  
02:28:36.786 CD2B.2    walk   None -60    80     74    80   264  
02:28:36.786 CD2B.0    stand  255  -170   320    0     80   264  
02:28:36.786 CD2B.1    stand  1    0      220    0     80   197  
02:28:37.785 CD2B.1    stand  1    0      220    0     80   0    
02:28:37.785 CD2B.2    walk   None -60    90     71    80   143  
02:28:37.785 CD2B.0    stand  255  -170   320    0     80   254  
02:28:38.786 CD2B.2    walk   None -60    90     74    80   254  
02:28:38.786 CD2B.1    stand  1    -10    200    0     80   120  
02:28:38.786 CD2B.0    stand  255  -170   320    0     80   200  
02:28:39.682 CD2B.2    walk   None -70    90     72    80   250  
02:28:39.682 CD2B.1    stand  1    -10    200    0     80   125  
02:28:39.682 CD2B.0    stand  255  -170   320    0     80   200  
02:28:40.696 CD2B.2    walk   None -70    90     76    80   250  
02:28:40.696 CD2B.0    stand  255  -170   320    0     80   250  
02:28:40.696 CD2B.1    stand  1    -10    200    0     80   200  
02:28:41.685 CD2B.0    stand  255  -170   320    0     80   200  
02:28:41.685 CD2B.2    walk   None -70    100    70    80   241  
02:28:41.685 CD2B.1    stand  1    -20    200    0     80   111  
02:28:42.690 CD2B.0    stand  255  -170   320    0     80   192  
02:28:42.690 CD2B.2    walk   None -70    90     70    80   250  
02:28:42.690 CD2B.1    stand  1    -20    200    0     80   120  
02:28:43.686 CD2B.2    walk   None -70    90     80    80   120  
02:28:43.686 CD2B.1    stand  1    0      230    0     80   156  
02:28:43.686 CD2B.0    stand  255  -170   320    0     80   192  
02:28:44.688 CD2B.1    stand  1    10     230    0     80   201  
02:28:44.688 CD2B.2    walk   None -70    80     76    80   170  
02:28:44.688 CD2B.0    stand  255  -170   320    0     80   260  
02:28:45.688 CD2B.0    stand  255  -170   320    0     80   0    
02:28:45.688 CD2B.2    walk   None -70    90     61    80   250  
02:28:45.688 CD2B.1    stand  1    0      230    0     80   156  
02:28:46.692 CD2B.0    stand  255  -170   320    0     80   192  
02:28:46.692 CD2B.2    walk   None -80    100    73    80   237  
02:28:46.692 CD2B.1    stand  1    10     240    0     80   166  
02:28:47.624 CD2B.2    walk   None -80    100    72    80   166  
02:28:47.624 CD2B.1    stand  1    0      240    0     80   161  
02:28:47.624 CD2B.0    stand  255  -170   320    0     80   187  
02:28:48.648 CD2B.2    walk   None -80    120    69    80   219  
02:28:48.648 CD2B.1    stand  1    0      240    0     80   144  
02:28:48.648 CD2B.0    stand  255  -170   320    0     80   187  
02:28:49.624 CD2B.1    stand  1    0      230    0     80   192  
02:28:49.624 CD2B.2    walk   None -80    150    75    80   113  
02:28:49.624 CD2B.0    stand  255  -170   320    0     80   192  
02:28:50.622 CD2B.1    stand  1    0      230    0     80   192  
02:28:50.622 CD2B.2    walk   None -70    140    76    80   114  
02:28:50.622 CD2B.0    stand  255  -170   320    0     80   205  
02:28:51.624 CD2B.2    sit    None -60    120    102   80   228  
02:28:51.624 CD2B.1    stand  1    0      230    0     80   125  
02:28:51.624 CD2B.0    stand  255  -170   320    0     80   192  
02:28:52.626 CD2B.1    stand  1    0      230    0     80   192  
02:28:52.626 CD2B.0    stand  255  -170   320    0     80   192  
02:28:52.626 CD2B.2    sit    None -70    150    74    80   197  
02:28:53.632 CD2B.0    stand  255  -170   320    0     80   197  
02:28:53.632 CD2B.2    sit    None -60    160    72    80   194  
02:28:53.632 CD2B.1    stand  1    0      230    0     80   92   
02:28:54.629 CD2B.1    stand  1    0      230    0     80   0    
02:28:54.629 CD2B.0    stand  255  -170   320    0     80   192  
02:28:54.629 CD2B.2    sit    None -70    150    84    80   197  
02:28:55.629 CD2B.1    stand  1    0      230    0     80   106  
02:28:55.629 CD2B.2    sit    None -60    130    72    80   116  
02:28:55.629 CD2B.0    stand  255  -170   320    0     80   219  
02:28:56.632 CD2B.1    stand  1    0      230    0     80   192  
02:28:56.632 CD2B.0    stand  255  -170   320    0     80   192  
02:28:56.632 CD2B.2    sit    None -50    120    75    80   233  
02:28:57.630 CD2B.2    sit    None -50    110    79    80   10   
02:28:57.630 CD2B.0    stand  255  -170   320    0     80   241  
02:28:57.630 CD2B.1    stand  1    0      230    0     80   192  
02:28:58.533 CD2B.2    sit    None -50    100    74    80   139  
02:28:58.533 CD2B.0    stand  255  -170   320    0     80   250  
02:28:58.533 CD2B.1    stand  1    0      230    0     80   192  
02:28:59.534 CD2B.1    stand  1    0      230    0     80   0    
02:28:59.534 CD2B.2    sit    None -50    120    70    80   120  
02:28:59.534 CD2B.0    stand  255  -170   320    0     80   233  
02:29:00.542 CD2B.0    stand  255  -170   320    0     80   0    
02:29:00.542 CD2B.1    stand  1    0      230    0     80   192  
02:29:00.542 CD2B.2    sit    None -50    110    75    80   130  
02:29:01.532 CD2B.0    stand  255  -170   320    0     80   241  
02:29:01.532 CD2B.2    sit    None -40    130    70    80   230  
02:29:01.532 CD2B.1    stand  1    0      230    0     80   107  
02:29:02.591 CD2B.0    stand  255  -170   320    0     80   192  
02:29:02.591 CD2B.2    sit    None -50    140    63    80   216  
02:29:02.591 CD2B.1    walk   1    -120   180    0     80   80   
02:29:03.494 CD2B.0    stand  255  -170   320    0     80   148  
02:29:03.494 CD2B.2    sit    None -50    120    71    80   233  
02:29:03.494 CD2B.1    walk   1    -50    210    0     80   90   
02:29:04.500 CD2B.0    stand  255  -170   320    0     80   162  
02:29:04.500 CD2B.2    sit    None -50    110    74    80   241  
02:29:04.500 CD2B.1    walk   1    -60    240    0     80   130  
02:29:05.496 CD2B.2    sit    None -50    110    69    80   130  
02:29:05.496 CD2B.0    stand  255  -170   320    0     80   241  
02:29:05.496 CD2B.1    walk   1    -60    240    0     80   136  
02:29:06.497 CD2B.1    walk   1    -70    230    0     80   14   
02:29:06.497 CD2B.2    sit    None -60    100    68    80   130  
02:29:06.497 CD2B.0    stand  255  -170   320    0     80   245  
02:29:07.497 CD2B.2    sit    None -90    90     92    80   243  
02:29:07.497 CD2B.1    stand  1    -70    230    0     80   141  
02:29:07.497 CD2B.0    stand  255  -170   320    0     80   134  
02:29:08.501 CD2B.1    stand  1    -70    230    0     80   134  
02:29:08.501 CD2B.0    stand  255  -170   320    0     80   134  
02:29:08.501 CD2B.2    sit    None -130   30     72    80   292  
02:29:09.502 CD2B.2    sit    None -80    80     75    80   70   
02:29:09.502 CD2B.1    stand  1    -70    230    0     80   150  
02:29:09.502 CD2B.0    stand  255  -170   320    0     80   134  
02:29:10.500 CD2B.2    sit    None -70    80     91    80   260  
02:29:10.500 CD2B.0    stand  255  -170   320    0     80   260  
02:29:10.500 CD2B.1    stand  1    -70    230    0     80   134  
02:29:11.501 CD2B.1    stand  1    -70    230    0     80   0    
02:29:11.501 CD2B.2    sit    None -70    90     83    80   140  
02:29:11.501 CD2B.0    stand  255  -170   320    0     80   250  
02:29:12.504 CD2B.0    stand  255  -170   320    0     80   0    
02:29:12.504 CD2B.2    sit    None -70    90     82    80   250  
02:29:12.504 CD2B.1    stand  1    -70    230    0     80   140  
02:29:13.502 CD2B.1    stand  1    -70    230    0     80   0    
02:29:13.502 CD2B.2    sit    None -60    70     71    80   160  
02:29:13.502 CD2B.0    stand  255  -170   320    0     80   273  
02:29:14.407 CD2B.2    sit    None -60    70     78    80   273  
02:29:14.407 CD2B.0    stand  255  -170   320    0     80   273  
02:29:14.407 CD2B.1    stand  1    -70    230    0     80   134  
02:29:15.404 CD2B.1    stand  1    -70    230    0     80   0    
02:29:15.404 CD2B.0    stand  255  -170   320    0     80   134  
02:29:15.404 CD2B.2    sit    None -60    70     66    80   273  
02:29:16.406 CD2B.1    stand  1    -70    230    0     80   160  
02:29:16.406 CD2B.0    stand  255  -170   320    0     80   134  
02:29:16.406 CD2B.2    sit    None -70    80     72    80   260  
02:29:17.404 CD2B.0    stand  255  -170   320    0     80   260  
02:29:17.404 CD2B.2    sit    None -70    80     72    80   260  
02:29:17.404 CD2B.1    stand  1    -70    230    0     80   150  
02:29:18.417 CD2B.1    stand  1    -70    230    0     80   0    
02:29:18.417 CD2B.2    sit    None -70    90     79    80   140  
02:29:18.417 CD2B.0    stand  255  -170   320    0     80   250  
02:29:19.408 CD2B.2    sit    None -60    100    58    80   245  
02:29:19.408 CD2B.0    stand  255  -170   320    0     80   245  
02:29:19.408 CD2B.1    stand  1    -70    230    0     80   134  
02:29:20.416 CD2B.0    stand  255  -170   320    0     80   134  
02:29:20.416 CD2B.2    sit    None -60    90     77    80   254  
02:29:20.416 CD2B.1    stand  1    -70    230    0     80   140  
02:29:21.411 CD2B.2    sit    None -50    70     101   80   161  
02:29:21.411 CD2B.0    stand  255  -170   320    0     80   277  
02:29:21.411 CD2B.1    stand  1    -70    230    0     80   134  
02:29:22.424 CD2B.1    stand  1    -70    230    0     80   0    
02:29:22.424 CD2B.0    stand  255  -170   320    0     80   134  
02:29:22.424 CD2B.2    sit    None -40    120    87    80   238  
02:29:23.413 CD2B.1    stand  1    -70    230    0     80   114  
02:29:23.413 CD2B.2    sit    None -30    130    84    80   107  
02:29:23.413 CD2B.0    stand  255  -170   320    0     80   236  
02:29:24.416 CD2B.2    sit    None -30    120    86    80   244  
02:29:24.416 CD2B.0    stand  255  -170   320    0     80   244  
02:29:24.416 CD2B.1    stand  1    -70    230    0     80   134  
02:29:25.314 CD2B.1    stand  1    -70    230    0     80   0    
02:29:25.314 CD2B.2    sit    None -40    110    83    80   123  
02:29:25.314 CD2B.0    stand  255  -170   320    0     80   246  
02:29:26.316 CD2B.2    sit    None -50    90     100   80   259  
02:29:26.316 CD2B.1    stand  1    -70    230    0     80   141  
02:29:26.316 CD2B.0    stand  255  -170   320    0     80   134  
02:29:27.320 CD2B.1    stand  1    -70    230    0     80   134  
02:29:27.320 CD2B.2    sit    None -50    90     89    80   141  
02:29:27.320 CD2B.0    stand  255  -170   320    0     80   259  
02:29:28.326 CD2B.1    stand  1    -70    230    0     80   134  
02:29:28.326 CD2B.2    sit    None -40    120    69    80   114  
02:29:28.326 CD2B.0    stand  255  -170   320    0     80   238  
02:29:29.336 CD2B.1    stand  1    -70    230    0     80   134  
02:29:29.336 CD2B.2    sit    None -50    140    55    80   92   
02:29:29.336 CD2B.0    stand  255  -170   320    0     80   216  
02:29:30.322 CD2B.2    sit    None -40    150    50    80   214  
02:29:30.322 CD2B.0    stand  255  -170   320    0     80   214  
02:29:30.322 CD2B.1    stand  1    -70    230    0     80   134  
02:29:31.336 CD2B.1    stand  1    -70    230    0     80   0    
02:29:31.336 CD2B.0    stand  255  -170   320    0     80   134  
02:29:31.336 CD2B.2    lying  None -70    160    50    80   188  
02:29:32.325 CD2B.1    stand  1    -70    230    0     80   70   
02:29:32.325 CD2B.0    stand  255  -170   320    0     80   134  
02:29:32.325 CD2B.2    lying  None -30    160    0     80   212  
02:29:33.326 CD2B.0    stand  255  -170   320    0     80   212  
02:29:33.326 CD2B.2    lying  None -10    160    56    80   226  
02:29:33.326 CD2B.1    stand  1    -70    230    0     80   92   
02:29:34.323 CD2B.0    stand  255  -170   320    0     80   134  
02:29:34.323 CD2B.2    lying  None -20    160    0     80   219  
02:29:34.323 CD2B.1    stand  1    -70    230    0     80   86   
02:29:35.234 CD2B.2    lying  None -20    160    0     80   86   
02:29:35.234 CD2B.1    stand  1    -70    230    0     80   86   
02:29:35.234 CD2B.0    stand  255  -170   320    0     80   134  
02:29:36.243 CD2B.1    stand  1    -70    230    0     80   134  
02:29:36.243 CD2B.0    stand  255  -170   320    0     80   134  
02:29:36.243 CD2B.2    lying  None -10    160    0     80   226  
02:29:37.232 CD2B.1    stand  1    -70    230    0     80   92   
02:29:37.232 CD2B.2    lying  None -10    160    49    80   92   
02:29:37.232 CD2B.0    stand  255  -170   320    0     80   226  
02:29:38.235 CD2B.0    stand  255  -170   320    0     80   0    
02:29:38.235 CD2B.2    lying  None -10    160    42    80   226  
02:29:38.235 CD2B.1    stand  1    -70    230    0     80   92   
02:29:39.235 CD2B.0    stand  255  -170   320    0     80   134  
02:29:39.235 CD2B.2    lying  None -60    160    48    80   194  
02:29:39.235 CD2B.1    stand  1    -70    230    0     80   70   
02:29:40.241 CD2B.1    stand  1    -70    230    0     80   0    
02:29:40.241 CD2B.0    stand  255  -170   320    0     80   134  
02:29:40.241 CD2B.2    lying  None -80    170    48    80   174  
02:29:41.238 CD2B.0    stand  255  -170   320    0     80   174  
02:29:41.238 CD2B.1    stand  1    -70    230    0     80   134  
02:29:41.238 CD2B.2    lying  None -80    170    49    80   60   
02:29:42.244 CD2B.1    stand  1    -70    230    0     80   60   
02:29:42.244 CD2B.2    lying  None -70    160    55    80   70   
02:29:42.244 CD2B.0    stand  255  -170   320    0     80   188  
02:29:43.244 CD2B.0    stand  255  -170   320    0     80   0    
02:29:43.244 CD2B.1    stand  1    -70    230    0     80   134  
02:29:43.244 CD2B.2    lying  None -70    160    59    80   70   
02:29:44.240 CD2B.2    lying  None -70    160    62    80   0    
02:29:44.240 CD2B.1    stand  1    -70    230    0     80   70   
02:29:44.240 CD2B.0    stand  255  -170   320    0     80   134  
02:29:45.245 CD2B.2    lying  None -70    160    52    80   188  
02:29:45.245 CD2B.1    stand  1    -70    230    0     80   70   
02:29:45.245 CD2B.0    stand  255  -170   320    0     80   134  
02:29:46.244 CD2B.1    stand  1    -70    230    0     80   134  
02:29:46.244 CD2B.0    stand  255  -170   320    0     80   134  
02:29:46.244 CD2B.2    lying  None -70    160    56    80   188  
02:29:47.140 CD2B.0    stand  255  -170   320    0     80   188  
02:29:47.140 CD2B.1    stand  1    -70    230    0     80   134  
02:29:47.140 CD2B.2    lying  None -70    170    54    80   60   
02:29:48.141 CD2B.2    lying  None -100   180    54    80   31   
02:29:48.141 CD2B.0    stand  255  -170   320    0     80   156  
02:29:48.141 CD2B.1    stand  1    -70    230    0     80   134  
02:29:49.147 CD2B.1    stand  1    -70    230    0     80   0    
02:29:49.147 CD2B.2    lying  None -30    150    0     80   89   
02:29:49.147 CD2B.0    stand  255  -170   320    0     80   220  
02:29:50.143 CD2B.0    stand  255  -170   320    0     80   0    
02:29:50.143 CD2B.2    lying  None -10    160    0     80   226  
02:29:50.143 CD2B.1    stand  1    -70    230    0     80   92   
02:29:51.104 CD2B.0    stand  255  -170   320    0     80   134  
02:29:51.104 CD2B.1    stand  1    -70    230    0     80   134  
02:29:51.104 CD2B.2    stand  None -40    130    0     80   104  
02:29:52.105 CD2B.2    stand  None -40    160    0     80   30   
02:29:52.105 CD2B.1    stand  1    -70    230    0     80   76   
02:29:52.105 CD2B.0    stand  255  -170   320    0     80   134  
02:29:53.107 CD2B.0    stand  255  -170   320    0     80   0    
02:29:53.107 CD2B.2    stand  None -40    150    0     80   214  
02:29:53.107 CD2B.1    stand  1    -70    230    0     80   85   
02:29:54.113 CD2B.0    stand  255  -170   320    0     80   134  
02:29:54.113 CD2B.2    stand  None -40    150    0     80   214  
02:29:54.113 CD2B.1    stand  1    -70    230    0     80   85   
02:29:55.118 CD2B.2    stand  None -40    150    0     80   85   
02:29:55.118 CD2B.1    stand  1    -70    230    0     80   85   
02:29:55.118 CD2B.0    stand  255  -170   320    0     80   134  
02:29:56.109 CD2B.2    stand  None -20    150    0     80   226  
02:29:56.109 CD2B.1    stand  1    -70    230    0     80   94   
02:29:56.109 CD2B.0    stand  255  -170   320    0     80   134  
02:29:57.111 CD2B.1    stand  1    -70    230    0     80   134  
02:29:57.111 CD2B.0    stand  255  -170   320    0     80   134  
02:29:57.111 CD2B.2    stand  None -20    160    0     80   219  
02:29:58.112 CD2B.0    stand  255  -170   320    0     80   219  
02:29:58.112 CD2B.1    stand  1    -70    230    0     80   134  
02:29:58.112 CD2B.2    stand  None -10    140    0     80   108  
02:29:59.117 CD2B.2    stand  None -10    140    0     80   0    
02:29:59.117 CD2B.0    stand  255  -170   320    0     80   240  
02:29:59.117 CD2B.1    stand  1    -70    230    0     80   134  
02:30:00.124 CD2B.2    stand  None -10    170    0     80   84   
02:30:00.124 CD2B.1    stand  1    -70    230    0     80   84   
02:30:00.124 CD2B.0    stand  255  -170   320    0     80   134  
02:30:01.115 CD2B.0    stand  255  -170   320    0     80   0    
02:30:01.115 CD2B.2    stand  None 0      170    0     80   226  
02:30:01.115 CD2B.1    stand  1    -70    230    0     80   92   
02:30:02.167 CD2B.0    stand  255  -170   320    0     80   134  
02:30:02.167 CD2B.2    stand  None 0      170    0     80   226  
02:30:02.167 CD2B.1    stand  1    -70    230    0     80   92   
02:30:03.006 CD2B.1    stand  1    -70    230    0     80   0    
02:30:03.006 CD2B.0    stand  255  -170   320    0     80   134  
02:30:03.006 CD2B.2    stand  None -20    170    56    80   212  
02:30:04.009 CD2B.2    stand  None -10    160    0     80   14   
02:30:04.009 CD2B.1    stand  1    -70    230    0     80   92   
02:30:04.009 CD2B.0    stand  255  -170   320    0     80   134  
02:30:05.010 CD2B.0    stand  255  -170   320    0     80   0    
02:30:05.010 CD2B.1    stand  1    -90    220    0     80   128  
02:30:05.010 CD2B.2    stand  None -10    160    0     80   100  
02:30:06.016 CD2B.2    stand  None -10    160    0     80   0    
02:30:06.016 CD2B.0    stand  255  -170   320    0     80   226  
02:30:06.016 CD2B.1    stand  1    -150   160    0     80   161  
02:30:07.020 CD2B.1    stand  1    -150   160    0     80   0    
02:30:07.020 CD2B.2    stand  None -10    160    0     80   140  
02:30:07.020 CD2B.0    stand  255  -170   320    0     80   226  
02:30:08.048 CD2B.0    stand  255  -170   320    0     80   0    
02:30:08.048 CD2B.2    stand  None -20    180    66    80   205  
02:30:08.048 CD2B.1    stand  1    -150   160    0     80   131  
02:30:09.048 CD2B.1    stand  1    -140   170    0     80   14   
02:30:09.048 CD2B.2    stand  None -40    180    0     80   100  
02:30:09.048 CD2B.0    stand  255  -170   320    0     80   191  
02:30:10.045 CD2B.2    stand  None -60    180    0     80   178  
02:30:10.045 CD2B.0    stand  255  -170   320    0     80   178  
02:30:10.045 CD2B.1    stand  1    -140   170    0     80   152  
02:30:10.942 CD2B.1    stand  1    -140   170    0     80   0    
02:30:10.942 CD2B.0    stand  255  -170   320    0     80   152  
02:30:10.942 CD2B.2    stand  None -40    180    0     80   191  
02:30:11.943 CD2B.1    stand  1    -140   170    0     80   100  
02:30:11.943 CD2B.2    stand  None -30    170    0     80   110  
02:30:11.943 CD2B.0    stand  255  -170   320    0     80   205  
02:30:12.945 CD2B.2    stand  None -30    170    0     80   205  
02:30:12.945 CD2B.1    stand  1    -140   170    0     80   110  
02:30:12.945 CD2B.0    stand  255  -170   320    0     80   152  
02:30:13.949 CD2B.1    stand  1    -140   170    0     80   152  
02:30:13.949 CD2B.0    stand  255  -170   320    0     80   152  
02:30:13.949 CD2B.2    stand  None -40    170    0     80   198  
02:30:14.953 CD2B.0    stand  255  -170   320    0     80   198  
02:30:14.953 CD2B.1    stand  1    -140   170    0     80   152  
02:30:14.953 CD2B.2    stand  None -40    170    0     80   100  
02:30:15.948 CD2B.0    stand  255  -170   320    0     80   198  
02:30:15.948 CD2B.2    stand  None -40    180    61    80   191  
02:30:15.948 CD2B.1    stand  1    -140   170    0     80   100  
02:30:16.948 CD2B.2    lying  None -70    170    51    80   70   
02:30:16.948 CD2B.0    stand  255  -170   320    0     80   180  
02:30:16.948 CD2B.1    stand  1    -140   170    0     80   152  
02:30:17.949 CD2B.2    lying  None -90    180    48    80   50   
02:30:17.949 CD2B.1    stand  1    -140   170    0     80   50   
02:30:17.949 CD2B.0    stand  255  -170   320    0     80   152  
02:30:18.950 CD2B.0    stand  255  -170   320    0     80   0    
02:30:18.950 CD2B.1    stand  1    -140   170    0     80   152  
02:30:18.950 CD2B.2    lying  None -100   180    0     80   41   
02:30:19.952 CD2B.1    stand  1    -140   170    0     80   41   
02:30:19.952 CD2B.0    stand  255  -170   320    0     80   152  
02:30:19.952 CD2B.2    lying  None -30    150    0     80   220  
02:30:20.956 CD2B.1    stand  1    -140   170    0     80   111  
02:30:20.956 CD2B.2    lying  None -80    180    52    80   60   
02:30:20.956 CD2B.0    stand  255  -170   320    0     80   166  
02:30:21.960 CD2B.2    lying  None -110   170    46    80   161  
02:30:21.960 CD2B.1    stand  1    -140   170    0     80   30   
02:30:21.960 CD2B.0    stand  255  -170   320    0     80   152  
02:30:22.849 CD2B.1    stand  1    -140   170    0     80   152  
02:30:22.849 CD2B.0    stand  255  -170   320    0     80   152  
02:30:22.849 CD2B.2    lying  None -100   170    54    80   165  
02:30:23.864 CD2B.2    lying  None -110   180    55    80   14   
02:30:23.864 CD2B.1    stand  1    -140   170    0     80   31   
02:30:23.864 CD2B.0    stand  255  -170   320    0     80   152  
02:30:24.860 CD2B.1    stand  1    -140   170    0     80   152  
02:30:24.860 CD2B.0    stand  255  -170   320    0     80   152  
02:30:24.860 CD2B.2    lying  None -100   180    55    80   156  
02:30:25.858 CD2B.2    lying  None -90    190    0     80   14   
02:30:25.858 CD2B.1    stand  1    -140   170    0     80   53   
02:30:25.858 CD2B.0    stand  255  -170   320    0     80   152  
02:30:26.860 CD2B.0    stand  255  -170   320    0     80   0    
02:30:26.860 CD2B.2    lying  None -100   180    0     80   156  
02:30:26.860 CD2B.1    stand  1    -150   170    0     80   50   
02:30:27.860 CD2B.0    stand  255  -170   320    0     80   151  
02:30:27.860 CD2B.1    stand  1    -150   170    0     80   151  
02:30:27.860 CD2B.2    lying  None -100   180    0     80   50   
02:30:28.652 CD2B.1    stand  1    -150   170    0     80   50   
02:30:28.652 CD2B.2    lying  None -100   180    0     80   50   
02:30:28.652 CD2B.3    stand  None -180   0      99    80   196  
02:30:28.652 CD2B.0    stand  255  -170   320    0     80   320  
02:30:28.808 CD2B.1    stand  1    -150   170    0     80   151  
02:30:28.808 CD2B.3    stand  None -180   10     99    80   162  
02:30:28.808 CD2B.2    lying  None -100   180    0     80   187  
02:30:28.808 CD2B.0    stand  255  -170   320    0     80   156  
02:30:29.802 CD2B.3    stand  None -200   20     102   80   301  
02:30:29.802 CD2B.0    stand  255  -170   320    0     80   301  
02:30:29.802 CD2B.1    stand  1    -150   170    0     80   151  
02:30:29.802 CD2B.2    lying  None -100   180    0     80   50   
02:30:30.808 CD2B.1    stand  1    -150   170    0     80   50   
02:30:30.808 CD2B.0    stand  255  -170   320    0     80   151  
02:30:30.808 CD2B.3    stand  None -180   40     0     80   280  
02:30:30.808 CD2B.2    lying  None -100   180    0     80   161  
02:30:31.814 CD2B.0    stand  255  -170   320    0     80   156  
02:30:31.814 CD2B.1    stand  1    -150   170    0     80   151  
02:30:31.814 CD2B.2    lying  None -100   180    0     80   50   
02:30:31.814 CD2B.3    stand  None -190   30     101   80   174  
02:30:32.806 CD2B.3    stand  None -170   20     84    80   22   
02:30:32.806 CD2B.2    lying  None -100   180    0     80   174  
02:30:32.806 CD2B.1    stand  1    -150   170    0     80   50   
02:30:32.806 CD2B.0    stand  255  -170   320    0     80   151  
02:30:33.807 CD2B.0    stand  255  -170   320    0     80   0    
02:30:33.807 CD2B.1    stand  1    -150   170    0     80   151  
02:30:33.807 CD2B.2    lying  None -100   180    0     80   50   
02:30:33.807 CD2B.3    stand  None -170   0      68    80   193  
02:30:34.810 CD2B.1    stand  1    -150   170    0     80   171  
02:30:34.810 CD2B.2    lying  None -100   180    0     80   50   
02:30:34.810 CD2B.3    stand  None -180   -10    100   80   206  
02:30:34.810 CD2B.0    stand  255  -170   320    0     80   330  
02:30:35.812 CD2B.1    stand  1    -150   170    0     80   151  
02:30:35.812 CD2B.0    stand  255  -170   320    0     80   151  
02:30:35.812 CD2B.3    stand  None -180   20     67    80   300  
02:30:35.812 CD2B.2    lying  None -100   180    0     80   178  
02:30:36.812 CD2B.2    lying  None -100   180    0     80   0    
02:30:36.812 CD2B.0    stand  255  -170   320    0     80   156  
02:30:36.812 CD2B.3    walk   None -220   80     74    80   245  
02:30:36.812 CD2B.1    stand  1    -150   170    0     80   114  
02:30:37.810 CD2B.2    lying  None -100   180    0     80   50   
02:30:37.810 CD2B.0    stand  255  -170   320    0     80   156  
02:30:37.810 CD2B.3    walk   None -260   140    75    80   201  
02:30:37.810 CD2B.1    stand  1    -150   170    0     80   114  
02:30:38.826 CD2B.2    lying  None -100   180    0     80   50   
02:30:38.826 CD2B.0    stand  255  -170   320    0     80   156  
02:30:38.826 CD2B.3    walk   None -270   200    76    80   156  
02:30:38.826 CD2B.1    stand  1    -150   170    0     80   123  
02:30:39.719 CD2B.2    lying  None -100   180    0     80   50   
02:30:39.719 CD2B.0    stand  255  -170   320    0     80   156  
02:30:39.719 CD2B.3    walk   None -290   220    61    80   156  
02:30:39.719 CD2B.1    stand  1    -150   170    0     80   148  
02:30:40.714 CD2B.1    stand  1    -150   170    0     80   0    
02:30:40.714 CD2B.0    stand  255  -170   320    0     80   151  
02:30:40.714 CD2B.3    walk   None -300   260    83    80   143  
02:30:40.714 CD2B.2    lying  None -100   180    0     80   215  
02:30:41.776 CD2B.3    walk   None -270   280    75    80   197  
02:30:41.776 CD2B.2    lying  None -100   180    0     80   197  
02:30:42.726 CD2B.3    walk   None -250   290    0     80   186  
02:30:42.726 CD2B.2    lying  None -100   180    0     80   186  
02:30:43.787 CD2B.3    walk   None -230   290    59    80   170  
02:30:44.738 CD2B.3    walk   None -190   310    73    80   44   
02:30:45.738 CD2B.3    walk   None -170   320    81    80   22   
02:30:46.742 CD2B.3    walk   None -170   310    86    80   10   
02:30:47.740 CD2B.3    walk   None -140   340    81    80   42   
02:30:48.640 CD2B.3    walk   None -120   340    72    80   20   
02:30:49.640 CD2B.3    walk   None -110   330    78    80   14   
02:30:50.646 CD2B.3    walk   None -140   330    91    80   30   
02:30:51.656 CD2B.3    walk   None -130   320    0     80   14   
02:30:52.642 CD2B.3    walk   None -160   320    0     80   30   
02:30:53.644 CD2B.3    walk   None -150   320    99    80   10   
02:30:54.646 CD2B.3    walk   None -140   320    101   80   10   
02:30:55.678 CD2B.3    walk   None -150   330    93    80   14   
02:30:56.576 CD2B.3    stand  None -140   330    0     80   10   
02:30:57.580 CD2B.3    stand  None -130   330    90    80   10   
02:30:58.588 CD2B.3    stand  None -150   340    111   80   22   
02:30:59.582 CD2B.3    stand  None -160   340    84    80   10   
02:31:00.578 CD2B.3    stand  None -160   340    61    80   0    
02:31:01.632 CD2B.3    stand  None -140   330    80    80   22   
02:31:02.580 CD2B.3    stand  None -160   330    0     80   20   
02:31:03.581 CD2B.3    stand  None -150   330    74    80   10   
02:31:04.583 CD2B.3    stand  None -150   310    79    80   20   
02:31:05.589 CD2B.3    stand  None -160   320    60    80   14   
02:31:06.586 CD2B.3    stand  None -180   320    79    80   20   
02:31:07.586 CD2B.3    stand  None -210   310    81    80   31   
02:31:08.480 CD2B.3    stand  None -230   280    85    80   36   
02:31:09.482 CD2B.3    stand  None -240   270    77    80   14   
02:31:10.481 CD2B.3    stand  None -220   260    0     80   22   
02:31:11.491 CD2B.3    stand  None -240   250    63    80   22   
02:31:12.495 CD2B.3    stand  None -240   240    84    80   10   
02:31:13.493 CD2B.3    stand  None -240   210    99    80   30   
02:31:14.494 CD2B.3    stand  None -240   170    98    80   40   
02:31:15.496 CD2B.3    stand  None -220   170    87    80   20   
02:31:16.496 CD2B.3    stand  None -230   160    96    80   14   
02:31:17.496 CD2B.3    stand  None -230   120    88    80   40   
02:31:18.496 CD2B.3    stand  None -210   90     66    80   36   
02:31:19.393 CD2B.3    walk   None -190   60     65    80   36   
02:31:20.400 CD2B.3    walk   None -190   30     73    80   30   
02:31:21.393 CD2B.3    walk   None -160   10     64    80   36   
02:31:22.396 CD2B.3    walk   None -140   0      91    80   22   
02:31:23.396 CD2B.3    walk   None -130   -10    0     80   14   
02:31:24.399 CD2B.3    sit    None -130   -10    0     80   0    
02:31:25.400 CD2B.3    sit    None -140   -10    0     80   10   
02:31:26.399 CD2B.3    sit    None -140   -10    0     80   0    
02:31:27.401 CD2B.3    sit    None -140   -10    0     80   0    
02:31:28.484 CD2B.88   88     -    -      -      -     -    -    
02:31:29.308 CD2B.88   88     -    -      -      -     -    -    
02:31:30.311 CD2B.88   88     -    -      -      -     -    -    
02:31:34.364 CD2B.0    stand  None -190   20     70    80   58   
02:31:35.320 CD2B.0    walk   None -240   120    67    80   111  
02:31:36.325 CD2B.0    walk   None -240   200    82    80   80   
02:31:37.321 CD2B.0    walk   None -230   260    88    80   60   
02:31:38.328 CD2B.0    walk   None -190   290    91    80   50   
02:31:39.330 CD2B.0    walk   None -110   280    99    80   80   
02:31:40.326 CD2B.0    walk   None -60    220    64    80   78   
02:31:41.216 CD2B.0    lying  None -80    220    53    80   20   
02:31:42.225 CD2B.0    lying  None -70    180    53    80   41   
02:31:43.200 CD2B.0    walk   None -30    130    0     80   64   
02:31:44.204 CD2B.0    sit    None 0      160    0     80   42   
02:31:45.202 CD2B.0    sit    None -10    160    0     80   10   
02:31:46.205 CD2B.0    sit    None -10    160    0     80   0    
02:31:47.204 CD2B.0    sit    None -10    160    0     80   0    
02:31:48.204 CD2B.0    sit    None -10    160    0     80   0    
02:31:49.205 CD2B.0    sit    None -10    160    0     80   0    
02:31:50.208 CD2B.0    sit    None -10    160    0     80   0    
02:31:51.210 CD2B.0    sit    None -10    160    0     80   0    
02:31:52.212 CD2B.0    sit    None -10    160    0     80   0    
02:31:53.209 CD2B.0    sit    None -10    160    0     80   0    
02:31:54.112 CD2B.0    sit    None -10    160    0     80   0    
02:31:55.109 CD2B.0    sit    None -10    160    0     80   0    
02:31:56.112 CD2B.0    sit    None -10    160    0     80   0    
02:31:57.112 CD2B.0    sit    None -10    160    0     80   0    
02:31:58.122 CD2B.0    sit    None -10    160    0     80   0    
02:31:59.118 CD2B.0    stand  None -20    170    0     80   14   
02:32:00.128 CD2B.0    stand  None -50    230    0     80   67   
02:32:01.178 CD2B.0    stand  None -50    230    0     80   0    
02:32:02.129 CD2B.0    stand  1    -50    230    0     80   0    
02:32:03.130 CD2B.0    stand  1    -50    230    0     80   0    
02:32:04.030 CD2B.0    stand  1    -50    230    0     80   0    
02:32:05.031 CD2B.0    stand  1    -50    230    0     80   0    
02:32:06.030 CD2B.0    stand  1    -40    240    0     80   14   
02:32:07.032 CD2B.0    stand  1    -30    240    0     80   10   
02:32:08.032 CD2B.0    stand  1    -30    240    0     80   0    
02:32:09.034 CD2B.0    stand  1    -30    240    0     80   0    
02:32:10.036 CD2B.0    stand  1    -30    240    0     80   0    
02:32:11.036 CD2B.0    stand  1    -30    240    0     80   0    
02:32:12.038 CD2B.0    stand  1    -30    240    0     80   0    
02:32:13.039 CD2B.0    stand  1    -30    240    0     80   0    
02:32:14.040 CD2B.0    stand  1    -30    240    0     80   0    
02:32:15.044 CD2B.0    stand  1    -30    240    0     80   0    
02:32:15.937 CD2B.0    stand  1    -30    240    0     80   0    
02:32:16.936 CD2B.0    walk   1    -90    200    49    80   72   
02:32:17.941 CD2B.0    walk   1    -20    150    0     80   86   
02:32:18.944 CD2B.0    walk   1    -20    160    44    80   10   
02:32:19.941 CD2B.0    walk   1    -20    150    0     80   10   
02:32:20.948 CD2B.0    walk   1    0      160    0     80   22   
02:32:21.948 CD2B.0    stand  1    0      160    0     80   0    
02:32:22.944 CD2B.0    stand  1    0      160    0     80   0    
02:32:23.946 CD2B.0    stand  1    0      170    0     80   10   
02:32:24.948 CD2B.0    stand  1    0      170    0     80   0    
02:32:25.950 CD2B.0    stand  1    0      170    0     80   0    
02:32:26.845 CD2B.0    stand  1    0      170    0     80   0    
02:32:27.846 CD2B.0    stand  1    0      170    0     80   0    
02:32:28.848 CD2B.0    stand  1    0      170    0     80   0    
02:32:29.849 CD2B.0    stand  1    0      170    0     80   0    
02:32:30.852 CD2B.0    stand  1    0      170    0     80   0    
02:32:31.806 CD2B.0    stand  1    0      170    0     80   0    
02:32:32.808 CD2B.0    stand  1    0      170    0     80   0    
02:32:33.807 CD2B.0    stand  1    0      170    0     80   0    
02:32:34.808 CD2B.0    stand  1    0      170    0     80   0    
02:32:35.809 CD2B.0    stand  1    0      170    0     80   0    
02:32:36.813 CD2B.0    stand  1    0      170    0     80   0    
02:32:37.816 CD2B.0    stand  1    -30    150    0     80   36   
02:32:38.813 CD2B.0    stand  1    -30    180    0     80   30   
02:32:39.817 CD2B.0    stand  1    -10    160    0     80   28   
02:32:40.818 CD2B.0    stand  1    -10    160    0     80   0    
02:32:41.816 CD2B.0    stand  1    0      160    0     80   10   
02:32:42.820 CD2B.0    stand  1    0      160    0     80   0    
02:32:43.708 CD2B.0    stand  1    0      160    0     80   0    
02:32:44.709 CD2B.0    stand  1    0      160    0     80   0    
02:32:45.716 CD2B.0    stand  1    0      160    0     80   0    
02:32:46.713 CD2B.0    stand  1    -10    160    65    80   10   
02:32:47.764 CD2B.0    stand  1    -10    160    0     80   0    
02:32:48.760 CD2B.0    stand  1    -10    160    0     80   0    
02:32:49.765 CD2B.0    stand  1    -10    160    0     80   0    
02:32:50.657 CD2B.0    stand  1    -10    160    0     80   0    
02:32:51.657 CD2B.0    stand  1    -10    160    0     80   0    
02:32:52.657 CD2B.0    stand  1    -10    160    0     80   0    
02:32:53.660 CD2B.0    stand  1    -10    160    0     80   0    
02:32:54.656 CD2B.0    stand  1    -10    160    0     80   0    
02:32:55.660 CD2B.0    stand  1    -10    160    0     80   0    
02:32:56.660 CD2B.0    lying  1    -50    170    45    80   41   
02:32:57.664 CD2B.0    lying  1    -110   180    44    80   60   
02:32:58.665 CD2B.0    lying  1    -120   190    0     80   14   

```

**汇总**: xray tick 635 | fire 1 | Fall 事件 0 () | 结论 = fire 命中

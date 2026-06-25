# case-b17f-0624-20492101 — 每 tick belief 时间线 (room fd00:0:3:111:2:300, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
20:49:00 B17F.2   B17F24900917  sit     0    NoReport sit                trk  0.50 Empty      3   0     0.00  0.03  0.08  0.00  0.85  0.04
20:49:00 B17F.0   B17F04900917  stand   0    NoReport stand              trk  0.50 Empty      3   0     0.00  0.02  0.26  0.00  0.69  0.03
20:49:00 B17F.1   B17F24900917  walk    97   NoReport walk               trk  0.50 Empty      3   0     0.00  0.03  0.08  0.00  0.85  0.04
20:49:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  0.51 OpenFloor  3   1     0.00  0.02  0.68  0.00  0.27  0.01
20:49:01 B17F.1   B17F24900917  walk    74   NoReport walk               trk  0.51 OpenFloor  3   1     0.00  0.03  0.09  0.00  0.74  0.01
20:49:01 B17F.2   B17F24900917  sit     0    NoReport sit                trk  0.51 OpenFloor  3   1     0.00  0.03  0.09  0.00  0.74  0.01
20:49:02 B17F.2   B17F24900917  sit     0    NoReport sit                trk  0.52 OpenFloor  3   2     0.00  0.02  0.07  0.01  0.42  0.01
20:49:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  0.52 OpenFloor  3   2     0.00  0.01  0.88  0.00  0.06  0.01
20:49:02 B17F.1   B17F24900917  walk    77   NoReport walk               trk  0.52 OpenFloor  3   2     0.00  0.02  0.07  0.01  0.42  0.01
20:49:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  0.53 OpenFloor  3   3     0.00  0.01  0.92  0.00  0.01  0.01
20:49:03 B17F.2   B17F24900917  sit     0    NoReport sit                trk  0.53 OpenFloor  3   3     0.00  0.01  0.03  0.01  0.11  0.01
20:49:03 B17F.1   B17F24900917  walk    72   NoReport walk               trk  0.53 OpenFloor  3   3     0.00  0.01  0.03  0.01  0.11  0.01
20:49:04 B17F.2   B17F24900917  sit     0    NoReport sit                trk  0.54 OpenFloor  3   3     0.00  0.00  0.02  0.01  0.02  0.00
20:49:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  0.54 OpenFloor  3   3     0.00  0.01  0.93  0.00  0.00  0.01
20:49:04 B17F.1   B17F24900917  walk    70   NoReport walk               trk  0.54 OpenFloor  3   3     0.00  0.00  0.02  0.01  0.02  0.00
20:49:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  0.55 OpenFloor  3   4     0.00  0.01  0.93  0.00  0.00  0.01
20:49:05 B17F.2   B17F24900917  sit     0    NoReport sit                trk  0.55 OpenFloor  3   4     0.00  0.00  0.02  0.01  0.00  0.00
20:49:05 B17F.1   B17F24900917  stand   0    NoReport stand              trk  0.55 OpenFloor  3   4     0.00  0.00  0.02  0.01  0.00  0.00
20:49:06 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   5     0.00  0.01  0.93  0.00  0.00  0.01
20:49:06 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   5     0.00  0.00  0.02  0.01  0.00  0.00
20:49:06 B17F.1   B17F24900917  stand   78   NoReport stand              trk  1.00 OpenFloor  3   5     0.00  0.00  0.02  0.01  0.00  0.00
20:49:07 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   6     0.00  0.00  0.02  0.01  0.00  0.00
20:49:07 B17F.1   B17F24900917  stand   71   NoReport stand              trk  1.00 OpenFloor  3   6     0.00  0.00  0.02  0.01  0.00  0.00
20:49:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   6     0.00  0.01  0.93  0.00  0.00  0.01
20:49:08 B17F.1   B17F24900917  stand   76   NoReport stand              trk  1.00 OpenFloor  3   7     0.00  0.00  0.02  0.01  0.00  0.00
20:49:08 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   7     0.00  0.00  0.02  0.01  0.00  0.00
20:49:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   7     0.00  0.01  0.93  0.00  0.00  0.01
20:49:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   8     0.00  0.01  0.93  0.00  0.00  0.01
20:49:09 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   8     0.00  0.00  0.02  0.01  0.00  0.00
20:49:09 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   8     0.00  0.00  0.02  0.01  0.00  0.00
20:49:10 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.01  0.93  0.00  0.00  0.01
20:49:10 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   9     0.00  0.00  0.02  0.01  0.00  0.00
20:49:10 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.00  0.02  0.01  0.00  0.00
20:49:11 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   10    0.00  0.01  0.93  0.00  0.00  0.01
20:49:11 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   10    0.00  0.00  0.02  0.01  0.00  0.00
20:49:11 B17F.1   B17F24900917  stand   99   NoReport stand              trk  1.00 OpenFloor  3   10    0.00  0.00  0.02  0.01  0.00  0.00
20:49:12 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   11    0.00  0.00  0.02  0.01  0.00  0.00
20:49:12 B17F.1   B17F24900917  stand   73   NoReport stand              trk  1.00 OpenFloor  3   11    0.00  0.00  0.02  0.01  0.00  0.00
20:49:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   11    0.00  0.01  0.93  0.00  0.00  0.01
20:49:13 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   12    0.00  0.00  0.02  0.01  0.00  0.00
20:49:13 B17F.1   B17F24900917  stand   85   NoReport stand              trk  1.00 OpenFloor  3   12    0.00  0.00  0.02  0.01  0.00  0.00
20:49:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   12    0.00  0.01  0.93  0.00  0.00  0.01
20:49:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   13    0.00  0.01  0.93  0.00  0.00  0.01
20:49:14 B17F.1   B17F24900917  stand   86   NoReport stand              trk  1.00 OpenFloor  3   13    0.00  0.00  0.02  0.01  0.00  0.00
20:49:14 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   13    0.00  0.00  0.02  0.01  0.00  0.00
20:49:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.01  0.93  0.00  0.00  0.01
20:49:15 B17F.1   B17F24900917  stand   81   NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.00  0.02  0.01  0.00  0.00
20:49:15 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   14    0.00  0.00  0.02  0.01  0.00  0.00
20:49:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.01  0.93  0.00  0.00  0.01
20:49:16 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   15    0.00  0.00  0.02  0.01  0.00  0.00
20:49:16 B17F.1   B17F24900917  stand   63   NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.00  0.02  0.01  0.00  0.00
20:49:17 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   16    0.00  0.00  0.02  0.01  0.00  0.00
20:49:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.01  0.93  0.00  0.00  0.01
20:49:17 B17F.1   B17F24900917  stand   90   NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.00  0.02  0.01  0.00  0.00
20:49:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.01  0.93  0.00  0.00  0.01
20:49:18 B17F.1   B17F24900917  stand   90   NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.00  0.02  0.01  0.00  0.00
20:49:18 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   17    0.00  0.00  0.02  0.01  0.00  0.00
20:49:19 B17F.1   B17F24900917  stand   78   NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.00  0.02  0.01  0.00  0.00
20:49:19 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   18    0.00  0.00  0.02  0.01  0.00  0.00
20:49:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.01  0.93  0.00  0.00  0.01
20:49:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.93  0.00  0.00  0.01
20:49:20 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   19    0.00  0.00  0.02  0.01  0.00  0.00
20:49:20 B17F.1   B17F24900917  stand   84   NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.00  0.02  0.01  0.00  0.00
20:49:21 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   20    0.00  0.00  0.02  0.01  0.00  0.00
20:49:21 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.01  0.93  0.00  0.00  0.01
20:49:21 B17F.1   B17F24900917  stand   82   NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.00  0.02  0.01  0.00  0.00
20:49:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.01  0.93  0.00  0.00  0.01
20:49:22 B17F.1   B17F24900917  stand   63   NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.00  0.02  0.01  0.00  0.00
20:49:22 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   21    0.00  0.00  0.02  0.01  0.00  0.00
20:49:23 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   22    0.00  0.00  0.02  0.01  0.00  0.00
20:49:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.01  0.93  0.00  0.00  0.01
20:49:23 B17F.1   B17F24900917  stand   85   NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.00  0.02  0.01  0.00  0.00
20:49:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.01  0.93  0.00  0.00  0.01
20:49:24 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   23    0.00  0.00  0.02  0.01  0.00  0.00
20:49:24 B17F.1   B17F24900917  stand   72   NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.00  0.02  0.01  0.00  0.00
20:49:25 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 Sit        3   24    0.00  0.00  0.02  0.01  0.00  0.00
20:49:25 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Sit        3   24    0.00  0.01  0.93  0.00  0.00  0.01
20:49:25 B17F.1   B17F24900917  stand   89   NoReport stand              trk  1.00 Sit        3   24    0.00  0.00  0.02  0.01  0.00  0.00
20:49:26 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Sit        3   25    0.00  0.01  0.93  0.00  0.00  0.01
20:49:26 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 Sit        3   25    0.00  0.00  0.02  0.01  0.00  0.00
20:49:26 B17F.1   B17F24900917  walk    64   NoReport walk               trk  1.00 Sit        3   25    0.00  0.00  0.02  0.01  0.00  0.00
20:49:27 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Sit        3   26    0.00  0.01  0.93  0.00  0.00  0.01
20:49:27 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 Sit        3   26    0.00  0.00  0.02  0.01  0.00  0.00
20:49:27 B17F.1   B17F24900917  walk    72   NoReport walk               trk  1.00 Sit        3   26    0.00  0.00  0.02  0.01  0.00  0.00
20:49:28 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 Sit        3   27    0.00  0.00  0.02  0.01  0.00  0.00
20:49:28 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 Sit        3   27    0.00  0.00  0.02  0.01  0.00  0.00
20:49:28 B17F.0   B17F04900917  walk    82   NoReport walk               trk  1.00 Sit        3   27    0.00  0.01  0.93  0.00  0.00  0.01
20:49:29 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 Sit        3   28    0.00  0.00  0.02  0.01  0.00  0.00
20:49:29 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 Sit        3   28    0.00  0.00  0.02  0.01  0.00  0.00
20:49:29 B17F.0   B17F04900917  walk    115  NoReport walk               trk  1.00 Sit        3   28    0.00  0.01  0.93  0.00  0.00  0.01
20:49:30 B17F.0   B17F04900917  walk    122  NoReport walk               trk  1.00 Sit        3   29    0.00  0.01  0.97  0.00  0.00  0.01
20:49:30 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 Sit        3   29    0.00  0.00  0.02  0.01  0.00  0.00
20:49:30 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Sit        3   29    0.00  0.00  0.02  0.01  0.00  0.00
20:49:31 B17F.0   B17F04900917  walk    118  NoReport walk               trk  1.00 OpenFloor  3   30    0.00  0.00  0.98  0.00  0.00  0.00
20:49:31 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   30    0.00  0.00  0.02  0.01  0.00  0.00
20:49:31 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   30    0.00  0.00  0.02  0.01  0.00  0.00
20:49:32 B17F.0   B17F04900917  walk    116  NoReport walk               trk  1.00 OpenFloor  3   31    0.00  0.00  0.99  0.00  0.00  0.00
20:49:32 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   31    0.00  0.00  0.02  0.01  0.00  0.00
20:49:32 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   31    0.00  0.00  0.02  0.01  0.00  0.00
20:49:33 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   32    0.00  0.00  0.02  0.01  0.00  0.00
20:49:33 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   32    0.00  0.00  0.02  0.01  0.00  0.00
20:49:33 B17F.0   B17F04900917  walk    113  NoReport walk               trk  1.00 OpenFloor  3   32    0.00  0.01  0.94  0.00  0.00  0.01
20:49:34 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 OpenFloor  3   33    0.00  0.01  0.93  0.00  0.00  0.01
20:49:34 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   33    0.00  0.00  0.02  0.01  0.00  0.00
20:49:34 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   33    0.00  0.00  0.02  0.01  0.00  0.00
20:49:35 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.00  0.02  0.01  0.00  0.01
20:49:35 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   34    0.00  0.00  0.02  0.01  0.00  0.01
20:49:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.01  0.90  0.00  0.01  0.01
20:49:36 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   0     0.00  0.00  0.02  0.01  0.00  0.00
20:49:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.01  0.96  0.00  0.00  0.01
20:49:36 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   0     0.00  0.00  0.02  0.01  0.00  0.00
20:49:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.01  0.94  0.00  0.00  0.01
20:49:37 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   9     0.00  0.00  0.02  0.01  0.00  0.00
20:49:37 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.00  0.02  0.01  0.00  0.00
20:49:38 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   10    0.00  0.00  0.02  0.01  0.00  0.00
20:49:38 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   10    0.00  0.00  0.02  0.01  0.00  0.00
20:49:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   10    0.00  0.01  0.93  0.00  0.00  0.01
20:49:39 B17F.0   B17F04900917  stand   122  NoReport stand              trk  1.00 OpenFloor  3   11    0.00  0.01  0.97  0.00  0.00  0.01
20:49:39 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   11    0.00  0.00  0.02  0.01  0.00  0.00
20:49:39 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   11    0.00  0.00  0.02  0.01  0.00  0.00
20:49:40 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   12    0.00  0.00  0.02  0.01  0.00  0.00
20:49:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   12    0.00  0.00  0.98  0.00  0.00  0.00
20:49:40 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   12    0.00  0.00  0.02  0.01  0.00  0.00
20:49:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   13    0.00  0.00  0.99  0.00  0.00  0.00
20:49:41 B17F.2   B17F24900917  sit     75   NoReport sit                trk  1.00 OpenFloor  3   13    0.00  0.00  0.02  0.01  0.00  0.00
20:49:41 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   13    0.00  0.00  0.02  0.01  0.00  0.00
20:49:42 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   14    0.00  0.00  0.02  0.01  0.00  0.00
20:49:42 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.00  0.99  0.00  0.00  0.00
20:49:42 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.00  0.02  0.01  0.00  0.00
20:49:43 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.00  0.99  0.00  0.00  0.00
20:49:43 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.00  0.02  0.01  0.00  0.00
20:49:43 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   15    0.00  0.00  0.02  0.01  0.00  0.00
20:49:44 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   16    0.00  0.00  0.02  0.01  0.00  0.00
20:49:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.00  0.99  0.00  0.00  0.00
20:49:44 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.00  0.02  0.01  0.00  0.00
20:49:45 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   17    0.00  0.00  0.02  0.01  0.00  0.00
20:49:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.00  0.99  0.00  0.00  0.00
20:49:45 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.00  0.02  0.01  0.00  0.00
20:49:45 B17F.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  3   17    0.00  0.01  0.93  0.00  0.00  0.01
20:49:46 B17F.E   -             -       0    NoReport np=2               room -    OpenFloor  3   17    0.00  0.01  0.93  0.00  0.00  0.01
20:49:46 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   18    0.00  0.00  0.03  0.01  0.00  0.01
20:49:46 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.00  0.03  0.01  0.00  0.01
20:49:47 B17F.E   -             -       0    NoReport np=3               room -    OpenFloor  3   18    0.00  0.02  0.87  0.00  0.00  0.02
20:49:47 B17F.0   B17F04900917  stand   83   NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.94  0.00  0.00  0.01
20:49:47 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.00  0.02  0.01  0.00  0.00
20:49:47 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   19    0.00  0.00  0.02  0.01  0.00  0.00
20:49:48 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   20    0.00  0.00  0.02  0.01  0.00  0.00
20:49:48 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.00  0.02  0.01  0.00  0.00
20:49:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.01  0.93  0.00  0.00  0.01
20:49:49 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   21    0.00  0.00  0.02  0.01  0.00  0.00
20:49:49 B17F.1   B17F24900917  stand   78   NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.00  0.02  0.01  0.00  0.00
20:49:49 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.01  0.93  0.00  0.00  0.01
20:49:50 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   9     0.00  0.00  0.02  0.01  0.00  0.00
20:49:50 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.01  0.93  0.00  0.00  0.01
20:49:50 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   9     0.00  0.00  0.02  0.01  0.00  0.00
20:49:51 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   10    0.00  0.01  0.93  0.00  0.00  0.01
20:49:51 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   10    0.00  0.00  0.02  0.01  0.00  0.00
20:49:51 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   10    0.00  0.00  0.02  0.01  0.00  0.00
20:49:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   11    0.00  0.01  0.93  0.00  0.00  0.01
20:49:52 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   11    0.00  0.00  0.02  0.01  0.00  0.00
20:49:52 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   11    0.00  0.00  0.02  0.01  0.00  0.00
20:49:53 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   12    0.00  0.00  0.02  0.01  0.00  0.00
20:49:53 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   12    0.00  0.00  0.02  0.01  0.00  0.00
20:49:53 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   12    0.00  0.01  0.93  0.00  0.00  0.01
20:49:54 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   13    0.00  0.00  0.02  0.01  0.00  0.00
20:49:54 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   13    0.00  0.00  0.02  0.01  0.00  0.00
20:49:54 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   13    0.00  0.01  0.93  0.00  0.00  0.01
20:49:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.01  0.93  0.00  0.00  0.01
20:49:55 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   14    0.00  0.00  0.02  0.01  0.00  0.00
20:49:55 B17F.1   B17F24900917  stand   84   NoReport stand              trk  1.00 OpenFloor  3   14    0.00  0.00  0.02  0.01  0.00  0.00
20:49:56 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.01  0.93  0.00  0.00  0.01
20:49:56 B17F.1   B17F24900917  stand   83   NoReport stand              trk  1.00 OpenFloor  3   15    0.00  0.00  0.02  0.01  0.00  0.00
20:49:56 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   15    0.00  0.00  0.02  0.01  0.00  0.00
20:49:57 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   16    0.00  0.00  0.02  0.01  0.00  0.00
20:49:57 B17F.1   B17F24900917  walk    40   NoReport walk               trk  1.00 OpenFloor  3   16    0.00  0.00  0.02  0.01  0.00  0.00
20:49:57 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   16    0.00  0.01  0.93  0.00  0.00  0.01
20:49:58 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   17    0.00  0.01  0.93  0.00  0.00  0.01
20:49:58 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  3   17    0.00  0.00  0.02  0.01  0.00  0.00
20:49:58 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   17    0.00  0.00  0.02  0.01  0.00  0.00
20:49:59 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   18    0.00  0.00  0.02  0.01  0.00  0.00
20:49:59 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  3   18    0.00  0.00  0.02  0.01  0.00  0.00
20:49:59 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   18    0.00  0.01  0.93  0.00  0.00  0.01
20:50:00 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   19    0.00  0.00  0.02  0.01  0.00  0.00
20:50:00 B17F.0   B17F04900917  stand   77   NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.01  0.93  0.00  0.00  0.01
20:50:00 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   19    0.00  0.00  0.02  0.01  0.00  0.00
20:50:01 B17F.0   B17F04900917  stand   77   NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.01  0.93  0.00  0.00  0.01
20:50:01 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   20    0.00  0.00  0.02  0.01  0.00  0.00
20:50:01 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   20    0.00  0.00  0.02  0.01  0.00  0.00
20:50:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.01  0.93  0.00  0.00  0.01
20:50:02 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   21    0.00  0.00  0.02  0.01  0.00  0.00
20:50:02 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   21    0.00  0.00  0.02  0.01  0.00  0.00
20:50:03 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   22    0.00  0.00  0.02  0.01  0.00  0.00
20:50:03 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.00  0.02  0.01  0.00  0.00
20:50:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   22    0.00  0.01  0.93  0.00  0.00  0.01
20:50:04 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   23    0.00  0.00  0.02  0.01  0.00  0.00
20:50:04 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.00  0.02  0.01  0.00  0.00
20:50:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   23    0.00  0.01  0.93  0.00  0.00  0.01
20:50:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   24    0.00  0.01  0.93  0.00  0.00  0.01
20:50:05 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   24    0.00  0.00  0.02  0.01  0.00  0.00
20:50:05 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   24    0.00  0.00  0.02  0.01  0.00  0.00
20:50:06 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   25    0.00  0.01  0.93  0.00  0.00  0.01
20:50:06 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   25    0.00  0.00  0.02  0.01  0.00  0.00
20:50:06 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   25    0.00  0.00  0.02  0.01  0.00  0.00
20:50:07 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   26    0.00  0.00  0.02  0.01  0.00  0.00
20:50:07 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   26    0.00  0.00  0.02  0.01  0.00  0.00
20:50:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   26    0.00  0.01  0.93  0.00  0.00  0.01
20:50:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   27    0.00  0.01  0.93  0.00  0.00  0.01
20:50:08 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   27    0.00  0.00  0.02  0.01  0.00  0.00
20:50:08 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   27    0.00  0.00  0.02  0.01  0.00  0.00
20:50:09 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   28    0.00  0.00  0.02  0.01  0.00  0.00
20:50:09 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   28    0.00  0.00  0.02  0.01  0.00  0.00
20:50:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   28    0.00  0.01  0.93  0.00  0.00  0.01
20:50:10 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   29    0.00  0.00  0.02  0.01  0.00  0.00
20:50:10 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   29    0.00  0.00  0.02  0.01  0.00  0.00
20:50:10 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   29    0.00  0.01  0.93  0.00  0.00  0.01
20:50:11 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   30    0.00  0.00  0.02  0.01  0.00  0.00
20:50:11 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   30    0.00  0.00  0.02  0.01  0.00  0.00
20:50:11 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   30    0.00  0.01  0.93  0.00  0.00  0.01
20:50:12 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   31    0.00  0.00  0.02  0.01  0.00  0.00
20:50:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   31    0.00  0.01  0.93  0.00  0.00  0.01
20:50:12 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   31    0.00  0.00  0.02  0.01  0.00  0.00
20:50:13 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   32    0.00  0.00  0.02  0.01  0.00  0.00
20:50:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   32    0.00  0.01  0.93  0.00  0.00  0.01
20:50:13 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   32    0.00  0.00  0.02  0.01  0.00  0.00
20:50:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   33    0.00  0.01  0.93  0.00  0.00  0.01
20:50:14 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   33    0.00  0.00  0.02  0.01  0.00  0.00
20:50:14 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   33    0.00  0.00  0.02  0.01  0.00  0.00
20:50:15 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   34    0.00  0.00  0.02  0.01  0.00  0.00
20:50:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.01  0.93  0.00  0.00  0.01
20:50:15 B17F.1   B17F24900917  stand   65   NoReport stand              trk  1.00 OpenFloor  3   34    0.00  0.00  0.02  0.01  0.00  0.00
20:50:16 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   35    0.00  0.00  0.02  0.01  0.00  0.00
20:50:16 B17F.0   B17F04900917  stand   80   NoReport stand              trk  1.00 OpenFloor  3   35    0.00  0.01  0.93  0.00  0.00  0.01
20:50:16 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   35    0.00  0.00  0.02  0.01  0.00  0.00
20:50:17 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   36    0.00  0.00  0.02  0.01  0.00  0.00
20:50:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   36    0.00  0.01  0.93  0.00  0.00  0.01
20:50:17 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   36    0.00  0.00  0.02  0.01  0.00  0.00
20:50:18 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   37    0.00  0.00  0.02  0.01  0.00  0.00
20:50:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   37    0.00  0.01  0.93  0.00  0.00  0.01
20:50:18 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   37    0.00  0.00  0.02  0.01  0.00  0.00
20:50:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   38    0.00  0.01  0.93  0.00  0.00  0.01
20:50:19 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   38    0.00  0.00  0.02  0.01  0.00  0.00
20:50:19 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   38    0.00  0.00  0.02  0.01  0.00  0.00
20:50:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   39    0.00  0.01  0.93  0.00  0.00  0.01
20:50:20 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   39    0.00  0.00  0.02  0.01  0.00  0.00
20:50:20 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   39    0.00  0.00  0.02  0.01  0.00  0.00
20:50:21 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   40    0.00  0.00  0.02  0.01  0.00  0.00
20:50:21 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   40    0.00  0.00  0.02  0.01  0.00  0.00
20:50:21 B17F.0   B17F04900917  stand   83   NoReport stand              trk  1.00 OpenFloor  3   40    0.00  0.01  0.93  0.00  0.00  0.01
20:50:22 B17F.0   B17F04900917  stand   72   NoReport stand              trk  1.00 OpenFloor  3   41    0.00  0.01  0.93  0.00  0.00  0.01
20:50:22 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   41    0.00  0.00  0.02  0.01  0.00  0.00
20:50:22 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   41    0.00  0.00  0.02  0.01  0.00  0.00
20:50:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   42    0.00  0.01  0.93  0.00  0.00  0.01
20:50:23 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   42    0.00  0.00  0.02  0.01  0.00  0.00
20:50:23 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   42    0.00  0.00  0.02  0.01  0.00  0.00
20:50:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   43    0.00  0.01  0.93  0.00  0.00  0.01
20:50:24 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   43    0.00  0.00  0.02  0.01  0.00  0.00
20:50:24 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   43    0.00  0.00  0.02  0.01  0.00  0.00
20:50:25 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   44    0.00  0.00  0.02  0.01  0.00  0.00
20:50:25 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   44    0.00  0.00  0.02  0.01  0.00  0.00
20:50:25 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   44    0.00  0.01  0.93  0.00  0.00  0.01
20:50:26 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.01  0.93  0.00  0.00  0.01
20:50:26 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   45    0.00  0.00  0.02  0.01  0.00  0.00
20:50:26 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   45    0.00  0.00  0.02  0.01  0.00  0.00
20:50:27 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.00  0.02  0.01  0.00  0.00
20:50:27 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   46    0.00  0.00  0.02  0.01  0.00  0.00
20:50:27 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   46    0.00  0.01  0.93  0.00  0.00  0.01
20:50:28 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   47    0.00  0.00  0.02  0.01  0.00  0.00
20:50:28 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   47    0.00  0.01  0.93  0.00  0.00  0.01
20:50:28 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   47    0.00  0.00  0.02  0.01  0.00  0.00
20:50:29 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   48    0.00  0.01  0.93  0.00  0.00  0.01
20:50:29 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   48    0.00  0.00  0.02  0.01  0.00  0.00
20:50:29 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   48    0.00  0.00  0.02  0.01  0.00  0.00
20:50:30 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   49    0.00  0.00  0.02  0.01  0.00  0.00
20:50:30 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   49    0.00  0.01  0.93  0.00  0.00  0.01
20:50:30 B17F.1   B17F24900917  stand   39   NoReport stand              trk  1.00 OpenFloor  3   49    0.00  0.00  0.02  0.01  0.00  0.00
20:50:31 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   50    0.00  0.01  0.93  0.00  0.00  0.01
20:50:31 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   50    0.00  0.00  0.02  0.01  0.00  0.00
20:50:31 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   50    0.00  0.00  0.02  0.01  0.00  0.00
20:50:32 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   51    0.00  0.01  0.93  0.00  0.00  0.01
20:50:32 B17F.1   B17F24900917  stand   38   NoReport stand              trk  1.00 OpenFloor  3   51    0.00  0.00  0.02  0.01  0.00  0.00
20:50:32 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   51    0.00  0.00  0.02  0.01  0.00  0.00
20:50:33 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   52    0.00  0.00  0.02  0.01  0.00  0.00
20:50:33 B17F.1   B17F24900917  stand   73   NoReport stand              trk  1.00 OpenFloor  3   52    0.00  0.00  0.02  0.01  0.00  0.00
20:50:33 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   52    0.00  0.01  0.93  0.00  0.00  0.01
20:50:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   53    0.00  0.01  0.93  0.00  0.00  0.01
20:50:34 B17F.1   B17F24900917  stand   75   NoReport stand              trk  1.00 OpenFloor  3   53    0.00  0.00  0.02  0.01  0.00  0.00
20:50:34 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   53    0.00  0.00  0.02  0.01  0.00  0.00
20:50:35 B17F.1   B17F24900917  walk    58   NoReport walk               trk  1.00 OpenFloor  3   54    0.00  0.00  0.02  0.01  0.00  0.01
20:50:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   54    0.00  0.01  0.90  0.00  0.01  0.01
20:50:35 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   54    0.00  0.00  0.02  0.01  0.00  0.01
20:50:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   55    0.00  0.01  0.93  0.00  0.00  0.01
20:50:36 B17F.1   B17F24900917  walk    97   NoReport walk               trk  1.00 OpenFloor  3   55    0.00  0.00  0.02  0.01  0.00  0.00
20:50:36 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   55    0.00  0.00  0.02  0.01  0.00  0.00
20:50:37 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   56    0.00  0.00  0.02  0.01  0.00  0.00
20:50:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   56    0.00  0.01  0.93  0.00  0.00  0.01
20:50:37 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  3   56    0.00  0.00  0.02  0.01  0.00  0.00
20:50:38 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   57    0.00  0.00  0.02  0.01  0.00  0.00
20:50:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   57    0.00  0.01  0.93  0.00  0.00  0.01
20:50:38 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   57    0.00  0.00  0.02  0.01  0.00  0.00
20:50:39 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   58    0.00  0.00  0.02  0.01  0.00  0.00
20:50:39 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   58    0.00  0.01  0.93  0.00  0.00  0.01
20:50:39 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   58    0.00  0.00  0.02  0.01  0.00  0.00
20:50:40 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   59    0.00  0.00  0.02  0.01  0.00  0.00
20:50:40 B17F.1   B17F24900917  stand   87   NoReport stand              trk  1.00 OpenFloor  3   59    0.00  0.00  0.02  0.01  0.00  0.00
20:50:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   59    0.00  0.01  0.93  0.00  0.00  0.01
20:50:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   60    0.00  0.01  0.93  0.00  0.00  0.01
20:50:41 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   60    0.00  0.00  0.02  0.01  0.00  0.00
20:50:41 B17F.1   B17F24900917  stand   101  NoReport stand              trk  1.00 OpenFloor  3   60    0.00  0.00  0.02  0.01  0.00  0.00
20:50:42 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   61    0.00  0.00  0.02  0.01  0.00  0.00
20:50:42 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   61    0.00  0.01  0.93  0.00  0.00  0.01
20:50:42 B17F.1   B17F24900917  stand   92   NoReport stand              trk  1.00 OpenFloor  3   61    0.00  0.00  0.02  0.01  0.00  0.00
20:50:43 B17F.1   B17F24900917  stand   98   NoReport stand              trk  1.00 OpenFloor  3   62    0.00  0.00  0.02  0.01  0.00  0.00
20:50:43 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   62    0.00  0.00  0.02  0.01  0.00  0.00
20:50:43 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   62    0.00  0.01  0.93  0.00  0.00  0.01
20:50:44 B17F.1   B17F24900917  stand   95   NoReport stand              trk  1.00 OpenFloor  3   63    0.00  0.00  0.02  0.01  0.00  0.00
20:50:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   63    0.00  0.01  0.93  0.00  0.00  0.01
20:50:44 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   63    0.00  0.00  0.02  0.01  0.00  0.00
20:50:45 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   64    0.00  0.00  0.02  0.01  0.00  0.00
20:50:45 B17F.1   B17F24900917  stand   103  NoReport stand              trk  1.00 OpenFloor  3   64    0.00  0.00  0.02  0.01  0.00  0.00
20:50:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   64    0.00  0.01  0.93  0.00  0.00  0.01
20:50:45 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   65    0.00  0.00  0.02  0.01  0.00  0.00
20:50:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   65    0.00  0.01  0.93  0.00  0.00  0.01
20:50:45 B17F.1   B17F24900917  stand   107  NoReport stand              trk  1.00 OpenFloor  3   65    0.00  0.00  0.02  0.01  0.00  0.00
20:50:46 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  3   65    0.00  0.01  0.93  0.00  0.00  0.01
20:50:47 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   66    0.00  0.01  0.93  0.00  0.00  0.01
20:50:47 B17F.2   B17F24900917  sit     0    NoReport sit                trk  1.00 OpenFloor  3   66    0.00  0.00  0.02  0.01  0.00  0.00
20:50:47 B17F.1   B17F24900917  stand   97   NoReport stand              trk  1.00 OpenFloor  3   66    0.00  0.00  0.02  0.01  0.00  0.00
20:50:48 B17F.E   -             -       0    NoReport np=2               room -    OpenFloor  3   66    0.00  0.01  0.93  0.00  0.00  0.01
20:50:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  3   67    0.00  0.02  0.87  0.00  0.00  0.02
20:50:48 B17F.1   B17F24900917  stand   92   NoReport stand              trk  1.00 OpenFloor  3   67    0.00  0.00  0.03  0.01  0.00  0.01
20:50:49 B17F.1   B17F24900917  stand   79   NoReport stand              trk  1.00 Sit        2   24    0.00  0.01  0.12  0.04  0.01  0.03
20:50:49 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Sit        2   24    0.00  0.02  0.86  0.00  0.01  0.02
20:50:50 B17F.1   B17F24900917  stand   64   NoReport stand              trk  1.00 Sit        2   25    0.01  0.03  0.16  0.06  0.03  0.03
20:50:50 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Sit        2   25    0.00  0.02  0.85  0.00  0.01  0.02
20:50:51 B17F.1   B17F24900917  stand   87   NoReport stand              trk  1.00 Sit        2   26    0.01  0.04  0.18  0.07  0.06  0.03
20:50:51 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Sit        2   26    0.00  0.02  0.85  0.00  0.01  0.02
20:50:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Sit        2   27    0.00  0.02  0.85  0.00  0.01  0.02
20:50:52 B17F.1   B17F24900917  stand   80   NoReport stand              trk  1.00 Sit        2   27    0.01  0.05  0.19  0.08  0.08  0.03
20:50:53 B17F.1   B17F24900917  stand   73   NoReport stand              trk  1.00 Sit        2   28    0.02  0.06  0.19  0.09  0.10  0.03
20:50:53 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Sit        2   28    0.00  0.02  0.85  0.00  0.01  0.02
20:50:54 B17F.1   B17F24900917  stand   67   NoReport stand              trk  1.00 Sit        2   29    0.02  0.07  0.19  0.09  0.12  0.03
20:50:54 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Sit        2   29    0.00  0.02  0.85  0.00  0.01  0.02
20:50:55 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Sit        2   30    0.02  0.08  0.19  0.10  0.13  0.03
20:50:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Sit        2   30    0.00  0.02  0.85  0.00  0.01  0.02
20:50:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Sit        2   31    0.00  0.02  0.85  0.00  0.01  0.02
20:50:55 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Sit        2   31    0.03  0.09  0.18  0.10  0.14  0.02
20:50:56 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   32    0.03  0.09  0.18  0.11  0.15  0.02
20:50:56 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   32    0.00  0.02  0.85  0.00  0.01  0.02
20:50:57 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   33    0.03  0.09  0.18  0.11  0.16  0.02
20:50:57 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   33    0.00  0.02  0.85  0.00  0.01  0.02
20:50:58 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   34    0.04  0.10  0.17  0.11  0.17  0.02
20:50:58 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   34    0.00  0.02  0.85  0.00  0.01  0.02
20:50:59 B17F.1   B17F24900917  stand   57   NoReport stand              trk  1.00 BlindOpen  2   35    0.04  0.10  0.17  0.12  0.17  0.02
20:50:59 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   35    0.00  0.02  0.85  0.00  0.01  0.02
20:51:00 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   36    0.00  0.02  0.85  0.00  0.01  0.02
20:51:00 B17F.1   B17F24900917  stand   75   NoReport stand              trk  1.00 BlindOpen  2   36    0.04  0.10  0.17  0.12  0.18  0.02
20:51:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   37    0.00  0.02  0.85  0.00  0.01  0.02
20:51:01 B17F.1   B17F24900917  stand   74   NoReport stand              trk  1.00 BlindOpen  2   37    0.05  0.10  0.17  0.12  0.18  0.02
20:51:02 B17F.1   B17F24900917  stand   60   NoReport stand              trk  1.00 BlindOpen  2   38    0.05  0.10  0.17  0.12  0.18  0.02
20:51:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   38    0.00  0.02  0.85  0.00  0.01  0.02
20:51:03 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.05  0.10  0.16  0.12  0.18  0.02
20:51:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:04 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.05  0.10  0.16  0.12  0.19  0.02
20:51:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:05 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.06  0.10  0.16  0.12  0.19  0.02
20:51:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:06 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:06 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 BlindOpen  2   0     0.06  0.10  0.16  0.13  0.19  0.02
20:51:07 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.06  0.11  0.16  0.13  0.19  0.02
20:51:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:08 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.06  0.11  0.16  0.13  0.19  0.02
20:51:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:09 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.07  0.11  0.16  0.13  0.19  0.02
20:51:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:10 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Empty      2   13    0.00  0.02  0.85  0.00  0.01  0.02
20:51:10 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   13    0.07  0.11  0.16  0.13  0.19  0.02
20:51:11 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   14    0.07  0.11  0.16  0.13  0.19  0.02
20:51:11 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Empty      2   14    0.00  0.02  0.85  0.00  0.01  0.02
20:51:12 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.07  0.11  0.16  0.13  0.19  0.02
20:51:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:13 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.08  0.11  0.15  0.13  0.19  0.02
20:51:14 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.08  0.11  0.15  0.13  0.19  0.02
20:51:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:15 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.08  0.11  0.15  0.13  0.19  0.02
20:51:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:16 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.08  0.11  0.15  0.13  0.19  0.02
20:51:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   13    0.00  0.02  0.85  0.00  0.01  0.02
20:51:17 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   13    0.08  0.11  0.15  0.13  0.19  0.02
20:51:18 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.09  0.11  0.15  0.13  0.19  0.02
20:51:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:19 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.09  0.11  0.15  0.13  0.19  0.02
20:51:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:20 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.09  0.11  0.15  0.13  0.19  0.02
20:51:21 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.09  0.11  0.15  0.13  0.19  0.02
20:51:21 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:22 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.10  0.11  0.15  0.13  0.19  0.02
20:51:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:23 B17F.1   B17F24900917  stand   38   NoReport stand              trk  1.00 Empty      2   0     0.10  0.10  0.15  0.13  0.19  0.02
20:51:24 B17F.1   B17F24900917  stand   81   NoReport stand              trk  1.00 Empty      2   0     0.10  0.10  0.15  0.13  0.19  0.02
20:51:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:25 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:25 B17F.1   B17F24900917  walk    68   NoReport walk               trk  1.00 Empty      2   0     0.10  0.10  0.15  0.13  0.19  0.02
20:51:26 B17F.1   B17F24900917  walk    107  NoReport walk               trk  1.00 Empty      2   0     0.10  0.10  0.15  0.13  0.19  0.02
20:51:26 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:27 B17F.1   B17F24900917  walk    110  NoReport walk               trk  1.00 Empty      2   0     0.11  0.10  0.15  0.13  0.19  0.02
20:51:27 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:28 B17F.1   B17F24900917  walk    114  NoReport walk               trk  1.00 Empty      2   0     0.11  0.10  0.15  0.13  0.19  0.02
20:51:28 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:29 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:29 B17F.1   B17F24900917  walk    94   NoReport walk               trk  1.00 Empty      2   0     0.11  0.10  0.15  0.13  0.19  0.02
20:51:30 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 Empty      2   0     0.11  0.10  0.15  0.13  0.19  0.02
20:51:30 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:31 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.11  0.10  0.15  0.13  0.19  0.02
20:51:31 B17F.0   B17F04900917  stand   76   NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:32 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.11  0.10  0.15  0.13  0.19  0.02
20:51:32 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:33 B17F.0   B17F04900917  stand   71   NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:33 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.12  0.10  0.15  0.13  0.19  0.02
20:51:34 B17F.1   B17F24900917  stand   113  NoReport stand              trk  1.00 Empty      2   0     0.12  0.10  0.15  0.13  0.19  0.02
20:51:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.03  0.81  0.00  0.02  0.02
20:51:35 B17F.1   B17F24900917  walk    119  NoReport walk               trk  1.00 Empty      2   0     0.12  0.10  0.15  0.13  0.19  0.02
20:51:36 B17F.1   B17F24900917  walk    105  NoReport walk               trk  1.00 Empty      2   0     0.12  0.10  0.15  0.13  0.19  0.02
20:51:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:51:37 B17F.1   B17F24900917  walk    88   NoReport walk               trk  1.00 Empty      2   0     0.12  0.10  0.15  0.13  0.19  0.02
20:51:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:51:38 B17F.1   B17F24900917  walk    76   NoReport walk               trk  1.00 Empty      2   0     0.13  0.10  0.14  0.13  0.19  0.02
20:51:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:39 B17F.1   B17F24900917  walk    97   NoReport walk               trk  1.00 Empty      2   0     0.13  0.10  0.14  0.13  0.19  0.02
20:51:39 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:40 B17F.1   B17F24900917  walk    112  NoReport walk               trk  1.00 Empty      2   0     0.13  0.10  0.14  0.13  0.19  0.02
20:51:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:41 B17F.1   B17F24900917  walk    110  NoReport walk               trk  1.00 Empty      2   0     0.13  0.10  0.14  0.12  0.19  0.02
20:51:42 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:42 B17F.1   B17F24900917  walk    99   NoReport walk               trk  1.00 Empty      2   0     0.13  0.10  0.14  0.12  0.19  0.02
20:51:43 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 Empty      2   0     0.13  0.10  0.14  0.12  0.19  0.02
20:51:43 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:44 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 Empty      2   0     0.14  0.10  0.14  0.12  0.18  0.02
20:51:44 B17F.0   B17F04900917  stand   75   NoReport stand              trk  1.00 Empty      2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:51:45 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   13    0.14  0.10  0.14  0.12  0.18  0.02
20:51:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   13    0.00  0.02  0.85  0.00  0.01  0.02
20:51:46 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   14    0.00  0.02  0.85  0.00  0.01  0.02
20:51:46 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   14    0.14  0.10  0.14  0.12  0.18  0.02
20:51:47 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   15    0.14  0.10  0.14  0.12  0.18  0.02
20:51:47 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   15    0.00  0.02  0.85  0.00  0.01  0.02
20:51:48 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 Empty      2   16    0.14  0.10  0.14  0.12  0.18  0.02
20:51:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   16    0.00  0.02  0.85  0.00  0.01  0.02
20:51:48 B17F.E   -             -       0    NoReport ExitRoom(rdr)      room -    Empty      2   16    0.14  0.10  0.14  0.12  0.18  0.02
20:51:49 B17F.E   -             -       0    NoReport np=1               room -    Empty      2   16    0.14  0.10  0.14  0.12  0.18  0.02
20:51:49 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      2   17    0.00  0.04  0.74  0.00  0.02  0.04
20:51:50 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   18    0.01  0.05  0.68  0.01  0.03  0.04
20:51:51 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   19    0.01  0.05  0.65  0.01  0.04  0.03
20:51:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   20    0.01  0.05  0.63  0.01  0.04  0.03
20:51:53 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   21    0.01  0.05  0.62  0.02  0.05  0.03
20:51:54 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   22    0.01  0.05  0.62  0.02  0.05  0.03
20:51:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.62  0.02  0.05  0.03
20:51:56 B17F.0   B17F04900917  stand   70   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:51:57 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:51:58 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:51:59 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:52:00 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:52:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:52:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:52:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:52:04 B17F.E   -             -       0    NoReport np=2               room -    Empty      1   0     0.16  0.10  0.14  0.12  0.18  0.02
20:52:04 B17F.1   B17F24900917  stand   97   NoReport stand              trk  1.00 OpenFloor  2   0     0.03  0.08  0.43  0.09  0.14  0.01
20:52:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.03  0.76  0.01  0.03  0.02
20:52:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.82  0.01  0.02  0.02
20:52:05 B17F.1   B17F24900917  walk    86   NoReport walk               trk  1.00 OpenFloor  2   0     0.01  0.07  0.50  0.07  0.11  0.03
20:52:06 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.01  0.07  0.54  0.05  0.09  0.03
20:52:06 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:52:07 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.06  0.57  0.04  0.08  0.03
20:52:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:52:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:08 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.08  0.41  0.05  0.09  0.04
20:52:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:09 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.07  0.48  0.04  0.09  0.03
20:52:10 B17F.0   B17F04900917  stand   75   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:10 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.04  0.69  0.02  0.05  0.02
20:52:11 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.79  0.01  0.03  0.02
20:52:11 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:12 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.83  0.01  0.02  0.02
20:52:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
20:52:13 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.02  0.84  0.00  0.01  0.02
20:52:14 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
20:52:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
20:52:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
20:52:15 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.04  0.73  0.00  0.02  0.04
20:52:16 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   20    0.00  0.03  0.81  0.00  0.02  0.02
20:52:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
20:52:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
20:52:17 B17F.1   B17F24900917  stand   96   NoReport stand              trk  1.00 OpenFloor  2   21    0.00  0.02  0.84  0.00  0.01  0.02
20:52:18 B17F.1   B17F24900917  stand   96   NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
20:52:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
20:52:19 B17F.0   B17F04900917  stand   63   NoReport stand              trk  1.00 OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
20:52:19 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
20:52:20 B17F.1   B17F24900917  walk    108  NoReport walk               trk  1.00 OpenFloor  1   24    0.00  0.04  0.74  0.00  0.02  0.04
20:52:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   24    0.00  0.02  0.85  0.00  0.01  0.02
20:52:21 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
20:52:21 B17F.1   B17F24900917  walk    110  NoReport walk               trk  1.00 OpenFloor  2   25    0.00  0.03  0.81  0.00  0.02  0.02
20:52:22 B17F.1   B17F24900917  walk    89   NoReport walk               trk  1.00 OpenFloor  2   26    0.00  0.02  0.84  0.00  0.01  0.02
20:52:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
20:52:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
20:52:23 B17F.1   B17F24900917  walk    87   NoReport walk               trk  1.00 OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
20:52:24 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   28    0.00  0.04  0.74  0.00  0.02  0.04
20:52:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   28    0.00  0.02  0.85  0.00  0.01  0.02
20:52:25 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   29    0.00  0.02  0.85  0.00  0.01  0.02
20:52:25 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   29    0.00  0.03  0.81  0.00  0.02  0.02
20:52:26 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   30    0.01  0.07  0.56  0.01  0.04  0.06
20:52:26 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
20:52:27 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:27 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.04  0.74  0.01  0.04  0.02
20:52:28 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:52:28 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:29 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:29 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.04  0.72  0.01  0.02  0.04
20:52:30 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.80  0.01  0.02  0.02
20:52:30 B17F.0   B17F04900917  stand   75   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:31 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.71  0.01  0.02  0.04
20:52:31 B17F.0   B17F04900917  stand   78   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:32 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:32 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.05  0.67  0.01  0.03  0.04
20:52:33 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:33 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.07  0.47  0.02  0.06  0.05
20:52:34 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.04  0.69  0.02  0.05  0.02
20:52:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.00  0.02  0.02
20:52:35 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.03  0.78  0.01  0.03  0.02
20:52:36 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.82  0.01  0.02  0.02
20:52:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:52:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:52:37 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:52:38 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.58  0.01  0.03  0.06
20:52:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:39 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:39 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.59  0.01  0.05  0.03
20:52:40 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.03  0.75  0.01  0.03  0.02
20:52:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:41 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:52:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:42 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:52:42 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:43 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:43 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:52:44 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   13    0.00  0.04  0.73  0.00  0.02  0.04
20:52:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
20:52:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
20:52:45 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.03  0.81  0.00  0.02  0.02
20:52:46 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.02  0.84  0.00  0.01  0.02
20:52:46 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
20:52:47 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
20:52:47 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   16    0.01  0.04  0.73  0.01  0.02  0.04
20:52:48 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.03  0.81  0.01  0.02  0.02
20:52:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
20:52:49 B17F.0   B17F04900917  stand   73   NoReport stand              trk  1.00 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
20:52:49 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   18    0.01  0.04  0.72  0.01  0.02  0.04
20:52:50 B17F.0   B17F04900917  stand   72   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:50 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.80  0.01  0.02  0.02
20:52:51 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:51 B17F.1   B17F24900917  stand   63   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.83  0.00  0.01  0.02
20:52:52 B17F.1   B17F24900917  walk    98   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.73  0.01  0.02  0.04
20:52:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:52 B17F.1   B17F24900917  walk    110  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:52:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:53 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:54 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:54 B17F.1   B17F24900917  walk    120  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:52:55 B17F.1   B17F24900917  walk    128  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.04  0.73  0.01  0.02  0.04
20:52:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:55 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:52:56 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:57 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:52:57 B17F.0   B17F04900917  stand   78   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:57 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:52:57 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.04  0.73  0.01  0.02  0.04
20:52:58 -.-      -             -       -    NoReport (no frame, held)   room -    OpenFloor  2   0     0.00  0.04  0.73  0.01  0.02  0.04
20:52:59 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:52:59 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:00 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:00 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.04  0.72  0.01  0.02  0.04
20:53:01 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.80  0.01  0.02  0.02
20:53:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:01 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.83  0.00  0.01  0.02
20:53:02 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:53:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:03 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:04 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:05 B17F.1   B17F24900917  stand   112  NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:06 B17F.1   B17F24900917  stand   100  NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:06 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:07 B17F.1   B17F24900917  walk    92   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.74  0.00  0.02  0.04
20:53:08 B17F.1   B17F24900917  walk    86   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.00  0.02  0.02
20:53:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:09 B17F.1   B17F24900917  walk    76   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.04  0.72  0.01  0.02  0.04
20:53:10 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.67  0.01  0.03  0.04
20:53:10 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:11 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:11 B17F.1   B17F24900917  walk    37   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.79  0.01  0.03  0.02
20:53:12 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.71  0.01  0.03  0.04
20:53:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:13 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.66  0.01  0.04  0.03
20:53:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:14 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.78  0.01  0.03  0.02
20:53:15 B17F.1   B17F24900917  stand   67   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.83  0.01  0.02  0.02
20:53:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:16 B17F.1   B17F24900917  stand   78   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:53:17 B17F.1   B17F24900917  stand   95   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:18 B17F.1   B17F24900917  stand   107  NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:19 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:20 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:21 B17F.1   B17F24900917  stand   65   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.04  0.74  0.00  0.02  0.04
20:53:21 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:22 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.05  0.68  0.01  0.03  0.04
20:53:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:23 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.65  0.01  0.04  0.03
20:53:23 B17F.0   B17F04900917  stand   70   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:24 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.78  0.01  0.03  0.02
20:53:24 B17F.0   B17F04900917  walk    74   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:25 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.82  0.01  0.02  0.02
20:53:25 B17F.0   B17F04900917  walk    66   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:26 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:26 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.72  0.01  0.02  0.04
20:53:27 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:53:27 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:28 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:28 B17F.1   B17F24900917  walk    69   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.72  0.01  0.02  0.04
20:53:29 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.67  0.01  0.03  0.04
20:53:29 B17F.0   B17F04900917  walk    66   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:30 B17F.0   B17F04900917  walk    95   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:30 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.79  0.01  0.03  0.02
20:53:31 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.83  0.00  0.02  0.02
20:53:31 B17F.0   B17F04900917  walk    102  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:32 B17F.0   B17F04900917  walk    101  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:32 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:53:33 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.05  0.68  0.01  0.03  0.04
20:53:33 B17F.0   B17F04900917  walk    106  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.00  0.02  0.02
20:53:34 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.05  0.65  0.01  0.04  0.03
20:53:34 B17F.0   B17F04900917  walk    111  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:53:35 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.78  0.01  0.03  0.02
20:53:35 B17F.0   B17F04900917  walk    107  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:36 B17F.0   B17F04900917  walk    103  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:36 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.04  0.70  0.01  0.03  0.03
20:53:37 B17F.0   B17F04900917  walk    117  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:37 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.05  0.66  0.01  0.04  0.03
20:53:38 B17F.0   B17F04900917  walk    108  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:38 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.78  0.01  0.03  0.02
20:53:39 B17F.0   B17F04900917  walk    110  NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:39 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.07  0.54  0.01  0.04  0.05
20:53:40 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.06  0.57  0.02  0.06  0.03
20:53:40 B17F.0   B17F04900917  walk    113  NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:41 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   13    0.01  0.04  0.74  0.01  0.04  0.02
20:53:41 B17F.0   B17F04900917  walk    124  NoReport walk               trk  1.00 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
20:53:42 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.03  0.81  0.01  0.02  0.02
20:53:42 B17F.0   B17F04900917  walk    118  NoReport walk               trk  1.00 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
20:53:43 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.02  0.84  0.00  0.01  0.02
20:53:43 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
20:53:44 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   16    0.00  0.02  0.84  0.00  0.01  0.02
20:53:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
20:53:45 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
20:53:45 B17F.0   B17F04900917  stand   113  NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
20:53:46 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.04  0.74  0.00  0.02  0.04
20:53:46 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
20:53:47 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.03  0.81  0.00  0.02  0.02
20:53:47 B17F.0   B17F04900917  stand   108  NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
20:53:48 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   17    0.01  0.04  0.72  0.01  0.02  0.04
20:53:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
20:53:49 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   21    0.00  0.03  0.80  0.01  0.02  0.02
20:53:49 B17F.0   B17F04900917  stand   103  NoReport stand              trk  1.00 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
20:53:50 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   22    0.01  0.04  0.71  0.01  0.02  0.04
20:53:50 B17F.0   B17F04900917  stand   97   NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
20:53:51 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   23    0.00  0.03  0.80  0.01  0.02  0.02
20:53:51 B17F.0   B17F04900917  stand   85   NoReport stand              trk  1.00 OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
20:53:52 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   24    0.01  0.04  0.71  0.01  0.02  0.04
20:53:52 B17F.0   B17F04900917  walk    94   NoReport walk               trk  1.00 OpenFloor  2   24    0.00  0.02  0.85  0.00  0.01  0.02
20:53:53 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.80  0.01  0.02  0.02
20:53:53 B17F.0   B17F04900917  walk    87   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:54 B17F.0   B17F04900917  walk    87   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:54 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.83  0.00  0.01  0.02
20:53:55 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.04  0.73  0.01  0.02  0.04
20:53:55 B17F.0   B17F04900917  walk    51   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:56 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:53:56 B17F.0   B17F04900917  walk    63   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:57 B17F.0   B17F04900917  walk    40   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:53:57 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:53:58 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.04  0.73  0.01  0.02  0.04
20:53:58 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.04  0.70  0.00  0.02  0.03
20:53:59 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:53:59 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.05  0.37  0.01  0.03  0.04
20:54:00 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.04  0.72  0.01  0.02  0.04
20:54:00 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.03  0.16  0.01  0.03  0.02
20:54:01 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.67  0.01  0.03  0.04
20:54:01 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.00  0.02  0.07  0.01  0.02  0.01
20:54:02 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.78  0.01  0.03  0.02
20:54:02 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.01  0.05  0.01  0.01  0.01
20:54:03 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.70  0.01  0.03  0.04
20:54:03 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.01  0.01
20:54:04 B17F.1   B17F24900917  stand   132  NoReport stand              trk  1.00 Sit        2   0     0.00  0.03  0.80  0.01  0.02  0.02
20:54:04 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.04  0.01  0.00  0.01
20:54:05 B17F.1   B17F24900917  stand   105  NoReport stand              trk  1.00 Sit        2   0     0.00  0.02  0.83  0.00  0.01  0.02
20:54:05 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.04  0.01  0.00  0.01
20:54:06 B17F.1   B17F24900917  stand   110  NoReport stand              trk  1.00 Sit        2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:54:06 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.04  0.01  0.00  0.01
20:54:07 B17F.1   B17F24900917  stand   107  NoReport stand              trk  1.00 Sit        2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Sit        2   0     0.00  0.01  0.13  0.02  0.01  0.02
20:54:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   13    0.00  0.02  0.45  0.03  0.01  0.02
20:54:08 B17F.1   B17F24900917  stand   96   NoReport stand              trk  1.00 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
20:54:09 B17F.1   B17F24900917  stand   88   NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
20:54:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.02  0.68  0.02  0.01  0.02
20:54:10 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:10 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.79  0.01  0.01  0.02
20:54:11 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.58  0.01  0.03  0.06
20:54:11 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.83  0.01  0.01  0.02
20:54:12 B17F.1   B17F24900917  stand   83   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.04  0.75  0.01  0.03  0.02
20:54:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:54:13 B17F.1   B17F24900917  stand   85   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.82  0.01  0.02  0.02
20:54:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:14 B17F.1   B17F24900917  stand   111  NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.04  0.72  0.01  0.02  0.04
20:54:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:15 B17F.1   B17F24900917  stand   114  NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.05  0.67  0.01  0.03  0.04
20:54:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:16 B17F.1   B17F24900917  stand   116  NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.79  0.01  0.03  0.02
20:54:17 B17F.1   B17F24900917  stand   109  NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.83  0.00  0.02  0.02
20:54:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:18 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:54:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:19 B17F.1   B17F24900917  stand   100  NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:20 B17F.1   B17F24900917  stand   83   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:21 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:21 B17F.0   B17F04900917  stand   80   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:22 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.04  0.74  0.00  0.02  0.04
20:54:22 B17F.0   B17F04900917  stand   75   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:23 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.00  0.02  0.02
20:54:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:24 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:54:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:25 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.57  0.01  0.03  0.06
20:54:25 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:26 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.06  0.59  0.01  0.05  0.03
20:54:26 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:27 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.03  0.75  0.01  0.03  0.02
20:54:27 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:28 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.04  0.69  0.01  0.03  0.03
20:54:28 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:54:29 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   13    0.00  0.03  0.79  0.01  0.02  0.02
20:54:29 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   13    0.00  0.02  0.85  0.00  0.01  0.02
20:54:30 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
20:54:30 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.03  0.83  0.00  0.01  0.02
20:54:31 B17F.1   B17F24900917  stand   44   NoReport stand              trk  1.00 OpenFloor  2   15    0.01  0.04  0.73  0.01  0.02  0.04
20:54:31 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
20:54:32 B17F.1   B17F24900917  stand   68   NoReport stand              trk  1.00 OpenFloor  2   16    0.01  0.05  0.67  0.01  0.03  0.04
20:54:32 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   16    0.00  0.02  0.85  0.00  0.01  0.02
20:54:33 B17F.1   B17F24900917  stand   92   NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.03  0.78  0.01  0.03  0.02
20:54:33 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.03  0.81  0.00  0.02  0.02
20:54:34 B17F.1   B17F24900917  walk    104  NoReport walk               trk  1.00 OpenFloor  2   18    0.00  0.03  0.83  0.01  0.02  0.02
20:54:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.02  0.84  0.00  0.01  0.02
20:54:35 B17F.1   B17F24900917  walk    112  NoReport walk               trk  1.00 OpenFloor  2   19    0.00  0.02  0.84  0.00  0.01  0.02
20:54:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.02  0.84  0.00  0.01  0.02
20:54:36 B17F.1   B17F24900917  walk    103  NoReport walk               trk  1.00 OpenFloor  1   20    0.00  0.04  0.73  0.01  0.02  0.04
20:54:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   20    0.00  0.02  0.85  0.00  0.01  0.02
20:54:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   21    0.00  0.02  0.85  0.00  0.01  0.02
20:54:37 B17F.1   B17F24900917  walk    99   NoReport walk               trk  1.00 OpenFloor  1   21    0.01  0.07  0.51  0.01  0.05  0.05
20:54:38 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   22    0.00  0.04  0.71  0.01  0.04  0.02
20:54:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
20:54:39 B17F.1   B17F24900917  walk    121  NoReport walk               trk  1.00 OpenFloor  2   23    0.00  0.03  0.80  0.01  0.02  0.02
20:54:39 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
20:54:40 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   24    0.00  0.02  0.83  0.00  0.01  0.02
20:54:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   24    0.00  0.02  0.85  0.00  0.01  0.02
20:54:41 B17F.1   B17F24900917  walk    120  NoReport walk               trk  1.00 OpenFloor  2   25    0.00  0.02  0.84  0.00  0.01  0.02
20:54:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
20:54:42 B17F.1   B17F24900917  walk    118  NoReport walk               trk  1.00 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
20:54:42 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
20:54:43 B17F.1   B17F24900917  walk    110  NoReport walk               trk  1.00 OpenFloor  2   27    0.00  0.04  0.74  0.00  0.02  0.04
20:54:43 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
20:54:44 B17F.1   B17F24900917  walk    113  NoReport walk               trk  1.00 OpenFloor  2   28    0.00  0.03  0.81  0.00  0.02  0.02
20:54:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   28    0.00  0.02  0.85  0.00  0.01  0.02
20:54:45 B17F.1   B17F24900917  walk    115  NoReport walk               trk  1.00 OpenFloor  1   29    0.01  0.07  0.56  0.01  0.04  0.06
20:54:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
20:54:46 B17F.1   B17F24900917  walk    113  NoReport walk               trk  1.00 OpenFloor  2   30    0.00  0.04  0.74  0.01  0.04  0.02
20:54:46 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   30    0.00  0.02  0.85  0.00  0.01  0.02
20:54:47 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   31    0.01  0.07  0.52  0.02  0.05  0.05
20:54:47 B17F.0   B17F04900917  stand   83   NoReport stand              trk  1.00 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
20:54:48 B17F.1   B17F24900917  walk    109  NoReport walk               trk  1.00 OpenFloor  1   32    0.00  0.06  0.56  0.02  0.07  0.03
20:54:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
20:54:49 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  2   33    0.00  0.04  0.74  0.01  0.04  0.02
20:54:49 B17F.0   B17F04900917  stand   84   NoReport stand              trk  1.00 OpenFloor  2   33    0.00  0.02  0.85  0.00  0.01  0.02
20:54:50 B17F.1   B17F24900917  walk    110  NoReport walk               trk  1.00 OpenFloor  2   34    0.00  0.03  0.81  0.01  0.02  0.02
20:54:50 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   34    0.00  0.02  0.85  0.00  0.01  0.02
20:54:51 B17F.1   B17F24900917  walk    117  NoReport walk               trk  1.00 OpenFloor  1   35    0.00  0.04  0.72  0.01  0.02  0.04
20:54:51 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   35    0.00  0.02  0.85  0.00  0.01  0.02
20:54:52 B17F.1   B17F24900917  walk    116  NoReport walk               trk  1.00 OpenFloor  1   36    0.00  0.05  0.67  0.01  0.03  0.04
20:54:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   36    0.00  0.02  0.85  0.00  0.01  0.02
20:54:53 B17F.1   B17F24900917  walk    107  NoReport walk               trk  1.00 OpenFloor  1   37    0.01  0.08  0.48  0.02  0.06  0.05
20:54:53 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   37    0.00  0.02  0.85  0.00  0.01  0.02
20:54:54 B17F.1   B17F24900917  walk    111  NoReport walk               trk  1.00 OpenFloor  2   38    0.00  0.06  0.53  0.02  0.07  0.03
20:54:54 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   38    0.00  0.02  0.85  0.00  0.01  0.02
20:54:55 B17F.1   B17F24900917  walk    93   NoReport walk               trk  1.00 OpenFloor  2   39    0.00  0.04  0.72  0.02  0.04  0.02
20:54:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   39    0.00  0.02  0.85  0.00  0.01  0.02
20:54:56 B17F.1   B17F24900917  stand   105  NoReport stand              trk  1.00 OpenFloor  2   40    0.00  0.03  0.81  0.01  0.02  0.02
20:54:56 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   40    0.00  0.02  0.85  0.00  0.01  0.02
20:54:57 B17F.1   B17F24900917  stand   111  NoReport stand              trk  1.00 OpenFloor  2   41    0.00  0.02  0.84  0.00  0.01  0.02
20:54:57 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   41    0.00  0.02  0.85  0.00  0.01  0.02
20:54:58 B17F.1   B17F24900917  stand   109  NoReport stand              trk  1.00 OpenFloor  2   42    0.00  0.02  0.85  0.00  0.01  0.02
20:54:58 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   42    0.00  0.02  0.85  0.00  0.01  0.02
20:54:58 B17F.1   B17F24900917  stand   106  NoReport stand              trk  1.00 OpenFloor  2   43    0.00  0.02  0.85  0.00  0.01  0.02
20:54:58 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   43    0.00  0.02  0.85  0.00  0.01  0.02
20:54:59 B17F.1   B17F24900917  stand   105  NoReport stand              trk  1.00 OpenFloor  2   44    0.00  0.02  0.85  0.00  0.01  0.02
20:54:59 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   44    0.00  0.02  0.85  0.00  0.01  0.02
20:55:00 B17F.1   B17F24900917  stand   112  NoReport stand              trk  1.00 OpenFloor  2   45    0.00  0.04  0.74  0.00  0.02  0.04
20:55:00 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   45    0.00  0.02  0.85  0.00  0.01  0.02
20:55:01 B17F.1   B17F24900917  stand   104  NoReport stand              trk  1.00 OpenFloor  1   46    0.00  0.05  0.68  0.01  0.03  0.04
20:55:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   46    0.00  0.02  0.85  0.00  0.01  0.02
20:55:02 B17F.1   B17F24900917  stand   105  NoReport stand              trk  1.00 OpenFloor  2   47    0.00  0.03  0.79  0.01  0.02  0.02
20:55:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   47    0.00  0.02  0.85  0.00  0.01  0.02
20:55:03 B17F.1   B17F24900917  stand   93   NoReport stand              trk  1.00 OpenFloor  2   48    0.00  0.04  0.71  0.01  0.03  0.04
20:55:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   48    0.00  0.02  0.85  0.00  0.01  0.02
20:55:04 B17F.1   B17F24900917  stand   89   NoReport stand              trk  1.00 OpenFloor  2   49    0.00  0.03  0.80  0.01  0.02  0.02
20:55:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   49    0.00  0.02  0.85  0.00  0.01  0.02
20:55:05 B17F.1   B17F24900917  stand   99   NoReport stand              trk  1.00 OpenFloor  2   50    0.00  0.02  0.83  0.00  0.01  0.02
20:55:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   50    0.00  0.02  0.85  0.00  0.01  0.02
20:55:06 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   51    0.00  0.02  0.84  0.00  0.01  0.02
20:55:06 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   51    0.00  0.02  0.85  0.00  0.01  0.02
20:55:07 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   52    0.00  0.04  0.73  0.00  0.02  0.04
20:55:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   52    0.00  0.02  0.85  0.00  0.01  0.02
20:55:08 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   53    0.01  0.05  0.68  0.01  0.03  0.04
20:55:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   53    0.00  0.02  0.85  0.00  0.01  0.02
20:55:09 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   54    0.00  0.03  0.79  0.01  0.02  0.02
20:55:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   54    0.00  0.02  0.85  0.00  0.01  0.02
20:55:10 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   55    0.00  0.03  0.83  0.00  0.02  0.02
20:55:10 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   55    0.00  0.02  0.85  0.00  0.01  0.02
20:55:11 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   56    0.01  0.06  0.57  0.01  0.03  0.06
20:55:11 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   56    0.00  0.02  0.85  0.00  0.01  0.02
20:55:12 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   57    0.01  0.04  0.74  0.01  0.03  0.02
20:55:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   57    0.00  0.02  0.85  0.00  0.01  0.02
20:55:13 B17F.1   B17F24900917  stand   54   NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:55:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:14 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:55:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:15 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.84  0.00  0.01  0.02
20:55:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:16 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.58  0.01  0.03  0.06
20:55:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:17 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.01  0.04  0.74  0.01  0.03  0.02
20:55:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:18 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:55:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:19 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.06  0.56  0.01  0.03  0.06
20:55:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
20:55:20 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.01  0.04  0.74  0.01  0.04  0.02
20:55:21 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
20:55:21 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.03  0.81  0.01  0.02  0.02
20:55:22 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.72  0.01  0.02  0.04
20:55:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:23 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.03  0.80  0.01  0.02  0.02
20:55:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.02  0.85  0.00  0.01  0.02
20:55:24 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.02  0.83  0.00  0.01  0.02
20:55:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.02  0.85  0.00  0.01  0.02
20:55:25 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.02  0.84  0.00  0.01  0.02
20:55:25 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
20:55:26 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.04  0.73  0.00  0.02  0.04
20:55:26 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:27 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   21    0.00  0.03  0.81  0.01  0.02  0.02
20:55:27 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
20:55:28 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.02  0.84  0.00  0.01  0.02
20:55:28 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
20:55:29 B17F.1   B17F24900917  stand   61   NoReport stand              trk  1.00 OpenFloor  2   23    0.01  0.04  0.73  0.01  0.02  0.04
20:55:29 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
20:55:30 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.03  0.81  0.01  0.02  0.02
20:55:30 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:55:31 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.02  0.85  0.00  0.01  0.02
20:55:31 B17F.1   B17F24900917  stand   44   NoReport stand              trk  1.00 OpenFloor  2   14    0.00  0.02  0.83  0.00  0.01  0.02
20:55:32 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.02  0.85  0.00  0.01  0.02
20:55:32 B17F.1   B17F24900917  stand   52   NoReport stand              trk  1.00 OpenFloor  2   15    0.00  0.02  0.84  0.00  0.01  0.02
20:55:33 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   16    0.00  0.03  0.81  0.00  0.02  0.02
20:55:33 B17F.1   B17F24900917  stand   40   NoReport stand              trk  1.00 OpenFloor  2   16    0.00  0.03  0.81  0.00  0.02  0.02
20:55:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.00  0.02  0.84  0.00  0.01  0.02
20:55:34 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   17    0.01  0.04  0.72  0.01  0.02  0.04
20:55:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.02  0.84  0.00  0.01  0.02
20:55:35 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   18    0.00  0.03  0.80  0.01  0.02  0.02
20:55:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.02  0.85  0.00  0.01  0.02
20:55:36 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   19    0.00  0.02  0.83  0.00  0.01  0.02
20:55:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   20    0.00  0.02  0.85  0.00  0.01  0.02
20:55:37 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   20    0.01  0.04  0.73  0.01  0.02  0.04
20:55:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   21    0.00  0.02  0.85  0.00  0.01  0.02
20:55:38 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   21    0.00  0.03  0.81  0.01  0.02  0.02
20:55:39 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.02  0.83  0.00  0.01  0.02
20:55:39 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.02  0.85  0.00  0.01  0.02
20:55:40 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   23    0.00  0.02  0.84  0.00  0.01  0.02
20:55:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   23    0.00  0.02  0.85  0.00  0.01  0.02
20:55:41 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   24    0.00  0.04  0.73  0.00  0.02  0.04
20:55:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   24    0.00  0.02  0.85  0.00  0.01  0.02
20:55:42 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   25    0.00  0.02  0.85  0.00  0.01  0.02
20:55:42 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   25    0.00  0.03  0.81  0.00  0.02  0.02
20:55:43 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   26    0.00  0.02  0.84  0.00  0.01  0.02
20:55:43 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   26    0.00  0.02  0.85  0.00  0.01  0.02
20:55:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   27    0.00  0.02  0.85  0.00  0.01  0.02
20:55:44 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   27    0.00  0.02  0.84  0.00  0.01  0.02
20:55:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   28    0.00  0.02  0.85  0.00  0.01  0.02
20:55:45 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   28    0.00  0.02  0.85  0.00  0.01  0.02
20:55:46 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   29    0.00  0.02  0.85  0.00  0.01  0.02
20:55:46 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   29    0.00  0.02  0.85  0.00  0.01  0.02
20:55:47 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   30    0.00  0.02  0.85  0.00  0.01  0.02
20:55:47 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   30    0.00  0.02  0.85  0.00  0.01  0.02
20:55:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
20:55:48 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   31    0.00  0.04  0.74  0.00  0.02  0.04
20:55:49 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   32    0.00  0.02  0.85  0.00  0.01  0.02
20:55:49 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   32    0.00  0.03  0.81  0.00  0.02  0.02
20:55:50 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
20:55:50 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   33    0.01  0.04  0.72  0.01  0.02  0.04
20:55:51 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   34    0.01  0.05  0.67  0.01  0.03  0.04
20:55:51 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   34    0.00  0.02  0.85  0.00  0.01  0.02
20:55:52 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   35    0.01  0.05  0.64  0.01  0.04  0.03
20:55:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   35    0.00  0.02  0.85  0.00  0.01  0.02
20:55:53 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   36    0.00  0.03  0.77  0.01  0.03  0.02
20:55:53 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   36    0.00  0.02  0.85  0.00  0.01  0.02
20:55:54 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   37    0.00  0.02  0.85  0.00  0.01  0.02
20:55:54 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   37    0.01  0.04  0.70  0.01  0.03  0.03
20:55:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   38    0.00  0.02  0.85  0.00  0.01  0.02
20:55:55 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   38    0.00  0.03  0.80  0.01  0.02  0.02
20:55:56 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   39    0.00  0.02  0.85  0.00  0.01  0.02
20:55:56 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   39    0.01  0.04  0.71  0.01  0.02  0.04
20:55:57 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   40    0.00  0.02  0.85  0.00  0.01  0.02
20:55:57 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   40    0.01  0.05  0.67  0.01  0.04  0.04
20:55:58 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   41    0.00  0.02  0.85  0.00  0.01  0.02
20:55:58 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   41    0.00  0.03  0.78  0.01  0.03  0.02
20:55:59 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   42    0.00  0.02  0.85  0.00  0.01  0.02
20:55:59 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   42    0.00  0.03  0.83  0.01  0.02  0.02
20:56:00 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   43    0.00  0.02  0.85  0.00  0.01  0.02
20:56:00 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   43    0.00  0.02  0.84  0.00  0.01  0.02
20:56:01 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   44    0.00  0.02  0.85  0.00  0.01  0.02
20:56:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   44    0.00  0.02  0.85  0.00  0.01  0.02
20:56:02 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   381   0.01  0.06  0.58  0.01  0.03  0.06
20:56:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   381   0.00  0.02  0.85  0.00  0.01  0.02
20:56:03 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   46    0.01  0.04  0.74  0.01  0.03  0.02
20:56:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   46    0.00  0.02  0.85  0.00  0.01  0.02
20:56:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   383   0.00  0.02  0.85  0.00  0.01  0.02
20:56:04 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   383   0.01  0.05  0.68  0.01  0.03  0.03
20:56:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   48    0.00  0.02  0.85  0.00  0.01  0.02
20:56:05 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   48    0.00  0.03  0.79  0.01  0.02  0.02
20:56:06 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   385   0.00  0.02  0.85  0.00  0.01  0.02
20:56:06 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   385   0.01  0.04  0.71  0.01  0.03  0.04
20:56:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   50    0.00  0.02  0.85  0.00  0.01  0.02
20:56:07 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   50    0.01  0.07  0.50  0.02  0.05  0.05
20:56:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   51    0.00  0.02  0.85  0.00  0.01  0.02
20:56:08 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   51    0.01  0.04  0.70  0.01  0.04  0.02
20:56:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   52    0.00  0.02  0.85  0.00  0.01  0.02
20:56:09 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   52    0.00  0.03  0.80  0.01  0.02  0.02
20:56:10 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   53    0.00  0.02  0.85  0.00  0.01  0.02
20:56:10 B17F.1   B17F24900917  stand   73   NoReport stand              trk  1.00 OpenFloor  2   53    0.00  0.02  0.83  0.00  0.01  0.02
20:56:11 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   54    0.00  0.02  0.85  0.00  0.01  0.02
20:56:11 B17F.1   B17F24900917  walk    73   NoReport walk               trk  1.00 OpenFloor  1   54    0.01  0.04  0.73  0.01  0.02  0.04
20:56:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   55    0.00  0.02  0.85  0.00  0.01  0.02
20:56:12 B17F.1   B17F24900917  walk    136  NoReport walk               trk  1.00 OpenFloor  1   55    0.00  0.05  0.68  0.01  0.03  0.04
20:56:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   56    0.00  0.02  0.85  0.00  0.01  0.02
20:56:13 B17F.1   B17F24900917  walk    132  NoReport walk               trk  1.00 OpenFloor  2   56    0.00  0.03  0.79  0.01  0.02  0.02
20:56:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   57    0.00  0.02  0.85  0.00  0.01  0.02
20:56:14 B17F.1   B17F24900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   57    0.00  0.04  0.71  0.01  0.03  0.04
20:56:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   58    0.00  0.02  0.85  0.00  0.01  0.02
20:56:15 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   58    0.00  0.03  0.80  0.01  0.02  0.02
20:56:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   59    0.00  0.02  0.85  0.00  0.01  0.02
20:56:16 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   59    0.00  0.02  0.83  0.00  0.01  0.02
20:56:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   60    0.00  0.02  0.85  0.00  0.01  0.02
20:56:17 B17F.1   B17F24900917  stand   127  NoReport stand              trk  1.00 OpenFloor  2   60    0.00  0.02  0.84  0.00  0.01  0.02
20:56:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   61    0.00  0.02  0.85  0.00  0.01  0.02
20:56:18 B17F.1   B17F24900917  stand   101  NoReport stand              trk  1.00 OpenFloor  2   61    0.00  0.02  0.85  0.00  0.01  0.02
20:56:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   62    0.00  0.02  0.85  0.00  0.01  0.02
20:56:19 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   62    0.00  0.02  0.85  0.00  0.01  0.02
20:56:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   63    0.00  0.02  0.85  0.00  0.01  0.02
20:56:20 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   63    0.00  0.02  0.85  0.00  0.01  0.02
20:56:21 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   64    0.00  0.01  0.92  0.00  0.01  0.01
20:56:21 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   64    0.00  0.02  0.85  0.00  0.01  0.02
20:56:22 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   65    0.00  0.02  0.87  0.00  0.01  0.02
20:56:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   65    0.00  0.02  0.85  0.00  0.01  0.02
20:56:23 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   66    0.00  0.01  0.96  0.00  0.00  0.01
20:56:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   66    0.00  0.02  0.85  0.00  0.01  0.02
20:56:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:56:24 B17F.1   B17F24900917  stand   0    NoReport stand              trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
20:56:24 B17F.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:56:25 B17F.E   -             -       0    NoReport np=1               room -    OpenFloor  2   0     0.00  0.02  0.85  0.00  0.01  0.02
20:56:25 B17F.0   B17F04900917  walk    74   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.04  0.74  0.00  0.02  0.04
20:56:26 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.68  0.01  0.03  0.04
20:56:27 B17F.0   B17F04900917  walk    88   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.65  0.01  0.04  0.03
20:56:28 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.64  0.01  0.05  0.03
20:56:29 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.63  0.02  0.05  0.03
20:56:30 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.62  0.02  0.05  0.03
20:56:31 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.62  0.02  0.05  0.03
20:56:32 B17F.0   B17F04900917  stand   70   NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:33 B17F.0   B17F04900917  stand   86   NoReport stand              trk  1.00 BlindOpen  1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:56:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:35 B17F.0   B17F04900917  stand   77   NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:39 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:42 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:43 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 BlindOpen  1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:46 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:47 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:48 B17F.0   B17F04900917  stand   84   NoReport stand              trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:56:49 B17F.0   B17F04900917  stand   90   NoReport stand              trk  1.00 Empty      1   24    0.00  0.05  0.61  0.02  0.05  0.03
20:56:50 B17F.0   B17F04900917  stand   78   NoReport stand              trk  1.00 Empty      1   25    0.01  0.05  0.61  0.02  0.05  0.03
20:56:51 B17F.0   B17F04900917  stand   77   NoReport stand              trk  1.00 Empty      1   26    0.01  0.05  0.61  0.02  0.05  0.03
20:56:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:53 B17F.0   B17F04900917  stand   75   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:56:54 B17F.0   B17F04900917  stand   104  NoReport stand              trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:56:55 B17F.0   B17F04900917  walk    92   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:56:56 B17F.0   B17F04900917  walk    50   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:56:57 B17F.0   B17F04900917  walk    38   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:56:58 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:56:59 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:57:00 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.07  0.41  0.03  0.07  0.04
20:57:01 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.07  0.28  0.03  0.09  0.03
20:57:02 B17F.E   -             -       0    NoReport Walking(rdr)       room -    Empty      1   0     0.10  0.10  0.15  0.13  0.19  0.02
20:57:02 B17F.0   B17F04900917  sitgnd  0    NoReport sitgnd             trk  1.00 Empty      1   0     0.01  0.08  0.24  0.05  0.11  0.03
20:57:02 B17F.0   B17F04900917  sitgnd  0    NoReport sitgnd             trk  1.00 Empty      1   0     0.02  0.08  0.22  0.06  0.13  0.03
20:57:03 B17F.0   B17F04900917  sitgnd  0    NoReport sitgnd             trk  1.00 Empty      1   0     0.02  0.09  0.21  0.07  0.14  0.03
20:57:04 B17F.0   B17F04900917  sitgnd  61   NoReport sitgnd             trk  1.00 Empty      1   0     0.02  0.09  0.20  0.08  0.15  0.03
20:57:05 B17F.0   B17F04900917  sit     83   NoReport sit                trk  1.00 Empty      1   0     0.01  0.07  0.29  0.07  0.12  0.02
20:57:05 B17F.0   B17F04900917  sit     83   NoReport sit                trk  1.00 Empty      1   0     0.00  0.06  0.34  0.05  0.09  0.02
20:57:06 B17F.0   B17F04900917  walk    94   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.44  0.05  0.08  0.03
20:57:07 B17F.0   B17F04900917  walk    93   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.51  0.04  0.07  0.03
20:57:08 B17F.0   B17F04900917  walk    66   NoReport walk               trk  1.00 Empty      1   0     0.00  0.06  0.55  0.04  0.07  0.03
20:57:09 B17F.0   B17F04900917  walk    61   NoReport walk               trk  1.00 Empty      1   0     0.00  0.06  0.58  0.03  0.06  0.03
20:57:10 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.07  0.38  0.03  0.07  0.04
20:57:11 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.00  0.06  0.47  0.04  0.08  0.03
20:57:12 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.07  0.31  0.04  0.09  0.03
20:57:13 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.06  0.22  0.04  0.09  0.03
20:57:14 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.05  0.16  0.05  0.08  0.02
20:57:15 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.04  0.13  0.04  0.07  0.02
20:57:16 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.04  0.11  0.04  0.06  0.02
20:57:17 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.03  0.10  0.04  0.04  0.02
20:57:18 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.03  0.10  0.04  0.04  0.02
20:57:19 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.02  0.09  0.04  0.03  0.02
20:57:20 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.02  0.09  0.04  0.03  0.02
20:57:21 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.02  0.09  0.03  0.03  0.02
20:57:22 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.02  0.09  0.03  0.02  0.02
20:57:23 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.02  0.09  0.03  0.02  0.02
20:57:24 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.02  0.09  0.03  0.02  0.02
20:57:25 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.02  0.09  0.03  0.02  0.02
20:57:26 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.02  0.09  0.03  0.02  0.02
20:57:27 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Empty      1   0     0.01  0.02  0.09  0.03  0.02  0.02
20:57:28 B17F.0   B17F04900917  stand   71   NoReport stand              trk  1.00 Empty      1   0     0.01  0.03  0.25  0.05  0.03  0.03
20:57:29 B17F.0   B17F04900917  stand   74   NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.38  0.05  0.04  0.03
20:57:30 B17F.0   B17F04900917  stand   58   NoReport stand              trk  1.00 Empty      1   0     0.01  0.04  0.47  0.04  0.05  0.03
20:57:31 B17F.0   B17F04900917  stand   41   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.56  0.03  0.05  0.03
20:57:32 B17F.0   B17F04900917  stand   60   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.58  0.03  0.05  0.03
20:57:33 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.59  0.02  0.05  0.03
20:57:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.60  0.02  0.05  0.03
20:57:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.60  0.02  0.05  0.03
20:57:36 B17F.0   B17F04900917  stand   66   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:57:37 B17F.0   B17F04900917  stand   73   NoReport stand              trk  1.00 Empty      1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:57:38 B17F.0   B17F04900917  stand   87   NoReport stand              trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:57:39 B17F.0   B17F04900917  stand   88   NoReport stand              trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:57:40 B17F.0   B17F04900917  stand   83   NoReport stand              trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:57:41 B17F.0   B17F04900917  walk    75   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:57:42 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:57:43 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:57:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Empty      1   0     0.00  0.05  0.61  0.02  0.05  0.03
20:57:45 B17F.0   B17F04900917  sit     21   NoReport sit                trk  1.00 Empty      1   0     0.01  0.07  0.41  0.03  0.07  0.04
20:57:46 B17F.0   B17F04900917  sit     65   NoReport sit                trk  1.00 Empty      1   0     0.01  0.07  0.28  0.03  0.09  0.03
20:57:47 B17F.0   B17F04900917  sit     99   NoReport sit                trk  1.00 Empty      1   0     0.00  0.05  0.33  0.03  0.07  0.02
20:57:48 B17F.0   B17F04900917  walk    80   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.44  0.04  0.07  0.03
20:57:49 B17F.0   B17F04900917  walk    91   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.51  0.03  0.06  0.03
20:57:50 B17F.0   B17F04900917  walk    53   NoReport walk               trk  1.00 Empty      1   0     0.00  0.05  0.55  0.03  0.06  0.03
20:57:51 B17F.0   B17F04900917  walk    59   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.58  0.03  0.06  0.03
20:57:52 B17F.0   B17F04900917  sit     78   NoReport sit                trk  1.00 Fallen     1   0     0.01  0.07  0.38  0.03  0.07  0.04
20:57:53 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.07  0.26  0.04  0.09  0.03
20:57:54 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.06  0.18  0.04  0.08  0.03
20:57:55 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.05  0.14  0.04  0.07  0.02
20:57:56 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.04  0.12  0.04  0.06  0.02
20:57:57 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.03  0.11  0.04  0.05  0.02
20:57:58 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.03  0.10  0.04  0.04  0.02
20:57:59 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.04  0.03  0.02
20:58:00 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.04  0.03  0.02
20:58:01 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.03  0.02
20:58:02 B17F.0   B17F04900917  sit     64   NoReport sit                trk  1.00 Fallen     1   0     0.01  0.02  0.09  0.03  0.02  0.02
20:58:03 B17F.0   B17F04900917  sit     85   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.17  0.03  0.02  0.02
20:58:04 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.12  0.03  0.02  0.02
20:58:05 B17F.0   B17F04900917  sit     70   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.10  0.03  0.02  0.02
20:58:06 B17F.0   B17F04900917  sit     0    NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.09  0.03  0.02  0.02
20:58:07 B17F.0   B17F04900917  sit     74   NoReport sit                trk  1.00 Fallen     1   0     0.00  0.02  0.09  0.03  0.02  0.02
20:58:08 B17F.0   B17F04900917  stand   62   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.03  0.25  0.05  0.03  0.03
20:58:09 B17F.0   B17F04900917  stand   32   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.04  0.38  0.05  0.04  0.03
20:58:10 B17F.0   B17F04900917  stand   64   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.04  0.47  0.04  0.05  0.03
20:58:11 B17F.0   B17F04900917  stand   61   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.53  0.04  0.05  0.03
20:58:12 B17F.0   B17F04900917  stand   62   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.56  0.03  0.05  0.03
20:58:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.58  0.03  0.05  0.03
20:58:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.59  0.02  0.05  0.03
20:58:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.60  0.02  0.05  0.03
20:58:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.60  0.02  0.05  0.03
20:58:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:21 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:25 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:26 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:27 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:28 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:29 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:30 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:31 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:32 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:33 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:39 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:42 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:43 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:58:44 B17F.0   B17F04900917  stand   65   NoReport stand              trk  1.00 Fallen     1   27    0.01  0.05  0.61  0.02  0.05  0.03
20:58:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   28    0.01  0.05  0.61  0.02  0.05  0.03
20:58:46 B17F.0   B17F04900917  stand   91   NoReport stand              trk  1.00 Fallen     1   29    0.00  0.05  0.61  0.02  0.05  0.03
20:58:47 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   30    0.01  0.05  0.61  0.02  0.05  0.03
20:58:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   31    0.01  0.05  0.61  0.02  0.05  0.03
20:58:49 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   32    0.01  0.05  0.61  0.02  0.05  0.03
20:58:50 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   33    0.01  0.05  0.61  0.02  0.05  0.03
20:58:51 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   34    0.01  0.05  0.61  0.02  0.05  0.03
20:58:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   35    0.01  0.05  0.61  0.02  0.05  0.03
20:58:53 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   36    0.01  0.05  0.61  0.02  0.05  0.03
20:58:54 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   37    0.01  0.05  0.61  0.02  0.05  0.03
20:58:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   38    0.01  0.05  0.61  0.02  0.05  0.03
20:58:56 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   39    0.01  0.05  0.61  0.02  0.05  0.03
20:58:57 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   40    0.01  0.05  0.61  0.02  0.05  0.03
20:58:58 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   41    0.01  0.05  0.61  0.02  0.05  0.03
20:58:59 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   42    0.01  0.05  0.61  0.02  0.05  0.03
20:59:00 B17F.0   B17F04900917  stand   61   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:06 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:10 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:11 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   28    0.01  0.05  0.61  0.02  0.05  0.03
20:59:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   29    0.01  0.05  0.61  0.02  0.05  0.03
20:59:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   30    0.01  0.05  0.61  0.02  0.05  0.03
20:59:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   31    0.01  0.05  0.61  0.02  0.05  0.03
20:59:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   32    0.01  0.05  0.61  0.02  0.05  0.03
20:59:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   33    0.01  0.05  0.61  0.02  0.05  0.03
20:59:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   34    0.01  0.05  0.61  0.02  0.05  0.03
20:59:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   35    0.01  0.05  0.61  0.02  0.05  0.03
20:59:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   36    0.01  0.05  0.61  0.02  0.05  0.03
20:59:21 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:25 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:26 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:27 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:28 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:29 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:30 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:31 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:32 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:33 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:39 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:42 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:43 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:46 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:47 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
20:59:49 B17F.0   B17F04900917  stand   65   NoReport stand              trk  1.00 Fallen     1   27    0.01  0.05  0.61  0.02  0.05  0.03
20:59:50 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   28    0.01  0.05  0.61  0.02  0.05  0.03
20:59:51 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   29    0.01  0.05  0.61  0.02  0.05  0.03
20:59:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   30    0.01  0.05  0.61  0.02  0.05  0.03
20:59:53 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   31    0.01  0.05  0.61  0.02  0.05  0.03
20:59:54 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   32    0.01  0.05  0.61  0.02  0.05  0.03
20:59:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   33    0.01  0.05  0.61  0.02  0.05  0.03
20:59:56 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   34    0.01  0.05  0.61  0.02  0.05  0.03
20:59:57 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   35    0.01  0.05  0.61  0.02  0.05  0.03
20:59:58 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   36    0.01  0.05  0.61  0.02  0.05  0.03
20:59:59 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   37    0.01  0.05  0.61  0.02  0.05  0.03
21:00:00 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   38    0.01  0.05  0.61  0.02  0.05  0.03
21:00:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   39    0.01  0.05  0.61  0.02  0.05  0.03
21:00:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   40    0.01  0.05  0.61  0.02  0.05  0.03
21:00:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   41    0.01  0.05  0.61  0.02  0.05  0.03
21:00:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   42    0.01  0.05  0.61  0.02  0.05  0.03
21:00:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   43    0.01  0.05  0.61  0.02  0.05  0.03
21:00:06 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   44    0.01  0.05  0.61  0.02  0.05  0.03
21:00:07 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   45    0.01  0.05  0.61  0.02  0.05  0.03
21:00:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   46    0.01  0.05  0.61  0.02  0.05  0.03
21:00:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   47    0.01  0.05  0.61  0.02  0.05  0.03
21:00:10 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   48    0.01  0.05  0.61  0.02  0.05  0.03
21:00:11 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   49    0.01  0.05  0.61  0.02  0.05  0.03
21:00:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   50    0.01  0.05  0.61  0.02  0.05  0.03
21:00:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   51    0.01  0.05  0.61  0.02  0.05  0.03
21:00:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   52    0.01  0.05  0.61  0.02  0.05  0.03
21:00:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   53    0.01  0.05  0.61  0.02  0.05  0.03
21:00:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   54    0.01  0.05  0.61  0.02  0.05  0.03
21:00:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   55    0.01  0.05  0.61  0.02  0.05  0.03
21:00:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   56    0.01  0.05  0.61  0.02  0.05  0.03
21:00:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   57    0.01  0.05  0.61  0.02  0.05  0.03
21:00:20 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   58    0.01  0.05  0.61  0.02  0.05  0.03
21:00:21 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   59    0.01  0.05  0.61  0.02  0.05  0.03
21:00:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   60    0.01  0.05  0.61  0.02  0.05  0.03
21:00:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   61    0.01  0.05  0.61  0.02  0.05  0.03
21:00:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   62    0.01  0.05  0.61  0.02  0.05  0.03
21:00:25 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   63    0.01  0.05  0.61  0.02  0.05  0.03
21:00:26 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   64    0.01  0.05  0.61  0.02  0.05  0.03
21:00:27 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   65    0.01  0.05  0.61  0.02  0.05  0.03
21:00:28 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   66    0.01  0.05  0.61  0.02  0.05  0.03
21:00:29 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   67    0.01  0.05  0.61  0.02  0.05  0.03
21:00:30 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   68    0.01  0.05  0.61  0.02  0.05  0.03
21:00:31 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   69    0.01  0.05  0.61  0.02  0.05  0.03
21:00:32 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   70    0.01  0.05  0.61  0.02  0.05  0.03
21:00:33 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   71    0.01  0.05  0.61  0.02  0.05  0.03
21:00:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   72    0.01  0.05  0.61  0.02  0.05  0.03
21:00:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   73    0.01  0.05  0.61  0.02  0.05  0.03
21:00:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   74    0.01  0.05  0.61  0.02  0.05  0.03
21:00:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   75    0.01  0.05  0.61  0.02  0.05  0.03
21:00:38 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   76    0.01  0.05  0.61  0.02  0.05  0.03
21:00:39 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   77    0.01  0.05  0.61  0.02  0.05  0.03
21:00:40 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   78    0.01  0.05  0.61  0.02  0.05  0.03
21:00:41 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   79    0.01  0.05  0.61  0.02  0.05  0.03
21:00:42 B17F.0   B17F04900917  stand   81   NoReport stand              trk  1.00 Fallen     1   80    0.00  0.05  0.61  0.02  0.05  0.03
21:00:43 B17F.0   B17F04900917  stand   73   NoReport stand              trk  1.00 Fallen     1   81    0.01  0.05  0.61  0.02  0.05  0.03
21:00:44 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   82    0.01  0.05  0.61  0.02  0.05  0.03
21:00:45 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   83    0.01  0.05  0.61  0.02  0.05  0.03
21:00:46 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   84    0.01  0.05  0.61  0.02  0.05  0.03
21:00:47 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   85    0.01  0.05  0.61  0.02  0.05  0.03
21:00:48 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   86    0.01  0.05  0.61  0.02  0.05  0.03
21:00:49 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   87    0.01  0.05  0.61  0.02  0.05  0.03
21:00:50 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   88    0.01  0.05  0.61  0.02  0.05  0.03
21:00:51 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   89    0.01  0.05  0.61  0.02  0.05  0.03
21:00:52 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   90    0.01  0.05  0.61  0.02  0.05  0.03
21:00:53 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   91    0.01  0.05  0.61  0.02  0.05  0.03
21:00:54 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   92    0.01  0.05  0.61  0.02  0.05  0.03
21:00:55 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   93    0.01  0.05  0.61  0.02  0.05  0.03
21:00:56 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   94    0.01  0.05  0.61  0.02  0.05  0.03
21:00:57 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   95    0.01  0.05  0.61  0.02  0.05  0.03
21:00:58 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   96    0.01  0.05  0.61  0.02  0.05  0.03
21:00:59 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   97    0.01  0.05  0.61  0.02  0.05  0.03
21:01:00 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   98    0.01  0.05  0.61  0.02  0.05  0.03
21:01:01 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   99    0.01  0.05  0.61  0.02  0.05  0.03
21:01:02 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   100   0.01  0.05  0.61  0.02  0.05  0.03
21:01:03 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   101   0.01  0.05  0.61  0.02  0.05  0.03
21:01:04 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   102   0.01  0.05  0.61  0.02  0.05  0.03
21:01:05 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   103   0.01  0.05  0.61  0.02  0.05  0.03
21:01:06 B17F.0   B17F04900917  stand   83   NoReport stand              trk  1.00 Fallen     1   104   0.00  0.05  0.61  0.02  0.05  0.03
21:01:07 B17F.0   B17F04900917  stand   67   NoReport stand              trk  1.00 Fallen     1   105   0.01  0.05  0.61  0.02  0.05  0.03
21:01:08 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   106   0.01  0.05  0.61  0.02  0.05  0.03
21:01:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   107   0.01  0.05  0.61  0.02  0.05  0.03
21:01:09 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   108   0.01  0.05  0.61  0.02  0.05  0.03
21:01:10 B17F.0   B17F04900917  stand   84   NoReport stand              trk  1.00 Fallen     1   109   0.00  0.05  0.61  0.02  0.05  0.03
21:01:11 B17F.0   B17F04900917  stand   64   NoReport stand              trk  1.00 Fallen     1   110   0.01  0.05  0.61  0.02  0.05  0.03
21:01:12 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   111   0.01  0.05  0.61  0.02  0.05  0.03
21:01:13 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   112   0.01  0.05  0.61  0.02  0.05  0.03
21:01:14 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   113   0.01  0.05  0.61  0.02  0.05  0.03
21:01:15 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   114   0.01  0.05  0.61  0.02  0.05  0.03
21:01:16 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   115   0.01  0.05  0.61  0.02  0.05  0.03
21:01:17 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   116   0.01  0.05  0.61  0.02  0.05  0.03
21:01:18 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   117   0.01  0.05  0.61  0.02  0.05  0.03
21:01:19 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   118   0.01  0.05  0.61  0.02  0.05  0.03
21:01:20 B17F.0   B17F04900917  stand   78   NoReport stand              trk  1.00 Fallen     1   119   0.01  0.05  0.61  0.02  0.05  0.03
21:01:21 B17F.0   B17F04900917  stand   78   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
21:01:22 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
21:01:23 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
21:01:24 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
21:01:25 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
21:01:26 B17F.0   B17F04900917  stand   73   NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.61  0.02  0.05  0.03
21:01:27 B17F.0   B17F04900917  stand   81   NoReport stand              trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
21:01:28 B17F.0   B17F04900917  walk    94   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
21:01:29 B17F.0   B17F04900917  walk    93   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
21:01:30 B17F.0   B17F04900917  walk    88   NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.61  0.02  0.05  0.03
21:01:31 B17F.0   B17F04900917  walk    123  NoReport walk               trk  1.00 Fallen     1   0     0.00  0.03  0.76  0.01  0.03  0.02
21:01:32 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.04  0.69  0.01  0.03  0.03
21:01:33 B17F.0   B17F04900917  walk    0    NoReport walk               trk  1.00 Fallen     1   0     0.00  0.05  0.66  0.01  0.04  0.03
21:01:34 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.05  0.64  0.01  0.04  0.03
21:01:35 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.63  0.02  0.05  0.03
21:01:36 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.01  0.05  0.62  0.02  0.05  0.03
21:01:37 B17F.0   B17F04900917  stand   0    NoReport stand              trk  1.00 Fallen     1   0     0.00  0.03  0.76  0.01  0.03  0.02
21:01:37 B17F.E   -             -       0    NoReport ExitRoom(rdr)      room -    Fallen     1   0     0.25  0.09  0.13  0.11  0.16  0.01
21:01:38 B17F.E   -             -       0    NoReport np=0  ★0           room -    Fallen     1   0     0.25  0.09  0.13  0.11  0.16  0.01
21:01:38 B17F.88  -             88      -    NoReport no-target(88)      room -    Fallen     1   0     0.25  0.09  0.13  0.11  0.16  0.01
21:01:39 B17F.88  -             88      -    NoReport no-target(88)      room -    Fallen     0   0     0.25  0.09  0.13  0.11  0.16  0.01
21:01:40 B17F.88  -             88      -    NoReport no-target(88)      room -    Fallen     0   0     0.25  0.09  0.13  0.11  0.16  0.01
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
20:49:00.917 B17F.2    sit    8    -50    40     0     80        
20:49:00.917 B17F.0    stand  8    -70    200    0     80   161  
20:49:00.917 B17F.1    walk   255  130    70     97    80   238  
20:49:01.918 B17F.0    stand  8    -70    200    0     80   238  
20:49:01.918 B17F.1    walk   255  80     80     74    80   192  
20:49:01.918 B17F.2    sit    8    -50    40     0     80   136  
20:49:02.919 B17F.2    sit    8    -50    40     0     80   0    
20:49:02.919 B17F.0    stand  8    -70    200    0     80   161  
20:49:02.919 B17F.1    walk   255  80     80     77    80   192  
20:49:03.921 B17F.0    stand  8    -70    200    0     80   192  
20:49:03.921 B17F.2    sit    8    -50    40     0     80   161  
20:49:03.921 B17F.1    walk   255  70     70     72    80   123  
20:49:04.811 B17F.2    sit    8    -50    40     0     80   123  
20:49:04.811 B17F.0    stand  8    -70    200    0     80   161  
20:49:04.811 B17F.1    walk   255  80     80     70    80   192  
20:49:05.818 B17F.0    stand  8    -70    200    0     80   192  
20:49:05.818 B17F.2    sit    8    -50    40     0     80   161  
20:49:05.818 B17F.1    stand  255  70     70     0     80   123  
20:49:06.814 B17F.0    stand  8    -70    200    0     80   191  
20:49:06.814 B17F.2    sit    8    -50    40     0     80   161  
20:49:06.814 B17F.1    stand  255  80     80     78    80   136  
20:49:07.816 B17F.2    sit    8    -50    40     0     80   136  
20:49:07.816 B17F.1    stand  255  70     70     71    80   123  
20:49:07.816 B17F.0    stand  8    -70    200    0     80   191  
20:49:08.816 B17F.1    stand  255  80     80     76    80   192  
20:49:08.816 B17F.2    sit    8    -50    40     0     80   136  
20:49:08.816 B17F.0    stand  8    -70    200    0     80   161  
20:49:09.824 B17F.0    stand  8    -70    200    0     80   0    
20:49:09.824 B17F.2    sit    8    -50    40     0     80   161  
20:49:09.824 B17F.1    stand  255  80     90     0     80   139  
20:49:10.837 B17F.0    stand  8    -70    200    0     80   186  
20:49:10.837 B17F.2    sit    8    -50    40     0     80   161  
20:49:10.837 B17F.1    stand  255  70     90     0     80   130  
20:49:11.824 B17F.0    stand  8    -70    200    0     80   178  
20:49:11.824 B17F.2    sit    8    -50    40     0     80   161  
20:49:11.824 B17F.1    stand  255  70     90     99    80   130  
20:49:12.827 B17F.2    sit    8    -50    40     0     80   130  
20:49:12.827 B17F.1    stand  255  70     70     73    80   123  
20:49:12.827 B17F.0    stand  8    -70    200    0     80   191  
20:49:13.828 B17F.2    sit    8    -50    40     0     80   161  
20:49:13.828 B17F.1    stand  255  60     90     85    80   120  
20:49:13.828 B17F.0    stand  8    -70    200    0     80   170  
20:49:14.829 B17F.0    stand  8    -70    200    0     80   0    
20:49:14.829 B17F.1    stand  255  60     90     86    80   170  
20:49:14.829 B17F.2    sit    8    -50    40     0     80   120  
20:49:15.728 B17F.0    stand  8    -70    200    0     80   161  
20:49:15.728 B17F.1    stand  255  70     80     81    80   184  
20:49:15.728 B17F.2    sit    8    -50    40     0     80   126  
20:49:16.729 B17F.0    stand  8    -70    200    0     80   161  
20:49:16.729 B17F.2    sit    8    -50    40     0     80   161  
20:49:16.729 B17F.1    stand  255  80     80     63    80   136  
20:49:17.726 B17F.2    sit    8    -50    40     0     80   136  
20:49:17.726 B17F.0    stand  8    -70    200    0     80   161  
20:49:17.726 B17F.1    stand  255  60     80     90    80   176  
20:49:18.727 B17F.0    stand  8    -70    200    0     80   176  
20:49:18.727 B17F.1    stand  255  70     70     90    80   191  
20:49:18.727 B17F.2    sit    8    -50    40     0     80   123  
20:49:19.728 B17F.1    stand  255  50     70     78    80   104  
20:49:19.728 B17F.2    sit    8    -50    40     0     80   104  
20:49:19.728 B17F.0    stand  8    -70    200    0     80   161  
20:49:20.735 B17F.0    stand  8    -70    200    0     80   0    
20:49:20.735 B17F.2    sit    8    -50    40     0     80   161  
20:49:20.735 B17F.1    stand  255  70     70     84    80   123  
20:49:21.734 B17F.2    sit    8    -50    40     0     80   123  
20:49:21.734 B17F.0    stand  8    -70    200    0     80   161  
20:49:21.734 B17F.1    stand  255  60     70     82    80   183  
20:49:22.732 B17F.0    stand  8    -70    200    0     80   183  
20:49:22.732 B17F.1    stand  255  60     70     63    80   183  
20:49:22.732 B17F.2    sit    8    -50    40     0     80   114  
20:49:23.736 B17F.2    sit    8    -50    40     0     80   0    
20:49:23.736 B17F.0    stand  8    -70    200    0     80   161  
20:49:23.736 B17F.1    stand  255  50     90     85    80   162  
20:49:24.742 B17F.0    stand  8    -70    200    0     80   162  
20:49:24.742 B17F.2    sit    8    -50    40     0     80   161  
20:49:24.742 B17F.1    stand  255  30     90     72    80   94   
20:49:25.646 B17F.2    sit    8    -80    30     0     80   125  
20:49:25.646 B17F.0    walk   8    -80    130    0     80   100  
20:49:25.646 B17F.1    stand  255  -10    50     89    80   106  
20:49:26.652 B17F.0    walk   8    -90    120    0     80   106  
20:49:26.652 B17F.2    sit    8    -80    30     0     80   90   
20:49:26.652 B17F.1    walk   255  -90    30     64    80   10   
20:49:27.652 B17F.0    walk   8    -120   120    0     80   94   
20:49:27.652 B17F.2    sit    8    -80    30     0     80   98   
20:49:27.652 B17F.1    walk   255  -160   50     72    80   82   
20:49:28.654 B17F.2    sit    8    -60    20     0     80   104  
20:49:28.654 B17F.1    walk   255  -190   60     0     80   136  
20:49:28.654 B17F.0    walk   8    -150   130    82    80   80   
20:49:29.658 B17F.1    walk   255  -180   60     0     80   76   
20:49:29.658 B17F.2    sit    8    -60    20     0     80   126  
20:49:29.658 B17F.0    walk   8    -170   170    115   80   186  
20:49:30.667 B17F.0    walk   8    -170   250    122   80   80   
20:49:30.667 B17F.2    sit    8    -60    30     0     80   245  
20:49:30.667 B17F.1    stand  255  -180   70     0     80   126  
20:49:31.663 B17F.0    walk   8    -160   310    118   80   240  
20:49:31.663 B17F.2    sit    8    -60    30     0     80   297  
20:49:31.663 B17F.1    stand  255  -180   70     0     80   126  
20:49:32.653 B17F.0    walk   8    -150   360    116   80   291  
20:49:32.653 B17F.2    sit    8    -60    30     0     80   342  
20:49:32.653 B17F.1    stand  255  -180   70     0     80   126  
20:49:33.668 B17F.2    sit    8    -80    10     0     80   116  
20:49:33.668 B17F.1    stand  255  -180   70     0     80   116  
20:49:33.668 B17F.0    walk   8    -120   410    113   80   345  
20:49:34.657 B17F.0    walk   8    -120   410    0     80   0    
20:49:34.657 B17F.2    sit    8    -80    10     0     80   401  
20:49:34.657 B17F.1    stand  255  -180   60     0     80   111  
20:49:35.717 B17F.1    stand  255  -180   60     0     80   0    
20:49:35.717 B17F.2    sit    8    -90    20     0     80   98   
20:49:35.717 B17F.0    stand  8    -120   400    0     80   381  
20:49:36.559 B17F.2    sit    8    -130   10     0     80   390  
20:49:36.559 B17F.0    stand  8    -120   400    0     80   390  
20:49:36.559 B17F.1    stand  255  -180   70     0     80   335  
20:49:37.562 B17F.0    stand  8    -120   400    0     80   335  
20:49:37.562 B17F.2    sit    8    -130   10     0     80   390  
20:49:37.562 B17F.1    stand  255  -180   70     0     80   78   
20:49:38.598 B17F.1    stand  255  -180   70     0     80   0    
20:49:38.598 B17F.2    sit    8    -120   20     0     80   78   
20:49:38.598 B17F.0    stand  8    -120   400    0     80   380  
20:49:39.567 B17F.0    stand  8    -180   360    122   80   72   
20:49:39.567 B17F.1    stand  255  -180   70     0     80   290  
20:49:39.567 B17F.2    sit    8    -120   20     0     80   78   
20:49:40.559 B17F.1    stand  255  -180   70     0     80   78   
20:49:40.559 B17F.0    stand  8    -190   350    0     80   280  
20:49:40.559 B17F.2    sit    8    -90    30     0     80   335  
20:49:41.523 B17F.0    stand  8    -190   350    0     80   335  
20:49:41.523 B17F.2    sit    8    -70    20     75    80   351  
20:49:41.523 B17F.1    stand  255  -180   60     0     80   117  
20:49:42.516 B17F.2    sit    8    -60    10     0     80   130  
20:49:42.516 B17F.0    stand  8    -190   350    0     80   364  
20:49:42.516 B17F.1    stand  255  -180   70     0     80   280  
20:49:43.519 B17F.0    stand  8    -190   350    0     80   280  
20:49:43.519 B17F.1    stand  255  -180   70     0     80   280  
20:49:43.519 B17F.2    sit    8    -70    20     0     80   120  
20:49:44.529 B17F.2    sit    8    -70    20     0     80   0    
20:49:44.529 B17F.0    stand  8    -190   350    0     80   351  
20:49:44.529 B17F.1    stand  255  -180   70     0     80   280  
20:49:45.529 B17F.2    sit    8    -70    20     0     80   120  
20:49:45.529 B17F.0    stand  8    -190   350    0     80   351  
20:49:45.529 B17F.1    stand  255  -170   40     0     80   310  
20:49:46.586 B17F.2    sit    8    -70    20     0     80   101  
20:49:46.586 B17F.1    stand  255  -200   30     0     80   130  
20:49:47.606 B17F.0    stand  255  130    90     83    80   335  
20:49:47.606 B17F.1    stand  255  -200   60     0     80   331  
20:49:47.606 B17F.2    sit    8    -70    20     0     80   136  
20:49:48.550 B17F.2    sit    8    -70    20     0     80   0    
20:49:48.550 B17F.1    stand  255  -200   60     0     80   136  
20:49:48.550 B17F.0    stand  255  140    90     0     80   341  
20:49:49.545 B17F.2    sit    8    -70    20     0     80   221  
20:49:49.545 B17F.1    stand  255  -210   70     78    80   148  
20:49:49.545 B17F.0    stand  255  140    90     0     80   350  
20:49:50.546 B17F.2    sit    8    -70    20     0     80   221  
20:49:50.546 B17F.0    stand  255  140    90     0     80   221  
20:49:50.546 B17F.1    stand  255  -210   100    0     80   350  
20:49:51.440 B17F.0    stand  255  140    90     0     80   350  
20:49:51.440 B17F.1    stand  255  -210   100    0     80   350  
20:49:51.440 B17F.2    sit    8    -70    20     0     80   161  
20:49:52.448 B17F.0    stand  255  120    130    0     80   219  
20:49:52.448 B17F.2    sit    8    -70    20     0     80   219  
20:49:52.448 B17F.1    stand  255  -210   80     0     80   152  
20:49:53.442 B17F.2    sit    8    -70    20     0     80   152  
20:49:53.442 B17F.1    stand  255  -210   80     0     80   152  
20:49:53.442 B17F.0    stand  255  120    130    0     80   333  
20:49:54.439 B17F.2    sit    8    -70    20     0     80   219  
20:49:54.439 B17F.1    stand  255  -210   90     0     80   156  
20:49:54.439 B17F.0    stand  255  120    130    0     80   332  
20:49:55.439 B17F.0    stand  255  120    130    0     80   0    
20:49:55.439 B17F.2    sit    8    -70    20     0     80   219  
20:49:55.439 B17F.1    stand  255  -190   90     84    80   138  
20:49:56.444 B17F.0    stand  255  120    120    0     80   311  
20:49:56.444 B17F.1    stand  255  -150   40     83    80   281  
20:49:56.444 B17F.2    sit    8    -70    20     0     80   82   
20:49:57.462 B17F.2    sit    8    -70    20     0     80   0    
20:49:57.462 B17F.1    walk   255  -100   10     40    80   31   
20:49:57.462 B17F.0    stand  255  120    120    0     80   245  
20:49:58.457 B17F.0    stand  255  120    120    0     80   0    
20:49:58.457 B17F.1    walk   255  -50    20     0     80   197  
20:49:58.457 B17F.2    sit    8    -70    20     0     80   20   
20:49:59.475 B17F.2    sit    8    -70    20     0     80   0    
20:49:59.475 B17F.1    walk   255  -50    20     0     80   20   
20:49:59.475 B17F.0    stand  255  130    120    0     80   205  
20:50:00.470 B17F.2    sit    8    -70    20     0     80   223  
20:50:00.470 B17F.0    stand  255  140    80     77    80   218  
20:50:00.470 B17F.1    stand  255  -50    20     0     80   199  
20:50:01.357 B17F.0    stand  255  130    70     77    80   186  
20:50:01.357 B17F.1    stand  255  -50    20     0     80   186  
20:50:01.357 B17F.2    sit    8    -70    20     0     80   20   
20:50:02.358 B17F.0    stand  255  150    80     0     80   228  
20:50:02.358 B17F.2    sit    8    -70    20     0     80   228  
20:50:02.358 B17F.1    stand  255  -60    20     0     80   10   
20:50:03.364 B17F.2    sit    8    -70    20     0     80   10   
20:50:03.364 B17F.1    stand  255  -60    20     0     80   10   
20:50:03.364 B17F.0    stand  255  150    90     0     80   221  
20:50:04.361 B17F.2    sit    8    -70    20     0     80   230  
20:50:04.361 B17F.1    stand  255  -60    20     0     80   10   
20:50:04.361 B17F.0    stand  255  140    90     0     80   211  
20:50:05.363 B17F.0    stand  255  130    100    0     80   14   
20:50:05.363 B17F.2    sit    8    -70    20     0     80   215  
20:50:05.363 B17F.1    stand  255  -60    20     0     80   10   
20:50:06.364 B17F.0    stand  255  140    110    0     80   219  
20:50:06.364 B17F.1    stand  255  -50    20     0     80   210  
20:50:06.364 B17F.2    sit    8    -70    20     0     80   20   
20:50:07.363 B17F.2    sit    8    -70    20     0     80   0    
20:50:07.363 B17F.1    stand  255  -50    20     0     80   20   
20:50:07.363 B17F.0    stand  255  130    90     0     80   193  
20:50:08.366 B17F.0    stand  255  130    90     0     80   0    
20:50:08.366 B17F.1    stand  255  -60    20     0     80   202  
20:50:08.366 B17F.2    sit    8    -70    20     0     80   10   
20:50:09.366 B17F.2    sit    8    -70    20     0     80   0    
20:50:09.366 B17F.1    stand  255  -70    10     0     80   10   
20:50:09.366 B17F.0    stand  255  130    90     0     80   215  
20:50:10.369 B17F.2    sit    8    -70    20     0     80   211  
20:50:10.369 B17F.1    stand  255  -70    10     0     80   10   
20:50:10.369 B17F.0    stand  255  130    90     0     80   215  
20:50:11.372 B17F.2    sit    8    -70    20     0     80   211  
20:50:11.372 B17F.1    stand  255  -70    10     0     80   10   
20:50:11.372 B17F.0    stand  255  130    90     0     80   215  
20:50:12.275 B17F.2    sit    8    -70    20     0     80   211  
20:50:12.275 B17F.0    stand  255  130    90     0     80   211  
20:50:12.275 B17F.1    stand  255  -70    10     0     80   215  
20:50:13.269 B17F.1    stand  255  -70    10     0     80   0    
20:50:13.269 B17F.0    stand  255  130    90     0     80   215  
20:50:13.269 B17F.2    sit    8    -70    20     0     80   211  
20:50:14.281 B17F.0    stand  255  130    90     0     80   211  
20:50:14.281 B17F.2    sit    8    -70    20     0     80   211  
20:50:14.281 B17F.1    stand  255  -50    20     0     80   20   
20:50:15.283 B17F.2    sit    8    -70    20     0     80   20   
20:50:15.283 B17F.0    stand  255  130    90     0     80   211  
20:50:15.283 B17F.1    stand  255  -80    20     65    80   221  
20:50:16.279 B17F.2    sit    8    -70    20     0     80   10   
20:50:16.279 B17F.0    stand  255  140    90     80    80   221  
20:50:16.279 B17F.1    stand  255  -100   10     0     80   252  
20:50:17.288 B17F.2    sit    8    -70    20     0     80   31   
20:50:17.288 B17F.0    stand  255  130    70     0     80   206  
20:50:17.288 B17F.1    stand  255  -100   10     0     80   237  
20:50:18.286 B17F.1    stand  255  -100   10     0     80   0    
20:50:18.286 B17F.0    stand  255  130    70     0     80   237  
20:50:18.286 B17F.2    sit    8    -70    20     0     80   206  
20:50:19.284 B17F.0    stand  255  130    70     0     80   206  
20:50:19.284 B17F.2    sit    8    -70    20     0     80   206  
20:50:19.284 B17F.1    stand  255  -90    20     0     80   20   
20:50:20.282 B17F.0    stand  255  110    90     0     80   211  
20:50:20.282 B17F.1    stand  255  -90    20     0     80   211  
20:50:20.282 B17F.2    sit    8    -70    20     0     80   20   
20:50:21.289 B17F.2    sit    8    -70    20     0     80   0    
20:50:21.289 B17F.1    stand  255  -90    20     0     80   20   
20:50:21.289 B17F.0    stand  255  110    110    83    80   219  
20:50:22.295 B17F.0    stand  255  140    90     72    80   36   
20:50:22.295 B17F.1    stand  255  -90    10     0     80   243  
20:50:22.295 B17F.2    sit    8    -70    20     0     80   22   
20:50:23.180 B17F.0    stand  255  140    60     0     80   213  
20:50:23.180 B17F.2    sit    8    -70    20     0     80   213  
20:50:23.180 B17F.1    stand  255  -90    10     0     80   22   
20:50:24.181 B17F.0    stand  255  110    40     0     80   202  
20:50:24.181 B17F.1    stand  255  -90    10     0     80   202  
20:50:24.181 B17F.2    sit    8    -70    20     0     80   22   
20:50:25.189 B17F.2    sit    8    -70    20     0     80   0    
20:50:25.189 B17F.1    stand  255  -90    10     0     80   22   
20:50:25.189 B17F.0    stand  255  100    40     0     80   192  
20:50:26.183 B17F.0    stand  255  110    40     0     80   10   
20:50:26.183 B17F.1    stand  255  -90    10     0     80   202  
20:50:26.183 B17F.2    sit    8    -70    20     0     80   22   
20:50:27.184 B17F.1    stand  255  -90    10     0     80   22   
20:50:27.184 B17F.2    sit    8    -70    20     0     80   22   
20:50:27.184 B17F.0    stand  255  110    40     0     80   181  
20:50:28.188 B17F.2    sit    8    -70    20     0     80   181  
20:50:28.188 B17F.0    stand  255  110    40     0     80   181  
20:50:28.188 B17F.1    stand  255  -90    10     0     80   202  
20:50:29.189 B17F.0    stand  255  110    40     0     80   202  
20:50:29.189 B17F.1    stand  255  -60    20     0     80   171  
20:50:29.189 B17F.2    sit    8    -70    20     0     80   10   
20:50:30.127 B17F.2    sit    8    -70    20     0     80   0    
20:50:30.127 B17F.0    stand  255  110    40     0     80   181  
20:50:30.127 B17F.1    stand  255  -120   20     39    80   230  
20:50:31.133 B17F.0    stand  255  120    0      0     80   240  
20:50:31.133 B17F.2    sit    8    -70    20     0     80   191  
20:50:31.133 B17F.1    stand  255  -120   30     0     80   50   
20:50:32.138 B17F.0    stand  255  120    0      0     80   241  
20:50:32.138 B17F.1    stand  255  -120   30     38    80   241  
20:50:32.138 B17F.2    sit    8    -70    20     0     80   50   
20:50:33.133 B17F.2    sit    8    -70    20     0     80   0    
20:50:33.133 B17F.1    stand  255  -150   30     73    80   80   
20:50:33.133 B17F.0    stand  255  120    0      0     80   271  
20:50:34.128 B17F.0    stand  255  120    0      0     80   0    
20:50:34.128 B17F.1    stand  255  -190   60     75    80   315  
20:50:34.128 B17F.2    sit    8    -70    20     0     80   126  
20:50:35.180 B17F.1    walk   255  -200   100    58    80   152  
20:50:35.180 B17F.0    stand  255  120    0      0     80   335  
20:50:35.180 B17F.2    sit    8    -70    20     0     80   191  
20:50:36.136 B17F.0    stand  255  120    0      0     80   191  
20:50:36.136 B17F.1    walk   255  -200   140    97    80   349  
20:50:36.136 B17F.2    sit    8    -70    20     0     80   176  
20:50:37.132 B17F.2    sit    8    -70    20     0     80   0    
20:50:37.132 B17F.0    stand  255  120    0      0     80   191  
20:50:37.132 B17F.1    walk   255  -200   130    0     80   345  
20:50:38.131 B17F.2    sit    8    -70    20     0     80   170  
20:50:38.131 B17F.0    stand  255  120    0      0     80   191  
20:50:38.131 B17F.1    stand  255  -180   140    0     80   331  
20:50:39.135 B17F.1    stand  255  -190   140    0     80   10   
20:50:39.135 B17F.0    stand  255  120    0      0     80   340  
20:50:39.135 B17F.2    sit    8    -70    20     0     80   191  
20:50:40.138 B17F.2    sit    8    -70    20     0     80   0    
20:50:40.138 B17F.1    stand  255  -180   140    87    80   162  
20:50:40.138 B17F.0    stand  255  120    0      0     80   331  
20:50:41.136 B17F.0    stand  255  120    0      0     80   0    
20:50:41.136 B17F.2    sit    8    -70    20     0     80   191  
20:50:41.136 B17F.1    stand  255  -180   160    101   80   178  
20:50:42.029 B17F.2    sit    8    -70    20     0     80   178  
20:50:42.029 B17F.0    stand  255  120    0      0     80   191  
20:50:42.029 B17F.1    stand  255  -160   130    92    80   308  
20:50:43.028 B17F.1    stand  255  -150   130    98    80   10   
20:50:43.028 B17F.2    sit    8    -70    20     0     80   136  
20:50:43.028 B17F.0    stand  255  120    0      0     80   191  
20:50:44.032 B17F.1    stand  255  -130   120    95    80   277  
20:50:44.032 B17F.0    stand  255  120    0      0     80   277  
20:50:44.032 B17F.2    sit    8    -70    20     0     80   191  
20:50:45.037 B17F.2    sit    8    -70    20     0     80   0    
20:50:45.037 B17F.1    stand  255  -140   130    103   80   130  
20:50:45.037 B17F.0    stand  255  120    0      0     80   290  
20:50:45.999 B17F.2    sit    8    -70    20     0     80   191  
20:50:45.999 B17F.0    stand  255  120    0      0     80   191  
20:50:45.999 B17F.1    stand  255  -140   130    107   80   290  
20:50:47.004 B17F.0    stand  255  120    0      0     80   290  
20:50:47.004 B17F.2    sit    8    -70    20     0     80   191  
20:50:47.004 B17F.1    stand  255  -150   120    97    80   128  
20:50:48.056 B17F.0    stand  255  120    0      0     80   295  
20:50:48.056 B17F.1    stand  255  -180   120    92    80   323  
20:50:49.013 B17F.1    stand  255  -180   120    79    80   0    
20:50:49.013 B17F.0    stand  255  120    0      0     80   323  
20:50:50.028 B17F.1    stand  255  -180   100    64    80   316  
20:50:50.028 B17F.0    stand  255  120    0      0     80   316  
20:50:51.014 B17F.1    stand  255  -170   110    87    80   310  
20:50:51.014 B17F.0    stand  255  120    0      0     80   310  
20:50:52.015 B17F.0    stand  255  120    0      0     80   0    
20:50:52.015 B17F.1    stand  255  -170   110    80    80   310  
20:50:53.019 B17F.1    stand  255  -170   70     73    80   40   
20:50:53.019 B17F.0    stand  255  120    0      0     80   298  
20:50:54.024 B17F.1    stand  255  -170   60     67    80   296  
20:50:54.024 B17F.0    stand  255  120    0      0     80   296  
20:50:55.021 B17F.1    stand  255  -180   20     0     80   300  
20:50:55.021 B17F.0    stand  255  120    0      0     80   300  
20:50:55.925 B17F.0    stand  255  120    0      0     80   0    
20:50:55.925 B17F.1    stand  255  -180   20     0     80   300  
20:50:56.955 B17F.1    stand  255  -170   30     0     80   14   
20:50:56.955 B17F.0    stand  255  120    0      0     80   291  
20:50:57.923 B17F.1    stand  255  -170   30     0     80   291  
20:50:57.923 B17F.0    stand  255  120    0      0     80   291  
20:50:58.918 B17F.1    stand  255  -170   30     0     80   291  
20:50:58.918 B17F.0    stand  255  120    0      0     80   291  
20:50:59.921 B17F.1    stand  255  -170   60     57    80   296  
20:50:59.921 B17F.0    stand  255  120    0      0     80   296  
20:51:00.923 B17F.0    stand  255  120    0      0     80   0    
20:51:00.923 B17F.1    stand  255  -170   80     75    80   300  
20:51:01.921 B17F.0    stand  255  120    0      0     80   300  
20:51:01.921 B17F.1    stand  255  -170   70     74    80   298  
20:51:02.927 B17F.1    stand  255  -190   60     60    80   22   
20:51:02.927 B17F.0    stand  255  120    0      0     80   315  
20:51:03.927 B17F.1    stand  255  -190   40     0     80   312  
20:51:03.927 B17F.0    stand  255  210    50     0     80   400  
20:51:04.927 B17F.1    stand  255  -180   50     0     80   390  
20:51:04.927 B17F.0    stand  255  210    50     0     80   390  
20:51:05.930 B17F.1    stand  255  -180   50     0     80   390  
20:51:05.930 B17F.0    stand  255  200    50     0     80   380  
20:51:06.929 B17F.0    stand  255  200    40     0     80   10   
20:51:06.929 B17F.1    stand  255  -180   50     0     80   380  
20:51:07.819 B17F.1    stand  255  -180   50     0     80   0    
20:51:07.819 B17F.0    stand  255  200    40     0     80   380  
20:51:08.823 B17F.1    stand  255  -170   50     0     80   370  
20:51:08.823 B17F.0    stand  255  200    40     0     80   370  
20:51:09.823 B17F.1    stand  255  -150   30     0     80   350  
20:51:09.823 B17F.0    stand  255  140    90     0     80   296  
20:51:10.834 B17F.0    walk   255  140    110    0     80   20   
20:51:10.834 B17F.1    stand  255  -150   30     0     80   300  
20:51:11.825 B17F.1    stand  255  -150   30     0     80   0    
20:51:11.825 B17F.0    walk   255  140    110    0     80   300  
20:51:12.832 B17F.1    stand  255  -140   0      0     80   300  
20:51:12.832 B17F.0    stand  255  140    110    0     80   300  
20:51:13.831 B17F.0    stand  255  140    110    0     80   0    
20:51:13.831 B17F.1    stand  255  -140   0      0     80   300  
20:51:14.828 B17F.1    stand  255  -140   0      0     80   0    
20:51:14.828 B17F.0    stand  255  140    100    0     80   297  
20:51:15.836 B17F.0    stand  255  140    100    0     80   0    
20:51:15.836 B17F.1    stand  255  -150   0      0     80   306  
20:51:16.829 B17F.0    stand  255  140    100    0     80   306  
20:51:16.829 B17F.1    stand  255  -150   0      0     80   306  
20:51:17.747 B17F.0    stand  255  130    10     0     80   280  
20:51:17.747 B17F.1    stand  255  -150   0      0     80   280  
20:51:18.739 B17F.1    stand  255  -120   30     0     80   42   
20:51:18.739 B17F.0    stand  255  130    0      0     80   251  
20:51:19.751 B17F.1    stand  255  -90    70     0     80   230  
20:51:19.751 B17F.0    stand  255  130    0      0     80   230  
20:51:20.751 B17F.0    stand  255  130    0      0     80   0    
20:51:20.751 B17F.1    stand  255  -90    70     0     80   230  
20:51:21.746 B17F.1    stand  255  -90    70     0     80   0    
20:51:21.746 B17F.0    stand  255  130    0      0     80   230  
20:51:22.751 B17F.0    stand  255  130    0      0     80   0    
20:51:22.751 B17F.1    stand  255  -70    20     0     80   200  
20:51:23.752 B17F.0    stand  255  130    0      0     80   200  
20:51:23.752 B17F.1    stand  255  -90    10     38    80   220  
20:51:24.761 B17F.1    stand  255  -130   30     81    80   44   
20:51:24.761 B17F.0    stand  255  130    0      0     80   261  
20:51:25.754 B17F.0    stand  255  130    0      0     80   0    
20:51:25.754 B17F.1    walk   255  -160   80     68    80   300  
20:51:26.751 B17F.1    walk   255  -170   160    107   80   80   
20:51:26.751 B17F.0    stand  255  130    0      0     80   340  
20:51:27.755 B17F.1    walk   255  -190   260    110   80   412  
20:51:27.755 B17F.0    stand  255  130    0      0     80   412  
20:51:28.751 B17F.1    walk   255  -180   350    114   80   467  
20:51:28.751 B17F.0    stand  255  130    0      0     80   467  
20:51:29.647 B17F.0    stand  255  130    0      0     80   0    
20:51:29.647 B17F.1    walk   255  -150   400    94    80   488  
20:51:30.649 B17F.1    walk   255  -160   410    0     80   14   
20:51:30.649 B17F.0    stand  255  130    0      0     80   502  
20:51:31.655 B17F.1    stand  255  -160   400    0     80   494  
20:51:31.655 B17F.0    stand  255  140    70     76    80   445  
20:51:32.660 B17F.1    stand  255  -160   400    0     80   445  
20:51:32.660 B17F.0    stand  255  130    100    0     80   417  
20:51:33.614 B17F.0    stand  255  140    90     71    80   14   
20:51:33.614 B17F.1    stand  255  -170   380    0     80   424  
20:51:34.633 B17F.1    stand  255  -190   310    113   80   72   
20:51:34.633 B17F.0    stand  255  130    120    0     80   372  
20:51:35.678 B17F.0    stand  255  130    120    0     80   0    
20:51:35.678 B17F.1    walk   255  -190   240    119   80   341  
20:51:36.617 B17F.1    walk   255  -180   160    105   80   80   
20:51:36.617 B17F.0    stand  255  130    120    0     80   312  
20:51:37.619 B17F.1    walk   255  -160   90     88    80   291  
20:51:37.619 B17F.0    stand  255  140    120    0     80   301  
20:51:38.617 B17F.1    walk   255  -150   90     76    80   291  
20:51:38.617 B17F.0    stand  255  140    120    0     80   291  
20:51:39.617 B17F.1    walk   255  -170   150    97    80   311  
20:51:39.617 B17F.0    stand  255  140    120    0     80   311  
20:51:40.626 B17F.1    walk   255  -190   250    112   80   354  
20:51:40.626 B17F.0    stand  255  140    120    0     80   354  
20:51:41.620 B17F.0    stand  255  140    120    0     80   0    
20:51:41.620 B17F.1    walk   255  -180   330    110   80   382  
20:51:42.632 B17F.0    stand  255  140    120    0     80   382  
20:51:42.632 B17F.1    walk   255  -190   390    99    80   426  
20:51:43.623 B17F.1    walk   255  -200   390    0     80   10   
20:51:43.623 B17F.0    stand  255  140    120    0     80   434  
20:51:44.634 B17F.1    walk   255  -200   390    0     80   434  
20:51:44.634 B17F.0    stand  255  140    80     75    80   460  
20:51:45.516 B17F.1    stand  255  -200   390    0     80   460  
20:51:45.516 B17F.0    stand  255  130    120    0     80   426  
20:51:46.515 B17F.0    stand  255  130    120    0     80   0    
20:51:46.515 B17F.1    stand  255  -200   380    0     80   420  
20:51:47.519 B17F.1    stand  255  -200   380    0     80   0    
20:51:47.519 B17F.0    stand  255  130    120    0     80   420  
20:51:48.520 B17F.1    stand  8    -200   380    0     80   420  
20:51:48.520 B17F.0    stand  255  130    120    0     80   420  
20:51:49.605 B17F.0    stand  255  130    120    0     80   0    
20:51:50.555 B17F.0    stand  255  130    130    0     80   10   
20:51:51.552 B17F.0    stand  255  130    130    0     80   0    
20:51:52.552 B17F.0    stand  255  130    130    0     80   0    
20:51:53.554 B17F.0    stand  255  130    130    0     80   0    
20:51:54.446 B17F.0    stand  255  120    130    0     80   10   
20:51:55.452 B17F.0    stand  255  150    70     0     80   67   
20:51:56.444 B17F.0    stand  255  140    80     70    80   14   
20:51:57.445 B17F.0    stand  255  150    90     0     80   14   
20:51:58.451 B17F.0    stand  255  150    100    0     80   10   
20:51:59.450 B17F.0    stand  255  150    100    0     80   0    
20:52:00.453 B17F.0    stand  255  150    100    0     80   0    
20:52:01.449 B17F.0    stand  255  140    100    0     80   10   
20:52:02.457 B17F.0    stand  255  130    110    0     80   14   
20:52:03.458 B17F.0    stand  255  140    110    0     80   10   
20:52:04.524 B17F.1    stand  255  -170   160    97    80   314  
20:52:04.524 B17F.0    stand  255  140    110    0     80   314  
20:52:05.366 B17F.0    stand  255  140    110    0     80   0    
20:52:05.366 B17F.1    walk   255  -190   60     86    80   333  
20:52:06.363 B17F.1    walk   255  -190   30     0     80   30   
20:52:06.363 B17F.0    stand  255  140    110    0     80   339  
20:52:07.361 B17F.1    walk   255  -190   40     0     80   337  
20:52:07.361 B17F.0    stand  255  140    110    0     80   337  
20:52:08.366 B17F.0    stand  255  140    110    0     80   0    
20:52:08.366 B17F.1    stand  255  -190   40     0     80   337  
20:52:09.372 B17F.0    stand  255  140    110    0     80   337  
20:52:09.372 B17F.1    stand  255  -190   50     0     80   335  
20:52:10.360 B17F.0    stand  255  140    90     75    80   332  
20:52:10.360 B17F.1    stand  255  -190   50     0     80   332  
20:52:11.384 B17F.1    stand  255  -190   50     0     80   0    
20:52:11.384 B17F.0    stand  255  140    90     0     80   332  
20:52:12.375 B17F.0    stand  255  130    110    0     80   22   
20:52:12.375 B17F.1    stand  255  -190   50     0     80   325  
20:52:13.365 B17F.0    stand  255  170    110    0     80   364  
20:52:13.365 B17F.1    stand  255  -190   50     0     80   364  
20:52:14.378 B17F.1    stand  255  -190   50     0     80   0    
20:52:14.378 B17F.0    stand  255  170    110    0     80   364  
20:52:15.367 B17F.0    stand  255  170    110    0     80   0    
20:52:15.367 B17F.1    stand  255  -190   50     0     80   364  
20:52:16.277 B17F.1    stand  255  -190   50     0     80   0    
20:52:16.277 B17F.0    stand  255  170    110    0     80   364  
20:52:17.267 B17F.0    stand  255  160    80     0     80   31   
20:52:17.267 B17F.1    stand  255  -200   80     96    80   360  
20:52:18.283 B17F.1    stand  255  -160   140    96    80   72   
20:52:18.283 B17F.0    stand  255  150    80     0     80   315  
20:52:19.271 B17F.0    stand  255  150    80     63    80   0    
20:52:19.271 B17F.1    walk   255  -140   180    0     80   306  
20:52:20.282 B17F.1    walk   255  -150   170    108   80   14   
20:52:20.282 B17F.0    stand  255  140    100    0     80   298  
20:52:21.279 B17F.0    stand  255  140    100    0     80   0    
20:52:21.279 B17F.1    walk   255  -160   160    110   80   305  
20:52:22.220 B17F.1    walk   255  -150   100    89    80   60   
20:52:22.220 B17F.0    stand  255  140    100    0     80   290  
20:52:23.234 B17F.0    stand  255  140    100    0     80   0    
20:52:23.234 B17F.1    walk   255  -130   20     87    80   281  
20:52:24.224 B17F.1    walk   255  -100   0      0     80   36   
20:52:24.224 B17F.0    stand  255  140    100    0     80   260  
20:52:25.224 B17F.0    stand  255  140    100    0     80   0    
20:52:25.224 B17F.1    walk   255  -110   0      0     80   269  
20:52:26.245 B17F.1    stand  255  -110   0      0     80   0    
20:52:26.245 B17F.0    stand  255  150    90     0     80   275  
20:52:27.232 B17F.0    stand  255  150    140    0     80   50   
20:52:27.232 B17F.1    stand  255  -110   0      0     80   295  
20:52:28.232 B17F.1    stand  255  -110   0      0     80   0    
20:52:28.232 B17F.0    stand  255  150    140    0     80   295  
20:52:29.228 B17F.0    stand  255  150    140    0     80   0    
20:52:29.228 B17F.1    stand  255  -110   0      0     80   295  
20:52:30.228 B17F.1    stand  255  -40    50     0     80   86   
20:52:30.228 B17F.0    stand  255  140    90     75    80   184  
20:52:31.232 B17F.1    stand  255  -40    50     0     80   184  
20:52:31.232 B17F.0    stand  255  140    90     78    80   184  
20:52:32.236 B17F.0    stand  255  130    60     0     80   31   
20:52:32.236 B17F.1    stand  255  -50    50     0     80   180  
20:52:33.133 B17F.0    stand  255  130    60     0     80   180  
20:52:33.133 B17F.1    stand  255  -50    50     0     80   180  
20:52:34.138 B17F.1    stand  255  -50    50     0     80   0    
20:52:34.138 B17F.0    stand  255  140    60     0     80   190  
20:52:35.196 B17F.0    stand  255  140    60     0     80   0    
20:52:35.196 B17F.1    stand  255  -50    50     0     80   190  
20:52:36.144 B17F.1    stand  255  -50    50     0     80   0    
20:52:36.144 B17F.0    stand  255  140    60     0     80   190  
20:52:37.140 B17F.0    stand  255  140    60     0     80   0    
20:52:37.140 B17F.1    stand  255  -50    50     0     80   190  
20:52:38.178 B17F.1    stand  255  -50    50     0     80   0    
20:52:38.178 B17F.0    stand  255  140    60     0     80   190  
20:52:39.175 B17F.0    stand  255  140    60     0     80   0    
20:52:39.175 B17F.1    stand  255  -50    50     0     80   190  
20:52:40.173 B17F.1    stand  255  -50    50     0     80   0    
20:52:40.173 B17F.0    stand  255  140    60     0     80   190  
20:52:41.071 B17F.1    stand  255  -70    40     0     80   210  
20:52:41.071 B17F.0    stand  255  140    60     0     80   210  
20:52:42.073 B17F.1    stand  255  -130   10     0     80   274  
20:52:42.073 B17F.0    stand  255  140    60     0     80   274  
20:52:43.079 B17F.0    stand  255  140    60     0     80   0    
20:52:43.079 B17F.1    stand  255  -130   10     0     80   274  
20:52:44.073 B17F.1    stand  255  -130   20     0     80   10   
20:52:44.073 B17F.0    stand  255  140    60     0     80   272  
20:52:45.071 B17F.0    stand  255  140    60     0     80   0    
20:52:45.071 B17F.1    stand  255  -120   20     0     80   263  
20:52:46.112 B17F.1    stand  255  -120   20     0     80   0    
20:52:46.112 B17F.0    stand  255  140    60     0     80   263  
20:52:47.083 B17F.0    stand  255  140    60     0     80   0    
20:52:47.083 B17F.1    stand  255  -120   20     0     80   263  
20:52:48.080 B17F.1    stand  255  -120   20     0     80   0    
20:52:48.080 B17F.0    stand  255  140    60     0     80   263  
20:52:49.077 B17F.0    stand  255  140    100    73    80   40   
20:52:49.077 B17F.1    stand  255  -120   20     0     80   272  
20:52:50.079 B17F.0    stand  255  150    40     72    80   270  
20:52:50.079 B17F.1    stand  255  -150   50     0     80   300  
20:52:51.079 B17F.0    stand  255  130    80     0     80   281  
20:52:51.079 B17F.1    stand  255  -160   40     63    80   292  
20:52:52.082 B17F.1    walk   255  -160   120    98    80   80   
20:52:52.082 B17F.0    stand  255  130    80     0     80   292  
20:52:52.985 B17F.1    walk   255  -150   180    110   80   297  
20:52:52.985 B17F.0    stand  255  130    80     0     80   297  
20:52:54.007 B17F.0    stand  255  120    110    0     80   31   
20:52:54.007 B17F.1    walk   255  -160   210    120   80   297  
20:52:55.026 B17F.1    walk   255  -170   240    128   80   31   
20:52:55.026 B17F.0    stand  255  120    110    0     80   317  
20:52:55.998 B17F.0    stand  255  130    120    0     80   14   
20:52:55.998 B17F.1    walk   255  -140   230    0     80   291  
20:52:57.003 B17F.1    stand  255  -140   230    0     80   0    
20:52:57.003 B17F.0    stand  255  130    120    78    80   291  
20:52:57.997 B17F.0    stand  255  130    120    0     80   0    
20:52:57.997 B17F.1    stand  255  -140   230    0     80   291  
20:52:59.002 B17F.1    stand  255  -140   230    0     80   0    
20:52:59.002 B17F.0    stand  255  130    120    0     80   291  
20:53:00.010 B17F.0    stand  255  130    120    0     80   0    
20:53:00.010 B17F.1    stand  255  -140   230    0     80   291  
20:53:01.007 B17F.1    stand  255  -140   230    0     80   0    
20:53:01.007 B17F.0    stand  255  130    120    0     80   291  
20:53:01.909 B17F.0    stand  255  140    120    0     80   10   
20:53:01.909 B17F.1    stand  255  -140   230    0     80   300  
20:53:02.904 B17F.1    stand  255  -140   230    0     80   0    
20:53:02.904 B17F.0    stand  255  150    60     0     80   336  
20:53:03.903 B17F.0    stand  255  150    60     0     80   0    
20:53:03.903 B17F.1    stand  255  -140   230    0     80   336  
20:53:04.912 B17F.1    stand  255  -110   230    0     80   30   
20:53:04.912 B17F.0    stand  255  150    60     0     80   310  
20:53:05.905 B17F.0    stand  255  150    70     0     80   10   
20:53:05.905 B17F.1    stand  255  -180   220    112   80   362  
20:53:06.905 B17F.1    stand  255  -160   180    100   80   44   
20:53:06.905 B17F.0    stand  255  150    70     0     80   328  
20:53:07.914 B17F.0    stand  255  150    70     0     80   0    
20:53:07.914 B17F.1    walk   255  -160   120    92    80   314  
20:53:08.912 B17F.1    walk   255  -140   80     86    80   44   
20:53:08.912 B17F.0    stand  255  150    70     0     80   290  
20:53:09.845 B17F.0    stand  255  150    70     0     80   0    
20:53:09.845 B17F.1    walk   255  -120   40     76    80   271  
20:53:10.844 B17F.1    walk   255  -80    10     0     80   50   
20:53:10.844 B17F.0    stand  255  150    70     0     80   237  
20:53:11.842 B17F.0    stand  255  150    70     0     80   0    
20:53:11.842 B17F.1    walk   255  -100   20     37    80   254  
20:53:12.853 B17F.1    walk   255  -130   20     0     80   30   
20:53:12.853 B17F.0    stand  255  150    70     0     80   284  
20:53:13.851 B17F.1    stand  255  -130   20     0     80   284  
20:53:13.851 B17F.0    stand  255  150    70     0     80   284  
20:53:14.843 B17F.0    stand  255  100    10     0     80   78   
20:53:14.843 B17F.1    stand  255  -130   20     0     80   230  
20:53:15.845 B17F.1    stand  255  -70    10     67    80   60   
20:53:15.845 B17F.0    stand  255  90     0      0     80   160  
20:53:16.846 B17F.0    stand  255  90     10     0     80   10   
20:53:16.846 B17F.1    stand  255  -70    20     78    80   160  
20:53:17.847 B17F.1    stand  255  -90    30     95    80   22   
20:53:17.847 B17F.0    stand  255  90     10     0     80   181  
20:53:18.850 B17F.0    stand  255  90     10     0     80   0    
20:53:18.850 B17F.1    stand  255  -80    40     107   80   172  
20:53:19.850 B17F.1    stand  255  -90    40     0     80   10   
20:53:19.850 B17F.0    stand  255  90     10     0     80   182  
20:53:20.750 B17F.0    stand  255  90     10     0     80   0    
20:53:20.750 B17F.1    stand  255  -90    40     0     80   182  
20:53:21.750 B17F.1    stand  255  -90    40     65    80   0    
20:53:21.750 B17F.0    stand  255  90     10     0     80   182  
20:53:22.751 B17F.1    stand  255  -80    20     0     80   170  
20:53:22.751 B17F.0    stand  255  90     30     0     80   170  
20:53:23.759 B17F.1    stand  255  -60    10     0     80   151  
20:53:23.759 B17F.0    stand  255  140    110    70    80   223  
20:53:24.761 B17F.1    stand  255  -60    10     0     80   223  
20:53:24.761 B17F.0    walk   255  150    90     74    80   224  
20:53:25.809 B17F.1    stand  255  -60    20     0     80   221  
20:53:25.809 B17F.0    walk   255  130    70     66    80   196  
20:53:26.703 B17F.0    walk   255  110    50     0     80   28   
20:53:26.703 B17F.1    stand  255  -100   30     0     80   210  
20:53:27.707 B17F.1    walk   255  -130   60     0     80   42   
20:53:27.707 B17F.0    walk   255  70     30     0     80   202  
20:53:28.704 B17F.0    walk   255  0      70     0     80   80   
20:53:28.704 B17F.1    walk   255  -140   60     69    80   140  
20:53:29.707 B17F.1    walk   255  -160   60     0     80   20   
20:53:29.707 B17F.0    walk   255  -50    80     66    80   111  
20:53:30.727 B17F.0    walk   255  -90    110    95    80   50   
20:53:30.727 B17F.1    walk   255  -150   50     0     80   84   
20:53:31.711 B17F.1    walk   255  -170   60     0     80   22   
20:53:31.711 B17F.0    walk   255  -130   120    102   80   72   
20:53:32.719 B17F.0    walk   255  -130   120    101   80   0    
20:53:32.719 B17F.1    walk   255  -160   60     0     80   67   
20:53:33.761 B17F.1    stand  255  -160   70     0     80   10   
20:53:33.761 B17F.0    walk   255  -130   120    106   80   58   
20:53:34.713 B17F.1    stand  255  -160   70     0     80   58   
20:53:34.713 B17F.0    walk   255  -130   110    111   80   50   
20:53:35.716 B17F.1    stand  255  -160   70     0     80   50   
20:53:35.716 B17F.0    walk   255  -130   110    107   80   50   
20:53:36.711 B17F.0    walk   255  -120   110    103   80   10   
20:53:36.711 B17F.1    stand  255  -160   70     0     80   56   
20:53:37.718 B17F.0    walk   255  -130   110    117   80   50   
20:53:37.718 B17F.1    stand  255  -160   70     0     80   50   
20:53:38.612 B17F.0    walk   255  -120   110    108   80   56   
20:53:38.612 B17F.1    stand  255  -160   70     0     80   56   
20:53:39.610 B17F.0    walk   255  -140   110    110   80   44   
20:53:39.610 B17F.1    stand  255  -160   70     0     80   44   
20:53:40.610 B17F.1    stand  255  -160   70     0     80   0    
20:53:40.610 B17F.0    walk   255  -130   100    113   80   42   
20:53:41.621 B17F.1    stand  255  -160   70     0     80   42   
20:53:41.621 B17F.0    walk   255  -130   100    124   80   42   
20:53:42.615 B17F.1    stand  255  -160   70     0     80   42   
20:53:42.615 B17F.0    walk   255  -130   100    118   80   42   
20:53:43.618 B17F.1    stand  255  -160   70     0     80   42   
20:53:43.618 B17F.0    walk   255  -130   100    0     80   42   
20:53:44.619 B17F.1    stand  255  -160   70     0     80   42   
20:53:44.619 B17F.0    stand  255  -120   80     0     80   41   
20:53:45.621 B17F.1    stand  255  -160   70     0     80   41   
20:53:45.621 B17F.0    stand  255  -130   90     113   80   36   
20:53:46.621 B17F.1    stand  255  -190   50     0     80   72   
20:53:46.621 B17F.0    stand  255  -120   90     0     80   80   
20:53:47.624 B17F.1    stand  255  -200   50     0     80   89   
20:53:47.624 B17F.0    stand  255  -120   90     108   80   89   
20:53:48.634 B17F.1    stand  255  -200   50     0     80   89   
20:53:48.634 B17F.0    stand  255  -130   90     0     80   80   
20:53:49.518 B17F.1    stand  255  -200   50     0     80   80   
20:53:49.518 B17F.0    stand  255  -130   110    103   80   92   
20:53:50.519 B17F.1    stand  255  -200   60     0     80   86   
20:53:50.519 B17F.0    stand  255  -140   120    97    80   84   
20:53:51.517 B17F.1    stand  255  -190   60     0     80   78   
20:53:51.517 B17F.0    stand  255  -110   70     85    80   80   
20:53:52.526 B17F.1    stand  255  -190   50     0     80   82   
20:53:52.526 B17F.0    walk   255  -10    80     94    80   182  
20:53:53.521 B17F.1    walk   255  -150   150    0     80   156  
20:53:53.521 B17F.0    walk   255  100    50     87    80   269  
20:53:54.523 B17F.0    walk   255  130    80     87    80   42   
20:53:54.523 B17F.1    walk   255  -170   140    0     80   305  
20:53:55.537 B17F.1    walk   255  -180   130    0     80   14   
20:53:55.537 B17F.0    walk   255  160    90     51    80   342  
20:53:56.541 B17F.1    walk   255  -160   160    0     80   327  
20:53:56.541 B17F.0    walk   255  160    80     63    80   329  
20:53:57.537 B17F.0    walk   255  150    70     40    80   14   
20:53:57.537 B17F.1    walk   255  -160   170    0     80   325  
20:53:58.528 B17F.1    walk   255  -150   180    0     80   14   
20:53:58.528 B17F.0    sit    255  140    70     0     80   310  
20:53:59.539 B17F.1    stand  255  -150   180    0     80   310  
20:53:59.539 B17F.0    sit    255  140    60     0     80   313  
20:54:00.434 B17F.1    stand  255  -150   180    0     80   313  
20:54:00.434 B17F.0    sit    255  140    60     0     80   313  
20:54:01.432 B17F.1    stand  255  -150   180    0     80   313  
20:54:01.432 B17F.0    sit    255  140    60     0     80   313  
20:54:02.435 B17F.1    stand  255  -150   180    0     80   313  
20:54:02.435 B17F.0    sit    255  140    100    0     80   300  
20:54:03.434 B17F.1    stand  255  -140   190    0     80   294  
20:54:03.434 B17F.0    sit    255  140    100    0     80   294  
20:54:04.434 B17F.1    stand  255  -150   180    132   80   300  
20:54:04.434 B17F.0    sit    255  140    90     0     80   303  
20:54:05.454 B17F.1    stand  255  -150   130    105   80   292  
20:54:05.454 B17F.0    sit    255  140    90     0     80   292  
20:54:06.439 B17F.1    stand  255  -140   110    110   80   280  
20:54:06.439 B17F.0    sit    255  140    90     0     80   280  
20:54:07.433 B17F.1    stand  255  -130   110    107   80   270  
20:54:07.433 B17F.0    stand  255  150    80     0     80   281  
20:54:08.455 B17F.0    stand  255  150    80     0     80   0    
20:54:08.455 B17F.1    stand  255  -140   90     96    80   290  
20:54:09.440 B17F.1    stand  255  -120   30     88    80   63   
20:54:09.440 B17F.0    stand  255  160    50     0     80   280  
20:54:10.448 B17F.1    stand  255  -100   30     0     80   260  
20:54:10.448 B17F.0    stand  255  160    0      0     80   261  
20:54:11.441 B17F.1    stand  255  -50    50     0     80   215  
20:54:11.441 B17F.0    stand  255  160    0      0     80   215  
20:54:12.334 B17F.1    stand  255  -60    40     83    80   223  
20:54:12.334 B17F.0    stand  255  140    0      0     80   203  
20:54:13.340 B17F.1    stand  255  -110   70     85    80   259  
20:54:13.340 B17F.0    stand  255  110    0      0     80   230  
20:54:14.347 B17F.1    stand  255  -110   110    111   80   245  
20:54:14.347 B17F.0    stand  255  110    0      0     80   245  
20:54:15.340 B17F.0    stand  255  130    100    0     80   101  
20:54:15.340 B17F.1    stand  255  -110   120    114   80   240  
20:54:16.344 B17F.0    stand  255  130    90     0     80   241  
20:54:16.344 B17F.1    stand  255  -110   120    116   80   241  
20:54:17.338 B17F.1    stand  255  -110   110    109   80   10   
20:54:17.338 B17F.0    stand  255  130    90     0     80   240  
20:54:18.341 B17F.1    stand  255  -110   110    0     80   240  
20:54:18.341 B17F.0    stand  255  130    100    0     80   240  
20:54:19.352 B17F.0    stand  255  130    120    0     80   20   
20:54:19.352 B17F.1    stand  255  -130   110    100   80   260  
20:54:20.348 B17F.1    stand  255  -140   70     83    80   41   
20:54:20.348 B17F.0    stand  255  130    110    0     80   272  
20:54:21.344 B17F.1    stand  255  -120   10     0     80   269  
20:54:21.344 B17F.0    stand  255  120    100    80    80   256  
20:54:22.349 B17F.1    stand  255  -120   20     0     80   252  
20:54:22.349 B17F.0    stand  255  110    90     75    80   240  
20:54:23.247 B17F.1    stand  255  -120   20     0     80   240  
20:54:23.247 B17F.0    stand  255  140    80     0     80   266  
20:54:24.251 B17F.1    stand  255  -120   20     0     80   266  
20:54:24.251 B17F.0    stand  255  140    80     0     80   266  
20:54:25.244 B17F.1    stand  255  -120   30     0     80   264  
20:54:25.244 B17F.0    stand  255  140    80     0     80   264  
20:54:26.250 B17F.1    stand  255  -120   30     0     80   264  
20:54:26.250 B17F.0    stand  255  140    80     0     80   264  
20:54:27.262 B17F.1    stand  255  -120   30     0     80   264  
20:54:27.262 B17F.0    stand  255  140    80     0     80   264  
20:54:28.249 B17F.1    stand  255  -120   30     0     80   264  
20:54:28.249 B17F.0    stand  255  140    80     0     80   264  
20:54:29.254 B17F.1    stand  255  -120   30     0     80   264  
20:54:29.254 B17F.0    stand  255  140    80     0     80   264  
20:54:30.272 B17F.0    stand  255  130    110    0     80   31   
20:54:30.272 B17F.1    stand  255  -120   30     0     80   262  
20:54:31.257 B17F.1    stand  255  -90    20     44    80   31   
20:54:31.257 B17F.0    stand  255  130    120    0     80   241  
20:54:32.255 B17F.1    stand  255  -140   40     68    80   281  
20:54:32.255 B17F.0    stand  255  130    120    0     80   281  
20:54:33.305 B17F.1    stand  255  -130   90     92    80   261  
20:54:33.305 B17F.0    stand  255  130    120    0     80   261  
20:54:34.254 B17F.1    walk   255  -120   110    104   80   250  
20:54:34.254 B17F.0    stand  255  120    90     0     80   240  
20:54:35.165 B17F.1    walk   255  -120   110    112   80   240  
20:54:35.165 B17F.0    stand  255  120    80     0     80   241  
20:54:36.165 B17F.1    walk   255  -130   100    103   80   250  
20:54:36.165 B17F.0    stand  255  120    80     0     80   250  
20:54:37.154 B17F.0    stand  255  130    70     0     80   14   
20:54:37.154 B17F.1    walk   255  -130   110    99    80   263  
20:54:38.153 B17F.1    walk   255  -120   110    0     80   10   
20:54:38.153 B17F.0    stand  255  140    80     0     80   261  
20:54:39.157 B17F.1    walk   255  -120   110    121   80   261  
20:54:39.157 B17F.0    stand  255  140    100    0     80   260  
20:54:40.158 B17F.1    walk   255  -120   110    0     80   260  
20:54:40.158 B17F.0    stand  255  140    110    0     80   260  
20:54:41.162 B17F.1    walk   255  -120   100    120   80   260  
20:54:41.162 B17F.0    stand  255  140    110    0     80   260  
20:54:42.157 B17F.1    walk   255  -120   110    118   80   260  
20:54:42.157 B17F.0    stand  255  140    110    0     80   260  
20:54:43.160 B17F.1    walk   255  -120   110    110   80   260  
20:54:43.160 B17F.0    stand  255  140    110    0     80   260  
20:54:44.157 B17F.1    walk   255  -120   100    113   80   260  
20:54:44.157 B17F.0    stand  255  140    110    0     80   260  
20:54:45.181 B17F.1    walk   255  -130   100    115   80   270  
20:54:45.181 B17F.0    stand  255  140    110    0     80   270  
20:54:46.165 B17F.1    walk   255  -120   100    113   80   260  
20:54:46.165 B17F.0    stand  255  140    110    0     80   260  
20:54:47.060 B17F.1    walk   255  -130   100    0     80   270  
20:54:47.060 B17F.0    stand  255  140    110    83    80   270  
20:54:48.056 B17F.1    walk   255  -120   100    109   80   260  
20:54:48.056 B17F.0    stand  255  140    100    0     80   260  
20:54:49.060 B17F.1    walk   255  -120   100    0     80   260  
20:54:49.060 B17F.0    stand  255  140    90     84    80   260  
20:54:50.058 B17F.1    walk   255  -120   110    110   80   260  
20:54:50.058 B17F.0    stand  255  150    100    0     80   270  
20:54:51.060 B17F.1    walk   255  -120   110    117   80   270  
20:54:51.060 B17F.0    stand  255  150    100    0     80   270  
20:54:52.058 B17F.1    walk   255  -120   110    116   80   270  
20:54:52.058 B17F.0    stand  255  150    100    0     80   270  
20:54:53.058 B17F.1    walk   255  -110   100    107   80   260  
20:54:53.058 B17F.0    stand  255  150    100    0     80   260  
20:54:54.060 B17F.1    walk   255  -120   100    111   80   270  
20:54:54.060 B17F.0    stand  255  150    100    0     80   270  
20:54:55.064 B17F.1    walk   255  -140   110    93    80   290  
20:54:55.064 B17F.0    stand  255  150    100    0     80   290  
20:54:56.062 B17F.1    stand  255  -140   120    105   80   290  
20:54:56.062 B17F.0    stand  255  140    60     0     80   286  
20:54:57.065 B17F.1    stand  255  -110   100    111   80   253  
20:54:57.065 B17F.0    stand  255  160    70     0     80   271  
20:54:58.074 B17F.1    stand  255  -120   100    109   80   281  
20:54:58.074 B17F.0    stand  255  160    80     0     80   280  
20:54:58.967 B17F.1    stand  255  -130   110    106   80   291  
20:54:58.967 B17F.0    stand  255  160    80     0     80   291  
20:54:59.958 B17F.1    stand  255  -130   110    105   80   291  
20:54:59.958 B17F.0    stand  255  160    80     0     80   291  
20:55:00.961 B17F.1    stand  255  -120   110    112   80   281  
20:55:00.961 B17F.0    stand  255  160    80     0     80   281  
20:55:01.963 B17F.1    stand  255  -120   110    104   80   281  
20:55:01.963 B17F.0    stand  255  160    80     0     80   281  
20:55:02.973 B17F.1    stand  255  -120   100    105   80   280  
20:55:02.973 B17F.0    stand  255  150    90     0     80   270  
20:55:03.963 B17F.1    stand  255  -130   120    93    80   281  
20:55:03.963 B17F.0    stand  255  140    80     0     80   272  
20:55:04.964 B17F.1    stand  255  -150   80     89    80   290  
20:55:04.964 B17F.0    stand  255  150    80     0     80   300  
20:55:05.965 B17F.1    stand  255  -100   20     99    80   257  
20:55:05.965 B17F.0    stand  255  150    80     0     80   257  
20:55:06.966 B17F.1    stand  255  -110   20     0     80   266  
20:55:06.966 B17F.0    stand  255  150    80     0     80   266  
20:55:07.966 B17F.1    stand  255  -110   20     0     80   266  
20:55:07.966 B17F.0    stand  255  140    60     0     80   253  
20:55:08.970 B17F.1    stand  255  -110   20     0     80   253  
20:55:08.970 B17F.0    stand  255  140    90     0     80   259  
20:55:09.979 B17F.1    stand  255  -110   20     0     80   259  
20:55:09.979 B17F.0    stand  255  140    90     0     80   259  
20:55:10.862 B17F.1    stand  255  -70    0      0     80   228  
20:55:10.862 B17F.0    stand  255  140    90     0     80   228  
20:55:11.869 B17F.1    stand  255  -80    0      0     80   237  
20:55:11.869 B17F.0    stand  255  140    90     0     80   237  
20:55:12.865 B17F.1    stand  255  -100   20     0     80   250  
20:55:12.865 B17F.0    stand  255  140    90     0     80   250  
20:55:13.864 B17F.1    stand  255  -110   20     54    80   259  
20:55:13.864 B17F.0    stand  255  140    30     0     80   250  
20:55:14.864 B17F.1    stand  255  -110   30     0     80   250  
20:55:14.864 B17F.0    stand  255  140    30     0     80   250  
20:55:15.871 B17F.1    stand  255  -60    10     0     80   200  
20:55:15.871 B17F.0    stand  255  130    30     0     80   191  
20:55:16.865 B17F.1    stand  255  -60    10     0     80   191  
20:55:16.865 B17F.0    stand  255  80     0      0     80   140  
20:55:17.884 B17F.1    stand  255  -60    10     0     80   140  
20:55:17.884 B17F.0    stand  255  70     0      0     80   130  
20:55:18.886 B17F.1    stand  255  -60    10     0     80   130  
20:55:18.886 B17F.0    stand  255  70     0      0     80   130  
20:55:19.887 B17F.1    stand  255  -60    10     0     80   130  
20:55:19.887 B17F.0    stand  255  70     0      0     80   130  
20:55:20.785 B17F.0    stand  255  70     0      0     80   0    
20:55:20.785 B17F.1    stand  255  -60    10     0     80   130  
20:55:21.787 B17F.0    stand  255  70     0      0     80   130  
20:55:21.787 B17F.1    stand  255  -60    10     0     80   130  
20:55:22.789 B17F.1    stand  255  -60    10     0     80   0    
20:55:22.789 B17F.0    stand  255  70     0      0     80   130  
20:55:23.781 B17F.1    stand  255  -60    10     0     80   130  
20:55:23.781 B17F.0    stand  255  70     0      0     80   130  
20:55:24.789 B17F.1    stand  255  -60    10     0     80   130  
20:55:24.789 B17F.0    stand  255  70     0      0     80   130  
20:55:25.785 B17F.1    stand  255  -70    10     0     80   140  
20:55:25.785 B17F.0    stand  255  70     0      0     80   140  
20:55:26.787 B17F.1    stand  255  -70    10     0     80   140  
20:55:26.787 B17F.0    stand  255  70     0      0     80   140  
20:55:27.785 B17F.1    stand  255  -70    10     0     80   140  
20:55:27.785 B17F.0    stand  255  70     0      0     80   140  
20:55:28.794 B17F.1    stand  255  -70    10     0     80   140  
20:55:28.794 B17F.0    stand  255  70     0      0     80   140  
20:55:29.796 B17F.1    stand  255  -90    10     61    80   160  
20:55:29.796 B17F.0    stand  255  70     0      0     80   160  
20:55:30.808 B17F.1    stand  255  -160   40     0     80   233  
20:55:30.808 B17F.0    stand  255  70     0      0     80   233  
20:55:31.791 B17F.0    stand  255  70     0      0     80   0    
20:55:31.791 B17F.1    stand  255  -140   30     44    80   212  
20:55:32.692 B17F.0    stand  255  70     0      0     80   212  
20:55:32.692 B17F.1    stand  255  -130   30     52    80   202  
20:55:33.754 B17F.0    stand  255  70     0      0     80   202  
20:55:33.754 B17F.1    stand  255  -150   40     40    80   223  
20:55:34.699 B17F.0    stand  255  70     0      0     80   223  
20:55:34.699 B17F.1    stand  255  -80    20     0     80   151  
20:55:35.697 B17F.0    stand  255  70     0      0     80   151  
20:55:35.697 B17F.1    stand  255  -60    10     0     80   130  
20:55:36.704 B17F.0    stand  255  70     0      0     80   130  
20:55:36.704 B17F.1    stand  255  -60    10     0     80   130  
20:55:37.701 B17F.0    stand  255  70     0      0     80   130  
20:55:37.701 B17F.1    stand  255  -60    10     0     80   130  
20:55:38.711 B17F.0    stand  255  70     0      0     80   130  
20:55:38.711 B17F.1    stand  255  -70    10     0     80   140  
20:55:39.701 B17F.1    stand  255  -70    10     0     80   0    
20:55:39.701 B17F.0    stand  255  70     0      0     80   140  
20:55:40.703 B17F.1    stand  255  -70    10     0     80   140  
20:55:40.703 B17F.0    stand  255  70     0      0     80   140  
20:55:41.707 B17F.1    stand  255  -70    10     0     80   140  
20:55:41.707 B17F.0    stand  255  70     0      0     80   140  
20:55:42.610 B17F.0    stand  255  70     0      0     80   0    
20:55:42.610 B17F.1    stand  255  -70    10     0     80   140  
20:55:43.609 B17F.1    stand  255  -70    10     0     80   0    
20:55:43.609 B17F.0    stand  255  70     0      0     80   140  
20:55:44.608 B17F.0    stand  255  70     0      0     80   0    
20:55:44.608 B17F.1    stand  255  -70    10     0     80   140  
20:55:45.608 B17F.0    stand  255  70     0      0     80   140  
20:55:45.608 B17F.1    stand  255  -70    10     0     80   140  
20:55:46.617 B17F.0    stand  255  70     0      0     80   140  
20:55:46.617 B17F.1    stand  255  -70    10     0     80   140  
20:55:47.607 B17F.0    stand  255  70     0      0     80   140  
20:55:47.607 B17F.1    stand  255  -70    10     0     80   140  
20:55:48.610 B17F.0    stand  255  70     0      0     80   140  
20:55:48.610 B17F.1    stand  255  -60    10     0     80   130  
20:55:49.619 B17F.0    stand  255  70     0      0     80   130  
20:55:49.619 B17F.1    stand  255  -60    10     0     80   130  
20:55:50.612 B17F.0    stand  255  70     0      0     80   130  
20:55:50.612 B17F.1    stand  255  -70    20     0     80   141  
20:55:51.611 B17F.1    stand  255  -70    20     0     80   0    
20:55:51.611 B17F.0    stand  255  70     0      0     80   141  
20:55:52.620 B17F.1    stand  255  -70    20     0     80   141  
20:55:52.620 B17F.0    stand  255  70     0      0     80   141  
20:55:53.616 B17F.1    stand  255  -70    20     0     80   141  
20:55:53.616 B17F.0    stand  255  70     0      0     80   141  
20:55:54.507 B17F.0    stand  255  70     0      0     80   0    
20:55:54.507 B17F.1    stand  255  -70    20     0     80   141  
20:55:55.509 B17F.0    stand  255  70     0      0     80   141  
20:55:55.509 B17F.1    stand  255  -70    20     0     80   141  
20:55:56.510 B17F.0    stand  255  70     0      0     80   141  
20:55:56.510 B17F.1    stand  255  -70    20     0     80   141  
20:55:57.514 B17F.0    stand  255  70     0      0     80   141  
20:55:57.514 B17F.1    stand  255  -70    20     0     80   141  
20:55:58.511 B17F.0    stand  255  70     0      0     80   141  
20:55:58.511 B17F.1    stand  255  -70    20     0     80   141  
20:55:59.514 B17F.0    stand  255  70     0      0     80   141  
20:55:59.514 B17F.1    stand  255  -70    20     0     80   141  
20:56:00.520 B17F.0    stand  255  70     0      0     80   141  
20:56:00.520 B17F.1    stand  255  -70    20     0     80   141  
20:56:01.513 B17F.1    stand  255  -70    20     0     80   0    
20:56:01.513 B17F.0    stand  255  70     0      0     80   141  
20:56:02.516 B17F.1    stand  255  -70    20     0     80   141  
20:56:02.516 B17F.0    stand  255  70     0      0     80   141  
20:56:03.519 B17F.1    stand  255  -70    20     0     80   141  
20:56:03.519 B17F.0    stand  255  70     0      0     80   141  
20:56:04.520 B17F.0    stand  255  70     0      0     80   0    
20:56:04.520 B17F.1    stand  255  -70    20     0     80   141  
20:56:05.519 B17F.0    stand  255  70     0      0     80   141  
20:56:05.519 B17F.1    stand  255  -70    20     0     80   141  
20:56:06.413 B17F.0    stand  255  70     0      0     80   141  
20:56:06.413 B17F.1    stand  255  -70    20     0     80   141  
20:56:07.414 B17F.0    stand  255  70     0      0     80   141  
20:56:07.414 B17F.1    stand  255  -90    20     0     80   161  
20:56:08.420 B17F.0    stand  255  70     0      0     80   161  
20:56:08.420 B17F.1    stand  255  -90    20     0     80   161  
20:56:09.416 B17F.0    stand  255  70     0      0     80   161  
20:56:09.416 B17F.1    stand  255  -120   40     0     80   194  
20:56:10.416 B17F.0    stand  255  70     0      0     80   194  
20:56:10.416 B17F.1    stand  255  -140   60     73    80   218  
20:56:11.417 B17F.0    stand  255  70     0      0     80   218  
20:56:11.417 B17F.1    walk   255  -160   120    73    80   259  
20:56:12.421 B17F.0    stand  255  70     0      0     80   259  
20:56:12.421 B17F.1    walk   255  -160   180    136   80   292  
20:56:13.421 B17F.0    stand  255  70     0      0     80   292  
20:56:13.421 B17F.1    walk   255  -140   180    132   80   276  
20:56:14.418 B17F.0    stand  255  70     0      0     80   276  
20:56:14.418 B17F.1    walk   255  -140   180    0     80   276  
20:56:15.425 B17F.0    stand  255  70     0      0     80   276  
20:56:15.425 B17F.1    stand  255  -140   180    0     80   276  
20:56:16.421 B17F.0    stand  255  70     0      0     80   276  
20:56:16.421 B17F.1    stand  255  -140   180    0     80   276  
20:56:17.427 B17F.0    stand  255  70     0      0     80   276  
20:56:17.427 B17F.1    stand  255  -160   190    127   80   298  
20:56:18.316 B17F.0    stand  255  70     0      0     80   298  
20:56:18.316 B17F.1    stand  255  -190   260    101   80   367  
20:56:19.318 B17F.0    stand  255  70     0      0     80   367  
20:56:19.318 B17F.1    stand  255  -210   330    0     80   432  
20:56:20.325 B17F.0    stand  255  70     0      0     80   432  
20:56:20.325 B17F.1    stand  255  -210   330    0     80   432  
20:56:21.322 B17F.1    stand  255  -200   320    0     80   14   
20:56:21.322 B17F.0    stand  255  70     0      0     80   418  
20:56:22.333 B17F.1    stand  255  -200   320    0     80   418  
20:56:22.333 B17F.0    stand  255  70     0      0     80   418  
20:56:23.335 B17F.1    stand  255  -200   320    0     80   418  
20:56:23.335 B17F.0    stand  255  80     10     0     80   417  
20:56:24.333 B17F.0    stand  255  150    90     0     80   106  
20:56:24.333 B17F.1    stand  8    -200   320    0     80   418  
20:56:25.386 B17F.0    walk   255  140    80     74    80   416  
20:56:26.345 B17F.0    walk   255  110    90     0     80   31   
20:56:27.247 B17F.0    walk   255  120    100    88    80   14   
20:56:28.245 B17F.0    walk   255  120    90     0     80   10   
20:56:29.244 B17F.0    stand  255  150    90     0     80   30   
20:56:30.252 B17F.0    stand  255  160    110    0     80   22   
20:56:31.245 B17F.0    stand  255  120    100    0     80   41   
20:56:32.304 B17F.0    stand  255  130    80     70    80   22   
20:56:33.254 B17F.0    stand  255  130    80     86    80   0    
20:56:34.248 B17F.0    stand  255  130    80     0     80   0    
20:56:35.260 B17F.0    stand  255  150    90     77    80   22   
20:56:36.257 B17F.0    stand  255  160    110    0     80   22   
20:56:37.252 B17F.0    stand  255  130    110    0     80   30   
20:56:38.157 B17F.0    stand  255  150    90     0     80   28   
20:56:39.155 B17F.0    stand  255  140    80     0     80   14   
20:56:40.155 B17F.0    stand  255  150    90     0     80   14   
20:56:41.162 B17F.0    stand  255  150    90     0     80   0    
20:56:42.158 B17F.0    stand  255  150    80     0     80   10   
20:56:43.160 B17F.0    stand  255  150    80     0     80   0    
20:56:44.160 B17F.0    stand  255  150    90     0     80   10   
20:56:45.165 B17F.0    stand  255  150    90     0     80   0    
20:56:46.161 B17F.0    stand  255  150    90     0     80   0    
20:56:47.164 B17F.0    stand  255  150    90     0     80   0    
20:56:48.164 B17F.0    stand  255  120    100    84    80   31   
20:56:49.166 B17F.0    stand  255  130    90     90    80   14   
20:56:50.061 B17F.0    stand  255  150    80     78    80   22   
20:56:51.060 B17F.0    stand  255  130    70     77    80   22   
20:56:52.062 B17F.0    stand  255  150    110    0     80   44   
20:56:53.069 B17F.0    stand  255  130    80     75    80   36   
20:56:54.035 B17F.0    stand  255  90     50     104   80   50   
20:56:55.031 B17F.0    walk   255  0      60     92    80   90   
20:56:56.036 B17F.0    walk   255  -70    60     50    80   70   
20:56:57.033 B17F.0    walk   255  -110   50     38    80   41   
20:56:58.036 B17F.0    walk   255  -130   30     0     80   28   
20:56:59.033 B17F.0    walk   255  -70    20     0     80   60   
20:57:00.040 B17F.0    sit    255  -70    20     0     80   0    
20:57:01.036 B17F.0    sit    255  -70    20     0     80   0    
20:57:02.116 B17F.0    sitgnd 255  -70    10     0     80   10   
20:57:02.957 B17F.0    sitgnd 255  -70    10     0     80   0    
20:57:03.959 B17F.0    sitgnd 255  -80    10     0     80   10   
20:57:04.445 B17F.0    sitgnd 255  -90    20     61    80   14   
20:57:05.021 B17F.0    sit    255  -60    40     83    80   36   
20:57:05.969 B17F.0    sit    255  20     60     83    80   82   
20:57:06.968 B17F.0    walk   255  90     60     94    80   70   
20:57:07.969 B17F.0    walk   255  110    70     93    80   22   
20:57:08.965 B17F.0    walk   255  140    80     66    80   31   
20:57:09.970 B17F.0    walk   255  140    70     61    80   10   
20:57:10.972 B17F.0    sit    255  140    90     0     80   20   
20:57:11.970 B17F.0    walk   255  150    80     0     80   14   
20:57:12.970 B17F.0    sit    255  130    80     0     80   20   
20:57:13.873 B17F.0    sit    255  140    100    0     80   22   
20:57:14.872 B17F.0    sit    255  130    100    0     80   10   
20:57:15.868 B17F.0    sit    255  130    100    0     80   0    
20:57:16.869 B17F.0    sit    255  130    100    0     80   0    
20:57:17.869 B17F.0    sit    255  130    120    0     80   20   
20:57:18.874 B17F.0    sit    255  120    110    0     80   14   
20:57:19.880 B17F.0    sit    255  120    110    0     80   0    
20:57:20.888 B17F.0    sit    255  130    120    0     80   14   
20:57:21.874 B17F.0    sit    255  130    120    0     80   0    
20:57:22.875 B17F.0    sit    255  130    120    0     80   0    
20:57:23.877 B17F.0    sit    255  130    130    0     80   10   
20:57:24.877 B17F.0    sit    255  130    130    0     80   0    
20:57:25.776 B17F.0    sit    255  130    120    0     80   10   
20:57:26.780 B17F.0    sit    255  130    100    0     80   20   
20:57:27.781 B17F.0    sit    255  140    80     0     80   22   
20:57:28.782 B17F.0    stand  255  140    70     71    80   10   
20:57:29.787 B17F.0    stand  255  140    70     74    80   0    
20:57:30.781 B17F.0    stand  255  140    70     58    80   0    
20:57:31.832 B17F.0    stand  255  150    70     41    80   10   
20:57:32.782 B17F.0    stand  255  150    90     60    80   20   
20:57:33.785 B17F.0    stand  255  150    90     0     80   0    
20:57:34.785 B17F.0    stand  255  150    90     0     80   0    
20:57:35.786 B17F.0    stand  255  150    90     0     80   0    
20:57:36.693 B17F.0    stand  255  130    100    66    80   22   
20:57:37.690 B17F.0    stand  255  140    70     73    80   31   
20:57:38.685 B17F.0    stand  255  120    80     87    80   22   
20:57:39.696 B17F.0    stand  255  100    70     88    80   22   
20:57:40.686 B17F.0    stand  255  70     60     83    80   31   
20:57:41.644 B17F.0    walk   255  -60    70     75    80   130  
20:57:42.645 B17F.0    walk   255  -70    40     0     80   31   
20:57:43.659 B17F.0    walk   255  -60    30     0     80   14   
20:57:44.650 B17F.0    stand  255  -60    30     0     80   0    
20:57:45.652 B17F.0    sit    255  -60    20     21    80   10   
20:57:46.654 B17F.0    sit    255  -80    20     65    80   20   
20:57:47.655 B17F.0    sit    255  -60    40     99    80   28   
20:57:48.655 B17F.0    walk   255  30     90     80    80   102  
20:57:49.656 B17F.0    walk   255  80     70     91    80   53   
20:57:50.665 B17F.0    walk   255  120    90     53    80   44   
20:57:51.699 B17F.0    walk   255  140    80     59    80   22   
20:57:52.556 B17F.0    sit    255  130    110    78    80   31   
20:57:53.558 B17F.0    sit    255  130    110    0     80   0    
20:57:54.562 B17F.0    sit    255  130    120    0     80   10   
20:57:55.558 B17F.0    sit    255  130    110    0     80   10   
20:57:56.573 B17F.0    sit    255  130    110    0     80   0    
20:57:57.566 B17F.0    sit    255  130    120    0     80   10   
20:57:58.610 B17F.0    sit    255  130    120    0     80   0    
20:57:59.598 B17F.0    sit    255  130    120    0     80   0    
20:58:00.493 B17F.0    sit    255  120    110    0     80   14   
20:58:01.493 B17F.0    sit    255  120    100    0     80   10   
20:58:02.497 B17F.0    sit    255  130    100    64    80   10   
20:58:03.493 B17F.0    sit    255  130    110    85    80   10   
20:58:04.497 B17F.0    sit    255  120    120    0     80   14   
20:58:05.496 B17F.0    sit    255  120    120    70    80   0    
20:58:06.498 B17F.0    sit    255  130    120    0     80   10   
20:58:07.500 B17F.0    sit    255  130    120    74    80   0    
20:58:08.500 B17F.0    stand  255  150    70     62    80   53   
20:58:09.498 B17F.0    stand  255  160    80     32    80   14   
20:58:10.500 B17F.0    stand  255  160    90     64    80   10   
20:58:11.510 B17F.0    stand  255  140    90     61    80   20   
20:58:12.397 B17F.0    stand  255  150    120    62    80   31   
20:58:13.396 B17F.0    stand  255  140    100    0     80   22   
20:58:14.408 B17F.0    stand  255  140    60     0     80   40   
20:58:15.419 B17F.0    stand  255  140    60     0     80   0    
20:58:16.411 B17F.0    stand  255  140    130    0     80   70   
20:58:17.414 B17F.0    stand  255  140    130    0     80   0    
20:58:18.412 B17F.0    stand  255  140    130    0     80   0    
20:58:19.413 B17F.0    stand  255  140    130    0     80   0    
20:58:20.421 B17F.0    stand  255  140    130    0     80   0    
20:58:21.417 B17F.0    stand  255  140    130    0     80   0    
20:58:22.319 B17F.0    stand  255  140    130    0     80   0    
20:58:23.316 B17F.0    stand  255  140    130    0     80   0    
20:58:24.319 B17F.0    stand  255  140    130    0     80   0    
20:58:25.347 B17F.0    stand  255  140    130    0     80   0    
20:58:26.329 B17F.0    stand  255  140    130    0     80   0    
20:58:27.320 B17F.0    stand  255  140    130    0     80   0    
20:58:28.322 B17F.0    stand  255  140    120    0     80   10   
20:58:29.333 B17F.0    stand  255  140    120    0     80   0    
20:58:30.256 B17F.0    stand  255  140    120    0     80   0    
20:58:31.252 B17F.0    stand  255  140    120    0     80   0    
20:58:32.313 B17F.0    stand  255  140    120    0     80   0    
20:58:33.264 B17F.0    stand  255  140    120    0     80   0    
20:58:34.254 B17F.0    stand  255  140    120    0     80   0    
20:58:35.256 B17F.0    stand  255  140    120    0     80   0    
20:58:36.258 B17F.0    stand  255  140    120    0     80   0    
20:58:37.260 B17F.0    stand  255  140    120    0     80   0    
20:58:38.260 B17F.0    stand  255  140    130    0     80   10   
20:58:39.261 B17F.0    stand  255  140    130    0     80   0    
20:58:40.264 B17F.0    stand  255  140    130    0     80   0    
20:58:41.260 B17F.0    stand  255  140    130    0     80   0    
20:58:42.154 B17F.0    stand  255  140    130    0     80   0    
20:58:43.161 B17F.0    stand  255  140    130    0     80   0    
20:58:44.156 B17F.0    stand  255  140    110    65    80   20   
20:58:45.158 B17F.0    stand  255  140    100    0     80   10   
20:58:46.124 B17F.0    stand  255  150    90     91    80   14   
20:58:47.124 B17F.0    stand  255  160    80     0     80   14   
20:58:48.124 B17F.0    stand  255  140    120    0     80   44   
20:58:49.130 B17F.0    stand  255  140    120    0     80   0    
20:58:50.128 B17F.0    stand  255  140    120    0     80   0    
20:58:51.128 B17F.0    stand  255  140    120    0     80   0    
20:58:52.133 B17F.0    stand  255  140    120    0     80   0    
20:58:53.128 B17F.0    stand  255  140    120    0     80   0    
20:58:54.129 B17F.0    stand  255  140    120    0     80   0    
20:58:55.139 B17F.0    stand  255  140    120    0     80   0    
20:58:56.136 B17F.0    stand  255  140    120    0     80   0    
20:58:57.134 B17F.0    stand  255  140    120    0     80   0    
20:58:58.028 B17F.0    stand  255  140    120    0     80   0    
20:58:59.027 B17F.0    stand  255  140    110    0     80   10   
20:59:00.028 B17F.0    stand  255  140    70     61    80   40   
20:59:01.031 B17F.0    stand  255  140    80     0     80   10   
20:59:02.045 B17F.0    stand  255  140    80     0     80   0    
20:59:03.040 B17F.0    stand  255  140    80     0     80   0    
20:59:04.045 B17F.0    stand  255  140    70     0     80   10   
20:59:05.045 B17F.0    stand  255  140    70     0     80   0    
20:59:06.044 B17F.0    stand  255  140    70     0     80   0    
20:59:07.045 B17F.0    stand  255  140    70     0     80   0    
20:59:07.945 B17F.0    stand  255  140    70     0     80   0    
20:59:08.949 B17F.0    stand  255  140    70     0     80   0    
20:59:09.960 B17F.0    stand  255  140    70     0     80   0    
20:59:10.949 B17F.0    stand  255  140    70     0     80   0    
20:59:11.949 B17F.0    stand  255  140    70     0     80   0    
20:59:12.950 B17F.0    stand  255  140    70     0     80   0    
20:59:13.960 B17F.0    stand  255  140    70     0     80   0    
20:59:14.953 B17F.0    stand  255  140    70     0     80   0    
20:59:15.956 B17F.0    stand  255  140    70     0     80   0    
20:59:16.955 B17F.0    stand  255  140    70     0     80   0    
20:59:17.869 B17F.0    stand  255  140    70     0     80   0    
20:59:18.868 B17F.0    stand  255  140    70     0     80   0    
20:59:19.884 B17F.0    stand  255  140    70     0     80   0    
20:59:20.871 B17F.0    stand  255  140    70     0     80   0    
20:59:21.875 B17F.0    stand  255  140    130    0     80   60   
20:59:22.876 B17F.0    stand  255  140    130    0     80   0    
20:59:23.889 B17F.0    stand  255  140    130    0     80   0    
20:59:24.873 B17F.0    stand  255  140    130    0     80   0    
20:59:25.874 B17F.0    stand  255  140    120    0     80   10   
20:59:26.875 B17F.0    stand  255  140    120    0     80   0    
20:59:27.880 B17F.0    stand  255  140    120    0     80   0    
20:59:28.881 B17F.0    stand  255  140    120    0     80   0    
20:59:29.775 B17F.0    stand  255  140    120    0     80   0    
20:59:30.772 B17F.0    stand  255  140    120    0     80   0    
20:59:31.824 B17F.0    stand  255  130    120    0     80   10   
20:59:32.775 B17F.0    stand  255  140    120    0     80   10   
20:59:33.743 B17F.0    stand  255  140    120    0     80   0    
20:59:34.740 B17F.0    stand  255  140    120    0     80   0    
20:59:35.743 B17F.0    stand  255  140    120    0     80   0    
20:59:36.750 B17F.0    stand  255  140    130    0     80   10   
20:59:37.745 B17F.0    stand  255  140    130    0     80   0    
20:59:38.743 B17F.0    stand  255  140    130    0     80   0    
20:59:39.751 B17F.0    stand  255  140    130    0     80   0    
20:59:40.745 B17F.0    stand  255  140    130    0     80   0    
20:59:41.748 B17F.0    stand  255  140    130    0     80   0    
20:59:42.748 B17F.0    stand  255  140    130    0     80   0    
20:59:43.749 B17F.0    stand  255  140    130    0     80   0    
20:59:44.749 B17F.0    stand  255  140    130    0     80   0    
20:59:45.642 B17F.0    stand  255  140    130    0     80   0    
20:59:46.652 B17F.0    stand  255  140    130    0     80   0    
20:59:47.645 B17F.0    stand  255  150    130    0     80   10   
20:59:48.651 B17F.0    stand  255  120    130    0     80   30   
20:59:49.647 B17F.0    stand  255  140    120    65    80   22   
20:59:50.676 B17F.0    stand  255  140    100    0     80   20   
20:59:51.677 B17F.0    stand  255  140    100    0     80   0    
20:59:52.683 B17F.0    stand  255  140    100    0     80   0    
20:59:53.687 B17F.0    stand  255  140    100    0     80   0    
20:59:54.570 B17F.0    stand  255  140    100    0     80   0    
20:59:55.572 B17F.0    stand  255  140    100    0     80   0    
20:59:56.574 B17F.0    stand  255  140    100    0     80   0    
20:59:57.576 B17F.0    stand  255  140    100    0     80   0    
20:59:58.575 B17F.0    stand  255  140    100    0     80   0    
20:59:59.575 B17F.0    stand  255  140    100    0     80   0    
21:00:00.580 B17F.0    stand  255  140    120    0     80   20   
21:00:01.591 B17F.0    stand  255  140    130    0     80   10   
21:00:02.580 B17F.0    stand  255  130    120    0     80   14   
21:00:03.579 B17F.0    stand  255  130    120    0     80   0    
21:00:04.580 B17F.0    stand  255  140    120    0     80   10   
21:00:05.582 B17F.0    stand  255  130    130    0     80   14   
21:00:06.477 B17F.0    stand  255  130    130    0     80   0    
21:00:07.482 B17F.0    stand  255  130    130    0     80   0    
21:00:08.483 B17F.0    stand  255  130    130    0     80   0    
21:00:09.476 B17F.0    stand  255  130    130    0     80   0    
21:00:10.484 B17F.0    stand  255  130    130    0     80   0    
21:00:11.484 B17F.0    stand  255  130    130    0     80   0    
21:00:12.482 B17F.0    stand  255  130    130    0     80   0    
21:00:13.480 B17F.0    stand  255  130    130    0     80   0    
21:00:14.482 B17F.0    stand  255  130    130    0     80   0    
21:00:15.490 B17F.0    stand  255  130    130    0     80   0    
21:00:16.483 B17F.0    stand  255  130    130    0     80   0    
21:00:17.498 B17F.0    stand  255  130    130    0     80   0    
21:00:18.400 B17F.0    stand  255  130    130    0     80   0    
21:00:19.389 B17F.0    stand  255  130    130    0     80   0    
21:00:20.383 B17F.0    stand  255  130    130    0     80   0    
21:00:21.382 B17F.0    stand  255  130    130    0     80   0    
21:00:22.384 B17F.0    stand  255  130    130    0     80   0    
21:00:23.384 B17F.0    stand  255  130    130    0     80   0    
21:00:24.388 B17F.0    stand  255  130    130    0     80   0    
21:00:25.388 B17F.0    stand  255  130    130    0     80   0    
21:00:26.388 B17F.0    stand  255  130    130    0     80   0    
21:00:27.406 B17F.0    stand  255  140    120    0     80   14   
21:00:28.400 B17F.0    stand  255  150    130    0     80   14   
21:00:29.393 B17F.0    stand  255  140    130    0     80   10   
21:00:30.283 B17F.0    stand  255  140    130    0     80   0    
21:00:31.337 B17F.0    stand  255  130    130    0     80   10   
21:00:32.291 B17F.0    stand  255  130    130    0     80   0    
21:00:33.288 B17F.0    stand  255  140    120    0     80   14   
21:00:34.288 B17F.0    stand  255  140    120    0     80   0    
21:00:35.292 B17F.0    stand  255  140    120    0     80   0    
21:00:36.292 B17F.0    stand  255  140    120    0     80   0    
21:00:37.309 B17F.0    stand  255  140    120    0     80   0    
21:00:38.309 B17F.0    stand  255  140    120    0     80   0    
21:00:39.312 B17F.0    stand  255  140    120    0     80   0    
21:00:40.203 B17F.0    stand  255  140    120    0     80   0    
21:00:41.203 B17F.0    stand  255  140    120    0     80   0    
21:00:42.205 B17F.0    stand  255  150    110    81    80   14   
21:00:43.217 B17F.0    stand  255  150    100    73    80   10   
21:00:44.209 B17F.0    stand  255  140    110    0     80   14   
21:00:45.212 B17F.0    stand  255  140    120    0     80   10   
21:00:46.209 B17F.0    stand  255  140    120    0     80   0    
21:00:47.214 B17F.0    stand  255  140    120    0     80   0    
21:00:48.212 B17F.0    stand  255  130    130    0     80   14   
21:00:49.213 B17F.0    stand  255  140    130    0     80   10   
21:00:50.216 B17F.0    stand  255  140    130    0     80   0    
21:00:51.213 B17F.0    stand  255  140    130    0     80   0    
21:00:52.110 B17F.0    stand  255  140    130    0     80   0    
21:00:53.108 B17F.0    stand  255  140    130    0     80   0    
21:00:54.124 B17F.0    stand  255  140    130    0     80   0    
21:00:55.125 B17F.0    stand  255  140    130    0     80   0    
21:00:56.128 B17F.0    stand  255  140    130    0     80   0    
21:00:57.124 B17F.0    stand  255  140    130    0     80   0    
21:00:58.124 B17F.0    stand  255  140    110    0     80   20   
21:00:59.124 B17F.0    stand  255  140    110    0     80   0    
21:01:00.144 B17F.0    stand  255  140    110    0     80   0    
21:01:01.130 B17F.0    stand  255  140    110    0     80   0    
21:01:02.029 B17F.0    stand  255  150    120    0     80   14   
21:01:03.032 B17F.0    stand  255  150    130    0     80   10   
21:01:04.039 B17F.0    stand  255  150    130    0     80   0    
21:01:05.030 B17F.0    stand  255  150    130    0     80   0    
21:01:06.035 B17F.0    stand  255  140    110    83    80   22   
21:01:07.032 B17F.0    stand  255  120    90     67    80   28   
21:01:08.034 B17F.0    stand  255  120    110    0     80   20   
21:01:09.034 B17F.0    stand  255  120    100    0     80   10   
21:01:09.969 B17F.0    stand  255  120    90     0     80   10   
21:01:10.970 B17F.0    stand  255  120    100    84    80   10   
21:01:11.988 B17F.0    stand  255  150    80     64    80   36   
21:01:12.967 B17F.0    stand  255  140    110    0     80   31   
21:01:13.967 B17F.0    stand  255  140    110    0     80   0    
21:01:14.969 B17F.0    stand  255  140    110    0     80   0    
21:01:15.970 B17F.0    stand  255  140    120    0     80   10   
21:01:16.969 B17F.0    stand  255  140    120    0     80   0    
21:01:17.972 B17F.0    stand  255  140    120    0     80   0    
21:01:18.972 B17F.0    stand  255  140    120    0     80   0    
21:01:19.972 B17F.0    stand  255  140    110    0     80   10   
21:01:20.976 B17F.0    stand  255  160    110    78    80   20   
21:01:21.869 B17F.0    stand  255  170    70     78    80   41   
21:01:22.869 B17F.0    stand  255  170    90     0     80   20   
21:01:23.872 B17F.0    stand  255  170    90     0     80   0    
21:01:24.873 B17F.0    stand  255  170    90     0     80   0    
21:01:25.838 B17F.0    stand  255  160    80     0     80   14   
21:01:26.843 B17F.0    stand  255  180    130    73    80   53   
21:01:27.847 B17F.0    stand  255  190    200    81    80   70   
21:01:28.839 B17F.0    walk   255  120    250    94    80   86   
21:01:29.902 B17F.0    walk   255  20     280    93    80   104  
21:01:30.848 B17F.0    walk   255  -90    310    88    80   114  
21:01:31.841 B17F.0    walk   255  -210   330    123   80   121  
21:01:32.842 B17F.0    walk   255  -230   330    0     80   20   
21:01:33.852 B17F.0    walk   255  -220   320    0     80   14   
21:01:34.864 B17F.0    stand  255  -220   320    0     80   0    
21:01:35.849 B17F.0    stand  255  -220   320    0     80   0    
21:01:36.846 B17F.0    stand  255  -210   320    0     80   10   
21:01:37.745 B17F.0    stand  8    -210   320    0     80   0    
21:01:38.796 B17F.88   88     -    -      -      -     -    -    
21:01:39.750 B17F.88   88     -    -      -      -     -    -    
21:01:40.754 B17F.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 1321 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire
